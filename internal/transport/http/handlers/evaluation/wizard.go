package evaluation

import (
	"net/http"

	evaluationDomain "brokle/internal/core/domain/evaluation"
	"brokle/internal/transport/http/httpctx"
	"brokle/pkg/request"
	"brokle/pkg/response"
)

func (h *Handler) WizardCreate(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
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

func (h *Handler) WizardValidate(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
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

func (h *Handler) WizardEstimate(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
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

func (h *Handler) WizardDatasetFields(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
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

func (h *Handler) WizardGetConfig(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
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
