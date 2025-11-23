package observability

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"

	"brokle/internal/core/domain/analytics"
	"brokle/internal/core/domain/observability"
	"brokle/pkg/ulid"
)

// MaxAttributeValueSize defines the maximum size for input/output attribute values
// Matches common OTEL collector limits (1MB) to prevent oversized spans
const MaxAttributeValueSize = 1024 * 1024 // 1MB

// OTLPConverterService handles conversion of OTLP traces to Brokle telemetry events
type OTLPConverterService struct {
	logger                 *logrus.Logger
	providerPricingService analytics.ProviderPricingService
}

// NewOTLPConverterService creates a new OTLP converter service
func NewOTLPConverterService(
	logger *logrus.Logger,
	providerPricingService analytics.ProviderPricingService,
) *OTLPConverterService {
	return &OTLPConverterService{
		logger:                 logger,
		providerPricingService: providerPricingService,
	}
}

// brokleEvent represents an internal converted event (before domain conversion)
type brokleEvent struct {
	Payload   map[string]interface{} `json:"payload"`
	Timestamp *int64                 `json:"timestamp,omitempty"`
	EventID   string                 `json:"event_id"`
	SpanID    string                 `json:"span_id"`
	TraceID   string                 `json:"trace_id"`
	EventType string                 `json:"event_type"`
}

// isRootSpanCheck determines if a span is a root span by checking if parent ID is nil, empty, or zero
// Most OTLP exporters (OpenTelemetry Collector, Java/Go SDKs) populate parentSpanId with zero values
// instead of omitting the field, so we need to check for all these cases:
// - nil (field omitted)
// - empty string ""
// - zero hex string "0000000000000000"
// - zero bytes array [0,0,0,0,0,0,0,0]
// - zero bytes in map format {data: [0,0,0,0,0,0,0,0]}
func isRootSpanCheck(parentSpanID interface{}) bool {
	// Case 1: No parent ID field (nil)
	if parentSpanID == nil {
		return true
	}

	// Case 2: Empty string or zero hex string
	if str, ok := parentSpanID.(string); ok {
		if str == "" || str == "0000000000000000" {
			return true
		}
	}

	// Case 3: Zero bytes in map format {data: Buffer}
	if mapVal, ok := parentSpanID.(map[string]interface{}); ok {
		if data, ok := mapVal["data"].([]interface{}); ok {
			// Check if all bytes are zero
			allZero := true
			for _, b := range data {
				if intVal, ok := b.(float64); ok && intVal != 0 {
					allZero = false
					break
				}
			}
			if allZero {
				return true
			}
		}
	}

	// Case 4: Zero bytes array
	if bytes, ok := parentSpanID.([]byte); ok {
		allZero := true
		for _, b := range bytes {
			if b != 0 {
				allZero = false
				break
			}
		}
		return allZero
	}

	return false
}

// ConvertOTLPToBrokleEvents converts OTLP resourceSpans to Brokle telemetry events
// For root spans, creates both trace AND span events
// ctx is used for cost calculation database queries
// projectID is required for cost calculation (model pricing lookup)
func (s *OTLPConverterService) ConvertOTLPToBrokleEvents(ctx context.Context, otlpReq *observability.OTLPRequest, projectID string) ([]*observability.TelemetryEventRequest, error) {
	var internalEvents []*brokleEvent
	tracesCreated := make(map[string]bool) // Track which traces we've created

	for _, resourceSpan := range otlpReq.ResourceSpans {
		// Extract resource attributes
		resourceAttrs := extractAttributes(resourceSpan.Resource)

		for _, scopeSpan := range resourceSpan.ScopeSpans {
			// Extract scope attributes
			scopeAttrs := extractAttributes(scopeSpan.Scope)

			for _, span := range scopeSpan.Spans {
				// Get trace ID
				traceID, err := convertTraceID(span.TraceID)
				if err != nil {
					return nil, fmt.Errorf("invalid trace_id: %w", err)
				}

				// Check if this is a root span (no parent or zero/empty parent)
				isRootSpan := isRootSpanCheck(span.ParentSpanID)

				// If root span and we haven't created a trace event yet, create it
				if isRootSpan && !tracesCreated[traceID] {
					traceEvent, err := s.createTraceEvent(span, resourceAttrs, scopeAttrs, scopeSpan.Scope, traceID)
					if err != nil {
						return nil, fmt.Errorf("failed to create trace event: %w", err)
					}
					internalEvents = append(internalEvents, traceEvent)
					tracesCreated[traceID] = true
				}

				// Convert OTLP span to Brokle span event (with cost calculation)
				obsEvent, err := s.createSpanEvent(ctx, span, resourceAttrs, scopeAttrs, scopeSpan.Scope, projectID)
				if err != nil {
					return nil, fmt.Errorf("failed to create span event: %w", err)
				}
				internalEvents = append(internalEvents, obsEvent)
			}
		}
	}

	// Convert internal events to domain events
	return convertToDomainEvents(internalEvents), nil
}

// truncateWithIndicator truncates a string value and sets a truncation indicator
// Returns the (possibly truncated) value and whether it was truncated
func truncateWithIndicator(value string, maxSize int) (string, bool) {
	if len(value) <= maxSize {
		return value, false
	}
	return value[:maxSize] + "...[truncated]", true
}

// validateMimeType validates MIME type against actual content, auto-detecting if missing
// Returns validated/corrected MIME type
func validateMimeType(value string, declaredMimeType string) string {
	// If MIME type missing, auto-detect
	if declaredMimeType == "" {
		if json.Valid([]byte(value)) {
			return "application/json"
		}
		return "text/plain"
	}

	// Validate declared type matches content
	if declaredMimeType == "application/json" && !json.Valid([]byte(value)) {
		// Content is not valid JSON but declared as JSON - override to text/plain
		return "text/plain"
	}

	return declaredMimeType
}

