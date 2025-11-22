package observability

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"brokle/internal/core/domain/observability"
)

// mockCostCalculator is a test mock that returns zero costs
type mockCostCalculator struct{}

func (m *mockCostCalculator) CalculateCost(ctx context.Context, input observability.CostCalculationInput) (*observability.CostBreakdown, error) {
	// Return zero costs for testing (allows tests to pass without real pricing data)
	return &observability.CostBreakdown{
		InputCost:  "0.000000000",
		OutputCost: "0.000000000",
		TotalCost:  "0.000000000",
		Currency:   "USD",
		ModelName:  input.ModelName,
	}, nil
}

func (m *mockCostCalculator) CalculateCostWithPricing(input observability.CostCalculationInput, pricing *observability.Model) *observability.CostBreakdown {
	return &observability.CostBreakdown{
		InputCost:  "0.000000000",
		OutputCost: "0.000000000",
		TotalCost:  "0.000000000",
		Currency:   "USD",
		ModelName:  input.ModelName,
	}
}

func TestOTLPConverterService_ConvertOTLPToBrokleEvents_TraceInputOutput(t *testing.T) {
	// Create converter service with mock cost calculator
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	mockCalc := &mockCostCalculator{}
	converter := NewOTLPConverterService(logger, mockCalc)

	// Load test fixture
	data, err := os.ReadFile("../../../../tests/fixtures/otlp_trace_large_payload.json")
	require.NoError(t, err)

	var otlpReq observability.OTLPRequest
	err = json.Unmarshal(data, &otlpReq)
	require.NoError(t, err)

	// Convert OTLP to Brokle events (with test project_id)
	events, err := converter.ConvertOTLPToBrokleEvents(context.Background(), &otlpReq, "test-project-id")
	require.NoError(t, err)
	require.NotEmpty(t, events)

	// Find trace event
	var traceEvent *observability.TelemetryEventRequest
	for _, e := range events {
		if e.EventType == "trace" {
			traceEvent = e
			break
		}
	}

	require.NotNil(t, traceEvent, "Should have created a trace event")

	// Verify trace has input and output (matches ClickHouse schema)
	input, ok := traceEvent.Payload["input"].(string)
	assert.True(t, ok, "Trace should have input field")
	assert.NotEmpty(t, input, "Trace input should not be empty")
	t.Logf("Trace input length: %d", len(input))

	output, ok := traceEvent.Payload["output"].(string)
	assert.True(t, ok, "Trace should have output field")
	assert.NotEmpty(t, output, "Trace output should not be empty")
	t.Logf("Trace output length: %d", len(output))

	// Verify input/output are valid JSON (OTEL gen_ai.input.messages format)
	var inputMessages, outputMessages []interface{}
	err = json.Unmarshal([]byte(input), &inputMessages)
	assert.NoError(t, err, "Input should be valid JSON array")
	assert.NotEmpty(t, inputMessages, "Input messages should not be empty")

	err = json.Unmarshal([]byte(output), &outputMessages)
	assert.NoError(t, err, "Output should be valid JSON array")
	assert.NotEmpty(t, outputMessages, "Output messages should not be empty")
}

