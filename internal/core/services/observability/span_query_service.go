package observability

import (
	"context"
	"log/slog"

	obsDomain "brokle/internal/core/domain/observability"
	appErrors "brokle/pkg/errors"
)

// SpanQueryService orchestrates the SDK span query flow.
// It parses filter expressions, builds safe SQL queries, and executes them.
// Parser and query builder are created per-request for thread safety.
type SpanQueryService struct {
	traceRepo obsDomain.TraceRepository
	logger    *slog.Logger
}

// NewSpanQueryService creates a new span query service.
func NewSpanQueryService(
	traceRepo obsDomain.TraceRepository,
	logger *slog.Logger,
) *SpanQueryService {
	return &SpanQueryService{
		traceRepo: traceRepo,
		logger:    logger.With("service", "span_query"),
	}
}

// QuerySpans executes a span query using the filter expression syntax.
func (s *SpanQueryService) QuerySpans(
	ctx context.Context,
	projectID string,
	req *obsDomain.SpanQueryRequest,
) (*obsDomain.SpanQueryResponse, error) {
	if errs := obsDomain.ValidateSpanQueryRequest(req); len(errs) > 0 {
		return nil, appErrors.InvalidParam("query", errs[0].Message)
	}

	obsDomain.NormalizeSpanQueryRequest(req)

	parser := NewFilterParser()
	queryBuilder := NewSpanQueryBuilder()

	filterNode, err := parser.Parse(req.Filter)
	if err != nil {
		s.logger.Debug("filter parse error",
			"filter", req.Filter,
			"error", err,
		)
		return nil, appErrors.InvalidParam("filter", err.Error())
	}

	// Calculate offset from page (page is 1-indexed)
	offset := (req.Page - 1) * req.Limit
	if offset < 0 {
		offset = 0
	}

	queryResult, err := queryBuilder.BuildQuery(
		filterNode,
		projectID,
		req.StartTime,
		req.EndTime,
		req.Limit,
		offset,
	)
	if err != nil {
		s.logger.Error("query build error",
			"project_id", projectID,
			"filter", req.Filter,
			"error", err,
		)
		return nil, appErrors.Internal("failed to build query", err)
	}

	countResult, err := queryBuilder.BuildCountQuery(
		filterNode,
		projectID,
		req.StartTime,
		req.EndTime,
	)
	if err != nil {
		s.logger.Error("count query build error",
			"project_id", projectID,
			"filter", req.Filter,
			"error", err,
		)
		return nil, appErrors.Internal("failed to build count query", err)
	}

	spans, err := s.traceRepo.QuerySpansByExpression(ctx, queryResult.Query, queryResult.Args)
	if err != nil {
		s.logger.Error("query execution error",
			"project_id", projectID,
			"error", err,
		)
		return nil, appErrors.Internal("query execution failed", err)
	}

	totalCount, err := s.traceRepo.CountSpansByExpression(ctx, countResult.Query, countResult.Args)
	if err != nil {
		s.logger.Error("count query execution error",
			"project_id", projectID,
			"error", err,
		)
		return nil, appErrors.Internal("count query failed", err)
	}

	s.logger.Debug("span query executed",
		"project_id", projectID,
		"filter", req.Filter,
		"result_count", len(spans),
		"total_count", totalCount,
	)

	return &obsDomain.SpanQueryResponse{
		Spans:      spans,
		TotalCount: totalCount,
		HasMore:    int64(offset+len(spans)) < totalCount,
	}, nil
}

// ValidateFilter validates a filter expression without executing it.
// This is useful for SDK clients to validate filters before submitting queries.
func (s *SpanQueryService) ValidateFilter(filter string) error {
	parser := NewFilterParser()
	_, err := parser.Parse(filter)
	if err != nil {
		// WithCode sets `error.code = "invalid_filter_expression"` on the
		// wire envelope. SDKs discriminate filter-syntax 422s from
		// generic input-validation 422s (invalid limit, page, etc.) on
		// that code, matching the Stripe/OpenAI/JSON:API convention.
		return appErrors.InvalidParam("filter", err.Error())
	}
	return nil
}
