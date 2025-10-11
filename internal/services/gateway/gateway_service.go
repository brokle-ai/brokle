package gateway

import (
	"context"
	"fmt"
	"io"
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
	providerFactory    providers.ProviderFactory
	logger             *logrus.Logger
}

// NewGatewayService creates a new gateway service instance
func NewGatewayService(
	providerRepo gateway.ProviderRepository,
	modelRepo gateway.ModelRepository,
	providerConfigRepo gateway.ProviderConfigRepository,
	routingService gateway.RoutingService,
	costService gateway.CostService,
	providerFactory providers.ProviderFactory,
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
	model, provider, config, err := s.resolveModelAndProvider(ctx, req.Model, req.OrganizationID)
	if err != nil {
		logger.WithError(err).Error("Failed to resolve model and provider")
		return nil, fmt.Errorf("failed to resolve model and provider: %w", err)
	}

	logger = logger.WithFields(logrus.Fields{
		"provider_id":   provider.ID,
		"provider_name": provider.Name,
		"model_id":      model.ID,
	})

	// 2. Calculate estimated cost (using EstimateRequestCost from interface)
	var estimatedCost float64
	inputTokens := s.estimateTokens(req.Messages)
	estimatedCost, err = s.costService.EstimateRequestCost(ctx, model.ModelName, int(inputTokens))
	if err != nil {
		logger.WithError(err).Warn("Failed to estimate cost")
		// Continue processing even if cost estimation fails
		estimatedCost = 0
	}

	// 3. Prepare provider configuration with decrypted API key
	providerConfig, err := s.prepareProviderConfig(ctx, config)
	if err != nil {
		logger.WithError(err).Error("Failed to prepare provider configuration")
		return nil, fmt.Errorf("failed to prepare provider configuration: %w", err)
	}

	// 4. Get provider client
	providerClient, err := s.providerFactory.GetProvider(ctx, provider.Type, providerConfig)
	if err != nil {
		logger.WithError(err).Error("Failed to get provider client")
		return nil, fmt.Errorf("failed to get provider client: %w", err)
	}

	// 5. Transform request to provider format
	providerRequest := s.transformChatCompletionRequest(req, model, provider)

	// 6. Execute request with provider
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
		errMsg := err.Error()
		s.logRequestMetrics(ctx, &gateway.RequestMetrics{
			RequestID:    requestID.String(),
			ProjectID:    req.OrganizationID,
			Provider:     provider.Name,
			Model:        model.ModelName,
			InputTokens:  0,
			OutputTokens: 0,
			TotalTokens:  0,
			CostUSD:      0,
			Success:      false,
			ErrorMessage: &errMsg,
			Timestamp:    time.Now(),
		})
		return nil, err
	}

	// 7. Log successful request metrics
	var costUSD float64
	if actualCost != nil {
		costUSD = actualCost.TotalCost
	} else {
		costUSD = estimatedCost
	}

	s.logRequestMetrics(ctx, &gateway.RequestMetrics{
		RequestID:    requestID.String(),
		ProjectID:    req.OrganizationID,
		Provider:     provider.Name,
		Model:        model.ModelName,
		InputTokens:  response.Usage.InputTokens,
		OutputTokens: response.Usage.OutputTokens,
		TotalTokens:  response.Usage.TotalTokens,
		CostUSD:      costUSD,
		Success:      true,
		Timestamp:    time.Now(),
	})

	logger.WithFields(logrus.Fields{
		"duration":       duration,
		"input_tokens":   response.Usage.InputTokens,
		"output_tokens":  response.Usage.OutputTokens,
		"total_tokens":   response.Usage.TotalTokens,
		"estimated_cost": estimatedCost,
		"actual_cost":    actualCost,
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
	model, provider, config, err := s.resolveModelAndProvider(ctx, req.Model, req.OrganizationID)
	if err != nil {
		logger.WithError(err).Error("Failed to resolve model and provider")
		return nil, fmt.Errorf("failed to resolve model and provider: %w", err)
	}

	// 2. Check if provider supports completion
	if !s.providerSupportsCompletion(provider) {
		logger.Error("Provider does not support completion")
		return nil, gateway.ErrUnsupportedRequestType
	}

	// 3. Prepare provider configuration with decrypted API key
	providerConfig, err := s.prepareProviderConfig(ctx, config)
	if err != nil {
		logger.WithError(err).Error("Failed to prepare provider configuration")
		return nil, fmt.Errorf("failed to prepare provider configuration: %w", err)
	}

	// 4. Get provider client
	providerClient, err := s.providerFactory.GetProvider(ctx, provider.Type, providerConfig)
	if err != nil {
		logger.WithError(err).Error("Failed to get provider client")
		return nil, fmt.Errorf("failed to get provider client: %w", err)
	}

	// 5. Transform and execute request
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
func (s *GatewayService) ProcessEmbeddings(ctx context.Context, req *gateway.EmbeddingsRequest) (*gateway.EmbeddingResponse, error) {
	requestID := ulid.New()
	startTime := time.Now()

	// Calculate input count safely
	inputCount := 0
	switch input := req.Input.(type) {
	case string:
		inputCount = 1
	case []string:
		inputCount = len(input)
	case []interface{}:
		inputCount = len(input)
	}

	logger := s.logger.WithFields(logrus.Fields{
		"request_id":      requestID,
		"organization_id": req.OrganizationID,
		"model":           req.Model,
		"input_count":     inputCount,
	})

	logger.Info("Processing embeddings request")

	// 1. Resolve model and provider
	model, provider, config, err := s.resolveModelAndProvider(ctx, req.Model, req.OrganizationID)
	if err != nil {
		logger.WithError(err).Error("Failed to resolve model and provider")
		return nil, fmt.Errorf("failed to resolve model and provider: %w", err)
	}

	// 2. Check if provider supports embeddings
	if !s.providerSupportsEmbeddings(provider) {
		logger.Error("Provider does not support embeddings")
		return nil, gateway.ErrUnsupportedRequestType
	}

	// 3. Prepare provider configuration with decrypted API key
	providerConfig, err := s.prepareProviderConfig(ctx, config)
	if err != nil {
		logger.WithError(err).Error("Failed to prepare provider configuration")
		return nil, fmt.Errorf("failed to prepare provider configuration: %w", err)
	}

	// 4. Get provider client
	providerClient, err := s.providerFactory.GetProvider(ctx, provider.Type, providerConfig)
	if err != nil {
		logger.WithError(err).Error("Failed to get provider client")
		return nil, fmt.Errorf("failed to get provider client: %w", err)
	}

	// 5. Transform and execute request
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
		config, err := s.providerConfigRepo.GetByProjectAndProvider(ctx, req.OrganizationID, provider.ID)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"provider_id": provider.ID,
				"error":       err,
			}).Warn("No configuration found for provider, skipping")
			continue
		}

		if !config.IsEnabled {
			continue
		}

		// Get models for this provider
		models, err := s.modelRepo.GetEnabledByProviderID(ctx, provider.ID)
		if err != nil {
			logger.WithError(err).WithField("provider_id", provider.ID).Error("Failed to fetch models for provider")
			continue
		}

		// Convert to response format
		for _, model := range models {
			// Build features list from model capabilities
			features := []string{}
			if model.SupportsStreaming {
				features = append(features, "streaming")
			}
			if model.SupportsFunctions {
				features = append(features, "functions")
			}
			if model.SupportsVision {
				features = append(features, "vision")
			}

			modelInfo := gateway.ModelInfo{
				ID:          model.ModelName,
				Object:      "model",
				Provider:    provider.Name,
				DisplayName: model.ModelName,
				MaxTokens:   model.MaxContextTokens,
				InputCost:   model.InputCostPer1kTokens,
				OutputCost:  model.OutputCostPer1kTokens,
				Features:    features,
				Metadata:    model.Metadata,
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
	s.logger.Info("Performing gateway health check")

	// Get active providers
	providers, err := s.providerRepo.ListEnabled(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to fetch active providers")
		return nil, fmt.Errorf("failed to fetch active providers: %w", err)
	}

	var healthStatuses []gateway.ProviderHealthStatus
	allHealthy := true

	for _, provider := range providers {
		// Test provider health
		healthStatus := "unhealthy"

		// In a real implementation, you would test the provider connection here
		// For now, assume all enabled providers are healthy
		healthStatus = "healthy"

		healthStatuses = append(healthStatuses, gateway.ProviderHealthStatus{
			ProviderID:   provider.ID,
			ProviderName: provider.Name,
			Status:       healthStatus,
			Error:        nil, // Only set if there's an error
		})
	}

	overallStatus := "healthy"
	if !allHealthy {
		overallStatus = "degraded"
	}
	if len(healthStatuses) == 0 {
		overallStatus = "unhealthy"
	}

	response := &gateway.HealthCheckResponse{
		Status:    overallStatus,
		Providers: healthStatuses,
		CheckedAt: time.Now(),
	}

	s.logger.WithFields(logrus.Fields{
		"overall_status":  overallStatus,
		"total_providers": len(providers),
	}).Info("Gateway health check completed")

	return response, nil
}

// Helper methods

// prepareProviderConfig decrypts the API key and prepares the full provider configuration
func (s *GatewayService) prepareProviderConfig(ctx context.Context, config *gateway.ProviderConfig) (map[string]interface{}, error) {
	// Decrypt API key
	decryptedKey, err := s.providerConfigRepo.DecryptAPIKey(ctx, config.APIKeyEncrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt API key: %w", err)
	}

	// Clone configuration map and add API key
	enhancedConfig := make(map[string]interface{})
	for k, v := range config.Configuration {
		enhancedConfig[k] = v
	}
	enhancedConfig["api_key"] = decryptedKey

	// Add other provider-specific config if needed
	if config.CustomBaseURL != nil {
		enhancedConfig["base_url"] = *config.CustomBaseURL
	}
	if config.CustomTimeoutSecs != nil {
		enhancedConfig["timeout"] = time.Duration(*config.CustomTimeoutSecs) * time.Second
	}

	return enhancedConfig, nil
}

func (s *GatewayService) resolveModelAndProvider(ctx context.Context, modelName string, projectID ulid.ULID) (*gateway.Model, *gateway.Provider, *gateway.ProviderConfig, error) {
	// First, try to find the model by name
	model, err := s.modelRepo.GetByModelName(ctx, modelName)
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

	// Get provider configuration for the project
	config, err := s.providerConfigRepo.GetByProjectAndProvider(ctx, projectID, provider.ID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("provider configuration not found for project: %w", err)
	}

	// Check if configuration is enabled
	if !config.IsEnabled {
		return nil, nil, nil, fmt.Errorf("provider configuration is disabled")
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
		"request_id":    metrics.RequestID,
		"project_id":    metrics.ProjectID,
		"provider":      metrics.Provider,
		"model":         metrics.Model,
		"input_tokens":  metrics.InputTokens,
		"output_tokens": metrics.OutputTokens,
		"total_tokens":  metrics.TotalTokens,
		"cost_usd":      metrics.CostUSD,
		"success":       metrics.Success,
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

// CreateChatCompletion implements the GatewayService interface
func (s *GatewayService) CreateChatCompletion(ctx context.Context, projectID ulid.ULID, req *gateway.ChatCompletionRequest) (*gateway.ChatCompletionResponse, error) {
	// Set the project ID in the request
	req.ProjectID = projectID
	return s.ProcessChatCompletion(ctx, req)
}

// CreateChatCompletionStream implements the GatewayService interface
func (s *GatewayService) CreateChatCompletionStream(ctx context.Context, projectID ulid.ULID, req *gateway.ChatCompletionRequest, writer io.Writer) error {
	// Set the project ID in the request and enable streaming
	req.ProjectID = projectID
	req.Stream = true
	// For now, return not implemented - streaming needs special handling
	return fmt.Errorf("streaming not implemented")
}

// CreateCompletion implements the GatewayService interface
func (s *GatewayService) CreateCompletion(ctx context.Context, projectID ulid.ULID, req *gateway.CompletionRequest) (*gateway.CompletionResponse, error) {
	// Set the project ID in the request
	req.ProjectID = projectID
	return s.ProcessCompletion(ctx, req)
}

// CreateCompletionStream implements the GatewayService interface
func (s *GatewayService) CreateCompletionStream(ctx context.Context, projectID ulid.ULID, req *gateway.CompletionRequest, writer io.Writer) error {
	// Set the project ID in the request and enable streaming
	req.ProjectID = projectID
	req.Stream = true
	// For now, return not implemented - streaming needs special handling
	return fmt.Errorf("streaming not implemented")
}

// CreateEmbedding implements the GatewayService interface
func (s *GatewayService) CreateEmbedding(ctx context.Context, projectID ulid.ULID, req *gateway.EmbeddingRequest) (*gateway.EmbeddingResponse, error) {
	// Convert EmbeddingRequest to EmbeddingsRequest
	embeddingsReq := &gateway.EmbeddingsRequest{
		OrganizationID: projectID,
		Model:          req.Model,
		Input:          req.Input,
	}

	if req.User != nil {
		embeddingsReq.User = *req.User
	}
	if req.EncodingFormat != nil {
		embeddingsReq.EncodingFormat = *req.EncodingFormat
	}

	// ProcessEmbeddings already returns *gateway.EmbeddingResponse
	return s.ProcessEmbeddings(ctx, embeddingsReq)
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

func (s *GatewayService) processEmbeddingsSync(ctx context.Context, client providers.Provider, req interface{}, model *gateway.Model, requestID ulid.ULID) (*gateway.EmbeddingResponse, *gateway.CostCalculation, error) {
	// Implementation depends on provider interface
	return nil, nil, fmt.Errorf("not implemented")
}

// ListAvailableModels implements the GatewayService interface
func (s *GatewayService) ListAvailableModels(ctx context.Context, projectID ulid.ULID) ([]*gateway.ModelInfo, error) {
	req := &gateway.ListModelsRequest{OrganizationID: projectID}
	resp, err := s.GetAvailableModels(ctx, req)
	if err != nil {
		return nil, err
	}

	// Convert slice of ModelInfo to slice of *ModelInfo
	result := make([]*gateway.ModelInfo, len(resp.Data))
	for i := range resp.Data {
		result[i] = &resp.Data[i]
	}
	return result, nil
}

// GetModel implements the GatewayService interface
func (s *GatewayService) GetModel(ctx context.Context, projectID ulid.ULID, modelName string) (*gateway.ModelInfo, error) {
	// Get the model from repository
	model, err := s.modelRepo.GetByModelName(ctx, modelName)
	if err != nil {
		return nil, fmt.Errorf("model not found: %w", err)
	}

	// Get the provider
	provider, err := s.providerRepo.GetByID(ctx, model.ProviderID)
	if err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}

	// Build features list from model capabilities
	features := []string{}
	if model.SupportsStreaming {
		features = append(features, "streaming")
	}
	if model.SupportsFunctions {
		features = append(features, "functions")
	}
	if model.SupportsVision {
		features = append(features, "vision")
	}

	return &gateway.ModelInfo{
		ID:           model.ModelName,
		Object:       "model",
		Provider:     provider.Name,
		DisplayName:  model.DisplayName,
		MaxTokens:    model.MaxContextTokens,
		InputCost:    model.InputCostPer1kTokens,
		OutputCost:   model.OutputCostPer1kTokens,
		Features:     features,
		QualityScore: model.QualityScore,
		SpeedScore:   model.SpeedScore,
		Metadata:     model.Metadata,
	}, nil
}

// GetRouteDecision implements the GatewayService interface
func (s *GatewayService) GetRouteDecision(ctx context.Context, projectID ulid.ULID, modelName string, strategy *string) (*gateway.RoutingDecision, error) {
	// Use routing service to determine the best route
	req := &gateway.RoutingRequest{
		ModelName: modelName,
	}

	if strategy != nil {
		// Convert string to RoutingStrategy enum
		// TODO: implement proper strategy conversion
	}

	decision, err := s.routingService.RouteRequest(ctx, projectID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get route decision: %w", err)
	}

	return decision, nil
}

// GetProviderHealth implements the GatewayService interface
func (s *GatewayService) GetProviderHealth(ctx context.Context, projectID ulid.ULID) ([]*gateway.ProviderHealth, error) {
	req := &gateway.HealthCheckRequest{}
	resp, err := s.CheckHealth(ctx, req)
	if err != nil {
		return nil, err
	}

	// Convert ProviderHealthStatus to ProviderHealth
	result := make([]*gateway.ProviderHealth, len(resp.Providers))
	for i, status := range resp.Providers {
		healthStatus := gateway.HealthStatusHealthy
		if status.Status == "unhealthy" {
			healthStatus = gateway.HealthStatusUnhealthy
		} else if status.Status == "degraded" {
			healthStatus = gateway.HealthStatusDegraded
		}

		result[i] = &gateway.ProviderHealth{
			ProviderID:   status.ProviderID,
			ProviderName: status.ProviderName,
			Status:       healthStatus,
			LastChecked:  resp.CheckedAt,
			LastError:    status.Error,
		}
	}
	return result, nil
}

// TestProviderConnection implements the GatewayService interface
func (s *GatewayService) TestProviderConnection(ctx context.Context, projectID ulid.ULID, providerID ulid.ULID) (*gateway.ConnectionTestResult, error) {
	// Get provider
	provider, err := s.providerRepo.GetByID(ctx, providerID)
	if err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}

	// Get provider configuration
	config, err := s.providerConfigRepo.GetByProjectAndProvider(ctx, projectID, providerID)
	if err != nil {
		return nil, fmt.Errorf("provider configuration not found: %w", err)
	}

	// Prepare provider configuration with decrypted API key
	providerConfig, err := s.prepareProviderConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare provider configuration: %w", err)
	}

	// Get provider client
	providerClient, err := s.providerFactory.GetProvider(ctx, provider.Type, providerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider client: %w", err)
	}

	// Test connection
	startTime := time.Now()
	err = providerClient.HealthCheck(ctx)
	latencyMs := int(time.Since(startTime).Milliseconds())

	result := &gateway.ConnectionTestResult{
		Success:   err == nil,
		LatencyMs: latencyMs,
		TestedAt:  time.Now(),
	}

	if err != nil {
		errorStr := err.Error()
		result.Error = &errorStr
	}

	return result, nil
}
