package observability

// ObservabilityValidationError represents a per-field validation failure
// on a span query / entity validation request. Returned by
// ValidateSpanQueryRequest and entity-level Validate() methods.
type ObservabilityValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
