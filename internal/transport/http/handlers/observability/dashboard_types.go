package observability

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"brokle/internal/core/domain/observability"
	"brokle/pkg/params"
)

// Dashboard-plane Huma operation types + trace/score/annotation DTOs +
// conversion helpers. All types here are consumed only by dashboard.go.

// ---- response DTOs + conversion helpers -----------------------------

// TraceScoreResponse is a DTO for score API responses.
// Metadata is json.RawMessage to preserve the stored JSON exactly as-is,
// avoiding lossy map[string]any conversion that silently drops non-object values.
type TraceScoreResponse struct {
	ID               uuid.UUID       `json:"id"`
	ProjectID        uuid.UUID       `json:"project_id"`
	TraceID          *string         `json:"trace_id,omitempty"`
	SpanID           *string         `json:"span_id,omitempty"`
	Name             string          `json:"name"`
	Value            *float64        `json:"value,omitempty"`
	StringValue      *string         `json:"string_value,omitempty"`
	Type             string          `json:"type"`
	Source           string          `json:"source"`
	Reason           *string         `json:"reason,omitempty"`
	Metadata         json.RawMessage `json:"metadata,omitempty"`
	ExperimentID     *uuid.UUID      `json:"experiment_id,omitempty"`
	ExperimentItemID *string         `json:"experiment_item_id,omitempty"`
	CreatedBy        *string         `json:"created_by,omitempty"`
	Timestamp        time.Time       `json:"timestamp"`
}

// toTraceScoreResponse converts a domain Score to the API DTO, preserving raw
// JSON metadata. Malformed legacy metadata bytes are escaped as a JSON
// string so the raw content is preserved losslessly rather than silently
// dropped or breaking the encoder.
func toTraceScoreResponse(s *observability.Score) *TraceScoreResponse {
	metadata := s.Metadata
	if len(metadata) > 0 && !json.Valid(metadata) {
		metadata, _ = json.Marshal(string(metadata))
	}

	return &TraceScoreResponse{
		ID:               s.ID,
		ProjectID:        s.ProjectID,
		TraceID:          s.TraceID,
		SpanID:           s.SpanID,
		Name:             s.Name,
		Value:            s.Value,
		StringValue:      s.StringValue,
		Type:             s.Type,
		Source:           s.Source,
		Reason:           s.Reason,
		Metadata:         metadata,
		ExperimentID:     s.ExperimentID,
		ExperimentItemID: s.ExperimentItemID,
		CreatedBy:        s.CreatedBy,
		Timestamp:        s.Timestamp,
	}
}

func toTraceScoreResponses(scores []*observability.Score) []*TraceScoreResponse {
	result := make([]*TraceScoreResponse, 0, len(scores))
	for _, s := range scores {
		result = append(result, toTraceScoreResponse(s))
	}
	return result
}

// AnnotationResponse represents a human annotation score returned from the API.
type AnnotationResponse struct {
	ID          uuid.UUID `json:"id"`
	ProjectID   uuid.UUID `json:"project_id"`
	TraceID     *string   `json:"trace_id,omitempty"`
	SpanID      *string   `json:"span_id,omitempty"`
	Name        string    `json:"name"`
	Value       *float64  `json:"value,omitempty"`
	StringValue *string   `json:"string_value,omitempty"`
	DataType    string    `json:"type"`
	Source      string    `json:"source"`
	Reason      *string   `json:"reason,omitempty"`
	CreatedBy   *string   `json:"created_by,omitempty"`
	Timestamp   string    `json:"timestamp"`
}

func toAnnotationResponse(s *observability.Score) *AnnotationResponse {
	return &AnnotationResponse{
		ID:          s.ID,
		ProjectID:   s.ProjectID,
		TraceID:     s.TraceID,
		SpanID:      s.SpanID,
		Name:        s.Name,
		Value:       s.Value,
		StringValue: s.StringValue,
		DataType:    s.Type,
		Source:      s.Source,
		Reason:      s.Reason,
		CreatedBy:   s.CreatedBy,
		Timestamp:   s.Timestamp.Format(time.RFC3339),
	}
}

// ---- pagination envelope --------------------------------------------

type paginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// ---- traces ---------------------------------------------------------

