package gateway

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"brokle/internal/core/domain/gateway"
	"brokle/pkg/ulid"
)

// Additional mock for provider config repository
type MockProviderConfigRepository struct {
	mock.Mock
}

func (m *MockProviderConfigRepository) Create(ctx context.Context, config *gateway.ProviderConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockProviderConfigRepository) GetByID(ctx context.Context, id ulid.ULID) (*gateway.ProviderConfig, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gateway.ProviderConfig), args.Error(1)
}

func (m *MockProviderConfigRepository) GetByProjectAndProvider(ctx context.Context, projectID, providerID ulid.ULID) (*gateway.ProviderConfig, error) {
	args := m.Called(ctx, projectID, providerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gateway.ProviderConfig), args.Error(1)
}

func (m *MockProviderConfigRepository) Update(ctx context.Context, config *gateway.ProviderConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockProviderConfigRepository) Delete(ctx context.Context, id ulid.ULID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProviderConfigRepository) ListByProject(ctx context.Context, projectID ulid.ULID) ([]*gateway.ProviderConfig, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*gateway.ProviderConfig), args.Error(1)
}

func (m *MockProviderConfigRepository) GetActiveConfigsForProject(ctx context.Context, projectID ulid.ULID) ([]*gateway.ProviderConfig, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*gateway.ProviderConfig), args.Error(1)
}

