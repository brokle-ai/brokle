package observability

import (
	"context"
	"errors"
	"time"

	"brokle/pkg/ulid"
)

// NOTE: Service interfaces are defined here to avoid circular imports between
// internal/workers and internal/services/observability packages.
// Concrete implementations are in internal/services/observability/.

// TraceService defines the interface for trace and span operations.
// Traces are virtual (derived from root spans where parent_span_id IS NULL).
type TraceService interface {
	// Span Ingestion
	IngestSpan(ctx context.Context, span *Span) error
	IngestSpanBatch(ctx context.Context, spans []*Span) error

	// Span Queries
	GetSpan(ctx context.Context, spanID string) (*Span, error)
	GetSpansByFilter(ctx context.Context, filter *SpanFilter) ([]*Span, error)
	CountSpans(ctx context.Context, filter *SpanFilter) (int64, error)

	// Trace Queries
	GetTrace(ctx context.Context, traceID string) (*TraceSummary, error)
	GetTraceSpans(ctx context.Context, traceID string) ([]*Span, error)
	GetTraceTree(ctx context.Context, traceID string) ([]*Span, error)
	GetRootSpan(ctx context.Context, traceID string) (*Span, error)
	ListTraces(ctx context.Context, filter *TraceFilter) ([]*TraceSummary, error)
	CountTraces(ctx context.Context, filter *TraceFilter) (int64, error)

	// Analytics
	GetTracesBySession(ctx context.Context, sessionID string) ([]*TraceSummary, error)
	GetTracesByUser(ctx context.Context, userID string, filter *TraceFilter) ([]*TraceSummary, error)
	CalculateTraceCost(ctx context.Context, traceID string) (float64, error)
	CalculateTraceTokens(ctx context.Context, traceID string) (uint64, error)

	// Deletion
	DeleteSpan(ctx context.Context, spanID string) error
	DeleteTrace(ctx context.Context, traceID string) error
}

// ScoreService defines the comprehensive interface for quality score operations
// Used by both workers (CreateScore) and handlers (GetScoresByTraceID, etc.)
type ScoreService interface {
	// Create operations
	CreateScore(ctx context.Context, score *Score) error
	CreateScoreBatch(ctx context.Context, scores []*Score) error

	// Read operations
	GetScoreByID(ctx context.Context, id string) (*Score, error)
	GetScoresByTraceID(ctx context.Context, traceID string) ([]*Score, error)
	GetScoresBySpanID(ctx context.Context, spanID string) ([]*Score, error)
	GetScoresByFilter(ctx context.Context, filter *ScoreFilter) ([]*Score, error)

	// Update operations
	UpdateScore(ctx context.Context, score *Score) error

	// Delete operations
	DeleteScore(ctx context.Context, id string) error

	// Analytics operations
	CountScores(ctx context.Context, filter *ScoreFilter) (int64, error)
}

// BlobStorageService defines the comprehensive interface for blob storage operations
type BlobStorageService interface {
	// Create operations
	CreateBlobReference(ctx context.Context, blob *BlobStorageFileLog) error

	// Read operations
	GetBlobByID(ctx context.Context, id string) (*BlobStorageFileLog, error)
	GetBlobsByEntityID(ctx context.Context, entityType, entityID string) ([]*BlobStorageFileLog, error)
	GetBlobsByProjectID(ctx context.Context, projectID string, filter *BlobStorageFilter) ([]*BlobStorageFileLog, error)

	// Update operations
	UpdateBlobReference(ctx context.Context, blob *BlobStorageFileLog) error

	// Delete operations
	DeleteBlobReference(ctx context.Context, id string) error

	// Storage operations
	ShouldOffload(content string) bool
	UploadToS3(ctx context.Context, content string, entityType, entityID, eventID string) (*BlobStorageFileLog, error)
	DownloadFromS3(ctx context.Context, blobID string) (string, error)
}

// MetricsService defines the interface for OTLP metrics operations
type MetricsService interface {
	CreateMetricSumBatch(ctx context.Context, metricsSums []*MetricSum) error
	CreateMetricGaugeBatch(ctx context.Context, metricsGauges []*MetricGauge) error
	CreateMetricHistogramBatch(ctx context.Context, metricsHistograms []*MetricHistogram) error
	CreateMetricExponentialHistogramBatch(ctx context.Context, metricsExpHistograms []*MetricExponentialHistogram) error
}

