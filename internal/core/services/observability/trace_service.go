package observability

import (
	"context"

	"github.com/sirupsen/logrus"

	"brokle/internal/core/domain/observability"
	appErrors "brokle/pkg/errors"
)

// TraceService implements OTEL-native trace operations
// Traces are virtual (derived from root spans where parent_span_id IS NULL)
type TraceService struct {
	traceRepo observability.TraceRepository
	spanRepo  observability.SpanRepository
	logger    *logrus.Logger
}

// NewTraceService creates a new trace service instance
func NewTraceService(
	traceRepo observability.TraceRepository,
	spanRepo observability.SpanRepository,
	logger *logrus.Logger,
) *TraceService {
	return &TraceService{
		traceRepo: traceRepo,
		spanRepo:  spanRepo,
		logger:    logger,
	}
}

// GetRootSpan retrieves the root span for a trace (OTEL-native: traces = root spans)
func (s *TraceService) GetRootSpan(ctx context.Context, traceID string) (*observability.Span, error) {
	// Validate trace_id format (32 hex chars per OTLP spec)
	if len(traceID) != 32 {
		return nil, appErrors.NewValidationError("invalid trace_id", "OTEL trace_id must be 32 hex characters")
	}

	rootSpan, err := s.traceRepo.GetRootSpan(ctx, traceID)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get root span", err)
	}

	return rootSpan, nil
}

// GetTraceMetrics calculates trace-level aggregations on-demand
func (s *TraceService) GetTraceMetrics(ctx context.Context, traceID string) (*observability.TraceMetrics, error) {
	// Validate trace_id format
	if len(traceID) != 32 {
		return nil, appErrors.NewValidationError("invalid trace_id", "OTEL trace_id must be 32 hex characters")
	}

	metrics, err := s.traceRepo.GetTraceMetrics(ctx, traceID)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get trace metrics", err)
	}

	return metrics, nil
}

// CalculateTraceCost calculates the total cost for a trace
func (s *TraceService) CalculateTraceCost(ctx context.Context, traceID string) (float64, error) {
	// Validate trace_id format
	if len(traceID) != 32 {
		return 0, appErrors.NewValidationError("invalid trace_id", "OTEL trace_id must be 32 hex characters")
	}

	totalCost, err := s.traceRepo.CalculateTotalCost(ctx, traceID)
	if err != nil {
		return 0, appErrors.NewInternalError("failed to calculate trace cost", err)
	}

	return totalCost, nil
}

// ListTraces retrieves traces with filters (queries root spans)
func (s *TraceService) ListTraces(ctx context.Context, filter *observability.TraceFilter) ([]*observability.TraceMetrics, error) {
	// Validate filter
	if filter == nil {
		return nil, appErrors.NewValidationError("filter is required", "trace filter cannot be nil")
	}

	if filter.ProjectID == "" {
		return nil, appErrors.NewValidationError("project_id is required", "filter must include project_id for scoping")
	}

	// Set pagination defaults
	filter.SetDefaults("start_time")

	traces, err := s.traceRepo.ListTraces(ctx, filter)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to list traces", err)
	}

	return traces, nil
}

// GetTraceWithAllSpans retrieves all spans for a trace
func (s *TraceService) GetTraceWithAllSpans(ctx context.Context, traceID string) ([]*observability.Span, error) {
	// Validate trace_id format
	if len(traceID) != 32 {
		return nil, appErrors.NewValidationError("invalid trace_id", "OTEL trace_id must be 32 hex characters")
	}

	spans, err := s.spanRepo.GetByTraceID(ctx, traceID)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get spans by trace", err)
	}

	return spans, nil
}

// GetTracesBySessionID retrieves traces by session ID (virtual session analytics)
func (s *TraceService) GetTracesBySessionID(ctx context.Context, sessionID string) ([]*observability.TraceMetrics, error) {
	if sessionID == "" {
		return nil, appErrors.NewValidationError("session_id is required", "session_id cannot be empty")
	}

	traces, err := s.traceRepo.GetBySessionID(ctx, sessionID)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get traces by session", err)
	}

	return traces, nil
}

// GetTracesByUserID retrieves traces by user ID
func (s *TraceService) GetTracesByUserID(ctx context.Context, userID string, filter *observability.TraceFilter) ([]*observability.TraceMetrics, error) {
	if userID == "" {
		return nil, appErrors.NewValidationError("user_id is required", "user_id cannot be empty")
	}

	traces, err := s.traceRepo.GetByUserID(ctx, userID, filter)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get traces by user", err)
	}

	return traces, nil
}

// CountTraces counts traces matching the filter
func (s *TraceService) CountTraces(ctx context.Context, filter *observability.TraceFilter) (int64, error) {
	if filter == nil {
		return 0, appErrors.NewValidationError("filter is required", "trace filter cannot be nil")
	}

	count, err := s.traceRepo.Count(ctx, filter)
	if err != nil {
		return 0, appErrors.NewInternalError("failed to count traces", err)
	}

	return count, nil
}
