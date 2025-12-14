package prompt

import (
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"

	promptDomain "brokle/internal/core/domain/prompt"
	"brokle/internal/transport/http/middleware"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/response"
	"brokle/pkg/ulid"
)

// ExecutePrompt handles POST /api/v1/projects/:projectId/prompts/:promptId/versions/:versionId/execute
// @Summary Execute a prompt with LLM
// @Description Execute a prompt version with variable substitution and LLM call
// @Tags Prompts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param projectId path string true "Project ID"
// @Param promptId path string true "Prompt ID"
// @Param versionId path string true "Version ID"
// @Param request body prompt.ExecutePromptRequest true "Execute prompt request"
// @Success 200 {object} response.APIResponse{data=prompt.ExecutePromptResponse} "Execution result"
// @Failure 400 {object} response.APIResponse{error=response.APIError} "Invalid request"
// @Failure 401 {object} response.APIResponse{error=response.APIError} "Unauthorized"
// @Failure 404 {object} response.APIResponse{error=response.APIError} "Version not found"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/projects/{projectId}/prompts/{promptId}/versions/{versionId}/execute [post]
func (h *Handler) ExecutePrompt(c *gin.Context) {
	projectID, err := ulid.Parse(c.Param("projectId"))
	if err != nil {
		response.ValidationError(c, "invalid project_id", "project_id must be a valid ULID")
		return
	}

	promptID, err := ulid.Parse(c.Param("promptId"))
	if err != nil {
		response.ValidationError(c, "invalid prompt_id", "prompt_id must be a valid ULID")
		return
	}

	versionID, err := ulid.Parse(c.Param("versionId"))
	if err != nil {
		response.ValidationError(c, "invalid version_id", "version_id must be a valid ULID")
		return
	}

	var req promptDomain.ExecutePromptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "invalid request body", err.Error())
		return
	}

	prompt, err := h.promptService.GetPromptByID(c.Request.Context(), projectID, promptID)
	if err != nil {
		h.logger.Error("Failed to get prompt", "prompt_id", promptID, "error", err)
		response.Error(c, err)
		return
	}

	version, err := h.promptService.GetVersionByID(c.Request.Context(), projectID, promptID, versionID)
	if err != nil {
		h.logger.Error("Failed to get version", "version_id", versionID, "error", err)
		response.Error(c, err)
		return
	}

	var template interface{}
	if err := json.Unmarshal(version.Template, &template); err != nil {
		h.logger.Error("Failed to unmarshal template", "version_id", versionID, "error", err)
		response.Error(c, appErrors.NewInternalError("failed to parse template", err))
		return
	}

	promptResp := &promptDomain.PromptResponse{
		ID:        prompt.ID.String(),
		Name:      prompt.Name,
		Type:      prompt.Type,
		Version:   version.Version,
		Template:  template,
		Variables: []string(version.Variables),
	}

	if len(version.Config) > 0 {
		config, _ := version.GetModelConfig()
		promptResp.Config = config
	}

	result, err := h.executionService.Execute(c.Request.Context(), promptResp, req.Variables, req.ConfigOverrides)
	if err != nil {
		h.logger.Error("Failed to execute prompt", "prompt_id", promptID, "version_id", versionID, "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

// ExecutePromptSDK handles POST /v1/prompts/:name/execute (SDK)
// @Summary Execute a prompt by name
// @Description Execute a prompt with variable substitution and optional LLM call
// @Tags SDK Prompts
// @Accept json
// @Produce json
// @Security APIKeyAuth
// @Param name path string true "Prompt name"
// @Param label query string false "Label to resolve (default: latest)"
// @Param version query int false "Specific version number (takes precedence over label)"
// @Param request body prompt.ExecutePromptRequest true "Execute prompt request"
// @Success 200 {object} response.APIResponse{data=prompt.ExecutePromptResponse} "Execution result"
// @Failure 400 {object} response.APIResponse{error=response.APIError} "Invalid request"
// @Failure 401 {object} response.APIResponse{error=response.APIError} "Unauthorized"
// @Failure 404 {object} response.APIResponse{error=response.APIError} "Prompt not found"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /v1/prompts/{name}/execute [post]
func (h *Handler) ExecutePromptSDK(c *gin.Context) {
	projectID, ok := middleware.GetProjectID(c)
	if !ok || projectID == nil {
		response.Unauthorized(c, "Invalid API key")
		return
	}

	name := c.Param("name")
	if name == "" {
		response.ValidationError(c, "name is required", "prompt name path parameter is required")
		return
	}

	var req promptDomain.ExecutePromptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "invalid request body", err.Error())
		return
	}

	opts := &promptDomain.GetPromptOptions{
		Label: c.DefaultQuery("label", "latest"),
	}
	if versionStr := c.Query("version"); versionStr != "" {
		version, err := strconv.Atoi(versionStr)
		if err != nil {
			response.ValidationError(c, "invalid version", "version must be an integer")
			return
		}
		opts.Version = &version
		opts.Label = ""
	}

	promptResp, err := h.promptService.GetPrompt(c.Request.Context(), *projectID, name, opts)
	if err != nil {
		h.logger.Error("Failed to get prompt", "name", name, "error", err)
		response.Error(c, err)
		return
	}

	result, err := h.executionService.Execute(c.Request.Context(), promptResp, req.Variables, req.ConfigOverrides)
	if err != nil {
		h.logger.Error("Failed to execute prompt", "name", name, "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}