// Test fixtures for routing
func createTestProviderConfig() *gateway.ProviderConfig {
	return &gateway.ProviderConfig{
		ID:         ulid.New(),
		ProjectID:  ulid.New(),
		ProviderID: ulid.New(),
		APIKey:     "test-api-key",
		Priority:   5,
		Enabled:    true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

func createTestRoutingRequest() *gateway.RoutingRequest {
	return &gateway.RoutingRequest{
		ModelName: "gpt-3.5-turbo",
		Strategy:  &gateway.RoutingStrategyCostOptimized,
		Requirements: &gateway.ModelRequirements{
			MaxTokens:     2048,
			MinQuality:    floatPtr(0.8),
			MaxLatency:    durationPtr(2 * time.Second),
			MaxCost:       floatPtr(0.01),
		},
		EstimatedTokens: intPtr(500),
		UserTier:        &gateway.UserTierPro,
		Priority:        &gateway.PriorityMedium,
	}
}

func createMultipleProviders() ([]*gateway.Provider, []*gateway.Model, []*gateway.ProviderConfig) {
	// Create providers
	provider1 := &gateway.Provider{
		ID:          ulid.New(),
		Name:        "openai",
		Type:        gateway.ProviderTypeOpenAI,
		DisplayName: "OpenAI",
		Status:      gateway.ProviderStatusActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	provider2 := &gateway.Provider{
		ID:          ulid.New(),
		Name:        "anthropic",
		Type:        gateway.ProviderTypeAnthropic,
		DisplayName: "Anthropic",
		Status:      gateway.ProviderStatusActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	providers := []*gateway.Provider{provider1, provider2}

	// Create models
	model1 := &gateway.Model{
		ID:          ulid.New(),
		Name:        "gpt-3.5-turbo",
		ProviderID:  provider1.ID,
		Type:        gateway.ModelTypeText,
		DisplayName: "GPT-3.5 Turbo",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	model2 := &gateway.Model{
		ID:          ulid.New(),
		Name:        "claude-3-sonnet",
		ProviderID:  provider2.ID,
		Type:        gateway.ModelTypeText,
		DisplayName: "Claude 3 Sonnet",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	models := []*gateway.Model{model1, model2}

	// Create provider configs
	config1 := &gateway.ProviderConfig{
		ID:         ulid.New(),
		ProjectID:  ulid.New(),
		ProviderID: provider1.ID,
		APIKey:     "openai-key",
		Priority:   5,
		Enabled:    true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	config2 := &gateway.ProviderConfig{
		ID:         ulid.New(),
		ProjectID:  config1.ProjectID, // Same project
		ProviderID: provider2.ID,
		APIKey:     "anthropic-key",
		Priority:   3,
		Enabled:    true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	configs := []*gateway.ProviderConfig{config1, config2}

	return providers, models, configs
}

func TestRoutingService_RouteRequest(t *testing.T) {
	tests := []struct {
		name          string
		projectID     ulid.ULID
		request       *gateway.RoutingRequest
		setupMocks    func(*MockProviderRepository, *MockModelRepository, *MockProviderConfigRepository, *MockCostService)
		expectedError string
		validateResult func(*testing.T, *gateway.RoutingDecision)
	}{
		{
			name:      "successful cost-optimized routing",
			projectID: ulid.New(),
			request:   createTestRoutingRequest(),
			setupMocks: func(providerRepo *MockProviderRepository, modelRepo *MockModelRepository, configRepo *MockProviderConfigRepository, costService *MockCostService) {
				providers, models, configs := createMultipleProviders()
				
				// Return active provider configs
				configRepo.On("GetActiveConfigsForProject", mock.Anything, mock.Anything).Return(configs, nil)
				
				// Return providers and models
				providerRepo.On("GetByID", mock.Anything, providers[0].ID).Return(providers[0], nil)
				providerRepo.On("GetByID", mock.Anything, providers[1].ID).Return(providers[1], nil)
				
				modelRepo.On("GetByName", mock.Anything, "gpt-3.5-turbo").Return(models[0], nil)
				
				// Mock cost calculations - OpenAI cheaper
				costService.On("EstimateRequestCost", mock.Anything, "gpt-3.5-turbo", 500).Return(0.01, nil)
				costService.On("EstimateRequestCost", mock.Anything, "claude-3-sonnet", 500).Return(0.02, nil)
			},
			validateResult: func(t *testing.T, decision *gateway.RoutingDecision) {
				assert.Equal(t, "openai", decision.ProviderName)
				assert.Equal(t, "gpt-3.5-turbo", decision.ModelName)
				assert.Equal(t, gateway.RoutingStrategyCostOptimized, decision.Strategy)
				assert.Contains(t, decision.Reason, "cost")
				assert.Greater(t, decision.Confidence, 0.0)
			},
		},
		{
			name:      "no available providers",
			projectID: ulid.New(),
			request:   createTestRoutingRequest(),
			setupMocks: func(providerRepo *MockProviderRepository, modelRepo *MockModelRepository, configRepo *MockProviderConfigRepository, costService *MockCostService) {
				configRepo.On("GetActiveConfigsForProject", mock.Anything, mock.Anything).Return([]*gateway.ProviderConfig{}, nil)
			},
			expectedError: "no available providers",
		},
		{
			name:      "model not found",
			projectID: ulid.New(),
			request: &gateway.RoutingRequest{
				ModelName: "non-existent-model",
				Strategy:  &gateway.RoutingStrategyCostOptimized,
			},
			setupMocks: func(providerRepo *MockProviderRepository, modelRepo *MockModelRepository, configRepo *MockProviderConfigRepository, costService *MockCostService) {
				_, _, configs := createMultipleProviders()
				configRepo.On("GetActiveConfigsForProject", mock.Anything, mock.Anything).Return(configs, nil)
				modelRepo.On("GetByName", mock.Anything, "non-existent-model").Return(nil, errors.New("model not found"))
			},
			expectedError: "model not found",
		},
		{
			name:      "latency-optimized routing",
			projectID: ulid.New(),
			request: &gateway.RoutingRequest{
				ModelName: "gpt-3.5-turbo",
				Strategy:  &gateway.RoutingStrategyLatencyOptimized,
			},
			setupMocks: func(providerRepo *MockProviderRepository, modelRepo *MockModelRepository, configRepo *MockProviderConfigRepository, costService *MockCostService) {
				providers, models, configs := createMultipleProviders()
				
				configRepo.On("GetActiveConfigsForProject", mock.Anything, mock.Anything).Return(configs, nil)
				providerRepo.On("GetByID", mock.Anything, providers[0].ID).Return(providers[0], nil)
				providerRepo.On("GetByID", mock.Anything, providers[1].ID).Return(providers[1], nil)
				modelRepo.On("GetByName", mock.Anything, "gpt-3.5-turbo").Return(models[0], nil)
				
				// For latency routing, we would check provider health/performance
				// Simplified: just return the first provider
			},
			validateResult: func(t *testing.T, decision *gateway.RoutingDecision) {
				assert.Equal(t, gateway.RoutingStrategyLatencyOptimized, decision.Strategy)
				assert.Contains(t, decision.Reason, "latency")
			},
		},
		{
			name:      "priority-based fallback",
			projectID: ulid.New(),
			request: &gateway.RoutingRequest{
				ModelName: "gpt-3.5-turbo",
				Strategy:  &gateway.RoutingStrategyFailover,
			},
			setupMocks: func(providerRepo *MockProviderRepository, modelRepo *MockModelRepository, configRepo *MockProviderConfigRepository, costService *MockCostService) {
				providers, models, configs := createMultipleProviders()
				
				// Higher priority config first
				configs[0].Priority = 10
				configs[1].Priority = 5
				
				configRepo.On("GetActiveConfigsForProject", mock.Anything, mock.Anything).Return(configs, nil)
				providerRepo.On("GetByID", mock.Anything, providers[0].ID).Return(providers[0], nil)
				modelRepo.On("GetByName", mock.Anything, "gpt-3.5-turbo").Return(models[0], nil)
			},
			validateResult: func(t *testing.T, decision *gateway.RoutingDecision) {
				assert.Equal(t, gateway.RoutingStrategyFailover, decision.Strategy)
				assert.Equal(t, "openai", decision.ProviderName)
				assert.Contains(t, decision.Reason, "priority")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			providerRepo := &MockProviderRepository{}
			modelRepo := &MockModelRepository{}
			configRepo := &MockProviderConfigRepository{}
			costService := &MockCostService{}

			tt.setupMocks(providerRepo, modelRepo, configRepo, costService)

			// Create service
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			service := NewRoutingService(
				providerRepo,
				modelRepo,
				configRepo,
				costService,
				logger,
			)

			// Execute test
			ctx := context.Background()
			decision, err := service.RouteRequest(ctx, tt.projectID, tt.request)

			// Verify results
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, decision)
			} else {
				require.NoError(t, err)
				require.NotNil(t, decision)
				if tt.validateResult != nil {
					tt.validateResult(t, decision)
				}
			}

			// Verify expectations
			providerRepo.AssertExpectations(t)
			modelRepo.AssertExpectations(t)
			configRepo.AssertExpectations(t)
			costService.AssertExpectations(t)
		})
	}
}

func TestRoutingService_GetBestProvider(t *testing.T) {
	tests := []struct {
		name           string
		projectID      ulid.ULID
		modelName      string
		strategy       gateway.RoutingStrategy
		setupMocks     func(*MockProviderRepository, *MockModelRepository, *MockProviderConfigRepository, *MockCostService)
		expectedError  string
		validateResult func(*testing.T, *gateway.RoutingDecision)
	}{
		{
			name:      "cost-optimized selection",
			projectID: ulid.New(),
			modelName: "gpt-3.5-turbo",
			strategy:  gateway.RoutingStrategyCostOptimized,
			setupMocks: func(providerRepo *MockProviderRepository, modelRepo *MockModelRepository, configRepo *MockProviderConfigRepository, costService *MockCostService) {
				providers, models, configs := createMultipleProviders()
				
				configRepo.On("GetActiveConfigsForProject", mock.Anything, mock.Anything).Return(configs, nil)
				providerRepo.On("GetByID", mock.Anything, providers[0].ID).Return(providers[0], nil)
				providerRepo.On("GetByID", mock.Anything, providers[1].ID).Return(providers[1], nil)
				modelRepo.On("GetByName", mock.Anything, "gpt-3.5-turbo").Return(models[0], nil)
				
				// Anthropic cheaper in this scenario
				costService.On("EstimateRequestCost", mock.Anything, "gpt-3.5-turbo", 1000).Return(0.02, nil)
				costService.On("EstimateRequestCost", mock.Anything, "claude-3-sonnet", 1000).Return(0.015, nil)
			},
			validateResult: func(t *testing.T, decision *gateway.RoutingDecision) {
				assert.Equal(t, gateway.RoutingStrategyCostOptimized, decision.Strategy)
				assert.NotEmpty(t, decision.ProviderName)
				assert.Greater(t, decision.Confidence, 0.0)
				assert.Equal(t, 0.015, decision.EstimatedCost)
			},
		},
		{
			name:      "quality-optimized selection",
			projectID: ulid.New(),
			modelName: "gpt-4",
			strategy:  gateway.RoutingStrategyQualityOptimized,
			setupMocks: func(providerRepo *MockProviderRepository, modelRepo *MockModelRepository, configRepo *MockProviderConfigRepository, costService *MockCostService) {
				providers, models, configs := createMultipleProviders()
				
				// Create a GPT-4 model
				gpt4Model := &gateway.Model{
					ID:          ulid.New(),
					Name:        "gpt-4",
					ProviderID:  providers[0].ID,
					Type:        gateway.ModelTypeText,
					DisplayName: "GPT-4",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				
				configRepo.On("GetActiveConfigsForProject", mock.Anything, mock.Anything).Return(configs, nil)
				providerRepo.On("GetByID", mock.Anything, providers[0].ID).Return(providers[0], nil)
				modelRepo.On("GetByName", mock.Anything, "gpt-4").Return(gpt4Model, nil)
			},
			validateResult: func(t *testing.T, decision *gateway.RoutingDecision) {
				assert.Equal(t, gateway.RoutingStrategyQualityOptimized, decision.Strategy)
				assert.Equal(t, "gpt-4", decision.ModelName)
				assert.Contains(t, decision.Reason, "quality")
			},
		},
		{
			name:      "load-balanced selection",
			projectID: ulid.New(),
			modelName: "gpt-3.5-turbo",
			strategy:  gateway.RoutingStrategyLoadBalance,
			setupMocks: func(providerRepo *MockProviderRepository, modelRepo *MockModelRepository, configRepo *MockProviderConfigRepository, costService *MockCostService) {
				providers, models, configs := createMultipleProviders()
				
				configRepo.On("GetActiveConfigsForProject", mock.Anything, mock.Anything).Return(configs, nil)
				providerRepo.On("GetByID", mock.Anything, providers[0].ID).Return(providers[0], nil)
				providerRepo.On("GetByID", mock.Anything, providers[1].ID).Return(providers[1], nil)
				modelRepo.On("GetByName", mock.Anything, "gpt-3.5-turbo").Return(models[0], nil)
			},
			validateResult: func(t *testing.T, decision *gateway.RoutingDecision) {
				assert.Equal(t, gateway.RoutingStrategyLoadBalance, decision.Strategy)
				assert.Contains(t, decision.Reason, "load")
				// In load balancing, we expect some randomization/distribution logic
			},
		},
		{
			name:          "no providers available",
			projectID:     ulid.New(),
			modelName:     "gpt-3.5-turbo",
			strategy:      gateway.RoutingStrategyCostOptimized,
			setupMocks: func(providerRepo *MockProviderRepository, modelRepo *MockModelRepository, configRepo *MockProviderConfigRepository, costService *MockCostService) {
				configRepo.On("GetActiveConfigsForProject", mock.Anything, mock.Anything).Return([]*gateway.ProviderConfig{}, nil)
			},
			expectedError: "no available providers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			providerRepo := &MockProviderRepository{}
			modelRepo := &MockModelRepository{}
			configRepo := &MockProviderConfigRepository{}
			costService := &MockCostService{}

			tt.setupMocks(providerRepo, modelRepo, configRepo, costService)

			// Create service
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			service := NewRoutingService(
				providerRepo,
				modelRepo,
				configRepo,
				costService,
				logger,
			)

			// Execute test
			ctx := context.Background()
			decision, err := service.GetBestProvider(ctx, tt.projectID, tt.modelName, tt.strategy)

			// Verify results
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, decision)
			} else {
				require.NoError(t, err)
				require.NotNil(t, decision)
				if tt.validateResult != nil {
					tt.validateResult(t, decision)
				}
			}

			// Verify expectations
			providerRepo.AssertExpectations(t)
			modelRepo.AssertExpectations(t)
			configRepo.AssertExpectations(t)
			costService.AssertExpectations(t)
		})
	}
}

func TestRoutingService_RouteByCost(t *testing.T) {
	// Setup mocks
	providerRepo := &MockProviderRepository{}
	modelRepo := &MockModelRepository{}
	configRepo := &MockProviderConfigRepository{}
	costService := &MockCostService{}

	providers, models, configs := createMultipleProviders()
	
	configRepo.On("GetActiveConfigsForProject", mock.Anything, mock.Anything).Return(configs, nil)
	providerRepo.On("GetByID", mock.Anything, providers[0].ID).Return(providers[0], nil)
	providerRepo.On("GetByID", mock.Anything, providers[1].ID).Return(providers[1], nil)
	modelRepo.On("GetByName", mock.Anything, "gpt-3.5-turbo").Return(models[0], nil)
	
	// OpenAI cheaper
	costService.On("EstimateRequestCost", mock.Anything, "gpt-3.5-turbo", 1000).Return(0.01, nil)
	costService.On("EstimateRequestCost", mock.Anything, "claude-3-sonnet", 1000).Return(0.02, nil)

	// Create service
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewRoutingService(
		providerRepo,
		modelRepo,
		configRepo,
		costService,
		logger,
	)

	// Execute test
	ctx := context.Background()
	decision, err := service.RouteByCost(ctx, ulid.New(), "gpt-3.5-turbo")

	// Verify results
	assert.NoError(t, err)
	assert.NotNil(t, decision)
	assert.Equal(t, gateway.RoutingStrategyCostOptimized, decision.Strategy)
	assert.Equal(t, "openai", decision.ProviderName)
	assert.Equal(t, 0.01, decision.EstimatedCost)
	assert.Contains(t, decision.Reason, "cost")

	// Verify expectations
	configRepo.AssertExpectations(t)
	providerRepo.AssertExpectations(t)
	modelRepo.AssertExpectations(t)
	costService.AssertExpectations(t)
}

func TestRoutingService_RouteByLatency(t *testing.T) {
	// Setup mocks
	providerRepo := &MockProviderRepository{}
	modelRepo := &MockModelRepository{}
	configRepo := &MockProviderConfigRepository{}
	costService := &MockCostService{}

	providers, models, configs := createMultipleProviders()
	
	configRepo.On("GetActiveConfigsForProject", mock.Anything, mock.Anything).Return(configs, nil)
	providerRepo.On("GetByID", mock.Anything, providers[0].ID).Return(providers[0], nil)
	providerRepo.On("GetByID", mock.Anything, providers[1].ID).Return(providers[1], nil)
	modelRepo.On("GetByName", mock.Anything, "gpt-3.5-turbo").Return(models[0], nil)

	// Create service
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewRoutingService(
		providerRepo,
		modelRepo,
		configRepo,
		costService,
		logger,
	)

	// Execute test
	ctx := context.Background()
	decision, err := service.RouteByLatency(ctx, ulid.New(), "gpt-3.5-turbo")

	// Verify results
	assert.NoError(t, err)
	assert.NotNil(t, decision)
	assert.Equal(t, gateway.RoutingStrategyLatencyOptimized, decision.Strategy)
	assert.Contains(t, decision.Reason, "latency")

	// Verify expectations
	configRepo.AssertExpectations(t)
	providerRepo.AssertExpectations(t)
	modelRepo.AssertExpectations(t)
}

func TestRoutingService_RouteByQuality(t *testing.T) {
	// Setup mocks
	providerRepo := &MockProviderRepository{}
	modelRepo := &MockModelRepository{}
	configRepo := &MockProviderConfigRepository{}
	costService := &MockCostService{}

	providers, models, configs := createMultipleProviders()
	
	configRepo.On("GetActiveConfigsForProject", mock.Anything, mock.Anything).Return(configs, nil)
	providerRepo.On("GetByID", mock.Anything, providers[0].ID).Return(providers[0], nil)
	providerRepo.On("GetByID", mock.Anything, providers[1].ID).Return(providers[1], nil)
	modelRepo.On("GetByName", mock.Anything, "gpt-3.5-turbo").Return(models[0], nil)

	// Create service
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewRoutingService(
		providerRepo,
		modelRepo,
		configRepo,
		costService,
		logger,
	)

	// Execute test
	ctx := context.Background()
	decision, err := service.RouteByQuality(ctx, ulid.New(), "gpt-3.5-turbo")

	// Verify results
	assert.NoError(t, err)
	assert.NotNil(t, decision)
	assert.Equal(t, gateway.RoutingStrategyQualityOptimized, decision.Strategy)
	assert.Contains(t, decision.Reason, "quality")

	// Verify expectations
	configRepo.AssertExpectations(t)
	providerRepo.AssertExpectations(t)
	modelRepo.AssertExpectations(t)
}

func TestRoutingService_GetFallbackProvider(t *testing.T) {
	// Setup mocks
	providerRepo := &MockProviderRepository{}
	modelRepo := &MockModelRepository{}
	configRepo := &MockProviderConfigRepository{}
	costService := &MockCostService{}

	providers, models, configs := createMultipleProviders()
	failedProviderID := providers[0].ID
	
	// Return configs excluding the failed provider
	fallbackConfigs := []*gateway.ProviderConfig{configs[1]} // Only anthropic
	
	configRepo.On("GetActiveConfigsForProject", mock.Anything, mock.Anything).Return(fallbackConfigs, nil)
	providerRepo.On("GetByID", mock.Anything, providers[1].ID).Return(providers[1], nil)
	modelRepo.On("GetByName", mock.Anything, "gpt-3.5-turbo").Return(models[0], nil)

	// Create service
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewRoutingService(
		providerRepo,
		modelRepo,
		configRepo,
		costService,
		logger,
	)

	// Execute test
	ctx := context.Background()
	decision, err := service.GetFallbackProvider(ctx, ulid.New(), failedProviderID, "gpt-3.5-turbo")

	// Verify results
	assert.NoError(t, err)
	assert.NotNil(t, decision)
	assert.NotEqual(t, "openai", decision.ProviderName) // Should not be the failed provider
	assert.Contains(t, decision.Reason, "fallback")

	// Verify expectations
	configRepo.AssertExpectations(t)
	providerRepo.AssertExpectations(t)
	modelRepo.AssertExpectations(t)
}

// Helper functions for tests
func floatPtr(f float64) *float64 {
	return &f
}

func durationPtr(d time.Duration) *time.Duration {
	return &d
}

func intPtr(i int) *int {
	return &i
}

// Benchmark tests
func BenchmarkRoutingService_RouteRequest(b *testing.B) {
	// Setup mocks
	providerRepo := &MockProviderRepository{}
	modelRepo := &MockModelRepository{}
	configRepo := &MockProviderConfigRepository{}
	costService := &MockCostService{}

	providers, models, configs := createMultipleProviders()
	
	configRepo.On("GetActiveConfigsForProject", mock.Anything, mock.Anything).Return(configs, nil)
	providerRepo.On("GetByID", mock.Anything, mock.Anything).Return(providers[0], nil)
	modelRepo.On("GetByName", mock.Anything, "gpt-3.5-turbo").Return(models[0], nil)
	costService.On("EstimateRequestCost", mock.Anything, mock.Anything, 500).Return(0.01, nil)

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewRoutingService(
		providerRepo,
		modelRepo,
		configRepo,
		costService,
		logger,
	)

	request := createTestRoutingRequest()
	ctx := context.Background()
	projectID := ulid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.RouteRequest(ctx, projectID, request)
		if err != nil {
			b.Fatal(err)
		}
	}
}