package billing

import (
	"time"

	"github.com/shopspring/decimal"

	"github.com/google/uuid"
)

// ============================================================================
// Usage-Based Billing Services (Spans + GB + Scores)
// ============================================================================

// UsageOverview represents the current usage overview for display
type UsageOverview struct {
	OrganizationID uuid.UUID `json:"organization_id"`
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
	FreeSpansTotal  int64 `json:"free_spans_total"`
	FreeBytesTotal  int64 `json:"free_bytes_total"`
	FreeScoresTotal int64 `json:"free_scores_total"`

	// Calculated cost
	EstimatedCost decimal.Decimal `json:"estimated_cost"`
}