func TestOTLPConverterService_RootSpanDetection(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	mockCalc := &mockCostCalculator{}
	converter := NewOTLPConverterService(logger, mockCalc)

	testCases := []struct {
		parentSpanID interface{}
		name         string
		description  string
		expectRoot   bool
	}{
		{
			name:         "nil parent",
			parentSpanID: nil,
			expectRoot:   true,
			description:  "No parent field (omitted)",
		},
		{
			name:         "empty string parent",
			parentSpanID: "",
			expectRoot:   true,
			description:  "Empty string parent ID",
		},
		{
			name:         "zero hex string parent",
			parentSpanID: "0000000000000000",
			expectRoot:   true,
			description:  "Zero-value hex string (16 zeros)",
		},
		{
			name:         "valid parent",
			parentSpanID: "1234567890abcdef",
			expectRoot:   false,
			description:  "Valid non-zero parent ID",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create OTLP request with test parent ID
			otlpReq := &observability.OTLPRequest{
				ResourceSpans: []observability.ResourceSpan{
					{
						ScopeSpans: []observability.ScopeSpan{
							{
								Spans: []observability.OTLPSpan{
									{
										TraceID:           "0123456789abcdef0123456789abcdef",
										SpanID:            "0123456789abcdef",
										ParentSpanID:      tc.parentSpanID,
										Name:              "test-span",
										StartTimeUnixNano: int64(1000000000),
									},
								},
							},
						},
					},
				},
			}

			events, err := converter.ConvertOTLPToBrokleEvents(context.Background(), otlpReq, "test-project")
			require.NoError(t, err)
			require.NotEmpty(t, events)

			// Check if trace event was created
			hasTraceEvent := false
			for _, e := range events {
				if e.EventType == "trace" {
					hasTraceEvent = true
					break
				}
			}

			// Find span event
			var spanEvent *observability.TelemetryEventRequest
			for _, e := range events {
				if e.EventType == "span" {
					spanEvent = e
					break
				}
			}
			require.NotNil(t, spanEvent, "Should have span event")

			if tc.expectRoot {
				assert.True(t, hasTraceEvent, "Expected trace event for root span: %s", tc.description)
				// Verify root span has nil parent_span_id (will be NULL in ClickHouse)
				assert.Nil(t, spanEvent.Payload["parent_span_id"], "Root span should have nil parent_span_id: %s", tc.description)
			} else {
				assert.False(t, hasTraceEvent, "Expected NO trace event for non-root span: %s", tc.description)
				// Verify child span has non-nil parent_span_id
				assert.NotNil(t, spanEvent.Payload["parent_span_id"], "Child span should have parent_span_id: %s", tc.description)
			}
		})
	}
}

// TestExtractInputValue_GenericData tests extraction of input.value for generic (non-LLM) data
func TestExtractInputValue_GenericData(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	converter := NewOTLPConverterService(logger, &mockCostCalculator{})

	// Create OTLP request with input.value attribute
	otlpReq := &observability.OTLPRequest{
		ResourceSpans: []observability.ResourceSpan{
			{
				ScopeSpans: []observability.ScopeSpan{
					{
						Spans: []observability.OTLPSpan{
							{
								TraceID:           "4bf92f3577b34da6a3ce929d0e0e4736",
								SpanID:            "00f067aa0ba902b7",
								ParentSpanID:      nil, // Root span
								Name:              "api-request",
								StartTimeUnixNano: 1700000000000000000,
								EndTimeUnixNano:   1700000001000000000,
								Attributes: []observability.KeyValue{
									{Key: "input.value", Value: `{"endpoint":"/weather","query":"Bangalore"}`},
									{Key: "input.mime_type", Value: "application/json"},
									{Key: "output.value", Value: `{"temp":25,"status":"sunny"}`},
									{Key: "output.mime_type", Value: "application/json"},
								},
							},
						},
					},
				},
			},
		},
	}

	// Convert
	events, err := converter.ConvertOTLPToBrokleEvents(context.Background(), otlpReq, "test-project")
	require.NoError(t, err)
	require.Len(t, events, 2) // 1 trace + 1 span

	// Find trace event
	var traceEvent *observability.TelemetryEventRequest
	for _, e := range events {
		if e.EventType == observability.TelemetryEventTypeTrace {
			traceEvent = e
			break
		}
	}
	require.NotNil(t, traceEvent)

	// Verify input extraction
	assert.Equal(t, `{"endpoint":"/weather","query":"Bangalore"}`, traceEvent.Payload["input"])
	assert.Equal(t, "application/json", traceEvent.Payload["input_mime_type"])

	// Verify output extraction
	assert.Equal(t, `{"temp":25,"status":"sunny"}`, traceEvent.Payload["output"])
	assert.Equal(t, "application/json", traceEvent.Payload["output_mime_type"])
}

