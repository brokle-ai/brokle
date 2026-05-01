package evaluation

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"

	evaluationDomain "brokle/internal/core/domain/evaluation"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/request"
	"brokle/pkg/response"
)

// ---- shared cores ---------------------------------------------------

func (h *Handler) createExperimentCore(ctx context.Context, projectID uuid.UUID, body *CreateExperimentRequest) (*evaluationDomain.ExperimentResponse, error) {
	exp, err := h.experimentSvc.Create(ctx, projectID, body)
	if err != nil {
		return nil, err
	}
	h.logger.InfoContext(ctx, "evaluation: experiment created",
		"experiment_id", exp.ID, "project_id", projectID, "name", exp.Name)
	return exp.ToResponse(), nil
}

func (h *Handler) listExperimentsCore(
	ctx context.Context,
	projectID uuid.UUID,
	datasetIDStr, statusStr, search, idsStr string,
	page, limit int,
) ([]*evaluationDomain.ExperimentResponse, int64, int, int, error) {
	var filter *evaluationDomain.ExperimentFilter
	if datasetIDStr != "" {
		datasetID, err := uuid.Parse(datasetIDStr)
		if err != nil {
			return nil, 0, page, limit, appErrors.InvalidParam("dataset_id", "must be a valid UUID")
		}
		filter = &evaluationDomain.ExperimentFilter{}
		filter.DatasetID = &datasetID
	}
	if statusStr != "" {
		status := evaluationDomain.ExperimentStatus(statusStr)
		switch status {
		case evaluationDomain.ExperimentStatusPending,
			evaluationDomain.ExperimentStatusRunning,
			evaluationDomain.ExperimentStatusCompleted,
			evaluationDomain.ExperimentStatusFailed,
			evaluationDomain.ExperimentStatusPartial,
			evaluationDomain.ExperimentStatusCancelled:
			if filter == nil {
				filter = &evaluationDomain.ExperimentFilter{}
			}
			filter.Status = &status
		default:
			return nil, 0, page, limit, appErrors.InvalidParam("status", "must be pending, running, completed, failed, partial, or cancelled")
		}
	}
	if search != "" {
		if filter == nil {
			filter = &evaluationDomain.ExperimentFilter{}
		}
		filter.Search = &search
	}
	if idsStr != "" {
		var ids []uuid.UUID
		for _, idStr := range strings.Split(idsStr, ",") {
			idStr = strings.TrimSpace(idStr)
			if idStr == "" {
				continue
			}
			id, err := uuid.Parse(idStr)
			if err != nil {
				return nil, 0, page, limit, appErrors.InvalidParam("ids", "invalid UUID: "+idStr)
			}
			ids = append(ids, id)
		}
		if len(ids) > 0 {
			if filter == nil {
				filter = &evaluationDomain.ExperimentFilter{}
			}
			filter.IDs = ids
		}
	}

	experiments, total, err := h.experimentSvc.List(ctx, projectID, filter, page, limit)
	if err != nil {
		return nil, 0, page, limit, err
	}
	out := make([]*evaluationDomain.ExperimentResponse, len(experiments))
	for i, e := range experiments {
		out[i] = e.ToResponse()
	}
	return out, total, page, limit, nil
}

