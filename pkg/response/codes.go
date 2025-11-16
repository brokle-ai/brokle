package response

import "net/http"

// Standard HTTP status codes for responses
const (
	StatusOK                  = http.StatusOK                  // 200
	StatusCreated             = http.StatusCreated             // 201
	StatusAccepted            = http.StatusAccepted            // 202
	StatusNoContent           = http.StatusNoContent           // 204
	StatusBadRequest          = http.StatusBadRequest          // 400
	StatusUnauthorized        = http.StatusUnauthorized        // 401
	StatusPaymentRequired     = http.StatusPaymentRequired     // 402
	StatusForbidden           = http.StatusForbidden           // 403
	StatusNotFound            = http.StatusNotFound            // 404
	StatusMethodNotAllowed    = http.StatusMethodNotAllowed    // 405
	StatusConflict            = http.StatusConflict            // 409
	StatusUnprocessableEntity = http.StatusUnprocessableEntity // 422
	StatusTooManyRequests     = http.StatusTooManyRequests     // 429
	StatusInternalServerError = http.StatusInternalServerError // 500
	StatusNotImplemented      = http.StatusNotImplemented      // 501
	StatusBadGateway          = http.StatusBadGateway          // 502
	StatusServiceUnavailable  = http.StatusServiceUnavailable  // 503
)

// Success response codes
const (
	CodeSuccess   = "SUCCESS"
	CodeCreated   = "CREATED"
	CodeUpdated   = "UPDATED"
	CodeDeleted   = "DELETED"
	CodeNoContent = "NO_CONTENT"
	CodeAccepted  = "ACCEPTED"
)

// Client error response codes (4xx)
const (
	CodeBadRequest       = "BAD_REQUEST"
	CodeUnauthorized     = "UNAUTHORIZED"
	CodePaymentRequired  = "PAYMENT_REQUIRED"
	CodeForbidden        = "FORBIDDEN"
	CodeNotFound         = "NOT_FOUND"
	CodeMethodNotAllowed = "METHOD_NOT_ALLOWED"
	CodeConflict         = "CONFLICT"
	CodeValidationFailed = "VALIDATION_FAILED"
	CodeTooManyRequests  = "TOO_MANY_REQUESTS"
	CodeQuotaExceeded    = "QUOTA_EXCEEDED"
)

// Server error response codes (5xx)
const (
	CodeInternalServerError = "INTERNAL_SERVER_ERROR"
	CodeNotImplemented      = "NOT_IMPLEMENTED"
	CodeBadGateway          = "BAD_GATEWAY"
	CodeServiceUnavailable  = "SERVICE_UNAVAILABLE"
	CodeGatewayTimeout      = "GATEWAY_TIMEOUT"
)

// AI Platform specific response codes
const (
	// Authentication & Authorization
	CodeTokenExpired      = "TOKEN_EXPIRED"
	CodeTokenInvalid      = "TOKEN_INVALID"
	CodeAPIKeyInvalid     = "API_KEY_INVALID"
	CodeInsufficientScope = "INSUFFICIENT_SCOPE"

	// Resource Management
	CodeResourceNotFound = "RESOURCE_NOT_FOUND"
	CodeResourceExists   = "RESOURCE_EXISTS"
	CodeResourceInactive = "RESOURCE_INACTIVE"
	CodeResourceLocked   = "RESOURCE_LOCKED"

	// AI Provider Integration
	CodeProviderUnavailable = "PROVIDER_UNAVAILABLE"
	CodeProviderError       = "PROVIDER_ERROR"
	CodeProviderTimeout     = "PROVIDER_TIMEOUT"
	CodeProviderRateLimit   = "PROVIDER_RATE_LIMIT"
	CodeModelUnsupported    = "MODEL_UNSUPPORTED"
	CodeModelConfigInvalid  = "MODEL_CONFIG_INVALID"

	// Billing & Usage
	CodeInsufficientCredits  = "INSUFFICIENT_CREDITS"
	CodeBillingSetupRequired = "BILLING_SETUP_REQUIRED"
	CodeSubscriptionExpired  = "SUBSCRIPTION_EXPIRED"
	CodeUsageQuotaExceeded   = "USAGE_QUOTA_EXCEEDED"

	// Analytics & Metrics
	CodeMetricsUnavailable   = "METRICS_UNAVAILABLE"
	CodeAnalyticsQueryFailed = "ANALYTICS_QUERY_FAILED"
	CodeDataNotAvailable     = "DATA_NOT_AVAILABLE"

	// Configuration & Settings
	CodeConfigurationInvalid = "CONFIGURATION_INVALID"
	CodeFeatureNotEnabled    = "FEATURE_NOT_ENABLED"
	CodeSettingsLocked       = "SETTINGS_LOCKED"

	// Real-time & WebSocket
	CodeWebSocketFailed       = "WEBSOCKET_FAILED"
	CodeEventDeliveryFailed   = "EVENT_DELIVERY_FAILED"
	CodeStreamingNotSupported = "STREAMING_NOT_SUPPORTED"

	// Cache & Performance
	CodeCacheUnavailable  = "CACHE_UNAVAILABLE"
	CodeSemanticCacheMiss = "SEMANTIC_CACHE_MISS"
	CodeRoutingFailed     = "ROUTING_FAILED"
)

