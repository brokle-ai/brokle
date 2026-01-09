package workers

import (
	"context"
	"log/slog"
	"sort"
	"strconv"
	"time"

	"brokle/internal/config"
	"brokle/internal/core/domain/billing"
	"brokle/internal/core/domain/organization"
	"brokle/pkg/ulid"
)

// UsageAggregationWorker syncs ClickHouse usage data to PostgreSQL billing state
// and checks budget thresholds to trigger alerts
type UsageAggregationWorker struct {
	config             *config.Config
	logger             *slog.Logger
	usageRepo          billing.BillableUsageRepository
	billingRepo        billing.OrganizationBillingRepository
	planRepo           billing.PlanRepository
	contractRepo       billing.ContractRepository
	tierRepo           billing.VolumeDiscountTierRepository
	budgetRepo         billing.UsageBudgetRepository
	alertRepo          billing.UsageAlertRepository
	orgRepo            organization.OrganizationRepository
	pricingService     billing.PricingService
	notificationWorker *NotificationWorker
	quit               chan bool
	ticker             *time.Ticker
}

// NewUsageAggregationWorker creates a new usage aggregation worker
func NewUsageAggregationWorker(
	config *config.Config,
	logger *slog.Logger,
	usageRepo billing.BillableUsageRepository,
	billingRepo billing.OrganizationBillingRepository,
	planRepo billing.PlanRepository,
	contractRepo billing.ContractRepository,
	tierRepo billing.VolumeDiscountTierRepository,
	budgetRepo billing.UsageBudgetRepository,
	alertRepo billing.UsageAlertRepository,
	orgRepo organization.OrganizationRepository,
	pricingService billing.PricingService,
	notificationWorker *NotificationWorker,
) *UsageAggregationWorker {
	return &UsageAggregationWorker{
		config:             config,
		logger:             logger,
		usageRepo:          usageRepo,
		billingRepo:        billingRepo,
		planRepo:           planRepo,
		contractRepo:       contractRepo,
		tierRepo:           tierRepo,
		budgetRepo:         budgetRepo,
		alertRepo:          alertRepo,
		orgRepo:            orgRepo,
		pricingService:     pricingService,
		notificationWorker: notificationWorker,
		quit:               make(chan bool),
	}
}

// Start starts the usage aggregation worker
func (w *UsageAggregationWorker) Start() {
	w.logger.Info("Starting usage aggregation worker")

	// Get sync interval from config (default 5 minutes)
	interval := 5 * time.Minute
	if w.config.Workers.UsageSyncIntervalMinutes > 0 {
		interval = time.Duration(w.config.Workers.UsageSyncIntervalMinutes) * time.Minute
	}

	w.ticker = time.NewTicker(interval)

	// Run immediately on start
	go w.run()

	// Then run on ticker
	go func() {
		for {
			select {
			case <-w.ticker.C:
				w.run()
			case <-w.quit:
				w.ticker.Stop()
				w.logger.Info("Usage aggregation worker stopped")
				return
			}
		}
	}()
}

// Stop stops the usage aggregation worker
func (w *UsageAggregationWorker) Stop() {
	w.logger.Info("Stopping usage aggregation worker")
	close(w.quit)
}

// run executes a single aggregation cycle
func (w *UsageAggregationWorker) run() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	w.logger.Debug("Starting usage aggregation cycle")
	startTime := time.Now()

	// Get all organizations
	orgs, err := w.orgRepo.List(ctx, nil)
	if err != nil {
		w.logger.Error("failed to list organizations", "error", err)
		return
	}

	var syncedCount, alertCount int

	for _, org := range orgs {
		// Sync billing state for each organization
		if err := w.syncOrganizationUsage(ctx, org.ID); err != nil {
			w.logger.Error("failed to sync organization usage",
				"error", err,
				"organization_id", org.ID,
			)
			continue
		}
		syncedCount++

		// Check budgets and trigger alerts
		alerts, err := w.checkBudgets(ctx, org.ID)
		if err != nil {
			w.logger.Error("failed to check budgets",
				"error", err,
				"organization_id", org.ID,
			)
			continue
		}
		alertCount += len(alerts)

		// Send notifications for new alerts
		for _, alert := range alerts {
			w.sendAlertNotification(ctx, org, alert)
		}
	}

	duration := time.Since(startTime)
	w.logger.Info("Usage aggregation cycle completed",
		"organizations_synced", syncedCount,
		"alerts_triggered", alertCount,
		"duration_ms", duration.Milliseconds(),
	)
}