type ListTracesInput struct {
	SessionID    string                `query:"session_id" required:"false"`
	UserID       string                `query:"user_id" required:"false"`
	ServiceName  string                `query:"service_name" required:"false"`
	ModelName    string                `query:"model_name" required:"false"`
	ProviderName string                `query:"provider_name" required:"false"`
	MinCost      float64               `query:"min_cost" required:"false"`
	MaxCost      float64               `query:"max_cost" required:"false"`
	MinTokens    int64                 `query:"min_tokens" required:"false"`
	MaxTokens    int64                 `query:"max_tokens" required:"false"`
	MinDuration  int64                 `query:"min_duration" required:"false"`
	MaxDuration  int64                 `query:"max_duration" required:"false"`
	HasError     params.Optional[bool] `query:"has_error" doc:"Filter by error presence; omit for all"`
	Search       string                `query:"search" required:"false"`
	SearchType   string                `query:"search_type" required:"false"`
	Status       string                `query:"status" required:"false" doc:"Comma-separated status filter: ok,error,unset"`
	StatusNot    string                `query:"status_not" required:"false" doc:"Comma-separated status exclusion"`
	StartTime    int64                 `query:"start_time" required:"false" doc:"Unix timestamp seconds"`
	EndTime      int64                 `query:"end_time" required:"false" doc:"Unix timestamp seconds"`
	Page         int                   `query:"page" required:"false" minimum:"1"`
	Limit        int                   `query:"limit" required:"false"`
	SortBy       string                `query:"sort_by" required:"false"`
	SortDir      string                `query:"sort_dir" required:"false" enum:"asc,desc"`
}

type ListTracesOutput struct {
	Body listTracesResponse
}

type listTracesResponse struct {
	Data       []*observability.TraceSummary `json:"data"`
	Pagination paginationMeta                `json:"pagination"`
}

type GetTraceInput struct {
	ID string `path:"id" doc:"Trace ID (W3C hex)"`
}

type GetTraceOutput struct {
	Body *observability.TraceSummary
}

type DeleteTraceInput struct {
	ID string `path:"id"`
}

type DeleteTraceOutput struct{}

type GetTraceSpansInput struct {
	ID string `path:"id"`
}

type GetTraceSpansOutput struct {
	Body []*observability.Span
}

type GetTraceScoresInput struct {
	ID        string `path:"id"`
	ProjectID string `query:"project_id" required:"true" format:"uuid"`
}

type GetTraceScoresOutput struct {
	Body []*AnnotationResponse
}

type CreateAnnotationRequest struct {
	Name        string   `json:"name" minLength:"1"`
	Value       *float64 `json:"value,omitempty"`
	StringValue *string  `json:"string_value,omitempty"`
	DataType    string   `json:"type" enum:"NUMERIC,CATEGORICAL,BOOLEAN"`
	Reason      *string  `json:"reason,omitempty"`
}

type CreateTraceScoreInput struct {
	ID        string `path:"id"`
	ProjectID string `query:"project_id" required:"true" format:"uuid"`
	Body      CreateAnnotationRequest
}

type CreateTraceScoreOutput struct {
	Body *AnnotationResponse
}

type DeleteTraceScoreInput struct {
	ID        string `path:"id"`
	ScoreID   string `path:"scoreId" format:"uuid"`
	ProjectID string `query:"project_id" required:"true" format:"uuid"`
}

type DeleteTraceScoreOutput struct{}

type UpdateTraceTagsInput struct {
	ID        string `path:"id"`
	ProjectID string `query:"project_id" required:"true" format:"uuid"`
	Body      observability.UpdateTraceTagsRequest
}

type UpdateTraceTagsOutput struct {
	Body struct {
		Tags []string `json:"tags"`
	}
}

type UpdateTraceBookmarkInput struct {
	ID        string `path:"id"`
	ProjectID string `query:"project_id" required:"true" format:"uuid"`
	Body      struct {
		Bookmarked bool `json:"bookmarked"`
	}
}

type UpdateTraceBookmarkOutput struct {
	Body struct {
		Bookmarked bool `json:"bookmarked"`
	}
}

type GetTraceFilterOptionsInput struct {
	ProjectID string `query:"project_id" required:"true" format:"uuid"`
}

type GetTraceFilterOptionsOutput struct {
	Body *observability.TraceFilterOptions
}

type DiscoverAttributesInput struct {
	ProjectID string `query:"project_id" required:"true" format:"uuid"`
	Prefix    string `query:"prefix" required:"false"`
	Source    string `query:"source" required:"false" enum:"span_attributes,resource_attributes"`
	Limit     int    `query:"limit" required:"false" minimum:"1" maximum:"500"`
}

type DiscoverAttributesOutput struct {
	Body *observability.AttributeDiscoveryResponse
}

// ---- spans ----------------------------------------------------------

type ListSpansInput struct {
	TraceID string `query:"trace_id" required:"false"`
	Type    string `query:"type" required:"false"`
	Model   string `query:"model" required:"false"`
	Level   string `query:"level" required:"false"`
	Page    int    `query:"page" required:"false" minimum:"1"`
	Limit   int    `query:"limit" required:"false"`
	SortBy  string `query:"sort_by" required:"false"`
	SortDir string `query:"sort_dir" required:"false" enum:"asc,desc"`
}

type ListSpansOutput struct {
	Body listSpansResponse
}

type listSpansResponse struct {
	Data       []*observability.Span `json:"data"`
	Pagination paginationMeta        `json:"pagination"`
}

