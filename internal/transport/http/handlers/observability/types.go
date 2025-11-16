package observability

import (
	"time"
)

// Trace request/response types

// CreateTraceRequest represents a request to create a new trace
type CreateTraceRequest struct {
	Tags            map[string]interface{} `json:"tags,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	ProjectID       string                 `json:"project_id" binding:"required"`
	ExternalTraceID string                 `json:"external_trace_id" binding:"required"`
	Name            string                 `json:"name" binding:"required"`
	UserID          string                 `json:"user_id,omitempty"`
	SessionID       string                 `json:"session_id,omitempty"`
	ParentTraceID   string                 `json:"parent_trace_id,omitempty"`
}

// UpdateTraceRequest represents a request to update an existing trace
type UpdateTraceRequest struct {
	Tags     map[string]interface{} `json:"tags,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Name     string                 `json:"name,omitempty"`
	UserID   string                 `json:"user_id,omitempty"`
}

// TraceResponse represents a trace response
type TraceResponse struct {
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	Tags            map[string]interface{} `json:"tags,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	ID              string                 `json:"id"`
	ProjectID       string                 `json:"project_id"`
	ExternalTraceID string                 `json:"external_trace_id"`
	Name            string                 `json:"name"`
	UserID          string                 `json:"user_id,omitempty"`
	SessionID       string                 `json:"session_id,omitempty"`
	ParentTraceID   string                 `json:"parent_trace_id,omitempty"`
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
	StartTime        time.Time              `json:"start_time" binding:"required"`
	Input            map[string]interface{} `json:"input,omitempty"`
	QualityScore     *float64               `json:"quality_score,omitempty"`
	LatencyMs        *int                   `json:"latency_ms,omitempty"`
	TotalCost        *float64               `json:"total_cost,omitempty"`
	OutputCost       *float64               `json:"output_cost,omitempty"`
	EndTime          *time.Time             `json:"end_time,omitempty"`
	InputCost        *float64               `json:"input_cost,omitempty"`
	ModelParameters  map[string]interface{} `json:"model_parameters,omitempty"`
	Output           map[string]interface{} `json:"output,omitempty"`
	Model            string                 `json:"model,omitempty"`
	Provider         string                 `json:"provider,omitempty"`
	TraceID          string                 `json:"trace_id" binding:"required"`
	Version          string                 `json:"version,omitempty"`
	StatusMessage    string                 `json:"status_message,omitempty"`
	Level            string                 `json:"level,omitempty"`
	Name             string                 `json:"name" binding:"required"`
	Type             string                 `json:"type" binding:"required"`
	ParentSpanID     string                 `json:"parent_span_id,omitempty"`
	ExternalSpanID   string                 `json:"external_span_id" binding:"required"`
	PromptTokens     int                    `json:"prompt_tokens,omitempty"`
	CompletionTokens int                    `json:"completion_tokens,omitempty"`
	TotalTokens      int                    `json:"total_tokens,omitempty"`
}

// UpdateSpanRequest represents a request to update an existing span
type UpdateSpanRequest struct {
	ModelParameters  map[string]interface{} `json:"model_parameters,omitempty"`
	EndTime          *time.Time             `json:"end_time,omitempty"`
	QualityScore     *float64               `json:"quality_score,omitempty"`
	LatencyMs        *int                   `json:"latency_ms,omitempty"`
	TotalCost        *float64               `json:"total_cost,omitempty"`
	OutputCost       *float64               `json:"output_cost,omitempty"`
	InputCost        *float64               `json:"input_cost,omitempty"`
	Input            map[string]interface{} `json:"input,omitempty"`
	Output           map[string]interface{} `json:"output,omitempty"`
	Version          string                 `json:"version,omitempty"`
	Provider         string                 `json:"provider,omitempty"`
	Model            string                 `json:"model,omitempty"`
	Name             string                 `json:"name,omitempty"`
	StatusMessage    string                 `json:"status_message,omitempty"`
	Level            string                 `json:"level,omitempty"`
	PromptTokens     int                    `json:"prompt_tokens,omitempty"`
	CompletionTokens int                    `json:"completion_tokens,omitempty"`
	TotalTokens      int                    `json:"total_tokens,omitempty"`
}

// CompleteSpanRequest represents a request to complete a span
type CompleteSpanRequest struct {
	EndTime          time.Time              `json:"end_time" binding:"required"`
	Output           map[string]interface{} `json:"output,omitempty"`
	InputCost        *float64               `json:"input_cost,omitempty"`
	OutputCost       *float64               `json:"output_cost,omitempty"`
	TotalCost        *float64               `json:"total_cost,omitempty"`
	QualityScore     *float64               `json:"quality_score,omitempty"`
	StatusMessage    string                 `json:"status_message,omitempty"`
	PromptTokens     int                    `json:"prompt_tokens,omitempty"`
	CompletionTokens int                    `json:"completion_tokens,omitempty"`
	TotalTokens      int                    `json:"total_tokens,omitempty"`
}

// SpanResponse represents an span response
type SpanResponse struct {
	UpdatedAt        time.Time              `json:"updated_at"`
	CreatedAt        time.Time              `json:"created_at"`
	StartTime        time.Time              `json:"start_time"`
	Input            map[string]interface{} `json:"input,omitempty"`
	QualityScore     *float64               `json:"quality_score,omitempty"`
	LatencyMs        *int                   `json:"latency_ms,omitempty"`
	TotalCost        *float64               `json:"total_cost,omitempty"`
	EndTime          *time.Time             `json:"end_time,omitempty"`
	OutputCost       *float64               `json:"output_cost,omitempty"`
	InputCost        *float64               `json:"input_cost,omitempty"`
	ModelParameters  map[string]interface{} `json:"model_parameters,omitempty"`
	Output           map[string]interface{} `json:"output,omitempty"`
	Provider         string                 `json:"provider,omitempty"`
	Type             string                 `json:"type"`
	Model            string                 `json:"model,omitempty"`
	Version          string                 `json:"version,omitempty"`
	TraceID          string                 `json:"trace_id"`
	ExternalSpanID   string                 `json:"external_span_id"`
	ParentSpanID     string                 `json:"parent_span_id,omitempty"`
	StatusMessage    string                 `json:"status_message,omitempty"`
	Level            string                 `json:"level"`
	Name             string                 `json:"name"`
	ID               string                 `json:"id"`
	TotalTokens      int                    `json:"total_tokens"`
	CompletionTokens int                    `json:"completion_tokens"`
	PromptTokens     int                    `json:"prompt_tokens"`
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
	UpdatedAt        time.Time `json:"updated_at"`
	CreatedAt        time.Time `json:"created_at"`
	ScoreValue       *float64  `json:"score_value,omitempty"`
	StringValue      *string   `json:"string_value,omitempty"`
	DataType         string    `json:"data_type"`
	ScoreName        string    `json:"score_name"`
	ID               string    `json:"id"`
	Source           string    `json:"source"`
	EvaluatorName    string    `json:"evaluator_name,omitempty"`
	EvaluatorVersion string    `json:"evaluator_version,omitempty"`
	Comment          string    `json:"comment,omitempty"`
	AuthorUserID     string    `json:"author_user_id,omitempty"`
	SpanID           string    `json:"span_id,omitempty"`
	TraceID          string    `json:"trace_id"`
}

// EvaluateRequest represents a request to evaluate traces or spans
type EvaluateRequest struct {
	EvaluatorName string   `json:"evaluator_name" binding:"required"`
	TraceIDs      []string `json:"trace_ids,omitempty"`
	SpanIDs       []string `json:"span_ids,omitempty"`
}

// EvaluateResponse represents an evaluation response
type EvaluateResponse struct {
	QualityScores  []QualityScoreResponse `json:"quality_scores"`
	Errors         []EvaluationError      `json:"errors,omitempty"`
	ProcessedCount int                    `json:"processed_count"`
	FailedCount    int                    `json:"failed_count"`
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
	TopProviders   []ProviderSummaryResponse `json:"top_providers"`
	RecentActivity []ActivityItemResponse    `json:"recent_activity"`
	CostTrend      []TimeSeriesPointResponse `json:"cost_trend"`
	LatencyTrend   []TimeSeriesPointResponse `json:"latency_trend"`
	QualityTrend   []TimeSeriesPointResponse `json:"quality_trend"`
	TotalTraces    int64                     `json:"total_traces"`
	TotalCost      float64                   `json:"total_cost"`
	AverageLatency float64                   `json:"average_latency"`
	ErrorRate      float64                   `json:"error_rate"`
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
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
}

// TimeSeriesPointResponse represents a time series data point
type TimeSeriesPointResponse struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// Error response types

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Details interface{} `json:"details,omitempty"`
	Error   string      `json:"error"`
	Message string      `json:"message"`
	Code    string      `json:"code,omitempty"`
}

// ValidationErrorResponse represents validation error details
type ValidationErrorResponse struct {
	FieldErrors map[string]string `json:"field_errors,omitempty"`
	Error       string            `json:"error"`
	Message     string            `json:"message"`
}

// Utility types for pagination

// SortInfo represents sorting information
type SortInfo struct {
	SortBy    string `json:"sort_by"`
	SortOrder string `json:"sort_order"`
}
