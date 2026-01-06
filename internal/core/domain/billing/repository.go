package billing

import (
	"context"
	"time"

	"brokle/pkg/ulid"
)

// ============================================================================
// Usage & Billing Repositories
// ============================================================================

// UsageRepository handles usage tracking data access
type UsageRepository interface {
	InsertUsageRecord(ctx context.Context, record *UsageRecord) error
	GetUsageRecords(ctx context.Context, orgID ulid.ULID, start, end time.Time) ([]*UsageRecord, error)
	UpdateUsageRecord(ctx context.Context, recordID ulid.ULID, record *UsageRecord) error
}

// BillingRecordRepository handles billing records and summaries persistence
type BillingRecordRepository interface {
	// Billing records
	InsertBillingRecord(ctx context.Context, record *BillingRecord) error
	UpdateBillingRecord(ctx context.Context, recordID ulid.ULID, record *BillingRecord) error
	GetBillingRecord(ctx context.Context, recordID ulid.ULID) (*BillingRecord, error)
	GetBillingHistory(ctx context.Context, orgID ulid.ULID, start, end time.Time) ([]*BillingRecord, error)

	// Billing summaries
	InsertBillingSummary(ctx context.Context, summary *BillingSummary) error
	GetBillingSummary(ctx context.Context, orgID ulid.ULID, period string) (*BillingSummary, error)
	GetBillingSummaryHistory(ctx context.Context, orgID ulid.ULID, start, end time.Time) ([]*BillingSummary, error)
}

// QuotaRepository handles usage quota management
type QuotaRepository interface {
	GetUsageQuota(ctx context.Context, orgID ulid.ULID) (*UsageQuota, error)
	UpdateUsageQuota(ctx context.Context, orgID ulid.ULID, quota *UsageQuota) error
}

// ============================================================================
// Usage-Based Billing Repositories (Spans + GB + Scores)
// ============================================================================

// BillableUsageFilter defines filters for querying billable usage
type BillableUsageFilter struct {
	OrganizationID ulid.ULID
	ProjectID      *ulid.ULID // nil for org-level
	Start          time.Time
	End            time.Time
	Granularity    string // "hourly" or "daily"
}

// BillableUsageRepository handles billable usage data access (ClickHouse)
type BillableUsageRepository interface {
	// Get aggregated usage for a time range
	GetUsage(ctx context.Context, filter *BillableUsageFilter) ([]*BillableUsage, error)

	// Get usage summary (totals) for a period
	GetUsageSummary(ctx context.Context, filter *BillableUsageFilter) (*BillableUsageSummary, error)

	// Get usage breakdown by project
	GetUsageByProject(ctx context.Context, orgID ulid.ULID, start, end time.Time) ([]*BillableUsageSummary, error)
}

// PricingConfigRepository handles pricing configuration (PostgreSQL)
type PricingConfigRepository interface {
	GetByID(ctx context.Context, id ulid.ULID) (*PricingConfig, error)
	GetByName(ctx context.Context, name string) (*PricingConfig, error)
	GetDefault(ctx context.Context) (*PricingConfig, error)
	GetActive(ctx context.Context) ([]*PricingConfig, error)
	Create(ctx context.Context, config *PricingConfig) error
	Update(ctx context.Context, config *PricingConfig) error
}

// OrganizationBillingRepository handles org billing state (PostgreSQL)
type OrganizationBillingRepository interface {
	GetByOrgID(ctx context.Context, orgID ulid.ULID) (*OrganizationBilling, error)
	Create(ctx context.Context, billing *OrganizationBilling) error
	Update(ctx context.Context, billing *OrganizationBilling) error
	UpdateUsage(ctx context.Context, orgID ulid.ULID, spans, bytes, scores int64, cost float64) error
	ResetPeriod(ctx context.Context, orgID ulid.ULID, newCycleStart time.Time) error
}

// UsageBudgetRepository handles budget CRUD (PostgreSQL)
type UsageBudgetRepository interface {
	GetByID(ctx context.Context, id ulid.ULID) (*UsageBudget, error)
	GetByOrgID(ctx context.Context, orgID ulid.ULID) ([]*UsageBudget, error)
	GetByProjectID(ctx context.Context, projectID ulid.ULID) ([]*UsageBudget, error)
	GetActive(ctx context.Context, orgID ulid.ULID) ([]*UsageBudget, error)
	Create(ctx context.Context, budget *UsageBudget) error
	Update(ctx context.Context, budget *UsageBudget) error
	UpdateUsage(ctx context.Context, budgetID ulid.ULID, spans, bytes, scores int64, cost float64) error
	Delete(ctx context.Context, id ulid.ULID) error
}

// UsageAlertRepository handles alert history (PostgreSQL)
type UsageAlertRepository interface {
	GetByID(ctx context.Context, id ulid.ULID) (*UsageAlert, error)
	GetByOrgID(ctx context.Context, orgID ulid.ULID, limit int) ([]*UsageAlert, error)
	GetByBudgetID(ctx context.Context, budgetID ulid.ULID) ([]*UsageAlert, error)
	GetUnacknowledged(ctx context.Context, orgID ulid.ULID) ([]*UsageAlert, error)
	Create(ctx context.Context, alert *UsageAlert) error
	Acknowledge(ctx context.Context, id ulid.ULID) error
	Resolve(ctx context.Context, id ulid.ULID) error
	MarkNotificationSent(ctx context.Context, id ulid.ULID) error
}
