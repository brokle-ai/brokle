package prompt

import (
	"github.com/gin-gonic/gin"

	promptDomain "brokle/internal/core/domain/prompt"
	"brokle/internal/transport/http/middleware"
	"brokle/pkg/response"
	"brokle/pkg/ulid"
)

// SetLabels handles PATCH /api/v1/projects/:projectId/prompts/:promptId/versions/:versionId/labels
// @Summary Set labels on a version
// @Description Add or update labels pointing to a specific version
// @Tags Prompts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param projectId path string true "Project ID"
// @Param promptId path string true "Prompt ID"
// @Param versionId path string true "Version ID"
// @Param request body prompt.SetLabelsRequest true "Set labels request"
// @Success 200 {object} response.APIResponse "Labels set successfully"
// @Failure 400 {object} response.APIResponse{error=response.APIError} "Invalid request"
// @Failure 401 {object} response.APIResponse{error=response.APIError} "Unauthorized"
// @Failure 403 {object} response.APIResponse{error=response.APIError} "Protected label modification forbidden"
// @Failure 404 {object} response.APIResponse{error=response.APIError} "Version not found"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/projects/{projectId}/prompts/{promptId}/versions/{versionId}/labels [patch]
func (h *Handler) SetLabels(c *gin.Context) {
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

	var req promptDomain.SetLabelsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "invalid request body", err.Error())
		return
	}

	var userID *ulid.ULID
	if uid, ok := middleware.GetUserIDULID(c); ok {
		userID = &uid
	}

	if err := h.promptService.SetLabels(c.Request.Context(), projectID, promptID, versionID, userID, req.Labels); err != nil {
		h.logger.Error("Failed to set labels", "version_id", versionID, "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "labels set successfully"})
}

// GetProtectedLabels handles GET /api/v1/projects/:projectId/prompts/settings/protected-labels
// @Summary Get protected labels
// @Description Retrieve the list of protected labels for a project
// @Tags Prompts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param projectId path string true "Project ID"
// @Success 200 {object} response.APIResponse{data=[]string} "List of protected labels"
// @Failure 400 {object} response.APIResponse{error=response.APIError} "Invalid parameters"
// @Failure 401 {object} response.APIResponse{error=response.APIError} "Unauthorized"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/projects/{projectId}/prompts/settings/protected-labels [get]
func (h *Handler) GetProtectedLabels(c *gin.Context) {
	projectID, err := ulid.Parse(c.Param("projectId"))
	if err != nil {
		response.ValidationError(c, "invalid project_id", "project_id must be a valid ULID")
		return
	}

	labels, err := h.promptService.GetProtectedLabels(c.Request.Context(), projectID)
	if err != nil {
		h.logger.Error("Failed to get protected labels", "project_id", projectID, "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"protected_labels": labels})
}

// SetProtectedLabels handles PUT /api/v1/projects/:projectId/prompts/settings/protected-labels
// @Summary Set protected labels
// @Description Update the list of protected labels for a project
// @Tags Prompts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param projectId path string true "Project ID"
// @Param request body prompt.ProtectedLabelsRequest true "Protected labels request"
// @Success 200 {object} response.APIResponse "Protected labels updated"
// @Failure 400 {object} response.APIResponse{error=response.APIError} "Invalid request"
// @Failure 401 {object} response.APIResponse{error=response.APIError} "Unauthorized"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/projects/{projectId}/prompts/settings/protected-labels [put]
func (h *Handler) SetProtectedLabels(c *gin.Context) {
	projectID, err := ulid.Parse(c.Param("projectId"))
	if err != nil {
		response.ValidationError(c, "invalid project_id", "project_id must be a valid ULID")
		return
	}

	var req promptDomain.ProtectedLabelsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "invalid request body", err.Error())
		return
	}

	var userID *ulid.ULID
	if uid, ok := middleware.GetUserIDULID(c); ok {
		userID = &uid
	}

	if err := h.promptService.SetProtectedLabels(c.Request.Context(), projectID, userID, req.ProtectedLabels); err != nil {
		h.logger.Error("Failed to set protected labels", "project_id", projectID, "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "protected labels updated successfully"})
}
