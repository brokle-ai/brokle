package gateway

import (
	"context"
	"time"

	"brokle/pkg/ulid"
)

// ProviderRepository defines the interface for provider data access
type ProviderRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, provider *Provider) error
	GetByID(ctx context.Context, id ulid.ULID) (*Provider, error)
	GetByName(ctx context.Context, name string) (*Provider, error)
	GetByType(ctx context.Context, providerType ProviderType) ([]*Provider, error)
	Update(ctx context.Context, provider *Provider) error
	Delete(ctx context.Context, id ulid.ULID) error

	// List operations
	List(ctx context.Context, limit, offset int) ([]*Provider, error)
	ListEnabled(ctx context.Context) ([]*Provider, error)
	ListByStatus(ctx context.Context, isEnabled bool, limit, offset int) ([]*Provider, error)

	// Search operations
	SearchProviders(ctx context.Context, filter *ProviderFilter) ([]*Provider, int, error)
	CountProviders(ctx context.Context, filter *ProviderFilter) (int64, error)

	// Batch operations
	CreateBatch(ctx context.Context, providers []*Provider) error
	UpdateBatch(ctx context.Context, providers []*Provider) error
	DeleteBatch(ctx context.Context, ids []ulid.ULID) error

	// Health operations
	UpdateHealthStatus(ctx context.Context, providerID ulid.ULID, status HealthStatus) error
	GetHealthyProviders(ctx context.Context) ([]*Provider, error)
}

// ModelRepository defines the interface for model data access
type ModelRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, model *Model) error
	GetByID(ctx context.Context, id ulid.ULID) (*Model, error)
	GetByModelName(ctx context.Context, modelName string) (*Model, error)
	GetByProviderAndModel(ctx context.Context, providerID ulid.ULID, modelName string) (*Model, error)
	Update(ctx context.Context, model *Model) error
	Delete(ctx context.Context, id ulid.ULID) error

	// Provider-scoped operations
	GetByProviderID(ctx context.Context, providerID ulid.ULID) ([]*Model, error)
	GetEnabledByProviderID(ctx context.Context, providerID ulid.ULID) ([]*Model, error)

	// Type and capability-based queries
	GetByModelType(ctx context.Context, modelType ModelType, limit, offset int) ([]*Model, error)
	GetStreamingModels(ctx context.Context) ([]*Model, error)
	GetFunctionModels(ctx context.Context) ([]*Model, error)
	GetVisionModels(ctx context.Context) ([]*Model, error)

	// Cost and performance queries
	GetModelsByCostRange(ctx context.Context, minCost, maxCost float64) ([]*Model, error)
	GetModelsByQualityRange(ctx context.Context, minQuality, maxQuality float64) ([]*Model, error)
	GetFastestModels(ctx context.Context, modelType ModelType, limit int) ([]*Model, error)
	GetCheapestModels(ctx context.Context, modelType ModelType, limit int) ([]*Model, error)

	// List operations
	List(ctx context.Context, limit, offset int) ([]*Model, error)
	ListEnabled(ctx context.Context) ([]*Model, error)
	ListWithProvider(ctx context.Context, limit, offset int) ([]*Model, error)

	// Search operations
	SearchModels(ctx context.Context, filter *ModelFilter) ([]*Model, int, error)
	CountModels(ctx context.Context, filter *ModelFilter) (int64, error)

	// Batch operations
	CreateBatch(ctx context.Context, models []*Model) error
	UpdateBatch(ctx context.Context, models []*Model) error
	DeleteBatch(ctx context.Context, ids []ulid.ULID) error

	// Availability operations
	GetAvailableModelsForProject(ctx context.Context, projectID ulid.ULID) ([]*Model, error)
	GetCompatibleModels(ctx context.Context, requirements *ModelRequirements) ([]*Model, error)
}

