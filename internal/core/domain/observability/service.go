package observability

import (
	"context"
	"fmt"
	"time"

	"brokle/pkg/ulid"
)

// NOTE: Service interfaces are defined here to avoid circular imports between
// internal/workers and internal/services/observability packages.
// Concrete implementations are in internal/services/observability/.

// TraceService defines the comprehensive interface for trace operations
// Used by both workers (CreateTrace) and handlers (GetTraceByID, GetTraceWithSpans, etc.)
type TraceService interface {
	// Create operations
	CreateTrace(ctx context.Context, trace *Trace) error
	CreateTraceBatch(ctx context.Context, traces []*Trace) error

	// Read operations
	GetTraceByID(ctx context.Context, id string) (*Trace, error)
	GetTraceWithSpans(ctx context.Context, id string) (*Trace, error)
	GetTraceWithScores(ctx context.Context, id string) (*Trace, error)
	GetTracesByProjectID(ctx context.Context, projectID string, filter *TraceFilter) ([]*Trace, error)
	GetTracesBySessionID(ctx context.Context, sessionID string) ([]*Trace, error) // Virtual session analytics
	GetTracesByUserID(ctx context.Context, userID string, filter *TraceFilter) ([]*Trace, error)

	// Update operations
	UpdateTrace(ctx context.Context, trace *Trace) error
	UpdateTraceMetrics(ctx context.Context, traceID string, totalCost float64, totalTokens, spanCount uint32) error

	// Delete operations
	DeleteTrace(ctx context.Context, id string) error

	// Analytics operations
	CountTraces(ctx context.Context, filter *TraceFilter) (int64, error)
}

// SpanService defines the comprehensive interface for span operations
// Used by both workers (CreateSpan) and handlers (GetSpansByFilter, etc.)
type SpanService interface {
	// Create operations
	CreateSpan(ctx context.Context, span *Span) error
	CreateSpanBatch(ctx context.Context, spans []*Span) error

	// Read operations
	GetSpanByID(ctx context.Context, id string) (*Span, error)
	GetSpansByTraceID(ctx context.Context, traceID string) ([]*Span, error)
	GetRootSpan(ctx context.Context, traceID string) (*Span, error)
	GetSpanTreeByTraceID(ctx context.Context, traceID string) ([]*Span, error)
	GetChildSpans(ctx context.Context, parentSpanID string) ([]*Span, error)
	GetSpansByFilter(ctx context.Context, filter *SpanFilter) ([]*Span, error)

	// Update operations
	UpdateSpan(ctx context.Context, span *Span) error
	SetSpanCost(ctx context.Context, spanID string, inputCost, outputCost float64) error
	SetSpanUsage(ctx context.Context, spanID string, promptTokens, completionTokens uint32) error

	// Delete operations
	DeleteSpan(ctx context.Context, id string) error

	// Analytics operations
	CountSpans(ctx context.Context, filter *SpanFilter) (int64, error)
	CalculateTraceCost(ctx context.Context, traceID string) (float64, error)
	CalculateTraceTokens(ctx context.Context, traceID string) (uint32, error)
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

	// Analytics operations
	CountBlobs(ctx context.Context, filter *BlobStorageFilter) (int64, error)
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
	InputCost  float64 `json:"input_cost"`
	OutputCost float64 `json:"output_cost"`
	TotalCost  float64 `json:"total_cost"`
	Currency   string  `json:"currency"`
	Provider   string  `json:"provider"`
	Model      string  `json:"model"`
}

// BatchIngestRequest represents a batch ingestion request
type BatchIngestRequest struct {
	ProjectID ulid.ULID `json:"project_id"`
	Traces    []*Trace  `json:"traces"`
	Async     bool      `json:"async"`
}

