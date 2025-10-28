package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"

	"brokle/internal/core/domain/observability"
)

// processTelemetryEvent processes a single telemetry event
func (w *TelemetryAnalyticsWorker) processTelemetryEvent(job *TelemetryEventJob) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Convert job to domain telemetry event format WITH context from job
	processedAt := time.Now()
	event := &observability.TelemetryEvent{
		ID:           job.EventID,
		BatchID:      job.BatchID,
		ProjectID:    job.ProjectID,
		EventType:    job.EventType,
		EventPayload: job.EventData,
		CreatedAt:    job.Timestamp,
		RetryCount:   job.RetryCount,
		ProcessedAt:  &processedAt,
	}

	// Insert into ClickHouse using domain type (context carried in struct)
	if err := w.repository.InsertTelemetryEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to insert telemetry event: %w", err)
	}

	return nil
}

// processTelemetryBatch processes a telemetry batch record
func (w *TelemetryAnalyticsWorker) processTelemetryBatch(job *TelemetryBatchJob) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Convert job to domain telemetry batch format WITH context from job
	processingTimeMs := int(job.ProcessingTime.Milliseconds())
	completedAt := time.Now()
	batch := &observability.TelemetryBatch{
		ID:               job.BatchID,
		ProjectID:        job.ProjectID,
		BatchMetadata:    job.Metadata,
		TotalEvents:      job.TotalEvents,
		ProcessedEvents:  job.ProcessedEvents,
		FailedEvents:     job.FailedEvents,
		Status:           job.Status,
		ProcessingTimeMs: &processingTimeMs,
		CreatedAt:        job.Timestamp,
		CompletedAt:      &completedAt,
	}

	// Insert into ClickHouse using domain type (context carried in struct)
	if err := w.repository.InsertTelemetryBatch(ctx, batch); err != nil {
		return fmt.Errorf("failed to insert telemetry batch: %w", err)
	}

	return nil
}

// processTelemetryMetric processes a telemetry metric
func (w *TelemetryAnalyticsWorker) processTelemetryMetric(job *TelemetryMetricJob) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Convert labels from map[string]string to map[string]interface{}
	labels := make(map[string]interface{}, len(job.Labels))
	for k, v := range job.Labels {
		labels[k] = v
	}

	// Convert job to domain telemetry metric format
	processedAt := time.Now()
	metric := &observability.TelemetryMetric{
		ProjectID:   job.ProjectID,
		MetricName:  job.MetricName,
		MetricType:  string(job.MetricType),
		MetricValue: job.MetricValue,
		Labels:      labels,
		Metadata:    job.Metadata,
		Timestamp:   job.Timestamp,
		ProcessedAt: &processedAt,
	}

	// Insert into ClickHouse using domain type
	if err := w.repository.InsertTelemetryMetric(ctx, metric); err != nil {
		return fmt.Errorf("failed to insert telemetry metric: %w", err)
	}

	return nil
}

