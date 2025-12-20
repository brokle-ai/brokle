// Package evaluation provides HTTP handlers for evaluation domain operations
// including score configuration management and SDK score ingestion.
package evaluation

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	evaluationDomain "brokle/internal/core/domain/evaluation"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/response"
	"brokle/pkg/ulid"
)

type ScoreConfigHandler struct {
	logger  *slog.Logger
	service evaluationDomain.ScoreConfigService
}

func NewScoreConfigHandler(
	logger *slog.Logger,
	service evaluationDomain.ScoreConfigService,
) *ScoreConfigHandler {
	return &ScoreConfigHandler{
		logger:  logger,
		service: service,
	}
}

type CreateRequest struct {
	Name        string                         `json:"name" binding:"required,min=1,max=100"`
	Description *string                        `json:"description,omitempty"`
	DataType    evaluationDomain.ScoreDataType `json:"data_type" binding:"required,oneof=NUMERIC CATEGORICAL BOOLEAN"`
	MinValue    *float64                       `json:"min_value,omitempty"`
	MaxValue    *float64                       `json:"max_value,omitempty"`
	Categories  []string                       `json:"categories,omitempty"`
	Metadata    map[string]interface{}         `json:"metadata,omitempty"`
}

type UpdateRequest struct {
	Name        *string                `json:"name,omitempty" binding:"omitempty,min=1,max=100"`
	Description *string                `json:"description,omitempty"`
	MinValue    *float64               `json:"min_value,omitempty"`
	MaxValue    *float64               `json:"max_value,omitempty"`
	Categories  []string               `json:"categories,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// @Summary Create score config
// @Description Creates a new score configuration for the project. Score configs define validation rules for scores.
// @Tags Score Configs
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID"
// @Param request body CreateRequest true "Score config request"
// @Success 201 {object} evaluation.ScoreConfigResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse "Name already exists"
// @Router /api/v1/projects/{projectId}/score-configs [post]
func (h *ScoreConfigHandler) Create(c *gin.Context) {
	projectID, err := ulid.Parse(c.Param("projectId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("projectId", "must be a valid ULID"))
		return
	}

	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	domainReq := &evaluationDomain.CreateScoreConfigRequest{
		Name:        req.Name,
		Description: req.Description,
		DataType:    req.DataType,
		MinValue:    req.MinValue,
		MaxValue:    req.MaxValue,
		Categories:  req.Categories,
		Metadata:    req.Metadata,
	}

	config, err := h.service.Create(c.Request.Context(), projectID, domainReq)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, config.ToResponse())
}

// @Summary List score configs
// @Description Returns all score configurations for the project.
// @Tags Score Configs
// @Produce json
// @Param projectId path string true "Project ID"
// @Success 200 {array} evaluation.ScoreConfigResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /api/v1/projects/{projectId}/score-configs [get]
func (h *ScoreConfigHandler) List(c *gin.Context) {
	projectID, err := ulid.Parse(c.Param("projectId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("projectId", "must be a valid ULID"))
		return
	}

	configs, err := h.service.List(c.Request.Context(), projectID)
	if err != nil {
		response.Error(c, err)
		return
	}

	responses := make([]*evaluationDomain.ScoreConfigResponse, len(configs))
	for i, config := range configs {
		responses[i] = config.ToResponse()
	}

	response.Success(c, responses)
}

// @Summary Get score config
// @Description Returns the score configuration for a specific config ID.
// @Tags Score Configs
// @Produce json
// @Param projectId path string true "Project ID"
// @Param configId path string true "Score Config ID"
// @Success 200 {object} evaluation.ScoreConfigResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/projects/{projectId}/score-configs/{configId} [get]
func (h *ScoreConfigHandler) Get(c *gin.Context) {
	projectID, err := ulid.Parse(c.Param("projectId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("projectId", "must be a valid ULID"))
		return
	}

	configID, err := ulid.Parse(c.Param("configId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("configId", "must be a valid ULID"))
		return
	}

	config, err := h.service.GetByID(c.Request.Context(), configID, projectID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, config.ToResponse())
}

// @Summary Update score config
// @Description Updates an existing score configuration by ID.
// @Tags Score Configs
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID"
// @Param configId path string true "Score Config ID"
// @Param request body UpdateRequest true "Update request"
// @Success 200 {object} evaluation.ScoreConfigResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse "Name already exists"
// @Router /api/v1/projects/{projectId}/score-configs/{configId} [put]
func (h *ScoreConfigHandler) Update(c *gin.Context) {
	projectID, err := ulid.Parse(c.Param("projectId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("projectId", "must be a valid ULID"))
		return
	}

	configID, err := ulid.Parse(c.Param("configId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("configId", "must be a valid ULID"))
		return
	}

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	domainReq := &evaluationDomain.UpdateScoreConfigRequest{
		Name:        req.Name,
		Description: req.Description,
		MinValue:    req.MinValue,
		MaxValue:    req.MaxValue,
		Categories:  req.Categories,
		Metadata:    req.Metadata,
	}

	config, err := h.service.Update(c.Request.Context(), configID, projectID, domainReq)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, config.ToResponse())
}

// @Summary Delete score config
// @Description Removes a score configuration by its ID.
// @Tags Score Configs
// @Produce json
// @Param projectId path string true "Project ID"
// @Param configId path string true "Score Config ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/projects/{projectId}/score-configs/{configId} [delete]
func (h *ScoreConfigHandler) Delete(c *gin.Context) {
	projectID, err := ulid.Parse(c.Param("projectId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("projectId", "must be a valid ULID"))
		return
	}

	configID, err := ulid.Parse(c.Param("configId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("configId", "must be a valid ULID"))
		return
	}

	if err := h.service.Delete(c.Request.Context(), configID, projectID); err != nil {
		response.Error(c, err)
		return
	}

	response.NoContent(c)
}