// ProviderConfigRepository defines the interface for provider configuration data access
type ProviderConfigRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, config *ProviderConfig) error
	GetByID(ctx context.Context, id ulid.ULID) (*ProviderConfig, error)
	GetByProjectAndProvider(ctx context.Context, projectID, providerID ulid.ULID) (*ProviderConfig, error)
	Update(ctx context.Context, config *ProviderConfig) error
	Delete(ctx context.Context, id ulid.ULID) error

	// Project-scoped operations
	GetByProjectID(ctx context.Context, projectID ulid.ULID) ([]*ProviderConfig, error)
	GetEnabledByProjectID(ctx context.Context, projectID ulid.ULID) ([]*ProviderConfig, error)
	GetByProjectIDWithProvider(ctx context.Context, projectID ulid.ULID) ([]*ProviderConfig, error)

	// Provider-scoped operations
	GetByProviderID(ctx context.Context, providerID ulid.ULID) ([]*ProviderConfig, error)
	CountProjectsForProvider(ctx context.Context, providerID ulid.ULID) (int64, error)

	// Priority and ordering
	GetOrderedByPriority(ctx context.Context, projectID ulid.ULID) ([]*ProviderConfig, error)
	UpdatePriority(ctx context.Context, configID ulid.ULID, priority int) error

	// List operations
	List(ctx context.Context, limit, offset int) ([]*ProviderConfig, error)
	ListEnabled(ctx context.Context) ([]*ProviderConfig, error)

	// Search operations
	SearchConfigs(ctx context.Context, filter *ProviderConfigFilter) ([]*ProviderConfig, int, error)
	CountConfigs(ctx context.Context, filter *ProviderConfigFilter) (int64, error)

	// Batch operations
	CreateBatch(ctx context.Context, configs []*ProviderConfig) error
	UpdateBatch(ctx context.Context, configs []*ProviderConfig) error
	DeleteBatch(ctx context.Context, ids []ulid.ULID) error

	// Encryption operations
	EncryptAPIKey(ctx context.Context, plaintext string) (string, error)
	DecryptAPIKey(ctx context.Context, encrypted string) (string, error)

	// Validation operations
	ValidateConfiguration(ctx context.Context, config *ProviderConfig) error
	TestProviderConnection(ctx context.Context, config *ProviderConfig) error
}

// HealthMetricsRepository defines the interface for provider health metrics data access
type HealthMetricsRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, metrics *ProviderHealthMetrics) error
	GetByID(ctx context.Context, id ulid.ULID) (*ProviderHealthMetrics, error)
	Update(ctx context.Context, metrics *ProviderHealthMetrics) error
	Delete(ctx context.Context, id ulid.ULID) error

	// Provider-scoped operations
	GetByProviderID(ctx context.Context, providerID ulid.ULID, limit, offset int) ([]*ProviderHealthMetrics, error)
	GetLatestByProviderID(ctx context.Context, providerID ulid.ULID) (*ProviderHealthMetrics, error)
	GetByProviderIDAndTimeRange(ctx context.Context, providerID ulid.ULID, startTime, endTime time.Time) ([]*ProviderHealthMetrics, error)

	// Health status operations
	GetByHealthStatus(ctx context.Context, status HealthStatus, limit, offset int) ([]*ProviderHealthMetrics, error)
	GetUnhealthyProviders(ctx context.Context) ([]*ProviderHealthMetrics, error)
	UpdateProviderHealth(ctx context.Context, providerID ulid.ULID, status HealthStatus, metrics *HealthMetricsData) error

	// Time-based queries
	GetByTimeRange(ctx context.Context, startTime, endTime time.Time, limit, offset int) ([]*ProviderHealthMetrics, error)
	GetRecentMetrics(ctx context.Context, providerID ulid.ULID, duration time.Duration) ([]*ProviderHealthMetrics, error)

	// Aggregation operations
	GetAverageLatency(ctx context.Context, providerID ulid.ULID, duration time.Duration) (float64, error)
	GetSuccessRate(ctx context.Context, providerID ulid.ULID, duration time.Duration) (float64, error)
	GetUptimePercentage(ctx context.Context, providerID ulid.ULID, duration time.Duration) (float64, error)

	// List operations
	List(ctx context.Context, limit, offset int) ([]*ProviderHealthMetrics, error)
	ListByProvider(ctx context.Context, providerID ulid.ULID, limit, offset int) ([]*ProviderHealthMetrics, error)

	// Batch operations
	CreateBatch(ctx context.Context, metrics []*ProviderHealthMetrics) error
	UpdateBatch(ctx context.Context, metrics []*ProviderHealthMetrics) error
	DeleteBatch(ctx context.Context, ids []ulid.ULID) error

	// Cleanup operations
	DeleteOldMetrics(ctx context.Context, olderThan time.Time) (int64, error)
	ArchiveMetrics(ctx context.Context, providerID ulid.ULID, olderThan time.Time) error
}

