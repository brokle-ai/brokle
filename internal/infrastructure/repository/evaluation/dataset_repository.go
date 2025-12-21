package evaluation

import (
	"context"
	"errors"
	"strings"

	"brokle/internal/core/domain/evaluation"
	"brokle/pkg/ulid"

	"gorm.io/gorm"
)

type DatasetRepository struct {
	db *gorm.DB
}

func NewDatasetRepository(db *gorm.DB) *DatasetRepository {
	return &DatasetRepository{db: db}
}

func (r *DatasetRepository) Create(ctx context.Context, dataset *evaluation.Dataset) error {
	result := r.db.WithContext(ctx).Create(dataset)
	if result.Error != nil {
		if isUniqueViolation(result.Error) {
			return evaluation.ErrDatasetExists
		}
		return result.Error
	}
	return nil
}

func (r *DatasetRepository) GetByID(ctx context.Context, id ulid.ULID, projectID ulid.ULID) (*evaluation.Dataset, error) {
	var dataset evaluation.Dataset
	result := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ?", id.String(), projectID.String()).
		First(&dataset)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, evaluation.ErrDatasetNotFound
		}
		return nil, result.Error
	}
	return &dataset, nil
}

func (r *DatasetRepository) GetByName(ctx context.Context, projectID ulid.ULID, name string) (*evaluation.Dataset, error) {
	var dataset evaluation.Dataset
	result := r.db.WithContext(ctx).
		Where("project_id = ? AND name = ?", projectID.String(), name).
		First(&dataset)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &dataset, nil
}

func (r *DatasetRepository) List(ctx context.Context, projectID ulid.ULID) ([]*evaluation.Dataset, error) {
	var datasets []*evaluation.Dataset
	result := r.db.WithContext(ctx).
		Where("project_id = ?", projectID.String()).
		Order("created_at DESC").
		Find(&datasets)

	if result.Error != nil {
		return nil, result.Error
	}
	return datasets, nil
}

func (r *DatasetRepository) Update(ctx context.Context, dataset *evaluation.Dataset, projectID ulid.ULID) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ?", dataset.ID.String(), projectID.String()).
		Save(dataset)

	if result.Error != nil {
		if isUniqueViolation(result.Error) {
			return evaluation.ErrDatasetExists
		}
		return result.Error
	}

	if result.RowsAffected == 0 {
		return evaluation.ErrDatasetNotFound
	}
	return nil
}

func (r *DatasetRepository) Delete(ctx context.Context, id ulid.ULID, projectID ulid.ULID) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ?", id.String(), projectID.String()).
		Delete(&evaluation.Dataset{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return evaluation.ErrDatasetNotFound
	}
	return nil
}

func (r *DatasetRepository) ExistsByName(ctx context.Context, projectID ulid.ULID, name string) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&evaluation.Dataset{}).
		Where("project_id = ? AND name = ?", projectID.String(), name).
		Count(&count)

	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

func isDatasetUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "23505") ||
		strings.Contains(errStr, "unique constraint") ||
		strings.Contains(errStr, "duplicate key")
}
