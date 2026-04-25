package observability

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"brokle/internal/core/domain/observability"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/pagination"
	"brokle/pkg/request"
	"brokle/pkg/response"
	"brokle/pkg/uid"
)

// registerDashboardOps wires every dashboard-plane observability route
// onto r. Called from RegisterRoutes in handlers.go.
func registerDashboardOps(r chi.Router, h *dashboardHandler) {
	r.Route("/api/v1/traces", func(r chi.Router) {
		r.Get("/", h.listTraces)
		r.Get("/filter-options", h.getTraceFilterOptions)
		r.Get("/attributes", h.discoverAttributes)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.getTrace)
			r.Delete("/", h.deleteTrace)
			r.Get("/spans", h.getTraceSpans)
			r.Get("/scores", h.getTraceScores)
			r.Post("/scores", h.createTraceScore)
			r.Delete("/scores/{scoreId}", h.deleteTraceScore)
			r.Put("/tags", h.updateTraceTags)
			r.Put("/bookmark", h.updateTraceBookmark)
		})
	})

	r.Route("/api/v1/spans", func(r chi.Router) {
		r.Get("/", h.listSpans)
		r.Get("/{id}", h.getSpan)
		r.Delete("/{id}", h.deleteSpan)
	})

	r.Route("/api/v1/scores", func(r chi.Router) {
		r.Get("/", h.listScores)
		r.Get("/{id}", h.getScore)
		r.Put("/{id}", h.updateScore)
	})

	// Project-scoped routes. Mounted as sibling prefixes rather than a
	// single r.Route("/api/v1/projects/{projectId}", ...) wrapper, so
	// they don't Mount-collide with the evaluation handler's
	// /api/v1/projects/{projectId}/{score-configs,datasets,...}
	// siblings on the same chi tree (chi panics on duplicate Mount
	// patterns; see gotcha #31).
	r.Route("/api/v1/projects/{projectId}/scores", func(r chi.Router) {
		r.Get("/", h.listProjectScores)
		r.Get("/analytics", h.getScoreAnalytics)
		r.Get("/names", h.getScoreNames)
	})

	r.Get("/api/v1/projects/{projectId}/sessions", h.listSessions)

	r.Route("/api/v1/projects/{projectId}/filter-presets", func(r chi.Router) {
		r.Post("/", h.createFilterPreset)
		r.Get("/", h.listFilterPresets)
		r.Get("/{id}", h.getFilterPreset)
		r.Patch("/{id}", h.updateFilterPreset)
		r.Delete("/{id}", h.deleteFilterPreset)
	})
}

// ---- shared helpers -------------------------------------------------

func parsePaginationQuery(r *http.Request) pagination.Params {
	q := r.URL.Query()
	page, _ := parsePositiveInt(q.Get("page"))
	limit, _ := parsePositiveInt(q.Get("limit"))
	p := pagination.Params{
		Page:    page,
		Limit:   limit,
		SortBy:  q.Get("sort_by"),
		SortDir: q.Get("sort_dir"),
	}
	p.SetDefaults("")
	return p
}

func parsePositiveInt(s string) (int, error) {
	if s == "" {
		return 0, nil
	}
	var v int
	_, err := fmtSscanf(s, "%d", &v)
	if err != nil || v < 0 {
		return 0, err
	}
	return v, nil
}

// fmtSscanf avoids pulling in the stdlib `fmt` for one call.
func fmtSscanf(s, format string, args ...any) (int, error) {
	var n int
	_ = format
	_ = args
	// Minimal positive-int parser — just strconv.Atoi.
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, appErrors.NewValidationError("Invalid number", "expected positive integer")
		}
		n = n*10 + int(s[i]-'0')
	}
	if len(args) > 0 {
		if ptr, ok := args[0].(*int); ok {
			*ptr = n
		}
	}
	return n, nil
}

