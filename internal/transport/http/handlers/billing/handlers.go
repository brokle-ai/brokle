// Package billing is the dashboard-plane billing handler domain.
// Exposes usage overview / time-series / export, budgets (CRUD + alerts)
// and enterprise contracts (CRUD + volume tiers + effective pricing)
// under /api/v1/organizations/{orgId}/... and /api/v1/billing/...
//
// Every op requires RequireAuth. The services own organization access
// validation; the handler validates path UUIDs and compares
// contract.OrganizationID to path-orgID where applicable to keep
// cross-tenant addressing honest.
package billing

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"brokle/internal/core/domain/analytics"
	billingDomain "brokle/internal/core/domain/billing"
	"brokle/internal/core/domain/shared"
	billingService "brokle/internal/core/services/billing"
	handlersShared "brokle/internal/transport/http/handlers/shared"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/request"
	"brokle/pkg/response"
	"brokle/pkg/uid"
	"brokle/pkg/units"
)

type Handler struct {
	usageSvc    *billingService.BillableUsageService
	budgetSvc   *billingService.BudgetService
	contractSvc *billingService.ContractService
	pricingSvc  *billingService.PricingService
	logger      *slog.Logger
}

// New constructs the billing handler.
func New(
	usageSvc *billingService.BillableUsageService,
	budgetSvc *billingService.BudgetService,
	contractSvc *billingService.ContractService,
	pricingSvc *billingService.PricingService,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		usageSvc:    usageSvc,
		budgetSvc:   budgetSvc,
		contractSvc: contractSvc,
		pricingSvc:  pricingSvc,
		logger:      logger,
	}
}

// ----- helpers ---------------------------------------------------------

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

func (h *Handler) GetUsageOverview(w http.ResponseWriter, r *http.Request) {
	orgID := httpctx.MustGetOrganizationID(r.Context())
	overview, err := h.usageSvc.GetUsageOverview(r.Context(), orgID)
	if err != nil {
		response.WriteError(w, appErrors.Internal("Failed to get usage overview", err))
		return
	}
	response.Success(w, overview)
}

func (h *Handler) GetUsageTimeSeries(w http.ResponseWriter, r *http.Request) {
	orgID := httpctx.MustGetOrganizationID(r.Context())
	q := r.URL.Query()
	from, to, err := handlersShared.ParseTimeRange(
		q.Get("from"), q.Get("to"), q.Get("time_range"),
		analytics.TimeRange30Days,
	)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	granularity := q.Get("granularity")
	if granularity == "" {
		if to.Sub(from) > 7*24*time.Hour {
			granularity = "daily"
		} else {
			granularity = "hourly"
		}
	}
	usage, err := h.usageSvc.GetUsageTimeSeries(r.Context(), orgID, from, to, granularity)
	if err != nil {
		response.WriteError(w, appErrors.Internal("Failed to get usage time series", err))
		return
	}
	response.Success(w, usage)
}

func (h *Handler) GetUsageByProject(w http.ResponseWriter, r *http.Request) {
	orgID := httpctx.MustGetOrganizationID(r.Context())
	q := r.URL.Query()
	from, to, err := handlersShared.ParseTimeRange(
		q.Get("from"), q.Get("to"), q.Get("time_range"),
		analytics.TimeRange30Days,
	)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	summaries, err := h.usageSvc.GetUsageByProject(r.Context(), orgID, from, to)
	if err != nil {
		response.WriteError(w, appErrors.Internal("Failed to get usage by project", err))
		return
	}
	response.Success(w, summaries)
}

