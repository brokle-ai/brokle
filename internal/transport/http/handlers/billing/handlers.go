// Package billing is the dashboard-plane billing handler domain.
// Exposes usage overview / time-series / export, budgets (CRUD + alerts)
// and enterprise contracts (CRUD + volume tiers + effective pricing)
// under /api/v1/organizations/{orgId}/... and /api/v1/billing/...
//
// Every op requires RequireAuth (apiAdmin plane). The services own
// organization access validation; the handler validates path UUIDs and
// compares contract.OrganizationID to path-orgID where applicable to
// keep cross-tenant addressing honest.
package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"brokle/internal/core/domain/analytics"
	billingDomain "brokle/internal/core/domain/billing"
	"brokle/internal/core/domain/shared"
	handlersShared "brokle/internal/transport/http/handlers/shared"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/uid"
	"brokle/pkg/units"
)

type handler struct {
	usageSvc    billingDomain.BillableUsageService
	budgetSvc   billingDomain.BudgetService
	contractSvc billingDomain.ContractService
	pricingSvc  billingDomain.PricingService
	logger      *slog.Logger
}

// RegisterRoutes registers every billing operation on apiAdmin.
func RegisterRoutes(
	api huma.API,
	usageSvc billingDomain.BillableUsageService,
	budgetSvc billingDomain.BudgetService,
	contractSvc billingDomain.ContractService,
	pricingSvc billingDomain.PricingService,
	logger *slog.Logger,
) {
	h := &handler{
		usageSvc:    usageSvc,
		budgetSvc:   budgetSvc,
		contractSvc: contractSvc,
		pricingSvc:  pricingSvc,
		logger:      logger,
	}

	// ----- Usage --------------------------------------------------------
	huma.Register(api, huma.Operation{
		OperationID: "get-usage-overview",
		Method:      http.MethodGet,
		Path:        "/api/v1/organizations/{orgId}/usage/overview",
		Tags:        []string{"billing"},
		Summary:     "Get current-period usage overview",
		Description: "Returns spans/bytes/scores usage for the current billing period plus free-tier remaining and estimated cost.",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getUsageOverview)

	huma.Register(api, huma.Operation{
		OperationID: "get-usage-timeseries",
		Method:      http.MethodGet,
		Path:        "/api/v1/organizations/{orgId}/usage/timeseries",
		Tags:        []string{"billing"},
		Summary:     "Get usage time series",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getUsageTimeSeries)

	huma.Register(api, huma.Operation{
		OperationID: "get-usage-by-project",
		Method:      http.MethodGet,
		Path:        "/api/v1/organizations/{orgId}/usage/by-project",
		Tags:        []string{"billing"},
		Summary:     "Get usage broken down by project",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getUsageByProject)

	huma.Register(api, huma.Operation{
		OperationID: "export-usage",
		Method:      http.MethodGet,
		Path:        "/api/v1/organizations/{orgId}/usage/export",
		Tags:        []string{"billing"},
		Summary:     "Export usage as CSV or JSON",
		Description: "Streams a file attachment (Content-Disposition) containing usage time-series and project breakdown for the requested range.",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.exportUsage)

	// ----- Budgets ------------------------------------------------------
	huma.Register(api, huma.Operation{
		OperationID: "list-budgets",
		Method:      http.MethodGet,
		Path:        "/api/v1/organizations/{orgId}/budgets",
		Tags:        []string{"billing"},
		Summary:     "List organization budgets",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.listBudgets)

	huma.Register(api, huma.Operation{
		OperationID:   "create-budget",
		Method:        http.MethodPost,
		Path:          "/api/v1/organizations/{orgId}/budgets",
		Tags:          []string{"billing"},
		Summary:       "Create a usage budget",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.createBudget)

	huma.Register(api, huma.Operation{
		OperationID: "get-budget-alerts",
		Method:      http.MethodGet,
		Path:        "/api/v1/organizations/{orgId}/budgets/alerts",
		Tags:        []string{"billing"},
		Summary:     "List recent budget alerts",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getBudgetAlerts)

	huma.Register(api, huma.Operation{
		OperationID: "acknowledge-budget-alert",
		Method:      http.MethodPost,
		Path:        "/api/v1/organizations/{orgId}/budgets/alerts/{alertId}/acknowledge",
		Tags:        []string{"billing"},
		Summary:     "Acknowledge a budget alert",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.acknowledgeBudgetAlert)

	huma.Register(api, huma.Operation{
		OperationID: "get-budget",
		Method:      http.MethodGet,
		Path:        "/api/v1/organizations/{orgId}/budgets/{budgetId}",
		Tags:        []string{"billing"},
		Summary:     "Get a budget by ID",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getBudget)

	huma.Register(api, huma.Operation{
		OperationID: "update-budget",
		Method:      http.MethodPut,
		Path:        "/api/v1/organizations/{orgId}/budgets/{budgetId}",
		Tags:        []string{"billing"},
		Summary:     "Update a budget",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.updateBudget)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-budget",
		Method:        http.MethodDelete,
		Path:          "/api/v1/organizations/{orgId}/budgets/{budgetId}",
		Tags:          []string{"billing"},
		Summary:       "Delete a budget",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.deleteBudget)

	// ----- Contracts ----------------------------------------------------
	huma.Register(api, huma.Operation{
		OperationID:   "create-contract",
		Method:        http.MethodPost,
		Path:          "/api/v1/billing/contracts",
		Tags:          []string{"billing-contracts"},
		Summary:       "Create an enterprise contract",
		Description:   "Timestamps must be RFC3339. Access rule: now < expires_at. Volume tiers may be attached in the same request.",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.createContract)

	huma.Register(api, huma.Operation{
		OperationID: "list-contracts",
		Method:      http.MethodGet,
		Path:        "/api/v1/billing/organizations/{orgId}/contracts",
		Tags:        []string{"billing-contracts"},
		Summary:     "List contracts for an organization",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.listContracts)

	huma.Register(api, huma.Operation{
		OperationID: "get-effective-pricing",
		Method:      http.MethodGet,
		Path:        "/api/v1/billing/organizations/{orgId}/effective-pricing",
		Tags:        []string{"billing-contracts"},
		Summary:     "Get effective pricing (contract overrides > plan defaults)",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getEffectivePricing)

	huma.Register(api, huma.Operation{
		OperationID: "get-contract",
		Method:      http.MethodGet,
		Path:        "/api/v1/billing/contracts/{contractId}",
		Tags:        []string{"billing-contracts"},
		Summary:     "Get a contract by ID",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getContract)

	huma.Register(api, huma.Operation{
		OperationID: "update-contract",
		Method:      http.MethodPut,
		Path:        "/api/v1/billing/contracts/{contractId}",
		Tags:        []string{"billing-contracts"},
		Summary:     "Update a contract",
		Description: "Cannot change pricing after activation. Timestamps must be RFC3339; minimum duration is 1 day.",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.updateContract)

	huma.Register(api, huma.Operation{
		OperationID: "cancel-contract",
		Method:      http.MethodDelete,
		Path:        "/api/v1/billing/contracts/{contractId}",
		Tags:        []string{"billing-contracts"},
		Summary:     "Cancel an active contract",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.cancelContract)

	huma.Register(api, huma.Operation{
		OperationID: "activate-contract",
		Method:      http.MethodPut,
		Path:        "/api/v1/billing/contracts/{contractId}/activate",
		Tags:        []string{"billing-contracts"},
		Summary:     "Activate a draft contract",
		Description: "Expires any currently-active contract for the organization.",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.activateContract)

	huma.Register(api, huma.Operation{
		OperationID: "update-contract-tiers",
		Method:      http.MethodPut,
		Path:        "/api/v1/billing/contracts/{contractId}/tiers",
		Tags:        []string{"billing-contracts"},
		Summary:     "Replace volume-discount tiers for a contract",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.updateContractTiers)

	huma.Register(api, huma.Operation{
		OperationID: "get-contract-history",
		Method:      http.MethodGet,
		Path:        "/api/v1/billing/contracts/{contractId}/history",
		Tags:        []string{"billing-contracts"},
		Summary:     "Get contract audit history",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getContractHistory)
}

// ----- shared parsers --------------------------------------------------

func parseOrg(orgIDStr string) (uuid.UUID, error) {
	id, err := uuid.Parse(orgIDStr)
	if err != nil {
		return uuid.Nil, appErrors.NewValidationError("Invalid organization ID", "orgId must be a valid UUID")
	}
	return id, nil
}

func parseBudget(budgetIDStr string) (uuid.UUID, error) {
	id, err := uuid.Parse(budgetIDStr)
	if err != nil {
		return uuid.Nil, appErrors.NewValidationError("Invalid budget ID", "budgetId must be a valid UUID")
	}
	return id, nil
}

func parseContract(contractIDStr string) (uuid.UUID, error) {
	id, err := uuid.Parse(contractIDStr)
	if err != nil {
		return uuid.Nil, appErrors.NewValidationError("Invalid contract ID", "contractId must be a valid UUID")
	}
	return id, nil
}

func parseAlert(alertIDStr string) (uuid.UUID, error) {
	id, err := uuid.Parse(alertIDStr)
	if err != nil {
		return uuid.Nil, appErrors.NewValidationError("Invalid alert ID", "alertId must be a valid UUID")
	}
	return id, nil
}

func float64ToDecimalPtr(f *float64) *decimal.Decimal {
	if f == nil {
		return nil
	}
	d := decimal.NewFromFloat(*f)
	return &d
}

// ============================================================================
// Usage
// ============================================================================

func (h *handler) getUsageOverview(ctx context.Context, in *GetUsageOverviewInput) (*GetUsageOverviewOutput, error) {
	orgID, err := parseOrg(in.OrgID)
	if err != nil {
		return nil, err
	}
	overview, err := h.usageSvc.GetUsageOverview(ctx, orgID)
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to get usage overview", err)
	}
	return &GetUsageOverviewOutput{Body: overview}, nil
}

func (h *handler) getUsageTimeSeries(ctx context.Context, in *UsageTimeSeriesInput) (*UsageTimeSeriesOutput, error) {
	orgID, err := parseOrg(in.OrgID)
	if err != nil {
		return nil, err
	}
	from, to, err := handlersShared.ParseTimeRange(in.From, in.To, in.TimeRange, analytics.TimeRange30Days)
	if err != nil {
		return nil, err
	}
	granularity := in.Granularity
	if granularity == "" {
		if to.Sub(from) > 7*24*time.Hour {
			granularity = "daily"
		} else {
			granularity = "hourly"
		}
	}
	usage, err := h.usageSvc.GetUsageTimeSeries(ctx, orgID, from, to, granularity)
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to get usage time series", err)
	}
	return &UsageTimeSeriesOutput{Body: usage}, nil
}

func (h *handler) getUsageByProject(ctx context.Context, in *UsageByProjectInput) (*UsageByProjectOutput, error) {
	orgID, err := parseOrg(in.OrgID)
	if err != nil {
		return nil, err
	}
	from, to, err := handlersShared.ParseTimeRange(in.From, in.To, in.TimeRange, analytics.TimeRange30Days)
	if err != nil {
		return nil, err
	}
	summaries, err := h.usageSvc.GetUsageByProject(ctx, orgID, from, to)
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to get usage by project", err)
	}
	return &UsageByProjectOutput{Body: summaries}, nil
}

func (h *handler) exportUsage(ctx context.Context, in *ExportUsageInput) (*ExportUsageOutput, error) {
	orgID, err := parseOrg(in.OrgID)
	if err != nil {
		return nil, err
	}
	from, to, err := handlersShared.ParseTimeRange(in.From, in.To, in.TimeRange, analytics.TimeRange30Days)
	if err != nil {
		return nil, err
	}
	format := in.Format
	if format == "" {
		format = "csv"
	}
	granularity := in.Granularity
	if granularity == "" {
		if to.Sub(from) > 7*24*time.Hour {
			granularity = "daily"
		} else {
			granularity = "hourly"
		}
	}

	usage, err := h.usageSvc.GetUsageTimeSeries(ctx, orgID, from, to, granularity)
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to get usage for export", err)
	}
	projectUsage, pErr := h.usageSvc.GetUsageByProject(ctx, orgID, from, to)
	if pErr != nil {
		// non-fatal — export without project breakdown
		projectUsage = nil
	}

	filenameBase := "usage_export_" + from.Format("2006-01-02") + "_to_" + to.Format("2006-01-02")

	if format == "json" {
		payload := map[string]any{
			"period": map[string]string{
				"from": from.Format(time.RFC3339),
				"to":   to.Format(time.RFC3339),
			},
			"time_series":  usage,
			"by_project":   projectUsage,
			"generated_at": time.Now().Format(time.RFC3339),
		}
		body, mErr := json.Marshal(payload)
		if mErr != nil {
			return nil, appErrors.NewInternalError("Failed to encode usage export", mErr)
		}
		return &ExportUsageOutput{
			ContentType:        "application/json",
			ContentDisposition: "attachment; filename=" + filenameBase + ".json",
			Body:               body,
		}, nil
	}

	// CSV
	var buf []byte
	buf = append(buf, "date,organization_id,project_id,spans,bytes_processed,gb_processed,scores,ai_provider_cost\n"...)
	for _, u := range usage {
		gb := float64(u.BytesProcessed) / float64(units.BytesPerGB)
		line := fmt.Sprintf("%s,%s,%s,%d,%d,%s,%d,%s\n",
			u.BucketTime.Format("2006-01-02"),
			u.OrganizationID.String(),
			u.ProjectID.String(),
			u.SpanCount,
			u.BytesProcessed,
			strconv.FormatFloat(gb, 'f', 4, 64),
			u.ScoreCount,
			u.AIProviderCost.StringFixed(2),
		)
		buf = append(buf, line...)
	}
	return &ExportUsageOutput{
		ContentType:        "text/csv",
		ContentDisposition: "attachment; filename=" + filenameBase + ".csv",
		Body:               buf,
	}, nil
}

// ============================================================================
// Budgets
// ============================================================================

func (h *handler) listBudgets(ctx context.Context, in *ListBudgetsInput) (*ListBudgetsOutput, error) {
	orgID, err := parseOrg(in.OrgID)
	if err != nil {
		return nil, err
	}
	budgets, err := h.budgetSvc.GetBudgetsByOrg(ctx, orgID)
	if err != nil {
		h.logger.Error("failed to list budgets", "error", err, "organization_id", orgID)
		return nil, appErrors.NewInternalError("Failed to list budgets", err)
	}
	return &ListBudgetsOutput{Body: budgets}, nil
}

func (h *handler) getBudget(ctx context.Context, in *GetBudgetInput) (*GetBudgetOutput, error) {
	orgID, err := parseOrg(in.OrgID)
	if err != nil {
		return nil, err
	}
	budgetID, err := parseBudget(in.BudgetID)
	if err != nil {
		return nil, err
	}
	budget, err := h.budgetSvc.GetBudget(ctx, budgetID)
	if err != nil {
		return nil, appErrors.NewNotFoundError("Budget not found")
	}
	if budget.OrganizationID != orgID {
		return nil, appErrors.NewForbiddenError("Access denied to this budget")
	}
	return &GetBudgetOutput{Body: budget}, nil
}

func (h *handler) createBudget(ctx context.Context, in *CreateBudgetInput) (*CreateBudgetOutput, error) {
	orgID, err := parseOrg(in.OrgID)
	if err != nil {
		return nil, err
	}

	if in.Body.SpanLimit == nil && in.Body.BytesLimit == nil && in.Body.ScoreLimit == nil && in.Body.CostLimit == nil {
		return nil, appErrors.NewValidationError(
			"At least one limit is required",
			"Set span_limit, bytes_limit, score_limit, or cost_limit",
		)
	}

	thresholds := in.Body.AlertThresholds
	if thresholds == nil {
		thresholds = []int64{50, 80, 100}
	}
	sort.Slice(thresholds, func(i, j int) bool { return thresholds[i] < thresholds[j] })

	budget := &billingDomain.UsageBudget{
		OrganizationID:  orgID,
		Name:            in.Body.Name,
		BudgetType:      billingDomain.BudgetType(in.Body.BudgetType),
		SpanLimit:       in.Body.SpanLimit,
		BytesLimit:      in.Body.BytesLimit,
		ScoreLimit:      in.Body.ScoreLimit,
		CostLimit:       float64ToDecimalPtr(in.Body.CostLimit),
		AlertThresholds: thresholds,
	}

	if in.Body.ProjectID != nil {
		projectID, pErr := uuid.Parse(*in.Body.ProjectID)
		if pErr != nil {
			return nil, appErrors.NewValidationError("Invalid project_id", "project_id must be a valid UUID")
		}
		budget.ProjectID = &projectID
	}

	if err := h.budgetSvc.CreateBudget(ctx, budget); err != nil {
		h.logger.Error("failed to create budget", "error", err, "organization_id", orgID)
		return nil, err
	}
	return &CreateBudgetOutput{Body: budget}, nil
}

func (h *handler) updateBudget(ctx context.Context, in *UpdateBudgetInput) (*UpdateBudgetOutput, error) {
	orgID, err := parseOrg(in.OrgID)
	if err != nil {
		return nil, err
	}
	budgetID, err := parseBudget(in.BudgetID)
	if err != nil {
		return nil, err
	}
	budget, err := h.budgetSvc.GetBudget(ctx, budgetID)
	if err != nil {
		return nil, appErrors.NewNotFoundError("Budget not found")
	}
	if budget.OrganizationID != orgID {
		return nil, appErrors.NewForbiddenError("Access denied to this budget")
	}

	if in.Body.Name != nil {
		budget.Name = *in.Body.Name
	}
	if in.Body.SpanLimit != nil {
		budget.SpanLimit = in.Body.SpanLimit
	}
	if in.Body.BytesLimit != nil {
		budget.BytesLimit = in.Body.BytesLimit
	}
	if in.Body.ScoreLimit != nil {
		budget.ScoreLimit = in.Body.ScoreLimit
	}
	if in.Body.CostLimit != nil {
		budget.CostLimit = float64ToDecimalPtr(in.Body.CostLimit)
	}
	if in.Body.AlertThresholds != nil {
		sort.Slice(in.Body.AlertThresholds, func(i, j int) bool {
			return in.Body.AlertThresholds[i] < in.Body.AlertThresholds[j]
		})
		budget.AlertThresholds = in.Body.AlertThresholds
	}
	if in.Body.IsActive != nil {
		budget.IsActive = *in.Body.IsActive
	}

	if err := h.budgetSvc.UpdateBudget(ctx, budget); err != nil {
		h.logger.Error("failed to update budget", "error", err, "budget_id", budgetID)
		return nil, appErrors.NewInternalError("Failed to update budget", err)
	}
	return &UpdateBudgetOutput{Body: budget}, nil
}

func (h *handler) deleteBudget(ctx context.Context, in *DeleteBudgetInput) (*DeleteBudgetOutput, error) {
	orgID, err := parseOrg(in.OrgID)
	if err != nil {
		return nil, err
	}
	budgetID, err := parseBudget(in.BudgetID)
	if err != nil {
		return nil, err
	}
	budget, err := h.budgetSvc.GetBudget(ctx, budgetID)
	if err != nil {
		return nil, appErrors.NewNotFoundError("Budget not found")
	}
	if budget.OrganizationID != orgID {
		return nil, appErrors.NewForbiddenError("Access denied to this budget")
	}
	if err := h.budgetSvc.DeleteBudget(ctx, budgetID); err != nil {
		h.logger.Error("failed to delete budget", "error", err, "budget_id", budgetID)
		return nil, appErrors.NewInternalError("Failed to delete budget", err)
	}
	return &DeleteBudgetOutput{}, nil
}

func (h *handler) getBudgetAlerts(ctx context.Context, in *GetBudgetAlertsInput) (*GetBudgetAlertsOutput, error) {
	orgID, err := parseOrg(in.OrgID)
	if err != nil {
		return nil, err
	}
	limit := in.Limit
	if limit <= 0 {
		limit = 50
	}
	alerts, err := h.budgetSvc.GetAlerts(ctx, orgID, limit)
	if err != nil {
		h.logger.Error("failed to get alerts", "error", err, "organization_id", orgID)
		return nil, appErrors.NewInternalError("Failed to get alerts", err)
	}
	return &GetBudgetAlertsOutput{Body: alerts}, nil
}

func (h *handler) acknowledgeBudgetAlert(ctx context.Context, in *AcknowledgeBudgetAlertInput) (*AcknowledgeBudgetAlertOutput, error) {
	orgID, err := parseOrg(in.OrgID)
	if err != nil {
		return nil, err
	}
	alertID, err := parseAlert(in.AlertID)
	if err != nil {
		return nil, err
	}
	if err := h.budgetSvc.AcknowledgeAlert(ctx, orgID, alertID); err != nil {
		h.logger.Error("failed to acknowledge alert", "error", err, "alert_id", alertID)
		if billingDomain.IsNotFoundError(err) {
			return nil, appErrors.NewNotFoundError("Alert not found")
		}
		return nil, appErrors.NewInternalError("Failed to acknowledge alert", err)
	}
	out := &AcknowledgeBudgetAlertOutput{}
	out.Body.Acknowledged = true
	return out, nil
}

// ============================================================================
// Contracts
// ============================================================================

func (h *handler) createContract(ctx context.Context, in *CreateContractInput) (*CreateContractOutput, error) {
	orgID, err := uuid.Parse(in.Body.OrganizationID)
	if err != nil {
		return nil, appErrors.NewValidationError("Invalid organization ID", "organization_id must be a valid UUID")
	}
	userID := httpctx.MustGetUserID(ctx)

	contract := &billingDomain.Contract{
		ID:                      uid.New(),
		OrganizationID:          orgID,
		ContractName:            in.Body.ContractName,
		ContractNumber:          in.Body.ContractNumber,
		StartDate:               in.Body.StartsAt,
		EndDate:                 in.Body.ExpiresAt,
		MinimumCommitAmount:     float64ToDecimalPtr(in.Body.MinimumCommitAmount),
		Currency:                "USD",
		AccountOwner:            shared.NilIfEmpty(in.Body.AccountOwner),
		SalesRepEmail:           shared.NilIfEmpty(in.Body.SalesRepEmail),
		Status:                  billingDomain.ContractStatusDraft,
		CustomFreeSpans:         in.Body.CustomFreeSpans,
		CustomPricePer100KSpans: float64ToDecimalPtr(in.Body.CustomPricePer100KSpans),
		CustomFreeGB:            float64ToDecimalPtr(in.Body.CustomFreeGB),
		CustomPricePerGB:        float64ToDecimalPtr(in.Body.CustomPricePerGB),
		CustomFreeScores:        in.Body.CustomFreeScores,
		CustomPricePer1KScores:  float64ToDecimalPtr(in.Body.CustomPricePer1KScores),
		CreatedBy:               userID.String(),
		CreatedAt:               time.Now(),
		UpdatedAt:               time.Now(),
		Notes:                   shared.NilIfEmpty(in.Body.Notes),
	}
	if in.Body.Currency != "" {
		contract.Currency = in.Body.Currency
	}

	if err := h.contractSvc.CreateContract(ctx, contract); err != nil {
		h.logger.Error("failed to create contract", "error", err, "organization_id", orgID)
		return nil, err
	}

	if len(in.Body.VolumeTiers) > 0 {
		tiers := make([]*billingDomain.VolumeDiscountTier, len(in.Body.VolumeTiers))
		for i, tr := range in.Body.VolumeTiers {
			tiers[i] = &billingDomain.VolumeDiscountTier{
				ID:           uid.New(),
				ContractID:   contract.ID,
				Dimension:    billingDomain.TierDimension(tr.Dimension),
				TierMin:      tr.TierMin,
				TierMax:      tr.TierMax,
				PricePerUnit: decimal.NewFromFloat(tr.PricePerUnit),
				Priority:     i,
				CreatedAt:    time.Now(),
			}
		}
		if err := h.contractSvc.AddVolumeTiers(ctx, contract.ID, tiers); err != nil {
			h.logger.Error("failed to add volume tiers", "error", err, "contract_id", contract.ID)
			return nil, err
		}
	}

	result, err := h.contractSvc.GetContract(ctx, contract.ID)
	if err != nil {
		return nil, err
	}
	return &CreateContractOutput{Body: result}, nil
}

func (h *handler) getContract(ctx context.Context, in *GetContractInput) (*GetContractOutput, error) {
	contractID, err := parseContract(in.ContractID)
	if err != nil {
		return nil, err
	}
	contract, err := h.contractSvc.GetContract(ctx, contractID)
	if err != nil {
		return nil, err
	}
	return &GetContractOutput{Body: contract}, nil
}

func (h *handler) listContracts(ctx context.Context, in *ListContractsInput) (*ListContractsOutput, error) {
	orgID, err := parseOrg(in.OrgID)
	if err != nil {
		return nil, err
	}
	contracts, err := h.contractSvc.GetContractsByOrg(ctx, orgID)
	if err != nil {
		h.logger.Error("failed to get contracts", "error", err, "organization_id", orgID)
		return nil, appErrors.NewInternalError("Failed to get contracts", err)
	}
	return &ListContractsOutput{Body: contracts}, nil
}

func (h *handler) updateContract(ctx context.Context, in *UpdateContractInput) (*UpdateContractOutput, error) {
	contractID, err := parseContract(in.ContractID)
	if err != nil {
		return nil, err
	}
	contract, err := h.contractSvc.GetContract(ctx, contractID)
	if err != nil {
		return nil, err
	}

	b := in.Body
	if b.ContractName != nil {
		contract.ContractName = *b.ContractName
	}
	if b.StartsAt != nil {
		contract.StartDate = *b.StartsAt
	}
	if b.ExpiresAt != nil {
		contract.EndDate = b.ExpiresAt
	}
	if b.MinimumCommitAmount != nil {
		contract.MinimumCommitAmount = float64ToDecimalPtr(b.MinimumCommitAmount)
	}
	if b.AccountOwner != nil {
		contract.AccountOwner = b.AccountOwner
	}
	if b.SalesRepEmail != nil {
		contract.SalesRepEmail = b.SalesRepEmail
	}
	if b.CustomFreeSpans != nil {
		contract.CustomFreeSpans = b.CustomFreeSpans
	}
	if b.CustomPricePer100KSpans != nil {
		contract.CustomPricePer100KSpans = float64ToDecimalPtr(b.CustomPricePer100KSpans)
	}
	if b.CustomFreeGB != nil {
		contract.CustomFreeGB = float64ToDecimalPtr(b.CustomFreeGB)
	}
	if b.CustomPricePerGB != nil {
		contract.CustomPricePerGB = float64ToDecimalPtr(b.CustomPricePerGB)
	}
	if b.CustomFreeScores != nil {
		contract.CustomFreeScores = b.CustomFreeScores
	}
	if b.CustomPricePer1KScores != nil {
		contract.CustomPricePer1KScores = float64ToDecimalPtr(b.CustomPricePer1KScores)
	}
	if b.Notes != nil {
		contract.Notes = b.Notes
	}
	contract.UpdatedAt = time.Now()

	if err := h.contractSvc.UpdateContract(ctx, contract); err != nil {
		h.logger.Error("failed to update contract", "error", err, "contract_id", contractID)
		return nil, err
	}
	return &UpdateContractOutput{Body: contract}, nil
}

func (h *handler) activateContract(ctx context.Context, in *ActivateContractInput) (*ActivateContractOutput, error) {
	contractID, err := parseContract(in.ContractID)
	if err != nil {
		return nil, err
	}
	userID := httpctx.MustGetUserID(ctx)

	if err := h.contractSvc.ActivateContract(ctx, contractID, userID); err != nil {
		h.logger.Error("failed to activate contract", "error", err, "contract_id", contractID)
		return nil, err
	}
	out := &ActivateContractOutput{}
	out.Body.Message = "Contract activated successfully"
	out.Body.Status = "active"
	return out, nil
}

func (h *handler) cancelContract(ctx context.Context, in *CancelContractInput) (*CancelContractOutput, error) {
	contractID, err := parseContract(in.ContractID)
	if err != nil {
		return nil, err
	}
	userID := httpctx.MustGetUserID(ctx)
	if err := h.contractSvc.CancelContract(ctx, contractID, in.Body.Reason, userID); err != nil {
		h.logger.Error("failed to cancel contract", "error", err, "contract_id", contractID)
		return nil, err
	}
	out := &CancelContractOutput{}
	out.Body.Message = "Contract cancelled successfully"
	out.Body.Status = "cancelled"
	return out, nil
}

func (h *handler) updateContractTiers(ctx context.Context, in *UpdateContractTiersInput) (*UpdateContractTiersOutput, error) {
	contractID, err := parseContract(in.ContractID)
	if err != nil {
		return nil, err
	}
	tiers := make([]*billingDomain.VolumeDiscountTier, len(in.Body.Tiers))
	for i, tr := range in.Body.Tiers {
		tiers[i] = &billingDomain.VolumeDiscountTier{
			ID:           uid.New(),
			ContractID:   contractID,
			Dimension:    billingDomain.TierDimension(tr.Dimension),
			TierMin:      tr.TierMin,
			TierMax:      tr.TierMax,
			PricePerUnit: decimal.NewFromFloat(tr.PricePerUnit),
			Priority:     i,
			CreatedAt:    time.Now(),
		}
	}
	if err := h.contractSvc.UpdateVolumeTiers(ctx, contractID, tiers); err != nil {
		h.logger.Error("failed to update volume tiers", "error", err, "contract_id", contractID)
		return nil, err
	}
	out := &UpdateContractTiersOutput{}
	out.Body.Message = "Volume tiers updated successfully"
	out.Body.TiersCount = len(tiers)
	return out, nil
}

func (h *handler) getContractHistory(ctx context.Context, in *GetContractHistoryInput) (*GetContractHistoryOutput, error) {
	contractID, err := parseContract(in.ContractID)
	if err != nil {
		return nil, err
	}
	history, err := h.contractSvc.GetContractHistory(ctx, contractID)
	if err != nil {
		h.logger.Error("failed to get contract history", "error", err, "contract_id", contractID)
		return nil, appErrors.NewInternalError("Failed to get contract history", err)
	}
	return &GetContractHistoryOutput{Body: history}, nil
}

func (h *handler) getEffectivePricing(ctx context.Context, in *GetEffectivePricingInput) (*GetEffectivePricingOutput, error) {
	orgID, err := parseOrg(in.OrgID)
	if err != nil {
		return nil, err
	}
	effective, err := h.pricingSvc.GetEffectivePricing(ctx, orgID)
	if err != nil {
		h.logger.Error("failed to get effective pricing", "error", err, "organization_id", orgID)
		return nil, appErrors.NewInternalError("Failed to get effective pricing", err)
	}
	return &GetEffectivePricingOutput{Body: effective}, nil
}
