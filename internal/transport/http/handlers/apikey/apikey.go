package apikey

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"brokle/internal/config"
	"brokle/pkg/response"
)

type Handler struct {
	config *config.Config
	logger *logrus.Logger
}

func NewHandler(config *config.Config, logger *logrus.Logger) *Handler {
	return &Handler{config: config, logger: logger}
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
func (h *Handler) List(c *gin.Context) { response.Success(c, gin.H{"message": "List API keys - TODO"}) }
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
func (h *Handler) Create(c *gin.Context) { response.Success(c, gin.H{"message": "Create API key - TODO"}) }
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
func (h *Handler) Delete(c *gin.Context) { response.Success(c, gin.H{"message": "Delete API key - TODO"}) }