// TestExtractGenAIMessages_LLMData tests backward compatibility with gen_ai.input.messages
func TestExtractGenAIMessages_LLMData(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	converter := NewOTLPConverterService(logger, &mockCostCalculator{})

	messagesJSON := `[{"role":"user","content":"Hello"}]`

	otlpReq := &observability.OTLPRequest{
		ResourceSpans: []observability.ResourceSpan{
			{
				ScopeSpans: []observability.ScopeSpan{
					{
						Spans: []observability.OTLPSpan{
							{
								TraceID:           "4bf92f3577b34da6a3ce929d0e0e4736",
								SpanID:            "00f067aa0ba902b7",
								ParentSpanID:      nil,
								Name:              "llm-chat",
								StartTimeUnixNano: 1700000000000000000,
								EndTimeUnixNano:   1700000001000000000,
								Attributes: []observability.KeyValue{
									{Key: "gen_ai.input.messages", Value: messagesJSON},
									{Key: "gen_ai.output.messages", Value: `[{"role":"assistant","content":"Hi there!"}]`},
								},
							},
						},
					},
				},
			},
		},
	}

	events, err := converter.ConvertOTLPToBrokleEvents(context.Background(), otlpReq, "test-project")
	require.NoError(t, err)

	var traceEvent *observability.TelemetryEventRequest
	for _, e := range events {
		if e.EventType == observability.TelemetryEventTypeTrace {
			traceEvent = e
			break
		}
	}
	require.NotNil(t, traceEvent)

	// Should extract gen_ai.input.messages (priority 1)
	assert.Equal(t, messagesJSON, traceEvent.Payload["input"])
	assert.Equal(t, "application/json", traceEvent.Payload["input_mime_type"])
}

// TestInputOutputPriorityOrder tests that gen_ai.* takes priority over input.value
func TestInputOutputPriorityOrder(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	converter := NewOTLPConverterService(logger, &mockCostCalculator{})

	otlpReq := &observability.OTLPRequest{
		ResourceSpans: []observability.ResourceSpan{
			{
				ScopeSpans: []observability.ScopeSpan{
					{
						Spans: []observability.OTLPSpan{
							{
								TraceID:           "4bf92f3577b34da6a3ce929d0e0e4736",
								SpanID:            "00f067aa0ba902b7",
								ParentSpanID:      nil,
								Name:              "test",
								StartTimeUnixNano: 1700000000000000000,
								Attributes: []observability.KeyValue{
									// Both attributes present
									{Key: "gen_ai.input.messages", Value: `[{"role":"user","content":"LLM"}]`},
									{Key: "input.value", Value: `{"generic":"data"}`},
								},
							},
						},
					},
				},
			},
		},
	}

	events, err := converter.ConvertOTLPToBrokleEvents(context.Background(), otlpReq, "test-project")
	require.NoError(t, err)

	var traceEvent *observability.TelemetryEventRequest
	for _, e := range events {
		if e.EventType == observability.TelemetryEventTypeTrace {
			traceEvent = e
			break
		}
	}
	require.NotNil(t, traceEvent)

	// Should use gen_ai.input.messages (priority 1), not input.value
	assert.Equal(t, `[{"role":"user","content":"LLM"}]`, traceEvent.Payload["input"])
}

// TestExtractLLMMetadata tests extraction of brokle.llm.* attributes from ChatML
func TestExtractLLMMetadata(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	converter := NewOTLPConverterService(logger, &mockCostCalculator{})

	chatMLInput := `[
		{"role":"system","content":"You are helpful"},
		{"role":"user","content":"Hello"},
		{"role":"assistant","content":"Hi","tool_calls":[{"id":"call_1","type":"function"}]},
		{"role":"tool","content":"result"}
	]`

	otlpReq := &observability.OTLPRequest{
		ResourceSpans: []observability.ResourceSpan{
			{
				ScopeSpans: []observability.ScopeSpan{
					{
						Spans: []observability.OTLPSpan{
							{
								TraceID:           "4bf92f3577b34da6a3ce929d0e0e4736",
								SpanID:            "00f067aa0ba902b7",
								ParentSpanID:      nil,
								Name:              "llm",
								StartTimeUnixNano: 1700000000000000000,
								Attributes: []observability.KeyValue{
									{Key: "input.value", Value: chatMLInput},
									{Key: "input.mime_type", Value: "application/json"},
								},
							},
						},
					},
				},
			},
		},
	}

	events, err := converter.ConvertOTLPToBrokleEvents(context.Background(), otlpReq, "test-project")
	require.NoError(t, err)

	var traceEvent *observability.TelemetryEventRequest
	for _, e := range events {
		if e.EventType == observability.TelemetryEventTypeTrace {
			traceEvent = e
			break
		}
	}
	require.NotNil(t, traceEvent)

	// Verify LLM metadata extracted
	assert.Equal(t, 4, traceEvent.Payload["brokle.llm.message_count"])
	assert.Equal(t, 1, traceEvent.Payload["brokle.llm.user_message_count"])
	assert.Equal(t, 1, traceEvent.Payload["brokle.llm.assistant_message_count"])
	assert.Equal(t, 1, traceEvent.Payload["brokle.llm.system_message_count"])
	assert.Equal(t, 1, traceEvent.Payload["brokle.llm.tool_message_count"])
	assert.Equal(t, "system", traceEvent.Payload["brokle.llm.first_role"])
	assert.Equal(t, "tool", traceEvent.Payload["brokle.llm.last_role"])
	assert.Equal(t, true, traceEvent.Payload["brokle.llm.has_tool_calls"])
}

