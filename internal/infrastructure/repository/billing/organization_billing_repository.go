package billing

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"brokle/internal/core/domain/billing"
	"brokle/internal/infrastructure/shared"
	"brokle/pkg/ulid"
)

type organizationBillingRepository struct {
	db *gorm.DB
}

func NewOrganizationBillingRepository(db *gorm.DB) billing.OrganizationBillingRepository {
	return &organizationBillingRepository{db: db}
}

// getDB extracts transaction from context if available
func (r *organizationBillingRepository) getDB(ctx context.Context) *gorm.DB {
	return shared.GetDB(ctx, r.db)
}

func (r *organizationBillingRepository) GetByOrgID(ctx context.Context, orgID ulid.ULID) (*billing.OrganizationBilling, error) {
	var orgBilling billing.OrganizationBilling
	err := r.getDB(ctx).WithContext(ctx).Where("organization_id = ?", orgID).First(&orgBilling).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, billing.NewBillingNotFoundError(orgID.String())
		}
		return nil, fmt.Errorf("get organization billing: %w", err)
	}
	return &orgBilling, nil
}

func (r *organizationBillingRepository) Create(ctx context.Context, orgBilling *billing.OrganizationBilling) error {
	return r.getDB(ctx).WithContext(ctx).Create(orgBilling).Error
}

func (r *organizationBillingRepository) Update(ctx context.Context, orgBilling *billing.OrganizationBilling) error {
	orgBilling.UpdatedAt = time.Now()
	return r.getDB(ctx).WithContext(ctx).Save(orgBilling).Error
}

func (r *organizationBillingRepository) UpdateUsage(ctx context.Context, orgID ulid.ULID, spans, bytes, scores int64, cost decimal.Decimal) error {
	return r.getDB(ctx).WithContext(ctx).
		Model(&billing.OrganizationBilling{}).
		Where("organization_id = ?", orgID).
		Updates(map[string]interface{}{
			"current_period_spans":  gorm.Expr("current_period_spans + ?", spans),
			"current_period_bytes":  gorm.Expr("current_period_bytes + ?", bytes),
			"current_period_scores": gorm.Expr("current_period_scores + ?", scores),
			"current_period_cost":   gorm.Expr("current_period_cost + ?", cost),
			"last_synced_at":        time.Now(),
			"updated_at":            time.Now(),
		}).Error
}

func (r *organizationBillingRepository) ResetPeriod(ctx context.Context, orgID ulid.ULID, newCycleStart time.Time) error {
	return r.getDB(ctx).WithContext(ctx).
		Model(&billing.OrganizationBilling{}).
		Where("organization_id = ?", orgID).
		Updates(map[string]interface{}{
			"billing_cycle_start":    newCycleStart,
			"current_period_spans":   0,
			"current_period_bytes":   0,
			"current_period_scores":  0,
			"current_period_cost":    0,
			"free_spans_remaining":   gorm.Expr("(SELECT free_spans FROM plans WHERE id = organization_billing.plan_id)"),
			"free_bytes_remaining":   gorm.Expr("(SELECT CAST(free_gb * 1073741824 AS BIGINT) FROM plans WHERE id = organization_billing.plan_id)"),
			"free_scores_remaining":  gorm.Expr("(SELECT free_scores FROM plans WHERE id = organization_billing.plan_id)"),
			"last_synced_at":         time.Now(),
			"updated_at":             time.Now(),
		}).Error
}
