package evaluation

import (
	evaluationDomain "brokle/internal/core/domain/evaluation"
)

// Huma operation types for the execution feature of the evaluation package.

// ExecutionListResponse mirrors the dashboard SPA's existing shape
// (`executions`/`total`/`page`/`limit`) rather than the canonical
// `data`/`total`/... envelope to avoid breaking the frontend.
type ExecutionListResponse struct {
	Executions []*evaluationDomain.EvaluatorExecutionResponse `json:"executions"`
	Total      int64                                          `json:"total"`
	Page       int                                            `json:"page"`
	Limit      int                                            `json:"limit"`
}

type ListExecutionsInput struct {
	ProjectID   string `path:"projectId" format:"uuid"`
	EvaluatorID string `path:"evaluatorId" format:"uuid"`
	Page        int    `query:"page" required:"false" minimum:"1"`
	Limit       int    `query:"limit" required:"false"`
	Status      string `query:"status" required:"false"`
	TriggerType string `query:"trigger_type" required:"false"`
}

type ListExecutionsOutput struct {
	Body *ExecutionListResponse
}

type GetLatestExecutionInput struct {
	ProjectID   string `path:"projectId" format:"uuid"`
	EvaluatorID string `path:"evaluatorId" format:"uuid"`
}

type ExecutionOutput struct {
	Body *evaluationDomain.EvaluatorExecutionResponse
}

type GetExecutionInput struct {
	ProjectID   string `path:"projectId" format:"uuid"`
	EvaluatorID string `path:"evaluatorId" format:"uuid"`
	ExecutionID string `path:"executionId" format:"uuid"`
}

type ExecutionDetailOutput struct {
	Body *evaluationDomain.EvaluatorExecutionDetailFlat
}
