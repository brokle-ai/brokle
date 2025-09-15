package observability

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"brokle/internal/core/domain/observability"
	"brokle/pkg/response"
	"brokle/pkg/ulid"
)

// CreateObservation handles POST /api/v1/observability/observations
// @Summary Create a new observation
// @Description Create a new LLM observation (LLM call, span, event, generation, etc.) within a trace
// @Tags Observability - Observations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateObservationRequest true "Observation creation data"
// @Success 201 {object} response.SuccessResponse{data=ObservationResponse} "Observation created successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload or observation data"
// @Failure 409 {object} response.ErrorResponse "Observation with external_observation_id already exists"
// @Failure 422 {object} response.ErrorResponse "Validation failed"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/observability/observations [post]
func (h *Handler) CreateObservation(c *gin.Context) {
	var req CreateObservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	// Convert request to domain entity
	observation, err := h.requestToObservation(&req)
	if err != nil {
		response.BadRequest(c, "Invalid observation data", err.Error())
		return
	}

	// Create observation via service
	createdObservation, err := h.services.GetObservationService().CreateObservation(c.Request.Context(), observation)
	if err != nil {
		if observability.IsValidationError(err) {
			response.ValidationError(c, "Validation failed", err.Error())
			return
		}
		if observability.IsConflictError(err) {
			response.Conflict(c, "Observation already exists")
			return
		}
		response.InternalServerError(c, "Failed to create observation")
		return
	}

	// Convert to response
	resp := h.observationToResponse(createdObservation)
	response.Created(c, resp)
}

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

// UpdateObservation handles PUT /api/v1/observability/observations/{id}
func (h *Handler) UpdateObservation(c *gin.Context) {
	idStr := c.Param("id")

	observationID, err := ulid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid observation ID", err.Error())
		return
	}

	var req UpdateObservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	// Get existing observation first
	observation, err := h.services.GetObservationService().GetObservation(c.Request.Context(), observationID)
	if err != nil {
		if observability.IsNotFoundError(err) {
			response.NotFound(c, "Observation")
			return
		}
		response.InternalServerError(c, "Failed to get observation")
		return
	}

	// Apply updates from request
	h.applyObservationUpdates(observation, &req)

	// Update observation via service
	updatedObservation, err := h.services.GetObservationService().UpdateObservation(c.Request.Context(), observation)
	if err != nil {
		if observability.IsValidationError(err) {
			response.ValidationError(c, "Validation failed", err.Error())
			return
		}
		response.InternalServerError(c, "Failed to update observation")
		return
	}

	// Convert to response
	resp := h.observationToResponse(updatedObservation)
	response.Success(c, resp)
}

// CompleteObservation handles POST /api/v1/observability/observations/{id}/complete
func (h *Handler) CompleteObservation(c *gin.Context) {
	idStr := c.Param("id")

	observationID, err := ulid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid observation ID", err.Error())
		return
	}

	var req CompleteObservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	// Convert request to completion data
	completionData := &observability.ObservationCompletion{
		EndTime: req.EndTime,
		Output:  req.Output,
	}

	if req.StatusMessage != "" {
		completionData.StatusMessage = &req.StatusMessage
	}

	if req.PromptTokens > 0 || req.CompletionTokens > 0 || req.TotalTokens > 0 {
		completionData.Usage = &observability.TokenUsage{
			PromptTokens:     req.PromptTokens,
			CompletionTokens: req.CompletionTokens,
			TotalTokens:      req.TotalTokens,
		}
	}

	if req.InputCost != nil || req.OutputCost != nil || req.TotalCost != nil {
		completionData.Cost = &observability.CostCalculation{
			Currency: "USD", // Default currency
		}
		if req.InputCost != nil {
			completionData.Cost.InputCost = *req.InputCost
		}
		if req.OutputCost != nil {
			completionData.Cost.OutputCost = *req.OutputCost
		}
		if req.TotalCost != nil {
			completionData.Cost.TotalCost = *req.TotalCost
		}
	}

	if req.QualityScore != nil {
		completionData.QualityScore = req.QualityScore
	}

	// Complete observation via service
	completedObservation, err := h.services.GetObservationService().CompleteObservation(c.Request.Context(), observationID, completionData)
	if err != nil {
		if observability.IsNotFoundError(err) {
			response.NotFound(c, "Observation")
			return
		}
		if observability.IsValidationError(err) {
			response.ValidationError(c, "Validation failed", err.Error())
			return
		}
		response.InternalServerError(c, "Failed to complete observation")
		return
	}

	// Convert to response
	resp := h.observationToResponse(completedObservation)
	response.Success(c, resp)
}

