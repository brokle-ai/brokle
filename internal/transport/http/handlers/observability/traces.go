package observability

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"brokle/internal/core/domain/observability"
	"brokle/pkg/response"
)

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
// @Success 200 {object} response.APIResponse{data=[]observability.TraceSummary} "List of trace summaries"
// @Failure 400 {object} response.APIResponse{error=response.APIError} "Invalid parameters"
// @Failure 401 {object} response.APIResponse{error=response.APIError} "Unauthorized"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/traces [get]
func (h *Handler) ListTraces(c *gin.Context) {
	projectID := c.Query("project_id")
	if projectID == "" {
		response.ValidationError(c, "project_id is required", "project_id query parameter is required")
		return
	}

	filter := &observability.TraceFilter{
		ProjectID: projectID,
	}

	if sessionID := c.Query("session_id"); sessionID != "" {
		filter.SessionID = &sessionID
	}
	if userID := c.Query("user_id"); userID != "" {
		filter.UserID = &userID
	}
	if serviceName := c.Query("service_name"); serviceName != "" {
		filter.ServiceName = &serviceName
	}
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

	params := response.ParsePaginationParams(
		c.Query("page"),
		c.Query("limit"),
		c.Query("sort_by"),
		c.Query("sort_dir"),
	)
	filter.Params = params

	traces, err := h.services.GetTraceService().ListTraces(c.Request.Context(), filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list traces")
		response.Error(c, err)
		return
	}

	totalCount, err := h.services.GetTraceService().CountTraces(c.Request.Context(), filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to count traces")
		response.Error(c, err)
		return
	}

	paginationMeta := response.NewPagination(params.Page, params.Limit, totalCount)

	response.SuccessWithPagination(c, traces, paginationMeta)
}

// GetTrace handles GET /api/v1/traces/:id
// @Summary Get trace by ID
// @Description Retrieve detailed trace information (aggregated from spans)
// @Tags Traces
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Trace ID"
// @Success 200 {object} response.APIResponse{data=observability.TraceSummary} "Trace summary"
// @Failure 404 {object} response.APIResponse{error=response.APIError} "Trace not found"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/traces/{id} [get]
func (h *Handler) GetTrace(c *gin.Context) {
	traceID := c.Param("id")
	if traceID == "" {
		response.ValidationError(c, "invalid trace_id", "trace_id is required")
		return
	}

	traceSummary, err := h.services.GetTraceService().GetTrace(c.Request.Context(), traceID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get trace summary")
		response.Error(c, err)
		return
	}

	response.Success(c, traceSummary)
}

// GetTraceWithSpans handles GET /api/v1/traces/:id/spans
// @Summary Get trace with spans tree
// @Description Retrieve trace with all spans in hierarchical structure
// @Tags Traces
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Trace ID"
// @Success 200 {object} response.APIResponse{data=[]observability.Span} "Spans for trace"
// @Failure 404 {object} response.APIResponse{error=response.APIError} "Trace not found"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/traces/{id}/spans [get]
func (h *Handler) GetTraceWithSpans(c *gin.Context) {
	traceID := c.Param("id")
	if traceID == "" {
		response.ValidationError(c, "invalid trace_id", "trace_id is required")
		return
	}

	spans, err := h.services.GetTraceService().GetTraceSpans(c.Request.Context(), traceID)
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
// @Success 200 {object} response.APIResponse{data=[]observability.Score} "Scores for trace"
// @Failure 404 {object} response.APIResponse{error=response.APIError} "Trace not found"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/traces/{id}/scores [get]
func (h *Handler) GetTraceWithScores(c *gin.Context) {
	traceID := c.Param("id")
	if traceID == "" {
		response.ValidationError(c, "invalid trace_id", "trace_id is required")
		return
	}

	scores, err := h.services.GetScoreService().GetScoresByTraceID(c.Request.Context(), traceID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get scores for trace")
		response.Error(c, err)
		return
	}

	response.Success(c, scores)
}

// DeleteTrace handles DELETE /api/v1/traces/:id
// @Summary Delete a trace
// @Description Delete all spans belonging to a trace
// @Tags Traces
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Trace ID"
// @Success 200 {object} response.APIResponse "Trace deleted"
// @Failure 404 {object} response.APIResponse{error=response.APIError} "Trace not found"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/traces/{id} [delete]
func (h *Handler) DeleteTrace(c *gin.Context) {
	traceID := c.Param("id")
	if traceID == "" {
		response.ValidationError(c, "invalid trace_id", "trace_id is required")
		return
	}

	if err := h.services.GetTraceService().DeleteTrace(c.Request.Context(), traceID); err != nil {
		h.logger.WithError(err).Error("Failed to delete trace")
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "trace deleted successfully"})
}
