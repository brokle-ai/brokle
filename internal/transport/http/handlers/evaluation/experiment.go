package evaluation

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	evaluationDomain "brokle/internal/core/domain/evaluation"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
)

// ---- dashboard routes -----------------------------------------------

func registerExperimentRoutes(api huma.API, h *handler) {
	huma.Register(api, huma.Operation{
		OperationID:   "create-experiment",
		Method:        http.MethodPost,
		Path:          "/api/v1/projects/{projectId}/experiments",
		Tags:          []string{"experiments"},
		Summary:       "Create an experiment",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.dashCreateExperiment)

	huma.Register(api, huma.Operation{
		OperationID: "list-experiments",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/experiments",
		Tags:        []string{"experiments"},
		Summary:     "List experiments",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.dashListExperiments)

	huma.Register(api, huma.Operation{
		OperationID: "compare-experiments",
		Method:      http.MethodPost,
		Path:        "/api/v1/projects/{projectId}/experiments/compare",
		Tags:        []string{"experiments"},
		Summary:     "Compare experiments",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.dashCompareExperiments)

	huma.Register(api, huma.Operation{
		OperationID: "get-experiment",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/experiments/{experimentId}",
		Tags:        []string{"experiments"},
		Summary:     "Get an experiment",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.dashGetExperiment)

	huma.Register(api, huma.Operation{
		OperationID: "update-experiment",
		Method:      http.MethodPut,
		Path:        "/api/v1/projects/{projectId}/experiments/{experimentId}",
		Tags:        []string{"experiments"},
		Summary:     "Update an experiment",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.dashUpdateExperiment)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-experiment",
		Method:        http.MethodDelete,
		Path:          "/api/v1/projects/{projectId}/experiments/{experimentId}",
		Tags:          []string{"experiments"},
		Summary:       "Delete an experiment",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.dashDeleteExperiment)

	huma.Register(api, huma.Operation{
		OperationID:   "rerun-experiment",
		Method:        http.MethodPost,
		Path:          "/api/v1/projects/{projectId}/experiments/{experimentId}/rerun",
		Tags:          []string{"experiments"},
		Summary:       "Re-run an experiment",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.dashRerunExperiment)

	huma.Register(api, huma.Operation{
		OperationID: "get-experiment-progress",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/experiments/{experimentId}/progress",
		Tags:        []string{"experiments"},
		Summary:     "Get experiment progress",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.dashGetExperimentProgress)

	huma.Register(api, huma.Operation{
		OperationID: "get-experiment-metrics",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/experiments/{experimentId}/metrics",
		Tags:        []string{"experiments"},
		Summary:     "Get experiment metrics",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.dashGetExperimentMetrics)

	huma.Register(api, huma.Operation{
		OperationID: "list-experiment-items",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/experiments/{experimentId}/items",
		Tags:        []string{"experiment-items"},
		Summary:     "List experiment items",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.dashListExperimentItems)
}

// ---- SDK routes ------------------------------------------------------

func registerSDKExperimentRoutes(api huma.API, h *handler) {
	huma.Register(api, huma.Operation{
		OperationID:   "sdk-create-experiment",
		Method:        http.MethodPost,
		Path:          "/v1/experiments",
		Tags:          []string{"SDK - experiments"},
		Summary:       "Create an experiment (SDK)",
		Security:      []map[string][]string{{"apiKey": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.sdkCreateExperiment)

	huma.Register(api, huma.Operation{
		OperationID: "sdk-list-experiments",
		Method:      http.MethodGet,
		Path:        "/v1/experiments",
		Tags:        []string{"SDK - experiments"},
		Summary:     "List experiments (SDK)",
		Security:    []map[string][]string{{"apiKey": {}}},
	}, h.sdkListExperiments)

	huma.Register(api, huma.Operation{
		OperationID: "sdk-compare-experiments",
		Method:      http.MethodPost,
		Path:        "/v1/experiments/compare",
		Tags:        []string{"SDK - experiments"},
		Summary:     "Compare experiments (SDK)",
		Security:    []map[string][]string{{"apiKey": {}}},
	}, h.sdkCompareExperiments)

	huma.Register(api, huma.Operation{
		OperationID: "sdk-get-experiment",
		Method:      http.MethodGet,
		Path:        "/v1/experiments/{experimentId}",
		Tags:        []string{"SDK - experiments"},
		Summary:     "Get an experiment (SDK)",
		Security:    []map[string][]string{{"apiKey": {}}},
	}, h.sdkGetExperiment)

	huma.Register(api, huma.Operation{
		OperationID: "sdk-update-experiment",
		Method:      http.MethodPatch,
		Path:        "/v1/experiments/{experimentId}",
		Tags:        []string{"SDK - experiments"},
		Summary:     "Update an experiment (SDK)",
		Security:    []map[string][]string{{"apiKey": {}}},
	}, h.sdkUpdateExperiment)

	huma.Register(api, huma.Operation{
		OperationID:   "sdk-rerun-experiment",
		Method:        http.MethodPost,
		Path:          "/v1/experiments/{experimentId}/rerun",
		Tags:          []string{"SDK - experiments"},
		Summary:       "Re-run an experiment (SDK)",
		Security:      []map[string][]string{{"apiKey": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.sdkRerunExperiment)

	huma.Register(api, huma.Operation{
		OperationID:   "sdk-batch-create-experiment-items",
		Method:        http.MethodPost,
		Path:          "/v1/experiments/{experimentId}/items",
		Tags:          []string{"SDK - experiments"},
		Summary:       "Batch-create experiment items (SDK)",
		Security:      []map[string][]string{{"apiKey": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.sdkBatchCreateExperimentItems)
}

// Request/response DTOs live in experiment_types.go.

// ---- shared experiment op bodies ------------------------------------

func (h *handler) createExperiment(ctx context.Context, projectID uuid.UUID, body *CreateExperimentRequest) (*evaluationDomain.ExperimentResponse, error) {
	exp, err := h.experimentSvc.Create(ctx, projectID, body)
	if err != nil {
		return nil, err
	}
	h.logger.InfoContext(ctx, "evaluation: experiment created",
		"experiment_id", exp.ID, "project_id", projectID, "name", exp.Name)
	return exp.ToResponse(), nil
}

func (h *handler) listExperiments(
	ctx context.Context,
	projectID uuid.UUID,
	datasetIDStr, statusStr, search, idsStr string,
	page, limit int,
) ([]*evaluationDomain.ExperimentResponse, int64, int, int, error) {
	page, limit = normalizePagination(page, limit)

	var filter *evaluationDomain.ExperimentFilter
	if datasetIDStr != "" {
		datasetID, err := uuid.Parse(datasetIDStr)
		if err != nil {
			return nil, 0, page, limit, appErrors.NewValidationError("dataset_id", "must be a valid UUID")
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
			return nil, 0, page, limit, appErrors.NewValidationError("status", "must be pending, running, completed, failed, partial, or cancelled")
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
				return nil, 0, page, limit, appErrors.NewValidationError("ids", "invalid UUID: "+idStr)
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

func (h *handler) compareExperiments(ctx context.Context, projectID uuid.UUID, body *CompareExperimentsRequest) (*CompareExperimentsResponse, error) {
	experimentIDs := make([]uuid.UUID, len(body.ExperimentIDs))
	for i, idStr := range body.ExperimentIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return nil, appErrors.NewValidationError("experiment_ids", "invalid UUID at index "+strconv.Itoa(i))
		}
		experimentIDs[i] = id
	}

	var baselineID *uuid.UUID
	if body.BaselineID != nil {
		id, err := uuid.Parse(*body.BaselineID)
		if err != nil {
			return nil, appErrors.NewValidationError("baseline_id", "must be a valid UUID")
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

// ---- dashboard handler methods --------------------------------------

type DashCreateExperimentInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	Body      CreateExperimentRequest
}
type ExperimentOutput struct {
	Body *evaluationDomain.ExperimentResponse
}

func (h *handler) dashCreateExperiment(ctx context.Context, in *DashCreateExperimentInput) (*ExperimentOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	body := in.Body
	resp, err := h.createExperiment(ctx, projectID, &body)
	if err != nil {
		return nil, err
	}
	return &ExperimentOutput{Body: resp}, nil
}

type DashListExperimentsInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	Page      int    `query:"page" required:"false" minimum:"1"`
	Limit     int    `query:"limit" required:"false"`
	DatasetID string `query:"dataset_id" required:"false" format:"uuid"`
	Status    string `query:"status" required:"false"`
	Search    string `query:"search" required:"false"`
	IDs       string `query:"ids" required:"false" doc:"comma-separated list of experiment UUIDs"`
}
type ExperimentListOutput struct {
	Body pageList[*evaluationDomain.ExperimentResponse]
}

func (h *handler) dashListExperiments(ctx context.Context, in *DashListExperimentsInput) (*ExperimentListOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	out, total, page, limit, err := h.listExperiments(ctx, projectID, in.DatasetID, in.Status, in.Search, in.IDs, in.Page, in.Limit)
	if err != nil {
		return nil, err
	}
	return &ExperimentListOutput{Body: pageList[*evaluationDomain.ExperimentResponse]{
		Data: out, Total: total, Page: page, Limit: limit,
	}}, nil
}

type DashCompareExperimentsInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	Body      CompareExperimentsRequest
}
type CompareExperimentsOutput struct {
	Body *CompareExperimentsResponse
}

func (h *handler) dashCompareExperiments(ctx context.Context, in *DashCompareExperimentsInput) (*CompareExperimentsOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	body := in.Body
	resp, err := h.compareExperiments(ctx, projectID, &body)
	if err != nil {
		return nil, err
	}
	return &CompareExperimentsOutput{Body: resp}, nil
}

type DashGetExperimentInput struct {
	ProjectID    string `path:"projectId" format:"uuid"`
	ExperimentID string `path:"experimentId" format:"uuid"`
}

func (h *handler) dashGetExperiment(ctx context.Context, in *DashGetExperimentInput) (*ExperimentOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	experimentID, err := parseExperimentID(in.ExperimentID)
	if err != nil {
		return nil, err
	}
	exp, err := h.experimentSvc.GetByID(ctx, experimentID, projectID)
	if err != nil {
		return nil, err
	}
	return &ExperimentOutput{Body: exp.ToResponse()}, nil
}

type DashUpdateExperimentInput struct {
	ProjectID    string `path:"projectId" format:"uuid"`
	ExperimentID string `path:"experimentId" format:"uuid"`
	Body         UpdateExperimentRequest
}

func (h *handler) dashUpdateExperiment(ctx context.Context, in *DashUpdateExperimentInput) (*ExperimentOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	experimentID, err := parseExperimentID(in.ExperimentID)
	if err != nil {
		return nil, err
	}
	body := in.Body
	exp, err := h.experimentSvc.Update(ctx, experimentID, projectID, &body)
	if err != nil {
		return nil, err
	}
	h.logger.InfoContext(ctx, "evaluation: experiment updated",
		"experiment_id", experimentID, "project_id", projectID, "status", exp.Status)
	return &ExperimentOutput{Body: exp.ToResponse()}, nil
}

type DashDeleteExperimentInput = DashGetExperimentInput

func (h *handler) dashDeleteExperiment(ctx context.Context, in *DashDeleteExperimentInput) (*EmptyOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	experimentID, err := parseExperimentID(in.ExperimentID)
	if err != nil {
		return nil, err
	}
	if err := h.experimentSvc.Delete(ctx, experimentID, projectID); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
}

type DashRerunExperimentInput struct {
	ProjectID    string `path:"projectId" format:"uuid"`
	ExperimentID string `path:"experimentId" format:"uuid"`
	Body         RerunExperimentRequest
}

func (h *handler) dashRerunExperiment(ctx context.Context, in *DashRerunExperimentInput) (*ExperimentOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	experimentID, err := parseExperimentID(in.ExperimentID)
	if err != nil {
		return nil, err
	}
	body := in.Body
	exp, err := h.experimentSvc.Rerun(ctx, experimentID, projectID, &body)
	if err != nil {
		return nil, err
	}
	h.logger.InfoContext(ctx, "evaluation: experiment rerun",
		"experiment_id", exp.ID, "source_experiment_id", experimentID, "project_id", projectID)
	return &ExperimentOutput{Body: exp.ToResponse()}, nil
}

type ExperimentProgressOutput struct {
	Body *evaluationDomain.ExperimentProgressResponse
}

func (h *handler) dashGetExperimentProgress(ctx context.Context, in *DashGetExperimentInput) (*ExperimentProgressOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	experimentID, err := parseExperimentID(in.ExperimentID)
	if err != nil {
		return nil, err
	}
	progress, err := h.experimentSvc.GetProgress(ctx, experimentID, projectID)
	if err != nil {
		return nil, err
	}
	return &ExperimentProgressOutput{Body: progress}, nil
}

type ExperimentMetricsOutput struct {
	Body *evaluationDomain.ExperimentMetricsResponse
}

func (h *handler) dashGetExperimentMetrics(ctx context.Context, in *DashGetExperimentInput) (*ExperimentMetricsOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	experimentID, err := parseExperimentID(in.ExperimentID)
	if err != nil {
		return nil, err
	}
	metrics, err := h.experimentSvc.GetMetrics(ctx, projectID, experimentID)
	if err != nil {
		return nil, err
	}
	return &ExperimentMetricsOutput{Body: metrics}, nil
}

type DashListExperimentItemsInput struct {
	ProjectID    string `path:"projectId" format:"uuid"`
	ExperimentID string `path:"experimentId" format:"uuid"`
	Limit        int    `query:"limit" required:"false" minimum:"1" maximum:"100"`
	Offset       int    `query:"offset" required:"false" minimum:"0"`
}
type ExperimentItemListOutput struct {
	Body *ExperimentItemListResponse
}

func (h *handler) dashListExperimentItems(ctx context.Context, in *DashListExperimentItemsInput) (*ExperimentItemListOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	experimentID, err := parseExperimentID(in.ExperimentID)
	if err != nil {
		return nil, err
	}
	limit := in.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := in.Offset
	if offset < 0 {
		offset = 0
	}

	items, total, err := h.experimentItemSvc.List(ctx, experimentID, projectID, limit, offset)
	if err != nil {
		return nil, err
	}
	out := make([]*ExperimentItemResponse, len(items))
	for i, it := range items {
		out[i] = toExperimentItemResponse(it)
	}
	return &ExperimentItemListOutput{Body: &ExperimentItemListResponse{Items: out, Total: total}}, nil
}

// ---- SDK handler methods --------------------------------------------

type SDKCreateExperimentInput struct {
	Body CreateExperimentRequest
}

func (h *handler) sdkCreateExperiment(ctx context.Context, in *SDKCreateExperimentInput) (*ExperimentOutput, error) {
	projectID := httpctx.MustGetProjectID(ctx)
	body := in.Body
	resp, err := h.createExperiment(ctx, projectID, &body)
	if err != nil {
		return nil, err
	}
	return &ExperimentOutput{Body: resp}, nil
}

type SDKListExperimentsInput struct {
	Page      int    `query:"page" required:"false" minimum:"1"`
	Limit     int    `query:"limit" required:"false"`
	DatasetID string `query:"dataset_id" required:"false" format:"uuid"`
	Status    string `query:"status" required:"false"`
	Search    string `query:"search" required:"false"`
	IDs       string `query:"ids" required:"false"`
}

func (h *handler) sdkListExperiments(ctx context.Context, in *SDKListExperimentsInput) (*ExperimentListOutput, error) {
	projectID := httpctx.MustGetProjectID(ctx)
	out, total, page, limit, err := h.listExperiments(ctx, projectID, in.DatasetID, in.Status, in.Search, in.IDs, in.Page, in.Limit)
	if err != nil {
		return nil, err
	}
	return &ExperimentListOutput{Body: pageList[*evaluationDomain.ExperimentResponse]{
		Data: out, Total: total, Page: page, Limit: limit,
	}}, nil
}

type SDKCompareExperimentsInput struct {
	Body CompareExperimentsRequest
}

func (h *handler) sdkCompareExperiments(ctx context.Context, in *SDKCompareExperimentsInput) (*CompareExperimentsOutput, error) {
	projectID := httpctx.MustGetProjectID(ctx)
	body := in.Body
	resp, err := h.compareExperiments(ctx, projectID, &body)
	if err != nil {
		return nil, err
	}
	return &CompareExperimentsOutput{Body: resp}, nil
}

type SDKExperimentInput struct {
	ExperimentID string `path:"experimentId" format:"uuid"`
}

func (h *handler) sdkGetExperiment(ctx context.Context, in *SDKExperimentInput) (*ExperimentOutput, error) {
	projectID := httpctx.MustGetProjectID(ctx)
	experimentID, err := parseExperimentID(in.ExperimentID)
	if err != nil {
		return nil, err
	}
	exp, err := h.experimentSvc.GetByID(ctx, experimentID, projectID)
	if err != nil {
		return nil, err
	}
	return &ExperimentOutput{Body: exp.ToResponse()}, nil
}

type SDKUpdateExperimentInput struct {
	ExperimentID string `path:"experimentId" format:"uuid"`
	Body         UpdateExperimentRequest
}

func (h *handler) sdkUpdateExperiment(ctx context.Context, in *SDKUpdateExperimentInput) (*ExperimentOutput, error) {
	projectID := httpctx.MustGetProjectID(ctx)
	experimentID, err := parseExperimentID(in.ExperimentID)
	if err != nil {
		return nil, err
	}
	body := in.Body
	exp, err := h.experimentSvc.Update(ctx, experimentID, projectID, &body)
	if err != nil {
		return nil, err
	}
	return &ExperimentOutput{Body: exp.ToResponse()}, nil
}

type SDKRerunExperimentInput struct {
	ExperimentID string `path:"experimentId" format:"uuid"`
	Body         RerunExperimentRequest
}

func (h *handler) sdkRerunExperiment(ctx context.Context, in *SDKRerunExperimentInput) (*ExperimentOutput, error) {
	projectID := httpctx.MustGetProjectID(ctx)
	experimentID, err := parseExperimentID(in.ExperimentID)
	if err != nil {
		return nil, err
	}
	body := in.Body
	exp, err := h.experimentSvc.Rerun(ctx, experimentID, projectID, &body)
	if err != nil {
		return nil, err
	}
	return &ExperimentOutput{Body: exp.ToResponse()}, nil
}

type SDKBatchCreateExperimentItemsInput struct {
	ExperimentID string `path:"experimentId" format:"uuid"`
	Body         BatchCreateExperimentItemsRequest
}

func (h *handler) sdkBatchCreateExperimentItems(ctx context.Context, in *SDKBatchCreateExperimentItemsInput) (*CountOutput, error) {
	projectID := httpctx.MustGetProjectID(ctx)
	experimentID, err := parseExperimentID(in.ExperimentID)
	if err != nil {
		return nil, err
	}
	body := in.Body
	count, err := h.experimentItemSvc.CreateBatch(ctx, experimentID, projectID, &body)
	if err != nil {
		return nil, err
	}
	h.logger.InfoContext(ctx, "evaluation: experiment items created (SDK)",
		"experiment_id", experimentID, "project_id", projectID, "count", count)
	return &CountOutput{Body: &CountResponse{Created: count}}, nil
}
