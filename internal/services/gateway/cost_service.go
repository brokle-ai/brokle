package gateway

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/sirupsen/logrus"

	"brokle/internal/core/domain/gateway"
	"brokle/pkg/ulid"
)

// CostService implements cost calculation and tracking functionality
type CostService struct {
	modelRepo          gateway.ModelRepository
	providerRepo       gateway.ProviderRepository
	providerConfigRepo gateway.ProviderConfigRepository
	logger             *logrus.Logger
}

// NewCostService creates a new cost service instance
func NewCostService(
	modelRepo gateway.ModelRepository,
	providerRepo gateway.ProviderRepository,
	providerConfigRepo gateway.ProviderConfigRepository,
	logger *logrus.Logger,
) gateway.CostService {
	return &CostService{
		modelRepo:          modelRepo,
		providerRepo:       providerRepo,
		providerConfigRepo: providerConfigRepo,
		logger:             logger,
	}
}

// EstimateChatCompletionCost estimates the cost for a chat completion request
func (c *CostService) EstimateChatCompletionCost(ctx context.Context, req *gateway.CostEstimationRequest) (*gateway.CostCalculation, error) {
	logger := c.logger.WithFields(logrus.Fields{
		"model_id":        req.Model.ID,
		"organization_id": req.OrganizationID,
		"input_tokens":    req.InputTokens,
		"max_tokens":      req.MaxTokens,
	})

	logger.Debug("Estimating chat completion cost")

	// Calculate input cost
	inputCost := float64(req.InputTokens) * req.Model.InputCostPerToken

	// Estimate output tokens if not provided
	outputTokens := req.MaxTokens
	if outputTokens == 0 {
		// Default estimation: assume 25% of input tokens as output
		outputTokens = int32(math.Max(float64(req.InputTokens)*0.25, 100))
	}

	// Calculate output cost
	outputCost := float64(outputTokens) * req.Model.OutputCostPerToken

	// Calculate total cost
	totalCost := inputCost + outputCost

	// Apply any organization-specific discounts
	totalCost = c.applyDiscounts(ctx, totalCost, req.OrganizationID, req.Model)

	calculation := &gateway.CostCalculation{
		ModelID:         req.Model.ID,
		ProviderID:      req.Model.ProviderID,
		InputTokens:     req.InputTokens,
		OutputTokens:    outputTokens,
		TotalTokens:     req.InputTokens + outputTokens,
		InputCost:       inputCost,
		OutputCost:      outputCost,
		TotalCost:       totalCost,
		Currency:        "USD",
		EstimatedAt:     time.Now(),
		CalculationType: "estimated",
	}

	logger.WithFields(logrus.Fields{
		"input_cost":    inputCost,
		"output_cost":   outputCost,
		"total_cost":    totalCost,
		"output_tokens": outputTokens,
	}).Debug("Chat completion cost estimated")

	return calculation, nil
}

// EstimateCompletionCost estimates the cost for a text completion request
func (c *CostService) EstimateCompletionCost(ctx context.Context, req *gateway.CostEstimationRequest) (*gateway.CostCalculation, error) {
	logger := c.logger.WithFields(logrus.Fields{
		"model_id":        req.Model.ID,
		"organization_id": req.OrganizationID,
		"input_tokens":    req.InputTokens,
		"max_tokens":      req.MaxTokens,
	})

	logger.Debug("Estimating completion cost")

	// For completion models, cost is typically per total token
	totalTokens := req.InputTokens + req.MaxTokens

	// Some models use input/output pricing, others use total token pricing
	var totalCost float64
	if req.Model.OutputCostPerToken > 0 {
		// Separate input/output pricing
		inputCost := float64(req.InputTokens) * req.Model.InputCostPerToken
		outputCost := float64(req.MaxTokens) * req.Model.OutputCostPerToken
		totalCost = inputCost + outputCost
	} else {
		// Total token pricing
		totalCost = float64(totalTokens) * req.Model.InputCostPerToken
	}

	// Apply discounts
	totalCost = c.applyDiscounts(ctx, totalCost, req.OrganizationID, req.Model)

	calculation := &gateway.CostCalculation{
		ModelID:         req.Model.ID,
		ProviderID:      req.Model.ProviderID,
		InputTokens:     req.InputTokens,
		OutputTokens:    req.MaxTokens,
		TotalTokens:     totalTokens,
		InputCost:       float64(req.InputTokens) * req.Model.InputCostPerToken,
		OutputCost:      float64(req.MaxTokens) * req.Model.OutputCostPerToken,
		TotalCost:       totalCost,
		Currency:        "USD",
		EstimatedAt:     time.Now(),
		CalculationType: "estimated",
	}

	logger.WithField("total_cost", totalCost).Debug("Completion cost estimated")

	return calculation, nil
}

