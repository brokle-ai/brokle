package observability

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"brokle/internal/core/domain/observability"
	"brokle/pkg/response"
	"brokle/pkg/ulid"
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
	projectIDStr := c.Query("project_id")
	if projectIDStr == "" {
		response.ValidationError(c, "project_id is required", "project_id query parameter is required")
		return
	}

	projectID, err := ulid.Parse(projectIDStr)
	if err != nil {
		response.ValidationError(c, "invalid project_id", "project_id must be a valid ULID")
		return
	}

	// Build filter from query parameters
	filter := &observability.TraceFilter{}

	// Session ID filter
	if sessionIDStr := c.Query("session_id"); sessionIDStr != "" {
		sessionID, err := ulid.Parse(sessionIDStr)
		if err != nil {
			response.ValidationError(c, "invalid session_id", "session_id must be a valid ULID")
			return
		}
		filter.SessionID = &sessionID
	}

	// User ID filter
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		userID, err := ulid.Parse(userIDStr)
		if err != nil {
			response.ValidationError(c, "invalid user_id", "user_id must be a valid ULID")
			return
		}
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
	traceIDStr := c.Param("id")
	traceID, err := ulid.Parse(traceIDStr)
	if err != nil {
		response.ValidationError(c, "invalid trace_id", "trace_id must be a valid ULID")
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
	traceIDStr := c.Param("id")
	traceID, err := ulid.Parse(traceIDStr)
	if err != nil {
		response.ValidationError(c, "invalid trace_id", "trace_id must be a valid ULID")
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
	traceIDStr := c.Param("id")
	traceID, err := ulid.Parse(traceIDStr)
	if err != nil {
		response.ValidationError(c, "invalid trace_id", "trace_id must be a valid ULID")
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
	if traceIDStr := c.Query("trace_id"); traceIDStr != "" {
		traceID, err := ulid.Parse(traceIDStr)
		if err != nil {
			response.ValidationError(c, "invalid trace_id", "trace_id must be a valid ULID")
			return
		}
		filter.TraceID = &traceID
	}

	// Type filter
	if typeStr := c.Query("type"); typeStr != "" {
		obsType := observability.ObservationType(typeStr)
		filter.Type = &obsType
	}

	// Model filter
	if model := c.Query("model"); model != "" {
		filter.Model = &model
	}

	// Level filter
	if levelStr := c.Query("level"); levelStr != "" {
		level := observability.ObservationLevel(levelStr)
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
	observationIDStr := c.Param("id")
	observationID, err := ulid.Parse(observationIDStr)
	if err != nil {
		response.ValidationError(c, "invalid observation_id", "observation_id must be a valid ULID")
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
	if traceIDStr := c.Query("trace_id"); traceIDStr != "" {
		traceID, err := ulid.Parse(traceIDStr)
		if err != nil {
			response.ValidationError(c, "invalid trace_id", "trace_id must be a valid ULID")
			return
		}
		filter.TraceID = &traceID
	}

	// Observation ID filter
	if observationIDStr := c.Query("observation_id"); observationIDStr != "" {
		observationID, err := ulid.Parse(observationIDStr)
		if err != nil {
			response.ValidationError(c, "invalid observation_id", "observation_id must be a valid ULID")
			return
		}
		filter.ObservationID = &observationID
	}

	// Session ID filter
	if sessionIDStr := c.Query("session_id"); sessionIDStr != "" {
		sessionID, err := ulid.Parse(sessionIDStr)
		if err != nil {
			response.ValidationError(c, "invalid session_id", "session_id must be a valid ULID")
			return
		}
		filter.SessionID = &sessionID
	}

	// Name filter
	if name := c.Query("name"); name != "" {
		filter.Name = &name
	}

	// Source filter
	if sourceStr := c.Query("source"); sourceStr != "" {
		source := observability.ScoreSource(sourceStr)
		filter.Source = &source
	}

	// Data type filter
	if dataTypeStr := c.Query("data_type"); dataTypeStr != "" {
		dataType := observability.ScoreDataType(dataTypeStr)
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
	scoreIDStr := c.Param("id")
	scoreID, err := ulid.Parse(scoreIDStr)
	if err != nil {
		response.ValidationError(c, "invalid score_id", "score_id must be a valid ULID")
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

// ===== Session Analytics =====

// ListSessions handles GET /api/v1/analytics/sessions
// @Summary List sessions with filtering
// @Description Retrieve paginated list of sessions
// @Tags Dashboard - Analytics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param project_id query string true "Project ID"
// @Param user_id query string false "Filter by user ID"
// @Param bookmarked query bool false "Filter by bookmarked"
// @Param public query bool false "Filter by public"
// @Param limit query int false "Limit (default 50, max 1000)"
// @Param offset query int false "Offset (default 0)"
// @Success 200 {object} response.APIResponse{data=[]observability.Session} "List of sessions"
// @Failure 400 {object} response.APIResponse{error=response.APIError} "Invalid parameters"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/analytics/sessions [get]
func (h *Handler) ListSessions(c *gin.Context) {
	// Get project ID from query
	projectIDStr := c.Query("project_id")
	if projectIDStr == "" {
		response.ValidationError(c, "project_id is required", "project_id query parameter is required")
		return
	}

	projectID, err := ulid.Parse(projectIDStr)
	if err != nil {
		response.ValidationError(c, "invalid project_id", "project_id must be a valid ULID")
		return
	}

	filter := &observability.SessionFilter{}

	// User ID filter
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		userID, err := ulid.Parse(userIDStr)
		if err != nil {
			response.ValidationError(c, "invalid user_id", "user_id must be a valid ULID")
			return
		}
		filter.UserID = &userID
	}

	// Bookmarked filter
	if bookmarkedStr := c.Query("bookmarked"); bookmarkedStr != "" {
		bookmarked, err := strconv.ParseBool(bookmarkedStr)
		if err != nil {
			response.ValidationError(c, "invalid bookmarked", "bookmarked must be a boolean")
			return
		}
		filter.Bookmarked = &bookmarked
	}

	// Public filter
	if publicStr := c.Query("public"); publicStr != "" {
		public, err := strconv.ParseBool(publicStr)
		if err != nil {
			response.ValidationError(c, "invalid public", "public must be a boolean")
			return
		}
		filter.Public = &public
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

	sessions, err := h.services.GetSessionService().GetSessionsByProjectID(c.Request.Context(), projectID, filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list sessions")
		response.Error(c, err)
		return
	}

	response.Success(c, sessions)
}

// GetSession handles GET /api/v1/analytics/sessions/:id
// @Summary Get session by ID
// @Description Retrieve detailed session information
// @Tags Dashboard - Analytics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Success 200 {object} response.APIResponse{data=observability.Session} "Session details"
// @Failure 404 {object} response.APIResponse{error=response.APIError} "Session not found"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/analytics/sessions/{id} [get]
func (h *Handler) GetSession(c *gin.Context) {
	sessionIDStr := c.Param("id")
	sessionID, err := ulid.Parse(sessionIDStr)
	if err != nil {
		response.ValidationError(c, "invalid session_id", "session_id must be a valid ULID")
		return
	}

	session, err := h.services.GetSessionService().GetSessionByID(c.Request.Context(), sessionID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get session")
		response.Error(c, err)
		return
	}

	response.Success(c, session)
}

// GetSessionWithTraces handles GET /api/v1/analytics/sessions/:id/traces
// @Summary Get session with traces
// @Description Retrieve session with all associated traces
// @Tags Dashboard - Analytics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Success 200 {object} response.APIResponse{data=observability.Session} "Session with traces"
// @Failure 404 {object} response.APIResponse{error=response.APIError} "Session not found"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/analytics/sessions/{id}/traces [get]
func (h *Handler) GetSessionWithTraces(c *gin.Context) {
	sessionIDStr := c.Param("id")
	sessionID, err := ulid.Parse(sessionIDStr)
	if err != nil {
		response.ValidationError(c, "invalid session_id", "session_id must be a valid ULID")
		return
	}

	session, err := h.services.GetSessionService().GetSessionWithTraces(c.Request.Context(), sessionID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get session with traces")
		response.Error(c, err)
		return
	}

	response.Success(c, session)
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
	traceIDStr := c.Param("id")
	traceID, err := ulid.Parse(traceIDStr)
	if err != nil {
		response.ValidationError(c, "invalid trace_id", "trace_id must be a valid ULID")
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
	observationIDStr := c.Param("id")
	observationID, err := ulid.Parse(observationIDStr)
	if err != nil {
		response.ValidationError(c, "invalid observation_id", "observation_id must be a valid ULID")
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
	scoreIDStr := c.Param("id")
	scoreID, err := ulid.Parse(scoreIDStr)
	if err != nil {
		response.ValidationError(c, "invalid score_id", "score_id must be a valid ULID")
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

// UpdateSession handles PUT /api/v1/analytics/sessions/:id
// @Summary Update session by ID
// @Description Update an existing session (for corrections/enrichment after initial creation). Supports partial updates - only send fields you want to change.
// @Tags Dashboard - Analytics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Param updateReq body observability.UpdateSessionRequest true "Session fields to update (partial)"
// @Success 200 {object} response.APIResponse{data=observability.Session} "Updated session"
// @Failure 400 {object} response.APIResponse{error=response.APIError} "Invalid request"
// @Failure 404 {object} response.APIResponse{error=response.APIError} "Session not found"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/analytics/sessions/{id} [put]
func (h *Handler) UpdateSession(c *gin.Context) {
	sessionIDStr := c.Param("id")
	sessionID, err := ulid.Parse(sessionIDStr)
	if err != nil {
		response.ValidationError(c, "invalid session_id", "session_id must be a valid ULID")
		return
	}

	// Bind to update request DTO (uses pointers to track field presence)
	var updateReq observability.UpdateSessionRequest
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		response.ValidationError(c, "invalid request body", err.Error())
		return
	}

	// Update via service with session ID and update request
	if err := h.services.GetSessionService().UpdateSession(c.Request.Context(), sessionID, &updateReq); err != nil {
		h.logger.WithError(err).Error("Failed to update session")
		response.Error(c, err)
		return
	}

	// Fetch updated session
	updated, err := h.services.GetSessionService().GetSessionByID(c.Request.Context(), sessionID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to fetch updated session")
		response.Error(c, err)
		return
	}

	response.Success(c, updated)
}
