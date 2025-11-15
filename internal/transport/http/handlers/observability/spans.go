package observability

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"brokle/internal/core/domain/observability"
	"brokle/pkg/response"
)

// Span Handlers for Dashboard (JWT-authenticated, read + update operations)

// ListSpans handles GET /api/v1/spans
// @Summary List spans with filtering
// @Description Retrieve paginated list of spans
// @Tags Spans
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param trace_id query string false "Filter by trace ID"
// @Param type query string false "Filter by span type"
// @Param model query string false "Filter by model"
// @Param level query string false "Filter by level"
// @Param limit query int false "Limit (default 50, max 1000)"
// @Param offset query int false "Offset (default 0)"
// @Success 200 {object} response.APIResponse{data=[]observability.Span} "List of spans"
// @Failure 400 {object} response.APIResponse{error=response.APIError} "Invalid parameters"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/spans [get]
func (h *Handler) ListSpans(c *gin.Context) {
	filter := &observability.SpanFilter{}

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

	spans, err := h.services.GetSpanService().GetSpansByFilter(c.Request.Context(), filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list spans")
		response.Error(c, err)
		return
	}

	response.Success(c, spans)
}

// GetSpan handles GET /api/v1/spans/:id
// @Summary Get span by ID
// @Description Retrieve detailed span information
// @Tags Spans
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Span ID"
// @Success 200 {object} response.APIResponse{data=observability.Span} "Span details"
// @Failure 404 {object} response.APIResponse{error=response.APIError} "Span not found"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/spans/{id} [get]
func (h *Handler) GetSpan(c *gin.Context) {
	spanID := c.Param("id")
	if spanID == "" {
		response.ValidationError(c, "invalid span_id", "span_id is required")
		return
	}

	span, err := h.services.GetSpanService().GetSpanByID(c.Request.Context(), spanID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get span")
		response.Error(c, err)
		return
	}

	response.Success(c, span)
}

// UpdateSpan handles PUT /api/v1/spans/:id
// @Summary Update span by ID
// @Description Update an existing span (for corrections/enrichment after initial creation)
// @Tags Spans
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Span ID"
// @Param span body observability.Span true "Updated span data"
// @Success 200 {object} response.APIResponse{data=observability.Span} "Updated span"
// @Failure 400 {object} response.APIResponse{error=response.APIError} "Invalid request"
// @Failure 404 {object} response.APIResponse{error=response.APIError} "Span not found"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/spans/{id} [put]
func (h *Handler) UpdateSpan(c *gin.Context) {
	spanID := c.Param("id")
	if spanID == "" {
		response.ValidationError(c, "invalid span_id", "span_id is required")
		return
	}

	var span observability.Span
	if err := c.ShouldBindJSON(&span); err != nil {
		response.ValidationError(c, "invalid request body", err.Error())
		return
	}

	// Ensure SpanID matches path parameter
	span.SpanID = spanID

	// Update via service
	if err := h.services.GetSpanService().UpdateSpan(c.Request.Context(), &span); err != nil {
		h.logger.WithError(err).Error("Failed to update span")
		response.Error(c, err)
		return
	}

	// Fetch updated span
	updated, err := h.services.GetSpanService().GetSpanByID(c.Request.Context(), spanID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to fetch updated span")
		response.Error(c, err)
		return
	}

	response.Success(c, updated)
}