// EstimateEmbeddingsCost estimates the cost for an embeddings request
func (c *CostService) EstimateEmbeddingsCost(ctx context.Context, req *gateway.CostEstimationRequest) (*gateway.CostCalculation, error) {
	logger := c.logger.WithFields(logrus.Fields{
		"model_id":        req.Model.ID,
		"organization_id": req.OrganizationID,
		"input_tokens":    req.InputTokens,
	})

	logger.Debug("Estimating embeddings cost")

	// Embeddings typically only have input cost
	inputCost := float64(req.InputTokens) * req.Model.InputCostPerToken
	totalCost := c.applyDiscounts(ctx, inputCost, req.OrganizationID, req.Model)

	calculation := &gateway.CostCalculation{
		ModelID:         req.Model.ID,
		ProviderID:      req.Model.ProviderID,
		InputTokens:     req.InputTokens,
		OutputTokens:    0,
		TotalTokens:     req.InputTokens,
		InputCost:       inputCost,
		OutputCost:      0,
		TotalCost:       totalCost,
		Currency:        "USD",
		EstimatedAt:     time.Now(),
		CalculationType: "estimated",
	}

	logger.WithField("total_cost", totalCost).Debug("Embeddings cost estimated")

	return calculation, nil
}

// CalculateActualCost calculates the actual cost based on usage
func (c *CostService) CalculateActualCost(ctx context.Context, req *gateway.ActualCostRequest) (*gateway.CostCalculation, error) {
	logger := c.logger.WithFields(logrus.Fields{
		"model_id":        req.Model.ID,
		"organization_id": req.OrganizationID,
		"input_tokens":    req.InputTokens,
		"output_tokens":   req.OutputTokens,
	})

	logger.Debug("Calculating actual cost")

	// Calculate actual costs
	inputCost := float64(req.InputTokens) * req.Model.InputCostPerToken
	outputCost := float64(req.OutputTokens) * req.Model.OutputCostPerToken
	totalCost := inputCost + outputCost

	// Apply discounts
	totalCost = c.applyDiscounts(ctx, totalCost, req.OrganizationID, req.Model)

	calculation := &gateway.CostCalculation{
		ModelID:         req.Model.ID,
		ProviderID:      req.Model.ProviderID,
		RequestID:       &req.RequestID,
		InputTokens:     req.InputTokens,
		OutputTokens:    req.OutputTokens,
		TotalTokens:     req.InputTokens + req.OutputTokens,
		InputCost:       inputCost,
		OutputCost:      outputCost,
		TotalCost:       totalCost,
		Currency:        "USD",
		EstimatedAt:     time.Now(),
		CalculationType: "actual",
		Duration:        req.Duration,
	}

	logger.WithFields(logrus.Fields{
		"input_cost":  inputCost,
		"output_cost": outputCost,
		"total_cost":  totalCost,
		"duration":    req.Duration,
	}).Debug("Actual cost calculated")

	return calculation, nil
}

// GetUsageStats retrieves usage statistics for an organization
func (c *CostService) GetUsageStats(ctx context.Context, req *gateway.UsageStatsRequest) (*gateway.UsageStatsResponse, error) {
	logger := c.logger.WithFields(logrus.Fields{
		"organization_id": req.OrganizationID,
		"start_date":      req.StartDate,
		"end_date":        req.EndDate,
	})

	logger.Info("Retrieving usage statistics")

	// TODO: Implement actual usage statistics retrieval from analytics database
	// This would typically query the ClickHouse analytics tables for aggregated usage data

	// Placeholder implementation
	stats := &gateway.UsageStatsResponse{
		OrganizationID: req.OrganizationID,
		StartDate:      req.StartDate,
		EndDate:        req.EndDate,
		TotalRequests:  0,
		TotalTokens:    0,
		TotalCost:      0.0,
		Currency:       "USD",
		ModelStats:     make(map[string]*gateway.ModelUsageStats),
		ProviderStats:  make(map[string]*gateway.ProviderUsageStats),
		DailyStats:     make([]*gateway.DailyUsageStats, 0),
	}

	logger.WithFields(logrus.Fields{
		"total_requests": stats.TotalRequests,
		"total_tokens":   stats.TotalTokens,
		"total_cost":     stats.TotalCost,
	}).Info("Usage statistics retrieved")

	return stats, nil
}

