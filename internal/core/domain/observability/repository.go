package observability

import (
	"context"
	"time"

	"brokle/pkg/ulid"
)

// TraceRepository defines the interface for trace data access
type TraceRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, trace *Trace) error
	GetByID(ctx context.Context, id ulid.ULID) (*Trace, error)
	GetByExternalTraceID(ctx context.Context, externalTraceID string) (*Trace, error)
	Update(ctx context.Context, trace *Trace) error
	Delete(ctx context.Context, id ulid.ULID) error

	// Project-scoped queries
	GetByProjectID(ctx context.Context, projectID ulid.ULID, limit, offset int) ([]*Trace, error)
	GetByUserID(ctx context.Context, userID ulid.ULID, limit, offset int) ([]*Trace, error)
	GetBySessionID(ctx context.Context, sessionID ulid.ULID) ([]*Trace, error)

	// Advanced queries
	SearchTraces(ctx context.Context, filter *TraceFilter) ([]*Trace, int, error)
	GetTraceWithObservations(ctx context.Context, id ulid.ULID) (*Trace, error)
	GetTraceStats(ctx context.Context, id ulid.ULID) (*TraceStats, error)

	// Batch operations
	CreateBatch(ctx context.Context, traces []*Trace) error
	UpdateBatch(ctx context.Context, traces []*Trace) error
	DeleteBatch(ctx context.Context, ids []ulid.ULID) error

	// Aggregation queries
	GetTracesByTimeRange(ctx context.Context, projectID ulid.ULID, startTime, endTime time.Time, limit, offset int) ([]*Trace, error)
	CountTraces(ctx context.Context, filter *TraceFilter) (int64, error)
	GetRecentTraces(ctx context.Context, projectID ulid.ULID, limit int) ([]*Trace, error)
}

// ObservationRepository defines the interface for observation data access
type ObservationRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, observation *Observation) error
	GetByID(ctx context.Context, id ulid.ULID) (*Observation, error)
	GetByExternalObservationID(ctx context.Context, externalObservationID string) (*Observation, error)
	Update(ctx context.Context, observation *Observation) error
	Delete(ctx context.Context, id ulid.ULID) error

	// Trace-scoped queries
	GetByTraceID(ctx context.Context, traceID ulid.ULID) ([]*Observation, error)
	GetByParentObservationID(ctx context.Context, parentID ulid.ULID) ([]*Observation, error)

	// Type and provider queries
	GetByType(ctx context.Context, obsType ObservationType, limit, offset int) ([]*Observation, error)
	GetByProvider(ctx context.Context, provider string, limit, offset int) ([]*Observation, error)
	GetByModel(ctx context.Context, provider, model string, limit, offset int) ([]*Observation, error)

	// Advanced queries
	SearchObservations(ctx context.Context, filter *ObservationFilter) ([]*Observation, int, error)
	GetObservationStats(ctx context.Context, filter *ObservationFilter) (*ObservationStats, error)

	// Batch operations
	CreateBatch(ctx context.Context, observations []*Observation) error
	UpdateBatch(ctx context.Context, observations []*Observation) error
	DeleteBatch(ctx context.Context, ids []ulid.ULID) error

	// Completion operations
	CompleteObservation(ctx context.Context, id ulid.ULID, endTime time.Time, output any, cost *float64) error
	GetIncompleteObservations(ctx context.Context, projectID ulid.ULID) ([]*Observation, error)

	// Analytics queries
	GetObservationsByTimeRange(ctx context.Context, filter *ObservationFilter, startTime, endTime time.Time) ([]*Observation, error)
	CountObservations(ctx context.Context, filter *ObservationFilter) (int64, error)
}

