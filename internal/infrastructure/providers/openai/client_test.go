package openai

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"brokle/internal/core/domain/gateway"
	"brokle/internal/infrastructure/providers"
)

// Test fixtures
var (
	validAPIKey        = "sk-test-key-123"
	invalidAPIKey      = ""
	testBaseURL        = "https://api.openai.com/v1"
	testOrgID          = "org-test-123"
	testTimeout        = 30 * time.Second
	testMaxRetries     = 3
	testModel          = "gpt-3.5-turbo"
	testEmbeddingModel = "text-embedding-ada-002"
)

// Mock HTTP server responses
const (
	chatCompletionResponse = `{
		"id": "chatcmpl-test-123",
		"object": "chat.completion",
		"created": 1677610602,
		"model": "gpt-3.5-turbo-0301",
		"choices": [{
			"index": 0,
			"message": {
				"role": "assistant",
				"content": "Hello! How can I help you today?"
			},
			"finish_reason": "stop"
		}],
		"usage": {
			"prompt_tokens": 13,
			"completion_tokens": 7,
			"total_tokens": 20
		}
	}`

	completionResponse = `{
		"id": "cmpl-test-123",
		"object": "text_completion",
		"created": 1677610602,
		"model": "text-davinci-003",
		"choices": [{
			"text": " World!",
			"index": 0,
			"logprobs": null,
			"finish_reason": "stop"
		}],
		"usage": {
			"prompt_tokens": 1,
			"completion_tokens": 2,
			"total_tokens": 3
		}
	}`

	embeddingResponse = `{
		"object": "list",
		"data": [{
			"object": "embedding",
			"index": 0,
			"embedding": [0.1, 0.2, 0.3, 0.4, 0.5]
		}],
		"model": "text-embedding-ada-002",
		"usage": {
			"prompt_tokens": 5,
			"total_tokens": 5
		}
	}`

	modelsResponse = `{
		"object": "list",
		"data": [{
			"id": "gpt-3.5-turbo",
			"object": "model",
			"created": 1677610602,
			"owned_by": "openai",
			"root": "gpt-3.5-turbo",
			"parent": null,
			"permission": []
		}]
	}`

	modelResponse = `{
		"id": "gpt-3.5-turbo",
		"object": "model",
		"created": 1677610602,
		"owned_by": "openai",
		"root": "gpt-3.5-turbo",
		"parent": null,
		"permission": []
	}`

	errorResponse = `{
		"error": {
			"message": "Invalid API key",
			"type": "invalid_request_error",
			"param": null,
			"code": "invalid_api_key"
		}
	}`

	rateLimitErrorResponse = `{
		"error": {
			"message": "Rate limit exceeded",
			"type": "rate_limit_exceeded",
			"param": null,
			"code": "rate_limit_exceeded"
		}
	}`
)

func createValidConfig() *providers.ProviderConfig {
	return &providers.ProviderConfig{
		APIKey:         validAPIKey,
		BaseURL:        testBaseURL,
		Timeout:        testTimeout,
		MaxRetries:     testMaxRetries,
		OrganizationID: &testOrgID,
	}
}

func createInvalidConfig() *providers.ProviderConfig {
	return &providers.ProviderConfig{
		APIKey:     invalidAPIKey,
		BaseURL:    testBaseURL,
		Timeout:    testTimeout,
		MaxRetries: testMaxRetries,
	}
}

func TestNewOpenAIProvider(t *testing.T) {
	tests := []struct {
		name        string
		config      *providers.ProviderConfig
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid configuration",
			config:      createValidConfig(),
			expectError: false,
		},
		{
			name:        "Missing API key",
			config:      createInvalidConfig(),
			expectError: true,
			errorMsg:    "OpenAI API key is required",
		},
		{
			name:        "Nil configuration",
			config:      nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewOpenAIProvider(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, provider)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, provider)

				// Verify provider properties
				openaiProvider := provider.(*OpenAIProvider)
				assert.Equal(t, "OpenAI", openaiProvider.GetName())
				assert.Equal(t, gateway.ProviderTypeOpenAI, openaiProvider.GetType())
				assert.Equal(t, tt.config.APIKey, openaiProvider.config.APIKey)
				assert.Equal(t, tt.config.Timeout, openaiProvider.timeout)
				assert.Equal(t, tt.config.MaxRetries, openaiProvider.maxRetries)
			}
		})
	}
}

