package evaluation

import "time"

// @Description Score data type (NUMERIC, CATEGORICAL, BOOLEAN)
type ScoreDataType string

const (
	ScoreDataTypeNumeric     ScoreDataType = "NUMERIC"
	ScoreDataTypeCategorical ScoreDataType = "CATEGORICAL"
	ScoreDataTypeBoolean     ScoreDataType = "BOOLEAN"
)

// @Description Dataset item data
type DatasetItemResponse struct {
	ID        string                 `json:"id"`
	DatasetID string                 `json:"dataset_id"`
	Input     map[string]interface{} `json:"input"`
	Expected  map[string]interface{} `json:"expected,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

// @Description Paginated dataset items response
type DatasetItemListResponse struct {
	Items []*DatasetItemResponse `json:"items"`
	Total int64                  `json:"total"`
}

// @Description Experiment item data
type ExperimentItemResponse struct {
	ID            string                 `json:"id"`
	ExperimentID  string                 `json:"experiment_id"`
	DatasetItemID *string                `json:"dataset_item_id,omitempty"`
	TraceID       *string                `json:"trace_id,omitempty"`
	Input         map[string]interface{} `json:"input"`
	Output        map[string]interface{} `json:"output,omitempty"`
	Expected      map[string]interface{} `json:"expected,omitempty"`
	TrialNumber   int                    `json:"trial_number"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
}

// @Description Paginated experiment items response
type ExperimentItemListResponse struct {
	Items []*ExperimentItemResponse `json:"items"`
	Total int64                     `json:"total"`
}

// @Description Score config creation request
type CreateRequest struct {
	Name        string                 `json:"name" binding:"required,min=1,max=100"`
	Description *string                `json:"description,omitempty"`
	DataType    ScoreDataType          `json:"data_type" binding:"required,oneof=NUMERIC CATEGORICAL BOOLEAN"`
	MinValue    *float64               `json:"min_value,omitempty"`
	MaxValue    *float64               `json:"max_value,omitempty"`
	Categories  []string               `json:"categories,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// @Description Score config update request
type UpdateRequest struct {
	Name        *string                `json:"name,omitempty" binding:"omitempty,min=1,max=100"`
	Description *string                `json:"description,omitempty"`
	DataType    *ScoreDataType         `json:"data_type,omitempty" binding:"omitempty,oneof=NUMERIC CATEGORICAL BOOLEAN"`
	MinValue    *float64               `json:"min_value,omitempty"`
	MaxValue    *float64               `json:"max_value,omitempty"`
	Categories  []string               `json:"categories,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// @Description Batch creation response with count
type SDKBatchCreateItemsResponse struct {
	Created int `json:"created"`
}

// @Description Batch experiment items creation response
type SDKBatchCreateExperimentItemsResponse struct {
	Created int `json:"created"`
}

// @Description SDK score creation request
type CreateScoreRequest struct {
	TraceID          string         `json:"trace_id" binding:"required"`
	SpanID           *string        `json:"span_id,omitempty"`
	Name             string         `json:"name" binding:"required"`
	Value            *float64       `json:"value,omitempty"`
	StringValue      *string        `json:"string_value,omitempty"`
	DataType         string         `json:"data_type" binding:"required,oneof=NUMERIC CATEGORICAL BOOLEAN"`
	Reason           *string        `json:"reason,omitempty"`
	Metadata         map[string]any `json:"metadata,omitempty"`
	ExperimentID     *string        `json:"experiment_id,omitempty"`
	ExperimentItemID *string        `json:"experiment_item_id,omitempty"`
}

// @Description Batch score creation request
type BatchScoreRequest struct {
	Scores []CreateScoreRequest `json:"scores" binding:"required,dive"`
}

// @Description Score data
type ScoreResponse struct {
	ID               string         `json:"id"`
	ProjectID        string         `json:"project_id"`
	TraceID          string         `json:"trace_id"`
	SpanID           string         `json:"span_id"`
	Name             string         `json:"name"`
	Value            *float64       `json:"value,omitempty"`
	StringValue      *string        `json:"string_value,omitempty"`
	DataType         string         `json:"data_type"`
	Source           string         `json:"source"`
	Reason           *string        `json:"reason,omitempty"`
	Metadata         map[string]any `json:"metadata,omitempty"`
	ExperimentID     *string        `json:"experiment_id,omitempty"`
	ExperimentItemID *string        `json:"experiment_item_id,omitempty"`
	Timestamp        time.Time      `json:"timestamp"`
}

// @Description Batch score creation response
type BatchScoreResponse struct {
	Created int `json:"created"`
}
