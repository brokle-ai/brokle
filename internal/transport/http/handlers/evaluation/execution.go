package evaluation

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	evaluationDomain "brokle/internal/core/domain/evaluation"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/pagination"
	"brokle/pkg/request"
	"brokle/pkg/response"
)

func registerExecutionRoutes(r chi.Router, h *handler) {
	r.Route("/evaluators/{evaluatorId}/executions", func(r chi.Router) {
		r.Get("/", h.listExecutions)
		r.Get("/latest", h.getLatestExecution)
		r.Get("/{executionId}", h.getExecution)
		r.Get("/{executionId}/detail", h.getExecutionDetail)
	})
}

func (h *handler) listExecutions(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
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
		Executions: out, Total: total, Page: params.Page, Limit: params.Limit,
	})
}

func (h *handler) getLatestExecution(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
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
		response.WriteError(w, appErrors.NewNotFoundError("no executions found for this evaluator"))
		return
	}
	response.Success(w, exec.ToResponse())
}

func (h *handler) getExecution(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
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

func (h *handler) getExecutionDetail(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
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
