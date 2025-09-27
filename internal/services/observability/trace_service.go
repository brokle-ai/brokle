package observability

import (
	"context"
	"fmt"
	"time"

	"brokle/internal/core/domain/observability"
	"brokle/pkg/ulid"
)

// traceService implements the TraceService interface
type traceService struct {
	traceRepo       observability.TraceRepository
	observationRepo observability.ObservationRepository
	eventPublisher  observability.EventPublisher
}

// NewTraceService creates a new trace service
func NewTraceService(
	traceRepo observability.TraceRepository,
	observationRepo observability.ObservationRepository,
	eventPublisher observability.EventPublisher,
) observability.TraceService {
	return &traceService{
		traceRepo:       traceRepo,
		observationRepo: observationRepo,
		eventPublisher:  eventPublisher,
	}
}

// CreateTrace creates a new trace
func (s *traceService) CreateTrace(ctx context.Context, trace *observability.Trace) (*observability.Trace, error) {
	if trace == nil {
		return nil, observability.NewObservabilityError(
			observability.ErrCodeValidationFailed,
			"trace cannot be nil",
		)
	}

	// Generate ID if not provided
	if trace.ID.IsZero() {
		trace.ID = ulid.New()
	}

	// Validate required fields
	if err := s.validateTrace(trace); err != nil {
		return nil, err
	}

	// Set timestamps
	now := time.Now()
	if trace.CreatedAt.IsZero() {
		trace.CreatedAt = now
	}
	trace.UpdatedAt = now

	// Create trace in repository
	if err := s.traceRepo.Create(ctx, trace); err != nil {
		return nil, fmt.Errorf("failed to create trace: %w", err)
	}

	// Publish trace created event
	event := observability.NewTraceCreatedEvent(trace, trace.UserID)
	if publishErr := s.eventPublisher.Publish(ctx, event); publishErr != nil {
		// Log error but don't fail the operation
		// In production, this might be logged or sent to a monitoring system
		_ = publishErr
	}

	return trace, nil
}

// CreateTraceWithObservations creates a trace with its initial observations
func (s *traceService) CreateTraceWithObservations(ctx context.Context, trace *observability.Trace) (*observability.Trace, error) {
	if trace == nil {
		return nil, observability.NewObservabilityError(
			observability.ErrCodeValidationFailed,
			"trace cannot be nil",
		)
	}

	// Create the trace first
	createdTrace, err := s.CreateTrace(ctx, trace)
	if err != nil {
		return nil, err
	}

	// Create associated observations if any
	if len(trace.Observations) > 0 {
		for _, obs := range trace.Observations {
			obs.TraceID = createdTrace.ID
			if obs.ID.IsZero() {
				obs.ID = ulid.New()
			}
		}

		var obsPointers []*observability.Observation
		for i := range trace.Observations {
			obsPointers = append(obsPointers, &trace.Observations[i])
		}

		err := s.observationRepo.CreateBatch(ctx, obsPointers)
		if err != nil {
			return nil, fmt.Errorf("failed to create observations: %w", err)
		}
		// The observations are updated in place, so we don't need to reassign
	}

	return createdTrace, nil
}

// GetTrace retrieves a trace by ID
func (s *traceService) GetTrace(ctx context.Context, id ulid.ULID) (*observability.Trace, error) {
	if id.IsZero() {
		return nil, observability.NewObservabilityError(
			observability.ErrCodeInvalidTraceID,
			"trace ID cannot be empty",
		)
	}

	trace, err := s.traceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get trace: %w", err)
	}

	return trace, nil
}

// GetTraceByExternalID retrieves a trace by external ID
func (s *traceService) GetTraceByExternalID(ctx context.Context, externalTraceID string) (*observability.Trace, error) {
	if externalTraceID == "" {
		return nil, observability.NewObservabilityError(
			observability.ErrCodeValidationFailed,
			"external trace ID cannot be empty",
		)
	}

	trace, err := s.traceRepo.GetByExternalTraceID(ctx, externalTraceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get trace by external ID: %w", err)
	}

	return trace, nil
}

