package apikey

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"brokle/internal/config"
	"brokle/internal/core/domain/auth"
	"brokle/internal/transport/http/middleware"
	"brokle/pkg/response"
	"brokle/pkg/ulid"
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

// APIKey represents an API key entity
type APIKey struct {
	ID            string    `json:"id" example:"key_1234567890" description:"Unique API key identifier"`
	Name          string    `json:"name" example:"Production API Key" description:"Human-readable name for the API key"`
	Key           string    `json:"key,omitempty" example:"bk_live_1234567890abcdef" description:"The actual API key (only shown on creation)"`
	KeyPreview    string    `json:"key_preview" example:"bk_live_...cdef" description:"Truncated version of the key for display"`
	EnvironmentID string    `json:"environment_id" example:"env_1234567890" description:"Environment ID this key belongs to"`
	Scopes        []string  `json:"scopes" example:"[\"read\", \"write\"]" description:"Permissions granted to this API key"`
	Status        string    `json:"status" example:"active" description:"API key status (active, inactive, revoked)"`
	LastUsed      time.Time `json:"last_used,omitempty" example:"2024-01-01T00:00:00Z" description:"Last time this key was used (null if never used)"`
	CreatedAt     time.Time `json:"created_at" example:"2024-01-01T00:00:00Z" description:"Creation timestamp"`
	ExpiresAt     time.Time `json:"expires_at,omitempty" example:"2024-12-31T23:59:59Z" description:"Expiration timestamp (null if never expires)"`
	CreatedBy     string    `json:"created_by" example:"usr_1234567890" description:"User ID who created this key"`
}

// CreateAPIKeyRequest represents the request to create an API key
type CreateAPIKeyRequest struct {
	Name          string    `json:"name" binding:"required,min=2,max=100" example:"Production API Key" description:"Human-readable name for the API key (2-100 characters)"`
	Scopes        []string  `json:"scopes" binding:"required,min=1" example:"[\"read\", \"write\"]" description:"Permissions to grant (read, write, admin)"`
	ExpiresAt     time.Time `json:"expires_at,omitempty" example:"2024-12-31T23:59:59Z" description:"Optional expiration date (null for no expiration)"`
	EnvironmentID string    `json:"environment_id" binding:"required" example:"env_1234567890" description:"Environment ID this key will belong to"`
}

// ListAPIKeysResponse represents the response when listing API keys
// NOTE: This struct is not used. When implementing, use response.SuccessWithPagination() 
// with []APIKey directly and response.NewPagination() for consistent pagination format.
type ListAPIKeysResponse struct {
	APIKeys []APIKey `json:"api_keys" description:"List of API keys"`
	// Pagination fields removed - use response.SuccessWithPagination() instead
}

