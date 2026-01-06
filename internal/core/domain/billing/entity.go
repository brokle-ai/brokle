package billing

import (
	"time"

	"github.com/lib/pq"

	"brokle/pkg/ulid"
)

// Usage & Billing Entities

// Note: provider_id and model_id are now stored as text (no foreign keys to gateway tables)
// These values come from ClickHouse spans for cost calculation
type UsageRecord struct {
	CreatedAt      time.Time  `json:"created_at"`
	ProcessedAt    *time.Time `json:"processed_at,omitempty"`
	RequestType    string     `json:"request_type"`
	BillingTier    string     `json:"billing_tier"`
	Currency       string     `json:"currency"`
	ProviderName   string     `json:"provider_name,omitempty"` // Human-readable provider name (e.g., "openai", "anthropic")
	ModelName      string     `json:"model_name,omitempty"`    // Human-readable model name (e.g., "gpt-4", "claude-3-opus")
	Cost           float64    `json:"cost"`
	NetCost        float64    `json:"net_cost"`
	Discounts      float64    `json:"discounts"`
	TotalTokens    int32      `json:"total_tokens"`
	OutputTokens   int32      `json:"output_tokens"`
	InputTokens    int32      `json:"input_tokens"`
	ID             ulid.ULID  `json:"id"`
	ModelID        ulid.ULID  `json:"model_id"`     // Model ID from models table (for pricing lookup)
	ProviderID     ulid.ULID  `json:"provider_id"`  // Provider identifier (text, not FK)
	RequestID      ulid.ULID  `json:"request_id"`
	OrganizationID ulid.ULID  `json:"organization_id"`
}

// UsageQuota represents organization usage quotas and limits
type UsageQuota struct {
	ResetDate           time.Time `json:"reset_date"`
	LastUpdated         time.Time `json:"last_updated"`
	BillingTier         string    `json:"billing_tier"`
	Currency            string    `json:"currency"`
	MonthlyRequestLimit int64     `json:"monthly_request_limit"`
	MonthlyTokenLimit   int64     `json:"monthly_token_limit"`
	MonthlyCostLimit    float64   `json:"monthly_cost_limit"`
	CurrentRequests     int64     `json:"current_requests"`
	CurrentTokens       int64     `json:"current_tokens"`
	CurrentCost         float64   `json:"current_cost"`
	OrganizationID      ulid.ULID `json:"organization_id"`
}

type PaymentMethod struct {
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Type           string    `json:"type"`
	Provider       string    `json:"provider"`
	ExternalID     string    `json:"external_id"`
	Last4          string    `json:"last_4"`
	ExpiryMonth    int       `json:"expiry_month"`
	ExpiryYear     int       `json:"expiry_year"`
	ID             ulid.ULID `json:"id"`
	OrganizationID ulid.ULID `json:"organization_id"`
	IsDefault      bool      `json:"is_default"`
}

type Invoice struct {
	DueDate          time.Time              `json:"due_date"`
	UpdatedAt        time.Time              `json:"updated_at"`
	CreatedAt        time.Time              `json:"created_at"`
	PeriodStart      time.Time              `json:"period_start"`
	PeriodEnd        time.Time              `json:"period_end"`
	IssueDate        time.Time              `json:"issue_date"`
	PaidAt           *time.Time             `json:"paid_at,omitempty"`
	BillingAddress   *BillingAddress        `json:"billing_address"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	Currency         string                 `json:"currency"`
	Period           string                 `json:"period"`
	InvoiceNumber    string                 `json:"invoice_number"`
	OrganizationName string                 `json:"organization_name"`
	Notes            string                 `json:"notes,omitempty"`
	PaymentTerms     string                 `json:"payment_terms"`
	Status           InvoiceStatus          `json:"status"`
	LineItems        []InvoiceLineItem      `json:"line_items"`
	TotalAmount      float64                `json:"total_amount"`
	DiscountAmount   float64                `json:"discount_amount"`
	TaxAmount        float64                `json:"tax_amount"`
	Subtotal         float64                `json:"subtotal"`
	ID               ulid.ULID              `json:"id"`
	OrganizationID   ulid.ULID              `json:"organization_id"`
}

type InvoiceStatus string

const (
	InvoiceStatusDraft     InvoiceStatus = "draft"
	InvoiceStatusSent      InvoiceStatus = "sent"
	InvoiceStatusPaid      InvoiceStatus = "paid"
	InvoiceStatusOverdue   InvoiceStatus = "overdue"
	InvoiceStatusCancelled InvoiceStatus = "cancelled"
	InvoiceStatusRefunded  InvoiceStatus = "refunded"
)

type InvoiceLineItem struct {
	ProviderID   *ulid.ULID `json:"provider_id,omitempty"`
	ModelID      *ulid.ULID `json:"model_id,omitempty"`
	Description  string     `json:"description"`
	ProviderName string     `json:"provider_name,omitempty"`
	ModelName    string     `json:"model_name,omitempty"`
	RequestType  string     `json:"request_type,omitempty"`
	Quantity     float64    `json:"quantity"`
	UnitPrice    float64    `json:"unit_price"`
	Amount       float64    `json:"amount"`
	Tokens       int64      `json:"tokens,omitempty"`
	Requests     int64      `json:"requests,omitempty"`
	ID           ulid.ULID  `json:"id"`
}

type BillingAddress struct {
	Company    string `json:"company"`
	Address1   string `json:"address_1"`
	Address2   string `json:"address_2,omitempty"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
	TaxID      string `json:"tax_id,omitempty"`
}

