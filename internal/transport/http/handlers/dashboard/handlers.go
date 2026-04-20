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
// additional membership validation (matches the pre-migration
// gin behaviour).
package dashboard

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	dashboardDomain "brokle/internal/core/domain/dashboard"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
)

type handler struct {
	svc      dashboardDomain.DashboardService
	query    dashboardDomain.WidgetQueryService
	template dashboardDomain.TemplateService
	logger   *slog.Logger
}

// RegisterRoutes registers every dashboard operation on apiAdmin.
func RegisterRoutes(
	api huma.API,
	svc dashboardDomain.DashboardService,
	query dashboardDomain.WidgetQueryService,
	template dashboardDomain.TemplateService,
	logger *slog.Logger,
) {
	h := &handler{svc: svc, query: query, template: template, logger: logger}

	// ---- core CRUD ---------------------------------------------------
	huma.Register(api, huma.Operation{
		OperationID: "list-dashboards",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/dashboards",
		Tags:        []string{"dashboards"},
		Summary:     "List dashboards for a project",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.list)

	huma.Register(api, huma.Operation{
		OperationID:   "create-dashboard",
		Method:        http.MethodPost,
		Path:          "/api/v1/projects/{projectId}/dashboards",
		Tags:          []string{"dashboards"},
		Summary:       "Create a dashboard",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.create)

	huma.Register(api, huma.Operation{
		OperationID: "get-dashboard",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/dashboards/{dashboardId}",
		Tags:        []string{"dashboards"},
		Summary:     "Get a dashboard",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.get)

	huma.Register(api, huma.Operation{
		OperationID: "update-dashboard",
		Method:      http.MethodPut,
		Path:        "/api/v1/projects/{projectId}/dashboards/{dashboardId}",
		Tags:        []string{"dashboards"},
		Summary:     "Update a dashboard",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.update)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-dashboard",
		Method:        http.MethodDelete,
		Path:          "/api/v1/projects/{projectId}/dashboards/{dashboardId}",
		Tags:          []string{"dashboards"},
		Summary:       "Delete a dashboard",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.delete)

	// ---- lifecycle ---------------------------------------------------
	huma.Register(api, huma.Operation{
		OperationID:   "duplicate-dashboard",
		Method:        http.MethodPost,
		Path:          "/api/v1/projects/{projectId}/dashboards/{dashboardId}/duplicate",
		Tags:          []string{"dashboards"},
		Summary:       "Duplicate a dashboard",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.duplicate)

	huma.Register(api, huma.Operation{
		OperationID: "lock-dashboard",
		Method:      http.MethodPost,
		Path:        "/api/v1/projects/{projectId}/dashboards/{dashboardId}/lock",
		Tags:        []string{"dashboards"},
		Summary:     "Lock a dashboard (read-only)",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.lock)

	huma.Register(api, huma.Operation{
		OperationID: "unlock-dashboard",
		Method:      http.MethodPost,
		Path:        "/api/v1/projects/{projectId}/dashboards/{dashboardId}/unlock",
		Tags:        []string{"dashboards"},
		Summary:     "Unlock a dashboard",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.unlock)

	huma.Register(api, huma.Operation{
		OperationID: "export-dashboard",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/dashboards/{dashboardId}/export",
		Tags:        []string{"dashboards"},
		Summary:     "Export a dashboard as JSON",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.export)

	huma.Register(api, huma.Operation{
		OperationID:   "import-dashboard",
		Method:        http.MethodPost,
		Path:          "/api/v1/projects/{projectId}/dashboards/import",
		Tags:          []string{"dashboards"},
		Summary:       "Import a dashboard from an exported JSON",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.importDashboard)

	// ---- query execution --------------------------------------------
	huma.Register(api, huma.Operation{
		OperationID: "execute-dashboard-queries",
		Method:      http.MethodPost,
		Path:        "/api/v1/projects/{projectId}/dashboards/{dashboardId}/execute",
		Tags:        []string{"dashboards"},
		Summary:     "Execute every widget query on a dashboard",
		Description: "Returns a map keyed by widget ID. Supports explicit or relative time ranges plus variable substitution.",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.executeDashboard)

	huma.Register(api, huma.Operation{
		OperationID: "execute-widget-query",
		Method:      http.MethodPost,
		Path:        "/api/v1/projects/{projectId}/dashboards/{dashboardId}/widgets/{widgetId}/execute",
		Tags:        []string{"dashboards"},
		Summary:     "Execute the query for a single widget",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.executeWidget)

	huma.Register(api, huma.Operation{
		OperationID: "get-view-definitions",
		Method:      http.MethodGet,
		Path:        "/api/v1/dashboards/view-definitions",
		Tags:        []string{"dashboards"},
		Summary:     "Get available query-builder view / measure / dimension definitions",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.viewDefinitions)

	huma.Register(api, huma.Operation{
		OperationID: "get-variable-options",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/dashboards/variable-options",
		Tags:        []string{"dashboards"},
		Summary:     "Get distinct values for a dimension (variable dropdowns)",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.variableOptions)

	// ---- templates ---------------------------------------------------
	huma.Register(api, huma.Operation{
		OperationID: "list-dashboard-templates",
		Method:      http.MethodGet,
		Path:        "/api/v1/dashboard-templates",
		Tags:        []string{"dashboard-templates"},
		Summary:     "List active dashboard templates",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.listTemplates)

	huma.Register(api, huma.Operation{
		OperationID: "get-dashboard-template",
		Method:      http.MethodGet,
		Path:        "/api/v1/dashboard-templates/{templateId}",
		Tags:        []string{"dashboard-templates"},
		Summary:     "Get a dashboard template",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getTemplate)

	huma.Register(api, huma.Operation{
		OperationID:   "create-dashboard-from-template",
		Method:        http.MethodPost,
		Path:          "/api/v1/projects/{projectId}/dashboards/from-template",
		Tags:          []string{"dashboard-templates"},
		Summary:       "Create a dashboard from a template",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.createFromTemplate)
}

// ---- shared parsers -------------------------------------------------

func parseProject(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, appErrors.NewValidationError("Invalid project ID", "projectId must be a valid UUID")
	}
	return id, nil
}

func parseDashboard(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, appErrors.NewValidationError("Invalid dashboard ID", "dashboardId must be a valid UUID")
	}
	return id, nil
}

func parseTemplate(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, appErrors.NewValidationError("Invalid template ID", "templateId must be a valid UUID")
	}
	return id, nil
}

// userIDPtr returns a *uuid.UUID when an authenticated user is
// present in context, nil otherwise. Matches the pre-migration
// "optional creator" semantics of CreateDashboard / ImportDashboard /
// CreateFromTemplate.
func userIDPtr(ctx context.Context) *uuid.UUID {
	uid, ok := httpctx.UserID(ctx)
	if !ok {
		return nil
	}
	return &uid
}

// ---- list -----------------------------------------------------------

func (h *handler) list(ctx context.Context, in *ListDashboardsInput) (*ListDashboardsOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}
	filter := &dashboardDomain.DashboardFilter{
		Name:   in.Name,
		Limit:  in.Limit,
		Offset: in.Offset,
	}
	resp, err := h.svc.ListDashboards(ctx, projectID, filter)
	if err != nil {
		h.logger.WarnContext(ctx, "dashboard: list failed", "project_id", projectID, "error", err)
		return nil, err
	}
	return &ListDashboardsOutput{Body: resp}, nil
}

// ---- create ---------------------------------------------------------

func (h *handler) create(ctx context.Context, in *CreateDashboardInput) (*CreateDashboardOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}
	if in.Body.Name == "" {
		return nil, appErrors.NewValidationError("Name is required", "dashboard name is required")
	}
	dash, err := h.svc.CreateDashboard(ctx, projectID, userIDPtr(ctx), &in.Body)
	if err != nil {
		h.logger.WarnContext(ctx, "dashboard: create failed", "project_id", projectID, "name", in.Body.Name, "error", err)
		return nil, err
	}
	return &CreateDashboardOutput{Body: dash}, nil
}

// ---- get ------------------------------------------------------------

func (h *handler) get(ctx context.Context, in *GetDashboardInput) (*GetDashboardOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}
	dashboardID, err := parseDashboard(in.DashboardID)
	if err != nil {
		return nil, err
	}
	dash, err := h.svc.GetDashboardByProject(ctx, projectID, dashboardID)
	if err != nil {
		return nil, err
	}
	return &GetDashboardOutput{Body: dash}, nil
}

// ---- update ---------------------------------------------------------

func (h *handler) update(ctx context.Context, in *UpdateDashboardInput) (*UpdateDashboardOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}
	dashboardID, err := parseDashboard(in.DashboardID)
	if err != nil {
		return nil, err
	}
	dash, err := h.svc.UpdateDashboard(ctx, projectID, dashboardID, &in.Body)
	if err != nil {
		return nil, err
	}
	return &UpdateDashboardOutput{Body: dash}, nil
}

// ---- delete ---------------------------------------------------------

func (h *handler) delete(ctx context.Context, in *DeleteDashboardInput) (*DeleteDashboardOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}
	dashboardID, err := parseDashboard(in.DashboardID)
	if err != nil {
		return nil, err
	}
	if err := h.svc.DeleteDashboard(ctx, projectID, dashboardID); err != nil {
		return nil, err
	}
	return &DeleteDashboardOutput{}, nil
}

// ---- duplicate ------------------------------------------------------

func (h *handler) duplicate(ctx context.Context, in *DuplicateDashboardInput) (*DuplicateDashboardOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}
	dashboardID, err := parseDashboard(in.DashboardID)
	if err != nil {
		return nil, err
	}
	if in.Body.Name == "" {
		return nil, appErrors.NewValidationError("Name is required", "dashboard name is required")
	}
	dash, err := h.svc.DuplicateDashboard(ctx, projectID, dashboardID, &in.Body)
	if err != nil {
		return nil, err
	}
	return &DuplicateDashboardOutput{Body: dash}, nil
}

// ---- lock / unlock --------------------------------------------------

func (h *handler) lock(ctx context.Context, in *LockDashboardInput) (*LockDashboardOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}
	dashboardID, err := parseDashboard(in.DashboardID)
	if err != nil {
		return nil, err
	}
	dash, err := h.svc.LockDashboard(ctx, projectID, dashboardID)
	if err != nil {
		return nil, err
	}
	return &LockDashboardOutput{Body: dash}, nil
}

func (h *handler) unlock(ctx context.Context, in *UnlockDashboardInput) (*UnlockDashboardOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}
	dashboardID, err := parseDashboard(in.DashboardID)
	if err != nil {
		return nil, err
	}
	dash, err := h.svc.UnlockDashboard(ctx, projectID, dashboardID)
	if err != nil {
		return nil, err
	}
	return &UnlockDashboardOutput{Body: dash}, nil
}

// ---- export / import ------------------------------------------------

func (h *handler) export(ctx context.Context, in *ExportDashboardInput) (*ExportDashboardOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}
	dashboardID, err := parseDashboard(in.DashboardID)
	if err != nil {
		return nil, err
	}
	exp, err := h.svc.ExportDashboard(ctx, projectID, dashboardID)
	if err != nil {
		return nil, err
	}
	return &ExportDashboardOutput{Body: exp}, nil
}

func (h *handler) importDashboard(ctx context.Context, in *ImportDashboardInput) (*ImportDashboardOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}
	dash, err := h.svc.ImportDashboard(ctx, projectID, userIDPtr(ctx), &in.Body)
	if err != nil {
		h.logger.WarnContext(ctx, "dashboard: import failed", "project_id", projectID, "error", err)
		return nil, err
	}
	return &ImportDashboardOutput{Body: dash}, nil
}

// ---- query execution ------------------------------------------------

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

func (h *handler) executeDashboard(ctx context.Context, in *ExecuteDashboardInput) (*ExecuteDashboardOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}
	dashboardID, err := parseDashboard(in.DashboardID)
	if err != nil {
		return nil, err
	}
	req := &dashboardDomain.QueryExecutionRequest{
		ProjectID:          projectID,
		DashboardID:        dashboardID,
		DashboardTimeRange: in.Body.toDomainTimeRange(),
		ForceRefresh:       in.Body.ForceRefresh,
		VariableValues:     in.Body.VariableValues,
	}
	results, err := h.query.ExecuteDashboardQueries(ctx, req)
	if err != nil {
		return nil, err
	}
	return &ExecuteDashboardOutput{Body: results}, nil
}

func (h *handler) executeWidget(ctx context.Context, in *ExecuteWidgetInput) (*ExecuteWidgetOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}
	dashboardID, err := parseDashboard(in.DashboardID)
	if err != nil {
		return nil, err
	}
	if in.WidgetID == "" {
		return nil, appErrors.NewValidationError("Invalid widget ID", "widgetId is required")
	}
	widgetID := in.WidgetID
	req := &dashboardDomain.QueryExecutionRequest{
		ProjectID:          projectID,
		DashboardID:        dashboardID,
		WidgetID:           &widgetID,
		DashboardTimeRange: in.Body.toDomainTimeRange(),
		ForceRefresh:       in.Body.ForceRefresh,
	}
	results, err := h.query.ExecuteDashboardQueries(ctx, req)
	if err != nil {
		return nil, err
	}
	result, ok := results.Results[widgetID]
	if !ok {
		return nil, appErrors.NewNotFoundError("widget")
	}
	return &ExecuteWidgetOutput{Body: result}, nil
}

