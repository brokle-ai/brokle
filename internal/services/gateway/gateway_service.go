package gateway

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"brokle/internal/core/domain/gateway"
	"brokle/internal/infrastructure/providers"
	"brokle/pkg/ulid"
)

// GatewayService implements the main gateway orchestration service
type GatewayService struct {
	providerRepo       gateway.ProviderRepository
	modelRepo          gateway.ModelRepository
	providerConfigRepo gateway.ProviderConfigRepository
	routingService     gateway.RoutingService
	costService        gateway.CostService
	providerFactory    providers.Factory
	logger             *logrus.Logger
}

// NewGatewayService creates a new gateway service instance
func NewGatewayService(
	providerRepo gateway.ProviderRepository,
	modelRepo gateway.ModelRepository,
	providerConfigRepo gateway.ProviderConfigRepository,
	routingService gateway.RoutingService,
	costService gateway.CostService,
	providerFactory providers.Factory,
	logger *logrus.Logger,
) gateway.GatewayService {
	return &GatewayService{
		providerRepo:       providerRepo,
		modelRepo:          modelRepo,
		providerConfigRepo: providerConfigRepo,
		routingService:     routingService,
		costService:        costService,
		providerFactory:    providerFactory,
		logger:             logger,
	}
}

// ProcessChatCompletion processes a chat completion request through the gateway
func (s *GatewayService) ProcessChatCompletion(ctx context.Context, req *gateway.ChatCompletionRequest) (*gateway.ChatCompletionResponse, error) {
	requestID := ulid.New()
	startTime := time.Now()

	logger := s.logger.WithFields(logrus.Fields{
		"request_id":      requestID,
		"organization_id": req.OrganizationID,
		"model":           req.Model,
		"stream":          req.Stream,
	})

	logger.Info("Processing chat completion request")

	// 1. Resolve model and provider
	model, provider, config, err := s.resolveModelAndProvider(ctx, req.Model, req.OrganizationID, req.Environment)
	if err != nil {
		logger.WithError(err).Error("Failed to resolve model and provider")
		return nil, fmt.Errorf("failed to resolve model and provider: %w", err)
	}

	logger = logger.WithFields(logrus.Fields{
		"provider_id":   provider.ID,
		"provider_name": provider.Name,
		"model_id":      model.ID,
	})

	// 2. Calculate estimated cost
	estimatedCost, err := s.costService.EstimateChatCompletionCost(ctx, &gateway.CostEstimationRequest{
		Model:          model,
		InputTokens:    s.estimateTokens(req.Messages),
		MaxTokens:      req.MaxTokens,
		OrganizationID: req.OrganizationID,
	})
	if err != nil {
		logger.WithError(err).Warn("Failed to estimate cost")
		// Continue processing even if cost estimation fails
	}

	// 3. Get provider client
	providerClient, err := s.providerFactory.GetProvider(ctx, provider.Type, config.ConfigData)
	if err != nil {
		logger.WithError(err).Error("Failed to get provider client")
		return nil, fmt.Errorf("failed to get provider client: %w", err)
	}

	// 4. Transform request to provider format
	providerRequest := s.transformChatCompletionRequest(req, model, provider)

	// 5. Execute request with provider
	var response *gateway.ChatCompletionResponse
	var actualCost *gateway.CostCalculation

	if req.Stream {
		// Handle streaming response
		response, actualCost, err = s.processChatCompletionStreaming(ctx, providerClient, providerRequest, model, requestID)
	} else {
		// Handle non-streaming response
		response, actualCost, err = s.processChatCompletionSync(ctx, providerClient, providerRequest, model, requestID)
	}

	duration := time.Since(startTime)

	if err != nil {
		logger.WithError(err).Error("Failed to process chat completion")
		// Log failed request for analytics
		s.logRequestMetrics(ctx, &gateway.RequestMetrics{
			RequestID:      requestID,
			OrganizationID: req.OrganizationID,
			Environment:    req.Environment,
			ProviderID:     provider.ID,
			ModelID:        model.ID,
			RequestType:    gateway.RequestTypeChatCompletion,
			Status:         "error",
			Duration:       duration,
			Error:          err.Error(),
			EstimatedCost:  estimatedCost,
		})
		return nil, err
	}

	// 6. Log successful request metrics
	s.logRequestMetrics(ctx, &gateway.RequestMetrics{
		RequestID:       requestID,
		OrganizationID:  req.OrganizationID,
		Environment:     req.Environment,
		ProviderID:      provider.ID,
		ModelID:         model.ID,
		RequestType:     gateway.RequestTypeChatCompletion,
		Status:          "success",
		Duration:        duration,
		InputTokens:     response.Usage.PromptTokens,
		OutputTokens:    response.Usage.CompletionTokens,
		TotalTokens:     response.Usage.TotalTokens,
		EstimatedCost:   estimatedCost,
		ActualCost:      actualCost,
		ResponseLength:  len(response.Choices),
	})

	logger.WithFields(logrus.Fields{
		"duration":        duration,
		"input_tokens":    response.Usage.PromptTokens,
		"output_tokens":   response.Usage.CompletionTokens,
		"total_tokens":    response.Usage.TotalTokens,
		"estimated_cost":  estimatedCost,
		"actual_cost":     actualCost,
	}).Info("Chat completion request completed successfully")

	return response, nil
}