// syncOrganizationUsage syncs ClickHouse usage to PostgreSQL billing state
func (w *UsageAggregationWorker) syncOrganizationUsage(ctx context.Context, orgID ulid.ULID) error {
	// Get current billing state
	orgBilling, err := w.billingRepo.GetByOrgID(ctx, orgID)
	if err != nil {
		// Organization might not have billing set up yet
		w.logger.Debug("no billing record for organization", "organization_id", orgID)
		return nil
	}

	// Get effective pricing (plan + contract overrides)
	effectivePricing, err := w.getEffectivePricing(ctx, orgID, orgBilling)
	if err != nil {
		return err
	}

	// Check if we need to reset the billing period
	if time.Now().After(w.calculatePeriodEnd(orgBilling.BillingCycleStart, orgBilling.BillingCycleAnchorDay)) {
		if err := w.resetBillingPeriod(ctx, orgID, orgBilling); err != nil {
			return err
		}
		// Refresh billing state after reset
		orgBilling, err = w.billingRepo.GetByOrgID(ctx, orgID)
		if err != nil {
			return err
		}
	}

	// Query current period usage from ClickHouse
	filter := &billing.BillableUsageFilter{
		OrganizationID: orgID,
		Start:          orgBilling.BillingCycleStart,
		End:            time.Now(),
		Granularity:    "hourly",
	}

	summary, err := w.usageRepo.GetUsageSummary(ctx, filter)
	if err != nil {
		return err
	}

	// Calculate cost (tier-aware)
	cost := w.calculateCost(summary, effectivePricing)

	// Calculate free tier remaining
	freeSpansRemaining := max(0, effectivePricing.FreeSpans-summary.TotalSpans)
	freeBytesRemaining := max(0, int64(effectivePricing.FreeGB*1073741824)-summary.TotalBytes)
	freeScoresRemaining := max(0, effectivePricing.FreeScores-summary.TotalScores)

	// Update billing state
	orgBilling.CurrentPeriodSpans = summary.TotalSpans
	orgBilling.CurrentPeriodBytes = summary.TotalBytes
	orgBilling.CurrentPeriodScores = summary.TotalScores
	orgBilling.CurrentPeriodCost = cost
	orgBilling.FreeSpansRemaining = freeSpansRemaining
	orgBilling.FreeBytesRemaining = freeBytesRemaining
	orgBilling.FreeScoresRemaining = freeScoresRemaining
	orgBilling.LastSyncedAt = time.Now()
	orgBilling.UpdatedAt = time.Now()

	if err := w.billingRepo.Update(ctx, orgBilling); err != nil {
		return err
	}

	// Also update budget usage
	if err := w.syncBudgetUsage(ctx, orgID, summary, cost, effectivePricing); err != nil {
		w.logger.Warn("failed to sync budget usage",
			"error", err,
			"organization_id", orgID,
		)
	}

	return nil
}

// resetBillingPeriod resets the billing period for an organization
func (w *UsageAggregationWorker) resetBillingPeriod(ctx context.Context, orgID ulid.ULID, current *billing.OrganizationBilling) error {
	w.logger.Info("Resetting billing period",
		"organization_id", orgID,
		"old_cycle_start", current.BillingCycleStart,
	)

	// Calculate new cycle start
	newCycleStart := w.calculatePeriodEnd(current.BillingCycleStart, current.BillingCycleAnchorDay)

	return w.billingRepo.ResetPeriod(ctx, orgID, newCycleStart)
}

