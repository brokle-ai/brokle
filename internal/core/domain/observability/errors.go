package observability

import (
	"errors"
	"fmt"
)

// Domain errors for observability operations
var (
	// Trace errors
	ErrTraceNotFound         = errors.New("trace not found")
	ErrTraceAlreadyExists    = errors.New("trace already exists")
	ErrInvalidTraceID        = errors.New("invalid trace ID")
	ErrExternalTraceIDExists = errors.New("external trace ID already exists")

	// Span errors
	ErrSpanNotFound         = errors.New("span not found")
	ErrSpanAlreadyExists    = errors.New("span already exists")
	ErrInvalidSpanID        = errors.New("invalid span ID")
	ErrSpanTraceNotFound    = errors.New("span trace not found")
	ErrInvalidSpanType      = errors.New("invalid span type")
	ErrSpanAlreadyCompleted = errors.New("span already completed")

	// Quality score errors
	ErrQualityScoreNotFound  = errors.New("quality score not found")
	ErrInvalidQualityScoreID = errors.New("invalid quality score ID")
	ErrInvalidScoreValue     = errors.New("invalid score value")
	ErrInvalidScoreDataType  = errors.New("invalid score data type")
	ErrEvaluatorNotFound     = errors.New("evaluator not found")
	ErrDuplicateQualityScore = errors.New("duplicate quality score for the same trace/span and score name")

	// Model pricing errors
	ErrModelNotFound         = errors.New("model not found")
	ErrInvalidPricingPattern = errors.New("invalid pricing pattern")
	ErrPricingDataIncomplete = errors.New("incomplete pricing data")
	ErrPricingExpired        = errors.New("pricing has expired")
	ErrInvalidPricingData    = errors.New("invalid pricing data")

	// General validation errors
	ErrValidationFailed        = errors.New("validation failed")
	ErrInvalidProjectID        = errors.New("invalid project ID")
	ErrInvalidUserID           = errors.New("invalid user ID")
	ErrInvalidSessionID        = errors.New("invalid session ID")
	ErrUnauthorizedAccess      = errors.New("unauthorized access")
	ErrInsufficientPermissions = errors.New("insufficient permissions")

	// Operation errors
	ErrBatchOperationFailed   = errors.New("batch operation failed")
	ErrConcurrentModification = errors.New("concurrent modification detected")
	ErrResourceLimitExceeded  = errors.New("resource limit exceeded")
	ErrInvalidFilter          = errors.New("invalid filter parameters")
	ErrInvalidPagination      = errors.New("invalid pagination parameters")
)

// ObservabilityError represents a structured error for observability operations
type ObservabilityError struct {
	Cause   error                  `json:"-"`
	Details map[string]interface{} `json:"details,omitempty"`
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
}

