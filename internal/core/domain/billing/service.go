package billing

import (
	"context"
	"time"

	"brokle/pkg/ulid"
)

// BillingService defines the interface for billing operations
type BillingService interface {
	// Usage recording
	RecordUsage(ctx context.Context, usage *CostMetric) error

	// Billing calculation
	CalculateBill(ctx context.Context, orgID ulid.ULID, period string) (*BillingSummary, error)
	GetBillingHistory(ctx context.Context, orgID ulid.ULID, start, end time.Time) ([]*BillingRecord, error)

	// Payment processing
	ProcessPayment(ctx context.Context, billingRecordID ulid.ULID) error
	CreateBillingRecord(ctx context.Context, summary *BillingSummary) (*BillingRecord, error)

	// Quota management
	CheckUsageQuotas(ctx context.Context, orgID ulid.ULID) (*QuotaStatus, error)

	// Health monitoring
	GetHealth() map[string]interface{}
}

// OrganizationService provides organization-related data for billing context
type OrganizationService interface {
	GetBillingTier(ctx context.Context, orgID ulid.ULID) (string, error)
	GetDiscountRate(ctx context.Context, orgID ulid.ULID) (float64, error)
	GetPaymentMethod(ctx context.Context, orgID ulid.ULID) (*PaymentMethod, error)
}

// QuotaStatus represents the current quota status for an organization
type QuotaStatus struct {
	Status               string    `json:"status"`
	RequestsUsagePercent float64   `json:"requests_usage_percent"`
	TokensUsagePercent   float64   `json:"tokens_usage_percent"`
	CostUsagePercent     float64   `json:"cost_usage_percent"`
	OrganizationID       ulid.ULID `json:"organization_id"`
	RequestsOK           bool      `json:"requests_ok"`
	TokensOK             bool      `json:"tokens_ok"`
	CostOK               bool      `json:"cost_ok"`
}

// ============================================================================
// Usage-Based Billing Services (Spans + GB + Scores)
// ============================================================================

// UsageOverview represents the current usage overview for display
type UsageOverview struct {
	OrganizationID ulid.ULID `json:"organization_id"`
	PeriodStart    time.Time `json:"period_start"`
	PeriodEnd      time.Time `json:"period_end"`

	// Current usage (3 dimensions)
	Spans  int64 `json:"spans"`
	Bytes  int64 `json:"bytes"`
	Scores int64 `json:"scores"`

	// Free tier remaining
	FreeSpansRemaining  int64 `json:"free_spans_remaining"`
	FreeBytesRemaining  int64 `json:"free_bytes_remaining"`
	FreeScoresRemaining int64 `json:"free_scores_remaining"`

	// Free tier totals (for progress display)
	FreeSpansTotal  int64   `json:"free_spans_total"`
	FreeBytesTotal  int64   `json:"free_bytes_total"`
	FreeScoresTotal int64   `json:"free_scores_total"`

	// Calculated cost
	EstimatedCost float64 `json:"estimated_cost"`
}

// BillableUsageService handles billable usage queries and cost calculation
type BillableUsageService interface {
	// Get current period overview (for dashboard cards)
	GetUsageOverview(ctx context.Context, orgID ulid.ULID) (*UsageOverview, error)

	// Get usage time series (for charts)
	GetUsageTimeSeries(ctx context.Context, orgID ulid.ULID, start, end time.Time, granularity string) ([]*BillableUsage, error)

	// Get usage breakdown by project
	GetUsageByProject(ctx context.Context, orgID ulid.ULID, start, end time.Time) ([]*BillableUsageSummary, error)

	// Calculate cost for usage
	CalculateCost(ctx context.Context, usage *BillableUsageSummary, config *PricingConfig) float64
}

// BudgetService handles budget CRUD and monitoring
type BudgetService interface {
	// CRUD
	CreateBudget(ctx context.Context, budget *UsageBudget) error
	GetBudget(ctx context.Context, id ulid.ULID) (*UsageBudget, error)
	GetBudgetsByOrg(ctx context.Context, orgID ulid.ULID) ([]*UsageBudget, error)
	UpdateBudget(ctx context.Context, budget *UsageBudget) error
	DeleteBudget(ctx context.Context, id ulid.ULID) error

	// Monitoring
	CheckBudgets(ctx context.Context, orgID ulid.ULID) ([]*UsageAlert, error)
	GetAlerts(ctx context.Context, orgID ulid.ULID, limit int) ([]*UsageAlert, error)
	AcknowledgeAlert(ctx context.Context, orgID, alertID ulid.ULID) error
}
