package billing

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"brokle/internal/core/domain/billing"
	"brokle/pkg/ulid"
)

type billableUsageService struct {
	usageRepo      billing.BillableUsageRepository
	billingRepo    billing.OrganizationBillingRepository
	pricingService billing.PricingService
	planRepo       billing.PlanRepository
	logger         *slog.Logger
}

func NewBillableUsageService(
	usageRepo billing.BillableUsageRepository,
	billingRepo billing.OrganizationBillingRepository,
	pricingService billing.PricingService,
	planRepo billing.PlanRepository,
	logger *slog.Logger,
) billing.BillableUsageService {
	return &billableUsageService{
		usageRepo:      usageRepo,
		billingRepo:    billingRepo,
		pricingService: pricingService,
		planRepo:       planRepo,
		logger:         logger,
	}
}

func (s *billableUsageService) GetUsageOverview(ctx context.Context, orgID ulid.ULID) (*billing.UsageOverview, error) {
	// 1. Get billing metadata from PostgreSQL (pricing config, free tier, period dates)
	orgBilling, err := s.billingRepo.GetByOrgID(ctx, orgID)
	if err != nil {
		s.logger.Error("failed to get organization billing",
			"error", err,
			"organization_id", orgID,
		)
		return nil, err
	}

	// Get effective pricing (contract overrides > plan defaults)
	effectivePricing, err := s.pricingService.GetEffectivePricing(ctx, orgID)
	if err != nil {
		s.logger.Error("failed to get effective pricing",
			"error", err,
			"organization_id", orgID,
		)
		return nil, err
	}

	periodEnd := s.calculatePeriodEnd(orgBilling.BillingCycleStart, orgBilling.BillingCycleAnchorDay)

	// 2. Get REAL-TIME usage from ClickHouse
	filter := &billing.BillableUsageFilter{
		OrganizationID: orgID,
		Start:          orgBilling.BillingCycleStart,
		End:            time.Now().UTC(),
		Granularity:    "hourly",
	}

	usageSummary, err := s.usageRepo.GetUsageSummary(ctx, filter)
	if err != nil {
		s.logger.Warn("failed to get real-time usage, falling back to cached state",
			"error", err,
			"organization_id", orgID,
		)
		// Fallback to cached PostgreSQL values
		usageSummary = &billing.BillableUsageSummary{
			TotalSpans:  orgBilling.CurrentPeriodSpans,
			TotalBytes:  orgBilling.CurrentPeriodBytes,
			TotalScores: orgBilling.CurrentPeriodScores,
		}
	}

	// 3. Calculate real-time cost with tier support (delegates to pricing service)
	estimatedCost, err := s.pricingService.CalculateCostWithTiers(ctx, orgID, usageSummary)
	if err != nil {
		s.logger.Error("failed to calculate cost",
			"error", err,
			"organization_id", orgID,
		)
		return nil, err
	}

	// Use effective pricing for free tier calculations (respects contract overrides)
	freeSpansRemaining := max(0, effectivePricing.FreeSpans-usageSummary.TotalSpans)
	freeGBTotal := int64(effectivePricing.FreeGB * 1073741824)
	freeBytesRemaining := max(0, freeGBTotal-usageSummary.TotalBytes)
	freeScoresRemaining := max(0, effectivePricing.FreeScores-usageSummary.TotalScores)

	// 4. Return real-time overview
	return &billing.UsageOverview{
		OrganizationID: orgID,
		PeriodStart:    orgBilling.BillingCycleStart,
		PeriodEnd:      periodEnd,

		Spans:  usageSummary.TotalSpans,
		Bytes:  usageSummary.TotalBytes,
		Scores: usageSummary.TotalScores,

		FreeSpansRemaining:  freeSpansRemaining,
		FreeBytesRemaining:  freeBytesRemaining,
		FreeScoresRemaining: freeScoresRemaining,

		FreeSpansTotal:  effectivePricing.FreeSpans,
		FreeBytesTotal:  freeGBTotal,
		FreeScoresTotal: effectivePricing.FreeScores,

		EstimatedCost: estimatedCost,
	}, nil
}

