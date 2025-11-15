package observability

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"brokle/internal/core/domain/observability"
	"brokle/pkg/response"
)

// Trace Handlers for Dashboard (JWT-authenticated, read + update operations)

// ListTraces handles GET /api/v1/traces
// @Summary List traces for a project
// @Description Retrieve paginated list of traces with optional filtering
// @Tags Traces
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
// @Router /api/v1/traces [get]
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

// GetTrace handles GET /api/v1/traces/:id
// @Summary Get trace by ID
// @Description Retrieve detailed trace information
// @Tags Traces
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Trace ID"
// @Success 200 {object} response.APIResponse{data=observability.Trace} "Trace details"
// @Failure 404 {object} response.APIResponse{error=response.APIError} "Trace not found"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/traces/{id} [get]
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

// GetTraceWithSpans handles GET /api/v1/traces/:id/spans
// @Summary Get trace with spans tree
// @Description Retrieve trace with all spans in hierarchical structure
// @Tags Traces
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Trace ID"
// @Success 200 {object} response.APIResponse{data=observability.Trace} "Trace with spans"
// @Failure 404 {object} response.APIResponse{error=response.APIError} "Trace not found"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/traces/{id}/spans [get]
func (h *Handler) GetTraceWithSpans(c *gin.Context) {
	traceID := c.Param("id")
	if traceID == "" {
		response.ValidationError(c, "invalid trace_id", "trace_id is required")
		return
	}

	trace, err := h.services.GetTraceService().GetTraceWithSpans(c.Request.Context(), traceID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get trace with spans")
		response.Error(c, err)
		return
	}

	response.Success(c, trace)
}

// GetTraceWithScores handles GET /api/v1/traces/:id/scores
// @Summary Get trace with quality scores
// @Description Retrieve trace with all associated quality scores
// @Tags Traces
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Trace ID"
// @Success 200 {object} response.APIResponse{data=observability.Trace} "Trace with scores"
// @Failure 404 {object} response.APIResponse{error=response.APIError} "Trace not found"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/traces/{id}/scores [get]
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

// UpdateTrace handles PUT /api/v1/traces/:id
// @Summary Update trace metadata
// @Description Update trace name, tags, or metadata (corrections/enrichment via dashboard)
// @Tags Traces
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Trace ID"
// @Param trace body observability.Trace true "Updated trace data"
// @Success 200 {object} response.APIResponse{data=observability.Trace} "Updated trace"
// @Failure 400 {object} response.APIResponse{error=response.APIError} "Invalid request"
// @Failure 404 {object} response.APIResponse{error=response.APIError} "Trace not found"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/traces/{id} [put]
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

	// Ensure TraceID matches path parameter
	trace.TraceID = traceID

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
