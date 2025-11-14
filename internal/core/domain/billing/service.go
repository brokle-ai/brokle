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
	OrganizationID       ulid.ULID `json:"organization_id"`
	RequestsOK           bool      `json:"requests_ok"`
	TokensOK             bool      `json:"tokens_ok"`
	CostOK               bool      `json:"cost_ok"`
	Status               string    `json:"status"`
	RequestsUsagePercent float64   `json:"requests_usage_percent"`
	TokensUsagePercent   float64   `json:"tokens_usage_percent"`
	CostUsagePercent     float64   `json:"cost_usage_percent"`
}
