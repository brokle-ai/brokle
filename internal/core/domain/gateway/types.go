package gateway

import (
	"time"

	"brokle/pkg/ulid"
)

// ProviderType defines the type of AI provider
type ProviderType string

const (
	ProviderTypeOpenAI      ProviderType = "openai"
	ProviderTypeAnthropic   ProviderType = "anthropic"
	ProviderTypeCohere      ProviderType = "cohere"
	ProviderTypeGoogle      ProviderType = "google"
	ProviderTypeAzure       ProviderType = "azure"
	ProviderTypeAWS         ProviderType = "aws"
	ProviderTypeHuggingFace ProviderType = "huggingface"
	ProviderTypeReplicate   ProviderType = "replicate"
)

// ModelType defines the type of AI model
type ModelType string

const (
	ModelTypeText       ModelType = "text"
	ModelTypeEmbedding  ModelType = "embedding"
	ModelTypeImage      ModelType = "image"
	ModelTypeAudio      ModelType = "audio"
	ModelTypeVideo      ModelType = "video"
	ModelTypeMultimodal ModelType = "multimodal"
)

// HealthStatus defines the health status of a provider
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// RoutingStrategy defines the strategy for routing requests to providers
type RoutingStrategy string

const (
	RoutingStrategyCostOptimized    RoutingStrategy = "cost_optimized"
	RoutingStrategyLatencyOptimized RoutingStrategy = "latency_optimized"
	RoutingStrategyQualityOptimized RoutingStrategy = "quality_optimized"
	RoutingStrategyRoundRobin       RoutingStrategy = "round_robin"
	RoutingStrategyWeightedRandom   RoutingStrategy = "weighted_random"
	RoutingStrategyFailover         RoutingStrategy = "failover"
	RoutingStrategyLoadBalance      RoutingStrategy = "load_balance"
)

// RequestType defines the type of AI request
type RequestType string

const (
	RequestTypeChatCompletion     RequestType = "chat_completion"
	RequestTypeCompletion         RequestType = "completion"
	RequestTypeEmbedding          RequestType = "embedding"
	RequestTypeImageGeneration    RequestType = "image_generation"
	RequestTypeAudioGeneration    RequestType = "audio_generation"
	RequestTypeAudioTranscription RequestType = "audio_transcription"
	RequestTypeModeration         RequestType = "moderation"
)

// CacheStrategy defines the caching strategy for requests
type CacheStrategy string

const (
	CacheStrategyNone     CacheStrategy = "none"
	CacheStrategySemantic CacheStrategy = "semantic"
	CacheStrategyExact    CacheStrategy = "exact"
	CacheStrategyTime     CacheStrategy = "time"
)

// ProviderStatus defines the current status of a provider
type ProviderStatus string

const (
	ProviderStatusActive      ProviderStatus = "active"
	ProviderStatusInactive    ProviderStatus = "inactive"
	ProviderStatusMaintenance ProviderStatus = "maintenance"
	ProviderStatusDeprecated  ProviderStatus = "deprecated"
)

// ErrorType defines the type of error that occurred
type ErrorType string

const (
	ErrorTypeRateLimit      ErrorType = "rate_limit"
	ErrorTypeTimeout        ErrorType = "timeout"
	ErrorTypeAuthentication ErrorType = "authentication"
	ErrorTypePermission     ErrorType = "permission"
	ErrorTypeQuotaExceeded  ErrorType = "quota_exceeded"
	ErrorTypeInvalidRequest ErrorType = "invalid_request"
	ErrorTypeServerError    ErrorType = "server_error"
	ErrorTypeNetworkError   ErrorType = "network_error"
	ErrorTypeUnknown        ErrorType = "unknown"
)

// RetryStrategy defines the retry strategy for failed requests
type RetryStrategy string

const (
	RetryStrategyNone        RetryStrategy = "none"
	RetryStrategyLinear      RetryStrategy = "linear"
	RetryStrategyExponential RetryStrategy = "exponential"
	RetryStrategyFixed       RetryStrategy = "fixed"
)

// LoadBalancingMethod defines the method for load balancing
type LoadBalancingMethod string

const (
	LoadBalancingMethodRoundRobin       LoadBalancingMethod = "round_robin"
	LoadBalancingMethodWeighted         LoadBalancingMethod = "weighted"
	LoadBalancingMethodLeastConnections LoadBalancingMethod = "least_connections"
	LoadBalancingMethodRandom           LoadBalancingMethod = "random"
	LoadBalancingMethodIPHash           LoadBalancingMethod = "ip_hash"
)

// Priority defines the priority level for requests
type Priority int

const (
	PriorityLowest  Priority = 1
	PriorityLow     Priority = 2
	PriorityMedium  Priority = 5
	PriorityHigh    Priority = 8
	PriorityHighest Priority = 10
)

// UserTier defines the user subscription tier
type UserTier string

const (
	UserTierFree       UserTier = "free"
	UserTierStarter    UserTier = "starter"
	UserTierPro        UserTier = "pro"
	UserTierBusiness   UserTier = "business"
	UserTierEnterprise UserTier = "enterprise"
)

// RateLimitScope defines the scope for rate limiting
type RateLimitScope string

const (
	RateLimitScopeUser         RateLimitScope = "user"
	RateLimitScopeProject      RateLimitScope = "project"
	RateLimitScopeOrganization RateLimitScope = "organization"
	RateLimitScopeGlobal       RateLimitScope = "global"
	RateLimitScopeProvider     RateLimitScope = "provider"
)

// MetricType defines the type of metric being tracked
type MetricType string