// LogsService defines the interface for OTLP logs operations
type LogsService interface {
	CreateLogBatch(ctx context.Context, logs []*Log) error
}

// GenAIEventsService defines the interface for OTLP GenAI events operations
type GenAIEventsService interface {
	CreateGenAIEventBatch(ctx context.Context, events []*GenAIEvent) error
}

// TelemetryDeduplicationService defines the interface for ULID-based deduplication
type TelemetryDeduplicationService interface {
	// Atomic deduplication operations (uses composite OTLP IDs: trace_id:span_id)
	ClaimEvents(ctx context.Context, projectID ulid.ULID, batchID ulid.ULID, dedupIDs []string, ttl time.Duration) (claimedIDs, duplicateIDs []string, err error)
	ReleaseEvents(ctx context.Context, dedupIDs []string) error

	// Legacy deduplication operations (deprecated - use ClaimEvents instead)
	CheckDuplicate(ctx context.Context, dedupID string) (bool, error)
	CheckBatchDuplicates(ctx context.Context, dedupIDs []string) ([]string, error)
	RegisterEvent(ctx context.Context, dedupID string, batchID ulid.ULID, projectID ulid.ULID, ttl time.Duration) error
	RegisterProcessedEventsBatch(ctx context.Context, projectID ulid.ULID, batchID ulid.ULID, dedupIDs []string) error

	// TTL management (string-based for composite IDs)
	CalculateOptimalTTL(ctx context.Context, dedupID string, defaultTTL time.Duration) (time.Duration, error)
	GetExpirationTime(dedupID string, baseTTL time.Duration) time.Time

	// Cleanup operations
	CleanupExpired(ctx context.Context) (int64, error)
	CleanupByProject(ctx context.Context, projectID ulid.ULID, olderThan time.Time) (int64, error)
	BatchCleanup(ctx context.Context, olderThan time.Time, batchSize int) (int64, error)

	// Redis fallback management
	SyncToRedis(ctx context.Context, entries []*TelemetryEventDeduplication) error
	ValidateRedisHealth(ctx context.Context) (*RedisHealthStatus, error)
	GetDeduplicationStats(ctx context.Context, projectID ulid.ULID) (*DeduplicationStats, error)

	// Performance monitoring
	GetCacheHitRate(ctx context.Context, timeWindow time.Duration) (float64, error)
	GetFallbackRate(ctx context.Context, timeWindow time.Duration) (float64, error)
}

// TelemetryService aggregates all telemetry-related services for health monitoring
type TelemetryService interface {
	// Service access
	Deduplication() TelemetryDeduplicationService

	// Health and monitoring
	GetHealth(ctx context.Context) (*TelemetryHealthStatus, error)
	GetMetrics(ctx context.Context) (*TelemetryMetrics, error)
	GetPerformanceStats(ctx context.Context, timeWindow time.Duration) (*TelemetryPerformanceStats, error)
}

// Quality evaluator interface
type QualityEvaluator interface {
	Name() string
	Version() string
	Description() string
	SupportedTypes() []string // Span types: span, generation, event, tool, etc.
	Evaluate(ctx context.Context, input *EvaluationInput) (*Score, error)
	ValidateInput(input *EvaluationInput) error
}

// Service request and response types

// SpanCompletion represents data for completing a span
type SpanCompletion struct {
	EndTime        time.Time        `json:"end_time"`
	Output         map[string]any   `json:"output,omitempty"`
	Usage          *TokenUsage      `json:"usage,omitempty"`
	Cost           *CostCalculation `json:"cost,omitempty"`
	QualityScore   *float64         `json:"quality_score,omitempty"`
	StatusMessage  *string          `json:"status_message,omitempty"`
	AdditionalData map[string]any   `json:"additional_data,omitempty"`
}

// TokenUsage represents token usage information
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// CostCalculation represents cost calculation details
type CostCalculation struct {
	Currency   string  `json:"currency"`
	Provider   string  `json:"provider"`
	Model      string  `json:"model"`
	InputCost  float64 `json:"input_cost"`
	OutputCost float64 `json:"output_cost"`
	TotalCost  float64 `json:"total_cost"`
}

