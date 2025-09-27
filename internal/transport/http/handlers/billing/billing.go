package billing

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"brokle/internal/config"
	"brokle/pkg/response"
)

type Handler struct {
	config *config.Config
	logger *logrus.Logger
}

func NewHandler(config *config.Config, logger *logrus.Logger) *Handler {
	return &Handler{config: config, logger: logger}
}

// Request/Response Models

// UsageMetrics represents usage metrics for an organization
type UsageMetrics struct {
	OrganizationID   string                    `json:"organization_id" example:"org_1234567890" description:"Organization identifier"`
	BillingPeriod    BillingPeriod            `json:"billing_period" description:"Current billing period"`
	TotalRequests    int64                    `json:"total_requests" example:"125000" description:"Total AI requests in period"`
	TotalTokens      int64                    `json:"total_tokens" example:"2500000" description:"Total tokens processed"`
	TotalCost        float64                  `json:"total_cost" example:"1250.75" description:"Total cost in USD"`
	UsageByProvider  []ProviderUsage          `json:"usage_by_provider" description:"Usage breakdown by AI provider"`
	UsageByProject   []ProjectUsage           `json:"usage_by_project" description:"Usage breakdown by project"`
	UsageByEnvironment []EnvironmentUsage      `json:"usage_by_environment" description:"Usage breakdown by environment"`
	DailyUsage       []DailyUsage             `json:"daily_usage" description:"Daily usage breakdown"`
	QuotaLimits      QuotaLimits              `json:"quota_limits" description:"Current quota limits and usage"`
	BillingAlerts    []BillingAlert           `json:"billing_alerts,omitempty" description:"Active billing alerts"`
}

// BillingPeriod represents a billing period
type BillingPeriod struct {
	StartDate time.Time `json:"start_date" example:"2024-01-01T00:00:00Z" description:"Billing period start date"`
	EndDate   time.Time `json:"end_date" example:"2024-01-31T23:59:59Z" description:"Billing period end date"`
	Status    string    `json:"status" example:"current" description:"Period status (current, closed, future)"`
}

// ProviderUsage represents usage metrics for a specific provider
type ProviderUsage struct {
	Provider string  `json:"provider" example:"openai" description:"AI provider name"`
	Requests int64   `json:"requests" example:"75000" description:"Number of requests"`
	Tokens   int64   `json:"tokens" example:"1500000" description:"Total tokens processed"`
	Cost     float64 `json:"cost" example:"750.50" description:"Total cost in USD"`
	Percent  float64 `json:"percent" example:"0.60" description:"Percentage of total usage"`
}

// ProjectUsage represents usage metrics for a specific project
type ProjectUsage struct {
	ProjectID   string  `json:"project_id" example:"proj_1234567890" description:"Project identifier"`
	ProjectName string  `json:"project_name" example:"AI Chatbot" description:"Project name"`
	Requests    int64   `json:"requests" example:"45000" description:"Number of requests"`
	Tokens      int64   `json:"tokens" example:"900000" description:"Total tokens processed"`
	Cost        float64 `json:"cost" example:"450.25" description:"Total cost in USD"`
	Percent     float64 `json:"percent" example:"0.36" description:"Percentage of total usage"`
}

// EnvironmentUsage represents usage metrics for a specific environment
type EnvironmentUsage struct {
	Environment string  `json:"environment" example:"production" description:"Environment tag"`
	Requests    int64   `json:"requests" example:"100000" description:"Number of requests"`
	Tokens      int64   `json:"tokens" example:"2000000" description:"Total tokens processed"`
	Cost        float64 `json:"cost" example:"1000.50" description:"Total cost in USD"`
	Percent     float64 `json:"percent" example:"0.80" description:"Percentage of total usage"`
}

// DailyUsage represents usage metrics for a specific day
type DailyUsage struct {
	Date     time.Time `json:"date" example:"2024-01-15T00:00:00Z" description:"Date for usage metrics"`
	Requests int64     `json:"requests" example:"4200" description:"Number of requests"`
	Tokens   int64     `json:"tokens" example:"84000" description:"Total tokens processed"`
	Cost     float64   `json:"cost" example:"42.15" description:"Total cost in USD"`
}