const (
	MetricTypeLatency      MetricType = "latency"
	MetricTypeThroughput   MetricType = "throughput"
	MetricTypeErrorRate    MetricType = "error_rate"
	MetricTypeCost         MetricType = "cost"
	MetricTypeTokens       MetricType = "tokens"
	MetricTypeQuality      MetricType = "quality"
	MetricTypeCacheHitRate MetricType = "cache_hit_rate"
)

// AggregationMethod defines how metrics are aggregated
type AggregationMethod string

const (
	AggregationMethodSum   AggregationMethod = "sum"
	AggregationMethodAvg   AggregationMethod = "avg"
	AggregationMethodMin   AggregationMethod = "min"
	AggregationMethodMax   AggregationMethod = "max"
	AggregationMethodCount AggregationMethod = "count"
	AggregationMethodP50   AggregationMethod = "p50"
	AggregationMethodP95   AggregationMethod = "p95"
	AggregationMethodP99   AggregationMethod = "p99"
)

// TimeWindow defines the time window for aggregation
type TimeWindow string

const (
	TimeWindowMinute TimeWindow = "minute"
	TimeWindowHour   TimeWindow = "hour"
	TimeWindowDay    TimeWindow = "day"
	TimeWindowWeek   TimeWindow = "week"
	TimeWindowMonth  TimeWindow = "month"
)

// String methods for enums

func (pt ProviderType) String() string {
	return string(pt)
}

func (mt ModelType) String() string {
	return string(mt)
}

func (hs HealthStatus) String() string {
	return string(hs)
}

func (rs RoutingStrategy) String() string {
	return string(rs)
}

func (rt RequestType) String() string {
	return string(rt)
}

func (cs CacheStrategy) String() string {
	return string(cs)
}

func (ps ProviderStatus) String() string {
	return string(ps)
}

func (et ErrorType) String() string {
	return string(et)
}

func (rs RetryStrategy) String() string {
	return string(rs)
}

// Validation methods for enums

// IsValidProviderType checks if a provider type is valid
func IsValidProviderType(pt ProviderType) bool {
	switch pt {
	case ProviderTypeOpenAI, ProviderTypeAnthropic, ProviderTypeCohere, ProviderTypeGoogle,
		ProviderTypeAzure, ProviderTypeAWS, ProviderTypeHuggingFace, ProviderTypeReplicate:
		return true
	default:
		return false
	}
}

// IsValidModelType checks if a model type is valid
func IsValidModelType(mt ModelType) bool {
	switch mt {
	case ModelTypeText, ModelTypeEmbedding, ModelTypeImage, ModelTypeAudio, ModelTypeVideo, ModelTypeMultimodal:
		return true
	default:
		return false
	}
}

// IsValidHealthStatus checks if a health status is valid
func IsValidHealthStatus(hs HealthStatus) bool {
	switch hs {
	case HealthStatusHealthy, HealthStatusDegraded, HealthStatusUnhealthy, HealthStatusUnknown:
		return true
	default:
		return false
	}
}

// IsValidRoutingStrategy checks if a routing strategy is valid
func IsValidRoutingStrategy(rs RoutingStrategy) bool {
	switch rs {
	case RoutingStrategyCostOptimized, RoutingStrategyLatencyOptimized, RoutingStrategyQualityOptimized,
		RoutingStrategyRoundRobin, RoutingStrategyWeightedRandom, RoutingStrategyFailover, RoutingStrategyLoadBalance:
		return true
	default:
		return false
	}
}

// IsValidRequestType checks if a request type is valid
func IsValidRequestType(rt RequestType) bool {
	switch rt {
	case RequestTypeChatCompletion, RequestTypeCompletion, RequestTypeEmbedding,
		RequestTypeImageGeneration, RequestTypeAudioGeneration, RequestTypeAudioTranscription, RequestTypeModeration:
		return true
	default:
		return false
	}
}

// IsValidCacheStrategy checks if a cache strategy is valid
func IsValidCacheStrategy(cs CacheStrategy) bool {
	switch cs {
	case CacheStrategyNone, CacheStrategySemantic, CacheStrategyExact, CacheStrategyTime:
		return true
	default:
		return false
	}
}

// IsValidUserTier checks if a user tier is valid
func IsValidUserTier(ut UserTier) bool {
	switch ut {
	case UserTierFree, UserTierStarter, UserTierPro, UserTierBusiness, UserTierEnterprise:
		return true
	default:
		return false
	}
}

// Helper functions

// GetDefaultRoutingStrategy returns the default routing strategy
func GetDefaultRoutingStrategy() RoutingStrategy {
	return RoutingStrategyCostOptimized
}

// GetDefaultRetryStrategy returns the default retry strategy
func GetDefaultRetryStrategy() RetryStrategy {
	return RetryStrategyExponential
}

// GetDefaultCacheStrategy returns the default cache strategy
func GetDefaultCacheStrategy() CacheStrategy {
	return CacheStrategyNone
}

// GetDefaultTimeWindow returns the default time window for aggregation
func GetDefaultTimeWindow() TimeWindow {
	return TimeWindowHour
}

// GetSupportedProviderTypes returns all supported provider types
func GetSupportedProviderTypes() []ProviderType {
	return []ProviderType{
		ProviderTypeOpenAI,
		ProviderTypeAnthropic,
		ProviderTypeCohere,
		ProviderTypeGoogle,
		ProviderTypeAzure,
		ProviderTypeAWS,
		ProviderTypeHuggingFace,
		ProviderTypeReplicate,
	}
}

// GetSupportedModelTypes returns all supported model types
func GetSupportedModelTypes() []ModelType {
	return []ModelType{
		ModelTypeText,
		ModelTypeEmbedding,
		ModelTypeImage,
		ModelTypeAudio,
		ModelTypeVideo,
		ModelTypeMultimodal,
	}
}

