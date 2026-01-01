package evaluation

import (
	"strconv"

	"github.com/gin-gonic/gin"

	evaluationDomain "brokle/internal/core/domain/evaluation"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/pagination"
	"brokle/pkg/response"
	"brokle/pkg/ulid"
)

type RuleExecutionHandler struct {
	service evaluationDomain.RuleExecutionService
}

func NewRuleExecutionHandler(
	service evaluationDomain.RuleExecutionService,
) *RuleExecutionHandler {
	return &RuleExecutionHandler{
		service: service,
	}
}

// ExecutionListResponse wraps the list response with pagination metadata.
type ExecutionListResponse struct {
	Executions []*evaluationDomain.RuleExecutionResponse `json:"executions"`
	Total      int64                                     `json:"total"`
	Page       int                                       `json:"page"`
	Limit      int                                       `json:"limit"`
}

// @Summary List rule executions
// @Description Returns execution history for an evaluation rule with optional filtering and pagination.
// @Tags Evaluation Rule Executions
// @Produce json
// @Param projectId path string true "Project ID"
// @Param ruleId path string true "Rule ID"
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Items per page (10, 25, 50, 100; default 25)"
// @Param status query string false "Filter by status (pending, running, completed, failed, cancelled)"
// @Param trigger_type query string false "Filter by trigger type (automatic, manual)"
// @Success 200 {object} ExecutionListResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /api/v1/projects/{projectId}/evaluations/rules/{ruleId}/executions [get]
func (h *RuleExecutionHandler) List(c *gin.Context) {
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
	params.SetDefaults("created_at")

	var filter evaluationDomain.ExecutionFilter
	if status := c.Query("status"); status != "" {
		s := evaluationDomain.ExecutionStatus(status)
		filter.Status = &s
	}
	if triggerType := c.Query("trigger_type"); triggerType != "" {
		t := evaluationDomain.TriggerType(triggerType)
		filter.TriggerType = &t
	}

	executions, total, err := h.service.ListByRuleID(c.Request.Context(), ruleID, projectID, &filter, params)
	if err != nil {
		response.Error(c, err)
		return
	}

	responses := make([]*evaluationDomain.RuleExecutionResponse, len(executions))
	for i, execution := range executions {
		responses[i] = execution.ToResponse()
	}

	response.Success(c, &ExecutionListResponse{
		Executions: responses,
		Total:      total,
		Page:       params.Page,
		Limit:      params.Limit,
	})
}

// @Summary Get rule execution
// @Description Returns a specific rule execution by ID.
// @Tags Evaluation Rule Executions
// @Produce json
// @Param projectId path string true "Project ID"
// @Param ruleId path string true "Rule ID"
// @Param executionId path string true "Execution ID"
// @Success 200 {object} evaluation.RuleExecutionResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/projects/{projectId}/evaluations/rules/{ruleId}/executions/{executionId} [get]
func (h *RuleExecutionHandler) Get(c *gin.Context) {
	projectID, err := ulid.Parse(c.Param("projectId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("projectId", "must be a valid ULID"))
		return
	}

	// Parse ruleId for validation (we don't use it in the query since execution ID is unique)
	_, err = ulid.Parse(c.Param("ruleId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("ruleId", "must be a valid ULID"))
		return
	}

	executionID, err := ulid.Parse(c.Param("executionId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("executionId", "must be a valid ULID"))
		return
	}

	execution, err := h.service.GetByID(c.Request.Context(), executionID, projectID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, execution.ToResponse())
}

// @Summary Get latest rule execution
// @Description Returns the most recent execution for an evaluation rule.
// @Tags Evaluation Rule Executions
// @Produce json
// @Param projectId path string true "Project ID"
// @Param ruleId path string true "Rule ID"
// @Success 200 {object} evaluation.RuleExecutionResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse "No executions found"
// @Router /api/v1/projects/{projectId}/evaluations/rules/{ruleId}/executions/latest [get]
func (h *RuleExecutionHandler) GetLatest(c *gin.Context) {
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

	execution, err := h.service.GetLatestByRuleID(c.Request.Context(), ruleID, projectID)
	if err != nil {
		response.Error(c, err)
		return
	}

	if execution == nil {
		response.Error(c, appErrors.NewNotFoundError("no executions found for this rule"))
		return
	}

	response.Success(c, execution.ToResponse())
}
