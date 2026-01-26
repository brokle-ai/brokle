package evaluation

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"brokle/internal/core/domain/analytics"
	evaluationDomain "brokle/internal/core/domain/evaluation"
	"brokle/internal/transport/http/handlers/shared"
	"brokle/internal/transport/http/middleware"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/pagination"
	"brokle/pkg/response"
	"brokle/pkg/ulid"
)

type RuleHandler struct {
	service evaluationDomain.RuleService
}

func NewRuleHandler(
	service evaluationDomain.RuleService,
) *RuleHandler {
	return &RuleHandler{
		service: service,
	}
}

// RuleListResponse wraps the list response with pagination metadata.
type RuleListResponse struct {
	Rules []*evaluationDomain.EvaluationRuleResponse `json:"rules"`
	Total int64                                      `json:"total"`
	Page  int                                        `json:"page"`
	Limit int                                        `json:"limit"`
}

// @Summary Create evaluation rule
// @Description Creates a new evaluation rule for automated span scoring.
// @Tags Evaluation Rules
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID"
// @Param request body evaluation.CreateEvaluationRuleRequest true "Rule request"
// @Success 201 {object} evaluation.EvaluationRuleResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse "Name already exists"
// @Router /api/v1/projects/{projectId}/evaluations/rules [post]
func (h *RuleHandler) Create(c *gin.Context) {
	projectID, err := ulid.Parse(c.Param("projectId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("projectId", "must be a valid ULID"))
		return
	}

	var req evaluationDomain.CreateEvaluationRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Get user ID if available (for created_by tracking)
	var userID *ulid.ULID
	if uid, ok := middleware.GetUserIDULID(c); ok {
		userID = &uid
	}

	rule, err := h.service.Create(c.Request.Context(), projectID, userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, rule.ToResponse())
}

// @Summary List evaluation rules
// @Description Returns all evaluation rules for the project with optional filtering and pagination.
// @Tags Evaluation Rules
// @Produce json
// @Param projectId path string true "Project ID"
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Items per page (10, 25, 50, 100; default 50)"
// @Param status query string false "Filter by status (active, inactive, paused)"
// @Param scorer_type query string false "Filter by scorer type (llm, builtin, regex)"
// @Param search query string false "Search by name"
// @Success 200 {object} RuleListResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /api/v1/projects/{projectId}/evaluations/rules [get]
func (h *RuleHandler) List(c *gin.Context) {
	projectID, err := ulid.Parse(c.Param("projectId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("projectId", "must be a valid ULID"))
		return
	}

	// Parse pagination params
	params := pagination.Params{}
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed >= 1 {
			params.Page = parsed
		}
	}
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			params.Limit = parsed
		}
	}

	// Parse and validate sort params
	allowedSortFields := []string{"name", "status", "sampling_rate", "created_at", "updated_at"}
	if sortBy := c.Query("sort_by"); sortBy != "" {
		validatedField, err := pagination.ValidateSortField(sortBy, allowedSortFields)
		if err != nil {
			response.Error(c, appErrors.NewValidationError("sort_by", err.Error()))
			return
		}
		params.SortBy = validatedField
	}
	if sortDir := c.Query("sort_dir"); sortDir != "" {
		if sortDir != "asc" && sortDir != "desc" {
			response.Error(c, appErrors.NewValidationError("sort_dir", "must be 'asc' or 'desc'"))
			return
		}
		params.SortDir = sortDir
	}

	params.SetDefaults("created_at")

	// Parse filter params
	var filter evaluationDomain.RuleFilter
	if status := c.Query("status"); status != "" {
		s := evaluationDomain.RuleStatus(status)
		filter.Status = &s
	}
	if scorerType := c.Query("scorer_type"); scorerType != "" {
		st := evaluationDomain.ScorerType(scorerType)
		filter.ScorerType = &st
	}
	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}

	rules, total, err := h.service.List(c.Request.Context(), projectID, &filter, params)
	if err != nil {
		response.Error(c, err)
		return
	}

	responses := make([]*evaluationDomain.EvaluationRuleResponse, len(rules))
	for i, rule := range rules {
		responses[i] = rule.ToResponse()
	}

	response.Success(c, &RuleListResponse{
		Rules: responses,
		Total: total,
		Page:  params.Page,
		Limit: params.Limit,
	})
}

