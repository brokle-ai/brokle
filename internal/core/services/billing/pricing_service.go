package billing

import (
	"context"
	"log/slog"
	"sort"

	"brokle/internal/core/domain/billing"
	"brokle/pkg/ulid"
)

type pricingService struct {
	billingRepo  billing.OrganizationBillingRepository
	planRepo     billing.PlanRepository
	contractRepo billing.ContractRepository
	tierRepo     billing.VolumeDiscountTierRepository
	logger       *slog.Logger
}

func NewPricingService(
	billingRepo billing.OrganizationBillingRepository,
	planRepo billing.PlanRepository,
	contractRepo billing.ContractRepository,
	tierRepo billing.VolumeDiscountTierRepository,
	logger *slog.Logger,
) billing.PricingService {
	return &pricingService{
		billingRepo:  billingRepo,
		planRepo:     planRepo,
		contractRepo: contractRepo,
		tierRepo:     tierRepo,
		logger:       logger,
	}
}

// GetEffectivePricing resolves pricing: contract overrides > plan defaults
func (s *pricingService) GetEffectivePricing(ctx context.Context, orgID ulid.ULID) (*billing.EffectivePricing, error) {
	// 1. Get organization's base plan
	orgBilling, err := s.billingRepo.GetByOrgID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	plan, err := s.planRepo.GetByID(ctx, orgBilling.PlanID)
	if err != nil {
		return nil, err
	}

	// 2. Check for active contract
	contract, err := s.contractRepo.GetActiveByOrgID(ctx, orgID)
	if err != nil {
		return nil, err // Real database error
	}
	// contract will be nil if no active contract exists (valid state)

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
		tiers, err := s.tierRepo.GetByContractID(ctx, contract.ID)
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

// CalculateCostWithTiers calculates cost with volume tier support
func (s *pricingService) CalculateCostWithTiers(ctx context.Context, orgID ulid.ULID, usage *billing.BillableUsageSummary) (float64, error) {
	effective, err := s.GetEffectivePricing(ctx, orgID)
	if err != nil {
		return 0, err
	}

	if effective.HasVolumeTiers {
		return s.calculateWithTiers(usage, effective), nil
	}

	return s.calculateFlat(usage, effective), nil
}

// CalculateCostWithTiersNoFreeTier calculates cost with tier support but without free tier deductions
// Used for project-level budgets where free tier is org-level only
func (s *pricingService) CalculateCostWithTiersNoFreeTier(ctx context.Context, orgID ulid.ULID, usage *billing.BillableUsageSummary) (float64, error) {
	effective, err := s.GetEffectivePricing(ctx, orgID)
	if err != nil {
		return 0, err
	}

	if effective.HasVolumeTiers {
		return s.calculateWithTiersNoFreeTier(usage, effective), nil
	}

	return s.calculateFlatNoFreeTier(usage, effective), nil
}

// calculateFlat uses simple linear pricing (current implementation)
func (s *pricingService) calculateFlat(usage *billing.BillableUsageSummary, pricing *billing.EffectivePricing) float64 {
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
func (s *pricingService) calculateWithTiers(usage *billing.BillableUsageSummary, pricing *billing.EffectivePricing) float64 {
	totalCost := 0.0

	// Calculate each dimension
	totalCost += s.CalculateDimensionWithTiers(usage.TotalSpans, pricing.FreeSpans, billing.TierDimensionSpans, pricing.VolumeTiers, pricing)

	freeBytes := int64(pricing.FreeGB * 1073741824)
	totalCost += s.CalculateDimensionWithTiers(usage.TotalBytes, freeBytes, billing.TierDimensionBytes, pricing.VolumeTiers, pricing)

	totalCost += s.CalculateDimensionWithTiers(usage.TotalScores, pricing.FreeScores, billing.TierDimensionScores, pricing.VolumeTiers, pricing)

	return totalCost
}

// calculateFlatNoFreeTier uses simple linear pricing without free tier deductions
func (s *pricingService) calculateFlatNoFreeTier(usage *billing.BillableUsageSummary, pricing *billing.EffectivePricing) float64 {
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
func (s *pricingService) calculateWithTiersNoFreeTier(usage *billing.BillableUsageSummary, pricing *billing.EffectivePricing) float64 {
	totalCost := 0.0

	// Calculate each dimension without free tier
	totalCost += s.CalculateDimensionWithTiers(usage.TotalSpans, 0, billing.TierDimensionSpans, pricing.VolumeTiers, pricing)
	totalCost += s.CalculateDimensionWithTiers(usage.TotalBytes, 0, billing.TierDimensionBytes, pricing.VolumeTiers, pricing)
	totalCost += s.CalculateDimensionWithTiers(usage.TotalScores, 0, billing.TierDimensionScores, pricing.VolumeTiers, pricing)

	return totalCost
}

// CalculateDimensionWithTiers applies progressive pricing using absolute position mapping with free tier offset
// The algorithm works in absolute coordinate space: billable range is [freeTier, usage), not [0, billableUsage)
// This ensures free tier correctly offsets tier boundaries (e.g., free=500 with tier [0-1k] charges usage 500-1k in that tier)
// Exported for use by workers that need per-dimension cost calculation
func (s *pricingService) CalculateDimensionWithTiers(usage, freeTier int64, dimension billing.TierDimension, allTiers []*billing.VolumeDiscountTier, pricing *billing.EffectivePricing) float64 {
	// Early exit: all usage covered by free tier
	if usage <= freeTier {
		return 0
	}

	// Filter and sort tiers for this dimension
	var tiers []*billing.VolumeDiscountTier
	for _, t := range allTiers {
		if t.Dimension == dimension {
			tiers = append(tiers, t)
		}
	}

	if len(tiers) == 0 {
		// No tiers defined for this dimension, fallback to flat pricing on billable amount
		billableUsage := usage - freeTier
		return s.calculateFlatDimension(billableUsage, dimension, pricing)
	}

	sort.Slice(tiers, func(i, j int) bool {
		return tiers[i].TierMin < tiers[j].TierMin
	})

	totalCost := 0.0

	for _, tier := range tiers {
		// Calculate overlap between billable range [freeTier, usage)
		// and tier range [tier.TierMin, tier.TierMax) in ABSOLUTE coordinates
		//
		// Example: free=500, tier=[0-1000], usage=1500
		//   overlapStart = max(500, 0) = 500
		//   overlapEnd = min(1500, 1000) = 1000
		//   usageInTier = 1000 - 500 = 500 (charges 500 units in this tier)

		overlapStart := max(freeTier, tier.TierMin)

		var overlapEnd int64
		if tier.TierMax == nil {
			overlapEnd = usage // Unlimited tier extends to total usage
		} else {
			overlapEnd = min(usage, *tier.TierMax)
		}

		// Skip if no overlap
		if overlapStart >= overlapEnd {
			continue
		}

		// Calculate billable usage in this tier
		usageInTier := overlapEnd - overlapStart

		// Convert to billable units and apply price
		unitSize := getDimensionUnitSize(dimension)
		units := float64(usageInTier) / float64(unitSize)
		cost := units * tier.PricePerUnit

		totalCost += cost

		// Optimization: stop if usage fully consumed
		if tier.TierMax == nil || usage <= *tier.TierMax {
			break
		}
	}

	return totalCost
}

// Helper functions

func getDimensionUnitSize(dimension billing.TierDimension) int64 {
	switch dimension {
	case billing.TierDimensionSpans:
		return 100000 // per 100K
	case billing.TierDimensionBytes:
		return 1073741824 // per GB
	case billing.TierDimensionScores:
		return 1000 // per 1K
	default:
		return 1
	}
}

// calculateFlatDimension calculates cost for a single dimension using flat pricing
// Used as fallback when no volume tiers are defined for a dimension
func (s *pricingService) calculateFlatDimension(billableUsage int64, dimension billing.TierDimension, pricing *billing.EffectivePricing) float64 {
	if billableUsage == 0 {
		return 0
	}

	unitSize := getDimensionUnitSize(dimension)
	units := float64(billableUsage) / float64(unitSize)

	switch dimension {
	case billing.TierDimensionSpans:
		return units * pricing.PricePer100KSpans
	case billing.TierDimensionBytes:
		return units * pricing.PricePerGB
	case billing.TierDimensionScores:
		return units * pricing.PricePer1KScores
	default:
		return 0
	}
}

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
