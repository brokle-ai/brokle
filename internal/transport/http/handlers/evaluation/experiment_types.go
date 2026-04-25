package evaluation

import (
	evaluationDomain "brokle/internal/core/domain/evaluation"
)

// Experiment request/response DTOs shared across dashboard and SDK planes.
// Per-route input/output structs stay colocated with their handler
// methods in experiment.go — they document the route's wire contract.

type CreateExperimentRequest = evaluationDomain.CreateExperimentRequest

type UpdateExperimentRequest = evaluationDomain.UpdateExperimentRequest

type RerunExperimentRequest = evaluationDomain.RerunExperimentRequest

type BatchCreateExperimentItemsRequest = evaluationDomain.CreateExperimentItemsBatchRequest

// CompareExperimentsRequest is the dual-plane compare payload (string UUIDs
// for ergonomic SDK use, parsed server-side).
type CompareExperimentsRequest struct {
	ExperimentIDs []string `json:"experiment_ids" minItems:"2" maxItems:"10"`
	BaselineID    *string  `json:"baseline_id,omitempty"`
}

type ScoreAggregationResponse struct {
	Mean   float64 `json:"mean"`
	StdDev float64 `json:"std_dev"`
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
	Count  uint64  `json:"count"`
}

type ScoreDiffResponse struct {
	Type       string  `json:"type"`
	Difference float64 `json:"difference,omitempty"`
	Direction  string  `json:"direction,omitempty"`
}

type ExperimentSummaryResponse struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type CompareExperimentsResponse struct {
	Experiments map[string]*ExperimentSummaryResponse           `json:"experiments"`
	Scores      map[string]map[string]*ScoreAggregationResponse `json:"scores"`
	Diffs       map[string]map[string]*ScoreDiffResponse        `json:"diffs,omitempty"`
}

type ExperimentItemListResponse struct {
	Items []*ExperimentItemResponse `json:"items"`
	Total int64                     `json:"total"`
}
