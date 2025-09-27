package observability

import (
	"context"
	"time"

	"brokle/pkg/ulid"
)

// TraceService defines the interface for trace business logic
type TraceService interface {
	// Core trace operations
	CreateTrace(ctx context.Context, trace *Trace) (*Trace, error)
	CreateTraceWithObservations(ctx context.Context, trace *Trace) (*Trace, error)
	GetTrace(ctx context.Context, id ulid.ULID) (*Trace, error)
	GetTraceByExternalID(ctx context.Context, externalTraceID string) (*Trace, error)
	UpdateTrace(ctx context.Context, trace *Trace) (*Trace, error)
	DeleteTrace(ctx context.Context, id ulid.ULID) error

	// Trace queries
	ListTraces(ctx context.Context, filter *TraceFilter) ([]*Trace, int, error)
	GetTraceWithObservations(ctx context.Context, id ulid.ULID) (*Trace, error)
	GetTraceStats(ctx context.Context, id ulid.ULID) (*TraceStats, error)
	GetRecentTraces(ctx context.Context, projectID ulid.ULID, limit int) ([]*Trace, error)

	// Batch operations
	CreateTracesBatch(ctx context.Context, traces []*Trace) ([]*Trace, error)
	IngestTraceBatch(ctx context.Context, request *BatchIngestRequest) (*BatchIngestResult, error)

	// Search and filtering
	SearchTraces(ctx context.Context, query string, filter *TraceFilter) ([]*Trace, int, error)
	GetTracesByTimeRange(ctx context.Context, projectID ulid.ULID, startTime, endTime time.Time, limit, offset int) ([]*Trace, error)

	// Analytics integration
	GetTraceAnalytics(ctx context.Context, filter *AnalyticsFilter) (*TraceAnalytics, error)
}

// ObservationService defines the interface for observation business logic
type ObservationService interface {
	// Core observation operations
	CreateObservation(ctx context.Context, observation *Observation) (*Observation, error)
	GetObservation(ctx context.Context, id ulid.ULID) (*Observation, error)
	GetObservationByExternalID(ctx context.Context, externalObservationID string) (*Observation, error)
	UpdateObservation(ctx context.Context, observation *Observation) (*Observation, error)
	CompleteObservation(ctx context.Context, id ulid.ULID, completionData *ObservationCompletion) (*Observation, error)
	DeleteObservation(ctx context.Context, id ulid.ULID) error

	// Observation queries
	ListObservations(ctx context.Context, filter *ObservationFilter) ([]*Observation, int, error)
	GetObservationsByTrace(ctx context.Context, traceID ulid.ULID) ([]*Observation, error)
	GetChildObservations(ctx context.Context, parentID ulid.ULID) ([]*Observation, error)

	// Batch operations
	CreateObservationsBatch(ctx context.Context, observations []*Observation) ([]*Observation, error)
	UpdateObservationsBatch(ctx context.Context, observations []*Observation) ([]*Observation, error)

	// Analytics and stats
	GetObservationStats(ctx context.Context, filter *ObservationFilter) (*ObservationStats, error)
	GetObservationAnalytics(ctx context.Context, filter *AnalyticsFilter) (*ObservationAnalytics, error)

	// Cost and token tracking
	CalculateCost(ctx context.Context, observation *Observation) (*CostCalculation, error)
	GetCostBreakdown(ctx context.Context, filter *AnalyticsFilter) ([]*CostBreakdown, error)

	// Performance monitoring
	GetLatencyPercentiles(ctx context.Context, filter *ObservationFilter) (*LatencyPercentiles, error)
	GetThroughputMetrics(ctx context.Context, filter *AnalyticsFilter) (*ThroughputMetrics, error)
}

