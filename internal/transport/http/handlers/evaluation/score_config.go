package evaluation

import (
	"net/http"

	evaluationDomain "brokle/internal/core/domain/evaluation"
	"brokle/internal/transport/http/httpctx"
	"brokle/pkg/request"
	"brokle/pkg/response"
)

func (h *Handler) CreateScoreConfig(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	var body CreateScoreConfigRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	domainReq := &evaluationDomain.CreateScoreConfigRequest{
		Name:        body.Name,
		Description: body.Description,
		Type:        evaluationDomain.ScoreType(body.Type),
		MinValue:    body.MinValue,
		MaxValue:    body.MaxValue,
		Categories:  body.Categories,
		Metadata:    body.Metadata,
	}
	cfg, err := h.scoreConfigSvc.Create(r.Context(), projectID, domainReq)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Created(w, cfg.ToResponse())
}

func (h *Handler) ListScoreConfigs(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	page, limit, err := readPagination(r)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	cfgs, total, err := h.scoreConfigSvc.List(r.Context(), projectID, page, limit)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	out := make([]*evaluationDomain.ScoreConfigResponse, len(cfgs))
	for i, c := range cfgs {
		out[i] = c.ToResponse()
	}
	response.Success(w, pageList[*evaluationDomain.ScoreConfigResponse]{
		Data: out, Pagination: response.BuildPagination(page, limit, total),
	})
}

func (h *Handler) GetScoreConfig(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	configID, err := request.URLParamUUID(r, "configId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	cfg, err := h.scoreConfigSvc.GetByID(r.Context(), configID, projectID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, cfg.ToResponse())
}

func (h *Handler) UpdateScoreConfig(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	configID, err := request.URLParamUUID(r, "configId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	var body UpdateScoreConfigRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	var scoreType *evaluationDomain.ScoreType
	if body.Type != nil {
		st := evaluationDomain.ScoreType(*body.Type)
		scoreType = &st
	}
	domainReq := &evaluationDomain.UpdateScoreConfigRequest{
		Name:        body.Name,
		Description: body.Description,
		Type:        scoreType,
		MinValue:    body.MinValue,
		MaxValue:    body.MaxValue,
		Categories:  body.Categories,
		Metadata:    body.Metadata,
	}
	cfg, err := h.scoreConfigSvc.Update(r.Context(), configID, projectID, domainReq)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, cfg.ToResponse())
}

func (h *Handler) DeleteScoreConfig(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	configID, err := request.URLParamUUID(r, "configId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	if err := h.scoreConfigSvc.Delete(r.Context(), configID, projectID); err != nil {
		response.WriteError(w, err)
		return
	}
	response.NoContent(w)
}
