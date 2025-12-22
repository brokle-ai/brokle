package observability

import (
	"context"
	"time"

	"brokle/pkg/pagination"
	"brokle/pkg/ulid"
)

// TraceRepository defines the interface for trace and span operations (ClickHouse).
// Traces are virtual (derived from root spans where parent_span_id IS NULL).
type TraceRepository interface {
	// Span Operations
	InsertSpan(ctx context.Context, span *Span) error
	InsertSpanBatch(ctx context.Context, spans []*Span) error
	DeleteSpan(ctx context.Context, spanID string) error
	GetSpan(ctx context.Context, spanID string) (*Span, error)
	GetSpansByTraceID(ctx context.Context, traceID string) ([]*Span, error)
	GetSpanChildren(ctx context.Context, parentSpanID string) ([]*Span, error)
	GetSpanTree(ctx context.Context, traceID string) ([]*Span, error)
	GetSpansByFilter(ctx context.Context, filter *SpanFilter) ([]*Span, error)
	CountSpansByFilter(ctx context.Context, filter *SpanFilter) (int64, error)

	// Trace Operations
	GetRootSpan(ctx context.Context, traceID string) (*Span, error)
	GetTraceSummary(ctx context.Context, traceID string) (*TraceSummary, error)
	ListTraces(ctx context.Context, filter *TraceFilter) ([]*TraceSummary, error)
	CountTraces(ctx context.Context, filter *TraceFilter) (int64, error)
	CountSpansInTrace(ctx context.Context, traceID string) (int64, error)
	DeleteTrace(ctx context.Context, traceID string) error

	// GetFilterOptions returns available filter values for populating the traces filter UI
	GetFilterOptions(ctx context.Context, projectID string) (*TraceFilterOptions, error)

	// Analytics
	GetTracesBySessionID(ctx context.Context, sessionID string) ([]*TraceSummary, error)
	GetTracesByUserID(ctx context.Context, userID string, filter *TraceFilter) ([]*TraceSummary, error)
	CalculateTotalCost(ctx context.Context, traceID string) (float64, error)
	CalculateTotalTokens(ctx context.Context, traceID string) (uint64, error)

	QuerySpansByExpression(ctx context.Context, query string, args []interface{}) ([]*Span, error)
	CountSpansByExpression(ctx context.Context, query string, args []interface{}) (int64, error)
}

// ScoreRepository defines the interface for score data access (ClickHouse)
type ScoreRepository interface {
	// Basic operations (ReplacingMergeTree pattern)
	Create(ctx context.Context, score *Score) error
	Update(ctx context.Context, score *Score) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*Score, error)

	// Queries
	GetByTraceID(ctx context.Context, traceID string) ([]*Score, error)
	GetBySpanID(ctx context.Context, spanID string) ([]*Score, error)

	// Filters
	GetByFilter(ctx context.Context, filter *ScoreFilter) ([]*Score, error)

	// Batch operations
	CreateBatch(ctx context.Context, scores []*Score) error

	// Count
	Count(ctx context.Context, filter *ScoreFilter) (int64, error)

	ExistsByConfigName(ctx context.Context, projectID, configName string) (bool, error)

	// Aggregations for experiment comparison
	// Returns: scoreName -> experimentID -> aggregation
	GetAggregationsByExperiments(ctx context.Context, projectID string, experimentIDs []string) (map[string]map[string]*ScoreAggregation, error)
}

// ScoreAggregation holds statistical metrics for score aggregations.
type ScoreAggregation struct {
	Mean   float64 `json:"mean"`
	StdDev float64 `json:"std_dev"`
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
	Count  int64   `json:"count"`
}

type TraceFilter struct {
	UserID      *string
	SessionID   *string
	StartTime   *time.Time
	EndTime     *time.Time
	Environment *string
	ServiceName *string // OTLP: service.name (materialized column for fast filtering)
	StatusCode  *string
	Bookmarked  *bool
	Public      *bool

	ModelName    *string
	ProviderName *string
	MinCost      *float64
	MaxCost      *float64
	MinTokens    *int64
	MaxTokens    *int64
	MinDuration  *int64
	MaxDuration  *int64
	HasError     *bool

	pagination.Params
	ProjectID string
	Tags      []string
}

