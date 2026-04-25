package evaluation

import (
	evaluationDomain "brokle/internal/core/domain/evaluation"
)

// Dataset request body DTOs shared across dashboard and SDK planes.
// Per-route input/output structs stay colocated with their handler
// methods in dataset.go — they document the route's wire contract.

type CreateDatasetRequest struct {
	Name        string         `json:"name" minLength:"1" maxLength:"255"`
	Description *string        `json:"description,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type UpdateDatasetRequest struct {
	Name        *string        `json:"name,omitempty" minLength:"1" maxLength:"255"`
	Description *string        `json:"description,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type CreateDatasetItemRequest = evaluationDomain.CreateDatasetItemRequest

type BatchCreateDatasetItemsRequest = evaluationDomain.CreateDatasetItemsBatchRequest

type ImportFromJSONRequest struct {
	Items       []map[string]any    `json:"items" minItems:"1"`
	KeysMapping *KeysMappingRequest `json:"keys_mapping,omitempty"`
	Deduplicate bool                `json:"deduplicate"`
	Source      string              `json:"source,omitempty"`
}

type ImportFromCSVRequest struct {
	Content       string           `json:"content"`
	ColumnMapping csvColumnMapping `json:"column_mapping"`
	HasHeader     bool             `json:"has_header"`
	Deduplicate   bool             `json:"deduplicate"`
}

type csvColumnMapping struct {
	InputColumn     string   `json:"input_column"`
	ExpectedColumn  string   `json:"expected_column,omitempty"`
	MetadataColumns []string `json:"metadata_columns,omitempty"`
}

type CreateFromTracesRequest struct {
	TraceIDs    []string            `json:"trace_ids" minItems:"1"`
	KeysMapping *KeysMappingRequest `json:"keys_mapping,omitempty"`
	Deduplicate bool                `json:"deduplicate"`
}

type CreateFromSpansRequest struct {
	SpanIDs     []string            `json:"span_ids" minItems:"1"`
	KeysMapping *KeysMappingRequest `json:"keys_mapping,omitempty"`
	Deduplicate bool                `json:"deduplicate"`
}

type CreateDatasetVersionRequest = evaluationDomain.CreateDatasetVersionRequest

type PinDatasetVersionRequest = evaluationDomain.PinDatasetVersionRequest