type GetSpanInput struct {
	ID string `path:"id" doc:"Span ID (OTEL 16-char hex)"`
}

type GetSpanOutput struct {
	Body *observability.Span
}

type DeleteSpanInput struct {
	ID string `path:"id"`
}

type DeleteSpanOutput struct{}

// ---- scores ---------------------------------------------------------

type scoreFilterQuery struct {
	TraceID string `query:"trace_id" required:"false"`
	SpanID  string `query:"span_id" required:"false"`
	Name    string `query:"name" required:"false"`
	Source  string `query:"source" required:"false"`
	Type    string `query:"type" required:"false"`
	Page    int    `query:"page" required:"false" minimum:"1"`
	Limit   int    `query:"limit" required:"false"`
	SortBy  string `query:"sort_by" required:"false"`
	SortDir string `query:"sort_dir" required:"false" enum:"asc,desc"`
}

type ListScoresInput struct {
	scoreFilterQuery
}

type ListScoresOutput struct {
	Body listScoresResponse
}

type listScoresResponse struct {
	Data       []*TraceScoreResponse `json:"data"`
	Pagination paginationMeta        `json:"pagination"`
}

type ListProjectScoresInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	scoreFilterQuery
}

type ListProjectScoresOutput struct {
	Body listScoresResponse
}

type GetScoreInput struct {
	ID string `path:"id" format:"uuid"`
}

type GetScoreOutput struct {
	Body *TraceScoreResponse
}

type UpdateScoreRequest struct {
	Name        string          `json:"name,omitempty"`
	Value       *float64        `json:"value,omitempty"`
	StringValue *string         `json:"string_value,omitempty"`
	Type        string          `json:"type,omitempty"`
	Source      string          `json:"source,omitempty"`
	Reason      *string         `json:"reason,omitempty"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
}

type UpdateScoreInput struct {
	ID   string `path:"id" format:"uuid"`
	Body UpdateScoreRequest
}

type UpdateScoreOutput struct {
	Body *TraceScoreResponse
}

type GetScoreAnalyticsInput struct {
	ProjectID        string `path:"projectId" format:"uuid"`
	ScoreName        string `query:"score_name" required:"true"`
	CompareScoreName string `query:"compare_score_name" required:"false"`
	FromTimestamp    string `query:"from_timestamp" required:"false" doc:"RFC3339 timestamp"`
	ToTimestamp      string `query:"to_timestamp" required:"false" doc:"RFC3339 timestamp"`
	Interval         string `query:"interval" required:"false" enum:"hour,day,week"`
}

type GetScoreAnalyticsOutput struct {
	Body *observability.ScoreAnalyticsResponse
}

type GetScoreNamesInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
}

type GetScoreNamesOutput struct {
	Body []string
}

// ---- trace sessions -------------------------------------------------

type ListTraceSessionsInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	Search    string `query:"search" required:"false"`
	UserID    string `query:"user_id" required:"false"`
	StartTime int64  `query:"start_time" required:"false" doc:"Unix timestamp seconds"`
	EndTime   int64  `query:"end_time" required:"false" doc:"Unix timestamp seconds"`
	Page      int    `query:"page" required:"false" minimum:"1"`
	Limit     int    `query:"limit" required:"false"`
	SortBy    string `query:"sort_by" required:"false"`
	SortDir   string `query:"sort_dir" required:"false" enum:"asc,desc"`
}

type ListTraceSessionsOutput struct {
	Body listTraceSessionsResponse
}

type listTraceSessionsResponse struct {
	Data       []*observability.TraceSessionSummary `json:"data"`
	Pagination paginationMeta                       `json:"pagination"`
}

// ---- filter presets -------------------------------------------------

type CreateFilterPresetInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	Body      observability.CreateFilterPresetRequest
}

type CreateFilterPresetOutput struct {
	Body *observability.FilterPreset
}

type ListFilterPresetsInput struct {
	ProjectID     string                `path:"projectId" format:"uuid"`
	TableName     string                `query:"table_name" required:"false" enum:"traces,spans"`
	IncludePublic params.Optional[bool] `query:"include_public" doc:"Include public presets (default: true)"`
}

type ListFilterPresetsOutput struct {
	Body []*observability.FilterPreset
}

type GetFilterPresetInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	ID        string `path:"id" format:"uuid"`
}

type GetFilterPresetOutput struct {
	Body *observability.FilterPreset
}

type UpdateFilterPresetInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	ID        string `path:"id" format:"uuid"`
	Body      observability.UpdateFilterPresetRequest
}

type UpdateFilterPresetOutput struct {
	Body *observability.FilterPreset
}

type DeleteFilterPresetInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	ID        string `path:"id" format:"uuid"`
}

type DeleteFilterPresetOutput struct{}