// processBulkOperations handles bulk insertions for better ClickHouse performance
func (w *TelemetryAnalyticsWorker) processBulkOperations() {
	w.bufferMutex.Lock()
	defer w.bufferMutex.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Process event buffer
	if len(w.eventBuffer) > 0 {
		events := make([]*observability.TelemetryEvent, len(w.eventBuffer))
		processedAt := time.Now()

		// Each event carries its own project_id from its job
		for i, job := range w.eventBuffer {
			events[i] = &observability.TelemetryEvent{
				ID:           job.EventID,
				BatchID:      job.BatchID,
				ProjectID:    job.ProjectID,
				EventType:    job.EventType,
				EventPayload: job.EventData,
				CreatedAt:    job.Timestamp,
				RetryCount:   job.RetryCount,
				ProcessedAt:  &processedAt,
			}
		}

		if err := w.repository.InsertTelemetryEventsBatch(ctx, events); err != nil {
			w.logger.WithError(err).WithField("count", len(events)).Error("Failed to insert telemetry events batch")
			// Retry individual events
			w.retryBulkEvents(w.eventBuffer)
		} else {
			w.logger.WithField("count", len(events)).Debug("Telemetry events batch processed")
		}

		// Clear buffer
		w.eventBuffer = w.eventBuffer[:0]
	}

	// Process batch buffer
	if len(w.batchBuffer) > 0 {
		batches := make([]*observability.TelemetryBatch, len(w.batchBuffer))
		completedAt := time.Now()

		// Each batch carries its own project_id from its job
		for i, job := range w.batchBuffer {
			processingTimeMs := int(job.ProcessingTime.Milliseconds())
			batches[i] = &observability.TelemetryBatch{
				ID:               job.BatchID,
				ProjectID:        job.ProjectID,
				BatchMetadata:    job.Metadata,
				TotalEvents:      job.TotalEvents,
				ProcessedEvents:  job.ProcessedEvents,
				FailedEvents:     job.FailedEvents,
				Status:           job.Status,
				ProcessingTimeMs: &processingTimeMs,
				CreatedAt:        job.Timestamp,
				CompletedAt:      &completedAt,
			}
		}

		if err := w.repository.InsertTelemetryBatchesBatch(ctx, batches); err != nil {
			w.logger.WithError(err).WithField("count", len(batches)).Error("Failed to insert telemetry batches batch")
			// Retry individual batches
			w.retryBulkBatches(w.batchBuffer)
		} else {
			w.logger.WithField("count", len(batches)).Debug("Telemetry batches batch processed")
		}

		// Clear buffer
		w.batchBuffer = w.batchBuffer[:0]
	}

	// Process metrics buffer
	if len(w.metricsBuffer) > 0 {
		metrics := make([]*observability.TelemetryMetric, len(w.metricsBuffer))
		processedAt := time.Now()
		for i, job := range w.metricsBuffer {
			// Convert labels from map[string]string to map[string]interface{}
			labels := make(map[string]interface{}, len(job.Labels))
			for k, v := range job.Labels {
				labels[k] = v
			}

			metrics[i] = &observability.TelemetryMetric{
				ProjectID:   job.ProjectID,
				MetricName:  job.MetricName,
				MetricType:  string(job.MetricType),
				MetricValue: job.MetricValue,
				Labels:      labels,
				Metadata:    job.Metadata,
				Timestamp:   job.Timestamp,
				ProcessedAt: &processedAt,
			}
		}

		if err := w.repository.InsertTelemetryMetricsBatch(ctx, metrics); err != nil {
			w.logger.WithError(err).WithField("count", len(metrics)).Error("Failed to insert telemetry metrics batch")
			// Retry individual metrics
			w.retryBulkMetrics(w.metricsBuffer)
		} else {
			w.logger.WithField("count", len(metrics)).Debug("Telemetry metrics batch processed")
		}

		// Clear buffer
		w.metricsBuffer = w.metricsBuffer[:0]
	}
}

// flushBuffers processes any remaining items in buffers during shutdown
func (w *TelemetryAnalyticsWorker) flushBuffers() {
	w.logger.Info("Flushing remaining buffered items")

	// Process remaining items in queues
	for {
		select {
		case job := <-w.eventQueue:
			if err := w.processTelemetryEvent(job); err != nil {
				w.logger.WithError(err).Error("Failed to process event during flush")
			}
		case job := <-w.batchQueue:
			if err := w.processTelemetryBatch(job); err != nil {
				w.logger.WithError(err).Error("Failed to process batch during flush")
			}
		case job := <-w.metricsQueue:
			if err := w.processTelemetryMetric(job); err != nil {
				w.logger.WithError(err).Error("Failed to process metric during flush")
			}
		default:
			// No more items in queues, process buffers
			w.processBulkOperations()
			return
		}
	}
}

// Error handling methods