// QualityService defines the interface for quality evaluation and scoring
type QualityService interface {
	// Quality scoring operations
	CreateQualityScore(ctx context.Context, score *QualityScore) (*QualityScore, error)
	GetQualityScore(ctx context.Context, id ulid.ULID) (*QualityScore, error)
	UpdateQualityScore(ctx context.Context, score *QualityScore) (*QualityScore, error)
	DeleteQualityScore(ctx context.Context, id ulid.ULID) error

	// Score queries
	GetQualityScoresByTrace(ctx context.Context, traceID ulid.ULID) ([]*QualityScore, error)
	GetQualityScoresByObservation(ctx context.Context, observationID ulid.ULID) ([]*QualityScore, error)
	ListQualityScores(ctx context.Context, filter *QualityScoreFilter) ([]*QualityScore, int, error)

	// Evaluation operations
	EvaluateTrace(ctx context.Context, traceID ulid.ULID, evaluatorName string) (*QualityScore, error)
	EvaluateObservation(ctx context.Context, observationID ulid.ULID, evaluatorName string) (*QualityScore, error)
	BulkEvaluate(ctx context.Context, request *BulkEvaluationRequest) (*BulkEvaluationResult, error)

	// Evaluator management
	RegisterEvaluator(ctx context.Context, evaluator QualityEvaluator) error
	GetEvaluator(ctx context.Context, name string) (QualityEvaluator, error)
	ListEvaluators(ctx context.Context) ([]QualityEvaluatorInfo, error)

	// Quality analytics
	GetQualityAnalytics(ctx context.Context, filter *AnalyticsFilter) (*QualityAnalytics, error)
	GetQualityTrends(ctx context.Context, filter *AnalyticsFilter, interval string) ([]*QualityTrendPoint, error)
	GetScoreDistribution(ctx context.Context, scoreName string, filter *QualityScoreFilter) (map[string]int, error)
}

// AnalyticsService defines the interface for analytics and reporting
type AnalyticsService interface {
	// Dashboard analytics
	GetDashboardOverview(ctx context.Context, projectID ulid.ULID, timeRange TimeRange) (*DashboardOverview, error)
	GetRealtimeMetrics(ctx context.Context, projectID ulid.ULID) (*RealtimeMetrics, error)

	// Cost analytics
	GetCostAnalytics(ctx context.Context, filter *AnalyticsFilter) (*CostAnalytics, error)
	GetCostOptimizationSuggestions(ctx context.Context, filter *AnalyticsFilter) ([]*OptimizationSuggestion, error)
	GetCostTrends(ctx context.Context, filter *AnalyticsFilter, interval string) ([]*TimeSeriesPoint, error)

	// Performance analytics
	GetPerformanceAnalytics(ctx context.Context, filter *AnalyticsFilter) (*PerformanceAnalytics, error)
	GetProviderComparison(ctx context.Context, filter *AnalyticsFilter) (*ProviderComparison, error)
	GetLatencyHeatmap(ctx context.Context, filter *AnalyticsFilter) (*LatencyHeatmap, error)

	// Time series analytics
	GetTimeSeries(ctx context.Context, filter *AnalyticsFilter, metric string, interval string) ([]*TimeSeriesPoint, error)
	GetCustomMetrics(ctx context.Context, projectID ulid.ULID, metricNames []string, timeRange TimeRange) (map[string][]*TimeSeriesPoint, error)

	// Export and reporting
	ExportTraces(ctx context.Context, filter *TraceFilter, format ExportFormat) (*ExportResult, error)
	GenerateReport(ctx context.Context, reportType ReportType, filter *AnalyticsFilter) (*Report, error)
}

// IngestionService defines the interface for high-throughput data ingestion
type IngestionService interface {
	// Batch ingestion
	IngestTraceBatch(ctx context.Context, request *BatchIngestRequest) (*BatchIngestResult, error)
	IngestObservationBatch(ctx context.Context, request *ObservationBatchRequest) (*BatchIngestResult, error)
	IngestQualityScoreBatch(ctx context.Context, request *QualityScoreBatchRequest) (*BatchIngestResult, error)

	// Queue management
	GetQueueStatus(ctx context.Context) (*QueueStatus, error)
	FlushQueue(ctx context.Context) error
	PauseIngestion(ctx context.Context) error
	ResumeIngestion(ctx context.Context) error

	// Health and monitoring
	GetIngestionHealth(ctx context.Context) (*IngestionHealth, error)
	GetIngestionMetrics(ctx context.Context) (*IngestionMetrics, error)
}