// @Summary Get evaluation rule
// @Description Returns the evaluation rule for a specific ID.
// @Tags Evaluation Rules
// @Produce json
// @Param projectId path string true "Project ID"
// @Param ruleId path string true "Rule ID"
// @Success 200 {object} evaluation.EvaluationRuleResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/projects/{projectId}/evaluations/rules/{ruleId} [get]
func (h *RuleHandler) Get(c *gin.Context) {
	projectID, err := ulid.Parse(c.Param("projectId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("projectId", "must be a valid ULID"))
		return
	}

	ruleID, err := ulid.Parse(c.Param("ruleId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("ruleId", "must be a valid ULID"))
		return
	}

	rule, err := h.service.GetByID(c.Request.Context(), ruleID, projectID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, rule.ToResponse())
}

// @Summary Update evaluation rule
// @Description Updates an existing evaluation rule by ID.
// @Tags Evaluation Rules
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID"
// @Param ruleId path string true "Rule ID"
// @Param request body evaluation.UpdateEvaluationRuleRequest true "Update request"
// @Success 200 {object} evaluation.EvaluationRuleResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse "Name already exists"
// @Router /api/v1/projects/{projectId}/evaluations/rules/{ruleId} [put]
func (h *RuleHandler) Update(c *gin.Context) {
	projectID, err := ulid.Parse(c.Param("projectId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("projectId", "must be a valid ULID"))
		return
	}

	ruleID, err := ulid.Parse(c.Param("ruleId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("ruleId", "must be a valid ULID"))
		return
	}

	var req evaluationDomain.UpdateEvaluationRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	rule, err := h.service.Update(c.Request.Context(), ruleID, projectID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, rule.ToResponse())
}

// @Summary Delete evaluation rule
// @Description Removes an evaluation rule by its ID.
// @Tags Evaluation Rules
// @Produce json
// @Param projectId path string true "Project ID"
// @Param ruleId path string true "Rule ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/projects/{projectId}/evaluations/rules/{ruleId} [delete]
func (h *RuleHandler) Delete(c *gin.Context) {
	projectID, err := ulid.Parse(c.Param("projectId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("projectId", "must be a valid ULID"))
		return
	}

	ruleID, err := ulid.Parse(c.Param("ruleId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("ruleId", "must be a valid ULID"))
		return
	}

	if err := h.service.Delete(c.Request.Context(), ruleID, projectID); err != nil {
		response.Error(c, err)
		return
	}

	response.NoContent(c)
}

// @Summary Activate evaluation rule
// @Description Activates an evaluation rule, enabling automatic span evaluation.
// @Tags Evaluation Rules
// @Produce json
// @Param projectId path string true "Project ID"
// @Param ruleId path string true "Rule ID"
// @Success 200 {object} response.SuccessResponse "Rule activated"
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/projects/{projectId}/evaluations/rules/{ruleId}/activate [post]
func (h *RuleHandler) Activate(c *gin.Context) {
	projectID, err := ulid.Parse(c.Param("projectId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("projectId", "must be a valid ULID"))
		return
	}

	ruleID, err := ulid.Parse(c.Param("ruleId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("ruleId", "must be a valid ULID"))
		return
	}

	if err := h.service.Activate(c.Request.Context(), ruleID, projectID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, map[string]string{"message": "rule activated"})
}

// @Summary Deactivate evaluation rule
// @Description Deactivates an evaluation rule, stopping automatic span evaluation.
// @Tags Evaluation Rules
// @Produce json
// @Param projectId path string true "Project ID"
// @Param ruleId path string true "Rule ID"
// @Success 200 {object} response.SuccessResponse "Rule deactivated"
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/projects/{projectId}/evaluations/rules/{ruleId}/deactivate [post]
func (h *RuleHandler) Deactivate(c *gin.Context) {
	projectID, err := ulid.Parse(c.Param("projectId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("projectId", "must be a valid ULID"))
		return
	}

	ruleID, err := ulid.Parse(c.Param("ruleId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("ruleId", "must be a valid ULID"))
		return
	}

	if err := h.service.Deactivate(c.Request.Context(), ruleID, projectID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, map[string]string{"message": "rule deactivated"})
}

// @Summary Trigger evaluation rule
// @Description Manually triggers an evaluation rule against matching spans. Returns 202 Accepted with execution ID for async processing.
// @Tags Evaluation Rules
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID"
// @Param ruleId path string true "Rule ID"
// @Param request body evaluation.TriggerOptions false "Optional trigger options"
// @Success 202 {object} evaluation.TriggerResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/projects/{projectId}/evaluations/rules/{ruleId}/trigger [post]
func (h *RuleHandler) Trigger(c *gin.Context) {
	projectID, err := ulid.Parse(c.Param("projectId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("projectId", "must be a valid ULID"))
		return
	}

	ruleID, err := ulid.Parse(c.Param("ruleId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("ruleId", "must be a valid ULID"))
		return
	}

	var opts *evaluationDomain.TriggerOptions
	if c.Request.ContentLength > 0 {
		var reqOpts evaluationDomain.TriggerOptions
		if err := c.ShouldBindJSON(&reqOpts); err != nil {
			response.ValidationError(c, "Invalid request body", err.Error())
			return
		}
		opts = &reqOpts
	}

	result, err := h.service.TriggerRule(c.Request.Context(), ruleID, projectID, opts)
	if err != nil {
		response.Error(c, err)
		return
	}

	// Return 202 Accepted for async processing (Opik pattern)
	response.Accepted(c, result)
}

// @Summary Test evaluation rule
// @Description Tests a rule against sample spans without persisting scores. Useful for validating rule configuration before activation.
// @Tags Evaluation Rules
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID"
// @Param ruleId path string true "Rule ID"
// @Param request body evaluation.TestRuleRequest false "Optional test parameters"
// @Success 200 {object} evaluation.TestRuleResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/projects/{projectId}/evaluations/rules/{ruleId}/test [post]
func (h *RuleHandler) Test(c *gin.Context) {
	projectID, err := ulid.Parse(c.Param("projectId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("projectId", "must be a valid ULID"))
		return
	}

	ruleID, err := ulid.Parse(c.Param("ruleId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("ruleId", "must be a valid ULID"))
		return
	}

	var req *evaluationDomain.TestRuleRequest
	if c.Request.ContentLength > 0 {
		var reqBody evaluationDomain.TestRuleRequest
		if err := c.ShouldBindJSON(&reqBody); err != nil {
			response.ValidationError(c, "Invalid request body", err.Error())
			return
		}
		req = &reqBody
	}

	result, err := h.service.TestRule(c.Request.Context(), ruleID, projectID, req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

// @Summary Get rule analytics
// @Description Returns performance analytics for an evaluation rule over a specified time period.
// @Tags Evaluation Rules
// @Produce json
// @Param projectId path string true "Project ID"
// @Param ruleId path string true "Rule ID"
// @Param period query string false "Time period: 24h, 7d, 30d (default: 7d)"
// @Param from_timestamp query string false "Custom start time (RFC3339 format)"
// @Param to_timestamp query string false "Custom end time (RFC3339 format)"
// @Success 200 {object} evaluation.RuleAnalyticsResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/projects/{projectId}/evaluations/rules/{ruleId}/analytics [get]
func (h *RuleHandler) GetAnalytics(c *gin.Context) {
	projectID, err := ulid.Parse(c.Param("projectId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("projectId", "must be a valid ULID"))
		return
	}

	ruleID, err := ulid.Parse(c.Param("ruleId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("ruleId", "must be a valid ULID"))
		return
	}

	// Parse time range using shared utility
	fromTimestamp := c.Query("from_timestamp")
	toTimestamp := c.Query("to_timestamp")
	period := c.DefaultQuery("period", "7d")

	fromTime, toTime, err := shared.ParseTimeRange(
		fromTimestamp,
		toTimestamp,
		period,
		analytics.TimeRange7Days,
	)
	if err != nil {
		response.Error(c, err)
		return
	}

	params := &evaluationDomain.RuleAnalyticsParams{
		ProjectID: projectID,
		RuleID:    ruleID,
		Period:    period,
		From:      &fromTime,
		To:        &toTime,
	}

	result, err := h.service.GetAnalytics(c.Request.Context(), ruleID, projectID, params)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}
