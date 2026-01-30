package evaluation

import (
	"time"

	"brokle/pkg/ulid"
)

// RuleAnalyticsParams defines the parameters for fetching rule analytics.
type RuleAnalyticsParams struct {
	ProjectID ulid.ULID
	RuleID    ulid.ULID
	Period    string     // "24h", "7d", "30d" (default: "7d")
	From      *time.Time // Optional custom start time
	To        *time.Time // Optional custom end time
}

// RuleAnalyticsResponse contains comprehensive analytics for a rule.
type RuleAnalyticsResponse struct {
	RuleID             string               `json:"rule_id"`
	Period             string               `json:"period"`
	TotalExecutions    int64                `json:"total_executions"`
	TotalSpansScored   int64                `json:"total_spans_scored"`
	SuccessRate        float64              `json:"success_rate"`        // Percentage of successful executions
	AverageScore       float64              `json:"average_score"`       // Mean score value across all scored spans
	ScoreDistribution  []DistributionBucket `json:"score_distribution"`  // Histogram of score values
	ExecutionTrend     []TimeSeriesPoint    `json:"execution_trend"`     // Executions over time
	ScoreTrend         []TimeSeriesPoint    `json:"score_trend"`         // Average score over time
	LatencyPercentiles LatencyStats         `json:"latency_percentiles"` // P50, P90, P99 latencies
	TopErrors          []ErrorSummary       `json:"top_errors"`          // Most common error types
	CostEstimate       *CostEstimate        `json:"cost_estimate,omitempty"` // Estimated cost for LLM rules
}

// DistributionBucket represents a bucket in a histogram.
type DistributionBucket struct {
	BinStart   float64 `json:"bin_start"`
	BinEnd     float64 `json:"bin_end"`
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage,omitempty"`
}

// TimeSeriesPoint represents a single data point in a time series.
// Used for execution trends and score trends in rule analytics.
type TimeSeriesPoint struct {
	Timestamp   time.Time `json:"timestamp"`
	Count       int64     `json:"count"`
	SuccessRate float64   `json:"success_rate"` // 0.0-1.0
	AvgScore    *float64  `json:"avg_score"`    // nullable - only present for score trends
}

// LatencyStats contains latency percentile statistics.
type LatencyStats struct {
	P50  int64   `json:"p50"`  // 50th percentile (median)
	P90  int64   `json:"p90"`  // 90th percentile
	P99  int64   `json:"p99"`  // 99th percentile
	Avg  float64 `json:"avg"`  // Average latency
	Max  int64   `json:"max"`  // Maximum latency
	Min  int64   `json:"min"`  // Minimum latency
}

// ErrorSummary groups errors by type with occurrence count.
type ErrorSummary struct {
	ErrorType    string    `json:"error_type"`
	Message      string    `json:"message"`
	Count        int64     `json:"count"`
	LastOccurred time.Time `json:"last_occurred"`
}

// CostEstimate provides cost breakdown for LLM-based rules.
type CostEstimate struct {
	TotalCost        float64 `json:"total_cost"`
	InputTokens      int64   `json:"input_tokens"`
	OutputTokens     int64   `json:"output_tokens"`
	EstimatedMonthly float64 `json:"estimated_monthly"`
}

// RuleAnalyticsRepository defines the interface for fetching rule analytics.
type RuleAnalyticsRepository interface {
	// GetAnalytics retrieves comprehensive analytics for a rule.
	GetAnalytics(params *RuleAnalyticsParams) (*RuleAnalyticsResponse, error)

	// GetScoreDistribution returns the histogram of score values.
	GetScoreDistribution(projectID, ruleID ulid.ULID, from, to time.Time, buckets int) ([]DistributionBucket, error)

	// GetExecutionTrend returns executions over time.
	GetExecutionTrend(projectID, ruleID ulid.ULID, from, to time.Time, interval string) ([]TimeSeriesPoint, error)

	// GetScoreTrend returns average scores over time.
	GetScoreTrend(projectID, ruleID ulid.ULID, from, to time.Time, interval string) ([]TimeSeriesPoint, error)

	// GetLatencyStats returns latency percentile statistics.
	GetLatencyStats(projectID, ruleID ulid.ULID, from, to time.Time) (*LatencyStats, error)

	// GetTopErrors returns the most common errors.
	GetTopErrors(projectID, ruleID ulid.ULID, from, to time.Time, limit int) ([]ErrorSummary, error)
}

