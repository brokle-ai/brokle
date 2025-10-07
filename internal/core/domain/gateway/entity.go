package gateway

import (
	"encoding/json"
	"fmt"
	"time"

	"brokle/pkg/ulid"
)

// Provider represents an AI provider (OpenAI, Anthropic, Cohere, etc.)
type Provider struct {
	ID                  ulid.ULID              `json:"id" db:"id"`
	Name                string                 `json:"name" db:"name"`
	Type                ProviderType           `json:"type" db:"type"`
	BaseURL             string                 `json:"base_url" db:"base_url"`
	IsEnabled           bool                   `json:"is_enabled" db:"is_enabled"`
	DefaultTimeoutSecs  int                    `json:"default_timeout_seconds" db:"default_timeout_seconds"`
	MaxRetries          int                    `json:"max_retries" db:"max_retries"`
	HealthCheckURL      *string                `json:"health_check_url,omitempty" db:"health_check_url"`
	SupportedFeatures   map[string]interface{} `json:"supported_features" db:"supported_features"`
	RateLimits          map[string]interface{} `json:"rate_limits" db:"rate_limits"`
	CreatedAt           time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at" db:"updated_at"`
}

// Model represents an AI model with pricing and capabilities
type Model struct {
	ID                    ulid.ULID  `json:"id" db:"id"`
	ProviderID            ulid.ULID  `json:"provider_id" db:"provider_id"`
	ModelName             string     `json:"model_name" db:"model_name"`
	DisplayName           string     `json:"display_name" db:"display_name"`
	InputCostPer1kTokens  float64    `json:"input_cost_per_1k_tokens" db:"input_cost_per_1k_tokens"`
	OutputCostPer1kTokens float64    `json:"output_cost_per_1k_tokens" db:"output_cost_per_1k_tokens"`
	MaxContextTokens      int        `json:"max_context_tokens" db:"max_context_tokens"`
	SupportsStreaming     bool       `json:"supports_streaming" db:"supports_streaming"`
	SupportsFunctions     bool       `json:"supports_functions" db:"supports_functions"`
	SupportsVision        bool       `json:"supports_vision" db:"supports_vision"`
	ModelType             ModelType  `json:"model_type" db:"model_type"`
	IsEnabled             bool       `json:"is_enabled" db:"is_enabled"`
	QualityScore          *float64   `json:"quality_score,omitempty" db:"quality_score"`
	SpeedScore            *float64   `json:"speed_score,omitempty" db:"speed_score"`
	Metadata              map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt             time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at" db:"updated_at"`

	// Populated relationships
	Provider *Provider `json:"provider,omitempty" db:"-"`
}

// ProviderConfig represents project-scoped provider configuration
type ProviderConfig struct {
	ID                   ulid.ULID              `json:"id" db:"id"`
	ProjectID            ulid.ULID              `json:"project_id" db:"project_id"`
	ProviderID           ulid.ULID              `json:"provider_id" db:"provider_id"`
	APIKeyEncrypted      string                 `json:"-" db:"api_key_encrypted"` // Never expose in JSON
	IsEnabled            bool                   `json:"is_enabled" db:"is_enabled"`
	CustomBaseURL        *string                `json:"custom_base_url,omitempty" db:"custom_base_url"`
	CustomTimeoutSecs    *int                   `json:"custom_timeout_seconds,omitempty" db:"custom_timeout_seconds"`
	RateLimitOverride    map[string]interface{} `json:"rate_limit_override" db:"rate_limit_override"`
	PriorityOrder        int                    `json:"priority_order" db:"priority_order"`
	Configuration        map[string]interface{} `json:"configuration" db:"configuration"`
	CreatedAt            time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at" db:"updated_at"`

	// Populated relationships
	Provider *Provider `json:"provider,omitempty" db:"-"`
}

// ProviderHealthMetrics represents provider health and performance data
type ProviderHealthMetrics struct {
	ID                   ulid.ULID  `json:"id" db:"id"`
	ProviderID           ulid.ULID  `json:"provider_id" db:"provider_id"`
	Timestamp            time.Time  `json:"timestamp" db:"timestamp"`
	Status               HealthStatus `json:"status" db:"status"`
	AvgLatencyMs         *int       `json:"avg_latency_ms,omitempty" db:"avg_latency_ms"`
	SuccessRate          *float64   `json:"success_rate,omitempty" db:"success_rate"`
	RequestsPerMinute    *int       `json:"requests_per_minute,omitempty" db:"requests_per_minute"`
	ErrorsPerMinute      *int       `json:"errors_per_minute,omitempty" db:"errors_per_minute"`
	LastError            *string    `json:"last_error,omitempty" db:"last_error"`
	ResponseTimeP95      *int       `json:"response_time_p95,omitempty" db:"response_time_p95"`
	ResponseTimeP99      *int       `json:"response_time_p99,omitempty" db:"response_time_p99"`
	UptimePercentage     *float64   `json:"uptime_percentage,omitempty" db:"uptime_percentage"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`

	// Populated relationships
	Provider *Provider `json:"provider,omitempty" db:"-"`
}