// extractLLMMetadata extracts LLM-specific metadata from ChatML formatted input
// Returns metadata map with brokle.llm.* attributes (empty if not ChatML format)
func extractLLMMetadata(inputValue string) map[string]interface{} {
	metadata := make(map[string]interface{})

	// Try to parse as JSON array (ChatML format)
	var messages []map[string]interface{}
	if err := json.Unmarshal([]byte(inputValue), &messages); err != nil {
		return metadata // Not ChatML, return empty
	}

	// Validate ChatML structure
	if len(messages) == 0 {
		return metadata // Empty array, not ChatML
	}

	// Check first message has required "role" field (ChatML requirement)
	if _, hasRole := messages[0]["role"]; !hasRole {
		return metadata // Missing role field, not ChatML format
	}

	// Extract metadata
	metadata["brokle.llm.message_count"] = len(messages)

	// Count messages by role
	var userCount, assistantCount, systemCount, toolCount int
	var firstRole, lastRole string
	hasToolCalls := false

	for i, msg := range messages {
		role, _ := msg["role"].(string)

		// Track first and last roles
		if i == 0 {
			firstRole = role
		}
		if i == len(messages)-1 {
			lastRole = role
		}

		// Count by role
		switch role {
		case "user":
			userCount++
		case "assistant":
			assistantCount++
		case "system":
			systemCount++
		case "tool":
			toolCount++
		}

		// Check for tool calls
		if toolCalls, ok := msg["tool_calls"]; ok && toolCalls != nil {
			hasToolCalls = true
		}
	}

	// Add role counts
	if userCount > 0 {
		metadata["brokle.llm.user_message_count"] = userCount
	}
	if assistantCount > 0 {
		metadata["brokle.llm.assistant_message_count"] = assistantCount
	}
	if systemCount > 0 {
		metadata["brokle.llm.system_message_count"] = systemCount
	}
	if toolCount > 0 {
		metadata["brokle.llm.tool_message_count"] = toolCount
	}

	// Add role tracking
	if firstRole != "" {
		metadata["brokle.llm.first_role"] = firstRole
	}
	if lastRole != "" {
		metadata["brokle.llm.last_role"] = lastRole
	}

	// Add tool call indicator
	metadata["brokle.llm.has_tool_calls"] = hasToolCalls

	return metadata
}

