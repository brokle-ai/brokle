package evaluation

import (
	evaluationDomain "brokle/internal/core/domain/evaluation"
)

// ExecutionListResponse mirrors the dashboard SPA's existing shape
// (`executions`/`total`/`page`/`limit`) rather than the canonical
// `data`/`pagination` envelope to avoid breaking the frontend.
type ExecutionListResponse struct {
	Executions []*evaluationDomain.EvaluatorExecutionResponse `json:"executions"`
	Total      int64                                          `json:"total"`
	Page       int                                            `json:"page"`
	Limit      int                                            `json:"limit"`
}