// BatchIngestRequest represents a batch ingestion request
// Note: Uses Span as traces are virtual in OTLP (root spans = traces)
type BatchIngestRequest struct {
	Spans     []*Span   `json:"spans"`
	ProjectID ulid.ULID `json:"project_id"`
	Async     bool      `json:"async"`
}

// BatchIngestResult represents the result of a batch ingestion
type BatchIngestResult struct {
	JobID          *string               `json:"job_id,omitempty"`
	Errors         []BatchIngestionError `json:"errors,omitempty"`
	ProcessedCount int                   `json:"processed_count"`
	FailedCount    int                   `json:"failed_count"`
	Duration       time.Duration         `json:"duration"`
}

// Service-specific request and response types for telemetry processing

// TelemetryBatchRequest represents a high-throughput telemetry batch request
type TelemetryBatchRequest struct {
	Metadata  map[string]any           `json:"metadata"`
	Events    []*TelemetryEventRequest `json:"events"`
	ProjectID ulid.ULID                `json:"project_id"`
	Async     bool                     `json:"async"`
}

// TelemetryEventRequest represents an individual telemetry event in a batch
type TelemetryEventRequest struct {
	Payload   map[string]any     `json:"payload"`
	Timestamp *time.Time         `json:"timestamp,omitempty"`
	SpanID    string             `json:"span_id"`
	TraceID   string             `json:"trace_id"`
	EventType TelemetryEventType `json:"event_type"`
	EventID   ulid.ULID          `json:"event_id"`
}

// Validate validates the event request
func (e *TelemetryEventRequest) Validate() error {
	if e.EventType == TelemetryEventTypeSpan && e.SpanID == "" {
		return errors.New("span events must have non-empty span_id")
	}
	if e.TraceID == "" {
		return errors.New("trace_id is required for all events")
	}
	return nil
}

// TelemetryBatchResponse represents the response for telemetry batch processing
type TelemetryBatchResponse struct {
	JobID             *string               `json:"job_id,omitempty"`
	Errors            []TelemetryEventError `json:"errors,omitempty"`
	DuplicateEventIDs []ulid.ULID           `json:"duplicate_event_ids,omitempty"`
	ProcessedEvents   int                   `json:"processed_events"`
	DuplicateEvents   int                   `json:"duplicate_events"`
	FailedEvents      int                   `json:"failed_events"`
	ProcessingTimeMs  int                   `json:"processing_time_ms"`
	BatchID           ulid.ULID             `json:"batch_id"`
}

// TelemetryEventError represents an error processing a telemetry event
type TelemetryEventError struct {
	EventType    TelemetryEventType `json:"event_type"`
	ErrorCode    string             `json:"error_code"`
	ErrorMessage string             `json:"error_message"`
	EventID      ulid.ULID          `json:"event_id"`
	Retryable    bool               `json:"retryable"`
}

// BatchProcessingResult represents the result of batch processing operations
type BatchProcessingResult struct {
	Errors           []TelemetryEventError `json:"errors,omitempty"`
	TotalEvents      int                   `json:"total_events"`
	ProcessedEvents  int                   `json:"processed_events"`
	FailedEvents     int                   `json:"failed_events"`
	SkippedEvents    int                   `json:"skipped_events"`
	ProcessingTimeMs int                   `json:"processing_time_ms"`
	ThroughputPerSec float64               `json:"throughput_per_sec"`
	SuccessRate      float64               `json:"success_rate"`
	BatchID          ulid.ULID             `json:"batch_id"`
}

// EventProcessingResult represents the result of event processing operations
type EventProcessingResult struct {
	ProcessedEventIDs []ulid.ULID           `json:"processed_event_ids"`
	NotProcessedIDs   []ulid.ULID           `json:"not_processed_ids"`
	Errors            []TelemetryEventError `json:"errors,omitempty"`
	ProcessedCount    int                   `json:"processed_count"`
	FailedCount       int                   `json:"failed_count"`
	NotProcessedCount int                   `json:"not_processed_count"`
	RetryCount        int                   `json:"retry_count"`
	ProcessingTimeMs  int                   `json:"processing_time_ms"`
	SuccessRate       float64               `json:"success_rate"`
}

