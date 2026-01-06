package billing

import (
	"context"
	"log/slog"
	"time"

	"brokle/internal/core/domain/billing"
	"brokle/pkg/ulid"
)

type billableUsageService struct {
	usageRepo   billing.BillableUsageRepository
	billingRepo billing.OrganizationBillingRepository
	pricingRepo billing.PricingConfigRepository
	logger      *slog.Logger
}

func NewBillableUsageService(
	usageRepo billing.BillableUsageRepository,
	billingRepo billing.OrganizationBillingRepository,
	pricingRepo billing.PricingConfigRepository,
	logger *slog.Logger,
) billing.BillableUsageService {
	return &billableUsageService{
		usageRepo:   usageRepo,
		billingRepo: billingRepo,
		pricingRepo: pricingRepo,
		logger:      logger,
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

	pricingConfig, err := s.pricingRepo.GetByID(ctx, orgBilling.PricingConfigID)
	if err != nil {
		s.logger.Error("failed to get pricing config",
			"error", err,
			"pricing_config_id", orgBilling.PricingConfigID,
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

	// 3. Calculate real-time cost and free tier remaining
	estimatedCost := s.CalculateCost(ctx, usageSummary, pricingConfig)

	freeSpansRemaining := max(0, pricingConfig.FreeSpans-usageSummary.TotalSpans)
	freeGBTotal := int64(pricingConfig.FreeGB * 1073741824)
	freeBytesRemaining := max(0, freeGBTotal-usageSummary.TotalBytes)
	freeScoresRemaining := max(0, pricingConfig.FreeScores-usageSummary.TotalScores)

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

		FreeSpansTotal:  pricingConfig.FreeSpans,
		FreeBytesTotal:  freeGBTotal,
		FreeScoresTotal: pricingConfig.FreeScores,

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

// CalculateCost computes total cost from three billable dimensions
// Formula: (spans - free_spans) / 100K × price_per_100k + (bytes - free_bytes) / GB × price_per_gb + (scores - free_scores) / 1K × price_per_1k
func (s *billableUsageService) CalculateCost(ctx context.Context, usage *billing.BillableUsageSummary, config *billing.PricingConfig) float64 {
	var totalCost float64

	// 1. Span cost: (spans - free_spans) / 100K × price_per_100k
	if config.PricePer100KSpans != nil {
		billableSpans := max(0, usage.TotalSpans-config.FreeSpans)
		spanCost := float64(billableSpans) / 100000.0 * *config.PricePer100KSpans
		totalCost += spanCost
	}

	// 2. Data cost: (bytes - free_bytes) / GB × price_per_gb
	if config.PricePerGB != nil {
		freeBytes := int64(config.FreeGB * 1073741824) // Convert GB to bytes
		billableBytes := max(0, usage.TotalBytes-freeBytes)
		billableGB := float64(billableBytes) / 1073741824.0
		dataCost := billableGB * *config.PricePerGB
		totalCost += dataCost
	}

	// 3. Score cost: (scores - free_scores) / 1K × price_per_1k
	if config.PricePer1KScores != nil {
		billableScores := max(0, usage.TotalScores-config.FreeScores)
		scoreCost := float64(billableScores) / 1000.0 * *config.PricePer1KScores
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