// QuotaLimits represents current quota limits and usage
type QuotaLimits struct {
	RequestsLimit     int64   `json:"requests_limit" example:"100000" description:"Monthly requests limit (0 = unlimited)"`
	RequestsUsed      int64   `json:"requests_used" example:"75000" description:"Requests used this month"`
	RequestsPercent   float64 `json:"requests_percent" example:"0.75" description:"Percentage of requests quota used"`
	CostLimit         float64 `json:"cost_limit" example:"1000.00" description:"Monthly cost limit in USD (0 = unlimited)"`
	CostUsed          float64 `json:"cost_used" example:"750.25" description:"Cost used this month in USD"`
	CostPercent       float64 `json:"cost_percent" example:"0.75" description:"Percentage of cost quota used"`
	TokensLimit       int64   `json:"tokens_limit" example:"2000000" description:"Monthly tokens limit (0 = unlimited)"`
	TokensUsed        int64   `json:"tokens_used" example:"1500000" description:"Tokens used this month"`
	TokensPercent     float64 `json:"tokens_percent" example:"0.75" description:"Percentage of tokens quota used"`
	OverageAllowed    bool    `json:"overage_allowed" example:"true" description:"Whether overage is allowed beyond limits"`
}

// BillingAlert represents an active billing alert
type BillingAlert struct {
	ID          string    `json:"id" example:"alert_1234567890" description:"Alert identifier"`
	Type        string    `json:"type" example:"cost_threshold" description:"Alert type (cost_threshold, quota_limit, usage_spike)"`
	Severity    string    `json:"severity" example:"warning" description:"Alert severity (info, warning, critical)"`
	Message     string    `json:"message" example:"Monthly cost has reached 80% of limit" description:"Alert message"`
	TriggeredAt time.Time `json:"triggered_at" example:"2024-01-15T10:30:00Z" description:"When alert was triggered"`
	Threshold   float64   `json:"threshold" example:"800.00" description:"Threshold value that triggered alert"`
	CurrentValue float64  `json:"current_value" example:"750.25" description:"Current value that triggered alert"`
	Status      string    `json:"status" example:"active" description:"Alert status (active, acknowledged, resolved)"`
}

// Invoice represents a billing invoice
type Invoice struct {
	ID              string           `json:"id" example:"inv_1234567890" description:"Invoice identifier"`
	InvoiceNumber   string           `json:"invoice_number" example:"INV-2024-001" description:"Human-readable invoice number"`
	OrganizationID  string           `json:"organization_id" example:"org_1234567890" description:"Organization identifier"`
	BillingPeriod   BillingPeriod    `json:"billing_period" description:"Billing period for this invoice"`
	Status          string           `json:"status" example:"paid" description:"Invoice status (draft, sent, paid, overdue, void)"`
	Subtotal        float64          `json:"subtotal" example:"1250.75" description:"Subtotal before taxes in USD"`
	TaxAmount       float64          `json:"tax_amount" example:"125.08" description:"Tax amount in USD"`
	Total           float64          `json:"total" example:"1375.83" description:"Total amount including taxes in USD"`
	Currency        string           `json:"currency" example:"USD" description:"Invoice currency"`
	IssueDate       time.Time        `json:"issue_date" example:"2024-02-01T00:00:00Z" description:"Invoice issue date"`
	DueDate         time.Time        `json:"due_date" example:"2024-02-15T23:59:59Z" description:"Payment due date"`
	PaidDate        time.Time        `json:"paid_date,omitempty" example:"2024-02-10T14:30:00Z" description:"Payment date (if paid)"`
	LineItems       []InvoiceLineItem `json:"line_items" description:"Invoice line items"`
	PaymentMethod   string           `json:"payment_method,omitempty" example:"credit_card" description:"Payment method used"`
	DownloadURL     string           `json:"download_url,omitempty" example:"https://invoices.brokle.ai/inv_1234567890.pdf" description:"PDF download URL"`
}

// InvoiceLineItem represents a line item on an invoice
type InvoiceLineItem struct {
	Description string  `json:"description" example:"OpenAI GPT-4 Usage" description:"Line item description"`
	Quantity    int64   `json:"quantity" example:"50000" description:"Quantity (e.g., number of requests or tokens)"`
	UnitPrice   float64 `json:"unit_price" example:"0.025" description:"Price per unit in USD"`
	Amount      float64 `json:"amount" example:"1250.00" description:"Line item total amount in USD"`
	Provider    string  `json:"provider,omitempty" example:"openai" description:"AI provider (if applicable)"`
	Model       string  `json:"model,omitempty" example:"gpt-4" description:"AI model (if applicable)"`
}