func (h *Handler) compareExperimentsCore(ctx context.Context, projectID uuid.UUID, body *CompareExperimentsRequest) (*CompareExperimentsResponse, error) {
	experimentIDs := make([]uuid.UUID, len(body.ExperimentIDs))
	for i, idStr := range body.ExperimentIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return nil, appErrors.InvalidParam("experiment_ids", "invalid UUID at index "+strconv.Itoa(i))
		}
		experimentIDs[i] = id
	}

	var baselineID *uuid.UUID
	if body.BaselineID != nil {
		id, err := uuid.Parse(*body.BaselineID)
		if err != nil {
			return nil, appErrors.InvalidParam("baseline_id", "must be a valid UUID")
		}
		baselineID = &id
	}

	result, err := h.experimentSvc.CompareExperiments(ctx, projectID, experimentIDs, baselineID)
	if err != nil {
		return nil, err
	}

	resp := &CompareExperimentsResponse{
		Experiments: make(map[string]*ExperimentSummaryResponse),
		Scores:      make(map[string]map[string]*ScoreAggregationResponse),
	}
	for id, exp := range result.Experiments {
		resp.Experiments[id] = &ExperimentSummaryResponse{Name: exp.Name, Status: exp.Status}
	}
	for scoreName, expScores := range result.Scores {
		resp.Scores[scoreName] = make(map[string]*ScoreAggregationResponse)
		for expID, agg := range expScores {
			resp.Scores[scoreName][expID] = &ScoreAggregationResponse{
				Mean:   agg.Mean,
				StdDev: agg.StdDev,
				Min:    agg.Min,
				Max:    agg.Max,
				Count:  agg.Count,
			}
		}
	}
	if result.Diffs != nil {
		resp.Diffs = make(map[string]map[string]*ScoreDiffResponse)
		for scoreName, expDiffs := range result.Diffs {
			resp.Diffs[scoreName] = make(map[string]*ScoreDiffResponse)
			for expID, diff := range expDiffs {
				if diff != nil {
					resp.Diffs[scoreName][expID] = &ScoreDiffResponse{
						Type:       string(diff.Type),
						Difference: diff.Difference,
						Direction:  diff.Direction,
					}
				}
			}
		}
	}
	return resp, nil
}

// ---- dashboard handlers --------------------------------------------

func (h *Handler) DashCreateExperiment(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	var body CreateExperimentRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	resp, err := h.createExperimentCore(r.Context(), projectID, &body)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Created(w, resp)
}

func (h *Handler) DashListExperiments(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	page, limit, err := readPagination(r)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	q := r.URL.Query()
	out, total, page, limit, err := h.listExperimentsCore(r.Context(), projectID,
		q.Get("dataset_id"), q.Get("status"), q.Get("search"), q.Get("ids"),
		page, limit)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, pageList[*evaluationDomain.ExperimentResponse]{
		Data: out, Total: total, Page: page, Limit: limit,
	})
}

func (h *Handler) DashCompareExperiments(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	var body CompareExperimentsRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	resp, err := h.compareExperimentsCore(r.Context(), projectID, &body)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, resp)
}

