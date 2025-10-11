//go:build integration
// +build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"brokle/internal/app"
	"brokle/internal/config"
	"brokle/internal/core/domain/gateway"
	"brokle/internal/core/domain/providers"
	"brokle/internal/transport/http/handlers/ai"
	"brokle/pkg/ulid"
)

// GatewayIntegrationTestSuite provides a comprehensive test suite for the gateway
type GatewayIntegrationTestSuite struct {
	suite.Suite
	app        *app.Application
	router     *gin.Engine
	server     *httptest.Server
	testOrgID  ulid.ULID
	testAPIKey string
}

// SetupSuite sets up the test suite with a real application instance
func (suite *GatewayIntegrationTestSuite) SetupSuite() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Load test configuration
	cfg, err := config.Load()
	require.NoError(suite.T(), err)

	// Override configuration for testing
	cfg.Server.Port = 0 // Use random port
	cfg.Database.PostgreSQL.Database = cfg.Database.PostgreSQL.Database + "_test"
	cfg.Database.ClickHouse.Database = cfg.Database.ClickHouse.Database + "_test"

	// Initialize application
	suite.app, err = app.New(cfg)
	require.NoError(suite.T(), err)

	// Set up router
	suite.router = gin.New()
	suite.router.Use(gin.Recovery())

	// Register routes
	suite.setupRoutes()

	// Create test server
	suite.server = httptest.NewServer(suite.router)

	// Set up test data
	suite.setupTestData()
}

// TearDownSuite cleans up after the test suite
func (suite *GatewayIntegrationTestSuite) TearDownSuite() {
	if suite.server != nil {
		suite.server.Close()
	}
	if suite.app != nil {
		_ = suite.app.Shutdown(context.Background())
	}
}

// setupRoutes configures the test routes
func (suite *GatewayIntegrationTestSuite) setupRoutes() {
	// Health endpoints
	suite.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// AI API endpoints
	aiGroup := suite.router.Group("/v1")
	{
		aiHandler := ai.NewAIHandler(
			suite.app.Services.Gateway,
			suite.app.Logger,
		)

		aiGroup.POST("/chat/completions", aiHandler.ChatCompletions)
		aiGroup.POST("/completions", aiHandler.Completions)
		aiGroup.POST("/embeddings", aiHandler.Embeddings)
		aiGroup.GET("/models", aiHandler.Models)
	}
}

// setupTestData creates test organization and API key
func (suite *GatewayIntegrationTestSuite) setupTestData() {
	suite.testOrgID = ulid.New()
	suite.testAPIKey = "bk-test-" + ulid.New().String()

	// TODO: Create test organization and API key in database
	// This would normally be done through the organization service
}

// TestHealthEndpoint tests that the health endpoint works
func (suite *GatewayIntegrationTestSuite) TestHealthEndpoint() {
	resp, err := http.Get(suite.server.URL + "/health")
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "healthy", result["status"])
}

// TestChatCompletionsSuccess tests successful chat completions request
func (suite *GatewayIntegrationTestSuite) TestChatCompletionsSuccess() {
	// Create a mock OpenAI server
	mockServer := suite.createMockOpenAIServer()
	defer mockServer.Close()

	// Create chat completion request
	request := ai.ChatCompletionRequest{
		Model: "gpt-3.5-turbo",
		Messages: []ai.ChatMessage{
			{
				Role:    "user",
				Content: "Hello, how are you?",
			},
		},
		MaxTokens:   100,
		Temperature: 0.7,
		Stream:      false,
	}

	// Make request
	response := suite.makeChatCompletionRequest(request)
	defer response.Body.Close()

	// Verify response
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode)

	var result ai.ChatCompletionResponse
	err := json.NewDecoder(response.Body).Decode(&result)
	require.NoError(suite.T(), err)

	// Verify response structure
	assert.NotEmpty(suite.T(), result.ID)
	assert.Equal(suite.T(), "chat.completion", result.Object)
	assert.NotEmpty(suite.T(), result.Created)
	assert.Equal(suite.T(), "gpt-3.5-turbo", result.Model)
	assert.Len(suite.T(), result.Choices, 1)
	assert.Equal(suite.T(), "stop", result.Choices[0].FinishReason)
	assert.NotEmpty(suite.T(), result.Choices[0].Message.Content)

	// Verify usage tracking
	assert.Greater(suite.T(), result.Usage.PromptTokens, int32(0))
	assert.Greater(suite.T(), result.Usage.CompletionTokens, int32(0))
	assert.Greater(suite.T(), result.Usage.TotalTokens, int32(0))
}