// BatchIngestResult represents the result of a batch ingestion
type BatchIngestResult struct {
	ProcessedCount int                   `json:"processed_count"`
	FailedCount    int                   `json:"failed_count"`
	Errors         []BatchIngestionError `json:"errors,omitempty"`
	Duration       time.Duration         `json:"duration"`
	JobID          *string               `json:"job_id,omitempty"` // For async operations
}

// Service-specific request and response types for telemetry processing

// TelemetryBatchRequest represents a high-throughput telemetry batch request
type TelemetryBatchRequest struct {
	ProjectID ulid.ULID                `json:"project_id"`
	Metadata  map[string]any           `json:"metadata"`
	Events    []*TelemetryEventRequest `json:"events"`
	Async     bool                     `json:"async"`
}

// TelemetryEventRequest represents an individual telemetry event in a batch
type TelemetryEventRequest struct {
	// Internal tracking (Brokle-specific)
	EventID ulid.ULID `json:"event_id"`

	// OTLP identity (for deduplication)
	SpanID  string `json:"span_id"`  // OTLP span_id (16 hex) - empty for traces
	TraceID string `json:"trace_id"` // OTLP trace_id (32 hex) - required

	// Event data
	EventType TelemetryEventType `json:"event_type"`
	Payload   map[string]any     `json:"payload"`
	Timestamp *time.Time         `json:"timestamp,omitempty"`
}

// Validate validates the event request
func (e *TelemetryEventRequest) Validate() error {
	if e.EventType == TelemetryEventTypeSpan && e.SpanID == "" {
		return fmt.Errorf("span events must have non-empty span_id")
	}
	if e.TraceID == "" {
		return fmt.Errorf("trace_id is required for all events")
	}
	return nil
}

// TelemetryBatchResponse represents the response for telemetry batch processing
type TelemetryBatchResponse struct {
	BatchID           ulid.ULID             `json:"batch_id"`
	ProcessedEvents   int                   `json:"processed_events"`
	DuplicateEvents   int                   `json:"duplicate_events"`
	FailedEvents      int                   `json:"failed_events"`
	ProcessingTimeMs  int                   `json:"processing_time_ms"`
	Errors            []TelemetryEventError `json:"errors,omitempty"`
	DuplicateEventIDs []ulid.ULID           `json:"duplicate_event_ids,omitempty"`
	JobID             *string               `json:"job_id,omitempty"` // For async processing
}

// TelemetryEventError represents an error processing a telemetry event
type TelemetryEventError struct {
	EventID      ulid.ULID          `json:"event_id"`
	EventType    TelemetryEventType `json:"event_type"`
	ErrorCode    string             `json:"error_code"`
	ErrorMessage string             `json:"error_message"`
	Retryable    bool               `json:"retryable"`
}

// BatchProcessingResult represents the result of batch processing operations
type BatchProcessingResult struct {
	BatchID          ulid.ULID             `json:"batch_id"`
	TotalEvents      int                   `json:"total_events"`
	ProcessedEvents  int                   `json:"processed_events"`
	FailedEvents     int                   `json:"failed_events"`
	SkippedEvents    int                   `json:"skipped_events"`
	ProcessingTimeMs int                   `json:"processing_time_ms"`
	ThroughputPerSec float64               `json:"throughput_per_sec"`
	Errors           []TelemetryEventError `json:"errors,omitempty"`
	SuccessRate      float64               `json:"success_rate"`
}

// EventProcessingResult represents the result of event processing operations
type EventProcessingResult struct {
	ProcessedCount    int `json:"processed_count"`
	FailedCount       int `json:"failed_count"`
	NotProcessedCount int `json:"not_processed_count"` // NEW: Events never attempted
	RetryCount        int `json:"retry_count"`
	ProcessingTimeMs  int `json:"processing_time_ms"`

	// Explicit event ID lists for precise tracking
	ProcessedEventIDs []ulid.ULID           `json:"processed_event_ids"` // NEW: Successfully processed
	NotProcessedIDs   []ulid.ULID           `json:"not_processed_ids"`   // NEW: Never attempted
	Errors            []TelemetryEventError `json:"errors,omitempty"`    // Failed events with details
	SuccessRate       float64               `json:"success_rate"`
}

