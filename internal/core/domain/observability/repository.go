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
	InsertSpan(ctx context.Context, span *Span) error
	InsertSpanBatch(ctx context.Context, spans []*Span) error
	DeleteSpan(ctx context.Context, spanID string) error
	GetSpan(ctx context.Context, spanID string) (*Span, error)
	GetSpansByTraceID(ctx context.Context, traceID string) ([]*Span, error)
	GetSpanChildren(ctx context.Context, parentSpanID string) ([]*Span, error)
	GetSpanTree(ctx context.Context, traceID string) ([]*Span, error)
	GetSpansByFilter(ctx context.Context, filter *SpanFilter) ([]*Span, error)
	CountSpansByFilter(ctx context.Context, filter *SpanFilter) (int64, error)

	GetRootSpan(ctx context.Context, traceID string) (*Span, error)
	// GetRootSpanByProject retrieves root span with project ownership validation.
	// Returns error if trace doesn't exist or doesn't belong to the specified project.
	GetRootSpanByProject(ctx context.Context, traceID string, projectID string) (*Span, error)
	// GetSpanByProject retrieves span with project ownership validation.
	// Returns error if span doesn't exist or doesn't belong to the specified project.
	GetSpanByProject(ctx context.Context, spanID string, projectID string) (*Span, error)
	GetTraceSummary(ctx context.Context, traceID string) (*TraceSummary, error)
	ListTraces(ctx context.Context, filter *TraceFilter) ([]*TraceSummary, error)
	CountTraces(ctx context.Context, filter *TraceFilter) (int64, error)
	CountSpansInTrace(ctx context.Context, traceID string) (int64, error)
	DeleteTrace(ctx context.Context, traceID string) error

	// GetFilterOptions returns available filter values for populating the traces filter UI
	GetFilterOptions(ctx context.Context, projectID string) (*TraceFilterOptions, error)

	GetTracesBySessionID(ctx context.Context, sessionID string) ([]*TraceSummary, error)
	GetTracesByUserID(ctx context.Context, userID string, filter *TraceFilter) ([]*TraceSummary, error)
	CalculateTotalCost(ctx context.Context, traceID string) (float64, error)
	CalculateTotalTokens(ctx context.Context, traceID string) (uint64, error)

	QuerySpansByExpression(ctx context.Context, query string, args []interface{}) ([]*Span, error)
	CountSpansByExpression(ctx context.Context, query string, args []interface{}) (int64, error)
}

// ScoreRepository uses ReplacingMergeTree pattern for eventual consistency.
type ScoreRepository interface {
	Create(ctx context.Context, score *Score) error
	Update(ctx context.Context, score *Score) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*Score, error)

	GetByTraceID(ctx context.Context, traceID string) ([]*Score, error)
	GetBySpanID(ctx context.Context, spanID string) ([]*Score, error)

	GetByFilter(ctx context.Context, filter *ScoreFilter) ([]*Score, error)

	CreateBatch(ctx context.Context, scores []*Score) error

	Count(ctx context.Context, filter *ScoreFilter) (int64, error)

	ExistsByConfigName(ctx context.Context, projectID, configName string) (bool, error)

	// Returns: scoreName -> experimentID -> aggregation
	GetAggregationsByExperiments(ctx context.Context, projectID string, experimentIDs []string) (map[string]map[string]*ScoreAggregation, error)
}

type ScoreAggregation struct {
	Mean   float64 `json:"mean"`
	StdDev float64 `json:"std_dev"`
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
	Count  int64   `json:"count"`
}

type ScoreAnalyticsFilter struct {
	ProjectID        string     `json:"project_id"`
	ScoreName        string     `json:"score_name"`
	CompareScoreName *string    `json:"compare_score_name,omitempty"`
	FromTimestamp    *time.Time `json:"from_timestamp,omitempty"`
	ToTimestamp      *time.Time `json:"to_timestamp,omitempty"`
	Interval         string     `json:"interval"` // hour, day, week
}

type ScoreStatistics struct {
	Count       int64    `json:"count"`
	Mean        float64  `json:"mean"`
	StdDev      float64  `json:"std_dev"`
	Min         float64  `json:"min"`
	Max         float64  `json:"max"`
	Median      float64  `json:"median"`
	Mode        *string  `json:"mode,omitempty"`         // For categorical
	ModePercent *float64 `json:"mode_percent,omitempty"` // For categorical
}

type TimeSeriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	AvgValue  float64   `json:"avg_value"`
	Count     int64     `json:"count"`
}

type DistributionBin struct {
	BinStart float64 `json:"bin_start"`
	BinEnd   float64 `json:"bin_end"`
	Count    int64   `json:"count"`
}

