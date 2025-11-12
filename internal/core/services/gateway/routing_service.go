package gateway

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/sirupsen/logrus"

	"brokle/internal/core/domain/gateway"
	"brokle/pkg/ulid"
)

// RoutingService implements intelligent provider routing
type RoutingService struct {
	providerRepo       gateway.ProviderRepository
	modelRepo          gateway.ModelRepository
	providerConfigRepo gateway.ProviderConfigRepository
	costService        gateway.CostService
	logger             *logrus.Logger
}

// NewRoutingService creates a new routing service instance
func NewRoutingService(
	providerRepo gateway.ProviderRepository,
	modelRepo gateway.ModelRepository,
	providerConfigRepo gateway.ProviderConfigRepository,
	costService gateway.CostService,
	logger *logrus.Logger,
) gateway.RoutingService {
	return &RoutingService{
		providerRepo:       providerRepo,
		modelRepo:          modelRepo,
		providerConfigRepo: providerConfigRepo,
		costService:        costService,
		logger:             logger,
	}
}

// RouteRequest routes a request to the best provider based on strategy
func (r *RoutingService) RouteRequest(ctx context.Context, projectID ulid.ULID, request *gateway.RoutingRequest) (*gateway.RoutingDecision, error) {
	logger := r.logger.WithFields(logrus.Fields{
		"project_id": projectID,
		"model_name": request.ModelName,
		"strategy":   request.Strategy,
	})

	logger.Info("Routing request")

	// Get model
	model, err := r.modelRepo.GetByModelName(ctx, request.ModelName)
	if err != nil {
		return nil, fmt.Errorf("model not found: %w", err)
	}

	// Get provider
	provider, err := r.providerRepo.GetByID(ctx, model.ProviderID)
	if err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}

	// Get provider config for project
	_, err = r.providerConfigRepo.GetByProjectAndProvider(ctx, projectID, provider.ID)
	if err != nil {
		return nil, fmt.Errorf("provider config not found: %w", err)
	}

	// Calculate estimated cost
	estimatedCost := 0.0
	if request.EstimatedTokens != nil && *request.EstimatedTokens > 0 {
		estimatedCost, _ = r.costService.EstimateRequestCost(ctx, model.ModelName, *request.EstimatedTokens)
	}

	strategyStr := "default"
	if request.Strategy != nil {
		strategyStr = string(*request.Strategy)
	}

	decision := &gateway.RoutingDecision{
		Provider:         provider.Name,
		ProviderID:       provider.ID,
		Model:            model.ModelName,
		ModelID:          model.ID,
		Strategy:         strategyStr,
		EstimatedCost:    estimatedCost,
		EstimatedLatency: 0, // TODO: Implement latency tracking
		RoutingReason:    fmt.Sprintf("Selected based on %s strategy", strategyStr),
	}

	logger.WithFields(logrus.Fields{
		"provider":       decision.Provider,
		"estimated_cost": decision.EstimatedCost,
	}).Info("Request routed successfully")

	return decision, nil
}

// GetBestProvider returns the best provider for a model using specified strategy
func (r *RoutingService) GetBestProvider(ctx context.Context, projectID ulid.ULID, modelName string, strategy gateway.RoutingStrategy) (*gateway.RoutingDecision, error) {
	request := &gateway.RoutingRequest{
		ModelName: modelName,
		Strategy:  &strategy,
	}
	return r.RouteRequest(ctx, projectID, request)
}

// RouteByCost routes by lowest cost
func (r *RoutingService) RouteByCost(ctx context.Context, projectID ulid.ULID, modelName string) (*gateway.RoutingDecision, error) {
	return r.GetBestProvider(ctx, projectID, modelName, gateway.RoutingStrategyCostOptimized)
}

// RouteByLatency routes by lowest latency
func (r *RoutingService) RouteByLatency(ctx context.Context, projectID ulid.ULID, modelName string) (*gateway.RoutingDecision, error) {
	return r.GetBestProvider(ctx, projectID, modelName, gateway.RoutingStrategyLatencyOptimized)
}

// RouteByQuality routes by highest quality
func (r *RoutingService) RouteByQuality(ctx context.Context, projectID ulid.ULID, modelName string) (*gateway.RoutingDecision, error) {
	return r.GetBestProvider(ctx, projectID, modelName, gateway.RoutingStrategyQualityOptimized)
}

// RouteByLoad routes by load balancing
func (r *RoutingService) RouteByLoad(ctx context.Context, projectID ulid.ULID, modelName string) (*gateway.RoutingDecision, error) {
	return r.GetBestProvider(ctx, projectID, modelName, gateway.RoutingStrategyLoadBalance)
}

