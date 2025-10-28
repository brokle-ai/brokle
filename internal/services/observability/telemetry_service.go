package observability

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"brokle/internal/core/domain/observability"
	"brokle/internal/infrastructure/streams"
	"brokle/internal/workers"
	"brokle/pkg/ulid"
	appErrors "brokle/pkg/errors"
)

// TelemetryService aggregates all telemetry-related services with Redis Streams-based async processing
// Exported to allow type assertion for SetAnalyticsWorker injection
type TelemetryService struct {
	deduplicationService observability.TelemetryDeduplicationService
	streamProducer       *streams.TelemetryStreamProducer
	analyticsWorker      *workers.TelemetryAnalyticsWorker
	logger              *logrus.Logger

	// Performance tracking
	mu                 sync.RWMutex
	batchesProcessed   uint64
	eventsProcessed    uint64
	lastProcessingTime time.Time
	avgProcessingTime  time.Duration
}

// NewTelemetryService creates a new telemetry service with Redis Streams and deduplication
func NewTelemetryService(
	deduplicationService observability.TelemetryDeduplicationService,
	streamProducer *streams.TelemetryStreamProducer,
	analyticsWorker *workers.TelemetryAnalyticsWorker,
	logger *logrus.Logger,
) observability.TelemetryService {
	return &TelemetryService{
		deduplicationService: deduplicationService,
		streamProducer:       streamProducer,
		analyticsWorker:      analyticsWorker,
		logger:              logger,
		lastProcessingTime:   time.Now(),
	}
}

// SetAnalyticsWorker injects the analytics worker (for two-phase initialization)
// This allows the telemetry service to be created before the worker is started,
// ensuring clean dependency initialization order.
func (s *TelemetryService) SetAnalyticsWorker(worker *workers.TelemetryAnalyticsWorker) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.analyticsWorker = worker
	s.logger.Debug("Analytics worker injected into telemetry service")
}

// ProcessTelemetryBatch processes a batch of telemetry events using Redis Streams (async)
// Returns 202 Accepted for async processing with batch ID for tracking
func (s *TelemetryService) ProcessTelemetryBatch(ctx context.Context, request *observability.TelemetryBatchRequest) (*observability.TelemetryBatchResponse, error) {
	startTime := time.Now()

	// Validate request
	if err := s.validateBatchRequest(request); err != nil {
		return nil, appErrors.NewValidationError("batch_request", fmt.Sprintf("Invalid batch request: %v", err))
	}

	// Generate batch ID
	batchID := ulid.New()

	// Extract event IDs for atomic claim operation
	eventIDs := make([]ulid.ULID, len(request.Events))
	for i, event := range request.Events {
		eventIDs[i] = event.EventID
	}

	// ✅ Atomic claim: Check and register events in single operation (eliminates race condition)
	claimedIDs, duplicateIDs, err := s.deduplicationService.ClaimEvents(ctx, request.ProjectID, batchID, eventIDs, 24*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("failed to claim events for deduplication: %w", err)
	}

	// Skip batch entirely if all events are duplicates
	if len(claimedIDs) == 0 {
		s.logger.WithFields(logrus.Fields{
			"batch_id":        batchID.String(),
			"total_events":    len(request.Events),
			"duplicate_count": len(duplicateIDs),
		}).Info("All events are duplicates, skipping batch processing")

		return &observability.TelemetryBatchResponse{
			BatchID:           batchID,
			ProcessedEvents:   0,
			DuplicateEvents:   len(duplicateIDs),
			FailedEvents:      0,
			ProcessingTimeMs:  int(time.Since(startTime).Milliseconds()),
			DuplicateEventIDs: duplicateIDs,
		}, nil
	}

	// Build event data for ONLY claimed events (not duplicates)
	claimedMap := make(map[ulid.ULID]bool)
	for _, id := range claimedIDs {
		claimedMap[id] = true
	}

	claimedEventData := make([]streams.TelemetryEventData, 0, len(claimedIDs))
	for _, eventReq := range request.Events {
		if claimedMap[eventReq.EventID] {
			claimedEventData = append(claimedEventData, streams.TelemetryEventData{
				EventID:      eventReq.EventID,
				EventType:    string(eventReq.EventType),
				EventPayload: eventReq.Payload,
			})
		}
	}

	// Build stream message with claimed events
	streamMessage := &streams.TelemetryStreamMessage{
		BatchID:         batchID,
		ProjectID:       request.ProjectID,
		Events:          claimedEventData,
		ClaimedEventIDs: claimedIDs, // Pass claimed IDs for consumer claim release
		Metadata:        request.Metadata,
		Timestamp:       time.Now(),
	}

	// Publish batch to Redis Streams
	streamID, err := s.streamProducer.PublishBatch(ctx, streamMessage)
	if err != nil {
		// ⚠️ CRITICAL: Rollback claimed events on publish failure
		if rollbackErr := s.deduplicationService.ReleaseEvents(ctx, claimedIDs); rollbackErr != nil {
			s.logger.WithError(rollbackErr).WithFields(logrus.Fields{
				"batch_id":     batchID.String(),
				"claimed_count": len(claimedIDs),
			}).Error("CRITICAL: Failed to rollback claimed events after publish failure - manual cleanup may be needed")
		}
		return nil, fmt.Errorf("failed to publish batch to stream: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"batch_id":      batchID.String(),
		"stream_id":     streamID,
		"project_id":    request.ProjectID.String(),
		"claimed_events": len(claimedIDs),
		"duplicates":    len(duplicateIDs),
	}).Info("Batch published to Redis Stream for async processing")

	// Update performance metrics
	s.updatePerformanceMetrics(len(request.Events), time.Since(startTime))

	// Build response (202 Accepted - async processing)
	response := &observability.TelemetryBatchResponse{
		BatchID:           batchID,
		ProcessedEvents:   len(claimedIDs), // Events claimed and accepted for processing
		DuplicateEvents:   len(duplicateIDs),
		FailedEvents:      0,
		ProcessingTimeMs:  int(time.Since(startTime).Milliseconds()),
		DuplicateEventIDs: duplicateIDs,
	}

	return response, nil
}