// GetSupportedRoutingStrategies returns all supported routing strategies
func GetSupportedRoutingStrategies() []RoutingStrategy {
	return []RoutingStrategy{
		RoutingStrategyCostOptimized,
		RoutingStrategyLatencyOptimized,
		RoutingStrategyQualityOptimized,
		RoutingStrategyRoundRobin,
		RoutingStrategyWeightedRandom,
		RoutingStrategyFailover,
		RoutingStrategyLoadBalance,
	}
}

// Additional types for service interfaces

// ProviderHealthUpdate represents an update to provider health status
type ProviderHealthUpdate struct {
	Status       HealthStatus `json:"status"`
	Latency      *float64     `json:"latency,omitempty"`
	ErrorRate    *float64     `json:"error_rate,omitempty"`
	Availability *float64     `json:"availability,omitempty"`
	Throughput   *float64     `json:"throughput,omitempty"`
	Message      *string      `json:"message,omitempty"`
	CheckedAt    time.Time    `json:"checked_at"`
}

// ModelPerformanceMetrics represents performance metrics for a model
type ModelPerformanceMetrics struct {
	ModelID            ulid.ULID `json:"model_id"`
	ModelName          string    `json:"model_name"`
	ProviderID         ulid.ULID `json:"provider_id"`
	AverageLatency     float64   `json:"average_latency"`
	P95Latency         float64   `json:"p95_latency"`
	P99Latency         float64   `json:"p99_latency"`
	ErrorRate          float64   `json:"error_rate"`
	Throughput         float64   `json:"throughput"`
	TokensPerSecond    float64   `json:"tokens_per_second"`
	CostPerToken       float64   `json:"cost_per_token"`
	QualityScore       *float64  `json:"quality_score,omitempty"`
	TotalRequests      int64     `json:"total_requests"`
	SuccessfulRequests int64     `json:"successful_requests"`
	FailedRequests     int64     `json:"failed_requests"`
	PeriodStart        time.Time `json:"period_start"`
	PeriodEnd          time.Time `json:"period_end"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// ComparisonCriteria defines criteria for comparing models
type ComparisonCriteria struct {
	WeightLatency    float64    `json:"weight_latency"`
	WeightCost       float64    `json:"weight_cost"`
	WeightQuality    float64    `json:"weight_quality"`
	WeightThroughput float64    `json:"weight_throughput"`
	MaxLatency       *float64   `json:"max_latency,omitempty"`
	MaxCost          *float64   `json:"max_cost,omitempty"`
	MinQuality       *float64   `json:"min_quality,omitempty"`
	MinThroughput    *float64   `json:"min_throughput,omitempty"`
	TimeRange        *TimeRange `json:"time_range,omitempty"`
}

// ModelComparison represents the result of comparing multiple models
type ModelComparison struct {
	Criteria       *ComparisonCriteria        `json:"criteria"`
	Models         []*ModelPerformanceMetrics `json:"models"`
	Ranking        []ModelRanking             `json:"ranking"`
	Recommendation *ModelRecommendation       `json:"recommendation,omitempty"`
	GeneratedAt    time.Time                  `json:"generated_at"`
}

// ModelRanking represents a ranked model in a comparison
type ModelRanking struct {
	ModelID    ulid.ULID `json:"model_id"`
	ModelName  string    `json:"model_name"`
	Score      float64   `json:"score"`
	Rank       int       `json:"rank"`
	Strengths  []string  `json:"strengths,omitempty"`
	Weaknesses []string  `json:"weaknesses,omitempty"`
}

// ModelRecommendation provides a recommendation for model selection
type ModelRecommendation struct {
	ModelID      ulid.ULID   `json:"model_id"`
	ModelName    string      `json:"model_name"`
	Reason       string      `json:"reason"`
	Confidence   float64     `json:"confidence"`
	Alternatives []ulid.ULID `json:"alternatives,omitempty"`
}

// RouteTestResult represents the result of testing a route
type RouteTestResult struct {
	TestID        ulid.ULID        `json:"test_id"`
	ProjectID     ulid.ULID        `json:"project_id"`
	Request       *RoutingRequest  `json:"request"`
	Decision      *RoutingDecision `json:"decision"`
	ActualLatency *float64         `json:"actual_latency,omitempty"`
	ActualCost    *float64         `json:"actual_cost,omitempty"`
	Success       bool             `json:"success"`
	Error         *string          `json:"error,omitempty"`
	FallbackUsed  bool             `json:"fallback_used"`
	RetryCount    int              `json:"retry_count"`
	ExecutionTime time.Duration    `json:"execution_time"`
	TestedAt      time.Time        `json:"tested_at"`
}

// RoutingAnalysis provides analysis of routing performance
type RoutingAnalysis struct {
	ProjectID         ulid.ULID                 `json:"project_id"`
	TimeRange         *TimeRange                `json:"time_range"`
	TotalRequests     int64                     `json:"total_requests"`
	SuccessRate       float64                   `json:"success_rate"`
	AverageLatency    float64                   `json:"average_latency"`
	AverageCost       float64                   `json:"average_cost"`
	ProviderBreakdown map[string]*ProviderStats `json:"provider_breakdown"`
	ModelBreakdown    map[string]*ModelStats    `json:"model_breakdown"`
	StrategyBreakdown map[string]*StrategyStats `json:"strategy_breakdown"`
	FallbackRate      float64                   `json:"fallback_rate"`
	RetryRate         float64                   `json:"retry_rate"`
	CacheHitRate      *float64                  `json:"cache_hit_rate,omitempty"`
	Recommendations   []*RoutingRecommendation  `json:"recommendations,omitempty"`
	GeneratedAt       time.Time                 `json:"generated_at"`
}

// ProviderStats represents statistics for a provider
type ProviderStats struct {
	ProviderID     ulid.ULID `json:"provider_id"`
	ProviderName   string    `json:"provider_name"`
	Requests       int64     `json:"requests"`
	SuccessRate    float64   `json:"success_rate"`
	AverageLatency float64   `json:"average_latency"`
	AverageCost    float64   `json:"average_cost"`
	TotalCost      float64   `json:"total_cost"`
	Uptime         float64   `json:"uptime"`
}

// ModelStats represents statistics for a model
type ModelStats struct {
	ModelID         ulid.ULID `json:"model_id"`
	ModelName       string    `json:"model_name"`
	Requests        int64     `json:"requests"`
	SuccessRate     float64   `json:"success_rate"`
	AverageLatency  float64   `json:"average_latency"`
	AverageCost     float64   `json:"average_cost"`
	TotalCost       float64   `json:"total_cost"`
	TokensProcessed int64     `json:"tokens_processed"`
}

// StrategyStats represents statistics for a routing strategy
type StrategyStats struct {
	Strategy       RoutingStrategy `json:"strategy"`
	Requests       int64           `json:"requests"`
	SuccessRate    float64         `json:"success_rate"`
	AverageLatency float64         `json:"average_latency"`
	AverageCost    float64         `json:"average_cost"`
	FallbackRate   float64         `json:"fallback_rate"`
}

// RoutingRecommendation provides recommendations for improving routing
type RoutingRecommendation struct {
	Type        string   `json:"type"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Impact      string   `json:"impact"`
	Priority    Priority `json:"priority"`
	Actions     []string `json:"actions,omitempty"`
}