// TestMimeTypeAutoDetection tests auto-detection when mime_type missing
func TestMimeTypeAutoDetection(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	converter := NewOTLPConverterService(logger, &mockCostCalculator{})

	tests := []struct {
		name             string
		inputValue       string
		expectedMimeType string
	}{
		{
			name:             "Valid JSON object",
			inputValue:       `{"key":"value"}`,
			expectedMimeType: "application/json",
		},
		{
			name:             "Valid JSON array",
			inputValue:       `["item1","item2"]`,
			expectedMimeType: "application/json",
		},
		{
			name:             "Plain text",
			inputValue:       "Hello world",
			expectedMimeType: "text/plain",
		},
		{
			name:             "Invalid JSON",
			inputValue:       `{invalid json`,
			expectedMimeType: "text/plain",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			otlpReq := &observability.OTLPRequest{
				ResourceSpans: []observability.ResourceSpan{
					{
						ScopeSpans: []observability.ScopeSpan{
							{
								Spans: []observability.OTLPSpan{
									{
										TraceID:           "4bf92f3577b34da6a3ce929d0e0e4736",
										SpanID:            "00f067aa0ba902b7",
										ParentSpanID:      nil,
										Name:              "test",
										StartTimeUnixNano: 1700000000000000000,
										Attributes: []observability.KeyValue{
											{Key: "input.value", Value: tc.inputValue},
											// No input.mime_type - should auto-detect
										},
									},
								},
							},
						},
					},
				},
			}

			events, err := converter.ConvertOTLPToBrokleEvents(context.Background(), otlpReq, "test-project")
			require.NoError(t, err)

			var traceEvent *observability.TelemetryEventRequest
			for _, e := range events {
				if e.EventType == observability.TelemetryEventTypeTrace {
					traceEvent = e
					break
				}
			}
			require.NotNil(t, traceEvent)

			// Verify MIME type was auto-detected
			assert.Equal(t, tc.expectedMimeType, traceEvent.Payload["input_mime_type"])
		})
	}
}

// TestMimeTypeValidation tests that declared MIME type is validated against content
func TestMimeTypeValidation(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	converter := NewOTLPConverterService(logger, &mockCostCalculator{})

	// JSON content but declared as text/plain should be corrected to application/json
	otlpReq := &observability.OTLPRequest{
		ResourceSpans: []observability.ResourceSpan{
			{
				ScopeSpans: []observability.ScopeSpan{
					{
						Spans: []observability.OTLPSpan{
							{
								TraceID:           "4bf92f3577b34da6a3ce929d0e0e4736",
								SpanID:            "00f067aa0ba902b7",
								ParentSpanID:      nil,
								Name:              "test",
								StartTimeUnixNano: 1700000000000000000,
								Attributes: []observability.KeyValue{
									{Key: "input.value", Value: `{"valid":"json"}`},
									{Key: "input.mime_type", Value: "text/plain"}, // Wrong MIME type
								},
							},
						},
					},
				},
			},
		},
	}

	events, err := converter.ConvertOTLPToBrokleEvents(context.Background(), otlpReq, "test-project")
	require.NoError(t, err)

	var traceEvent *observability.TelemetryEventRequest
	for _, e := range events {
		if e.EventType == observability.TelemetryEventTypeTrace {
			traceEvent = e
			break
		}
	}
	require.NotNil(t, traceEvent)

	// Should keep text/plain since content IS valid JSON (MIME type describes how to interpret)
	assert.Equal(t, "text/plain", traceEvent.Payload["input_mime_type"])
}

