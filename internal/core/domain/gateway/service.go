package gateway

import (
	"context"
	"io"
	"time"

	"brokle/pkg/ulid"
)

// GatewayService defines the main gateway service interface for AI requests
type GatewayService interface {
	// Chat completion operations
	CreateChatCompletion(ctx context.Context, projectID ulid.ULID, environment string, req *ChatCompletionRequest) (*ChatCompletionResponse, error)
	CreateChatCompletionStream(ctx context.Context, projectID ulid.ULID, environment string, req *ChatCompletionRequest, writer io.Writer) error

	// Text completion operations
	CreateCompletion(ctx context.Context, projectID ulid.ULID, environment string, req *CompletionRequest) (*CompletionResponse, error)
	CreateCompletionStream(ctx context.Context, projectID ulid.ULID, environment string, req *CompletionRequest, writer io.Writer) error

	// Embedding operations
	CreateEmbedding(ctx context.Context, projectID ulid.ULID, environment string, req *EmbeddingRequest) (*EmbeddingResponse, error)

	// Model operations
	ListAvailableModels(ctx context.Context, projectID ulid.ULID) ([]*ModelInfo, error)
	GetModel(ctx context.Context, projectID ulid.ULID, modelName string) (*ModelInfo, error)

	// Route decision operations (for debugging/analysis)
	GetRouteDecision(ctx context.Context, projectID ulid.ULID, modelName string, strategy *string) (*RoutingDecision, error)

	// Health operations
	GetProviderHealth(ctx context.Context, projectID ulid.ULID) ([]*ProviderHealth, error)
	TestProviderConnection(ctx context.Context, projectID ulid.ULID, providerID ulid.ULID) (*ConnectionTestResult, error)
}

// ProviderService defines the interface for provider management
type ProviderService interface {
	// Provider CRUD operations
	CreateProvider(ctx context.Context, provider *Provider) error
	GetProvider(ctx context.Context, id ulid.ULID) (*Provider, error)
	GetProviderByName(ctx context.Context, name string) (*Provider, error)
	UpdateProvider(ctx context.Context, provider *Provider) error
	DeleteProvider(ctx context.Context, id ulid.ULID) error

	// Provider listing and search
	ListProviders(ctx context.Context, filter *ProviderFilter, limit, offset int) ([]*Provider, int, error)
	SearchProviders(ctx context.Context, query string, filter *ProviderFilter) ([]*Provider, error)

	// Provider status management
	EnableProvider(ctx context.Context, id ulid.ULID) error
	DisableProvider(ctx context.Context, id ulid.ULID) error
	UpdateProviderHealth(ctx context.Context, id ulid.ULID, health *ProviderHealthUpdate) error

	// Provider validation
	ValidateProvider(ctx context.Context, provider *Provider) error
	TestProviderEndpoint(ctx context.Context, id ulid.ULID) error
}

// ModelService defines the interface for model management
type ModelService interface {
	// Model CRUD operations
	CreateModel(ctx context.Context, model *Model) error
	GetModel(ctx context.Context, id ulid.ULID) (*Model, error)
	GetModelByName(ctx context.Context, modelName string) (*Model, error)
	UpdateModel(ctx context.Context, model *Model) error
	DeleteModel(ctx context.Context, id ulid.ULID) error

	// Model listing and search
	ListModels(ctx context.Context, filter *ModelFilter, limit, offset int) ([]*Model, int, error)
	SearchModels(ctx context.Context, query string, filter *ModelFilter) ([]*Model, error)

	// Model availability
	GetAvailableModels(ctx context.Context, projectID ulid.ULID, requirements *ModelRequirements) ([]*Model, error)
	GetCompatibleModels(ctx context.Context, requirements *ModelRequirements) ([]*Model, error)

	// Model pricing and performance
	CalculateModelCost(ctx context.Context, modelID ulid.ULID, inputTokens, outputTokens int) (float64, error)
	GetModelPerformanceMetrics(ctx context.Context, modelID ulid.ULID, timeRange *TimeRange) (*ModelPerformanceMetrics, error)
	CompareModels(ctx context.Context, modelIDs []ulid.ULID, criteria *ComparisonCriteria) (*ModelComparison, error)

	// Model validation
	ValidateModel(ctx context.Context, model *Model) error
	ValidateModelCompatibility(ctx context.Context, modelID, providerID ulid.ULID) error
}

