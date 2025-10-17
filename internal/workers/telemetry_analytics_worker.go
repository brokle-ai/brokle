package workers

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"

	"brokle/internal/config"
	"brokle/internal/core/domain/observability"
	"brokle/pkg/ulid"
)

// TelemetryAnalyticsWorker handles high-performance telemetry analytics processing
// Designed for 10k events/sec throughput with comprehensive buffering and error handling
type TelemetryAnalyticsWorker struct {
	config     *config.Config
	logger     *logrus.Logger
	repository observability.TelemetryAnalyticsRepository

	// Channel-based queue system with configurable buffers
	eventQueue    chan *TelemetryEventJob
	batchQueue    chan *TelemetryBatchJob
	metricsQueue  chan *TelemetryMetricJob
	quit          chan bool

	// Buffering system for batch processing
	eventBuffer   []*TelemetryEventJob
	batchBuffer   []*TelemetryBatchJob
	metricsBuffer []*TelemetryMetricJob
	bufferMutex   sync.RWMutex

	// Performance tracking
	stats         *WorkerStats
	statsMutex    sync.RWMutex // Protects stats struct from concurrent access
	healthMetrics *HealthMetrics
	healthMutex   sync.RWMutex // Protects healthMetrics struct from concurrent access

	// Configuration
	maxWorkers      int
	bufferSize      int
	batchInterval   time.Duration
	maxRetries      int
	retryBackoff    time.Duration

	// Actual worker counts (after minimum constraints)
	eventWorkers   int
	batchWorkers   int
	metricsWorkers int

	// Worker pool management
	workerWg    sync.WaitGroup
	running     int64
	lifecycleMu sync.Mutex // Protects start/stop operations
}

// TelemetryEventJob represents a telemetry event processing job
type TelemetryEventJob struct {
	BatchID     ulid.ULID                        `json:"batch_id"`
	EventID     ulid.ULID                        `json:"event_id"`
	ProjectID   ulid.ULID                        `json:"project_id"`
	Environment string                           `json:"environment,omitempty"`
	EventType   observability.TelemetryEventType `json:"event_type"`
	EventData   map[string]interface{}           `json:"event_data"`
	Timestamp   time.Time                        `json:"timestamp"`
	RetryCount  int                              `json:"retry_count"`
	Priority    JobPriority                      `json:"priority"`
}

// TelemetryBatchJob represents a telemetry batch processing job
type TelemetryBatchJob struct {
	BatchID         ulid.ULID                    `json:"batch_id"`
	ProjectID       ulid.ULID                    `json:"project_id"`
	Environment     string                       `json:"environment,omitempty"`
	Status          observability.BatchStatus    `json:"status"`
	TotalEvents     int                          `json:"total_events"`
	ProcessedEvents int                          `json:"processed_events"`
	FailedEvents    int                          `json:"failed_events"`
	ProcessingTime  time.Duration                `json:"processing_time"`
	Metadata        map[string]interface{}       `json:"metadata"`
	Timestamp       time.Time                    `json:"timestamp"`
	RetryCount      int                          `json:"retry_count"`
	Priority        JobPriority                  `json:"priority"`
}

// TelemetryMetricJob represents a telemetry metric processing job
type TelemetryMetricJob struct {
	ProjectID      ulid.ULID              `json:"project_id"`
	Environment    string                 `json:"environment,omitempty"`
	MetricName     string                 `json:"metric_name"`
	MetricType     MetricType             `json:"metric_type"`
	MetricValue    float64                `json:"metric_value"`
	Labels         map[string]string      `json:"labels"`
	Metadata       map[string]interface{} `json:"metadata"`
	Timestamp      time.Time              `json:"timestamp"`
	RetryCount     int                    `json:"retry_count"`
	Priority       JobPriority            `json:"priority"`
}

// JobPriority defines the priority levels for telemetry jobs
type JobPriority int

const (
	PriorityLow JobPriority = iota
	PriorityNormal
	PriorityHigh
	PriorityCritical
)

// MetricType defines the type of telemetry metric
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
	MetricTypeSummary   MetricType = "summary"
)