// UpdateTrace updates an existing trace
func (s *traceService) UpdateTrace(ctx context.Context, trace *observability.Trace) (*observability.Trace, error) {
	if trace == nil {
		return nil, observability.NewObservabilityError(
			observability.ErrCodeValidationFailed,
			"trace cannot be nil",
		)
	}

	if trace.ID.IsZero() {
		return nil, observability.NewObservabilityError(
			observability.ErrCodeInvalidTraceID,
			"trace ID cannot be empty",
		)
	}

	// Validate trace data
	if err := s.validateTrace(trace); err != nil {
		return nil, err
	}

	// Get original trace for change detection
	originalTrace, err := s.traceRepo.GetByID(ctx, trace.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get original trace: %w", err)
	}

	// Update timestamp
	trace.UpdatedAt = time.Now()

	// Update in repository
	if err := s.traceRepo.Update(ctx, trace); err != nil {
		return nil, fmt.Errorf("failed to update trace: %w", err)
	}

	// Detect changes and publish event
	changes := s.detectTraceChanges(originalTrace, trace)
	if len(changes) > 0 {
		event := observability.NewTraceUpdatedEvent(trace, trace.UserID, changes)
		if publishErr := s.eventPublisher.Publish(ctx, event); publishErr != nil {
			_ = publishErr
		}
	}

	return trace, nil
}

// DeleteTrace deletes a trace by ID
func (s *traceService) DeleteTrace(ctx context.Context, id ulid.ULID) error {
	if id.IsZero() {
		return observability.NewObservabilityError(
			observability.ErrCodeInvalidTraceID,
			"trace ID cannot be empty",
		)
	}

	// Get trace for event publishing
	trace, err := s.traceRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get trace for deletion: %w", err)
	}

	// Delete from repository
	if err := s.traceRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete trace: %w", err)
	}

	// Publish deletion event
	event := observability.NewTraceDeletedEvent(id, trace.ProjectID, trace.UserID)
	if publishErr := s.eventPublisher.Publish(ctx, event); publishErr != nil {
		_ = publishErr
	}

	return nil
}

// ListTraces retrieves traces based on filter criteria
func (s *traceService) ListTraces(ctx context.Context, filter *observability.TraceFilter) ([]*observability.Trace, int, error) {
	if filter == nil {
		filter = &observability.TraceFilter{}
	}

	// Set default pagination if not provided
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Limit > 1000 {
		filter.Limit = 1000
	}

	traces, total, err := s.traceRepo.SearchTraces(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list traces: %w", err)
	}

	return traces, total, nil
}

// GetTraceWithObservations retrieves a trace with all its observations
func (s *traceService) GetTraceWithObservations(ctx context.Context, id ulid.ULID) (*observability.Trace, error) {
	trace, err := s.GetTrace(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get all observations for this trace
	observations, err := s.observationRepo.GetByTraceID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get trace observations: %w", err)
	}

	// Convert pointers to values for the trace
	var obsValues []observability.Observation
	for _, obs := range observations {
		obsValues = append(obsValues, *obs)
	}
	trace.Observations = obsValues
	return trace, nil
}

// GetTraceStats retrieves statistics for a trace
func (s *traceService) GetTraceStats(ctx context.Context, id ulid.ULID) (*observability.TraceStats, error) {
	// This would typically aggregate statistics from the observations
	// For now, return a basic implementation
	trace, err := s.GetTraceWithObservations(ctx, id)
	if err != nil {
		return nil, err
	}

	stats := &observability.TraceStats{
		TraceID:     id,
		TotalCost:   0,
		TotalTokens: 0,
	}

	// Calculate aggregated stats from observations
	if len(trace.Observations) > 0 {
		var startTime, endTime *time.Time

		for _, obs := range trace.Observations {
			// Aggregate costs
			if obs.TotalCost != nil {
				stats.TotalCost += *obs.TotalCost
			}

			// Aggregate tokens
			stats.TotalTokens += obs.TotalTokens

			// Track time bounds
			if startTime == nil || obs.StartTime.Before(*startTime) {
				startTime = &obs.StartTime
			}
			if obs.EndTime != nil && (endTime == nil || obs.EndTime.After(*endTime)) {
				endTime = obs.EndTime
			}
		}

		// Calculate duration if we have both start and end times
		if startTime != nil && endTime != nil {
			// Duration would be calculated here if needed in stats
		}
	}

	return stats, nil
}

// GetRecentTraces retrieves recent traces for a project
func (s *traceService) GetRecentTraces(ctx context.Context, projectID ulid.ULID, limit int) ([]*observability.Trace, error) {
	if projectID.IsZero() {
		return nil, observability.NewObservabilityError(
			observability.ErrCodeInvalidProjectID,
			"project ID cannot be empty",
		)
	}

	if limit <= 0 || limit > 1000 {
		limit = 50
	}

	traces, err := s.traceRepo.GetRecentTraces(ctx, projectID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent traces: %w", err)
	}

	return traces, nil
}