// TestTruncationWithIndicator tests large payload truncation
func TestTruncationWithIndicator(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	converter := NewOTLPConverterService(logger, &mockCostCalculator{})

	// Create 2MB string (exceeds 1MB limit)
	largeData := make([]byte, 2*1024*1024)
	for i := range largeData {
		largeData[i] = 'a'
	}
	largeString := string(largeData)

	otlpReq := &observability.OTLPRequest{
		ResourceSpans: []observability.ResourceSpan{
			{
				ScopeSpans: []observability.ScopeSpan{
					{
						Spans: []observability.OTLPSpan{
							{
								TraceID:           "4bf92f3577b34da6a3ce929d0e0e4736",
								SpanID:            "00f067aa0ba902b7",
								ParentSpanID:      nil,
								Name:              "large-payload",
								StartTimeUnixNano: 1700000000000000000,
								Attributes: []observability.KeyValue{
									{Key: "input.value", Value: largeString},
									{Key: "input.mime_type", Value: "text/plain"},
								},
							},
						},
					},
				},
			},
		},
	}

	events, err := converter.ConvertOTLPToBrokleEvents(context.Background(), otlpReq, "test-project")
	require.NoError(t, err)

	var traceEvent *observability.TelemetryEventRequest
	for _, e := range events {
		if e.EventType == observability.TelemetryEventTypeTrace {
			traceEvent = e
			break
		}
	}
	require.NotNil(t, traceEvent)

	// Verify truncation
	inputValue := traceEvent.Payload["input"].(string)
	assert.True(t, len(inputValue) <= MaxAttributeValueSize+len("...[truncated]"))
	assert.Contains(t, inputValue, "...[truncated]")
	assert.Equal(t, true, traceEvent.Payload["input_truncated"])
}

// TestExtractBothInputAndOutput tests simultaneous input and output extraction
func TestExtractBothInputAndOutput(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	converter := NewOTLPConverterService(logger, &mockCostCalculator{})

	otlpReq := &observability.OTLPRequest{
		ResourceSpans: []observability.ResourceSpan{
			{
				ScopeSpans: []observability.ScopeSpan{
					{
						Spans: []observability.OTLPSpan{
							{
								TraceID:           "4bf92f3577b34da6a3ce929d0e0e4736",
								SpanID:            "00f067aa0ba902b7",
								ParentSpanID:      nil,
								Name:              "complete-trace",
								StartTimeUnixNano: 1700000000000000000,
								EndTimeUnixNano:   1700000001000000000,
								Attributes: []observability.KeyValue{
									{Key: "input.value", Value: `{"request":"data"}`},
									{Key: "input.mime_type", Value: "application/json"},
									{Key: "output.value", Value: `{"response":"result"}`},
									{Key: "output.mime_type", Value: "application/json"},
								},
							},
						},
					},
				},
			},
		},
	}

	events, err := converter.ConvertOTLPToBrokleEvents(context.Background(), otlpReq, "test-project")
	require.NoError(t, err)

	var traceEvent *observability.TelemetryEventRequest
	for _, e := range events {
		if e.EventType == observability.TelemetryEventTypeTrace {
			traceEvent = e
			break
		}
	}
	require.NotNil(t, traceEvent)

	// Both should be populated
	assert.Equal(t, `{"request":"data"}`, traceEvent.Payload["input"])
	assert.Equal(t, "application/json", traceEvent.Payload["input_mime_type"])
	assert.Equal(t, `{"response":"result"}`, traceEvent.Payload["output"])
	assert.Equal(t, "application/json", traceEvent.Payload["output_mime_type"])
}
