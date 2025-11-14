package observability

import (
	"context"
	"errors"
	"fmt"

	"brokle/internal/core/domain/observability"
)

// CostCalculatorService implements observability.CostCalculator
type CostCalculatorService struct {
	modelRepo observability.ModelRepository
}

// NewCostCalculatorService creates a new cost calculator service
func NewCostCalculatorService(modelRepo observability.ModelRepository) *CostCalculatorService {
	return &CostCalculatorService{
		modelRepo: modelRepo,
	}
}

// CalculateCost calculates cost breakdown for a span based on token usage
// Returns nil breakdown with zero costs if no pricing found (ingestion continues without failure)
func (s *CostCalculatorService) CalculateCost(
	ctx context.Context,
	input observability.CostCalculationInput,
) (*observability.CostBreakdown, error) {
	// Validate input
	if input.ModelName == "" {
		// Cost calculation skipped: no model name provided
		return s.zeroCostBreakdown(input), nil
	}

	if input.InputTokens == 0 && input.OutputTokens == 0 {
		// Cost calculation skipped: zero tokens
		return s.zeroCostBreakdown(input), nil
	}

	// Lookup model pricing (project-scoped with global fallback)
	pricing, err := s.modelRepo.FindByModelName(ctx, input.ModelName, input.ProjectID)
	if err != nil {
		// Check if it's expected "not found" error
		if errors.Is(err, observability.ErrModelNotFound) {
			// Expected - no pricing configured, return zero costs (graceful degradation)
			return s.zeroCostBreakdown(input), nil
		}
		// Unexpected database error - graceful degradation for telemetry ingestion
		// Don't fail ingestion due to pricing lookup errors
		return s.zeroCostBreakdown(input), nil
	}

	if pricing == nil {
		// No pricing found for model - continuing with zero costs
		return s.zeroCostBreakdown(input), nil
	}

	// Calculate costs using pricing entity methods
	return s.CalculateCostWithPricing(input, pricing), nil
}

// CalculateCostWithPricing calculates cost using provided pricing (for testing/preview)
func (s *CostCalculatorService) CalculateCostWithPricing(
	input observability.CostCalculationInput,
	pricing *observability.Model,
) *observability.CostBreakdown {
	if pricing == nil {
		return s.zeroCostBreakdown(input)
	}

	// Calculate costs using pricing entity methods
	inputCost := pricing.CalculateInputCost(input.InputTokens, input.CacheHit)
	outputCost := pricing.CalculateOutputCost(input.OutputTokens)
	totalCost := pricing.CalculateTotalCost(input.InputTokens, input.OutputTokens, input.CacheHit, input.BatchMode)

	// Calculate savings for attribution
	var cacheSavings, batchSavings *float64

	// Cache savings (if applicable)
	if input.CacheHit && pricing.CacheReadMultiplier > 0 && pricing.CacheReadMultiplier < 1.0 {
		baseCost := pricing.CalculateInputCost(input.InputTokens, false)
		savings := baseCost - inputCost
		cacheSavings = &savings
	}

	// Batch savings (if applicable)
	if input.BatchMode && pricing.BatchDiscountPercentage > 0 {
		baseTotal := inputCost + outputCost
		savings := baseTotal - totalCost
		batchSavings = &savings
	}

	// Format as strings with 9 decimal precision (ClickHouse Decimal(18,9))
	// CRITICAL: Costs stored as STRINGS in ClickHouse attributes (OTEL pattern)
	return &observability.CostBreakdown{
		InputCost:  fmt.Sprintf("%.9f", inputCost),
		OutputCost: fmt.Sprintf("%.9f", outputCost),
		TotalCost:  fmt.Sprintf("%.9f", totalCost),
		Currency:   "USD",

		// Metadata for attribution
		ModelName:    input.ModelName,
		Provider:     pricing.Provider,
		InputTokens:  input.InputTokens,
		OutputTokens: input.OutputTokens,
		CacheHit:     input.CacheHit,
		BatchMode:    input.BatchMode,

		// Savings attribution
		CacheSavings: cacheSavings,
		BatchSavings: batchSavings,
	}
}

// zeroCostBreakdown returns a cost breakdown with all zero costs
func (s *CostCalculatorService) zeroCostBreakdown(
	input observability.CostCalculationInput,
) *observability.CostBreakdown {
	return &observability.CostBreakdown{
		InputCost:    "0.000000000",
		OutputCost:   "0.000000000",
		TotalCost:    "0.000000000",
		Currency:     "USD",
		ModelName:    input.ModelName,
		Provider:     "",
		InputTokens:  input.InputTokens,
		OutputTokens: input.OutputTokens,
		CacheHit:     input.CacheHit,
		BatchMode:    input.BatchMode,
	}
}
