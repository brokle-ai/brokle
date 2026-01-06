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
	pricingRepo        billing.PricingConfigRepository
	budgetRepo         billing.UsageBudgetRepository
	alertRepo          billing.UsageAlertRepository
	orgRepo            organization.OrganizationRepository
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
	pricingRepo billing.PricingConfigRepository,
	budgetRepo billing.UsageBudgetRepository,
	alertRepo billing.UsageAlertRepository,
	orgRepo organization.OrganizationRepository,
	notificationWorker *NotificationWorker,
) *UsageAggregationWorker {
	return &UsageAggregationWorker{
		config:             config,
		logger:             logger,
		usageRepo:          usageRepo,
		billingRepo:        billingRepo,
		pricingRepo:        pricingRepo,
		budgetRepo:         budgetRepo,
		alertRepo:          alertRepo,
		orgRepo:            orgRepo,
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

	// Get pricing config
	pricingConfig, err := w.pricingRepo.GetByID(ctx, orgBilling.PricingConfigID)
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

	// Calculate cost
	cost := w.calculateCost(summary, pricingConfig)

	// Calculate free tier remaining
	freeSpansRemaining := max(0, pricingConfig.FreeSpans-summary.TotalSpans)
	freeBytesRemaining := max(0, int64(pricingConfig.FreeGB*1073741824)-summary.TotalBytes)
	freeScoresRemaining := max(0, pricingConfig.FreeScores-summary.TotalScores)

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
	if err := w.syncBudgetUsage(ctx, orgID, summary, cost, pricingConfig); err != nil {
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
func (w *UsageAggregationWorker) syncBudgetUsage(ctx context.Context, orgID ulid.ULID, summary *billing.BillableUsageSummary, cost float64, pricingConfig *billing.PricingConfig) error {
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
			budgetCost = w.calculateRawCost(projectSummary, pricingConfig)
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
						budgetCost = w.calculateRawCost(orgPeriodSummary, pricingConfig)
					} else {
						costBeforeBudget := w.calculateCost(preBudgetSummary, pricingConfig)
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

// calculateCost computes total cost from three billable dimensions
func (w *UsageAggregationWorker) calculateCost(usage *billing.BillableUsageSummary, config *billing.PricingConfig) float64 {
	var totalCost float64

	// 1. Span cost: (spans - free_spans) / 100K * price_per_100k
	if config.PricePer100KSpans != nil {
		billableSpans := max(0, usage.TotalSpans-config.FreeSpans)
		spanCost := float64(billableSpans) / 100000.0 * *config.PricePer100KSpans
		totalCost += spanCost
	}

	// 2. Data cost: (bytes - free_bytes) / GB * price_per_gb
	if config.PricePerGB != nil {
		freeBytes := int64(config.FreeGB * 1073741824) // Convert GB to bytes
		billableBytes := max(0, usage.TotalBytes-freeBytes)
		billableGB := float64(billableBytes) / 1073741824.0
		dataCost := billableGB * *config.PricePerGB
		totalCost += dataCost
	}

	// 3. Score cost: (scores - free_scores) / 1K * price_per_1k
	if config.PricePer1KScores != nil {
		billableScores := max(0, usage.TotalScores-config.FreeScores)
		scoreCost := float64(billableScores) / 1000.0 * *config.PricePer1KScores
		totalCost += scoreCost
	}

	return totalCost
}

// calculateRawCost computes cost for usage without applying free tier deductions.
// Used for project-level budgets where free tier is already accounted at org level.
func (w *UsageAggregationWorker) calculateRawCost(usage *billing.BillableUsageSummary, config *billing.PricingConfig) float64 {
	var totalCost float64

	// 1. Span cost: spans / 100K * price_per_100k (no free tier)
	if config.PricePer100KSpans != nil {
		spanCost := float64(usage.TotalSpans) / 100000.0 * *config.PricePer100KSpans
		totalCost += spanCost
	}

	// 2. Data cost: bytes / GB * price_per_gb (no free tier)
	if config.PricePerGB != nil {
		billableGB := float64(usage.TotalBytes) / 1073741824.0
		dataCost := billableGB * *config.PricePerGB
		totalCost += dataCost
	}

	// 3. Score cost: scores / 1K * price_per_1k (no free tier)
	if config.PricePer1KScores != nil {
		scoreCost := float64(usage.TotalScores) / 1000.0 * *config.PricePer1KScores
		totalCost += scoreCost
	}

	return totalCost
}

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
