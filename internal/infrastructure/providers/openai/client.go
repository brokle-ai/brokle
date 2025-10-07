package openai

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"

	"brokle/internal/core/domain/gateway"
	"brokle/internal/infrastructure/providers"
)

// OpenAIProvider implements the Provider interface for OpenAI
type OpenAIProvider struct {
	client     *openai.Client
	logger     *logrus.Logger
	config     *providers.ProviderConfig
	name       string
	timeout    time.Duration
	maxRetries int
}

// NewOpenAIProvider creates a new OpenAI provider instance
func NewOpenAIProvider(config *providers.ProviderConfig) (providers.Provider, error) {
	if config.APIKey == "" {
		return nil, providers.NewProviderError(
			"MISSING_API_KEY",
			"OpenAI API key is required",
			400,
		)
	}

	// Create OpenAI client configuration
	clientConfig := openai.DefaultConfig(config.APIKey)
	
	if config.BaseURL != "" {
		clientConfig.BaseURL = config.BaseURL
	}

	if config.OrganizationID != nil {
		clientConfig.OrgID = *config.OrganizationID
	}

	// Set timeout
	timeout := 30 * time.Second
	if config.Timeout > 0 {
		timeout = config.Timeout
	}
	
	httpClient := &http.Client{
		Timeout: timeout,
	}
	clientConfig.HTTPClient = httpClient

	// Create the client
	client := openai.NewClientWithConfig(clientConfig)

	return &OpenAIProvider{
		client:     client,
		logger:     logrus.New(),
		config:     config,
		name:       "OpenAI",
		timeout:    timeout,
		maxRetries: config.MaxRetries,
	}, nil
}

// Provider identification methods

func (p *OpenAIProvider) GetName() string {
	return p.name
}

func (p *OpenAIProvider) GetType() gateway.ProviderType {
	return gateway.ProviderTypeOpenAI
}

// Core AI operations

func (p *OpenAIProvider) ChatCompletion(ctx context.Context, req *providers.ChatCompletionRequest) (*providers.ChatCompletionResponse, error) {
	// Transform request to OpenAI format
	openaiReq := p.transformChatCompletionRequest(req)
	
	// Add timeout context
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	// Make API call with retry logic
	var resp openai.ChatCompletionResponse
	var err error
	
	for attempt := 0; attempt <= p.maxRetries; attempt++ {
		resp, err = p.client.CreateChatCompletion(ctx, openaiReq)
		if err == nil {
			break
		}
		
		// Check if error is retryable
		if !p.isRetryableError(err) {
			break
		}
		
		// Wait before retry
		if attempt < p.maxRetries {
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}

	if err != nil {
		return nil, p.transformError(err)
	}

	// Transform response to provider format
	return p.transformChatCompletionResponse(&resp), nil
}

func (p *OpenAIProvider) ChatCompletionStream(ctx context.Context, req *providers.ChatCompletionRequest, writer io.Writer) error {
	// Transform request to OpenAI format
	openaiReq := p.transformChatCompletionRequest(req)
	openaiReq.Stream = true

	// Add timeout context
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	// Create streaming request
	stream, err := p.client.CreateChatCompletionStream(ctx, openaiReq)
	if err != nil {
		return p.transformError(err)
	}
	defer stream.Close()

	// Process stream
	for {
		response, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return p.transformError(err)
		}

		// Transform and write response chunk
		chunk := p.transformChatCompletionStreamResponse(&response)
		data := fmt.Sprintf("data: %s\n\n", p.marshalJSON(chunk))
		
		if _, err := writer.Write([]byte(data)); err != nil {
			return providers.NewProviderErrorWithCause(
				"STREAM_WRITE_ERROR",
				"Failed to write stream data",
				500,
				err,
			)
		}
	}

	// Write final message
	if _, err := writer.Write([]byte("data: [DONE]\n\n")); err != nil {
		return providers.NewProviderErrorWithCause(
			"STREAM_WRITE_ERROR",
			"Failed to write stream completion",
			500,
			err,
		)
	}

	return nil
}

