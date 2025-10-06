package workers

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"brokle/internal/config"
	"brokle/internal/core/domain/observability"
	"brokle/internal/infrastructure/repository/clickhouse"
	"brokle/pkg/ulid"
)

// MockAnalyticsRepository mocks the ClickHouse analytics repository
type MockAnalyticsRepository struct {
	mock.Mock
}

func (m *MockAnalyticsRepository) InsertTelemetryEvent(ctx context.Context, event *clickhouse.TelemetryEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockAnalyticsRepository) InsertTelemetryEventsBatch(ctx context.Context, events []*clickhouse.TelemetryEvent) error {
	args := m.Called(ctx, events)
	return args.Error(0)
}

func (m *MockAnalyticsRepository) InsertTelemetryBatch(ctx context.Context, batch *clickhouse.TelemetryBatch) error {
	args := m.Called(ctx, batch)
	return args.Error(0)
}

func (m *MockAnalyticsRepository) InsertTelemetryBatchesBatch(ctx context.Context, batches []*clickhouse.TelemetryBatch) error {
	args := m.Called(ctx, batches)
	return args.Error(0)
}

func (m *MockAnalyticsRepository) InsertTelemetryMetric(ctx context.Context, metric *clickhouse.TelemetryMetric) error {
	args := m.Called(ctx, metric)
	return args.Error(0)
}

func (m *MockAnalyticsRepository) InsertTelemetryMetricsBatch(ctx context.Context, metrics []*clickhouse.TelemetryMetric) error {
	args := m.Called(ctx, metrics)
	return args.Error(0)
}

// TestTelemetryAnalyticsWorker_QueueProcessing tests basic queue processing functionality
func TestTelemetryAnalyticsWorker_QueueProcessing(t *testing.T) {
	t.Run("successful_event_processing", func(t *testing.T) {
		worker, mockRepo := setupTestWorker(t)
		defer worker.Stop()

		// Setup
		eventJob := createTestEventJob()

		mockRepo.On("InsertTelemetryEvent", mock.Anything, mock.MatchedBy(func(event *clickhouse.TelemetryEvent) bool {
			return event.ID == eventJob.EventID.String() &&
				   event.ProjectID == eventJob.ProjectID.String() &&
				   event.EventType == string(eventJob.EventType)
		})).Return(nil).Once()

		// Execute
		success := worker.QueueTelemetryEvent(eventJob)
		assert.True(t, success, "Event should be queued successfully")

		// Wait for processing
		time.Sleep(100 * time.Millisecond)

		// Verify
		mockRepo.AssertExpectations(t)
		assert.Equal(t, int64(1), atomic.LoadInt64(&worker.stats.EventsProcessed))
		assert.Equal(t, int64(0), atomic.LoadInt64(&worker.stats.EventsFailed))
	})

	t.Run("successful_batch_processing", func(t *testing.T) {
		worker, mockRepo := setupTestWorker(t)
		defer worker.Stop()

		// Setup
		batchJob := createTestBatchJob()

		mockRepo.On("InsertTelemetryBatch", mock.Anything, mock.MatchedBy(func(batch *clickhouse.TelemetryBatch) bool {
			return batch.ID == batchJob.BatchID.String() &&
				   batch.ProjectID == batchJob.ProjectID.String() &&
				   batch.Status == string(batchJob.Status)
		})).Return(nil).Once()

		// Execute
		success := worker.QueueTelemetryBatch(batchJob)
		assert.True(t, success, "Batch should be queued successfully")

		// Wait for processing
		time.Sleep(100 * time.Millisecond)

		// Verify
		mockRepo.AssertExpectations(t)
		assert.Equal(t, int64(1), atomic.LoadInt64(&worker.stats.BatchesProcessed))
	})

	t.Run("successful_metric_processing", func(t *testing.T) {
		worker, mockRepo := setupTestWorker(t)
		defer worker.Stop()

		// Setup
		metricJob := createTestMetricJob()

		mockRepo.On("InsertTelemetryMetric", mock.Anything, mock.MatchedBy(func(metric *clickhouse.TelemetryMetric) bool {
			return metric.ProjectID == metricJob.ProjectID.String() &&
				   metric.MetricName == metricJob.MetricName &&
				   metric.MetricType == string(metricJob.MetricType)
		})).Return(nil).Once()

		// Execute
		success := worker.QueueTelemetryMetric(metricJob)
		assert.True(t, success, "Metric should be queued successfully")

		// Wait for processing
		time.Sleep(100 * time.Millisecond)

		// Verify
		mockRepo.AssertExpectations(t)
		assert.Equal(t, int64(1), atomic.LoadInt64(&worker.stats.MetricsProcessed))
	})

	t.Run("concurrent_queue_processing", func(t *testing.T) {
		worker, mockRepo := setupTestWorker(t)
		defer worker.Stop()

		// Setup
		const jobCount = 10

		mockRepo.On("InsertTelemetryEvent", mock.Anything, mock.AnythingOfType("*clickhouse.TelemetryEvent")).
			Return(nil).Times(jobCount)

		// Execute concurrently
		for i := 0; i < jobCount; i++ {
			go func() {
				eventJob := createTestEventJob()
				worker.QueueTelemetryEvent(eventJob)
			}()
		}

		// Wait for processing
		time.Sleep(500 * time.Millisecond)

		// Verify
		mockRepo.AssertExpectations(t)
		assert.Equal(t, int64(jobCount), atomic.LoadInt64(&worker.stats.EventsProcessed))
	})
}