type TaxConfiguration struct {
	TaxName     string  `json:"tax_name"`
	TaxID       string  `json:"tax_id"`
	TaxRate     float64 `json:"tax_rate"`
	IsInclusive bool    `json:"is_inclusive"`
}

type DiscountRule struct {
	UpdatedAt       time.Time          `json:"updated_at"`
	CreatedAt       time.Time          `json:"created_at"`
	ValidFrom       time.Time          `json:"valid_from"`
	Conditions      *DiscountCondition `json:"conditions,omitempty"`
	OrganizationID  *ulid.ULID         `json:"organization_id,omitempty"`
	UsageLimit      *int               `json:"usage_limit,omitempty"`
	ValidUntil      *time.Time         `json:"valid_until,omitempty"`
	Type            DiscountType       `json:"type"`
	Description     string             `json:"description"`
	Name            string             `json:"name"`
	MaximumDiscount float64            `json:"maximum_discount"`
	MinimumAmount   float64            `json:"minimum_amount"`
	Value           float64            `json:"value"`
	UsageCount      int                `json:"usage_count"`
	Priority        int                `json:"priority"`
	ID              ulid.ULID          `json:"id"`
	IsActive        bool               `json:"is_active"`
}

type DiscountType string

const (
	DiscountTypePercentage DiscountType = "percentage"
	DiscountTypeFixed      DiscountType = "fixed"
	DiscountTypeTiered     DiscountType = "tiered"
)

type DiscountCondition struct {
	MinUsage          *UsageThreshold `json:"min_usage,omitempty"`
	TimeOfDay         *TimeRange      `json:"time_of_day,omitempty"`
	VolumeThreshold   *VolumeDiscount `json:"volume_threshold,omitempty"`
	BillingTiers      []string        `json:"billing_tiers,omitempty"`
	RequestTypes      []string        `json:"request_types,omitempty"`
	Providers         []ulid.ULID     `json:"providers,omitempty"`
	Models            []ulid.ULID     `json:"models,omitempty"`
	DaysOfWeek        []time.Weekday  `json:"days_of_week,omitempty"`
	FirstTimeCustomer bool            `json:"first_time_customer"`
}

type UsageThreshold struct {
	Requests int64   `json:"requests"`
	Tokens   int64   `json:"tokens"`
	Cost     float64 `json:"cost"`
}

type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type VolumeDiscount struct {
	Tiers []VolumeTier `json:"tiers"`
}

type VolumeTier struct {
	MinAmount float64 `json:"min_amount"`
	Discount  float64 `json:"discount"` // percentage or fixed amount
}

// BillingRecord represents a billing record (moved from deleted analytics worker)
type BillingRecord struct {
	UpdatedAt      time.Time              `json:"updated_at" db:"updated_at"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
	Metadata       map[string]interface{} `json:"metadata" db:"metadata"`
	TransactionID  *string                `json:"transaction_id,omitempty" db:"transaction_id"`
	PaymentMethod  *string                `json:"payment_method,omitempty" db:"payment_method"`
	ProcessedAt    *time.Time             `json:"processed_at,omitempty" db:"processed_at"`
	Period         string                 `json:"period" db:"period"`
	Currency       string                 `json:"currency" db:"currency"`
	Status         string                 `json:"status" db:"status"`
	Amount         float64                `json:"amount" db:"amount"`
	NetCost        float64                `json:"net_cost" db:"net_cost"`
	ID             ulid.ULID              `json:"id" db:"id"`
	OrganizationID ulid.ULID              `json:"organization_id" db:"organization_id"`
}

// BillingSummary represents aggregated billing data (moved from deleted analytics worker)
type BillingSummary struct {
	PeriodStart       time.Time              `json:"period_start" db:"period_start"`
	PeriodEnd         time.Time              `json:"period_end" db:"period_end"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	GeneratedAt       time.Time              `json:"generated_at" db:"generated_at"`
	ModelBreakdown    map[string]interface{} `json:"model_breakdown"`
	ProviderBreakdown map[string]interface{} `json:"provider_breakdown"`
	Currency          string                 `json:"currency" db:"currency"`
	Period            string                 `json:"period" db:"period"`
	Status            string                 `json:"status" db:"status"`
	TotalAmount       float64                `json:"total_amount" db:"total_amount"`
	Discounts         float64                `json:"discounts" db:"discounts"`
	NetCost           float64                `json:"net_cost" db:"net_cost"`
	RecordCount       int                    `json:"record_count" db:"record_count"`
	TotalCost         float64                `json:"total_cost" db:"total_cost"`
	TotalTokens       int                    `json:"total_tokens" db:"total_tokens"`
	TotalRequests     int                    `json:"total_requests" db:"total_requests"`
	ID                ulid.ULID              `json:"id" db:"id"`
	OrganizationID    ulid.ULID              `json:"organization_id" db:"organization_id"`
}