func (p *OpenAIProvider) Completion(ctx context.Context, req *providers.CompletionRequest) (*providers.CompletionResponse, error) {
	// Transform request to OpenAI format
	openaiReq := p.transformCompletionRequest(req)
	
	// Add timeout context
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	// Make API call with retry logic
	var resp openai.CompletionResponse
	var err error
	
	for attempt := 0; attempt <= p.maxRetries; attempt++ {
		resp, err = p.client.CreateCompletion(ctx, openaiReq)
		if err == nil {
			break
		}
		
		// Check if error is retryable
		if !p.isRetryableError(err) {
			break
		}
		
		// Wait before retry
		if attempt < p.maxRetries {
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}

	if err != nil {
		return nil, p.transformError(err)
	}

	// Transform response to provider format
	return p.transformCompletionResponse(&resp), nil
}

func (p *OpenAIProvider) CompletionStream(ctx context.Context, req *providers.CompletionRequest, writer io.Writer) error {
	// Transform request to OpenAI format
	openaiReq := p.transformCompletionRequest(req)
	openaiReq.Stream = true

	// Add timeout context
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	// Create streaming request
	stream, err := p.client.CreateCompletionStream(ctx, openaiReq)
	if err != nil {
		return p.transformError(err)
	}
	defer stream.Close()

	// Process stream
	for {
		response, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return p.transformError(err)
		}

		// Transform and write response chunk
		chunk := p.transformCompletionStreamResponse(&response)
		data := fmt.Sprintf("data: %s\n\n", p.marshalJSON(chunk))
		
		if _, err := writer.Write([]byte(data)); err != nil {
			return providers.NewProviderErrorWithCause(
				"STREAM_WRITE_ERROR",
				"Failed to write stream data",
				500,
				err,
			)
		}
	}

	// Write final message
	if _, err := writer.Write([]byte("data: [DONE]\n\n")); err != nil {
		return providers.NewProviderErrorWithCause(
			"STREAM_WRITE_ERROR",
			"Failed to write stream completion",
			500,
			err,
		)
	}

	return nil
}