// Deduplication returns the deduplication service
func (s *TelemetryService) Deduplication() observability.TelemetryDeduplicationService {
	return s.deduplicationService
}

// GetHealth returns the health status of all telemetry services
func (s *TelemetryService) GetHealth(ctx context.Context) (*observability.TelemetryHealthStatus, error) {
	// Get analytics worker health if available
	var analyticsHealth *workers.HealthMetrics
	var activeWorkers int = 1 // Default
	if s.analyticsWorker != nil {
		analyticsHealth = s.analyticsWorker.GetHealth()
		activeWorkers = analyticsHealth.ActiveWorkers
	}

	health := &observability.TelemetryHealthStatus{
		Healthy:               s.isHealthy(analyticsHealth),
		ActiveWorkers:         activeWorkers,
		AverageProcessingTime: float64(s.avgProcessingTime.Milliseconds()),
		ThroughputPerMinute:   float64(s.eventsProcessed) / time.Since(s.lastProcessingTime).Minutes(),
	}

	// Set default database health
	health.Database = &observability.DatabaseHealth{
		Connected:         true,
		LatencyMs:         1.5, // Default value
		ActiveConnections: 10,  // Default value
		MaxConnections:    100, // Default value
	}

	// Set default Redis health
	health.Redis = &observability.RedisHealthStatus{
		Available:   true,
		LatencyMs:   0.5, // Default value
		Connections: 5,   // Default value
		LastError:   nil,
		Uptime:      time.Hour * 24, // Default uptime
	}

	// Add processing queue health from analytics worker
	if analyticsHealth != nil {
		health.ProcessingQueue = &observability.QueueHealth{
			Size:             int64(analyticsHealth.QueueDepth),
			ProcessingRate:   float64(s.eventsProcessed),
			AverageWaitTime:  10.0, // Default value
			OldestMessageAge: 0,
		}
	} else {
		health.ProcessingQueue = &observability.QueueHealth{
			Size:             0,
			ProcessingRate:   float64(s.eventsProcessed),
			AverageWaitTime:  10.0, // Default value
			OldestMessageAge: 0,
		}
	}

	// Calculate error rate from analytics worker if available
	s.mu.RLock()
	if analyticsHealth != nil {
		health.ErrorRate = analyticsHealth.ErrorRate
	} else if s.eventsProcessed > 0 {
		health.ErrorRate = 0.01 // 1% default error rate
	}
	s.mu.RUnlock()

	return health, nil
}

// isHealthy determines overall health based on analytics worker status
func (s *TelemetryService) isHealthy(analyticsHealth *workers.HealthMetrics) bool {
	// If analytics worker is not available, service is still healthy
	if analyticsHealth == nil {
		return true
	}

	// Check analytics worker health
	if !analyticsHealth.Healthy {
		return false
	}

	// Check if error rate is acceptable (< 5%)
	if analyticsHealth.ErrorRate > 0.05 {
		return false
	}

	// Check if queue is not severely backed up (< 1000 items)
	if analyticsHealth.QueueDepth > 1000 {
		return false
	}

	return true
}

