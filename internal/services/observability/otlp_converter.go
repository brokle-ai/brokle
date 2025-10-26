package observability

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"brokle/internal/core/domain/observability"
	"brokle/pkg/ulid"
)

// OTLPConverterService handles conversion of OTLP traces to Brokle telemetry events
type OTLPConverterService struct {
	logger *logrus.Logger
}

// NewOTLPConverterService creates a new OTLP converter service
func NewOTLPConverterService(logger *logrus.Logger) *OTLPConverterService {
	return &OTLPConverterService{
		logger: logger,
	}
}

// brokleEvent represents an internal converted event (before domain conversion)
type brokleEvent struct {
	EventID   string                 `json:"event_id"`
	EventType string                 `json:"event_type"` // "trace", "observation"
	Payload   map[string]interface{} `json:"payload"`
	Timestamp *int64                 `json:"timestamp,omitempty"`
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
// For root spans, creates both trace AND observation events
func (s *OTLPConverterService) ConvertOTLPToBrokleEvents(otlpReq *observability.OTLPRequest) ([]*observability.TelemetryEventRequest, error) {
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
					traceEvent, err := s.createTraceEvent(span, resourceAttrs, scopeAttrs, traceID)
					if err != nil {
						return nil, fmt.Errorf("failed to create trace event: %w", err)
					}
					internalEvents = append(internalEvents, traceEvent)
					tracesCreated[traceID] = true
				}

				// Convert OTLP span to Brokle observation event
				obsEvent, err := s.convertSpanToEvent(span, resourceAttrs, scopeAttrs)
				if err != nil {
					return nil, fmt.Errorf("failed to convert span: %w", err)
				}
				internalEvents = append(internalEvents, obsEvent)
			}
		}
	}

	// Convert internal events to domain events
	return convertToDomainEvents(internalEvents), nil
}

