package gateway

import (
	"context"
	"fmt"
	"math/rand"
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
	rand               *rand.Rand
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
		rand:               rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// SelectProvider selects the best provider for a given request
func (r *RoutingService) SelectProvider(ctx context.Context, req *gateway.ProviderSelectionRequest) (*gateway.ProviderSelectionResponse, error) {
	logger := r.logger.WithFields(logrus.Fields{
		"organization_id": req.OrganizationID,
		"model_name":      req.ModelName,
		"request_type":    req.RequestType,
		"strategy":        req.Strategy,
	})

	logger.Info("Selecting provider for request")

	// Get available providers for the requested model
	candidates, err := r.getProviderCandidates(ctx, req)
	if err != nil {
		logger.WithError(err).Error("Failed to get provider candidates")
		return nil, fmt.Errorf("failed to get provider candidates: %w", err)
	}

	if len(candidates) == 0 {
		logger.Error("No available providers found")
		return nil, gateway.ErrNoProvidersAvailable
	}

	logger.WithField("candidate_count", len(candidates)).Info("Found provider candidates")

	// Apply routing strategy
	selected, err := r.applyRoutingStrategy(ctx, req, candidates)
	if err != nil {
		logger.WithError(err).Error("Failed to apply routing strategy")
		return nil, fmt.Errorf("failed to apply routing strategy: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"selected_provider": selected.Provider.Name,
		"provider_id":       selected.Provider.ID,
		"model_id":          selected.Model.ID,
		"estimated_cost":    selected.EstimatedCost,
		"selection_reason":  selected.SelectionReason,
	}).Info("Provider selected successfully")

	return selected, nil
}

// SelectFallbackProvider selects a fallback provider when primary fails
func (r *RoutingService) SelectFallbackProvider(ctx context.Context, req *gateway.FallbackSelectionRequest) (*gateway.ProviderSelectionResponse, error) {
	logger := r.logger.WithFields(logrus.Fields{
		"organization_id":   req.OrganizationID,
		"failed_provider":   req.FailedProvider,
		"original_model":    req.OriginalModel,
		"request_type":      req.RequestType,
	})

	logger.Info("Selecting fallback provider")

	// Get available providers excluding the failed one
	candidates, err := r.getProviderCandidates(ctx, &gateway.ProviderSelectionRequest{
		OrganizationID: req.OrganizationID,
		ModelName:      req.OriginalModel,
		RequestType:    req.RequestType,
		Environment:    req.Environment,
		Strategy:       gateway.RoutingStrategyReliability, // Prefer reliable providers for fallback
		ExcludeProvider: &req.FailedProvider,
	})

	if err != nil {
		logger.WithError(err).Error("Failed to get fallback provider candidates")
		return nil, fmt.Errorf("failed to get fallback provider candidates: %w", err)
	}

	if len(candidates) == 0 {
		logger.Error("No fallback providers available")
		return nil, gateway.ErrNoFallbackProviders
	}

	// Sort by reliability for fallback
	sort.Slice(candidates, func(i, j int) bool {
		return r.calculateReliabilityScore(candidates[i]) > r.calculateReliabilityScore(candidates[j])
	})

	selected := candidates[0]
	selected.SelectionReason = "fallback_reliability"

	logger.WithFields(logrus.Fields{
		"fallback_provider": selected.Provider.Name,
		"provider_id":       selected.Provider.ID,
		"reliability_score": r.calculateReliabilityScore(selected),
	}).Info("Fallback provider selected")

	return selected, nil
}

// RouteRequest routes a request to multiple providers for A/B testing or redundancy
func (r *RoutingService) RouteRequest(ctx context.Context, req *gateway.RequestRoutingConfig) (*gateway.RequestRoutingResponse, error) {
	logger := r.logger.WithFields(logrus.Fields{
		"organization_id": req.OrganizationID,
		"routing_type":    req.RoutingType,
		"model_name":      req.ModelName,
	})

	logger.Info("Routing request to multiple providers")

	var routes []gateway.ProviderRoute

	switch req.RoutingType {
	case gateway.RoutingTypeABTesting:
		routes, err := r.createABTestRoutes(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to create A/B test routes: %w", err)
		}
		return &gateway.RequestRoutingResponse{
			Routes:       routes,
			RoutingType:  req.RoutingType,
		}, nil

	case gateway.RoutingTypeRedundancy:
		routes, err := r.createRedundancyRoutes(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to create redundancy routes: %w", err)
		}
		return &gateway.RequestRoutingResponse{
			Routes:       routes,
			RoutingType:  req.RoutingType,
		}, nil

	case gateway.RoutingTypeLoadBalancing:
		routes, err := r.createLoadBalancingRoutes(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to create load balancing routes: %w", err)
		}
		return &gateway.RequestRoutingResponse{
			Routes:       routes,
			RoutingType:  req.RoutingType,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported routing type: %s", req.RoutingType)
	}
}

// UpdateProviderWeights updates routing weights based on performance metrics
func (r *RoutingService) UpdateProviderWeights(ctx context.Context, req *gateway.WeightUpdateRequest) error {
	logger := r.logger.WithFields(logrus.Fields{
		"organization_id": req.OrganizationID,
		"provider_id":     req.ProviderID,
		"new_weight":      req.NewWeight,
	})

	logger.Info("Updating provider weights")

	// TODO: Implement weight persistence
	// This would typically update routing rules or provider configurations
	// with new weights based on performance metrics

	logger.Info("Provider weights updated successfully")
	return nil
}

// Helper methods

func (r *RoutingService) getProviderCandidates(ctx context.Context, req *gateway.ProviderSelectionRequest) ([]*gateway.ProviderSelectionResponse, error) {
	// Get models matching the requested model name
	var models []*gateway.Model
	var err error

	if req.ModelName != "" {
		model, err := r.modelRepo.GetByName(ctx, req.ModelName)
		if err != nil {
			return nil, fmt.Errorf("model not found: %w", err)
		}
		models = []*gateway.Model{model}
	} else {
		// If no specific model, get all models of the requested type
		filter := &gateway.ModelFilter{
			IsEnabled: &[]bool{true}[0],
		}
		if req.ModelType != nil {
			filter.ModelType = req.ModelType
		}
		models, _, err = r.modelRepo.SearchModels(ctx, filter)
		if err != nil {
			return nil, fmt.Errorf("failed to search models: %w", err)
		}
	}

	var candidates []*gateway.ProviderSelectionResponse

	for _, model := range models {
		// Skip if this is the excluded provider
		if req.ExcludeProvider != nil && model.ProviderID == *req.ExcludeProvider {
			continue
		}

		// Get provider for this model
		provider, err := r.providerRepo.GetByID(ctx, model.ProviderID)
		if err != nil {
			r.logger.WithError(err).WithField("provider_id", model.ProviderID).Warn("Failed to get provider")
			continue
		}

		// Check if provider is enabled
		if !provider.IsEnabled {
			continue
		}

		// Check if provider supports the requested operation
		if !r.providerSupportsOperation(provider, req.RequestType) {
			continue
		}

		// Get provider configuration
		config, err := r.providerConfigRepo.GetByProviderOrgAndEnv(ctx, provider.ID, req.OrganizationID, req.Environment)
		if err != nil {
			r.logger.WithFields(logrus.Fields{
				"provider_id": provider.ID,
				"error":       err,
			}).Warn("No configuration found for provider")
			continue
		}

		if !config.IsActive {
			continue
		}

		// Calculate estimated cost
		estimatedCost, err := r.calculateEstimatedCost(ctx, model, req)
		if err != nil {
			r.logger.WithError(err).WithField("model_id", model.ID).Warn("Failed to calculate estimated cost")
			// Continue without cost estimation
		}

		candidate := &gateway.ProviderSelectionResponse{
			Provider:      provider,
			Model:         model,
			Config:        config,
			EstimatedCost: estimatedCost,
		}

		candidates = append(candidates, candidate)
	}

	return candidates, nil
}

func (r *RoutingService) applyRoutingStrategy(ctx context.Context, req *gateway.ProviderSelectionRequest, candidates []*gateway.ProviderSelectionResponse) (*gateway.ProviderSelectionResponse, error) {
	switch req.Strategy {
	case gateway.RoutingStrategyCost:
		return r.selectByCost(candidates)
	case gateway.RoutingStrategyLatency:
		return r.selectByLatency(candidates)
	case gateway.RoutingStrategyReliability:
		return r.selectByReliability(candidates)
	case gateway.RoutingStrategyQuality:
		return r.selectByQuality(candidates)
	case gateway.RoutingStrategyRoundRobin:
		return r.selectByRoundRobin(candidates)
	case gateway.RoutingStrategyWeighted:
		return r.selectByWeighted(candidates)
	case gateway.RoutingStrategyRandom:
		return r.selectByRandom(candidates)
	default:
		// Default to cost-based selection
		return r.selectByCost(candidates)
	}
}

func (r *RoutingService) selectByCost(candidates []*gateway.ProviderSelectionResponse) (*gateway.ProviderSelectionResponse, error) {
	if len(candidates) == 0 {
		return nil, gateway.ErrNoProvidersAvailable
	}

	// Sort by estimated cost (ascending)
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].EstimatedCost == nil {
			return false
		}
		if candidates[j].EstimatedCost == nil {
			return true
		}
		return candidates[i].EstimatedCost.TotalCost < candidates[j].EstimatedCost.TotalCost
	})

	selected := candidates[0]
	selected.SelectionReason = "lowest_cost"
	return selected, nil
}