// RoutingRuleRepository defines the interface for routing rule data access
type RoutingRuleRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, rule *RoutingRule) error
	GetByID(ctx context.Context, id ulid.ULID) (*RoutingRule, error)
	GetByProjectAndName(ctx context.Context, projectID ulid.ULID, ruleName string) (*RoutingRule, error)
	Update(ctx context.Context, rule *RoutingRule) error
	Delete(ctx context.Context, id ulid.ULID) error

	// Project-scoped operations
	GetByProjectID(ctx context.Context, projectID ulid.ULID) ([]*RoutingRule, error)
	GetEnabledByProjectID(ctx context.Context, projectID ulid.ULID) ([]*RoutingRule, error)
	GetOrderedByPriority(ctx context.Context, projectID ulid.ULID) ([]*RoutingRule, error)

	// Rule matching operations
	GetMatchingRules(ctx context.Context, projectID ulid.ULID, conditions map[string]interface{}) ([]*RoutingRule, error)
	GetRulesByStrategy(ctx context.Context, strategy RoutingStrategy) ([]*RoutingRule, error)

	// List operations
	List(ctx context.Context, limit, offset int) ([]*RoutingRule, error)
	ListEnabled(ctx context.Context) ([]*RoutingRule, error)

	// Search operations
	SearchRules(ctx context.Context, filter *RoutingRuleFilter) ([]*RoutingRule, int, error)
	CountRules(ctx context.Context, filter *RoutingRuleFilter) (int64, error)

	// Batch operations
	CreateBatch(ctx context.Context, rules []*RoutingRule) error
	UpdateBatch(ctx context.Context, rules []*RoutingRule) error
	DeleteBatch(ctx context.Context, ids []ulid.ULID) error

	// Priority management
	UpdatePriority(ctx context.Context, ruleID ulid.ULID, priority int) error
	ReorderRules(ctx context.Context, projectID ulid.ULID, ruleIDs []ulid.ULID) error

	// Validation operations
	ValidateRule(ctx context.Context, rule *RoutingRule) error
	TestRuleMatching(ctx context.Context, rule *RoutingRule, testConditions map[string]interface{}) (bool, error)
}

// CacheRepository defines the interface for request cache data access
type CacheRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, cache *RequestCache) error
	GetByID(ctx context.Context, id ulid.ULID) (*RequestCache, error)
	GetByCacheKey(ctx context.Context, cacheKey string) (*RequestCache, error)
	Update(ctx context.Context, cache *RequestCache) error
	Delete(ctx context.Context, id ulid.ULID) error

	// Cache key operations
	GetByProjectAndKey(ctx context.Context, projectID ulid.ULID, cacheKey string) (*RequestCache, error)
	GetByRequestHash(ctx context.Context, projectID ulid.ULID, requestHash string) (*RequestCache, error)

	// Project-scoped operations
	GetByProjectID(ctx context.Context, projectID ulid.ULID, limit, offset int) ([]*RequestCache, error)
	GetProjectCacheStats(ctx context.Context, projectID ulid.ULID) (*CacheStats, error)

	// Model-scoped operations
	GetByModelName(ctx context.Context, modelName string, limit, offset int) ([]*RequestCache, error)
	GetModelCacheStats(ctx context.Context, modelName string) (*CacheStats, error)

	// Time-based operations
	GetByTimeRange(ctx context.Context, startTime, endTime time.Time, limit, offset int) ([]*RequestCache, error)
	GetRecentCache(ctx context.Context, projectID ulid.ULID, duration time.Duration) ([]*RequestCache, error)

	// Cache hit operations
	IncrementHit(ctx context.Context, cacheKey string) error
	UpdateLastAccessed(ctx context.Context, cacheKey string) error

	// List operations
	List(ctx context.Context, limit, offset int) ([]*RequestCache, error)
	ListExpired(ctx context.Context) ([]*RequestCache, error)
	ListByHitCount(ctx context.Context, minHits int, limit, offset int) ([]*RequestCache, error)

	// Batch operations
	CreateBatch(ctx context.Context, caches []*RequestCache) error
	UpdateBatch(ctx context.Context, caches []*RequestCache) error
	DeleteBatch(ctx context.Context, ids []ulid.ULID) error

	// Cleanup operations
	DeleteExpired(ctx context.Context) (int64, error)
	DeleteByProjectID(ctx context.Context, projectID ulid.ULID) (int64, error)
	DeleteOldCache(ctx context.Context, olderThan time.Time) (int64, error)

	// Statistics operations
	GetCacheStats(ctx context.Context, filter *CacheFilter) (*CacheStats, error)
	GetHitRate(ctx context.Context, projectID ulid.ULID, duration time.Duration) (float64, error)
	GetCacheSizeByProject(ctx context.Context, projectID ulid.ULID) (int64, error)
}

// Filter structures for repository queries

