package gateway

import "fmt"

// Domain errors for gateway operations
var (
	// Provider errors
	ErrProviderNotFound          = fmt.Errorf("provider not found")
	ErrProviderAlreadyExists     = fmt.Errorf("provider already exists")
	ErrProviderDisabled          = fmt.Errorf("provider is disabled")
	ErrProviderUnhealthy         = fmt.Errorf("provider is unhealthy")
	ErrProviderUnavailable       = fmt.Errorf("provider is unavailable")
	ErrInvalidProviderID         = fmt.Errorf("invalid provider ID")
	ErrInvalidProviderType       = fmt.Errorf("invalid provider type")
	ErrProviderConfigMissing     = fmt.Errorf("provider configuration missing")
	ErrProviderAPIKeyInvalid     = fmt.Errorf("provider API key is invalid")
	ErrProviderRateLimitExceeded = fmt.Errorf("provider rate limit exceeded")

	// Model errors
	ErrModelNotFound           = fmt.Errorf("model not found")
	ErrModelAlreadyExists      = fmt.Errorf("model already exists")
	ErrModelDisabled           = fmt.Errorf("model is disabled")
	ErrModelNotSupported       = fmt.Errorf("model not supported by provider")
	ErrInvalidModelID          = fmt.Errorf("invalid model ID")
	ErrInvalidModelType        = fmt.Errorf("invalid model type")
	ErrModelContextExceeded    = fmt.Errorf("model context length exceeded")
	ErrModelTokenLimitExceeded = fmt.Errorf("model token limit exceeded")

	// Routing errors
	ErrNoProvidersAvailable    = fmt.Errorf("no providers available for routing")
	ErrNoModelsAvailable       = fmt.Errorf("no models available for routing")
	ErrRoutingStrategyInvalid  = fmt.Errorf("invalid routing strategy")
	ErrRoutingDecisionFailed   = fmt.Errorf("routing decision failed")
	ErrAllProvidersFailed      = fmt.Errorf("all providers failed")
	ErrFallbackProviderFailed  = fmt.Errorf("fallback provider failed")
	ErrCircularRoutingDetected = fmt.Errorf("circular routing detected")

	// Configuration errors
	ErrProviderConfigNotFound = fmt.Errorf("provider configuration not found")
	ErrProviderConfigInvalid  = fmt.Errorf("provider configuration is invalid")
	ErrAPIKeyNotConfigured    = fmt.Errorf("API key not configured for provider")
	ErrAPIKeyEncryptionFailed = fmt.Errorf("API key encryption failed")
	ErrAPIKeyDecryptionFailed = fmt.Errorf("API key decryption failed")
	ErrConfigValidationFailed = fmt.Errorf("configuration validation failed")

	// Cache errors
	ErrCacheNotFound              = fmt.Errorf("cache entry not found")
	ErrCacheExpired               = fmt.Errorf("cache entry expired")
	ErrCacheKeyGenerationFailed   = fmt.Errorf("cache key generation failed")
	ErrCacheSerializationFailed   = fmt.Errorf("cache serialization failed")
	ErrCacheDeserializationFailed = fmt.Errorf("cache deserialization failed")

	// Request errors
	ErrRequestValidationFailed = fmt.Errorf("request validation failed")
	ErrRequestTimeout          = fmt.Errorf("request timeout")
	ErrRequestTooLarge         = fmt.Errorf("request too large")
	ErrInvalidRequestFormat    = fmt.Errorf("invalid request format")
	ErrUnsupportedRequestType  = fmt.Errorf("unsupported request type")
	ErrRequestQuotaExceeded    = fmt.Errorf("request quota exceeded")

	// Authentication errors
	ErrAuthenticationFailed    = fmt.Errorf("authentication failed")
	ErrAuthorizationFailed     = fmt.Errorf("authorization failed")
	ErrInvalidAPIKey           = fmt.Errorf("invalid API key")
	ErrAPIKeyExpired           = fmt.Errorf("API key expired")
	ErrInsufficientPermissions = fmt.Errorf("insufficient permissions")

	// Health monitoring errors
	ErrHealthCheckFailed        = fmt.Errorf("health check failed")
	ErrHealthMetricsUnavailable = fmt.Errorf("health metrics unavailable")
	ErrProviderHealthUnknown    = fmt.Errorf("provider health status unknown")

	// General validation errors
	ErrValidationFailed   = fmt.Errorf("validation failed")
	ErrInvalidProjectID   = fmt.Errorf("invalid project ID")
	ErrInvalidUserID      = fmt.Errorf("invalid user ID")
	ErrInvalidEnvironment = fmt.Errorf("invalid environment")
	ErrResourceNotFound   = fmt.Errorf("resource not found")
	ErrUnauthorizedAccess = fmt.Errorf("unauthorized access")

	// Operation errors
	ErrOperationFailed        = fmt.Errorf("operation failed")
	ErrConcurrentModification = fmt.Errorf("concurrent modification detected")
	ErrResourceLimitExceeded  = fmt.Errorf("resource limit exceeded")
	ErrInvalidFilter          = fmt.Errorf("invalid filter parameters")
	ErrInvalidPagination      = fmt.Errorf("invalid pagination parameters")
)