// createTraceEvent creates a trace event from a root span
func (s *OTLPConverterService) createTraceEvent(span observability.OTLPSpan, resourceAttrs, scopeAttrs map[string]interface{}, scope *observability.Scope, traceID string) (*brokleEvent, error) {
	// Extract span attributes
	spanAttrs := extractAttributesFromKeyValues(span.Attributes)

	// Merge all attributes
	allAttrs := mergeAttributes(resourceAttrs, scopeAttrs, spanAttrs)

	// Convert times
	startTime := convertUnixNano(span.StartTimeUnixNano)
	endTime := convertUnixNano(span.EndTimeUnixNano)

	// Convert status
	statusCode := convertStatusCode(span.Status)

	// Build trace payload with new OTEL-native field names
	payload := map[string]interface{}{
		"trace_id":    traceID, // Renamed from "id"
		"name":        span.Name,
		"status_code": statusCode, // Now UInt8 (0-2)
	}

	// Add start_time if available
	if startTime != nil {
		payload["start_time"] = startTime.Format(time.RFC3339Nano)
	}

	if endTime != nil {
		payload["end_time"] = endTime.Format(time.RFC3339Nano)
		if startTime != nil {
			duration := uint32(endTime.Sub(*startTime).Milliseconds())
			payload["duration_ms"] = duration
		}
	}

	if span.Status != nil && span.Status.Message != "" {
		payload["status_message"] = span.Status.Message
	}

	// Extract service info from resource attributes
	if serviceName, ok := allAttrs["service.name"].(string); ok {
		payload["service_name"] = serviceName
	}
	if serviceVersion, ok := allAttrs["service.version"].(string); ok {
		payload["service_version"] = serviceVersion
	}

	// Extract trace-level input/output with priority order:
	// Priority 1: gen_ai.input.messages (OTLP standard for LLM)
	// Priority 2: input.value (OpenInference pattern for generic data)
	// Note: Data stored directly in ClickHouse with ZSTD compression

	// ========== INPUT EXTRACTION ==========
	var inputValue string
	var inputMimeType string

	// Try gen_ai.input.messages first (LLM messages - OTLP standard)
	if messages, ok := allAttrs["gen_ai.input.messages"].([]interface{}); ok {
		if messagesJSON, err := json.Marshal(messages); err == nil {
			inputValue = string(messagesJSON)
			inputMimeType = "application/json" // LLM messages are always JSON
		}
	} else if messages, ok := allAttrs["gen_ai.input.messages"].(string); ok && messages != "" {
		// Already JSON string from SDK
		inputValue = messages
		inputMimeType = "application/json"
	}

	// Fallback to input.value (generic input - OpenInference pattern)
	if inputValue == "" {
		if genericInput, ok := allAttrs["input.value"].(string); ok && genericInput != "" {
			inputValue = genericInput
			// Get MIME type (auto-detect if missing)
			if mimeType, ok := allAttrs["input.mime_type"].(string); ok {
				inputMimeType = validateMimeType(inputValue, mimeType)
			} else {
				inputMimeType = validateMimeType(inputValue, "")
			}
		}
	}

	// Truncate if needed (1MB limit)
	if inputValue != "" {
		truncated := false
		inputValue, truncated = truncateWithIndicator(inputValue, MaxAttributeValueSize)
		if truncated {
			payload["input_truncated"] = true
		}
		payload["input"] = inputValue
		if inputMimeType != "" {
			payload["input_mime_type"] = inputMimeType
		}

		// Extract LLM metadata if input is ChatML format
		if inputMimeType == "application/json" {
			if llmMetadata := extractLLMMetadata(inputValue); len(llmMetadata) > 0 {
				// Add LLM metadata to payload for attributes column
				for key, value := range llmMetadata {
					payload[key] = value
				}
			}
		}
	}

	// ========== OUTPUT EXTRACTION ==========
	var outputValue string
	var outputMimeType string

	// Try gen_ai.output.messages first (LLM messages - OTLP standard)
	if messages, ok := allAttrs["gen_ai.output.messages"].([]interface{}); ok {
		if messagesJSON, err := json.Marshal(messages); err == nil {
			outputValue = string(messagesJSON)
			outputMimeType = "application/json"
		}
	} else if messages, ok := allAttrs["gen_ai.output.messages"].(string); ok && messages != "" {
		outputValue = messages
		outputMimeType = "application/json"
	}

	// Fallback to output.value (generic output - OpenInference pattern)
	if outputValue == "" {
		if genericOutput, ok := allAttrs["output.value"].(string); ok && genericOutput != "" {
			outputValue = genericOutput
			// Get MIME type (auto-detect if missing)
			if mimeType, ok := allAttrs["output.mime_type"].(string); ok {
				outputMimeType = validateMimeType(outputValue, mimeType)
			} else {
				outputMimeType = validateMimeType(outputValue, "")
			}
		}
	}

	// Truncate if needed (1MB limit)
	if outputValue != "" {
		truncated := false
		outputValue, truncated = truncateWithIndicator(outputValue, MaxAttributeValueSize)
		if truncated {
			payload["output_truncated"] = true
		}
		payload["output"] = outputValue
		if outputMimeType != "" {
			payload["output_mime_type"] = outputMimeType
		}
	}

	// Extract user_id (OTEL standard: user.id)
	if userID, ok := allAttrs["user.id"].(string); ok && userID != "" {
		payload["user_id"] = userID
	}

	// Extract session_id (OTEL standard: session.id)
	if sessionID, ok := allAttrs["session.id"].(string); ok && sessionID != "" {
		payload["session_id"] = sessionID
	}

	// Extract environment (default to "default" if not set)
	environment := "default"
	if env, ok := allAttrs["brokle.environment"].(string); ok && env != "" {
		environment = env
	}
	payload["environment"] = environment

	// ========== Build metadata JSON (OTEL resource attributes + Brokle attributes) ==========
	metadata := buildOTELMetadata(resourceAttrs, scopeAttrs, scope)

	// Add Brokle-specific attributes to metadata
	if release, ok := allAttrs["brokle.release"].(string); ok && release != "" {
		metadata["brokle.release"] = release
	}

	if version, ok := allAttrs["brokle.version"].(string); ok && version != "" {
		metadata["brokle.version"] = version
	}

	// Extract user-provided custom metadata (brokle.trace.metadata)
	if userMetadata, ok := allAttrs["brokle.trace.metadata"].(string); ok && userMetadata != "" {
		var metadataMap map[string]interface{}
		if err := json.Unmarshal([]byte(userMetadata), &metadataMap); err == nil {
			// Merge user metadata at top level (matches brokle.release/version pattern)
			for k, v := range metadataMap {
				metadata[k] = v
			}
		}
	}

	// Always store metadata (even if empty, for schema consistency)
	payload["metadata"] = metadata

	// Extract tags (trace-level categorization)
	if tagsAttr, ok := allAttrs["brokle.trace.tags"]; ok {
		var tags []string

		// Handle JSON string format (current SDK sends this)
		if tagsJSON, ok := tagsAttr.(string); ok && tagsJSON != "" {
			if err := json.Unmarshal([]byte(tagsJSON), &tags); err == nil {
				payload["tags"] = tags
			}
		} else if tagsArray, ok := tagsAttr.([]interface{}); ok {
			// Handle array format (OTLP arrayValue - future-proofing)
			tags = make([]string, 0, len(tagsArray))
			for _, tag := range tagsArray {
				if tagStr, ok := tag.(string); ok {
					tags = append(tags, tagStr)
				} else {
					tags = append(tags, fmt.Sprintf("%v", tag))
				}
			}
			payload["tags"] = tags
		}
	}

	// Note: resource_attributes already set above with pure OTEL attributes
	// No need for separate metadata field - consolidated into resource_attributes

	// Create trace event
	event := &brokleEvent{
		EventID:   ulid.New().String(),
		SpanID:    "",      // Traces don't have span_id
		TraceID:   traceID, // OTLP trace_id (32 hex chars)
		EventType: "trace",
		Payload:   payload,
		Timestamp: func() *int64 {
			if startTime != nil {
				ts := startTime.Unix()
				return &ts
			}
			return nil
		}(),
	}

	return event, nil
}

