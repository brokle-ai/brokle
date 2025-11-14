package billing

import (
	"time"

	"brokle/pkg/ulid"
)

// UsageRecord represents a usage tracking record for billing
type UsageRecord struct {
	ID             ulid.ULID  `json:"id"`
	OrganizationID ulid.ULID  `json:"organization_id"`
	RequestID      ulid.ULID  `json:"request_id"`
	ProviderID     ulid.ULID  `json:"provider_id"`
	ModelID        ulid.ULID  `json:"model_id"`
	RequestType    string     `json:"request_type"`
	InputTokens    int32      `json:"input_tokens"`
	OutputTokens   int32      `json:"output_tokens"`
	TotalTokens    int32      `json:"total_tokens"`
	Cost           float64    `json:"cost"`
	Currency       string     `json:"currency"`
	BillingTier    string     `json:"billing_tier"`
	Discounts      float64    `json:"discounts"`
	NetCost        float64    `json:"net_cost"`
	CreatedAt      time.Time  `json:"created_at"`
	ProcessedAt    *time.Time `json:"processed_at,omitempty"`
}

// UsageQuota represents organization usage quotas and limits
type UsageQuota struct {
	OrganizationID      ulid.ULID `json:"organization_id"`
	BillingTier         string    `json:"billing_tier"`
	MonthlyRequestLimit int64     `json:"monthly_request_limit"`
	MonthlyTokenLimit   int64     `json:"monthly_token_limit"`
	MonthlyCostLimit    float64   `json:"monthly_cost_limit"`
	CurrentRequests     int64     `json:"current_requests"`
	CurrentTokens       int64     `json:"current_tokens"`
	CurrentCost         float64   `json:"current_cost"`
	Currency            string    `json:"currency"`
	ResetDate           time.Time `json:"reset_date"`
	LastUpdated         time.Time `json:"last_updated"`
}

// PaymentMethod represents organization payment information
type PaymentMethod struct {
	ID             ulid.ULID `json:"id"`
	OrganizationID ulid.ULID `json:"organization_id"`
	Type           string    `json:"type"`     // card, bank_transfer, etc.
	Provider       string    `json:"provider"` // stripe, etc.
	ExternalID     string    `json:"external_id"`
	Last4          string    `json:"last_4"`
	ExpiryMonth    int       `json:"expiry_month"`
	ExpiryYear     int       `json:"expiry_year"`
	IsDefault      bool      `json:"is_default"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Invoice represents a generated invoice
type Invoice struct {
	ID               ulid.ULID              `json:"id"`
	InvoiceNumber    string                 `json:"invoice_number"`
	OrganizationID   ulid.ULID              `json:"organization_id"`
	OrganizationName string                 `json:"organization_name"`
	BillingAddress   *BillingAddress        `json:"billing_address"`
	Period           string                 `json:"period"`
	PeriodStart      time.Time              `json:"period_start"`
	PeriodEnd        time.Time              `json:"period_end"`
	IssueDate        time.Time              `json:"issue_date"`
	DueDate          time.Time              `json:"due_date"`
	LineItems        []InvoiceLineItem      `json:"line_items"`
	Subtotal         float64                `json:"subtotal"`
	TaxAmount        float64                `json:"tax_amount"`
	DiscountAmount   float64                `json:"discount_amount"`
	TotalAmount      float64                `json:"total_amount"`
	Currency         string                 `json:"currency"`
	Status           InvoiceStatus          `json:"status"`
	PaymentTerms     string                 `json:"payment_terms"`
	Notes            string                 `json:"notes,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
	PaidAt           *time.Time             `json:"paid_at,omitempty"`
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
	ID           ulid.ULID  `json:"id"`
	Description  string     `json:"description"`
	Quantity     float64    `json:"quantity"`
	UnitPrice    float64    `json:"unit_price"`
	Amount       float64    `json:"amount"`
	ProviderID   *ulid.ULID `json:"provider_id,omitempty"`
	ProviderName string     `json:"provider_name,omitempty"`
	ModelID      *ulid.ULID `json:"model_id,omitempty"`
	ModelName    string     `json:"model_name,omitempty"`
	RequestType  string     `json:"request_type,omitempty"`
	Tokens       int64      `json:"tokens,omitempty"`
	Requests     int64      `json:"requests,omitempty"`
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
	TaxRate     float64 `json:"tax_rate"`     // e.g., 0.08 for 8%
	TaxName     string  `json:"tax_name"`     // e.g., "VAT", "GST", "Sales Tax"
	TaxID       string  `json:"tax_id"`       // Tax identification number
	IsInclusive bool    `json:"is_inclusive"` // Whether tax is included in prices
}

