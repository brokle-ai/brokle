// Package evaluation provides domain entities for quality scoring, datasets, and experiments.
package evaluation

import (
	"time"

	"github.com/google/uuid"

	"brokle/pkg/uid"
)

type ExperimentStatus string

const (
	ExperimentStatusPending   ExperimentStatus = "pending"
	ExperimentStatusRunning   ExperimentStatus = "running"
	ExperimentStatusCompleted ExperimentStatus = "completed"
	ExperimentStatusFailed    ExperimentStatus = "failed"
	ExperimentStatusPartial   ExperimentStatus = "partial"   // Some items completed, some failed
	ExperimentStatusCancelled ExperimentStatus = "cancelled" // User cancelled the experiment
)

// Experiment represents a batch evaluation run.
type Experiment struct {
	ID          uuid.UUID              `json:"id"`
	ProjectID   uuid.UUID              `json:"project_id"`
	DatasetID   *uuid.UUID             `json:"dataset_id,omitempty"`
	Name        string                 `json:"name"`
	Description *string                `json:"description,omitempty"`
	Status      ExperimentStatus       `json:"status"`
	Source      ExperimentSource       `json:"source"`
	ConfigID    *uuid.UUID             `json:"config_id,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	// Progress tracking fields
	TotalItems     int        `json:"total_items"`
	CompletedItems int        `json:"completed_items"`
	FailedItems    int        `json:"failed_items"`
	StartedAt      *time.Time `json:"started_at,omitempty"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	// Relationships (optional, loaded when needed)
	Config *ExperimentConfig `json:"config,omitempty"`
}

