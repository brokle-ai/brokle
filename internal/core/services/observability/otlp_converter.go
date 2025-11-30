package observability

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"

	"brokle/internal/config"
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
	config                 *config.ObservabilityConfig
}

// NewOTLPConverterService creates a new OTLP converter service
func NewOTLPConverterService(
	logger *logrus.Logger,
	providerPricingService analytics.ProviderPricingService,
	observabilityConfig *config.ObservabilityConfig,
) *OTLPConverterService {
	return &OTLPConverterService{
		logger:                 logger,
		providerPricingService: providerPricingService,
		config:                 observabilityConfig,
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

// isRootSpanCheck determines if a span is a root span by checking if parent ID is nil, empty, or zero.
// Handles: nil, empty string, "0000000000000000", zero bytes array, and {data: Buffer} format.
func isRootSpanCheck(parentSpanID interface{}) bool {
	if parentSpanID == nil {
		return true
	}

	if str, ok := parentSpanID.(string); ok {
		if str == "" || str == "0000000000000000" {
			return true
		}
	}

	if mapVal, ok := parentSpanID.(map[string]interface{}); ok {
		if data, ok := mapVal["data"].([]interface{}); ok {
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

// ConvertOTLPToBrokleEvents converts OTLP resourceSpans to Brokle span events.
// Traces are derived asynchronously from root spans (parent_span_id IS NULL).
func (s *OTLPConverterService) ConvertOTLPToBrokleEvents(ctx context.Context, otlpReq *observability.OTLPRequest, projectID string) ([]*observability.TelemetryEventRequest, error) {
	var internalEvents []*brokleEvent

	for _, resourceSpan := range otlpReq.ResourceSpans {
		resourceAttrs := extractAttributes(resourceSpan.Resource)

		for _, scopeSpan := range resourceSpan.ScopeSpans {
			scopeAttrs := extractAttributes(scopeSpan.Scope)

			for _, span := range scopeSpan.Spans {
				obsEvent, err := s.createSpanEvent(ctx, span, resourceAttrs, scopeAttrs, resourceSpan.Resource, scopeSpan.Scope, projectID)
				if err != nil {
					return nil, fmt.Errorf("failed to create span event: %w", err)
				}
				internalEvents = append(internalEvents, obsEvent)
			}
		}
	}

	return convertToDomainEvents(internalEvents), nil
}

func truncateWithIndicator(value string, maxSize int) (string, bool) {
	if len(value) <= maxSize {
		return value, false
	}
	return value[:maxSize] + "...[truncated]", true
}

// validateMimeType validates MIME type against actual content, auto-detecting if missing.
func validateMimeType(value string, declaredMimeType string) string {
	if declaredMimeType == "" {
		if json.Valid([]byte(value)) {
			return "application/json"
		}
		return "text/plain"
	}

	if declaredMimeType == "application/json" && !json.Valid([]byte(value)) {
		return "text/plain"
	}

	return declaredMimeType
}

// extractLLMMetadata extracts LLM-specific metadata from ChatML formatted input.
func extractLLMMetadata(inputValue string) map[string]interface{} {
	metadata := make(map[string]interface{})

	var messages []map[string]interface{}
	if err := json.Unmarshal([]byte(inputValue), &messages); err != nil {
		return metadata
	}

	if len(messages) == 0 {
		return metadata
	}

	if _, hasRole := messages[0]["role"]; !hasRole {
		return metadata
	}

	metadata["brokle.llm.message_count"] = len(messages)

	var userCount, assistantCount, systemCount, toolCount int
	var firstRole, lastRole string
	hasToolCalls := false

	for i, msg := range messages {
		role, _ := msg["role"].(string)

		if i == 0 {
			firstRole = role
		}
		if i == len(messages)-1 {
			lastRole = role
		}

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

		if toolCalls, ok := msg["tool_calls"]; ok && toolCalls != nil {
			hasToolCalls = true
		}
	}

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

	if firstRole != "" {
		metadata["brokle.llm.first_role"] = firstRole
	}
	if lastRole != "" {
		metadata["brokle.llm.last_role"] = lastRole
	}

	metadata["brokle.llm.has_tool_calls"] = hasToolCalls

	return metadata
}

func (s *OTLPConverterService) createSpanEvent(ctx context.Context, span observability.OTLPSpan, resourceAttrs, scopeAttrs map[string]interface{}, resource *observability.Resource, scope *observability.Scope, projectID string) (*brokleEvent, error) {
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

	spanAttrs := extractAttributesFromKeyValues(span.Attributes)
	allAttrs := mergeAttributes(resourceAttrs, scopeAttrs, spanAttrs)
	startTime := convertUnixNano(span.StartTimeUnixNano)
	endTime := convertUnixNano(span.EndTimeUnixNano)
	spanKind := convertSpanKind(span.Kind)
	statusCode := convertStatusCode(span.Status)

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

	// Extract input with priority: gen_ai.input.messages > input.value
	var inputValue string
	var inputMimeType string

	if messages, ok := allAttrs["gen_ai.input.messages"].([]interface{}); ok {
		if messagesJSON, err := json.Marshal(messages); err == nil {
			inputValue = string(messagesJSON)
			inputMimeType = "application/json"
		}
	} else if messages, ok := allAttrs["gen_ai.input.messages"].(string); ok && messages != "" {
		inputValue = messages
		inputMimeType = "application/json"
	}

	if inputValue == "" {
		if genericInput, ok := allAttrs["input.value"].(string); ok && genericInput != "" {
			inputValue = genericInput
			if mimeType, ok := allAttrs["input.mime_type"].(string); ok {
				inputMimeType = validateMimeType(inputValue, mimeType)
			} else {
				inputMimeType = validateMimeType(inputValue, "")
			}
		}
	}

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

		if inputMimeType == "application/json" {
			if llmMetadata := extractLLMMetadata(inputValue); len(llmMetadata) > 0 {
				for key, value := range llmMetadata {
					payload[key] = value
				}
			}
		}
	}

	// Extract output with priority: gen_ai.output.messages > output.value
	var outputValue string
	var outputMimeType string

	if messages, ok := allAttrs["gen_ai.output.messages"].([]interface{}); ok {
		if messagesJSON, err := json.Marshal(messages); err == nil {
			outputValue = string(messagesJSON)
			outputMimeType = "application/json"
		}
	} else if messages, ok := allAttrs["gen_ai.output.messages"].(string); ok && messages != "" {
		outputValue = messages
		outputMimeType = "application/json"
	}

	if outputValue == "" {
		if genericOutput, ok := allAttrs["output.value"].(string); ok && genericOutput != "" {
			outputValue = genericOutput
			if mimeType, ok := allAttrs["output.mime_type"].(string); ok {
				outputMimeType = validateMimeType(outputValue, mimeType)
			} else {
				outputMimeType = validateMimeType(outputValue, "")
			}
		}
	}

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

	extractGenAIFields(allAttrs, payload)
	s.calculateProviderCostsAtIngestion(ctx, allAttrs, payload, projectID)

	payload["span_attributes"] = spanAttrs

	if scope != nil {
		if scope.Name != "" {
			payload["scope_name"] = scope.Name
		}
		if scope.Version != "" {
			payload["scope_version"] = scope.Version
		}
		if len(scopeAttrs) > 0 {
			payload["scope_attributes"] = scopeAttrs
		}
	}

	if traceState, ok := allAttrs["trace_state"].(string); ok && traceState != "" {
		payload["trace_state"] = traceState
	}

	if len(span.Events) > 0 {
		events := make([]map[string]interface{}, len(span.Events))
		for i, event := range span.Events {
			eventMap := make(map[string]interface{})
			if eventTime := convertUnixNano(event.TimeUnixNano); eventTime != nil {
				eventMap["timestamp"] = eventTime.Format(time.RFC3339Nano)
			}
			eventMap["name"] = event.Name
			eventMap["attributes"] = convertToStringMap(extractAttributesFromKeyValues(event.Attributes))
			eventMap["dropped_attributes_count"] = event.DroppedAttributesCount
			events[i] = eventMap
		}
		payload["events"] = events
	}

	if len(span.Links) > 0 {
		links := make([]map[string]interface{}, len(span.Links))
		for i, link := range span.Links {
			linkMap := make(map[string]interface{})
			if traceID, err := convertTraceID(link.TraceID); err == nil {
				linkMap["trace_id"] = traceID
			}
			if spanID, err := convertSpanID(link.SpanID); err == nil {
				linkMap["span_id"] = spanID
			}
			if link.TraceState != nil {
				if ts, ok := link.TraceState.(string); ok {
					linkMap["trace_state"] = ts
				}
			}
			linkMap["attributes"] = convertToStringMap(extractAttributesFromKeyValues(link.Attributes))
			linkMap["dropped_attributes_count"] = link.DroppedAttributesCount
			links[i] = linkMap
		}
		payload["links"] = links
	}

	payload["resource_attributes"] = convertToStringMap(resourceAttrs)
	payload["span_attributes"] = convertToStringMap(spanAttrs)
	payload["scope_attributes"] = convertToStringMap(scopeAttrs)

	if scope != nil {
		payload["scope_name"] = scope.Name
		payload["scope_version"] = scope.Version
	}

	if resource != nil && resource.SchemaUrl != "" {
		payload["resource_schema_url"] = resource.SchemaUrl
	}
	if scope != nil && scope.SchemaUrl != "" {
		payload["scope_schema_url"] = scope.SchemaUrl
	}

	if s.config.PreserveRawOTLP {
		rawOTLPJSON, err := json.Marshal(span)
		if err == nil {
			payload["otlp_span_raw"] = string(rawOTLPJSON)
		} else {
			s.logger.WithError(err).Warn("Failed to marshal raw OTLP span")
		}

		if len(resourceAttrs) > 0 {
			resourceJSON, err := json.Marshal(resourceAttrs)
			if err == nil {
				payload["otlp_resource_attributes"] = string(resourceJSON)
			}
		}

		if len(scopeAttrs) > 0 {
			scopeJSON, err := json.Marshal(scopeAttrs)
			if err == nil {
				payload["otlp_scope_attributes"] = string(scopeJSON)
			}
		}
	}

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

func convertToStringMap(attrs map[string]interface{}) map[string]string {
	if attrs == nil {
		return make(map[string]string)
	}

	result := make(map[string]string, len(attrs))
	for k, v := range attrs {
		result[k] = fmt.Sprintf("%v", v)
	}
	return result
}

func convertTraceID(traceID interface{}) (string, error) {
	switch v := traceID.(type) {
	case string:
		if len(v) == 32 {
			return v, nil
		}
		return "", fmt.Errorf("invalid trace_id length: %d (expected 32)", len(v))
	case map[string]interface{}:
		if data, ok := v["data"].([]interface{}); ok {
			return bytesToHex(data), nil
		}
	case []byte:
		return hex.EncodeToString(v), nil
	}
	return "", fmt.Errorf("unsupported trace_id type: %T", traceID)
}

func convertSpanID(spanID interface{}) (string, error) {
	switch v := spanID.(type) {
	case string:
		if len(v) == 16 {
			return v, nil
		}
		return "", fmt.Errorf("invalid span_id length: %d (expected 16)", len(v))
	case map[string]interface{}:
		if data, ok := v["data"].([]interface{}); ok {
			return bytesToHex(data), nil
		}
	case []byte:
		return hex.EncodeToString(v), nil
	}
	return "", fmt.Errorf("unsupported span_id type: %T", spanID)
}

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

func convertSpanKind(kind int) uint8 {
	switch kind {
	case 0:
		return observability.SpanKindUnspecified
	case 1:
		return observability.SpanKindInternal
	case 2:
		return observability.SpanKindServer
	case 3:
		return observability.SpanKindClient
	case 4:
		return observability.SpanKindProducer
	case 5:
		return observability.SpanKindConsumer
	default:
		return observability.SpanKindInternal
	}
}

func convertStatusCode(status *observability.Status) uint8 {
	if status == nil {
		return observability.StatusCodeUnset
	}
	switch status.Code {
	case 0:
		return observability.StatusCodeUnset
	case 1:
		return observability.StatusCodeOK
	case 2:
		return observability.StatusCodeError
	default:
		return observability.StatusCodeUnset
	}
}

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

func extractAttributesFromKeyValues(kvs []observability.KeyValue) map[string]interface{} {
	attrs := make(map[string]interface{})

	for _, kv := range kvs {
		value := extractValue(kv.Value)
		if value != nil {
			attrs[kv.Key] = value
		}
	}

	return attrs
}

func extractValue(v interface{}) interface{} {
	if v == nil {
		return nil
	}

	switch val := v.(type) {
	case map[string]interface{}:
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

func mergeAttributes(resource, scope, span map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})

	for k, v := range resource {
		merged[k] = v
	}

	for k, v := range scope {
		merged[k] = v
	}

	for k, v := range span {
		merged[k] = v
	}

	return merged
}

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

// extractGenAIFields extracts Gen AI semantic conventions from attributes.
func extractGenAIFields(attrs map[string]interface{}, payload map[string]interface{}) {
	if provider, ok := attrs["gen_ai.provider.name"].(string); ok {
		payload["provider"] = provider
	}

	if responseModel, ok := attrs["gen_ai.response.model"].(string); ok {
		payload["model_name"] = responseModel
	} else if requestModel, ok := attrs["gen_ai.request.model"].(string); ok {
		payload["model_name"] = requestModel
	}

	if messages, ok := attrs["gen_ai.input.messages"].(string); ok {
		payload["input"] = messages
	} else if messages, ok := attrs["gen_ai.input.messages"].([]interface{}); ok {
		if messagesJSON, err := json.Marshal(messages); err == nil {
			payload["input"] = string(messagesJSON)
		}
	}

	if messages, ok := attrs["gen_ai.output.messages"].(string); ok {
		payload["output"] = messages
	} else if messages, ok := attrs["gen_ai.output.messages"].([]interface{}); ok {
		if messagesJSON, err := json.Marshal(messages); err == nil {
			payload["output"] = string(messagesJSON)
		}
	}

	modelParams := make(map[string]interface{})
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

	if len(modelParams) > 0 {
		if paramsJSON, err := json.Marshal(modelParams); err == nil {
			payload["model_parameters"] = string(paramsJSON)
		}
	}

	usageDetails := make(map[string]uint64)
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

	if totalTokens, ok := attrs["brokle.usage.total_tokens"].(float64); ok {
		usageDetails["total"] = uint64(totalTokens)
	} else if totalTokens, ok := attrs["brokle.usage.total_tokens"].(int64); ok {
		usageDetails["total"] = uint64(totalTokens)
	}

	if len(usageDetails) > 0 {
		payload["usage_details"] = usageDetails
	}
}

func bytesToHex(data []interface{}) string {
	bytes := make([]byte, len(data))
	for i, v := range data {
		if f, ok := v.(float64); ok {
			bytes[i] = byte(f)
		}
	}
	return hex.EncodeToString(bytes)
}

func convertToDomainEvents(events []*brokleEvent) []*observability.TelemetryEventRequest {
	result := make([]*observability.TelemetryEventRequest, 0, len(events))
	for _, e := range events {
		eventID, err := ulid.Parse(e.EventID)
		if err != nil {
			continue
		}
		result = append(result, &observability.TelemetryEventRequest{
			EventID:   eventID,
			SpanID:    e.SpanID,
			TraceID:   e.TraceID,
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

// calculateProviderCostsAtIngestion calculates provider costs for cost visibility.
func (s *OTLPConverterService) calculateProviderCostsAtIngestion(
	ctx context.Context,
	attrs map[string]interface{},
	payload map[string]interface{},
	projectID string,
) {
	modelName := extractStringFromInterface(attrs["gen_ai.request.model"])
	if modelName == "" {
		return
	}

	usage := make(map[string]uint64)

	if val := extractUint64FromInterface(attrs["gen_ai.usage.input_tokens"]); val > 0 {
		usage["input"] = val
	}
	if val := extractUint64FromInterface(attrs["gen_ai.usage.output_tokens"]); val > 0 {
		usage["output"] = val
	}
	if val := extractUint64FromInterface(attrs["gen_ai.usage.input_tokens.cache_read"]); val > 0 {
		usage["cache_read_input_tokens"] = val
	}
	if val := extractUint64FromInterface(attrs["gen_ai.usage.input_tokens.cache_creation"]); val > 0 {
		usage["cache_creation_input_tokens"] = val
	}
	if val := extractUint64FromInterface(attrs["gen_ai.usage.input_audio_tokens"]); val > 0 {
		usage["audio_input"] = val
	}
	if val := extractUint64FromInterface(attrs["gen_ai.usage.output_audio_tokens"]); val > 0 {
		usage["audio_output"] = val
	}
	if val := extractUint64FromInterface(attrs["gen_ai.usage.reasoning_tokens"]); val > 0 {
		usage["reasoning_tokens"] = val
	}
	if val := extractUint64FromInterface(attrs["gen_ai.usage.image_tokens"]); val > 0 {
		usage["image_tokens"] = val
	}
	if val := extractUint64FromInterface(attrs["gen_ai.usage.video_tokens"]); val > 0 {
		usage["video_tokens"] = val
	}

	// Calculate total (excludes cache subsets which are already counted in input)
	var total uint64
	if input, ok := usage["input"]; ok {
		total += input
	}
	if output, ok := usage["output"]; ok {
		total += output
	}
	if reasoning, ok := usage["reasoning_tokens"]; ok {
		total += reasoning
	}
	if audioIn, ok := usage["audio_input"]; ok {
		total += audioIn
	}
	if audioOut, ok := usage["audio_output"]; ok {
		total += audioOut
	}
	if image, ok := usage["image_tokens"]; ok {
		total += image
	}
	if video, ok := usage["video_tokens"]; ok {
		total += video
	}

	if total > 0 {
		usage["total"] = total
	}

	if len(usage) == 0 {
		return
	}

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
		payload["usage_details"] = usage
		return
	}

	providerCost := s.providerPricingService.CalculateProviderCost(usage, providerPricing)

	providerPricingSnapshot := make(map[string]decimal.Decimal)
	for usageType, price := range providerPricing.Pricing {
		key := fmt.Sprintf("%s_price_per_million", usageType)
		providerPricingSnapshot[key] = price
	}

	payload["usage_details"] = usage
	payload["cost_details"] = providerCost
	payload["pricing_snapshot"] = providerPricingSnapshot

	if totalCost, ok := providerCost["total"]; ok {
		payload["total_cost"] = totalCost
	}

	s.logger.WithFields(logrus.Fields{
		"model":                 modelName,
		"usage_types":           len(usage),
		"total_tokens":          total,
		"provider_cost_usd":     providerCost["total"],
		"provider_pricing_date": providerPricing.SnapshotTime,
	}).Debug("Provider costs calculated successfully")
}

func extractStringFromInterface(val interface{}) string {
	if str, ok := val.(string); ok {
		return str
	}
	return ""
}

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

func extractBoolFromInterface(val interface{}) bool {
	if b, ok := val.(bool); ok {
		return b
	}
	return false
}
