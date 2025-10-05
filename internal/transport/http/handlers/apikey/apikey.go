package apikey

import (
	"strconv"
	"time"

	"brokle/internal/config"
	"brokle/internal/core/domain/auth"
	"brokle/internal/transport/http/middleware"
	"brokle/pkg/response"
	"brokle/pkg/ulid"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	config        *config.Config
	logger        *logrus.Logger
	apiKeyService auth.APIKeyService
}

func NewHandler(config *config.Config, logger *logrus.Logger, apiKeyService auth.APIKeyService) *Handler {
	return &Handler{
		config:        config,
		logger:        logger,
		apiKeyService: apiKeyService,
	}
}

// Request/Response Models

// APIKey represents an API key entity for response
type APIKey struct {
	ID         string    `json:"id" example:"key_01234567890123456789012345" description:"Unique API key identifier"`
	Name       string    `json:"name" example:"Production API Key" description:"Human-readable name for the API key"`
	Key        string    `json:"key,omitempty" example:"bk_proj_01234567890123456789012345_abcdef1234567890abcdef1234567890" description:"The actual API key (only shown on creation)"`
	KeyPreview string    `json:"key_preview" example:"bk_proj_...7890" description:"Truncated version of the key for display"`
	ProjectID  string    `json:"project_id" example:"proj_01234567890123456789012345" description:"Project ID this key belongs to"`
	Status     string    `json:"status" example:"active" description:"API key status (active, inactive, expired)"`
	LastUsed   time.Time `json:"last_used,omitempty" example:"2024-01-01T00:00:00Z" description:"Last time this key was used (null if never used)"`
	CreatedAt  time.Time `json:"created_at" example:"2024-01-01T00:00:00Z" description:"Creation timestamp"`
	ExpiresAt  time.Time `json:"expires_at,omitempty" example:"2024-12-31T23:59:59Z" description:"Expiration timestamp (null if never expires)"`
	CreatedBy  string    `json:"created_by" example:"usr_01234567890123456789012345" description:"User ID who created this key"`
}

// CreateAPIKeyRequest represents the request to create an API key
type CreateAPIKeyRequest struct {
	Name         string `json:"name" binding:"required,min=2,max=100" example:"Production API Key" description:"Human-readable name for the API key (2-100 characters)"`
	ExpiryOption string `json:"expiry_option" binding:"required,oneof=30days 90days never" example:"90days" description:"Expiration option: '30days', '90days', or 'never'"`
}

// ListAPIKeysResponse represents the response when listing API keys
// NOTE: This struct is not used. When implementing, use response.SuccessWithPagination()
// with []APIKey directly and response.NewPagination() for consistent pagination format.
type ListAPIKeysResponse struct {
	APIKeys []APIKey `json:"api_keys" description:"List of API keys"`
	// Pagination fields removed - use response.SuccessWithPagination() instead
}