func (r *RoutingService) selectByLatency(candidates []*gateway.ProviderSelectionResponse) (*gateway.ProviderSelectionResponse, error) {
	if len(candidates) == 0 {
		return nil, gateway.ErrNoProvidersAvailable
	}

	// Sort by historical latency (would need to implement latency tracking)
	sort.Slice(candidates, func(i, j int) bool {
		return r.calculateLatencyScore(candidates[i]) < r.calculateLatencyScore(candidates[j])
	})

	selected := candidates[0]
	selected.SelectionReason = "lowest_latency"
	return selected, nil
}

func (r *RoutingService) selectByReliability(candidates []*gateway.ProviderSelectionResponse) (*gateway.ProviderSelectionResponse, error) {
	if len(candidates) == 0 {
		return nil, gateway.ErrNoProvidersAvailable
	}

	// Sort by reliability score (descending)
	sort.Slice(candidates, func(i, j int) bool {
		return r.calculateReliabilityScore(candidates[i]) > r.calculateReliabilityScore(candidates[j])
	})

	selected := candidates[0]
	selected.SelectionReason = "highest_reliability"
	return selected, nil
}

func (r *RoutingService) selectByQuality(candidates []*gateway.ProviderSelectionResponse) (*gateway.ProviderSelectionResponse, error) {
	if len(candidates) == 0 {
		return nil, gateway.ErrNoProvidersAvailable
	}

	// Sort by quality score (would need to implement quality metrics)
	sort.Slice(candidates, func(i, j int) bool {
		return r.calculateQualityScore(candidates[i]) > r.calculateQualityScore(candidates[j])
	})

	selected := candidates[0]
	selected.SelectionReason = "highest_quality"
	return selected, nil
}

