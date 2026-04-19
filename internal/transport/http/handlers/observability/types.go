package observability

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"brokle/internal/core/domain/observability"
)

// ScoreResponse is a DTO for score API responses.
// Metadata is json.RawMessage to preserve the stored JSON exactly as-is,
// avoiding lossy map[string]any conversion that silently drops non-object values.
type ScoreResponse struct {
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

// toScoreResponse converts a domain Score to the API DTO, preserving raw
// JSON metadata. Malformed legacy metadata bytes are escaped as a JSON
// string so the raw content is preserved losslessly rather than silently
// dropped or breaking the encoder.
func toScoreResponse(s *observability.Score) *ScoreResponse {
	metadata := s.Metadata
	if len(metadata) > 0 && !json.Valid(metadata) {
		metadata, _ = json.Marshal(string(metadata))
	}

	return &ScoreResponse{
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

func toScoreResponses(scores []*observability.Score) []*ScoreResponse {
	result := make([]*ScoreResponse, 0, len(scores))
	for _, s := range scores {
		result = append(result, toScoreResponse(s))
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