func (h *Handler) DashGetExperiment(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	experimentID, err := request.URLParamUUID(r, "experimentId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	exp, err := h.experimentSvc.GetByID(r.Context(), experimentID, projectID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, exp.ToResponse())
}

func (h *Handler) DashUpdateExperiment(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	experimentID, err := request.URLParamUUID(r, "experimentId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	var body UpdateExperimentRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	exp, err := h.experimentSvc.Update(r.Context(), experimentID, projectID, &body)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	h.logger.InfoContext(r.Context(), "evaluation: experiment updated",
		"experiment_id", experimentID, "project_id", projectID, "status", exp.Status)
	response.Success(w, exp.ToResponse())
}

func (h *Handler) DashDeleteExperiment(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	experimentID, err := request.URLParamUUID(r, "experimentId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	if err := h.experimentSvc.Delete(r.Context(), experimentID, projectID); err != nil {
		response.WriteError(w, err)
		return
	}
	response.NoContent(w)
}

func (h *Handler) DashRerunExperiment(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	experimentID, err := request.URLParamUUID(r, "experimentId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	var body RerunExperimentRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	exp, err := h.experimentSvc.Rerun(r.Context(), experimentID, projectID, &body)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	h.logger.InfoContext(r.Context(), "evaluation: experiment rerun",
		"experiment_id", exp.ID, "source_experiment_id", experimentID, "project_id", projectID)
	response.Created(w, exp.ToResponse())
}

func (h *Handler) DashGetExperimentProgress(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	experimentID, err := request.URLParamUUID(r, "experimentId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	progress, err := h.experimentSvc.GetProgress(r.Context(), experimentID, projectID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, progress)
}

func (h *Handler) DashGetExperimentMetrics(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	experimentID, err := request.URLParamUUID(r, "experimentId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	metrics, err := h.experimentSvc.GetMetrics(r.Context(), projectID, experimentID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, metrics)
}

func (h *Handler) DashListExperimentItems(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	experimentID, err := request.URLParamUUID(r, "experimentId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	limit, err := request.QueryInt(r, "limit", 50)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset, err := request.QueryInt(r, "offset", 0)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	if offset < 0 {
		offset = 0
	}

	items, total, err := h.experimentItemSvc.List(r.Context(), experimentID, projectID, limit, offset)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	out := make([]*ExperimentItemResponse, len(items))
	for i, it := range items {
		out[i] = toExperimentItemResponse(it)
	}
	response.Success(w, &ExperimentItemListResponse{Items: out, Total: total})
}

// ---- SDK handlers --------------------------------------------------

func (h *Handler) SdkCreateExperiment(w http.ResponseWriter, r *http.Request) {
	projectID := projectIDForSDK(r)
	var body CreateExperimentRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	resp, err := h.createExperimentCore(r.Context(), projectID, &body)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Created(w, resp)
}

func (h *Handler) SdkListExperiments(w http.ResponseWriter, r *http.Request) {
	projectID := projectIDForSDK(r)
	page, limit, err := readPagination(r)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	q := r.URL.Query()
	out, total, page, limit, err := h.listExperimentsCore(r.Context(), projectID,
		q.Get("dataset_id"), q.Get("status"), q.Get("search"), q.Get("ids"),
		page, limit)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, pageList[*evaluationDomain.ExperimentResponse]{
		Data: out, Total: total, Page: page, Limit: limit,
	})
}

func (h *Handler) SdkCompareExperiments(w http.ResponseWriter, r *http.Request) {
	projectID := projectIDForSDK(r)
	var body CompareExperimentsRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	resp, err := h.compareExperimentsCore(r.Context(), projectID, &body)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, resp)
}

func (h *Handler) SdkGetExperiment(w http.ResponseWriter, r *http.Request) {
	projectID := projectIDForSDK(r)
	experimentID, err := request.URLParamUUID(r, "experimentId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	exp, err := h.experimentSvc.GetByID(r.Context(), experimentID, projectID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, exp.ToResponse())
}

func (h *Handler) SdkUpdateExperiment(w http.ResponseWriter, r *http.Request) {
	projectID := projectIDForSDK(r)
	experimentID, err := request.URLParamUUID(r, "experimentId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	var body UpdateExperimentRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	exp, err := h.experimentSvc.Update(r.Context(), experimentID, projectID, &body)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, exp.ToResponse())
}

func (h *Handler) SdkRerunExperiment(w http.ResponseWriter, r *http.Request) {
	projectID := projectIDForSDK(r)
	experimentID, err := request.URLParamUUID(r, "experimentId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	var body RerunExperimentRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	exp, err := h.experimentSvc.Rerun(r.Context(), experimentID, projectID, &body)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Created(w, exp.ToResponse())
}

func (h *Handler) SdkBatchCreateExperimentItems(w http.ResponseWriter, r *http.Request) {
	projectID := projectIDForSDK(r)
	experimentID, err := request.URLParamUUID(r, "experimentId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	var body BatchCreateExperimentItemsRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	count, err := h.experimentItemSvc.CreateBatch(r.Context(), experimentID, projectID, &body)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	h.logger.InfoContext(r.Context(), "evaluation: experiment items created (SDK)",
		"experiment_id", experimentID, "project_id", projectID, "count", count)
	response.Created(w, &CountResponse{Created: count})
}
