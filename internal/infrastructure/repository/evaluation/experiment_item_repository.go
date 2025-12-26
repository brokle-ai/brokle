package evaluation

import (
	"context"

	"brokle/internal/core/domain/evaluation"
	"brokle/pkg/ulid"

	"gorm.io/gorm"
)

type ExperimentItemRepository struct {
	db *gorm.DB
}

func NewExperimentItemRepository(db *gorm.DB) *ExperimentItemRepository {
	return &ExperimentItemRepository{db: db}
}

func (r *ExperimentItemRepository) Create(ctx context.Context, item *evaluation.ExperimentItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *ExperimentItemRepository) CreateBatch(ctx context.Context, items []*evaluation.ExperimentItem) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(items, 100).Error
}

func (r *ExperimentItemRepository) List(ctx context.Context, experimentID ulid.ULID, limit, offset int) ([]*evaluation.ExperimentItem, int64, error) {
	var items []*evaluation.ExperimentItem
	var total int64

	baseQuery := r.db.WithContext(ctx).
		Model(&evaluation.ExperimentItem{}).
		Where("experiment_id = ?", experimentID.String())

	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	result := r.db.WithContext(ctx).
		Where("experiment_id = ?", experimentID.String()).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&items)

	if result.Error != nil {
		return nil, 0, result.Error
	}
	return items, total, nil
}

func (r *ExperimentItemRepository) CountByExperiment(ctx context.Context, experimentID ulid.ULID) (int64, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&evaluation.ExperimentItem{}).
		Where("experiment_id = ?", experimentID.String()).
		Count(&count)

	if result.Error != nil {
		return 0, result.Error
	}
	return count, nil
}
