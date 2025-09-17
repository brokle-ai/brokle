package observability

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"brokle/internal/core/domain/observability"
	"brokle/pkg/response"
	"brokle/pkg/ulid"
)

// CreateQualityScore handles POST /api/v1/observability/quality-scores
// @Summary Create a quality score
// @Description Create a new quality evaluation score for a trace or observation (numeric, categorical, or boolean)
// @Tags Observability - Quality
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateQualityScoreRequest true "Quality score creation data"
// @Success 201 {object} response.SuccessResponse{data=QualityScoreResponse} "Quality score created successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload or quality score data"
// @Failure 409 {object} response.ErrorResponse "Duplicate quality score for same trace/observation and score name"
// @Failure 422 {object} response.ErrorResponse "Validation failed"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/observability/quality-scores [post]
func (h *Handler) CreateQualityScore(c *gin.Context) {
	var req CreateQualityScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	// Convert request to domain entity
	score, err := h.requestToQualityScore(&req)
	if err != nil {
		response.BadRequest(c, "Invalid quality score data", err.Error())
		return
	}

	// Create quality score via service
	createdScore, err := h.services.GetQualityService().CreateQualityScore(c.Request.Context(), score)
	if err != nil {
		if observability.IsValidationError(err) {
			response.ValidationError(c, "Validation failed", err.Error())
			return
		}
		if observability.IsConflictError(err) {
			response.Conflict(c, "Quality score already exists")
			return
		}
		response.InternalServerError(c, "Failed to create quality score")
		return
	}

	// Convert to response
	resp := h.qualityScoreToResponse(createdScore)
	response.Created(c, resp)
}

// GetQualityScore handles GET /api/v1/observability/quality-scores/{id}
func (h *Handler) GetQualityScore(c *gin.Context) {
	idStr := c.Param("id")

	scoreID, err := ulid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid quality score ID", err.Error())
		return
	}

	// Get quality score via service
	score, err := h.services.GetQualityService().GetQualityScore(c.Request.Context(), scoreID)
	if err != nil {
		if observability.IsNotFoundError(err) {
			response.NotFound(c, "Quality score")
			return
		}
		response.InternalServerError(c, "Failed to get quality score")
		return
	}

	// Convert to response
	resp := h.qualityScoreToResponse(score)
	response.Success(c, resp)
}

// UpdateQualityScore handles PUT /api/v1/observability/quality-scores/{id}
func (h *Handler) UpdateQualityScore(c *gin.Context) {
	idStr := c.Param("id")

	scoreID, err := ulid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid quality score ID", err.Error())
		return
	}

	var req UpdateQualityScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	// Get existing score first
	score, err := h.services.GetQualityService().GetQualityScore(c.Request.Context(), scoreID)
	if err != nil {
		if observability.IsNotFoundError(err) {
			response.NotFound(c, "Quality score")
			return
		}
		response.InternalServerError(c, "Failed to get quality score")
		return
	}

	// Apply updates from request
	h.applyQualityScoreUpdates(score, &req)

	// Update quality score via service
	updatedScore, err := h.services.GetQualityService().UpdateQualityScore(c.Request.Context(), score)
	if err != nil {
		if observability.IsValidationError(err) {
			response.ValidationError(c, "Validation failed", err.Error())
			return
		}
		response.InternalServerError(c, "Failed to update quality score")
		return
	}

	// Convert to response
	resp := h.qualityScoreToResponse(updatedScore)
	response.Success(c, resp)
}

// DeleteQualityScore handles DELETE /api/v1/observability/quality-scores/{id}
func (h *Handler) DeleteQualityScore(c *gin.Context) {
	idStr := c.Param("id")

	scoreID, err := ulid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid quality score ID", err.Error())
		return
	}

	// Delete quality score via service
	if err := h.services.GetQualityService().DeleteQualityScore(c.Request.Context(), scoreID); err != nil {
		if observability.IsNotFoundError(err) {
			response.NotFound(c, "Quality score")
			return
		}
		response.InternalServerError(c, "Failed to delete quality score")
		return
	}

	response.Success(c, gin.H{"message": "Quality score deleted successfully"})
}

