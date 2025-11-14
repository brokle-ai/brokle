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
	inputCost := (float64(req.InputTokens) / 1000.0) * req.Model.InputCostPer1kTokens

	// Estimate output tokens if not provided
	outputTokens := req.MaxTokens
	if outputTokens == 0 {
		// Default estimation: assume 25% of input tokens as output
		outputTokens = int32(math.Max(float64(req.InputTokens)*0.25, 100))
	}

	// Calculate output cost
	outputCost := (float64(outputTokens) / 1000.0) * req.Model.OutputCostPer1kTokens

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
	if req.Model.OutputCostPer1kTokens > 0 {
		// Separate input/output pricing
		inputCost := (float64(req.InputTokens) / 1000.0) * req.Model.InputCostPer1kTokens
		outputCost := (float64(req.MaxTokens) / 1000.0) * req.Model.OutputCostPer1kTokens
		totalCost = inputCost + outputCost
	} else {
		// Total token pricing
		totalCost = (float64(totalTokens) / 1000.0) * req.Model.InputCostPer1kTokens
	}

	// Apply discounts
	totalCost = c.applyDiscounts(ctx, totalCost, req.OrganizationID, req.Model)

	calculation := &gateway.CostCalculation{
		ModelID:         req.Model.ID,
		ProviderID:      req.Model.ProviderID,
		InputTokens:     req.InputTokens,
		OutputTokens:    req.MaxTokens,
		TotalTokens:     totalTokens,
		InputCost:       (float64(req.InputTokens) / 1000.0) * req.Model.InputCostPer1kTokens,
		OutputCost:      (float64(req.MaxTokens) / 1000.0) * req.Model.OutputCostPer1kTokens,
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
	inputCost := (float64(req.InputTokens) / 1000.0) * req.Model.InputCostPer1kTokens
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
	inputCost := (float64(req.InputTokens) / 1000.0) * req.Model.InputCostPer1kTokens
	outputCost := (float64(req.OutputTokens) / 1000.0) * req.Model.OutputCostPer1kTokens
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
		Duration:        &req.Duration,
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
		ModelID:            model.ID,
		ModelName:          model.ModelName,
		ProviderID:         model.ProviderID,
		InputCostPerToken:  model.InputCostPer1kTokens / 1000.0,
		OutputCostPerToken: model.OutputCostPer1kTokens / 1000.0,
		Currency:           "USD",
		EffectiveDate:      model.CreatedAt,
		IsActive:           model.IsEnabled,
	}

	logger.WithFields(logrus.Fields{
		"input_cost":  pricing.InputCostPerToken,
		"output_cost": pricing.OutputCostPerToken,
	}).Debug("Model pricing retrieved")

	return pricing, nil
}

// CompareProviderCosts compares costs across different providers for the same request
func (c *CostService) CompareProviderCosts(ctx context.Context, req *gateway.CostCalculationRequest) (*gateway.CostComparison, error) {
	// TODO: Implement cost comparison logic
	return nil, fmt.Errorf("not implemented")
}

// CalculateBatchCost calculates costs for multiple requests in batch
func (c *CostService) CalculateBatchCost(ctx context.Context, requests []*gateway.CostCalculationRequest) (*gateway.BatchCostResult, error) {
	logger := c.logger.WithField("request_count", len(requests))
	logger.Debug("Calculating batch cost")

	results := make([]*gateway.CostCalculationResult, len(requests))
	totalCost := 0.0

	for i, req := range requests {
		// Get model for this request
		model, err := c.modelRepo.GetByID(ctx, req.ModelID)
		if err != nil {
			errorMsg := fmt.Sprintf("failed to get model: %v", err)
			results[i] = &gateway.CostCalculationResult{
				RequestIndex: i,
				InputCost:    0,
				OutputCost:   0,
				TotalCost:    0,
				Currency:     "USD",
				ProviderID:   req.ModelID, // Use ModelID as fallback
				Error:        &errorMsg,
			}
			continue
		}

		// Calculate costs for this request
		inputCost := (float64(req.InputTokens) / 1000.0) * model.InputCostPer1kTokens
		outputCost := (float64(req.OutputTokens) / 1000.0) * model.OutputCostPer1kTokens
		requestCost := inputCost + outputCost

		results[i] = &gateway.CostCalculationResult{
			RequestIndex: i,
			InputCost:    inputCost,
			OutputCost:   outputCost,
			TotalCost:    requestCost,
			Currency:     "USD",
			ProviderID:   model.ProviderID,
			Error:        nil,
		}

		totalCost += requestCost
	}

	batchResult := &gateway.BatchCostResult{
		Requests:     requests,
		Results:      results,
		TotalCost:    totalCost,
		Currency:     "USD",
		CalculatedAt: time.Now(),
	}

	logger.WithField("total_cost", totalCost).Debug("Batch cost calculation completed")
	return batchResult, nil
}

// Helper methods

// applyDiscounts applies organization-specific discounts to the cost
func (c *CostService) applyDiscounts(ctx context.Context, cost float64, orgID ulid.ULID, model *gateway.Model) float64 {
	// TODO: Implement discount logic based on organization tier, usage volume, etc.
	return cost
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
			if model.ModelName == modelName {
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
		TokensPerSecond:    tokensPerSecond,
		OutputToInputRatio: outputRatio,
		TotalTokens:        totalTokens,
		Duration:           duration,
	}
}

// Interface method implementations required by gateway.CostService