// RedisHealthStatus represents Redis health status for deduplication
type RedisHealthStatus struct {
	Available   bool          `json:"available"`
	LatencyMs   float64       `json:"latency_ms"`
	MemoryUsage int64         `json:"memory_usage_bytes"`
	Connections int           `json:"connections"`
	LastError   *string       `json:"last_error,omitempty"`
	Uptime      time.Duration `json:"uptime"`
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
	Healthy               bool               `json:"healthy"`
	Database              *DatabaseHealth    `json:"database"`
	Redis                 *RedisHealthStatus `json:"redis"`
	ProcessingQueue       *QueueHealth       `json:"processing_queue"`
	ActiveWorkers         int                `json:"active_workers"`
	AverageProcessingTime float64            `json:"average_processing_time_ms"`
	ThroughputPerMinute   float64            `json:"throughput_per_minute"`
	ErrorRate             float64            `json:"error_rate"`
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
	ProjectID ulid.ULID `json:"project_id"`
	Spans     []*Span   `json:"spans"`
	Async     bool      `json:"async"`
}

// QualityScoreBatchRequest represents a batch quality score ingestion request
type QualityScoreBatchRequest struct {
	ProjectID     ulid.ULID `json:"project_id"`
	QualityScores []*Score  `json:"quality_scores"`
	Async         bool      `json:"async"`
}

// BatchIngestionError represents an error during batch ingestion
type BatchIngestionError struct {
	Index   int    `json:"index"`
	Error   string `json:"error"`
	Details any    `json:"details,omitempty"`
}

// BulkEvaluationRequest represents a bulk evaluation request
type BulkEvaluationRequest struct {
	TraceIDs       []ulid.ULID      `json:"trace_ids,omitempty"`
	SpanIDs        []ulid.ULID      `json:"span_ids,omitempty"`
	EvaluatorNames []string         `json:"evaluator_names"`
	Filter         *AnalyticsFilter `json:"filter,omitempty"`
	Async          bool             `json:"async"`
}

// BulkEvaluationResult represents the result of bulk evaluation
type BulkEvaluationResult struct {
	ProcessedCount int                   `json:"processed_count"`
	FailedCount    int                   `json:"failed_count"`
	Scores         []*Score              `json:"scores,omitempty"`
	Errors         []BulkEvaluationError `json:"errors,omitempty"`
	JobID          *string               `json:"job_id,omitempty"`
}

// BulkEvaluationError represents an error during bulk evaluation
type BulkEvaluationError struct {
	ItemID  ulid.ULID `json:"item_id"`
	Error   string    `json:"error"`
	Details any       `json:"details,omitempty"`
}

// EvaluationInput represents input for quality evaluation
type EvaluationInput struct {
	TraceID *ulid.ULID     `json:"trace_id,omitempty"`
	SpanID  *ulid.ULID     `json:"span_id,omitempty"`
	Trace   *Trace         `json:"trace,omitempty"`
	Span    *Span          `json:"span,omitempty"`
	Context map[string]any `json:"context,omitempty"`
}

// QualityEvaluatorInfo represents information about a quality evaluator
type QualityEvaluatorInfo struct {
	Name           string         `json:"name"`
	Version        string         `json:"version"`
	Description    string         `json:"description"`
	SupportedTypes []string       `json:"supported_types"` // Span types
	IsBuiltIn      bool           `json:"is_built_in"`
	Configuration  map[string]any `json:"configuration,omitempty"`
}

// Dashboard and reporting types

