package evaluation

import (
	evaluationDomain "brokle/internal/core/domain/evaluation"
)

// Huma operation types for the score-config feature of the evaluation package.

type CreateScoreConfigRequest struct {
	Name        string         `json:"name" minLength:"1" maxLength:"100"`
	Description *string        `json:"description,omitempty"`
	Type        ScoreType      `json:"type" enum:"NUMERIC,CATEGORICAL,BOOLEAN"`
	MinValue    *float64       `json:"min_value,omitempty"`
	MaxValue    *float64       `json:"max_value,omitempty"`
	Categories  []string       `json:"categories,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type UpdateScoreConfigRequest struct {
	Name        *string        `json:"name,omitempty" minLength:"1" maxLength:"100"`
	Description *string        `json:"description,omitempty"`
	Type        *ScoreType     `json:"type,omitempty" enum:"NUMERIC,CATEGORICAL,BOOLEAN"`
	MinValue    *float64       `json:"min_value,omitempty"`
	MaxValue    *float64       `json:"max_value,omitempty"`
	Categories  []string       `json:"categories,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type CreateScoreConfigInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	Body      CreateScoreConfigRequest
}

type ScoreConfigOutput struct {
	Body *evaluationDomain.ScoreConfigResponse
}

type ListScoreConfigsInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	Page      int    `query:"page" required:"false" minimum:"1"`
	Limit     int    `query:"limit" required:"false"`
}

type ListScoreConfigsOutput struct {
	Body pageList[*evaluationDomain.ScoreConfigResponse]
}

type GetScoreConfigInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	ConfigID  string `path:"configId" format:"uuid"`
}

type UpdateScoreConfigInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	ConfigID  string `path:"configId" format:"uuid"`
	Body      UpdateScoreConfigRequest
}

type DeleteScoreConfigInput = GetScoreConfigInput