// TestChatCompletionsStreaming tests streaming chat completions
func (suite *GatewayIntegrationTestSuite) TestChatCompletionsStreaming() {
	// Create a mock OpenAI server for streaming
	mockServer := suite.createMockStreamingServer()
	defer mockServer.Close()

	// Create streaming chat completion request
	request := ai.ChatCompletionRequest{
		Model: "gpt-3.5-turbo",
		Messages: []ai.ChatMessage{
			{
				Role:    "user",
				Content: "Tell me a short story",
			},
		},
		MaxTokens: 150,
		Stream:    true,
	}

	// Make streaming request
	response := suite.makeChatCompletionRequest(request)
	defer response.Body.Close()

	// Verify streaming response
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode)
	assert.Equal(suite.T(), "text/plain", response.Header.Get("Content-Type"))
	assert.Equal(suite.T(), "keep-alive", response.Header.Get("Connection"))

	// Read streaming chunks
	chunks := suite.readStreamingChunks(response.Body)
	assert.Greater(suite.T(), len(chunks), 0)

	// Verify streaming format
	for _, chunk := range chunks {
		if strings.HasPrefix(chunk, "data: ") && !strings.Contains(chunk, "[DONE]") {
			var streamResponse ai.ChatCompletionStreamResponse
			data := strings.TrimPrefix(chunk, "data: ")
			err := json.Unmarshal([]byte(data), &streamResponse)
			require.NoError(suite.T(), err)

			assert.Equal(suite.T(), "chat.completion.chunk", streamResponse.Object)
			assert.NotEmpty(suite.T(), streamResponse.ID)
		}
	}
}

// TestProviderRouting tests intelligent provider routing
func (suite *GatewayIntegrationTestSuite) TestProviderRouting() {
	testCases := []struct {
		name           string
		model          string
		expectedRoute  string
		routingReason  string
	}{
		{
			name:          "OpenAI model routes to OpenAI",
			model:         "gpt-4",
			expectedRoute: "openai",
			routingReason: "model_preference",
		},
		{
			name:          "Anthropic model routes to Anthropic",
			model:         "claude-3-haiku",
			expectedRoute: "anthropic",
			routingReason: "model_preference",
		},
		{
			name:          "Generic model uses cost optimization",
			model:         "text-davinci-003",
			expectedRoute: "openai",
			routingReason: "cost_optimization",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			request := ai.ChatCompletionRequest{
				Model: tc.model,
				Messages: []ai.ChatMessage{
					{
						Role:    "user",
						Content: "Test routing",
					},
				},
				MaxTokens: 50,
			}

			response := suite.makeChatCompletionRequest(request)
			defer response.Body.Close()

			assert.Equal(suite.T(), http.StatusOK, response.StatusCode)

			// Check routing headers (these would be added by our gateway)
			routedProvider := response.Header.Get("X-Brokle-Provider")
			routingReason := response.Header.Get("X-Brokle-Routing-Reason")

			assert.Equal(suite.T(), tc.expectedRoute, routedProvider)
			assert.Equal(suite.T(), tc.routingReason, routingReason)
		})
	}
}

// TestCostCalculation tests cost calculation and tracking
func (suite *GatewayIntegrationTestSuite) TestCostCalculation() {
	request := ai.ChatCompletionRequest{
		Model: "gpt-3.5-turbo",
		Messages: []ai.ChatMessage{
			{
				Role:    "user",
				Content: "Calculate my costs",
			},
		},
		MaxTokens: 100,
	}

	response := suite.makeChatCompletionRequest(request)
	defer response.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, response.StatusCode)

	// Check cost headers
	estimatedCost := response.Header.Get("X-Brokle-Estimated-Cost")
	actualCost := response.Header.Get("X-Brokle-Actual-Cost")
	currency := response.Header.Get("X-Brokle-Currency")

	assert.NotEmpty(suite.T(), estimatedCost)
	assert.NotEmpty(suite.T(), actualCost)
	assert.Equal(suite.T(), "USD", currency)

	// Verify costs are reasonable (should be small amounts)
	assert.Contains(suite.T(), estimatedCost, "0.00")
	assert.Contains(suite.T(), actualCost, "0.00")
}

// TestRateLimiting tests rate limiting functionality
func (suite *GatewayIntegrationTestSuite) TestRateLimiting() {
	// Make multiple requests rapidly
	const numRequests = 10
	const maxAllowed = 5 // Assuming rate limit of 5/minute for test

	var responses []*http.Response
	for i := 0; i < numRequests; i++ {
		request := ai.ChatCompletionRequest{
			Model: "gpt-3.5-turbo",
			Messages: []ai.ChatMessage{
				{
					Role:    "user",
					Content: fmt.Sprintf("Request %d", i),
				},
			},
			MaxTokens: 10,
		}

		response := suite.makeChatCompletionRequest(request)
		responses = append(responses, response)
	}

	// Check responses
	successCount := 0
	rateLimitedCount := 0

	for _, resp := range responses {
		if resp.StatusCode == http.StatusOK {
			successCount++
		} else if resp.StatusCode == http.StatusTooManyRequests {
			rateLimitedCount++
		}
		resp.Body.Close()
	}

	// Should have some rate limited requests
	assert.Greater(suite.T(), rateLimitedCount, 0)
	assert.LessOrEqual(suite.T(), successCount, maxAllowed)
}