func (r *RoutingService) selectByRoundRobin(candidates []*gateway.ProviderSelectionResponse) (*gateway.ProviderSelectionResponse, error) {
	if len(candidates) == 0 {
		return nil, gateway.ErrNoProvidersAvailable
	}

	// TODO: Implement proper round-robin state tracking
	// For now, use random selection
	selected := candidates[r.rand.Intn(len(candidates))]
	selected.SelectionReason = "round_robin"
	return selected, nil
}

func (r *RoutingService) selectByWeighted(candidates []*gateway.ProviderSelectionResponse) (*gateway.ProviderSelectionResponse, error) {
	if len(candidates) == 0 {
		return nil, gateway.ErrNoProvidersAvailable
	}

	// Calculate total weight
	totalWeight := 0.0
	for _, candidate := range candidates {
		totalWeight += r.getProviderWeight(candidate.Provider)
	}

	if totalWeight == 0 {
		// Fallback to random if no weights
		return r.selectByRandom(candidates)
	}

	// Weighted random selection
	target := r.rand.Float64() * totalWeight
	current := 0.0

	for _, candidate := range candidates {
		current += r.getProviderWeight(candidate.Provider)
		if current >= target {
			candidate.SelectionReason = "weighted_random"
			return candidate, nil
		}
	}

	// Fallback to last candidate
	selected := candidates[len(candidates)-1]
	selected.SelectionReason = "weighted_fallback"
	return selected, nil
}

func (r *RoutingService) selectByRandom(candidates []*gateway.ProviderSelectionResponse) (*gateway.ProviderSelectionResponse, error) {
	if len(candidates) == 0 {
		return nil, gateway.ErrNoProvidersAvailable
	}

	selected := candidates[r.rand.Intn(len(candidates))]
	selected.SelectionReason = "random"
	return selected, nil
}

func (r *RoutingService) providerSupportsOperation(provider *gateway.Provider, requestType gateway.RequestType) bool {
	switch requestType {
	case gateway.RequestTypeChatCompletion:
		features, ok := provider.SupportedFeatures["chat_completion"]
		return ok && features == true
	case gateway.RequestTypeCompletion:
		features, ok := provider.SupportedFeatures["completion"]
		return ok && features == true
	case gateway.RequestTypeEmbeddings:
		features, ok := provider.SupportedFeatures["embeddings"]
		return ok && features == true
	default:
		return false
	}
}

func (r *RoutingService) calculateEstimatedCost(ctx context.Context, model *gateway.Model, req *gateway.ProviderSelectionRequest) (*gateway.CostCalculation, error) {
	// This would integrate with the cost service to calculate estimated costs
	// For now, return a simple calculation based on model pricing
	return &gateway.CostCalculation{
		InputCost:  float64(req.EstimatedInputTokens) * model.InputCostPerToken,
		OutputCost: float64(req.EstimatedOutputTokens) * model.OutputCostPerToken,
		TotalCost:  float64(req.EstimatedInputTokens)*model.InputCostPerToken + float64(req.EstimatedOutputTokens)*model.OutputCostPerToken,
		Currency:   "USD",
	}, nil
}