// createSpanEvent creates a span event with complete input/output extraction
// Mirrors createTraceEvent() logic for consistency (spans also have input/output columns)
func (s *OTLPConverterService) createSpanEvent(ctx context.Context, span observability.OTLPSpan, resourceAttrs, scopeAttrs map[string]interface{}, scope *observability.Scope, projectID string) (*brokleEvent, error) {
	// Convert OTLP IDs to hex strings
	traceID, err := convertTraceID(span.TraceID)
	if err != nil {
		return nil, fmt.Errorf("invalid trace_id: %w", err)
	}

	spanID, err := convertSpanID(span.SpanID)
	if err != nil {
		return nil, fmt.Errorf("invalid span_id: %w", err)
	}

	var parentSpanID *string
	if !isRootSpanCheck(span.ParentSpanID) {
		parentID, err := convertSpanID(span.ParentSpanID)
		if err == nil {
			parentSpanID = &parentID
		}
	}

	// Extract span attributes
	spanAttrs := extractAttributesFromKeyValues(span.Attributes)

	// Merge all attributes (resource + scope + span)
	allAttrs := mergeAttributes(resourceAttrs, scopeAttrs, spanAttrs)

	// Convert start/end times
	startTime := convertUnixNano(span.StartTimeUnixNano)
	endTime := convertUnixNano(span.EndTimeUnixNano)

	// Convert OTEL span kind to string
	spanKind := convertSpanKind(span.Kind)

	// Convert OTEL status to string
	statusCode := convertStatusCode(span.Status)

	// Build span payload with new OTEL-native field names
	payload := map[string]interface{}{
		"span_id":        spanID,
		"trace_id":       traceID,
		"parent_span_id": parentSpanID,
		"span_name":      span.Name,
		"span_kind":      spanKind,
		"status_code":    statusCode,
	}

	if startTime != nil {
		payload["start_time"] = startTime.Format(time.RFC3339Nano)
	}

	if endTime != nil {
		payload["end_time"] = endTime.Format(time.RFC3339Nano)
	}
	if span.Status != nil && span.Status.Message != "" {
		payload["status_message"] = span.Status.Message
	}

	// ========== SPAN-LEVEL INPUT/OUTPUT EXTRACTION (Same as Traces) ==========
	// Extract input/output with priority order (identical to createTraceEvent logic)
	var inputValue string
	var inputMimeType string

	// Try gen_ai.input.messages first (LLM messages - OTLP standard)
	if messages, ok := allAttrs["gen_ai.input.messages"].([]interface{}); ok {
		if messagesJSON, err := json.Marshal(messages); err == nil {
			inputValue = string(messagesJSON)
			inputMimeType = "application/json"
		}
	} else if messages, ok := allAttrs["gen_ai.input.messages"].(string); ok && messages != "" {
		inputValue = messages
		inputMimeType = "application/json"
	}

	// Fallback to input.value (generic input - OpenInference pattern)
	if inputValue == "" {
		if genericInput, ok := allAttrs["input.value"].(string); ok && genericInput != "" {
			inputValue = genericInput
			// Get MIME type (auto-detect if missing)
			if mimeType, ok := allAttrs["input.mime_type"].(string); ok {
				inputMimeType = validateMimeType(inputValue, mimeType)
			} else {
				inputMimeType = validateMimeType(inputValue, "")
			}
		}
	}

	// Truncate if needed (1MB limit)
	if inputValue != "" {
		truncated := false
		inputValue, truncated = truncateWithIndicator(inputValue, MaxAttributeValueSize)
		if truncated {
			payload["input_truncated"] = true
		}
		payload["input"] = inputValue
		if inputMimeType != "" {
			payload["input_mime_type"] = inputMimeType
		}

		// Extract LLM metadata if input is ChatML format
		if inputMimeType == "application/json" {
			if llmMetadata := extractLLMMetadata(inputValue); len(llmMetadata) > 0 {
				// Add LLM metadata to payload for attributes column
				for key, value := range llmMetadata {
					payload[key] = value
				}
			}
		}
	}

	// ========== OUTPUT EXTRACTION ==========
	var outputValue string
	var outputMimeType string

	// Try gen_ai.output.messages first (LLM messages - OTLP standard)
	if messages, ok := allAttrs["gen_ai.output.messages"].([]interface{}); ok {
		if messagesJSON, err := json.Marshal(messages); err == nil {
			outputValue = string(messagesJSON)
			outputMimeType = "application/json"
		}
	} else if messages, ok := allAttrs["gen_ai.output.messages"].(string); ok && messages != "" {
		outputValue = messages
		outputMimeType = "application/json"
	}

	// Fallback to output.value (generic output - OpenInference pattern)
	if outputValue == "" {
		if genericOutput, ok := allAttrs["output.value"].(string); ok && genericOutput != "" {
			outputValue = genericOutput
			// Get MIME type (auto-detect if missing)
			if mimeType, ok := allAttrs["output.mime_type"].(string); ok {
				outputMimeType = validateMimeType(outputValue, mimeType)
			} else {
				outputMimeType = validateMimeType(outputValue, "")
			}
		}
	}

	// Truncate if needed (1MB limit)
	if outputValue != "" {
		truncated := false
		outputValue, truncated = truncateWithIndicator(outputValue, MaxAttributeValueSize)
		if truncated {
			payload["output_truncated"] = true
		}
		payload["output"] = outputValue
		if outputMimeType != "" {
			payload["output_mime_type"] = outputMimeType
		}
	}

	// ========== Continue with existing Gen AI extraction ==========
	// Extract Gen AI semantic conventions (model, provider, tokens, etc.)
	extractGenAIFields(allAttrs, payload)

	// ========== NEW: PROVIDER PRICING - Calculate Provider Costs + Store Maps ==========
	s.calculateProviderCostsAtIngestion(ctx, allAttrs, payload, projectID)

	// Store all span attributes in payload (will become attributes JSON in ClickHouse)
	// All OTEL + Brokle attributes already in spanAttrs from line 516
	payload["span_attributes"] = spanAttrs

	// ========== OTEL 1.38+ INSTRUMENTATION SCOPE ==========
	// Extract scope information from scopeSpan
	if scope != nil {
		if scope.Name != "" {
			payload["scope_name"] = scope.Name
		}
		if scope.Version != "" {
			payload["scope_version"] = scope.Version
		}
		// Extract scope attributes if present
		if len(scopeAttrs) > 0 {
			payload["scope_attributes"] = scopeAttrs
		}
	}

	// ========== W3C TRACE CONTEXT: TraceState ==========
	// Extract TraceState from span (vendor-specific tracing data)
	// Note: TraceState not directly in OTLP span, comes from SpanContext
	// For now, extract from attributes if SDK sets it
	if traceState, ok := allAttrs["trace_state"].(string); ok && traceState != "" {
		payload["trace_state"] = traceState
	}

	// Extract OTLP Events (timeline annotations within span)
	if len(span.Events) > 0 {
		eventsTimestamp := make([]string, len(span.Events))
		eventsName := make([]string, len(span.Events))
		eventsAttributes := make([]map[string]interface{}, len(span.Events)) // Map type (OTEL standard)
		eventsDroppedCount := make([]uint32, len(span.Events))

		for i, event := range span.Events {
			// Convert event timestamp (nanosecond precision)
			if eventTime := convertUnixNano(event.TimeUnixNano); eventTime != nil {
				eventsTimestamp[i] = eventTime.Format(time.RFC3339Nano)
			}

			// Extract event name
			eventsName[i] = event.Name

			// Extract event attributes as Map (not JSON string)
			eventAttrs := extractAttributesFromKeyValues(event.Attributes)
			eventsAttributes[i] = eventAttrs // Direct map, ClickHouse converts to Map(String,String)

			// Track dropped attributes
			eventsDroppedCount[i] = event.DroppedAttributesCount
		}

		payload["events_timestamp"] = eventsTimestamp
		payload["events_name"] = eventsName
		payload["events_attributes"] = eventsAttributes // Array of maps
		payload["events_dropped_attributes_count"] = eventsDroppedCount
	}

	// Extract OTLP Links (cross-trace references)
	if len(span.Links) > 0 {
		linksTraceID := make([]string, len(span.Links))
		linksSpanID := make([]string, len(span.Links))
		linksTraceState := make([]string, len(span.Links))                 // W3C TraceState for links
		linksAttributes := make([]map[string]interface{}, len(span.Links)) // Map type (OTEL standard)
		linksDroppedCount := make([]uint32, len(span.Links))

		for i, link := range span.Links {
			// Convert linked trace ID
			if traceID, err := convertTraceID(link.TraceID); err == nil {
				linksTraceID[i] = traceID
			}

			// Convert linked span ID
			if spanID, err := convertSpanID(link.SpanID); err == nil {
				linksSpanID[i] = spanID
			}

			// Extract TraceState for this link (W3C Trace Context)
			if link.TraceState != nil {
				if ts, ok := link.TraceState.(string); ok {
					linksTraceState[i] = ts
				}
			}

			// Extract link attributes as Map (not JSON string)
			linkAttrs := extractAttributesFromKeyValues(link.Attributes)
			linksAttributes[i] = linkAttrs // Direct map, ClickHouse converts to Map(String,String)

			// Track dropped attributes
			linksDroppedCount[i] = link.DroppedAttributesCount
		}

		payload["links_trace_id"] = linksTraceID
		payload["links_span_id"] = linksSpanID
		payload["links_trace_state"] = linksTraceState // NEW: W3C TraceState array
		payload["links_attributes"] = linksAttributes  // Array of maps
		payload["links_dropped_attributes_count"] = linksDroppedCount
	}

	// ========== Build attributes JSON (all OTEL + Brokle span attributes) ==========
	// Get the attributes map from payload (extractBrokleFields may have modified it)
	attributes, ok := payload["span_attributes"].(map[string]interface{})
	if !ok {
		attributes = spanAttrs
	}
	payload["attributes"] = attributes
	delete(payload, "span_attributes") // Remove old key

	// ========== Build metadata JSON (OTEL resource attributes + scope) ==========
	spanMetadata := buildOTELMetadata(resourceAttrs, scopeAttrs, scope)
	payload["metadata"] = spanMetadata

	// Remove old scope fields (now in metadata)
	delete(payload, "scope_name")
	delete(payload, "scope_version")
	delete(payload, "scope_attributes")

	// Create span event
	event := &brokleEvent{
		EventID:   ulid.New().String(),
		SpanID:    spanID,
		TraceID:   traceID,
		EventType: "span",
		Payload:   payload,
		Timestamp: func() *int64 {
			if startTime != nil {
				ts := startTime.Unix()
				return &ts
			}
			return nil
		}(),
	}

	return event, nil
}

