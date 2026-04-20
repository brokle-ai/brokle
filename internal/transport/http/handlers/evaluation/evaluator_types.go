package evaluation

import (
	evaluationDomain "brokle/internal/core/domain/evaluation"
	"brokle/internal/transport/http/handlers/shared"
)

// Huma operation types for the evaluator feature of the evaluation package.

type CreateEvaluatorRequest = evaluationDomain.CreateEvaluatorRequest

type UpdateEvaluatorRequest = evaluationDomain.UpdateEvaluatorRequest

type EvaluatorOutput struct {
	Body *evaluationDomain.EvaluatorResponse
}

type CreateEvaluatorInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	Body      CreateEvaluatorRequest
}

type ListEvaluatorsInput struct {
	ProjectID  string `path:"projectId" format:"uuid"`
	Page       int    `query:"page" required:"false" minimum:"1"`
	Limit      int    `query:"limit" required:"false"`
	SortBy     string `query:"sort_by" required:"false"`
	SortDir    string `query:"sort_dir" required:"false" enum:"asc,desc"`
	Status     string `query:"status" required:"false"`
	ScorerType string `query:"scorer_type" required:"false"`
	Search     string `query:"search" required:"false"`
}

type EvaluatorListOutput struct {
	Body pageList[*evaluationDomain.EvaluatorResponse]
}

type GetEvaluatorInput struct {
	ProjectID   string `path:"projectId" format:"uuid"`
	EvaluatorID string `path:"evaluatorId" format:"uuid"`
}

type UpdateEvaluatorInput struct {
	ProjectID   string `path:"projectId" format:"uuid"`
	EvaluatorID string `path:"evaluatorId" format:"uuid"`
	Body        UpdateEvaluatorRequest
}

type DeleteEvaluatorInput = GetEvaluatorInput

// MessageOutput wraps the shared MessageResponse for activate/deactivate
// endpoints (status message responses).
type MessageOutput struct {
	Body *shared.MessageResponse
}

type TriggerEvaluatorInput struct {
	ProjectID   string                          `path:"projectId" format:"uuid"`
	EvaluatorID string                          `path:"evaluatorId" format:"uuid"`
	Body        evaluationDomain.TriggerOptions `doc:"All fields optional; empty object acceptable"`
}

type TriggerEvaluatorOutput struct {
	Body *evaluationDomain.TriggerResponse
}

type TestEvaluatorInput struct {
	ProjectID   string                                `path:"projectId" format:"uuid"`
	EvaluatorID string                                `path:"evaluatorId" format:"uuid"`
	Body        evaluationDomain.TestEvaluatorRequest `doc:"All fields optional; empty object acceptable"`
}

type TestEvaluatorOutput struct {
	Body *evaluationDomain.TestEvaluatorResponse
}

type GetEvaluatorAnalyticsInput struct {
	ProjectID     string `path:"projectId" format:"uuid"`
	EvaluatorID   string `path:"evaluatorId" format:"uuid"`
	Period        string `query:"period" required:"false" doc:"24h, 7d, 30d"`
	FromTimestamp string `query:"from_timestamp" required:"false" doc:"RFC3339"`
	ToTimestamp   string `query:"to_timestamp" required:"false" doc:"RFC3339"`
}

type EvaluatorAnalyticsOutput struct {
	Body *evaluationDomain.EvaluatorAnalyticsResponse
}
