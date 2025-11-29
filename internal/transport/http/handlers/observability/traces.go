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
// @Param service_name query string false "Filter by service name (OTLP resource attribute)"
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
	filter := &observability.TraceFilter{
		ProjectID: projectID, // Set project scope
	}

	// Session ID filter (virtual session)
	if sessionID := c.Query("session_id"); sessionID != "" {
		filter.SessionID = &sessionID
	}

	// User ID filter
	if userID := c.Query("user_id"); userID != "" {
		filter.UserID = &userID
	}

	// Service name filter (OTLP resource attribute)
	if serviceName := c.Query("service_name"); serviceName != "" {
		filter.ServiceName = &serviceName
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

	// Offset pagination
	params := response.ParsePaginationParams(
		c.Query("page"),
		c.Query("limit"),
		c.Query("sort_by"),
		c.Query("sort_dir"),
	)

	// Set embedded pagination fields
	filter.Params = params

	// Get traces from service (OTEL-native: returns TraceMetrics)
	traces, err := h.services.GetTraceService().ListTraces(c.Request.Context(), filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list traces")
		response.Error(c, err)
		return
	}

	// Get total count for pagination metadata
	totalCount, err := h.services.GetTraceService().CountTraces(c.Request.Context(), filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to count traces")
		response.Error(c, err)
		return
	}

	// Build pagination metadata (NewPagination calculates has_next, has_prev, total_pages)
	paginationMeta := response.NewPagination(params.Page, params.Limit, totalCount)

	response.SuccessWithPagination(c, traces, paginationMeta)
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

	// OTEL-Native: Get trace metrics (aggregated from spans)
	traceMetrics, err := h.services.GetTraceService().GetTraceMetrics(c.Request.Context(), traceID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get trace metrics")
		response.Error(c, err)
		return
	}

	response.Success(c, traceMetrics)
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

	// OTEL-Native: Get all spans for trace
	spans, err := h.services.GetTraceService().GetTraceWithAllSpans(c.Request.Context(), traceID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get spans for trace")
		response.Error(c, err)
		return
	}

	response.Success(c, spans)
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

	// OTEL-Native: Get scores directly via ScoreService
	scores, err := h.services.GetScoreService().GetScoresByTraceID(c.Request.Context(), traceID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get scores for trace")
		response.Error(c, err)
		return
	}

	response.Success(c, scores)
}

// UpdateTrace handles PUT /api/v1/traces/:id
// @Summary Update trace metadata
// @Description Update root span attributes (OTEL-native: traces are root spans)
// @Tags Traces
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Trace ID"
// @Param span body observability.Span true "Updated root span data"
// @Success 200 {object} response.APIResponse{data=observability.Span} "Updated root span"
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

	// OTEL-Native: Get existing root span first to get span_id
	existingRootSpan, err := h.services.GetTraceService().GetRootSpan(c.Request.Context(), traceID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get root span for update")
		response.Error(c, err)
		return
	}

	// Bind update data from request body
	var rootSpan observability.Span
	if err := c.ShouldBindJSON(&rootSpan); err != nil {
		response.ValidationError(c, "invalid request body", err.Error())
		return
	}

	// CRITICAL: Set IDs from existing root span (span_id is required for UpdateSpan to work)
	rootSpan.SpanID = existingRootSpan.SpanID
	rootSpan.TraceID = existingRootSpan.TraceID
	rootSpan.ParentSpanID = existingRootSpan.ParentSpanID
	rootSpan.ProjectID = existingRootSpan.ProjectID

	// Update root span via SpanService
	if err := h.services.GetSpanService().UpdateSpan(c.Request.Context(), &rootSpan); err != nil {
		h.logger.WithError(err).Error("Failed to update root span")
		response.Error(c, err)
		return
	}

	// Fetch updated root span
	updated, err := h.services.GetTraceService().GetRootSpan(c.Request.Context(), traceID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to fetch updated root span")
		response.Error(c, err)
		return
	}

	response.Success(c, updated)
}