// DeleteObservation handles DELETE /api/v1/observability/observations/{id}
func (h *Handler) DeleteObservation(c *gin.Context) {
	idStr := c.Param("id")

	observationID, err := ulid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid observation ID", err.Error())
		return
	}

	// Delete observation via service
	if err := h.services.GetObservationService().DeleteObservation(c.Request.Context(), observationID); err != nil {
		if observability.IsNotFoundError(err) {
			response.NotFound(c, "Observation")
			return
		}
		response.InternalServerError(c, "Failed to delete observation")
		return
	}

	response.Success(c, gin.H{"message": "Observation deleted successfully"})
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

// GetObservationsByTrace handles GET /api/v1/observability/traces/{trace_id}/observations
func (h *Handler) GetObservationsByTrace(c *gin.Context) {
	traceIDStr := c.Param("trace_id")

	traceID, err := ulid.Parse(traceIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid trace ID", err.Error())
		return
	}

	// Get observations by trace via service
	observations, err := h.services.GetObservationService().GetObservationsByTrace(c.Request.Context(), traceID)
	if err != nil {
		response.InternalServerError(c, "Failed to get observations by trace")
		return
	}

	// Convert to response
	var observationResponses []ObservationResponse
	for _, obs := range observations {
		observationResponses = append(observationResponses, h.observationToResponse(obs))
	}

	resp := ListObservationsResponse{
		Observations: observationResponses,
		Total:        len(observationResponses),
		Limit:        len(observationResponses),
		Offset:       0,
	}

	response.Success(c, resp)
}

