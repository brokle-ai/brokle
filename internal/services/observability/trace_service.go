package observability

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"brokle/internal/core/domain/observability"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/ulid"
)

// TraceService implements business logic for trace management
type TraceService struct {
	traceRepo       observability.TraceRepository
	observationRepo observability.ObservationRepository
	scoreRepo       observability.ScoreRepository
}

// NewTraceService creates a new trace service instance
func NewTraceService(
	traceRepo observability.TraceRepository,
	observationRepo observability.ObservationRepository,
	scoreRepo observability.ScoreRepository,
) *TraceService {
	return &TraceService{
		traceRepo:       traceRepo,
		observationRepo: observationRepo,
		scoreRepo:       scoreRepo,
	}
}

// CreateTrace creates a new trace with validation
func (s *TraceService) CreateTrace(ctx context.Context, trace *observability.Trace) error {
	// Validate required fields
	if trace.ProjectID.IsZero() {
		return appErrors.NewValidationError("project_id is required", "trace must have a valid project_id")
	}
	if trace.Name == "" {
		return appErrors.NewValidationError("name is required", "trace name cannot be empty")
	}

	// Generate new ID if not provided
	if trace.ID.IsZero() {
		trace.ID = ulid.New()
	}

	// Validate parent trace exists if provided
	if trace.ParentTraceID != nil && !trace.ParentTraceID.IsZero() {
		_, err := s.traceRepo.GetByID(ctx, *trace.ParentTraceID)
		if err != nil {
			return appErrors.NewNotFoundError(fmt.Sprintf("parent trace %s", trace.ParentTraceID.String()))
		}
	}

	// Create trace
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
			return appErrors.NewNotFoundError(fmt.Sprintf("trace %s", trace.ID.String()))
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
	if src.UserID != nil && !src.UserID.IsZero() {
		dst.UserID = src.UserID
	}
	if src.SessionID != nil && !src.SessionID.IsZero() {
		dst.SessionID = src.SessionID
	}
	if src.ParentTraceID != nil && !src.ParentTraceID.IsZero() {
		dst.ParentTraceID = src.ParentTraceID
	}
	if src.Input != nil {
		dst.Input = src.Input
	}
	if src.Output != nil {
		dst.Output = src.Output
	}
	// Allow clearing metadata/tags by sending empty map/slice
	// nil = not sent (preserve), {} or [] = clear, {...} = update
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
	if !src.Timestamp.IsZero() {
		dst.Timestamp = src.Timestamp
	}
}

// DeleteTrace soft deletes a trace
func (s *TraceService) DeleteTrace(ctx context.Context, id ulid.ULID) error {
	// Validate trace exists
	_, err := s.traceRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appErrors.NewNotFoundError(fmt.Sprintf("trace %s", id.String()))
		}
		return appErrors.NewInternalError("failed to get trace", err)
	}

	// Delete trace
	if err := s.traceRepo.Delete(ctx, id); err != nil {
		return appErrors.NewInternalError("failed to delete trace", err)
	}

	return nil
}

// GetTraceByID retrieves a trace by ID
func (s *TraceService) GetTraceByID(ctx context.Context, id ulid.ULID) (*observability.Trace, error) {
	trace, err := s.traceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, appErrors.NewNotFoundError(fmt.Sprintf("trace %s", id.String()))
	}

	return trace, nil
}

// GetTraceWithObservations retrieves a trace with all its observations in hierarchical tree structure
func (s *TraceService) GetTraceWithObservations(ctx context.Context, id ulid.ULID) (*observability.Trace, error) {
	// Get trace
	trace, err := s.traceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, appErrors.NewNotFoundError(fmt.Sprintf("trace %s", id.String()))
	}

	// Get observations tree
	observations, err := s.observationRepo.GetTreeByTraceID(ctx, id)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get observations", err)
	}

	trace.Observations = observations

	return trace, nil
}

// GetTraceWithScores retrieves a trace with all its quality scores
func (s *TraceService) GetTraceWithScores(ctx context.Context, id ulid.ULID) (*observability.Trace, error) {
	// Get trace
	trace, err := s.traceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, appErrors.NewNotFoundError(fmt.Sprintf("trace %s", id.String()))
	}

	// Get scores
	scores, err := s.scoreRepo.GetByTraceID(ctx, id)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get scores", err)
	}

	trace.Scores = scores

	return trace, nil
}

// GetTracesByProjectID retrieves traces for a project with optional filtering
func (s *TraceService) GetTracesByProjectID(ctx context.Context, projectID ulid.ULID, filter *observability.TraceFilter) ([]*observability.Trace, error) {
	if projectID.IsZero() {
		return nil, appErrors.NewValidationError("project_id is required", "traces query requires a valid project_id")
	}

	traces, err := s.traceRepo.GetByProjectID(ctx, projectID, filter)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get traces", err)
	}

	return traces, nil
}

// GetTracesBySessionID retrieves all traces in a session
func (s *TraceService) GetTracesBySessionID(ctx context.Context, sessionID ulid.ULID) ([]*observability.Trace, error) {
	if sessionID.IsZero() {
		return nil, appErrors.NewValidationError("session_id is required", "traces query requires a valid session_id")
	}

	traces, err := s.traceRepo.GetBySessionID(ctx, sessionID)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get traces", err)
	}

	return traces, nil
}

// GetChildTraces retrieves child traces of a parent trace
func (s *TraceService) GetChildTraces(ctx context.Context, parentTraceID ulid.ULID) ([]*observability.Trace, error) {
	if parentTraceID.IsZero() {
		return nil, appErrors.NewValidationError("parent_trace_id is required", "parent_trace_id cannot be empty")
	}

	// Validate parent exists
	_, err := s.traceRepo.GetByID(ctx, parentTraceID)
	if err != nil {
		return nil, appErrors.NewNotFoundError(fmt.Sprintf("parent trace %s", parentTraceID.String()))
	}

	traces, err := s.traceRepo.GetChildren(ctx, parentTraceID)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get child traces", err)
	}

	return traces, nil
}

// GetTracesByUserID retrieves traces for a user with optional filtering
func (s *TraceService) GetTracesByUserID(ctx context.Context, userID ulid.ULID, filter *observability.TraceFilter) ([]*observability.Trace, error) {
	if userID.IsZero() {
		return nil, appErrors.NewValidationError("user_id is required", "traces query requires a valid user_id")
	}

	traces, err := s.traceRepo.GetByUserID(ctx, userID, filter)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get traces", err)
	}

	return traces, nil
}

// CreateTraceBatch creates multiple traces in a single batch operation
func (s *TraceService) CreateTraceBatch(ctx context.Context, traces []*observability.Trace) error {
	if len(traces) == 0 {
		return appErrors.NewValidationError("traces array cannot be empty", "batch create requires at least one trace")
	}

	// Validate all traces
	for i, trace := range traces {
		if trace.ProjectID.IsZero() {
			return appErrors.NewValidationError(
				fmt.Sprintf("trace[%d]: project_id is required", i),
				"all traces must have valid project_id",
			)
		}
		if trace.Name == "" {
			return appErrors.NewValidationError(
				fmt.Sprintf("trace[%d]: name is required", i),
				"all traces must have a name",
			)
		}

		// Generate ID if not provided
		if trace.ID.IsZero() {
			trace.ID = ulid.New()
		}
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
