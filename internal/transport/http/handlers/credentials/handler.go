package credentials

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"brokle/internal/config"
	credentialsDomain "brokle/internal/core/domain/credentials"
	credentialsService "brokle/internal/core/services/credentials"
	"brokle/internal/transport/http/middleware"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/response"
	"brokle/pkg/ulid"
)

type Handler struct {
	config       *config.Config
	logger       *slog.Logger
	service      credentialsDomain.ProviderCredentialService
	modelCatalog credentialsService.ModelCatalogService
}

func NewHandler(
	cfg *config.Config,
	logger *slog.Logger,
	service credentialsDomain.ProviderCredentialService,
	modelCatalog credentialsService.ModelCatalogService,
) *Handler {
	return &Handler{
		config:       cfg,
		logger:       logger,
		service:      service,
		modelCatalog: modelCatalog,
	}
}

// serviceUnavailable checks if the credential service is configured.
// Returns true and sends error response if service is nil.
func (h *Handler) serviceUnavailable(c *gin.Context) bool {
	if h.service == nil {
		response.Error(c, appErrors.NewServiceUnavailableError(
			"Credentials feature not configured: AI_KEY_ENCRYPTION_KEY is required",
		))
		return true
	}
	return false
}

type CreateRequest struct {
	Name         string            `json:"name" binding:"required,min=1,max=100"`
	Adapter      string            `json:"adapter" binding:"required,oneof=openai anthropic azure gemini openrouter custom"`
	APIKey       string            `json:"api_key" binding:"required,min=10"`
	BaseURL      *string           `json:"base_url,omitempty"`
	Config       map[string]any    `json:"config,omitempty"`
	CustomModels []string          `json:"custom_models,omitempty"`
	Headers      map[string]string `json:"headers,omitempty"`
}

type UpdateRequest struct {
	Name         *string            `json:"name,omitempty" binding:"omitempty,min=1,max=100"`
	APIKey       *string            `json:"api_key,omitempty" binding:"omitempty,min=10"`
	BaseURL      *string            `json:"base_url,omitempty"`
	Config       map[string]any     `json:"config,omitempty"`
	CustomModels []string           `json:"custom_models,omitempty"`
	Headers      *map[string]string `json:"headers,omitempty"` // Pointer allows clearing with empty map
}