// List handles GET /environments/:envId/api-keys
// @Summary List API keys
// @Description Get a paginated list of API keys for a specific environment
// @Tags API Keys
// @Accept json
// @Produce json
// @Param envId path string true "Environment ID" example("env_1234567890")
// @Param status query string false "Filter by API key status" Enums(active,inactive,revoked)
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Items per page" default(20) minimum(1) maximum(100)
// @Success 200 {object} response.APIResponse{data=[]APIKey,meta=response.Meta{pagination=response.Pagination}} "List of API keys with pagination"
// @Failure 400 {object} response.ErrorResponse "Bad request - invalid environment ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - insufficient permissions to view API keys"
// @Failure 404 {object} response.ErrorResponse "Environment not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/environments/{envId}/api-keys [get]
func (h *Handler) List(c *gin.Context) {
	// Get environment ID from URL path
	envID, err := ulid.Parse(c.Param("envId"))
	if err != nil {
		h.logger.WithError(err).Error("Invalid environment ID")
		response.BadRequest(c, "Invalid environment ID", err.Error())
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
		EnvironmentID: &envID,
		Limit:         limit,
		Offset:        (page - 1) * limit,
		SortBy:        "created_at",
		SortOrder:     "desc",
	}

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
			ID:            key.ID.String(),
			Name:          key.Name,
			KeyPreview:    key.KeyPrefix + "...",
			EnvironmentID: envID.String(),
			Scopes:        key.Scopes,
			Status:        getKeyStatus(*key),
			CreatedAt:     key.CreatedAt,
			CreatedBy:     key.UserID.String(),
		}

		if key.LastUsedAt != nil {
			responseKeys[i].LastUsed = *key.LastUsedAt
		}
		if key.ExpiresAt != nil {
			responseKeys[i].ExpiresAt = *key.ExpiresAt
		}
	}

	h.logger.WithFields(map[string]interface{}{
		"user_id":        userID,
		"environment_id": envID,
		"count":          len(responseKeys),
		"page":           page,
		"limit":          limit,
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
// Create handles POST /environments/:envId/api-keys
// @Summary Create API key
// @Description Create a new API key for an environment. The key will only be displayed once upon creation.
// @Tags API Keys
// @Accept json
// @Produce json
// @Param envId path string true "Environment ID" example("env_1234567890")
// @Param request body CreateAPIKeyRequest true "API key details"
// @Success 201 {object} response.SuccessResponse{data=APIKey} "API key created successfully (key only shown once)"
// @Failure 400 {object} response.ErrorResponse "Bad request - invalid input or validation errors"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - insufficient permissions to create API keys"
// @Failure 404 {object} response.ErrorResponse "Environment not found"
// @Failure 409 {object} response.ErrorResponse "Conflict - API key name already exists in environment"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/environments/{envId}/api-keys [post]
func (h *Handler) Create(c *gin.Context) {
	// Get environment ID from URL path
	envID, err := ulid.Parse(c.Param("envId"))
	if err != nil {
		h.logger.WithError(err).Error("Invalid environment ID")
		response.BadRequest(c, "Invalid environment ID", err.Error())
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

	// Create service request
	serviceReq := &auth.CreateAPIKeyRequest{
		Name:           req.Name,
		OrganizationID: envID, // TODO: Get actual organization ID from environment
		ProjectID:      nil,   // TODO: Get project ID from environment if needed
		EnvironmentID:  &envID,
		Scopes:         req.Scopes,
		RateLimitRPM:   1000, // Default rate limit
		ExpiresAt:      nil,  // No expiration by default
	}

	// Set expiration if provided
	if !req.ExpiresAt.IsZero() {
		serviceReq.ExpiresAt = &req.ExpiresAt
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
		ID:            apiKeyResp.ID.String(),
		Name:          apiKeyResp.Name,
		Key:           apiKeyResp.Key, // Only shown once
		KeyPreview:    apiKeyResp.KeyPrefix + "...",
		EnvironmentID: envID.String(),
		Scopes:        apiKeyResp.Scopes,
		Status:        "active",
		CreatedAt:     time.Now(),
		CreatedBy:     userID,
	}

	if apiKeyResp.ExpiresAt != nil {
		responseKey.ExpiresAt = *apiKeyResp.ExpiresAt
	}

	h.logger.WithFields(map[string]interface{}{
		"user_id":        userID,
		"api_key_id":     apiKeyResp.ID,
		"environment_id": envID,
		"key_name":       req.Name,
	}).Info("API key created successfully")

	response.Created(c, responseKey)
}
// Delete handles DELETE /environments/:envId/api-keys/:keyId
// @Summary Delete API key
// @Description Permanently revoke and delete an API key. This action cannot be undone and will immediately invalidate the key.
// @Tags API Keys
// @Accept json
// @Produce json
// @Param envId path string true "Environment ID" example("env_1234567890")
// @Param keyId path string true "API Key ID" example("key_1234567890")
// @Success 204 "API key deleted successfully"
// @Failure 400 {object} response.ErrorResponse "Bad request - invalid environment ID or key ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - insufficient permissions to delete API keys"
// @Failure 404 {object} response.ErrorResponse "Environment or API key not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/environments/{envId}/api-keys/{keyId} [delete]
func (h *Handler) Delete(c *gin.Context) {
	// Get environment ID from URL path
	envID, err := ulid.Parse(c.Param("envId"))
	if err != nil {
		h.logger.WithError(err).Error("Invalid environment ID")
		response.BadRequest(c, "Invalid environment ID", err.Error())
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

	// Verify the API key belongs to the specified environment
	if apiKey.EnvironmentID == nil || *apiKey.EnvironmentID != envID {
		h.logger.WithFields(map[string]interface{}{
			"api_key_id":     keyID,
			"environment_id": envID,
			"key_env_id":     apiKey.EnvironmentID,
		}).Warn("API key does not belong to specified environment")
		response.NotFound(c, "API key not found in this environment")
		return
	}

	// Revoke the API key
	if err := h.apiKeyService.RevokeAPIKey(c.Request.Context(), keyID); err != nil {
		h.logger.WithError(err).Error("Failed to delete API key")
		response.Error(c, err)
		return
	}

	h.logger.WithFields(map[string]interface{}{
		"user_id":        userID,
		"api_key_id":     keyID,
		"environment_id": envID,
		"key_name":       apiKey.Name,
	}).Info("API key deleted successfully")

	response.NoContent(c)
}