// Response code to HTTP status code mapping
var CodeToStatusMap = map[string]int{
	// Success codes
	CodeSuccess:   StatusOK,
	CodeCreated:   StatusCreated,
	CodeUpdated:   StatusOK,
	CodeDeleted:   StatusNoContent,
	CodeNoContent: StatusNoContent,
	CodeAccepted:  StatusAccepted,

	// Client error codes
	CodeBadRequest:       StatusBadRequest,
	CodeUnauthorized:     StatusUnauthorized,
	CodePaymentRequired:  StatusPaymentRequired,
	CodeForbidden:        StatusForbidden,
	CodeNotFound:         StatusNotFound,
	CodeMethodNotAllowed: StatusMethodNotAllowed,
	CodeConflict:         StatusConflict,
	CodeValidationFailed: StatusUnprocessableEntity,
	CodeTooManyRequests:  StatusTooManyRequests,
	CodeQuotaExceeded:    StatusTooManyRequests,

	// Server error codes
	CodeInternalServerError: StatusInternalServerError,
	CodeNotImplemented:      StatusNotImplemented,
	CodeBadGateway:          StatusBadGateway,
	CodeServiceUnavailable:  StatusServiceUnavailable,
	CodeGatewayTimeout:      http.StatusGatewayTimeout,

	// AI Platform specific codes
	CodeTokenExpired:          StatusUnauthorized,
	CodeTokenInvalid:          StatusUnauthorized,
	CodeAPIKeyInvalid:         StatusUnauthorized,
	CodeInsufficientScope:     StatusForbidden,
	CodeResourceNotFound:      StatusNotFound,
	CodeResourceExists:        StatusConflict,
	CodeResourceInactive:      StatusForbidden,
	CodeResourceLocked:        StatusLocked,
	CodeProviderUnavailable:   StatusServiceUnavailable,
	CodeProviderError:         StatusBadGateway,
	CodeProviderTimeout:       http.StatusGatewayTimeout,
	CodeProviderRateLimit:     StatusTooManyRequests,
	CodeModelUnsupported:      StatusBadRequest,
	CodeModelConfigInvalid:    StatusBadRequest,
	CodeInsufficientCredits:   StatusPaymentRequired,
	CodeBillingSetupRequired:  StatusPaymentRequired,
	CodeSubscriptionExpired:   StatusPaymentRequired,
	CodeUsageQuotaExceeded:    StatusTooManyRequests,
	CodeMetricsUnavailable:    StatusServiceUnavailable,
	CodeAnalyticsQueryFailed:  StatusInternalServerError,
	CodeDataNotAvailable:      StatusNotFound,
	CodeConfigurationInvalid:  StatusBadRequest,
	CodeFeatureNotEnabled:     StatusForbidden,
	CodeSettingsLocked:        StatusForbidden,
	CodeWebSocketFailed:       StatusInternalServerError,
	CodeEventDeliveryFailed:   StatusInternalServerError,
	CodeStreamingNotSupported: StatusNotImplemented,
	CodeCacheUnavailable:      StatusServiceUnavailable,
	CodeSemanticCacheMiss:     StatusNotFound,
	CodeRoutingFailed:         StatusInternalServerError,
}

// Additional HTTP status codes not in standard library
const (
	StatusLocked = 423
)

// GetStatusCode returns the HTTP status code for a given response code
func GetStatusCode(code string) int {
	if statusCode, exists := CodeToStatusMap[code]; exists {
		return statusCode
	}
	return StatusInternalServerError
}

// IsSuccessCode returns true if the code represents a successful operation
func IsSuccessCode(code string) bool {
	statusCode := GetStatusCode(code)
	return statusCode >= 200 && statusCode < 300
}

// IsClientErrorCode returns true if the code represents a client error
func IsClientErrorCode(code string) bool {
	statusCode := GetStatusCode(code)
	return statusCode >= 400 && statusCode < 500
}

// IsServerErrorCode returns true if the code represents a server error
func IsServerErrorCode(code string) bool {
	statusCode := GetStatusCode(code)
	return statusCode >= 500 && statusCode < 600
}
