package observability

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"brokle/internal/core/domain/observability"
	appErrors "brokle/pkg/errors"
)

// TraceService implements business logic for trace management
type TraceService struct {
	traceRepo       observability.TraceRepository
	observationRepo observability.ObservationRepository
	scoreRepo       observability.ScoreRepository
	logger          *logrus.Logger
}

// NewTraceService creates a new trace service instance
func NewTraceService(
	traceRepo observability.TraceRepository,
	observationRepo observability.ObservationRepository,
	scoreRepo observability.ScoreRepository,
	logger *logrus.Logger,
) *TraceService {
	return &TraceService{
		traceRepo:       traceRepo,
		observationRepo: observationRepo,
		scoreRepo:       scoreRepo,
		logger:          logger,
	}
}

// CreateTrace creates a new OTEL trace with validation
func (s *TraceService) CreateTrace(ctx context.Context, trace *observability.Trace) error {
	// Validate required fields
	if trace.ProjectID == "" {
		return appErrors.NewValidationError("project_id is required", "trace must have a valid project_id")
	}
	if trace.Name == "" {
		return appErrors.NewValidationError("name is required", "trace name cannot be empty")
	}
	if trace.ID == "" {
		return appErrors.NewValidationError("id is required", "trace must have OTEL trace_id")
	}

	// Validate OTEL trace_id format (32 hex chars)
	if len(trace.ID) != 32 {
		return appErrors.NewValidationError("invalid trace_id", "OTEL trace_id must be 32 hex characters")
	}

	// Set defaults
	if trace.StatusCode == "" {
		trace.StatusCode = observability.StatusCodeUnset
	}
	if trace.Environment == "" {
		trace.Environment = "production"
	}
	if trace.Attributes == "" {
		trace.Attributes = "{}"
	}
	if trace.CreatedAt.IsZero() {
		trace.CreatedAt = time.Now()
	}

	// Calculate duration if not set
	trace.CalculateDuration()

	// Store trace directly in ClickHouse (ZSTD compression handles all sizes)
	if err := s.traceRepo.Create(ctx, trace); err != nil {
		return appErrors.NewInternalError("failed to create trace", err)
	}

	return nil
}

// UpdateTrace updates an existing trace
func (s *TraceService) UpdateTrace(ctx context.Context, trace *observability.Trace) error {
	// Validate trace exists
	existing, err := s.traceRepo.GetByID(ctx, trace.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appErrors.NewNotFoundError(fmt.Sprintf("trace %s", trace.ID))
		}
		return appErrors.NewInternalError("failed to get trace", err)
	}

	// Merge non-zero fields from incoming trace into existing
	mergeTraceFields(existing, trace)

	// Preserve version for increment in repository layer
	existing.Version = existing.Version

	// Update trace
	if err := s.traceRepo.Update(ctx, existing); err != nil {
		return appErrors.NewInternalError("failed to update trace", err)
	}

	return nil
}

// UpdateTraceMetrics updates aggregate metrics for a trace (called after observation changes)
func (s *TraceService) UpdateTraceMetrics(ctx context.Context, traceID string, totalCost float64, totalTokens, observationCount uint32) error {
	// Get existing trace
	trace, err := s.traceRepo.GetByID(ctx, traceID)
	if err != nil {
		return appErrors.NewNotFoundError(fmt.Sprintf("trace %s", traceID))
	}

	// Update aggregate metrics
	trace.TotalCost = &totalCost
	trace.TotalTokens = &totalTokens
	trace.ObservationCount = &observationCount

	// Update trace
	if err := s.traceRepo.Update(ctx, trace); err != nil {
		return appErrors.NewInternalError("failed to update trace metrics", err)
	}

	return nil
}

// mergeTraceFields merges non-zero fields from src into dst
// This prevents zero-value corruption from partial JSON updates
func mergeTraceFields(dst *observability.Trace, src *observability.Trace) {
	// Immutable fields (never update):
	// - ID (primary key)
	// - ProjectID (security boundary)
	// - Version (managed by repository)
	// - EventTs (managed by repository)
	// - IsDeleted (managed by Delete method)

	// Update optional fields only if non-zero
	if src.Name != "" {
		dst.Name = src.Name
	}
	if src.UserID != nil && *src.UserID != "" {
		dst.UserID = src.UserID
	}
	if src.SessionID != nil && *src.SessionID != "" {
		dst.SessionID = src.SessionID
	}
	if src.Input != nil {
		dst.Input = src.Input
	}
	if src.Output != nil {
		dst.Output = src.Output
	}
	if src.Metadata != nil {
		dst.Metadata = src.Metadata
	}
	if src.Tags != nil {
		dst.Tags = src.Tags
	}
	if src.Environment != "" {
		dst.Environment = src.Environment
	}
	if src.Release != nil {
		dst.Release = src.Release
	}
	if !src.StartTime.IsZero() {
		dst.StartTime = src.StartTime
	}
	if src.EndTime != nil {
		dst.EndTime = src.EndTime
		// Recalculate duration when end time is updated
		dst.CalculateDuration()
	}
	if src.StatusCode != "" {
		dst.StatusCode = src.StatusCode
	}
	if src.StatusMessage != nil {
		dst.StatusMessage = src.StatusMessage
	}
	if src.Attributes != "" {
		dst.Attributes = src.Attributes
	}
	if src.ServiceName != nil {
		dst.ServiceName = src.ServiceName
	}
	if src.ServiceVersion != nil {
		dst.ServiceVersion = src.ServiceVersion
	}
	if src.TotalCost != nil {
		dst.TotalCost = src.TotalCost
	}
	if src.TotalTokens != nil {
		dst.TotalTokens = src.TotalTokens
	}
	if src.ObservationCount != nil {
		dst.ObservationCount = src.ObservationCount
	}
	// Bookmarked and Public are bool, so always update
	dst.Bookmarked = src.Bookmarked
	dst.Public = src.Public
}

