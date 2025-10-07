package providers

import (
	"context"
	"io"
	"time"

	"brokle/internal/core/domain/gateway"
)

// Provider defines the common interface that all AI providers must implement
// This interface provides OpenAI-compatible methods that can be used across different providers
type Provider interface {
	// Provider identification
	GetName() string
	GetType() gateway.ProviderType
	
	// Core AI operations
	ChatCompletion(ctx context.Context, req *ChatCompletionRequest) (*ChatCompletionResponse, error)
	ChatCompletionStream(ctx context.Context, req *ChatCompletionRequest, writer io.Writer) error
	Completion(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)
	CompletionStream(ctx context.Context, req *CompletionRequest, writer io.Writer) error
	Embedding(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error)
	
	// Model operations
	ListModels(ctx context.Context) ([]*Model, error)
	GetModel(ctx context.Context, modelName string) (*Model, error)
	
	// Health and connectivity
	HealthCheck(ctx context.Context) error
	TestConnection(ctx context.Context) (*ConnectionTestResult, error)
	
	// Configuration
	SetAPIKey(apiKey string)
	SetBaseURL(baseURL string)
	SetTimeout(timeout time.Duration)
	SetMaxRetries(maxRetries int)
	
	// Provider-specific capabilities
	GetSupportedFeatures() []string
	SupportsStreaming() bool
	SupportsFunctions() bool
	SupportsVision() bool
	SupportsEmbeddings() bool
}

// ProviderConfig contains configuration for provider initialization
type ProviderConfig struct {
	APIKey          string        `json:"api_key"`
	BaseURL         string        `json:"base_url"`
	Timeout         time.Duration `json:"timeout"`
	MaxRetries      int           `json:"max_retries"`
	CustomHeaders   map[string]string `json:"custom_headers,omitempty"`
	OrganizationID  *string       `json:"organization_id,omitempty"`
	ProjectID       *string       `json:"project_id,omitempty"`
}

// Request/Response types that match OpenAI's API format
// These types are used internally by providers and can be converted from/to domain types

// ChatCompletionRequest represents a chat completion request
type ChatCompletionRequest struct {
	Model            string                 `json:"model"`
	Messages         []ChatMessage          `json:"messages"`
	MaxTokens        *int                   `json:"max_tokens,omitempty"`
	Temperature      *float64               `json:"temperature,omitempty"`
	TopP             *float64               `json:"top_p,omitempty"`
	N                *int                   `json:"n,omitempty"`
	Stream           bool                   `json:"stream,omitempty"`
	Stop             []string               `json:"stop,omitempty"`
	PresencePenalty  *float64               `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float64               `json:"frequency_penalty,omitempty"`
	LogitBias        map[string]interface{} `json:"logit_bias,omitempty"`
	User             *string                `json:"user,omitempty"`
	Functions        []Function             `json:"functions,omitempty"`
	FunctionCall     interface{}            `json:"function_call,omitempty"`
	Tools            []Tool                 `json:"tools,omitempty"`
	ToolChoice       interface{}            `json:"tool_choice,omitempty"`
	ResponseFormat   *ResponseFormat        `json:"response_format,omitempty"`
	Seed             *int                   `json:"seed,omitempty"`
}

// ChatMessage represents a message in a chat conversation
type ChatMessage struct {
	Role         string      `json:"role"`
	Content      interface{} `json:"content"`
	Name         *string     `json:"name,omitempty"`
	FunctionCall interface{} `json:"function_call,omitempty"`
	ToolCalls    []ToolCall  `json:"tool_calls,omitempty"`
	ToolCallID   *string     `json:"tool_call_id,omitempty"`
}

// ChatCompletionResponse represents the response from a chat completion
type ChatCompletionResponse struct {
	ID                string                    `json:"id"`
	Object            string                    `json:"object"`
	Created           int64                     `json:"created"`
	Model             string                    `json:"model"`
	Choices           []ChatCompletionChoice    `json:"choices"`
	Usage             *TokenUsage               `json:"usage,omitempty"`
	SystemFingerprint *string                   `json:"system_fingerprint,omitempty"`
}

// ChatCompletionChoice represents a choice in the chat completion response
type ChatCompletionChoice struct {
	Index        int          `json:"index"`
	Message      *ChatMessage `json:"message,omitempty"`
	Delta        *ChatMessage `json:"delta,omitempty"`
	FinishReason *string      `json:"finish_reason,omitempty"`
	Logprobs     interface{}  `json:"logprobs,omitempty"`
}

// CompletionRequest represents a text completion request
type CompletionRequest struct {
	Model            string    `json:"model"`
	Prompt           interface{} `json:"prompt"`
	MaxTokens        *int      `json:"max_tokens,omitempty"`
	Temperature      *float64  `json:"temperature,omitempty"`
	TopP             *float64  `json:"top_p,omitempty"`
	N                *int      `json:"n,omitempty"`
	Stream           bool      `json:"stream,omitempty"`
	Logprobs         *int      `json:"logprobs,omitempty"`
	Echo             bool      `json:"echo,omitempty"`
	Stop             []string  `json:"stop,omitempty"`
	PresencePenalty  *float64  `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float64  `json:"frequency_penalty,omitempty"`
	BestOf           *int      `json:"best_of,omitempty"`
	LogitBias        map[string]interface{} `json:"logit_bias,omitempty"`
	User             *string   `json:"user,omitempty"`
	Suffix           *string   `json:"suffix,omitempty"`
}

