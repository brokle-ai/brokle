package observability

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"brokle/internal/core/domain/observability"
)

// Dashboard-plane response DTOs + conversion helpers.

// TraceScoreResponse is a DTO for score API responses. Metadata is
// json.RawMessage to preserve the stored JSON exactly as-is, avoiding
// lossy map[string]any conversion that silently drops non-object values.
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

// AnnotationResponse represents a human annotation score.
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

// ---- list response wrappers ----------------------------------------

type listTracesResponse struct {
	Data       []*observability.TraceSummary `json:"data"`
	Pagination paginationMeta                `json:"pagination"`
}

type listSpansResponse struct {
	Data       []*observability.Span `json:"data"`
	Pagination paginationMeta        `json:"pagination"`
}

type listScoresResponse struct {
	Data       []*TraceScoreResponse `json:"data"`
	Pagination paginationMeta        `json:"pagination"`
}

type listTraceSessionsResponse struct {
	Data       []*observability.TraceSessionSummary `json:"data"`
	Pagination paginationMeta                       `json:"pagination"`
}

// ---- request bodies -------------------------------------------------

type CreateAnnotationRequest struct {
	Name        string   `json:"name"                     validate:"required,min=1"`
	Value       *float64 `json:"value,omitempty"`
	StringValue *string  `json:"string_value,omitempty"`
	DataType    string   `json:"type"                     validate:"required,oneof=NUMERIC CATEGORICAL BOOLEAN"`
	Reason      *string  `json:"reason,omitempty"`
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