// TestErrorHandling tests various error scenarios
func (suite *GatewayIntegrationTestSuite) TestErrorHandling() {
	testCases := []struct {
		name           string
		request        ai.ChatCompletionRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Invalid model",
			request: ai.ChatCompletionRequest{
				Model: "invalid-model",
				Messages: []ai.ChatMessage{
					{Role: "user", Content: "Test"},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "model not supported",
		},
		{
			name: "Empty messages",
			request: ai.ChatCompletionRequest{
				Model:    "gpt-3.5-turbo",
				Messages: []ai.ChatMessage{},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "messages cannot be empty",
		},
		{
			name: "Invalid API key",
			request: ai.ChatCompletionRequest{
				Model: "gpt-3.5-turbo",
				Messages: []ai.ChatMessage{
					{Role: "user", Content: "Test"},
				},
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid API key",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Use invalid API key for the third test case
			apiKey := suite.testAPIKey
			if tc.name == "Invalid API key" {
				apiKey = "invalid-key"
			}

			response := suite.makeChatCompletionRequestWithAPIKey(tc.request, apiKey)
			defer response.Body.Close()

			assert.Equal(suite.T(), tc.expectedStatus, response.StatusCode)

			var errorResponse map[string]interface{}
			err := json.NewDecoder(response.Body).Decode(&errorResponse)
			require.NoError(suite.T(), err)

			errorMessage := errorResponse["error"].(map[string]interface{})["message"].(string)
			assert.Contains(suite.T(), strings.ToLower(errorMessage), strings.ToLower(tc.expectedError))
		})
	}
}

// TestEmbeddingsEndpoint tests the embeddings endpoint
func (suite *GatewayIntegrationTestSuite) TestEmbeddingsEndpoint() {
	// Create embeddings request
	request := ai.EmbeddingRequest{
		Model: "text-embedding-ada-002",
		Input: []string{"Hello world", "Test embedding"},
	}

	jsonData, err := json.Marshal(request)
	require.NoError(suite.T(), err)

	// Make request
	httpReq, err := http.NewRequest("POST", suite.server.URL+"/v1/embeddings", bytes.NewBuffer(jsonData))
	require.NoError(suite.T(), err)

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+suite.testAPIKey)

	client := &http.Client{}
	response, err := client.Do(httpReq)
	require.NoError(suite.T(), err)
	defer response.Body.Close()

	// Verify response
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode)

	var result ai.EmbeddingResponse
	err = json.NewDecoder(response.Body).Decode(&result)
	require.NoError(suite.T(), err)

	// Verify response structure
	assert.Equal(suite.T(), "list", result.Object)
	assert.Len(suite.T(), result.Data, 2)
	assert.NotEmpty(suite.T(), result.Model)

	// Verify embeddings
	for i, embedding := range result.Data {
		assert.Equal(suite.T(), "embedding", embedding.Object)
		assert.Equal(suite.T(), i, embedding.Index)
		assert.Greater(suite.T(), len(embedding.Embedding), 0)
	}
}

// TestModelsEndpoint tests the models listing endpoint
func (suite *GatewayIntegrationTestSuite) TestModelsEndpoint() {
	// Make request
	response, err := http.Get(suite.server.URL + "/v1/models")
	require.NoError(suite.T(), err)
	defer response.Body.Close()

	// Verify response
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode)

	var result ai.ModelsResponse
	err = json.NewDecoder(response.Body).Decode(&result)
	require.NoError(suite.T(), err)

	// Verify response structure
	assert.Equal(suite.T(), "list", result.Object)
	assert.Greater(suite.T(), len(result.Data), 0)

	// Verify model structure
	for _, model := range result.Data {
		assert.Equal(suite.T(), "model", model.Object)
		assert.NotEmpty(suite.T(), model.ID)
		assert.NotEmpty(suite.T(), model.OwnedBy)
		assert.Greater(suite.T(), model.Created, int64(0))
	}
}

// TestCacheHitScenario tests semantic caching functionality
func (suite *GatewayIntegrationTestSuite) TestCacheHitScenario() {
	request := ai.ChatCompletionRequest{
		Model: "gpt-3.5-turbo",
		Messages: []ai.ChatMessage{
			{
				Role:    "user",
				Content: "What is the capital of France?",
			},
		},
		MaxTokens: 50,
	}

	// Make first request
	response1 := suite.makeChatCompletionRequest(request)
	defer response1.Body.Close()
	assert.Equal(suite.T(), http.StatusOK, response1.StatusCode)

	// Check cache miss
	cacheHit1 := response1.Header.Get("X-Brokle-Cache-Hit")
	assert.Equal(suite.T(), "false", cacheHit1)

	// Make identical second request
	response2 := suite.makeChatCompletionRequest(request)
	defer response2.Body.Close()
	assert.Equal(suite.T(), http.StatusOK, response2.StatusCode)

	// Check cache hit
	cacheHit2 := response2.Header.Get("X-Brokle-Cache-Hit")
	assert.Equal(suite.T(), "true", cacheHit2)
}

