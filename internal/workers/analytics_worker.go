package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"brokle/internal/config"
	"brokle/internal/infrastructure/repository/observability"
)

// AnalyticsWorker handles async analytics data processing
type AnalyticsWorker struct {
	config     *config.Config
	logger     *logrus.Logger
	repository *observability.AnalyticsRepository
	queue      chan AnalyticsJob
	quit       chan bool
}

// AnalyticsJob represents an analytics processing job
type AnalyticsJob struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	Retry     int         `json:"retry"`
}

// RequestLogJob represents a request log processing job
type RequestLogJob struct {
	RequestID      string    `json:"request_id"`
	UserID         string    `json:"user_id"`
	OrganizationID string    `json:"organization_id"`
	ProjectID      string    `json:"project_id"`
	Environment    string    `json:"environment"`
	Provider       string    `json:"provider"`
	Model          string    `json:"model"`
	Method         string    `json:"method"`
	Endpoint       string    `json:"endpoint"`
	StatusCode     int       `json:"status_code"`
	InputTokens    int64     `json:"input_tokens"`
	OutputTokens   int64     `json:"output_tokens"`
	TotalTokens    int64     `json:"total_tokens"`
	Cost           float64   `json:"cost"`
	Latency        int64     `json:"latency"`
	QualityScore   float64   `json:"quality_score"`
	Cached         bool      `json:"cached"`
	Error          string    `json:"error"`
	Timestamp      time.Time `json:"timestamp"`
}

// MetricJob represents a metric processing job
type MetricJob struct {
	OrganizationID string                 `json:"organization_id"`
	ProjectID      string                 `json:"project_id"`
	Environment    string                 `json:"environment"`
	MetricName     string                 `json:"metric_name"`
	MetricValue    float64                `json:"metric_value"`
	Tags           map[string]interface{} `json:"tags"`
	Timestamp      time.Time              `json:"timestamp"`
}

// NewAnalyticsWorker creates a new analytics worker
func NewAnalyticsWorker(
	config *config.Config,
	logger *logrus.Logger,
	repository *observability.AnalyticsRepository,
) *AnalyticsWorker {
	return &AnalyticsWorker{
		config:     config,
		logger:     logger,
		repository: repository,
		queue:      make(chan AnalyticsJob, 1000), // Buffer for 1000 jobs
		quit:       make(chan bool),
	}
}

// Start starts the analytics worker
func (w *AnalyticsWorker) Start() {
	w.logger.Info("Starting analytics worker")

	// Start multiple worker goroutines for parallel processing
	numWorkers := w.config.Workers.AnalyticsWorkers
	if numWorkers == 0 {
		numWorkers = 3 // Default
	}

	for i := 0; i < numWorkers; i++ {
		go w.worker(i)
	}

	// Start batch processor for efficient DB writes
	go w.batchProcessor()
}

// Stop stops the analytics worker
func (w *AnalyticsWorker) Stop() {
	w.logger.Info("Stopping analytics worker")
	close(w.quit)
}

// QueueJob queues an analytics job for processing
func (w *AnalyticsWorker) QueueJob(jobType string, data interface{}) {
	job := AnalyticsJob{
		Type:      jobType,
		Data:      data,
		Timestamp: time.Now(),
		Retry:     0,
	}

	select {
	case w.queue <- job:
		w.logger.WithField("type", jobType).Debug("Analytics job queued")
	default:
		w.logger.WithField("type", jobType).Warn("Analytics queue full, dropping job")
	}
}

// QueueRequestLog queues a request log for processing
func (w *AnalyticsWorker) QueueRequestLog(log RequestLogJob) {
	w.QueueJob("request_log", log)
}

// QueueMetric queues a metric for processing
func (w *AnalyticsWorker) QueueMetric(metric MetricJob) {
	w.QueueJob("metric", metric)
}

// worker processes jobs from the queue
func (w *AnalyticsWorker) worker(id int) {
	w.logger.WithField("worker_id", id).Info("Analytics worker started")

	for {
		select {
		case job := <-w.queue:
			w.processJob(id, job)

		case <-w.quit:
			w.logger.WithField("worker_id", id).Info("Analytics worker stopping")
			return
		}
	}
}

// processJob processes a single analytics job
func (w *AnalyticsWorker) processJob(workerID int, job AnalyticsJob) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	logger := w.logger.WithFields(logrus.Fields{
		"worker_id": workerID,
		"job_type":  job.Type,
		"retry":     job.Retry,
	})

	logger.Debug("Processing analytics job")

	var err error
	switch job.Type {
	case "request_log":
		err = w.processRequestLog(ctx, job.Data)
	case "metric":
		err = w.processMetric(ctx, job.Data)
	default:
		logger.Warn("Unknown job type")
		return
	}

	if err != nil {
		logger.WithError(err).Error("Failed to process analytics job")

		// Retry logic
		if job.Retry < 3 {
			job.Retry++
			// Exponential backoff
			time.Sleep(time.Duration(job.Retry*job.Retry) * time.Second)
			w.queue <- job
			logger.WithField("retry", job.Retry).Info("Retrying analytics job")
		} else {
			logger.Error("Max retries exceeded, dropping analytics job")
		}
	} else {
		logger.Debug("Analytics job processed successfully")
	}
}

