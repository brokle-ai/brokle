package evaluation

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	evaluationDomain "brokle/internal/core/domain/evaluation"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/pagination"
)

func registerExecutionRoutes(api huma.API, h *handler) {
	huma.Register(api, huma.Operation{
		OperationID: "list-evaluator-executions",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/evaluators/{evaluatorId}/executions",
		Tags:        []string{"evaluator-executions"},
		Summary:     "List executions for an evaluator",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.listExecutions)

	huma.Register(api, huma.Operation{
		OperationID: "get-latest-evaluator-execution",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/evaluators/{evaluatorId}/executions/latest",
		Tags:        []string{"evaluator-executions"},
		Summary:     "Get the most recent evaluator execution",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getLatestExecution)

	huma.Register(api, huma.Operation{
		OperationID: "get-evaluator-execution",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/evaluators/{evaluatorId}/executions/{executionId}",
		Tags:        []string{"evaluator-executions"},
		Summary:     "Get an evaluator execution",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getExecution)

	huma.Register(api, huma.Operation{
		OperationID: "get-evaluator-execution-detail",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/evaluators/{evaluatorId}/executions/{executionId}/detail",
		Tags:        []string{"evaluator-executions"},
		Summary:     "Get detailed execution info including span-level results",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getExecutionDetail)
}

func (h *handler) listExecutions(ctx context.Context, in *ListExecutionsInput) (*ListExecutionsOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	evaluatorID, err := parseEvaluatorID(in.EvaluatorID)
	if err != nil {
		return nil, err
	}
	params := pagination.Params{Page: in.Page, Limit: in.Limit}
	params.SetDefaults("created_at")

	var filter evaluationDomain.ExecutionFilter
	if in.Status != "" {
		s := evaluationDomain.ExecutionStatus(in.Status)
		filter.Status = &s
	}
	if in.TriggerType != "" {
		t := evaluationDomain.TriggerType(in.TriggerType)
		filter.TriggerType = &t
	}

	execs, total, err := h.evaluatorExecSvc.ListByEvaluatorID(ctx, evaluatorID, projectID, &filter, params)
	if err != nil {
		return nil, err
	}
	out := make([]*evaluationDomain.EvaluatorExecutionResponse, len(execs))
	for i, e := range execs {
		out[i] = e.ToResponse()
	}
	return &ListExecutionsOutput{Body: &ExecutionListResponse{
		Executions: out, Total: total, Page: params.Page, Limit: params.Limit,
	}}, nil
}

func (h *handler) getLatestExecution(ctx context.Context, in *GetLatestExecutionInput) (*ExecutionOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	evaluatorID, err := parseEvaluatorID(in.EvaluatorID)
	if err != nil {
		return nil, err
	}
	exec, err := h.evaluatorExecSvc.GetLatestByEvaluatorID(ctx, evaluatorID, projectID)
	if err != nil {
		return nil, err
	}
	if exec == nil {
		return nil, appErrors.NewNotFoundError("no executions found for this evaluator")
	}
	return &ExecutionOutput{Body: exec.ToResponse()}, nil
}

func (h *handler) getExecution(ctx context.Context, in *GetExecutionInput) (*ExecutionOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	// evaluatorId is parsed for validation only; the service lookup is keyed
	// by execution ID (which is globally unique).
	if _, err := parseEvaluatorID(in.EvaluatorID); err != nil {
		return nil, err
	}
	executionID, err := parseExecutionID(in.ExecutionID)
	if err != nil {
		return nil, err
	}
	exec, err := h.evaluatorExecSvc.GetByID(ctx, executionID, projectID)
	if err != nil {
		return nil, err
	}
	return &ExecutionOutput{Body: exec.ToResponse()}, nil
}

func (h *handler) getExecutionDetail(ctx context.Context, in *GetExecutionInput) (*ExecutionDetailOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	evaluatorID, err := parseEvaluatorID(in.EvaluatorID)
	if err != nil {
		return nil, err
	}
	executionID, err := parseExecutionID(in.ExecutionID)
	if err != nil {
		return nil, err
	}
	detail, err := h.evaluatorExecSvc.GetExecutionDetail(ctx, executionID, projectID, evaluatorID)
	if err != nil {
		return nil, err
	}
	return &ExecutionDetailOutput{Body: detail.ToFlat()}, nil
}