func (s *billableUsageService) GetUsageTimeSeries(ctx context.Context, orgID ulid.ULID, start, end time.Time, granularity string) ([]*billing.BillableUsage, error) {
	filter := &billing.BillableUsageFilter{
		OrganizationID: orgID,
		Start:          start,
		End:            end,
		Granularity:    granularity,
	}

	usage, err := s.usageRepo.GetUsage(ctx, filter)
	if err != nil {
		s.logger.Error("failed to get usage time series",
			"error", err,
			"organization_id", orgID,
		)
		return nil, err
	}

	return usage, nil
}

func (s *billableUsageService) GetUsageByProject(ctx context.Context, orgID ulid.ULID, start, end time.Time) ([]*billing.BillableUsageSummary, error) {
	summaries, err := s.usageRepo.GetUsageByProject(ctx, orgID, start, end)
	if err != nil {
		s.logger.Error("failed to get usage by project",
			"error", err,
			"organization_id", orgID,
		)
		return nil, err
	}

	return summaries, nil
}

// CalculateCost delegates to pricing service for tier-aware cost calculation
// This maintains the interface contract while supporting enterprise custom pricing
func (s *billableUsageService) CalculateCost(ctx context.Context, usage *billing.BillableUsageSummary, plan *billing.Plan) float64 {
	// For backward compatibility with interface, we still accept plan parameter
	// but delegate to pricing service which handles contracts and volume tiers

	// Extract orgID from usage summary if available, otherwise use simple calculation
	// This is a transitional method - new code should use pricingService directly
	s.logger.Warn("CalculateCost called with plan parameter - consider using pricingService.CalculateCostWithTiers directly")

	// Simple flat calculation without contract awareness (legacy behavior)
	var totalCost float64

	if plan.PricePer100KSpans != nil {
		billableSpans := max(0, usage.TotalSpans-plan.FreeSpans)
		spanCost := float64(billableSpans) / 100000.0 * *plan.PricePer100KSpans
		totalCost += spanCost
	}

	if plan.PricePerGB != nil {
		freeBytes := int64(plan.FreeGB * 1073741824)
		billableBytes := max(0, usage.TotalBytes-freeBytes)
		billableGB := float64(billableBytes) / 1073741824.0
		dataCost := billableGB * *plan.PricePerGB
		totalCost += dataCost
	}

	if plan.PricePer1KScores != nil {
		billableScores := max(0, usage.TotalScores-plan.FreeScores)
		scoreCost := float64(billableScores) / 1000.0 * *plan.PricePer1KScores
		totalCost += scoreCost
	}

	return totalCost
}

func (s *billableUsageService) calculatePeriodEnd(cycleStart time.Time, anchorDay int) time.Time {
	nextMonth := cycleStart.AddDate(0, 1, 0)

	year, month, _ := nextMonth.Date()
	loc := nextMonth.Location()

	// Handle months with fewer days than anchor day
	lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, loc).Day()
	day := anchorDay
	if day > lastDay {
		day = lastDay
	}

	return time.Date(year, month, day, 0, 0, 0, 0, loc)
}

func (s *billableUsageService) ProvisionOrganizationBilling(ctx context.Context, orgID ulid.ULID) error {
	// Get default plan
	defaultPlan, err := s.planRepo.GetDefault(ctx)
	if err != nil {
		s.logger.Error("failed to get default pricing plan",
			"error", err,
			"organization_id", orgID,
		)
		return fmt.Errorf("get default pricing plan: %w", err)
	}

	now := time.Now()
	billingRecord := &billing.OrganizationBilling{
		OrganizationID:        orgID,
		PlanID:                defaultPlan.ID,
		BillingCycleStart:     now,
		BillingCycleAnchorDay: 1,
		FreeSpansRemaining:    defaultPlan.FreeSpans,
		FreeBytesRemaining:    int64(defaultPlan.FreeGB * 1024 * 1024 * 1024),
		FreeScoresRemaining:   defaultPlan.FreeScores,
		CurrentPeriodSpans:    0,
		CurrentPeriodBytes:    0,
		CurrentPeriodScores:   0,
		CurrentPeriodCost:     0,
		LastSyncedAt:          now,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	if err := s.billingRepo.Create(ctx, billingRecord); err != nil {
		// Idempotency check
		if strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "unique constraint") {
			s.logger.Info("billing record already exists", "organization_id", orgID)
			return nil // Success - already provisioned
		}
		return fmt.Errorf("create billing record: %w", err)
	}

	s.logger.Info("provisioned billing",
		"organization_id", orgID,
		"plan", defaultPlan.Name,
	)
	return nil
}