// ProcessCompletion processes a text completion request through the gateway
func (s *GatewayService) ProcessCompletion(ctx context.Context, req *gateway.CompletionRequest) (*gateway.CompletionResponse, error) {
	requestID := ulid.New()
	startTime := time.Now()

	logger := s.logger.WithFields(logrus.Fields{
		"request_id":      requestID,
		"organization_id": req.OrganizationID,
		"model":           req.Model,
		"stream":          req.Stream,
	})

	logger.Info("Processing completion request")

	// 1. Resolve model and provider
	model, provider, config, err := s.resolveModelAndProvider(ctx, req.Model, req.OrganizationID, req.Environment)
	if err != nil {
		logger.WithError(err).Error("Failed to resolve model and provider")
		return nil, fmt.Errorf("failed to resolve model and provider: %w", err)
	}

	// 2. Check if provider supports completion
	if !s.providerSupportsCompletion(provider) {
		logger.Error("Provider does not support completion")
		return nil, gateway.ErrUnsupportedOperation
	}

	// 3. Get provider client and process request
	providerClient, err := s.providerFactory.GetProvider(ctx, provider.Type, config.ConfigData)
	if err != nil {
		logger.WithError(err).Error("Failed to get provider client")
		return nil, fmt.Errorf("failed to get provider client: %w", err)
	}

	// 4. Transform and execute request
	providerRequest := s.transformCompletionRequest(req, model, provider)
	
	var response *gateway.CompletionResponse
	var actualCost *gateway.CostCalculation

	if req.Stream {
		response, actualCost, err = s.processCompletionStreaming(ctx, providerClient, providerRequest, model, requestID)
	} else {
		response, actualCost, err = s.processCompletionSync(ctx, providerClient, providerRequest, model, requestID)
	}

	duration := time.Since(startTime)

	if err != nil {
		logger.WithError(err).Error("Failed to process completion")
		return nil, err
	}

	logger.WithFields(logrus.Fields{
		"duration":     duration,
		"total_tokens": response.Usage.TotalTokens,
		"actual_cost":  actualCost,
	}).Info("Completion request completed successfully")

	return response, nil
}

