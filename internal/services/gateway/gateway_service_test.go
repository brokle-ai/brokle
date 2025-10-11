package gateway

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"brokle/internal/core/domain/gateway"
	"brokle/pkg/ulid"
)

// Mock implementations for testing

type MockProviderRepository struct {
	mock.Mock
}

func (m *MockProviderRepository) Create(ctx context.Context, provider *gateway.Provider) error {
	args := m.Called(ctx, provider)
	return args.Error(0)
}

func (m *MockProviderRepository) GetByID(ctx context.Context, id ulid.ULID) (*gateway.Provider, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gateway.Provider), args.Error(1)
}

func (m *MockProviderRepository) GetByName(ctx context.Context, name string) (*gateway.Provider, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gateway.Provider), args.Error(1)
}

func (m *MockProviderRepository) Update(ctx context.Context, provider *gateway.Provider) error {
	args := m.Called(ctx, provider)
	return args.Error(0)
}

func (m *MockProviderRepository) Delete(ctx context.Context, id ulid.ULID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProviderRepository) List(ctx context.Context, filter *gateway.ProviderFilter, limit, offset int) ([]*gateway.Provider, int, error) {
	args := m.Called(ctx, filter, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*gateway.Provider), args.Int(1), args.Error(2)
}

func (m *MockProviderRepository) GetActiveProviders(ctx context.Context, projectID ulid.ULID) ([]*gateway.Provider, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*gateway.Provider), args.Error(1)
}

type MockModelRepository struct {
	mock.Mock
}

func (m *MockModelRepository) Create(ctx context.Context, model *gateway.Model) error {
	args := m.Called(ctx, model)
	return args.Error(0)
}

func (m *MockModelRepository) GetByID(ctx context.Context, id ulid.ULID) (*gateway.Model, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gateway.Model), args.Error(1)
}

func (m *MockModelRepository) GetByName(ctx context.Context, name string) (*gateway.Model, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gateway.Model), args.Error(1)
}

func (m *MockModelRepository) Update(ctx context.Context, model *gateway.Model) error {
	args := m.Called(ctx, model)
	return args.Error(0)
}

func (m *MockModelRepository) Delete(ctx context.Context, id ulid.ULID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockModelRepository) List(ctx context.Context, filter *gateway.ModelFilter, limit, offset int) ([]*gateway.Model, int, error) {
	args := m.Called(ctx, filter, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*gateway.Model), args.Int(1), args.Error(2)
}

func (m *MockModelRepository) GetActiveModels(ctx context.Context, projectID ulid.ULID) ([]*gateway.Model, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*gateway.Model), args.Error(1)
}

type MockRoutingService struct {
	mock.Mock
}

func (m *MockRoutingService) RouteRequest(ctx context.Context, projectID ulid.ULID, request *gateway.RoutingRequest) (*gateway.RoutingDecision, error) {
	args := m.Called(ctx, projectID, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gateway.RoutingDecision), args.Error(1)
}

func (m *MockRoutingService) GetBestProvider(ctx context.Context, projectID ulid.ULID, modelName string, strategy gateway.RoutingStrategy) (*gateway.RoutingDecision, error) {
	args := m.Called(ctx, projectID, modelName, strategy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gateway.RoutingDecision), args.Error(1)
}

type MockCostService struct {
	mock.Mock
}

func (m *MockCostService) CalculateRequestCost(ctx context.Context, modelID ulid.ULID, inputTokens, outputTokens int) (float64, error) {
	args := m.Called(ctx, modelID, inputTokens, outputTokens)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockCostService) EstimateRequestCost(ctx context.Context, modelName string, estimatedTokens int) (float64, error) {
	args := m.Called(ctx, modelName, estimatedTokens)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockCostService) TrackRequestCost(ctx context.Context, metrics *gateway.RequestMetrics) error {
	args := m.Called(ctx, metrics)
	return args.Error(0)
}

type MockProviderClient struct {
	mock.Mock
}

func (m *MockProviderClient) CreateChatCompletion(ctx context.Context, request *gateway.ChatCompletionRequest) (*gateway.ChatCompletionResponse, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gateway.ChatCompletionResponse), args.Error(1)
}

func (m *MockProviderClient) CreateChatCompletionStream(ctx context.Context, request *gateway.ChatCompletionRequest, writer io.Writer) error {
	args := m.Called(ctx, request, writer)
	return args.Error(0)
}