// WorkerStats tracks performance statistics
type WorkerStats struct {
	EventsProcessed    int64     `json:"events_processed"`
	BatchesProcessed   int64     `json:"batches_processed"`
	MetricsProcessed   int64     `json:"metrics_processed"`
	EventsDropped      int64     `json:"events_dropped"`
	EventsRetried      int64     `json:"events_retried"`
	EventsFailed       int64     `json:"events_failed"`
	TotalProcessingTime time.Duration `json:"total_processing_time"`
	AverageLatency     time.Duration `json:"average_latency"`
	ThroughputPerSec   float64   `json:"throughput_per_sec"`
	LastProcessedTime  time.Time `json:"last_processed_time"`
	StartTime          time.Time `json:"start_time"`
}

// HealthMetrics tracks worker health metrics
type HealthMetrics struct {
	Healthy           bool          `json:"healthy"`
	ActiveWorkers     int           `json:"active_workers"`
	QueueDepth        int           `json:"queue_depth"`
	BufferUtilization float64       `json:"buffer_utilization"`
	ErrorRate         float64       `json:"error_rate"`
	MemoryUsage       int64         `json:"memory_usage_bytes"`
	LastHeartbeat     time.Time     `json:"last_heartbeat"`
}

// NewTelemetryAnalyticsWorker creates a new high-performance telemetry analytics worker
func NewTelemetryAnalyticsWorker(
	config *config.Config,
	logger *logrus.Logger,
	repository observability.TelemetryAnalyticsRepository,
) *TelemetryAnalyticsWorker {
	// Configure worker parameters for high throughput
	maxWorkers := config.Workers.AnalyticsWorkers
	if maxWorkers == 0 {
		maxWorkers = 8 // Default for high throughput
	}

	bufferSize := 4500 // Slightly below ClickHouse connection pool capacity
	// TODO: Add AnalyticsBufferSize to config if needed

	batchInterval := 2 * time.Second // Optimized for high throughput
	// TODO: Add AnalyticsBatchInterval to config if needed

	return &TelemetryAnalyticsWorker{
		config:     config,
		logger:     logger,
		repository: repository,

		// High-capacity queues for 10k events/sec
		eventQueue:   make(chan *TelemetryEventJob, bufferSize),
		batchQueue:   make(chan *TelemetryBatchJob, bufferSize/10), // Fewer batches than events
		metricsQueue: make(chan *TelemetryMetricJob, bufferSize/5), // Metrics in between
		quit:         make(chan bool),

		// Pre-allocated buffers for batch processing
		eventBuffer:   make([]*TelemetryEventJob, 0, bufferSize),
		batchBuffer:   make([]*TelemetryBatchJob, 0, bufferSize/10),
		metricsBuffer: make([]*TelemetryMetricJob, 0, bufferSize/5),

		// Performance tracking
		stats: &WorkerStats{
			StartTime: time.Now(),
		},
		healthMetrics: &HealthMetrics{
			Healthy:       true,
			LastHeartbeat: time.Now(),
		},

		// Configuration
		maxWorkers:    maxWorkers,
		bufferSize:    bufferSize,
		batchInterval: batchInterval,
		maxRetries:    3,
		retryBackoff:  500 * time.Millisecond,
	}
}

// Start starts the telemetry analytics worker with full worker pool
func (w *TelemetryAnalyticsWorker) Start() {
	w.lifecycleMu.Lock()
	defer w.lifecycleMu.Unlock()
	w.logger.WithFields(logrus.Fields{
		"max_workers":     w.maxWorkers,
		"buffer_size":     w.bufferSize,
		"batch_interval":  w.batchInterval,
		"max_retries":     w.maxRetries,
	}).Info("Starting telemetry analytics worker")

	atomic.StoreInt64(&w.running, 1)

	// Calculate actual worker counts with minimum constraints
	// Ensure at least 1 batch and 1 metrics worker to prevent queue starvation
	// in small deployments (e.g., maxWorkers=1-3 would cause 0 batch/metrics workers)
	w.eventWorkers = w.maxWorkers
	w.batchWorkers = w.maxWorkers / 2
	if w.batchWorkers == 0 {
		w.batchWorkers = 1 // Minimum 1 to drain batch queue
	}
	w.metricsWorkers = w.maxWorkers / 4
	if w.metricsWorkers == 0 {
		w.metricsWorkers = 1 // Minimum 1 to drain metrics queue
	}

	// Start event processing workers
	for i := 0; i < w.eventWorkers; i++ {
		w.workerWg.Add(1)
		go w.eventWorker(i)
	}

	// Start batch processing workers
	for i := 0; i < w.batchWorkers; i++ {
		w.workerWg.Add(1)
		go w.batchWorker(i)
	}

	// Start metrics processing workers
	for i := 0; i < w.metricsWorkers; i++ {
		w.workerWg.Add(1)
		go w.metricsWorker(i)
	}

	// Start batch processor for efficient bulk operations
	w.workerWg.Add(1)
	go w.batchProcessor()

	// Start health monitor
	w.workerWg.Add(1)
	go w.healthMonitor()

	w.logger.Info("Telemetry analytics worker started successfully")
}