// TestAnalyticsCollection tests that analytics are properly collected
func (suite *GatewayIntegrationTestSuite) TestAnalyticsCollection() {
	request := ai.ChatCompletionRequest{
		Model: "gpt-3.5-turbo",
		Messages: []ai.ChatMessage{
			{
				Role:    "user",
				Content: "Test analytics collection",
			},
		},
		MaxTokens: 50,
	}

	response := suite.makeChatCompletionRequest(request)
	defer response.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, response.StatusCode)

	// Give some time for analytics to be processed
	time.Sleep(100 * time.Millisecond)

	// Check analytics headers
	requestID := response.Header.Get("X-Brokle-Request-ID")
	assert.NotEmpty(suite.T(), requestID)

	// TODO: Query analytics database to verify data was stored
	// This would require access to the analytics repository
}

// Helper methods

// makeChatCompletionRequest makes a chat completion request with default API key
func (suite *GatewayIntegrationTestSuite) makeChatCompletionRequest(request ai.ChatCompletionRequest) *http.Response {
	return suite.makeChatCompletionRequestWithAPIKey(request, suite.testAPIKey)
}

// makeChatCompletionRequestWithAPIKey makes a chat completion request with specific API key
func (suite *GatewayIntegrationTestSuite) makeChatCompletionRequestWithAPIKey(request ai.ChatCompletionRequest, apiKey string) *http.Response {
	jsonData, err := json.Marshal(request)
	require.NoError(suite.T(), err)

	httpReq, err := http.NewRequest("POST", suite.server.URL+"/v1/chat/completions", bytes.NewBuffer(jsonData))
	require.NoError(suite.T(), err)

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	response, err := client.Do(httpReq)
	require.NoError(suite.T(), err)

	return response
}

// createMockOpenAIServer creates a mock OpenAI-compatible server
func (suite *GatewayIntegrationTestSuite) createMockOpenAIServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := ai.ChatCompletionResponse{
			ID:      "chatcmpl-test-" + ulid.New().String(),
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "gpt-3.5-turbo",
			Choices: []ai.ChatCompletionChoice{
				{
					Index: 0,
					Message: ai.ChatMessage{
						Role:    "assistant",
						Content: "I'm doing well, thank you for asking! How can I help you today?",
					},
					FinishReason: "stop",
				},
			},
			Usage: ai.Usage{
				PromptTokens:     10,
				CompletionTokens: 20,
				TotalTokens:      30,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
}

// createMockStreamingServer creates a mock server for streaming responses
func (suite *GatewayIntegrationTestSuite) createMockStreamingServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		require.True(suite.T(), ok)

		// Send streaming chunks
		chunks := []string{
			"data: {\"id\":\"chatcmpl-test\",\"object\":\"chat.completion.chunk\",\"choices\":[{\"delta\":{\"content\":\"Once\"}}]}\n\n",
			"data: {\"id\":\"chatcmpl-test\",\"object\":\"chat.completion.chunk\",\"choices\":[{\"delta\":{\"content\":\" upon\"}}]}\n\n",
			"data: {\"id\":\"chatcmpl-test\",\"object\":\"chat.completion.chunk\",\"choices\":[{\"delta\":{\"content\":\" a\"}}]}\n\n",
			"data: {\"id\":\"chatcmpl-test\",\"object\":\"chat.completion.chunk\",\"choices\":[{\"delta\":{\"content\":\" time...\"}}]}\n\n",
			"data: [DONE]\n\n",
		}

		for _, chunk := range chunks {
			w.Write([]byte(chunk))
			flusher.Flush()
			time.Sleep(10 * time.Millisecond)
		}
	}))
}

// readStreamingChunks reads all chunks from a streaming response
func (suite *GatewayIntegrationTestSuite) readStreamingChunks(body io.Reader) []string {
	var chunks []string
	buf := make([]byte, 1024)

	for {
		n, err := body.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			require.NoError(suite.T(), err)
		}

		content := string(buf[:n])
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				chunks = append(chunks, line)
			}
		}

		if strings.Contains(content, "[DONE]") {
			break
		}
	}

	return chunks
}

// TestGatewayIntegrationSuite runs the complete integration test suite
func TestGatewayIntegrationSuite(t *testing.T) {
	suite.Run(t, new(GatewayIntegrationTestSuite))
}