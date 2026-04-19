// Package observability exposes observability operations on both
// planes:
//
//   - Dashboard plane (apiAdmin, RequireAuth) — traces browse, spans
//     list/get/delete, scores CRUD, sessions list, filter-preset CRUD.
//   - SDK plane (apiPublic, RequireSDKAuth) — span query + filter
//     validation (see handlers_sdk.go).
//   - OTLP ingestion (/v1/traces, /v1/logs, /v1/metrics) is HUMA-EXEMPT
//     and mounted as plain chi handlers — see otlp.go.
package observability

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"brokle/internal/core/domain/observability"
	obsServices "brokle/internal/core/services/observability"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/pagination"
	"brokle/pkg/uid"
)

type dashboardHandler struct {
	traces          *obsServices.TraceService
	scores          *obsServices.ScoreService
	scoreAnalytics  *obsServices.ScoreAnalyticsService
	filterPresets   *obsServices.FilterPresetService
	logger          *slog.Logger
}

// RegisterRoutes wires the dashboard-plane observability operations onto
// apiAdmin. All routes require RequireAuth. Routes scoped to a specific
// project carry {projectId}; routes on /traces or /spans expect the
// RequireProjectAccess middleware to pin the project via
// httpctx.WithProjectID (access via query param + middleware).
func RegisterRoutes(
	api huma.API,
	traces *obsServices.TraceService,
	scores *obsServices.ScoreService,
	scoreAnalytics *obsServices.ScoreAnalyticsService,
	filterPresets *obsServices.FilterPresetService,
	logger *slog.Logger,
) {
	h := &dashboardHandler{
		traces:         traces,
		scores:         scores,
		scoreAnalytics: scoreAnalytics,
		filterPresets:  filterPresets,
		logger:         logger,
	}

	// ---- traces ---------------------------------------------------
	huma.Register(api, huma.Operation{
		OperationID: "list-traces",
		Method:      http.MethodGet,
		Path:        "/api/v1/traces",
		Tags:        []string{"traces"},
		Summary:     "List traces for a project",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.listTraces)

	huma.Register(api, huma.Operation{
		OperationID: "get-trace-filter-options",
		Method:      http.MethodGet,
		Path:        "/api/v1/traces/filter-options",
		Tags:        []string{"traces"},
		Summary:     "Get available filter options for traces",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getTraceFilterOptions)

	huma.Register(api, huma.Operation{
		OperationID: "discover-trace-attributes",
		Method:      http.MethodGet,
		Path:        "/api/v1/traces/attributes",
		Tags:        []string{"traces"},
		Summary:     "Discover attribute keys from trace data",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.discoverAttributes)

	huma.Register(api, huma.Operation{
		OperationID: "get-trace",
		Method:      http.MethodGet,
		Path:        "/api/v1/traces/{id}",
		Tags:        []string{"traces"},
		Summary:     "Get trace by ID",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getTrace)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-trace",
		Method:        http.MethodDelete,
		Path:          "/api/v1/traces/{id}",
		Tags:          []string{"traces"},
		Summary:       "Delete a trace",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.deleteTrace)

	huma.Register(api, huma.Operation{
		OperationID: "get-trace-spans",
		Method:      http.MethodGet,
		Path:        "/api/v1/traces/{id}/spans",
		Tags:        []string{"traces"},
		Summary:     "Get spans for a trace",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getTraceSpans)

	huma.Register(api, huma.Operation{
		OperationID: "get-trace-scores",
		Method:      http.MethodGet,
		Path:        "/api/v1/traces/{id}/scores",
		Tags:        []string{"traces", "scores"},
		Summary:     "List scores for a trace",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getTraceScores)

	huma.Register(api, huma.Operation{
		OperationID:   "create-trace-score",
		Method:        http.MethodPost,
		Path:          "/api/v1/traces/{id}/scores",
		Tags:          []string{"traces", "scores"},
		Summary:       "Create annotation score for a trace",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.createTraceScore)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-trace-score",
		Method:        http.MethodDelete,
		Path:          "/api/v1/traces/{id}/scores/{scoreId}",
		Tags:          []string{"traces", "scores"},
		Summary:       "Delete annotation score",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.deleteTraceScore)

	huma.Register(api, huma.Operation{
		OperationID: "update-trace-tags",
		Method:      http.MethodPut,
		Path:        "/api/v1/traces/{id}/tags",
		Tags:        []string{"traces"},
		Summary:     "Update trace tags",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.updateTraceTags)

	huma.Register(api, huma.Operation{
		OperationID: "update-trace-bookmark",
		Method:      http.MethodPut,
		Path:        "/api/v1/traces/{id}/bookmark",
		Tags:        []string{"traces"},
		Summary:     "Update trace bookmark status",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.updateTraceBookmark)

	// ---- spans ----------------------------------------------------
	huma.Register(api, huma.Operation{
		OperationID: "list-spans",
		Method:      http.MethodGet,
		Path:        "/api/v1/spans",
		Tags:        []string{"spans"},
		Summary:     "List spans",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.listSpans)

	huma.Register(api, huma.Operation{
		OperationID: "get-span",
		Method:      http.MethodGet,
		Path:        "/api/v1/spans/{id}",
		Tags:        []string{"spans"},
		Summary:     "Get span by ID",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getSpan)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-span",
		Method:        http.MethodDelete,
		Path:          "/api/v1/spans/{id}",
		Tags:          []string{"spans"},
		Summary:       "Delete a span",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.deleteSpan)

	// ---- scores (global + project scoped) -------------------------
	huma.Register(api, huma.Operation{
		OperationID: "list-scores",
		Method:      http.MethodGet,
		Path:        "/api/v1/scores",
		Tags:        []string{"scores"},
		Summary:     "List quality scores",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.listScores)

	huma.Register(api, huma.Operation{
		OperationID: "get-score",
		Method:      http.MethodGet,
		Path:        "/api/v1/scores/{id}",
		Tags:        []string{"scores"},
		Summary:     "Get score by ID",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getScore)

	huma.Register(api, huma.Operation{
		OperationID: "update-score",
		Method:      http.MethodPut,
		Path:        "/api/v1/scores/{id}",
		Tags:        []string{"scores"},
		Summary:     "Update score by ID",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.updateScore)

	huma.Register(api, huma.Operation{
		OperationID: "list-project-scores",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/scores",
		Tags:        []string{"scores"},
		Summary:     "List quality scores for a project",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.listProjectScores)

	huma.Register(api, huma.Operation{
		OperationID: "get-score-analytics",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/scores/analytics",
		Tags:        []string{"scores"},
		Summary:     "Get score analytics for a project",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getScoreAnalytics)

	huma.Register(api, huma.Operation{
		OperationID: "get-score-names",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/scores/names",
		Tags:        []string{"scores"},
		Summary:     "Get distinct score names for a project",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getScoreNames)

	// ---- sessions -------------------------------------------------
	huma.Register(api, huma.Operation{
		OperationID: "list-sessions",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/sessions",
		Tags:        []string{"sessions"},
		Summary:     "List sessions for a project",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.listSessions)

	// ---- filter presets -------------------------------------------
	huma.Register(api, huma.Operation{
		OperationID:   "create-filter-preset",
		Method:        http.MethodPost,
		Path:          "/api/v1/projects/{projectId}/filter-presets",
		Tags:          []string{"filter-presets"},
		Summary:       "Create a filter preset",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.createFilterPreset)

	huma.Register(api, huma.Operation{
		OperationID: "list-filter-presets",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/filter-presets",
		Tags:        []string{"filter-presets"},
		Summary:     "List filter presets for a project",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.listFilterPresets)

	huma.Register(api, huma.Operation{
		OperationID: "get-filter-preset",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/filter-presets/{id}",
		Tags:        []string{"filter-presets"},
		Summary:     "Get filter preset",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getFilterPreset)

	huma.Register(api, huma.Operation{
		OperationID: "update-filter-preset",
		Method:      http.MethodPatch,
		Path:        "/api/v1/projects/{projectId}/filter-presets/{id}",
		Tags:        []string{"filter-presets"},
		Summary:     "Update filter preset",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.updateFilterPreset)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-filter-preset",
		Method:        http.MethodDelete,
		Path:          "/api/v1/projects/{projectId}/filter-presets/{id}",
		Tags:          []string{"filter-presets"},
		Summary:       "Delete filter preset",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.deleteFilterPreset)
}

// ---- shared helpers -----------------------------------------------

func parseUUIDParam(v, field, humanField string) (uuid.UUID, error) {
	id, err := uuid.Parse(v)
	if err != nil {
		return uuid.Nil, appErrors.NewValidationError("Invalid "+humanField, field+" must be a valid UUID")
	}
	return id, nil
}

func paginationParams(page, limit int, sortBy, sortDir string) pagination.Params {
	p := pagination.Params{Page: page, Limit: limit, SortBy: sortBy, SortDir: sortDir}
	p.SetDefaults("")
	return p
}

type paginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

func newPaginationMeta(p pagination.Params, total int64) paginationMeta {
	pages := 0
	if p.Limit > 0 {
		pages = int((total + int64(p.Limit) - 1) / int64(p.Limit))
	}
	return paginationMeta{Page: p.Page, Limit: p.Limit, Total: total, TotalPages: pages}
}

// ==================================================================
// TRACES
// ==================================================================

type ListTracesInput struct {
	SessionID    string  `query:"session_id" required:"false"`
	UserID       string  `query:"user_id" required:"false"`
	ServiceName  string  `query:"service_name" required:"false"`
	ModelName    string  `query:"model_name" required:"false"`
	ProviderName string  `query:"provider_name" required:"false"`
	MinCost      float64 `query:"min_cost" required:"false"`
	MaxCost      float64 `query:"max_cost" required:"false"`
	MinTokens    int64   `query:"min_tokens" required:"false"`
	MaxTokens    int64   `query:"max_tokens" required:"false"`
	MinDuration  int64   `query:"min_duration" required:"false"`
	MaxDuration  int64   `query:"max_duration" required:"false"`
	HasError     *bool   `query:"has_error" required:"false"`
	Search       string  `query:"search" required:"false"`
	SearchType   string  `query:"search_type" required:"false"`
	Status       string  `query:"status" required:"false" doc:"Comma-separated status filter: ok,error,unset"`
	StatusNot    string  `query:"status_not" required:"false" doc:"Comma-separated status exclusion"`
	StartTime    int64   `query:"start_time" required:"false" doc:"Unix timestamp seconds"`
	EndTime      int64   `query:"end_time" required:"false" doc:"Unix timestamp seconds"`
	Page         int     `query:"page" required:"false" minimum:"1"`
	Limit        int     `query:"limit" required:"false"`
	SortBy       string  `query:"sort_by" required:"false"`
	SortDir      string  `query:"sort_dir" required:"false" enum:"asc,desc"`
}

type ListTracesOutput struct {
	Body listTracesResponse
}

type listTracesResponse struct {
	Data       []*observability.TraceSummary `json:"data"`
	Pagination paginationMeta                `json:"pagination"`
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	out := []string{}
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ',' {
			if start < i {
				seg := s[start:i]
				// trim spaces
				for len(seg) > 0 && seg[0] == ' ' {
					seg = seg[1:]
				}
				for len(seg) > 0 && seg[len(seg)-1] == ' ' {
					seg = seg[:len(seg)-1]
				}
				if seg != "" {
					out = append(out, seg)
				}
			}
			start = i + 1
		}
	}
	return out
}

func validateStatusList(values []string, field string) error {
	for _, v := range values {
		if v != "ok" && v != "error" && v != "unset" {
			return appErrors.NewValidationError("Invalid "+field+" value", field+" must be one of: ok, error, unset (got: "+v+")")
		}
	}
	return nil
}

func (h *dashboardHandler) listTraces(ctx context.Context, in *ListTracesInput) (*ListTracesOutput, error) {
	projectID := httpctx.MustGetProjectID(ctx)

	filter := &observability.TraceFilter{ProjectID: projectID}

	if in.SessionID != "" {
		filter.SessionID = &in.SessionID
	}
	if in.UserID != "" {
		filter.UserID = &in.UserID
	}
	if in.ServiceName != "" {
		filter.ServiceName = &in.ServiceName
	}
	if in.ModelName != "" {
		filter.ModelName = &in.ModelName
	}
	if in.ProviderName != "" {
		filter.ProviderName = &in.ProviderName
	}
	if in.MinCost != 0 {
		v := in.MinCost
		filter.MinCost = &v
	}
	if in.MaxCost != 0 {
		v := in.MaxCost
		filter.MaxCost = &v
	}
	if in.MinTokens != 0 {
		v := in.MinTokens
		filter.MinTokens = &v
	}
	if in.MaxTokens != 0 {
		v := in.MaxTokens
		filter.MaxTokens = &v
	}
	if in.MinDuration != 0 {
		v := in.MinDuration
		filter.MinDuration = &v
	}
	if in.MaxDuration != 0 {
		v := in.MaxDuration
		filter.MaxDuration = &v
	}
	if in.HasError != nil {
		filter.HasError = in.HasError
	}
	if in.Search != "" {
		filter.Search = &in.Search
	}
	if in.SearchType != "" {
		filter.SearchType = &in.SearchType
	}
	if in.Status != "" {
		filter.Statuses = splitCSV(in.Status)
		if err := validateStatusList(filter.Statuses, "status"); err != nil {
			return nil, err
		}
	}
	if in.StatusNot != "" {
		filter.StatusesNot = splitCSV(in.StatusNot)
		if err := validateStatusList(filter.StatusesNot, "status_not"); err != nil {
			return nil, err
		}
	}
	if in.StartTime != 0 {
		ts := time.Unix(in.StartTime, 0)
		filter.StartTime = &ts
	}
	if in.EndTime != 0 {
		ts := time.Unix(in.EndTime, 0)
		filter.EndTime = &ts
	}

	params := paginationParams(in.Page, in.Limit, in.SortBy, in.SortDir)
	filter.Params = params

	traces, err := h.traces.ListTraces(ctx, filter)
	if err != nil {
		h.logger.ErrorContext(ctx, "observability: list traces failed", "error", err)
		return nil, err
	}
	total, err := h.traces.CountTraces(ctx, filter)
	if err != nil {
		h.logger.ErrorContext(ctx, "observability: count traces failed", "error", err)
		return nil, err
	}

	return &ListTracesOutput{Body: listTracesResponse{
		Data:       traces,
		Pagination: newPaginationMeta(params, total),
	}}, nil
}

type GetTraceInput struct {
	ID string `path:"id" doc:"Trace ID (W3C hex)"`
}

type GetTraceOutput struct {
	Body *observability.TraceSummary
}

func (h *dashboardHandler) getTrace(ctx context.Context, in *GetTraceInput) (*GetTraceOutput, error) {
	if in.ID == "" {
		return nil, appErrors.NewValidationError("Missing trace ID", "id is required")
	}
	summary, err := h.traces.GetTrace(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	return &GetTraceOutput{Body: summary}, nil
}

type DeleteTraceInput struct {
	ID string `path:"id"`
}
type DeleteTraceOutput struct{}

func (h *dashboardHandler) deleteTrace(ctx context.Context, in *DeleteTraceInput) (*DeleteTraceOutput, error) {
	if in.ID == "" {
		return nil, appErrors.NewValidationError("Missing trace ID", "id is required")
	}
	if err := h.traces.DeleteTrace(ctx, in.ID); err != nil {
		return nil, err
	}
	return &DeleteTraceOutput{}, nil
}

type GetTraceSpansInput struct {
	ID string `path:"id"`
}
type GetTraceSpansOutput struct {
	Body []*observability.Span
}

func (h *dashboardHandler) getTraceSpans(ctx context.Context, in *GetTraceSpansInput) (*GetTraceSpansOutput, error) {
	if in.ID == "" {
		return nil, appErrors.NewValidationError("Missing trace ID", "id is required")
	}
	spans, err := h.traces.GetTraceSpans(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	return &GetTraceSpansOutput{Body: spans}, nil
}

type GetTraceScoresInput struct {
	ID        string `path:"id"`
	ProjectID string `query:"project_id" required:"true" format:"uuid"`
}
type GetTraceScoresOutput struct {
	Body []*AnnotationResponse
}

func (h *dashboardHandler) getTraceScores(ctx context.Context, in *GetTraceScoresInput) (*GetTraceScoresOutput, error) {
	if in.ID == "" {
		return nil, appErrors.NewValidationError("Missing trace ID", "id is required")
	}
	if _, err := parseUUIDParam(in.ProjectID, "project_id", "project ID"); err != nil {
		return nil, err
	}

	scores, err := h.scores.GetScoresByTraceID(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	out := make([]*AnnotationResponse, 0, len(scores))
	for _, s := range scores {
		out = append(out, toAnnotationResponse(s))
	}
	return &GetTraceScoresOutput{Body: out}, nil
}

type CreateAnnotationRequest struct {
	Name        string   `json:"name" minLength:"1"`
	Value       *float64 `json:"value,omitempty"`
	StringValue *string  `json:"string_value,omitempty"`
	DataType    string   `json:"type" enum:"NUMERIC,CATEGORICAL,BOOLEAN"`
	Reason      *string  `json:"reason,omitempty"`
}

type CreateTraceScoreInput struct {
	ID        string `path:"id"`
	ProjectID string `query:"project_id" required:"true" format:"uuid"`
	Body      CreateAnnotationRequest
}

type CreateTraceScoreOutput struct {
	Body *AnnotationResponse
}

func (h *dashboardHandler) createTraceScore(ctx context.Context, in *CreateTraceScoreInput) (*CreateTraceScoreOutput, error) {
	if in.ID == "" {
		return nil, appErrors.NewValidationError("Missing trace ID", "id is required")
	}
	projectID, err := parseUUIDParam(in.ProjectID, "project_id", "project ID")
	if err != nil {
		return nil, err
	}

	userID := httpctx.MustGetUserID(ctx)

	rootSpan, err := h.traces.GetRootSpan(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	if rootSpan.ProjectID != projectID {
		return nil, appErrors.NewValidationError("project_id", "does not match trace's project")
	}

	userIDStr := userID.String()
	score := &observability.Score{
		ID:             uid.New(),
		ProjectID:      projectID,
		OrganizationID: rootSpan.OrganizationID,
		TraceID:        &in.ID,
		SpanID:         &rootSpan.SpanID,
		Name:           in.Body.Name,
		Value:          in.Body.Value,
		StringValue:    in.Body.StringValue,
		Type:           in.Body.DataType,
		Source:         observability.ScoreSourceAnnotation,
		Reason:         in.Body.Reason,
		Metadata:       json.RawMessage("{}"),
		CreatedBy:      &userIDStr,
		Timestamp:      time.Now(),
	}

	if err := h.scores.CreateScore(ctx, score); err != nil {
		return nil, err
	}
	h.logger.InfoContext(ctx, "annotation created",
		"score_id", score.ID, "project_id", projectID, "trace_id", in.ID, "user_id", userID, "name", score.Name,
	)
	return &CreateTraceScoreOutput{Body: toAnnotationResponse(score)}, nil
}

type DeleteTraceScoreInput struct {
	ID        string `path:"id"`
	ScoreID   string `path:"scoreId" format:"uuid"`
	ProjectID string `query:"project_id" required:"true" format:"uuid"`
}
type DeleteTraceScoreOutput struct{}

func (h *dashboardHandler) deleteTraceScore(ctx context.Context, in *DeleteTraceScoreInput) (*DeleteTraceScoreOutput, error) {
	if in.ID == "" {
		return nil, appErrors.NewValidationError("Missing trace ID", "id is required")
	}
	scoreID, err := parseUUIDParam(in.ScoreID, "scoreId", "score ID")
	if err != nil {
		return nil, err
	}
	if _, err := parseUUIDParam(in.ProjectID, "project_id", "project ID"); err != nil {
		return nil, err
	}
	userID := httpctx.MustGetUserID(ctx)

	score, err := h.scores.GetScoreByID(ctx, scoreID)
	if err != nil {
		return nil, err
	}
	if score.TraceID == nil || *score.TraceID != in.ID {
		return nil, appErrors.NewNotFoundError("score")
	}
	if score.Source != observability.ScoreSourceAnnotation {
		return nil, appErrors.NewForbiddenError("only annotation scores can be deleted")
	}
	if score.CreatedBy == nil || *score.CreatedBy != userID.String() {
		return nil, appErrors.NewForbiddenError("only the creator can delete this annotation")
	}
	if err := h.scores.DeleteScore(ctx, scoreID); err != nil {
		return nil, err
	}
	h.logger.InfoContext(ctx, "annotation deleted",
		"score_id", scoreID, "project_id", in.ProjectID, "trace_id", in.ID, "user_id", userID,
	)
	return &DeleteTraceScoreOutput{}, nil
}

type UpdateTraceTagsInput struct {
	ID        string                                `path:"id"`
	ProjectID string                                `query:"project_id" required:"true" format:"uuid"`
	Body      observability.UpdateTraceTagsRequest
}
type UpdateTraceTagsOutput struct {
	Body struct {
		Tags []string `json:"tags"`
	}
}

func (h *dashboardHandler) updateTraceTags(ctx context.Context, in *UpdateTraceTagsInput) (*UpdateTraceTagsOutput, error) {
	if in.ID == "" {
		return nil, appErrors.NewValidationError("Missing trace ID", "id is required")
	}
	projectID, err := parseUUIDParam(in.ProjectID, "project_id", "project ID")
	if err != nil {
		return nil, err
	}
	if errs := in.Body.Validate(); len(errs) > 0 {
		return nil, appErrors.NewValidationError("Validation failed", errs[0].Message)
	}
	tags, err := h.traces.UpdateTraceTags(ctx, projectID, in.ID, in.Body.Tags)
	if err != nil {
		return nil, err
	}
	out := &UpdateTraceTagsOutput{}
	out.Body.Tags = tags
	return out, nil
}

type UpdateTraceBookmarkInput struct {
	ID        string `path:"id"`
	ProjectID string `query:"project_id" required:"true" format:"uuid"`
	Body      struct {
		Bookmarked bool `json:"bookmarked"`
	}
}
type UpdateTraceBookmarkOutput struct {
	Body struct {
		Bookmarked bool `json:"bookmarked"`
	}
}

func (h *dashboardHandler) updateTraceBookmark(ctx context.Context, in *UpdateTraceBookmarkInput) (*UpdateTraceBookmarkOutput, error) {
	if in.ID == "" {
		return nil, appErrors.NewValidationError("Missing trace ID", "id is required")
	}
	projectID, err := parseUUIDParam(in.ProjectID, "project_id", "project ID")
	if err != nil {
		return nil, err
	}
	if err := h.traces.UpdateTraceBookmark(ctx, projectID, in.ID, in.Body.Bookmarked); err != nil {
		return nil, err
	}
	out := &UpdateTraceBookmarkOutput{}
	out.Body.Bookmarked = in.Body.Bookmarked
	return out, nil
}

type GetTraceFilterOptionsInput struct {
	ProjectID string `query:"project_id" required:"true" format:"uuid"`
}
type GetTraceFilterOptionsOutput struct {
	Body *observability.TraceFilterOptions
}

func (h *dashboardHandler) getTraceFilterOptions(ctx context.Context, in *GetTraceFilterOptionsInput) (*GetTraceFilterOptionsOutput, error) {
	projectID, err := parseUUIDParam(in.ProjectID, "project_id", "project ID")
	if err != nil {
		return nil, err
	}
	opts, err := h.traces.GetFilterOptions(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return &GetTraceFilterOptionsOutput{Body: opts}, nil
}

type DiscoverAttributesInput struct {
	ProjectID string `query:"project_id" required:"true" format:"uuid"`
	Prefix    string `query:"prefix" required:"false"`
	Source    string `query:"source" required:"false" enum:"span_attributes,resource_attributes"`
	Limit     int    `query:"limit" required:"false" minimum:"1" maximum:"500"`
}
type DiscoverAttributesOutput struct {
	Body *observability.AttributeDiscoveryResponse
}

func (h *dashboardHandler) discoverAttributes(ctx context.Context, in *DiscoverAttributesInput) (*DiscoverAttributesOutput, error) {
	projectID, err := parseUUIDParam(in.ProjectID, "project_id", "project ID")
	if err != nil {
		return nil, err
	}
	req := &observability.AttributeDiscoveryRequest{
		ProjectID: projectID,
		Prefix:    in.Prefix,
		Limit:     in.Limit,
	}
	switch in.Source {
	case "":
		// default to all sources (handled by normalizer)
	case "span_attributes":
		req.Sources = []observability.AttributeSource{observability.AttributeSourceSpan}
	case "resource_attributes":
		req.Sources = []observability.AttributeSource{observability.AttributeSourceResource}
	}
	resp, err := h.traces.DiscoverAttributes(ctx, req)
	if err != nil {
		return nil, err
	}
	return &DiscoverAttributesOutput{Body: resp}, nil
}

// ==================================================================
// SPANS
// ==================================================================

type ListSpansInput struct {
	TraceID string `query:"trace_id" required:"false"`
	Type    string `query:"type" required:"false"`
	Model   string `query:"model" required:"false"`
	Level   string `query:"level" required:"false"`
	Page    int    `query:"page" required:"false" minimum:"1"`
	Limit   int    `query:"limit" required:"false"`
	SortBy  string `query:"sort_by" required:"false"`
	SortDir string `query:"sort_dir" required:"false" enum:"asc,desc"`
}
type ListSpansOutput struct {
	Body listSpansResponse
}
type listSpansResponse struct {
	Data       []*observability.Span `json:"data"`
	Pagination paginationMeta        `json:"pagination"`
}

func (h *dashboardHandler) listSpans(ctx context.Context, in *ListSpansInput) (*ListSpansOutput, error) {
	projectID := httpctx.MustGetProjectID(ctx)
	filter := &observability.SpanFilter{ProjectID: projectID}
	if in.TraceID != "" {
		filter.TraceID = &in.TraceID
	}
	if in.Type != "" {
		filter.Type = &in.Type
	}
	if in.Model != "" {
		filter.Model = &in.Model
	}
	if in.Level != "" {
		filter.Level = &in.Level
	}
	params := paginationParams(in.Page, in.Limit, in.SortBy, in.SortDir)
	filter.Params = params

	spans, err := h.traces.GetSpansByFilter(ctx, filter)
	if err != nil {
		return nil, err
	}
	total, err := h.traces.CountSpans(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &ListSpansOutput{Body: listSpansResponse{Data: spans, Pagination: newPaginationMeta(params, total)}}, nil
}

type GetSpanInput struct {
	ID string `path:"id" doc:"Span ID (OTEL 16-char hex)"`
}
type GetSpanOutput struct {
	Body *observability.Span
}

func (h *dashboardHandler) getSpan(ctx context.Context, in *GetSpanInput) (*GetSpanOutput, error) {
	if in.ID == "" {
		return nil, appErrors.NewValidationError("Missing span ID", "id is required")
	}
	span, err := h.traces.GetSpan(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	return &GetSpanOutput{Body: span}, nil
}

type DeleteSpanInput struct {
	ID string `path:"id"`
}
type DeleteSpanOutput struct{}

func (h *dashboardHandler) deleteSpan(ctx context.Context, in *DeleteSpanInput) (*DeleteSpanOutput, error) {
	if in.ID == "" {
		return nil, appErrors.NewValidationError("Missing span ID", "id is required")
	}
	if err := h.traces.DeleteSpan(ctx, in.ID); err != nil {
		return nil, err
	}
	return &DeleteSpanOutput{}, nil
}

// ==================================================================
// SCORES (project-scoped + global)
// ==================================================================

type scoreFilterQuery struct {
	TraceID string `query:"trace_id" required:"false"`
	SpanID  string `query:"span_id" required:"false"`
	Name    string `query:"name" required:"false"`
	Source  string `query:"source" required:"false"`
	Type    string `query:"type" required:"false"`
	Page    int    `query:"page" required:"false" minimum:"1"`
	Limit   int    `query:"limit" required:"false"`
	SortBy  string `query:"sort_by" required:"false"`
	SortDir string `query:"sort_dir" required:"false" enum:"asc,desc"`
}

func (q *scoreFilterQuery) apply(f *observability.ScoreFilter) pagination.Params {
	if q.TraceID != "" {
		f.TraceID = &q.TraceID
	}
	if q.SpanID != "" {
		f.SpanID = &q.SpanID
	}
	if q.Name != "" {
		f.Name = &q.Name
	}
	if q.Source != "" {
		f.Source = &q.Source
	}
	if q.Type != "" {
		f.Type = &q.Type
	}
	params := paginationParams(q.Page, q.Limit, q.SortBy, q.SortDir)
	f.Params = params
	return params
}

type ListScoresInput struct {
	scoreFilterQuery
}
type ListScoresOutput struct {
	Body listScoresResponse
}
type listScoresResponse struct {
	Data       []*ScoreResponse `json:"data"`
	Pagination paginationMeta   `json:"pagination"`
}

func (h *dashboardHandler) listScores(ctx context.Context, in *ListScoresInput) (*ListScoresOutput, error) {
	projectID := httpctx.MustGetProjectID(ctx)
	filter := &observability.ScoreFilter{ProjectID: projectID}
	params := in.scoreFilterQuery.apply(filter)

	scores, err := h.scores.GetScoresByFilter(ctx, filter)
	if err != nil {
		return nil, err
	}
	total, err := h.scores.CountScores(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &ListScoresOutput{Body: listScoresResponse{Data: toScoreResponses(scores), Pagination: newPaginationMeta(params, total)}}, nil
}

type ListProjectScoresInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	scoreFilterQuery
}
type ListProjectScoresOutput struct {
	Body listScoresResponse
}

func (h *dashboardHandler) listProjectScores(ctx context.Context, in *ListProjectScoresInput) (*ListProjectScoresOutput, error) {
	projectID, err := parseUUIDParam(in.ProjectID, "projectId", "project ID")
	if err != nil {
		return nil, err
	}
	filter := &observability.ScoreFilter{ProjectID: projectID}
	params := in.scoreFilterQuery.apply(filter)

	scores, err := h.scores.GetScoresByFilter(ctx, filter)
	if err != nil {
		return nil, err
	}
	total, err := h.scores.CountScores(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &ListProjectScoresOutput{Body: listScoresResponse{Data: toScoreResponses(scores), Pagination: newPaginationMeta(params, total)}}, nil
}

type GetScoreInput struct {
	ID string `path:"id" format:"uuid"`
}
type GetScoreOutput struct {
	Body *ScoreResponse
}

func (h *dashboardHandler) getScore(ctx context.Context, in *GetScoreInput) (*GetScoreOutput, error) {
	id, err := parseUUIDParam(in.ID, "id", "score ID")
	if err != nil {
		return nil, err
	}
	score, err := h.scores.GetScoreByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &GetScoreOutput{Body: toScoreResponse(score)}, nil
}

type UpdateScoreRequest struct {
	Name        string          `json:"name,omitempty"`
	Value       *float64        `json:"value,omitempty"`
	StringValue *string         `json:"string_value,omitempty"`
	Type        string          `json:"type,omitempty"`
	Source      string          `json:"source,omitempty"`
	Reason      *string         `json:"reason,omitempty"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
}

type UpdateScoreInput struct {
	ID   string `path:"id" format:"uuid"`
	Body UpdateScoreRequest
}
type UpdateScoreOutput struct {
	Body *ScoreResponse
}

func (h *dashboardHandler) updateScore(ctx context.Context, in *UpdateScoreInput) (*UpdateScoreOutput, error) {
	id, err := parseUUIDParam(in.ID, "id", "score ID")
	if err != nil {
		return nil, err
	}
	score := observability.Score{
		ID:          id,
		Name:        in.Body.Name,
		Value:       in.Body.Value,
		StringValue: in.Body.StringValue,
		Type:        in.Body.Type,
		Source:      in.Body.Source,
		Reason:      in.Body.Reason,
	}
	if len(in.Body.Metadata) > 0 {
		if !json.Valid(in.Body.Metadata) {
			return nil, appErrors.NewValidationError("Invalid metadata", "metadata must be valid JSON")
		}
		score.Metadata = in.Body.Metadata
	}
	if err := h.scores.UpdateScore(ctx, &score); err != nil {
		return nil, err
	}
	updated, err := h.scores.GetScoreByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &UpdateScoreOutput{Body: toScoreResponse(updated)}, nil
}

type GetScoreAnalyticsInput struct {
	ProjectID        string `path:"projectId" format:"uuid"`
	ScoreName        string `query:"score_name" required:"true"`
	CompareScoreName string `query:"compare_score_name" required:"false"`
	FromTimestamp    string `query:"from_timestamp" required:"false" doc:"RFC3339 timestamp"`
	ToTimestamp      string `query:"to_timestamp" required:"false" doc:"RFC3339 timestamp"`
	Interval         string `query:"interval" required:"false" enum:"hour,day,week"`
}
type GetScoreAnalyticsOutput struct {
	Body *observability.ScoreAnalyticsResponse
}

func (h *dashboardHandler) getScoreAnalytics(ctx context.Context, in *GetScoreAnalyticsInput) (*GetScoreAnalyticsOutput, error) {
	projectID, err := parseUUIDParam(in.ProjectID, "projectId", "project ID")
	if err != nil {
		return nil, err
	}
	if in.ScoreName == "" {
		return nil, appErrors.NewValidationError("Missing score name", "score_name is required")
	}
	interval := in.Interval
	if interval == "" {
		interval = "day"
	}
	filter := &observability.ScoreAnalyticsFilter{
		ProjectID: projectID,
		ScoreName: in.ScoreName,
		Interval:  interval,
	}
	if in.CompareScoreName != "" {
		filter.CompareScoreName = &in.CompareScoreName
	}
	if in.FromTimestamp != "" {
		ts, err := time.Parse(time.RFC3339, in.FromTimestamp)
		if err != nil {
			return nil, appErrors.NewValidationError("Invalid from_timestamp", "must be RFC3339 format (e.g., 2024-01-15T00:00:00Z)")
		}
		filter.FromTimestamp = &ts
	}
	if in.ToTimestamp != "" {
		ts, err := time.Parse(time.RFC3339, in.ToTimestamp)
		if err != nil {
			return nil, appErrors.NewValidationError("Invalid to_timestamp", "must be RFC3339 format (e.g., 2024-01-15T23:59:59Z)")
		}
		filter.ToTimestamp = &ts
	}
	analytics, err := h.scoreAnalytics.GetAnalytics(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &GetScoreAnalyticsOutput{Body: analytics}, nil
}

type GetScoreNamesInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
}
type GetScoreNamesOutput struct {
	Body []string
}

func (h *dashboardHandler) getScoreNames(ctx context.Context, in *GetScoreNamesInput) (*GetScoreNamesOutput, error) {
	if _, err := parseUUIDParam(in.ProjectID, "projectId", "project ID"); err != nil {
		return nil, err
	}
	names, err := h.scoreAnalytics.GetDistinctScoreNames(ctx, in.ProjectID)
	if err != nil {
		return nil, err
	}
	return &GetScoreNamesOutput{Body: names}, nil
}

// ==================================================================
// SESSIONS
// ==================================================================

type ListSessionsInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	Search    string `query:"search" required:"false"`
	UserID    string `query:"user_id" required:"false"`
	StartTime int64  `query:"start_time" required:"false" doc:"Unix timestamp seconds"`
	EndTime   int64  `query:"end_time" required:"false" doc:"Unix timestamp seconds"`
	Page      int    `query:"page" required:"false" minimum:"1"`
	Limit     int    `query:"limit" required:"false"`
	SortBy    string `query:"sort_by" required:"false"`
	SortDir   string `query:"sort_dir" required:"false" enum:"asc,desc"`
}
type ListSessionsOutput struct {
	Body listSessionsResponse
}
type listSessionsResponse struct {
	Data       []*observability.SessionSummary `json:"data"`
	Pagination paginationMeta                  `json:"pagination"`
}

func (h *dashboardHandler) listSessions(ctx context.Context, in *ListSessionsInput) (*ListSessionsOutput, error) {
	projectID, err := parseUUIDParam(in.ProjectID, "projectId", "project ID")
	if err != nil {
		return nil, err
	}
	filter := &observability.SessionFilter{ProjectID: projectID}
	if in.Search != "" {
		filter.Search = &in.Search
	}
	if in.UserID != "" {
		filter.UserID = &in.UserID
	}
	if in.StartTime != 0 {
		ts := time.Unix(in.StartTime, 0)
		filter.StartTime = &ts
	}
	if in.EndTime != 0 {
		ts := time.Unix(in.EndTime, 0)
		filter.EndTime = &ts
	}
	params := paginationParams(in.Page, in.Limit, in.SortBy, in.SortDir)
	filter.Params = params

	sessions, err := h.traces.ListSessions(ctx, filter)
	if err != nil {
		h.logger.ErrorContext(ctx, "observability: list sessions failed", "error", err, "project_id", projectID)
		return nil, err
	}
	total, err := h.traces.CountSessions(ctx, filter)
	if err != nil {
		h.logger.ErrorContext(ctx, "observability: count sessions failed", "error", err, "project_id", projectID)
		return nil, err
	}
	return &ListSessionsOutput{Body: listSessionsResponse{Data: sessions, Pagination: newPaginationMeta(params, total)}}, nil
}

// ==================================================================
// FILTER PRESETS
// ==================================================================

type CreateFilterPresetInput struct {
	ProjectID string                                    `path:"projectId" format:"uuid"`
	Body      observability.CreateFilterPresetRequest
}
type CreateFilterPresetOutput struct {
	Body *observability.FilterPreset
}

func (h *dashboardHandler) createFilterPreset(ctx context.Context, in *CreateFilterPresetInput) (*CreateFilterPresetOutput, error) {
	projectID, err := parseUUIDParam(in.ProjectID, "projectId", "project ID")
	if err != nil {
		return nil, err
	}
	userID := httpctx.MustGetUserID(ctx)
	preset, err := h.filterPresets.Create(ctx, projectID, userID, &in.Body)
	if err != nil {
		return nil, err
	}
	return &CreateFilterPresetOutput{Body: preset}, nil
}

type ListFilterPresetsInput struct {
	ProjectID     string `path:"projectId" format:"uuid"`
	TableName     string `query:"table_name" required:"false" enum:"traces,spans"`
	IncludePublic *bool  `query:"include_public" required:"false" doc:"Include public presets (default: true)"`
}
type ListFilterPresetsOutput struct {
	Body []*observability.FilterPreset
}

func (h *dashboardHandler) listFilterPresets(ctx context.Context, in *ListFilterPresetsInput) (*ListFilterPresetsOutput, error) {
	projectID, err := parseUUIDParam(in.ProjectID, "projectId", "project ID")
	if err != nil {
		return nil, err
	}
	userID := httpctx.MustGetUserID(ctx)
	var tableName *string
	if in.TableName != "" {
		tableName = &in.TableName
	}
	includePublic := true
	if in.IncludePublic != nil {
		includePublic = *in.IncludePublic
	}
	presets, err := h.filterPresets.List(ctx, projectID, userID, tableName, includePublic)
	if err != nil {
		return nil, err
	}
	return &ListFilterPresetsOutput{Body: presets}, nil
}

type GetFilterPresetInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	ID        string `path:"id" format:"uuid"`
}
type GetFilterPresetOutput struct {
	Body *observability.FilterPreset
}

func (h *dashboardHandler) getFilterPreset(ctx context.Context, in *GetFilterPresetInput) (*GetFilterPresetOutput, error) {
	projectID, err := parseUUIDParam(in.ProjectID, "projectId", "project ID")
	if err != nil {
		return nil, err
	}
	presetID, err := parseUUIDParam(in.ID, "id", "filter preset ID")
	if err != nil {
		return nil, err
	}
	userID := httpctx.MustGetUserID(ctx)
	preset, err := h.filterPresets.GetByID(ctx, projectID, presetID, userID)
	if err != nil {
		return nil, err
	}
	return &GetFilterPresetOutput{Body: preset}, nil
}

type UpdateFilterPresetInput struct {
	ProjectID string                                    `path:"projectId" format:"uuid"`
	ID        string                                    `path:"id" format:"uuid"`
	Body      observability.UpdateFilterPresetRequest
}
type UpdateFilterPresetOutput struct {
	Body *observability.FilterPreset
}

func (h *dashboardHandler) updateFilterPreset(ctx context.Context, in *UpdateFilterPresetInput) (*UpdateFilterPresetOutput, error) {
	projectID, err := parseUUIDParam(in.ProjectID, "projectId", "project ID")
	if err != nil {
		return nil, err
	}
	presetID, err := parseUUIDParam(in.ID, "id", "filter preset ID")
	if err != nil {
		return nil, err
	}
	userID := httpctx.MustGetUserID(ctx)
	preset, err := h.filterPresets.Update(ctx, projectID, presetID, userID, &in.Body)
	if err != nil {
		return nil, err
	}
	return &UpdateFilterPresetOutput{Body: preset}, nil
}

type DeleteFilterPresetInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	ID        string `path:"id" format:"uuid"`
}
type DeleteFilterPresetOutput struct{}

func (h *dashboardHandler) deleteFilterPreset(ctx context.Context, in *DeleteFilterPresetInput) (*DeleteFilterPresetOutput, error) {
	projectID, err := parseUUIDParam(in.ProjectID, "projectId", "project ID")
	if err != nil {
		return nil, err
	}
	presetID, err := parseUUIDParam(in.ID, "id", "filter preset ID")
	if err != nil {
		return nil, err
	}
	userID := httpctx.MustGetUserID(ctx)
	if err := h.filterPresets.Delete(ctx, projectID, presetID, userID); err != nil {
		return nil, err
	}
	return &DeleteFilterPresetOutput{}, nil
}