// processRequestLog processes a request log job
func (w *AnalyticsWorker) processRequestLog(ctx context.Context, data interface{}) error {
	jobData, ok := data.(RequestLogJob)
	if !ok {
		// Try to unmarshal if it's a map
		if mapData, ok := data.(map[string]interface{}); ok {
			jsonData, err := json.Marshal(mapData)
			if err != nil {
				return fmt.Errorf("failed to marshal request log data: %w", err)
			}
			if err := json.Unmarshal(jsonData, &jobData); err != nil {
				return fmt.Errorf("failed to unmarshal request log data: %w", err)
			}
		} else {
			return fmt.Errorf("invalid request log data type")
		}
	}

	// Convert to ClickHouse format
	log := &observability.RequestLog{
		ID:             jobData.RequestID,
		UserID:         jobData.UserID,
		OrganizationID: jobData.OrganizationID,
		ProjectID:      jobData.ProjectID,
		Environment:    jobData.Environment,
		Provider:       jobData.Provider,
		Model:          jobData.Model,
		Method:         jobData.Method,
		Endpoint:       jobData.Endpoint,
		StatusCode:     jobData.StatusCode,
		InputTokens:    jobData.InputTokens,
		OutputTokens:   jobData.OutputTokens,
		TotalTokens:    jobData.TotalTokens,
		Cost:           jobData.Cost,
		Latency:        jobData.Latency,
		QualityScore:   jobData.QualityScore,
		Cached:         jobData.Cached,
		Error:          jobData.Error,
		Timestamp:      jobData.Timestamp,
	}

	return w.repository.InsertRequestLog(ctx, log)
}

// processMetric processes a metric job
func (w *AnalyticsWorker) processMetric(ctx context.Context, data interface{}) error {
	jobData, ok := data.(MetricJob)
	if !ok {
		// Try to unmarshal if it's a map
		if mapData, ok := data.(map[string]interface{}); ok {
			jsonData, err := json.Marshal(mapData)
			if err != nil {
				return fmt.Errorf("failed to marshal metric data: %w", err)
			}
			if err := json.Unmarshal(jsonData, &jobData); err != nil {
				return fmt.Errorf("failed to unmarshal metric data: %w", err)
			}
		} else {
			return fmt.Errorf("invalid metric data type")
		}
	}

	// Convert tags to JSON string
	tagsJSON := ""
	if jobData.Tags != nil {
		tagsBytes, err := json.Marshal(jobData.Tags)
		if err != nil {
			w.logger.WithError(err).Warn("Failed to marshal metric tags")
		} else {
			tagsJSON = string(tagsBytes)
		}
	}

	// Convert to ClickHouse format
	metric := &observability.MetricPoint{
		Timestamp:      jobData.Timestamp,
		OrganizationID: jobData.OrganizationID,
		ProjectID:      jobData.ProjectID,
		Environment:    jobData.Environment,
		MetricName:     jobData.MetricName,
		MetricValue:    jobData.MetricValue,
		Tags:           tagsJSON,
	}

	return w.repository.InsertMetric(ctx, metric)
}

// batchProcessor processes jobs in batches for better performance
func (w *AnalyticsWorker) batchProcessor() {
	ticker := time.NewTicker(10 * time.Second) // Batch every 10 seconds
	defer ticker.Stop()

	var requestLogs []*observability.RequestLog
	var metrics []*observability.MetricPoint

	for {
		select {
		case <-ticker.C:
			if len(requestLogs) > 0 || len(metrics) > 0 {
				w.processBatches(requestLogs, metrics)
				requestLogs = nil
				metrics = nil
			}

		case <-w.quit:
			// Process remaining batches before stopping
			if len(requestLogs) > 0 || len(metrics) > 0 {
				w.processBatches(requestLogs, metrics)
			}
			return
		}
	}
}

// processBatches processes batched data
func (w *AnalyticsWorker) processBatches(requestLogs []*observability.RequestLog, metrics []*observability.MetricPoint) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Process request logs batch
	if len(requestLogs) > 0 {
		if err := w.repository.InsertBatchRequestLogs(ctx, requestLogs); err != nil {
			w.logger.WithError(err).WithField("count", len(requestLogs)).Error("Failed to insert request logs batch")
		} else {
			w.logger.WithField("count", len(requestLogs)).Debug("Request logs batch processed")
		}
	}

	// Process metrics batch
	if len(metrics) > 0 {
		for _, metric := range metrics {
			if err := w.repository.InsertMetric(ctx, metric); err != nil {
				w.logger.WithError(err).Error("Failed to insert metric")
			}
		}
		w.logger.WithField("count", len(metrics)).Debug("Metrics batch processed")
	}
}

// GetQueueLength returns the current queue length
func (w *AnalyticsWorker) GetQueueLength() int {
	return len(w.queue)
}

// GetStats returns worker statistics
func (w *AnalyticsWorker) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"queue_length": w.GetQueueLength(),
		"queue_capacity": cap(w.queue),
		"workers": w.config.Workers.AnalyticsWorkers,
	}
}