package observability

import (
	"github.com/gin-gonic/gin"

	"brokle/internal/core/domain/observability"
	"brokle/pkg/response"
)

// Quality Score Handlers for Dashboard (JWT-authenticated, read + update operations)

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

	// Trace ID filter
	if traceID := c.Query("trace_id"); traceID != "" {
		filter.TraceID = &traceID
	}

	// Span ID filter
	if spanID := c.Query("span_id"); spanID != "" {
		filter.SpanID = &spanID
	}

	// Name filter
	if name := c.Query("name"); name != "" {
		filter.Name = &name
	}

	// Source filter
	if source := c.Query("source"); source != "" {
		filter.Source = &source
	}

	// Data type filter
	if dataType := c.Query("data_type"); dataType != "" {
		filter.DataType = &dataType
	}

	// Offset pagination
	params := response.ParsePaginationParams(
		c.Query("page"),
		c.Query("limit"),
		c.Query("sort_by"),
		c.Query("sort_dir"),
	)

	// Set embedded pagination fields
	filter.Params = params

	// Get scores from service
	scores, err := h.services.GetScoreService().GetScoresByFilter(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to list scores", "error", err)
		response.Error(c, err)
		return
	}

	// Get total count for pagination metadata
	totalCount, err := h.services.GetScoreService().CountScores(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to count scores", "error", err)
		response.Error(c, err)
		return
	}

	// Build pagination metadata (NewPagination calculates has_next, has_prev, total_pages)
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
		h.logger.Error("Failed to get score", "error", err)
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

	// Update via service
	if err := h.services.GetScoreService().UpdateScore(c.Request.Context(), &score); err != nil {
		h.logger.Error("Failed to update score", "error", err)
		response.Error(c, err)
		return
	}

	// Fetch updated score
	updated, err := h.services.GetScoreService().GetScoreByID(c.Request.Context(), scoreID)
	if err != nil {
		h.logger.Error("Failed to fetch updated score", "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, updated)
}