type HeatmapCell struct {
	Row      int     `json:"row"`
	Col      int     `json:"col"`
	Value    int64   `json:"value"`
	RowLabel string  `json:"row_label"`
	ColLabel string  `json:"col_label"`
}

type ComparisonMetrics struct {
	MatchedCount        int64   `json:"matched_count"`
	PearsonCorrelation  float64 `json:"pearson_correlation"`
	SpearmanCorrelation float64 `json:"spearman_correlation"`
	MAE                 float64 `json:"mae"`
	RMSE                float64 `json:"rmse"`
	CohensKappa         float64 `json:"cohens_kappa,omitempty"`
	OverallAgreement    float64 `json:"overall_agreement,omitempty"`
}

type ScoreAnalyticsResponse struct {
	Statistics   *ScoreStatistics   `json:"statistics"`
	TimeSeries   []TimeSeriesPoint  `json:"time_series"`
	Distribution []DistributionBin  `json:"distribution"`
	Heatmap      []HeatmapCell      `json:"heatmap,omitempty"`
	Comparison   *ComparisonMetrics `json:"comparison,omitempty"`
}

type ScoreAnalyticsRepository interface {
	GetStatistics(ctx context.Context, filter *ScoreAnalyticsFilter) (*ScoreStatistics, error)
	GetTimeSeries(ctx context.Context, filter *ScoreAnalyticsFilter) ([]TimeSeriesPoint, error)
	GetDistribution(ctx context.Context, filter *ScoreAnalyticsFilter, bins int) ([]DistributionBin, error)
	GetHeatmap(ctx context.Context, filter *ScoreAnalyticsFilter, bins int) ([]HeatmapCell, error)
	GetComparisonMetrics(ctx context.Context, filter *ScoreAnalyticsFilter) (*ComparisonMetrics, error)
	GetDistinctScoreNames(ctx context.Context, projectID string) ([]string, error)
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
	ProjectID string // Required for scoping queries to project

	TraceID      *string
	ParentID     *string
	Type         *string
	SpanKind     *string
	Model        *string
	ServiceName  *string   // OTLP: service.name (materialized column for fast filtering)
	SpanNames    []string  // Filter by one or more span names (OR logic)
	StartTime    *time.Time
	EndTime      *time.Time
	MinLatencyMs *uint32
	MaxLatencyMs *uint32
	MinCost      *float64
	MaxCost      *float64
	Level        *string
	IsCompleted  *bool

	pagination.Params
}

type ScoreFilter struct {
	ProjectID string // Required for scoping queries to project

	TraceID   *string
	SpanID    *string
	Name      *string
	Source    *string
	DataType  *string
	MinValue  *float64
	MaxValue  *float64
	StartTime *time.Time
	EndTime   *time.Time

	pagination.Params
}

// TraceFilterOptions represents available values for filter UI dropdowns.
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

type Range struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

type TelemetryDeduplicationRepository interface {
	// Atomic claim for deduplication - returns which IDs were successfully claimed vs already processed.
	ClaimEvents(ctx context.Context, projectID ulid.ULID, batchID ulid.ULID, dedupIDs []string, ttl time.Duration) (claimed []string, duplicates []string, err error)
	ReleaseEvents(ctx context.Context, dedupIDs []string) error

	CheckDuplicate(ctx context.Context, dedupID string) (bool, error)
	RegisterEvent(ctx context.Context, dedupID string, batchID ulid.ULID, projectID ulid.ULID, ttl time.Duration) error
	Exists(ctx context.Context, dedupID string) (bool, error)
	Create(ctx context.Context, dedup *TelemetryEventDeduplication) error
	Delete(ctx context.Context, dedupID string) error

	CheckBatchDuplicates(ctx context.Context, dedupIDs []string) ([]string, error)
	CreateBatch(ctx context.Context, dedups []*TelemetryEventDeduplication) error

	CountByProjectID(ctx context.Context, projectID ulid.ULID) (int64, error)
}

type MetricsRepository interface {
	CreateMetricSumBatch(ctx context.Context, metricsSums []*MetricSum) error
	CreateMetricGaugeBatch(ctx context.Context, metricsGauges []*MetricGauge) error
	CreateMetricHistogramBatch(ctx context.Context, metricsHistograms []*MetricHistogram) error
	CreateMetricExponentialHistogramBatch(ctx context.Context, metricsExpHistograms []*MetricExponentialHistogram) error
}

type LogsRepository interface {
	CreateLogBatch(ctx context.Context, logs []*Log) error
}

type GenAIEventsRepository interface {
	CreateGenAIEventBatch(ctx context.Context, events []*GenAIEvent) error
}

// ModelRepository removed - use analytics.ProviderModelRepository instead