// Helper functions for OTLP conversion

// convertTraceID converts OTLP trace_id to 32-char hex string
func convertTraceID(traceID interface{}) (string, error) {
	switch v := traceID.(type) {
	case string:
		// Already hex string
		if len(v) == 32 {
			return v, nil
		}
		return "", fmt.Errorf("invalid trace_id length: %d (expected 32)", len(v))
	case map[string]interface{}:
		// Handle {data: Buffer} format
		if data, ok := v["data"].([]interface{}); ok {
			return bytesToHex(data), nil
		}
	case []byte:
		return hex.EncodeToString(v), nil
	}
	return "", fmt.Errorf("unsupported trace_id type: %T", traceID)
}

// convertSpanID converts OTLP span_id to 16-char hex string
func convertSpanID(spanID interface{}) (string, error) {
	switch v := spanID.(type) {
	case string:
		// Already hex string
		if len(v) == 16 {
			return v, nil
		}
		return "", fmt.Errorf("invalid span_id length: %d (expected 16)", len(v))
	case map[string]interface{}:
		// Handle {data: Buffer} format
		if data, ok := v["data"].([]interface{}); ok {
			return bytesToHex(data), nil
		}
	case []byte:
		return hex.EncodeToString(v), nil
	}
	return "", fmt.Errorf("unsupported span_id type: %T", spanID)
}

// convertUnixNano converts OTLP nanosecond timestamp to time.Time
func convertUnixNano(ts interface{}) *time.Time {
	if ts == nil {
		return nil
	}

	var nanos int64
	switch v := ts.(type) {
	case int64:
		nanos = v
	case float64:
		nanos = int64(v)
	case map[string]interface{}:
		// Handle {low, high} format from protobuf
		low, lowOk := v["low"].(float64)
		high, highOk := v["high"].(float64)
		if !lowOk || !highOk {
			return nil
		}
		nanos = int64(high)*4294967296 + int64(low)
	default:
		return nil
	}

	if nanos == 0 {
		return nil
	}

	t := time.Unix(0, nanos)
	return &t
}

