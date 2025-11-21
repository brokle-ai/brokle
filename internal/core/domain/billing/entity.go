package billing

import (
	"time"

	"brokle/pkg/ulid"
)

// UsageRecord represents a usage tracking record for billing
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

// PaymentMethod represents organization payment information
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

// Invoice represents a generated invoice
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

// InvoiceStatus represents the status of an invoice
type InvoiceStatus string

const (
	InvoiceStatusDraft     InvoiceStatus = "draft"
	InvoiceStatusSent      InvoiceStatus = "sent"
	InvoiceStatusPaid      InvoiceStatus = "paid"
	InvoiceStatusOverdue   InvoiceStatus = "overdue"
	InvoiceStatusCancelled InvoiceStatus = "cancelled"
	InvoiceStatusRefunded  InvoiceStatus = "refunded"
)

// InvoiceLineItem represents a line item on an invoice
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

// BillingAddress represents an organization's billing address
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

// TaxConfiguration represents tax configuration for billing
type TaxConfiguration struct {
	TaxName     string  `json:"tax_name"`
	TaxID       string  `json:"tax_id"`
	TaxRate     float64 `json:"tax_rate"`
	IsInclusive bool    `json:"is_inclusive"`
}

// DiscountRule represents a discount rule
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

// DiscountType represents the type of discount
type DiscountType string

const (
	DiscountTypePercentage DiscountType = "percentage"
	DiscountTypeFixed      DiscountType = "fixed"
	DiscountTypeTiered     DiscountType = "tiered"
)

// DiscountCondition represents conditions for applying discounts
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

// UsageThreshold represents minimum usage requirements
type UsageThreshold struct {
	Requests int64   `json:"requests"`
	Tokens   int64   `json:"tokens"`
	Cost     float64 `json:"cost"`
}

// TimeRange represents a time range for discounts
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// VolumeDiscount represents volume-based discount tiers
type VolumeDiscount struct {
	Tiers []VolumeTier `json:"tiers"`
}

// VolumeTier represents a single volume discount tier
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
