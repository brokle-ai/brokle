package observability

import (
	"context"
	"fmt"
	"time"

	"brokle/internal/core/domain/observability"
	"brokle/pkg/ulid"
)

// telemetryEventService implements the TelemetryEventService interface
type telemetryEventService struct {
	eventRepo observability.TelemetryEventRepository
	batchRepo observability.TelemetryBatchRepository
}

// NewTelemetryEventService creates a new telemetry event service
func NewTelemetryEventService(
	eventRepo observability.TelemetryEventRepository,
	batchRepo observability.TelemetryBatchRepository,
) observability.TelemetryEventService {
	return &telemetryEventService{
		eventRepo: eventRepo,
		batchRepo: batchRepo,
	}
}

// CreateEvent creates a new telemetry event with validation
func (s *telemetryEventService) CreateEvent(ctx context.Context, event *observability.TelemetryEvent) (*observability.TelemetryEvent, error) {
	if event == nil {
		return nil, fmt.Errorf("event cannot be nil")
	}

	// Generate ID if not provided
	if event.ID.IsZero() {
		event.ID = ulid.New()
	}

	// Validate event
	if validationErrors := event.Validate(); len(validationErrors) > 0 {
		return nil, fmt.Errorf("event validation failed: %v", validationErrors)
	}

	// Set creation timestamp
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}

	// Create event in repository
	if err := s.eventRepo.Create(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	return event, nil
}

// CreateEventsBatch creates multiple telemetry events using efficient bulk operations
func (s *telemetryEventService) CreateEventsBatch(ctx context.Context, events []*observability.TelemetryEvent) error {
	if len(events) == 0 {
		return nil
	}

	// Validate all events before storing
	for i, event := range events {
		if event == nil {
			return fmt.Errorf("event at index %d cannot be nil", i)
		}

		// Generate ID if not provided
		if event.ID.IsZero() {
			event.ID = ulid.New()
		}

		// Validate event
		if validationErrors := event.Validate(); len(validationErrors) > 0 {
			return fmt.Errorf("event validation failed for event %s: %v", event.ID.String(), validationErrors)
		}

		// Set creation timestamp
		if event.CreatedAt.IsZero() {
			event.CreatedAt = time.Now()
		}
	}

	// Use repository's optimized bulk create
	if err := s.eventRepo.CreateBatch(ctx, events); err != nil {
		return fmt.Errorf("failed to create events batch: %w", err)
	}

	return nil
}

// UpdateEventsBatch updates multiple telemetry events using efficient bulk operations
func (s *telemetryEventService) UpdateEventsBatch(ctx context.Context, events []*observability.TelemetryEvent) error {
	if len(events) == 0 {
		return nil
	}

	// Validate all events before updating
	for i, event := range events {
		if event == nil {
			return fmt.Errorf("event at index %d cannot be nil", i)
		}
		if event.ID.IsZero() {
			return fmt.Errorf("event at index %d must have a valid ID for update", i)
		}

		// Validate event
		if validationErrors := event.Validate(); len(validationErrors) > 0 {
			return fmt.Errorf("event validation failed for event %s: %v", event.ID.String(), validationErrors)
		}
	}

	// Use repository's optimized bulk update
	if err := s.eventRepo.UpdateBatch(ctx, events); err != nil {
		return fmt.Errorf("failed to update events batch: %w", err)
	}

	return nil
}

// GetEvent retrieves a telemetry event by ID
func (s *telemetryEventService) GetEvent(ctx context.Context, id ulid.ULID) (*observability.TelemetryEvent, error) {
	if id.IsZero() {
		return nil, fmt.Errorf("event ID cannot be zero")
	}

	event, err := s.eventRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	return event, nil
}

// UpdateEvent updates an existing telemetry event
func (s *telemetryEventService) UpdateEvent(ctx context.Context, event *observability.TelemetryEvent) (*observability.TelemetryEvent, error) {
	if event == nil {
		return nil, fmt.Errorf("event cannot be nil")
	}

	if event.ID.IsZero() {
		return nil, fmt.Errorf("event ID cannot be zero")
	}

	// Validate event
	if validationErrors := event.Validate(); len(validationErrors) > 0 {
		return nil, fmt.Errorf("event validation failed: %v", validationErrors)
	}

	// Update event in repository
	if err := s.eventRepo.Update(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to update event: %w", err)
	}

	return event, nil
}

// DeleteEvent deletes a telemetry event
func (s *telemetryEventService) DeleteEvent(ctx context.Context, id ulid.ULID) error {
	if id.IsZero() {
		return fmt.Errorf("event ID cannot be zero")
	}

	if err := s.eventRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}

	return nil
}