// ProviderConfigService defines the interface for provider configuration management
type ProviderConfigService interface {
	// Configuration CRUD operations
	CreateProviderConfig(ctx context.Context, config *ProviderConfig) error
	GetProviderConfig(ctx context.Context, id ulid.ULID) (*ProviderConfig, error)
	GetProjectProviderConfig(ctx context.Context, projectID, providerID ulid.ULID) (*ProviderConfig, error)
	UpdateProviderConfig(ctx context.Context, config *ProviderConfig) error
	DeleteProviderConfig(ctx context.Context, id ulid.ULID) error

	// Project-scoped operations
	ListProjectProviderConfigs(ctx context.Context, projectID ulid.ULID) ([]*ProviderConfig, error)
	GetEnabledProviderConfigs(ctx context.Context, projectID ulid.ULID) ([]*ProviderConfig, error)

	// API key management
	SetProviderAPIKey(ctx context.Context, projectID, providerID ulid.ULID, apiKey string) error
	RotateProviderAPIKey(ctx context.Context, configID ulid.ULID, newAPIKey string) error
	TestProviderAPIKey(ctx context.Context, configID ulid.ULID) error

	// Priority management
	SetProviderPriority(ctx context.Context, configID ulid.ULID, priority int) error
	ReorderProviderConfigs(ctx context.Context, projectID ulid.ULID, configIDs []ulid.ULID) error

	// Configuration validation
	ValidateProviderConfig(ctx context.Context, config *ProviderConfig) error
	TestProviderConfiguration(ctx context.Context, config *ProviderConfig) error
}

// RoutingService defines the interface for intelligent request routing
type RoutingService interface {
	// Route decision making
	RouteRequest(ctx context.Context, projectID ulid.ULID, request *RoutingRequest) (*RoutingDecision, error)
	GetBestProvider(ctx context.Context, projectID ulid.ULID, modelName string, strategy RoutingStrategy) (*RoutingDecision, error)

	// Routing strategy implementation
	RouteByCost(ctx context.Context, projectID ulid.ULID, modelName string) (*RoutingDecision, error)
	RouteByLatency(ctx context.Context, projectID ulid.ULID, modelName string) (*RoutingDecision, error)
	RouteByQuality(ctx context.Context, projectID ulid.ULID, modelName string) (*RoutingDecision, error)
	RouteByLoad(ctx context.Context, projectID ulid.ULID, modelName string) (*RoutingDecision, error)

	// Fallback handling
	GetFallbackProvider(ctx context.Context, projectID ulid.ULID, failedProviderID ulid.ULID, modelName string) (*RoutingDecision, error)
	HandleProviderFailure(ctx context.Context, projectID ulid.ULID, providerID ulid.ULID, errorType ErrorType) (*RoutingDecision, error)

	// Routing rules management
	CreateRoutingRule(ctx context.Context, rule *RoutingRule) error
	UpdateRoutingRule(ctx context.Context, rule *RoutingRule) error
	DeleteRoutingRule(ctx context.Context, ruleID ulid.ULID) error
	ListProjectRoutingRules(ctx context.Context, projectID ulid.ULID) ([]*RoutingRule, error)

	// Route testing and analysis
	TestRoute(ctx context.Context, projectID ulid.ULID, request *RoutingRequest) (*RouteTestResult, error)
	AnalyzeRoutingPerformance(ctx context.Context, projectID ulid.ULID, timeRange *TimeRange) (*RoutingAnalysis, error)
}

// CostService defines the interface for cost calculation and management
type CostService interface {
	// Cost calculation
	CalculateRequestCost(ctx context.Context, modelID ulid.ULID, inputTokens, outputTokens int) (float64, error)
	EstimateRequestCost(ctx context.Context, modelName string, estimatedTokens int) (float64, error)
	CalculateBatchCost(ctx context.Context, requests []*CostCalculationRequest) (*BatchCostResult, error)

	// Cost optimization
	GetCostOptimizedProvider(ctx context.Context, projectID ulid.ULID, modelName string) (*RoutingDecision, error)
	CompareCosts(ctx context.Context, modelNames []string, tokenCount int) (*CostComparison, error)
	GetCostSavingsReport(ctx context.Context, projectID ulid.ULID, timeRange *TimeRange) (*CostSavingsReport, error)

	// Cost tracking and analytics
	TrackRequestCost(ctx context.Context, metrics *RequestMetrics) error
	GetProjectCostAnalytics(ctx context.Context, projectID ulid.ULID, timeRange *TimeRange) (*CostAnalytics, error)
	GetProviderCostBreakdown(ctx context.Context, projectID ulid.ULID, timeRange *TimeRange) (*ProviderCostBreakdown, error)

	// Budget and quota management
	CheckBudgetLimits(ctx context.Context, projectID ulid.ULID, estimatedCost float64) (*BudgetCheckResult, error)
	UpdateBudgetUsage(ctx context.Context, projectID ulid.ULID, actualCost float64) error
	GetBudgetStatus(ctx context.Context, projectID ulid.ULID) (*BudgetStatus, error)
}