type SpanFilter struct {
	// Scope
	ProjectID string // Required for scoping queries to project

	// Domain filters
	TraceID      *string
	ParentID     *string
	Type         *string
	SpanKind     *string
	Model        *string
	ServiceName  *string // OTLP: service.name (materialized column for fast filtering)
	StartTime    *time.Time
	EndTime      *time.Time
	MinLatencyMs *uint32
	MaxLatencyMs *uint32
	MinCost      *float64
	MaxCost      *float64
	Level        *string
	IsCompleted  *bool

	// Pagination (embedded for DRY)
	pagination.Params
}

type ScoreFilter struct {
	// Scope
	ProjectID string // Required for scoping queries to project

	// Domain filters
	TraceID   *string
	SpanID    *string
	Name      *string
	Source    *string
	DataType  *string
	MinValue  *float64
	MaxValue  *float64
	StartTime *time.Time
	EndTime   *time.Time

	// Pagination (embedded for DRY)
	pagination.Params
}

// TraceFilterOptions represents available filter values for populating filter UI
type TraceFilterOptions struct {
	Models        []string `json:"models"`
	Providers     []string `json:"providers"`
	Services      []string `json:"services"`
	Environments  []string `json:"environments"`
	Users         []string `json:"users"`
	Sessions      []string `json:"sessions"`
	CostRange     *Range   `json:"cost_range"`
	TokenRange    *Range   `json:"token_range"`
	DurationRange *Range   `json:"duration_range"`
}

// Range represents a min/max numeric range for filter options
type Range struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// TelemetryDeduplicationRepository defines methods for telemetry deduplication
type TelemetryDeduplicationRepository interface {
	// Atomic claim operations for deduplication
	ClaimEvents(ctx context.Context, projectID ulid.ULID, batchID ulid.ULID, dedupIDs []string, ttl time.Duration) (claimed []string, duplicates []string, err error)
	ReleaseEvents(ctx context.Context, dedupIDs []string) error

	// Individual operations
	CheckDuplicate(ctx context.Context, dedupID string) (bool, error)
	RegisterEvent(ctx context.Context, dedupID string, batchID ulid.ULID, projectID ulid.ULID, ttl time.Duration) error
	Exists(ctx context.Context, dedupID string) (bool, error)
	Create(ctx context.Context, dedup *TelemetryEventDeduplication) error
	Delete(ctx context.Context, dedupID string) error

	// Batch operations
	CheckBatchDuplicates(ctx context.Context, dedupIDs []string) ([]string, error)
	CreateBatch(ctx context.Context, dedups []*TelemetryEventDeduplication) error

	// Statistics
	CountByProjectID(ctx context.Context, projectID ulid.ULID) (int64, error)
}

// MetricsRepository defines the interface for OTLP metrics data access (ClickHouse)
type MetricsRepository interface {
	// Batch operations (primary API for metrics ingestion)
	CreateMetricSumBatch(ctx context.Context, metricsSums []*MetricSum) error
	CreateMetricGaugeBatch(ctx context.Context, metricsGauges []*MetricGauge) error
	CreateMetricHistogramBatch(ctx context.Context, metricsHistograms []*MetricHistogram) error
	CreateMetricExponentialHistogramBatch(ctx context.Context, metricsExpHistograms []*MetricExponentialHistogram) error
}

// LogsRepository defines the interface for OTLP logs data access (ClickHouse)
type LogsRepository interface {
	// Batch operations (primary API for logs ingestion)
	CreateLogBatch(ctx context.Context, logs []*Log) error
}

// GenAIEventsRepository defines the interface for OTLP GenAI events data access (ClickHouse)
type GenAIEventsRepository interface {
	// Batch operations (primary API for GenAI events ingestion)
	CreateGenAIEventBatch(ctx context.Context, events []*GenAIEvent) error
}

// ModelRepository removed - use analytics.ProviderModelRepository instead