// GetCostBreakdown provides detailed cost breakdown for an organization
func (c *CostService) GetCostBreakdown(ctx context.Context, req *gateway.CostBreakdownRequest) (*gateway.CostBreakdownResponse, error) {
	logger := c.logger.WithFields(logrus.Fields{
		"organization_id": req.OrganizationID,
		"start_date":      req.StartDate,
		"end_date":        req.EndDate,
		"group_by":        req.GroupBy,
	})

	logger.Info("Retrieving cost breakdown")

	// TODO: Implement actual cost breakdown retrieval from analytics database
	// This would typically query aggregated cost data grouped by the requested dimensions

	breakdown := &gateway.CostBreakdownResponse{
		OrganizationID: req.OrganizationID,
		StartDate:      req.StartDate,
		EndDate:        req.EndDate,
		TotalCost:      0.0,
		Currency:       "USD",
		GroupBy:        req.GroupBy,
		Breakdown:      make([]*gateway.CostBreakdownItem, 0),
	}

	logger.WithField("total_cost", breakdown.TotalCost).Info("Cost breakdown retrieved")

	return breakdown, nil
}

// TrackUsage tracks usage for analytics and billing
func (c *CostService) TrackUsage(ctx context.Context, req *gateway.UsageTrackingRequest) error {
	logger := c.logger.WithFields(logrus.Fields{
		"request_id":      req.RequestID,
		"organization_id": req.OrganizationID,
		"model_id":        req.ModelID,
		"provider_id":     req.ProviderID,
		"total_cost":      req.Cost.TotalCost,
	})

	logger.Debug("Tracking usage")

	// TODO: Implement usage tracking to analytics database
	// This would typically write to ClickHouse for real-time analytics
	// and potentially to a billing system for cost tracking

	logger.Debug("Usage tracked successfully")

	return nil
}

// GetModelPricing retrieves current pricing for a model
func (c *CostService) GetModelPricing(ctx context.Context, modelID ulid.ULID) (*gateway.ModelPricing, error) {
	logger := c.logger.WithField("model_id", modelID)
	logger.Debug("Retrieving model pricing")

	model, err := c.modelRepo.GetByID(ctx, modelID)
	if err != nil {
		logger.WithError(err).Error("Failed to get model")
		return nil, fmt.Errorf("failed to get model: %w", err)
	}

	pricing := &gateway.ModelPricing{
		ModelID:             model.ID,
		ModelName:           model.Name,
		ProviderID:          model.ProviderID,
		InputCostPerToken:   model.InputCostPerToken,
		OutputCostPerToken:  model.OutputCostPerToken,
		Currency:            "USD",
		EffectiveDate:       model.CreatedAt,
		IsActive:            model.IsEnabled,
	}

	logger.WithFields(logrus.Fields{
		"input_cost":  pricing.InputCostPerToken,
		"output_cost": pricing.OutputCostPerToken,
	}).Debug("Model pricing retrieved")

	return pricing, nil
}