func (p *OpenAIProvider) Embedding(ctx context.Context, req *providers.EmbeddingRequest) (*providers.EmbeddingResponse, error) {
	// Transform request to OpenAI format
	openaiReq := p.transformEmbeddingRequest(req)
	
	// Add timeout context
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	// Make API call with retry logic
	var resp openai.EmbeddingResponse
	var err error
	
	for attempt := 0; attempt <= p.maxRetries; attempt++ {
		resp, err = p.client.CreateEmbeddings(ctx, openaiReq)
		if err == nil {
			break
		}
		
		// Check if error is retryable
		if !p.isRetryableError(err) {
			break
		}
		
		// Wait before retry
		if attempt < p.maxRetries {
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}

	if err != nil {
		return nil, p.transformError(err)
	}

	// Transform response to provider format
	return p.transformEmbeddingResponse(&resp), nil
}

// Model operations

func (p *OpenAIProvider) ListModels(ctx context.Context) ([]*providers.Model, error) {
	// Add timeout context
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	models, err := p.client.ListModels(ctx)
	if err != nil {
		return nil, p.transformError(err)
	}

	// Transform models to provider format
	result := make([]*providers.Model, len(models.Models))
	for i, model := range models.Models {
		result[i] = p.transformModel(&model)
	}

	return result, nil
}

func (p *OpenAIProvider) GetModel(ctx context.Context, modelName string) (*providers.Model, error) {
	// Add timeout context
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	model, err := p.client.RetrieveModel(ctx, modelName)
	if err != nil {
		return nil, p.transformError(err)
	}

	return p.transformModel(&model), nil
}

// Health and connectivity

func (p *OpenAIProvider) HealthCheck(ctx context.Context) error {
	// Simple health check by listing models
	_, err := p.ListModels(ctx)
	return err
}

func (p *OpenAIProvider) TestConnection(ctx context.Context) (*providers.ConnectionTestResult, error) {
	startTime := time.Now()
	
	err := p.HealthCheck(ctx)
	latency := time.Since(startTime).Milliseconds()

	result := &providers.ConnectionTestResult{
		Success:   err == nil,
		LatencyMs: latency,
		TestedAt:  startTime,
	}

	if err != nil {
		errMsg := err.Error()
		result.Error = &errMsg
		
		if providerErr, ok := err.(*providers.ProviderError); ok {
			result.StatusCode = &providerErr.HTTPStatusCode
		}
	}

	return result, nil
}

// Configuration

func (p *OpenAIProvider) SetAPIKey(apiKey string) {
	p.config.APIKey = apiKey
	
	// Update client configuration
	clientConfig := openai.DefaultConfig(apiKey)
	if p.config.BaseURL != "" {
		clientConfig.BaseURL = p.config.BaseURL
	}
	if p.config.OrganizationID != nil {
		clientConfig.OrgID = *p.config.OrganizationID
	}
	
	httpClient := &http.Client{
		Timeout: p.timeout,
	}
	clientConfig.HTTPClient = httpClient
	
	p.client = openai.NewClientWithConfig(clientConfig)
}

func (p *OpenAIProvider) SetBaseURL(baseURL string) {
	p.config.BaseURL = baseURL
	p.SetAPIKey(p.config.APIKey) // Recreate client with new base URL
}

func (p *OpenAIProvider) SetTimeout(timeout time.Duration) {
	p.timeout = timeout
	p.SetAPIKey(p.config.APIKey) // Recreate client with new timeout
}

func (p *OpenAIProvider) SetMaxRetries(maxRetries int) {
	p.maxRetries = maxRetries
}

// Provider-specific capabilities

func (p *OpenAIProvider) GetSupportedFeatures() []string {
	return []string{
		"chat_completions",
		"completions",
		"embeddings",
		"streaming",
		"function_calling",
		"tool_calling",
		"vision",
		"json_mode",
	}
}

func (p *OpenAIProvider) SupportsStreaming() bool {
	return true
}

func (p *OpenAIProvider) SupportsFunctions() bool {
	return true
}

func (p *OpenAIProvider) SupportsVision() bool {
	return true
}

func (p *OpenAIProvider) SupportsEmbeddings() bool {
	return true
}

// Helper methods for error handling and retries

func (p *OpenAIProvider) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for OpenAI API errors
	if apiErr, ok := err.(*openai.APIError); ok {
		// Retry on rate limits, timeouts, and server errors
		return apiErr.HTTPStatusCode == 429 || // Rate limit
			apiErr.HTTPStatusCode == 408 || // Request timeout
			apiErr.HTTPStatusCode >= 500 // Server errors
	}

	// Retry on context timeouts
	if err == context.DeadlineExceeded {
		return true
	}

	return false
}

func (p *OpenAIProvider) transformError(err error) *providers.ProviderError {
	if err == nil {
		return nil
	}

	// Handle OpenAI API errors
	if apiErr, ok := err.(*openai.APIError); ok {
		return providers.NewProviderErrorWithCause(
			apiErr.Type,
			apiErr.Message,
			apiErr.HTTPStatusCode,
			err,
		)
	}

	// Handle context timeout
	if err == context.DeadlineExceeded {
		return providers.NewProviderErrorWithCause(
			"REQUEST_TIMEOUT",
			"Request timed out",
			408,
			err,
		)
	}

	// Handle other errors
	return providers.NewProviderErrorWithCause(
		"PROVIDER_ERROR",
		"OpenAI provider error",
		500,
		err,
	)
}

// Factory function for OpenAI provider
func NewProvider(config *providers.ProviderConfig) (providers.Provider, error) {
	return NewOpenAIProvider(config)
}

// Register the OpenAI provider
func init() {
	providers.RegisterProvider(gateway.ProviderTypeOpenAI, NewProvider)
}