// ListQualityScores handles GET /api/v1/observability/quality-scores
func (h *Handler) ListQualityScores(c *gin.Context) {
	// Parse query parameters
	filter, err := h.parseQualityScoreFilter(c)
	if err != nil {
		response.BadRequest(c, "Invalid filter parameters", err.Error())
		return
	}

	// Get quality scores via service
	scores, total, err := h.services.GetQualityService().ListQualityScores(c.Request.Context(), filter)
	if err != nil {
		response.InternalServerError(c, "Failed to list quality scores")
		return
	}

	// Convert to response
	var scoreResponses []QualityScoreResponse
	for _, score := range scores {
		scoreResponses = append(scoreResponses, h.qualityScoreToResponse(score))
	}

	resp := ListQualityScoresResponse{
		QualityScores: scoreResponses,
		Total:         total,
		Limit:         filter.Limit,
		Offset:        filter.Offset,
	}

	// Create pagination metadata
	pagination := response.NewPagination(filter.Offset/filter.Limit+1, filter.Limit, int64(total))
	response.SuccessWithPagination(c, resp, pagination)
}

// GetQualityScoresByTrace handles GET /api/v1/observability/traces/{id}/quality-scores
func (h *Handler) GetQualityScoresByTrace(c *gin.Context) {
	traceIDStr := c.Param("id")

	traceID, err := ulid.Parse(traceIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid trace ID", err.Error())
		return
	}

	// Get quality scores by trace via service
	scores, err := h.services.GetQualityService().GetQualityScoresByTrace(c.Request.Context(), traceID)
	if err != nil {
		response.InternalServerError(c, "Failed to get quality scores by trace")
		return
	}

	// Convert to response
	var scoreResponses []QualityScoreResponse
	for _, score := range scores {
		scoreResponses = append(scoreResponses, h.qualityScoreToResponse(score))
	}

	resp := ListQualityScoresResponse{
		QualityScores: scoreResponses,
		Total:  len(scoreResponses),
		Limit:  len(scoreResponses),
		Offset: 0,
	}

	response.Success(c, resp)
}

// GetQualityScoresByObservation handles GET /api/v1/observability/observations/{id}/quality-scores
func (h *Handler) GetQualityScoresByObservation(c *gin.Context) {
	observationIDStr := c.Param("id")

	observationID, err := ulid.Parse(observationIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid observation ID", err.Error())
		return
	}

	// Get quality scores by observation via service
	scores, err := h.services.GetQualityService().GetQualityScoresByObservation(c.Request.Context(), observationID)
	if err != nil {
		response.InternalServerError(c, "Failed to get quality scores by observation")
		return
	}

	// Convert to response
	var scoreResponses []QualityScoreResponse
	for _, score := range scores {
		scoreResponses = append(scoreResponses, h.qualityScoreToResponse(score))
	}

	resp := ListQualityScoresResponse{
		QualityScores: scoreResponses,
		Total:  len(scoreResponses),
		Limit:  len(scoreResponses),
		Offset: 0,
	}

	response.Success(c, resp)
}

// Helper methods