// CalculateRequestCost calculates the cost for a specific model with token counts
func (c *CostService) CalculateRequestCost(ctx context.Context, modelID ulid.ULID, inputTokens, outputTokens int) (float64, error) {
	logger := c.logger.WithFields(logrus.Fields{
		"model_id":      modelID,
		"input_tokens":  inputTokens,
		"output_tokens": outputTokens,
	})

	logger.Debug("Calculating request cost")

	model, err := c.modelRepo.GetByID(ctx, modelID)
	if err != nil {
		logger.WithError(err).Error("Failed to get model")
		return 0, fmt.Errorf("failed to get model: %w", err)
	}

	inputCost := (float64(inputTokens) / 1000.0) * model.InputCostPer1kTokens
	outputCost := (float64(outputTokens) / 1000.0) * model.OutputCostPer1kTokens
	totalCost := inputCost + outputCost

	logger.WithField("total_cost", totalCost).Debug("Request cost calculated")
	return totalCost, nil
}

// EstimateRequestCost estimates the cost for a named model with estimated tokens
func (c *CostService) EstimateRequestCost(ctx context.Context, modelName string, estimatedTokens int) (float64, error) {
	logger := c.logger.WithFields(logrus.Fields{
		"model_name":       modelName,
		"estimated_tokens": estimatedTokens,
	})

	logger.Debug("Estimating request cost")

	model, err := c.modelRepo.GetByModelName(ctx, modelName)
	if err != nil {
		logger.WithError(err).Error("Failed to get model by name")
		return 0, fmt.Errorf("failed to get model: %w", err)
	}

	// Estimate assuming 75% input tokens, 25% output tokens
	inputTokens := int(float64(estimatedTokens) * 0.75)
	outputTokens := estimatedTokens - inputTokens

	inputCost := (float64(inputTokens) / 1000.0) * model.InputCostPer1kTokens
	outputCost := (float64(outputTokens) / 1000.0) * model.OutputCostPer1kTokens
	totalCost := inputCost + outputCost

	logger.WithField("estimated_cost", totalCost).Debug("Request cost estimated")
	return totalCost, nil
}

// GetCostOptimizedProvider returns the provider with the lowest cost for a given model
func (c *CostService) GetCostOptimizedProvider(ctx context.Context, projectID ulid.ULID, modelName string) (*gateway.RoutingDecision, error) {
	// TODO: Implement cost-optimized provider selection logic
	return nil, fmt.Errorf("not implemented")
}

// CompareCosts compares costs across different models for a given token count
func (c *CostService) CompareCosts(ctx context.Context, modelNames []string, tokenCount int) (*gateway.CostComparison, error) {
	// TODO: Implement cost comparison logic
	return nil, fmt.Errorf("not implemented")
}

// GetCostSavingsReport generates a cost savings report for a project
func (c *CostService) GetCostSavingsReport(ctx context.Context, projectID ulid.ULID, timeRange *gateway.TimeRange) (*gateway.CostSavingsReport, error) {
	// TODO: Implement cost savings report generation
	return nil, fmt.Errorf("not implemented")
}

// TrackRequestCost tracks the cost of a request for analytics
func (c *CostService) TrackRequestCost(ctx context.Context, metrics *gateway.RequestMetrics) error {
	logger := c.logger.WithFields(logrus.Fields{
		"request_id": metrics.RequestID,
		"project_id": metrics.ProjectID,
		"cost_usd":   metrics.CostUSD,
	})

	logger.Debug("Tracking request cost")

	// TODO: Implement actual cost tracking to analytics database
	// This would typically write to ClickHouse for real-time analytics

	logger.Debug("Request cost tracked successfully")
	return nil
}

// GetProjectCostAnalytics retrieves cost analytics for a project
func (c *CostService) GetProjectCostAnalytics(ctx context.Context, projectID ulid.ULID, timeRange *gateway.TimeRange) (*gateway.CostAnalytics, error) {
	// TODO: Implement cost analytics retrieval from ClickHouse
	return nil, fmt.Errorf("not implemented")
}

// GetProviderCostBreakdown retrieves cost breakdown by provider for a project
func (c *CostService) GetProviderCostBreakdown(ctx context.Context, projectID ulid.ULID, timeRange *gateway.TimeRange) (*gateway.ProviderCostBreakdown, error) {
	// TODO: Implement provider cost breakdown retrieval
	return nil, fmt.Errorf("not implemented")
}

// CheckBudgetLimits checks if a request would exceed budget limits
func (c *CostService) CheckBudgetLimits(ctx context.Context, projectID ulid.ULID, estimatedCost float64) (*gateway.BudgetCheckResult, error) {
	// TODO: Implement budget limits checking
	return &gateway.BudgetCheckResult{
		ProjectID:        projectID,
		CurrentUsage:     0.0,
		BudgetLimit:      0.0,
		RemainingBudget:  0.0,
		WillExceedBudget: false,
		CheckedAt:        time.Now(),
	}, nil
}

// UpdateBudgetUsage updates the budget usage for a project
func (c *CostService) UpdateBudgetUsage(ctx context.Context, projectID ulid.ULID, actualCost float64) error {
	// TODO: Implement budget usage updating
	return nil
}

// GetBudgetStatus retrieves the current budget status for a project
func (c *CostService) GetBudgetStatus(ctx context.Context, projectID ulid.ULID) (*gateway.BudgetStatus, error) {
	// TODO: Implement budget status retrieval
	return &gateway.BudgetStatus{
		ProjectID:         projectID,
		CurrentUsage:      0.0,
		BudgetLimit:       0.0,
		RemainingBudget:   0.0,
		BudgetUtilization: 0.0,
		OnTrack:           true,
		UpdatedAt:         time.Now(),
	}, nil
}
