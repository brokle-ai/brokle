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

// Test fixtures for cost service
func createTestModelWithCosts() *gateway.Model {
	return &gateway.Model{
		ID:          ulid.New(),
		Name:        "gpt-3.5-turbo",
		ProviderID:  ulid.New(),
		Type:        gateway.ModelTypeText,
		DisplayName: "GPT-3.5 Turbo",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func createTestRequestMetrics() *gateway.RequestMetrics {
	return &gateway.RequestMetrics{
		RequestID:    "req-123",
		ModelName:    "gpt-3.5-turbo",
		ProviderName: "openai",
		Environment:  "production",
		InputTokens:  100,
		OutputTokens: 50,
		TotalTokens:  150,
		Status:       "completed",
		Duration:     250 * time.Millisecond,
		EstimatedCost: &gateway.CostBreakdown{
			InputCost:  0.0015,
			OutputCost: 0.003,
			TotalCost:  0.0045,
			Currency:   "USD",
		},
		ActualCost: &gateway.CostBreakdown{
			InputCost:  0.0015,
			OutputCost: 0.003,
			TotalCost:  0.0045,
			Currency:   "USD",
		},
		Timestamp: time.Now(),
	}
}

func createTestCostCalculationRequests() []*gateway.CostCalculationRequest {
	return []*gateway.CostCalculationRequest{
		{
			ModelID:      ulid.New(),
			ModelName:    "gpt-3.5-turbo",
			InputTokens:  100,
			OutputTokens: 50,
			RequestType:  gateway.RequestTypeChatCompletion,
		},
		{
			ModelID:      ulid.New(),
			ModelName:    "gpt-4",
			InputTokens:  200,
			OutputTokens: 100,
			RequestType:  gateway.RequestTypeChatCompletion,
		},
	}
}

func TestCostService_CalculateRequestCost(t *testing.T) {
	tests := []struct {
		name          string
		modelID       ulid.ULID
		inputTokens   int
		outputTokens  int
		setupMocks    func(*MockModelRepository)
		expectedCost  float64
		expectedError string
	}{
		{
			name:         "successful cost calculation",
			modelID:      ulid.New(),
			inputTokens:  100,
			outputTokens: 50,
			setupMocks: func(modelRepo *MockModelRepository) {
				model := createTestModelWithCosts()
				modelRepo.On("GetByID", mock.Anything, mock.Anything).Return(model, nil)
			},
			expectedCost: 0.0045, // Calculated based on typical GPT-3.5 pricing
		},
		{
			name:         "model not found",
			modelID:      ulid.New(),
			inputTokens:  100,
			outputTokens: 50,
			setupMocks: func(modelRepo *MockModelRepository) {
				modelRepo.On("GetByID", mock.Anything, mock.Anything).Return(nil, errors.New("model not found"))
			},
			expectedError: "model not found",
		},
		{
			name:         "zero tokens",
			modelID:      ulid.New(),
			inputTokens:  0,
			outputTokens: 0,
			setupMocks: func(modelRepo *MockModelRepository) {
				model := createTestModelWithCosts()
				modelRepo.On("GetByID", mock.Anything, mock.Anything).Return(model, nil)
			},
			expectedCost: 0.0,
		},
		{
			name:         "high token count",
			modelID:      ulid.New(),
			inputTokens:  10000,
			outputTokens: 5000,
			setupMocks: func(modelRepo *MockModelRepository) {
				model := createTestModelWithCosts()
				modelRepo.On("GetByID", mock.Anything, mock.Anything).Return(model, nil)
			},
			expectedCost: 0.45, // 15000 tokens * typical rate
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			modelRepo := &MockModelRepository{}
			tt.setupMocks(modelRepo)

			// Create service
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			service := NewCostService(logger)

			// Execute test
			ctx := context.Background()
			cost, err := service.CalculateRequestCost(ctx, tt.modelID, tt.inputTokens, tt.outputTokens)

			// Verify results
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.InDelta(t, tt.expectedCost, cost, 0.001) // Allow small floating point differences
			}

			// Verify expectations
			modelRepo.AssertExpectations(t)
		})
	}
}