// syncBudgetUsage syncs usage to all budgets for an organization
func (w *UsageAggregationWorker) syncBudgetUsage(ctx context.Context, orgID ulid.ULID, summary *billing.BillableUsageSummary, cost float64, effectivePricing *billing.EffectivePricing) error {
	budgets, err := w.budgetRepo.GetActive(ctx, orgID)
	if err != nil {
		return err
	}

	for _, budget := range budgets {
		var spans, bytes, scores int64
		var budgetCost float64

		if budget.ProjectID != nil {
			// Project-level budget - query project-specific usage
			filter := &billing.BillableUsageFilter{
				OrganizationID: orgID,
				ProjectID:      budget.ProjectID,
				Start:          w.getBudgetPeriodStart(budget),
				End:            time.Now(),
				Granularity:    "hourly",
			}
			projectSummary, err := w.usageRepo.GetUsageSummary(ctx, filter)
			if err != nil {
				w.logger.Warn("failed to get project usage",
					"error", err,
					"project_id", budget.ProjectID,
				)
				continue
			}
			spans = projectSummary.TotalSpans
			bytes = projectSummary.TotalBytes
			scores = projectSummary.TotalScores
			budgetCost = w.calculateRawCost(projectSummary, effectivePricing)
		} else {
			// Org-level budget - check if budget period differs from billing cycle
			budgetStart := w.getBudgetPeriodStart(budget)
			if !budgetStart.Equal(summary.PeriodStart) {
				// Budget period differs from billing cycle (weekly budget, or mid-month billing start)
				// Use marginal cost: cost(cycle_start→now) - cost(cycle_start→budget_start)
				// This properly accounts for free tier across the billing cycle

				// Clamp usage window start to billing cycle start
				// When budget starts before billing cycle, we only count usage within the billing cycle
				usageWindowStart := budgetStart
				if budgetStart.Before(summary.PeriodStart) {
					usageWindowStart = summary.PeriodStart
				}

				// Query usage for the budget window (usageWindowStart → now)
				budgetFilter := &billing.BillableUsageFilter{
					OrganizationID: orgID,
					Start:          usageWindowStart,
					End:            time.Now(),
					Granularity:    "hourly",
				}
				orgPeriodSummary, err := w.usageRepo.GetUsageSummary(ctx, budgetFilter)
				if err != nil {
					w.logger.Warn("failed to get org period usage for budget",
						"error", err,
						"budget_id", budget.ID,
						"budget_type", budget.BudgetType,
					)
					continue
				}
				spans = orgPeriodSummary.TotalSpans
				bytes = orgPeriodSummary.TotalBytes
				scores = orgPeriodSummary.TotalScores

				// Calculate marginal cost for budget window
				// Marginal cost = total_cycle_cost - cost_before_budget_window
				if budgetStart.Before(summary.PeriodStart) || budgetStart.Equal(summary.PeriodStart) {
					// Budget window starts at or before billing cycle - no pre-budget period
					// Full billing cycle cost applies to this budget window
					budgetCost = cost
				} else {
					// Query usage from billing cycle start to budget window start
					preBudgetFilter := &billing.BillableUsageFilter{
						OrganizationID: orgID,
						Start:          summary.PeriodStart,
						End:            budgetStart,
						Granularity:    "hourly",
					}
					preBudgetSummary, err := w.usageRepo.GetUsageSummary(ctx, preBudgetFilter)
					if err != nil {
						w.logger.Warn("failed to get pre-budget usage",
							"error", err,
							"budget_id", budget.ID,
						)
						// Fall back to raw cost if we can't calculate marginal
						budgetCost = w.calculateRawCost(orgPeriodSummary, effectivePricing)
					} else {
						costBeforeBudget := w.calculateCost(preBudgetSummary, effectivePricing)
						budgetCost = max(0, cost-costBeforeBudget)
					}
				}
			} else {
				// Budget period matches billing cycle - use pre-calculated values
				spans = summary.TotalSpans
				bytes = summary.TotalBytes
				scores = summary.TotalScores
				budgetCost = cost
			}
		}

		if err := w.budgetRepo.UpdateUsage(ctx, budget.ID, spans, bytes, scores, budgetCost); err != nil {
			w.logger.Warn("failed to update budget usage",
				"error", err,
				"budget_id", budget.ID,
			)
		}
	}

	return nil
}