func (h *handler) viewDefinitions(ctx context.Context, _ *struct{}) (*ViewDefinitionsOutput, error) {
	resp, err := h.query.GetViewDefinitions(ctx)
	if err != nil {
		return nil, err
	}
	return &ViewDefinitionsOutput{Body: resp}, nil
}

func (h *handler) variableOptions(ctx context.Context, in *VariableOptionsInput) (*VariableOptionsOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}
	if in.View == "" {
		return nil, appErrors.NewValidationError("View required", "view query parameter is required")
	}
	if in.Dimension == "" {
		return nil, appErrors.NewValidationError("Dimension required", "dimension query parameter is required")
	}
	limit := in.Limit
	if limit <= 0 {
		limit = 100
	}
	req := &dashboardDomain.VariableOptionsRequest{
		ProjectID: projectID,
		View:      dashboardDomain.ViewType(in.View),
		Dimension: in.Dimension,
		Limit:     limit,
	}
	result, err := h.query.GetVariableOptions(ctx, req)
	if err != nil {
		return nil, err
	}
	return &VariableOptionsOutput{Body: result}, nil
}

// ---- templates ------------------------------------------------------

func (h *handler) listTemplates(ctx context.Context, _ *struct{}) (*ListTemplatesOutput, error) {
	templates, err := h.template.ListTemplates(ctx)
	if err != nil {
		return nil, err
	}
	return &ListTemplatesOutput{Body: templates}, nil
}

func (h *handler) getTemplate(ctx context.Context, in *GetTemplateInput) (*GetTemplateOutput, error) {
	templateID, err := parseTemplate(in.TemplateID)
	if err != nil {
		return nil, err
	}
	tmpl, err := h.template.GetTemplate(ctx, templateID)
	if err != nil {
		return nil, err
	}
	return &GetTemplateOutput{Body: tmpl}, nil
}

func (h *handler) createFromTemplate(ctx context.Context, in *CreateFromTemplateInput) (*CreateFromTemplateOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}
	dash, err := h.template.CreateFromTemplate(ctx, projectID, userIDPtr(ctx), &in.Body)
	if err != nil {
		h.logger.WarnContext(ctx, "dashboard: create-from-template failed", "project_id", projectID, "template_id", in.Body.TemplateID, "error", err)
		return nil, err
	}
	return &CreateFromTemplateOutput{Body: dash}, nil
}