// DeleteTrace soft deletes a trace
func (s *TraceService) DeleteTrace(ctx context.Context, id string) error {
	// Validate trace exists
	_, err := s.traceRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appErrors.NewNotFoundError(fmt.Sprintf("trace %s", id))
		}
		return appErrors.NewInternalError("failed to get trace", err)
	}

	// Delete trace
	if err := s.traceRepo.Delete(ctx, id); err != nil {
		return appErrors.NewInternalError("failed to delete trace", err)
	}

	return nil
}

// GetTraceByID retrieves a trace by its OTEL trace_id
func (s *TraceService) GetTraceByID(ctx context.Context, id string) (*observability.Trace, error) {
	trace, err := s.traceRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, appErrors.NewNotFoundError(fmt.Sprintf("trace %s", id))
		}
		return nil, appErrors.NewInternalError("failed to get trace", err)
	}

	return trace, nil
}

// GetTraceWithObservations retrieves a trace with all its observations
func (s *TraceService) GetTraceWithObservations(ctx context.Context, id string) (*observability.Trace, error) {
	// Get trace
	trace, err := s.traceRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, appErrors.NewNotFoundError(fmt.Sprintf("trace %s", id))
		}
		return nil, appErrors.NewInternalError("failed to get trace", err)
	}

	// Get observations
	observations, err := s.observationRepo.GetByTraceID(ctx, id)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get observations", err)
	}

	trace.Observations = observations

	return trace, nil
}

// GetTraceWithScores retrieves a trace with all its scores
func (s *TraceService) GetTraceWithScores(ctx context.Context, id string) (*observability.Trace, error) {
	// Get trace
	trace, err := s.traceRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, appErrors.NewNotFoundError(fmt.Sprintf("trace %s", id))
		}
		return nil, appErrors.NewInternalError("failed to get trace", err)
	}

	// Get scores
	scores, err := s.scoreRepo.GetByTraceID(ctx, id)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get scores", err)
	}

	trace.Scores = scores

	return trace, nil
}

// GetTracesByProjectID retrieves traces by project ID with optional filters
func (s *TraceService) GetTracesByProjectID(ctx context.Context, projectID string, filter *observability.TraceFilter) ([]*observability.Trace, error) {
	traces, err := s.traceRepo.GetByProjectID(ctx, projectID, filter)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get traces", err)
	}

	return traces, nil
}

// GetTracesBySessionID retrieves all traces in a virtual session
func (s *TraceService) GetTracesBySessionID(ctx context.Context, sessionID string) ([]*observability.Trace, error) {
	traces, err := s.traceRepo.GetBySessionID(ctx, sessionID)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get traces by session", err)
	}

	return traces, nil
}

// GetTracesByUserID retrieves traces by user ID
func (s *TraceService) GetTracesByUserID(ctx context.Context, userID string, filter *observability.TraceFilter) ([]*observability.Trace, error) {
	traces, err := s.traceRepo.GetByUserID(ctx, userID, filter)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get traces by user", err)
	}

	return traces, nil
}

// CreateTraceBatch creates multiple traces in a batch
func (s *TraceService) CreateTraceBatch(ctx context.Context, traces []*observability.Trace) error {
	if len(traces) == 0 {
		return nil
	}

	// Validate all traces
	for i, trace := range traces {
		if trace.ProjectID == "" {
			return appErrors.NewValidationError(fmt.Sprintf("trace[%d].project_id", i), "project_id is required")
		}
		if trace.Name == "" {
			return appErrors.NewValidationError(fmt.Sprintf("trace[%d].name", i), "name is required")
		}
		if trace.ID == "" {
			return appErrors.NewValidationError(fmt.Sprintf("trace[%d].id", i), "OTEL trace_id is required")
		}

		// Set defaults
		if trace.StatusCode == "" {
			trace.StatusCode = observability.StatusCodeUnset
		}
		if trace.Environment == "" {
			trace.Environment = "production"
		}
		if trace.Attributes == "" {
			trace.Attributes = "{}"
		}
		if trace.CreatedAt.IsZero() {
			trace.CreatedAt = time.Now()
		}

		// Calculate duration
		trace.CalculateDuration()
	}

	// Create batch
	if err := s.traceRepo.CreateBatch(ctx, traces); err != nil {
		return appErrors.NewInternalError("failed to create trace batch", err)
	}

	return nil
}

// CountTraces returns the count of traces matching the filter
func (s *TraceService) CountTraces(ctx context.Context, filter *observability.TraceFilter) (int64, error) {
	count, err := s.traceRepo.Count(ctx, filter)
	if err != nil {
		return 0, appErrors.NewInternalError("failed to count traces", err)
	}

	return count, nil
}