// ListEvents retrieves telemetry events with filtering
func (s *telemetryEventService) ListEvents(ctx context.Context, filter *observability.TelemetryEventFilter) ([]*observability.TelemetryEvent, int, error) {
	if filter == nil {
		filter = &observability.TelemetryEventFilter{
			Limit: 50,
		}
	}

	// Apply default limits
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Limit > 1000 {
		filter.Limit = 1000
	}

	// Get total count
	totalCount, err := s.eventRepo.CountEvents(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count events: %w", err)
	}

	// For listing, we need to implement SearchEvents method in repository
	// For now, we'll use a simpler approach
	var events []*observability.TelemetryEvent

	if filter.BatchID != nil {
		events, err = s.eventRepo.GetByBatchID(ctx, *filter.BatchID)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get events by batch: %w", err)
		}
	} else if filter.EventType != nil {
		events, err = s.eventRepo.GetByEventType(ctx, *filter.EventType, filter.Limit, filter.Offset)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get events by type: %w", err)
		}
	} else {
		// This would require a generic search method in the repository
		return nil, 0, fmt.Errorf("search method not implemented for complex filters")
	}

	return events, int(totalCount), nil
}

// GetEventsByBatch retrieves all events for a specific batch
func (s *telemetryEventService) GetEventsByBatch(ctx context.Context, batchID ulid.ULID) ([]*observability.TelemetryEvent, error) {
	if batchID.IsZero() {
		return nil, fmt.Errorf("batch ID cannot be zero")
	}

	events, err := s.eventRepo.GetByBatchID(ctx, batchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get events by batch: %w", err)
	}

	return events, nil
}

// GetUnprocessedEvents retrieves unprocessed events for a batch
func (s *telemetryEventService) GetUnprocessedEvents(ctx context.Context, batchID ulid.ULID) ([]*observability.TelemetryEvent, error) {
	if batchID.IsZero() {
		return nil, fmt.Errorf("batch ID cannot be zero")
	}

	events, err := s.eventRepo.GetUnprocessedByBatchID(ctx, batchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unprocessed events: %w", err)
	}

	return events, nil
}

// GetFailedEvents retrieves failed events with optional batch filtering
func (s *telemetryEventService) GetFailedEvents(ctx context.Context, batchID *ulid.ULID, limit, offset int) ([]*observability.TelemetryEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}
	if offset < 0 {
		offset = 0
	}

	events, err := s.eventRepo.GetFailedEvents(ctx, batchID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get failed events: %w", err)
	}

	return events, nil
}

// ProcessEvent processes a single telemetry event
func (s *telemetryEventService) ProcessEvent(ctx context.Context, eventID ulid.ULID) error {
	if eventID.IsZero() {
		return fmt.Errorf("event ID cannot be zero")
	}

	// Get the event
	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return fmt.Errorf("failed to get event: %w", err)
	}

	// Check if already processed
	if event.IsProcessed() {
		return fmt.Errorf("event %s is already processed", eventID.String())
	}

	// Process the event based on its type
	if err := s.processEventByType(ctx, event); err != nil {
		// Mark as failed and increment retry count
		if markErr := s.eventRepo.MarkAsFailed(ctx, eventID, err.Error()); markErr != nil {
			return fmt.Errorf("failed to mark event as failed: %w", markErr)
		}
		if retryErr := s.eventRepo.IncrementRetryCount(ctx, eventID); retryErr != nil {
			return fmt.Errorf("failed to increment retry count: %w", retryErr)
		}
		return fmt.Errorf("failed to process event: %w", err)
	}

	// Mark as processed
	if err := s.eventRepo.MarkAsProcessed(ctx, eventID, time.Now()); err != nil {
		return fmt.Errorf("failed to mark event as processed: %w", err)
	}

	return nil
}

