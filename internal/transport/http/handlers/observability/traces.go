package observability

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"brokle/internal/core/domain/observability"
	"brokle/pkg/response"
	"brokle/pkg/ulid"
)

// GetTrace handles GET /api/v1/observability/traces/{id}
// @Summary Get trace by ID
// @Description Retrieve a specific LLM observability trace by its unique identifier
// @Tags Observability - Traces
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Trace ID (ULID format)"
// @Success 200 {object} response.SuccessResponse{data=TraceResponse} "Trace retrieved successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid trace ID format"
// @Failure 404 {object} response.ErrorResponse "Trace not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/observability/traces/{id} [get]
func (h *Handler) GetTrace(c *gin.Context) {
	idStr := c.Param("id")

	traceID, err := ulid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid trace ID", err.Error())
		return
	}

	// Get trace via service
	trace, err := h.services.GetTraceService().GetTrace(c.Request.Context(), traceID)
	if err != nil {
		if observability.IsNotFoundError(err) {
			response.NotFound(c, "Trace")
			return
		}
		response.InternalServerError(c, "Failed to get trace")
		return
	}

	// Convert to response
	resp := h.traceToResponse(trace)
	response.Success(c, resp)
}

// GetTraceWithObservations handles GET /api/v1/observability/traces/{id}/observations
// @Summary Get trace with observations
// @Description Retrieve a trace along with all its associated observations (LLM calls, spans, events)
// @Tags Observability - Traces
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Trace ID (ULID format)"
// @Success 200 {object} response.SuccessResponse{data=TraceWithObservationsResponse} "Trace with observations retrieved successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid trace ID format"
// @Failure 404 {object} response.ErrorResponse "Trace not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/observability/traces/{id}/observations [get]
func (h *Handler) GetTraceWithObservations(c *gin.Context) {
	idStr := c.Param("id")

	traceID, err := ulid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid trace ID", err.Error())
		return
	}

	// Get trace with observations via service
	trace, err := h.services.GetTraceService().GetTraceWithObservations(c.Request.Context(), traceID)
	if err != nil {
		if observability.IsNotFoundError(err) {
			response.NotFound(c, "Trace")
			return
		}
		response.InternalServerError(c, "Failed to get trace with observations")
		return
	}

	// Convert to response
	resp := h.traceWithObservationsToResponse(trace)
	response.Success(c, resp)
}

// ListTraces handles GET /api/v1/observability/traces
// @Summary List traces
// @Description Get a paginated list of LLM observability traces with filtering options
// @Tags Observability - Traces
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param project_id query string false "Project ID filter"
// @Param user_id query string false "User ID filter"
// @Param session_id query string false "Session ID filter"
// @Param name query string false "Trace name filter"
// @Param external_trace_id query string false "External trace ID filter"
// @Param start_time query string false "Start time filter (RFC3339 format)"
// @Param end_time query string false "End time filter (RFC3339 format)"
// @Param limit query int false "Maximum number of results" default(50)
// @Param offset query int false "Number of results to skip" default(0)
// @Param sort_by query string false "Sort field" default(created_at)
// @Param sort_order query string false "Sort order (asc/desc)" default(desc)
// @Success 200 {object} response.APIResponse{data=ListTracesResponse,meta=response.Meta{pagination=response.Pagination}} "Traces retrieved successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid filter parameters"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/observability/traces [get]
func (h *Handler) ListTraces(c *gin.Context) {
	// Parse query parameters
	filter, err := h.parseTraceFilter(c)
	if err != nil {
		response.BadRequest(c, "Invalid filter parameters", err.Error())
		return
	}

	// Get traces via service
	traces, total, err := h.services.GetTraceService().ListTraces(c.Request.Context(), filter)
	if err != nil {
		response.InternalServerError(c, "Failed to list traces")
		return
	}

	// Convert to response
	var traceResponses []TraceResponse
	for _, trace := range traces {
		traceResponses = append(traceResponses, h.traceToResponse(trace))
	}

	resp := ListTracesResponse{
		Traces: traceResponses,
		Total:  total,
		Limit:  filter.Limit,
		Offset: filter.Offset,
	}

	// Create pagination metadata
	pagination := response.NewPagination(filter.Offset/filter.Limit+1, filter.Limit, int64(total))
	response.SuccessWithPagination(c, resp, pagination)
}

