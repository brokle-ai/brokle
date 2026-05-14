// Package evaluation provides domain entities for quality scoring, datasets, and experiments.
package evaluation

import (
	"time"

	"github.com/google/uuid"

	"brokle/pkg/uid"
)

type DatasetVersion struct {
	ID          uuid.UUID              `json:"id"`
	DatasetID   uuid.UUID              `json:"dataset_id"`
	Version     int                    `json:"version"`
	ItemCount   int                    `json:"item_count"`
	Description *string                `json:"description,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	CreatedBy   *uuid.UUID             `json:"created_by,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// NewDatasetVersion creates a new dataset version.
func NewDatasetVersion(datasetID uuid.UUID, version int, itemCount int) *DatasetVersion {
	return &DatasetVersion{
		ID:        uid.New(),
		DatasetID: datasetID,
		Version:   version,
		ItemCount: itemCount,
		Metadata:  make(map[string]any),
		CreatedAt: time.Now(),
	}
}

func (dv *DatasetVersion) Validate() []EvaluationValidationError {
	var errors []EvaluationValidationError

	if dv.Version < 1 {
		errors = append(errors, EvaluationValidationError{Field: "version", Message: "version must be at least 1"})
	}
	if dv.ItemCount < 0 {
		errors = append(errors, EvaluationValidationError{Field: "item_count", Message: "item_count cannot be negative"})
	}

	return errors
}

// DatasetVersionResponse is the API response for a dataset version.
type DatasetVersionResponse struct {
	ID          uuid.UUID      `json:"id"`
	DatasetID   uuid.UUID      `json:"dataset_id"`
	Version     int            `json:"version"`
	ItemCount   int            `json:"item_count"`
	Description *string        `json:"description,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	CreatedBy   *uuid.UUID     `json:"created_by,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
}

func (dv *DatasetVersion) ToResponse() *DatasetVersionResponse {
	return &DatasetVersionResponse{
		ID:          dv.ID,
		DatasetID:   dv.DatasetID,
		Version:     dv.Version,
		ItemCount:   dv.ItemCount,
		Description: dv.Description,
		Metadata:    dv.Metadata,
		CreatedBy:   dv.CreatedBy,
		CreatedAt:   dv.CreatedAt,
	}
}

// DatasetItemVersion is the join table linking items to versions.
type DatasetItemVersion struct {
	DatasetVersionID uuid.UUID `json:"dataset_version_id"`
	DatasetItemID    uuid.UUID `json:"dataset_item_id"`
}

// CreateDatasetVersionRequest is the request to create a new version manually.
type CreateDatasetVersionRequest struct {
	Description *string                `json:"description,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// PinDatasetVersionRequest is the request to pin a dataset to a specific version.
type PinDatasetVersionRequest struct {
	VersionID *uuid.UUID `json:"version_id"` // nil to unpin (use latest)
}

// DatasetVersionFilter is used for filtering versions.
type DatasetVersionFilter struct {
	DatasetID *uuid.UUID
}

// DatasetWithVersion extends Dataset to include version info in responses.
type DatasetWithVersionResponse struct {
	ID               uuid.UUID      `json:"id"`
	ProjectID        uuid.UUID      `json:"project_id"`
	Name             string         `json:"name"`
	Description      *string        `json:"description,omitempty"`
	Metadata         map[string]any `json:"metadata,omitempty"`
	CurrentVersionID *uuid.UUID     `json:"current_version_id,omitempty"`
	CurrentVersion   *int           `json:"current_version,omitempty"`
	LatestVersion    *int           `json:"latest_version,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

// ============================================================================
// Dataset Filter and Pagination Types
// ============================================================================

// ============================================================================
// Experiment Metrics Types
// ============================================================================

// ExperimentMetricsResponse is the response for GET /experiments/{id}/metrics.
// It provides comprehensive metrics including progress, performance, and scores.