func newPaginationMeta(p pagination.Params, total int64) paginationMeta {
	pages := 0
	if p.Limit > 0 {
		pages = int((total + int64(p.Limit) - 1) / int64(p.Limit))
	}
	return paginationMeta{Page: p.Page, Limit: p.Limit, Total: total, TotalPages: pages}
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func validateStatusList(values []string, field string) error {
	for _, v := range values {
		if v != "ok" && v != "error" && v != "unset" {
			return appErrors.NewValidationError(
				"Invalid "+field+" value",
				field+" must be one of: ok, error, unset (got: "+v+")",
				appErrors.WithParam(field),
			)
		}
	}
	return nil
}

// parseProjectIDQuery parses the `project_id` query parameter as a UUID.
func parseProjectIDQuery(r *http.Request) (uuid.UUID, error) {
	raw := r.URL.Query().Get("project_id")
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, appErrors.NewValidationError(
			"Invalid project ID", "project_id must be a valid UUID",
			appErrors.WithParam("project_id"),
		)
	}
	return id, nil
}

// ==================================================================
// TRACES
// ==================================================================

func (h *dashboardHandler) listTraces(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	q := r.URL.Query()

	filter := &observability.TraceFilter{ProjectID: projectID}
	if v := q.Get("session_id"); v != "" {
		filter.SessionID = &v
	}
	if v := q.Get("user_id"); v != "" {
		filter.UserID = &v
	}
	if v := q.Get("service_name"); v != "" {
		filter.ServiceName = &v
	}
	if v := q.Get("model_name"); v != "" {
		filter.ModelName = &v
	}
	if v := q.Get("provider_name"); v != "" {
		filter.ProviderName = &v
	}

	if v, err := parseFloat(q.Get("min_cost")); err != nil {
		response.WriteError(w, err)
		return
	} else if v != 0 {
		filter.MinCost = &v
	}
	if v, err := parseFloat(q.Get("max_cost")); err != nil {
		response.WriteError(w, err)
		return
	} else if v != 0 {
		filter.MaxCost = &v
	}
	if v, err := parseInt64(q.Get("min_tokens")); err != nil {
		response.WriteError(w, err)
		return
	} else if v != 0 {
		filter.MinTokens = &v
	}
	if v, err := parseInt64(q.Get("max_tokens")); err != nil {
		response.WriteError(w, err)
		return
	} else if v != 0 {
		filter.MaxTokens = &v
	}
	if v, err := parseInt64(q.Get("min_duration")); err != nil {
		response.WriteError(w, err)
		return
	} else if v != 0 {
		filter.MinDuration = &v
	}
	if v, err := parseInt64(q.Get("max_duration")); err != nil {
		response.WriteError(w, err)
		return
	} else if v != 0 {
		filter.MaxDuration = &v
	}

	hasErrorPtr, err := request.QueryOptionalBool(r, "has_error")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	filter.HasError = hasErrorPtr

	if v := q.Get("search"); v != "" {
		filter.Search = &v
	}
	if v := q.Get("search_type"); v != "" {
		filter.SearchType = &v
	}
	if v := q.Get("status"); v != "" {
		filter.Statuses = splitCSV(v)
		if err := validateStatusList(filter.Statuses, "status"); err != nil {
			response.WriteError(w, err)
			return
		}
	}
	if v := q.Get("status_not"); v != "" {
		filter.StatusesNot = splitCSV(v)
		if err := validateStatusList(filter.StatusesNot, "status_not"); err != nil {
			response.WriteError(w, err)
			return
		}
	}
	if v, err := parseInt64(q.Get("start_time")); err != nil {
		response.WriteError(w, err)
		return
	} else if v != 0 {
		ts := time.Unix(v, 0)
		filter.StartTime = &ts
	}
	if v, err := parseInt64(q.Get("end_time")); err != nil {
		response.WriteError(w, err)
		return
	} else if v != 0 {
		ts := time.Unix(v, 0)
		filter.EndTime = &ts
	}

	params := parsePaginationQuery(r)
	filter.Params = params

	traces, err := h.traces.ListTraces(r.Context(), filter)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "observability: list traces failed", "error", err)
		response.WriteError(w, err)
		return
	}
	total, err := h.traces.CountTraces(r.Context(), filter)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "observability: count traces failed", "error", err)
		response.WriteError(w, err)
		return
	}

	response.Success(w, listTracesResponse{
		Data:       traces,
		Pagination: newPaginationMeta(params, total),
	})
}

func (h *dashboardHandler) getTrace(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.WriteError(w, appErrors.NewValidationError("Missing trace ID", "id is required"))
		return
	}
	summary, err := h.traces.GetTrace(r.Context(), id)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, summary)
}

