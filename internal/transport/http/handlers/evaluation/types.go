package evaluation

import (
	"time"

	"github.com/google/uuid"

	evaluationDomain "brokle/internal/core/domain/evaluation"
)

// Package-level shared DTOs — types used by two or more feature files
// (dataset.go, experiment.go, evaluator.go, execution.go, wizard.go,
// handlers.go). Feature-scoped types live in <feature>_types.go.

// pageList is the canonical paginated list envelope for evaluation responses.
type pageList[T any] struct {
	Data  []T   `json:"data"`
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
}

// ScoreType is the score-config enum carried in request bodies.
type ScoreType string

const (
	ScoreTypeNumeric     ScoreType = "NUMERIC"
	ScoreTypeCategorical ScoreType = "CATEGORICAL"
	ScoreTypeBoolean     ScoreType = "BOOLEAN"
)

// DatasetItemResponse is the serialisation shape used by all item-returning
// endpoints. Identical to the underlying evaluationDomain.DatasetItemResponse
// but with `source` pre-stringified for stable JSON regardless of future
// enum changes.
type DatasetItemResponse struct {
	ID            uuid.UUID      `json:"id"`
	DatasetID     uuid.UUID      `json:"dataset_id"`
	Input         map[string]any `json:"input"`
	Expected      map[string]any `json:"expected,omitempty"`
	Metadata      map[string]any `json:"metadata,omitempty"`
	Source        string         `json:"source"`
	SourceTraceID *string        `json:"source_trace_id,omitempty"`
	SourceSpanID  *string        `json:"source_span_id,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
}

func toDatasetItemResponse(item *evaluationDomain.DatasetItem) *DatasetItemResponse {
	r := item.ToResponse()
	return &DatasetItemResponse{
		ID:            r.ID,
		DatasetID:     r.DatasetID,
		Input:         r.Input,
		Expected:      r.Expected,
		Metadata:      r.Metadata,
		Source:        string(r.Source),
		SourceTraceID: r.SourceTraceID,
		SourceSpanID:  r.SourceSpanID,
		CreatedAt:     r.CreatedAt,
	}
}

// ExperimentItemResponse mirrors evaluationDomain.ExperimentItemResponse
// so the handler can keep a single shape across endpoints.
type ExperimentItemResponse struct {
	ID            uuid.UUID      `json:"id"`
	ExperimentID  uuid.UUID      `json:"experiment_id"`
	DatasetItemID *uuid.UUID     `json:"dataset_item_id,omitempty"`
	TraceID       *string        `json:"trace_id,omitempty"`
	Input         map[string]any `json:"input"`
	Output        any            `json:"output,omitempty"`
	Expected      any            `json:"expected,omitempty"`
	TrialNumber   int            `json:"trial_number"`
	Metadata      map[string]any `json:"metadata,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
}

func toExperimentItemResponse(item *evaluationDomain.ExperimentItem) *ExperimentItemResponse {
	r := item.ToResponse()
	return &ExperimentItemResponse{
		ID:            r.ID,
		ExperimentID:  r.ExperimentID,
		DatasetItemID: r.DatasetItemID,
		TraceID:       r.TraceID,
		Input:         r.Input,
		Output:        r.Output,
		Expected:      r.Expected,
		TrialNumber:   r.TrialNumber,
		Metadata:      r.Metadata,
		CreatedAt:     r.CreatedAt,
	}
}

// BulkImportResponse is the result of a bulk dataset-item import.
type BulkImportResponse struct {
	Created int      `json:"created"`
	Skipped int      `json:"skipped"`
	Errors  []string `json:"errors,omitempty"`
}

func toBulkImportResponse(r *evaluationDomain.BulkImportResult) *BulkImportResponse {
	return &BulkImportResponse{
		Created: r.Created,
		Skipped: r.Skipped,
		Errors:  r.Errors,
	}
}

// KeysMappingRequest is the optional field-extraction mapping shared by
// import endpoints.
type KeysMappingRequest struct {
	InputKeys    []string `json:"input_keys,omitempty"`
	ExpectedKeys []string `json:"expected_keys,omitempty"`
	MetadataKeys []string `json:"metadata_keys,omitempty"`
}

func (k *KeysMappingRequest) toDomain() *evaluationDomain.KeysMapping {
	if k == nil {
		return nil
	}
	return &evaluationDomain.KeysMapping{
		InputKeys:    k.InputKeys,
		ExpectedKeys: k.ExpectedKeys,
		MetadataKeys: k.MetadataKeys,
	}
}

// CountResponse is returned by SDK batch creates (items + experiment items).
type CountResponse struct {
	Created int `json:"created"`
}

// EmptyOutput is the Huma output wrapper for 204-style empty responses.
type EmptyOutput struct{}