// CostCalculationRequest represents a request for cost calculation
type CostCalculationRequest struct {
	ModelID      ulid.ULID   `json:"model_id"`
	ModelName    string      `json:"model_name"`
	InputTokens  int         `json:"input_tokens"`
	OutputTokens int         `json:"output_tokens"`
	RequestType  RequestType `json:"request_type"`
	ProviderID   *ulid.ULID  `json:"provider_id,omitempty"`
}

// BatchCostResult represents the result of batch cost calculations
type BatchCostResult struct {
	Requests     []*CostCalculationRequest `json:"requests"`
	Results      []*CostCalculationResult  `json:"results"`
	TotalCost    float64                   `json:"total_cost"`
	Currency     string                    `json:"currency"`
	CalculatedAt time.Time                 `json:"calculated_at"`
}

// CostCalculationResult represents the result of a single cost calculation
type CostCalculationResult struct {
	RequestIndex int       `json:"request_index"`
	InputCost    float64   `json:"input_cost"`
	OutputCost   float64   `json:"output_cost"`
	TotalCost    float64   `json:"total_cost"`
	Currency     string    `json:"currency"`
	ProviderID   ulid.ULID `json:"provider_id"`
	Error        *string   `json:"error,omitempty"`
}

// CostComparison represents a comparison of costs across providers/models
type CostComparison struct {
	ModelNames    []string                  `json:"model_names"`
	TokenCount    int                       `json:"token_count"`
	Providers     []*ProviderCostComparison `json:"providers"`
	Cheapest      *ProviderCostComparison   `json:"cheapest"`
	MostExpensive *ProviderCostComparison   `json:"most_expensive"`
	AverageCost   float64                   `json:"average_cost"`
	CostRange     float64                   `json:"cost_range"`
	Currency      string                    `json:"currency"`
	ComparedAt    time.Time                 `json:"compared_at"`
}

// ProviderCostComparison represents cost data for a provider in a comparison
type ProviderCostComparison struct {
	ProviderID   ulid.ULID `json:"provider_id"`
	ProviderName string    `json:"provider_name"`
	ModelName    string    `json:"model_name"`
	InputCost    float64   `json:"input_cost"`
	OutputCost   float64   `json:"output_cost"`
	TotalCost    float64   `json:"total_cost"`
	CostPerToken float64   `json:"cost_per_token"`
	RelativeCost float64   `json:"relative_cost"` // Compared to cheapest
	Available    bool      `json:"available"`
	Healthy      bool      `json:"healthy"`
}

// CostSavingsReport represents potential cost savings analysis
type CostSavingsReport struct {
	ProjectID             ulid.ULID                         `json:"project_id"`
	TimeRange             *TimeRange                        `json:"time_range"`
	CurrentCost           float64                           `json:"current_cost"`
	OptimizedCost         float64                           `json:"optimized_cost"`
	PotentialSavings      float64                           `json:"potential_savings"`
	SavingsPercentage     float64                           `json:"savings_percentage"`
	Currency              string                            `json:"currency"`
	Recommendations       []*CostOptimizationRecommendation `json:"recommendations"`
	ProviderOptimizations []*ProviderOptimization           `json:"provider_optimizations"`
	ModelOptimizations    []*ModelOptimization              `json:"model_optimizations"`
	GeneratedAt           time.Time                         `json:"generated_at"`
}

// CostOptimizationRecommendation provides specific cost optimization advice
type CostOptimizationRecommendation struct {
	Type                 string   `json:"type"`
	Title                string   `json:"title"`
	Description          string   `json:"description"`
	PotentialSavings     float64  `json:"potential_savings"`
	ImplementationEffort string   `json:"implementation_effort"`
	Priority             Priority `json:"priority"`
	Actions              []string `json:"actions,omitempty"`
}

// ProviderOptimization represents optimization opportunities for a provider
type ProviderOptimization struct {
	ProviderID       ulid.ULID `json:"provider_id"`
	ProviderName     string    `json:"provider_name"`
	CurrentCost      float64   `json:"current_cost"`
	OptimizedCost    float64   `json:"optimized_cost"`
	PotentialSavings float64   `json:"potential_savings"`
	Recommendation   string    `json:"recommendation"`
}