func (h *dashboardHandler) deleteTrace(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.WriteError(w, appErrors.NewValidationError("Missing trace ID", "id is required"))
		return
	}
	if err := h.traces.DeleteTrace(r.Context(), id); err != nil {
		response.WriteError(w, err)
		return
	}
	response.NoContent(w)
}

func (h *dashboardHandler) getTraceSpans(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.WriteError(w, appErrors.NewValidationError("Missing trace ID", "id is required"))
		return
	}
	spans, err := h.traces.GetTraceSpans(r.Context(), id)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, spans)
}

func (h *dashboardHandler) getTraceScores(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.WriteError(w, appErrors.NewValidationError("Missing trace ID", "id is required"))
		return
	}
	if _, err := parseProjectIDQuery(r); err != nil {
		response.WriteError(w, err)
		return
	}
	scores, err := h.scores.GetScoresByTraceID(r.Context(), id)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	out := make([]*AnnotationResponse, 0, len(scores))
	for _, s := range scores {
		out = append(out, toAnnotationResponse(s))
	}
	response.Success(w, out)
}

func (h *dashboardHandler) createTraceScore(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.WriteError(w, appErrors.NewValidationError("Missing trace ID", "id is required"))
		return
	}
	projectID, err := parseProjectIDQuery(r)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	var body CreateAnnotationRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	userID := httpctx.MustGetUserID(r.Context())

	rootSpan, err := h.traces.GetRootSpan(r.Context(), id)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	if rootSpan.ProjectID != projectID {
		response.WriteError(w, appErrors.NewValidationError("project_id", "does not match trace's project"))
		return
	}

	userIDStr := userID.String()
	score := &observability.Score{
		ID:             uid.New(),
		ProjectID:      projectID,
		OrganizationID: rootSpan.OrganizationID,
		TraceID:        &id,
		SpanID:         &rootSpan.SpanID,
		Name:           body.Name,
		Value:          body.Value,
		StringValue:    body.StringValue,
		Type:           body.DataType,
		Source:         observability.ScoreSourceAnnotation,
		Reason:         body.Reason,
		Metadata:       json.RawMessage("{}"),
		CreatedBy:      &userIDStr,
		Timestamp:      time.Now(),
	}

	if err := h.scores.CreateScore(r.Context(), score); err != nil {
		response.WriteError(w, err)
		return
	}
	h.logger.InfoContext(r.Context(), "annotation created",
		"score_id", score.ID, "project_id", projectID, "trace_id", id, "user_id", userID, "name", score.Name)
	response.Created(w, toAnnotationResponse(score))
}

func (h *dashboardHandler) deleteTraceScore(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.WriteError(w, appErrors.NewValidationError("Missing trace ID", "id is required"))
		return
	}
	scoreID, err := request.URLParamUUID(r, "scoreId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	if _, err := parseProjectIDQuery(r); err != nil {
		response.WriteError(w, err)
		return
	}
	userID := httpctx.MustGetUserID(r.Context())

	score, err := h.scores.GetScoreByID(r.Context(), scoreID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	if score.TraceID == nil || *score.TraceID != id {
		response.WriteError(w, appErrors.NewNotFoundError("score"))
		return
	}
	if score.Source != observability.ScoreSourceAnnotation {
		response.WriteError(w, appErrors.NewForbiddenError("only annotation scores can be deleted"))
		return
	}
	if score.CreatedBy == nil || *score.CreatedBy != userID.String() {
		response.WriteError(w, appErrors.NewForbiddenError("only the creator can delete this annotation"))
		return
	}
	if err := h.scores.DeleteScore(r.Context(), scoreID); err != nil {
		response.WriteError(w, err)
		return
	}
	h.logger.InfoContext(r.Context(), "annotation deleted",
		"score_id", scoreID, "trace_id", id, "user_id", userID)
	response.NoContent(w)
}

func (h *dashboardHandler) updateTraceTags(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.WriteError(w, appErrors.NewValidationError("Missing trace ID", "id is required"))
		return
	}
	projectID, err := parseProjectIDQuery(r)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	var body observability.UpdateTraceTagsRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	if errs := body.Validate(); len(errs) > 0 {
		response.WriteError(w, appErrors.NewValidationError("Validation failed", errs[0].Message))
		return
	}
	tags, err := h.traces.UpdateTraceTags(r.Context(), projectID, id, body.Tags)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, map[string]any{"tags": tags})
}