// CreateTracesBatch creates multiple traces in a batch operation
func (s *traceService) CreateTracesBatch(ctx context.Context, traces []*observability.Trace) ([]*observability.Trace, error) {
	if len(traces) == 0 {
		return []*observability.Trace{}, nil
	}

	// Validate and prepare traces
	now := time.Now()
	for _, trace := range traces {
		if trace.ID.IsZero() {
			trace.ID = ulid.New()
		}

		if err := s.validateTrace(trace); err != nil {
			return nil, err
		}

		if trace.CreatedAt.IsZero() {
			trace.CreatedAt = now
		}
		trace.UpdatedAt = now
	}

	// Create batch in repository
	if err := s.traceRepo.CreateBatch(ctx, traces); err != nil {
		return nil, fmt.Errorf("failed to create traces batch: %w", err)
	}

	// Publish events for created traces
	var events []*observability.Event
	for _, trace := range traces {
		event := observability.NewTraceCreatedEvent(trace, trace.UserID)
		events = append(events, event)
	}

	if len(events) > 0 {
		if publishErr := s.eventPublisher.PublishBatch(ctx, events); publishErr != nil {
			_ = publishErr
		}
	}

	return traces, nil
}

// IngestTraceBatch ingests a batch of traces (implementation placeholder)
func (s *traceService) IngestTraceBatch(ctx context.Context, request *observability.BatchIngestRequest) (*observability.BatchIngestResult, error) {
	// This would implement high-throughput ingestion logic
	// For now, delegate to CreateTracesBatch
	startTime := time.Now()

	createdTraces, err := s.CreateTracesBatch(ctx, request.Traces)
	if err != nil {
		return &observability.BatchIngestResult{
			ProcessedCount: 0,
			FailedCount:    len(request.Traces),
			Duration:       time.Since(startTime),
			Errors: []observability.BatchIngestionError{
				{
					Index:   0,
					Error:   err.Error(),
					Details: nil,
				},
			},
		}, err
	}

	return &observability.BatchIngestResult{
		ProcessedCount: len(createdTraces),
		FailedCount:    0,
		Duration:       time.Since(startTime),
		Errors:         []observability.BatchIngestionError{},
	}, nil
}

// SearchTraces searches traces by query string (implementation placeholder)
func (s *traceService) SearchTraces(ctx context.Context, query string, filter *observability.TraceFilter) ([]*observability.Trace, int, error) {
	// This would implement full-text search logic
	// For now, delegate to ListTraces
	return s.ListTraces(ctx, filter)
}

// GetTracesByTimeRange retrieves traces within a time range
func (s *traceService) GetTracesByTimeRange(ctx context.Context, projectID ulid.ULID, startTime, endTime time.Time, limit, offset int) ([]*observability.Trace, error) {
	if projectID.IsZero() {
		return nil, observability.NewObservabilityError(
			observability.ErrCodeInvalidProjectID,
			"project ID cannot be empty",
		)
	}

	traces, err := s.traceRepo.GetTracesByTimeRange(ctx, projectID, startTime, endTime, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get traces by time range: %w", err)
	}

	return traces, nil
}

// GetTraceAnalytics retrieves analytics data for traces (implementation placeholder)
func (s *traceService) GetTraceAnalytics(ctx context.Context, filter *observability.AnalyticsFilter) (*observability.TraceAnalytics, error) {
	// This would implement analytics aggregation logic
	// Return placeholder for now
	return &observability.TraceAnalytics{
		TotalTraces: 0,
	}, nil
}

// Helper methods

// validateTrace validates a trace object
func (s *traceService) validateTrace(trace *observability.Trace) error {
	if trace.ProjectID.IsZero() {
		return observability.NewValidationError("project_id", "project ID is required")
	}

	if trace.ExternalTraceID == "" {
		return observability.NewValidationError("external_trace_id", "external trace ID is required")
	}

	if trace.Name == "" {
		return observability.NewValidationError("name", "trace name is required")
	}

	return nil
}

// detectTraceChanges compares two traces and returns a map of changes
func (s *traceService) detectTraceChanges(original, updated *observability.Trace) map[string]any {
	changes := make(map[string]any)

	if original.Name != updated.Name {
		changes["name"] = map[string]string{
			"from": original.Name,
			"to":   updated.Name,
		}
	}

	if (original.UserID == nil) != (updated.UserID == nil) ||
		(original.UserID != nil && updated.UserID != nil && *original.UserID != *updated.UserID) {
		changes["user_id"] = map[string]any{
			"from": original.UserID,
			"to":   updated.UserID,
		}
	}

	// Add more change detection logic as needed

	return changes
}