func TestOpenAIProvider_ChatCompletion(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/chat/completions", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Contains(t, r.Header.Get("Authorization"), "Bearer "+validAPIKey)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(chatCompletionResponse))
	}))
	defer server.Close()

	// Create provider with mock server
	config := createValidConfig()
	config.BaseURL = server.URL
	provider, err := NewOpenAIProvider(config)
	require.NoError(t, err)

	// Create test request
	req := &providers.ChatCompletionRequest{
		Model: testModel,
		Messages: []providers.ChatMessage{
			{
				Role:    "user",
				Content: "Hello, how are you?",
			},
		},
		MaxTokens:   &[]int{100}[0],
		Temperature: &[]float64{0.7}[0],
	}

	// Test successful request
	resp, err := provider.ChatCompletion(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "chatcmpl-test-123", resp.ID)
	assert.Equal(t, "chat.completion", resp.Object)
	assert.Equal(t, "gpt-3.5-turbo-0301", resp.Model)
	assert.Len(t, resp.Choices, 1)
	assert.Equal(t, "assistant", resp.Choices[0].Message.Role)
	assert.Equal(t, "Hello! How can I help you today?", resp.Choices[0].Message.Content)
	assert.NotNil(t, resp.Usage)
	assert.Equal(t, 13, resp.Usage.PromptTokens)
	assert.Equal(t, 7, resp.Usage.CompletionTokens)
	assert.Equal(t, 20, resp.Usage.TotalTokens)
}

func TestOpenAIProvider_ChatCompletion_Error(t *testing.T) {
	// Create a mock HTTP server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(errorResponse))
	}))
	defer server.Close()

	// Create provider with mock server
	config := createValidConfig()
	config.BaseURL = server.URL
	provider, err := NewOpenAIProvider(config)
	require.NoError(t, err)

	// Create test request
	req := &providers.ChatCompletionRequest{
		Model: testModel,
		Messages: []providers.ChatMessage{
			{
				Role:    "user",
				Content: "Hello, how are you?",
			},
		},
	}

	// Test error response
	resp, err := provider.ChatCompletion(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.IsType(t, &providers.ProviderError{}, err)

	providerErr := err.(*providers.ProviderError)
	assert.Equal(t, 401, providerErr.HTTPStatusCode)
	assert.Contains(t, providerErr.Message, "Invalid API key")
}

func TestOpenAIProvider_ChatCompletion_Retry(t *testing.T) {
	// Track number of requests
	requestCount := 0

	// Create a mock HTTP server that fails twice then succeeds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		if requestCount <= 2 {
			// First two requests fail with rate limit
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(rateLimitErrorResponse))
		} else {
			// Third request succeeds
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(chatCompletionResponse))
		}
	}))
	defer server.Close()

	// Create provider with mock server
	config := createValidConfig()
	config.BaseURL = server.URL
	provider, err := NewOpenAIProvider(config)
	require.NoError(t, err)

	// Create test request
	req := &providers.ChatCompletionRequest{
		Model: testModel,
		Messages: []providers.ChatMessage{
			{
				Role:    "user",
				Content: "Hello, how are you?",
			},
		},
	}

	// Test successful request after retries
	resp, err := provider.ChatCompletion(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 3, requestCount) // Should have made 3 requests
}

func TestOpenAIProvider_ChatCompletionStream(t *testing.T) {
	// Create a mock HTTP server that returns streaming data
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/chat/completions", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.Header.Get("Authorization"), "Bearer "+validAPIKey)

		// Simulate streaming response
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// Write streaming chunks
		chunks := []string{
			`data: {"id":"chatcmpl-test","object":"chat.completion.chunk","created":1677610602,"model":"gpt-3.5-turbo","choices":[{"index":0,"delta":{"role":"assistant","content":""},"finish_reason":null}]}`,
			`data: {"id":"chatcmpl-test","object":"chat.completion.chunk","created":1677610602,"model":"gpt-3.5-turbo","choices":[{"index":0,"delta":{"content":"Hello"},"finish_reason":null}]}`,
			`data: {"id":"chatcmpl-test","object":"chat.completion.chunk","created":1677610602,"model":"gpt-3.5-turbo","choices":[{"index":0,"delta":{"content":"!"},"finish_reason":"stop"}]}`,
			`data: [DONE]`,
		}

		for _, chunk := range chunks {
			w.Write([]byte(chunk + "\n\n"))
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			time.Sleep(10 * time.Millisecond) // Small delay to simulate streaming
		}
	}))
	defer server.Close()

	// Create provider with mock server
	config := createValidConfig()
	config.BaseURL = server.URL
	provider, err := NewOpenAIProvider(config)
	require.NoError(t, err)

	// Create test request
	req := &providers.ChatCompletionRequest{
		Model: testModel,
		Messages: []providers.ChatMessage{
			{
				Role:    "user",
				Content: "Hello, how are you?",
			},
		},
	}

	// Test streaming request
	var buf bytes.Buffer
	err = provider.ChatCompletionStream(context.Background(), req, &buf)

	assert.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "data:")
	assert.Contains(t, output, "Hello")
	assert.Contains(t, output, "[DONE]")
}

