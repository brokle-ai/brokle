package billing

import (
	billingDomain "brokle/internal/core/domain/billing"
	"time"
)

// Huma operation types for the billing package.

type GetUsageOverviewInput struct {
	OrgID string `path:"orgId" format:"uuid"`
}

type GetUsageOverviewOutput struct {
	Body *billingDomain.UsageOverview
}

type UsageTimeSeriesInput struct {
	OrgID       string `path:"orgId" format:"uuid"`
	TimeRange   string `query:"time_range" required:"false" enum:"15m,30m,1h,3h,6h,12h,24h,7d,14d,30d"`
	From        string `query:"from" required:"false" doc:"Custom range start (RFC3339)"`
	To          string `query:"to" required:"false" doc:"Custom range end (RFC3339)"`
	Granularity string `query:"granularity" required:"false" enum:"hourly,daily"`
}

type UsageTimeSeriesOutput struct {
	Body []*billingDomain.BillableUsage
}

type UsageByProjectInput struct {
	OrgID     string `path:"orgId" format:"uuid"`
	TimeRange string `query:"time_range" required:"false" enum:"15m,30m,1h,3h,6h,12h,24h,7d,14d,30d"`
	From      string `query:"from" required:"false"`
	To        string `query:"to" required:"false"`
}

type UsageByProjectOutput struct {
	Body []*billingDomain.BillableUsageSummary
}

type ExportUsageInput struct {
	OrgID       string `path:"orgId" format:"uuid"`
	TimeRange   string `query:"time_range" required:"false" enum:"15m,30m,1h,3h,6h,12h,24h,7d,14d,30d"`
	From        string `query:"from" required:"false"`
	To          string `query:"to" required:"false"`
	Format      string `query:"format" required:"false" enum:"csv,json"`
	Granularity string `query:"granularity" required:"false" enum:"hourly,daily"`
}

// ExportUsageOutput streams a file attachment. Body bypasses the
// APIResponse envelope because the client is expected to save the
// payload directly to disk.
type ExportUsageOutput struct {
	ContentType        string `header:"Content-Type"`
	ContentDisposition string `header:"Content-Disposition"`
	Body               []byte
}

type ListBudgetsInput struct {
	OrgID string `path:"orgId" format:"uuid"`
}

type ListBudgetsOutput struct {
	Body []*billingDomain.UsageBudget
}

type GetBudgetInput struct {
	OrgID    string `path:"orgId" format:"uuid"`
	BudgetID string `path:"budgetId" format:"uuid"`
}

type GetBudgetOutput struct {
	Body *billingDomain.UsageBudget
}

type CreateBudgetInput struct {
	OrgID string `path:"orgId" format:"uuid"`
	Body  createBudgetBody
}

type createBudgetBody struct {
	Name            string   `json:"name" minLength:"1" maxLength:"100"`
	ProjectID       *string  `json:"project_id,omitempty" doc:"Optional project scope (UUID)"`
	BudgetType      string   `json:"budget_type" enum:"monthly,weekly"`
	SpanLimit       *int64   `json:"span_limit,omitempty"`
	BytesLimit      *int64   `json:"bytes_limit,omitempty"`
	ScoreLimit      *int64   `json:"score_limit,omitempty"`
	CostLimit       *float64 `json:"cost_limit,omitempty"`
	AlertThresholds []int64  `json:"alert_thresholds,omitempty" doc:"Percentages (e.g. [50,80,100]). Omit for default; empty [] disables alerts."`
}

type CreateBudgetOutput struct {
	Body *billingDomain.UsageBudget
}

type UpdateBudgetInput struct {
	OrgID    string `path:"orgId" format:"uuid"`
	BudgetID string `path:"budgetId" format:"uuid"`
	Body     updateBudgetBody
}

type updateBudgetBody struct {
	Name            *string  `json:"name,omitempty" minLength:"1" maxLength:"100"`
	SpanLimit       *int64   `json:"span_limit,omitempty"`
	BytesLimit      *int64   `json:"bytes_limit,omitempty"`
	ScoreLimit      *int64   `json:"score_limit,omitempty"`
	CostLimit       *float64 `json:"cost_limit,omitempty"`
	AlertThresholds []int64  `json:"alert_thresholds,omitempty"`
	IsActive        *bool    `json:"is_active,omitempty"`
}

type UpdateBudgetOutput struct {
	Body *billingDomain.UsageBudget
}

type DeleteBudgetInput struct {
	OrgID    string `path:"orgId" format:"uuid"`
	BudgetID string `path:"budgetId" format:"uuid"`
}

type DeleteBudgetOutput struct{}

type GetBudgetAlertsInput struct {
	OrgID string `path:"orgId" format:"uuid"`
	Limit int    `query:"limit" required:"false" minimum:"1" maximum:"100" default:"50"`
}

type GetBudgetAlertsOutput struct {
	Body []*billingDomain.UsageAlert
}

type AcknowledgeBudgetAlertInput struct {
	OrgID   string `path:"orgId" format:"uuid"`
	AlertID string `path:"alertId" format:"uuid"`
}