// CompletionResponse represents the response from a text completion
type CompletionResponse struct {
	ID      string             `json:"id"`
	Object  string             `json:"object"`
	Created int64              `json:"created"`
	Model   string             `json:"model"`
	Choices []CompletionChoice `json:"choices"`
	Usage   *TokenUsage        `json:"usage,omitempty"`
}

// CompletionChoice represents a choice in the completion response
type CompletionChoice struct {
	Text         string      `json:"text"`
	Index        int         `json:"index"`
	Logprobs     interface{} `json:"logprobs,omitempty"`
	FinishReason *string     `json:"finish_reason,omitempty"`
}

// EmbeddingRequest represents an embedding request
type EmbeddingRequest struct {
	Model          string      `json:"model"`
	Input          interface{} `json:"input"`
	EncodingFormat *string     `json:"encoding_format,omitempty"`
	Dimensions     *int        `json:"dimensions,omitempty"`
	User           *string     `json:"user,omitempty"`
}

// EmbeddingResponse represents the response from an embedding request
type EmbeddingResponse struct {
	Object string      `json:"object"`
	Data   []Embedding `json:"data"`
	Model  string      `json:"model"`
	Usage  *TokenUsage `json:"usage,omitempty"`
}

// Embedding represents a single embedding vector
type Embedding struct {
	Object    string    `json:"object"`
	Index     int       `json:"index"`
	Embedding []float64 `json:"embedding"`
}

// TokenUsage represents token usage information
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Function represents a function definition for function calling
type Function struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// Tool represents a tool definition for tool calling
type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

// ToolCall represents a tool call in a message
type ToolCall struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Function FunctionCall `json:"function"`
}

// FunctionCall represents a function call
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ResponseFormat represents the response format specification
type ResponseFormat struct {
	Type string `json:"type"`
}

// Model represents model information
type Model struct {
	ID         string                 `json:"id"`
	Object     string                 `json:"object"`
	Created    int64                  `json:"created"`
	OwnedBy    string                 `json:"owned_by"`
	Permission []ModelPermission      `json:"permission,omitempty"`
	Root       string                 `json:"root,omitempty"`
	Parent     *string                `json:"parent,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// ModelPermission represents model permissions
type ModelPermission struct {
	ID                 string  `json:"id"`
	Object             string  `json:"object"`
	Created            int64   `json:"created"`
	AllowCreateEngine  bool    `json:"allow_create_engine"`
	AllowSampling      bool    `json:"allow_sampling"`
	AllowLogprobs      bool    `json:"allow_logprobs"`
	AllowSearchIndices bool    `json:"allow_search_indices"`
	AllowView          bool    `json:"allow_view"`
	AllowFineTuning    bool    `json:"allow_fine_tuning"`
	Organization       string  `json:"organization"`
	Group              *string `json:"group"`
	IsBlocking         bool    `json:"is_blocking"`
}

// ConnectionTestResult represents the result of testing a provider connection
type ConnectionTestResult struct {
	Success      bool                   `json:"success"`
	LatencyMs    int64                  `json:"latency_ms"`
	Error        *string                `json:"error,omitempty"`
	StatusCode   *int                   `json:"status_code,omitempty"`
	ResponseData map[string]interface{} `json:"response_data,omitempty"`
	TestedAt     time.Time              `json:"tested_at"`
}

// ProviderError represents an error from a provider
type ProviderError struct {
	Code           string                 `json:"code"`
	Message        string                 `json:"message"`
	Type           string                 `json:"type,omitempty"`
	Param          *string                `json:"param,omitempty"`
	HTTPStatusCode int                    `json:"http_status_code"`
	Details        map[string]interface{} `json:"details,omitempty"`
	InnerError     error                  `json:"-"`
}

// Error implements the error interface
func (e *ProviderError) Error() string {
	if e.InnerError != nil {
		return e.Message + ": " + e.InnerError.Error()
	}
	return e.Message
}

// Unwrap returns the inner error for error unwrapping
func (e *ProviderError) Unwrap() error {
	return e.InnerError
}

// NewProviderError creates a new provider error
func NewProviderError(code, message string, httpStatusCode int) *ProviderError {
	return &ProviderError{
		Code:           code,
		Message:        message,
		HTTPStatusCode: httpStatusCode,
		Details:        make(map[string]interface{}),
	}
}

// NewProviderErrorWithCause creates a new provider error with an inner error
func NewProviderErrorWithCause(code, message string, httpStatusCode int, cause error) *ProviderError {
	return &ProviderError{
		Code:           code,
		Message:        message,
		HTTPStatusCode: httpStatusCode,
		InnerError:     cause,
		Details:        make(map[string]interface{}),
	}
}

// Factory function type for creating providers
type ProviderFactory func(config *ProviderConfig) (Provider, error)

// Registry of provider factories
var providerFactories = make(map[gateway.ProviderType]ProviderFactory)

// RegisterProvider registers a provider factory
func RegisterProvider(providerType gateway.ProviderType, factory ProviderFactory) {
	providerFactories[providerType] = factory
}

// CreateProvider creates a new provider instance
func CreateProvider(providerType gateway.ProviderType, config *ProviderConfig) (Provider, error) {
	factory, exists := providerFactories[providerType]
	if !exists {
		return nil, NewProviderError(
			"PROVIDER_NOT_SUPPORTED",
			"Provider type not supported: "+string(providerType),
			400,
		)
	}
	
	return factory(config)
}

// GetSupportedProviders returns a list of supported provider types
func GetSupportedProviders() []gateway.ProviderType {
	types := make([]gateway.ProviderType, 0, len(providerFactories))
	for providerType := range providerFactories {
		types = append(types, providerType)
	}
	return types
}