package observability

import (
	"time"

	"github.com/gin-gonic/gin"

	"brokle/internal/core/domain/observability"
	"brokle/pkg/response"
)

// Quality Score Handlers for Dashboard (JWT-authenticated, read + update operations)

// ListProjectScores handles GET /api/v1/projects/:projectId/scores
// @Summary List quality scores for a project
// @Description Retrieve paginated list of quality scores scoped to a project
// @Tags Scores
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param projectId path string true "Project ID"
// @Param trace_id query string false "Filter by trace ID"
// @Param span_id query string false "Filter by span ID"
// @Param name query string false "Filter by score name"
// @Param source query string false "Filter by source (API, AUTO, HUMAN, EVAL)"
// @Param data_type query string false "Filter by data type (NUMERIC, CATEGORICAL, BOOLEAN)"
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Page size (default 50, max 1000)"
// @Success 200 {object} response.APIResponse{data=[]observability.Score} "List of scores"
// @Failure 400 {object} response.APIResponse{error=response.APIError} "Invalid parameters"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/projects/{projectId}/scores [get]
func (h *Handler) ListProjectScores(c *gin.Context) {
	projectID := c.Param("projectId")
	if projectID == "" {
		response.ValidationError(c, "invalid project_id", "project_id is required")
		return
	}

	filter := &observability.ScoreFilter{
		ProjectID: projectID,
	}

	if traceID := c.Query("trace_id"); traceID != "" {
		filter.TraceID = &traceID
	}
	if spanID := c.Query("span_id"); spanID != "" {
		filter.SpanID = &spanID
	}
	if name := c.Query("name"); name != "" {
		filter.Name = &name
	}
	if source := c.Query("source"); source != "" {
		filter.Source = &source
	}
	if dataType := c.Query("data_type"); dataType != "" {
		filter.DataType = &dataType
	}

	params := response.ParsePaginationParams(
		c.Query("page"),
		c.Query("limit"),
		c.Query("sort_by"),
		c.Query("sort_dir"),
	)
	filter.Params = params

	scores, err := h.services.GetScoreService().GetScoresByFilter(c.Request.Context(), filter)
	if err != nil {
		response.Error(c, err)
		return
	}

	totalCount, err := h.services.GetScoreService().CountScores(c.Request.Context(), filter)
	if err != nil {
		response.Error(c, err)
		return
	}

	paginationMeta := response.NewPagination(params.Page, params.Limit, totalCount)
	response.SuccessWithPagination(c, scores, paginationMeta)
}

// ListScores handles GET /api/v1/scores
// @Summary List quality scores with filtering
// @Description Retrieve paginated list of quality scores
// @Tags Scores
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param trace_id query string false "Filter by trace ID"
// @Param span_id query string false "Filter by span ID"
// @Param session_id query string false "Filter by session ID"
// @Param name query string false "Filter by score name"
// @Param source query string false "Filter by source (API, AUTO, HUMAN, EVAL)"
// @Param data_type query string false "Filter by data type (NUMERIC, CATEGORICAL, BOOLEAN)"
// @Param limit query int false "Limit (default 50, max 1000)"
// @Param offset query int false "Offset (default 0)"
// @Success 200 {object} response.APIResponse{data=[]observability.Score} "List of scores"
// @Failure 400 {object} response.APIResponse{error=response.APIError} "Invalid parameters"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/scores [get]
func (h *Handler) ListScores(c *gin.Context) {
	filter := &observability.ScoreFilter{}

	if traceID := c.Query("trace_id"); traceID != "" {
		filter.TraceID = &traceID
	}
	if spanID := c.Query("span_id"); spanID != "" {
		filter.SpanID = &spanID
	}
	if name := c.Query("name"); name != "" {
		filter.Name = &name
	}
	if source := c.Query("source"); source != "" {
		filter.Source = &source
	}
	if dataType := c.Query("data_type"); dataType != "" {
		filter.DataType = &dataType
	}

	params := response.ParsePaginationParams(
		c.Query("page"),
		c.Query("limit"),
		c.Query("sort_by"),
		c.Query("sort_dir"),
	)
	filter.Params = params

	scores, err := h.services.GetScoreService().GetScoresByFilter(c.Request.Context(), filter)
	if err != nil {
		response.Error(c, err)
		return
	}

	totalCount, err := h.services.GetScoreService().CountScores(c.Request.Context(), filter)
	if err != nil {
		response.Error(c, err)
		return
	}

	paginationMeta := response.NewPagination(params.Page, params.Limit, totalCount)

	response.SuccessWithPagination(c, scores, paginationMeta)
}

