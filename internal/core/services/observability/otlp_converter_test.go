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