// QualityScoreRepository defines the interface for quality score data access
type QualityScoreRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, score *QualityScore) error
	GetByID(ctx context.Context, id ulid.ULID) (*QualityScore, error)
	Update(ctx context.Context, score *QualityScore) error
	Delete(ctx context.Context, id ulid.ULID) error

	// Trace and observation scoped queries
	GetByTraceID(ctx context.Context, traceID ulid.ULID) ([]*QualityScore, error)
	GetByObservationID(ctx context.Context, observationID ulid.ULID) ([]*QualityScore, error)

	// Score queries
	GetByScoreName(ctx context.Context, scoreName string, limit, offset int) ([]*QualityScore, error)
	GetBySource(ctx context.Context, source ScoreSource, limit, offset int) ([]*QualityScore, error)
	GetByEvaluator(ctx context.Context, evaluatorName string, limit, offset int) ([]*QualityScore, error)

	// Unique score operations
	GetByTraceAndScoreName(ctx context.Context, traceID ulid.ULID, scoreName string) (*QualityScore, error)
	GetByObservationAndScoreName(ctx context.Context, observationID ulid.ULID, scoreName string) (*QualityScore, error)

	// Batch operations
	CreateBatch(ctx context.Context, scores []*QualityScore) error
	UpdateBatch(ctx context.Context, scores []*QualityScore) error
	DeleteBatch(ctx context.Context, ids []ulid.ULID) error

	// Analytics queries
	GetAverageScoreByName(ctx context.Context, scoreName string, filter *QualityScoreFilter) (float64, error)
	GetScoreDistribution(ctx context.Context, scoreName string, filter *QualityScoreFilter) (map[string]int, error)
	GetScoresByTimeRange(ctx context.Context, filter *QualityScoreFilter, startTime, endTime time.Time) ([]*QualityScore, error)
}

// AnalyticsRepository defines the interface for analytics and aggregated data access
type AnalyticsRepository interface {
	// Trace analytics
	GetTraceAnalytics(ctx context.Context, filter *AnalyticsFilter) (*TraceAnalytics, error)
	GetTraceTimeSeries(ctx context.Context, filter *AnalyticsFilter, interval string) ([]*TimeSeriesPoint, error)

	// Observation analytics
	GetObservationAnalytics(ctx context.Context, filter *AnalyticsFilter) (*ObservationAnalytics, error)
	GetObservationTimeSeries(ctx context.Context, filter *AnalyticsFilter, interval string, metric string) ([]*TimeSeriesPoint, error)

	// Cost analytics
	GetCostAnalytics(ctx context.Context, filter *AnalyticsFilter) (*CostAnalytics, error)
	GetCostBreakdown(ctx context.Context, filter *AnalyticsFilter, groupBy string) ([]*CostBreakdown, error)

	// Performance analytics
	GetPerformanceAnalytics(ctx context.Context, filter *AnalyticsFilter) (*PerformanceAnalytics, error)
	GetLatencyPercentiles(ctx context.Context, filter *AnalyticsFilter) (*LatencyPercentiles, error)

	// Provider analytics
	GetProviderAnalytics(ctx context.Context, filter *AnalyticsFilter) ([]*ProviderAnalytics, error)
	GetProviderComparison(ctx context.Context, filter *AnalyticsFilter) (*ProviderComparison, error)

	// Quality analytics
	GetQualityAnalytics(ctx context.Context, filter *AnalyticsFilter) (*QualityAnalytics, error)
	GetQualityTrends(ctx context.Context, filter *AnalyticsFilter, interval string) ([]*QualityTrendPoint, error)

	// Real-time analytics
	GetRealtimeMetrics(ctx context.Context, projectID ulid.ULID) (*RealtimeMetrics, error)
}

// Repository aggregates all observability-related repositories
type Repository interface {
	Traces() TraceRepository
	Observations() ObservationRepository
	QualityScores() QualityScoreRepository
	Analytics() AnalyticsRepository
}

// Filter structures for repository queries

// TraceFilter represents filters for trace queries
type TraceFilter struct {
	ProjectID         *ulid.ULID             `json:"project_id,omitempty"`
	UserID            *ulid.ULID             `json:"user_id,omitempty"`
	SessionID         *ulid.ULID             `json:"session_id,omitempty"`
	Name              *string                `json:"name,omitempty"`
	ExternalTraceID   *string                `json:"external_trace_id,omitempty"`
	StartTime         *time.Time             `json:"start_time,omitempty"`
	EndTime           *time.Time             `json:"end_time,omitempty"`
	Tags              map[string]any         `json:"tags,omitempty"`
	ExcludeTags       map[string]any         `json:"exclude_tags,omitempty"`
	HasObservationType *ObservationType      `json:"has_observation_type,omitempty"`
	MinObservations   *int                   `json:"min_observations,omitempty"`
	MaxObservations   *int                   `json:"max_observations,omitempty"`
	MinCost           *float64               `json:"min_cost,omitempty"`
	MaxCost           *float64               `json:"max_cost,omitempty"`
	SortBy            string                 `json:"sort_by"`         // created_at, updated_at, name
	SortOrder         string                 `json:"sort_order"`      // asc, desc
	Limit             int                    `json:"limit"`
	Offset            int                    `json:"offset"`
}