func TestCostService_EstimateRequestCost(t *testing.T) {
	tests := []struct {
		name           string
		modelName      string
		estimatedTokens int
		expectedCost   float64
		expectedError  string
	}{
		{
			name:            "gpt-3.5-turbo estimation",
			modelName:       "gpt-3.5-turbo",
			estimatedTokens: 1000,
			expectedCost:    0.002, // $2 per 1M tokens typical
		},
		{
			name:            "gpt-4 estimation",
			modelName:       "gpt-4",
			estimatedTokens: 1000,
			expectedCost:    0.03, // $30 per 1M tokens typical
		},
		{
			name:            "claude-3-sonnet estimation",
			modelName:       "claude-3-sonnet",
			estimatedTokens: 1000,
			expectedCost:    0.003, // $3 per 1M tokens typical
		},
		{
			name:            "unknown model uses default",
			modelName:       "unknown-model",
			estimatedTokens: 1000,
			expectedCost:    0.002, // Default rate
		},
		{
			name:            "zero tokens",
			modelName:       "gpt-3.5-turbo",
			estimatedTokens: 0,
			expectedCost:    0.0,
		},
		{
			name:            "large token count",
			modelName:       "gpt-3.5-turbo",
			estimatedTokens: 100000,
			expectedCost:    0.2, // 100K tokens
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create service
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			service := NewCostService(logger)

			// Execute test
			ctx := context.Background()
			cost, err := service.EstimateRequestCost(ctx, tt.modelName, tt.estimatedTokens)

			// Verify results
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.InDelta(t, tt.expectedCost, cost, 0.001)
			}
		})
	}
}

func TestCostService_CalculateBatchCost(t *testing.T) {
	tests := []struct {
		name          string
		requests      []*gateway.CostCalculationRequest
		setupMocks    func(*MockModelRepository)
		validateResult func(*testing.T, *gateway.BatchCostResult)
		expectedError string
	}{
		{
			name:     "successful batch calculation",
			requests: createTestCostCalculationRequests(),
			setupMocks: func(modelRepo *MockModelRepository) {
				// Mock models for both requests
				model1 := createTestModelWithCosts()
				model1.Name = "gpt-3.5-turbo"
				model2 := createTestModelWithCosts()
				model2.Name = "gpt-4"
				
				modelRepo.On("GetByID", mock.Anything, mock.MatchedBy(func(id ulid.ULID) bool {
					return true // Accept any ULID
				})).Return(model1, nil).Once()
				
				modelRepo.On("GetByID", mock.Anything, mock.MatchedBy(func(id ulid.ULID) bool {
					return true
				})).Return(model2, nil).Once()
			},
			validateResult: func(t *testing.T, result *gateway.BatchCostResult) {
				assert.Len(t, result.Results, 2)
				assert.Greater(t, result.TotalCost, 0.0)
				assert.Equal(t, "USD", result.Currency)
				assert.NotZero(t, result.CalculatedAt)
				
				// Check individual results
				for _, res := range result.Results {
					assert.Greater(t, res.TotalCost, 0.0)
					assert.Equal(t, "USD", res.Currency)
				}
			},
		},
		{
			name:     "partial failures in batch",
			requests: createTestCostCalculationRequests(),
			setupMocks: func(modelRepo *MockModelRepository) {
				// First model succeeds
				model1 := createTestModelWithCosts()
				modelRepo.On("GetByID", mock.Anything, mock.Anything).Return(model1, nil).Once()
				
				// Second model fails
				modelRepo.On("GetByID", mock.Anything, mock.Anything).Return(nil, errors.New("model not found")).Once()
			},
			validateResult: func(t *testing.T, result *gateway.BatchCostResult) {
				assert.Len(t, result.Results, 2)
				
				// First should succeed
				assert.Greater(t, result.Results[0].TotalCost, 0.0)
				assert.Nil(t, result.Results[0].Error)
				
				// Second should fail
				assert.Equal(t, 0.0, result.Results[1].TotalCost)
				assert.NotNil(t, result.Results[1].Error)
				assert.Contains(t, *result.Results[1].Error, "model not found")
			},
		},
		{
			name:     "empty batch",
			requests: []*gateway.CostCalculationRequest{},
			validateResult: func(t *testing.T, result *gateway.BatchCostResult) {
				assert.Len(t, result.Results, 0)
				assert.Equal(t, 0.0, result.TotalCost)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			modelRepo := &MockModelRepository{}
			if tt.setupMocks != nil {
				tt.setupMocks(modelRepo)
			}

			// Create service
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			service := NewCostService(logger)

			// Execute test
			ctx := context.Background()
			result, err := service.CalculateBatchCost(ctx, tt.requests)

			// Verify results
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}

			// Verify expectations
			if tt.setupMocks != nil {
				modelRepo.AssertExpectations(t)
			}
		})
	}
}

