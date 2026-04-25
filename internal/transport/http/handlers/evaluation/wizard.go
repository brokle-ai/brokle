package evaluation

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	evaluationDomain "brokle/internal/core/domain/evaluation"
	"brokle/pkg/request"
	"brokle/pkg/response"
)

func registerWizardRoutes(r chi.Router, h *handler) {
	r.Route("/experiments/wizard", func(r chi.Router) {
		r.Post("/", h.wizardCreate)
		r.Post("/validate", h.wizardValidate)
		r.Post("/estimate", h.wizardEstimate)
	})
	r.Get("/datasets/{datasetId}/fields", h.wizardDatasetFields)
	r.Get("/experiments/{experimentId}/config", h.wizardGetConfig)
}

func (h *handler) wizardCreate(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	var body evaluationDomain.CreateExperimentFromWizardRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	exp, err := h.experimentWizardSvc.CreateFromWizard(r.Context(), projectID, userIDPtr(r.Context()), &body)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Created(w, exp.ToResponse())
}

func (h *handler) wizardValidate(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	var body evaluationDomain.ValidateStepRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	res, err := h.experimentWizardSvc.ValidateStep(r.Context(), projectID, &body)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, res)
}

func (h *handler) wizardEstimate(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	var body evaluationDomain.EstimateCostRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	res, err := h.experimentWizardSvc.EstimateCost(r.Context(), projectID, &body)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, res)
}

func (h *handler) wizardDatasetFields(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	datasetID, err := request.URLParamUUID(r, "datasetId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	res, err := h.experimentWizardSvc.GetDatasetFields(r.Context(), projectID, datasetID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, res)
}

func (h *handler) wizardGetConfig(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	experimentID, err := request.URLParamUUID(r, "experimentId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	cfg, err := h.experimentWizardSvc.GetExperimentConfig(r.Context(), experimentID, projectID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, cfg.ToResponse())
}