// getEffectivePricing resolves pricing: contract overrides > plan defaults
func (w *UsageAggregationWorker) getEffectivePricing(ctx context.Context, orgID ulid.ULID, orgBilling *billing.OrganizationBilling) (*billing.EffectivePricing, error) {
	// 1. Get organization's base plan
	plan, err := w.planRepo.GetByID(ctx, orgBilling.PlanID)
	if err != nil {
		return nil, err
	}

	// 2. Check for active contract
	contract, err := w.contractRepo.GetActiveByOrgID(ctx, orgID)
	if err != nil && err.Error() != "no active contract found for organization: "+orgID.String() {
		return nil, err
	}

	effective := &billing.EffectivePricing{
		OrganizationID: orgID,
		BasePlan:       plan,
		Contract:       contract,
	}

	// 3. Resolve pricing (contract overrides plan)
	if contract != nil {
		effective.FreeSpans = coalesceInt64(contract.CustomFreeSpans, plan.FreeSpans)
		effective.PricePer100KSpans = coalesceFloat64Ptr(contract.CustomPricePer100KSpans, plan.PricePer100KSpans)
		effective.FreeGB = coalesceFloat64Ptr(contract.CustomFreeGB, &plan.FreeGB)
		effective.PricePerGB = coalesceFloat64Ptr(contract.CustomPricePerGB, plan.PricePerGB)
		effective.FreeScores = coalesceInt64(contract.CustomFreeScores, plan.FreeScores)
		effective.PricePer1KScores = coalesceFloat64Ptr(contract.CustomPricePer1KScores, plan.PricePer1KScores)

		// Load volume tiers
		tiers, err := w.tierRepo.GetByContractID(ctx, contract.ID)
		if err != nil {
			return nil, err
		}

		if len(tiers) > 0 {
			effective.HasVolumeTiers = true
			effective.VolumeTiers = tiers
		}
	} else {
		// No contract, use plan defaults
		effective.FreeSpans = plan.FreeSpans
		effective.PricePer100KSpans = derefFloat64(plan.PricePer100KSpans)
		effective.FreeGB = plan.FreeGB
		effective.PricePerGB = derefFloat64(plan.PricePerGB)
		effective.FreeScores = plan.FreeScores
		effective.PricePer1KScores = derefFloat64(plan.PricePer1KScores)
	}

	return effective, nil
}

// checkBudgets checks all budgets and returns any new alerts
func (w *UsageAggregationWorker) checkBudgets(ctx context.Context, orgID ulid.ULID) ([]*billing.UsageAlert, error) {
	budgets, err := w.budgetRepo.GetActive(ctx, orgID)
	if err != nil {
		return nil, err
	}

	var newAlerts []*billing.UsageAlert

	for _, budget := range budgets {
		alerts := w.evaluateBudget(budget)
		for _, alert := range alerts {
			// Check if we already have a recent alert for this budget/threshold/dimension
			if w.hasRecentAlert(ctx, budget.ID, alert.AlertThreshold, alert.Dimension) {
				continue
			}

			if err := w.alertRepo.Create(ctx, alert); err != nil {
				w.logger.Error("failed to create alert",
					"error", err,
					"budget_id", budget.ID,
				)
				continue
			}
			newAlerts = append(newAlerts, alert)

			w.logger.Warn("budget alert triggered",
				"alert_id", alert.ID,
				"budget_id", budget.ID,
				"budget_name", budget.Name,
				"alert_threshold", alert.AlertThreshold,
				"dimension", alert.Dimension,
				"percent_used", alert.PercentUsed,
			)
		}
	}

	return newAlerts, nil
}