type TestConnectionRequest struct {
	Adapter string            `json:"adapter" binding:"required,oneof=openai anthropic azure gemini openrouter custom"`
	APIKey  string            `json:"api_key" binding:"required"`
	BaseURL *string           `json:"base_url,omitempty"`
	Config  map[string]any    `json:"config,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// Create creates a new AI provider credential.
// @Summary Create AI provider credential
// @Description Creates a new credential configuration. Each configuration has a unique name within the project.
// @Tags Credentials
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID"
// @Param request body CreateRequest true "Credential request"
// @Success 201 {object} credentials.ProviderCredentialResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse "Name already exists"
// @Failure 422 {object} response.ErrorResponse "API key validation failed"
// @Router /api/v1/projects/{projectId}/credentials/ai [post]
func (h *Handler) Create(c *gin.Context) {
	if h.serviceUnavailable(c) {
		return
	}

	projectID, err := ulid.Parse(c.Param("projectId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("Invalid project ID", "projectId must be a valid ULID"))
		return
	}

	var req CreateRequest
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
		ProjectID:    projectID,
		Name:         req.Name,
		Adapter:      credentialsDomain.Provider(req.Adapter),
		APIKey:       req.APIKey,
		BaseURL:      req.BaseURL,
		Config:       req.Config,
		CustomModels: req.CustomModels,
		Headers:      req.Headers,
		CreatedBy:    userIDPtr,
	}

	credential, err := h.service.Create(c.Request.Context(), domainReq)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, credential)
}

// Update updates an existing AI provider credential.
// @Summary Update AI provider credential
// @Description Updates an existing credential configuration by ID. Name can be changed if unique.
// @Tags Credentials
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID"
// @Param credentialId path string true "Credential ID"
// @Param request body UpdateRequest true "Update request"
// @Success 200 {object} credentials.ProviderCredentialResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse "Name already exists"
// @Failure 422 {object} response.ErrorResponse "API key validation failed"
// @Router /api/v1/projects/{projectId}/credentials/ai/{credentialId} [patch]
func (h *Handler) Update(c *gin.Context) {
	if h.serviceUnavailable(c) {
		return
	}

	projectID, err := ulid.Parse(c.Param("projectId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("Invalid project ID", "projectId must be a valid ULID"))
		return
	}

	credentialID, err := ulid.Parse(c.Param("credentialId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("Invalid credential ID", "credentialId must be a valid ULID"))
		return
	}

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	domainReq := &credentialsDomain.UpdateCredentialRequest{
		Name:         req.Name,
		APIKey:       req.APIKey,
		BaseURL:      req.BaseURL,
		Config:       req.Config,
		CustomModels: req.CustomModels,
		Headers:      req.Headers,
	}

	credential, err := h.service.Update(c.Request.Context(), credentialID, projectID, domainReq)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, credential)
}

// List lists all AI provider credentials for a project.
// @Summary List AI provider credentials
// @Description Returns all configured AI provider credentials for the project (with masked keys).
// @Tags Credentials
// @Produce json
// @Param projectId path string true "Project ID"
// @Success 200 {array} credentials.ProviderCredentialResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /api/v1/projects/{projectId}/credentials/ai [get]
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

// Get retrieves a specific AI provider credential by ID.
// @Summary Get AI provider credential
// @Description Returns the credential configuration for a specific credential ID (with masked key).
// @Tags Credentials
// @Produce json
// @Param projectId path string true "Project ID"
// @Param credentialId path string true "Credential ID"
// @Success 200 {object} credentials.ProviderCredentialResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/projects/{projectId}/credentials/ai/{credentialId} [get]
func (h *Handler) Get(c *gin.Context) {
	if h.serviceUnavailable(c) {
		return
	}

	projectID, err := ulid.Parse(c.Param("projectId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("Invalid project ID", "projectId must be a valid ULID"))
		return
	}

	credentialID, err := ulid.Parse(c.Param("credentialId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("Invalid credential ID", "credentialId must be a valid ULID"))
		return
	}

	credential, err := h.service.GetByID(c.Request.Context(), credentialID, projectID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, credential)
}

// Delete removes an AI provider credential by ID.
// @Summary Delete AI provider credential
// @Description Removes a credential configuration by its ID.
// @Tags Credentials
// @Produce json
// @Param projectId path string true "Project ID"
// @Param credentialId path string true "Credential ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/projects/{projectId}/credentials/ai/{credentialId} [delete]
func (h *Handler) Delete(c *gin.Context) {
	if h.serviceUnavailable(c) {
		return
	}

	projectID, err := ulid.Parse(c.Param("projectId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("Invalid project ID", "projectId must be a valid ULID"))
		return
	}

	credentialID, err := ulid.Parse(c.Param("credentialId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("Invalid credential ID", "credentialId must be a valid ULID"))
		return
	}

	if err := h.service.Delete(c.Request.Context(), credentialID, projectID); err != nil {
		response.Error(c, err)
		return
	}

	response.NoContent(c)
}

// TestConnection tests an AI provider connection without saving.
// @Summary Test AI provider connection
// @Description Tests the provided API key and configuration without saving. Returns success or error message.
// @Tags Credentials
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID"
// @Param request body TestConnectionRequest true "Connection test request"
// @Success 200 {object} credentials.TestConnectionResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /api/v1/projects/{projectId}/credentials/ai/test [post]
func (h *Handler) TestConnection(c *gin.Context) {
	if h.serviceUnavailable(c) {
		return
	}

	// We don't need projectId for testing, but keep it for route consistency
	_, err := ulid.Parse(c.Param("projectId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("Invalid project ID", "projectId must be a valid ULID"))
		return
	}

	var req TestConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	domainReq := &credentialsDomain.TestConnectionRequest{
		Adapter: credentialsDomain.Provider(req.Adapter),
		APIKey:  req.APIKey,
		BaseURL: req.BaseURL,
		Config:  req.Config,
		Headers: req.Headers,
	}

	result := h.service.TestConnection(c.Request.Context(), domainReq)
	response.Success(c, result)
}

// GetAvailableModels returns all available LLM models for a project.
// @Summary Get available models
// @Description Returns LLM models available based on configured AI providers. Standard providers return default models plus any custom models. Custom providers return only user-defined models.
// @Tags Credentials
// @Produce json
// @Param projectId path string true "Project ID"
// @Success 200 {array} analytics.AvailableModel
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /api/v1/projects/{projectId}/credentials/ai/models [get]
func (h *Handler) GetAvailableModels(c *gin.Context) {
	if h.modelCatalog == nil {
		response.Error(c, appErrors.NewServiceUnavailableError(
			"Model catalog feature not configured",
		))
		return
	}

	projectID, err := ulid.Parse(c.Param("projectId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("Invalid project ID", "projectId must be a valid ULID"))
		return
	}

	models, err := h.modelCatalog.GetAvailableModels(c.Request.Context(), projectID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, models)
}