// GetScore handles GET /api/v1/scores/:id
// @Summary Get quality score by ID
// @Description Retrieve detailed score information
// @Tags Scores
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Score ID"
// @Success 200 {object} response.APIResponse{data=observability.Score} "Score details"
// @Failure 404 {object} response.APIResponse{error=response.APIError} "Score not found"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/scores/{id} [get]
func (h *Handler) GetScore(c *gin.Context) {
	scoreID := c.Param("id")
	if scoreID == "" {
		response.ValidationError(c, "invalid score_id", "score_id is required")
		return
	}

	score, err := h.services.GetScoreService().GetScoreByID(c.Request.Context(), scoreID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, score)
}

// UpdateScore handles PUT /api/v1/scores/:id
// @Summary Update quality score by ID
// @Description Update an existing score (for corrections/enrichment after initial creation)
// @Tags Scores
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Score ID"
// @Param score body observability.Score true "Updated score data"
// @Success 200 {object} response.APIResponse{data=observability.Score} "Updated score"
// @Failure 400 {object} response.APIResponse{error=response.APIError} "Invalid request"
// @Failure 404 {object} response.APIResponse{error=response.APIError} "Score not found"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/scores/{id} [put]
func (h *Handler) UpdateScore(c *gin.Context) {
	scoreID := c.Param("id")
	if scoreID == "" {
		response.ValidationError(c, "invalid score_id", "score_id is required")
		return
	}

	var score observability.Score
	if err := c.ShouldBindJSON(&score); err != nil {
		response.ValidationError(c, "invalid request body", err.Error())
		return
	}

	// Ensure ID matches path parameter
	score.ID = scoreID

	if err := h.services.GetScoreService().UpdateScore(c.Request.Context(), &score); err != nil {
		response.Error(c, err)
		return
	}

	updated, err := h.services.GetScoreService().GetScoreByID(c.Request.Context(), scoreID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, updated)
}

// GetScoreAnalytics handles GET /api/v1/projects/:projectId/scores/analytics
// @Summary Get score analytics for a project
// @Description Retrieve comprehensive analytics for a score including statistics, time series, and distribution
// @Tags Scores
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param projectId path string true "Project ID"
// @Param score_name query string true "Score name to analyze"
// @Param compare_score_name query string false "Optional second score for comparison"
// @Param from_timestamp query string false "Start of time range (RFC3339)"
// @Param to_timestamp query string false "End of time range (RFC3339)"
// @Param interval query string false "Aggregation interval (hour, day, week)"
// @Success 200 {object} response.APIResponse{data=observability.ScoreAnalyticsResponse} "Analytics data"
// @Failure 400 {object} response.APIResponse{error=response.APIError} "Invalid parameters"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/projects/{projectId}/scores/analytics [get]
func (h *Handler) GetScoreAnalytics(c *gin.Context) {
	projectID := c.Param("projectId")
	if projectID == "" {
		response.ValidationError(c, "invalid project_id", "project_id is required")
		return
	}

	scoreName := c.Query("score_name")
	if scoreName == "" {
		response.ValidationError(c, "score_name is required", "score_name query parameter is required")
		return
	}

	filter := &observability.ScoreAnalyticsFilter{
		ProjectID: projectID,
		ScoreName: scoreName,
		Interval:  c.DefaultQuery("interval", "day"),
	}

	if compareScoreName := c.Query("compare_score_name"); compareScoreName != "" {
		filter.CompareScoreName = &compareScoreName
	}

	if fromStr := c.Query("from_timestamp"); fromStr != "" {
		fromTime, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			response.ValidationError(c, "invalid from_timestamp", "must be RFC3339 format (e.g., 2024-01-15T00:00:00Z)")
			return
		}
		filter.FromTimestamp = &fromTime
	}
	if toStr := c.Query("to_timestamp"); toStr != "" {
		toTime, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			response.ValidationError(c, "invalid to_timestamp", "must be RFC3339 format (e.g., 2024-01-15T23:59:59Z)")
			return
		}
		filter.ToTimestamp = &toTime
	}

	analytics, err := h.services.GetScoreAnalyticsService().GetAnalytics(c.Request.Context(), filter)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, analytics)
}

// GetScoreNames handles GET /api/v1/projects/:projectId/scores/names
// @Summary Get distinct score names for a project
// @Description Retrieve all unique score names available in a project (for dropdown selection)
// @Tags Scores
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param projectId path string true "Project ID"
// @Success 200 {object} response.APIResponse{data=[]string} "List of score names"
// @Failure 400 {object} response.APIResponse{error=response.APIError} "Invalid parameters"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/projects/{projectId}/scores/names [get]
func (h *Handler) GetScoreNames(c *gin.Context) {
	projectID := c.Param("projectId")
	if projectID == "" {
		response.ValidationError(c, "invalid project_id", "project_id is required")
		return
	}

	names, err := h.services.GetScoreAnalyticsService().GetDistinctScoreNames(c.Request.Context(), projectID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, names)
}
