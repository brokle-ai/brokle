package app

import (
	"log/slog"

	dashboardService "brokle/internal/core/services/dashboard"
)

// ProvideDashboardServices wires the dashboard-domain services for
// the dashboard CRUD plane plus the widget query engine and template
// catalog.
func ProvideDashboardServices(
	dashboardRepos *DashboardRepositories,
	logger *slog.Logger,
) *DashboardServices {
	dashboardSvc := dashboardService.NewDashboardService(
		dashboardRepos.Dashboard,
		logger,
	)

	widgetQuerySvc := dashboardService.NewWidgetQueryService(
		dashboardRepos.Dashboard,
		dashboardRepos.WidgetQuery,
		logger,
	)

	templateSvc := dashboardService.NewTemplateService(
		dashboardRepos.Template,
		dashboardRepos.Dashboard,
		logger,
	)

	return &DashboardServices{
		Dashboard:   dashboardSvc,
		WidgetQuery: widgetQuerySvc,
		Template:    templateSvc,
	}
}
