package observability

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"brokle/internal/core/domain/observability"
	"brokle/pkg/response"
)

// Analytics Handlers for Dashboard (JWT-authenticated, read-only queries)

// ===== Trace Analytics =====

// ListTraces handles GET /api/v1/analytics/traces
// @Summary List traces for a project
// @Description Retrieve paginated list of traces with optional filtering
// @Tags Dashboard - Analytics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param project_id query string true "Project ID"
// @Param session_id query string false "Filter by session ID"
// @Param user_id query string false "Filter by user ID"
// @Param start_time query int64 false "Start time (Unix timestamp)"
// @Param end_time query int64 false "End time (Unix timestamp)"
// @Param limit query int false "Limit (default 50, max 1000)"
// @Param offset query int false "Offset (default 0)"
// @Success 200 {object} response.APIResponse{data=[]observability.Trace} "List of traces"
// @Failure 400 {object} response.APIResponse{error=response.APIError} "Invalid parameters"
// @Failure 401 {object} response.APIResponse{error=response.APIError} "Unauthorized"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/analytics/traces [get]
func (h *Handler) ListTraces(c *gin.Context) {
	// Get project ID from query
	projectID := c.Query("project_id")
	if projectID == "" {
		response.ValidationError(c, "project_id is required", "project_id query parameter is required")
		return
	}

	// Build filter from query parameters
	filter := &observability.TraceFilter{}

	// Session ID filter (virtual session)
	if sessionID := c.Query("session_id"); sessionID != "" {
		filter.SessionID = &sessionID
	}

	// User ID filter
	if userID := c.Query("user_id"); userID != "" {
		filter.UserID = &userID
	}

	// Time range filters
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		startTimeInt, err := strconv.ParseInt(startTimeStr, 10, 64)
		if err != nil {
			response.ValidationError(c, "invalid start_time", "start_time must be a Unix timestamp")
			return
		}
		startTime := time.Unix(startTimeInt, 0)
		filter.StartTime = &startTime
	}

	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		endTimeInt, err := strconv.ParseInt(endTimeStr, 10, 64)
		if err != nil {
			response.ValidationError(c, "invalid end_time", "end_time must be a Unix timestamp")
			return
		}
		endTime := time.Unix(endTimeInt, 0)
		filter.EndTime = &endTime
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

	// Get traces from service
	traces, err := h.services.GetTraceService().GetTracesByProjectID(c.Request.Context(), projectID, filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list traces")
		response.Error(c, err)
		return
	}

	response.Success(c, traces)
}

// GetTrace handles GET /api/v1/analytics/traces/:id
// @Summary Get trace by ID
// @Description Retrieve detailed trace information
// @Tags Dashboard - Analytics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Trace ID"
// @Success 200 {object} response.APIResponse{data=observability.Trace} "Trace details"
// @Failure 404 {object} response.APIResponse{error=response.APIError} "Trace not found"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/analytics/traces/{id} [get]
func (h *Handler) GetTrace(c *gin.Context) {
	traceID := c.Param("id")
	if traceID == "" {
		response.ValidationError(c, "invalid trace_id", "trace_id is required")
		return
	}

	trace, err := h.services.GetTraceService().GetTraceByID(c.Request.Context(), traceID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get trace")
		response.Error(c, err)
		return
	}

	response.Success(c, trace)
}

// GetTraceWithObservations handles GET /api/v1/analytics/traces/:id/observations
// @Summary Get trace with observations tree
// @Description Retrieve trace with all observations in hierarchical structure
// @Tags Dashboard - Analytics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Trace ID"
// @Success 200 {object} response.APIResponse{data=observability.Trace} "Trace with observations"
// @Failure 404 {object} response.APIResponse{error=response.APIError} "Trace not found"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/analytics/traces/{id}/observations [get]
func (h *Handler) GetTraceWithObservations(c *gin.Context) {
	traceID := c.Param("id")
	if traceID == "" {
		response.ValidationError(c, "invalid trace_id", "trace_id is required")
		return
	}

	trace, err := h.services.GetTraceService().GetTraceWithObservations(c.Request.Context(), traceID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get trace with observations")
		response.Error(c, err)
		return
	}

	response.Success(c, trace)
}

