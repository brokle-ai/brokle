package evaluation

// Request-body DTOs for the score-config feature.

type CreateScoreConfigRequest struct {
	Name        string         `json:"name"                  validate:"required,min=1,max=100"`
	Description *string        `json:"description,omitempty"`
	Type        ScoreType      `json:"type"                  validate:"required,oneof=NUMERIC CATEGORICAL BOOLEAN"`
	MinValue    *float64       `json:"min_value,omitempty"`
	MaxValue    *float64       `json:"max_value,omitempty"`
	Categories  []string       `json:"categories,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type UpdateScoreConfigRequest struct {
	Name        *string        `json:"name,omitempty"        validate:"omitempty,min=1,max=100"`
	Description *string        `json:"description,omitempty"`
	Type        *ScoreType     `json:"type,omitempty"        validate:"omitempty,oneof=NUMERIC CATEGORICAL BOOLEAN"`
	MinValue    *float64       `json:"min_value,omitempty"`
	MaxValue    *float64       `json:"max_value,omitempty"`
	Categories  []string       `json:"categories,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}