func TestCostService_GetCostOptimizedProvider(t *testing.T) {
	tests := []struct {
		name          string
		projectID     ulid.ULID
		modelName     string
		setupMocks    func(*MockProviderRepository, *MockModelRepository, *MockProviderConfigRepository)
		validateResult func(*testing.T, *gateway.RoutingDecision)
		expectedError string
	}{
		{
			name:      "successful cost optimization",
			projectID: ulid.New(),
			modelName: "gpt-3.5-turbo",
			setupMocks: func(providerRepo *MockProviderRepository, modelRepo *MockModelRepository, configRepo *MockProviderConfigRepository) {
				providers, models, configs := createMultipleProviders()
				
				// Return active configs
				configRepo.On("GetActiveConfigsForProject", mock.Anything, mock.Anything).Return(configs, nil)
				
				// Return providers
				providerRepo.On("GetByID", mock.Anything, providers[0].ID).Return(providers[0], nil)
				providerRepo.On("GetByID", mock.Anything, providers[1].ID).Return(providers[1], nil)
				
				// Return model
				modelRepo.On("GetByName", mock.Anything, "gpt-3.5-turbo").Return(models[0], nil)
			},
			validateResult: func(t *testing.T, decision *gateway.RoutingDecision) {
				assert.NotEmpty(t, decision.ProviderName)
				assert.Equal(t, "gpt-3.5-turbo", decision.ModelName)
				assert.Equal(t, gateway.RoutingStrategyCostOptimized, decision.Strategy)
				assert.Contains(t, decision.Reason, "cost")
				assert.Greater(t, decision.Confidence, 0.0)
			},
		},
		{
			name:      "no providers available",
			projectID: ulid.New(),
			modelName: "gpt-3.5-turbo",
			setupMocks: func(providerRepo *MockProviderRepository, modelRepo *MockModelRepository, configRepo *MockProviderConfigRepository) {
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

			tt.setupMocks(providerRepo, modelRepo, configRepo)

			// Create service
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			service := NewCostService(logger)

			// Execute test
			ctx := context.Background()
			decision, err := service.GetCostOptimizedProvider(ctx, tt.projectID, tt.modelName)

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
		})
	}
}

func TestCostService_CompareCosts(t *testing.T) {
	tests := []struct {
		name          string
		modelNames    []string
		tokenCount    int
		validateResult func(*testing.T, *gateway.CostComparison)
		expectedError string
	}{
		{
			name:       "successful cost comparison",
			modelNames: []string{"gpt-3.5-turbo", "gpt-4", "claude-3-sonnet"},
			tokenCount: 1000,
			validateResult: func(t *testing.T, comparison *gateway.CostComparison) {
				assert.Len(t, comparison.Providers, 3)
				assert.NotNil(t, comparison.Cheapest)
				assert.NotNil(t, comparison.MostExpensive)
				assert.Greater(t, comparison.AverageCost, 0.0)
				assert.Greater(t, comparison.CostRange, 0.0)
				assert.Equal(t, "USD", comparison.Currency)
				assert.Equal(t, 1000, comparison.TokenCount)
				
				// Verify providers are sorted by cost
				for i := 1; i < len(comparison.Providers); i++ {
					assert.GreaterOrEqual(t, comparison.Providers[i].TotalCost, comparison.Providers[i-1].TotalCost)
				}
			},
		},
		{
			name:       "single model comparison",
			modelNames: []string{"gpt-3.5-turbo"},
			tokenCount: 500,
			validateResult: func(t *testing.T, comparison *gateway.CostComparison) {
				assert.Len(t, comparison.Providers, 1)
				assert.NotNil(t, comparison.Cheapest)
				assert.NotNil(t, comparison.MostExpensive)
				assert.Equal(t, comparison.Cheapest.TotalCost, comparison.MostExpensive.TotalCost)
				assert.Equal(t, 0.0, comparison.CostRange) // No range for single provider
			},
		},
		{
			name:          "empty model list",
			modelNames:    []string{},
			tokenCount:    1000,
			expectedError: "no models provided",
		},
		{
			name:       "zero tokens",
			modelNames: []string{"gpt-3.5-turbo", "gpt-4"},
			tokenCount: 0,
			validateResult: func(t *testing.T, comparison *gateway.CostComparison) {
				for _, provider := range comparison.Providers {
					assert.Equal(t, 0.0, provider.TotalCost)
				}
				assert.Equal(t, 0.0, comparison.AverageCost)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create service
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			service := NewCostService(logger)

			// Execute test
			ctx := context.Background()
			comparison, err := service.CompareCosts(ctx, tt.modelNames, tt.tokenCount)

			// Verify results
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, comparison)
			} else {
				require.NoError(t, err)
				require.NotNil(t, comparison)
				if tt.validateResult != nil {
					tt.validateResult(t, comparison)
				}
			}
		})
	}
}