// List handles GET /projects/:projectId/api-keys
// @Summary List project-scoped API keys
// @Description Get a paginated list of project-scoped API keys for a specific project. Keys are shown with preview format (bk_proj_...7890) for security.
// @Tags API Keys
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID" example("proj_01234567890123456789012345")
// @Param status query string false "Filter by API key status" Enums(active,inactive,expired)
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Items per page" default(20) minimum(1) maximum(100)
// @Success 200 {object} response.APIResponse{data=[]APIKey,meta=response.Meta{pagination=response.Pagination}} "List of project-scoped API keys with pagination"
// @Failure 400 {object} response.ErrorResponse "Bad request - invalid project ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - insufficient permissions to view API keys"
// @Failure 404 {object} response.ErrorResponse "Project not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/projects/{projectId}/api-keys [get]
func (h *Handler) List(c *gin.Context) {
	// Get project ID from URL path
	projectID, err := ulid.Parse(c.Param("projectId"))
	if err != nil {
		h.logger.WithError(err).Error("Invalid project ID")
		response.BadRequest(c, "Invalid project ID", err.Error())
		return
	}

	// Get authenticated user from context
	userID, exists := middleware.GetUserID(c)
	if !exists {
		h.logger.Error("User ID not found in context")
		response.Unauthorized(c, "Authentication required")
		return
	}

	// Parse query parameters
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	status := c.Query("status")

	// Create filters
	filters := &auth.APIKeyFilters{
		Limit:     limit,
		Offset:    (page - 1) * limit,
		SortBy:    "created_at",
		SortOrder: "desc",
	}

	// Note: Environment filtering removed - environments are handled via SDK headers/tags

	// Filter by status if provided
	if status != "" {
		isActive := status == "active"
		filters.IsActive = &isActive
	}

	// Get API keys
	apiKeys, err := h.apiKeyService.GetAPIKeys(c.Request.Context(), filters)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list API keys")
		response.Error(c, err)
		return
	}

	// Convert to response format
	responseKeys := make([]APIKey, len(apiKeys))
	for i, key := range apiKeys {
		responseKeys[i] = APIKey{
			ID:         key.ID.String(),
			Name:       key.Name,
			KeyPreview: key.KeyPreview, // Use stored preview
			ProjectID:  key.ProjectID.String(),
			Status:     getKeyStatus(*key),
			CreatedAt:  key.CreatedAt,
			CreatedBy:  key.UserID.String(),
		}

		if key.LastUsedAt != nil {
			responseKeys[i].LastUsed = *key.LastUsedAt
		}
		if key.ExpiresAt != nil {
			responseKeys[i].ExpiresAt = *key.ExpiresAt
		}
	}

	h.logger.WithFields(map[string]interface{}{
		"user_id":    userID,
		"project_id": projectID,
		"count":      len(responseKeys),
		"page":       page,
		"limit":      limit,
	}).Debug("Listed API keys")

	// Use response.SuccessWithPagination for consistent pagination
	pagination := response.NewPagination(page, limit, int64(len(responseKeys)))
	response.SuccessWithPagination(c, responseKeys, pagination)
}

// getKeyStatus determines the status of an API key
func getKeyStatus(key auth.APIKey) string {
	if !key.IsActive {
		return "inactive"
	}
	if key.IsExpired() {
		return "expired"
	}
	return "active"
}

// Create handles POST /projects/:projectId/api-keys
// @Summary Create project-scoped API key
// @Description Create a new project-scoped API key with embedded project context. The full key will only be displayed once upon creation. Format: bk_proj_{project_id}_{secret}
// @Tags API Keys
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID" example("proj_01234567890123456789012345")
// @Param request body CreateAPIKeyRequest true "API key details"
// @Success 201 {object} response.SuccessResponse{data=APIKey} "Project-scoped API key created successfully (full key only shown once)"
// @Failure 400 {object} response.ErrorResponse "Bad request - invalid input or validation errors"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - insufficient permissions to create API keys"
// @Failure 404 {object} response.ErrorResponse "Project not found"
// @Failure 409 {object} response.ErrorResponse "Conflict - API key name already exists in project"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/projects/{projectId}/api-keys [post]
func (h *Handler) Create(c *gin.Context) {
	// Get project ID from URL path
	projectID, err := ulid.Parse(c.Param("projectId"))
	if err != nil {
		h.logger.WithError(err).Error("Invalid project ID")
		response.BadRequest(c, "Invalid project ID", err.Error())
		return
	}

	// Parse request body
	var req CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid create API key request")
		response.BadRequest(c, "Invalid request payload", err.Error())
		return
	}

	// Get authenticated user from context
	userID, exists := middleware.GetUserID(c)
	if !exists {
		h.logger.Error("User ID not found in context")
		response.Unauthorized(c, "Authentication required")
		return
	}

	userIDParsed, err := ulid.Parse(userID)
	if err != nil {
		h.logger.WithError(err).Error("Invalid user ID format")
		response.InternalServerError(c, "Authentication error")
		return
	}

	// Convert expiry option to timestamp
	var expiresAt *time.Time
	switch req.ExpiryOption {
	case "30days":
		t := time.Now().Add(30 * 24 * time.Hour)
		expiresAt = &t
	case "90days":
		t := time.Now().Add(90 * 24 * time.Hour)
		expiresAt = &t
	case "never":
		expiresAt = nil
	}

	// Create service request
	serviceReq := &auth.CreateAPIKeyRequest{
		Name:      req.Name,
		ProjectID: projectID,
		ExpiresAt: expiresAt,
	}

	// Create the API key
	apiKeyResp, err := h.apiKeyService.CreateAPIKey(c.Request.Context(), userIDParsed, serviceReq)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create API key")
		response.Error(c, err)
		return
	}

	// Convert to response format
	responseKey := APIKey{
		ID:         apiKeyResp.ID,
		Name:       apiKeyResp.Name,
		Key:        apiKeyResp.Key, // Only shown once
		KeyPreview: apiKeyResp.KeyPreview,
		ProjectID:  apiKeyResp.ProjectID,
		Status:     "active",
		CreatedAt:  apiKeyResp.CreatedAt,
		CreatedBy:  userID,
	}

	if apiKeyResp.ExpiresAt != nil {
		responseKey.ExpiresAt = *apiKeyResp.ExpiresAt
	}

	h.logger.WithFields(map[string]interface{}{
		"user_id":    userID,
		"api_key_id": apiKeyResp.ID,
		"project_id": projectID,
		"key_name":   req.Name,
	}).Info("API key created successfully")

	response.Created(c, responseKey)
}