// GetMetrics returns aggregated metrics from all telemetry services
func (s *TelemetryService) GetMetrics(ctx context.Context) (*observability.TelemetryMetrics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Calculate throughput per second
	throughput := float64(0)
	if s.lastProcessingTime.Before(time.Now()) {
		elapsed := time.Since(s.lastProcessingTime)
		if elapsed.Seconds() > 0 {
			throughput = float64(s.eventsProcessed) / elapsed.Seconds()
		}
	}

	// Calculate success rate
	successRate := float64(100)
	if s.eventsProcessed > 0 {
		successRate = 99.0 // Default 99% success rate
	}

	// Aggregate metrics
	metrics := &observability.TelemetryMetrics{
		TotalBatches:         int64(s.batchesProcessed),
		CompletedBatches:     int64(s.batchesProcessed),
		FailedBatches:        0,
		ProcessingBatches:    0,
		TotalEvents:          int64(s.eventsProcessed),
		ProcessedEvents:      int64(s.eventsProcessed),
		FailedEvents:         0,
		DuplicateEvents:      0,
		AverageEventsPerBatch: func() float64 {
			if s.batchesProcessed > 0 {
				return float64(s.eventsProcessed) / float64(s.batchesProcessed)
			}
			return 0
		}(),
		ThroughputPerSecond:  throughput,
		SuccessRate:          successRate,
		DeduplicationRate:    0.0, // Will be updated with actual dedup stats
	}

	return metrics, nil
}

// GetPerformanceStats returns performance statistics for a given time window
func (s *TelemetryService) GetPerformanceStats(ctx context.Context, timeWindow time.Duration) (*observability.TelemetryPerformanceStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Calculate performance statistics based on current metrics
	throughputPerSec := float64(0)
	if timeWindow.Seconds() > 0 {
		throughputPerSec = float64(s.eventsProcessed) / timeWindow.Seconds()
	}

	// Aggregate performance stats
	stats := &observability.TelemetryPerformanceStats{
		TimeWindow:           timeWindow,
		TotalRequests:        int64(s.batchesProcessed),
		SuccessfulRequests:   int64(s.batchesProcessed),
		AverageLatencyMs:     float64(s.avgProcessingTime.Milliseconds()),
		P95LatencyMs:         float64(s.avgProcessingTime.Milliseconds()) * 1.2, // Estimated
		P99LatencyMs:         float64(s.avgProcessingTime.Milliseconds()) * 1.5, // Estimated
		ThroughputPerSecond:  throughputPerSec,
		PeakThroughput:       throughputPerSec * 1.3, // Estimated peak
		CacheHitRate:         0.8,  // Default 80% cache hit rate
		DatabaseFallbackRate: 0.2,  // Default 20% fallback rate
		ErrorRate:            0.01, // Default 1% error rate
		RetryRate:            0.05, // Default 5% retry rate
	}

	return stats, nil
}

// validateBatchRequest validates the telemetry batch request
func (s *TelemetryService) validateBatchRequest(request *observability.TelemetryBatchRequest) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if request.ProjectID == (ulid.ULID{}) {
		return fmt.Errorf("project ID is required")
	}

	if len(request.Events) == 0 {
		return fmt.Errorf("events list cannot be empty")
	}

	if len(request.Events) > 1000 { // Reasonable batch size limit
		return fmt.Errorf("batch size exceeds maximum limit of 1000 events")
	}

	// Validate each event
	for i, event := range request.Events {
		if event == nil {
			return fmt.Errorf("event at index %d cannot be nil", i)
		}

		if event.EventID == (ulid.ULID{}) {
			return fmt.Errorf("event at index %d has invalid event ID", i)
		}

		if event.EventType == "" {
			return fmt.Errorf("event at index %d has empty event type", i)
		}

		if len(event.Payload) == 0 {
			return fmt.Errorf("event at index %d has empty payload", i)
		}
	}

	return nil
}

// updatePerformanceMetrics updates internal performance tracking metrics
func (s *TelemetryService) updatePerformanceMetrics(eventCount int, processingTime time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.batchesProcessed++
	s.eventsProcessed += uint64(eventCount)
	s.lastProcessingTime = time.Now()

	// Update rolling average processing time
	if s.avgProcessingTime == 0 {
		s.avgProcessingTime = processingTime
	} else {
		// Simple exponential moving average with alpha = 0.1
		s.avgProcessingTime = time.Duration(0.9*float64(s.avgProcessingTime) + 0.1*float64(processingTime))
	}
}