// handleEventError handles errors in event processing with retry logic
func (w *TelemetryAnalyticsWorker) handleEventError(job *TelemetryEventJob, err error, logger *logrus.Entry) {
	logger.WithError(err).WithFields(logrus.Fields{
		"event_id":    job.EventID.String(),
		"batch_id":    job.BatchID.String(),
		"retry_count": job.RetryCount,
	}).Error("Failed to process telemetry event")

	// Implement retry logic
	if job.RetryCount < w.maxRetries {
		// Create a copy of the job for retry to avoid race conditions
		retryJob := *job
		retryJob.RetryCount++

		// Exponential backoff
		backoffDelay := w.retryBackoff * time.Duration(1<<uint(retryJob.RetryCount-1))

		go func() {
			time.Sleep(backoffDelay)
			if !w.QueueTelemetryEvent(&retryJob) {
				logger.WithField("event_id", retryJob.EventID.String()).Error("Failed to requeue event after backoff")
			}
		}()

		atomic.AddInt64(&w.stats.EventsRetried, 1)
		logger.WithFields(logrus.Fields{
			"event_id":     retryJob.EventID.String(),
			"retry_count":  retryJob.RetryCount,
			"backoff_delay": backoffDelay,
		}).Info("Retrying telemetry event")
	} else {
		atomic.AddInt64(&w.stats.EventsFailed, 1)
		logger.WithField("event_id", job.EventID.String()).Error("Max retries exceeded for telemetry event")
	}
}

// handleBatchError handles errors in batch processing with retry logic
func (w *TelemetryAnalyticsWorker) handleBatchError(job *TelemetryBatchJob, err error, logger *logrus.Entry) {
	logger.WithError(err).WithFields(logrus.Fields{
		"batch_id":    job.BatchID.String(),
		"retry_count": job.RetryCount,
	}).Error("Failed to process telemetry batch")

	// Implement retry logic
	if job.RetryCount < w.maxRetries {
		// Create a copy of the job for retry to avoid race conditions
		retryJob := *job
		retryJob.RetryCount++

		// Exponential backoff
		backoffDelay := w.retryBackoff * time.Duration(1<<uint(retryJob.RetryCount-1))

		go func() {
			time.Sleep(backoffDelay)
			if !w.QueueTelemetryBatch(&retryJob) {
				logger.WithField("batch_id", retryJob.BatchID.String()).Error("Failed to requeue batch after backoff")
			}
		}()

		atomic.AddInt64(&w.stats.EventsRetried, 1)
		logger.WithFields(logrus.Fields{
			"batch_id":     retryJob.BatchID.String(),
			"retry_count":  retryJob.RetryCount,
			"backoff_delay": backoffDelay,
		}).Info("Retrying telemetry batch")
	} else {
		atomic.AddInt64(&w.stats.EventsFailed, 1)
		logger.WithField("batch_id", job.BatchID.String()).Error("Max retries exceeded for telemetry batch")
	}
}

// handleMetricError handles errors in metric processing with retry logic
func (w *TelemetryAnalyticsWorker) handleMetricError(job *TelemetryMetricJob, err error, logger *logrus.Entry) {
	logger.WithError(err).WithFields(logrus.Fields{
		"metric_name": job.MetricName,
		"retry_count": job.RetryCount,
	}).Error("Failed to process telemetry metric")

	// Implement retry logic
	if job.RetryCount < w.maxRetries {
		// Create a copy of the job for retry to avoid race conditions
		retryJob := *job
		retryJob.RetryCount++

		// Exponential backoff
		backoffDelay := w.retryBackoff * time.Duration(1<<uint(retryJob.RetryCount-1))

		go func() {
			time.Sleep(backoffDelay)
			if !w.QueueTelemetryMetric(&retryJob) {
				logger.WithField("metric_name", retryJob.MetricName).Error("Failed to requeue metric after backoff")
			}
		}()

		atomic.AddInt64(&w.stats.EventsRetried, 1)
		logger.WithFields(logrus.Fields{
			"metric_name":  retryJob.MetricName,
			"retry_count":  retryJob.RetryCount,
			"backoff_delay": backoffDelay,
		}).Info("Retrying telemetry metric")
	} else {
		atomic.AddInt64(&w.stats.EventsFailed, 1)
		logger.WithField("metric_name", job.MetricName).Error("Max retries exceeded for telemetry metric")
	}
}

// Bulk retry methods