// ObservationFilter represents filters for observation queries
type ObservationFilter struct {
	TraceID         *ulid.ULID       `json:"trace_id,omitempty"`
	Type            *ObservationType `json:"type,omitempty"`
	Provider        *string          `json:"provider,omitempty"`
	Model           *string          `json:"model,omitempty"`
	StartTime       *time.Time       `json:"start_time,omitempty"`
	EndTime         *time.Time       `json:"end_time,omitempty"`
	MinLatency      *int             `json:"min_latency,omitempty"`
	MaxLatency      *int             `json:"max_latency,omitempty"`
	MinCost         *float64         `json:"min_cost,omitempty"`
	MaxCost         *float64         `json:"max_cost,omitempty"`
	MinTokens       *int             `json:"min_tokens,omitempty"`
	MaxTokens       *int             `json:"max_tokens,omitempty"`
	MinQualityScore *float64         `json:"min_quality_score,omitempty"`
	MaxQualityScore *float64         `json:"max_quality_score,omitempty"`
	Level           *ObservationLevel `json:"level,omitempty"`
	IsCompleted     *bool            `json:"is_completed,omitempty"`
	HasError        *bool            `json:"has_error,omitempty"`
	SortBy          string           `json:"sort_by"`    // start_time, end_time, latency, cost
	SortOrder       string           `json:"sort_order"` // asc, desc
	Limit           int              `json:"limit"`
	Offset          int              `json:"offset"`
}

// QualityScoreFilter represents filters for quality score queries
type QualityScoreFilter struct {
	TraceID        *ulid.ULID    `json:"trace_id,omitempty"`
	ObservationID  *ulid.ULID    `json:"observation_id,omitempty"`
	ScoreName      *string       `json:"score_name,omitempty"`
	Source         *ScoreSource  `json:"source,omitempty"`
	DataType       *ScoreDataType `json:"data_type,omitempty"`
	EvaluatorName  *string       `json:"evaluator_name,omitempty"`
	AuthorUserID   *ulid.ULID    `json:"author_user_id,omitempty"`
	MinScore       *float64      `json:"min_score,omitempty"`
	MaxScore       *float64      `json:"max_score,omitempty"`
	StartTime      *time.Time    `json:"start_time,omitempty"`
	EndTime        *time.Time    `json:"end_time,omitempty"`
	SortBy         string        `json:"sort_by"`    // created_at, score_value, score_name
	SortOrder      string        `json:"sort_order"` // asc, desc
	Limit          int           `json:"limit"`
	Offset         int           `json:"offset"`
}

// AnalyticsFilter represents filters for analytics queries
type AnalyticsFilter struct {
	ProjectID      ulid.ULID      `json:"project_id"`
	UserID         *ulid.ULID     `json:"user_id,omitempty"`
	SessionID      *ulid.ULID     `json:"session_id,omitempty"`
	StartTime      time.Time      `json:"start_time"`
	EndTime        time.Time      `json:"end_time"`
	Provider       *string        `json:"provider,omitempty"`
	Model          *string        `json:"model,omitempty"`
	ObservationType *ObservationType `json:"observation_type,omitempty"`
	Tags           map[string]any `json:"tags,omitempty"`
}

// Analytics result structures

// TraceAnalytics represents aggregated trace analytics
type TraceAnalytics struct {
	TotalTraces       int64   `json:"total_traces"`
	CompletedTraces   int64   `json:"completed_traces"`
	AverageLatency    float64 `json:"average_latency"`
	TotalCost         float64 `json:"total_cost"`
	AverageCost       float64 `json:"average_cost"`
	TotalObservations int64   `json:"total_observations"`
	UniqueUsers       int64   `json:"unique_users"`
	UniqueSessions    int64   `json:"unique_sessions"`
}

