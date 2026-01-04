package overview

import (
	"log/slog"

	"brokle/internal/config"
	"brokle/internal/core/domain/analytics"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/response"
	"brokle/pkg/ulid"

	"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests for project overview
type Handler struct {
	config          *config.Config
	logger          *slog.Logger
	overviewService analytics.OverviewService
}

// NewHandler creates a new overview handler
func NewHandler(
	config *config.Config,
	logger *slog.Logger,
	overviewService analytics.OverviewService,
) *Handler {
	return &Handler{
		config:          config,
		logger:          logger,
		overviewService: overviewService,
	}
}

// OverviewRequest represents the query parameters for the overview endpoint
type OverviewRequest struct {
	TimeRange string `form:"time_range" binding:"omitempty,oneof=24h 7d 30d"`
}

// GetOverview handles GET /api/v1/projects/:projectId/overview
// @Summary Get project overview
// @Description Get comprehensive overview data for a project including stats, charts, and onboarding status
// @Tags Projects
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID"
// @Param time_range query string false "Time range for data" default("24h") Enums(24h,7d,30d)
// @Success 200 {object} response.SuccessResponse{data=analytics.OverviewResponse} "Project overview data"
// @Failure 400 {object} response.ErrorResponse "Bad request - invalid parameters"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - no access to project"
// @Failure 404 {object} response.ErrorResponse "Project not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/projects/{projectId}/overview [get]
func (h *Handler) GetOverview(c *gin.Context) {
	// Parse project ID from path parameter
	projectIDStr := c.Param("projectId")
	if projectIDStr == "" {
		response.Error(c, appErrors.NewValidationError("project_id is required", "projectId path parameter is missing"))
		return
	}

	projectID, err := ulid.Parse(projectIDStr)
	if err != nil {
		response.Error(c, appErrors.NewValidationError("Invalid project ID", "projectId must be a valid ULID"))
		return
	}

	// Parse query parameters
	var req OverviewRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, appErrors.NewValidationError("Invalid query parameters", err.Error()))
		return
	}

	// Default to 24h time range
	timeRange := analytics.ParseTimeRange(req.TimeRange)

	// Get overview data from service
	overview, err := h.overviewService.GetOverview(c.Request.Context(), projectID, timeRange)
	if err != nil {
		h.logger.Error("failed to get project overview",
			"error", err,
			"project_id", projectID,
			"time_range", timeRange,
		)
		response.Error(c, appErrors.NewInternalError("Failed to get project overview", err))
		return
	}

	response.Success(c, overview)
}