// Subscription represents a billing subscription
type Subscription struct {
	ID              string              `json:"id" example:"sub_1234567890" description:"Subscription identifier"`
	OrganizationID  string              `json:"organization_id" example:"org_1234567890" description:"Organization identifier"`
	Plan            SubscriptionPlan    `json:"plan" description:"Current subscription plan"`
	Status          string              `json:"status" example:"active" description:"Subscription status (active, canceled, past_due, unpaid)"`
	CurrentPeriod   BillingPeriod       `json:"current_period" description:"Current billing period"`
	NextBillingDate time.Time           `json:"next_billing_date" example:"2024-02-01T00:00:00Z" description:"Next billing date"`
	CreatedAt       time.Time           `json:"created_at" example:"2024-01-01T00:00:00Z" description:"Subscription creation date"`
	UpdatedAt       time.Time           `json:"updated_at" example:"2024-01-15T10:30:00Z" description:"Last update date"`
	PaymentMethod   PaymentMethod       `json:"payment_method" description:"Payment method for subscription"`
	AddOns          []SubscriptionAddOn `json:"add_ons,omitempty" description:"Subscription add-ons"`
	Discounts       []Discount          `json:"discounts,omitempty" description:"Applied discounts"`
	CancelAt        time.Time           `json:"cancel_at,omitempty" example:"2024-12-31T23:59:59Z" description:"Scheduled cancellation date"`
}

// SubscriptionPlan represents a subscription plan
type SubscriptionPlan struct {
	ID            string  `json:"id" example:"plan_pro" description:"Plan identifier"`
	Name          string  `json:"name" example:"Pro Plan" description:"Plan name"`
	Price         float64 `json:"price" example:"29.00" description:"Monthly price in USD"`
	Currency      string  `json:"currency" example:"USD" description:"Plan currency"`
	Interval      string  `json:"interval" example:"month" description:"Billing interval (month, year)"`
	RequestsLimit int64   `json:"requests_limit" example:"100000" description:"Monthly requests limit (0 = unlimited)"`
	FeaturesIncluded []string `json:"features_included" example:"[\"advanced_analytics\", \"semantic_caching\"]" description:"Features included in plan"`
}

// PaymentMethod represents a payment method
type PaymentMethod struct {
	ID     string `json:"id" example:"pm_1234567890" description:"Payment method identifier"`
	Type   string `json:"type" example:"credit_card" description:"Payment method type"`
	Brand  string `json:"brand,omitempty" example:"visa" description:"Card brand (for credit cards)"`
	Last4  string `json:"last4,omitempty" example:"1234" description:"Last 4 digits (for credit cards)"`
	Expiry string `json:"expiry,omitempty" example:"12/2025" description:"Expiry date (for credit cards)"`
	Default bool  `json:"default" example:"true" description:"Whether this is the default payment method"`
}

// SubscriptionAddOn represents a subscription add-on
type SubscriptionAddOn struct {
	ID       string  `json:"id" example:"addon_extra_requests" description:"Add-on identifier"`
	Name     string  `json:"name" example:"Extra Requests" description:"Add-on name"`
	Price    float64 `json:"price" example:"10.00" description:"Add-on price in USD"`
	Quantity int     `json:"quantity" example:"2" description:"Add-on quantity"`
}

// Discount represents a discount applied to subscription
type Discount struct {
	ID       string  `json:"id" example:"discount_1234567890" description:"Discount identifier"`
	Name     string  `json:"name" example:"New Customer 20% Off" description:"Discount name"`
	Type     string  `json:"type" example:"percentage" description:"Discount type (percentage, fixed_amount)"`
	Value    float64 `json:"value" example:"20.0" description:"Discount value (percentage or amount)"`
	ValidUntil time.Time `json:"valid_until,omitempty" example:"2024-06-01T00:00:00Z" description:"Discount expiration date"`
}

// UpdateSubscriptionRequest represents a request to update subscription
type UpdateSubscriptionRequest struct {
	PlanID string `json:"plan_id,omitempty" example:"plan_business" description:"New plan ID"`
	AddOns []SubscriptionAddOn `json:"add_ons,omitempty" description:"Updated add-ons"`
	PaymentMethodID string `json:"payment_method_id,omitempty" example:"pm_0987654321" description:"New payment method ID"`
	CancelAt time.Time `json:"cancel_at,omitempty" example:"2024-12-31T23:59:59Z" description:"Schedule cancellation date"`
}

