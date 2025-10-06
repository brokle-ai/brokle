package observability

import (
	"context"
	"fmt"
	"time"

	"brokle/internal/core/domain/observability"
	"brokle/pkg/ulid"
)

// telemetryBatchService implements the TelemetryBatchService interface
type telemetryBatchService struct {
	batchRepo         observability.TelemetryBatchRepository
	eventRepo         observability.TelemetryEventRepository
	deduplicationRepo observability.TelemetryDeduplicationRepository
}

// NewTelemetryBatchService creates a new telemetry batch service
func NewTelemetryBatchService(
	batchRepo observability.TelemetryBatchRepository,
	eventRepo observability.TelemetryEventRepository,
	deduplicationRepo observability.TelemetryDeduplicationRepository,
) observability.TelemetryBatchService {
	return &telemetryBatchService{
		batchRepo:         batchRepo,
		eventRepo:         eventRepo,
		deduplicationRepo: deduplicationRepo,
	}
}

// CreateBatch creates a new telemetry batch with validation
func (s *telemetryBatchService) CreateBatch(ctx context.Context, batch *observability.TelemetryBatch) (*observability.TelemetryBatch, error) {
	if batch == nil {
		return nil, fmt.Errorf("batch cannot be nil")
	}

	// Generate ID if not provided
	if batch.ID.IsZero() {
		batch.ID = ulid.New()
	}

	// Validate batch
	if validationErrors := batch.Validate(); len(validationErrors) > 0 {
		return nil, fmt.Errorf("batch validation failed: %v", validationErrors)
	}

	// Set initial status if not provided
	if batch.Status == "" {
		batch.Status = observability.BatchStatusProcessing
	}

	// Set creation timestamp
	if batch.CreatedAt.IsZero() {
		batch.CreatedAt = time.Now()
	}

	// Create batch in repository
	if err := s.batchRepo.Create(ctx, batch); err != nil {
		return nil, fmt.Errorf("failed to create batch: %w", err)
	}

	return batch, nil
}

// GetBatch retrieves a telemetry batch by ID
func (s *telemetryBatchService) GetBatch(ctx context.Context, id ulid.ULID) (*observability.TelemetryBatch, error) {
	if id.IsZero() {
		return nil, fmt.Errorf("batch ID cannot be zero")
	}

	batch, err := s.batchRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get batch: %w", err)
	}

	return batch, nil
}

// UpdateBatch updates an existing telemetry batch
func (s *telemetryBatchService) UpdateBatch(ctx context.Context, batch *observability.TelemetryBatch) (*observability.TelemetryBatch, error) {
	if batch == nil {
		return nil, fmt.Errorf("batch cannot be nil")
	}

	if batch.ID.IsZero() {
		return nil, fmt.Errorf("batch ID cannot be zero")
	}

	// Validate batch
	if validationErrors := batch.Validate(); len(validationErrors) > 0 {
		return nil, fmt.Errorf("batch validation failed: %v", validationErrors)
	}

	// Update batch in repository
	if err := s.batchRepo.Update(ctx, batch); err != nil {
		return nil, fmt.Errorf("failed to update batch: %w", err)
	}

	return batch, nil
}

// DeleteBatch deletes a telemetry batch
func (s *telemetryBatchService) DeleteBatch(ctx context.Context, id ulid.ULID) error {
	if id.IsZero() {
		return fmt.Errorf("batch ID cannot be zero")
	}

	if err := s.batchRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete batch: %w", err)
	}

	return nil
}

