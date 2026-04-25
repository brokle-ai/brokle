package billing

import "time"

// ---- budgets ---------------------------------------------------------

type createBudgetBody struct {
	Name            string   `json:"name"                      validate:"required,min=1,max=100"`
	ProjectID       *string  `json:"project_id,omitempty"`
	BudgetType      string   `json:"budget_type"               validate:"required,oneof=monthly weekly"`
	SpanLimit       *int64   `json:"span_limit,omitempty"`
	BytesLimit      *int64   `json:"bytes_limit,omitempty"`
	ScoreLimit      *int64   `json:"score_limit,omitempty"`
	CostLimit       *float64 `json:"cost_limit,omitempty"`
	AlertThresholds []int64  `json:"alert_thresholds,omitempty"`
}

type updateBudgetBody struct {
	Name            *string  `json:"name,omitempty"             validate:"omitempty,min=1,max=100"`
	SpanLimit       *int64   `json:"span_limit,omitempty"`
	BytesLimit      *int64   `json:"bytes_limit,omitempty"`
	ScoreLimit      *int64   `json:"score_limit,omitempty"`
	CostLimit       *float64 `json:"cost_limit,omitempty"`
	AlertThresholds []int64  `json:"alert_thresholds,omitempty"`
	IsActive        *bool    `json:"is_active,omitempty"`
}

// ---- contracts -------------------------------------------------------

type createContractBody struct {
	OrganizationID          string                    `json:"organization_id"                   validate:"required"`
	ContractName            string                    `json:"contract_name"                     validate:"required,min=1"`
	ContractNumber          string                    `json:"contract_number"                   validate:"required,min=1"`
	StartsAt                time.Time                 `json:"starts_at"                         validate:"required"`
	ExpiresAt               *time.Time                `json:"expires_at,omitempty"`
	MinimumCommitAmount     *float64                  `json:"minimum_commit_amount,omitempty"`
	Currency                string                    `json:"currency,omitempty"`
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
	Dimension    string  `json:"dimension"        validate:"required,oneof=spans bytes scores"`
	TierMin      int64   `json:"tier_min"         validate:"min=0"`
	TierMax      *int64  `json:"tier_max,omitempty"`
	PricePerUnit float64 `json:"price_per_unit"   validate:"min=0"`
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

type cancelContractBody struct {
	Reason string `json:"reason" validate:"required,min=1"`
}

type updateContractTiersBody struct {
	Tiers []createVolumeTierRequest `json:"tiers"`
}
