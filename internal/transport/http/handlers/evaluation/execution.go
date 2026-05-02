package evaluation

import (
	"net/http"

	evaluationDomain "brokle/internal/core/domain/evaluation"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/pagination"
	"brokle/pkg/request"
	"brokle/pkg/response"
)

func (h *Handler) ListExecutions(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	evaluatorID, err := request.URLParamUUID(r, "evaluatorId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	page, limit, err := readPagination(r)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	params := pagination.Params{Page: page, Limit: limit}
	params.SetDefaults("created_at")

	var filter evaluationDomain.ExecutionFilter
	if v := r.URL.Query().Get("status"); v != "" {
		s := evaluationDomain.ExecutionStatus(v)
		filter.Status = &s
	}
	if v := r.URL.Query().Get("trigger_type"); v != "" {
		t := evaluationDomain.TriggerType(v)
		filter.TriggerType = &t
	}

	execs, total, err := h.evaluatorExecSvc.ListByEvaluatorID(r.Context(), evaluatorID, projectID, &filter, params)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	out := make([]*evaluationDomain.EvaluatorExecutionResponse, len(execs))
	for i, e := range execs {
		out[i] = e.ToResponse()
	}
	response.Success(w, &ExecutionListResponse{
		Data: out, Pagination: response.BuildPagination(params.Page, params.Limit, total),
	})
}

func (h *Handler) GetLatestExecution(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	evaluatorID, err := request.URLParamUUID(r, "evaluatorId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	exec, err := h.evaluatorExecSvc.GetLatestByEvaluatorID(r.Context(), evaluatorID, projectID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	if exec == nil {
		response.WriteError(w, appErrors.NotFound("execution", appErrors.WithMessage("no executions found for this evaluator")))
		return
	}
	response.Success(w, exec.ToResponse())
}

func (h *Handler) GetExecution(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	// evaluatorId is parsed for validation only.
	if _, err := request.URLParamUUID(r, "evaluatorId"); err != nil {
		response.WriteError(w, err)
		return
	}
	executionID, err := request.URLParamUUID(r, "executionId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	exec, err := h.evaluatorExecSvc.GetByID(r.Context(), executionID, projectID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, exec.ToResponse())
}

func (h *Handler) GetExecutionDetail(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	evaluatorID, err := request.URLParamUUID(r, "evaluatorId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	executionID, err := request.URLParamUUID(r, "executionId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	detail, err := h.evaluatorExecSvc.GetExecutionDetail(r.Context(), executionID, projectID, evaluatorID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, detail.ToFlat())
}