// Stop gracefully stops the telemetry analytics worker
func (w *TelemetryAnalyticsWorker) Stop() {
	w.lifecycleMu.Lock()
	defer w.lifecycleMu.Unlock()

	w.logger.Info("Stopping telemetry analytics worker")

	atomic.StoreInt64(&w.running, 0)
	close(w.quit)

	// Wait for all workers to finish processing current jobs
	w.workerWg.Wait()

	// Process any remaining buffered items
	w.flushBuffers()

	// Log final stats with proper synchronization
	w.statsMutex.RLock()
	finalStats := logrus.Fields{
		"events_processed":  w.stats.EventsProcessed,
		"batches_processed": w.stats.BatchesProcessed,
		"metrics_processed": w.stats.MetricsProcessed,
		"events_dropped":    w.stats.EventsDropped,
		"uptime":           time.Since(w.stats.StartTime),
	}
	w.statsMutex.RUnlock()

	w.logger.WithFields(finalStats).Info("Telemetry analytics worker stopped")
}

// QueueTelemetryEvent queues a telemetry event for processing
func (w *TelemetryAnalyticsWorker) QueueTelemetryEvent(job *TelemetryEventJob) bool {
	if atomic.LoadInt64(&w.running) == 0 {
		return false
	}

	select {
	case w.eventQueue <- job:
		w.logger.WithFields(logrus.Fields{
			"batch_id":   job.BatchID.String(),
			"event_id":   job.EventID.String(),
			"event_type": job.EventType,
			"priority":   job.Priority,
		}).Debug("Telemetry event queued")
		return true
	default:
		// Queue is full, handle based on priority
		if job.Priority >= PriorityHigh {
			// Try to drop a lower priority job
			return w.tryPriorityDrop(job)
		}

		atomic.AddInt64(&w.stats.EventsDropped, 1)
		w.logger.WithFields(logrus.Fields{
			"batch_id":   job.BatchID.String(),
			"event_id":   job.EventID.String(),
			"event_type": job.EventType,
			"queue_size": len(w.eventQueue),
		}).Warn("Event queue full, dropping telemetry event")
		return false
	}
}

// QueueTelemetryBatch queues a telemetry batch for processing
func (w *TelemetryAnalyticsWorker) QueueTelemetryBatch(job *TelemetryBatchJob) bool {
	if atomic.LoadInt64(&w.running) == 0 {
		return false
	}

	select {
	case w.batchQueue <- job:
		w.logger.WithFields(logrus.Fields{
			"batch_id":       job.BatchID.String(),
			"total_events":   job.TotalEvents,
			"status":         job.Status,
			"priority":       job.Priority,
		}).Debug("Telemetry batch queued")
		return true
	default:
		atomic.AddInt64(&w.stats.EventsDropped, 1)
		w.logger.WithFields(logrus.Fields{
			"batch_id":   job.BatchID.String(),
			"queue_size": len(w.batchQueue),
		}).Warn("Batch queue full, dropping telemetry batch")
		return false
	}
}

// QueueTelemetryMetric queues a telemetry metric for processing
func (w *TelemetryAnalyticsWorker) QueueTelemetryMetric(job *TelemetryMetricJob) bool {
	if atomic.LoadInt64(&w.running) == 0 {
		return false
	}

	select {
	case w.metricsQueue <- job:
		w.logger.WithFields(logrus.Fields{
			"metric_name": job.MetricName,
			"metric_type": job.MetricType,
			"value":       job.MetricValue,
			"priority":    job.Priority,
		}).Debug("Telemetry metric queued")
		return true
	default:
		atomic.AddInt64(&w.stats.EventsDropped, 1)
		w.logger.WithFields(logrus.Fields{
			"metric_name": job.MetricName,
			"queue_size":  len(w.metricsQueue),
		}).Warn("Metrics queue full, dropping telemetry metric")
		return false
	}
}