// RoutingRule represents project-specific routing configuration
type RoutingRule struct {
	ID                ulid.ULID              `json:"id" db:"id"`
	ProjectID         ulid.ULID              `json:"project_id" db:"project_id"`
	RuleName          string                 `json:"rule_name" db:"rule_name"`
	IsEnabled         bool                   `json:"is_enabled" db:"is_enabled"`
	Priority          int                    `json:"priority" db:"priority"`
	Conditions        map[string]interface{} `json:"conditions" db:"conditions"`
	RoutingStrategy   RoutingStrategy        `json:"routing_strategy" db:"routing_strategy"`
	TargetProviders   []ProviderWeight       `json:"target_providers" db:"-"`
	FallbackProviders []ProviderWeight       `json:"fallback_providers" db:"-"`
	RateLimits        map[string]interface{} `json:"rate_limits" db:"rate_limits"`
	CreatedBy         ulid.ULID              `json:"created_by" db:"created_by"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" db:"updated_at"`
}

// ProviderWeight represents a provider with routing weight
type ProviderWeight struct {
	ProviderID ulid.ULID `json:"provider_id"`
	Weight     int       `json:"weight"`
}

// RequestCache represents cached AI responses for semantic caching
type RequestCache struct {
	ID               ulid.ULID              `json:"id" db:"id"`
	CacheKey         string                 `json:"cache_key" db:"cache_key"`
	ProjectID        ulid.ULID              `json:"project_id" db:"project_id"`
	ModelName        string                 `json:"model_name" db:"model_name"`
	RequestHash      string                 `json:"request_hash" db:"request_hash"`
	ResponseData     map[string]interface{} `json:"response_data" db:"response_data"`
	TokenUsage       TokenUsage             `json:"token_usage" db:"token_usage"`
	CostUSD          float64                `json:"cost_usd" db:"cost_usd"`
	HitCount         int                    `json:"hit_count" db:"hit_count"`
	LastAccessedAt   time.Time              `json:"last_accessed_at" db:"last_accessed_at"`
	ExpiresAt        time.Time              `json:"expires_at" db:"expires_at"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
}

// TokenUsage represents token usage information
type TokenUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// RoutingDecision represents the result of routing logic
type RoutingDecision struct {
	Provider         string    `json:"provider"`
	ProviderID       ulid.ULID `json:"provider_id"`
	Model            string    `json:"model"`
	ModelID          ulid.ULID `json:"model_id"`
	Strategy         string    `json:"strategy"`
	EstimatedCost    float64   `json:"estimated_cost"`
	EstimatedLatency int       `json:"estimated_latency"`
	ProviderHealth   float64   `json:"provider_health"`
	RoutingReason    string    `json:"routing_reason"`
	FallbackOptions  []string  `json:"fallback_options,omitempty"`
}

// RequestMetrics represents metrics for a gateway request
type RequestMetrics struct {
	RequestID          string        `json:"request_id"`
	ProjectID          ulid.ULID     `json:"project_id"`
	Environment        string        `json:"environment"`
	Provider           string        `json:"provider"`
	Model              string        `json:"model"`
	RoutingStrategy    string        `json:"routing_strategy"`
	InputTokens        int           `json:"input_tokens"`
	OutputTokens       int           `json:"output_tokens"`
	TotalTokens        int           `json:"total_tokens"`
	CostUSD            float64       `json:"cost_usd"`
	LatencyMs          int           `json:"latency_ms"`
	CacheHit           bool          `json:"cache_hit"`
	FallbackTriggered  bool          `json:"fallback_triggered"`
	PrimaryProvider    string        `json:"primary_provider,omitempty"`
	FallbackProvider   string        `json:"fallback_provider,omitempty"`
	QualityScore       *float64      `json:"quality_score,omitempty"`
	Success            bool          `json:"success"`
	ErrorCode          *string       `json:"error_code,omitempty"`
	ErrorMessage       *string       `json:"error_message,omitempty"`
	Timestamp          time.Time     `json:"timestamp"`
}

// Validation methods

// Validate validates a Provider entity
func (p *Provider) Validate() []ValidationError {
	var errors []ValidationError

	if p.Name == "" {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "provider name is required",
		})
	}

	if p.Type == "" {
		errors = append(errors, ValidationError{
			Field:   "type",
			Message: "provider type is required",
		})
	}

	if !p.isValidProviderType() {
		errors = append(errors, ValidationError{
			Field:   "type",
			Message: "invalid provider type",
		})
	}

	if p.BaseURL == "" {
		errors = append(errors, ValidationError{
			Field:   "base_url",
			Message: "base URL is required",
		})
	}

	if p.DefaultTimeoutSecs < 1 || p.DefaultTimeoutSecs > 300 {
		errors = append(errors, ValidationError{
			Field:   "default_timeout_seconds",
			Message: "timeout must be between 1 and 300 seconds",
		})
	}

	if p.MaxRetries < 0 || p.MaxRetries > 10 {
		errors = append(errors, ValidationError{
			Field:   "max_retries",
			Message: "max retries must be between 0 and 10",
		})
	}

	return errors
}

// Validate validates a Model entity
func (m *Model) Validate() []ValidationError {
	var errors []ValidationError

	if m.ModelName == "" {
		errors = append(errors, ValidationError{
			Field:   "model_name",
			Message: "model name is required",
		})
	}

	if m.DisplayName == "" {
		errors = append(errors, ValidationError{
			Field:   "display_name",
			Message: "display name is required",
		})
	}

	if m.InputCostPer1kTokens < 0 {
		errors = append(errors, ValidationError{
			Field:   "input_cost_per_1k_tokens",
			Message: "input cost cannot be negative",
		})
	}

	if m.OutputCostPer1kTokens < 0 {
		errors = append(errors, ValidationError{
			Field:   "output_cost_per_1k_tokens",
			Message: "output cost cannot be negative",
		})
	}

	if m.MaxContextTokens < 1 {
		errors = append(errors, ValidationError{
			Field:   "max_context_tokens",
			Message: "max context tokens must be at least 1",
		})
	}

	if !m.isValidModelType() {
		errors = append(errors, ValidationError{
			Field:   "model_type",
			Message: "invalid model type",
		})
	}

	if m.QualityScore != nil && (*m.QualityScore < 0 || *m.QualityScore > 1) {
		errors = append(errors, ValidationError{
			Field:   "quality_score",
			Message: "quality score must be between 0 and 1",
		})
	}

	if m.SpeedScore != nil && (*m.SpeedScore < 0 || *m.SpeedScore > 1) {
		errors = append(errors, ValidationError{
			Field:   "speed_score",
			Message: "speed score must be between 0 and 1",
		})
	}

	return errors
}

// Validate validates a ProviderConfig entity
func (pc *ProviderConfig) Validate() []ValidationError {
	var errors []ValidationError

	if pc.APIKeyEncrypted == "" {
		errors = append(errors, ValidationError{
			Field:   "api_key_encrypted",
			Message: "API key is required",
		})
	}

	if pc.PriorityOrder < 0 {
		errors = append(errors, ValidationError{
			Field:   "priority_order",
			Message: "priority order cannot be negative",
		})
	}

	if pc.CustomTimeoutSecs != nil && (*pc.CustomTimeoutSecs < 1 || *pc.CustomTimeoutSecs > 300) {
		errors = append(errors, ValidationError{
			Field:   "custom_timeout_seconds",
			Message: "timeout must be between 1 and 300 seconds",
		})
	}

	return errors
}

// Helper validation methods

func (p *Provider) isValidProviderType() bool {
	switch p.Type {
	case ProviderTypeOpenAI, ProviderTypeAnthropic, ProviderTypeCohere, ProviderTypeGoogle:
		return true
	default:
		return false
	}
}

func (m *Model) isValidModelType() bool {
	switch m.ModelType {
	case ModelTypeText, ModelTypeEmbedding, ModelTypeImage, ModelTypeAudio:
		return true
	default:
		return false
	}
}

// JSON marshaling helpers to handle JSONB fields

// MarshalSupportedFeatures marshals supported features to JSON
func (p *Provider) MarshalSupportedFeatures() ([]byte, error) {
	return json.Marshal(p.SupportedFeatures)
}

// UnmarshalSupportedFeatures unmarshals supported features from JSON
func (p *Provider) UnmarshalSupportedFeatures(data []byte) error {
	return json.Unmarshal(data, &p.SupportedFeatures)
}

// MarshalRateLimits marshals rate limits to JSON
func (p *Provider) MarshalRateLimits() ([]byte, error) {
	return json.Marshal(p.RateLimits)
}

// UnmarshalRateLimits unmarshals rate limits from JSON
func (p *Provider) UnmarshalRateLimits(data []byte) error {
	return json.Unmarshal(data, &p.RateLimits)
}

// GetEffectiveTimeout returns the effective timeout for a provider config
func (pc *ProviderConfig) GetEffectiveTimeout(defaultTimeout int) int {
	if pc.CustomTimeoutSecs != nil {
		return *pc.CustomTimeoutSecs
	}
	return defaultTimeout
}

// GetEffectiveBaseURL returns the effective base URL for a provider config
func (pc *ProviderConfig) GetEffectiveBaseURL(defaultURL string) string {
	if pc.CustomBaseURL != nil && *pc.CustomBaseURL != "" {
		return *pc.CustomBaseURL
	}
	return defaultURL
}

// CalculateCost calculates the total cost for a request
func (m *Model) CalculateCost(inputTokens, outputTokens int) float64 {
	inputCost := (float64(inputTokens) / 1000.0) * m.InputCostPer1kTokens
	outputCost := (float64(outputTokens) / 1000.0) * m.OutputCostPer1kTokens
	return inputCost + outputCost
}

// IsExpired checks if a cache entry has expired
func (rc *RequestCache) IsExpired() bool {
	return time.Now().After(rc.ExpiresAt)
}

// IncrementHit increments the cache hit count and updates last accessed time
func (rc *RequestCache) IncrementHit() {
	rc.HitCount++
	rc.LastAccessedAt = time.Now()
}

// ValidationError represents a domain validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error implements the error interface
func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}