func NewExperiment(projectID uuid.UUID, name string) *Experiment {
	now := time.Now()
	return &Experiment{
		ID:        uid.New(),
		ProjectID: projectID,
		Name:      name,
		Status:    ExperimentStatusPending,
		Source:    ExperimentSourceSDK, // Default to SDK for backwards compatibility
		Metadata:  make(map[string]any),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewExperimentFromDashboard creates a new experiment from the dashboard wizard.
func NewExperimentFromDashboard(projectID uuid.UUID, name string) *Experiment {
	now := time.Now()
	return &Experiment{
		ID:        uid.New(),
		ProjectID: projectID,
		Name:      name,
		Status:    ExperimentStatusPending,
		Source:    ExperimentSourceDashboard,
		Metadata:  make(map[string]any),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (e *Experiment) Validate() []EvaluationValidationError {
	var errors []EvaluationValidationError

	if e.Name == "" {
		errors = append(errors, EvaluationValidationError{Field: "name", Message: "name is required"})
	}
	if len(e.Name) > 255 {
		errors = append(errors, EvaluationValidationError{Field: "name", Message: "name must be 255 characters or less"})
	}

	switch e.Status {
	case ExperimentStatusPending, ExperimentStatusRunning, ExperimentStatusCompleted, ExperimentStatusFailed, ExperimentStatusPartial, ExperimentStatusCancelled:
	default:
		errors = append(errors, EvaluationValidationError{Field: "status", Message: "invalid status"})
	}

	return errors
}

type CreateExperimentRequest struct {
	Name        string         `json:"name" binding:"required,min=1,max=255"`
	DatasetID   *uuid.UUID     `json:"dataset_id,omitempty"`
	Description *string        `json:"description,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// RerunExperimentRequest is the request to create a new experiment based on an existing one.
// The new experiment will have the same dataset but can have a different name and metadata.
type RerunExperimentRequest struct {
	Name        *string                `json:"name,omitempty" binding:"omitempty,min=1,max=255"`
	Description *string                `json:"description,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type UpdateExperimentRequest struct {
	Name        *string                `json:"name,omitempty" binding:"omitempty,min=1,max=255"`
	Description *string                `json:"description,omitempty"`
	Status      *ExperimentStatus      `json:"status,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type ExperimentFilter struct {
	DatasetID *uuid.UUID
	Status    *ExperimentStatus
	Search    *string
	IDs       []uuid.UUID // Filter by specific experiment IDs
}

// DatasetFilter defines filter criteria for listing datasets.
type DatasetFilter struct {
	Search *string
}

type ExperimentResponse struct {
	ID             uuid.UUID        `json:"id"`
	ProjectID      uuid.UUID        `json:"project_id"`
	DatasetID      *uuid.UUID       `json:"dataset_id,omitempty"`
	Name           string           `json:"name"`
	Description    *string          `json:"description,omitempty"`
	Status         ExperimentStatus `json:"status"`
	Source         ExperimentSource `json:"source"`
	ConfigID       *uuid.UUID       `json:"config_id,omitempty"`
	Metadata       map[string]any   `json:"metadata,omitempty"`
	TotalItems     int              `json:"total_items"`
	CompletedItems int              `json:"completed_items"`
	FailedItems    int              `json:"failed_items"`
	StartedAt      *time.Time       `json:"started_at,omitempty"`
	CompletedAt    *time.Time       `json:"completed_at,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

func (e *Experiment) ToResponse() *ExperimentResponse {
	return &ExperimentResponse{
		ID:             e.ID,
		ProjectID:      e.ProjectID,
		DatasetID:      e.DatasetID,
		Name:           e.Name,
		Description:    e.Description,
		Status:         e.Status,
		Source:         e.Source,
		ConfigID:       e.ConfigID,
		Metadata:       e.Metadata,
		TotalItems:     e.TotalItems,
		CompletedItems: e.CompletedItems,
		FailedItems:    e.FailedItems,
		StartedAt:      e.StartedAt,
		CompletedAt:    e.CompletedAt,
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
	}
}

// ExperimentProgressResponse is a lightweight response for progress polling.
type ExperimentProgressResponse struct {
	ID             uuid.UUID        `json:"id"`
	Status         ExperimentStatus `json:"status"`
	TotalItems     int              `json:"total_items"`
	CompletedItems int              `json:"completed_items"`
	FailedItems    int              `json:"failed_items"`
	PendingItems   int              `json:"pending_items"`
	ProgressPct    float64          `json:"progress_pct"`
	StartedAt      *time.Time       `json:"started_at,omitempty"`
	CompletedAt    *time.Time       `json:"completed_at,omitempty"`
	ElapsedSeconds *float64         `json:"elapsed_seconds,omitempty"`
	ETASeconds     *float64         `json:"eta_seconds,omitempty"`
}

// ToProgressResponse creates a progress response with derived fields.
func (e *Experiment) ToProgressResponse() *ExperimentProgressResponse {
	resp := &ExperimentProgressResponse{
		ID:             e.ID,
		Status:         e.Status,
		TotalItems:     e.TotalItems,
		CompletedItems: e.CompletedItems,
		FailedItems:    e.FailedItems,
		PendingItems:   e.TotalItems - e.CompletedItems - e.FailedItems,
		StartedAt:      e.StartedAt,
		CompletedAt:    e.CompletedAt,
	}

	// Calculate progress percentage
	if e.TotalItems > 0 {
		resp.ProgressPct = float64(e.CompletedItems+e.FailedItems) / float64(e.TotalItems) * 100
	}

	// Calculate elapsed time
	if e.StartedAt != nil {
		var elapsed float64

		if e.CompletedAt != nil {
			// Finished experiment: use fixed duration from start to completion
			elapsed = e.CompletedAt.Sub(*e.StartedAt).Seconds()
		} else if e.Status == ExperimentStatusRunning {
			// Running experiment: use live elapsed time
			elapsed = time.Since(*e.StartedAt).Seconds()
		}

		// Only set elapsed if we calculated it (skip pending experiments without completion)
		if e.CompletedAt != nil || e.Status == ExperimentStatusRunning {
			resp.ElapsedSeconds = &elapsed
		}

		// Calculate ETA only for running experiments
		if e.Status == ExperimentStatusRunning {
			processed := e.CompletedItems + e.FailedItems
			if processed > 0 && e.TotalItems > processed {
				rate := elapsed / float64(processed)
				remaining := float64(e.TotalItems - processed)
				eta := rate * remaining
				resp.ETASeconds = &eta
			}
		}
	}

	return resp
}


type EvaluatorScoreAggregation struct {
	Mean   float64 `json:"mean"`
	StdDev float64 `json:"std_dev"`
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
	Count  uint64  `json:"count"`
}

// ScoreDiffType represents the type of score difference.
type ScoreDiffType string

const (
	ScoreDiffTypeNumeric     ScoreDiffType = "NUMERIC"
	ScoreDiffTypeCategorical ScoreDiffType = "CATEGORICAL"
)

// ScoreDiff represents the difference between a score and its baseline.
type ScoreDiff struct {
	Type        ScoreDiffType `json:"type"`
	Difference  float64       `json:"difference,omitempty"`   // Absolute difference for NUMERIC
	Direction   string        `json:"direction,omitempty"`    // "+" or "-" for NUMERIC
	IsDifferent bool          `json:"is_different,omitempty"` // For CATEGORICAL
}

// ExperimentSummary contains basic info about an experiment for comparison.
type ExperimentSummary struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	ItemCount int64  `json:"item_count"`
}

// CompareExperimentsRequest is the request body for comparing experiments.
type CompareExperimentsRequest struct {
	ExperimentIDs []string `json:"experiment_ids" binding:"required,min=2,max=10"`
	BaselineID    *string  `json:"baseline_id,omitempty"`
}

// CompareExperimentsResponse contains the comparison results.
type CompareExperimentsResponse struct {
	Experiments map[string]*ExperimentSummary           `json:"experiments"`
	Scores      map[string]map[string]*EvaluatorScoreAggregation `json:"scores"`          // scoreName -> experimentID -> aggregation
	Diffs       map[string]map[string]*ScoreDiff        `json:"diffs,omitempty"` // scoreName -> experimentID -> diff (vs baseline)
}

// CalculateDiff computes the difference between two score aggregations.
func CalculateDiff(baseline, current *EvaluatorScoreAggregation) *ScoreDiff {
	if baseline == nil || current == nil {
		return nil
	}

	difference := current.Mean - baseline.Mean
	direction := "+"
	if difference < 0 {
		direction = "-"
	}

	return &ScoreDiff{
		Type:       ScoreDiffTypeNumeric,
		Difference: abs(difference),
		Direction:  direction,
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// ============================================================================
// Dataset Versioning Types
// ============================================================================

// DatasetVersion represents a snapshot of a dataset at a point in time.
// Versions are created automatically when items are added or removed.

type ExperimentMetricsResponse struct {
	ExperimentID uuid.UUID                    `json:"experiment_id"`
	Status       ExperimentStatus             `json:"status"`
	Progress     ExperimentProgressMetrics    `json:"progress"`
	Performance  ExperimentPerformanceMetrics `json:"performance"`
	Scores       map[string]*ScoreMetrics     `json:"scores,omitempty"`
}

// ExperimentProgressMetrics contains progress and success/error rate metrics.
type ExperimentProgressMetrics struct {
	TotalItems     int     `json:"total_items"`
	CompletedItems int     `json:"completed_items"`
	FailedItems    int     `json:"failed_items"`
	PendingItems   int     `json:"pending_items"`
	ProgressPct    float64 `json:"progress_pct"`
	SuccessRate    float64 `json:"success_rate"` // completed / (completed + failed) * 100
	ErrorRate      float64 `json:"error_rate"`   // failed / (completed + failed) * 100
}

// ExperimentPerformanceMetrics contains timing and ETA metrics.
type ExperimentPerformanceMetrics struct {
	StartedAt      *time.Time `json:"started_at,omitempty"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	ElapsedSeconds *float64   `json:"elapsed_seconds,omitempty"`
	ETASeconds     *float64   `json:"eta_seconds,omitempty"`
}

// ScoreMetrics contains statistical metrics for a score type.
type ScoreMetrics struct {
	Mean   float64 `json:"mean"`
	StdDev float64 `json:"std_dev"`
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
	Count  uint64  `json:"count"`
}

// DatasetWithItemCount extends Dataset to include item count in list responses.