// retryBulkEvents retries individual events from a failed bulk operation (thread-safe)
func (w *TelemetryAnalyticsWorker) retryBulkEvents(jobs []*TelemetryEventJob) {
	for _, job := range jobs {
		// Create a copy to avoid race conditions
		retryJob := *job
		retryJob.RetryCount++
		if retryJob.RetryCount <= w.maxRetries {
			go func(j TelemetryEventJob) {
				backoffDelay := w.retryBackoff * time.Duration(1<<uint(j.RetryCount-1))
				time.Sleep(backoffDelay)
				w.QueueTelemetryEvent(&j)
			}(retryJob)
		} else {
			atomic.AddInt64(&w.stats.EventsFailed, 1)
		}
	}
}

// retryBulkBatches retries individual batches from a failed bulk operation (thread-safe)
func (w *TelemetryAnalyticsWorker) retryBulkBatches(jobs []*TelemetryBatchJob) {
	for _, job := range jobs {
		// Create a copy to avoid race conditions
		retryJob := *job
		retryJob.RetryCount++
		if retryJob.RetryCount <= w.maxRetries {
			go func(j TelemetryBatchJob) {
				backoffDelay := w.retryBackoff * time.Duration(1<<uint(j.RetryCount-1))
				time.Sleep(backoffDelay)
				w.QueueTelemetryBatch(&j)
			}(retryJob)
		} else {
			atomic.AddInt64(&w.stats.EventsFailed, 1)
		}
	}
}

// retryBulkMetrics retries individual metrics from a failed bulk operation (thread-safe)
func (w *TelemetryAnalyticsWorker) retryBulkMetrics(jobs []*TelemetryMetricJob) {
	for _, job := range jobs {
		// Create a copy to avoid race conditions
		retryJob := *job
		retryJob.RetryCount++
		if retryJob.RetryCount <= w.maxRetries {
			go func(j TelemetryMetricJob) {
				backoffDelay := w.retryBackoff * time.Duration(1<<uint(j.RetryCount-1))
				time.Sleep(backoffDelay)
				w.QueueTelemetryMetric(&j)
			}(retryJob)
		} else {
			atomic.AddInt64(&w.stats.EventsFailed, 1)
		}
	}
}

// Utility methods

// marshalEventData marshals event data to JSON string for ClickHouse storage
func (w *TelemetryAnalyticsWorker) marshalEventData(data map[string]interface{}) string {
	if data == nil || len(data) == 0 {
		return "{}"
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		w.logger.WithError(err).Warn("Failed to marshal event data")
		return "{}"
	}

	return string(jsonBytes)
}

// marshalLabels marshals labels to JSON string for ClickHouse storage
func (w *TelemetryAnalyticsWorker) marshalLabels(labels map[string]string) string {
	if labels == nil || len(labels) == 0 {
		return "{}"
	}

	jsonBytes, err := json.Marshal(labels)
	if err != nil {
		w.logger.WithError(err).Warn("Failed to marshal labels")
		return "{}"
	}

	return string(jsonBytes)
}

// bufferEvent adds an event to the buffer for bulk processing
func (w *TelemetryAnalyticsWorker) bufferEvent(job *TelemetryEventJob) {
	w.bufferMutex.Lock()
	defer w.bufferMutex.Unlock()

	w.eventBuffer = append(w.eventBuffer, job)

	// Trigger batch processing if buffer is full
	if len(w.eventBuffer) >= w.bufferSize {
		go w.processBulkOperations()
	}
}

// bufferBatch adds a batch to the buffer for bulk processing
func (w *TelemetryAnalyticsWorker) bufferBatch(job *TelemetryBatchJob) {
	w.bufferMutex.Lock()
	defer w.bufferMutex.Unlock()

	w.batchBuffer = append(w.batchBuffer, job)

	// Trigger batch processing if buffer is full
	if len(w.batchBuffer) >= w.bufferSize/10 {
		go w.processBulkOperations()
	}
}

// bufferMetric adds a metric to the buffer for bulk processing
func (w *TelemetryAnalyticsWorker) bufferMetric(job *TelemetryMetricJob) {
	w.bufferMutex.Lock()
	defer w.bufferMutex.Unlock()

	w.metricsBuffer = append(w.metricsBuffer, job)

	// Trigger batch processing if buffer is full
	if len(w.metricsBuffer) >= w.bufferSize/5 {
		go w.processBulkOperations()
	}
}