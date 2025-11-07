package observability

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"brokle/internal/core/domain/observability"
)

func TestOTLPConverterService_ConvertOTLPToBrokleEvents_TraceInputOutput(t *testing.T) {
	// Create converter service
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	converter := NewOTLPConverterService(logger)

	// Load test fixture
	data, err := os.ReadFile("../../../../tests/fixtures/otlp_trace_large_payload.json")
	require.NoError(t, err)

	var otlpReq observability.OTLPRequest
	err = json.Unmarshal(data, &otlpReq)
	require.NoError(t, err)

	// Convert OTLP to Brokle events
	events, err := converter.ConvertOTLPToBrokleEvents(&otlpReq)
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

	// Verify trace has input and output
	input, ok := traceEvent.Payload["input"].(string)
	assert.True(t, ok, "Trace should have input field")
	assert.NotEmpty(t, input, "Trace input should not be empty")
	t.Logf("Trace input length: %d", len(input))

	output, ok := traceEvent.Payload["output"].(string)
	assert.True(t, ok, "Trace should have output field")
	assert.NotEmpty(t, output, "Trace output should not be empty")
	t.Logf("Trace output length: %d", len(output))

	// Verify trace has input_preview and output_preview
	inputPreview, ok := traceEvent.Payload["input_preview"].(string)
	assert.True(t, ok, "Trace should have input_preview field")
	assert.NotEmpty(t, inputPreview, "Trace input_preview should not be empty")
	t.Logf("Trace input_preview length: %d", len(inputPreview))

	outputPreview, ok := traceEvent.Payload["output_preview"].(string)
	assert.True(t, ok, "Trace should have output_preview field")
	assert.NotEmpty(t, outputPreview, "Trace output_preview should not be empty")
	t.Logf("Trace output_preview length: %d", len(outputPreview))

	// Verify preview format
	assert.Contains(t, inputPreview, "[", "Preview should have format header")
	assert.Contains(t, outputPreview, "[", "Preview should have format header")
}

func TestOTLPConverterService_RootSpanDetection(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	converter := NewOTLPConverterService(logger)

	testCases := []struct {
		name         string
		parentSpanID interface{}
		expectRoot   bool
		description  string
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
								Spans: []observability.Span{
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

			events, err := converter.ConvertOTLPToBrokleEvents(otlpReq)
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

			// Find observation event
			var obsEvent *observability.TelemetryEventRequest
			for _, e := range events {
				if e.EventType == "observation" {
					obsEvent = e
					break
				}
			}
			require.NotNil(t, obsEvent, "Should have observation event")

			if tc.expectRoot {
				assert.True(t, hasTraceEvent, "Expected trace event for root span: %s", tc.description)
				// Verify root span has nil parent_observation_id (will be NULL in ClickHouse)
				assert.Nil(t, obsEvent.Payload["parent_observation_id"], "Root span should have nil parent_observation_id: %s", tc.description)
			} else {
				assert.False(t, hasTraceEvent, "Expected NO trace event for non-root span: %s", tc.description)
				// Verify child span has non-nil parent_observation_id
				assert.NotNil(t, obsEvent.Payload["parent_observation_id"], "Child span should have parent_observation_id: %s", tc.description)
			}
		})
	}
}