// DiscountRule represents a discount rule
type DiscountRule struct {
	ID              ulid.ULID          `json:"id"`
	OrganizationID  *ulid.ULID         `json:"organization_id,omitempty"` // nil for global rules
	Name            string             `json:"name"`
	Description     string             `json:"description"`
	Type            DiscountType       `json:"type"`
	Value           float64            `json:"value"` // percentage (0.1 = 10%) or fixed amount
	MinimumAmount   float64            `json:"minimum_amount"`
	MaximumDiscount float64            `json:"maximum_discount"`
	Conditions      *DiscountCondition `json:"conditions,omitempty"`
	ValidFrom       time.Time          `json:"valid_from"`
	ValidUntil      *time.Time         `json:"valid_until,omitempty"`
	UsageLimit      *int               `json:"usage_limit,omitempty"`
	UsageCount      int                `json:"usage_count"`
	IsActive        bool               `json:"is_active"`
	Priority        int                `json:"priority"` // Higher priority rules are applied first
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
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
	BillingTiers      []string        `json:"billing_tiers,omitempty"`    // Apply only to specific tiers
	MinUsage          *UsageThreshold `json:"min_usage,omitempty"`        // Minimum usage requirements
	RequestTypes      []string        `json:"request_types,omitempty"`    // Specific request types
	Providers         []ulid.ULID     `json:"providers,omitempty"`        // Specific providers
	Models            []ulid.ULID     `json:"models,omitempty"`           // Specific models
	TimeOfDay         *TimeRange      `json:"time_of_day,omitempty"`      // Time-based discounts
	DaysOfWeek        []time.Weekday  `json:"days_of_week,omitempty"`     // Day-based discounts
	FirstTimeCustomer bool            `json:"first_time_customer"`        // First-time customer discount
	VolumeThreshold   *VolumeDiscount `json:"volume_threshold,omitempty"` // Volume-based discounts
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
	ID             ulid.ULID              `json:"id" db:"id"`
	OrganizationID ulid.ULID              `json:"organization_id" db:"organization_id"`
	Period         string                 `json:"period" db:"period"` // e.g., "2024-11"
	Amount         float64                `json:"amount" db:"amount"`
	Currency       string                 `json:"currency" db:"currency"`
	Status         string                 `json:"status" db:"status"`
	TransactionID  *string                `json:"transaction_id,omitempty" db:"transaction_id"`
	PaymentMethod  *string                `json:"payment_method,omitempty" db:"payment_method"`
	ProcessedAt    *time.Time             `json:"processed_at,omitempty" db:"processed_at"`
	NetCost        float64                `json:"net_cost" db:"net_cost"`
	Metadata       map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at" db:"updated_at"`
}

// BillingSummary represents aggregated billing data (moved from deleted analytics worker)
type BillingSummary struct {
	ID                ulid.ULID              `json:"id" db:"id"`
	OrganizationID    ulid.ULID              `json:"organization_id" db:"organization_id"`
	Period            string                 `json:"period" db:"period"`
	PeriodStart       time.Time              `json:"period_start" db:"period_start"`
	PeriodEnd         time.Time              `json:"period_end" db:"period_end"`
	TotalRequests     int                    `json:"total_requests" db:"total_requests"`
	TotalTokens       int                    `json:"total_tokens" db:"total_tokens"`
	TotalCost         float64                `json:"total_cost" db:"total_cost"`
	TotalAmount       float64                `json:"total_amount" db:"total_amount"`
	Currency          string                 `json:"currency" db:"currency"`
	Discounts         float64                `json:"discounts" db:"discounts"`
	NetCost           float64                `json:"net_cost" db:"net_cost"`
	RecordCount       int                    `json:"record_count" db:"record_count"`
	Status            string                 `json:"status" db:"status"`
	GeneratedAt       time.Time              `json:"generated_at" db:"generated_at"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	ProviderBreakdown map[string]interface{} `json:"provider_breakdown"`
	ModelBreakdown    map[string]interface{} `json:"model_breakdown"`
}

// CostMetric represents cost tracking data (moved from deleted analytics worker)
type CostMetric struct {
	RequestID      ulid.ULID `json:"request_id"`
	OrganizationID ulid.ULID `json:"organization_id"`
	ProjectID      ulid.ULID `json:"project_id"`
	ProviderID     ulid.ULID `json:"provider_id"`
	ModelID        ulid.ULID `json:"model_id"`
	Model          string    `json:"model"`
	Provider       string    `json:"provider"`
	RequestType    string    `json:"request_type"`
	InputTokens    int32     `json:"input_tokens"`
	OutputTokens   int32     `json:"output_tokens"`
	TotalTokens    int32     `json:"total_tokens"`
	TotalCost      float64   `json:"total_cost"`
	Currency       string    `json:"currency"`
	Timestamp      time.Time `json:"timestamp"`
}