// ProcessEmbeddings processes an embeddings request through the gateway
func (s *GatewayService) ProcessEmbeddings(ctx context.Context, req *gateway.EmbeddingsRequest) (*gateway.EmbeddingsResponse, error) {
	requestID := ulid.New()
	startTime := time.Now()

	logger := s.logger.WithFields(logrus.Fields{
		"request_id":      requestID,
		"organization_id": req.OrganizationID,
		"model":           req.Model,
		"input_count":     len(req.Input),
	})

	logger.Info("Processing embeddings request")

	// 1. Resolve model and provider
	model, provider, config, err := s.resolveModelAndProvider(ctx, req.Model, req.OrganizationID, req.Environment)
	if err != nil {
		logger.WithError(err).Error("Failed to resolve model and provider")
		return nil, fmt.Errorf("failed to resolve model and provider: %w", err)
	}

	// 2. Check if provider supports embeddings
	if !s.providerSupportsEmbeddings(provider) {
		logger.Error("Provider does not support embeddings")
		return nil, gateway.ErrUnsupportedOperation
	}

	// 3. Get provider client and process request
	providerClient, err := s.providerFactory.GetProvider(ctx, provider.Type, config.ConfigData)
	if err != nil {
		logger.WithError(err).Error("Failed to get provider client")
		return nil, fmt.Errorf("failed to get provider client: %w", err)
	}

	// 4. Transform and execute request
	providerRequest := s.transformEmbeddingsRequest(req, model, provider)
	response, actualCost, err := s.processEmbeddingsSync(ctx, providerClient, providerRequest, model, requestID)

	duration := time.Since(startTime)

	if err != nil {
		logger.WithError(err).Error("Failed to process embeddings")
		return nil, err
	}

	logger.WithFields(logrus.Fields{
		"duration":         duration,
		"total_tokens":     response.Usage.TotalTokens,
		"embeddings_count": len(response.Data),
		"actual_cost":      actualCost,
	}).Info("Embeddings request completed successfully")

	return response, nil
}

// GetAvailableModels returns a list of available models
func (s *GatewayService) GetAvailableModels(ctx context.Context, req *gateway.ListModelsRequest) (*gateway.ListModelsResponse, error) {
	logger := s.logger.WithFields(logrus.Fields{
		"organization_id": req.OrganizationID,
		"environment":     req.Environment,
	})

	logger.Info("Fetching available models")

	// Get active providers for the organization
	providers, err := s.providerRepo.ListEnabled(ctx)
	if err != nil {
		logger.WithError(err).Error("Failed to fetch active providers")
		return nil, fmt.Errorf("failed to fetch active providers: %w", err)
	}

	var allModels []gateway.ModelInfo
	for _, provider := range providers {
		// Get provider configuration
		config, err := s.providerConfigRepo.GetByProviderOrgAndEnv(ctx, provider.ID, req.OrganizationID, req.Environment)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"provider_id": provider.ID,
				"error":       err,
			}).Warn("No configuration found for provider, skipping")
			continue
		}

		if !config.IsActive {
			continue
		}

		// Get models for this provider
		models, err := s.modelRepo.ListByProviderAndStatus(ctx, provider.ID, true)
		if err != nil {
			logger.WithError(err).WithField("provider_id", provider.ID).Error("Failed to fetch models for provider")
			continue
		}

		// Convert to response format
		for _, model := range models {
			modelInfo := gateway.ModelInfo{
				ID:           model.Name,
				Object:       "model",
				Created:      model.CreatedAt.Unix(),
				OwnedBy:      provider.Name,
				Provider:     provider.Name,
				Type:         string(model.Type),
				ContextLength: model.ContextLength,
				MaxTokens:    model.MaxTokens,
				InputCost:    model.InputCostPerToken,
				OutputCost:   model.OutputCostPerToken,
				Capabilities: model.Capabilities,
			}
			allModels = append(allModels, modelInfo)
		}
	}

	logger.WithField("model_count", len(allModels)).Info("Successfully fetched available models")

	return &gateway.ListModelsResponse{
		Object: "list",
		Data:   allModels,
	}, nil
}