func (m *MockProviderClient) CreateCompletion(ctx context.Context, request *gateway.CompletionRequest) (*gateway.CompletionResponse, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gateway.CompletionResponse), args.Error(1)
}

func (m *MockProviderClient) CreateCompletionStream(ctx context.Context, request *gateway.CompletionRequest, writer io.Writer) error {
	args := m.Called(ctx, request, writer)
	return args.Error(0)
}

func (m *MockProviderClient) CreateEmbedding(ctx context.Context, request *gateway.EmbeddingRequest) (*gateway.EmbeddingResponse, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gateway.EmbeddingResponse), args.Error(1)
}

func (m *MockProviderClient) ListModels(ctx context.Context) ([]*gateway.ModelInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*gateway.ModelInfo), args.Error(1)
}

func (m *MockProviderClient) GetModel(ctx context.Context, modelID string) (*gateway.ModelInfo, error) {
	args := m.Called(ctx, modelID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gateway.ModelInfo), args.Error(1)
}

func (m *MockProviderClient) ValidateCredentials(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Test fixtures
func createTestProvider() *gateway.Provider {
	return &gateway.Provider{
		ID:          ulid.New(),
		Name:        "openai",
		Type:        gateway.ProviderTypeOpenAI,
		DisplayName: "OpenAI",
		Status:      gateway.ProviderStatusActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func createTestModel() *gateway.Model {
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

func createTestChatRequest() *gateway.ChatCompletionRequest {
	return &gateway.ChatCompletionRequest{
		Model: "gpt-3.5-turbo",
		Messages: []gateway.ChatMessage{
			{
				Role:    "user",
				Content: "Hello, world!",
			},
		},
	}
}

func createTestRoutingDecision() *gateway.RoutingDecision {
	return &gateway.RoutingDecision{
		ProviderID:    ulid.New(),
		ProviderName:  "openai",
		ModelID:       ulid.New(),
		ModelName:     "gpt-3.5-turbo",
		Strategy:      gateway.RoutingStrategyCostOptimized,
		Reason:        "Lowest cost provider available",
		Confidence:    0.95,
		EstimatedCost: 0.002,
		DecisionTime:  time.Now(),
	}
}

// Test suite
func TestGatewayService_CreateChatCompletion(t *testing.T) {
	tests := []struct {
		name          string
		projectID     ulid.ULID
		environment   string
		request       *gateway.ChatCompletionRequest
		setupMocks    func(*MockProviderRepository, *MockModelRepository, *MockRoutingService, *MockCostService, *MockProviderClient)
		expectedError string
	}{
		{
			name:        "successful chat completion",
			projectID:   ulid.New(),
			environment: "production",
			request:     createTestChatRequest(),
			setupMocks: func(providerRepo *MockProviderRepository, modelRepo *MockModelRepository, routingService *MockRoutingService, costService *MockCostService, providerClient *MockProviderClient) {
				// Setup routing decision
				decision := createTestRoutingDecision()
				routingRequest := &gateway.RoutingRequest{
					ModelName: "gpt-3.5-turbo",
				}
				routingService.On("RouteRequest", mock.Anything, mock.Anything, routingRequest).Return(decision, nil)

				// Setup provider response
				response := &gateway.ChatCompletionResponse{
					ID:     "chatcmpl-123",
					Object: "chat.completion",
					Model:  "gpt-3.5-turbo",
					Choices: []gateway.ChatChoice{
						{
							Index: 0,
							Message: gateway.ChatMessage{
								Role:    "assistant",
								Content: "Hello! How can I help you today?",
							},
							FinishReason: stringPtr("stop"),
						},
					},
					Usage: &gateway.TokenUsage{
						PromptTokens:     10,
						CompletionTokens: 15,
						TotalTokens:      25,
					},
				}
				providerClient.On("CreateChatCompletion", mock.Anything, mock.Anything).Return(response, nil)

				// Setup cost calculation
				costService.On("CalculateRequestCost", decision.ModelID, 10, 15).Return(0.002, nil)

				// Setup cost tracking
				costService.On("TrackRequestCost", mock.Anything, mock.AnythingOfType("*gateway.RequestMetrics")).Return(nil)
			},
		},
		{
			name:        "routing failure",
			projectID:   ulid.New(),
			environment: "production",
			request:     createTestChatRequest(),
			setupMocks: func(providerRepo *MockProviderRepository, modelRepo *MockModelRepository, routingService *MockRoutingService, costService *MockCostService, providerClient *MockProviderClient) {
				routingRequest := &gateway.RoutingRequest{
					ModelName: "gpt-3.5-turbo",
				}
				routingService.On("RouteRequest", mock.Anything, mock.Anything, routingRequest).Return(nil, errors.New("no available providers"))
			},
			expectedError: "failed to route request",
		},
		{
			name:        "provider client failure",
			projectID:   ulid.New(),
			environment: "production",
			request:     createTestChatRequest(),
			setupMocks: func(providerRepo *MockProviderRepository, modelRepo *MockModelRepository, routingService *MockRoutingService, costService *MockCostService, providerClient *MockProviderClient) {
				decision := createTestRoutingDecision()
				routingRequest := &gateway.RoutingRequest{
					ModelName: "gpt-3.5-turbo",
				}
				routingService.On("RouteRequest", mock.Anything, mock.Anything, routingRequest).Return(decision, nil)

				providerClient.On("CreateChatCompletion", mock.Anything, mock.Anything).Return(nil, errors.New("provider API error"))
			},
			expectedError: "failed to create chat completion",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			providerRepo := &MockProviderRepository{}
			modelRepo := &MockModelRepository{}
			routingService := &MockRoutingService{}
			costService := &MockCostService{}
			providerClient := &MockProviderClient{}

			tt.setupMocks(providerRepo, modelRepo, routingService, costService, providerClient)

			// Create service
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel) // Reduce noise during tests

			clients := map[string]gateway.ProviderClient{
				"openai": providerClient,
			}

			service := NewGatewayService(
				providerRepo,
				modelRepo,
				nil, // provider config repo not needed for this test
				routingService,
				costService,
				clients,
				logger,
			)

			// Execute test
			ctx := context.Background()
			response, err := service.CreateChatCompletion(ctx, tt.projectID, tt.request)

			// Verify results
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, response)
			} else {
				require.NoError(t, err)
				require.NotNil(t, response)
				assert.Equal(t, "chatcmpl-123", response.ID)
				assert.Equal(t, "gpt-3.5-turbo", response.Model)
				assert.Len(t, response.Choices, 1)
				assert.Equal(t, "Hello! How can I help you today?", response.Choices[0].Message.Content)
			}

			// Verify all expectations were met
			providerRepo.AssertExpectations(t)
			modelRepo.AssertExpectations(t)
			routingService.AssertExpectations(t)
			costService.AssertExpectations(t)
			providerClient.AssertExpectations(t)
		})
	}
}