func TestOpenAIProvider_Completion(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/completions", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(completionResponse))
	}))
	defer server.Close()

	// Create provider with mock server
	config := createValidConfig()
	config.BaseURL = server.URL
	provider, err := NewOpenAIProvider(config)
	require.NoError(t, err)

	// Create test request
	req := &providers.CompletionRequest{
		Model:       "text-davinci-003",
		Prompt:      "Hello",
		MaxTokens:   &[]int{10}[0],
		Temperature: &[]float64{0.5}[0],
	}

	// Test successful request
	resp, err := provider.Completion(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "cmpl-test-123", resp.ID)
	assert.Equal(t, "text_completion", resp.Object)
	assert.Equal(t, "text-davinci-003", resp.Model)
	assert.Len(t, resp.Choices, 1)
	assert.Equal(t, " World!", resp.Choices[0].Text)
	assert.NotNil(t, resp.Usage)
	assert.Equal(t, 1, resp.Usage.PromptTokens)
	assert.Equal(t, 2, resp.Usage.CompletionTokens)
	assert.Equal(t, 3, resp.Usage.TotalTokens)
}

func TestOpenAIProvider_Embedding(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/embeddings", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(embeddingResponse))
	}))
	defer server.Close()

	// Create provider with mock server
	config := createValidConfig()
	config.BaseURL = server.URL
	provider, err := NewOpenAIProvider(config)
	require.NoError(t, err)

	// Create test request
	req := &providers.EmbeddingRequest{
		Model: testEmbeddingModel,
		Input: "Hello, world!",
	}

	// Test successful request
	resp, err := provider.Embedding(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "list", resp.Object)
	assert.Equal(t, "text-embedding-ada-002", resp.Model)
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "embedding", resp.Data[0].Object)
	assert.Equal(t, 0, resp.Data[0].Index)
	// Use InDelta for float comparisons to account for float32→float64 precision loss
	expectedEmbedding := []float64{0.1, 0.2, 0.3, 0.4, 0.5}
	assert.Len(t, resp.Data[0].Embedding, len(expectedEmbedding))
	for i, expected := range expectedEmbedding {
		assert.InDelta(t, expected, resp.Data[0].Embedding[i], 0.0001)
	}
	assert.NotNil(t, resp.Usage)
	assert.Equal(t, 5, resp.Usage.PromptTokens)
	assert.Equal(t, 5, resp.Usage.TotalTokens)
}

func TestOpenAIProvider_ListModels(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/models", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(modelsResponse))
	}))
	defer server.Close()

	// Create provider with mock server
	config := createValidConfig()
	config.BaseURL = server.URL
	provider, err := NewOpenAIProvider(config)
	require.NoError(t, err)

	// Test successful request
	models, err := provider.ListModels(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, models)
	assert.Len(t, models, 1)
	assert.Equal(t, "gpt-3.5-turbo", models[0].ID)
	assert.Equal(t, "model", models[0].Object)
	assert.Equal(t, "openai", models[0].OwnedBy)
}

func TestOpenAIProvider_GetModel(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/models/gpt-3.5-turbo", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(modelResponse))
	}))
	defer server.Close()

	// Create provider with mock server
	config := createValidConfig()
	config.BaseURL = server.URL
	provider, err := NewOpenAIProvider(config)
	require.NoError(t, err)

	// Test successful request
	model, err := provider.GetModel(context.Background(), "gpt-3.5-turbo")

	assert.NoError(t, err)
	assert.NotNil(t, model)
	assert.Equal(t, "gpt-3.5-turbo", model.ID)
	assert.Equal(t, "model", model.Object)
	assert.Equal(t, "openai", model.OwnedBy)
}

func TestOpenAIProvider_HealthCheck(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(modelsResponse))
	}))
	defer server.Close()

	// Create provider with mock server
	config := createValidConfig()
	config.BaseURL = server.URL
	provider, err := NewOpenAIProvider(config)
	require.NoError(t, err)

	// Test health check
	err = provider.HealthCheck(context.Background())
	assert.NoError(t, err)
}