func (h *dashboardHandler) updateTraceBookmark(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.WriteError(w, appErrors.NewValidationError("Missing trace ID", "id is required"))
		return
	}
	projectID, err := parseProjectIDQuery(r)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	var body struct {
		Bookmarked bool `json:"bookmarked"`
	}
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	if err := h.traces.UpdateTraceBookmark(r.Context(), projectID, id, body.Bookmarked); err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, map[string]any{"bookmarked": body.Bookmarked})
}

func (h *dashboardHandler) getTraceFilterOptions(w http.ResponseWriter, r *http.Request) {
	projectID, err := parseProjectIDQuery(r)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	opts, err := h.traces.GetFilterOptions(r.Context(), projectID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, opts)
}

func (h *dashboardHandler) discoverAttributes(w http.ResponseWriter, r *http.Request) {
	projectID, err := parseProjectIDQuery(r)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	q := r.URL.Query()
	limit, err := request.QueryInt(r, "limit", 0)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	req := &observability.AttributeDiscoveryRequest{
		ProjectID: projectID,
		Prefix:    q.Get("prefix"),
		Limit:     limit,
	}
	switch q.Get("source") {
	case "":
	case "span_attributes":
		req.Sources = []observability.AttributeSource{observability.AttributeSourceSpan}
	case "resource_attributes":
		req.Sources = []observability.AttributeSource{observability.AttributeSourceResource}
	}
	resp, err := h.traces.DiscoverAttributes(r.Context(), req)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, resp)
}

// ==================================================================
// SPANS
// ==================================================================

func (h *dashboardHandler) listSpans(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	q := r.URL.Query()
	filter := &observability.SpanFilter{ProjectID: projectID}
	if v := q.Get("trace_id"); v != "" {
		filter.TraceID = &v
	}
	if v := q.Get("type"); v != "" {
		filter.Type = &v
	}
	if v := q.Get("model"); v != "" {
		filter.Model = &v
	}
	if v := q.Get("level"); v != "" {
		filter.Level = &v
	}
	params := parsePaginationQuery(r)
	filter.Params = params

	spans, err := h.traces.GetSpansByFilter(r.Context(), filter)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	total, err := h.traces.CountSpans(r.Context(), filter)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, listSpansResponse{Data: spans, Pagination: newPaginationMeta(params, total)})
}

func (h *dashboardHandler) getSpan(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.WriteError(w, appErrors.NewValidationError("Missing span ID", "id is required"))
		return
	}
	span, err := h.traces.GetSpan(r.Context(), id)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, span)
}

func (h *dashboardHandler) deleteSpan(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.WriteError(w, appErrors.NewValidationError("Missing span ID", "id is required"))
		return
	}
	if err := h.traces.DeleteSpan(r.Context(), id); err != nil {
		response.WriteError(w, err)
		return
	}
	response.NoContent(w)
}

// ==================================================================
// SCORES (project-scoped + global)
// ==================================================================

func applyScoreFilter(r *http.Request, f *observability.ScoreFilter) pagination.Params {
	q := r.URL.Query()
	if v := q.Get("trace_id"); v != "" {
		f.TraceID = &v
	}
	if v := q.Get("span_id"); v != "" {
		f.SpanID = &v
	}
	if v := q.Get("name"); v != "" {
		f.Name = &v
	}
	if v := q.Get("source"); v != "" {
		f.Source = &v
	}
	if v := q.Get("type"); v != "" {
		f.Type = &v
	}
	params := parsePaginationQuery(r)
	f.Params = params
	return params
}

func (h *dashboardHandler) listScores(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	filter := &observability.ScoreFilter{ProjectID: projectID}
	params := applyScoreFilter(r, filter)

	scores, err := h.scores.GetScoresByFilter(r.Context(), filter)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	total, err := h.scores.CountScores(r.Context(), filter)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, listScoresResponse{
		Data:       toTraceScoreResponses(scores),
		Pagination: newPaginationMeta(params, total),
	})
}