type AcknowledgeBudgetAlertOutput struct {
	Body struct {
		Acknowledged bool `json:"acknowledged"`
	}
}

type CreateContractInput struct {
	Body createContractBody
}

type createContractBody struct {
	OrganizationID          string                    `json:"organization_id" format:"uuid"`
	ContractName            string                    `json:"contract_name" minLength:"1"`
	ContractNumber          string                    `json:"contract_number" minLength:"1"`
	StartsAt                time.Time                 `json:"starts_at" doc:"RFC3339 contract start"`
	ExpiresAt               *time.Time                `json:"expires_at,omitempty" doc:"RFC3339 expiry (null = no expiration)"`
	MinimumCommitAmount     *float64                  `json:"minimum_commit_amount,omitempty"`
	Currency                string                    `json:"currency,omitempty" doc:"Defaults to USD"`
	AccountOwner            string                    `json:"account_owner,omitempty"`
	SalesRepEmail           string                    `json:"sales_rep_email,omitempty"`
	CustomFreeSpans         *int64                    `json:"custom_free_spans,omitempty"`
	CustomPricePer100KSpans *float64                  `json:"custom_price_per_100k_spans,omitempty"`
	CustomFreeGB            *float64                  `json:"custom_free_gb,omitempty"`
	CustomPricePerGB        *float64                  `json:"custom_price_per_gb,omitempty"`
	CustomFreeScores        *int64                    `json:"custom_free_scores,omitempty"`
	CustomPricePer1KScores  *float64                  `json:"custom_price_per_1k_scores,omitempty"`
	Notes                   string                    `json:"notes,omitempty"`
	VolumeTiers             []createVolumeTierRequest `json:"volume_tiers,omitempty"`
}

type createVolumeTierRequest struct {
	Dimension    string  `json:"dimension" enum:"spans,bytes,scores"`
	TierMin      int64   `json:"tier_min" minimum:"0"`
	TierMax      *int64  `json:"tier_max,omitempty"`
	PricePerUnit float64 `json:"price_per_unit" minimum:"0"`
}

type CreateContractOutput struct {
	Body *billingDomain.Contract
}

type GetContractInput struct {
	ContractID string `path:"contractId" format:"uuid"`
}

type GetContractOutput struct {
	Body *billingDomain.Contract
}

type ListContractsInput struct {
	OrgID string `path:"orgId" format:"uuid"`
}

type ListContractsOutput struct {
	Body []*billingDomain.Contract
}

type UpdateContractInput struct {
	ContractID string `path:"contractId" format:"uuid"`
	Body       updateContractBody
}

type updateContractBody struct {
	ContractName            *string    `json:"contract_name,omitempty"`
	StartsAt                *time.Time `json:"starts_at,omitempty"`
	ExpiresAt               *time.Time `json:"expires_at,omitempty"`
	MinimumCommitAmount     *float64   `json:"minimum_commit_amount,omitempty"`
	AccountOwner            *string    `json:"account_owner,omitempty"`
	SalesRepEmail           *string    `json:"sales_rep_email,omitempty"`
	CustomFreeSpans         *int64     `json:"custom_free_spans,omitempty"`
	CustomPricePer100KSpans *float64   `json:"custom_price_per_100k_spans,omitempty"`
	CustomFreeGB            *float64   `json:"custom_free_gb,omitempty"`
	CustomPricePerGB        *float64   `json:"custom_price_per_gb,omitempty"`
	CustomFreeScores        *int64     `json:"custom_free_scores,omitempty"`
	CustomPricePer1KScores  *float64   `json:"custom_price_per_1k_scores,omitempty"`
	Notes                   *string    `json:"notes,omitempty"`
}

type UpdateContractOutput struct {
	Body *billingDomain.Contract
}

type ActivateContractInput struct {
	ContractID string `path:"contractId" format:"uuid"`
}

type ActivateContractOutput struct {
	Body struct {
		Message string `json:"message"`
		Status  string `json:"status"`
	}
}

type CancelContractInput struct {
	ContractID string `path:"contractId" format:"uuid"`
	Body       struct {
		Reason string `json:"reason" minLength:"1"`
	}
}

type CancelContractOutput struct {
	Body struct {
		Message string `json:"message"`
		Status  string `json:"status"`
	}
}

type UpdateContractTiersInput struct {
	ContractID string `path:"contractId" format:"uuid"`
	Body       struct {
		Tiers []createVolumeTierRequest `json:"tiers"`
	}
}

type UpdateContractTiersOutput struct {
	Body struct {
		Message    string `json:"message"`
		TiersCount int    `json:"tiers_count"`
	}
}

type GetContractHistoryInput struct {
	ContractID string `path:"contractId" format:"uuid"`
}

type GetContractHistoryOutput struct {
	Body []*billingDomain.ContractHistory
}

type GetEffectivePricingInput struct {
	OrgID string `path:"orgId" format:"uuid"`
}

type GetEffectivePricingOutput struct {
	Body *billingDomain.EffectivePricing
}