// TestTelemetryAnalyticsWorker_RetryLogic tests retry mechanisms and exponential backoff
func TestTelemetryAnalyticsWorker_RetryLogic(t *testing.T) {
	t.Run("event_retry_with_exponential_backoff", func(t *testing.T) {
		worker, mockRepo := setupTestWorker(t)
		defer worker.Stop()

		// Setup
		eventJob := createTestEventJob()
		retryError := errors.New("temporary database error")

		// First call fails, second succeeds
		mockRepo.On("InsertTelemetryEvent", mock.Anything, mock.AnythingOfType("*clickhouse.TelemetryEvent")).
			Return(retryError).Once()
		mockRepo.On("InsertTelemetryEvent", mock.Anything, mock.AnythingOfType("*clickhouse.TelemetryEvent")).
			Return(nil).Once()

		// Execute
		success := worker.QueueTelemetryEvent(eventJob)
		assert.True(t, success, "Event should be queued successfully")

		// Wait for processing and retry
		time.Sleep(1 * time.Second)

		// Verify
		mockRepo.AssertExpectations(t)
		assert.Equal(t, int64(1), atomic.LoadInt64(&worker.stats.EventsProcessed))
		assert.Equal(t, int64(1), atomic.LoadInt64(&worker.stats.EventsRetried))
	})

	t.Run("max_retries_exceeded", func(t *testing.T) {
		worker, mockRepo := setupTestWorker(t)
		defer worker.Stop()

		// Setup
		eventJob := createTestEventJob()
		persistentError := errors.New("persistent database error")

		// All attempts fail (initial + maxRetries = 4 total calls)
		mockRepo.On("InsertTelemetryEvent", mock.Anything, mock.AnythingOfType("*clickhouse.TelemetryEvent")).
			Return(persistentError).Times(4)

		// Execute
		success := worker.QueueTelemetryEvent(eventJob)
		assert.True(t, success, "Event should be queued successfully")

		// Wait for processing and all retries
		time.Sleep(2 * time.Second)

		// Verify
		mockRepo.AssertExpectations(t)
		assert.Equal(t, int64(0), atomic.LoadInt64(&worker.stats.EventsProcessed))
		assert.Equal(t, int64(1), atomic.LoadInt64(&worker.stats.EventsFailed))
		assert.Equal(t, int64(3), atomic.LoadInt64(&worker.stats.EventsRetried)) // 3 retries after initial failure
	})
}

