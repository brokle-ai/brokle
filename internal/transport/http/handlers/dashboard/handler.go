package dashboard

import (
	"log/slog"

	"brokle/internal/config"
	dashboardDomain "brokle/internal/core/domain/dashboard"
)

// Handler handles dashboard HTTP requests
type Handler struct {
	config       *config.Config
	logger       *slog.Logger
	service      dashboardDomain.DashboardService
	queryService dashboardDomain.WidgetQueryService
}

// NewHandler creates a new dashboard handler instance
func NewHandler(
	cfg *config.Config,
	logger *slog.Logger,
	service dashboardDomain.DashboardService,
	widgetQueryService dashboardDomain.WidgetQueryService,
) *Handler {
	return &Handler{
		config:       cfg,
		logger:       logger,
		service:      service,
		queryService: widgetQueryService,
	}
}