// GatewayError represents a structured error for gateway operations
type GatewayError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
	Cause   error                  `json:"-"`
}

// Error implements the error interface
func (e *GatewayError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying cause
func (e *GatewayError) Unwrap() error {
	return e.Cause
}

// NewGatewayError creates a new gateway error
func NewGatewayError(code, message string) *GatewayError {
	return &GatewayError{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// NewGatewayErrorWithCause creates a new gateway error with a cause
func NewGatewayErrorWithCause(code, message string, cause error) *GatewayError {
	return &GatewayError{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
		Cause:   cause,
	}
}

// WithDetail adds a detail to the error
func (e *GatewayError) WithDetail(key string, value interface{}) *GatewayError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// Error codes for different types of errors
const (
	// Provider error codes
	ErrCodeProviderNotFound          = "PROVIDER_NOT_FOUND"
	ErrCodeProviderAlreadyExists     = "PROVIDER_ALREADY_EXISTS"
	ErrCodeProviderDisabled          = "PROVIDER_DISABLED"
	ErrCodeProviderUnhealthy         = "PROVIDER_UNHEALTHY"
	ErrCodeProviderUnavailable       = "PROVIDER_UNAVAILABLE"
	ErrCodeInvalidProviderID         = "INVALID_PROVIDER_ID"
	ErrCodeInvalidProviderType       = "INVALID_PROVIDER_TYPE"
	ErrCodeProviderConfigMissing     = "PROVIDER_CONFIG_MISSING"
	ErrCodeProviderAPIKeyInvalid     = "PROVIDER_API_KEY_INVALID"
	ErrCodeProviderRateLimitExceeded = "PROVIDER_RATE_LIMIT_EXCEEDED"

	// Model error codes
	ErrCodeModelNotFound           = "MODEL_NOT_FOUND"
	ErrCodeModelAlreadyExists      = "MODEL_ALREADY_EXISTS"
	ErrCodeModelDisabled           = "MODEL_DISABLED"
	ErrCodeModelNotSupported       = "MODEL_NOT_SUPPORTED"
	ErrCodeInvalidModelID          = "INVALID_MODEL_ID"
	ErrCodeInvalidModelType        = "INVALID_MODEL_TYPE"
	ErrCodeModelContextExceeded    = "MODEL_CONTEXT_EXCEEDED"
	ErrCodeModelTokenLimitExceeded = "MODEL_TOKEN_LIMIT_EXCEEDED"

	// Routing error codes
	ErrCodeNoProvidersAvailable    = "NO_PROVIDERS_AVAILABLE"
	ErrCodeNoModelsAvailable       = "NO_MODELS_AVAILABLE"
	ErrCodeRoutingStrategyInvalid  = "ROUTING_STRATEGY_INVALID"
	ErrCodeRoutingDecisionFailed   = "ROUTING_DECISION_FAILED"
	ErrCodeAllProvidersFailed      = "ALL_PROVIDERS_FAILED"
	ErrCodeFallbackProviderFailed  = "FALLBACK_PROVIDER_FAILED"
	ErrCodeCircularRoutingDetected = "CIRCULAR_ROUTING_DETECTED"

	// Configuration error codes
	ErrCodeProviderConfigNotFound = "PROVIDER_CONFIG_NOT_FOUND"
	ErrCodeProviderConfigInvalid  = "PROVIDER_CONFIG_INVALID"
	ErrCodeAPIKeyNotConfigured    = "API_KEY_NOT_CONFIGURED"
	ErrCodeAPIKeyEncryptionFailed = "API_KEY_ENCRYPTION_FAILED"
	ErrCodeAPIKeyDecryptionFailed = "API_KEY_DECRYPTION_FAILED"
	ErrCodeConfigValidationFailed = "CONFIG_VALIDATION_FAILED"

	// Cache error codes
	ErrCodeCacheNotFound              = "CACHE_NOT_FOUND"
	ErrCodeCacheExpired               = "CACHE_EXPIRED"
	ErrCodeCacheKeyGenerationFailed   = "CACHE_KEY_GENERATION_FAILED"
	ErrCodeCacheSerializationFailed   = "CACHE_SERIALIZATION_FAILED"
	ErrCodeCacheDeserializationFailed = "CACHE_DESERIALIZATION_FAILED"

	// Request error codes
	ErrCodeRequestValidationFailed = "REQUEST_VALIDATION_FAILED"
	ErrCodeRequestTimeout          = "REQUEST_TIMEOUT"
	ErrCodeRequestTooLarge         = "REQUEST_TOO_LARGE"
	ErrCodeInvalidRequestFormat    = "INVALID_REQUEST_FORMAT"
	ErrCodeUnsupportedRequestType  = "UNSUPPORTED_REQUEST_TYPE"
	ErrCodeRequestQuotaExceeded    = "REQUEST_QUOTA_EXCEEDED"

	// Authentication error codes
	ErrCodeAuthenticationFailed    = "AUTHENTICATION_FAILED"
	ErrCodeAuthorizationFailed     = "AUTHORIZATION_FAILED"
	ErrCodeInvalidAPIKey           = "INVALID_API_KEY"
	ErrCodeAPIKeyExpired           = "API_KEY_EXPIRED"
	ErrCodeInsufficientPermissions = "INSUFFICIENT_PERMISSIONS"

	// Health monitoring error codes
	ErrCodeHealthCheckFailed        = "HEALTH_CHECK_FAILED"
	ErrCodeHealthMetricsUnavailable = "HEALTH_METRICS_UNAVAILABLE"
	ErrCodeProviderHealthUnknown    = "PROVIDER_HEALTH_UNKNOWN"

	// General validation error codes
	ErrCodeValidationFailed   = "VALIDATION_FAILED"
	ErrCodeInvalidProjectID   = "INVALID_PROJECT_ID"
	ErrCodeInvalidUserID      = "INVALID_USER_ID"
	ErrCodeInvalidEnvironment = "INVALID_ENVIRONMENT"
	ErrCodeResourceNotFound   = "RESOURCE_NOT_FOUND"
	ErrCodeUnauthorizedAccess = "UNAUTHORIZED_ACCESS"

	// Operation error codes
	ErrCodeOperationFailed        = "OPERATION_FAILED"
	ErrCodeConcurrentModification = "CONCURRENT_MODIFICATION"
	ErrCodeResourceLimitExceeded  = "RESOURCE_LIMIT_EXCEEDED"
	ErrCodeInvalidFilter          = "INVALID_FILTER"
	ErrCodeInvalidPagination      = "INVALID_PAGINATION"
)

// Convenience functions for creating common errors

// NewProviderNotFoundError creates a provider not found error
func NewProviderNotFoundError(providerID string) *GatewayError {
	return NewGatewayError(ErrCodeProviderNotFound, "provider not found").
		WithDetail("provider_id", providerID)
}

// NewModelNotFoundError creates a model not found error
func NewModelNotFoundError(modelName string) *GatewayError {
	return NewGatewayError(ErrCodeModelNotFound, "model not found").
		WithDetail("model_name", modelName)
}

// NewProviderConfigNotFoundError creates a provider config not found error
func NewProviderConfigNotFoundError(projectID, providerID string) *GatewayError {
	return NewGatewayError(ErrCodeProviderConfigNotFound, "provider configuration not found").
		WithDetail("project_id", projectID).
		WithDetail("provider_id", providerID)
}

// NewProviderConfigNotFoundByIDError creates a provider config not found error with just ID
func NewProviderConfigNotFoundByIDError(configID string) *GatewayError {
	return NewGatewayError(ErrCodeProviderConfigNotFound, "provider configuration not found").
		WithDetail("config_id", configID)
}

// NewValidationError creates a validation error with field details
func NewValidationError(field, message string) *GatewayError {
	return NewGatewayError(ErrCodeValidationFailed, "validation failed").
		WithDetail("field", field).
		WithDetail("message", message)
}

// NewValidationErrors creates a validation error with multiple field errors
func NewValidationErrors(fieldErrors []ValidationError) *GatewayError {
	err := NewGatewayError(ErrCodeValidationFailed, "validation failed")

	fields := make(map[string]string)
	for _, fieldErr := range fieldErrors {
		fields[fieldErr.Field] = fieldErr.Message
	}

	return err.WithDetail("field_errors", fields)
}

// NewProviderUnavailableError creates a provider unavailable error
func NewProviderUnavailableError(providerName string, reason string) *GatewayError {
	return NewGatewayError(ErrCodeProviderUnavailable, "provider is unavailable").
		WithDetail("provider", providerName).
		WithDetail("reason", reason)
}

// NewRoutingFailedError creates a routing failed error
func NewRoutingFailedError(modelName string, strategy string, reason string) *GatewayError {
	return NewGatewayError(ErrCodeRoutingDecisionFailed, "routing decision failed").
		WithDetail("model", modelName).
		WithDetail("strategy", strategy).
		WithDetail("reason", reason)
}

// NewAuthenticationError creates an authentication error
func NewAuthenticationError(provider string) *GatewayError {
	return NewGatewayError(ErrCodeAuthenticationFailed, "authentication failed").
		WithDetail("provider", provider)
}

// NewRateLimitError creates a rate limit error
func NewRateLimitError(provider string, limit int, resetTime string) *GatewayError {
	return NewGatewayError(ErrCodeProviderRateLimitExceeded, "provider rate limit exceeded").
		WithDetail("provider", provider).
		WithDetail("limit", limit).
		WithDetail("reset_time", resetTime)
}

// NewRequestTimeoutError creates a request timeout error
func NewRequestTimeoutError(provider string, timeout int) *GatewayError {
	return NewGatewayError(ErrCodeRequestTimeout, "request timeout").
		WithDetail("provider", provider).
		WithDetail("timeout_seconds", timeout)
}

// NewTokenLimitError creates a token limit error
func NewTokenLimitError(modelName string, requestTokens int, maxTokens int) *GatewayError {
	return NewGatewayError(ErrCodeModelTokenLimitExceeded, "model token limit exceeded").
		WithDetail("model", modelName).
		WithDetail("request_tokens", requestTokens).
		WithDetail("max_tokens", maxTokens)
}

// NewQuotaExceededError creates a quota exceeded error
func NewQuotaExceededError(projectID string, quotaType string, limit int, used int) *GatewayError {
	return NewGatewayError(ErrCodeRequestQuotaExceeded, "request quota exceeded").
		WithDetail("project_id", projectID).
		WithDetail("quota_type", quotaType).
		WithDetail("limit", limit).
		WithDetail("used", used)
}

// Error classification functions

// IsNotFoundError checks if the error is a not found error
func IsNotFoundError(err error) bool {
	if gatewayErr, ok := err.(*GatewayError); ok {
		return gatewayErr.Code == ErrCodeProviderNotFound ||
			gatewayErr.Code == ErrCodeModelNotFound ||
			gatewayErr.Code == ErrCodeProviderConfigNotFound ||
			gatewayErr.Code == ErrCodeCacheNotFound ||
			gatewayErr.Code == ErrCodeResourceNotFound
	}
	return false
}

// IsValidationError checks if the error is a validation error
func IsValidationError(err error) bool {
	if gatewayErr, ok := err.(*GatewayError); ok {
		return gatewayErr.Code == ErrCodeValidationFailed ||
			gatewayErr.Code == ErrCodeInvalidProviderID ||
			gatewayErr.Code == ErrCodeInvalidModelID ||
			gatewayErr.Code == ErrCodeInvalidProviderType ||
			gatewayErr.Code == ErrCodeInvalidModelType ||
			gatewayErr.Code == ErrCodeInvalidRequestFormat ||
			gatewayErr.Code == ErrCodeConfigValidationFailed
	}
	return false
}

// IsAuthenticationError checks if the error is an authentication error
func IsAuthenticationError(err error) bool {
	if gatewayErr, ok := err.(*GatewayError); ok {
		return gatewayErr.Code == ErrCodeAuthenticationFailed ||
			gatewayErr.Code == ErrCodeAuthorizationFailed ||
			gatewayErr.Code == ErrCodeInvalidAPIKey ||
			gatewayErr.Code == ErrCodeAPIKeyExpired ||
			gatewayErr.Code == ErrCodeInsufficientPermissions
	}
	return false
}

// IsRateLimitError checks if the error is a rate limit error
func IsRateLimitError(err error) bool {
	if gatewayErr, ok := err.(*GatewayError); ok {
		return gatewayErr.Code == ErrCodeProviderRateLimitExceeded
	}
	return false
}

// IsTimeoutError checks if the error is a timeout error
func IsTimeoutError(err error) bool {
	if gatewayErr, ok := err.(*GatewayError); ok {
		return gatewayErr.Code == ErrCodeRequestTimeout
	}
	return false
}

// IsQuotaExceededError checks if the error is a quota exceeded error
func IsQuotaExceededError(err error) bool {
	if gatewayErr, ok := err.(*GatewayError); ok {
		return gatewayErr.Code == ErrCodeRequestQuotaExceeded ||
			gatewayErr.Code == ErrCodeModelTokenLimitExceeded ||
			gatewayErr.Code == ErrCodeResourceLimitExceeded
	}
	return false
}

// IsProviderError checks if the error is related to provider issues
func IsProviderError(err error) bool {
	if gatewayErr, ok := err.(*GatewayError); ok {
		return gatewayErr.Code == ErrCodeProviderNotFound ||
			gatewayErr.Code == ErrCodeProviderDisabled ||
			gatewayErr.Code == ErrCodeProviderUnhealthy ||
			gatewayErr.Code == ErrCodeProviderUnavailable ||
			gatewayErr.Code == ErrCodeAllProvidersFailed ||
			gatewayErr.Code == ErrCodeFallbackProviderFailed
	}
	return false
}

// IsRoutingError checks if the error is related to routing issues
func IsRoutingError(err error) bool {
	if gatewayErr, ok := err.(*GatewayError); ok {
		return gatewayErr.Code == ErrCodeNoProvidersAvailable ||
			gatewayErr.Code == ErrCodeNoModelsAvailable ||
			gatewayErr.Code == ErrCodeRoutingStrategyInvalid ||
			gatewayErr.Code == ErrCodeRoutingDecisionFailed ||
			gatewayErr.Code == ErrCodeCircularRoutingDetected
	}
	return false
}

// IsRetryableError checks if the error is retryable
func IsRetryableError(err error) bool {
	if gatewayErr, ok := err.(*GatewayError); ok {
		// These errors are typically retryable
		return gatewayErr.Code == ErrCodeRequestTimeout ||
			gatewayErr.Code == ErrCodeProviderUnavailable ||
			gatewayErr.Code == ErrCodeProviderRateLimitExceeded ||
			gatewayErr.Code == ErrCodeHealthCheckFailed
	}
	return false
}