func TestGatewayService_CreateChatCompletionStream(t *testing.T) {
	// Setup mocks
	providerRepo := &MockProviderRepository{}
	modelRepo := &MockModelRepository{}
	routingService := &MockRoutingService{}
	costService := &MockCostService{}
	providerClient := &MockProviderClient{}

	// Setup expectations
	decision := createTestRoutingDecision()
	routingRequest := &gateway.RoutingRequest{
		ModelName: "gpt-3.5-turbo",
	}
	routingService.On("RouteRequest", mock.Anything, mock.Anything, routingRequest).Return(decision, nil)

	var writer strings.Builder
	providerClient.On("CreateChatCompletionStream", mock.Anything, mock.Anything, mock.AnythingOfType("*strings.Builder")).Return(nil)

	// Create service
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	clients := map[string]gateway.ProviderClient{
		"openai": providerClient,
	}

	service := NewGatewayService(
		providerRepo,
		modelRepo,
		nil,
		routingService,
		costService,
		clients,
		logger,
	)

	// Execute test
	ctx := context.Background()
	request := createTestChatRequest()
	request.Stream = boolPtr(true)

	err := service.CreateChatCompletionStream(ctx, ulid.New(), request, &writer)

	// Verify results
	assert.NoError(t, err)

	// Verify all expectations were met
	routingService.AssertExpectations(t)
	providerClient.AssertExpectations(t)
}

