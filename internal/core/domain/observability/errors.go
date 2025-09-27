package observability

import "fmt"

// Domain errors for observability operations
var (
	// Trace errors
	ErrTraceNotFound         = fmt.Errorf("trace not found")
	ErrTraceAlreadyExists    = fmt.Errorf("trace already exists")
	ErrInvalidTraceID        = fmt.Errorf("invalid trace ID")
	ErrExternalTraceIDExists = fmt.Errorf("external trace ID already exists")

	// Observation errors
	ErrObservationNotFound         = fmt.Errorf("observation not found")
	ErrObservationAlreadyExists    = fmt.Errorf("observation already exists")
	ErrInvalidObservationID        = fmt.Errorf("invalid observation ID")
	ErrObservationTraceNotFound    = fmt.Errorf("observation trace not found")
	ErrInvalidObservationType      = fmt.Errorf("invalid observation type")
	ErrObservationAlreadyCompleted = fmt.Errorf("observation already completed")

	// Quality score errors
	ErrQualityScoreNotFound    = fmt.Errorf("quality score not found")
	ErrInvalidQualityScoreID   = fmt.Errorf("invalid quality score ID")
	ErrInvalidScoreValue       = fmt.Errorf("invalid score value")
	ErrInvalidScoreDataType    = fmt.Errorf("invalid score data type")
	ErrEvaluatorNotFound       = fmt.Errorf("evaluator not found")
	ErrDuplicateQualityScore   = fmt.Errorf("duplicate quality score for the same trace/observation and score name")

	// General validation errors
	ErrValidationFailed     = fmt.Errorf("validation failed")
	ErrInvalidProjectID     = fmt.Errorf("invalid project ID")
	ErrInvalidUserID        = fmt.Errorf("invalid user ID")
	ErrInvalidSessionID     = fmt.Errorf("invalid session ID")
	ErrUnauthorizedAccess   = fmt.Errorf("unauthorized access")
	ErrInsufficientPermissions = fmt.Errorf("insufficient permissions")

	// Operation errors
	ErrBatchOperationFailed = fmt.Errorf("batch operation failed")
	ErrConcurrentModification = fmt.Errorf("concurrent modification detected")
	ErrResourceLimitExceeded = fmt.Errorf("resource limit exceeded")
	ErrInvalidFilter         = fmt.Errorf("invalid filter parameters")
	ErrInvalidPagination     = fmt.Errorf("invalid pagination parameters")
)

// ObservabilityError represents a structured error for observability operations
type ObservabilityError struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Cause     error                  `json:"-"`
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

	// Observation error codes
	ErrCodeObservationNotFound              = "OBSERVATION_NOT_FOUND"
	ErrCodeObservationAlreadyExists         = "OBSERVATION_ALREADY_EXISTS"
	ErrCodeInvalidObservationID             = "INVALID_OBSERVATION_ID"
	ErrCodeObservationTraceNotFound         = "OBSERVATION_TRACE_NOT_FOUND"
	ErrCodeInvalidObservationType           = "INVALID_OBSERVATION_TYPE"
	ErrCodeObservationAlreadyCompleted      = "OBSERVATION_ALREADY_COMPLETED"
	ErrCodeExternalObservationIDExists      = "EXTERNAL_OBSERVATION_ID_EXISTS"
	ErrCodeValidation                       = "VALIDATION_ERROR"

	// Quality score error codes
	ErrCodeQualityScoreNotFound  = "QUALITY_SCORE_NOT_FOUND"
	ErrCodeInvalidQualityScoreID = "INVALID_QUALITY_SCORE_ID"
	ErrCodeInvalidScoreValue     = "INVALID_SCORE_VALUE"
	ErrCodeInvalidScoreDataType  = "INVALID_SCORE_DATA_TYPE"
	ErrCodeEvaluatorNotFound     = "EVALUATOR_NOT_FOUND"
	ErrCodeDuplicateQualityScore = "DUPLICATE_QUALITY_SCORE"

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

// NewObservationNotFoundError creates an observation not found error
func NewObservationNotFoundError(observationID string) *ObservabilityError {
	return NewObservabilityError(ErrCodeObservationNotFound, "observation not found").
		WithDetail("observation_id", observationID)
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
			obsErr.Code == ErrCodeObservationNotFound ||
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
			obsErr.Code == ErrCodeInvalidObservationID ||
			obsErr.Code == ErrCodeInvalidQualityScoreID ||
			obsErr.Code == ErrCodeInvalidObservationType ||
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
			obsErr.Code == ErrCodeObservationAlreadyExists ||
			obsErr.Code == ErrCodeExternalTraceIDExists ||
			obsErr.Code == ErrCodeDuplicateQualityScore ||
			obsErr.Code == ErrCodeConcurrentModification
	}
	return false
}