// createTraceEvent creates a trace event from a root span
func (s *OTLPConverterService) createTraceEvent(span observability.Span, resourceAttrs, scopeAttrs map[string]interface{}, traceID string) (*brokleEvent, error) {
	// Extract span attributes
	spanAttrs := extractAttributesFromKeyValues(span.Attributes)

	// Merge all attributes
	allAttrs := mergeAttributes(resourceAttrs, scopeAttrs, spanAttrs)

	// Convert times
	startTime := convertUnixNano(span.StartTimeUnixNano)
	endTime := convertUnixNano(span.EndTimeUnixNano)

	// Convert status
	statusCode := convertStatusCode(span.Status)

	// Build trace payload
	payload := map[string]interface{}{
		"id":          traceID,
		"name":        span.Name,
		"start_time":  startTime.Format(time.RFC3339Nano),
		"status_code": statusCode,
		"attributes":  marshalAttributes(allAttrs),
	}

	if endTime != nil {
		payload["end_time"] = endTime.Format(time.RFC3339Nano)
		duration := uint32(endTime.Sub(*startTime).Milliseconds())
		payload["duration_ms"] = duration
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

	// Extract environment
	if env, ok := allAttrs["deployment.environment"].(string); ok {
		payload["environment"] = env
	}

	// Extract trace-level input/output using official OTel format
	// Supports: gen_ai.input.messages and gen_ai.output.messages (OTel 1.28+)
	// Note: Data stored directly in ClickHouse with ZSTD compression
	if messages, ok := allAttrs["gen_ai.input.messages"].([]interface{}); ok {
		if messagesJSON, err := json.Marshal(messages); err == nil {
			payload["input"] = string(messagesJSON)
		}
	}

	if messages, ok := allAttrs["gen_ai.output.messages"].([]interface{}); ok {
		if messagesJSON, err := json.Marshal(messages); err == nil {
			payload["output"] = string(messagesJSON)
		}
	}

	// Create trace event
	return &brokleEvent{
		EventID:   ulid.New().String(),
		EventType: "trace",
		Payload:   payload,
		Timestamp: func() *int64 {
			ts := startTime.Unix()
			return &ts
		}(),
	}, nil
}

// convertSpanToEvent converts a single OTLP span to a Brokle telemetry event
func (s *OTLPConverterService) convertSpanToEvent(span observability.Span, resourceAttrs, scopeAttrs map[string]interface{}) (*brokleEvent, error) {
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

	// Extract Brokle type from attributes (default: "span")
	brokleType := "span"
	if t, ok := allAttrs["brokle.type"].(string); ok {
		brokleType = t
	}

	// Convert start/end times
	startTime := convertUnixNano(span.StartTimeUnixNano)
	endTime := convertUnixNano(span.EndTimeUnixNano)

	// Convert OTEL span kind to string
	spanKind := convertSpanKind(span.Kind)

	// Convert OTEL status to string
	statusCode := convertStatusCode(span.Status)

	// Build observation payload
	payload := map[string]interface{}{
		"id":                    spanID,
		"trace_id":              traceID,
		"parent_observation_id": parentSpanID,
		"name":                  span.Name,
		"span_kind":             spanKind,
		"type":                  brokleType,
		"start_time":            startTime.Format(time.RFC3339Nano),
		"status_code":           statusCode,
		"attributes":            marshalAttributes(allAttrs),
	}

	if endTime != nil {
		payload["end_time"] = endTime.Format(time.RFC3339Nano)
	}
	if span.Status != nil && span.Status.Message != "" {
		payload["status_message"] = span.Status.Message
	}

	// Extract Gen AI semantic conventions from attributes
	extractGenAIFields(allAttrs, payload)

	// Extract Brokle extensions from attributes
	extractBrokleFields(allAttrs, payload)

	// Create observation event
	event := &brokleEvent{
		EventID:   ulid.New().String(),
		EventType: "observation",
		Payload:   payload,
		Timestamp: func() *int64 {
			ts := startTime.Unix()
			return &ts
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
		low, _ := v["low"].(float64)
		high, _ := v["high"].(float64)
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

// convertSpanKind converts OTLP span kind enum to string
func convertSpanKind(kind int) string {
	switch kind {
	case 0:
		return "UNSPECIFIED"
	case 1:
		return "INTERNAL"
	case 2:
		return "SERVER"
	case 3:
		return "CLIENT"
	case 4:
		return "PRODUCER"
	case 5:
		return "CONSUMER"
	default:
		return "INTERNAL"
	}
}

// convertStatusCode converts OTLP status code to string
func convertStatusCode(status *observability.Status) string {
	if status == nil {
		return "UNSET"
	}
	switch status.Code {
	case 0:
		return "UNSET"
	case 1:
		return "OK"
	case 2:
		return "ERROR"
	default:
		return "UNSET"
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

// extractBrokleFields extracts Brokle extension fields from attributes
func extractBrokleFields(attrs map[string]interface{}, payload map[string]interface{}) {
	// Initialize Maps
	metadata := make(map[string]string)
	costDetails := make(map[string]float64)

	// Routing → metadata Map
	if provider, ok := attrs["brokle.routing.provider"].(string); ok {
		metadata["brokle.routing.provider"] = provider
	}
	if strategy, ok := attrs["brokle.routing.strategy"].(string); ok {
		metadata["brokle.routing.strategy"] = strategy
	}

	// Cost → cost_details Map
	if costTotal, ok := attrs["brokle.cost.total"].(float64); ok {
		costDetails["total"] = costTotal
		payload["total_cost"] = costTotal // Denormalized for fast queries
	}
	if costInput, ok := attrs["brokle.cost.input"].(float64); ok {
		costDetails["input"] = costInput
	}
	if costOutput, ok := attrs["brokle.cost.output"].(float64); ok {
		costDetails["output"] = costOutput
	}

	// Cache → metadata Map
	if cacheHit, ok := attrs["brokle.cache.hit"].(bool); ok {
		if cacheHit {
			metadata["brokle.cache.hit"] = "true"
		} else {
			metadata["brokle.cache.hit"] = "false"
		}
	}
	if similarity, ok := attrs["brokle.cache.similarity"].(float64); ok {
		metadata["brokle.cache.similarity"] = fmt.Sprintf("%.2f", similarity)
	}

	// Governance → metadata Map
	if passed, ok := attrs["brokle.governance.passed"].(bool); ok {
		if passed {
			metadata["brokle.governance.passed"] = "true"
		} else {
			metadata["brokle.governance.passed"] = "false"
		}
	}
	if policy, ok := attrs["brokle.governance.policy"].(string); ok {
		metadata["brokle.governance.policy"] = policy
	}

	// Prompt management (keep as dedicated fields)
	if promptID, ok := attrs["brokle.prompt.id"].(string); ok {
		payload["prompt_id"] = promptID
	}
	if promptName, ok := attrs["brokle.prompt.name"].(string); ok {
		payload["prompt_name"] = promptName
	}
	if promptVersion, ok := attrs["brokle.prompt.version"].(int64); ok {
		payload["prompt_version"] = uint16(promptVersion)
	}

	// Add Maps to payload
	if len(metadata) > 0 {
		payload["metadata"] = metadata
	}
	if len(costDetails) > 0 {
		payload["cost_details"] = costDetails
	}
}

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
	result := make([]*observability.TelemetryEventRequest, len(events))
	for i, e := range events {
		eventID, _ := ulid.Parse(e.EventID)
		result[i] = &observability.TelemetryEventRequest{
			EventID:   eventID,
			EventType: observability.TelemetryEventType(e.EventType),
			Payload:   e.Payload,
			Timestamp: func() *time.Time {
				if e.Timestamp != nil {
					t := time.Unix(*e.Timestamp, 0)
					return &t
				}
				return nil
			}(),
		}
	}
	return result
}