// Quality evaluator interface
type QualityEvaluator interface {
	Name() string
	Version() string
	Description() string
	SupportedTypes() []ObservationType
	Evaluate(ctx context.Context, input *EvaluationInput) (*QualityScore, error)
	ValidateInput(input *EvaluationInput) error
}

// Service request and response types

// ObservationCompletion represents data for completing an observation
type ObservationCompletion struct {
	EndTime         time.Time              `json:"end_time"`
	Output          map[string]any         `json:"output,omitempty"`
	Usage           *TokenUsage            `json:"usage,omitempty"`
	Cost            *CostCalculation       `json:"cost,omitempty"`
	QualityScore    *float64               `json:"quality_score,omitempty"`
	StatusMessage   *string                `json:"status_message,omitempty"`
	AdditionalData  map[string]any         `json:"additional_data,omitempty"`
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
	ProcessedCount int                    `json:"processed_count"`
	FailedCount    int                    `json:"failed_count"`
	Errors         []BatchIngestionError  `json:"errors,omitempty"`
	Duration       time.Duration          `json:"duration"`
	JobID          *string                `json:"job_id,omitempty"` // For async operations
}

// ObservationBatchRequest represents a batch observation ingestion request
type ObservationBatchRequest struct {
	ProjectID    ulid.ULID      `json:"project_id"`
	Observations []*Observation `json:"observations"`
	Async        bool           `json:"async"`
}

// QualityScoreBatchRequest represents a batch quality score ingestion request
type QualityScoreBatchRequest struct {
	ProjectID     ulid.ULID       `json:"project_id"`
	QualityScores []*QualityScore `json:"quality_scores"`
	Async         bool            `json:"async"`
}

// BatchIngestionError represents an error during batch ingestion
type BatchIngestionError struct {
	Index   int    `json:"index"`
	Error   string `json:"error"`
	Details any    `json:"details,omitempty"`
}

// BulkEvaluationRequest represents a bulk evaluation request
type BulkEvaluationRequest struct {
	TraceIDs        []ulid.ULID `json:"trace_ids,omitempty"`
	ObservationIDs  []ulid.ULID `json:"observation_ids,omitempty"`
	EvaluatorNames  []string    `json:"evaluator_names"`
	Filter          *AnalyticsFilter `json:"filter,omitempty"`
	Async           bool        `json:"async"`
}

// BulkEvaluationResult represents the result of bulk evaluation
type BulkEvaluationResult struct {
	ProcessedCount int                     `json:"processed_count"`
	FailedCount    int                     `json:"failed_count"`
	Scores         []*QualityScore         `json:"scores,omitempty"`
	Errors         []BulkEvaluationError   `json:"errors,omitempty"`
	JobID          *string                 `json:"job_id,omitempty"`
}

// BulkEvaluationError represents an error during bulk evaluation
type BulkEvaluationError struct {
	ItemID  ulid.ULID `json:"item_id"`
	Error   string    `json:"error"`
	Details any       `json:"details,omitempty"`
}

// EvaluationInput represents input for quality evaluation
type EvaluationInput struct {
	TraceID       *ulid.ULID     `json:"trace_id,omitempty"`
	ObservationID *ulid.ULID     `json:"observation_id,omitempty"`
	Trace         *Trace         `json:"trace,omitempty"`
	Observation   *Observation   `json:"observation,omitempty"`
	Context       map[string]any `json:"context,omitempty"`
}

// QualityEvaluatorInfo represents information about a quality evaluator
type QualityEvaluatorInfo struct {
	Name            string              `json:"name"`
	Version         string              `json:"version"`
	Description     string              `json:"description"`
	SupportedTypes  []ObservationType   `json:"supported_types"`
	IsBuiltIn       bool                `json:"is_built_in"`
	Configuration   map[string]any      `json:"configuration,omitempty"`
}

// Dashboard and reporting types

// DashboardOverview represents overview metrics for dashboard
type DashboardOverview struct {
	TotalTraces     int64                `json:"total_traces"`
	TotalCost       float64              `json:"total_cost"`
	AverageLatency  float64              `json:"average_latency"`
	ErrorRate       float64              `json:"error_rate"`
	TopProviders    []*ProviderSummary   `json:"top_providers"`
	RecentActivity  []*ActivityItem      `json:"recent_activity"`
	CostTrend       []*TimeSeriesPoint   `json:"cost_trend"`
	LatencyTrend    []*TimeSeriesPoint   `json:"latency_trend"`
	QualityTrend    []*TimeSeriesPoint   `json:"quality_trend"`
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
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]any         `json:"metadata,omitempty"`
}