// RedisHealthStatus represents Redis health status for deduplication
type RedisHealthStatus struct {
	LastError   *string       `json:"last_error,omitempty"`
	LatencyMs   float64       `json:"latency_ms"`
	MemoryUsage int64         `json:"memory_usage_bytes"`
	Connections int           `json:"connections"`
	Uptime      time.Duration `json:"uptime"`
	Available   bool          `json:"available"`
}

// DeduplicationStats represents deduplication performance statistics
type DeduplicationStats struct {
	ProjectID         ulid.ULID `json:"project_id"`
	TotalChecks       int64     `json:"total_checks"`
	CacheHits         int64     `json:"cache_hits"`
	CacheMisses       int64     `json:"cache_misses"`
	DatabaseFallbacks int64     `json:"database_fallbacks"`
	DuplicatesFound   int64     `json:"duplicates_found"`
	CacheHitRate      float64   `json:"cache_hit_rate"`
	FallbackRate      float64   `json:"fallback_rate"`
	AverageLatencyMs  float64   `json:"average_latency_ms"`
}

// TelemetryHealthStatus represents overall telemetry service health
type TelemetryHealthStatus struct {
	Database              *DatabaseHealth    `json:"database"`
	Redis                 *RedisHealthStatus `json:"redis"`
	ProcessingQueue       *QueueHealth       `json:"processing_queue"`
	ActiveWorkers         int                `json:"active_workers"`
	AverageProcessingTime float64            `json:"average_processing_time_ms"`
	ThroughputPerMinute   float64            `json:"throughput_per_minute"`
	ErrorRate             float64            `json:"error_rate"`
	Healthy               bool               `json:"healthy"`
}

// DatabaseHealth represents database health status
type DatabaseHealth struct {
	Connected         bool    `json:"connected"`
	LatencyMs         float64 `json:"latency_ms"`
	ActiveConnections int     `json:"active_connections"`
	MaxConnections    int     `json:"max_connections"`
}

// QueueHealth represents processing queue health
type QueueHealth struct {
	Size             int64   `json:"size"`
	ProcessingRate   float64 `json:"processing_rate"`
	AverageWaitTime  float64 `json:"average_wait_time_ms"`
	OldestMessageAge float64 `json:"oldest_message_age_ms"`
}

// TelemetryMetrics represents comprehensive telemetry service metrics
type TelemetryMetrics struct {
	TotalBatches          int64   `json:"total_batches"`
	CompletedBatches      int64   `json:"completed_batches"`
	FailedBatches         int64   `json:"failed_batches"`
	ProcessingBatches     int64   `json:"processing_batches"`
	TotalEvents           int64   `json:"total_events"`
	ProcessedEvents       int64   `json:"processed_events"`
	FailedEvents          int64   `json:"failed_events"`
	DuplicateEvents       int64   `json:"duplicate_events"`
	AverageEventsPerBatch float64 `json:"average_events_per_batch"`
	ThroughputPerSecond   float64 `json:"throughput_per_second"`
	SuccessRate           float64 `json:"success_rate"`
	DeduplicationRate     float64 `json:"deduplication_rate"`
}

// TelemetryPerformanceStats represents performance statistics over a time window
type TelemetryPerformanceStats struct {
	TimeWindow           time.Duration `json:"time_window"`
	TotalRequests        int64         `json:"total_requests"`
	SuccessfulRequests   int64         `json:"successful_requests"`
	AverageLatencyMs     float64       `json:"average_latency_ms"`
	P95LatencyMs         float64       `json:"p95_latency_ms"`
	P99LatencyMs         float64       `json:"p99_latency_ms"`
	ThroughputPerSecond  float64       `json:"throughput_per_second"`
	PeakThroughput       float64       `json:"peak_throughput"`
	CacheHitRate         float64       `json:"cache_hit_rate"`
	DatabaseFallbackRate float64       `json:"database_fallback_rate"`
	ErrorRate            float64       `json:"error_rate"`
	RetryRate            float64       `json:"retry_rate"`
}