// ModelOptimization represents optimization opportunities for a model
type ModelOptimization struct {
	ModelID          ulid.ULID `json:"model_id"`
	ModelName        string    `json:"model_name"`
	CurrentCost      float64   `json:"current_cost"`
	OptimizedCost    float64   `json:"optimized_cost"`
	PotentialSavings float64   `json:"potential_savings"`
	AlternativeModel *string   `json:"alternative_model,omitempty"`
	Recommendation   string    `json:"recommendation"`
}

// Note: TimeRange is defined in service.go to avoid circular dependencies

// CostCalculation represents cost calculation details
type CostCalculation struct {
	ModelID         ulid.ULID      `json:"model_id"`
	ProviderID      ulid.ULID      `json:"provider_id"`
	RequestID       *ulid.ULID     `json:"request_id,omitempty"`
	InputTokens     int32          `json:"input_tokens"`
	OutputTokens    int32          `json:"output_tokens"`
	TotalTokens     int32          `json:"total_tokens"`
	InputCost       float64        `json:"input_cost"`
	OutputCost      float64        `json:"output_cost"`
	TotalCost       float64        `json:"total_cost"`
	Currency        string         `json:"currency"`
	EstimatedAt     time.Time      `json:"estimated_at"`
	CalculationType string         `json:"calculation_type"`
	Duration        *time.Duration `json:"duration,omitempty"`
	EstimatedCost   bool           `json:"estimated_cost"`
	RatePerToken    float64        `json:"rate_per_token,omitempty"`
	ProviderPricing string         `json:"provider_pricing,omitempty"`
}

// CostEstimationRequest represents a request for cost estimation
type CostEstimationRequest struct {
	Model          *Model    `json:"model"`
	OrganizationID ulid.ULID `json:"organization_id"`
	InputTokens    int32     `json:"input_tokens"`
	MaxTokens      int32     `json:"max_tokens,omitempty"`
}

// ActualCostRequest represents a request for actual cost calculation
type ActualCostRequest struct {
	Model          *Model        `json:"model"`
	OrganizationID ulid.ULID     `json:"organization_id"`
	RequestID      ulid.ULID     `json:"request_id"`
	InputTokens    int32         `json:"input_tokens"`
	OutputTokens   int32         `json:"output_tokens"`
	Duration       time.Duration `json:"duration"`
}

// UsageStatsRequest represents a request for usage statistics
type UsageStatsRequest struct {
	OrganizationID ulid.ULID `json:"organization_id"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
}

// UsageStatsResponse represents usage statistics response
type UsageStatsResponse struct {
	OrganizationID ulid.ULID                      `json:"organization_id"`
	StartDate      time.Time                      `json:"start_date"`
	EndDate        time.Time                      `json:"end_date"`
	TotalRequests  int64                          `json:"total_requests"`
	TotalTokens    int64                          `json:"total_tokens"`
	TotalCost      float64                        `json:"total_cost"`
	Currency       string                         `json:"currency"`
	ModelStats     map[string]*ModelUsageStats    `json:"model_stats"`
	ProviderStats  map[string]*ProviderUsageStats `json:"provider_stats"`
	DailyStats     []*DailyUsageStats             `json:"daily_stats"`
}

// ModelUsageStats represents usage statistics for a model
type ModelUsageStats struct {
	ModelID        ulid.ULID `json:"model_id"`
	ModelName      string    `json:"model_name"`
	Requests       int64     `json:"requests"`
	Tokens         int64     `json:"tokens"`
	Cost           float64   `json:"cost"`
	AverageLatency float64   `json:"average_latency"`
}

// ProviderUsageStats represents usage statistics for a provider
type ProviderUsageStats struct {
	ProviderID     ulid.ULID `json:"provider_id"`
	ProviderName   string    `json:"provider_name"`
	Requests       int64     `json:"requests"`
	Tokens         int64     `json:"tokens"`
	Cost           float64   `json:"cost"`
	AverageLatency float64   `json:"average_latency"`
}

// DailyUsageStats represents daily usage statistics
type DailyUsageStats struct {
	Date     time.Time `json:"date"`
	Requests int64     `json:"requests"`
	Tokens   int64     `json:"tokens"`
	Cost     float64   `json:"cost"`
}

// CostBreakdownRequest represents a request for cost breakdown
type CostBreakdownRequest struct {
	OrganizationID ulid.ULID `json:"organization_id"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	GroupBy        []string  `json:"group_by"`
}

// CostBreakdownResponse represents cost breakdown response
type CostBreakdownResponse struct {
	OrganizationID ulid.ULID            `json:"organization_id"`
	StartDate      time.Time            `json:"start_date"`
	EndDate        time.Time            `json:"end_date"`
	TotalCost      float64              `json:"total_cost"`
	Currency       string               `json:"currency"`
	GroupBy        []string             `json:"group_by"`
	Breakdown      []*CostBreakdownItem `json:"breakdown"`
}

// CostBreakdownItem represents a single cost breakdown item
type CostBreakdownItem struct {
	Dimension string  `json:"dimension"`
	Value     string  `json:"value"`
	Cost      float64 `json:"cost"`
	Percent   float64 `json:"percent"`
}

// UsageTrackingRequest represents a request to track usage
type UsageTrackingRequest struct {
	RequestID      ulid.ULID        `json:"request_id"`
	OrganizationID ulid.ULID        `json:"organization_id"`
	ModelID        ulid.ULID        `json:"model_id"`
	ProviderID     ulid.ULID        `json:"provider_id"`
	Cost           *CostCalculation `json:"cost"`
}

