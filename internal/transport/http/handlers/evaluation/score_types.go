package evaluation

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Huma operation types for the SDK-plane score ingestion feature.

type CreateScoreRequest struct {
	TraceID          *string        `json:"trace_id,omitempty"`
	SpanID           *string        `json:"span_id,omitempty"`
	Name             string         `json:"name"`
	Value            *float64       `json:"value,omitempty"`
	StringValue      *string        `json:"string_value,omitempty"`
	Type             string         `json:"type" enum:"NUMERIC,CATEGORICAL,BOOLEAN"`
	Reason           *string        `json:"reason,omitempty"`
	Metadata         map[string]any `json:"metadata,omitempty"`
	ExperimentID     *uuid.UUID     `json:"experiment_id,omitempty"`
	ExperimentItemID *string        `json:"experiment_item_id,omitempty"`
}

type BatchScoreRequest struct {
	Scores []CreateScoreRequest `json:"scores"`
}

// SubmittedScoreResponse exposes the observability.Score DTO. Metadata is a
// json.RawMessage on the domain entity — we sanitize it for safe JSON
// marshaling in toSubmittedScoreResponse.
type SubmittedScoreResponse struct {
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
	Timestamp        time.Time       `json:"timestamp"`
}

type BatchScoreResponse struct {
	Created int `json:"created"`
}

type SDKCreateScoreInput struct {
	Body CreateScoreRequest
}

type SDKCreateScoreOutput struct {
	Body *SubmittedScoreResponse
}

type SDKCreateScoreBatchInput struct {
	Body BatchScoreRequest
}

type SDKCreateScoreBatchOutput struct {
	Body *BatchScoreResponse
}
