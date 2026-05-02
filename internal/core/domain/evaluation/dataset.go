// Package evaluation provides domain entities for quality scoring, datasets, and experiments.
package evaluation

import (
	"time"

	"github.com/google/uuid"

	"brokle/pkg/uid"
)

type Dataset struct {
	ID               uuid.UUID              `json:"id"`
	ProjectID        uuid.UUID              `json:"project_id"`
	Name             string                 `json:"name"`
	Description      *string                `json:"description,omitempty"`
	Metadata         map[string]any `json:"metadata,omitempty"`
	CurrentVersionID *uuid.UUID             `json:"current_version_id,omitempty"` // Pinned version (nil = use latest)
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

func NewDataset(projectID uuid.UUID, name string) *Dataset {
	now := time.Now()
	return &Dataset{
		ID:        uid.New(),
		ProjectID: projectID,
		Name:      name,
		Metadata:  make(map[string]any),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (d *Dataset) Validate() []EvaluationValidationError {
	var errors []EvaluationValidationError

	if d.Name == "" {
		errors = append(errors, EvaluationValidationError{Field: "name", Message: "name is required"})
	}
	if len(d.Name) > 255 {
		errors = append(errors, EvaluationValidationError{Field: "name", Message: "name must be 255 characters or less"})
	}

	return errors
}

type CreateDatasetRequest struct {
	Name        string                 `json:"name" binding:"required,min=1,max=255"`
	Description *string                `json:"description,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type UpdateDatasetRequest struct {
	Name        *string                `json:"name,omitempty" binding:"omitempty,min=1,max=255"`
	Description *string                `json:"description,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type DatasetResponse struct {
	ID               uuid.UUID      `json:"id"`
	ProjectID        uuid.UUID      `json:"project_id"`
	Name             string         `json:"name"`
	Description      *string        `json:"description,omitempty"`
	Metadata         map[string]any `json:"metadata,omitempty"`
	CurrentVersionID *uuid.UUID     `json:"current_version_id,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

func (d *Dataset) ToResponse() *DatasetResponse {
	return &DatasetResponse{
		ID:               d.ID,
		ProjectID:        d.ProjectID,
		Name:             d.Name,
		Description:      d.Description,
		Metadata:         d.Metadata,
		CurrentVersionID: d.CurrentVersionID,
		CreatedAt:        d.CreatedAt,
		UpdatedAt:        d.UpdatedAt,
	}
}

// DatasetItemSource represents the origin of a dataset item.
type DatasetItemSource string

const (
	DatasetItemSourceManual DatasetItemSource = "manual"
	DatasetItemSourceTrace  DatasetItemSource = "trace"
	DatasetItemSourceSpan   DatasetItemSource = "span"
	DatasetItemSourceCSV    DatasetItemSource = "csv"
	DatasetItemSourceJSON   DatasetItemSource = "json"
	DatasetItemSourceSDK    DatasetItemSource = "sdk"
)

// IsValid checks if the source is a valid DatasetItemSource value.
func (s DatasetItemSource) IsValid() bool {
	switch s {
	case DatasetItemSourceManual, DatasetItemSourceTrace, DatasetItemSourceSpan,
		DatasetItemSourceCSV, DatasetItemSourceJSON, DatasetItemSourceSDK:
		return true
	default:
		return false
	}
}

// DatasetItem represents an individual test case within a dataset.
type DatasetItem struct {
	ID            uuid.UUID              `json:"id"`
	DatasetID     uuid.UUID              `json:"dataset_id"`
	Input         map[string]any `json:"input"`
	Expected      map[string]any `json:"expected,omitempty"`
	Metadata      map[string]any `json:"metadata,omitempty"`
	Source        DatasetItemSource      `json:"source"`
	SourceTraceID *string                `json:"source_trace_id,omitempty"`
	SourceSpanID  *string                `json:"source_span_id,omitempty"`
	ContentHash   *string                `json:"content_hash,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
}

func NewDatasetItem(datasetID uuid.UUID, input map[string]any) *DatasetItem {
	return &DatasetItem{
		ID:        uid.New(),
		DatasetID: datasetID,
		Input:     input,
		Metadata:  make(map[string]any),
		Source:    DatasetItemSourceManual,
		CreatedAt: time.Now(),
	}
}

// NewDatasetItemWithSource creates a new dataset item with explicit source tracking.
func NewDatasetItemWithSource(datasetID uuid.UUID, input map[string]any, source DatasetItemSource) *DatasetItem {
	return &DatasetItem{
		ID:        uid.New(),
		DatasetID: datasetID,
		Input:     input,
		Metadata:  make(map[string]any),
		Source:    source,
		CreatedAt: time.Now(),
	}
}

func (di *DatasetItem) Validate() []EvaluationValidationError {
	var errors []EvaluationValidationError

	if di.Input == nil || len(di.Input) == 0 {
		errors = append(errors, EvaluationValidationError{Field: "input", Message: "input is required"})
	}

	return errors
}

type CreateDatasetItemRequest struct {
	Input    map[string]any `json:"input" binding:"required"`
	Expected map[string]any `json:"expected,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type CreateDatasetItemsBatchRequest struct {
	Items       []CreateDatasetItemRequest `json:"items" binding:"required,dive"`
	Deduplicate bool                       `json:"deduplicate,omitempty"`
}

// ============================================================================
// Bulk Import Types for Dataset Items
// ============================================================================

// KeysMapping defines how to map source fields to dataset item fields.
type KeysMapping struct {
	InputKeys    []string `json:"input_keys"`
	ExpectedKeys []string `json:"expected_keys"`
	MetadataKeys []string `json:"metadata_keys"`
}

// BulkImportResult contains the result of a bulk import operation.
type BulkImportResult struct {
	Created int      `json:"created"`
	Skipped int      `json:"skipped"`
	Errors  []string `json:"errors,omitempty"`
}

// CreateDatasetItemsFromTracesRequest is the request to create dataset items from production traces.
type CreateDatasetItemsFromTracesRequest struct {
	TraceIDs    []string     `json:"trace_ids" binding:"required,min=1"`
	KeysMapping *KeysMapping `json:"keys_mapping,omitempty"`
	Deduplicate bool         `json:"deduplicate"`
}

// CreateDatasetItemsFromSpansRequest is the request to create dataset items from production spans.
type CreateDatasetItemsFromSpansRequest struct {
	SpanIDs     []string     `json:"span_ids" binding:"required,min=1"`
	KeysMapping *KeysMapping `json:"keys_mapping,omitempty"`
	Deduplicate bool         `json:"deduplicate"`
}

// ImportDatasetItemsFromJSONRequest is the request to import dataset items from JSON data.
type ImportDatasetItemsFromJSONRequest struct {
	Items       []map[string]any `json:"items" binding:"required,min=1"`
	KeysMapping *KeysMapping             `json:"keys_mapping,omitempty"`
	Deduplicate bool                     `json:"deduplicate"`
	Source      DatasetItemSource        `json:"source,omitempty"`
}

// CSVColumnMapping defines how CSV columns map to dataset item fields.
type CSVColumnMapping struct {
	InputColumn     string   `json:"input_column" binding:"required"`
	ExpectedColumn  string   `json:"expected_column,omitempty"`
	MetadataColumns []string `json:"metadata_columns,omitempty"`
}

// ImportDatasetItemsFromCSVRequest is the request to import dataset items from CSV data.
type ImportDatasetItemsFromCSVRequest struct {
	Content       string           `json:"content" binding:"required"`
	ColumnMapping CSVColumnMapping `json:"column_mapping" binding:"required"`
	HasHeader     bool             `json:"has_header"`
	Deduplicate   bool             `json:"deduplicate"`
}

// ExportFormat specifies the format for exporting dataset items.
type ExportFormat string

const (
	ExportFormatJSON ExportFormat = "json"
	ExportFormatCSV  ExportFormat = "csv"
)

// ExportDatasetItemsRequest is the request to export dataset items.
type ExportDatasetItemsRequest struct {
	Format ExportFormat `json:"format" binding:"required,oneof=json csv"`
}

type DatasetItemResponse struct {
	ID            uuid.UUID         `json:"id"`
	DatasetID     uuid.UUID         `json:"dataset_id"`
	Input         map[string]any    `json:"input"`
	Expected      map[string]any    `json:"expected,omitempty"`
	Metadata      map[string]any    `json:"metadata,omitempty"`
	Source        DatasetItemSource `json:"source"`
	SourceTraceID *string           `json:"source_trace_id,omitempty"` // W3C hex
	SourceSpanID  *string           `json:"source_span_id,omitempty"`  // W3C hex
	CreatedAt     time.Time         `json:"created_at"`
}

func (di *DatasetItem) ToResponse() *DatasetItemResponse {
	return &DatasetItemResponse{
		ID:            di.ID,
		DatasetID:     di.DatasetID,
		Input:         di.Input,
		Expected:      di.Expected,
		Metadata:      di.Metadata,
		Source:        di.Source,
		SourceTraceID: di.SourceTraceID,
		SourceSpanID:  di.SourceSpanID,
		CreatedAt:     di.CreatedAt,
	}
}


type DatasetWithItemCount struct {
	Dataset
	ItemCount int64 `json:"item_count"`
}

// DatasetWithItemCountResponse is the API response for a dataset with item count.
type DatasetWithItemCountResponse struct {
	ID               uuid.UUID      `json:"id"`
	ProjectID        uuid.UUID      `json:"project_id"`
	Name             string         `json:"name"`
	Description      *string        `json:"description,omitempty"`
	Metadata         map[string]any `json:"metadata,omitempty"`
	CurrentVersionID *uuid.UUID     `json:"current_version_id,omitempty"`
	ItemCount        int64          `json:"item_count"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

// ToResponse converts DatasetWithItemCount to its API response format.
func (d *DatasetWithItemCount) ToResponse() *DatasetWithItemCountResponse {
	return &DatasetWithItemCountResponse{
		ID:               d.ID,
		ProjectID:        d.ProjectID,
		Name:             d.Name,
		Description:      d.Description,
		Metadata:         d.Metadata,
		CurrentVersionID: d.CurrentVersionID,
		ItemCount:        d.ItemCount,
		CreatedAt:        d.CreatedAt,
		UpdatedAt:        d.UpdatedAt,
	}
}

