package observability

import (
	"time"

	"github.com/gin-gonic/gin"

	"brokle/internal/core/domain/observability"
	"brokle/internal/transport/http/middleware"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/response"
	"brokle/pkg/ulid"
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
// @Param type query string false "Filter by type (NUMERIC, CATEGORICAL, BOOLEAN)"
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
	if scoreType := c.Query("type"); scoreType != "" {
		filter.Type = &scoreType
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
// @Param type query string false "Filter by type (NUMERIC, CATEGORICAL, BOOLEAN)"
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
	if scoreType := c.Query("type"); scoreType != "" {
		filter.Type = &scoreType
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

// CreateTraceScore handles POST /api/v1/traces/:id/scores
// @Summary Create annotation score for a trace
// @Description Creates a human annotation score for a trace (JWT-authenticated dashboard endpoint)
// @Tags Scores
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Trace ID"
// @Param project_id query string true "Project ID"
// @Param request body CreateAnnotationRequest true "Annotation data"
// @Success 201 {object} response.APIResponse{data=AnnotationResponse} "Created annotation"
// @Failure 400 {object} response.APIResponse{error=response.APIError} "Invalid parameters"
// @Failure 401 {object} response.APIResponse{error=response.APIError} "Authentication required"
// @Failure 404 {object} response.APIResponse{error=response.APIError} "Trace not found"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/traces/{id}/scores [post]
func (h *Handler) CreateTraceScore(c *gin.Context) {
	ctx := c.Request.Context()

	// Get trace ID from path
	traceID := c.Param("id")
	if traceID == "" {
		response.Error(c, appErrors.NewValidationError("id", "is required"))
		return
	}

	// Get project ID from query
	projectIDStr := c.Query("project_id")
	if projectIDStr == "" {
		response.Error(c, appErrors.NewValidationError("project_id", "is required"))
		return
	}

	// Get user ID from JWT (for audit trail)
	userID, exists := middleware.GetUserIDULID(c)
	if !exists {
		response.Error(c, appErrors.NewUnauthorizedError("authentication required"))
		return
	}

	// Parse request body
	var req CreateAnnotationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Get root span to retrieve organization ID
	rootSpan, err := h.services.GetTraceService().GetRootSpan(ctx, traceID)
	if err != nil {
		response.Error(c, err)
		return
	}

	// Validate that the trace belongs to the requested project
	if rootSpan.ProjectID != projectIDStr {
		response.Error(c, appErrors.NewValidationError("project_id", "does not match trace's project"))
		return
	}

	// Build the score entity
	userIDStr := userID.String()
	score := &observability.Score{
		ID:             ulid.New().String(),
		ProjectID:      projectIDStr,
		OrganizationID: rootSpan.OrganizationID,
		TraceID:        &traceID,
		SpanID:         &rootSpan.SpanID, // Use the actual root span's ID
		Name:           req.Name,
		Value:          req.Value,
		StringValue:    req.StringValue,
		DataType:       req.DataType,
		Source:         observability.ScoreSourceAnnotation,
		Reason:         req.Reason,
		Metadata:       "{}",
		CreatedBy:      &userIDStr,
		Timestamp:      time.Now(),
	}

	// Create the score
	if err := h.services.GetScoreService().CreateScore(ctx, score); err != nil {
		response.Error(c, err)
		return
	}

	h.logger.Info("annotation created",
		"score_id", score.ID,
		"project_id", projectIDStr,
		"trace_id", traceID,
		"user_id", userID,
		"name", req.Name,
	)

	// Return response
	response.Created(c, &AnnotationResponse{
		ID:          score.ID,
		ProjectID:   score.ProjectID,
		TraceID:     score.TraceID,
		SpanID:      score.SpanID,
		Name:        score.Name,
		Value:       score.Value,
		StringValue: score.StringValue,
		DataType:    score.DataType,
		Source:      score.Source,
		Reason:      score.Reason,
		CreatedBy:   score.CreatedBy,
		Timestamp:   score.Timestamp.Format(time.RFC3339),
	})
}

// GetTraceScores handles GET /api/v1/traces/:id/scores
// @Summary List scores for a trace
// @Description Retrieve all scores (annotations and automated) for a specific trace
// @Tags Scores
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Trace ID"
// @Param project_id query string true "Project ID"
// @Success 200 {object} response.APIResponse{data=[]AnnotationResponse} "List of scores"
// @Failure 400 {object} response.APIResponse{error=response.APIError} "Invalid parameters"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/traces/{id}/scores [get]
func (h *Handler) GetTraceScores(c *gin.Context) {
	ctx := c.Request.Context()

	// Get trace ID from path
	traceID := c.Param("id")
	if traceID == "" {
		response.Error(c, appErrors.NewValidationError("id", "is required"))
		return
	}

	// Get project ID from query (for authorization)
	projectIDStr := c.Query("project_id")
	if projectIDStr == "" {
		response.Error(c, appErrors.NewValidationError("project_id", "is required"))
		return
	}

	// Get scores for the trace
	scores, err := h.services.GetScoreService().GetScoresByTraceID(ctx, traceID)
	if err != nil {
		response.Error(c, err)
		return
	}

	// Convert to response format
	responses := make([]*AnnotationResponse, 0, len(scores))
	for _, score := range scores {
		responses = append(responses, &AnnotationResponse{
			ID:          score.ID,
			ProjectID:   score.ProjectID,
			TraceID:     score.TraceID,
			SpanID:      score.SpanID,
			Name:        score.Name,
			Value:       score.Value,
			StringValue: score.StringValue,
			DataType:    score.DataType,
			Source:      score.Source,
			Reason:      score.Reason,
			CreatedBy:   score.CreatedBy,
			Timestamp:   score.Timestamp.Format(time.RFC3339),
		})
	}

	response.Success(c, responses)
}

// DeleteTraceScore handles DELETE /api/v1/traces/:id/scores/:score_id
// @Summary Delete annotation score
// @Description Deletes a human annotation score. Only the creator can delete their annotation.
// @Tags Scores
// @Produce json
// @Security BearerAuth
// @Param id path string true "Trace ID"
// @Param score_id path string true "Score ID"
// @Param project_id query string true "Project ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.APIResponse{error=response.APIError} "Invalid parameters"
// @Failure 401 {object} response.APIResponse{error=response.APIError} "Authentication required"
// @Failure 403 {object} response.APIResponse{error=response.APIError} "Not the annotation owner"
// @Failure 404 {object} response.APIResponse{error=response.APIError} "Score not found"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/traces/{id}/scores/{score_id} [delete]
func (h *Handler) DeleteTraceScore(c *gin.Context) {
	ctx := c.Request.Context()

	// Get trace ID from path
	traceID := c.Param("id")
	if traceID == "" {
		response.Error(c, appErrors.NewValidationError("id", "is required"))
		return
	}

	// Get score ID from path
	scoreID := c.Param("score_id")
	if scoreID == "" {
		response.Error(c, appErrors.NewValidationError("score_id", "is required"))
		return
	}

	// Get project ID from query
	projectIDStr := c.Query("project_id")
	if projectIDStr == "" {
		response.Error(c, appErrors.NewValidationError("project_id", "is required"))
		return
	}

	// Get user ID from JWT
	userID, exists := middleware.GetUserIDULID(c)
	if !exists {
		response.Error(c, appErrors.NewUnauthorizedError("authentication required"))
		return
	}

	// Get the score to verify ownership
	score, err := h.services.GetScoreService().GetScoreByID(ctx, scoreID)
	if err != nil {
		response.Error(c, err)
		return
	}

	// Verify the score belongs to the trace
	if score.TraceID == nil || *score.TraceID != traceID {
		response.Error(c, appErrors.NewNotFoundError("score"))
		return
	}

	// Only allow deletion of annotation scores by their creator
	if score.Source != observability.ScoreSourceAnnotation {
		response.Error(c, appErrors.NewForbiddenError("only annotation scores can be deleted"))
		return
	}

	if score.CreatedBy == nil || *score.CreatedBy != userID.String() {
		response.Error(c, appErrors.NewForbiddenError("only the creator can delete this annotation"))
		return
	}

	// Delete the score
	if err := h.services.GetScoreService().DeleteScore(ctx, scoreID); err != nil {
		response.Error(c, err)
		return
	}

	h.logger.Info("annotation deleted",
		"score_id", scoreID,
		"project_id", projectIDStr,
		"trace_id", traceID,
		"user_id", userID,
	)

	response.NoContent(c)
}
