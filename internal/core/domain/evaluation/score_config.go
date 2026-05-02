// Package evaluation provides domain entities for quality scoring, datasets, and experiments.
package evaluation

import (
	"time"

	"github.com/google/uuid"

	"brokle/pkg/uid"
)

type ScoreType string

const (
	ScoreTypeNumeric     ScoreType = "NUMERIC"
	ScoreTypeCategorical ScoreType = "CATEGORICAL"
	ScoreTypeBoolean     ScoreType = "BOOLEAN"
)

// ScoreConfig defines metadata and validation rules for a score type.
// Stored in PostgreSQL for transactional consistency.
type ScoreConfig struct {
	ID          uuid.UUID              `json:"id"`
	ProjectID   uuid.UUID              `json:"project_id"`
	Name        string                 `json:"name"`
	Description *string                `json:"description,omitempty"`
	Type        ScoreType              `json:"type"`
	MinValue    *float64               `json:"min_value,omitempty"`
	MaxValue    *float64               `json:"max_value,omitempty"`
	Categories  []string               `json:"categories,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

func NewScoreConfig(projectID uuid.UUID, name string, scoreType ScoreType) *ScoreConfig {
	now := time.Now()
	return &ScoreConfig{
		ID:        uid.New(),
		ProjectID: projectID,
		Name:      name,
		Type:      scoreType,
		Metadata:  make(map[string]any),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

type EvaluationValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (sc *ScoreConfig) Validate() []EvaluationValidationError {
	var errors []EvaluationValidationError

	if sc.Name == "" {
		errors = append(errors, EvaluationValidationError{Field: "name", Message: "name is required"})
	}
	if len(sc.Name) > 100 {
		errors = append(errors, EvaluationValidationError{Field: "name", Message: "name must be 100 characters or less"})
	}

	switch sc.Type {
	case ScoreTypeNumeric:
		if sc.MinValue != nil && sc.MaxValue != nil && *sc.MinValue > *sc.MaxValue {
			errors = append(errors, EvaluationValidationError{Field: "max_value", Message: "max_value must be greater than or equal to min_value"})
		}
	case ScoreTypeCategorical:
		if len(sc.Categories) == 0 {
			errors = append(errors, EvaluationValidationError{Field: "categories", Message: "categories are required for CATEGORICAL type"})
		}
	case ScoreTypeBoolean:
	default:
		errors = append(errors, EvaluationValidationError{Field: "type", Message: "invalid type, must be NUMERIC, CATEGORICAL, or BOOLEAN"})
	}

	return errors
}

type CreateScoreConfigRequest struct {
	Name        string                 `json:"name" binding:"required,min=1,max=100"`
	Description *string                `json:"description,omitempty"`
	Type        ScoreType              `json:"type" binding:"required,oneof=NUMERIC CATEGORICAL BOOLEAN"`
	MinValue    *float64               `json:"min_value,omitempty"`
	MaxValue    *float64               `json:"max_value,omitempty"`
	Categories  []string               `json:"categories,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type UpdateScoreConfigRequest struct {
	Name        *string                `json:"name,omitempty" binding:"omitempty,min=1,max=100"`
	Description *string                `json:"description,omitempty"`
	Type        *ScoreType             `json:"type,omitempty" binding:"omitempty,oneof=NUMERIC CATEGORICAL BOOLEAN"`
	MinValue    *float64               `json:"min_value,omitempty"`
	MaxValue    *float64               `json:"max_value,omitempty"`
	Categories  []string               `json:"categories,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type ScoreConfigResponse struct {
	ID          uuid.UUID      `json:"id"`
	ProjectID   uuid.UUID      `json:"project_id"`
	Name        string         `json:"name"`
	Description *string        `json:"description,omitempty"`
	Type        ScoreType      `json:"type"`
	MinValue    *float64       `json:"min_value,omitempty"`
	MaxValue    *float64       `json:"max_value,omitempty"`
	Categories  []string       `json:"categories,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

func (sc *ScoreConfig) ToResponse() *ScoreConfigResponse {
	return &ScoreConfigResponse{
		ID:          sc.ID,
		ProjectID:   sc.ProjectID,
		Name:        sc.Name,
		Description: sc.Description,
		Type:        sc.Type,
		MinValue:    sc.MinValue,
		MaxValue:    sc.MaxValue,
		Categories:  sc.Categories,
		Metadata:    sc.Metadata,
		CreatedAt:   sc.CreatedAt,
		UpdatedAt:   sc.UpdatedAt,
	}
}