// evaluateBudget checks a single budget and returns any triggered alerts
func (w *UsageAggregationWorker) evaluateBudget(budget *billing.UsageBudget) []*billing.UsageAlert {
	var alerts []*billing.UsageAlert

	// Ensure thresholds are sorted ascending for correct reverse iteration
	sort.Slice(budget.AlertThresholds, func(i, j int) bool {
		return budget.AlertThresholds[i] < budget.AlertThresholds[j]
	})

	// Check each dimension
	dimensions := []struct {
		dimension    billing.AlertDimension
		current      int64
		limit        *int64
		currentFloat float64
		limitFloat   *float64
	}{
		{billing.AlertDimensionSpans, budget.CurrentSpans, budget.SpanLimit, 0, nil},
		{billing.AlertDimensionBytes, budget.CurrentBytes, budget.BytesLimit, 0, nil},
		{billing.AlertDimensionScores, budget.CurrentScores, budget.ScoreLimit, 0, nil},
		{billing.AlertDimensionCost, 0, nil, budget.CurrentCost, budget.CostLimit},
	}

	for _, dim := range dimensions {
		var percentUsed float64
		var actualValue int64
		var thresholdValue int64

		if dim.dimension == billing.AlertDimensionCost {
			if dim.limitFloat == nil || *dim.limitFloat == 0 {
				continue
			}
			percentUsed = (dim.currentFloat / *dim.limitFloat) * 100
			actualValue = int64(dim.currentFloat * 100) // Store as cents
			thresholdValue = int64(*dim.limitFloat * 100)
		} else {
			if dim.limit == nil || *dim.limit == 0 {
				continue
			}
			percentUsed = (float64(dim.current) / float64(*dim.limit)) * 100
			actualValue = dim.current
			thresholdValue = *dim.limit
		}

		// Iterate over flexible thresholds (sorted descending to trigger highest first)
		for i := len(budget.AlertThresholds) - 1; i >= 0; i-- {
			threshold := budget.AlertThresholds[i]
			if percentUsed >= float64(threshold) {
				alert := &billing.UsageAlert{
					ID:             ulid.New(),
					BudgetID:       &budget.ID,
					OrganizationID: budget.OrganizationID,
					ProjectID:      budget.ProjectID,
					AlertThreshold: threshold,
					Dimension:      dim.dimension,
					Severity:       getSeverityForThreshold(threshold),
					ThresholdValue: thresholdValue,
					ActualValue:    actualValue,
					PercentUsed:    percentUsed,
					Status:         billing.AlertStatusTriggered,
					TriggeredAt:    time.Now(),
				}
				alerts = append(alerts, alert)
				break // Only trigger the highest threshold per dimension
			}
		}
	}

	return alerts
}

// getSeverityForThreshold returns the appropriate severity based on threshold percentage
func getSeverityForThreshold(threshold int64) billing.AlertSeverity {
	switch {
	case threshold >= 100:
		return billing.AlertSeverityCritical
	case threshold >= 80:
		return billing.AlertSeverityWarning
	default:
		return billing.AlertSeverityInfo
	}
}

// hasRecentAlert checks if there's a recent unresolved alert for the same budget/threshold/dimension
func (w *UsageAggregationWorker) hasRecentAlert(ctx context.Context, budgetID ulid.ULID, alertThreshold int64, dimension billing.AlertDimension) bool {
	alerts, err := w.alertRepo.GetByBudgetID(ctx, budgetID)
	if err != nil {
		return false
	}

	// Check for recent alerts (within last 24 hours) that match
	cutoff := time.Now().Add(-24 * time.Hour)
	for _, alert := range alerts {
		if alert.AlertThreshold == alertThreshold &&
			alert.Dimension == dimension &&
			alert.TriggeredAt.After(cutoff) &&
			alert.Status != billing.AlertStatusResolved {
			return true
		}
	}

	return false
}

// sendAlertNotification sends notification for a new alert
func (w *UsageAggregationWorker) sendAlertNotification(ctx context.Context, org *organization.Organization, alert *billing.UsageAlert) {
	if w.notificationWorker == nil {
		return
	}

	// Get budget name for context
	budgetName := "Organization"
	if alert.BudgetID != nil {
		budget, err := w.budgetRepo.GetByID(ctx, *alert.BudgetID)
		if err == nil {
			budgetName = budget.Name
		}
	}

	// Format the dimension value
	var valueStr string
	switch alert.Dimension {
	case billing.AlertDimensionSpans:
		valueStr = formatNumber(alert.ActualValue)
	case billing.AlertDimensionBytes:
		valueStr = formatBytes(alert.ActualValue)
	case billing.AlertDimensionScores:
		valueStr = formatNumber(alert.ActualValue)
	case billing.AlertDimensionCost:
		valueStr = formatCurrency(float64(alert.ActualValue) / 100)
	}

	// Send email notification
	if org.BillingEmail != "" {
		w.notificationWorker.QueueEmail(EmailJob{
			To:       []string{org.BillingEmail},
			Subject:  "Usage Alert: " + string(alert.Dimension) + " threshold exceeded",
			Template: "usage_alert",
			TemplateData: map[string]interface{}{
				"organization_name": org.Name,
				"budget_name":       budgetName,
				"dimension":         string(alert.Dimension),
				"percent_used":      alert.PercentUsed,
				"current_value":     valueStr,
				"severity":          string(alert.Severity),
			},
			Priority: "high",
		})
	}

	// Mark notification as sent
	if err := w.alertRepo.MarkNotificationSent(ctx, alert.ID); err != nil {
		w.logger.Warn("failed to mark notification sent",
			"error", err,
			"alert_id", alert.ID,
		)
	}
}