// CheckHealth performs a health check on all active providers
func (s *GatewayService) CheckHealth(ctx context.Context, req *gateway.HealthCheckRequest) (*gateway.HealthCheckResponse, error) {
	logger := s.logger.WithField("organization_id", req.OrganizationID)
	logger.Info("Performing gateway health check")

	// Get active providers
	providers, err := s.providerRepo.ListEnabled(ctx)
	if err != nil {
		logger.WithError(err).Error("Failed to fetch active providers")
		return nil, fmt.Errorf("failed to fetch active providers: %w", err)
	}

	var providerHealth []gateway.ProviderHealth
	allHealthy := true

	for _, provider := range providers {
		// Get provider configuration
		config, err := s.providerConfigRepo.GetByProviderOrgAndEnv(ctx, provider.ID, req.OrganizationID, req.Environment)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"provider_id": provider.ID,
				"error":       err,
			}).Warn("No configuration found for provider")

			providerHealth = append(providerHealth, gateway.ProviderHealth{
				ProviderID:   provider.ID,
				ProviderName: provider.Name,
				Status:       gateway.HealthStatusUnhealthy,
				Error:        "No configuration found",
				LastChecked:  time.Now(),
			})
			allHealthy = false
			continue
		}

		// Get provider client and check health
		providerClient, err := s.providerFactory.GetProvider(ctx, provider.Type, config.ConfigData)
		if err != nil {
			logger.WithError(err).WithField("provider_id", provider.ID).Error("Failed to get provider client")
			
			providerHealth = append(providerHealth, gateway.ProviderHealth{
				ProviderID:   provider.ID,
				ProviderName: provider.Name,
				Status:       gateway.HealthStatusUnhealthy,
				Error:        err.Error(),
				LastChecked:  time.Now(),
			})
			allHealthy = false
			continue
		}

		// Perform health check
		healthStatus, err := providerClient.HealthCheck(ctx)
		status := gateway.HealthStatusHealthy
		errorMsg := ""

		if err != nil || !healthStatus {
			status = gateway.HealthStatusUnhealthy
			errorMsg = ""
			if err != nil {
				errorMsg = err.Error()
			}
			allHealthy = false
		}

		providerHealth = append(providerHealth, gateway.ProviderHealth{
			ProviderID:   provider.ID,
			ProviderName: provider.Name,
			Status:       status,
			Error:        errorMsg,
			LastChecked:  time.Now(),
		})
	}

	overallStatus := gateway.HealthStatusHealthy
	if !allHealthy {
		overallStatus = gateway.HealthStatusDegraded
	}
	if len(providerHealth) == 0 {
		overallStatus = gateway.HealthStatusUnhealthy
	}

	response := &gateway.HealthCheckResponse{
		Status:           overallStatus,
		Timestamp:        time.Now(),
		ProviderHealth:   providerHealth,
		TotalProviders:   len(providers),
		HealthyProviders: 0,
	}

	// Count healthy providers
	for _, health := range providerHealth {
		if health.Status == gateway.HealthStatusHealthy {
			response.HealthyProviders++
		}
	}

	logger.WithFields(logrus.Fields{
		"overall_status":    overallStatus,
		"total_providers":   response.TotalProviders,
		"healthy_providers": response.HealthyProviders,
	}).Info("Gateway health check completed")

	return response, nil
}

// Helper methods

func (s *GatewayService) resolveModelAndProvider(ctx context.Context, modelName string, orgID ulid.ULID, env gateway.Environment) (*gateway.Model, *gateway.Provider, *gateway.ProviderConfig, error) {
	// First, try to find the model by name
	model, err := s.modelRepo.GetByName(ctx, modelName)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("model not found: %w", err)
	}

	// Get the provider for this model
	provider, err := s.providerRepo.GetByID(ctx, model.ProviderID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("provider not found: %w", err)
	}

	// Check if provider is enabled
	if !provider.IsEnabled {
		return nil, nil, nil, gateway.ErrProviderDisabled
	}

	// Get provider configuration for the organization and environment
	config, err := s.providerConfigRepo.GetByProviderOrgAndEnv(ctx, provider.ID, orgID, env)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("provider configuration not found: %w", err)
	}

	// Check if configuration is active
	if !config.IsActive {
		return nil, nil, nil, gateway.ErrProviderConfigInactive
	}

	return model, provider, config, nil
}