// ModelPricing represents pricing information for a model
type ModelPricing struct {
	ModelID            ulid.ULID `json:"model_id"`
	ModelName          string    `json:"model_name"`
	ProviderID         ulid.ULID `json:"provider_id"`
	InputCostPerToken  float64   `json:"input_cost_per_token"`
	OutputCostPerToken float64   `json:"output_cost_per_token"`
	Currency           string    `json:"currency"`
	EffectiveDate      time.Time `json:"effective_date"`
	IsActive           bool      `json:"is_active"`
}

// EfficiencyMetrics represents efficiency metrics for usage analysis
type EfficiencyMetrics struct {
	TokensPerSecond    float64       `json:"tokens_per_second"`
	OutputToInputRatio float64       `json:"output_to_input_ratio"`
	TotalTokens        int32         `json:"total_tokens"`
	Duration           time.Duration `json:"duration"`
}

// ListModelsRequest represents a request to list available models
type ListModelsRequest struct {
	OrganizationID ulid.ULID `json:"organization_id"`
	ProjectID      ulid.ULID `json:"project_id"`
}

// HealthCheckRequest represents a health check request
type HealthCheckRequest struct {
	ProviderID *ulid.ULID `json:"provider_id,omitempty"`
}

// HealthCheckResponse represents a health check response
type HealthCheckResponse struct {
	Status    string                 `json:"status"`
	Providers []ProviderHealthStatus `json:"providers"`
	CheckedAt time.Time              `json:"checked_at"`
}

// ProviderHealthStatus represents health status for a provider
type ProviderHealthStatus struct {
	ProviderID   ulid.ULID `json:"provider_id"`
	ProviderName string    `json:"provider_name"`
	Status       string    `json:"status"`
	Latency      *float64  `json:"latency,omitempty"`
	Error        *string   `json:"error,omitempty"`
}

// ProviderSelectionRequest represents a request for provider selection
type ProviderSelectionRequest struct {
	OrganizationID ulid.ULID          `json:"organization_id"`
	ProjectID      ulid.ULID          `json:"project_id"`
	ModelName      string             `json:"model_name"`
	RequestType    RequestType        `json:"request_type"`
	Strategy       RoutingStrategy    `json:"strategy"`
	Requirements   *ModelRequirements `json:"requirements,omitempty"`
}

// ProviderSelectionResponse represents the response from provider selection
type ProviderSelectionResponse struct {
	ProviderID       ulid.ULID `json:"provider_id"`
	ProviderName     string    `json:"provider_name"`
	ModelID          ulid.ULID `json:"model_id"`
	ModelName        string    `json:"model_name"`
	EstimatedCost    *float64  `json:"estimated_cost,omitempty"`
	EstimatedLatency *float64  `json:"estimated_latency,omitempty"`
	QualityScore     *float64  `json:"quality_score,omitempty"`
	Reason           string    `json:"reason"`
}

// FallbackSelectionRequest represents a request for fallback provider selection
type FallbackSelectionRequest struct {
	OrganizationID   ulid.ULID          `json:"organization_id"`
	ProjectID        ulid.ULID          `json:"project_id"`
	OriginalProvider ulid.ULID          `json:"original_provider"`
	ModelName        string             `json:"model_name"`
	RequestType      RequestType        `json:"request_type"`
	ExcludeProviders []ulid.ULID        `json:"exclude_providers,omitempty"`
	Requirements     *ModelRequirements `json:"requirements,omitempty"`
}

// RequestRoutingConfig represents configuration for request routing
type RequestRoutingConfig struct {
	Strategy        RoutingStrategy `json:"strategy"`
	FallbackEnabled bool            `json:"fallback_enabled"`
	RetryAttempts   int             `json:"retry_attempts"`
	TimeoutSeconds  int             `json:"timeout_seconds"`
	LoadBalancing   bool            `json:"load_balancing"`
}

// RequestRoutingResponse represents the response from request routing
type RequestRoutingResponse struct {
	ProviderID       ulid.ULID `json:"provider_id"`
	ProviderName     string    `json:"provider_name"`
	ModelID          ulid.ULID `json:"model_id"`
	ModelName        string    `json:"model_name"`
	RoutingDecision  string    `json:"routing_decision"`
	EstimatedCost    *float64  `json:"estimated_cost,omitempty"`
	EstimatedLatency *float64  `json:"estimated_latency,omitempty"`
	FallbackUsed     bool      `json:"fallback_used"`
	RetryCount       int       `json:"retry_count"`
}

// WeightUpdateRequest represents a request to update provider weights
type WeightUpdateRequest struct {
	ProviderID ulid.ULID `json:"provider_id"`
	Weight     float64   `json:"weight"`
	Reason     string    `json:"reason"`
}

// ProviderRoute represents routing information for a provider
type ProviderRoute struct {
	ProviderID       ulid.ULID `json:"provider_id"`
	ProviderName     string    `json:"provider_name"`
	ModelID          ulid.ULID `json:"model_id"`
	ModelName        string    `json:"model_name"`
	Weight           float64   `json:"weight"`
	Priority         int       `json:"priority"`
	HealthStatus     string    `json:"health_status"`
	EstimatedCost    *float64  `json:"estimated_cost,omitempty"`
	EstimatedLatency *float64  `json:"estimated_latency,omitempty"`
	IsAvailable      bool      `json:"is_available"`
}

// EmbeddingsRequest represents an embeddings request
type EmbeddingsRequest struct {
	OrganizationID ulid.ULID   `json:"organization_id"`
	Model          string      `json:"model"`
	Input          interface{} `json:"input"`
	User           string      `json:"user,omitempty"`
	EncodingFormat string      `json:"encoding_format,omitempty"`
}

// EmbeddingResponse represents an embeddings response
type EmbeddingResponse struct {
	Object          string          `json:"object"`
	Data            []EmbeddingData `json:"data"`
	Model           string          `json:"model"`
	Usage           TokenUsage      `json:"usage"`
	Provider        string          `json:"provider"`
	Cost            *float64        `json:"cost,omitempty"`
	RoutingDecision string          `json:"routing_decision,omitempty"`
}

