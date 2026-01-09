package billing

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"brokle/internal/core/domain/billing"
	"brokle/pkg/ulid"
)

type planRepository struct {
	db *gorm.DB
}

func NewPlanRepository(db *gorm.DB) billing.PlanRepository {
	return &planRepository{db: db}
}

func (r *planRepository) GetByID(ctx context.Context, id ulid.ULID) (*billing.Plan, error) {
	var plan billing.Plan
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&plan).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("plan not found: %s", id)
		}
		return nil, fmt.Errorf("get plan: %w", err)
	}
	return &plan, nil
}

func (r *planRepository) GetByName(ctx context.Context, name string) (*billing.Plan, error) {
	var plan billing.Plan
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&plan).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("plan not found: %s", name)
		}
		return nil, fmt.Errorf("get plan by name: %w", err)
	}
	return &plan, nil
}

func (r *planRepository) GetDefault(ctx context.Context) (*billing.Plan, error) {
	var plan billing.Plan
	err := r.db.WithContext(ctx).Where("is_default = ? AND is_active = ?", true, true).First(&plan).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("no default plan found")
		}
		return nil, fmt.Errorf("get default plan: %w", err)
	}
	return &plan, nil
}

func (r *planRepository) GetActive(ctx context.Context) ([]*billing.Plan, error) {
	var plans []*billing.Plan
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Find(&plans).Error
	if err != nil {
		return nil, fmt.Errorf("get active plans: %w", err)
	}
	return plans, nil
}

func (r *planRepository) Create(ctx context.Context, plan *billing.Plan) error {
	return r.db.WithContext(ctx).Create(plan).Error
}

func (r *planRepository) Update(ctx context.Context, plan *billing.Plan) error {
	return r.db.WithContext(ctx).Save(plan).Error
}