// Error implements the error interface
func (e *ObservabilityError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying cause
func (e *ObservabilityError) Unwrap() error {
	return e.Cause
}

// NewObservabilityError creates a new observability error
func NewObservabilityError(code, message string) *ObservabilityError {
	return &ObservabilityError{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// NewObservabilityErrorWithCause creates a new observability error with a cause
func NewObservabilityErrorWithCause(code, message string, cause error) *ObservabilityError {
	return &ObservabilityError{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
		Cause:   cause,
	}
}

// WithDetail adds a detail to the error
func (e *ObservabilityError) WithDetail(key string, value interface{}) *ObservabilityError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// Error codes for different types of errors
const (
	// Trace error codes
	ErrCodeTraceNotFound         = "TRACE_NOT_FOUND"
	ErrCodeTraceAlreadyExists    = "TRACE_ALREADY_EXISTS"
	ErrCodeInvalidTraceID        = "INVALID_TRACE_ID"
	ErrCodeExternalTraceIDExists = "EXTERNAL_TRACE_ID_EXISTS"

	// Span error codes
	ErrCodeSpanNotFound         = "SPAN_NOT_FOUND"
	ErrCodeSpanAlreadyExists    = "SPAN_ALREADY_EXISTS"
	ErrCodeInvalidSpanID        = "INVALID_SPAN_ID"
	ErrCodeSpanTraceNotFound    = "SPAN_TRACE_NOT_FOUND"
	ErrCodeInvalidSpanType      = "INVALID_SPAN_TYPE"
	ErrCodeSpanAlreadyCompleted = "SPAN_ALREADY_COMPLETED"
	ErrCodeExternalSpanIDExists = "EXTERNAL_SPAN_ID_EXISTS"
	ErrCodeValidation           = "VALIDATION_ERROR"

	// Quality score error codes
	ErrCodeQualityScoreNotFound  = "QUALITY_SCORE_NOT_FOUND"
	ErrCodeInvalidQualityScoreID = "INVALID_QUALITY_SCORE_ID"
	ErrCodeInvalidScoreValue     = "INVALID_SCORE_VALUE"
	ErrCodeInvalidScoreDataType  = "INVALID_SCORE_DATA_TYPE"
	ErrCodeEvaluatorNotFound     = "EVALUATOR_NOT_FOUND"
	ErrCodeDuplicateQualityScore = "DUPLICATE_QUALITY_SCORE"

	// Model pricing error codes
	ErrCodeModelNotFound         = "MODEL_NOT_FOUND"
	ErrCodeInvalidPricingPattern = "INVALID_PRICING_PATTERN"
	ErrCodePricingDataIncomplete = "PRICING_DATA_INCOMPLETE"
	ErrCodePricingExpired        = "PRICING_EXPIRED"
	ErrCodeInvalidPricingData    = "INVALID_PRICING_DATA"

	// General validation error codes
	ErrCodeValidationFailed        = "VALIDATION_FAILED"
	ErrCodeInvalidProjectID        = "INVALID_PROJECT_ID"
	ErrCodeInvalidUserID           = "INVALID_USER_ID"
	ErrCodeInvalidSessionID        = "INVALID_SESSION_ID"
	ErrCodeUnauthorizedAccess      = "UNAUTHORIZED_ACCESS"
	ErrCodeInsufficientPermissions = "INSUFFICIENT_PERMISSIONS"

	// Operation error codes
	ErrCodeBatchOperationFailed   = "BATCH_OPERATION_FAILED"
	ErrCodeConcurrentModification = "CONCURRENT_MODIFICATION"
	ErrCodeResourceLimitExceeded  = "RESOURCE_LIMIT_EXCEEDED"
	ErrCodeInvalidFilter          = "INVALID_FILTER"
	ErrCodeInvalidPagination      = "INVALID_PAGINATION"
)

// Convenience functions for creating common errors

// NewTraceNotFoundError creates a trace not found error
func NewTraceNotFoundError(traceID string) *ObservabilityError {
	return NewObservabilityError(ErrCodeTraceNotFound, "trace not found").
		WithDetail("trace_id", traceID)
}

// NewSpanNotFoundError creates a span not found error
func NewSpanNotFoundError(spanID string) *ObservabilityError {
	return NewObservabilityError(ErrCodeSpanNotFound, "span not found").
		WithDetail("span_id", spanID)
}

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// NewValidationError creates a validation error with field details
func NewValidationError(field, message string) *ObservabilityError {
	return NewObservabilityError(ErrCodeValidationFailed, "validation failed").
		WithDetail("field", field).
		WithDetail("message", message)
}

// NewValidationErrors creates a validation error with multiple field errors
func NewValidationErrors(fieldErrors []ValidationError) *ObservabilityError {
	err := NewObservabilityError(ErrCodeValidationFailed, "validation failed")

	fields := make(map[string]string)
	for _, fieldErr := range fieldErrors {
		fields[fieldErr.Field] = fieldErr.Message
	}

	return err.WithDetail("field_errors", fields)
}

// NewUnauthorizedError creates an unauthorized access error
func NewUnauthorizedError(resource string) *ObservabilityError {
	return NewObservabilityError(ErrCodeUnauthorizedAccess, "unauthorized access").
		WithDetail("resource", resource)
}

// NewInsufficientPermissionsError creates an insufficient permissions error
func NewInsufficientPermissionsError(operation string) *ObservabilityError {
	return NewObservabilityError(ErrCodeInsufficientPermissions, "insufficient permissions").
		WithDetail("operation", operation)
}

// NewBatchOperationError creates a batch operation error
func NewBatchOperationError(operation string, cause error) *ObservabilityError {
	return NewObservabilityErrorWithCause(ErrCodeBatchOperationFailed, "batch operation failed", cause).
		WithDetail("operation", operation)
}

// NewResourceLimitError creates a resource limit exceeded error
func NewResourceLimitError(resource string, limit int) *ObservabilityError {
	return NewObservabilityError(ErrCodeResourceLimitExceeded, "resource limit exceeded").
		WithDetail("resource", resource).
		WithDetail("limit", limit)
}

// IsNotFoundError checks if the error is a not found error
func IsNotFoundError(err error) bool {
	if obsErr, ok := err.(*ObservabilityError); ok {
		return obsErr.Code == ErrCodeTraceNotFound ||
			obsErr.Code == ErrCodeSpanNotFound ||
			obsErr.Code == ErrCodeQualityScoreNotFound ||
			obsErr.Code == ErrCodeEvaluatorNotFound
	}
	return false
}

// IsValidationError checks if the error is a validation error
func IsValidationError(err error) bool {
	if obsErr, ok := err.(*ObservabilityError); ok {
		return obsErr.Code == ErrCodeValidationFailed ||
			obsErr.Code == ErrCodeInvalidTraceID ||
			obsErr.Code == ErrCodeInvalidSpanID ||
			obsErr.Code == ErrCodeInvalidQualityScoreID ||
			obsErr.Code == ErrCodeInvalidSpanType ||
			obsErr.Code == ErrCodeInvalidScoreValue ||
			obsErr.Code == ErrCodeInvalidScoreDataType
	}
	return false
}

// IsUnauthorizedError checks if the error is an authorization error
func IsUnauthorizedError(err error) bool {
	if obsErr, ok := err.(*ObservabilityError); ok {
		return obsErr.Code == ErrCodeUnauthorizedAccess ||
			obsErr.Code == ErrCodeInsufficientPermissions
	}
	return false
}

// IsConflictError checks if the error is a conflict error
func IsConflictError(err error) bool {
	if obsErr, ok := err.(*ObservabilityError); ok {
		return obsErr.Code == ErrCodeTraceAlreadyExists ||
			obsErr.Code == ErrCodeSpanAlreadyExists ||
			obsErr.Code == ErrCodeExternalTraceIDExists ||
			obsErr.Code == ErrCodeDuplicateQualityScore ||
			obsErr.Code == ErrCodeConcurrentModification
	}
	return false
}