// SpanBatchRequest represents a batch span ingestion request
type SpanBatchRequest struct {
	Spans     []*Span   `json:"spans"`
	ProjectID ulid.ULID `json:"project_id"`
	Async     bool      `json:"async"`
}

// QualityScoreBatchRequest represents a batch quality score ingestion request
type QualityScoreBatchRequest struct {
	QualityScores []*Score  `json:"quality_scores"`
	ProjectID     ulid.ULID `json:"project_id"`
	Async         bool      `json:"async"`
}

// BatchIngestionError represents an error during batch ingestion
type BatchIngestionError struct {
	Details any    `json:"details,omitempty"`
	Error   string `json:"error"`
	Index   int    `json:"index"`
}

// BulkEvaluationRequest represents a bulk evaluation request
type BulkEvaluationRequest struct {
	Filter         *AnalyticsFilter `json:"filter,omitempty"`
	TraceIDs       []ulid.ULID      `json:"trace_ids,omitempty"`
	SpanIDs        []ulid.ULID      `json:"span_ids,omitempty"`
	EvaluatorNames []string         `json:"evaluator_names"`
	Async          bool             `json:"async"`
}

// BulkEvaluationResult represents the result of bulk evaluation
type BulkEvaluationResult struct {
	JobID          *string               `json:"job_id,omitempty"`
	Scores         []*Score              `json:"scores,omitempty"`
	Errors         []BulkEvaluationError `json:"errors,omitempty"`
	ProcessedCount int                   `json:"processed_count"`
	FailedCount    int                   `json:"failed_count"`
}

// BulkEvaluationError represents an error during bulk evaluation
type BulkEvaluationError struct {
	Details any       `json:"details,omitempty"`
	Error   string    `json:"error"`
	ItemID  ulid.ULID `json:"item_id"`
}

// EvaluationInput represents input for quality evaluation
// Note: Uses TraceSummary as traces are virtual in OTLP
type EvaluationInput struct {
	TraceID      *ulid.ULID    `json:"trace_id,omitempty"`
	SpanID       *ulid.ULID    `json:"span_id,omitempty"`
	TraceSummary *TraceSummary `json:"trace_summary,omitempty"`
	Span         *Span         `json:"span,omitempty"`
	Context      map[string]any `json:"context,omitempty"`
}

// QualityEvaluatorInfo represents information about a quality evaluator
type QualityEvaluatorInfo struct {
	Configuration  map[string]any `json:"configuration,omitempty"`
	Name           string         `json:"name"`
	Version        string         `json:"version"`
	Description    string         `json:"description"`
	SupportedTypes []string       `json:"supported_types"`
	IsBuiltIn      bool           `json:"is_built_in"`
}

// Dashboard and reporting types

// DashboardOverview represents overview metrics for dashboard
type DashboardOverview struct {
	TopProviders   []*ProviderSummary `json:"top_providers"`
	RecentActivity []*ActivityItem    `json:"recent_activity"`
	CostTrend      []*TimeSeriesPoint `json:"cost_trend"`
	LatencyTrend   []*TimeSeriesPoint `json:"latency_trend"`
	QualityTrend   []*TimeSeriesPoint `json:"quality_trend"`
	TotalTraces    int64              `json:"total_traces"`
	TotalCost      float64            `json:"total_cost"`
	AverageLatency float64            `json:"average_latency"`
	ErrorRate      float64            `json:"error_rate"`
}

// ProviderSummary represents summary information for a provider
type ProviderSummary struct {
	Provider       string  `json:"provider"`
	RequestCount   int64   `json:"request_count"`
	TotalCost      float64 `json:"total_cost"`
	AverageLatency float64 `json:"average_latency"`
	ErrorRate      float64 `json:"error_rate"`
}