func (s *GatewayService) estimateTokens(messages []gateway.ChatMessage) int32 {
	// Simple token estimation - in production, use a proper tokenizer
	totalChars := 0
	for _, msg := range messages {
		totalChars += len(msg.Content)
	}
	// Rough approximation: 4 characters per token
	return int32(totalChars / 4)
}

func (s *GatewayService) providerSupportsCompletion(provider *gateway.Provider) bool {
	features, ok := provider.SupportedFeatures["completion"]
	return ok && features == true
}

func (s *GatewayService) providerSupportsEmbeddings(provider *gateway.Provider) bool {
	features, ok := provider.SupportedFeatures["embeddings"]
	return ok && features == true
}

func (s *GatewayService) logRequestMetrics(ctx context.Context, metrics *gateway.RequestMetrics) {
	// This would typically send metrics to an analytics service
	// For now, just log the metrics
	s.logger.WithFields(logrus.Fields{
		"request_id":       metrics.RequestID,
		"organization_id":  metrics.OrganizationID,
		"provider_id":      metrics.ProviderID,
		"model_id":         metrics.ModelID,
		"request_type":     metrics.RequestType,
		"status":           metrics.Status,
		"duration":         metrics.Duration,
		"input_tokens":     metrics.InputTokens,
		"output_tokens":    metrics.OutputTokens,
		"total_tokens":     metrics.TotalTokens,
		"estimated_cost":   metrics.EstimatedCost,
		"actual_cost":      metrics.ActualCost,
		"response_length":  metrics.ResponseLength,
	}).Info("Request metrics logged")
}

// Request transformation methods (to be implemented based on provider interfaces)

func (s *GatewayService) transformChatCompletionRequest(req *gateway.ChatCompletionRequest, model *gateway.Model, provider *gateway.Provider) interface{} {
	// Transform gateway request to provider-specific format
	// Implementation depends on provider interface
	return req
}

func (s *GatewayService) transformCompletionRequest(req *gateway.CompletionRequest, model *gateway.Model, provider *gateway.Provider) interface{} {
	// Transform gateway request to provider-specific format
	return req
}

func (s *GatewayService) transformEmbeddingsRequest(req *gateway.EmbeddingsRequest, model *gateway.Model, provider *gateway.Provider) interface{} {
	// Transform gateway request to provider-specific format
	return req
}

// Processing methods (to be implemented based on provider interfaces)

func (s *GatewayService) processChatCompletionSync(ctx context.Context, client providers.Provider, req interface{}, model *gateway.Model, requestID ulid.ULID) (*gateway.ChatCompletionResponse, *gateway.CostCalculation, error) {
	// Implementation depends on provider interface
	// This is a placeholder
	return nil, nil, fmt.Errorf("not implemented")
}

func (s *GatewayService) processChatCompletionStreaming(ctx context.Context, client providers.Provider, req interface{}, model *gateway.Model, requestID ulid.ULID) (*gateway.ChatCompletionResponse, *gateway.CostCalculation, error) {
	// Implementation depends on provider interface
	return nil, nil, fmt.Errorf("not implemented")
}

func (s *GatewayService) processCompletionSync(ctx context.Context, client providers.Provider, req interface{}, model *gateway.Model, requestID ulid.ULID) (*gateway.CompletionResponse, *gateway.CostCalculation, error) {
	// Implementation depends on provider interface
	return nil, nil, fmt.Errorf("not implemented")
}

func (s *GatewayService) processCompletionStreaming(ctx context.Context, client providers.Provider, req interface{}, model *gateway.Model, requestID ulid.ULID) (*gateway.CompletionResponse, *gateway.CostCalculation, error) {
	// Implementation depends on provider interface
	return nil, nil, fmt.Errorf("not implemented")
}

func (s *GatewayService) processEmbeddingsSync(ctx context.Context, client providers.Provider, req interface{}, model *gateway.Model, requestID ulid.ULID) (*gateway.EmbeddingsResponse, *gateway.CostCalculation, error) {
	// Implementation depends on provider interface
	return nil, nil, fmt.Errorf("not implemented")
}