// CostMetric represents cost tracking data (moved from deleted analytics worker)
type CostMetric struct {
	Timestamp      time.Time `json:"timestamp"`
	Provider       string    `json:"provider"`
	Currency       string    `json:"currency"`
	RequestType    string    `json:"request_type"`
	Model          string    `json:"model"`
	TotalCost      float64   `json:"total_cost"`
	OutputTokens   int32     `json:"output_tokens"`
	InputTokens    int32     `json:"input_tokens"`
	TotalTokens    int32     `json:"total_tokens"`
	ModelID        ulid.ULID `json:"model_id"`
	RequestID      ulid.ULID `json:"request_id"`
	ProviderID     ulid.ULID `json:"provider_id"`
	ProjectID      ulid.ULID `json:"project_id"`
	OrganizationID ulid.ULID `json:"organization_id"`
}

// Usage-Based Billing Entities

// Queried from ClickHouse billable_usage_hourly/daily tables
type BillableUsage struct {
	OrganizationID ulid.ULID `json:"organization_id"`
	ProjectID      ulid.ULID `json:"project_id"`
	BucketTime     time.Time `json:"bucket_time"`

	// Three billable dimensions
	SpanCount      int64 `json:"span_count"`       // All spans (traces + child spans)
	BytesProcessed int64 `json:"bytes_processed"`  // Total payload bytes (input + output)
	ScoreCount     int64 `json:"score_count"`      // Quality scores

	// Informational (not billable by Brokle)
	AIProviderCost float64 `json:"ai_provider_cost"`

	LastUpdated time.Time `json:"last_updated"`
}

type BillableUsageSummary struct {
	OrganizationID ulid.ULID  `json:"organization_id"`
	ProjectID      *ulid.ULID `json:"project_id,omitempty"` // nil for org-level summary
	PeriodStart    time.Time  `json:"period_start"`
	PeriodEnd      time.Time  `json:"period_end"`

	// Totals for period
	TotalSpans  int64 `json:"total_spans"`
	TotalBytes  int64 `json:"total_bytes"`
	TotalScores int64 `json:"total_scores"`

	// Calculated cost
	TotalCost float64 `json:"total_cost"`

	// Informational
	TotalAIProviderCost float64 `json:"total_ai_provider_cost"`
}

