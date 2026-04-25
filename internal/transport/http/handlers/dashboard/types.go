package dashboard

import (
	"time"

	dashboardDomain "brokle/internal/core/domain/dashboard"
	"brokle/pkg/response"
)

// listDashboardsBody is the canonical {data, pagination} wire envelope
// returned by GET /api/v1/projects/{projectId}/dashboards.
type listDashboardsBody struct {
	Data       []*dashboardDomain.Dashboard `json:"data"`
	Pagination *response.Pagination         `json:"pagination"`
}

// timeRangeBody is the shared time-range shape for execute endpoints.
type timeRangeBody struct {
	From     *time.Time `json:"from,omitempty"`
	To       *time.Time `json:"to,omitempty"`
	Relative string     `json:"relative,omitempty"`
}

// executeDashboardBody — POST /dashboards/{id}/execute  and  widgets/{widgetId}/execute
type executeDashboardBody struct {
	DashboardTimeRange *timeRangeBody `json:"time_range,omitempty"`
	ForceRefresh       bool           `json:"force_refresh,omitempty"`
	VariableValues     map[string]any `json:"variable_values,omitempty"`
}
