package evaluation

import (
	"context"
	"errors"

	"brokle/internal/core/domain/evaluation"
	"brokle/pkg/ulid"

	"gorm.io/gorm"
)

type DatasetItemRepository struct {
	db *gorm.DB
}

func NewDatasetItemRepository(db *gorm.DB) *DatasetItemRepository {
	return &DatasetItemRepository{db: db}
}

func (r *DatasetItemRepository) Create(ctx context.Context, item *evaluation.DatasetItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *DatasetItemRepository) CreateBatch(ctx context.Context, items []*evaluation.DatasetItem) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(items, 100).Error
}

func (r *DatasetItemRepository) GetByID(ctx context.Context, id ulid.ULID, datasetID ulid.ULID) (*evaluation.DatasetItem, error) {
	var item evaluation.DatasetItem
	result := r.db.WithContext(ctx).
		Where("id = ? AND dataset_id = ?", id.String(), datasetID.String()).
		First(&item)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, evaluation.ErrDatasetItemNotFound
		}
		return nil, result.Error
	}
	return &item, nil
}

func (r *DatasetItemRepository) GetByIDForProject(ctx context.Context, id ulid.ULID, projectID ulid.ULID) (*evaluation.DatasetItem, error) {
	var item evaluation.DatasetItem
	result := r.db.WithContext(ctx).
		Joins("JOIN datasets ON datasets.id = dataset_items.dataset_id").
		Where("dataset_items.id = ? AND datasets.project_id = ?", id.String(), projectID.String()).
		First(&item)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, evaluation.ErrDatasetItemNotFound
		}
		return nil, result.Error
	}
	return &item, nil
}

func (r *DatasetItemRepository) List(ctx context.Context, datasetID ulid.ULID, limit, offset int) ([]*evaluation.DatasetItem, int64, error) {
	var items []*evaluation.DatasetItem
	var total int64

	baseQuery := r.db.WithContext(ctx).
		Model(&evaluation.DatasetItem{}).
		Where("dataset_id = ?", datasetID.String())

	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	result := r.db.WithContext(ctx).
		Where("dataset_id = ?", datasetID.String()).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&items)

	if result.Error != nil {
		return nil, 0, result.Error
	}
	return items, total, nil
}

func (r *DatasetItemRepository) Delete(ctx context.Context, id ulid.ULID, datasetID ulid.ULID) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND dataset_id = ?", id.String(), datasetID.String()).
		Delete(&evaluation.DatasetItem{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return evaluation.ErrDatasetItemNotFound
	}
	return nil
}

func (r *DatasetItemRepository) CountByDataset(ctx context.Context, datasetID ulid.ULID) (int64, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&evaluation.DatasetItem{}).
		Where("dataset_id = ?", datasetID.String()).
		Count(&count)

	if result.Error != nil {
		return 0, result.Error
	}
	return count, nil
}
