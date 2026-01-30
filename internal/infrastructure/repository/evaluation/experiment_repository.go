package evaluation

import (
	"context"
	"errors"
	"strings"

	"brokle/internal/core/domain/evaluation"
	"brokle/pkg/ulid"

	"gorm.io/gorm"
)

type ExperimentRepository struct {
	db *gorm.DB
}

func NewExperimentRepository(db *gorm.DB) *ExperimentRepository {
	return &ExperimentRepository{db: db}
}

func (r *ExperimentRepository) Create(ctx context.Context, experiment *evaluation.Experiment) error {
	return r.db.WithContext(ctx).Create(experiment).Error
}

func (r *ExperimentRepository) GetByID(ctx context.Context, id ulid.ULID, projectID ulid.ULID) (*evaluation.Experiment, error) {
	var experiment evaluation.Experiment
	result := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ?", id.String(), projectID.String()).
		First(&experiment)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, evaluation.ErrExperimentNotFound
		}
		return nil, result.Error
	}
	return &experiment, nil
}

func (r *ExperimentRepository) List(ctx context.Context, projectID ulid.ULID, filter *evaluation.ExperimentFilter, offset, limit int) ([]*evaluation.Experiment, int64, error) {
	var experiments []*evaluation.Experiment
	var total int64

	query := r.db.WithContext(ctx).
		Where("project_id = ?", projectID.String())

	if filter != nil {
		if filter.DatasetID != nil {
			query = query.Where("dataset_id = ?", filter.DatasetID.String())
		}
		if filter.Status != nil {
			query = query.Where("status = ?", string(*filter.Status))
		}
		if filter.Search != nil && *filter.Search != "" {
			search := "%" + strings.ToLower(*filter.Search) + "%"
			query = query.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", search, search)
		}
		if len(filter.IDs) > 0 {
			idStrings := make([]string, len(filter.IDs))
			for i, id := range filter.IDs {
				idStrings[i] = id.String()
			}
			query = query.Where("id IN ?", idStrings)
		}
	}

	if err := query.Model(&evaluation.Experiment{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	result := query.Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&experiments)

	if result.Error != nil {
		return nil, 0, result.Error
	}
	return experiments, total, nil
}

func (r *ExperimentRepository) Update(ctx context.Context, experiment *evaluation.Experiment, projectID ulid.ULID) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ?", experiment.ID.String(), projectID.String()).
		Save(experiment)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return evaluation.ErrExperimentNotFound
	}
	return nil
}

func (r *ExperimentRepository) Delete(ctx context.Context, id ulid.ULID, projectID ulid.ULID) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ?", id.String(), projectID.String()).
		Delete(&evaluation.Experiment{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return evaluation.ErrExperimentNotFound
	}
	return nil
}