func (h *Handler) ExportUsage(w http.ResponseWriter, r *http.Request) {
	orgID := httpctx.MustGetOrganizationID(r.Context())
	q := r.URL.Query()
	from, to, err := handlersShared.ParseTimeRange(
		q.Get("from"), q.Get("to"), q.Get("time_range"),
		analytics.TimeRange30Days,
	)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	format := q.Get("format")
	if format == "" {
		format = "csv"
	}
	granularity := q.Get("granularity")
	if granularity == "" {
		if to.Sub(from) > 7*24*time.Hour {
			granularity = "daily"
		} else {
			granularity = "hourly"
		}
	}

	usage, err := h.usageSvc.GetUsageTimeSeries(r.Context(), orgID, from, to, granularity)
	if err != nil {
		response.WriteError(w, appErrors.Internal("Failed to get usage for export", err))
		return
	}
	projectUsage, pErr := h.usageSvc.GetUsageByProject(r.Context(), orgID, from, to)
	if pErr != nil {
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
			response.WriteError(w, appErrors.Internal("Failed to encode usage export", mErr))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename="+filenameBase+".json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
		return
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
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename="+filenameBase+".csv")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf)
}

// ============================================================================
// Budgets
// ============================================================================

func (h *Handler) ListBudgets(w http.ResponseWriter, r *http.Request) {
	orgID := httpctx.MustGetOrganizationID(r.Context())
	budgets, err := h.budgetSvc.GetBudgetsByOrg(r.Context(), orgID)
	if err != nil {
		h.logger.Error("Failed to list budgets", "error", err, "organization_id", orgID)
		response.WriteError(w, appErrors.Internal("Failed to list budgets", err))
		return
	}
	response.Success(w, budgets)
}

func (h *Handler) GetBudget(w http.ResponseWriter, r *http.Request) {
	orgID := httpctx.MustGetOrganizationID(r.Context())
	budgetID, err := request.URLParamUUID(r, "budgetId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	budget, err := h.budgetSvc.GetBudget(r.Context(), budgetID)
	if err != nil {
		response.WriteError(w, appErrors.NotFound("budget", appErrors.WithMessage("Budget not found")))
		return
	}
	if budget.OrganizationID != orgID {
		response.WriteError(w, appErrors.PermissionDenied("budget", "Access denied to this budget"))
		return
	}
	response.Success(w, budget)
}

func (h *Handler) CreateBudget(w http.ResponseWriter, r *http.Request) {
	orgID := httpctx.MustGetOrganizationID(r.Context())

	var body createBudgetBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	if body.SpanLimit == nil && body.BytesLimit == nil && body.ScoreLimit == nil && body.CostLimit == nil {
		response.WriteError(w, appErrors.InvalidParam("limits", "at least one of span_limit, bytes_limit, score_limit, or cost_limit is required"))
		return
	}

	thresholds := body.AlertThresholds
	if thresholds == nil {
		thresholds = []int64{50, 80, 100}
	}
	sort.Slice(thresholds, func(i, j int) bool { return thresholds[i] < thresholds[j] })

	budget := &billingDomain.UsageBudget{
		OrganizationID:  orgID,
		Name:            body.Name,
		BudgetType:      billingDomain.BudgetType(body.BudgetType),
		SpanLimit:       body.SpanLimit,
		BytesLimit:      body.BytesLimit,
		ScoreLimit:      body.ScoreLimit,
		CostLimit:       float64ToDecimalPtr(body.CostLimit),
		AlertThresholds: thresholds,
	}

	if body.ProjectID != nil {
		projectID, pErr := uuid.Parse(*body.ProjectID)
		if pErr != nil {
			response.WriteError(w, appErrors.InvalidParam("project_id", "must be a valid UUID"))
			return
		}
		budget.ProjectID = &projectID
	}

	if err := h.budgetSvc.CreateBudget(r.Context(), budget); err != nil {
		h.logger.Error("Failed to create budget", "error", err, "organization_id", orgID)
		response.WriteError(w, err)
		return
	}
	response.Created(w, budget)
}

func (h *Handler) UpdateBudget(w http.ResponseWriter, r *http.Request) {
	orgID := httpctx.MustGetOrganizationID(r.Context())
	budgetID, err := request.URLParamUUID(r, "budgetId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	budget, err := h.budgetSvc.GetBudget(r.Context(), budgetID)
	if err != nil {
		response.WriteError(w, appErrors.NotFound("budget", appErrors.WithMessage("Budget not found")))
		return
	}
	if budget.OrganizationID != orgID {
		response.WriteError(w, appErrors.PermissionDenied("budget", "Access denied to this budget"))
		return
	}

	var body updateBudgetBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	if body.Name != nil {
		budget.Name = *body.Name
	}
	if body.SpanLimit != nil {
		budget.SpanLimit = body.SpanLimit
	}
	if body.BytesLimit != nil {
		budget.BytesLimit = body.BytesLimit
	}
	if body.ScoreLimit != nil {
		budget.ScoreLimit = body.ScoreLimit
	}
	if body.CostLimit != nil {
		budget.CostLimit = float64ToDecimalPtr(body.CostLimit)
	}
	if body.AlertThresholds != nil {
		sort.Slice(body.AlertThresholds, func(i, j int) bool {
			return body.AlertThresholds[i] < body.AlertThresholds[j]
		})
		budget.AlertThresholds = body.AlertThresholds
	}
	if body.IsActive != nil {
		budget.IsActive = *body.IsActive
	}

	if err := h.budgetSvc.UpdateBudget(r.Context(), budget); err != nil {
		h.logger.Error("Failed to update budget", "error", err, "budget_id", budgetID)
		response.WriteError(w, appErrors.Internal("Failed to update budget", err))
		return
	}
	response.Success(w, budget)
}

func (h *Handler) DeleteBudget(w http.ResponseWriter, r *http.Request) {
	orgID := httpctx.MustGetOrganizationID(r.Context())
	budgetID, err := request.URLParamUUID(r, "budgetId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	budget, err := h.budgetSvc.GetBudget(r.Context(), budgetID)
	if err != nil {
		response.WriteError(w, appErrors.NotFound("budget", appErrors.WithMessage("Budget not found")))
		return
	}
	if budget.OrganizationID != orgID {
		response.WriteError(w, appErrors.PermissionDenied("budget", "Access denied to this budget"))
		return
	}
	if err := h.budgetSvc.DeleteBudget(r.Context(), budgetID); err != nil {
		h.logger.Error("Failed to delete budget", "error", err, "budget_id", budgetID)
		response.WriteError(w, appErrors.Internal("Failed to delete budget", err))
		return
	}
	response.NoContent(w)
}

func (h *Handler) GetBudgetAlerts(w http.ResponseWriter, r *http.Request) {
	orgID := httpctx.MustGetOrganizationID(r.Context())
	limit, err := request.QueryInt(r, "limit", 50)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	if limit <= 0 {
		limit = 50
	}
	alerts, err := h.budgetSvc.GetAlerts(r.Context(), orgID, limit)
	if err != nil {
		h.logger.Error("Failed to get alerts", "error", err, "organization_id", orgID)
		response.WriteError(w, appErrors.Internal("Failed to get alerts", err))
		return
	}
	response.Success(w, alerts)
}

func (h *Handler) AcknowledgeBudgetAlert(w http.ResponseWriter, r *http.Request) {
	orgID := httpctx.MustGetOrganizationID(r.Context())
	alertID, err := request.URLParamUUID(r, "alertId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	if err := h.budgetSvc.AcknowledgeAlert(r.Context(), orgID, alertID); err != nil {
		h.logger.Error("Failed to acknowledge alert", "error", err, "alert_id", alertID)
		if appErrors.IsNotFound(err) {
			response.WriteError(w, appErrors.NotFound("alert", appErrors.WithMessage("Alert not found")))
			return
		}
		response.WriteError(w, appErrors.Internal("Failed to acknowledge alert", err))
		return
	}
	response.Success(w, map[string]any{"acknowledged": true})
}

// ============================================================================
// Contracts
// ============================================================================

func (h *Handler) CreateContract(w http.ResponseWriter, r *http.Request) {
	var body createContractBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	orgID, err := uuid.Parse(body.OrganizationID)
	if err != nil {
		response.WriteError(w, appErrors.InvalidParam("organization_id", "must be a valid UUID"))
		return
	}
	userID := httpctx.MustGetUserID(r.Context())

	contract := &billingDomain.Contract{
		ID:                      uid.New(),
		OrganizationID:          orgID,
		ContractName:            body.ContractName,
		ContractNumber:          body.ContractNumber,
		StartDate:               body.StartsAt,
		EndDate:                 body.ExpiresAt,
		MinimumCommitAmount:     float64ToDecimalPtr(body.MinimumCommitAmount),
		Currency:                "USD",
		AccountOwner:            shared.NilIfEmpty(body.AccountOwner),
		SalesRepEmail:           shared.NilIfEmpty(body.SalesRepEmail),
		Status:                  billingDomain.ContractStatusDraft,
		CustomFreeSpans:         body.CustomFreeSpans,
		CustomPricePer100KSpans: float64ToDecimalPtr(body.CustomPricePer100KSpans),
		CustomFreeGB:            float64ToDecimalPtr(body.CustomFreeGB),
		CustomPricePerGB:        float64ToDecimalPtr(body.CustomPricePerGB),
		CustomFreeScores:        body.CustomFreeScores,
		CustomPricePer1KScores:  float64ToDecimalPtr(body.CustomPricePer1KScores),
		CreatedBy:               userID.String(),
		CreatedAt:               time.Now(),
		UpdatedAt:               time.Now(),
		Notes:                   shared.NilIfEmpty(body.Notes),
	}
	if body.Currency != "" {
		contract.Currency = body.Currency
	}

	if err := h.contractSvc.CreateContract(r.Context(), contract); err != nil {
		h.logger.Error("Failed to create contract", "error", err, "organization_id", orgID)
		response.WriteError(w, err)
		return
	}

	if len(body.VolumeTiers) > 0 {
		tiers := make([]*billingDomain.VolumeDiscountTier, len(body.VolumeTiers))
		for i, tr := range body.VolumeTiers {
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
		if err := h.contractSvc.AddVolumeTiers(r.Context(), contract.ID, tiers); err != nil {
			h.logger.Error("Failed to add volume tiers", "error", err, "contract_id", contract.ID)
			response.WriteError(w, err)
			return
		}
	}

	result, err := h.contractSvc.GetContract(r.Context(), contract.ID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Created(w, result)
}

func (h *Handler) GetContract(w http.ResponseWriter, r *http.Request) {
	contractID, err := request.URLParamUUID(r, "contractId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	contract, err := h.contractSvc.GetContract(r.Context(), contractID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, contract)
}

func (h *Handler) ListContracts(w http.ResponseWriter, r *http.Request) {
	orgID, err := request.URLParamUUID(r, "orgId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	contracts, err := h.contractSvc.GetContractsByOrg(r.Context(), orgID)
	if err != nil {
		h.logger.Error("Failed to get contracts", "error", err, "organization_id", orgID)
		response.WriteError(w, appErrors.Internal("Failed to get contracts", err))
		return
	}
	response.Success(w, contracts)
}

func (h *Handler) UpdateContract(w http.ResponseWriter, r *http.Request) {
	contractID, err := request.URLParamUUID(r, "contractId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	contract, err := h.contractSvc.GetContract(r.Context(), contractID)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	var b updateContractBody
	if err := request.DecodeJSON(r, &b); err != nil {
		response.WriteError(w, err)
		return
	}

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

	if err := h.contractSvc.UpdateContract(r.Context(), contract); err != nil {
		h.logger.Error("Failed to update contract", "error", err, "contract_id", contractID)
		response.WriteError(w, err)
		return
	}
	response.Success(w, contract)
}

func (h *Handler) ActivateContract(w http.ResponseWriter, r *http.Request) {
	contractID, err := request.URLParamUUID(r, "contractId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	userID := httpctx.MustGetUserID(r.Context())

	if err := h.contractSvc.ActivateContract(r.Context(), contractID, userID); err != nil {
		h.logger.Error("Failed to activate contract", "error", err, "contract_id", contractID)
		response.WriteError(w, err)
		return
	}
	response.Success(w, map[string]any{
		"message": "Contract activated successfully",
		"status":  "active",
	})
}

func (h *Handler) CancelContract(w http.ResponseWriter, r *http.Request) {
	contractID, err := request.URLParamUUID(r, "contractId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	userID := httpctx.MustGetUserID(r.Context())

	var body cancelContractBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	if err := h.contractSvc.CancelContract(r.Context(), contractID, body.Reason, userID); err != nil {
		h.logger.Error("Failed to cancel contract", "error", err, "contract_id", contractID)
		response.WriteError(w, err)
		return
	}
	response.NoContent(w)
}

func (h *Handler) UpdateContractTiers(w http.ResponseWriter, r *http.Request) {
	contractID, err := request.URLParamUUID(r, "contractId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	var body updateContractTiersBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	tiers := make([]*billingDomain.VolumeDiscountTier, len(body.Tiers))
	for i, tr := range body.Tiers {
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
	if err := h.contractSvc.UpdateVolumeTiers(r.Context(), contractID, tiers); err != nil {
		h.logger.Error("Failed to update volume tiers", "error", err, "contract_id", contractID)
		response.WriteError(w, err)
		return
	}
	response.Success(w, map[string]any{
		"message":     "Volume tiers updated successfully",
		"tiers_count": len(tiers),
	})
}

func (h *Handler) GetContractHistory(w http.ResponseWriter, r *http.Request) {
	contractID, err := request.URLParamUUID(r, "contractId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	history, err := h.contractSvc.GetContractHistory(r.Context(), contractID)
	if err != nil {
		h.logger.Error("Failed to get contract history", "error", err, "contract_id", contractID)
		response.WriteError(w, appErrors.Internal("Failed to get contract history", err))
		return
	}
	response.Success(w, history)
}

func (h *Handler) GetEffectivePricing(w http.ResponseWriter, r *http.Request) {
	orgID, err := request.URLParamUUID(r, "orgId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	effective, err := h.pricingSvc.GetEffectivePricing(r.Context(), orgID)
	if err != nil {
		h.logger.Error("Failed to get effective pricing", "error", err, "organization_id", orgID)
		response.WriteError(w, appErrors.Internal("Failed to get effective pricing", err))
		return
	}
	response.Success(w, effective)
}