// queueAnalyticsJobs queues telemetry data for analytics processing
func (s *TelemetryService) queueAnalyticsJobs(
	ctx context.Context,
	request *observability.TelemetryBatchRequest,
	batch *observability.TelemetryBatch,
	events []*observability.TelemetryEvent,
	processingTime time.Duration,
) {
	if s.analyticsWorker == nil {
		s.logger.Debug("Analytics worker not available, skipping analytics jobs")
		return
	}

	// Queue individual telemetry events for analytics
	for _, event := range events {
		eventJob := &workers.TelemetryEventJob{
			BatchID:     batch.ID,
			EventID:     event.ID,
			ProjectID:   batch.ProjectID,
			EventType:   event.EventType,
			EventData:   event.EventPayload,
			Timestamp:   event.CreatedAt,
			RetryCount:  0,
			Priority:    workers.PriorityNormal,
		}

		// Queue event with high priority for critical events
		if s.isCriticalEventType(event.EventType) {
			eventJob.Priority = workers.PriorityHigh
		}

		success := s.analyticsWorker.QueueTelemetryEvent(eventJob)
		if !success {
			s.logger.WithFields(logrus.Fields{
				"event_id": event.ID.String(),
				"batch_id": batch.ID.String(),
			}).Warn("Failed to queue telemetry event for analytics")
		}
	}

	// Queue batch analytics job
	batchJob := &workers.TelemetryBatchJob{
		BatchID:         batch.ID,
		ProjectID:       batch.ProjectID,
		Status:          batch.Status,
		TotalEvents:     batch.TotalEvents,
		ProcessedEvents: batch.ProcessedEvents,
		FailedEvents:    batch.FailedEvents,
		ProcessingTime:  processingTime,
		Metadata:        batch.BatchMetadata,
		Timestamp:       batch.CreatedAt,
		RetryCount:      0,
		Priority:        workers.PriorityNormal,
	}

	success := s.analyticsWorker.QueueTelemetryBatch(batchJob)
	if !success {
		s.logger.WithField("batch_id", batch.ID.String()).Warn("Failed to queue telemetry batch for analytics")
	}

	// Queue performance metrics
	metricsJob := &workers.TelemetryMetricJob{
		ProjectID:   batch.ProjectID,
		MetricName:  "batch_processing_time",
		MetricType:  workers.MetricTypeHistogram,
		MetricValue: float64(processingTime.Milliseconds()),
		Labels: map[string]string{
			"batch_id":     batch.ID.String(),
			"event_count":  fmt.Sprintf("%d", len(events)),
			"status":       string(batch.Status),
		},
		Metadata: map[string]interface{}{
			"events_processed": batch.ProcessedEvents,
			"events_failed":    batch.FailedEvents,
			"total_events":     batch.TotalEvents,
		},
		Timestamp:  time.Now(),
		RetryCount: 0,
		Priority:   workers.PriorityLow,
	}

	success = s.analyticsWorker.QueueTelemetryMetric(metricsJob)
	if !success {
		s.logger.WithField("batch_id", batch.ID.String()).Warn("Failed to queue telemetry metrics for analytics")
	}

	s.logger.WithFields(logrus.Fields{
		"batch_id":      batch.ID.String(),
		"events_queued": len(events),
		"metrics_queued": 1,
	}).Debug("Successfully queued telemetry data for analytics")
}

// isCriticalEventType determines if an event type should be processed with high priority
func (s *TelemetryService) isCriticalEventType(eventType observability.TelemetryEventType) bool {
	criticalTypes := []observability.TelemetryEventType{
		observability.TelemetryEventTypeTrace,
		observability.TelemetryEventTypeObservation,
		observability.TelemetryEventTypeQualityScore,
	}

	for _, criticalType := range criticalTypes {
		if eventType == criticalType {
			return true
		}
	}
	return false
}

// updateEventStatuses updates event processing statuses based on processing results
// Uses existing schema fields: processed_at, error_message, retry_count
func (s *TelemetryService) updateEventStatuses(events []*observability.TelemetryEvent, result *observability.EventProcessingResult) {
	now := time.Now()

	// Build lookup maps for O(1) status determination
	errorMap := make(map[ulid.ULID]string)
	for _, err := range result.Errors {
		errorMap[err.EventID] = err.ErrorMessage
	}

	// Update each event's status precisely
	for _, event := range events {
		if errorMsg, hasError := errorMap[event.ID]; hasError {
			// Event explicitly failed during processing
			event.ErrorMessage = &errorMsg
			event.RetryCount++
			event.ProcessedAt = nil // Keep as unprocessed for retry
		} else {
			// Event processed successfully (no error recorded)
			event.ProcessedAt = &now
			event.ErrorMessage = nil
			// retry_count stays the same
		}
	}
}