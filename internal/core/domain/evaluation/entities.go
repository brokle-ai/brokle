// Package evaluation provides domain entities for quality scoring, datasets, and experiments.
package evaluation

import (
	"time"

	"brokle/pkg/ulid"
)

type ScoreDataType string

const (
	ScoreDataTypeNumeric     ScoreDataType = "NUMERIC"
	ScoreDataTypeCategorical ScoreDataType = "CATEGORICAL"
	ScoreDataTypeBoolean     ScoreDataType = "BOOLEAN"
)

// ScoreConfig defines metadata and validation rules for a score type.
// Stored in PostgreSQL for transactional consistency.
type ScoreConfig struct {
	ID          ulid.ULID              `json:"id" gorm:"type:char(26);primaryKey"`
	ProjectID   ulid.ULID              `json:"project_id" gorm:"type:char(26);not null;index"`
	Name        string                 `json:"name" gorm:"type:varchar(100);not null"`
	Description *string                `json:"description,omitempty" gorm:"type:text"`
	DataType    ScoreDataType          `json:"data_type" gorm:"column:data_type;type:varchar(20);not null;default:'NUMERIC'"`
	MinValue    *float64               `json:"min_value,omitempty" gorm:"type:decimal(10,4)"`
	MaxValue    *float64               `json:"max_value,omitempty" gorm:"type:decimal(10,4)"`
	Categories  []string               `json:"categories,omitempty" gorm:"type:jsonb;serializer:json"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" gorm:"type:jsonb;serializer:json;default:'{}'"`
	CreatedAt   time.Time              `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt   time.Time              `json:"updated_at" gorm:"not null;autoUpdateTime"`
}

func (ScoreConfig) TableName() string {
	return "score_configs"
}

func NewScoreConfig(projectID ulid.ULID, name string, dataType ScoreDataType) *ScoreConfig {
	now := time.Now()
	return &ScoreConfig{
		ID:        ulid.New(),
		ProjectID: projectID,
		Name:      name,
		DataType:  dataType,
		Metadata:  make(map[string]interface{}),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (sc *ScoreConfig) Validate() []ValidationError {
	var errors []ValidationError

	if sc.Name == "" {
		errors = append(errors, ValidationError{Field: "name", Message: "name is required"})
	}
	if len(sc.Name) > 100 {
		errors = append(errors, ValidationError{Field: "name", Message: "name must be 100 characters or less"})
	}

	switch sc.DataType {
	case ScoreDataTypeNumeric:
		if sc.MinValue != nil && sc.MaxValue != nil && *sc.MinValue > *sc.MaxValue {
			errors = append(errors, ValidationError{Field: "max_value", Message: "max_value must be greater than or equal to min_value"})
		}
	case ScoreDataTypeCategorical:
		if len(sc.Categories) == 0 {
			errors = append(errors, ValidationError{Field: "categories", Message: "categories are required for CATEGORICAL type"})
		}
	case ScoreDataTypeBoolean:
	default:
		errors = append(errors, ValidationError{Field: "data_type", Message: "invalid data type, must be NUMERIC, CATEGORICAL, or BOOLEAN"})
	}

	return errors
}

type CreateScoreConfigRequest struct {
	Name        string                 `json:"name" binding:"required,min=1,max=100"`
	Description *string                `json:"description,omitempty"`
	DataType    ScoreDataType          `json:"data_type" binding:"required,oneof=NUMERIC CATEGORICAL BOOLEAN"`
	MinValue    *float64               `json:"min_value,omitempty"`
	MaxValue    *float64               `json:"max_value,omitempty"`
	Categories  []string               `json:"categories,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type UpdateScoreConfigRequest struct {
	Name        *string                `json:"name,omitempty" binding:"omitempty,min=1,max=100"`
	Description *string                `json:"description,omitempty"`
	DataType    *ScoreDataType         `json:"data_type,omitempty" binding:"omitempty,oneof=NUMERIC CATEGORICAL BOOLEAN"`
	MinValue    *float64               `json:"min_value,omitempty"`
	MaxValue    *float64               `json:"max_value,omitempty"`
	Categories  []string               `json:"categories,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type ScoreConfigResponse struct {
	ID          string                 `json:"id"`
	ProjectID   string                 `json:"project_id"`
	Name        string                 `json:"name"`
	Description *string                `json:"description,omitempty"`
	DataType    ScoreDataType          `json:"data_type"`
	MinValue    *float64               `json:"min_value,omitempty"`
	MaxValue    *float64               `json:"max_value,omitempty"`
	Categories  []string               `json:"categories,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

func (sc *ScoreConfig) ToResponse() *ScoreConfigResponse {
	return &ScoreConfigResponse{
		ID:          sc.ID.String(),
		ProjectID:   sc.ProjectID.String(),
		Name:        sc.Name,
		Description: sc.Description,
		DataType:    sc.DataType,
		MinValue:    sc.MinValue,
		MaxValue:    sc.MaxValue,
		Categories:  sc.Categories,
		Metadata:    sc.Metadata,
		CreatedAt:   sc.CreatedAt,
		UpdatedAt:   sc.UpdatedAt,
	}
}

// Dataset represents a collection of test cases for evaluation.
type Dataset struct {
	ID          ulid.ULID              `json:"id" gorm:"type:char(26);primaryKey"`
	ProjectID   ulid.ULID              `json:"project_id" gorm:"type:char(26);not null;index"`
	Name        string                 `json:"name" gorm:"type:varchar(255);not null"`
	Description *string                `json:"description,omitempty" gorm:"type:text"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" gorm:"type:jsonb;serializer:json;default:'{}'"`
	CreatedAt   time.Time              `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt   time.Time              `json:"updated_at" gorm:"not null;autoUpdateTime"`
}

func (Dataset) TableName() string {
	return "datasets"
}

func NewDataset(projectID ulid.ULID, name string) *Dataset {
	now := time.Now()
	return &Dataset{
		ID:        ulid.New(),
		ProjectID: projectID,
		Name:      name,
		Metadata:  make(map[string]interface{}),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (d *Dataset) Validate() []ValidationError {
	var errors []ValidationError

	if d.Name == "" {
		errors = append(errors, ValidationError{Field: "name", Message: "name is required"})
	}
	if len(d.Name) > 255 {
		errors = append(errors, ValidationError{Field: "name", Message: "name must be 255 characters or less"})
	}

	return errors
}

type CreateDatasetRequest struct {
	Name        string                 `json:"name" binding:"required,min=1,max=255"`
	Description *string                `json:"description,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type UpdateDatasetRequest struct {
	Name        *string                `json:"name,omitempty" binding:"omitempty,min=1,max=255"`
	Description *string                `json:"description,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type DatasetResponse struct {
	ID          string                 `json:"id"`
	ProjectID   string                 `json:"project_id"`
	Name        string                 `json:"name"`
	Description *string                `json:"description,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

func (d *Dataset) ToResponse() *DatasetResponse {
	return &DatasetResponse{
		ID:          d.ID.String(),
		ProjectID:   d.ProjectID.String(),
		Name:        d.Name,
		Description: d.Description,
		Metadata:    d.Metadata,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

// DatasetItem represents an individual test case within a dataset.
type DatasetItem struct {
	ID        ulid.ULID              `json:"id" gorm:"type:char(26);primaryKey"`
	DatasetID ulid.ULID              `json:"dataset_id" gorm:"type:char(26);not null;index"`
	Input     map[string]interface{} `json:"input" gorm:"type:jsonb;serializer:json;not null"`
	Expected  map[string]interface{} `json:"expected,omitempty" gorm:"type:jsonb;serializer:json"`
	Metadata  map[string]interface{} `json:"metadata,omitempty" gorm:"type:jsonb;serializer:json;default:'{}'"`
	CreatedAt time.Time              `json:"created_at" gorm:"not null;autoCreateTime"`
}

func (DatasetItem) TableName() string {
	return "dataset_items"
}

func NewDatasetItem(datasetID ulid.ULID, input map[string]interface{}) *DatasetItem {
	return &DatasetItem{
		ID:        ulid.New(),
		DatasetID: datasetID,
		Input:     input,
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
	}
}

func (di *DatasetItem) Validate() []ValidationError {
	var errors []ValidationError

	if di.Input == nil || len(di.Input) == 0 {
		errors = append(errors, ValidationError{Field: "input", Message: "input is required"})
	}

	return errors
}

type CreateDatasetItemRequest struct {
	Input    map[string]interface{} `json:"input" binding:"required"`
	Expected map[string]interface{} `json:"expected,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type CreateDatasetItemsBatchRequest struct {
	Items []CreateDatasetItemRequest `json:"items" binding:"required,dive"`
}

type DatasetItemResponse struct {
	ID        string                 `json:"id"`
	DatasetID string                 `json:"dataset_id"`
	Input     map[string]interface{} `json:"input"`
	Expected  map[string]interface{} `json:"expected,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

func (di *DatasetItem) ToResponse() *DatasetItemResponse {
	return &DatasetItemResponse{
		ID:        di.ID.String(),
		DatasetID: di.DatasetID.String(),
		Input:     di.Input,
		Expected:  di.Expected,
		Metadata:  di.Metadata,
		CreatedAt: di.CreatedAt,
	}
}

// ExperimentStatus represents the current state of an experiment.
type ExperimentStatus string

const (
	ExperimentStatusPending   ExperimentStatus = "pending"
	ExperimentStatusRunning   ExperimentStatus = "running"
	ExperimentStatusCompleted ExperimentStatus = "completed"
	ExperimentStatusFailed    ExperimentStatus = "failed"
)

// Experiment represents a batch evaluation run.
type Experiment struct {
	ID          ulid.ULID              `json:"id" gorm:"type:char(26);primaryKey"`
	ProjectID   ulid.ULID              `json:"project_id" gorm:"type:char(26);not null;index"`
	DatasetID   *ulid.ULID             `json:"dataset_id,omitempty" gorm:"type:char(26);index"`
	Name        string                 `json:"name" gorm:"type:varchar(255);not null"`
	Description *string                `json:"description,omitempty" gorm:"type:text"`
	Status      ExperimentStatus       `json:"status" gorm:"type:varchar(20);not null;default:'pending'"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" gorm:"type:jsonb;serializer:json;default:'{}'"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	CreatedAt   time.Time              `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt   time.Time              `json:"updated_at" gorm:"not null;autoUpdateTime"`
}

func (Experiment) TableName() string {
	return "experiments"
}

func NewExperiment(projectID ulid.ULID, name string) *Experiment {
	now := time.Now()
	return &Experiment{
		ID:        ulid.New(),
		ProjectID: projectID,
		Name:      name,
		Status:    ExperimentStatusPending,
		Metadata:  make(map[string]interface{}),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (e *Experiment) Validate() []ValidationError {
	var errors []ValidationError

	if e.Name == "" {
		errors = append(errors, ValidationError{Field: "name", Message: "name is required"})
	}
	if len(e.Name) > 255 {
		errors = append(errors, ValidationError{Field: "name", Message: "name must be 255 characters or less"})
	}

	switch e.Status {
	case ExperimentStatusPending, ExperimentStatusRunning, ExperimentStatusCompleted, ExperimentStatusFailed:
	default:
		errors = append(errors, ValidationError{Field: "status", Message: "invalid status"})
	}

	return errors
}

type CreateExperimentRequest struct {
	Name        string                 `json:"name" binding:"required,min=1,max=255"`
	DatasetID   *string                `json:"dataset_id,omitempty"`
	Description *string                `json:"description,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type UpdateExperimentRequest struct {
	Name        *string                `json:"name,omitempty" binding:"omitempty,min=1,max=255"`
	Description *string                `json:"description,omitempty"`
	Status      *ExperimentStatus      `json:"status,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type ExperimentFilter struct {
	DatasetID *ulid.ULID
	Status    *ExperimentStatus
}

type ExperimentResponse struct {
	ID          string                 `json:"id"`
	ProjectID   string                 `json:"project_id"`
	DatasetID   *string                `json:"dataset_id,omitempty"`
	Name        string                 `json:"name"`
	Description *string                `json:"description,omitempty"`
	Status      ExperimentStatus       `json:"status"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

func (e *Experiment) ToResponse() *ExperimentResponse {
	var datasetID *string
	if e.DatasetID != nil {
		id := e.DatasetID.String()
		datasetID = &id
	}
	return &ExperimentResponse{
		ID:          e.ID.String(),
		ProjectID:   e.ProjectID.String(),
		DatasetID:   datasetID,
		Name:        e.Name,
		Description: e.Description,
		Status:      e.Status,
		Metadata:    e.Metadata,
		StartedAt:   e.StartedAt,
		CompletedAt: e.CompletedAt,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

// ExperimentItem represents an individual result from an experiment run.
type ExperimentItem struct {
	ID            ulid.ULID              `json:"id" gorm:"type:char(26);primaryKey"`
	ExperimentID  ulid.ULID              `json:"experiment_id" gorm:"type:char(26);not null;index"`
	DatasetItemID *ulid.ULID             `json:"dataset_item_id,omitempty" gorm:"type:char(26)"`
	TraceID       *string                `json:"trace_id,omitempty" gorm:"type:varchar(32);index"`
	Input         map[string]interface{} `json:"input" gorm:"type:jsonb;serializer:json;not null"`
	Output        map[string]interface{} `json:"output,omitempty" gorm:"type:jsonb;serializer:json"`
	Expected      map[string]interface{} `json:"expected,omitempty" gorm:"type:jsonb;serializer:json"`
	TrialNumber   int                    `json:"trial_number" gorm:"not null;default:1"`
	Metadata      map[string]interface{} `json:"metadata,omitempty" gorm:"type:jsonb;serializer:json;default:'{}'"`
	CreatedAt     time.Time              `json:"created_at" gorm:"not null;autoCreateTime"`
}

func (ExperimentItem) TableName() string {
	return "experiment_items"
}

func NewExperimentItem(experimentID ulid.ULID, input map[string]interface{}) *ExperimentItem {
	return &ExperimentItem{
		ID:           ulid.New(),
		ExperimentID: experimentID,
		Input:        input,
		TrialNumber:  1,
		Metadata:     make(map[string]interface{}),
		CreatedAt:    time.Now(),
	}
}

func (ei *ExperimentItem) Validate() []ValidationError {
	var errors []ValidationError

	if ei.Input == nil || len(ei.Input) == 0 {
		errors = append(errors, ValidationError{Field: "input", Message: "input is required"})
	}
	if ei.TrialNumber < 1 {
		errors = append(errors, ValidationError{Field: "trial_number", Message: "trial_number must be at least 1"})
	}

	return errors
}

type CreateExperimentItemRequest struct {
	DatasetItemID *string                `json:"dataset_item_id,omitempty"`
	TraceID       *string                `json:"trace_id,omitempty"`
	Input         map[string]interface{} `json:"input" binding:"required"`
	Output        map[string]interface{} `json:"output,omitempty"`
	Expected      map[string]interface{} `json:"expected,omitempty"`
	TrialNumber   *int                   `json:"trial_number,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

type CreateExperimentItemsBatchRequest struct {
	Items []CreateExperimentItemRequest `json:"items" binding:"required,dive"`
}

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

func (ei *ExperimentItem) ToResponse() *ExperimentItemResponse {
	var datasetItemID *string
	if ei.DatasetItemID != nil {
		id := ei.DatasetItemID.String()
		datasetItemID = &id
	}
	return &ExperimentItemResponse{
		ID:            ei.ID.String(),
		ExperimentID:  ei.ExperimentID.String(),
		DatasetItemID: datasetItemID,
		TraceID:       ei.TraceID,
		Input:         ei.Input,
		Output:        ei.Output,
		Expected:      ei.Expected,
		TrialNumber:   ei.TrialNumber,
		Metadata:      ei.Metadata,
		CreatedAt:     ei.CreatedAt,
	}
}
