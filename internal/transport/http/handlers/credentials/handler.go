// Package credentials provides HTTP handlers for LLM provider credential management.
package credentials

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"brokle/internal/config"
	credentialsDomain "brokle/internal/core/domain/credentials"
	"brokle/internal/transport/http/middleware"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/response"
	"brokle/pkg/ulid"
)

// Handler contains all credential-related HTTP handlers.
type Handler struct {
	config  *config.Config
	logger  *slog.Logger
	service credentialsDomain.LLMProviderCredentialService
}

// NewHandler creates a new credentials handler.
func NewHandler(
	cfg *config.Config,
	logger *slog.Logger,
	service credentialsDomain.LLMProviderCredentialService,
) *Handler {
	return &Handler{
		config:  cfg,
		logger:  logger,
		service: service,
	}
}

// serviceUnavailable checks if the credential service is configured.
// Returns true and sends error response if service is nil.
func (h *Handler) serviceUnavailable(c *gin.Context) bool {
	if h.service == nil {
		response.Error(c, appErrors.NewServiceUnavailableError(
			"Credentials feature not configured: LLM_KEY_ENCRYPTION_KEY is required",
		))
		return true
	}
	return false
}

// CreateOrUpdateRequest represents the request body for creating/updating a credential.
type CreateOrUpdateRequest struct {
	Provider string  `json:"provider" binding:"required,oneof=openai anthropic"`
	APIKey   string  `json:"api_key" binding:"required,min=10"`
	BaseURL  *string `json:"base_url,omitempty"`
}

// CreateOrUpdate creates or updates an LLM provider credential.
// @Summary Create or update LLM provider credential
// @Description Creates a new credential or updates an existing one for the specified provider. The API key is validated before storing.
// @Tags Credentials
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID"
// @Param request body CreateOrUpdateRequest true "Credential request"
// @Success 200 {object} credentials.LLMProviderCredentialResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 422 {object} response.ErrorResponse "API key validation failed"
// @Router /api/v1/projects/{projectId}/credentials/llm [post]
func (h *Handler) CreateOrUpdate(c *gin.Context) {
	if h.serviceUnavailable(c) {
		return
	}

	projectID, err := ulid.Parse(c.Param("projectId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("Invalid project ID", "projectId must be a valid ULID"))
		return
	}

	var req CreateOrUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	userID, exists := middleware.GetUserIDULID(c)
	var userIDPtr *ulid.ULID
	if exists {
		userIDPtr = &userID
	}

	domainReq := &credentialsDomain.CreateCredentialRequest{
		ProjectID: projectID,
		Provider:  credentialsDomain.LLMProvider(req.Provider),
		APIKey:    req.APIKey,
		BaseURL:   req.BaseURL,
		CreatedBy: userIDPtr,
	}

	credential, err := h.service.CreateOrUpdate(c.Request.Context(), domainReq)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, credential)
}

// List lists all LLM provider credentials for a project.
// @Summary List LLM provider credentials
// @Description Returns all configured LLM provider credentials for the project (with masked keys).
// @Tags Credentials
// @Produce json
// @Param projectId path string true "Project ID"
// @Success 200 {array} credentials.LLMProviderCredentialResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /api/v1/projects/{projectId}/credentials/llm [get]
func (h *Handler) List(c *gin.Context) {
	if h.serviceUnavailable(c) {
		return
	}

	projectID, err := ulid.Parse(c.Param("projectId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("Invalid project ID", "projectId must be a valid ULID"))
		return
	}

	credentials, err := h.service.List(c.Request.Context(), projectID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, credentials)
}

// Get retrieves a specific LLM provider credential.
// @Summary Get LLM provider credential
// @Description Returns the credential configuration for a specific provider (with masked key).
// @Tags Credentials
// @Produce json
// @Param projectId path string true "Project ID"
// @Param provider path string true "Provider (openai or anthropic)"
// @Success 200 {object} credentials.LLMProviderCredentialResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/projects/{projectId}/credentials/llm/{provider} [get]
func (h *Handler) Get(c *gin.Context) {
	if h.serviceUnavailable(c) {
		return
	}

	projectID, err := ulid.Parse(c.Param("projectId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("Invalid project ID", "projectId must be a valid ULID"))
		return
	}

	provider := credentialsDomain.LLMProvider(c.Param("provider"))
	if !provider.IsValid() {
		response.Error(c, appErrors.NewValidationError("Invalid provider", "provider must be 'openai' or 'anthropic'"))
		return
	}

	credential, err := h.service.Get(c.Request.Context(), projectID, provider)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, credential)
}

// Delete removes an LLM provider credential.
// @Summary Delete LLM provider credential
// @Description Removes the credential for a specific provider from the project.
// @Tags Credentials
// @Produce json
// @Param projectId path string true "Project ID"
// @Param provider path string true "Provider (openai or anthropic)"
// @Success 204 "No Content"
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/projects/{projectId}/credentials/llm/{provider} [delete]
func (h *Handler) Delete(c *gin.Context) {
	if h.serviceUnavailable(c) {
		return
	}

	projectID, err := ulid.Parse(c.Param("projectId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("Invalid project ID", "projectId must be a valid ULID"))
		return
	}

	provider := credentialsDomain.LLMProvider(c.Param("provider"))
	if !provider.IsValid() {
		response.Error(c, appErrors.NewValidationError("Invalid provider", "provider must be 'openai' or 'anthropic'"))
		return
	}

	if err := h.service.Delete(c.Request.Context(), projectID, provider); err != nil {
		response.Error(c, err)
		return
	}

	response.NoContent(c)
}
