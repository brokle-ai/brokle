package evaluation

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	evaluationDomain "brokle/internal/core/domain/evaluation"
)

// registerScoreConfigRoutes wires the dashboard-plane score-config CRUD
// operations. Called from RegisterRoutes in handlers.go.
func registerScoreConfigRoutes(api huma.API, h *handler) {
	huma.Register(api, huma.Operation{
		OperationID:   "create-score-config",
		Method:        http.MethodPost,
		Path:          "/api/v1/projects/{projectId}/score-configs",
		Tags:          []string{"score-configs"},
		Summary:       "Create a score config",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.createScoreConfig)

	huma.Register(api, huma.Operation{
		OperationID: "list-score-configs",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/score-configs",
		Tags:        []string{"score-configs"},
		Summary:     "List score configs",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.listScoreConfigs)

	huma.Register(api, huma.Operation{
		OperationID: "get-score-config",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/score-configs/{configId}",
		Tags:        []string{"score-configs"},
		Summary:     "Get a score config",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getScoreConfig)

	huma.Register(api, huma.Operation{
		OperationID: "update-score-config",
		Method:      http.MethodPut,
		Path:        "/api/v1/projects/{projectId}/score-configs/{configId}",
		Tags:        []string{"score-configs"},
		Summary:     "Update a score config",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.updateScoreConfig)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-score-config",
		Method:        http.MethodDelete,
		Path:          "/api/v1/projects/{projectId}/score-configs/{configId}",
		Tags:          []string{"score-configs"},
		Summary:       "Delete a score config",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.deleteScoreConfig)
}

func (h *handler) createScoreConfig(ctx context.Context, in *CreateScoreConfigInput) (*ScoreConfigOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	domainReq := &evaluationDomain.CreateScoreConfigRequest{
		Name:        in.Body.Name,
		Description: in.Body.Description,
		Type:        evaluationDomain.ScoreType(in.Body.Type),
		MinValue:    in.Body.MinValue,
		MaxValue:    in.Body.MaxValue,
		Categories:  in.Body.Categories,
		Metadata:    in.Body.Metadata,
	}
	cfg, err := h.scoreConfigSvc.Create(ctx, projectID, domainReq)
	if err != nil {
		return nil, err
	}
	return &ScoreConfigOutput{Body: cfg.ToResponse()}, nil
}

func (h *handler) listScoreConfigs(ctx context.Context, in *ListScoreConfigsInput) (*ListScoreConfigsOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	page, limit := normalizePagination(in.Page, in.Limit)
	cfgs, total, err := h.scoreConfigSvc.List(ctx, projectID, page, limit)
	if err != nil {
		return nil, err
	}
	out := make([]*evaluationDomain.ScoreConfigResponse, len(cfgs))
	for i, c := range cfgs {
		out[i] = c.ToResponse()
	}
	return &ListScoreConfigsOutput{Body: pageList[*evaluationDomain.ScoreConfigResponse]{
		Data: out, Total: total, Page: page, Limit: limit,
	}}, nil
}

func (h *handler) getScoreConfig(ctx context.Context, in *GetScoreConfigInput) (*ScoreConfigOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	configID, err := parseScoreConfigID(in.ConfigID)
	if err != nil {
		return nil, err
	}
	cfg, err := h.scoreConfigSvc.GetByID(ctx, configID, projectID)
	if err != nil {
		return nil, err
	}
	return &ScoreConfigOutput{Body: cfg.ToResponse()}, nil
}

func (h *handler) updateScoreConfig(ctx context.Context, in *UpdateScoreConfigInput) (*ScoreConfigOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	configID, err := parseScoreConfigID(in.ConfigID)
	if err != nil {
		return nil, err
	}

	var scoreType *evaluationDomain.ScoreType
	if in.Body.Type != nil {
		st := evaluationDomain.ScoreType(*in.Body.Type)
		scoreType = &st
	}
	domainReq := &evaluationDomain.UpdateScoreConfigRequest{
		Name:        in.Body.Name,
		Description: in.Body.Description,
		Type:        scoreType,
		MinValue:    in.Body.MinValue,
		MaxValue:    in.Body.MaxValue,
		Categories:  in.Body.Categories,
		Metadata:    in.Body.Metadata,
	}
	cfg, err := h.scoreConfigSvc.Update(ctx, configID, projectID, domainReq)
	if err != nil {
		return nil, err
	}
	return &ScoreConfigOutput{Body: cfg.ToResponse()}, nil
}

func (h *handler) deleteScoreConfig(ctx context.Context, in *DeleteScoreConfigInput) (*EmptyOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	configID, err := parseScoreConfigID(in.ConfigID)
	if err != nil {
		return nil, err
	}
	if err := h.scoreConfigSvc.Delete(ctx, configID, projectID); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
}
