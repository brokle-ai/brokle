package observability

import (
	"context"

	"github.com/shopspring/decimal"

	"brokle/internal/core/domain/observability"
)

type AggregationService struct {
	traceRepo observability.TraceRepository
}

func NewAggregationService(traceRepo observability.TraceRepository) *AggregationService {
	return &AggregationService{
		traceRepo: traceRepo,
	}
}

type TraceAggregations struct {
	TotalCost   decimal.Decimal `json:"total_cost"`
	TotalTokens uint32          `json:"total_tokens"`
	SpanCount   uint32          `json:"span_count"`
}

func (s *AggregationService) CalculateTraceAggregations(ctx context.Context, traceID string) (*TraceAggregations, error) {
	spans, err := s.traceRepo.GetSpansByTraceID(ctx, traceID)
	if err != nil {
		return nil, err
	}

	totalCost := decimal.Zero
	var totalTokens uint32

	for _, span := range spans {
		if span.TotalCost != nil {
			totalCost = totalCost.Add(*span.TotalCost)
		}

		if span.UsageDetails != nil {
			if total, ok := span.UsageDetails["total"]; ok {
				totalTokens += uint32(total)
			}
		}
	}

	return &TraceAggregations{
		TotalCost:   totalCost,
		TotalTokens: totalTokens,
		SpanCount:   uint32(len(spans)),
	}, nil
}

func (s *AggregationService) CalculateBatchAggregations(ctx context.Context, traceIDs []string) (map[string]*TraceAggregations, error) {
	result := make(map[string]*TraceAggregations, len(traceIDs))

	for _, traceID := range traceIDs {
		aggs, err := s.CalculateTraceAggregations(ctx, traceID)
		if err != nil {
			continue
		}
		result[traceID] = aggs
	}

	return result, nil
}
