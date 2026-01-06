package billing

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"brokle/internal/core/domain/billing"
	"brokle/pkg/ulid"
)

type pricingConfigRepository struct {
	db *gorm.DB
}

func NewPricingConfigRepository(db *gorm.DB) billing.PricingConfigRepository {
	return &pricingConfigRepository{db: db}
}

func (r *pricingConfigRepository) GetByID(ctx context.Context, id ulid.ULID) (*billing.PricingConfig, error) {
	var config billing.PricingConfig
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&config).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("pricing config not found: %s", id)
		}
		return nil, fmt.Errorf("get pricing config: %w", err)
	}
	return &config, nil
}

func (r *pricingConfigRepository) GetByName(ctx context.Context, name string) (*billing.PricingConfig, error) {
	var config billing.PricingConfig
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&config).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("pricing config not found: %s", name)
		}
		return nil, fmt.Errorf("get pricing config by name: %w", err)
	}
	return &config, nil
}

func (r *pricingConfigRepository) GetDefault(ctx context.Context) (*billing.PricingConfig, error) {
	var config billing.PricingConfig
	err := r.db.WithContext(ctx).Where("is_default = ? AND is_active = ?", true, true).First(&config).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("no default pricing config found")
		}
		return nil, fmt.Errorf("get default pricing config: %w", err)
	}
	return &config, nil
}

func (r *pricingConfigRepository) GetActive(ctx context.Context) ([]*billing.PricingConfig, error) {
	var configs []*billing.PricingConfig
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Find(&configs).Error
	if err != nil {
		return nil, fmt.Errorf("get active pricing configs: %w", err)
	}
	return configs, nil
}

func (r *pricingConfigRepository) Create(ctx context.Context, config *billing.PricingConfig) error {
	return r.db.WithContext(ctx).Create(config).Error
}

func (r *pricingConfigRepository) Update(ctx context.Context, config *billing.PricingConfig) error {
	return r.db.WithContext(ctx).Save(config).Error
}
