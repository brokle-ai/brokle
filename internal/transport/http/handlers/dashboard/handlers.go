// Package dashboard is the dashboard-plane dashboards handler domain.
// Exposes:
//   - /api/v1/projects/{projectId}/dashboards       — CRUD + list
//   - /api/v1/projects/{projectId}/dashboards/{id}  — single + lifecycle
//     (duplicate, lock/unlock, export, execute, from-template, import)
//   - /api/v1/projects/{projectId}/dashboards/{id}/widgets/{widgetId}/execute
//   - /api/v1/dashboard-templates                   — template catalogue
//   - /api/v1/dashboards/view-definitions           — query-builder metadata
//
// Every op requires RequireAuth. The dashboard service enforces
// project-scoped checks on its own queries; the handler does no
// additional membership validation.
package dashboard

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	dashboardDomain "brokle/internal/core/domain/dashboard"
	dashboardService "brokle/internal/core/services/dashboard"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/request"
	"brokle/pkg/response"
)

type Handler struct {
	svc      *dashboardService.DashboardService
	query    *dashboardService.WidgetQueryService
	template *dashboardService.TemplateService
	logger   *slog.Logger
}

// New constructs the dashboard handler.
func New(
	svc *dashboardService.DashboardService,
	query *dashboardService.WidgetQueryService,
	template *dashboardService.TemplateService,
	logger *slog.Logger,
) *Handler {
	return &Handler{svc: svc, query: query, template: template, logger: logger}
}

// userIDPtr returns a *uuid.UUID when an authenticated user is present
// in context, nil otherwise. Matches the pre-migration "optional
// creator" semantics of CreateDashboard / ImportDashboard /
// CreateFromTemplate.
func userIDPtr(ctx context.Context) *uuid.UUID {
	uid, ok := httpctx.UserID(ctx)
	if !ok {
		return nil
	}
	return &uid
}

// ---- list -------------------------------------------------------------

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	limit, err := request.QueryInt(r, "limit", 0)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	offset, err := request.QueryInt(r, "offset", 0)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	filter := &dashboardDomain.DashboardFilter{
		Name:   r.URL.Query().Get("name"),
		Limit:  limit,
		Offset: offset,
	}
	items, total, err := h.svc.ListDashboards(r.Context(), projectID, filter)
	if err != nil {
		h.logger.WarnContext(r.Context(), "dashboard: list failed",
			"project_id", projectID, "error", err)
		response.WriteError(w, err)
		return
	}
	pageLimit := limit
	if pageLimit <= 0 {
		pageLimit = 50
	}
	page := offset/pageLimit + 1
	response.Success(w, listDashboardsBody{
		Data:       items,
		Pagination: response.BuildPagination(page, pageLimit, total),
	})
}

// ---- create -----------------------------------------------------------

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	var body dashboardDomain.CreateDashboardRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	if body.Name == "" {
		response.WriteError(w, appErrors.NewValidationError(
			"Name is required", "dashboard name is required",
			appErrors.WithParam("name"),
		))
		return
	}
	dash, err := h.svc.CreateDashboard(r.Context(), projectID, userIDPtr(r.Context()), &body)
	if err != nil {
		h.logger.WarnContext(r.Context(), "dashboard: create failed",
			"project_id", projectID, "name", body.Name, "error", err)
		response.WriteError(w, err)
		return
	}
	response.Created(w, dash)
}

// ---- get --------------------------------------------------------------

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	dashboardID, err := request.URLParamUUID(r, "dashboardId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	dash, err := h.svc.GetDashboardByProject(r.Context(), projectID, dashboardID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, dash)
}

// ---- update -----------------------------------------------------------

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	dashboardID, err := request.URLParamUUID(r, "dashboardId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	var body dashboardDomain.UpdateDashboardRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	dash, err := h.svc.UpdateDashboard(r.Context(), projectID, dashboardID, &body)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, dash)
}

// ---- delete -----------------------------------------------------------

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	dashboardID, err := request.URLParamUUID(r, "dashboardId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	if err := h.svc.DeleteDashboard(r.Context(), projectID, dashboardID); err != nil {
		response.WriteError(w, err)
		return
	}
	response.NoContent(w)
}

// ---- duplicate --------------------------------------------------------

func (h *Handler) Duplicate(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	dashboardID, err := request.URLParamUUID(r, "dashboardId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	var body dashboardDomain.DuplicateDashboardRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	if body.Name == "" {
		response.WriteError(w, appErrors.NewValidationError(
			"Name is required", "dashboard name is required",
			appErrors.WithParam("name"),
		))
		return
	}
	dash, err := h.svc.DuplicateDashboard(r.Context(), projectID, dashboardID, &body)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Created(w, dash)
}

// ---- lock / unlock ----------------------------------------------------

