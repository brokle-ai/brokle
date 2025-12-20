package evaluation

import (
	"context"
	"errors"
	"strings"

	"brokle/internal/core/domain/evaluation"
	"brokle/pkg/ulid"

	"gorm.io/gorm"
)

type ScoreConfigRepository struct {
	db *gorm.DB
}

func NewScoreConfigRepository(db *gorm.DB) *ScoreConfigRepository {
	return &ScoreConfigRepository{db: db}
}

func (r *ScoreConfigRepository) Create(ctx context.Context, config *evaluation.ScoreConfig) error {
	result := r.db.WithContext(ctx).Create(config)
	if result.Error != nil {
		if isUniqueViolation(result.Error) {
			return evaluation.ErrScoreConfigExists
		}
		return result.Error
	}
	return nil
}

func (r *ScoreConfigRepository) GetByID(ctx context.Context, id ulid.ULID, projectID ulid.ULID) (*evaluation.ScoreConfig, error) {
	var config evaluation.ScoreConfig
	result := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ?", id.String(), projectID.String()).
		First(&config)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, evaluation.ErrScoreConfigNotFound
		}
		return nil, result.Error
	}
	return &config, nil
}

// GetByName returns nil, nil if not found (for uniqueness checks).
func (r *ScoreConfigRepository) GetByName(ctx context.Context, projectID ulid.ULID, name string) (*evaluation.ScoreConfig, error) {
	var config evaluation.ScoreConfig
	result := r.db.WithContext(ctx).
		Where("project_id = ? AND name = ?", projectID.String(), name).
		First(&config)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Not found is not an error for uniqueness checks
		}
		return nil, result.Error
	}
	return &config, nil
}

func (r *ScoreConfigRepository) List(ctx context.Context, projectID ulid.ULID) ([]*evaluation.ScoreConfig, error) {
	var configs []*evaluation.ScoreConfig
	result := r.db.WithContext(ctx).
		Where("project_id = ?", projectID.String()).
		Order("created_at DESC").
		Find(&configs)

	if result.Error != nil {
		return nil, result.Error
	}
	return configs, nil
}

func (r *ScoreConfigRepository) Update(ctx context.Context, config *evaluation.ScoreConfig, projectID ulid.ULID) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ?", config.ID.String(), projectID.String()).
		Save(config)

	if result.Error != nil {
		if isUniqueViolation(result.Error) {
			return evaluation.ErrScoreConfigExists
		}
		return result.Error
	}

	if result.RowsAffected == 0 {
		return evaluation.ErrScoreConfigNotFound
	}
	return nil
}

func (r *ScoreConfigRepository) Delete(ctx context.Context, id ulid.ULID, projectID ulid.ULID) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ?", id.String(), projectID.String()).
		Delete(&evaluation.ScoreConfig{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return evaluation.ErrScoreConfigNotFound
	}
	return nil
}

func (r *ScoreConfigRepository) ExistsByName(ctx context.Context, projectID ulid.ULID, name string) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&evaluation.ScoreConfig{}).
		Where("project_id = ? AND name = ?", projectID.String(), name).
		Count(&count)

	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "23505") ||
		strings.Contains(errStr, "unique constraint") ||
		strings.Contains(errStr, "duplicate key")
}