// ProcessEventsBatch processes multiple events in a batch
func (s *telemetryEventService) ProcessEventsBatch(ctx context.Context, events []*observability.TelemetryEvent) (*observability.EventProcessingResult, error) {
	if len(events) == 0 {
		return &observability.EventProcessingResult{
			ProcessedCount:    0,
			FailedCount:       0,
			NotProcessedCount: 0,
			ProcessedEventIDs: []ulid.ULID{},
			NotProcessedIDs:   []ulid.ULID{},
			SuccessRate:       100.0,
		}, nil
	}

	startTime := time.Now()
	var processedCount, failedCount, notProcessedCount, retryCount int
	var errors []observability.TelemetryEventError
	var processedEventIDs []ulid.ULID
	var notProcessedIDs []ulid.ULID

	for _, event := range events {
		if event == nil {
			notProcessedCount++
			continue
		}

		// Attempt to process the event
		if err := s.processEventByType(ctx, event); err != nil {
			failedCount++
			retryCount++

			errors = append(errors, observability.TelemetryEventError{
				EventID:      event.ID,
				EventType:    event.EventType,
				ErrorCode:    "PROCESSING_FAILED",
				ErrorMessage: err.Error(),
				Retryable:    true,
			})

			// Mark event as failed and increment retry count
			_ = s.eventRepo.MarkAsFailed(ctx, event.ID, err.Error())
			_ = s.eventRepo.IncrementRetryCount(ctx, event.ID)
		} else {
			processedCount++
			processedEventIDs = append(processedEventIDs, event.ID)
			// Mark event as processed
			_ = s.eventRepo.MarkAsProcessed(ctx, event.ID, time.Now())
		}
	}

	// Identify events that were never attempted (nil events or validation failures)
	for _, event := range events {
		if event == nil {
			continue // Can't track ID for nil events
		}

		// Check if this event ID is in processed list or error list
		isProcessed := false
		isFailed := false

		for _, processedID := range processedEventIDs {
			if processedID == event.ID {
				isProcessed = true
				break
			}
		}

		if !isProcessed {
			for _, err := range errors {
				if err.EventID == event.ID {
					isFailed = true
					break
				}
			}
		}

		// If neither processed nor failed, it was not processed
		if !isProcessed && !isFailed {
			notProcessedIDs = append(notProcessedIDs, event.ID)
		}
	}

	// Calculate metrics
	processingTime := time.Since(startTime)
	processingTimeMs := int(processingTime.Milliseconds())
	successRate := float64(processedCount) / float64(len(events)) * 100.0

	return &observability.EventProcessingResult{
		ProcessedCount:    processedCount,
		FailedCount:       failedCount,
		NotProcessedCount: notProcessedCount,
		RetryCount:        retryCount,
		ProcessingTimeMs:  processingTimeMs,
		ProcessedEventIDs: processedEventIDs,
		NotProcessedIDs:   notProcessedIDs,
		Errors:            errors,
		SuccessRate:       successRate,
	}, nil
}

// MarkEventProcessed marks an event as processed
func (s *telemetryEventService) MarkEventProcessed(ctx context.Context, eventID ulid.ULID) error {
	if eventID.IsZero() {
		return fmt.Errorf("event ID cannot be zero")
	}

	if err := s.eventRepo.MarkAsProcessed(ctx, eventID, time.Now()); err != nil {
		return fmt.Errorf("failed to mark event as processed: %w", err)
	}

	return nil
}

// MarkEventFailed marks an event as failed
func (s *telemetryEventService) MarkEventFailed(ctx context.Context, eventID ulid.ULID, errorMessage string) error {
	if eventID.IsZero() {
		return fmt.Errorf("event ID cannot be zero")
	}

	if errorMessage == "" {
		return fmt.Errorf("error message cannot be empty")
	}

	if err := s.eventRepo.MarkAsFailed(ctx, eventID, errorMessage); err != nil {
		return fmt.Errorf("failed to mark event as failed: %w", err)
	}

	return nil
}

// RetryEvent retries processing a single event
func (s *telemetryEventService) RetryEvent(ctx context.Context, eventID ulid.ULID) error {
	if eventID.IsZero() {
		return fmt.Errorf("event ID cannot be zero")
	}

	// Get the event
	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return fmt.Errorf("failed to get event: %w", err)
	}

	// Check if event can be retried
	if !event.HasErrors() {
		return fmt.Errorf("event %s has no errors to retry", eventID.String())
	}

	if event.IsProcessed() {
		return fmt.Errorf("event %s is already processed", eventID.String())
	}

	// Increment retry count
	if err := s.eventRepo.IncrementRetryCount(ctx, eventID); err != nil {
		return fmt.Errorf("failed to increment retry count: %w", err)
	}

	// Process the event
	return s.ProcessEvent(ctx, eventID)
}

// GetEventsForRetry retrieves events that should be retried
func (s *telemetryEventService) GetEventsForRetry(ctx context.Context, maxRetries int, limit int) ([]*observability.TelemetryEvent, error) {
	if maxRetries <= 0 {
		maxRetries = 3 // Default max retries
	}
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	events, err := s.eventRepo.GetEventsForRetry(ctx, maxRetries, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get events for retry: %w", err)
	}

	return events, nil
}

