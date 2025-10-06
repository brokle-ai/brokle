package observability

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"brokle/internal/core/domain/observability"
	"brokle/pkg/response"
	"brokle/pkg/ulid"
)

// GetObservation handles GET /api/v1/observability/observations/{id}
func (h *Handler) GetObservation(c *gin.Context) {
	idStr := c.Param("id")

	observationID, err := ulid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid observation ID", err.Error())
		return
	}

	// Get observation via service
	observation, err := h.services.GetObservationService().GetObservation(c.Request.Context(), observationID)
	if err != nil {
		if observability.IsNotFoundError(err) {
			response.NotFound(c, "Observation")
			return
		}
		response.InternalServerError(c, "Failed to get observation")
		return
	}

	// Convert to response
	resp := h.observationToResponse(observation)
	response.Success(c, resp)
}

// ListObservations handles GET /api/v1/observability/observations
func (h *Handler) ListObservations(c *gin.Context) {
	// Parse query parameters
	filter, err := h.parseObservationFilter(c)
	if err != nil {
		response.BadRequest(c, "Invalid filter parameters", err.Error())
		return
	}

	// Get observations via service
	observations, total, err := h.services.GetObservationService().ListObservations(c.Request.Context(), filter)
	if err != nil {
		response.InternalServerError(c, "Failed to list observations")
		return
	}

	// Convert to response
	var observationResponses []ObservationResponse
	for _, obs := range observations {
		observationResponses = append(observationResponses, h.observationToResponse(obs))
	}

	resp := ListObservationsResponse{
		Observations: observationResponses,
		Total:        total,
		Limit:        filter.Limit,
		Offset:       filter.Offset,
	}

	// Create pagination metadata
	pagination := response.NewPagination(filter.Offset/filter.Limit+1, filter.Limit, int64(total))
	response.SuccessWithPagination(c, resp, pagination)
}

// Helper methods

// parseObservationFilter parses query parameters into an ObservationFilter
func (h *Handler) parseObservationFilter(c *gin.Context) (*observability.ObservationFilter, error) {
	filter := &observability.ObservationFilter{}

	// Parse trace_id
	if traceIDStr := c.Query("trace_id"); traceIDStr != "" {
		traceID, err := ulid.Parse(traceIDStr)
		if err != nil {
			return nil, err
		}
		filter.TraceID = &traceID
	}

	// Parse type
	if obsType := c.Query("type"); obsType != "" {
		obsTypeEnum := observability.ObservationType(obsType)
		filter.Type = &obsTypeEnum
	}

	// Parse provider
	if provider := c.Query("provider"); provider != "" {
		filter.Provider = &provider
	}

	// Parse model
	if model := c.Query("model"); model != "" {
		filter.Model = &model
	}

	// Parse level
	if level := c.Query("level"); level != "" {
		levelEnum := observability.ObservationLevel(level)
		filter.Level = &levelEnum
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
		filter.SortBy = "start_time" // Default sort
	}

	if sortOrder := c.Query("sort_order"); sortOrder != "" {
		filter.SortOrder = sortOrder
	} else {
		filter.SortOrder = "desc" // Default order
	}

	return filter, nil
}

// observationToResponse converts an Observation domain entity to an ObservationResponse
func (h *Handler) observationToResponse(obs *observability.Observation) ObservationResponse {
	resp := ObservationResponse{
		ID:                    obs.ID.String(),
		TraceID:               obs.TraceID.String(),
		ExternalObservationID: obs.ExternalObservationID,
		Type:                  string(obs.Type),
		Name:                  obs.Name,
		StartTime:             obs.StartTime,
		Level:                 string(obs.Level),
		PromptTokens:          obs.PromptTokens,
		CompletionTokens:      obs.CompletionTokens,
		TotalTokens:           obs.TotalTokens,
		Input:                 obs.Input,
		Output:                obs.Output,
		ModelParameters:       obs.ModelParameters,
		CreatedAt:             obs.CreatedAt,
		UpdatedAt:             obs.UpdatedAt,
	}

	if obs.ParentObservationID != nil {
		resp.ParentObservationID = obs.ParentObservationID.String()
	}

	if obs.EndTime != nil {
		resp.EndTime = obs.EndTime
	}

	if obs.StatusMessage != nil {
		resp.StatusMessage = *obs.StatusMessage
	}

	if obs.Version != nil {
		resp.Version = *obs.Version
	}

	if obs.Model != nil {
		resp.Model = *obs.Model
	}

	if obs.Provider != nil {
		resp.Provider = *obs.Provider
	}

	if obs.InputCost != nil {
		resp.InputCost = obs.InputCost
	}

	if obs.OutputCost != nil {
		resp.OutputCost = obs.OutputCost
	}

	if obs.TotalCost != nil {
		resp.TotalCost = obs.TotalCost
	}

	if obs.LatencyMs != nil {
		resp.LatencyMs = obs.LatencyMs
	}

	if obs.QualityScore != nil {
		resp.QualityScore = obs.QualityScore
	}

	return resp
}