// calculateCost computes total cost from three billable dimensions with tier support
func (w *UsageAggregationWorker) calculateCost(usage *billing.BillableUsageSummary, pricing *billing.EffectivePricing) float64 {
	if pricing.HasVolumeTiers {
		return w.calculateWithTiers(usage, pricing)
	}
	return w.calculateFlat(usage, pricing)
}

// calculateFlat uses simple linear pricing
func (w *UsageAggregationWorker) calculateFlat(usage *billing.BillableUsageSummary, pricing *billing.EffectivePricing) float64 {
	var totalCost float64

	// Spans
	billableSpans := max(0, usage.TotalSpans-pricing.FreeSpans)
	spanCost := float64(billableSpans) / 100000.0 * pricing.PricePer100KSpans
	totalCost += spanCost

	// Bytes
	freeBytes := int64(pricing.FreeGB * 1073741824)
	billableBytes := max(0, usage.TotalBytes-freeBytes)
	billableGB := float64(billableBytes) / 1073741824.0
	dataCost := billableGB * pricing.PricePerGB
	totalCost += dataCost

	// Scores
	billableScores := max(0, usage.TotalScores-pricing.FreeScores)
	scoreCost := float64(billableScores) / 1000.0 * pricing.PricePer1KScores
	totalCost += scoreCost

	return totalCost
}

// calculateWithTiers uses progressive tier pricing
// Delegates to PricingService for correct tier calculation logic
func (w *UsageAggregationWorker) calculateWithTiers(usage *billing.BillableUsageSummary, pricing *billing.EffectivePricing) float64 {
	totalCost := 0.0

	// Delegate to PricingService for tier calculations
	totalCost += w.pricingService.CalculateDimensionWithTiers(usage.TotalSpans, pricing.FreeSpans, billing.TierDimensionSpans, pricing.VolumeTiers, pricing)

	freeBytes := int64(pricing.FreeGB * 1073741824)
	totalCost += w.pricingService.CalculateDimensionWithTiers(usage.TotalBytes, freeBytes, billing.TierDimensionBytes, pricing.VolumeTiers, pricing)

	totalCost += w.pricingService.CalculateDimensionWithTiers(usage.TotalScores, pricing.FreeScores, billing.TierDimensionScores, pricing.VolumeTiers, pricing)

	return totalCost
}

// calculateRawCost computes cost for usage without applying free tier deductions.
// Used for project-level budgets where free tier is already accounted at org level.
func (w *UsageAggregationWorker) calculateRawCost(usage *billing.BillableUsageSummary, pricing *billing.EffectivePricing) float64 {
	if pricing.HasVolumeTiers {
		return w.calculateWithTiersNoFreeTier(usage, pricing)
	}
	return w.calculateFlatNoFreeTier(usage, pricing)
}

// calculateFlatNoFreeTier uses simple linear pricing without free tier deductions
func (w *UsageAggregationWorker) calculateFlatNoFreeTier(usage *billing.BillableUsageSummary, pricing *billing.EffectivePricing) float64 {
	var totalCost float64

	// Spans
	spanCost := float64(usage.TotalSpans) / 100000.0 * pricing.PricePer100KSpans
	totalCost += spanCost

	// Bytes
	billableGB := float64(usage.TotalBytes) / 1073741824.0
	dataCost := billableGB * pricing.PricePerGB
	totalCost += dataCost

	// Scores
	scoreCost := float64(usage.TotalScores) / 1000.0 * pricing.PricePer1KScores
	totalCost += scoreCost

	return totalCost
}