// convertSpanKind converts OTLP span kind enum to UInt8 for ClickHouse
func convertSpanKind(kind int) uint8 {
	switch kind {
	case 0:
		return observability.SpanKindUnspecified // 0
	case 1:
		return observability.SpanKindInternal // 1
	case 2:
		return observability.SpanKindServer // 2
	case 3:
		return observability.SpanKindClient // 3
	case 4:
		return observability.SpanKindProducer // 4
	case 5:
		return observability.SpanKindConsumer // 5
	default:
		return observability.SpanKindInternal // default to INTERNAL
	}
}

// convertStatusCode converts OTLP status code to UInt8 for ClickHouse
func convertStatusCode(status *observability.Status) uint8 {
	if status == nil {
		return observability.StatusCodeUnset // 0
	}
	switch status.Code {
	case 0:
		return observability.StatusCodeUnset // 0
	case 1:
		return observability.StatusCodeOK // 1
	case 2:
		return observability.StatusCodeError // 2
	default:
		return observability.StatusCodeUnset // default to UNSET
	}
}

// extractAttributes extracts attributes from resource or scope
func extractAttributes(obj interface{}) map[string]interface{} {
	attrs := make(map[string]interface{})

	switch v := obj.(type) {
	case *observability.Resource:
		if v != nil {
			return extractAttributesFromKeyValues(v.Attributes)
		}
	case *observability.Scope:
		if v != nil {
			return extractAttributesFromKeyValues(v.Attributes)
		}
	}

	return attrs
}

// extractAttributesFromKeyValues converts OTLP KeyValue array to map
func extractAttributesFromKeyValues(kvs []observability.KeyValue) map[string]interface{} {
	attrs := make(map[string]interface{})

	for _, kv := range kvs {
		// Extract value from OTLP value union type
		value := extractValue(kv.Value)
		if value != nil {
			attrs[kv.Key] = value
		}
	}

	return attrs
}

// extractValue extracts the actual value from OTLP value union
func extractValue(v interface{}) interface{} {
	if v == nil {
		return nil
	}

	switch val := v.(type) {
	case map[string]interface{}:
		// Handle {stringValue: "...", intValue: 123, ...} format
		if sv, ok := val["stringValue"].(string); ok {
			return sv
		}
		if iv, ok := val["intValue"].(float64); ok {
			return int64(iv)
		}
		if bv, ok := val["boolValue"].(bool); ok {
			return bv
		}
		if dv, ok := val["doubleValue"].(float64); ok {
			return dv
		}
		if av, ok := val["arrayValue"].(map[string]interface{}); ok {
			// Handle array values
			if values, ok := av["values"].([]interface{}); ok {
				result := make([]interface{}, len(values))
				for i, item := range values {
					result[i] = extractValue(item)
				}
				return result
			}
		}
		return val
	case string, int, int64, float64, bool:
		return val
	}

	return v
}

// mergeAttributes merges resource, scope, and span attributes (span takes precedence)
func mergeAttributes(resource, scope, span map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})

	// Add resource attributes
	for k, v := range resource {
		merged[k] = v
	}

	// Add scope attributes (override resource)
	for k, v := range scope {
		merged[k] = v
	}

	// Add span attributes (override all)
	for k, v := range span {
		merged[k] = v
	}

	return merged
}

// marshalAttributes converts attributes map to JSON string
func marshalAttributes(attrs map[string]interface{}) string {
	if len(attrs) == 0 {
		return "{}"
	}

	jsonBytes, err := json.Marshal(attrs)
	if err != nil {
		return "{}"
	}

	return string(jsonBytes)
}

// buildOTELMetadata builds OTEL metadata structure
// Contains resource attributes and instrumentation scope information
func buildOTELMetadata(resourceAttrs, scopeAttrs map[string]interface{}, scope *observability.Scope) map[string]interface{} {
	metadata := make(map[string]interface{})

	// Filter out brokle.* attributes from resource attributes
	// These are SDK-internal attributes, not standard OTEL resource attributes
	filteredResourceAttrs := make(map[string]interface{})
	for k, v := range resourceAttrs {
		if !strings.HasPrefix(k, "brokle.") {
			filteredResourceAttrs[k] = v
		}
	}

	// Add filtered resource attributes (pure OTEL only)
	metadata["resourceAttributes"] = filteredResourceAttrs

	// Add instrumentation scope
	scopeInfo := make(map[string]interface{})
	if scope != nil {
		scopeInfo["name"] = scope.Name
		scopeInfo["version"] = scope.Version
		if len(scopeAttrs) > 0 {
			scopeInfo["attributes"] = scopeAttrs
		}
	}
	metadata["scope"] = scopeInfo

	return metadata
}