// CompareProviderCosts compares costs across different providers for the same request
func (c *CostService) CompareProviderCosts(ctx context.Context, req *gateway.ProviderCostComparisonRequest) (*gateway.ProviderCostComparisonResponse, error) {
	logger := c.logger.WithFields(logrus.Fields{
		"organization_id": req.OrganizationID,
		"model_name":      req.ModelName,
		"input_tokens":    req.InputTokens,
		"output_tokens":   req.OutputTokens,
	})

	logger.Info("Comparing provider costs")

	// Find all models with the same name or similar capabilities
	models, err := c.findSimilarModels(ctx, req.ModelName, req.ModelType)
	if err != nil {
		logger.WithError(err).Error("Failed to find similar models")
		return nil, fmt.Errorf("failed to find similar models: %w", err)
	}

	var comparisons []*gateway.ProviderCostComparison
	for _, model := range models {
		// Calculate cost for this model
		inputCost := float64(req.InputTokens) * model.InputCostPerToken
		outputCost := float64(req.OutputTokens) * model.OutputCostPerToken
		totalCost := inputCost + outputCost

		// Apply discounts
		totalCost = c.applyDiscounts(ctx, totalCost, req.OrganizationID, model)

		// Get provider info
		provider, err := c.providerRepo.GetByID(ctx, model.ProviderID)
		if err != nil {
			logger.WithError(err).WithField("provider_id", model.ProviderID).Warn("Failed to get provider")
			continue
		}

		comparison := &gateway.ProviderCostComparison{
			ProviderID:     provider.ID,
			ProviderName:   provider.Name,
			ModelID:        model.ID,
			ModelName:      model.Name,
			InputCost:      inputCost,
			OutputCost:     outputCost,
			TotalCost:      totalCost,
			Currency:       "USD",
			ContextLength:  model.ContextLength,
			MaxTokens:      model.MaxTokens,
			IsAvailable:    model.IsEnabled && provider.IsEnabled,
		}

		comparisons = append(comparisons, comparison)
	}

	// Sort by total cost
	for i := 0; i < len(comparisons)-1; i++ {
		for j := i + 1; j < len(comparisons); j++ {
			if comparisons[i].TotalCost > comparisons[j].TotalCost {
				comparisons[i], comparisons[j] = comparisons[j], comparisons[i]
			}
		}
	}

	response := &gateway.ProviderCostComparisonResponse{
		OrganizationID: req.OrganizationID,
		ModelName:      req.ModelName,
		InputTokens:    req.InputTokens,
		OutputTokens:   req.OutputTokens,
		Currency:       "USD",
		Comparisons:    comparisons,
		ComparedAt:     time.Now(),
	}

	logger.WithField("comparison_count", len(comparisons)).Info("Provider cost comparison completed")

	return response, nil
}

// Helper methods

func (c *CostService) applyDiscounts(ctx context.Context, baseCost float64, orgID ulid.ULID, model *gateway.Model) float64 {
	// TODO: Implement organization-specific discount logic
	// This could include:
	// - Volume discounts based on usage tiers
	// - Enterprise contract discounts
	// - Promotional discounts
	// - Provider-specific negotiated rates

	// For now, return the base cost
	return baseCost
}

func (c *CostService) findSimilarModels(ctx context.Context, modelName string, modelType *gateway.ModelType) ([]*gateway.Model, error) {
	// Find models with similar capabilities or the same name
	filter := &gateway.ModelFilter{
		IsEnabled: &[]bool{true}[0],
	}

	if modelType != nil {
		filter.ModelType = modelType
	}

	models, _, err := c.modelRepo.SearchModels(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to search models: %w", err)
	}

	// If specific model name is provided, try to find exact matches first
	if modelName != "" {
		var exactMatches []*gateway.Model
		for _, model := range models {
			if model.Name == modelName {
				exactMatches = append(exactMatches, model)
			}
		}
		if len(exactMatches) > 0 {
			return exactMatches, nil
		}
	}

	return models, nil
}

// Utility functions for cost calculations

// EstimateTokensFromText estimates token count from text
func (c *CostService) EstimateTokensFromText(text string) int32 {
	// Simple approximation: 4 characters per token
	// In production, use a proper tokenizer for accurate counts
	return int32(len(text) / 4)
}

// EstimateTokensFromMessages estimates token count from chat messages
func (c *CostService) EstimateTokensFromMessages(messages []gateway.ChatMessage) int32 {
	totalChars := 0
	for _, msg := range messages {
		// Add some overhead for role and formatting
		totalChars += len(msg.Content) + len(msg.Role) + 10
	}
	return int32(totalChars / 4)
}

// CalculateTokenEfficiency calculates efficiency metrics for usage analysis
func (c *CostService) CalculateTokenEfficiency(inputTokens, outputTokens int32, duration time.Duration) *gateway.EfficiencyMetrics {
	totalTokens := inputTokens + outputTokens
	tokensPerSecond := 0.0
	if duration > 0 {
		tokensPerSecond = float64(totalTokens) / duration.Seconds()
	}

	outputRatio := 0.0
	if totalTokens > 0 {
		outputRatio = float64(outputTokens) / float64(totalTokens)
	}

	return &gateway.EfficiencyMetrics{
		TokensPerSecond:     tokensPerSecond,
		OutputToInputRatio:  outputRatio,
		TotalTokens:         totalTokens,
		Duration:            duration,
	}
}