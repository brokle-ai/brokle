package evaluation

import (
	evaluationDomain "brokle/internal/core/domain/evaluation"
	"brokle/pkg/response"
)

// ExecutionListResponse is the canonical paginated envelope for evaluator
// execution lists.
type ExecutionListResponse struct {
	Data       []*evaluationDomain.EvaluatorExecutionResponse `json:"data"`
	Pagination *response.Pagination                           `json:"pagination"`
}