// ListInvoicesResponse represents the response when listing invoices
// NOTE: This struct is not used. When implementing, use response.SuccessWithPagination()
// with []Invoice directly and response.NewPagination() for consistent pagination format.
type ListInvoicesResponse struct {
	Invoices []Invoice `json:"invoices" description:"List of invoices"`
	// Pagination fields removed - use response.SuccessWithPagination() instead
}

// GetUsage handles GET /billing/:orgId/usage
// @Summary Get organization usage metrics
// @Description Get detailed usage metrics and billing information for an organization
// @Tags Billing
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID" example("org_1234567890")
// @Param period query string false "Billing period" default("current") Enums(current,previous,custom)
// @Param start_date query string false "Custom period start date (RFC3339)" example("2024-01-01T00:00:00Z")
// @Param end_date query string false "Custom period end date (RFC3339)" example("2024-01-31T23:59:59Z")
// @Param breakdown query string false "Include usage breakdown" default("all") Enums(all,provider,project,environment,none)
// @Success 200 {object} response.SuccessResponse{data=UsageMetrics} "Organization usage metrics"
// @Failure 400 {object} response.ErrorResponse "Bad request - invalid organization ID or parameters"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - insufficient permissions to view billing information"
// @Failure 404 {object} response.ErrorResponse "Organization not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/billing/{orgId}/usage [get]
func (h *Handler) GetUsage(c *gin.Context) { response.Success(c, gin.H{"message": "Get usage - TODO"}) }
// ListInvoices handles GET /billing/:orgId/invoices
// @Summary List organization invoices
// @Description Get a paginated list of invoices for an organization
// @Tags Billing
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID" example("org_1234567890")
// @Param status query string false "Filter by invoice status" Enums(draft,sent,paid,overdue,void)
// @Param start_date query string false "Filter invoices from date (RFC3339)" example("2024-01-01T00:00:00Z")
// @Param end_date query string false "Filter invoices until date (RFC3339)" example("2024-01-31T23:59:59Z")
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Items per page" default(20) minimum(1) maximum(100)
// @Success 200 {object} response.APIResponse{data=[]Invoice,meta=response.Meta{pagination=response.Pagination}} "List of organization invoices with pagination"
// @Failure 400 {object} response.ErrorResponse "Bad request - invalid organization ID or parameters"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - insufficient permissions to view billing information"
// @Failure 404 {object} response.ErrorResponse "Organization not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/billing/{orgId}/invoices [get]
func (h *Handler) ListInvoices(c *gin.Context) { response.Success(c, gin.H{"message": "List invoices - TODO"}) }
// GetSubscription handles GET /billing/:orgId/subscription
// @Summary Get organization subscription
// @Description Get current subscription details for an organization
// @Tags Billing
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID" example("org_1234567890")
// @Success 200 {object} response.SuccessResponse{data=Subscription} "Organization subscription details"
// @Failure 400 {object} response.ErrorResponse "Bad request - invalid organization ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - insufficient permissions to view billing information"
// @Failure 404 {object} response.ErrorResponse "Organization or subscription not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/billing/{orgId}/subscription [get]
func (h *Handler) GetSubscription(c *gin.Context) { response.Success(c, gin.H{"message": "Get subscription - TODO"}) }
// UpdateSubscription handles POST /billing/:orgId/subscription
// @Summary Update organization subscription
// @Description Update subscription plan, add-ons, or payment method for an organization
// @Tags Billing
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID" example("org_1234567890")
// @Param request body UpdateSubscriptionRequest true "Subscription update details"
// @Success 200 {object} response.SuccessResponse{data=Subscription} "Updated subscription details"
// @Failure 400 {object} response.ErrorResponse "Bad request - invalid parameters or subscription update not allowed"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - insufficient permissions to modify billing information"
// @Failure 404 {object} response.ErrorResponse "Organization or subscription not found"
// @Failure 422 {object} response.ErrorResponse "Unprocessable entity - payment method declined or plan change not allowed"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/billing/{orgId}/subscription [post]
func (h *Handler) UpdateSubscription(c *gin.Context) { response.Success(c, gin.H{"message": "Update subscription - TODO"}) }