// ListBatches retrieves telemetry batches with filtering
func (s *telemetryBatchService) ListBatches(ctx context.Context, filter *observability.TelemetryBatchFilter) ([]*observability.TelemetryBatch, int, error) {
	if filter == nil {
		filter = &observability.TelemetryBatchFilter{
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

	batches, totalCount, err := s.batchRepo.SearchBatches(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list batches: %w", err)
	}

	return batches, totalCount, nil
}

// GetBatchWithEvents retrieves a batch with all its events
func (s *telemetryBatchService) GetBatchWithEvents(ctx context.Context, id ulid.ULID) (*observability.TelemetryBatch, error) {
	if id.IsZero() {
		return nil, fmt.Errorf("batch ID cannot be zero")
	}

	batch, err := s.batchRepo.GetBatchWithEvents(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get batch with events: %w", err)
	}

	return batch, nil
}

// GetActiveBatches retrieves active batches for a project
func (s *telemetryBatchService) GetActiveBatches(ctx context.Context, projectID ulid.ULID) ([]*observability.TelemetryBatch, error) {
	if projectID.IsZero() {
		return nil, fmt.Errorf("project ID cannot be zero")
	}

	batches, err := s.batchRepo.GetActiveByProjectID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active batches: %w", err)
	}

	return batches, nil
}

// GetProcessingBatches retrieves all batches currently being processed
func (s *telemetryBatchService) GetProcessingBatches(ctx context.Context) ([]*observability.TelemetryBatch, error) {
	batches, err := s.batchRepo.GetProcessingBatches(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get processing batches: %w", err)
	}

	return batches, nil
}

// ProcessBatch processes all events in a telemetry batch
func (s *telemetryBatchService) ProcessBatch(ctx context.Context, batchID ulid.ULID) (*observability.BatchProcessingResult, error) {
	if batchID.IsZero() {
		return nil, fmt.Errorf("batch ID cannot be zero")
	}

	startTime := time.Now()

	// Get the batch
	batch, err := s.batchRepo.GetByID(ctx, batchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get batch: %w", err)
	}

	// Get unprocessed events for the batch
	events, err := s.eventRepo.GetUnprocessedByBatchID(ctx, batchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unprocessed events: %w", err)
	}

	var processedCount, failedCount int
	var errors []observability.TelemetryEventError

	// Process each event
	for _, event := range events {
		if err := s.processEvent(ctx, event); err != nil {
			failedCount++
			errors = append(errors, observability.TelemetryEventError{
				EventID:      event.ID,
				EventType:    event.EventType,
				ErrorCode:    "PROCESSING_FAILED",
				ErrorMessage: err.Error(),
				Retryable:    true,
			})

			// Mark event as failed
			if markErr := s.eventRepo.MarkAsFailed(ctx, event.ID, err.Error()); markErr != nil {
				// Log error but continue processing
				_ = markErr
			}
		} else {
			processedCount++
			// Mark event as processed
			if markErr := s.eventRepo.MarkAsProcessed(ctx, event.ID, time.Now()); markErr != nil {
				// Log error but continue processing
				_ = markErr
			}
		}
	}

	// Update batch counts
	batch.ProcessedEvents = processedCount
	batch.FailedEvents = failedCount

	// Calculate processing time
	processingTime := time.Since(startTime)
	processingTimeMs := int(processingTime.Milliseconds())
	batch.ProcessingTimeMs = &processingTimeMs

	// Determine final batch status
	if failedCount == 0 {
		batch.Status = observability.BatchStatusCompleted
	} else if processedCount == 0 {
		batch.Status = observability.BatchStatusFailed
	} else {
		batch.Status = observability.BatchStatusPartial
	}

	// Update batch with final status
	if updateErr := s.batchRepo.Update(ctx, batch); updateErr != nil {
		return nil, fmt.Errorf("failed to update batch status: %w", updateErr)
	}

	// Calculate success rate and throughput
	successRate := float64(processedCount) / float64(len(events)) * 100.0
	throughputPerSec := float64(processedCount) / processingTime.Seconds()

	return &observability.BatchProcessingResult{
		BatchID:          batchID,
		TotalEvents:      len(events),
		ProcessedEvents:  processedCount,
		FailedEvents:     failedCount,
		ProcessingTimeMs: processingTimeMs,
		ThroughputPerSec: throughputPerSec,
		Errors:           errors,
		SuccessRate:      successRate,
	}, nil
}

// ProcessEventsBatch processes a batch of events directly
func (s *telemetryBatchService) ProcessEventsBatch(ctx context.Context, events []*observability.TelemetryEvent) (*observability.BatchProcessingResult, error) {
	if len(events) == 0 {
		return &observability.BatchProcessingResult{
			TotalEvents:     0,
			ProcessedEvents: 0,
			FailedEvents:    0,
			SuccessRate:     100.0,
		}, nil
	}

	startTime := time.Now()
	var processedCount, failedCount int
	var errors []observability.TelemetryEventError

	// Process each event
	for _, event := range events {
		if event == nil {
			failedCount++
			continue
		}

		if err := s.processEvent(ctx, event); err != nil {
			failedCount++
			errors = append(errors, observability.TelemetryEventError{
				EventID:      event.ID,
				EventType:    event.EventType,
				ErrorCode:    "PROCESSING_FAILED",
				ErrorMessage: err.Error(),
				Retryable:    true,
			})
		} else {
			processedCount++
		}
	}

	// Calculate metrics
	processingTime := time.Since(startTime)
	processingTimeMs := int(processingTime.Milliseconds())
	successRate := float64(processedCount) / float64(len(events)) * 100.0
	throughputPerSec := float64(processedCount) / processingTime.Seconds()

	return &observability.BatchProcessingResult{
		TotalEvents:      len(events),
		ProcessedEvents:  processedCount,
		FailedEvents:     failedCount,
		ProcessingTimeMs: processingTimeMs,
		ThroughputPerSec: throughputPerSec,
		Errors:           errors,
		SuccessRate:      successRate,
	}, nil
}

// RetryFailedEvents retries failed events in a batch
func (s *telemetryBatchService) RetryFailedEvents(ctx context.Context, batchID ulid.ULID, maxRetries int) (*observability.BatchProcessingResult, error) {
	if batchID.IsZero() {
		return nil, fmt.Errorf("batch ID cannot be zero")
	}

	if maxRetries <= 0 {
		maxRetries = 3 // Default max retries
	}

	// Get failed events that can be retried
	failedEvents, err := s.eventRepo.GetFailedByBatchID(ctx, batchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get failed events: %w", err)
	}

	// Filter events that should be retried
	var retryableEvents []*observability.TelemetryEvent
	for _, event := range failedEvents {
		if event.ShouldRetry(maxRetries) {
			retryableEvents = append(retryableEvents, event)
		}
	}

	if len(retryableEvents) == 0 {
		return &observability.BatchProcessingResult{
			BatchID:         batchID,
			TotalEvents:     0,
			ProcessedEvents: 0,
			FailedEvents:    0,
			SuccessRate:     100.0,
		}, nil
	}

	// Process retryable events
	result, err := s.ProcessEventsBatch(ctx, retryableEvents)
	if err != nil {
		return nil, fmt.Errorf("failed to retry events: %w", err)
	}

	result.BatchID = batchID
	return result, nil
}

// GetBatchStats retrieves statistics for a telemetry batch
func (s *telemetryBatchService) GetBatchStats(ctx context.Context, id ulid.ULID) (*observability.BatchStats, error) {
	if id.IsZero() {
		return nil, fmt.Errorf("batch ID cannot be zero")
	}

	stats, err := s.batchRepo.GetBatchStats(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get batch stats: %w", err)
	}

	return stats, nil
}

// GetBatchMetrics retrieves processing metrics for telemetry batches
func (s *telemetryBatchService) GetBatchMetrics(ctx context.Context, filter *observability.TelemetryBatchFilter) (*observability.BatchProcessingMetrics, error) {
	metrics, err := s.batchRepo.GetBatchProcessingMetrics(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get batch metrics: %w", err)
	}

	return metrics, nil
}

// GetThroughputStats retrieves throughput statistics for a project
func (s *telemetryBatchService) GetThroughputStats(ctx context.Context, projectID ulid.ULID, timeWindow time.Duration) (*observability.BatchThroughputStats, error) {
	if projectID.IsZero() {
		return nil, fmt.Errorf("project ID cannot be zero")
	}

	if timeWindow <= 0 {
		timeWindow = time.Hour // Default to 1 hour
	}

	stats, err := s.batchRepo.GetBatchThroughputStats(ctx, projectID, timeWindow)
	if err != nil {
		return nil, fmt.Errorf("failed to get throughput stats: %w", err)
	}

	return stats, nil
}

// MarkBatchCompleted marks a batch as completed
func (s *telemetryBatchService) MarkBatchCompleted(ctx context.Context, batchID ulid.ULID, processingTimeMs int) error {
	if batchID.IsZero() {
		return fmt.Errorf("batch ID cannot be zero")
	}

	if err := s.batchRepo.UpdateBatchStatus(ctx, batchID, observability.BatchStatusCompleted, &processingTimeMs); err != nil {
		return fmt.Errorf("failed to mark batch completed: %w", err)
	}

	return nil
}

// MarkBatchFailed marks a batch as failed
func (s *telemetryBatchService) MarkBatchFailed(ctx context.Context, batchID ulid.ULID, errorMessage string) error {
	if batchID.IsZero() {
		return fmt.Errorf("batch ID cannot be zero")
	}

	if err := s.batchRepo.UpdateBatchStatus(ctx, batchID, observability.BatchStatusFailed, nil); err != nil {
		return fmt.Errorf("failed to mark batch failed: %w", err)
	}

	return nil
}

// MarkBatchPartial marks a batch as partially completed
func (s *telemetryBatchService) MarkBatchPartial(ctx context.Context, batchID ulid.ULID, processingTimeMs int) error {
	if batchID.IsZero() {
		return fmt.Errorf("batch ID cannot be zero")
	}

	if err := s.batchRepo.UpdateBatchStatus(ctx, batchID, observability.BatchStatusPartial, &processingTimeMs); err != nil {
		return fmt.Errorf("failed to mark batch partial: %w", err)
	}

	return nil
}

// processEvent processes a single telemetry event based on its type
func (s *telemetryBatchService) processEvent(ctx context.Context, event *observability.TelemetryEvent) error {
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
func (s *telemetryBatchService) processTraceCreateEvent(ctx context.Context, event *observability.TelemetryEvent) error {
	// Extract trace data from event payload
	// In a real implementation, this would unmarshal the payload into a Trace struct
	// and create it using the trace service

	// For now, simulate processing
	time.Sleep(time.Millisecond * 2) // Simulate processing time
	return nil
}

func (s *telemetryBatchService) processTraceUpdateEvent(ctx context.Context, event *observability.TelemetryEvent) error {
	// Process trace update event
	time.Sleep(time.Millisecond * 1)
	return nil
}

func (s *telemetryBatchService) processObservationCreateEvent(ctx context.Context, event *observability.TelemetryEvent) error {
	// Process observation create event
	time.Sleep(time.Millisecond * 3)
	return nil
}

func (s *telemetryBatchService) processObservationUpdateEvent(ctx context.Context, event *observability.TelemetryEvent) error {
	// Process observation update event
	time.Sleep(time.Millisecond * 2)
	return nil
}

func (s *telemetryBatchService) processObservationCompleteEvent(ctx context.Context, event *observability.TelemetryEvent) error {
	// Process observation complete event
	time.Sleep(time.Millisecond * 2)
	return nil
}

func (s *telemetryBatchService) processQualityScoreCreateEvent(ctx context.Context, event *observability.TelemetryEvent) error {
	// Process quality score create event
	time.Sleep(time.Millisecond * 1)
	return nil
}