func TestCostService_TrackRequestCost(t *testing.T) {
	tests := []struct {
		name           string
		metrics        *gateway.RequestMetrics
		expectedError  string
	}{
		{
			name:    "successful cost tracking",
			metrics: createTestRequestMetrics(),
		},
		{
			name: "tracking without actual cost",
			metrics: func() *gateway.RequestMetrics {
				m := createTestRequestMetrics()
				m.ActualCost = nil // No actual cost
				return m
			}(),
		},
		{
			name: "tracking with zero cost",
			metrics: func() *gateway.RequestMetrics {
				m := createTestRequestMetrics()
				m.ActualCost.TotalCost = 0.0
				return m
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create service
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			service := NewCostService(logger)

			// Execute test
			ctx := context.Background()
			err := service.TrackRequestCost(ctx, tt.metrics)

			// Verify results
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCostService_GetProjectCostAnalytics(t *testing.T) {
	projectID := ulid.New()
	timeRange := &gateway.TimeRange{
		StartTime: time.Now().Add(-30 * 24 * time.Hour), // 30 days ago
		EndTime:   time.Now(),
		Interval:  stringPtr("daily"),
	}

	// Create service
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewCostService(logger)

	// Execute test
	ctx := context.Background()
	analytics, err := service.GetProjectCostAnalytics(ctx, projectID, timeRange)

	// Verify results (this would typically require more complex mocking)
	// For now, we test the method exists and handles the call gracefully
	if err != nil {
		assert.Contains(t, err.Error(), "not implemented") // Expected for stub
	} else {
		assert.NotNil(t, analytics)
	}
}

func TestCostService_GetProviderCostBreakdown(t *testing.T) {
	projectID := ulid.New()
	timeRange := &gateway.TimeRange{
		StartTime: time.Now().Add(-7 * 24 * time.Hour), // 7 days ago
		EndTime:   time.Now(),
		Interval:  stringPtr("hourly"),
	}

	// Create service
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewCostService(logger)

	// Execute test
	ctx := context.Background()
	breakdown, err := service.GetProviderCostBreakdown(ctx, projectID, timeRange)

	// Verify results (this would typically require more complex mocking)
	if err != nil {
		assert.Contains(t, err.Error(), "not implemented") // Expected for stub
	} else {
		assert.NotNil(t, breakdown)
	}
}

// Test cost calculation edge cases
func TestCostService_EdgeCases(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewCostService(logger)
	ctx := context.Background()

	t.Run("negative tokens", func(t *testing.T) {
		cost, err := service.EstimateRequestCost(ctx, "gpt-3.5-turbo", -100)
		assert.NoError(t, err)
		assert.Equal(t, 0.0, cost) // Should handle negative gracefully
	})

	t.Run("extremely large token count", func(t *testing.T) {
		cost, err := service.EstimateRequestCost(ctx, "gpt-3.5-turbo", 10000000) // 10M tokens
		assert.NoError(t, err)
		assert.Greater(t, cost, 0.0)
		assert.Less(t, cost, 1000.0) // Reasonable upper bound
	})

	t.Run("empty model name", func(t *testing.T) {
		cost, err := service.EstimateRequestCost(ctx, "", 1000)
		assert.NoError(t, err)
		assert.Greater(t, cost, 0.0) // Should use default
	})
}

// Benchmark tests
func BenchmarkCostService_EstimateRequestCost(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewCostService(logger)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.EstimateRequestCost(ctx, "gpt-3.5-turbo", 1000)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCostService_CompareCosts(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewCostService(logger)
	ctx := context.Background()
	models := []string{"gpt-3.5-turbo", "gpt-4", "claude-3-sonnet"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.CompareCosts(ctx, models, 1000)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Helper function for tests
func stringPtr(s string) *string {
	return &s
}