// GetTraceStats handles GET /api/v1/observability/traces/{id}/stats
// @Summary Get trace statistics
// @Description Retrieve aggregated statistics for a specific LLM trace (cost, tokens, latency, etc.)
// @Tags Observability - Traces
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Trace ID (ULID format)"
// @Success 200 {object} response.SuccessResponse{data=TraceStatsResponse} "Trace statistics retrieved successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid trace ID format"
// @Failure 404 {object} response.ErrorResponse "Trace not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/observability/traces/{id}/stats [get]
func (h *Handler) GetTraceStats(c *gin.Context) {
	idStr := c.Param("id")

	traceID, err := ulid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid trace ID", err.Error())
		return
	}

	// Get trace stats via service
	stats, err := h.services.GetTraceService().GetTraceStats(c.Request.Context(), traceID)
	if err != nil {
		if observability.IsNotFoundError(err) {
			response.NotFound(c, "Trace")
			return
		}
		response.InternalServerError(c, "Failed to get trace stats")
		return
	}

	// Convert to response
	resp := h.traceStatsToResponse(stats)
	response.Success(c, resp)
}

// Helper methods

// parseTraceFilter parses query parameters into a TraceFilter
func (h *Handler) parseTraceFilter(c *gin.Context) (*observability.TraceFilter, error) {
	filter := &observability.TraceFilter{}

	// Parse project_id
	if projectIDStr := c.Query("project_id"); projectIDStr != "" {
		projectID, err := ulid.Parse(projectIDStr)
		if err != nil {
			return nil, err
		}
		filter.ProjectID = &projectID
	}

	// Parse user_id
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		userID, err := ulid.Parse(userIDStr)
		if err != nil {
			return nil, err
		}
		filter.UserID = &userID
	}

	// Parse session_id
	if sessionIDStr := c.Query("session_id"); sessionIDStr != "" {
		sessionID, err := ulid.Parse(sessionIDStr)
		if err != nil {
			return nil, err
		}
		filter.SessionID = &sessionID
	}

	// Parse name
	if name := c.Query("name"); name != "" {
		filter.Name = &name
	}

	// Parse external_trace_id
	if externalTraceID := c.Query("external_trace_id"); externalTraceID != "" {
		filter.ExternalTraceID = &externalTraceID
	}

	// Parse time range
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		startTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			return nil, err
		}
		filter.StartTime = &startTime
	}

	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		endTime, err := time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			return nil, err
		}
		filter.EndTime = &endTime
	}

	// Parse pagination
	if limitStr := c.Query("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return nil, err
		}
		filter.Limit = limit
	} else {
		filter.Limit = 50 // Default limit
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			return nil, err
		}
		filter.Offset = offset
	}

	// Parse sorting
	if sortBy := c.Query("sort_by"); sortBy != "" {
		filter.SortBy = sortBy
	} else {
		filter.SortBy = "created_at" // Default sort
	}

	if sortOrder := c.Query("sort_order"); sortOrder != "" {
		filter.SortOrder = sortOrder
	} else {
		filter.SortOrder = "desc" // Default order
	}

	return filter, nil
}

// traceToResponse converts a Trace domain entity to a TraceResponse
func (h *Handler) traceToResponse(trace *observability.Trace) TraceResponse {
	resp := TraceResponse{
		ID:              trace.ID.String(),
		ProjectID:       trace.ProjectID.String(),
		ExternalTraceID: trace.ExternalTraceID,
		Name:            trace.Name,
		Tags:            trace.Tags,
		Metadata:        trace.Metadata,
		CreatedAt:       trace.CreatedAt,
		UpdatedAt:       trace.UpdatedAt,
	}

	if trace.UserID != nil {
		resp.UserID = trace.UserID.String()
	}

	if trace.SessionID != nil {
		resp.SessionID = trace.SessionID.String()
	}

	if trace.ParentTraceID != nil {
		resp.ParentTraceID = trace.ParentTraceID.String()
	}

	return resp
}

// traceWithObservationsToResponse converts a Trace with observations to response
func (h *Handler) traceWithObservationsToResponse(trace *observability.Trace) TraceWithObservationsResponse {
	resp := TraceWithObservationsResponse{
		TraceResponse: h.traceToResponse(trace),
		Observations:  make([]ObservationResponse, 0, len(trace.Observations)),
	}

	// Convert observations
	for _, obs := range trace.Observations {
		obsResp := h.observationToResponse(&obs)
		resp.Observations = append(resp.Observations, obsResp)
	}

	return resp
}

// traceStatsToResponse converts TraceStats to response
func (h *Handler) traceStatsToResponse(stats *observability.TraceStats) TraceStatsResponse {
	return TraceStatsResponse{
		TraceID:     stats.TraceID.String(),
		TotalCost:   stats.TotalCost,
		TotalTokens: stats.TotalTokens,
	}
}
