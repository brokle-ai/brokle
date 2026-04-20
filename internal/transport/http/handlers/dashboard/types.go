package dashboard

import (
	dashboardDomain "brokle/internal/core/domain/dashboard"
	"time"
)

// Huma operation types for the dashboard package.

type ListDashboardsInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	Name      string `query:"name" required:"false" doc:"Filter by name (partial match)"`
	Limit     int    `query:"limit" required:"false" minimum:"0" doc:"Items per page (default 50)"`
	Offset    int    `query:"offset" required:"false" minimum:"0" doc:"Pagination offset"`
}

type ListDashboardsOutput struct {
	Body *dashboardDomain.DashboardListResponse
}

type CreateDashboardInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	Body      dashboardDomain.CreateDashboardRequest
}

type CreateDashboardOutput struct {
	Body *dashboardDomain.Dashboard
}

type GetDashboardInput struct {
	ProjectID   string `path:"projectId" format:"uuid"`
	DashboardID string `path:"dashboardId" format:"uuid"`
}

type GetDashboardOutput struct {
	Body *dashboardDomain.Dashboard
}

type UpdateDashboardInput struct {
	ProjectID   string `path:"projectId" format:"uuid"`
	DashboardID string `path:"dashboardId" format:"uuid"`
	Body        dashboardDomain.UpdateDashboardRequest
}

type UpdateDashboardOutput struct {
	Body *dashboardDomain.Dashboard
}

type DeleteDashboardInput struct {
	ProjectID   string `path:"projectId" format:"uuid"`
	DashboardID string `path:"dashboardId" format:"uuid"`
}

type DeleteDashboardOutput struct{}

type DuplicateDashboardInput struct {
	ProjectID   string `path:"projectId" format:"uuid"`
	DashboardID string `path:"dashboardId" format:"uuid"`
	Body        dashboardDomain.DuplicateDashboardRequest
}

type DuplicateDashboardOutput struct {
	Body *dashboardDomain.Dashboard
}

type LockDashboardInput struct {
	ProjectID   string `path:"projectId" format:"uuid"`
	DashboardID string `path:"dashboardId" format:"uuid"`
}

type LockDashboardOutput struct {
	Body *dashboardDomain.Dashboard
}

type UnlockDashboardInput = LockDashboardInput

type UnlockDashboardOutput = LockDashboardOutput

type ExportDashboardInput struct {
	ProjectID   string `path:"projectId" format:"uuid"`
	DashboardID string `path:"dashboardId" format:"uuid"`
}

type ExportDashboardOutput struct {
	Body *dashboardDomain.DashboardExport
}

type ImportDashboardInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	Body      dashboardDomain.DashboardImportRequest
}

type ImportDashboardOutput struct {
	Body *dashboardDomain.Dashboard
}

type timeRangeBody struct {
	From     *time.Time `json:"from,omitempty"`
	To       *time.Time `json:"to,omitempty"`
	Relative string     `json:"relative,omitempty" doc:"Relative time preset (1h, 24h, 7d, 30d)"`
}

type executeDashboardBody struct {
	DashboardTimeRange *timeRangeBody `json:"time_range,omitempty"`
	ForceRefresh       bool           `json:"force_refresh,omitempty"`
	VariableValues     map[string]any `json:"variable_values,omitempty"`
}

type ExecuteDashboardInput struct {
	ProjectID   string `path:"projectId" format:"uuid"`
	DashboardID string `path:"dashboardId" format:"uuid"`
	Body        executeDashboardBody
}

type ExecuteDashboardOutput struct {
	Body *dashboardDomain.DashboardQueryResults
}

type ExecuteWidgetInput struct {
	ProjectID   string `path:"projectId" format:"uuid"`
	DashboardID string `path:"dashboardId" format:"uuid"`
	WidgetID    string `path:"widgetId" doc:"Widget ID within the dashboard"`
	Body        executeDashboardBody
}

type ExecuteWidgetOutput struct {
	Body *dashboardDomain.DashboardQueryResult
}

type ViewDefinitionsOutput struct {
	Body *dashboardDomain.ViewDefinitionResponse
}

type VariableOptionsInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	View      string `query:"view" doc:"View type (traces, spans, scores)"`
	Dimension string `query:"dimension" doc:"Dimension field name"`
	Limit     int    `query:"limit" required:"false" minimum:"1" doc:"Maximum options (default 100)"`
}

type VariableOptionsOutput struct {
	Body *dashboardDomain.VariableOptionsResponse
}

type ListTemplatesOutput struct {
	Body []*dashboardDomain.Template
}

type GetTemplateInput struct {
	TemplateID string `path:"templateId" format:"uuid"`
}

type GetTemplateOutput struct {
	Body *dashboardDomain.Template
}

type CreateFromTemplateInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	Body      dashboardDomain.CreateFromTemplateRequest
}

type CreateFromTemplateOutput struct {
	Body *dashboardDomain.Dashboard
}