// EmbeddingData represents a single embedding
type EmbeddingData struct {
	Object    string    `json:"object"`
	Index     int       `json:"index"`
	Embedding []float64 `json:"embedding"`
}

// ListModelsResponse represents list models response
type ListModelsResponse struct {
	Object string      `json:"object"`
	Data   []ModelInfo `json:"data"`
}

// RoutingResponse represents routing decision response
type RoutingResponse struct {
	Provider         string                 `json:"provider"`
	Model            string                 `json:"model"`
	Endpoint         string                 `json:"endpoint"`
	Strategy         string                 `json:"strategy"`
	EstimatedCost    *float64               `json:"estimated_cost,omitempty"`
	EstimatedLatency *int                   `json:"estimated_latency,omitempty"`
	QualityScore     *float64               `json:"quality_score,omitempty"`
	CacheHit         bool                   `json:"cache_hit"`
	ProviderHealth   *float64               `json:"provider_health,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// CacheStatusRequest represents cache status request
type CacheStatusRequest struct {
	OrganizationID ulid.ULID `json:"organization_id"`
}

// CacheStatusResponse represents cache status response
type CacheStatusResponse struct {
	Status            string                 `json:"status"`
	HitRate           float64                `json:"hit_rate"`
	TotalEntries      int64                  `json:"total_entries"`
	SizeBytes         int64                  `json:"size_bytes"`
	MemoryUsage       float64                `json:"memory_usage"`
	EvictionCount     int64                  `json:"eviction_count"`
	LastEviction      *int64                 `json:"last_eviction,omitempty"`
	ProviderBreakdown map[string]interface{} `json:"provider_breakdown,omitempty"`
}

// InvalidateCacheRequest represents cache invalidation request
type InvalidateCacheRequest struct {
	OrganizationID ulid.ULID `json:"organization_id"`
	Provider       string    `json:"provider,omitempty"`
	Model          string    `json:"model,omitempty"`
	Keys           []string  `json:"keys,omitempty"`
	ClearAll       bool      `json:"clear_all,omitempty"`
	MaxAge         int64     `json:"max_age,omitempty"`
	Pattern        string    `json:"pattern,omitempty"`
}

// CacheInvalidationResponse represents cache invalidation result
type InvalidateCacheResponse struct {
	Success         bool     `json:"success"`
	InvalidatedKeys []string `json:"invalidated_keys,omitempty"`
	Count           int      `json:"count"`
	Message         string   `json:"message"`
	Error           string   `json:"error,omitempty"`
}

// Streaming response types for real-time AI responses

// ChatCompletionStreamResponse represents streaming chat completion response
type ChatCompletionStreamResponse struct {
	ID      string                       `json:"id"`
	Object  string                       `json:"object"`
	Created int64                        `json:"created"`
	Model   string                       `json:"model"`
	Choices []ChatCompletionStreamChoice `json:"choices"`

	// Brokle extensions
	Provider        *string `json:"provider,omitempty"`
	RoutingDecision *string `json:"routing_decision,omitempty"`
}

// ChatCompletionStreamChoice represents streaming choice in chat completion
type ChatCompletionStreamChoice struct {
	Index        int                       `json:"index"`
	Delta        ChatCompletionStreamDelta `json:"delta"`
	FinishReason *string                   `json:"finish_reason,omitempty"`
}

// ChatCompletionStreamDelta represents delta content in streaming response
type ChatCompletionStreamDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// CompletionStreamResponse represents streaming completion response
type CompletionStreamResponse struct {
	ID      string                   `json:"id"`
	Object  string                   `json:"object"`
	Created int64                    `json:"created"`
	Model   string                   `json:"model"`
	Choices []CompletionStreamChoice `json:"choices"`

	// Brokle extensions
	Provider        *string `json:"provider,omitempty"`
	RoutingDecision *string `json:"routing_decision,omitempty"`
}

// CompletionStreamChoice represents streaming choice in completion
type CompletionStreamChoice struct {
	Index        int         `json:"index"`
	Text         string      `json:"text"`
	FinishReason *string     `json:"finish_reason,omitempty"`
	LogProbs     interface{} `json:"logprobs,omitempty"`
}

// CostAnalytics represents cost analytics for a project
type CostAnalytics struct {
	ProjectID         ulid.ULID                `json:"project_id"`
	TimeRange         *TimeRange               `json:"time_range"`
	TotalCost         float64                  `json:"total_cost"`
	Currency          string                   `json:"currency"`
	ProviderBreakdown []*ProviderCostBreakdown `json:"provider_breakdown"`
	ModelBreakdown    []*ModelCostBreakdown    `json:"model_breakdown"`
	DailyBreakdown    []*DailyCostBreakdown    `json:"daily_breakdown"`
	Trends            *CostTrends              `json:"trends"`
	GeneratedAt       time.Time                `json:"generated_at"`
}

// ProviderCostBreakdown represents cost breakdown by provider
type ProviderCostBreakdown struct {
	ProviderID        ulid.ULID `json:"provider_id"`
	ProviderName      string    `json:"provider_name"`
	TotalCost         float64   `json:"total_cost"`
	RequestCount      int64     `json:"request_count"`
	AverageCost       float64   `json:"average_cost"`
	PercentageOfTotal float64   `json:"percentage_of_total"`
	TokensProcessed   int64     `json:"tokens_processed"`
}

// ModelCostBreakdown represents cost breakdown by model
type ModelCostBreakdown struct {
	ModelID           ulid.ULID `json:"model_id"`
	ModelName         string    `json:"model_name"`
	TotalCost         float64   `json:"total_cost"`
	RequestCount      int64     `json:"request_count"`
	AverageCost       float64   `json:"average_cost"`
	PercentageOfTotal float64   `json:"percentage_of_total"`
	TokensProcessed   int64     `json:"tokens_processed"`
}

// DailyCostBreakdown represents daily cost breakdown
type DailyCostBreakdown struct {
	Date            time.Time `json:"date"`
	TotalCost       float64   `json:"total_cost"`
	RequestCount    int64     `json:"request_count"`
	AverageCost     float64   `json:"average_cost"`
	TokensProcessed int64     `json:"tokens_processed"`
}

// CostTrends represents cost trends over time
type CostTrends struct {
	CostTrend        string  `json:"cost_trend"` // "increasing", "decreasing", "stable"
	PercentageChange float64 `json:"percentage_change"`
	VolumeTrend      string  `json:"volume_trend"`
	VolumeChange     float64 `json:"volume_change"`
}

// BudgetCheckResult represents the result of a budget check
type BudgetCheckResult struct {
	ProjectID          ulid.ULID `json:"project_id"`
	BudgetLimit        float64   `json:"budget_limit"`
	CurrentUsage       float64   `json:"current_usage"`
	EstimatedCost      float64   `json:"estimated_cost"`
	TotalProjectedCost float64   `json:"total_projected_cost"`
	RemainingBudget    float64   `json:"remaining_budget"`
	BudgetUtilization  float64   `json:"budget_utilization"`
	WillExceedBudget   bool      `json:"will_exceed_budget"`
	WarningThreshold   float64   `json:"warning_threshold"`
	ExceededWarning    bool      `json:"exceeded_warning"`
	CheckedAt          time.Time `json:"checked_at"`
}

// BudgetStatus represents the current budget status
type BudgetStatus struct {
	ProjectID         ulid.ULID `json:"project_id"`
	BudgetLimit       float64   `json:"budget_limit"`
	CurrentUsage      float64   `json:"current_usage"`
	RemainingBudget   float64   `json:"remaining_budget"`
	BudgetUtilization float64   `json:"budget_utilization"`
	DaysRemaining     int       `json:"days_remaining"`
	ProjectedUsage    float64   `json:"projected_usage"`
	OnTrack           bool      `json:"on_track"`
	PeriodStart       time.Time `json:"period_start"`
	PeriodEnd         time.Time `json:"period_end"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// CachedResponse represents a cached response
type CachedResponse struct {
	CacheKey        string         `json:"cache_key"`
	Response        interface{}    `json:"response"`
	Metadata        *CacheMetadata `json:"metadata"`
	HitCount        int            `json:"hit_count"`
	CachedAt        time.Time      `json:"cached_at"`
	ExpiresAt       time.Time      `json:"expires_at"`
	SimilarityScore *float64       `json:"similarity_score,omitempty"`
}

// CacheMetadata represents metadata for cached responses
type CacheMetadata struct {
	OriginalRequest interface{} `json:"original_request"`
	ProjectID       ulid.ULID   `json:"project_id"`
	ModelName       string      `json:"model_name"`
	Provider        string      `json:"provider"`
	CostSaved       float64     `json:"cost_saved"`
	LatencySaved    int         `json:"latency_saved_ms"`
}

// CacheOptimizationResult represents cache optimization results
type CacheOptimizationResult struct {
	ProjectID        ulid.ULID     `json:"project_id"`
	EntriesRemoved   int           `json:"entries_removed"`
	SpaceFreed       int64         `json:"space_freed_bytes"`
	OptimizationTime time.Duration `json:"optimization_time"`
	CacheHitRate     float64       `json:"cache_hit_rate"`
	Recommendations  []string      `json:"recommendations,omitempty"`
	OptimizedAt      time.Time     `json:"optimized_at"`
}

// CacheabilityResult represents whether a request can be cached
type CacheabilityResult struct {
	Cacheable           bool          `json:"cacheable"`
	Reason              string        `json:"reason"`
	TTL                 *int          `json:"ttl_seconds,omitempty"`
	CacheStrategy       CacheStrategy `json:"cache_strategy"`
	SimilarityThreshold *float64      `json:"similarity_threshold,omitempty"`
}

// CacheSavingsReport represents cache savings analysis
type CacheSavingsReport struct {
	ProjectID        ulid.ULID  `json:"project_id"`
	TimeRange        *TimeRange `json:"time_range"`
	TotalRequests    int64      `json:"total_requests"`
	CacheHits        int64      `json:"cache_hits"`
	CacheMisses      int64      `json:"cache_misses"`
	HitRate          float64    `json:"hit_rate"`
	CostSaved        float64    `json:"cost_saved"`
	LatencySaved     int64      `json:"latency_saved_ms"`
	Currency         string     `json:"currency"`
	ProjectedSavings float64    `json:"projected_monthly_savings"`
	GeneratedAt      time.Time  `json:"generated_at"`
}

// HealthAlert represents a health alert for a provider
type HealthAlert struct {
	AlertID        ulid.ULID  `json:"alert_id"`
	ProjectID      ulid.ULID  `json:"project_id"`
	ProviderID     ulid.ULID  `json:"provider_id"`
	ProviderName   string     `json:"provider_name"`
	AlertType      string     `json:"alert_type"`
	Severity       string     `json:"severity"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	Threshold      *float64   `json:"threshold,omitempty"`
	CurrentValue   *float64   `json:"current_value,omitempty"`
	Triggered      bool       `json:"triggered"`
	Acknowledged   bool       `json:"acknowledged"`
	AcknowledgedBy *ulid.ULID `json:"acknowledged_by,omitempty"`
	AcknowledgedAt *time.Time `json:"acknowledged_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
