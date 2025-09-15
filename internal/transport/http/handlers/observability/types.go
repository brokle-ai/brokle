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

// TraceWithObservationsResponse represents a trace with its observations
type TraceWithObservationsResponse struct {
	TraceResponse
	Observations []ObservationResponse `json:"observations"`
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

// Observation request/response types

// CreateObservationRequest represents a request to create a new observation
type CreateObservationRequest struct {
	TraceID                string                 `json:"trace_id" binding:"required"`
	ExternalObservationID  string                 `json:"external_observation_id" binding:"required"`
	ParentObservationID    string                 `json:"parent_observation_id,omitempty"`
	Type                   string                 `json:"type" binding:"required"`
	Name                   string                 `json:"name" binding:"required"`
	StartTime              time.Time              `json:"start_time" binding:"required"`
	EndTime                *time.Time             `json:"end_time,omitempty"`
	Level                  string                 `json:"level,omitempty"`
	StatusMessage          string                 `json:"status_message,omitempty"`
	Version                string                 `json:"version,omitempty"`
	Model                  string                 `json:"model,omitempty"`
	Provider               string                 `json:"provider,omitempty"`
	Input                  map[string]interface{} `json:"input,omitempty"`
	Output                 map[string]interface{} `json:"output,omitempty"`
	ModelParameters        map[string]interface{} `json:"model_parameters,omitempty"`
	PromptTokens           int                    `json:"prompt_tokens,omitempty"`
	CompletionTokens       int                    `json:"completion_tokens,omitempty"`
	TotalTokens            int                    `json:"total_tokens,omitempty"`
	InputCost              *float64               `json:"input_cost,omitempty"`
	OutputCost             *float64               `json:"output_cost,omitempty"`
	TotalCost              *float64               `json:"total_cost,omitempty"`
	LatencyMs              *int                   `json:"latency_ms,omitempty"`
	QualityScore           *float64               `json:"quality_score,omitempty"`
}

// UpdateObservationRequest represents a request to update an existing observation
type UpdateObservationRequest struct {
	Name                   string                 `json:"name,omitempty"`
	EndTime                *time.Time             `json:"end_time,omitempty"`
	Level                  string                 `json:"level,omitempty"`
	StatusMessage          string                 `json:"status_message,omitempty"`
	Version                string                 `json:"version,omitempty"`
	Model                  string                 `json:"model,omitempty"`
	Provider               string                 `json:"provider,omitempty"`
	Input                  map[string]interface{} `json:"input,omitempty"`
	Output                 map[string]interface{} `json:"output,omitempty"`
	ModelParameters        map[string]interface{} `json:"model_parameters,omitempty"`
	PromptTokens           int                    `json:"prompt_tokens,omitempty"`
	CompletionTokens       int                    `json:"completion_tokens,omitempty"`
	TotalTokens            int                    `json:"total_tokens,omitempty"`
	InputCost              *float64               `json:"input_cost,omitempty"`
	OutputCost             *float64               `json:"output_cost,omitempty"`
	TotalCost              *float64               `json:"total_cost,omitempty"`
	LatencyMs              *int                   `json:"latency_ms,omitempty"`
	QualityScore           *float64               `json:"quality_score,omitempty"`
}

// CompleteObservationRequest represents a request to complete an observation
type CompleteObservationRequest struct {
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

// ObservationResponse represents an observation response
type ObservationResponse struct {
	ID                     string                 `json:"id"`
	TraceID               string                 `json:"trace_id"`
	ExternalObservationID string                 `json:"external_observation_id"`
	ParentObservationID   string                 `json:"parent_observation_id,omitempty"`
	Type                  string                 `json:"type"`
	Name                  string                 `json:"name"`
	StartTime             time.Time              `json:"start_time"`
	EndTime               *time.Time             `json:"end_time,omitempty"`
	Level                 string                 `json:"level"`
	StatusMessage         string                 `json:"status_message,omitempty"`
	Version               string                 `json:"version,omitempty"`
	Model                 string                 `json:"model,omitempty"`
	Provider              string                 `json:"provider,omitempty"`
	Input                 map[string]interface{} `json:"input,omitempty"`
	Output                map[string]interface{} `json:"output,omitempty"`
	ModelParameters       map[string]interface{} `json:"model_parameters,omitempty"`
	PromptTokens          int                    `json:"prompt_tokens"`
	CompletionTokens      int                    `json:"completion_tokens"`
	TotalTokens           int                    `json:"total_tokens"`
	InputCost             *float64               `json:"input_cost,omitempty"`
	OutputCost            *float64               `json:"output_cost,omitempty"`
	TotalCost             *float64               `json:"total_cost,omitempty"`
	LatencyMs             *int                   `json:"latency_ms,omitempty"`
	QualityScore          *float64               `json:"quality_score,omitempty"`
	CreatedAt             time.Time              `json:"created_at"`
	UpdatedAt             time.Time              `json:"updated_at"`
}

// ListObservationsResponse represents a paginated list of observations
type ListObservationsResponse struct {
	Observations []ObservationResponse `json:"observations"`
	Total        int                   `json:"total"`
	Limit        int                   `json:"limit"`
	Offset       int                   `json:"offset"`
}

// BatchCreateObservationsRequest represents a batch create request
type BatchCreateObservationsRequest struct {
	Observations []CreateObservationRequest `json:"observations" binding:"required"`
}

// BatchCreateObservationsResponse represents a batch create response
type BatchCreateObservationsResponse struct {
	Observations   []ObservationResponse `json:"observations"`
	ProcessedCount int                   `json:"processed_count"`
}

// Quality Score request/response types

// CreateQualityScoreRequest represents a request to create a new quality score
type CreateQualityScoreRequest struct {
	TraceID           string   `json:"trace_id" binding:"required"`
	ObservationID     string   `json:"observation_id,omitempty"`
	ScoreName         string   `json:"score_name" binding:"required"`
	ScoreValue        *float64 `json:"score_value,omitempty"`
	StringValue       *string  `json:"string_value,omitempty"`
	DataType          string   `json:"data_type" binding:"required"`
	Source            string   `json:"source,omitempty"`
	EvaluatorName     string   `json:"evaluator_name,omitempty"`
	EvaluatorVersion  string   `json:"evaluator_version,omitempty"`
	Comment           string   `json:"comment,omitempty"`
	AuthorUserID      string   `json:"author_user_id,omitempty"`
}

// UpdateQualityScoreRequest represents a request to update an existing quality score
type UpdateQualityScoreRequest struct {
	ScoreValue        *float64 `json:"score_value,omitempty"`
	StringValue       *string  `json:"string_value,omitempty"`
	Comment           string   `json:"comment,omitempty"`
	EvaluatorName     string   `json:"evaluator_name,omitempty"`
	EvaluatorVersion  string   `json:"evaluator_version,omitempty"`
}

// QualityScoreResponse represents a quality score response
type QualityScoreResponse struct {
	ID                string    `json:"id"`
	TraceID          string    `json:"trace_id"`
	ObservationID    string    `json:"observation_id,omitempty"`
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

// EvaluateRequest represents a request to evaluate traces or observations
type EvaluateRequest struct {
	TraceIDs       []string `json:"trace_ids,omitempty"`
	ObservationIDs []string `json:"observation_ids,omitempty"`
	EvaluatorName  string   `json:"evaluator_name" binding:"required"`
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
	ProjectID     string     `json:"project_id,omitempty"`
	UserID        string     `json:"user_id,omitempty"`
	StartTime     *time.Time `json:"start_time,omitempty"`
	EndTime       *time.Time `json:"end_time,omitempty"`
	Provider      string     `json:"provider,omitempty"`
	Model         string     `json:"model,omitempty"`
	ObsType       string     `json:"observation_type,omitempty"`
	Tags          []string   `json:"tags,omitempty"`
}

// DashboardOverviewResponse represents dashboard overview metrics
type DashboardOverviewResponse struct {
	TotalTraces     int64                     `json:"total_traces"`
	TotalCost       float64                   `json:"total_cost"`
	AverageLatency  float64                   `json:"average_latency"`
	ErrorRate       float64                   `json:"error_rate"`
	TopProviders    []ProviderSummaryResponse `json:"top_providers"`
	RecentActivity  []ActivityItemResponse    `json:"recent_activity"`
	CostTrend       []TimeSeriesPointResponse `json:"cost_trend"`
	LatencyTrend    []TimeSeriesPointResponse `json:"latency_trend"`
	QualityTrend    []TimeSeriesPointResponse `json:"quality_trend"`
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
	Error        string            `json:"error"`
	Message      string            `json:"message"`
	FieldErrors  map[string]string `json:"field_errors,omitempty"`
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