// GetFallbackProvider gets a fallback provider when primary fails
func (r *RoutingService) GetFallbackProvider(ctx context.Context, projectID ulid.ULID, failedProviderID ulid.ULID, modelName string) (*gateway.RoutingDecision, error) {
	logger := r.logger.WithFields(logrus.Fields{
		"project_id":         projectID,
		"failed_provider_id": failedProviderID,
		"model_name":         modelName,
	})

	logger.Info("Getting fallback provider")

	// Get model
	model, err := r.modelRepo.GetByModelName(ctx, modelName)
	if err != nil {
		return nil, fmt.Errorf("model not found: %w", err)
	}

	// Get all providers for this model type
	allProviders, err := r.providerRepo.ListEnabled(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list providers: %w", err)
	}

	// Find alternative providers (excluding failed one)
	for _, provider := range allProviders {
		if provider.ID == failedProviderID {
			continue
		}

		// Check if provider has a model with this name
		models, err := r.modelRepo.GetEnabledByProviderID(ctx, provider.ID)
		if err != nil {
			continue
		}

		for _, m := range models {
			if m.ModelName == modelName || m.ModelType == model.ModelType {
				// Found a fallback
				config, err := r.providerConfigRepo.GetByProjectAndProvider(ctx, projectID, provider.ID)
				if err != nil {
					continue
				}

				if !config.IsEnabled {
					continue
				}

				return &gateway.RoutingDecision{
					Provider:         provider.Name,
					ProviderID:       provider.ID,
					Model:            m.ModelName,
					ModelID:          m.ID,
					Strategy:         "fallback",
					EstimatedCost:    0,
					EstimatedLatency: 0,
					RoutingReason:    "Fallback provider after primary failure",
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("no fallback provider available")
}

// HandleProviderFailure handles provider failure and returns alternative
func (r *RoutingService) HandleProviderFailure(ctx context.Context, projectID ulid.ULID, providerID ulid.ULID, errorType gateway.ErrorType) (*gateway.RoutingDecision, error) {
	r.logger.WithFields(logrus.Fields{
		"project_id":  projectID,
		"provider_id": providerID,
		"error_type":  errorType,
	}).Warn("Handling provider failure")

	// For now, just return an error - in production, this would:
	// 1. Update provider health metrics
	// 2. Find alternative provider
	// 3. Update routing weights
	return nil, fmt.Errorf("provider failure handling not fully implemented")
}

// CreateRoutingRule creates a new routing rule
func (r *RoutingService) CreateRoutingRule(ctx context.Context, rule *gateway.RoutingRule) error {
	// TODO: Implement routing rule creation
	return fmt.Errorf("not implemented")
}

// UpdateRoutingRule updates an existing routing rule
func (r *RoutingService) UpdateRoutingRule(ctx context.Context, rule *gateway.RoutingRule) error {
	// TODO: Implement routing rule update
	return fmt.Errorf("not implemented")
}

// DeleteRoutingRule deletes a routing rule
func (r *RoutingService) DeleteRoutingRule(ctx context.Context, ruleID ulid.ULID) error {
	// TODO: Implement routing rule deletion
	return fmt.Errorf("not implemented")
}

// ListProjectRoutingRules lists all routing rules for a project
func (r *RoutingService) ListProjectRoutingRules(ctx context.Context, projectID ulid.ULID) ([]*gateway.RoutingRule, error) {
	// TODO: Implement routing rule listing
	return []*gateway.RoutingRule{}, nil
}

// TestRoute tests a route without executing it
func (r *RoutingService) TestRoute(ctx context.Context, projectID ulid.ULID, request *gateway.RoutingRequest) (*gateway.RouteTestResult, error) {
	logger := r.logger.WithFields(logrus.Fields{
		"project_id": projectID,
		"model_name": request.ModelName,
	})

	logger.Info("Testing route")

	// Get routing decision
	decision, err := r.RouteRequest(ctx, projectID, request)
	if err != nil {
		errMsg := err.Error()
		return &gateway.RouteTestResult{
			TestID:    ulid.New(),
			ProjectID: projectID,
			Request:   request,
			Decision:  nil,
			Success:   false,
			Error:     &errMsg,
			TestedAt:  time.Now(),
		}, nil
	}

	return &gateway.RouteTestResult{
		TestID:    ulid.New(),
		ProjectID: projectID,
		Request:   request,
		Decision:  decision,
		Success:   true,
		TestedAt:  time.Now(),
	}, nil
}

// AnalyzeRoutingPerformance analyzes routing performance over time
func (r *RoutingService) AnalyzeRoutingPerformance(ctx context.Context, projectID ulid.ULID, timeRange *gateway.TimeRange) (*gateway.RoutingAnalysis, error) {
	logger := r.logger.WithFields(logrus.Fields{
		"project_id": projectID,
		"start_time": timeRange.StartTime,
		"end_time":   timeRange.EndTime,
	})

	logger.Info("Analyzing routing performance")

	// TODO: Implement actual analytics retrieval from ClickHouse
	// This would query routing decisions, success rates, costs, latencies

	analysis := &gateway.RoutingAnalysis{
		ProjectID:         projectID,
		TimeRange:         timeRange,
		TotalRequests:     0,
		SuccessRate:       0.0,
		AverageLatency:    0.0,
		AverageCost:       0.0,
		ProviderBreakdown: make(map[string]*gateway.ProviderStats),
		ModelBreakdown:    make(map[string]*gateway.ModelStats),
		StrategyBreakdown: make(map[string]*gateway.StrategyStats),
		FallbackRate:      0.0,
		RetryRate:         0.0,
		GeneratedAt:       time.Now(),
	}

	return analysis, nil
}

// Helper methods for advanced routing strategies

// selectByCost selects provider with lowest cost
func (r *RoutingService) selectByCost(ctx context.Context, projectID ulid.ULID, modelName string) (*gateway.RoutingDecision, error) {
	// Get all providers that support this model
	model, err := r.modelRepo.GetByModelName(ctx, modelName)
	if err != nil {
		return nil, err
	}

	// Get models with same type
	similarModels, err := r.modelRepo.GetByModelType(ctx, model.ModelType, 100, 0)
	if err != nil {
		return nil, err
	}

	// Sort by cost
	sort.Slice(similarModels, func(i, j int) bool {
		return similarModels[i].InputCostPer1kTokens < similarModels[j].InputCostPer1kTokens
	})

	if len(similarModels) == 0 {
		return nil, fmt.Errorf("no models available")
	}

	// Use cheapest model
	cheapestModel := similarModels[0]
	provider, err := r.providerRepo.GetByID(ctx, cheapestModel.ProviderID)
	if err != nil {
		return nil, err
	}

	return &gateway.RoutingDecision{
		Provider:         provider.Name,
		ProviderID:       provider.ID,
		Model:            cheapestModel.ModelName,
		ModelID:          cheapestModel.ID,
		Strategy:         "cost_optimized",
		EstimatedCost:    cheapestModel.InputCostPer1kTokens,
		EstimatedLatency: 0,
		RoutingReason:    "Selected for lowest cost",
	}, nil
}

// selectByLatency selects provider with lowest latency
func (r *RoutingService) selectByLatency(ctx context.Context, projectID ulid.ULID, modelName string) (*gateway.RoutingDecision, error) {
	// TODO: Implement latency-based selection using health metrics
	strategy := gateway.RoutingStrategyLatencyOptimized
	return r.RouteRequest(ctx, projectID, &gateway.RoutingRequest{
		ModelName: modelName,
		Strategy:  &strategy,
	})
}

// selectByQuality selects provider with highest quality score
func (r *RoutingService) selectByQuality(ctx context.Context, projectID ulid.ULID, modelName string) (*gateway.RoutingDecision, error) {
	model, err := r.modelRepo.GetByModelName(ctx, modelName)
	if err != nil {
		return nil, err
	}

	// Get models with same type
	similarModels, err := r.modelRepo.GetByModelType(ctx, model.ModelType, 100, 0)
	if err != nil {
		return nil, err
	}

	// Sort by quality score (if available)
	sort.Slice(similarModels, func(i, j int) bool {
		qi := 0.0
		qj := 0.0
		if similarModels[i].QualityScore != nil {
			qi = *similarModels[i].QualityScore
		}
		if similarModels[j].QualityScore != nil {
			qj = *similarModels[j].QualityScore
		}
		return qi > qj
	})

	if len(similarModels) == 0 {
		return nil, fmt.Errorf("no models available")
	}

	bestModel := similarModels[0]
	provider, err := r.providerRepo.GetByID(ctx, bestModel.ProviderID)
	if err != nil {
		return nil, err
	}

	return &gateway.RoutingDecision{
		Provider:         provider.Name,
		ProviderID:       provider.ID,
		Model:            bestModel.ModelName,
		ModelID:          bestModel.ID,
		Strategy:         "quality_optimized",
		EstimatedCost:    bestModel.InputCostPer1kTokens,
		EstimatedLatency: 0,
		RoutingReason:    "Selected for highest quality",
	}, nil
}
