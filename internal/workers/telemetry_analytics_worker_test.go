package workers

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"brokle/internal/config"
	"brokle/internal/core/domain/observability"
	"brokle/pkg/ulid"
)

// MockAnalyticsRepository mocks the domain analytics repository
type MockAnalyticsRepository struct {
	mock.Mock
}

func (m *MockAnalyticsRepository) InsertTelemetryBatch(ctx context.Context, batch *observability.TelemetryBatch) error {
	args := m.Called(ctx, batch)
	return args.Error(0)
}

func (m *MockAnalyticsRepository) InsertTelemetryBatchesBatch(ctx context.Context, batches []*observability.TelemetryBatch) error {
	args := m.Called(ctx, batches)
	return args.Error(0)
}

func (m *MockAnalyticsRepository) InsertTelemetryMetric(ctx context.Context, metric *observability.TelemetryMetric) error {
	args := m.Called(ctx, metric)
	return args.Error(0)
}

func (m *MockAnalyticsRepository) InsertTelemetryMetricsBatch(ctx context.Context, metrics []*observability.TelemetryMetric) error {
	args := m.Called(ctx, metrics)
	return args.Error(0)
}

// TestTelemetryAnalyticsWorker_QueueProcessing tests basic queue processing functionality
func TestTelemetryAnalyticsWorker_QueueProcessing(t *testing.T) {
	t.Run("successful_batch_processing", func(t *testing.T) {
		worker, mockRepo := setupTestWorker(t)
		defer worker.Stop()

		// Setup
		batchJob := createTestBatchJob()

		mockRepo.On("InsertTelemetryBatch", mock.Anything, mock.MatchedBy(func(batch *observability.TelemetryBatch) bool {
			return batch.ID == batchJob.BatchID &&
				   batch.ProjectID == batchJob.ProjectID &&
				   batch.Status == batchJob.Status
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

		mockRepo.On("InsertTelemetryMetric", mock.Anything, mock.MatchedBy(func(metric *observability.TelemetryMetric) bool {
			return metric.ProjectID == metricJob.ProjectID &&
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

}

// TestTelemetryAnalyticsWorker_HealthMetrics tests health monitoring and metrics
func TestTelemetryAnalyticsWorker_HealthMetrics(t *testing.T) {
	t.Run("health_metrics_available", func(t *testing.T) {
		worker, _ := setupTestWorker(t)
		defer worker.Stop()

		// Get health metrics
		health := worker.GetHealth()
		assert.NotNil(t, health, "Health metrics should be available")

		// Verify initial state
		assert.True(t, health.Healthy, "Worker should start healthy")
		assert.Equal(t, 0, health.QueueDepth, "Queue should start empty")

		// Check queue depths
		depths := worker.GetQueueDepths()
		assert.NotNil(t, depths, "Queue depths should be available")
		assert.Contains(t, depths, "batches", "Should track batch queue")
		assert.Contains(t, depths, "metrics", "Should track metrics queue")

		// Wait for processing
		time.Sleep(200 * time.Millisecond)

		// Check final health
		health = worker.GetHealth()
		assert.True(t, health.Healthy, "Worker should remain healthy after processing")
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
		batchQueue:    make(chan *TelemetryBatchJob, 100),
		metricsQueue:  make(chan *TelemetryMetricJob, 100),
		quit:          make(chan bool),
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

// Test helper to create test telemetry batch job
func createTestBatchJob() *TelemetryBatchJob {
	return &TelemetryBatchJob{
		BatchID:         ulid.New(),
		ProjectID:       ulid.New(),
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