// ProviderFilter represents filtering options for provider queries
type ProviderFilter struct {
	ProjectID     *ulid.ULID    `json:"project_id,omitempty"`
	ProviderType  *ProviderType `json:"provider_type,omitempty"`
	IsEnabled     *bool         `json:"is_enabled,omitempty"`
	HealthStatus  *HealthStatus `json:"health_status,omitempty"`
	Search        *string       `json:"search,omitempty"`
	CreatedAfter  *time.Time    `json:"created_after,omitempty"`
	CreatedBefore *time.Time    `json:"created_before,omitempty"`
}

// ModelFilter represents filtering options for model queries
type ModelFilter struct {
	ProviderID        *ulid.ULID `json:"provider_id,omitempty"`
	ModelType         *ModelType `json:"model_type,omitempty"`
	IsEnabled         *bool      `json:"is_enabled,omitempty"`
	SupportsStreaming *bool      `json:"supports_streaming,omitempty"`
	SupportsFunctions *bool      `json:"supports_functions,omitempty"`
	SupportsVision    *bool      `json:"supports_vision,omitempty"`
	MinCostPer1k      *float64   `json:"min_cost_per_1k,omitempty"`
	MaxCostPer1k      *float64   `json:"max_cost_per_1k,omitempty"`
	MinContextTokens  *int       `json:"min_context_tokens,omitempty"`
	MaxContextTokens  *int       `json:"max_context_tokens,omitempty"`
	Search            *string    `json:"search,omitempty"`
}

// ProviderConfigFilter represents filtering options for provider config queries
type ProviderConfigFilter struct {
	ProjectID   *ulid.ULID `json:"project_id,omitempty"`
	ProviderID  *ulid.ULID `json:"provider_id,omitempty"`
	IsEnabled   *bool      `json:"is_enabled,omitempty"`
	MinPriority *int       `json:"min_priority,omitempty"`
	MaxPriority *int       `json:"max_priority,omitempty"`
}

// RoutingRuleFilter represents filtering options for routing rule queries
type RoutingRuleFilter struct {
	ProjectID       *ulid.ULID       `json:"project_id,omitempty"`
	IsEnabled       *bool            `json:"is_enabled,omitempty"`
	RoutingStrategy *RoutingStrategy `json:"routing_strategy,omitempty"`
	MinPriority     *int             `json:"min_priority,omitempty"`
	MaxPriority     *int             `json:"max_priority,omitempty"`
	CreatedBy       *ulid.ULID       `json:"created_by,omitempty"`
}

// CacheFilter represents filtering options for cache queries
type CacheFilter struct {
	ProjectID     *ulid.ULID `json:"project_id,omitempty"`
	ModelName     *string    `json:"model_name,omitempty"`
	MinHitCount   *int       `json:"min_hit_count,omitempty"`
	IsExpired     *bool      `json:"is_expired,omitempty"`
	CreatedAfter  *time.Time `json:"created_after,omitempty"`
	CreatedBefore *time.Time `json:"created_before,omitempty"`
}

// Supporting structures

// ModelRequirements defines requirements for model compatibility
type ModelRequirements struct {
	ModelType         ModelType `json:"model_type"`
	SupportsStreaming *bool     `json:"supports_streaming,omitempty"`
	SupportsFunctions *bool     `json:"supports_functions,omitempty"`
	SupportsVision    *bool     `json:"supports_vision,omitempty"`
	MinContextTokens  *int      `json:"min_context_tokens,omitempty"`
	MaxCostPer1k      *float64  `json:"max_cost_per_1k,omitempty"`
	MinQualityScore   *float64  `json:"min_quality_score,omitempty"`
}

// HealthMetricsData contains health metrics data for updates
type HealthMetricsData struct {
	AvgLatencyMs      *int     `json:"avg_latency_ms,omitempty"`
	SuccessRate       *float64 `json:"success_rate,omitempty"`
	RequestsPerMinute *int     `json:"requests_per_minute,omitempty"`
	ErrorsPerMinute   *int     `json:"errors_per_minute,omitempty"`
	ResponseTimeP95   *int     `json:"response_time_p95,omitempty"`
	ResponseTimeP99   *int     `json:"response_time_p99,omitempty"`
	UptimePercentage  *float64 `json:"uptime_percentage,omitempty"`
	LastError         *string  `json:"last_error,omitempty"`
}

// CacheStats contains cache statistics
type CacheStats struct {
	TotalEntries   int64    `json:"total_entries"`
	HitRate        float64  `json:"hit_rate"`
	TotalHits      int64    `json:"total_hits"`
	TotalMisses    int64    `json:"total_misses"`
	TotalSize      int64    `json:"total_size"`
	ExpiredEntries int64    `json:"expired_entries"`
	AvgHitCount    float64  `json:"avg_hit_count"`
	TopModels      []string `json:"top_models"`
	CostSavings    float64  `json:"cost_savings"`
}