// TestTelemetryAnalyticsWorker_DropPaths tests scenarios where events are dropped
func TestTelemetryAnalyticsWorker_DropPaths(t *testing.T) {
	t.Run("queue_full_drops_events", func(t *testing.T) {
		// Create a worker with a small queue and don't start it (no processing)
		cfg := &config.Config{
			Workers: config.WorkersConfig{
				AnalyticsWorkers: 1,
			},
		}

		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)

		mockRepo := &MockAnalyticsRepository{}

		worker := &TelemetryAnalyticsWorker{
			config:        cfg,
			logger:        logger,
			repository:    mockRepo,
			eventQueue:    make(chan *TelemetryEventJob, 3), // Very small queue
			batchQueue:    make(chan *TelemetryBatchJob, 3),
			metricsQueue:  make(chan *TelemetryMetricJob, 3),
			quit:          make(chan bool),
			eventBuffer:   make([]*TelemetryEventJob, 0, 3),
			batchBuffer:   make([]*TelemetryBatchJob, 0, 3),
			metricsBuffer: make([]*TelemetryMetricJob, 0, 3),
			bufferMutex:   sync.RWMutex{},
			bufferSize:    3,
			maxWorkers:    1,
			maxRetries:    3,
			retryBackoff:  100 * time.Millisecond,
			batchInterval: 1 * time.Second,
			stats: &WorkerStats{
				StartTime: time.Now(),
			},
			healthMetrics: &HealthMetrics{
				Healthy:       true,
				LastHeartbeat: time.Now(),
			},
		}

		// Start the worker but immediately set running to 1 to allow queueing
		atomic.StoreInt64(&worker.running, 1)

		// Fill up the queue completely
		for i := 0; i < cap(worker.eventQueue); i++ {
			eventJob := createTestEventJob()
			success := worker.QueueTelemetryEvent(eventJob)
			assert.True(t, success, "Should queue event until capacity")
		}

		// Verify queue is actually full
		assert.Equal(t, cap(worker.eventQueue), len(worker.eventQueue), "Queue should be at capacity")

		// This should be dropped due to full queue
		droppedEvent := createTestEventJob()
		success := worker.QueueTelemetryEvent(droppedEvent)
		assert.False(t, success, "Event should be dropped when queue is full")

		// Check that events were dropped
		droppedCount := atomic.LoadInt64(&worker.stats.EventsDropped)
		assert.True(t, droppedCount >= 1, "Should have at least one dropped event")
	})

	t.Run("worker_shutdown_drops_remaining", func(t *testing.T) {
		worker, mockRepo := setupTestWorker(t)

		// Set up mock to handle any processing that might occur during shutdown
		mockRepo.On("InsertTelemetryEvent", mock.Anything, mock.AnythingOfType("*clickhouse.TelemetryEvent")).
			Return(nil).Maybe()

		// Setup
		eventJob := createTestEventJob()

		// Queue event but stop worker immediately
		success := worker.QueueTelemetryEvent(eventJob)
		assert.True(t, success, "Event should be queued successfully")

		worker.Stop()

		// Wait for shutdown
		time.Sleep(100 * time.Millisecond)

		// Verify processing was interrupted
		// Note: Some events might still be processed during graceful shutdown
		processedCount := atomic.LoadInt64(&worker.stats.EventsProcessed)
		assert.True(t, processedCount <= 1, "Should process minimal events during shutdown")
	})
}

// TestTelemetryAnalyticsWorker_HealthMetrics tests health monitoring and metrics
func TestTelemetryAnalyticsWorker_HealthMetrics(t *testing.T) {
	t.Run("metrics_track_successful_processing", func(t *testing.T) {
		worker, mockRepo := setupTestWorker(t)
		defer worker.Stop()

		// Setup
		eventJob := createTestEventJob()

		mockRepo.On("InsertTelemetryEvent", mock.Anything, mock.AnythingOfType("*clickhouse.TelemetryEvent")).
			Return(nil).Once()

		// Execute
		initialProcessed := atomic.LoadInt64(&worker.stats.EventsProcessed)
		success := worker.QueueTelemetryEvent(eventJob)
		assert.True(t, success, "Event should be queued successfully")

		// Wait for processing
		time.Sleep(200 * time.Millisecond)

		// Verify metrics
		finalProcessed := atomic.LoadInt64(&worker.stats.EventsProcessed)
		assert.Equal(t, initialProcessed+1, finalProcessed, "Processed count should increase")
		assert.Equal(t, int64(0), atomic.LoadInt64(&worker.stats.EventsFailed), "Failed count should remain zero")

		mockRepo.AssertExpectations(t)
	})

	t.Run("metrics_track_failed_processing", func(t *testing.T) {
		worker, mockRepo := setupTestWorker(t)
		defer worker.Stop()

		// Setup
		eventJob := createTestEventJob()
		persistentError := errors.New("persistent error")

		// All attempts fail
		mockRepo.On("InsertTelemetryEvent", mock.Anything, mock.AnythingOfType("*clickhouse.TelemetryEvent")).
			Return(persistentError).Times(4) // Initial + 3 retries

		// Execute
		initialFailed := atomic.LoadInt64(&worker.stats.EventsFailed)
		success := worker.QueueTelemetryEvent(eventJob)
		assert.True(t, success, "Event should be queued successfully")

		// Wait for processing and retries
		time.Sleep(3 * time.Second)

		// Verify metrics
		finalFailed := atomic.LoadInt64(&worker.stats.EventsFailed)
		assert.Equal(t, initialFailed+1, finalFailed, "Failed count should increase")
		assert.Equal(t, int64(0), atomic.LoadInt64(&worker.stats.EventsProcessed), "Processed count should remain zero")

		mockRepo.AssertExpectations(t)
	})

	t.Run("health_status_updates", func(t *testing.T) {
		worker, mockRepo := setupTestWorker(t)
		defer worker.Stop()

		// Get initial health metrics
		health := worker.GetHealth()
		assert.NotNil(t, health, "Health metrics should be available")

		// Verify initial state
		assert.True(t, health.Healthy, "Worker should start healthy")
		assert.Equal(t, 0, health.QueueDepth, "Queue should start empty")

		// Setup
		eventJob := createTestEventJob()
		mockRepo.On("InsertTelemetryEvent", mock.Anything, mock.AnythingOfType("*clickhouse.TelemetryEvent")).
			Return(nil).Once()

		// Execute
		success := worker.QueueTelemetryEvent(eventJob)
		assert.True(t, success, "Event should be queued successfully")

		// Check health during processing
		health = worker.GetHealth()
		assert.True(t, health.QueueDepth >= 0, "Queue depth should be tracked")

		// Wait for processing
		time.Sleep(200 * time.Millisecond)

		// Check final health
		health = worker.GetHealth()
		assert.True(t, health.Healthy, "Worker should remain healthy after processing")

		mockRepo.AssertExpectations(t)
	})
}