// ObservationAnalytics represents aggregated observation analytics
type ObservationAnalytics struct {
	TotalObservations   int64   `json:"total_observations"`
	CompletedObservations int64 `json:"completed_observations"`
	AverageLatency      float64 `json:"average_latency"`
	TotalTokens         int64   `json:"total_tokens"`
	TotalCost           float64 `json:"total_cost"`
	AverageQualityScore float64 `json:"average_quality_score"`
	ErrorRate           float64 `json:"error_rate"`
	ThroughputPerHour   float64 `json:"throughput_per_hour"`
}

// CostAnalytics represents cost-related analytics
type CostAnalytics struct {
	TotalCost           float64 `json:"total_cost"`
	InputCost           float64 `json:"input_cost"`
	OutputCost          float64 `json:"output_cost"`
	AverageCostPerToken float64 `json:"average_cost_per_token"`
	CostSavings         float64 `json:"cost_savings"`
	TopCostProviders    []ProviderCost `json:"top_cost_providers"`
}

// PerformanceAnalytics represents performance-related analytics
type PerformanceAnalytics struct {
	AverageLatency    float64 `json:"average_latency"`
	MedianLatency     float64 `json:"median_latency"`
	P95Latency        float64 `json:"p95_latency"`
	P99Latency        float64 `json:"p99_latency"`
	Throughput        float64 `json:"throughput"`
	ErrorRate         float64 `json:"error_rate"`
	TimeoutRate       float64 `json:"timeout_rate"`
}

// ProviderAnalytics represents provider-specific analytics
type ProviderAnalytics struct {
	Provider        string  `json:"provider"`
	TotalRequests   int64   `json:"total_requests"`
	AverageLatency  float64 `json:"average_latency"`
	ErrorRate       float64 `json:"error_rate"`
	TotalCost       float64 `json:"total_cost"`
	AverageQuality  float64 `json:"average_quality"`
	MarketShare     float64 `json:"market_share"`
}

// QualityAnalytics represents quality-related analytics
type QualityAnalytics struct {
	AverageQuality    float64            `json:"average_quality"`
	QualityDistribution map[string]int   `json:"quality_distribution"`
	TopEvaluators     []EvaluatorStats   `json:"top_evaluators"`
	QualityTrend      string             `json:"quality_trend"` // improving, declining, stable
}

// Supporting types for analytics

// TimeSeriesPoint represents a time series data point
type TimeSeriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// CostBreakdown represents cost breakdown by dimension
type CostBreakdown struct {
	Dimension   string  `json:"dimension"`
	TotalCost   float64 `json:"total_cost"`
	RequestCount int64  `json:"request_count"`
	AverageCost float64 `json:"average_cost"`
}

// LatencyPercentiles represents latency percentile data
type LatencyPercentiles struct {
	P50 float64 `json:"p50"`
	P75 float64 `json:"p75"`
	P90 float64 `json:"p90"`
	P95 float64 `json:"p95"`
	P99 float64 `json:"p99"`
}

// ProviderComparison represents comparison between providers
type ProviderComparison struct {
	Providers []ProviderAnalytics `json:"providers"`
	Winner    string              `json:"winner"`
	Criteria  string              `json:"criteria"` // cost, latency, quality
}

// QualityTrendPoint represents quality trend data
type QualityTrendPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Score     float64   `json:"score"`
	Count     int64     `json:"count"`
}

// RealtimeMetrics represents real-time metrics
type RealtimeMetrics struct {
	ActiveTraces        int64   `json:"active_traces"`
	RequestsPerMinute   float64 `json:"requests_per_minute"`
	AverageLatency      float64 `json:"average_latency"`
	ErrorRate           float64 `json:"error_rate"`
	CostPerHour         float64 `json:"cost_per_hour"`
	TopProviders        []string `json:"top_providers"`
	LastUpdated         time.Time `json:"last_updated"`
}

// ProviderCost represents cost information for a provider
type ProviderCost struct {
	Provider  string  `json:"provider"`
	TotalCost float64 `json:"total_cost"`
	Requests  int64   `json:"requests"`
}

// EvaluatorStats represents statistics for an evaluator
type EvaluatorStats struct {
	Name           string  `json:"name"`
	TotalScores    int64   `json:"total_scores"`
	AverageScore   float64 `json:"average_score"`
	LastEvaluation time.Time `json:"last_evaluation"`
}