// ExecutionDetailRequest defines the parameters for fetching execution details.
type ExecutionDetailRequest struct {
	ProjectID   ulid.ULID
	RuleID      ulid.ULID
	ExecutionID ulid.ULID
}

// ExecutionDetailResponse contains detailed information about a specific execution.
type ExecutionDetailResponse struct {
	Execution    *RuleExecutionResponse `json:"execution"`
	Spans        []SpanExecutionDetail  `json:"spans"`
	RuleSnapshot *RuleSnapshot          `json:"rule_snapshot,omitempty"`
}

// SpanExecutionDetail contains detailed execution info for a single span.
type SpanExecutionDetail struct {
	SpanID            string              `json:"span_id"`
	TraceID           string              `json:"trace_id"`
	SpanName          string              `json:"span_name"`
	Status            string              `json:"status"` // success, failed, skipped
	ScoreResults      []ExecutionScoreResult `json:"score_results"`
	PromptSent        []LLMMessage        `json:"prompt_sent,omitempty"`
	LLMResponseRaw    string              `json:"llm_response_raw,omitempty"`
	LLMResponseParsed map[string]any      `json:"llm_response_parsed,omitempty"`
	VariablesResolved []ResolvedVariable  `json:"variables_resolved"`
	ErrorMessage      string              `json:"error_message,omitempty"`
	ErrorStack        string              `json:"error_stack,omitempty"`
	LatencyMs         int64               `json:"latency_ms,omitempty"`
	CreatedAt         time.Time           `json:"created_at"`
}

// ExecutionScoreResult represents a score written during execution.
type ExecutionScoreResult struct {
	ScoreName  string  `json:"score_name"`
	Value      any     `json:"value"`
	Reasoning  string  `json:"reasoning,omitempty"`
	Confidence float64 `json:"confidence,omitempty"`
	RawOutput  any     `json:"raw_output,omitempty"`
}

// RuleSnapshot captures the rule configuration at execution time.
type RuleSnapshot struct {
	Name            string         `json:"name"`
	ScorerType      ScorerType     `json:"scorer_type"`
	ScorerConfig    map[string]any `json:"scorer_config"`
	VariableMapping []VariableMap  `json:"variable_mapping"`
	Filter          []FilterClause `json:"filter"`
}

// RuleExecutionDetailFlat is the API response for execution detail endpoint.
// Uses flattened structure for frontend compatibility (execution fields at root level).
type RuleExecutionDetailFlat struct {
	// Embedded execution fields (flattened from RuleExecutionResponse)
	ID           string          `json:"id"`
	RuleID       string          `json:"rule_id"`
	ProjectID    string          `json:"project_id"`
	Status       ExecutionStatus `json:"status"`
	TriggerType  TriggerType     `json:"trigger_type"`
	SpansMatched int             `json:"spans_matched"`
	SpansScored  int             `json:"spans_scored"`
	ErrorsCount  int             `json:"errors_count"`
	ErrorMessage *string         `json:"error_message,omitempty"`
	StartedAt    *time.Time      `json:"started_at,omitempty"`
	CompletedAt  *time.Time      `json:"completed_at,omitempty"`
	DurationMs   *int            `json:"duration_ms,omitempty"`
	Metadata     map[string]any  `json:"metadata,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`

	// Detail-specific fields
	Spans        []SpanExecutionDetail `json:"spans"`
	RuleSnapshot *RuleSnapshot         `json:"rule_snapshot,omitempty"`
}

// ToFlat converts ExecutionDetailResponse to flattened API response.
func (r *ExecutionDetailResponse) ToFlat() *RuleExecutionDetailFlat {
	if r == nil || r.Execution == nil {
		return nil
	}
	return &RuleExecutionDetailFlat{
		ID:           r.Execution.ID,
		RuleID:       r.Execution.RuleID,
		ProjectID:    r.Execution.ProjectID,
		Status:       r.Execution.Status,
		TriggerType:  r.Execution.TriggerType,
		SpansMatched: r.Execution.SpansMatched,
		SpansScored:  r.Execution.SpansScored,
		ErrorsCount:  r.Execution.ErrorsCount,
		ErrorMessage: r.Execution.ErrorMessage,
		StartedAt:    r.Execution.StartedAt,
		CompletedAt:  r.Execution.CompletedAt,
		DurationMs:   r.Execution.DurationMs,
		Metadata:     r.Execution.Metadata,
		CreatedAt:    r.Execution.CreatedAt,
		Spans:        r.Spans,
		RuleSnapshot: r.RuleSnapshot,
	}
}