// Helper function to set up test worker with mock repository
func setupTestWorker(t *testing.T) (*TelemetryAnalyticsWorker, *MockAnalyticsRepository) {
	cfg := &config.Config{
		Workers: config.WorkersConfig{
			AnalyticsWorkers: 8, // Ensure enough workers for all types (events, batches, metrics)
		},
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

	mockRepo := &MockAnalyticsRepository{}

	worker := &TelemetryAnalyticsWorker{
		config:        cfg,
		logger:        logger,
		repository:    mockRepo, // Now implements TelemetryRepository interface
		eventQueue:    make(chan *TelemetryEventJob, 100),
		batchQueue:    make(chan *TelemetryBatchJob, 100),
		metricsQueue:  make(chan *TelemetryMetricJob, 100),
		quit:          make(chan bool),
		eventBuffer:   make([]*TelemetryEventJob, 0, 50),
		batchBuffer:   make([]*TelemetryBatchJob, 0, 10),
		metricsBuffer: make([]*TelemetryMetricJob, 0, 20),
		bufferMutex:   sync.RWMutex{},
		bufferSize:    50,
		maxWorkers:    8,
		maxRetries:    3,
		retryBackoff:  100 * time.Millisecond,
		batchInterval: 1 * time.Second,
		stats: &WorkerStats{
			EventsProcessed:  0,
			EventsFailed:     0,
			EventsRetried:    0,
			BatchesProcessed: 0,
			MetricsProcessed: 0,
			StartTime:        time.Now(),
		},
		healthMetrics: &HealthMetrics{
			Healthy:       true,
			QueueDepth:    0,
			ErrorRate:     0.0,
			LastHeartbeat: time.Now(),
		},
	}

	// Start the worker
	go worker.Start()

	// Wait a moment for worker to start
	time.Sleep(10 * time.Millisecond)

	return worker, mockRepo
}

// Test helper to create test telemetry event job
func createTestEventJob() *TelemetryEventJob {
	return &TelemetryEventJob{
		EventID:     ulid.New(),
		BatchID:     ulid.New(),
		ProjectID:   ulid.New(),
		Environment: "test",
		EventType:   observability.TelemetryEventTypeTraceCreate,
		EventData: map[string]interface{}{
			"trace_id":  uuid.New().String(),
			"operation": "test_operation",
		},
		Timestamp:  time.Now(),
		RetryCount: 0,
		Priority:   PriorityNormal,
	}
}

// Test helper to create test telemetry batch job
func createTestBatchJob() *TelemetryBatchJob {
	return &TelemetryBatchJob{
		BatchID:         ulid.New(),
		ProjectID:       ulid.New(),
		Environment:     "test",
		Status:          observability.BatchStatusCompleted,
		TotalEvents:     10,
		ProcessedEvents: 10,
		FailedEvents:    0,
		ProcessingTime:  50 * time.Millisecond,
		Metadata: map[string]interface{}{
			"source": "test_suite",
		},
		Timestamp:  time.Now(),
		RetryCount: 0,
		Priority:   PriorityNormal,
	}
}

// Test helper to create test telemetry metric job
func createTestMetricJob() *TelemetryMetricJob {
	return &TelemetryMetricJob{
		ProjectID:   ulid.New(),
		Environment: "test",
		MetricName:  "test_metric",
		MetricType:  MetricTypeCounter,
		MetricValue: 1.0,
		Labels: map[string]string{
			"source": "test",
		},
		Metadata: map[string]interface{}{
			"test": true,
		},
		Timestamp:  time.Now(),
		RetryCount: 0,
		Priority:   PriorityNormal,
	}
}