// extractGenAIFields extracts Gen AI semantic conventions from attributes
func extractGenAIFields(attrs map[string]interface{}, payload map[string]interface{}) {
	// ========== OTEL GenAI 1.28+ Attributes ==========
	// Strategy: Use existing schema columns + attributes JSON
	// - provider column (indexed, fast queries)
	// - model_name column (indexed, fast queries)
	// - input/output columns (messages, ZSTD compressed)
	// - model_parameters column (JSON for request params)
	// - usage_details Map (token counts)
	// - attributes JSON column (ALL OTEL attributes, ZSTD compressed)

	// ========== Provider (existing column) ==========
	if provider, ok := attrs["gen_ai.provider.name"].(string); ok {
		payload["provider"] = provider
	}

	// ========== Model (existing column) ==========
	// Prefer response model (authoritative) over request model
	if responseModel, ok := attrs["gen_ai.response.model"].(string); ok {
		payload["model_name"] = responseModel
	} else if requestModel, ok := attrs["gen_ai.request.model"].(string); ok {
		payload["model_name"] = requestModel
	}

	// ========== Messages → existing input/output columns ==========
	// Input messages (excluding system)
	if messages, ok := attrs["gen_ai.input.messages"].(string); ok {
		payload["input"] = messages
	} else if messages, ok := attrs["gen_ai.input.messages"].([]interface{}); ok {
		if messagesJSON, err := json.Marshal(messages); err == nil {
			payload["input"] = string(messagesJSON)
		}
	}

	// Output messages
	if messages, ok := attrs["gen_ai.output.messages"].(string); ok {
		payload["output"] = messages
	} else if messages, ok := attrs["gen_ai.output.messages"].([]interface{}); ok {
		if messagesJSON, err := json.Marshal(messages); err == nil {
			payload["output"] = string(messagesJSON)
		}
	}

	// ========== Model Parameters → existing model_parameters JSON ==========
	modelParams := make(map[string]interface{})

	// Extract standard OTEL GenAI request parameters
	if temp, ok := attrs["gen_ai.request.temperature"].(float64); ok {
		modelParams["temperature"] = temp
	}
	if maxTokens, ok := attrs["gen_ai.request.max_tokens"].(float64); ok {
		modelParams["max_tokens"] = int(maxTokens)
	} else if maxTokens, ok := attrs["gen_ai.request.max_tokens"].(int64); ok {
		modelParams["max_tokens"] = int(maxTokens)
	}
	if topP, ok := attrs["gen_ai.request.top_p"].(float64); ok {
		modelParams["top_p"] = topP
	}
	if topK, ok := attrs["gen_ai.request.top_k"].(float64); ok {
		modelParams["top_k"] = int(topK)
	} else if topK, ok := attrs["gen_ai.request.top_k"].(int64); ok {
		modelParams["top_k"] = int(topK)
	}
	if freqPenalty, ok := attrs["gen_ai.request.frequency_penalty"].(float64); ok {
		modelParams["frequency_penalty"] = freqPenalty
	}
	if presPenalty, ok := attrs["gen_ai.request.presence_penalty"].(float64); ok {
		modelParams["presence_penalty"] = presPenalty
	}

	// Store model parameters as JSON
	if len(modelParams) > 0 {
		if paramsJSON, err := json.Marshal(modelParams); err == nil {
			payload["model_parameters"] = string(paramsJSON)
		}
	}

	// ========== Usage Tokens → existing usage_details Map ==========
	usageDetails := make(map[string]uint64)

	// OTEL 1.28+ standard attributes (input_tokens, output_tokens)
	if inputTokens, ok := attrs["gen_ai.usage.input_tokens"].(float64); ok {
		usageDetails["input"] = uint64(inputTokens)
	} else if inputTokens, ok := attrs["gen_ai.usage.input_tokens"].(int64); ok {
		usageDetails["input"] = uint64(inputTokens)
	}

	if outputTokens, ok := attrs["gen_ai.usage.output_tokens"].(float64); ok {
		usageDetails["output"] = uint64(outputTokens)
	} else if outputTokens, ok := attrs["gen_ai.usage.output_tokens"].(int64); ok {
		usageDetails["output"] = uint64(outputTokens)
	}

	// Brokle convenience attribute (brokle.usage.total_tokens)
	if totalTokens, ok := attrs["brokle.usage.total_tokens"].(float64); ok {
		usageDetails["total"] = uint64(totalTokens)
	} else if totalTokens, ok := attrs["brokle.usage.total_tokens"].(int64); ok {
		usageDetails["total"] = uint64(totalTokens)
	}

	if len(usageDetails) > 0 {
		payload["usage_details"] = usageDetails
	}

	// Note: ALL OTEL GenAI attributes (including operation, response_id, finish_reasons, etc.)
	// are already stored in the "attributes" JSON column via marshalAttributes(allAttrs).
	// This includes:
	// - gen_ai.operation.name
	// - gen_ai.response.id
	// - gen_ai.response.finish_reasons
	// - gen_ai.system_instructions
	// - All provider-specific attributes (openai.*, anthropic.*, etc.)
	// No need to duplicate them in separate columns!
}

// extractBrokleFields function removed - no longer needed
// Costs now stored in cost_details Map
// All span attributes directly stored in payload["span_attributes"]
// Zero users, clean architecture - no legacy compatibility needed

// Helper function to convert byte array to hex
func bytesToHex(data []interface{}) string {
	bytes := make([]byte, len(data))
	for i, v := range data {
		if f, ok := v.(float64); ok {
			bytes[i] = byte(f)
		}
	}
	return hex.EncodeToString(bytes)
}

// convertToDomainEvents converts internal events to domain events
func convertToDomainEvents(events []*brokleEvent) []*observability.TelemetryEventRequest {
	result := make([]*observability.TelemetryEventRequest, 0, len(events))
	for _, e := range events {
		eventID, err := ulid.Parse(e.EventID)
		if err != nil {
			// Skip invalid event IDs
			continue
		}
		result = append(result, &observability.TelemetryEventRequest{
			EventID:   eventID,
			SpanID:    e.SpanID,  // OTLP span_id (populated from brokleEvent)
			TraceID:   e.TraceID, // OTLP trace_id (populated from brokleEvent)
			EventType: observability.TelemetryEventType(e.EventType),
			Payload:   e.Payload,
			Timestamp: func() *time.Time {
				if e.Timestamp != nil {
					t := time.Unix(*e.Timestamp, 0)
					return &t
				}
				return nil
			}(),
		})
	}
	return result
}