// CacheService defines the interface for semantic caching (future implementation)
type CacheService interface {
	// Cache operations
	GetCachedResponse(ctx context.Context, projectID ulid.ULID, requestHash string) (*CachedResponse, error)
	CacheResponse(ctx context.Context, cache *RequestCache) error
	InvalidateCache(ctx context.Context, projectID ulid.ULID, patterns []string) (int64, error)

	// Cache management
	GetCacheStats(ctx context.Context, projectID ulid.ULID, timeRange *TimeRange) (*CacheStats, error)
	OptimizeCache(ctx context.Context, projectID ulid.ULID) (*CacheOptimizationResult, error)
	CleanupExpiredCache(ctx context.Context) (int64, error)

	// Cache strategy
	DetermineCacheability(ctx context.Context, request interface{}) (*CacheabilityResult, error)
	GenerateCacheKey(ctx context.Context, request interface{}) (string, error)
	CalculateCacheSavings(ctx context.Context, projectID ulid.ULID, timeRange *TimeRange) (*CacheSavingsReport, error)
}

// HealthService defines the interface for provider health monitoring
type HealthService interface {
	// Health monitoring
	CheckProviderHealth(ctx context.Context, providerID ulid.ULID) (*ProviderHealth, error)
	CheckAllProvidersHealth(ctx context.Context) ([]*ProviderHealth, error)
	UpdateProviderHealthMetrics(ctx context.Context, providerID ulid.ULID, metrics *HealthMetricsData) error

	// Health status management
	SetProviderHealthStatus(ctx context.Context, providerID ulid.ULID, status HealthStatus, reason string) error
	GetProviderHealthHistory(ctx context.Context, providerID ulid.ULID, timeRange *TimeRange) ([]*ProviderHealthMetrics, error)

	// Health-based routing
	GetHealthyProviders(ctx context.Context, projectID ulid.ULID) ([]*Provider, error)
	FilterUnhealthyProviders(ctx context.Context, providers []*Provider) ([]*Provider, error)

	// Health alerts and notifications
	CheckHealthThresholds(ctx context.Context, providerID ulid.ULID) ([]*HealthAlert, error)
	GetHealthAlerts(ctx context.Context, projectID ulid.ULID) ([]*HealthAlert, error)
	AcknowledgeHealthAlert(ctx context.Context, alertID ulid.ULID) error
}

// Request/Response types for service operations

// ChatCompletionRequest represents a chat completion request
type ChatCompletionRequest struct {
	Model            string                 `json:"model" binding:"required"`
	Messages         []ChatMessage          `json:"messages" binding:"required,min=1"`
	MaxTokens        *int                   `json:"max_tokens,omitempty"`
	Temperature      *float64               `json:"temperature,omitempty" binding:"omitempty,min=0,max=2"`
	TopP             *float64               `json:"top_p,omitempty" binding:"omitempty,min=0,max=1"`
	N                *int                   `json:"n,omitempty" binding:"omitempty,min=1,max=10"`
	Stream           bool                   `json:"stream"`
	Stop             []string               `json:"stop,omitempty"`
	PresencePenalty  *float64               `json:"presence_penalty,omitempty" binding:"omitempty,min=-2,max=2"`
	FrequencyPenalty *float64               `json:"frequency_penalty,omitempty" binding:"omitempty,min=-2,max=2"`
	LogitBias        map[string]interface{} `json:"logit_bias,omitempty"`
	User             *string                `json:"user,omitempty"`
	Functions        []Function             `json:"functions,omitempty"`
	FunctionCall     interface{}            `json:"function_call,omitempty"`
	RoutingStrategy  *string                `json:"routing_strategy,omitempty"`
}