// BulkRetryEvents retries multiple events in bulk
func (s *telemetryEventService) BulkRetryEvents(ctx context.Context, eventIDs []ulid.ULID) (*observability.EventProcessingResult, error) {
	if len(eventIDs) == 0 {
		return &observability.EventProcessingResult{
			ProcessedCount: 0,
			FailedCount:    0,
			SuccessRate:    100.0,
		}, nil
	}

	// Get events to retry
	var events []*observability.TelemetryEvent
	for _, eventID := range eventIDs {
		event, err := s.eventRepo.GetByID(ctx, eventID)
		if err != nil {
			continue // Skip events that can't be found
		}
		events = append(events, event)
	}

	// Process the events
	return s.ProcessEventsBatch(ctx, events)
}

// GetEventStats retrieves statistics for telemetry events
func (s *telemetryEventService) GetEventStats(ctx context.Context, filter *observability.TelemetryEventFilter) (*observability.TelemetryEventStats, error) {
	stats, err := s.eventRepo.GetEventStats(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get event stats: %w", err)
	}

	return stats, nil
}

// GetEventTypeDistribution retrieves event type distribution
func (s *telemetryEventService) GetEventTypeDistribution(ctx context.Context, batchID *ulid.ULID) (map[observability.TelemetryEventType]int, error) {
	distribution, err := s.eventRepo.GetEventTypeDistribution(ctx, batchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get event type distribution: %w", err)
	}

	return distribution, nil
}

// CleanupFailedEvents removes failed events older than the specified time
func (s *telemetryEventService) CleanupFailedEvents(ctx context.Context, olderThan time.Time) (int64, error) {
	deletedCount, err := s.eventRepo.DeleteFailedEvents(ctx, olderThan)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup failed events: %w", err)
	}

	return deletedCount, nil
}

// processEventByType processes an event based on its type
func (s *telemetryEventService) processEventByType(ctx context.Context, event *observability.TelemetryEvent) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	// Validate event
	if validationErrors := event.Validate(); len(validationErrors) > 0 {
		return fmt.Errorf("event validation failed: %v", validationErrors)
	}

	// Process based on event type
	switch event.EventType {
	case observability.TelemetryEventTypeTraceCreate:
		return s.processTraceCreateEvent(ctx, event)
	case observability.TelemetryEventTypeTraceUpdate:
		return s.processTraceUpdateEvent(ctx, event)
	case observability.TelemetryEventTypeObservationCreate:
		return s.processObservationCreateEvent(ctx, event)
	case observability.TelemetryEventTypeObservationUpdate:
		return s.processObservationUpdateEvent(ctx, event)
	case observability.TelemetryEventTypeObservationComplete:
		return s.processObservationCompleteEvent(ctx, event)
	case observability.TelemetryEventTypeQualityScoreCreate:
		return s.processQualityScoreCreateEvent(ctx, event)
	default:
		return fmt.Errorf("unknown event type: %s", event.EventType)
	}
}

// Event processing methods for different event types
// In a real implementation, these would unmarshal the event payload and call appropriate services

func (s *telemetryEventService) processTraceCreateEvent(ctx context.Context, event *observability.TelemetryEvent) error {
	// Extract trace data from event payload and create trace
	// For now, simulate processing
	time.Sleep(time.Millisecond * 2)
	return nil
}

func (s *telemetryEventService) processTraceUpdateEvent(ctx context.Context, event *observability.TelemetryEvent) error {
	// Extract trace data from event payload and update trace
	time.Sleep(time.Millisecond * 1)
	return nil
}

func (s *telemetryEventService) processObservationCreateEvent(ctx context.Context, event *observability.TelemetryEvent) error {
	// Extract observation data from event payload and create observation
	time.Sleep(time.Millisecond * 3)
	return nil
}

func (s *telemetryEventService) processObservationUpdateEvent(ctx context.Context, event *observability.TelemetryEvent) error {
	// Extract observation data from event payload and update observation
	time.Sleep(time.Millisecond * 2)
	return nil
}

func (s *telemetryEventService) processObservationCompleteEvent(ctx context.Context, event *observability.TelemetryEvent) error {
	// Extract completion data from event payload and complete observation
	time.Sleep(time.Millisecond * 2)
	return nil
}

func (s *telemetryEventService) processQualityScoreCreateEvent(ctx context.Context, event *observability.TelemetryEvent) error {
	// Extract quality score data from event payload and create quality score
	time.Sleep(time.Millisecond * 1)
	return nil
}