// calculateProviderCostsAtIngestion calculates user spending with AI providers at ingestion time
// Purpose: Cost visibility ("You spent $X with OpenAI") - NOT for billing users
// Creates: usage_details, cost_details, pricing_snapshot maps for ClickHouse
// Pricing snapshot = audit trail for "What was OpenAI's rate on this date?"
func (s *OTLPConverterService) calculateProviderCostsAtIngestion(
	ctx context.Context,
	attrs map[string]interface{},
	payload map[string]interface{},
	projectID string,
) {
	// Extract model name
	modelName := extractStringFromInterface(attrs["gen_ai.request.model"])
	if modelName == "" {
		return
	}

	// ========== STEP 1: Extract Usage from OTEL Attributes ==========
	usage := make(map[string]uint64)

	// Standard tokens (OTEL semantic conventions)
	if val := extractUint64FromInterface(attrs["gen_ai.usage.input_tokens"]); val > 0 {
		usage["input"] = val
	}
	if val := extractUint64FromInterface(attrs["gen_ai.usage.output_tokens"]); val > 0 {
		usage["output"] = val
	}

	// Cache tokens (OTEL proposed convention - not yet standard)
	if val := extractUint64FromInterface(attrs["gen_ai.usage.input_tokens.cache_read"]); val > 0 {
		usage["cache_read_input_tokens"] = val
	}
	if val := extractUint64FromInterface(attrs["gen_ai.usage.input_tokens.cache_creation"]); val > 0 {
		usage["cache_creation_input_tokens"] = val
	}

	// Audio tokens (multi-modal)
	if val := extractUint64FromInterface(attrs["gen_ai.usage.input_audio_tokens"]); val > 0 {
		usage["audio_input"] = val
	}
	if val := extractUint64FromInterface(attrs["gen_ai.usage.output_audio_tokens"]); val > 0 {
		usage["audio_output"] = val
	}

	// Reasoning tokens (OpenAI o1 models)
	if val := extractUint64FromInterface(attrs["gen_ai.usage.reasoning_tokens"]); val > 0 {
		usage["reasoning_tokens"] = val
	}

	// Image tokens (vision models)
	if val := extractUint64FromInterface(attrs["gen_ai.usage.image_tokens"]); val > 0 {
		usage["image_tokens"] = val
	}

	// Video tokens (future)
	if val := extractUint64FromInterface(attrs["gen_ai.usage.video_tokens"]); val > 0 {
		usage["video_tokens"] = val
	}

	// Calculate total (only sum base + additive tokens, exclude cache subsets)
	var total uint64

	// Base text tokens (always additive)
	if input, ok := usage["input"]; ok {
		total += input
	}
	if output, ok := usage["output"]; ok {
		total += output
	}

	// Reasoning tokens (additive for OpenAI o1 models)
	if reasoning, ok := usage["reasoning_tokens"]; ok {
		total += reasoning
	}

	// Audio tokens (separate pricing, additive)
	if audioIn, ok := usage["audio_input"]; ok {
		total += audioIn
	}
	if audioOut, ok := usage["audio_output"]; ok {
		total += audioOut
	}

	// Multimodal tokens (separate pricing, additive)
	if image, ok := usage["image_tokens"]; ok {
		total += image
	}
	if video, ok := usage["video_tokens"]; ok {
		total += video
	}

	// Store total (excludes cache subsets which are already counted in input)
	if total > 0 {
		usage["total"] = total
	}

	// Note: Cache tokens explicitly EXCLUDED from total:
	// - cache_read_input_tokens (subset of input, reused from previous calls)
	// - cache_creation_input_tokens (subset of input, written to cache)
	// These are tracked separately for cost breakdown but not added to total count

	if len(usage) == 0 {
		return
	}

	// ========== STEP 2: Lookup Pricing from PostgreSQL ==========
	projectIDPtr := (*ulid.ULID)(nil)
	if projectID != "" {
		if pid, err := ulid.Parse(projectID); err == nil {
			projectIDPtr = &pid
		}
	}

	providerPricing, err := s.providerPricingService.GetProviderPricingSnapshot(ctx, projectIDPtr, modelName, time.Now())
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"model":      modelName,
			"project_id": projectID,
			"error":      err,
		}).Warn("Failed to get provider pricing - continuing without cost data")

		// Store usage even without pricing
		payload["usage_details"] = usage
		return
	}

	// ========== STEP 3: Calculate Provider Costs ==========
	providerCost := s.providerPricingService.CalculateProviderCost(usage, providerPricing)

	// ========== STEP 4: Build Provider Pricing Snapshot for Audit Trail ==========
	providerPricingSnapshot := make(map[string]decimal.Decimal)
	for usageType, price := range providerPricing.Pricing {
		// Store with descriptive key
		key := fmt.Sprintf("%s_price_per_million", usageType)
		providerPricingSnapshot[key] = price
	}

	// ========== STEP 5: Store in Payload (for ClickHouse) ==========
	payload["usage_details"] = usage
	payload["cost_details"] = providerCost               // Store as map[string]decimal.Decimal (matches ClickHouse schema)
	payload["pricing_snapshot"] = providerPricingSnapshot // Store as map[string]decimal.Decimal (matches ClickHouse schema)

	// Extract total_cost for fast aggregation
	if totalCost, ok := providerCost["total"]; ok {
		payload["total_cost"] = totalCost // Store as decimal.Decimal (matches ClickHouse Decimal(18,12))
	}

	s.logger.WithFields(logrus.Fields{
		"model":                 modelName,
		"usage_types":           len(usage),
		"total_tokens":          total,
		"provider_cost_usd":     providerCost["total"],
		"provider_pricing_date": providerPricing.SnapshotTime,
	}).Debug("Provider costs calculated successfully")
}

// extractStringFromInterface safely extracts a string from interface{}
func extractStringFromInterface(val interface{}) string {
	if str, ok := val.(string); ok {
		return str
	}
	return ""
}

// extractUint64FromInterface safely extracts a uint64 from interface{}
func extractUint64FromInterface(val interface{}) uint64 {
	switch v := val.(type) {
	case float64:
		return uint64(v)
	case int64:
		return uint64(v)
	case int32:
		return uint64(v)
	case int:
		return uint64(v)
	case uint64:
		return v
	case uint32:
		return uint64(v)
	case uint:
		return uint64(v)
	default:
		return 0
	}
}

// extractBoolFromInterface safely extracts a boolean from interface{}
func extractBoolFromInterface(val interface{}) bool {
	if b, ok := val.(bool); ok {
		return b
	}
	return false
}