// ChatMessage represents a message in a chat conversation
type ChatMessage struct {
	Role         string      `json:"role" binding:"required"`
	Content      string      `json:"content"`
	Name         *string     `json:"name,omitempty"`
	FunctionCall interface{} `json:"function_call,omitempty"`
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
	// Brokle-specific extensions
	Provider       string    `json:"x-brokle-provider,omitempty"`
	RoutingReason  string    `json:"x-brokle-routing-reason,omitempty"`
	Cost           *float64  `json:"x-brokle-cost,omitempty"`
	CacheHit       bool      `json:"x-brokle-cache-hit,omitempty"`
	ProcessingTime int       `json:"x-brokle-processing-time-ms,omitempty"`
	RequestID      string    `json:"x-brokle-request-id,omitempty"`
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
	Model            string    `json:"model" binding:"required"`
	Prompt           string    `json:"prompt" binding:"required"`
	MaxTokens        *int      `json:"max_tokens,omitempty"`
	Temperature      *float64  `json:"temperature,omitempty"`
	TopP             *float64  `json:"top_p,omitempty"`
	N                *int      `json:"n,omitempty"`
	Stream           bool      `json:"stream"`
	Logprobs         *int      `json:"logprobs,omitempty"`
	Echo             bool      `json:"echo"`
	Stop             []string  `json:"stop,omitempty"`
	PresencePenalty  *float64  `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float64  `json:"frequency_penalty,omitempty"`
	BestOf           *int      `json:"best_of,omitempty"`
	User             *string   `json:"user,omitempty"`
	RoutingStrategy  *string   `json:"routing_strategy,omitempty"`
}

// CompletionResponse represents the response from a text completion
type CompletionResponse struct {
	ID                string             `json:"id"`
	Object            string             `json:"object"`
	Created           int64              `json:"created"`
	Model             string             `json:"model"`
	Choices           []CompletionChoice `json:"choices"`
	Usage             *TokenUsage        `json:"usage,omitempty"`
	// Brokle-specific extensions
	Provider       string   `json:"x-brokle-provider,omitempty"`
	RoutingReason  string   `json:"x-brokle-routing-reason,omitempty"`
	Cost           *float64 `json:"x-brokle-cost,omitempty"`
	CacheHit       bool     `json:"x-brokle-cache-hit,omitempty"`
	ProcessingTime int      `json:"x-brokle-processing-time-ms,omitempty"`
	RequestID      string   `json:"x-brokle-request-id,omitempty"`
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
	Model           string      `json:"model" binding:"required"`
	Input           interface{} `json:"input" binding:"required"`
	EncodingFormat  *string     `json:"encoding_format,omitempty"`
	Dimensions      *int        `json:"dimensions,omitempty"`
	User            *string     `json:"user,omitempty"`
	RoutingStrategy *string     `json:"routing_strategy,omitempty"`
}

// EmbeddingResponse represents the response from an embedding request
type EmbeddingResponse struct {
	Object string      `json:"object"`
	Data   []Embedding `json:"data"`
	Model  string      `json:"model"`
	Usage  *TokenUsage `json:"usage,omitempty"`
	// Brokle-specific extensions
	Provider       string   `json:"x-brokle-provider,omitempty"`
	RoutingReason  string   `json:"x-brokle-routing-reason,omitempty"`
	Cost           *float64 `json:"x-brokle-cost,omitempty"`
	ProcessingTime int      `json:"x-brokle-processing-time-ms,omitempty"`
	RequestID      string   `json:"x-brokle-request-id,omitempty"`
}

// Embedding represents a single embedding vector
type Embedding struct {
	Object    string    `json:"object"`
	Index     int       `json:"index"`
	Embedding []float64 `json:"embedding"`
}

// Function represents a function definition for function calling
type Function struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// Supporting types for service operations

// ModelInfo represents model information for listing
type ModelInfo struct {
	ID            string                 `json:"id"`
	Object        string                 `json:"object"`
	Provider      string                 `json:"provider"`
	DisplayName   string                 `json:"display_name"`
	MaxTokens     int                    `json:"max_tokens"`
	InputCost     float64                `json:"input_cost_per_1k_tokens"`
	OutputCost    float64                `json:"output_cost_per_1k_tokens"`
	Features      []string               `json:"features"`
	QualityScore  *float64               `json:"quality_score,omitempty"`
	SpeedScore    *float64               `json:"speed_score,omitempty"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// ProviderHealth represents provider health status
type ProviderHealth struct {
	ProviderID       ulid.ULID `json:"provider_id"`
	ProviderName     string    `json:"provider_name"`
	Status           HealthStatus `json:"status"`
	LastChecked      time.Time `json:"last_checked"`
	AvgLatencyMs     *int      `json:"avg_latency_ms,omitempty"`
	SuccessRate      *float64  `json:"success_rate,omitempty"`
	UptimePercentage *float64  `json:"uptime_percentage,omitempty"`
	LastError        *string   `json:"last_error,omitempty"`
}

// ConnectionTestResult represents the result of testing a provider connection
type ConnectionTestResult struct {
	Success      bool      `json:"success"`
	LatencyMs    int       `json:"latency_ms"`
	Error        *string   `json:"error,omitempty"`
	TestedAt     time.Time `json:"tested_at"`
	ResponseData interface{} `json:"response_data,omitempty"`
}

// RoutingRequest represents a routing request for decision making
type RoutingRequest struct {
	ModelName       string                 `json:"model_name"`
	Strategy        *RoutingStrategy       `json:"strategy,omitempty"`
	Requirements    *ModelRequirements     `json:"requirements,omitempty"`
	Context         map[string]interface{} `json:"context,omitempty"`
	EstimatedTokens *int                   `json:"estimated_tokens,omitempty"`
	UserTier        *UserTier              `json:"user_tier,omitempty"`
	Priority        *Priority              `json:"priority,omitempty"`
}

// TimeRange represents a time range for analytics queries
type TimeRange struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Interval  *string   `json:"interval,omitempty"`
}

// More supporting types would be defined here as needed...
// This provides a comprehensive foundation for the gateway service interfaces