// ActivityItem represents a recent activity item
type ActivityItem struct {
	Timestamp   time.Time      `json:"timestamp"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	Type        string         `json:"type"`
	Description string         `json:"description"`
}

// TimeRange represents a time range for analytics
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// OptimizationSuggestion represents a cost optimization suggestion
type OptimizationSuggestion struct {
	Metadata         map[string]any `json:"metadata,omitempty"`
	Type             string         `json:"type"`
	Title            string         `json:"title"`
	Description      string         `json:"description"`
	ActionItems      []string       `json:"action_items"`
	PotentialSavings float64        `json:"potential_savings"`
	Confidence       float64        `json:"confidence"`
}

// LatencyHeatmap represents latency data in heatmap format
type LatencyHeatmap struct {
	Data     [][]float64 `json:"data"`
	XLabels  []string    `json:"x_labels"`
	YLabels  []string    `json:"y_labels"`
	MinValue float64     `json:"min_value"`
	MaxValue float64     `json:"max_value"`
}

// ThroughputMetrics represents throughput metrics
type ThroughputMetrics struct {
	TimeSeries        []*TimeSeriesPoint `json:"time_series"`
	RequestsPerSecond float64            `json:"requests_per_second"`
	RequestsPerMinute float64            `json:"requests_per_minute"`
	RequestsPerHour   float64            `json:"requests_per_hour"`
	PeakThroughput    float64            `json:"peak_throughput"`
}

// Queue and ingestion monitoring types

// QueueStatus represents the status of ingestion queues
type QueueStatus struct {
	TraceQueue   *QueueInfo `json:"trace_queue"`
	SpanQueue    *QueueInfo `json:"span_queue"`
	QualityQueue *QueueInfo `json:"quality_queue"`
	TotalPending int64      `json:"total_pending"`
	IsHealthy    bool       `json:"is_healthy"`
}

// QueueInfo represents information about a specific queue
type QueueInfo struct {
	LastUpdated time.Time `json:"last_updated"`
	Name        string    `json:"name"`
	Size        int64     `json:"size"`
	Processing  int64     `json:"processing"`
	Failed      int64     `json:"failed"`
}

// IngestionHealth represents the health status of ingestion
type IngestionHealth struct {
	LastCheck      time.Time      `json:"last_check"`
	Details        map[string]any `json:"details,omitempty"`
	Status         string         `json:"status"`
	Bottlenecks    []string       `json:"bottlenecks,omitempty"`
	IngestionRate  float64        `json:"ingestion_rate"`
	ProcessingRate float64        `json:"processing_rate"`
	ErrorRate      float64        `json:"error_rate"`
}

// IngestionMetrics represents metrics for the ingestion system
type IngestionMetrics struct {
	Errors         []*IngestionError `json:"recent_errors,omitempty"`
	TotalIngested  int64             `json:"total_ingested"`
	IngestedToday  int64             `json:"ingested_today"`
	ProcessingRate float64           `json:"processing_rate"`
	AverageLatency float64           `json:"average_latency"`
	QueueBacklog   int64             `json:"queue_backlog"`
	WorkerCount    int               `json:"worker_count"`
}

// IngestionError represents an ingestion error
type IngestionError struct {
	Timestamp time.Time      `json:"timestamp"`
	Details   map[string]any `json:"details,omitempty"`
	Type      string         `json:"type"`
	Message   string         `json:"message"`
	Count     int64          `json:"count"`
}

// Export and reporting types

// ExportFormat represents supported export formats
type ExportFormat string

const (
	ExportFormatJSON    ExportFormat = "json"
	ExportFormatCSV     ExportFormat = "csv"
	ExportFormatParquet ExportFormat = "parquet"
)

// ExportResult represents the result of an export operation
type ExportResult struct {
	ExpiresAt   time.Time    `json:"expires_at"`
	DownloadURL string       `json:"download_url"`
	Format      ExportFormat `json:"format"`
	Status      string       `json:"status"`
	RecordCount int64        `json:"record_count"`
	FileSize    int64        `json:"file_size"`
}

// ReportType represents different types of reports
type ReportType string

const (
	ReportTypeCost        ReportType = "cost"
	ReportTypePerformance ReportType = "performance"
	ReportTypeQuality     ReportType = "quality"
	ReportTypeUsage       ReportType = "usage"
)

// Report represents a generated report
type Report struct {
	GeneratedAt time.Time      `json:"generated_at"`
	Data        map[string]any `json:"data"`
	Type        ReportType     `json:"type"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Format      ExportFormat   `json:"format"`
	DownloadURL string         `json:"download_url,omitempty"`
	ID          ulid.ULID      `json:"id"`
}

// Placeholder types (will be implemented with analytics features)

type AnalyticsFilter struct{}

type TimeSeriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}