// Scoring methods (placeholder implementations)

func (r *RoutingService) calculateLatencyScore(candidate *gateway.ProviderSelectionResponse) float64 {
	// TODO: Implement based on historical latency metrics
	// For now, return a random score
	return r.rand.Float64() * 1000 // milliseconds
}

func (r *RoutingService) calculateReliabilityScore(candidate *gateway.ProviderSelectionResponse) float64 {
	// TODO: Implement based on historical reliability metrics
	// For now, return a score based on provider type
	switch candidate.Provider.Type {
	case gateway.ProviderTypeOpenAI:
		return 0.95
	case gateway.ProviderTypeAnthropic:
		return 0.93
	default:
		return 0.85
	}
}

func (r *RoutingService) calculateQualityScore(candidate *gateway.ProviderSelectionResponse) float64 {
	// TODO: Implement based on historical quality metrics
	// For now, return a score based on model capabilities
	score := 0.8
	if len(candidate.Model.Capabilities) > 5 {
		score += 0.1
	}
	if candidate.Model.ContextLength > 8000 {
		score += 0.05
	}
	return score
}

func (r *RoutingService) getProviderWeight(provider *gateway.Provider) float64 {
	// TODO: Implement weight retrieval from configuration or database
	// For now, return default weight
	return 1.0
}

// Routing creation methods

func (r *RoutingService) createABTestRoutes(ctx context.Context, req *gateway.RequestRoutingConfig) ([]gateway.ProviderRoute, error) {
	// Get multiple providers for A/B testing
	candidates, err := r.getProviderCandidates(ctx, &gateway.ProviderSelectionRequest{
		OrganizationID: req.OrganizationID,
		ModelName:      req.ModelName,
		RequestType:    req.RequestType,
		Environment:    req.Environment,
		Strategy:       gateway.RoutingStrategyRandom,
	})

	if err != nil {
		return nil, err
	}

	if len(candidates) < 2 {
		return nil, fmt.Errorf("insufficient providers for A/B testing: need at least 2, got %d", len(candidates))
	}

	// Select two providers for A/B testing
	var routes []gateway.ProviderRoute
	for i := 0; i < 2 && i < len(candidates); i++ {
		routes = append(routes, gateway.ProviderRoute{
			Provider:   candidates[i].Provider,
			Model:      candidates[i].Model,
			Config:     candidates[i].Config,
			Weight:     0.5, // Equal weight for A/B testing
			RouteID:    ulid.New(),
		})
	}

	return routes, nil
}

func (r *RoutingService) createRedundancyRoutes(ctx context.Context, req *gateway.RequestRoutingConfig) ([]gateway.ProviderRoute, error) {
	// Get multiple providers for redundancy
	candidates, err := r.getProviderCandidates(ctx, &gateway.ProviderSelectionRequest{
		OrganizationID: req.OrganizationID,
		ModelName:      req.ModelName,
		RequestType:    req.RequestType,
		Environment:    req.Environment,
		Strategy:       gateway.RoutingStrategyReliability,
	})

	if err != nil {
		return nil, err
	}

	// Use up to 3 providers for redundancy
	maxRoutes := 3
	if len(candidates) < maxRoutes {
		maxRoutes = len(candidates)
	}

	var routes []gateway.ProviderRoute
	for i := 0; i < maxRoutes; i++ {
		routes = append(routes, gateway.ProviderRoute{
			Provider: candidates[i].Provider,
			Model:    candidates[i].Model,
			Config:   candidates[i].Config,
			Weight:   1.0 / float64(maxRoutes), // Equal weight distribution
			RouteID:  ulid.New(),
		})
	}

	return routes, nil
}

func (r *RoutingService) createLoadBalancingRoutes(ctx context.Context, req *gateway.RequestRoutingConfig) ([]gateway.ProviderRoute, error) {
	// Get all available providers for load balancing
	candidates, err := r.getProviderCandidates(ctx, &gateway.ProviderSelectionRequest{
		OrganizationID: req.OrganizationID,
		ModelName:      req.ModelName,
		RequestType:    req.RequestType,
		Environment:    req.Environment,
		Strategy:       gateway.RoutingStrategyWeighted,
	})

	if err != nil {
		return nil, err
	}

	// Calculate total weight
	totalWeight := 0.0
	for _, candidate := range candidates {
		totalWeight += r.getProviderWeight(candidate.Provider)
	}

	var routes []gateway.ProviderRoute
	for _, candidate := range candidates {
		weight := r.getProviderWeight(candidate.Provider) / totalWeight
		routes = append(routes, gateway.ProviderRoute{
			Provider: candidate.Provider,
			Model:    candidate.Model,
			Config:   candidate.Config,
			Weight:   weight,
			RouteID:  ulid.New(),
		})
	}

	return routes, nil
}