type PricingConfig struct {
	ID        ulid.ULID `json:"id" gorm:"column:id;primaryKey"`
	Name      string    `json:"name" gorm:"column:name"` // free, pro, enterprise
	IsActive  bool      `json:"is_active" gorm:"column:is_active"`
	IsDefault bool      `json:"is_default" gorm:"column:is_default"` // Default plan for new organizations
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`

	// Span pricing (per 100K)
	FreeSpans         int64    `json:"free_spans" gorm:"column:free_spans"`
	PricePer100KSpans *float64 `json:"price_per_100k_spans,omitempty" gorm:"column:price_per_100k_spans"` // nil = unlimited in free tier

	// Data volume pricing (per GB)
	FreeGB     float64  `json:"free_gb" gorm:"column:free_gb"`
	PricePerGB *float64 `json:"price_per_gb,omitempty" gorm:"column:price_per_gb"` // nil = unlimited in free tier

	// Score pricing (per 1K)
	FreeScores       int64    `json:"free_scores" gorm:"column:free_scores"`
	PricePer1KScores *float64 `json:"price_per_1k_scores,omitempty" gorm:"column:price_per_1k_scores"` // nil = unlimited in free tier
}

type OrganizationBilling struct {
	OrganizationID        ulid.ULID `json:"organization_id" db:"organization_id" gorm:"type:char(26);primaryKey"`
	PricingConfigID       ulid.ULID `json:"pricing_config_id" db:"pricing_config_id"`
	BillingCycleStart     time.Time `json:"billing_cycle_start" db:"billing_cycle_start"`
	BillingCycleAnchorDay int       `json:"billing_cycle_anchor_day" db:"billing_cycle_anchor_day"` // Day of month (1-28)

	// Current period usage (three dimensions)
	CurrentPeriodSpans  int64 `json:"current_period_spans" db:"current_period_spans"`
	CurrentPeriodBytes  int64 `json:"current_period_bytes" db:"current_period_bytes"`
	CurrentPeriodScores int64 `json:"current_period_scores" db:"current_period_scores"`

	// Calculated cost this period
	CurrentPeriodCost float64 `json:"current_period_cost" db:"current_period_cost"`

	// Free tier remaining (three dimensions)
	FreeSpansRemaining  int64 `json:"free_spans_remaining" db:"free_spans_remaining"`
	FreeBytesRemaining  int64 `json:"free_bytes_remaining" db:"free_bytes_remaining"`
	FreeScoresRemaining int64 `json:"free_scores_remaining" db:"free_scores_remaining"`

	LastSyncedAt time.Time `json:"last_synced_at" db:"last_synced_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type BudgetType string

const (
	BudgetTypeMonthly BudgetType = "monthly"
	BudgetTypeWeekly  BudgetType = "weekly"
)

type UsageBudget struct {
	ID             ulid.ULID  `json:"id" db:"id"`
	OrganizationID ulid.ULID  `json:"organization_id" db:"organization_id"`
	ProjectID      *ulid.ULID `json:"project_id,omitempty" db:"project_id"` // nil for org-level budget
	Name           string     `json:"name" db:"name"`
	BudgetType     BudgetType `json:"budget_type" db:"budget_type"`

	// Limits (any can be set, nil = no limit)
	SpanLimit  *int64   `json:"span_limit,omitempty" db:"span_limit"`
	BytesLimit *int64   `json:"bytes_limit,omitempty" db:"bytes_limit"`
	ScoreLimit *int64   `json:"score_limit,omitempty" db:"score_limit"`
	CostLimit  *float64 `json:"cost_limit,omitempty" db:"cost_limit"`

	// Current usage
	CurrentSpans  int64   `json:"current_spans" db:"current_spans"`
	CurrentBytes  int64   `json:"current_bytes" db:"current_bytes"`
	CurrentScores int64   `json:"current_scores" db:"current_scores"`
	CurrentCost   float64 `json:"current_cost" db:"current_cost"`

	// Alert thresholds (flexible array of percentages, e.g., [50, 80, 100])
	AlertThresholds pq.Int64Array `json:"alert_thresholds" gorm:"column:alert_thresholds;type:integer[];default:'{50,80,100}'"`

	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type AlertDimension string

const (
	AlertDimensionSpans  AlertDimension = "spans"
	AlertDimensionBytes  AlertDimension = "bytes"
	AlertDimensionScores AlertDimension = "scores"
	AlertDimensionCost   AlertDimension = "cost"
)

type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityCritical AlertSeverity = "critical"
)

type AlertStatus string

const (
	AlertStatusTriggered    AlertStatus = "triggered"
	AlertStatusAcknowledged AlertStatus = "acknowledged"
	AlertStatusResolved     AlertStatus = "resolved"
)

type UsageAlert struct {
	ID             ulid.ULID  `json:"id" db:"id"`
	BudgetID       *ulid.ULID `json:"budget_id,omitempty" db:"budget_id"`
	OrganizationID ulid.ULID  `json:"organization_id" db:"organization_id"`
	ProjectID      *ulid.ULID `json:"project_id,omitempty" db:"project_id"`

	AlertThreshold int64          `json:"alert_threshold" db:"alert_threshold"`
	Dimension      AlertDimension `json:"dimension" db:"dimension"`
	Severity       AlertSeverity  `json:"severity" db:"severity"`
	ThresholdValue int64          `json:"threshold_value" db:"threshold_value"`
	ActualValue    int64          `json:"actual_value" db:"actual_value"`
	PercentUsed    float64        `json:"percent_used" db:"percent_used"`

	Status           AlertStatus `json:"status" db:"status"`
	TriggeredAt      time.Time   `json:"triggered_at" db:"triggered_at"`
	AcknowledgedAt   *time.Time  `json:"acknowledged_at,omitempty" db:"acknowledged_at"`
	ResolvedAt       *time.Time  `json:"resolved_at,omitempty" db:"resolved_at"`
	NotificationSent bool        `json:"notification_sent" db:"notification_sent"`
}
