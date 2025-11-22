package observability

import (
	"context"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"brokle/internal/core/domain/observability"
)

// TestMalformedChatML_GracefulDegradation tests handling of malformed ChatML data
func TestMalformedChatML_GracefulDegradation(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	converter := NewOTLPConverterService(logger, &mockCostCalculator{})

	tests := []struct {
		name        string
		inputValue  string
		description string
	}{
		{
			name:        "Invalid JSON",
			inputValue:  `{invalid json`,
			description: "Should not extract metadata",
		},
		{
			name:        "Missing role field",
			inputValue:  `[{"content":"Hello"}]`,
			description: "Should not extract metadata (no role field)",
		},
		{
			name:        "Empty messages array",
			inputValue:  `[]`,
			description: "Should not extract metadata (empty array)",
		},
		{
			name:        "Not an array",
			inputValue:  `{"role":"user","content":"Hello"}`,
			description: "Should not extract metadata (object, not array)",
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
											{Key: "input.mime_type", Value: "application/json"},
										},
									},
								},
							},
						},
					},
				},
			}

			// Should not panic or error
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

			// Input should still be populated (raw value)
			assert.Equal(t, tc.inputValue, traceEvent.Payload["input"])

			// LLM metadata should NOT be present (malformed ChatML)
			assert.Nil(t, traceEvent.Payload["brokle.llm.message_count"])
			assert.Nil(t, traceEvent.Payload["brokle.llm.has_tool_calls"])
		})
	}
}

// TestHelperFunctions_TruncateWithIndicator tests the truncation helper
func TestHelperFunctions_TruncateWithIndicator(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		maxSize           int
		expectedTruncated bool
	}{
		{
			name:              "Short string not truncated",
			input:             "hello",
			maxSize:           100,
			expectedTruncated: false,
		},
		{
			name:              "Long string truncated",
			input:             string(make([]byte, 1000)),
			maxSize:           100,
			expectedTruncated: true,
		},
		{
			name:              "Exact size not truncated",
			input:             string(make([]byte, 100)),
			maxSize:           100,
			expectedTruncated: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, truncated := truncateWithIndicator(tc.input, tc.maxSize)
			assert.Equal(t, tc.expectedTruncated, truncated)

			if truncated {
				assert.Contains(t, result, "...[truncated]")
				assert.LessOrEqual(t, len(result), tc.maxSize+len("...[truncated]"))
			} else {
				assert.Equal(t, tc.input, result)
			}
		})
	}
}

// TestHelperFunctions_ValidateMimeType tests MIME type validation helper
func TestHelperFunctions_ValidateMimeType(t *testing.T) {
	tests := []struct {
		name         string
		value        string
		declaredType string
		expected     string
	}{
		{
			name:         "Missing MIME type with JSON",
			value:        `{"key":"value"}`,
			declaredType: "",
			expected:     "application/json",
		},
		{
			name:         "Missing MIME type with text",
			value:        "plain text",
			declaredType: "",
			expected:     "text/plain",
		},
		{
			name:         "Correct JSON declaration",
			value:        `{"key":"value"}`,
			declaredType: "application/json",
			expected:     "application/json",
		},
		{
			name:         "Incorrect JSON declaration",
			value:        "not json",
			declaredType: "application/json",
			expected:     "text/plain", // Corrected
		},
		{
			name:         "Text declared correctly",
			value:        "hello",
			declaredType: "text/plain",
			expected:     "text/plain",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := validateMimeType(tc.value, tc.declaredType)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestHelperFunctions_ExtractLLMMetadata tests LLM metadata extraction
func TestHelperFunctions_ExtractLLMMetadata(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]interface{}
	}{
		{
			name:  "Valid ChatML",
			input: `[{"role":"user","content":"Hello"},{"role":"assistant","content":"Hi"}]`,
			expected: map[string]interface{}{
				"brokle.llm.message_count":          2,
				"brokle.llm.user_message_count":     1,
				"brokle.llm.assistant_message_count": 1,
				"brokle.llm.first_role":             "user",
				"brokle.llm.last_role":              "assistant",
				"brokle.llm.has_tool_calls":         false,
			},
		},
		{
			name:  "ChatML with tool calls",
			input: `[{"role":"assistant","content":"Using tool","tool_calls":[{"id":"1"}]}]`,
			expected: map[string]interface{}{
				"brokle.llm.message_count":           1,
				"brokle.llm.assistant_message_count": 1,
				"brokle.llm.first_role":              "assistant",
				"brokle.llm.last_role":               "assistant",
				"brokle.llm.has_tool_calls":          true,
			},
		},
		{
			name:     "Invalid JSON",
			input:    `{not json}`,
			expected: map[string]interface{}{}, // Empty metadata
		},
		{
			name:     "Empty array",
			input:    `[]`,
			expected: map[string]interface{}{}, // Empty metadata
		},
		{
			name:     "Missing role field",
			input:    `[{"content":"Hello"}]`,
			expected: map[string]interface{}{}, // Empty metadata
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := extractLLMMetadata(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