// eventWorker processes telemetry events from the queue
func (w *TelemetryAnalyticsWorker) eventWorker(id int) {
	defer w.workerWg.Done()

	logger := w.logger.WithField("worker_type", "event").WithField("worker_id", id)
	logger.Info("Event worker started")

	for {
		select {
		case job := <-w.eventQueue:
			startTime := time.Now()

			if err := w.processTelemetryEvent(job); err != nil {
				w.handleEventError(job, err, logger)
			} else {
				atomic.AddInt64(&w.stats.EventsProcessed, 1)
				w.updateLatencyStats(time.Since(startTime))
			}

		case <-w.quit:
			logger.Info("Event worker stopping")
			return
		}
	}
}

// batchWorker processes telemetry batches from the queue
func (w *TelemetryAnalyticsWorker) batchWorker(id int) {
	defer w.workerWg.Done()

	logger := w.logger.WithField("worker_type", "batch").WithField("worker_id", id)
	logger.Info("Batch worker started")

	for {
		select {
		case job := <-w.batchQueue:
			startTime := time.Now()

			if err := w.processTelemetryBatch(job); err != nil {
				w.handleBatchError(job, err, logger)
			} else {
				atomic.AddInt64(&w.stats.BatchesProcessed, 1)
				w.updateLatencyStats(time.Since(startTime))
			}

		case <-w.quit:
			logger.Info("Batch worker stopping")
			return
		}
	}
}

// metricsWorker processes telemetry metrics from the queue
func (w *TelemetryAnalyticsWorker) metricsWorker(id int) {
	defer w.workerWg.Done()

	logger := w.logger.WithField("worker_type", "metrics").WithField("worker_id", id)
	logger.Info("Metrics worker started")

	for {
		select {
		case job := <-w.metricsQueue:
			startTime := time.Now()

			if err := w.processTelemetryMetric(job); err != nil {
				w.handleMetricError(job, err, logger)
			} else {
				atomic.AddInt64(&w.stats.MetricsProcessed, 1)
				w.updateLatencyStats(time.Since(startTime))
			}

		case <-w.quit:
			logger.Info("Metrics worker stopping")
			return
		}
	}
}

// batchProcessor handles bulk operations for improved ClickHouse performance
func (w *TelemetryAnalyticsWorker) batchProcessor() {
	defer w.workerWg.Done()

	ticker := time.NewTicker(w.batchInterval)
	defer ticker.Stop()

	logger := w.logger.WithField("worker_type", "batch_processor")
	logger.Info("Batch processor started")

	for {
		select {
		case <-ticker.C:
			w.processBulkOperations()

		case <-w.quit:
			// Process remaining items before stopping
			w.processBulkOperations()
			logger.Info("Batch processor stopping")
			return
		}
	}
}

// healthMonitor monitors worker health and updates metrics
func (w *TelemetryAnalyticsWorker) healthMonitor() {
	defer w.workerWg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	logger := w.logger.WithField("worker_type", "health_monitor")
	logger.Info("Health monitor started")

	for {
		select {
		case <-ticker.C:
			w.updateHealthMetrics()

		case <-w.quit:
			logger.Info("Health monitor stopping")
			return
		}
	}
}

// tryPriorityDrop attempts to drop a lower priority job to make room for a high priority one
func (w *TelemetryAnalyticsWorker) tryPriorityDrop(newJob *TelemetryEventJob) bool {
	// This is a simplified version - in production, you might want a more sophisticated priority queue
	select {
	case oldJob := <-w.eventQueue:
		if oldJob.Priority < newJob.Priority {
			// Drop the old job and queue the new one
			select {
			case w.eventQueue <- newJob:
				atomic.AddInt64(&w.stats.EventsDropped, 1)
				w.logger.WithFields(logrus.Fields{
					"dropped_event_id": oldJob.EventID.String(),
					"new_event_id":     newJob.EventID.String(),
				}).Debug("Dropped lower priority event for higher priority")
				return true
			default:
				// Put the old job back and fail
				select {
				case w.eventQueue <- oldJob:
				default:
					atomic.AddInt64(&w.stats.EventsDropped, 1)
				}
				return false
			}
		} else {
			// Put the old job back
			select {
			case w.eventQueue <- oldJob:
			default:
				atomic.AddInt64(&w.stats.EventsDropped, 1)
			}
			return false
		}
	default:
		return false
	}
}