func TestGatewayService_ListAvailableModels(t *testing.T) {
	tests := []struct {
		name        string
		projectID   ulid.ULID
		setupMocks  func(*MockProviderRepository, *MockModelRepository)
		expectError bool
		expectCount int
	}{
		{
			name:      "successful model listing",
			projectID: ulid.New(),
			setupMocks: func(providerRepo *MockProviderRepository, modelRepo *MockModelRepository) {
				models := []*gateway.Model{
					createTestModel(),
					{
						ID:          ulid.New(),
						Name:        "gpt-4",
						ProviderID:  ulid.New(),
						Type:        gateway.ModelTypeText,
						DisplayName: "GPT-4",
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					},
				}
				modelRepo.On("GetActiveModels", mock.Anything, mock.Anything).Return(models, nil)
			},
			expectCount: 2,
		},
		{
			name:      "repository error",
			projectID: ulid.New(),
			setupMocks: func(providerRepo *MockProviderRepository, modelRepo *MockModelRepository) {
				modelRepo.On("GetActiveModels", mock.Anything, mock.Anything).Return(nil, errors.New("database error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			providerRepo := &MockProviderRepository{}
			modelRepo := &MockModelRepository{}
			routingService := &MockRoutingService{}
			costService := &MockCostService{}

			tt.setupMocks(providerRepo, modelRepo)

			// Create service
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			service := NewGatewayService(
				providerRepo,
				modelRepo,
				nil,
				routingService,
				costService,
				map[string]gateway.ProviderClient{},
				logger,
			)

			// Execute test
			ctx := context.Background()
			models, err := service.ListAvailableModels(ctx, tt.projectID)

			// Verify results
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, models)
			} else {
				assert.NoError(t, err)
				assert.Len(t, models, tt.expectCount)
			}

			// Verify expectations
			modelRepo.AssertExpectations(t)
		})
	}
}

func TestGatewayService_GetRouteDecision(t *testing.T) {
	// Setup mocks
	providerRepo := &MockProviderRepository{}
	modelRepo := &MockModelRepository{}
	routingService := &MockRoutingService{}
	costService := &MockCostService{}

	decision := createTestRoutingDecision()
	strategy := gateway.RoutingStrategyCostOptimized
	routingService.On("GetBestProvider", mock.Anything, mock.Anything, "gpt-3.5-turbo", strategy).Return(decision, nil)

	// Create service
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewGatewayService(
		providerRepo,
		modelRepo,
		nil,
		routingService,
		costService,
		map[string]gateway.ProviderClient{},
		logger,
	)

	// Execute test
	ctx := context.Background()
	result, err := service.GetRouteDecision(ctx, ulid.New(), "gpt-3.5-turbo", &strategy)

	// Verify results
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, decision.ProviderName, result.ProviderName)
	assert.Equal(t, decision.ModelName, result.ModelName)
	assert.Equal(t, decision.Strategy, result.Strategy)

	// Verify expectations
	routingService.AssertExpectations(t)
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

// Benchmark tests
func BenchmarkGatewayService_CreateChatCompletion(b *testing.B) {
	// Setup mocks
	providerRepo := &MockProviderRepository{}
	modelRepo := &MockModelRepository{}
	routingService := &MockRoutingService{}
	costService := &MockCostService{}
	providerClient := &MockProviderClient{}

	decision := createTestRoutingDecision()
	routingService.On("RouteRequest", mock.Anything, mock.Anything, mock.Anything).Return(decision, nil)

	response := &gateway.ChatCompletionResponse{
		ID:     "chatcmpl-123",
		Object: "chat.completion",
		Model:  "gpt-3.5-turbo",
		Choices: []gateway.ChatChoice{
			{
				Index: 0,
				Message: gateway.ChatMessage{
					Role:    "assistant",
					Content: "Hello!",
				},
			},
		},
		Usage: &gateway.TokenUsage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}
	providerClient.On("CreateChatCompletion", mock.Anything, mock.Anything).Return(response, nil)
	costService.On("CalculateRequestCost", mock.Anything, 10, 5).Return(0.001, nil)
	costService.On("TrackRequestCost", mock.Anything, mock.Anything).Return(nil)

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	clients := map[string]gateway.ProviderClient{
		"openai": providerClient,
	}

	service := NewGatewayService(
		providerRepo,
		modelRepo,
		nil,
		routingService,
		costService,
		clients,
		logger,
	)

	request := createTestChatRequest()
	ctx := context.Background()
	projectID := ulid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.CreateChatCompletion(ctx, projectID, request)
		if err != nil {
			b.Fatal(err)
		}
	}
}