// GetTraceWithScores handles GET /api/v1/analytics/traces/:id/scores
// @Summary Get trace with quality scores
// @Description Retrieve trace with all associated quality scores
// @Tags Dashboard - Analytics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Trace ID"
// @Success 200 {object} response.APIResponse{data=observability.Trace} "Trace with scores"
// @Failure 404 {object} response.APIResponse{error=response.APIError} "Trace not found"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/analytics/traces/{id}/scores [get]
func (h *Handler) GetTraceWithScores(c *gin.Context) {
	traceID := c.Param("id")
	if traceID == "" {
		response.ValidationError(c, "invalid trace_id", "trace_id is required")
		return
	}

	trace, err := h.services.GetTraceService().GetTraceWithScores(c.Request.Context(), traceID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get trace with scores")
		response.Error(c, err)
		return
	}

	response.Success(c, trace)
}

// ===== Observation Analytics =====

// ListObservations handles GET /api/v1/analytics/observations
// @Summary List observations with filtering
// @Description Retrieve paginated list of observations
// @Tags Dashboard - Analytics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param trace_id query string false "Filter by trace ID"
// @Param type query string false "Filter by observation type"
// @Param model query string false "Filter by model"
// @Param level query string false "Filter by level"
// @Param limit query int false "Limit (default 50, max 1000)"
// @Param offset query int false "Offset (default 0)"
// @Success 200 {object} response.APIResponse{data=[]observability.Observation} "List of observations"
// @Failure 400 {object} response.APIResponse{error=response.APIError} "Invalid parameters"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/analytics/observations [get]
func (h *Handler) ListObservations(c *gin.Context) {
	filter := &observability.ObservationFilter{}

	// Trace ID filter
	if traceID := c.Query("trace_id"); traceID != "" {
		filter.TraceID = &traceID
	}

	// Type filter
	if obsType := c.Query("type"); obsType != "" {
		filter.Type = &obsType
	}

	// Model filter
	if model := c.Query("model"); model != "" {
		filter.Model = &model
	}

	// Level filter
	if level := c.Query("level"); level != "" {
		filter.Level = &level
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

	observations, err := h.services.GetObservationService().GetObservationsByFilter(c.Request.Context(), filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list observations")
		response.Error(c, err)
		return
	}

	response.Success(c, observations)
}

// GetObservation handles GET /api/v1/analytics/observations/:id
// @Summary Get observation by ID
// @Description Retrieve detailed observation information
// @Tags Dashboard - Analytics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Observation ID"
// @Success 200 {object} response.APIResponse{data=observability.Observation} "Observation details"
// @Failure 404 {object} response.APIResponse{error=response.APIError} "Observation not found"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/analytics/observations/{id} [get]
func (h *Handler) GetObservation(c *gin.Context) {
	observationID := c.Param("id")
	if observationID == "" {
		response.ValidationError(c, "invalid observation_id", "observation_id is required")
		return
	}

	observation, err := h.services.GetObservationService().GetObservationByID(c.Request.Context(), observationID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get observation")
		response.Error(c, err)
		return
	}

	response.Success(c, observation)
}

// ===== Score Analytics =====

// ListScores handles GET /api/v1/analytics/scores
// @Summary List quality scores with filtering
// @Description Retrieve paginated list of quality scores
// @Tags Dashboard - Analytics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param trace_id query string false "Filter by trace ID"
// @Param observation_id query string false "Filter by observation ID"
// @Param session_id query string false "Filter by session ID"
// @Param name query string false "Filter by score name"
// @Param source query string false "Filter by source (API, AUTO, HUMAN, EVAL)"
// @Param data_type query string false "Filter by data type (NUMERIC, CATEGORICAL, BOOLEAN)"
// @Param limit query int false "Limit (default 50, max 1000)"
// @Param offset query int false "Offset (default 0)"
// @Success 200 {object} response.APIResponse{data=[]observability.Score} "List of scores"
// @Failure 400 {object} response.APIResponse{error=response.APIError} "Invalid parameters"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/analytics/scores [get]
func (h *Handler) ListScores(c *gin.Context) {
	filter := &observability.ScoreFilter{}

	// Trace ID filter
	if traceID := c.Query("trace_id"); traceID != "" {
		filter.TraceID = &traceID
	}

	// Observation ID filter
	if observationID := c.Query("observation_id"); observationID != "" {
		filter.ObservationID = &observationID
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

// GetScore handles GET /api/v1/analytics/scores/:id
// @Summary Get quality score by ID
// @Description Retrieve detailed score information
// @Tags Dashboard - Analytics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Score ID"
// @Success 200 {object} response.APIResponse{data=observability.Score} "Score details"
// @Failure 404 {object} response.APIResponse{error=response.APIError} "Score not found"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/analytics/scores/{id} [get]
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

// ===== Update Endpoints (Mutable Operations via REST API) =====

// UpdateTrace handles PUT /api/v1/analytics/traces/:id
// @Summary Update trace by ID
// @Description Update an existing trace (for corrections/enrichment after initial creation)
// @Tags Dashboard - Analytics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Trace ID"
// @Param trace body observability.Trace true "Updated trace data"
// @Success 200 {object} response.APIResponse{data=observability.Trace} "Updated trace"
// @Failure 400 {object} response.APIResponse{error=response.APIError} "Invalid request"
// @Failure 404 {object} response.APIResponse{error=response.APIError} "Trace not found"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/analytics/traces/{id} [put]
func (h *Handler) UpdateTrace(c *gin.Context) {
	traceID := c.Param("id")
	if traceID == "" {
		response.ValidationError(c, "invalid trace_id", "trace_id is required")
		return
	}

	var trace observability.Trace
	if err := c.ShouldBindJSON(&trace); err != nil {
		response.ValidationError(c, "invalid request body", err.Error())
		return
	}

	// Ensure ID matches path parameter
	trace.ID = traceID

	// Update via service
	if err := h.services.GetTraceService().UpdateTrace(c.Request.Context(), &trace); err != nil {
		h.logger.WithError(err).Error("Failed to update trace")
		response.Error(c, err)
		return
	}

	// Fetch updated trace
	updated, err := h.services.GetTraceService().GetTraceByID(c.Request.Context(), traceID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to fetch updated trace")
		response.Error(c, err)
		return
	}

	response.Success(c, updated)
}

// UpdateObservation handles PUT /api/v1/analytics/observations/:id
// @Summary Update observation by ID
// @Description Update an existing observation (for corrections/enrichment after initial creation)
// @Tags Dashboard - Analytics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Observation ID"
// @Param observation body observability.Observation true "Updated observation data"
// @Success 200 {object} response.APIResponse{data=observability.Observation} "Updated observation"
// @Failure 400 {object} response.APIResponse{error=response.APIError} "Invalid request"
// @Failure 404 {object} response.APIResponse{error=response.APIError} "Observation not found"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/analytics/observations/{id} [put]
func (h *Handler) UpdateObservation(c *gin.Context) {
	observationID := c.Param("id")
	if observationID == "" {
		response.ValidationError(c, "invalid observation_id", "observation_id is required")
		return
	}

	var observation observability.Observation
	if err := c.ShouldBindJSON(&observation); err != nil {
		response.ValidationError(c, "invalid request body", err.Error())
		return
	}

	// Ensure ID matches path parameter
	observation.ID = observationID

	// Update via service
	if err := h.services.GetObservationService().UpdateObservation(c.Request.Context(), &observation); err != nil {
		h.logger.WithError(err).Error("Failed to update observation")
		response.Error(c, err)
		return
	}

	// Fetch updated observation
	updated, err := h.services.GetObservationService().GetObservationByID(c.Request.Context(), observationID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to fetch updated observation")
		response.Error(c, err)
		return
	}

	response.Success(c, updated)
}

// UpdateScore handles PUT /api/v1/analytics/scores/:id
// @Summary Update quality score by ID
// @Description Update an existing score (for corrections/enrichment after initial creation)
// @Tags Dashboard - Analytics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Score ID"
// @Param score body observability.Score true "Updated score data"
// @Success 200 {object} response.APIResponse{data=observability.Score} "Updated score"
// @Failure 400 {object} response.APIResponse{error=response.APIError} "Invalid request"
// @Failure 404 {object} response.APIResponse{error=response.APIError} "Score not found"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/analytics/scores/{id} [put]
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

