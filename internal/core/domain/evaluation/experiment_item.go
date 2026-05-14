// Package evaluation provides domain entities for quality scoring, datasets, and experiments.
package evaluation

import (
	"time"

	"github.com/google/uuid"

	"brokle/pkg/uid"
)

type ExperimentItem struct {
	ID            uuid.UUID              `json:"id"`
	ExperimentID  uuid.UUID              `json:"experiment_id"`
	DatasetItemID *uuid.UUID             `json:"dataset_item_id,omitempty"`
	TraceID       *string                `json:"trace_id,omitempty"`
	Input         map[string]any `json:"input"`
	Output        any            `json:"output,omitempty"`
	Expected      any            `json:"expected,omitempty"`
	TrialNumber   int                    `json:"trial_number"`
	Metadata      map[string]any `json:"metadata,omitempty"`
	Error         *string                `json:"error,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
}

func NewExperimentItem(experimentID uuid.UUID, input map[string]any) *ExperimentItem {
	return &ExperimentItem{
		ID:           uid.New(),
		ExperimentID: experimentID,
		Input:        input,
		TrialNumber:  1,
		Metadata:     make(map[string]any),
		CreatedAt:    time.Now(),
	}
}

func (ei *ExperimentItem) Validate() []EvaluationValidationError {
	var errors []EvaluationValidationError

	if ei.Input == nil || len(ei.Input) == 0 {
		errors = append(errors, EvaluationValidationError{Field: "input", Message: "input is required"})
	}
	if ei.TrialNumber < 1 {
		errors = append(errors, EvaluationValidationError{Field: "trial_number", Message: "trial_number must be at least 1"})
	}

	return errors
}

// ExperimentItemScore represents a score submitted with an experiment item from SDK.
// These scores are computed by SDK evaluators and bundled with experiment items.
type ExperimentItemScore struct {
	Name          string                 `json:"name" binding:"required"`
	Value         *float64               `json:"value,omitempty"`
	Type          string                 `json:"type,omitempty"` // NUMERIC, CATEGORICAL, BOOLEAN
	StringValue   *string                `json:"string_value,omitempty"`
	Reason        *string                `json:"reason,omitempty"`
	Metadata      map[string]any `json:"metadata,omitempty"`
	ScoringFailed *bool                  `json:"scoring_failed,omitempty"`
}

type CreateExperimentItemRequest struct {
	DatasetItemID *string                `json:"dataset_item_id,omitempty"`
	TraceID       *string                `json:"trace_id,omitempty"`
	Input         map[string]any `json:"input" binding:"required"`
	Output        any            `json:"output,omitempty"`
	Expected      any            `json:"expected,omitempty"`
	TrialNumber   *int                   `json:"trial_number,omitempty"`
	Metadata      map[string]any `json:"metadata,omitempty"`
	// Scores computed by SDK evaluators, bundled with the experiment item
	Scores []ExperimentItemScore `json:"scores,omitempty"`
	// Error message if task execution failed
	Error *string `json:"error,omitempty"`
}

type CreateExperimentItemsBatchRequest struct {
	Items []CreateExperimentItemRequest `json:"items" binding:"required,dive"`
}

type ExperimentItemResponse struct {
	ID            uuid.UUID      `json:"id"`
	ExperimentID  uuid.UUID      `json:"experiment_id"`
	DatasetItemID *uuid.UUID     `json:"dataset_item_id,omitempty"`
	TraceID       *string        `json:"trace_id,omitempty"` // W3C hex
	Input         map[string]any `json:"input"`
	Output        any            `json:"output,omitempty"`
	Expected      any            `json:"expected,omitempty"`
	TrialNumber   int            `json:"trial_number"`
	Metadata      map[string]any `json:"metadata,omitempty"`
	Error         *string        `json:"error,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
}

func (ei *ExperimentItem) ToResponse() *ExperimentItemResponse {
	return &ExperimentItemResponse{
		ID:            ei.ID,
		ExperimentID:  ei.ExperimentID,
		DatasetItemID: ei.DatasetItemID,
		TraceID:       ei.TraceID,
		Input:         ei.Input,
		Output:        ei.Output,
		Expected:      ei.Expected,
		TrialNumber:   ei.TrialNumber,
		Metadata:      ei.Metadata,
		Error:         ei.Error,
		CreatedAt:     ei.CreatedAt,
	}
}

// ============================================================================
// Experiment Comparison Types
// ============================================================================

// EvaluatorScoreAggregation holds statistical metrics for a score across experiment items.