func (h *Handler) Lock(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	dashboardID, err := request.URLParamUUID(r, "dashboardId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	dash, err := h.svc.LockDashboard(r.Context(), projectID, dashboardID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, dash)
}

func (h *Handler) Unlock(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	dashboardID, err := request.URLParamUUID(r, "dashboardId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	dash, err := h.svc.UnlockDashboard(r.Context(), projectID, dashboardID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, dash)
}

// ---- export / import --------------------------------------------------

func (h *Handler) Export(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	dashboardID, err := request.URLParamUUID(r, "dashboardId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	exp, err := h.svc.ExportDashboard(r.Context(), projectID, dashboardID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, exp)
}

func (h *Handler) ImportDashboard(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	var body dashboardDomain.DashboardImportRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	dash, err := h.svc.ImportDashboard(r.Context(), projectID, userIDPtr(r.Context()), &body)
	if err != nil {
		h.logger.WarnContext(r.Context(), "dashboard: import failed",
			"project_id", projectID, "error", err)
		response.WriteError(w, err)
		return
	}
	response.Created(w, dash)
}

// ---- query execution --------------------------------------------------

func (b *executeDashboardBody) toDomainTimeRange() *dashboardDomain.DashboardTimeRange {
	if b == nil || b.DashboardTimeRange == nil {
		return nil
	}
	return &dashboardDomain.DashboardTimeRange{
		From:     b.DashboardTimeRange.From,
		To:       b.DashboardTimeRange.To,
		Relative: b.DashboardTimeRange.Relative,
	}
}

func (h *Handler) ExecuteDashboard(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	dashboardID, err := request.URLParamUUID(r, "dashboardId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	var body executeDashboardBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	req := &dashboardDomain.QueryExecutionRequest{
		ProjectID:          projectID,
		DashboardID:        dashboardID,
		DashboardTimeRange: body.toDomainTimeRange(),
		ForceRefresh:       body.ForceRefresh,
		VariableValues:     body.VariableValues,
	}
	results, err := h.query.ExecuteDashboardQueries(r.Context(), req)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, results)
}

func (h *Handler) ExecuteWidget(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	dashboardID, err := request.URLParamUUID(r, "dashboardId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	widgetID := chi.URLParam(r, "widgetId")
	if widgetID == "" {
		response.WriteError(w, appErrors.NewValidationError(
			"Invalid widget ID", "widgetId is required",
			appErrors.WithParam("widgetId"),
		))
		return
	}
	var body executeDashboardBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	req := &dashboardDomain.QueryExecutionRequest{
		ProjectID:          projectID,
		DashboardID:        dashboardID,
		WidgetID:           &widgetID,
		DashboardTimeRange: body.toDomainTimeRange(),
		ForceRefresh:       body.ForceRefresh,
	}
	results, err := h.query.ExecuteDashboardQueries(r.Context(), req)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	result, ok := results.Results[widgetID]
	if !ok {
		response.WriteError(w, appErrors.NewNotFoundError("widget"))
		return
	}
	response.Success(w, result)
}

func (h *Handler) ViewDefinitions(w http.ResponseWriter, r *http.Request) {
	resp, err := h.query.GetViewDefinitions(r.Context())
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, resp)
}

func (h *Handler) VariableOptions(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	view := r.URL.Query().Get("view")
	dimension := r.URL.Query().Get("dimension")
	if view == "" {
		response.WriteError(w, appErrors.NewValidationError(
			"View required", "view query parameter is required",
			appErrors.WithParam("view"),
		))
		return
	}
	if dimension == "" {
		response.WriteError(w, appErrors.NewValidationError(
			"Dimension required", "dimension query parameter is required",
			appErrors.WithParam("dimension"),
		))
		return
	}
	limit, err := request.QueryInt(r, "limit", 100)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	if limit <= 0 {
		limit = 100
	}
	req := &dashboardDomain.VariableOptionsRequest{
		ProjectID: projectID,
		View:      dashboardDomain.ViewType(view),
		Dimension: dimension,
		Limit:     limit,
	}
	result, err := h.query.GetVariableOptions(r.Context(), req)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, result)
}

// ---- templates --------------------------------------------------------

func (h *Handler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	templates, err := h.template.ListTemplates(r.Context())
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, templates)
}

func (h *Handler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	templateID, err := request.URLParamUUID(r, "templateId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	tmpl, err := h.template.GetTemplate(r.Context(), templateID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, tmpl)
}

func (h *Handler) CreateFromTemplate(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	var body dashboardDomain.CreateFromTemplateRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	dash, err := h.template.CreateFromTemplate(r.Context(), projectID, userIDPtr(r.Context()), &body)
	if err != nil {
		h.logger.WarnContext(r.Context(), "dashboard: create-from-template failed",
			"project_id", projectID, "template_id", body.TemplateID, "error", err)
		response.WriteError(w, err)
		return
	}
	response.Created(w, dash)
}
