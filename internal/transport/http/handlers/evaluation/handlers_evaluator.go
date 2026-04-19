package evaluation

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	analyticsDomain "brokle/internal/core/domain/analytics"
	evaluationDomain "brokle/internal/core/domain/evaluation"
	"brokle/internal/transport/http/handlers/shared"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/pagination"
)

// ---- route registration ---------------------------------------------

func registerEvaluatorRoutes(api huma.API, h *handler) {
	huma.Register(api, huma.Operation{
		OperationID:   "create-evaluator",
		Method:        http.MethodPost,
		Path:          "/api/v1/projects/{projectId}/evaluators",
		Tags:          []string{"evaluators"},
		Summary:       "Create an evaluator",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.createEvaluator)

	huma.Register(api, huma.Operation{
		OperationID: "list-evaluators",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/evaluators",
		Tags:        []string{"evaluators"},
		Summary:     "List evaluators",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.listEvaluators)

	huma.Register(api, huma.Operation{
		OperationID: "get-evaluator",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/evaluators/{evaluatorId}",
		Tags:        []string{"evaluators"},
		Summary:     "Get an evaluator",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getEvaluator)

	huma.Register(api, huma.Operation{
		OperationID: "update-evaluator",
		Method:      http.MethodPut,
		Path:        "/api/v1/projects/{projectId}/evaluators/{evaluatorId}",
		Tags:        []string{"evaluators"},
		Summary:     "Update an evaluator",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.updateEvaluator)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-evaluator",
		Method:        http.MethodDelete,
		Path:          "/api/v1/projects/{projectId}/evaluators/{evaluatorId}",
		Tags:          []string{"evaluators"},
		Summary:       "Delete an evaluator",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.deleteEvaluator)

	huma.Register(api, huma.Operation{
		OperationID: "activate-evaluator",
		Method:      http.MethodPost,
		Path:        "/api/v1/projects/{projectId}/evaluators/{evaluatorId}/activate",
		Tags:        []string{"evaluators"},
		Summary:     "Activate an evaluator",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.activateEvaluator)

	huma.Register(api, huma.Operation{
		OperationID: "deactivate-evaluator",
		Method:      http.MethodPost,
		Path:        "/api/v1/projects/{projectId}/evaluators/{evaluatorId}/deactivate",
		Tags:        []string{"evaluators"},
		Summary:     "Deactivate an evaluator",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.deactivateEvaluator)

	huma.Register(api, huma.Operation{
		OperationID:   "trigger-evaluator",
		Method:        http.MethodPost,
		Path:          "/api/v1/projects/{projectId}/evaluators/{evaluatorId}/trigger",
		Tags:          []string{"evaluators"},
		Summary:       "Trigger an evaluator against matching spans",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusAccepted,
	}, h.triggerEvaluator)

	huma.Register(api, huma.Operation{
		OperationID: "test-evaluator",
		Method:      http.MethodPost,
		Path:        "/api/v1/projects/{projectId}/evaluators/{evaluatorId}/test",
		Tags:        []string{"evaluators"},
		Summary:     "Dry-run an evaluator against sample spans",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.testEvaluator)

	huma.Register(api, huma.Operation{
		OperationID: "get-evaluator-analytics",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/evaluators/{evaluatorId}/analytics",
		Tags:        []string{"evaluators"},
		Summary:     "Get evaluator analytics",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getEvaluatorAnalytics)
}

// ---- DTOs ------------------------------------------------------------

type CreateEvaluatorRequest = evaluationDomain.CreateEvaluatorRequest
type UpdateEvaluatorRequest = evaluationDomain.UpdateEvaluatorRequest

type EvaluatorOutput struct {
	Body *evaluationDomain.EvaluatorResponse
}

// ---- handlers --------------------------------------------------------

type CreateEvaluatorInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	Body      CreateEvaluatorRequest
}

func (h *handler) createEvaluator(ctx context.Context, in *CreateEvaluatorInput) (*EvaluatorOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	body := in.Body
	ev, err := h.evaluatorSvc.Create(ctx, projectID, userIDPtr(ctx), &body)
	if err != nil {
		return nil, err
	}
	return &EvaluatorOutput{Body: ev.ToResponse()}, nil
}

type ListEvaluatorsInput struct {
	ProjectID  string `path:"projectId" format:"uuid"`
	Page       int    `query:"page" required:"false" minimum:"1"`
	Limit      int    `query:"limit" required:"false"`
	SortBy     string `query:"sort_by" required:"false"`
	SortDir    string `query:"sort_dir" required:"false" enum:"asc,desc"`
	Status     string `query:"status" required:"false"`
	ScorerType string `query:"scorer_type" required:"false"`
	Search     string `query:"search" required:"false"`
}
type EvaluatorListOutput struct {
	Body pageList[*evaluationDomain.EvaluatorResponse]
}

func (h *handler) listEvaluators(ctx context.Context, in *ListEvaluatorsInput) (*EvaluatorListOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	allowedSortFields := []string{"name", "status", "sampling_rate", "created_at", "updated_at"}
	sortBy, err := pagination.ValidateSortField(in.SortBy, allowedSortFields)
	if err != nil {
		return nil, appErrors.NewValidationError("sort_by", err.Error())
	}
	params := pagination.Params{
		Page:    in.Page,
		Limit:   in.Limit,
		SortBy:  sortBy,
		SortDir: in.SortDir,
	}
	if in.SortDir != "" && in.SortDir != "asc" && in.SortDir != "desc" {
		return nil, appErrors.NewValidationError("sort_dir", "must be 'asc' or 'desc'")
	}
	params.SetDefaults("created_at")

	var filter evaluationDomain.EvaluatorFilter
	if in.Status != "" {
		s := evaluationDomain.EvaluatorStatus(in.Status)
		filter.Status = &s
	}
	if in.ScorerType != "" {
		st := evaluationDomain.ScorerType(in.ScorerType)
		filter.ScorerType = &st
	}
	if in.Search != "" {
		s := in.Search
		filter.Search = &s
	}

	evs, total, err := h.evaluatorSvc.List(ctx, projectID, &filter, params)
	if err != nil {
		return nil, err
	}
	out := make([]*evaluationDomain.EvaluatorResponse, len(evs))
	for i, e := range evs {
		out[i] = e.ToResponse()
	}
	return &EvaluatorListOutput{Body: pageList[*evaluationDomain.EvaluatorResponse]{
		Data: out, Total: total, Page: params.Page, Limit: params.Limit,
	}}, nil
}

type GetEvaluatorInput struct {
	ProjectID   string `path:"projectId" format:"uuid"`
	EvaluatorID string `path:"evaluatorId" format:"uuid"`
}

func (h *handler) getEvaluator(ctx context.Context, in *GetEvaluatorInput) (*EvaluatorOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	evaluatorID, err := parseEvaluatorID(in.EvaluatorID)
	if err != nil {
		return nil, err
	}
	ev, err := h.evaluatorSvc.GetByID(ctx, evaluatorID, projectID)
	if err != nil {
		return nil, err
	}
	return &EvaluatorOutput{Body: ev.ToResponse()}, nil
}

type UpdateEvaluatorInput struct {
	ProjectID   string `path:"projectId" format:"uuid"`
	EvaluatorID string `path:"evaluatorId" format:"uuid"`
	Body        UpdateEvaluatorRequest
}

func (h *handler) updateEvaluator(ctx context.Context, in *UpdateEvaluatorInput) (*EvaluatorOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	evaluatorID, err := parseEvaluatorID(in.EvaluatorID)
	if err != nil {
		return nil, err
	}
	body := in.Body
	ev, err := h.evaluatorSvc.Update(ctx, evaluatorID, projectID, &body)
	if err != nil {
		return nil, err
	}
	return &EvaluatorOutput{Body: ev.ToResponse()}, nil
}

type DeleteEvaluatorInput = GetEvaluatorInput

func (h *handler) deleteEvaluator(ctx context.Context, in *DeleteEvaluatorInput) (*EmptyOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	evaluatorID, err := parseEvaluatorID(in.EvaluatorID)
	if err != nil {
		return nil, err
	}
	if err := h.evaluatorSvc.Delete(ctx, evaluatorID, projectID); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
}

// MessageResponse is the legacy status-message shape retained for the
// activate/deactivate endpoints to avoid breaking the dashboard SPA.
type MessageResponse struct {
	Message string `json:"message"`
}

type MessageOutput struct {
	Body *MessageResponse
}

func (h *handler) activateEvaluator(ctx context.Context, in *GetEvaluatorInput) (*MessageOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	evaluatorID, err := parseEvaluatorID(in.EvaluatorID)
	if err != nil {
		return nil, err
	}
	if err := h.evaluatorSvc.Activate(ctx, evaluatorID, projectID); err != nil {
		return nil, err
	}
	return &MessageOutput{Body: &MessageResponse{Message: "evaluator activated"}}, nil
}

func (h *handler) deactivateEvaluator(ctx context.Context, in *GetEvaluatorInput) (*MessageOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	evaluatorID, err := parseEvaluatorID(in.EvaluatorID)
	if err != nil {
		return nil, err
	}
	if err := h.evaluatorSvc.Deactivate(ctx, evaluatorID, projectID); err != nil {
		return nil, err
	}
	return &MessageOutput{Body: &MessageResponse{Message: "evaluator deactivated"}}, nil
}

type TriggerEvaluatorInput struct {
	ProjectID   string                          `path:"projectId" format:"uuid"`
	EvaluatorID string                          `path:"evaluatorId" format:"uuid"`
	Body        evaluationDomain.TriggerOptions `doc:"All fields optional; empty object acceptable"`
}
type TriggerEvaluatorOutput struct {
	Body *evaluationDomain.TriggerResponse
}

func (h *handler) triggerEvaluator(ctx context.Context, in *TriggerEvaluatorInput) (*TriggerEvaluatorOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	evaluatorID, err := parseEvaluatorID(in.EvaluatorID)
	if err != nil {
		return nil, err
	}
	opts := in.Body
	res, err := h.evaluatorSvc.TriggerEvaluator(ctx, evaluatorID, projectID, &opts)
	if err != nil {
		return nil, err
	}
	return &TriggerEvaluatorOutput{Body: res}, nil
}

type TestEvaluatorInput struct {
	ProjectID   string                                `path:"projectId" format:"uuid"`
	EvaluatorID string                                `path:"evaluatorId" format:"uuid"`
	Body        evaluationDomain.TestEvaluatorRequest `doc:"All fields optional; empty object acceptable"`
}
type TestEvaluatorOutput struct {
	Body *evaluationDomain.TestEvaluatorResponse
}

func (h *handler) testEvaluator(ctx context.Context, in *TestEvaluatorInput) (*TestEvaluatorOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	evaluatorID, err := parseEvaluatorID(in.EvaluatorID)
	if err != nil {
		return nil, err
	}
	req := in.Body
	res, err := h.evaluatorSvc.TestEvaluator(ctx, evaluatorID, projectID, &req)
	if err != nil {
		return nil, err
	}
	return &TestEvaluatorOutput{Body: res}, nil
}

type GetEvaluatorAnalyticsInput struct {
	ProjectID     string `path:"projectId" format:"uuid"`
	EvaluatorID   string `path:"evaluatorId" format:"uuid"`
	Period        string `query:"period" required:"false" doc:"24h, 7d, 30d"`
	FromTimestamp string `query:"from_timestamp" required:"false" doc:"RFC3339"`
	ToTimestamp   string `query:"to_timestamp" required:"false" doc:"RFC3339"`
}
type EvaluatorAnalyticsOutput struct {
	Body *evaluationDomain.EvaluatorAnalyticsResponse
}

func (h *handler) getEvaluatorAnalytics(ctx context.Context, in *GetEvaluatorAnalyticsInput) (*EvaluatorAnalyticsOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	evaluatorID, err := parseEvaluatorID(in.EvaluatorID)
	if err != nil {
		return nil, err
	}
	period := in.Period
	if period == "" {
		period = "7d"
	}
	fromTime, toTime, err := shared.ParseTimeRange(
		in.FromTimestamp,
		in.ToTimestamp,
		period,
		analyticsDomain.TimeRange7Days,
	)
	if err != nil {
		return nil, err
	}
	params := &evaluationDomain.EvaluatorAnalyticsParams{
		ProjectID:   projectID,
		EvaluatorID: evaluatorID,
		Period:      period,
		From:        &fromTime,
		To:          &toTime,
	}
	res, err := h.evaluatorSvc.GetAnalytics(ctx, evaluatorID, projectID, params)
	if err != nil {
		return nil, err
	}
	return &EvaluatorAnalyticsOutput{Body: res}, nil
}