func (h *dashboardHandler) listProjectScores(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	filter := &observability.ScoreFilter{ProjectID: projectID}
	params := applyScoreFilter(r, filter)

	scores, err := h.scores.GetScoresByFilter(r.Context(), filter)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	total, err := h.scores.CountScores(r.Context(), filter)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, listScoresResponse{
		Data:       toTraceScoreResponses(scores),
		Pagination: newPaginationMeta(params, total),
	})
}

func (h *dashboardHandler) getScore(w http.ResponseWriter, r *http.Request) {
	id, err := request.URLParamUUID(r, "id")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	score, err := h.scores.GetScoreByID(r.Context(), id)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, toTraceScoreResponse(score))
}

func (h *dashboardHandler) updateScore(w http.ResponseWriter, r *http.Request) {
	id, err := request.URLParamUUID(r, "id")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	var body UpdateScoreRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	score := observability.Score{
		ID:          id,
		Name:        body.Name,
		Value:       body.Value,
		StringValue: body.StringValue,
		Type:        body.Type,
		Source:      body.Source,
		Reason:      body.Reason,
	}
	if len(body.Metadata) > 0 {
		if !json.Valid(body.Metadata) {
			response.WriteError(w, appErrors.NewValidationError("Invalid metadata", "metadata must be valid JSON"))
			return
		}
		score.Metadata = body.Metadata
	}
	if err := h.scores.UpdateScore(r.Context(), &score); err != nil {
		response.WriteError(w, err)
		return
	}
	updated, err := h.scores.GetScoreByID(r.Context(), id)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, toTraceScoreResponse(updated))
}

func (h *dashboardHandler) getScoreAnalytics(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	q := r.URL.Query()
	scoreName := q.Get("score_name")
	if scoreName == "" {
		response.WriteError(w, appErrors.NewValidationError("Missing score name", "score_name is required", appErrors.WithParam("score_name")))
		return
	}
	interval := q.Get("interval")
	if interval == "" {
		interval = "day"
	}
	filter := &observability.ScoreAnalyticsFilter{
		ProjectID: projectID,
		ScoreName: scoreName,
		Interval:  interval,
	}
	if v := q.Get("compare_score_name"); v != "" {
		filter.CompareScoreName = &v
	}
	if v := q.Get("from_timestamp"); v != "" {
		ts, err := time.Parse(time.RFC3339, v)
		if err != nil {
			response.WriteError(w, appErrors.NewValidationError("Invalid from_timestamp", "must be RFC3339 format", appErrors.WithParam("from_timestamp")))
			return
		}
		filter.FromTimestamp = &ts
	}
	if v := q.Get("to_timestamp"); v != "" {
		ts, err := time.Parse(time.RFC3339, v)
		if err != nil {
			response.WriteError(w, appErrors.NewValidationError("Invalid to_timestamp", "must be RFC3339 format", appErrors.WithParam("to_timestamp")))
			return
		}
		filter.ToTimestamp = &ts
	}
	analytics, err := h.scoreAnalytics.GetAnalytics(r.Context(), filter)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, analytics)
}

func (h *dashboardHandler) getScoreNames(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	names, err := h.scoreAnalytics.GetDistinctScoreNames(r.Context(), projectID.String())
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, names)
}

// ==================================================================
// SESSIONS
// ==================================================================

func (h *dashboardHandler) listSessions(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	q := r.URL.Query()
	filter := &observability.SessionFilter{ProjectID: projectID}
	if v := q.Get("search"); v != "" {
		filter.Search = &v
	}
	if v := q.Get("user_id"); v != "" {
		filter.UserID = &v
	}
	if v, err := parseInt64(q.Get("start_time")); err != nil {
		response.WriteError(w, err)
		return
	} else if v != 0 {
		ts := time.Unix(v, 0)
		filter.StartTime = &ts
	}
	if v, err := parseInt64(q.Get("end_time")); err != nil {
		response.WriteError(w, err)
		return
	} else if v != 0 {
		ts := time.Unix(v, 0)
		filter.EndTime = &ts
	}
	params := parsePaginationQuery(r)
	filter.Params = params

	sessions, err := h.traces.ListSessions(r.Context(), filter)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "observability: list sessions failed",
			"error", err, "project_id", projectID)
		response.WriteError(w, err)
		return
	}
	total, err := h.traces.CountSessions(r.Context(), filter)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "observability: count sessions failed",
			"error", err, "project_id", projectID)
		response.WriteError(w, err)
		return
	}
	response.Success(w, listTraceSessionsResponse{
		Data:       sessions,
		Pagination: newPaginationMeta(params, total),
	})
}

