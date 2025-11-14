package observability

import (
	"time"
)

// Trace request/response types

// CreateTraceRequest represents a request to create a new trace
type CreateTraceRequest struct {
	ProjectID       string                 `json:"project_id" binding:"required"`
	ExternalTraceID string                 `json:"external_trace_id" binding:"required"`
	Name            string                 `json:"name" binding:"required"`
	UserID          string                 `json:"user_id,omitempty"`
	SessionID       string                 `json:"session_id,omitempty"`
	ParentTraceID   string                 `json:"parent_trace_id,omitempty"`
	Tags            map[string]interface{} `json:"tags,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateTraceRequest represents a request to update an existing trace
type UpdateTraceRequest struct {
	Name     string                 `json:"name,omitempty"`
	UserID   string                 `json:"user_id,omitempty"`
	Tags     map[string]interface{} `json:"tags,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// TraceResponse represents a trace response
type TraceResponse struct {
	ID              string                 `json:"id"`
	ProjectID       string                 `json:"project_id"`
	ExternalTraceID string                 `json:"external_trace_id"`
	Name            string                 `json:"name"`
	UserID          string                 `json:"user_id,omitempty"`
	SessionID       string                 `json:"session_id,omitempty"`
	ParentTraceID   string                 `json:"parent_trace_id,omitempty"`
	Tags            map[string]interface{} `json:"tags,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// TraceWithSpansResponse represents a trace with its spans
type TraceWithSpansResponse struct {
	TraceResponse
	Spans []SpanResponse `json:"spans"`
}

// TraceStatsResponse represents trace statistics
type TraceStatsResponse struct {
	TraceID     string  `json:"trace_id"`
	TotalCost   float64 `json:"total_cost"`
	TotalTokens int     `json:"total_tokens"`
}

// ListTracesResponse represents a paginated list of traces
type ListTracesResponse struct {
	Traces []TraceResponse `json:"traces"`
	Total  int             `json:"total"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}

// BatchCreateTracesRequest represents a batch create request
type BatchCreateTracesRequest struct {
	Traces []CreateTraceRequest `json:"traces" binding:"required"`
}

// BatchCreateTracesResponse represents a batch create response
type BatchCreateTracesResponse struct {
	Traces         []TraceResponse `json:"traces"`
	ProcessedCount int             `json:"processed_count"`
}

// Span request/response types

// CreateSpanRequest represents a request to create a new span
type CreateSpanRequest struct {
	TraceID          string                 `json:"trace_id" binding:"required"`
	ExternalSpanID   string                 `json:"external_span_id" binding:"required"`
	ParentSpanID     string                 `json:"parent_span_id,omitempty"`
	Type             string                 `json:"type" binding:"required"`
	Name             string                 `json:"name" binding:"required"`
	StartTime        time.Time              `json:"start_time" binding:"required"`
	EndTime          *time.Time             `json:"end_time,omitempty"`
	Level            string                 `json:"level,omitempty"`
	StatusMessage    string                 `json:"status_message,omitempty"`
	Version          string                 `json:"version,omitempty"`
	Model            string                 `json:"model,omitempty"`
	Provider         string                 `json:"provider,omitempty"`
	Input            map[string]interface{} `json:"input,omitempty"`
	Output           map[string]interface{} `json:"output,omitempty"`
	ModelParameters  map[string]interface{} `json:"model_parameters,omitempty"`
	PromptTokens     int                    `json:"prompt_tokens,omitempty"`
	CompletionTokens int                    `json:"completion_tokens,omitempty"`
	TotalTokens      int                    `json:"total_tokens,omitempty"`
	InputCost        *float64               `json:"input_cost,omitempty"`
	OutputCost       *float64               `json:"output_cost,omitempty"`
	TotalCost        *float64               `json:"total_cost,omitempty"`
	LatencyMs        *int                   `json:"latency_ms,omitempty"`
	QualityScore     *float64               `json:"quality_score,omitempty"`
}

// UpdateSpanRequest represents a request to update an existing span
type UpdateSpanRequest struct {
	Name             string                 `json:"name,omitempty"`
	EndTime          *time.Time             `json:"end_time,omitempty"`
	Level            string                 `json:"level,omitempty"`
	StatusMessage    string                 `json:"status_message,omitempty"`
	Version          string                 `json:"version,omitempty"`
	Model            string                 `json:"model,omitempty"`
	Provider         string                 `json:"provider,omitempty"`
	Input            map[string]interface{} `json:"input,omitempty"`
	Output           map[string]interface{} `json:"output,omitempty"`
	ModelParameters  map[string]interface{} `json:"model_parameters,omitempty"`
	PromptTokens     int                    `json:"prompt_tokens,omitempty"`
	CompletionTokens int                    `json:"completion_tokens,omitempty"`
	TotalTokens      int                    `json:"total_tokens,omitempty"`
	InputCost        *float64               `json:"input_cost,omitempty"`
	OutputCost       *float64               `json:"output_cost,omitempty"`
	TotalCost        *float64               `json:"total_cost,omitempty"`
	LatencyMs        *int                   `json:"latency_ms,omitempty"`
	QualityScore     *float64               `json:"quality_score,omitempty"`
}

// CompleteSpanRequest represents a request to complete a span
type CompleteSpanRequest struct {
	EndTime          time.Time              `json:"end_time" binding:"required"`
	Output           map[string]interface{} `json:"output,omitempty"`
	StatusMessage    string                 `json:"status_message,omitempty"`
	PromptTokens     int                    `json:"prompt_tokens,omitempty"`
	CompletionTokens int                    `json:"completion_tokens,omitempty"`
	TotalTokens      int                    `json:"total_tokens,omitempty"`
	InputCost        *float64               `json:"input_cost,omitempty"`
	OutputCost       *float64               `json:"output_cost,omitempty"`
	TotalCost        *float64               `json:"total_cost,omitempty"`
	QualityScore     *float64               `json:"quality_score,omitempty"`
}

// SpanResponse represents an span response
type SpanResponse struct {
	ID               string                 `json:"id"`
	TraceID          string                 `json:"trace_id"`
	ExternalSpanID   string                 `json:"external_span_id"`
	ParentSpanID     string                 `json:"parent_span_id,omitempty"`
	Type             string                 `json:"type"`
	Name             string                 `json:"name"`
	StartTime        time.Time              `json:"start_time"`
	EndTime          *time.Time             `json:"end_time,omitempty"`
	Level            string                 `json:"level"`
	StatusMessage    string                 `json:"status_message,omitempty"`
	Version          string                 `json:"version,omitempty"`
	Model            string                 `json:"model,omitempty"`
	Provider         string                 `json:"provider,omitempty"`
	Input            map[string]interface{} `json:"input,omitempty"`
	Output           map[string]interface{} `json:"output,omitempty"`
	ModelParameters  map[string]interface{} `json:"model_parameters,omitempty"`
	PromptTokens     int                    `json:"prompt_tokens"`
	CompletionTokens int                    `json:"completion_tokens"`
	TotalTokens      int                    `json:"total_tokens"`
	InputCost        *float64               `json:"input_cost,omitempty"`
	OutputCost       *float64               `json:"output_cost,omitempty"`
	TotalCost        *float64               `json:"total_cost,omitempty"`
	LatencyMs        *int                   `json:"latency_ms,omitempty"`
	QualityScore     *float64               `json:"quality_score,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// ListSpansResponse represents a paginated list of spans
type ListSpansResponse struct {
	Spans  []SpanResponse `json:"spans"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

// BatchCreateSpansRequest represents a batch create request
type BatchCreateSpansRequest struct {
	Spans []CreateSpanRequest `json:"spans" binding:"required"`
}

// BatchCreateSpansResponse represents a batch create response
type BatchCreateSpansResponse struct {
	Spans          []SpanResponse `json:"spans"`
	ProcessedCount int            `json:"processed_count"`
}

// Quality Score request/response types

// CreateQualityScoreRequest represents a request to create a new quality score
type CreateQualityScoreRequest struct {
	TraceID          string   `json:"trace_id" binding:"required"`
	SpanID           string   `json:"span_id,omitempty"`
	ScoreName        string   `json:"score_name" binding:"required"`
	ScoreValue       *float64 `json:"score_value,omitempty"`
	StringValue      *string  `json:"string_value,omitempty"`
	DataType         string   `json:"data_type" binding:"required"`
	Source           string   `json:"source,omitempty"`
	EvaluatorName    string   `json:"evaluator_name,omitempty"`
	EvaluatorVersion string   `json:"evaluator_version,omitempty"`
	Comment          string   `json:"comment,omitempty"`
	AuthorUserID     string   `json:"author_user_id,omitempty"`
}

// UpdateQualityScoreRequest represents a request to update an existing quality score
type UpdateQualityScoreRequest struct {
	ScoreValue       *float64 `json:"score_value,omitempty"`
	StringValue      *string  `json:"string_value,omitempty"`
	Comment          string   `json:"comment,omitempty"`
	EvaluatorName    string   `json:"evaluator_name,omitempty"`
	EvaluatorVersion string   `json:"evaluator_version,omitempty"`
}

// QualityScoreResponse represents a quality score response
type QualityScoreResponse struct {
	ID               string    `json:"id"`
	TraceID          string    `json:"trace_id"`
	SpanID           string    `json:"span_id,omitempty"`
	ScoreName        string    `json:"score_name"`
	ScoreValue       *float64  `json:"score_value,omitempty"`
	StringValue      *string   `json:"string_value,omitempty"`
	DataType         string    `json:"data_type"`
	Source           string    `json:"source"`
	EvaluatorName    string    `json:"evaluator_name,omitempty"`
	EvaluatorVersion string    `json:"evaluator_version,omitempty"`
	Comment          string    `json:"comment,omitempty"`
	AuthorUserID     string    `json:"author_user_id,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ListQualityScoresResponse represents a paginated list of quality scores
type ListQualityScoresResponse struct {
	QualityScores []QualityScoreResponse `json:"quality_scores"`
	Total         int                    `json:"total"`
	Limit         int                    `json:"limit"`
	Offset        int                    `json:"offset"`
}

// EvaluateRequest represents a request to evaluate traces or spans
type EvaluateRequest struct {
	TraceIDs      []string `json:"trace_ids,omitempty"`
	SpanIDs       []string `json:"span_ids,omitempty"`
	EvaluatorName string   `json:"evaluator_name" binding:"required"`
}

// EvaluateResponse represents an evaluation response
type EvaluateResponse struct {
	QualityScores  []QualityScoreResponse `json:"quality_scores"`
	ProcessedCount int                    `json:"processed_count"`
	FailedCount    int                    `json:"failed_count"`
	Errors         []EvaluationError      `json:"errors,omitempty"`
}

// EvaluationError represents an error during evaluation
type EvaluationError struct {
	ItemID  string `json:"item_id"`
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

// Analytics types

// AnalyticsFilter represents filters for analytics queries
type AnalyticsFilter struct {
	ProjectID string     `json:"project_id,omitempty"`
	UserID    string     `json:"user_id,omitempty"`
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Provider  string     `json:"provider,omitempty"`
	Model     string     `json:"model,omitempty"`
	SpanType  string     `json:"span_type,omitempty"`
	Tags      []string   `json:"tags,omitempty"`
}

// DashboardOverviewResponse represents dashboard overview metrics
type DashboardOverviewResponse struct {
	TotalTraces    int64                     `json:"total_traces"`
	TotalCost      float64                   `json:"total_cost"`
	AverageLatency float64                   `json:"average_latency"`
	ErrorRate      float64                   `json:"error_rate"`
	TopProviders   []ProviderSummaryResponse `json:"top_providers"`
	RecentActivity []ActivityItemResponse    `json:"recent_activity"`
	CostTrend      []TimeSeriesPointResponse `json:"cost_trend"`
	LatencyTrend   []TimeSeriesPointResponse `json:"latency_trend"`
	QualityTrend   []TimeSeriesPointResponse `json:"quality_trend"`
}

// ProviderSummaryResponse represents provider summary information
type ProviderSummaryResponse struct {
	Provider       string  `json:"provider"`
	RequestCount   int64   `json:"request_count"`
	TotalCost      float64 `json:"total_cost"`
	AverageLatency float64 `json:"average_latency"`
	ErrorRate      float64 `json:"error_rate"`
}

// ActivityItemResponse represents a recent activity item
type ActivityItemResponse struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// TimeSeriesPointResponse represents a time series data point
type TimeSeriesPointResponse struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// Error response types

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error   string      `json:"error"`
	Message string      `json:"message"`
	Code    string      `json:"code,omitempty"`
	Details interface{} `json:"details,omitempty"`
}

// ValidationErrorResponse represents validation error details
type ValidationErrorResponse struct {
	Error       string            `json:"error"`
	Message     string            `json:"message"`
	FieldErrors map[string]string `json:"field_errors,omitempty"`
}

// Utility types for pagination

// PaginationInfo represents pagination information
type PaginationInfo struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

// SortInfo represents sorting information
type SortInfo struct {
	SortBy    string `json:"sort_by"`
	SortOrder string `json:"sort_order"`
}
