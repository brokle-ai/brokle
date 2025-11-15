package observability

import (
	"strconv"

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

	// Pagination
	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil || l < 1 || l > 1000 {
			response.ValidationError(c, "invalid limit", "limit must be between 1 and 1000")
			return
		}
		limit = l
	}
	filter.Limit = limit

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		o, err := strconv.Atoi(offsetStr)
		if err != nil || o < 0 {
			response.ValidationError(c, "invalid offset", "offset must be >= 0")
			return
		}
		offset = o
	}
	filter.Offset = offset

	scores, err := h.services.GetScoreService().GetScoresByFilter(c.Request.Context(), filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list scores")
		response.Error(c, err)
		return
	}

	response.Success(c, scores)
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
		h.logger.WithError(err).Error("Failed to get score")
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
		h.logger.WithError(err).Error("Failed to update score")
		response.Error(c, err)
		return
	}

	// Fetch updated score
	updated, err := h.services.GetScoreService().GetScoreByID(c.Request.Context(), scoreID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to fetch updated score")
		response.Error(c, err)
		return
	}

	response.Success(c, updated)
}
