package billing

import (
	"context"
	"time"

	"brokle/pkg/ulid"
)

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
