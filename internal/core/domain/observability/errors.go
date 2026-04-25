package observability

import "errors"

// Domain errors for observability operations.
//
// Span/trace/score sentinels live in this package only when actively
// produced by a repository and consumed by a service via errors.Is.
// Anything else gets deleted on sight per the producer/consumer parity
// rule (CLAUDE.md "Sentinel producer/consumer must match").
var (
	// Filter preset CRUD on the dashboard plane.
	ErrFilterPresetNotFound = errors.New("filter preset not found")
)

// ObservabilityValidationError represents a per-field validation failure
// on a span query / entity validation request. Returned by
// ValidateSpanQueryRequest and entity-level Validate() methods.
type ObservabilityValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
