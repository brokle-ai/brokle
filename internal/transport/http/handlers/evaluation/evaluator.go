package evaluation

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	analyticsDomain "brokle/internal/core/domain/analytics"
	evaluationDomain "brokle/internal/core/domain/evaluation"
	"brokle/internal/transport/http/handlers/shared"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/pagination"
	"brokle/pkg/request"
	"brokle/pkg/response"
)

// registerEvaluatorRoutes mounts evaluator CRUD + lifecycle under
// /api/v1/projects/{projectId}/evaluators.
func registerEvaluatorRoutes(r chi.Router, h *handler) {
	r.Route("/evaluators", func(r chi.Router) {
		r.Post("/", h.createEvaluator)
		r.Get("/", h.listEvaluators)
		r.Route("/{evaluatorId}", func(r chi.Router) {
			r.Get("/", h.getEvaluator)
			r.Put("/", h.updateEvaluator)
			r.Delete("/", h.deleteEvaluator)
			r.Post("/activate", h.activateEvaluator)
			r.Post("/deactivate", h.deactivateEvaluator)
			r.Post("/trigger", h.triggerEvaluator)
			r.Post("/test", h.testEvaluator)
			r.Get("/analytics", h.getEvaluatorAnalytics)
		})
	})
}

// ---- handlers -------------------------------------------------------

func (h *handler) createEvaluator(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	var body evaluationDomain.CreateEvaluatorRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	ev, err := h.evaluatorSvc.Create(r.Context(), projectID, userIDPtr(r.Context()), &body)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Created(w, ev.ToResponse())
}

func (h *handler) listEvaluators(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	q := r.URL.Query()
	allowedSortFields := []string{"name", "status", "sampling_rate", "created_at", "updated_at"}
	sortBy, err := pagination.ValidateSortField(q.Get("sort_by"), allowedSortFields)
	if err != nil {
		response.WriteError(w, appErrors.NewValidationError("sort_by", err.Error()))
		return
	}
	sortDir := q.Get("sort_dir")
	if sortDir != "" && sortDir != "asc" && sortDir != "desc" {
		response.WriteError(w, appErrors.NewValidationError("sort_dir", "must be 'asc' or 'desc'"))
		return
	}

	page, limit, err := readPagination(r)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	params := pagination.Params{
		Page:    page,
		Limit:   limit,
		SortBy:  sortBy,
		SortDir: sortDir,
	}
	params.SetDefaults("created_at")

	var filter evaluationDomain.EvaluatorFilter
	if v := q.Get("status"); v != "" {
		s := evaluationDomain.EvaluatorStatus(v)
		filter.Status = &s
	}
	if v := q.Get("scorer_type"); v != "" {
		st := evaluationDomain.ScorerType(v)
		filter.ScorerType = &st
	}
	if v := q.Get("search"); v != "" {
		filter.Search = &v
	}

	evs, total, err := h.evaluatorSvc.List(r.Context(), projectID, &filter, params)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	out := make([]*evaluationDomain.EvaluatorResponse, len(evs))
	for i, e := range evs {
		out[i] = e.ToResponse()
	}
	response.Success(w, pageList[*evaluationDomain.EvaluatorResponse]{
		Data: out, Total: total, Page: params.Page, Limit: params.Limit,
	})
}

func (h *handler) getEvaluator(w http.ResponseWriter, r *http.Request) {
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
	ev, err := h.evaluatorSvc.GetByID(r.Context(), evaluatorID, projectID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, ev.ToResponse())
}

func (h *handler) updateEvaluator(w http.ResponseWriter, r *http.Request) {
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
	var body evaluationDomain.UpdateEvaluatorRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	ev, err := h.evaluatorSvc.Update(r.Context(), evaluatorID, projectID, &body)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, ev.ToResponse())
}

func (h *handler) deleteEvaluator(w http.ResponseWriter, r *http.Request) {
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
	if err := h.evaluatorSvc.Delete(r.Context(), evaluatorID, projectID); err != nil {
		response.WriteError(w, err)
		return
	}
	response.NoContent(w)
}

func (h *handler) activateEvaluator(w http.ResponseWriter, r *http.Request) {
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
	if err := h.evaluatorSvc.Activate(r.Context(), evaluatorID, projectID); err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, &shared.MessageResponse{Message: "evaluator activated"})
}

func (h *handler) deactivateEvaluator(w http.ResponseWriter, r *http.Request) {
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
	if err := h.evaluatorSvc.Deactivate(r.Context(), evaluatorID, projectID); err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, &shared.MessageResponse{Message: "evaluator deactivated"})
}

func (h *handler) triggerEvaluator(w http.ResponseWriter, r *http.Request) {
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
	var body evaluationDomain.TriggerOptions
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	res, err := h.evaluatorSvc.TriggerEvaluator(r.Context(), evaluatorID, projectID, &body)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.JSON(w, http.StatusAccepted, res)
}

func (h *handler) testEvaluator(w http.ResponseWriter, r *http.Request) {
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
	var body evaluationDomain.TestEvaluatorRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	res, err := h.evaluatorSvc.TestEvaluator(r.Context(), evaluatorID, projectID, &body)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, res)
}

func (h *handler) getEvaluatorAnalytics(w http.ResponseWriter, r *http.Request) {
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
	q := r.URL.Query()
	period := q.Get("period")
	if period == "" {
		period = "7d"
	}
	fromTime, toTime, err := shared.ParseTimeRange(
		q.Get("from_timestamp"),
		q.Get("to_timestamp"),
		period,
		analyticsDomain.TimeRange7Days,
	)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	params := &evaluationDomain.EvaluatorAnalyticsParams{
		ProjectID:   projectID,
		EvaluatorID: evaluatorID,
		Period:      period,
		From:        &fromTime,
		To:          &toTime,
	}
	res, err := h.evaluatorSvc.GetAnalytics(r.Context(), evaluatorID, projectID, params)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, res)
}
