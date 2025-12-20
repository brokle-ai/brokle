// Package evaluation provides the evaluation domain model.
//
// The evaluation domain handles quality scoring configurations,
// datasets for offline evaluation, experiments, and evaluation rules.
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