// parseQualityScoreFilter parses query parameters into a QualityScoreFilter
func (h *Handler) parseQualityScoreFilter(c *gin.Context) (*observability.QualityScoreFilter, error) {
	filter := &observability.QualityScoreFilter{}

	// Parse trace_id
	if traceIDStr := c.Query("trace_id"); traceIDStr != "" {
		traceID, err := ulid.Parse(traceIDStr)
		if err != nil {
			return nil, err
		}
		filter.TraceID = &traceID
	}

	// Parse observation_id
	if observationIDStr := c.Query("observation_id"); observationIDStr != "" {
		observationID, err := ulid.Parse(observationIDStr)
		if err != nil {
			return nil, err
		}
		filter.ObservationID = &observationID
	}

	// Parse score_name
	if scoreName := c.Query("score_name"); scoreName != "" {
		filter.ScoreName = &scoreName
	}

	// Parse source
	if source := c.Query("source"); source != "" {
		sourceEnum := observability.ScoreSource(source)
		filter.Source = &sourceEnum
	}

	// Parse data_type
	if dataType := c.Query("data_type"); dataType != "" {
		dataTypeEnum := observability.ScoreDataType(dataType)
		filter.DataType = &dataTypeEnum
	}

	// Parse evaluator_name
	if evaluatorName := c.Query("evaluator_name"); evaluatorName != "" {
		filter.EvaluatorName = &evaluatorName
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

// requestToQualityScore converts a CreateQualityScoreRequest to a QualityScore domain entity
func (h *Handler) requestToQualityScore(req *CreateQualityScoreRequest) (*observability.QualityScore, error) {
	score := &observability.QualityScore{
		ScoreName: req.ScoreName,
		DataType:  observability.ScoreDataType(req.DataType),
		Source:    observability.ScoreSource(req.Source),
	}

	// Parse trace_id
	traceID, err := ulid.Parse(req.TraceID)
	if err != nil {
		return nil, err
	}
	score.TraceID = traceID

	// Parse observation_id (optional)
	if req.ObservationID != "" {
		observationID, err := ulid.Parse(req.ObservationID)
		if err != nil {
			return nil, err
		}
		score.ObservationID = &observationID
	}

	// Parse author_user_id (optional)
	if req.AuthorUserID != "" {
		authorUserID, err := ulid.Parse(req.AuthorUserID)
		if err != nil {
			return nil, err
		}
		score.AuthorUserID = &authorUserID
	}

	// Set optional fields
	if req.ScoreValue != nil {
		score.ScoreValue = req.ScoreValue
	}

	if req.StringValue != nil && *req.StringValue != "" {
		score.StringValue = req.StringValue
	}

	if req.EvaluatorName != "" {
		score.EvaluatorName = &req.EvaluatorName
	}

	if req.EvaluatorVersion != "" {
		score.EvaluatorVersion = &req.EvaluatorVersion
	}

	if req.Comment != "" {
		score.Comment = &req.Comment
	}

	return score, nil
}

// applyQualityScoreUpdates applies updates from UpdateQualityScoreRequest to a quality score
func (h *Handler) applyQualityScoreUpdates(score *observability.QualityScore, req *UpdateQualityScoreRequest) {
	if req.ScoreValue != nil {
		score.ScoreValue = req.ScoreValue
	}

	if req.StringValue != nil {
		score.StringValue = req.StringValue
	}

	if req.Comment != "" {
		comment := req.Comment
		score.Comment = &comment
	}

	if req.EvaluatorVersion != "" {
		evaluatorVersion := req.EvaluatorVersion
		score.EvaluatorVersion = &evaluatorVersion
	}
}

// qualityScoreToResponse converts a QualityScore domain entity to a QualityScoreResponse
func (h *Handler) qualityScoreToResponse(score *observability.QualityScore) QualityScoreResponse {
	resp := QualityScoreResponse{
		ID:        score.ID.String(),
		TraceID:   score.TraceID.String(),
		ScoreName: score.ScoreName,
		DataType:  string(score.DataType),
		Source:    string(score.Source),
		CreatedAt: score.CreatedAt,
		UpdatedAt: score.UpdatedAt,
	}

	if score.ObservationID != nil {
		resp.ObservationID = score.ObservationID.String()
	}

	if score.AuthorUserID != nil {
		resp.AuthorUserID = score.AuthorUserID.String()
	}

	if score.ScoreValue != nil {
		resp.ScoreValue = score.ScoreValue
	}

	if score.StringValue != nil {
		resp.StringValue = score.StringValue
	}

	if score.EvaluatorName != nil {
		resp.EvaluatorName = *score.EvaluatorName
	}

	if score.EvaluatorVersion != nil {
		resp.EvaluatorVersion = *score.EvaluatorVersion
	}

	if score.Comment != nil {
		resp.Comment = *score.Comment
	}

	return resp
}