func TestOpenAIProvider_TestConnection(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		expectSuccess  bool
	}{
		{
			name: "Successful connection",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(modelsResponse))
			},
			expectSuccess: true,
		},
		{
			name: "Failed connection",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(errorResponse))
			},
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock HTTP server
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()

			// Create provider with mock server
			config := createValidConfig()
			config.BaseURL = server.URL
			provider, err := NewOpenAIProvider(config)
			require.NoError(t, err)

			// Test connection
			result, err := provider.TestConnection(context.Background())

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expectSuccess, result.Success)
			assert.GreaterOrEqual(t, result.LatencyMs, int64(0)) // Latency can be 0 for fast mock responses
			assert.NotZero(t, result.TestedAt)

			if !tt.expectSuccess {
				assert.NotNil(t, result.Error)
				assert.Contains(t, *result.Error, "Invalid API key")
			} else {
				assert.Nil(t, result.Error)
			}
		})
	}
}

func TestOpenAIProvider_Capabilities(t *testing.T) {
	provider, err := NewOpenAIProvider(createValidConfig())
	require.NoError(t, err)

	// Test supported features
	features := provider.GetSupportedFeatures()
	expectedFeatures := []string{
		"chat_completions",
		"completions",
		"embeddings",
		"streaming",
		"function_calling",
		"tool_calling",
		"vision",
		"json_mode",
	}
	assert.ElementsMatch(t, expectedFeatures, features)

	// Test capability checks
	assert.True(t, provider.SupportsStreaming())
	assert.True(t, provider.SupportsFunctions())
	assert.True(t, provider.SupportsVision())
	assert.True(t, provider.SupportsEmbeddings())
}

func TestOpenAIProvider_ContextTimeout(t *testing.T) {
	// Create provider with short timeout
	config := createValidConfig()
	config.Timeout = 10 * time.Millisecond
	provider, err := NewOpenAIProvider(config)
	require.NoError(t, err)

	// Create a slow mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond) // Longer than timeout
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Update provider base URL
	provider.SetBaseURL(server.URL)

	// Create test request
	req := &providers.ChatCompletionRequest{
		Model: testModel,
		Messages: []providers.ChatMessage{
			{Role: "user", Content: "Hello"},
		},
	}

	// Test timeout error
	resp, err := provider.ChatCompletion(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.IsType(t, &providers.ProviderError{}, err)

	providerErr := err.(*providers.ProviderError)
	assert.Equal(t, 408, providerErr.HTTPStatusCode)
	assert.Contains(t, providerErr.Message, "Request timed out")
}

func TestOpenAIProvider_RetryableErrors(t *testing.T) {
	provider, err := NewOpenAIProvider(createValidConfig())
	require.NoError(t, err)

	openaiProvider := provider.(*OpenAIProvider)

	// Test retryable errors
	tests := []struct {
		name        string
		error       error
		isRetryable bool
	}{
		{
			name:        "Rate limit error",
			error:       &providers.ProviderError{HTTPStatusCode: 429},
			isRetryable: true,
		},
		{
			name:        "Request timeout",
			error:       &providers.ProviderError{HTTPStatusCode: 408},
			isRetryable: true,
		},
		{
			name:        "Server error",
			error:       &providers.ProviderError{HTTPStatusCode: 500},
			isRetryable: true,
		},
		{
			name:        "Context timeout",
			error:       context.DeadlineExceeded,
			isRetryable: true,
		},
		{
			name:        "Client error",
			error:       &providers.ProviderError{HTTPStatusCode: 400},
			isRetryable: false,
		},
		{
			name:        "Unauthorized error",
			error:       &providers.ProviderError{HTTPStatusCode: 401},
			isRetryable: false,
		},
		{
			name:        "Not found error",
			error:       &providers.ProviderError{HTTPStatusCode: 404},
			isRetryable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := openaiProvider.isRetryableError(tt.error)
			assert.Equal(t, tt.isRetryable, result)
		})
	}
}

// Benchmark tests
func BenchmarkOpenAIProvider_ChatCompletion(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(chatCompletionResponse))
	}))
	defer server.Close()

	config := createValidConfig()
	config.BaseURL = server.URL
	provider, err := NewOpenAIProvider(config)
	require.NoError(b, err)

	req := &providers.ChatCompletionRequest{
		Model: testModel,
		Messages: []providers.ChatMessage{
			{Role: "user", Content: "Hello"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := provider.ChatCompletion(context.Background(), req)
		require.NoError(b, err)
	}
}