// CreateObservationsBatch handles POST /api/v1/observability/observations/batch
func (h *Handler) CreateObservationsBatch(c *gin.Context) {
	var req BatchCreateObservationsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	// Convert requests to domain entities
	var observations []*observability.Observation
	for _, obsReq := range req.Observations {
		obs, err := h.requestToObservation(&obsReq)
		if err != nil {
			response.BadRequest(c, "Invalid observation data", err.Error())
			return
		}
		observations = append(observations, obs)
	}

	// Create observations batch via service
	createdObservations, err := h.services.GetObservationService().CreateObservationsBatch(c.Request.Context(), observations)
	if err != nil {
		if observability.IsValidationError(err) {
			response.ValidationError(c, "Validation failed", err.Error())
			return
		}
		response.InternalServerError(c, "Failed to create observations batch")
		return
	}

	// Convert to response
	var observationResponses []ObservationResponse
	for _, obs := range createdObservations {
		observationResponses = append(observationResponses, h.observationToResponse(obs))
	}

	resp := BatchCreateObservationsResponse{
		Observations:   observationResponses,
		ProcessedCount: len(createdObservations),
	}

	response.Created(c, resp)
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

// requestToObservation converts a CreateObservationRequest to an Observation domain entity
func (h *Handler) requestToObservation(req *CreateObservationRequest) (*observability.Observation, error) {
	observation := &observability.Observation{
		ExternalObservationID: req.ExternalObservationID,
		Type:                  observability.ObservationType(req.Type),
		Name:                  req.Name,
		StartTime:             req.StartTime,
		PromptTokens:          req.PromptTokens,
		CompletionTokens:      req.CompletionTokens,
		TotalTokens:           req.TotalTokens,
		Input:                 req.Input,
		Output:                req.Output,
		ModelParameters:       req.ModelParameters,
	}

	// Parse trace_id
	traceID, err := ulid.Parse(req.TraceID)
	if err != nil {
		return nil, err
	}
	observation.TraceID = traceID

	// Parse parent_observation_id
	if req.ParentObservationID != "" {
		parentObsID, err := ulid.Parse(req.ParentObservationID)
		if err != nil {
			return nil, err
		}
		observation.ParentObservationID = &parentObsID
	}

	// Set optional fields
	if req.EndTime != nil {
		observation.EndTime = req.EndTime
	}

	if req.Level != "" {
		observation.Level = observability.ObservationLevel(req.Level)
	} else {
		observation.Level = observability.ObservationLevelDefault
	}

	if req.StatusMessage != "" {
		observation.StatusMessage = &req.StatusMessage
	}

	if req.Version != "" {
		observation.Version = &req.Version
	}

	if req.Model != "" {
		observation.Model = &req.Model
	}

	if req.Provider != "" {
		observation.Provider = &req.Provider
	}

	if req.InputCost != nil {
		observation.InputCost = req.InputCost
	}

	if req.OutputCost != nil {
		observation.OutputCost = req.OutputCost
	}

	if req.TotalCost != nil {
		observation.TotalCost = req.TotalCost
	}

	if req.LatencyMs != nil {
		observation.LatencyMs = req.LatencyMs
	}

	if req.QualityScore != nil {
		observation.QualityScore = req.QualityScore
	}

	return observation, nil
}

// applyObservationUpdates applies updates from UpdateObservationRequest to an observation
func (h *Handler) applyObservationUpdates(obs *observability.Observation, req *UpdateObservationRequest) {
	if req.Name != "" {
		obs.Name = req.Name
	}

	if req.EndTime != nil {
		obs.EndTime = req.EndTime
	}

	if req.Level != "" {
		obs.Level = observability.ObservationLevel(req.Level)
	}

	if req.StatusMessage != "" {
		statusMessage := req.StatusMessage
		obs.StatusMessage = &statusMessage
	}

	if req.Version != "" {
		version := req.Version
		obs.Version = &version
	}

	if req.Model != "" {
		model := req.Model
		obs.Model = &model
	}

	if req.Provider != "" {
		provider := req.Provider
		obs.Provider = &provider
	}

	if req.Input != nil {
		obs.Input = req.Input
	}

	if req.Output != nil {
		obs.Output = req.Output
	}

	if req.ModelParameters != nil {
		obs.ModelParameters = req.ModelParameters
	}

	if req.PromptTokens > 0 {
		obs.PromptTokens = req.PromptTokens
	}

	if req.CompletionTokens > 0 {
		obs.CompletionTokens = req.CompletionTokens
	}

	if req.TotalTokens > 0 {
		obs.TotalTokens = req.TotalTokens
	}

	if req.InputCost != nil {
		obs.InputCost = req.InputCost
	}

	if req.OutputCost != nil {
		obs.OutputCost = req.OutputCost
	}

	if req.TotalCost != nil {
		obs.TotalCost = req.TotalCost
	}

	if req.LatencyMs != nil {
		obs.LatencyMs = req.LatencyMs
	}

	if req.QualityScore != nil {
		obs.QualityScore = req.QualityScore
	}
}

// observationToResponse converts an Observation domain entity to an ObservationResponse
func (h *Handler) observationToResponse(obs *observability.Observation) ObservationResponse {
	resp := ObservationResponse{
		ID:                    obs.ID.String(),
		TraceID:              obs.TraceID.String(),
		ExternalObservationID: obs.ExternalObservationID,
		Type:                 string(obs.Type),
		Name:                 obs.Name,
		StartTime:            obs.StartTime,
		Level:                string(obs.Level),
		PromptTokens:         obs.PromptTokens,
		CompletionTokens:     obs.CompletionTokens,
		TotalTokens:          obs.TotalTokens,
		Input:                obs.Input,
		Output:               obs.Output,
		ModelParameters:      obs.ModelParameters,
		CreatedAt:            obs.CreatedAt,
		UpdatedAt:            obs.UpdatedAt,
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