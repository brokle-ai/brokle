package observability

import (
	"fmt"
	"strings"

	"brokle/internal/core/domain/observability"
	"brokle/pkg/ulid"
)

// ValidateTelemetryBatchRequest validates the incoming telemetry batch request
func ValidateTelemetryBatchRequest(req *TelemetryBatchRequest) []string {
	var errors []string

	if req == nil {
		return []string{"Request cannot be nil"}
	}

	// Validate events
	if len(req.Events) == 0 {
		errors = append(errors, "Events list cannot be empty")
	}

	if len(req.Events) > 1000 {
		errors = append(errors, "Batch size exceeds maximum limit of 1000 events")
	}

	// Validate each event
	for i, event := range req.Events {
		if eventErrors := validateTelemetryEvent(event, i); len(eventErrors) > 0 {
			errors = append(errors, eventErrors...)
		}
	}

	// Note: Deduplication is always enforced with server-controlled 24h TTL (production-grade pattern)

	return errors
}

// validateTelemetryEvent validates an individual telemetry event
func validateTelemetryEvent(event *TelemetryEventRequest, index int) []string {
	var errors []string

	if event == nil {
		return []string{fmt.Sprintf("Event at index %d cannot be nil", index)}
	}

	// Validate event ID
	if event.EventID == "" {
		errors = append(errors, fmt.Sprintf("Event at index %d has empty event ID", index))
	} else {
		if _, err := ulid.Parse(event.EventID); err != nil {
			errors = append(errors, fmt.Sprintf("Event at index %d has invalid ULID format: %v", index, err))
		}
	}

	// Validate event type
	if event.EventType == "" {
		errors = append(errors, fmt.Sprintf("Event at index %d has empty event type", index))
	} else {
		if !isValidTelemetryEventType(event.EventType) {
			errors = append(errors, fmt.Sprintf("Event at index %d has invalid event type: %s", index, event.EventType))
		}
	}

	// Validate payload
	if len(event.Payload) == 0 {
		errors = append(errors, fmt.Sprintf("Event at index %d has empty payload", index))
	}

	// Validate timestamp if provided
	if event.Timestamp != nil && *event.Timestamp < 0 {
		errors = append(errors, fmt.Sprintf("Event at index %d has invalid timestamp (negative value)", index))
	}

	return errors
}

// isValidTelemetryEventType checks if the event type is valid
func isValidTelemetryEventType(eventType string) bool {
	validTypes := []string{
		string(observability.TelemetryEventTypeEvent),
		string(observability.TelemetryEventTypeTrace),
		string(observability.TelemetryEventTypeObservation),
		string(observability.TelemetryEventTypeQualityScore),
	}

	for _, validType := range validTypes {
		if eventType == validType {
			return true
		}
	}
	return false
}

// GetValidTelemetryEventTypes returns a list of valid telemetry event types
func GetValidTelemetryEventTypes() []string {
	return []string{
		string(observability.TelemetryEventTypeEvent),
		string(observability.TelemetryEventTypeTrace),
		string(observability.TelemetryEventTypeObservation),
		string(observability.TelemetryEventTypeQualityScore),
	}
}

// FormatValidationErrors formats validation errors into a readable string
func FormatValidationErrors(errors []string) string {
	if len(errors) == 0 {
		return ""
	}

	if len(errors) == 1 {
		return errors[0]
	}

	return strings.Join(errors, "; ")
}

// ValidateEnvironmentTag validates the environment tag format
func ValidateEnvironmentTag(environment string) error {
	if environment == "" {
		return nil // Environment is optional
	}

	// Environment tag validation rules
	if len(environment) > 50 {
		return fmt.Errorf("environment tag cannot exceed 50 characters")
	}

	// Allow alphanumeric, hyphens, underscores, and dots
	for _, char := range environment {
		if !((char >= 'a' && char <= 'z') ||
			 (char >= 'A' && char <= 'Z') ||
			 (char >= '0' && char <= '9') ||
			 char == '-' || char == '_' || char == '.') {
			return fmt.Errorf("environment tag contains invalid character: %c", char)
		}
	}

	return nil
}

// ValidateULIDString validates a ULID string format
func ValidateULIDString(ulidStr string) error {
	if ulidStr == "" {
		return fmt.Errorf("ULID cannot be empty")
	}

	_, err := ulid.Parse(ulidStr)
	if err != nil {
		return fmt.Errorf("invalid ULID format: %v", err)
	}

	return nil
}

// SanitizeMetadata sanitizes metadata to prevent oversized payloads
func SanitizeMetadata(metadata map[string]interface{}) map[string]interface{} {
	if metadata == nil {
		return nil
	}

	sanitized := make(map[string]interface{})

	for key, value := range metadata {
		// Limit key length
		if len(key) > 100 {
			key = key[:100]
		}

		// Limit string value length
		if strValue, ok := value.(string); ok {
			if len(strValue) > 1000 {
				strValue = strValue[:1000]
			}
			sanitized[key] = strValue
		} else {
			sanitized[key] = value
		}
	}

	return sanitized
}

// CalculateRequestSize estimates the size of a telemetry batch request in bytes
func CalculateRequestSize(req *TelemetryBatchRequest) int {
	if req == nil {
		return 0
	}

	size := 0

	// Environment string
	if req.Environment != nil {
		size += len(*req.Environment)
	}

	// Metadata (rough estimation)
	for key, value := range req.Metadata {
		size += len(key)
		if strValue, ok := value.(string); ok {
			size += len(strValue)
		} else {
			size += 50 // Rough estimate for other types
		}
	}

	// Events
	for _, event := range req.Events {
		size += len(event.EventID)
		size += len(event.EventType)

		// Payload (rough estimation)
		for key, value := range event.Payload {
			size += len(key)
			if strValue, ok := value.(string); ok {
				size += len(strValue)
			} else {
				size += 50 // Rough estimate for other types
			}
		}
	}

	return size
}

// ValidateRequestSize validates that the request size is within limits
func ValidateRequestSize(req *TelemetryBatchRequest) error {
	const maxRequestSize = 10 * 1024 * 1024 // 10MB limit

	size := CalculateRequestSize(req)
	if size > maxRequestSize {
		return fmt.Errorf("request size (%d bytes) exceeds maximum limit of %d bytes", size, maxRequestSize)
	}

	return nil
}