// ==================================================================
// FILTER PRESETS
// ==================================================================

func (h *dashboardHandler) createFilterPreset(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	var body observability.CreateFilterPresetRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	userID := httpctx.MustGetUserID(r.Context())
	preset, err := h.filterPresets.Create(r.Context(), projectID, userID, &body)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Created(w, preset)
}

func (h *dashboardHandler) listFilterPresets(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	userID := httpctx.MustGetUserID(r.Context())
	var tableName *string
	if v := r.URL.Query().Get("table_name"); v != "" {
		tableName = &v
	}

	includePublic := true
	if v, err := request.QueryOptionalBool(r, "include_public"); err != nil {
		response.WriteError(w, err)
		return
	} else if v != nil {
		includePublic = *v
	}

	presets, err := h.filterPresets.List(r.Context(), projectID, userID, tableName, includePublic)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, presets)
}

func (h *dashboardHandler) getFilterPreset(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	presetID, err := request.URLParamUUID(r, "id")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	userID := httpctx.MustGetUserID(r.Context())
	preset, err := h.filterPresets.GetByID(r.Context(), projectID, presetID, userID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, preset)
}

func (h *dashboardHandler) updateFilterPreset(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	presetID, err := request.URLParamUUID(r, "id")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	var body observability.UpdateFilterPresetRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	userID := httpctx.MustGetUserID(r.Context())
	preset, err := h.filterPresets.Update(r.Context(), projectID, presetID, userID, &body)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, preset)
}

func (h *dashboardHandler) deleteFilterPreset(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	presetID, err := request.URLParamUUID(r, "id")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	userID := httpctx.MustGetUserID(r.Context())
	if err := h.filterPresets.Delete(r.Context(), projectID, presetID, userID); err != nil {
		response.WriteError(w, err)
		return
	}
	response.NoContent(w)
}

// ---- tiny numeric parsers -------------------------------------------

func parseFloat(s string) (float64, error) {
	if s == "" {
		return 0, nil
	}
	// strconv.ParseFloat via fmt.Sscanf to keep imports small.
	var v float64
	n, err := fmtSscanfFloat(s, &v)
	if err != nil || n == 0 {
		return 0, appErrors.NewValidationError("Invalid number", "expected a number")
	}
	return v, nil
}

func fmtSscanfFloat(s string, out *float64) (int, error) {
	// Minimal float parser — support digits + '.' + optional leading '-'.
	sign := 1.0
	i := 0
	if i < len(s) && s[i] == '-' {
		sign = -1
		i++
	}
	seenDigit := false
	intPart, fracPart, fracDiv := 0.0, 0.0, 1.0
	for ; i < len(s); i++ {
		c := s[i]
		if c == '.' {
			break
		}
		if c < '0' || c > '9' {
			return 0, appErrors.NewValidationError("Invalid number", "expected numeric digits")
		}
		intPart = intPart*10 + float64(c-'0')
		seenDigit = true
	}
	if i < len(s) && s[i] == '.' {
		i++
		for ; i < len(s); i++ {
			c := s[i]
			if c < '0' || c > '9' {
				return 0, appErrors.NewValidationError("Invalid number", "expected numeric digits after decimal")
			}
			fracPart = fracPart*10 + float64(c-'0')
			fracDiv *= 10
			seenDigit = true
		}
	}
	if !seenDigit {
		return 0, nil
	}
	*out = sign * (intPart + fracPart/fracDiv)
	return 1, nil
}

func parseInt64(s string) (int64, error) {
	if s == "" {
		return 0, nil
	}
	sign := int64(1)
	i := 0
	if i < len(s) && s[i] == '-' {
		sign = -1
		i++
	}
	var v int64
	seenDigit := false
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return 0, appErrors.NewValidationError("Invalid number", "expected integer")
		}
		v = v*10 + int64(c-'0')
		seenDigit = true
	}
	if !seenDigit {
		return 0, appErrors.NewValidationError("Invalid number", "expected integer")
	}
	return sign * v, nil
}