// DashboardOverview represents overview metrics for dashboard
type DashboardOverview struct {
	TotalTraces    int64              `json:"total_traces"`
	TotalCost      float64            `json:"total_cost"`
	AverageLatency float64            `json:"average_latency"`
	ErrorRate      float64            `json:"error_rate"`
	TopProviders   []*ProviderSummary `json:"top_providers"`
	RecentActivity []*ActivityItem    `json:"recent_activity"`
	CostTrend      []*TimeSeriesPoint `json:"cost_trend"`
	LatencyTrend   []*TimeSeriesPoint `json:"latency_trend"`
	QualityTrend   []*TimeSeriesPoint `json:"quality_trend"`
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
	Type        string         `json:"type"`
	Description string         `json:"description"`
	Timestamp   time.Time      `json:"timestamp"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// TimeRange represents a time range for analytics
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// OptimizationSuggestion represents a cost optimization suggestion
type OptimizationSuggestion struct {
	Type             string         `json:"type"`
	Title            string         `json:"title"`
	Description      string         `json:"description"`
	PotentialSavings float64        `json:"potential_savings"`
	Confidence       float64        `json:"confidence"`
	ActionItems      []string       `json:"action_items"`
	Metadata         map[string]any `json:"metadata,omitempty"`
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
	RequestsPerSecond float64            `json:"requests_per_second"`
	RequestsPerMinute float64            `json:"requests_per_minute"`
	RequestsPerHour   float64            `json:"requests_per_hour"`
	PeakThroughput    float64            `json:"peak_throughput"`
	TimeSeries        []*TimeSeriesPoint `json:"time_series"`
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
	Name        string    `json:"name"`
	Size        int64     `json:"size"`
	Processing  int64     `json:"processing"`
	Failed      int64     `json:"failed"`
	LastUpdated time.Time `json:"last_updated"`
}

// IngestionHealth represents the health status of ingestion
type IngestionHealth struct {
	Status         string         `json:"status"` // healthy, degraded, unhealthy
	IngestionRate  float64        `json:"ingestion_rate"`
	ProcessingRate float64        `json:"processing_rate"`
	ErrorRate      float64        `json:"error_rate"`
	Bottlenecks    []string       `json:"bottlenecks,omitempty"`
	LastCheck      time.Time      `json:"last_check"`
	Details        map[string]any `json:"details,omitempty"`
}

// IngestionMetrics represents metrics for the ingestion system
type IngestionMetrics struct {
	TotalIngested  int64             `json:"total_ingested"`
	IngestedToday  int64             `json:"ingested_today"`
	ProcessingRate float64           `json:"processing_rate"`
	AverageLatency float64           `json:"average_latency"`
	QueueBacklog   int64             `json:"queue_backlog"`
	WorkerCount    int               `json:"worker_count"`
	Errors         []*IngestionError `json:"recent_errors,omitempty"`
}

// IngestionError represents an ingestion error
type IngestionError struct {
	Timestamp time.Time      `json:"timestamp"`
	Type      string         `json:"type"`
	Message   string         `json:"message"`
	Count     int64          `json:"count"`
	Details   map[string]any `json:"details,omitempty"`
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
	DownloadURL string       `json:"download_url"`
	Format      ExportFormat `json:"format"`
	RecordCount int64        `json:"record_count"`
	FileSize    int64        `json:"file_size"`
	ExpiresAt   time.Time    `json:"expires_at"`
	Status      string       `json:"status"`
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
	ID          ulid.ULID      `json:"id"`
	Type        ReportType     `json:"type"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Data        map[string]any `json:"data"`
	GeneratedAt time.Time      `json:"generated_at"`
	Format      ExportFormat   `json:"format"`
	DownloadURL string         `json:"download_url,omitempty"`
}

// ==================================
// Legacy/Placeholder Types (not actively used in clean implementation)
// ==================================

// AnalyticsFilter - placeholder for analytics filtering (not actively used)
type AnalyticsFilter struct {
	// Placeholder - will be implemented when analytics features are added
}

// TimeSeriesPoint - placeholder for time series data (not actively used)
type TimeSeriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}