// updateLatencyStats updates latency statistics (thread-safe)
func (w *TelemetryAnalyticsWorker) updateLatencyStats(duration time.Duration) {
	w.statsMutex.Lock()
	defer w.statsMutex.Unlock()

	// Simple moving average calculation
	currentAvg := w.stats.AverageLatency
	if currentAvg == 0 {
		w.stats.AverageLatency = duration
	} else {
		// Exponential moving average with alpha = 0.1
		w.stats.AverageLatency = time.Duration(0.9*float64(currentAvg) + 0.1*float64(duration))
	}
	w.stats.LastProcessedTime = time.Now()
}

// updateHealthMetrics updates health metrics (thread-safe)
func (w *TelemetryAnalyticsWorker) updateHealthMetrics() {
	// Calculate queue metrics (no locking needed for channel length operations)
	queueDepth := len(w.eventQueue) + len(w.batchQueue) + len(w.metricsQueue)
	activeWorkers := w.eventWorkers + w.batchWorkers + w.metricsWorkers + 1
	bufferUtilization := float64(queueDepth) / float64(w.bufferSize) * 100

	// Access stats with proper locking
	w.statsMutex.Lock()
	// Calculate error rate (as fraction 0.0-1.0)
	totalProcessed := w.stats.EventsProcessed + w.stats.BatchesProcessed + w.stats.MetricsProcessed
	var errorRate float64
	if totalProcessed > 0 {
		errorRate = float64(w.stats.EventsFailed) / float64(totalProcessed)
	}

	// Calculate throughput
	uptime := time.Since(w.stats.StartTime)
	if uptime.Seconds() > 0 {
		w.stats.ThroughputPerSec = float64(totalProcessed) / uptime.Seconds()
	}
	w.statsMutex.Unlock()

	// Update health metrics with proper locking
	w.healthMutex.Lock()
	w.healthMetrics.ActiveWorkers = activeWorkers
	w.healthMetrics.QueueDepth = queueDepth
	w.healthMetrics.BufferUtilization = bufferUtilization
	w.healthMetrics.ErrorRate = errorRate
	w.healthMetrics.Healthy = w.healthMetrics.ErrorRate < 0.05 && w.healthMetrics.BufferUtilization < 90.0
	w.healthMetrics.LastHeartbeat = time.Now()
	w.healthMutex.Unlock()
}

// GetStats returns current worker statistics (thread-safe copy)
func (w *TelemetryAnalyticsWorker) GetStats() *WorkerStats {
	w.statsMutex.RLock()
	defer w.statsMutex.RUnlock()

	// Return a copy to avoid external modifications
	statsCopy := *w.stats
	return &statsCopy
}

// GetHealth returns current worker health metrics (thread-safe copy)
func (w *TelemetryAnalyticsWorker) GetHealth() *HealthMetrics {
	w.healthMutex.RLock()
	healthCopy := &HealthMetrics{
		ActiveWorkers:     w.healthMetrics.ActiveWorkers,
		QueueDepth:        w.healthMetrics.QueueDepth,
		BufferUtilization: w.healthMetrics.BufferUtilization,
		ErrorRate:         w.healthMetrics.ErrorRate,
		Healthy:           w.healthMetrics.Healthy,
		LastHeartbeat:     w.healthMetrics.LastHeartbeat,
	}
	w.healthMutex.RUnlock()
	return healthCopy
}

// GetQueueDepths returns current queue depths for monitoring
func (w *TelemetryAnalyticsWorker) GetQueueDepths() map[string]int {
	return map[string]int{
		"events":  len(w.eventQueue),
		"batches": len(w.batchQueue),
		"metrics": len(w.metricsQueue),
	}
}