// calculateWithTiersNoFreeTier uses progressive tier pricing without free tier deductions
// Delegates to PricingService for correct tier calculation logic
func (w *UsageAggregationWorker) calculateWithTiersNoFreeTier(usage *billing.BillableUsageSummary, pricing *billing.EffectivePricing) float64 {
	totalCost := 0.0

	// Delegate to PricingService for tier calculations without free tier
	totalCost += w.pricingService.CalculateDimensionWithTiers(usage.TotalSpans, 0, billing.TierDimensionSpans, pricing.VolumeTiers, pricing)
	totalCost += w.pricingService.CalculateDimensionWithTiers(usage.TotalBytes, 0, billing.TierDimensionBytes, pricing.VolumeTiers, pricing)
	totalCost += w.pricingService.CalculateDimensionWithTiers(usage.TotalScores, 0, billing.TierDimensionScores, pricing.VolumeTiers, pricing)

	return totalCost
}

// NOTE: Removed duplicate calculateDimensionWithTiers, calculateFlatDimension, and getDimensionUnitSize.
// Worker now delegates to PricingService.CalculateDimensionWithTiers for all tier calculations.
// This ensures single source of truth for billing logic and prevents bugs from duplicate code.

// calculatePeriodEnd calculates the end of the current billing period
func (w *UsageAggregationWorker) calculatePeriodEnd(cycleStart time.Time, anchorDay int) time.Time {
	nextMonth := cycleStart.AddDate(0, 1, 0)
	year, month, _ := nextMonth.Date()
	loc := nextMonth.Location()

	lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, loc).Day()
	day := anchorDay
	if day > lastDay {
		day = lastDay
	}

	return time.Date(year, month, day, 0, 0, 0, 0, loc)
}

// getBudgetPeriodStart returns the start of the current budget period
func (w *UsageAggregationWorker) getBudgetPeriodStart(budget *billing.UsageBudget) time.Time {
	now := time.Now()
	switch budget.BudgetType {
	case billing.BudgetTypeWeekly:
		// Start of current week (Monday)
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		return time.Date(now.Year(), now.Month(), now.Day()-weekday+1, 0, 0, 0, 0, now.Location())
	case billing.BudgetTypeMonthly:
		// Start of current month
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	default:
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	}
}

// Helper functions for pricing resolution
func coalesceInt64(custom *int64, defaultVal int64) int64 {
	if custom != nil {
		return *custom
	}
	return defaultVal
}

func coalesceFloat64Ptr(custom *float64, defaultVal *float64) float64 {
	if custom != nil {
		return *custom
	}
	if defaultVal != nil {
		return *defaultVal
	}
	return 0
}

func derefFloat64(ptr *float64) float64 {
	if ptr != nil {
		return *ptr
	}
	return 0
}

// Helper formatting functions
func formatNumber(n int64) string {
	if n >= 1000000 {
		return formatFloat(float64(n)/1000000) + "M"
	}
	if n >= 1000 {
		return formatFloat(float64(n)/1000) + "K"
	}
	return formatInt(n)
}

func formatBytes(b int64) string {
	if b >= 1073741824 {
		return formatFloat(float64(b)/1073741824) + " GB"
	}
	if b >= 1048576 {
		return formatFloat(float64(b)/1048576) + " MB"
	}
	if b >= 1024 {
		return formatFloat(float64(b)/1024) + " KB"
	}
	return formatInt(b) + " B"
}

func formatCurrency(f float64) string {
	return "$" + formatFloat(f)
}

func formatFloat(f float64) string {
	if f == float64(int64(f)) {
		return formatInt(int64(f))
	}
	return strconv.FormatFloat(f, 'f', 2, 64)
}

func formatInt(n int64) string {
	if n == 0 {
		return "0"
	}
	var result []byte
	for n > 0 {
		result = append([]byte{byte('0' + n%10)}, result...)
		n /= 10
	}
	return string(result)
}