// Delete handles DELETE /projects/:projectId/api-keys/:keyId
// @Summary Delete project-scoped API key
// @Description Permanently revoke and delete a project-scoped API key. This action cannot be undone and will immediately invalidate the key across all environments.
// @Tags API Keys
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID" example("proj_01234567890123456789012345")
// @Param keyId path string true "API Key ID" example("key_01234567890123456789012345")
// @Success 204 "Project-scoped API key deleted successfully"
// @Failure 400 {object} response.ErrorResponse "Bad request - invalid project ID or key ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - insufficient permissions to delete API keys"
// @Failure 404 {object} response.ErrorResponse "Project or API key not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/projects/{projectId}/api-keys/{keyId} [delete]
func (h *Handler) Delete(c *gin.Context) {
	// Get project ID from URL path
	projectID, err := ulid.Parse(c.Param("projectId"))
	if err != nil {
		h.logger.WithError(err).Error("Invalid project ID")
		response.BadRequest(c, "Invalid project ID", err.Error())
		return
	}

	// Get API key ID from URL path
	keyID, err := ulid.Parse(c.Param("keyId"))
	if err != nil {
		h.logger.WithError(err).Error("Invalid API key ID")
		response.BadRequest(c, "Invalid API key ID", err.Error())
		return
	}

	// Get authenticated user from context
	userID, exists := middleware.GetUserID(c)
	if !exists {
		h.logger.Error("User ID not found in context")
		response.Unauthorized(c, "Authentication required")
		return
	}

	// Get the API key to verify it exists and belongs to the environment
	apiKey, err := h.apiKeyService.GetAPIKey(c.Request.Context(), keyID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get API key for deletion")
		response.Error(c, err)
		return
	}

	// Verify the API key belongs to the specified project (since API keys are now project-scoped)
	if apiKey.ProjectID != projectID {
		h.logger.WithFields(map[string]interface{}{
			"api_key_id":     keyID,
			"project_id":     projectID,
			"key_project_id": apiKey.ProjectID,
		}).Warn("API key does not belong to specified project")
		response.NotFound(c, "API key not found in this project")
		return
	}

	// Revoke the API key
	if err := h.apiKeyService.RevokeAPIKey(c.Request.Context(), keyID); err != nil {
		h.logger.WithError(err).Error("Failed to delete API key")
		response.Error(c, err)
		return
	}

	h.logger.WithFields(map[string]interface{}{
		"user_id":    userID,
		"api_key_id": keyID,
		"project_id": projectID,
		"key_name":   apiKey.Name,
	}).Info("API key deleted successfully")

	response.NoContent(c)
}