// TimeRange represents a time range for analytics
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// OptimizationSuggestion represents a cost optimization suggestion
type OptimizationSuggestion struct {
	Type            string                 `json:"type"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	PotentialSavings float64              `json:"potential_savings"`
	Confidence      float64               `json:"confidence"`
	ActionItems     []string              `json:"action_items"`
	Metadata        map[string]any        `json:"metadata,omitempty"`
}

// LatencyHeatmap represents latency data in heatmap format
type LatencyHeatmap struct {
	Data      [][]float64 `json:"data"`
	XLabels   []string    `json:"x_labels"`
	YLabels   []string    `json:"y_labels"`
	MinValue  float64     `json:"min_value"`
	MaxValue  float64     `json:"max_value"`
}

// ThroughputMetrics represents throughput metrics
type ThroughputMetrics struct {
	RequestsPerSecond float64              `json:"requests_per_second"`
	RequestsPerMinute float64              `json:"requests_per_minute"`
	RequestsPerHour   float64              `json:"requests_per_hour"`
	PeakThroughput    float64              `json:"peak_throughput"`
	TimeSeries        []*TimeSeriesPoint   `json:"time_series"`
}

// Queue and ingestion monitoring types

// QueueStatus represents the status of ingestion queues
type QueueStatus struct {
	TraceQueue       *QueueInfo `json:"trace_queue"`
	ObservationQueue *QueueInfo `json:"observation_queue"`
	QualityQueue     *QueueInfo `json:"quality_queue"`
	TotalPending     int64      `json:"total_pending"`
	IsHealthy        bool       `json:"is_healthy"`
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
	Status          string                 `json:"status"` // healthy, degraded, unhealthy
	IngestionRate   float64                `json:"ingestion_rate"`
	ProcessingRate  float64                `json:"processing_rate"`
	ErrorRate       float64                `json:"error_rate"`
	Bottlenecks     []string               `json:"bottlenecks,omitempty"`
	LastCheck       time.Time              `json:"last_check"`
	Details         map[string]any         `json:"details,omitempty"`
}

// IngestionMetrics represents metrics for the ingestion system
type IngestionMetrics struct {
	TotalIngested     int64                `json:"total_ingested"`
	IngestedToday     int64                `json:"ingested_today"`
	ProcessingRate    float64              `json:"processing_rate"`
	AverageLatency    float64              `json:"average_latency"`
	QueueBacklog      int64                `json:"queue_backlog"`
	WorkerCount       int                  `json:"worker_count"`
	Errors            []*IngestionError    `json:"recent_errors,omitempty"`
}

// IngestionError represents an ingestion error
type IngestionError struct {
	Timestamp time.Time              `json:"timestamp"`
	Type      string                 `json:"type"`
	Message   string                 `json:"message"`
	Count     int64                  `json:"count"`
	Details   map[string]any         `json:"details,omitempty"`
}

// Export and reporting types

// ExportFormat represents supported export formats
type ExportFormat string

const (
	ExportFormatJSON ExportFormat = "json"
	ExportFormatCSV  ExportFormat = "csv"
	ExportFormatParquet ExportFormat = "parquet"
)

// ExportResult represents the result of an export operation
type ExportResult struct {
	DownloadURL  string        `json:"download_url"`
	Format       ExportFormat  `json:"format"`
	RecordCount  int64         `json:"record_count"`
	FileSize     int64         `json:"file_size"`
	ExpiresAt    time.Time     `json:"expires_at"`
	Status       string        `json:"status"`
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
	ID          ulid.ULID              `json:"id"`
	Type        ReportType             `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Data        map[string]any         `json:"data"`
	GeneratedAt time.Time              `json:"generated_at"`
	Format      ExportFormat           `json:"format"`
	DownloadURL string                 `json:"download_url,omitempty"`
}