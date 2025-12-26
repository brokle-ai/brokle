package evaluation

import (
	"context"
	"errors"

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

func (r *ExperimentRepository) List(ctx context.Context, projectID ulid.ULID, filter *evaluation.ExperimentFilter) ([]*evaluation.Experiment, error) {
	var experiments []*evaluation.Experiment

	query := r.db.WithContext(ctx).
		Where("project_id = ?", projectID.String())

	if filter != nil {
		if filter.DatasetID != nil {
			query = query.Where("dataset_id = ?", filter.DatasetID.String())
		}
		if filter.Status != nil {
			query = query.Where("status = ?", string(*filter.Status))
		}
	}

	result := query.Order("created_at DESC").Find(&experiments)
	if result.Error != nil {
		return nil, result.Error
	}
	return experiments, nil
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
