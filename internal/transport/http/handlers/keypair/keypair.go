package keypair

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"brokle/internal/config"
	"brokle/internal/core/domain/auth"
	"brokle/internal/transport/http/middleware"
	"brokle/pkg/response"
	"brokle/pkg/ulid"
)

// Handler handles key pair management endpoints
type Handler struct {
	config         *config.Config
	logger         *logrus.Logger
	keyPairService auth.KeyPairService
}

// NewHandler creates a new key pair handler
func NewHandler(config *config.Config, logger *logrus.Logger, keyPairService auth.KeyPairService) *Handler {
	return &Handler{
		config:         config,
		logger:         logger,
		keyPairService: keyPairService,
	}
}

// Request/Response Models

// CreateKeyPairRequest represents the request to create a key pair
// @Description Request to create a new public+secret key pair
type CreateKeyPairRequest struct {
	Name         string   `json:"name" binding:"required,min=2,max=100" example:"Production Key Pair" description:"Human-readable name for the key pair (2-100 characters)"`
	ProjectID    string   `json:"project_id" binding:"required" example:"01K4FHGHT3XX9WFM293QPZ5G9V" description:"Project ID this key pair will belong to"`
	EnvironmentID *string `json:"environment_id,omitempty" example:"01K4FHGHT3XX9WFM293QPZ5G9V" description:"Optional environment ID for environment-specific access"`
	Scopes       []string `json:"scopes" binding:"required,min=1" example:"[\"gateway:read\", \"analytics:read\"]" description:"Permissions to grant (gateway:read, gateway:write, analytics:read, config:read, config:write, admin)"`
	RateLimitRPM int      `json:"rate_limit_rpm" binding:"min=1,max=10000" example:"1000" description:"Rate limit in requests per minute"`
	ExpiresAt    *string  `json:"expires_at,omitempty" example:"2024-12-31T23:59:59Z" description:"Optional expiration date (ISO 8601 format)"`
}

// KeyPairResponse represents a key pair entity (without secret key)
// @Description Key pair information (secret key not included for security)
type KeyPairResponse struct {
	ID           string   `json:"id" example:"01K4FHGHT3XX9WFM293QPZ5G9V" description:"Unique key pair identifier"`
	Name         string   `json:"name" example:"Production Key Pair" description:"Human-readable name"`
	PublicKey    string   `json:"public_key" example:"pk_01K4FHGHT3XX9WFM293QPZ5G9V_abc123def456" description:"Public key (safe to display)"`
	ProjectID    string   `json:"project_id" example:"01K4FHGHT3XX9WFM293QPZ5G9V" description:"Project ID this key belongs to"`
	EnvironmentID *string `json:"environment_id,omitempty" example:"01K4FHGHT3XX9WFM293QPZ5G9V" description:"Environment ID if environment-specific"`
	Scopes       []string `json:"scopes" example:"[\"gateway:read\", \"analytics:read\"]" description:"Permissions granted to this key pair"`
	RateLimitRPM int      `json:"rate_limit_rpm" example:"1000" description:"Rate limit in requests per minute"`
	IsActive     bool     `json:"is_active" example:"true" description:"Whether the key pair is active"`
	LastUsedAt   *string  `json:"last_used_at,omitempty" example:"2024-01-01T00:00:00Z" description:"Last time this key was used (null if never used)"`
	CreatedAt    string   `json:"created_at" example:"2024-01-01T00:00:00Z" description:"Creation timestamp"`
	ExpiresAt    *string  `json:"expires_at,omitempty" example:"2024-12-31T23:59:59Z" description:"Expiration timestamp (null if never expires)"`
}

// CreateKeyPairResponse represents the response when creating a key pair
// @Description Response when creating a key pair (includes secret key - only shown once)
type CreateKeyPairResponse struct {
	ID           string   `json:"id" example:"01K4FHGHT3XX9WFM293QPZ5G9V" description:"Unique key pair identifier"`
	Name         string   `json:"name" example:"Production Key Pair" description:"Human-readable name"`
	PublicKey    string   `json:"public_key" example:"pk_01K4FHGHT3XX9WFM293QPZ5G9V_abc123def456" description:"Public key (safe to display)"`
	SecretKey    string   `json:"secret_key" example:"sk_xyz789uvw456rst123" description:"Secret key - STORE SECURELY, shown only once"`
	ProjectID    string   `json:"project_id" example:"01K4FHGHT3XX9WFM293QPZ5G9V" description:"Project ID"`
	EnvironmentID *string `json:"environment_id,omitempty" example:"01K4FHGHT3XX9WFM293QPZ5G9V" description:"Environment ID if specified"`
	Scopes       []string `json:"scopes" example:"[\"gateway:read\", \"analytics:read\"]" description:"Permissions granted"`
	RateLimitRPM int      `json:"rate_limit_rpm" example:"1000" description:"Rate limit in requests per minute"`
	CreatedAt    string   `json:"created_at" example:"2024-01-01T00:00:00Z" description:"Creation timestamp"`
	ExpiresAt    *string  `json:"expires_at,omitempty" example:"2024-12-31T23:59:59Z" description:"Expiration timestamp"`
}

// UpdateKeyPairRequest represents the request to update a key pair
// @Description Request to update an existing key pair
type UpdateKeyPairRequest struct {
	Name         *string  `json:"name,omitempty" example:"Updated Key Pair Name" description:"New name for the key pair"`
	Scopes       []string `json:"scopes,omitempty" example:"[\"gateway:read\", \"analytics:read\"]" description:"New permissions"`
	RateLimitRPM *int     `json:"rate_limit_rpm,omitempty" example:"2000" description:"New rate limit"`
	IsActive     *bool    `json:"is_active,omitempty" example:"false" description:"Activate or deactivate the key pair"`
	ExpiresAt    *string  `json:"expires_at,omitempty" example:"2025-12-31T23:59:59Z" description:"New expiration date"`
}

// ListKeyPairsResponse represents the response when listing key pairs
// @Description List of key pairs with pagination
type ListKeyPairsResponse struct {
	KeyPairs []KeyPairResponse `json:"key_pairs" description:"List of key pairs"`
	// Note: Pagination fields handled by response.SuccessWithPagination()
}

// List handles GET /projects/:projectId/key-pairs
// @Summary List key pairs
// @Description Get a paginated list of key pairs for a specific project
// @Tags Key Pairs
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID" example("01K4FHGHT3XX9WFM293QPZ5G9V")
// @Param environment_id query string false "Filter by environment ID"
// @Param is_active query bool false "Filter by active status"
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Items per page" default(20) minimum(1) maximum(100)
// @Success 200 {object} response.APIResponse{data=[]KeyPairResponse,meta=response.Meta{pagination=response.Pagination}} "List of key pairs with pagination"
// @Failure 400 {object} response.ErrorResponse "Bad request - invalid project ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - insufficient permissions to view key pairs"
// @Failure 404 {object} response.ErrorResponse "Project not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Security KeyPairAuth
// @Router /api/v1/projects/{projectId}/key-pairs [get]
func (h *Handler) List(c *gin.Context) {
	// Get authenticated user context
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	// Get project ID from URL
	projectIDStr := c.Param("projectId")
	projectID, err := ulid.Parse(projectIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid project ID", err.Error())
		return
	}

	// Get key pairs for the project
	keyPairs, err := h.keyPairService.GetKeyPairsByProject(c.Request.Context(), projectID)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"user_id":    authCtx.UserID,
			"project_id": projectID,
		}).Error("Failed to get key pairs")
		response.InternalServerError(c, "Failed to retrieve key pairs")
		return
	}

	// Convert to response format
	keyPairResponses := make([]KeyPairResponse, len(keyPairs))
	for i, kp := range keyPairs {
		keyPairResponses[i] = KeyPairResponse{
			ID:           kp.ID.String(),
			Name:         kp.Name,
			PublicKey:    kp.PublicKey,
			ProjectID:    kp.ProjectID.String(),
			EnvironmentID: stringPtr(kp.EnvironmentID),
			Scopes:       kp.Scopes,
			RateLimitRPM: kp.RateLimitRPM,
			IsActive:     kp.IsActive,
			LastUsedAt:   formatTimePtr(kp.LastUsedAt),
			CreatedAt:    kp.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			ExpiresAt:    formatTimePtr(kp.ExpiresAt),
		}
	}

	response.Success(c, keyPairResponses)
}

// Create handles POST /projects/:projectId/key-pairs
// @Summary Create key pair
// @Description Create a new public+secret key pair for a project. The secret key will only be displayed once upon creation.
// @Tags Key Pairs
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID" example("01K4FHGHT3XX9WFM293QPZ5G9V")
// @Param request body CreateKeyPairRequest true "Key pair details"
// @Success 201 {object} response.SuccessResponse{data=CreateKeyPairResponse} "Key pair created successfully (secret key only shown once)"
// @Failure 400 {object} response.ErrorResponse "Bad request - invalid input or validation errors"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - insufficient permissions to create key pairs"
// @Failure 404 {object} response.ErrorResponse "Project not found"
// @Failure 409 {object} response.ErrorResponse "Conflict - key pair name already exists in project"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Security KeyPairAuth
// @Router /api/v1/projects/{projectId}/key-pairs [post]
func (h *Handler) Create(c *gin.Context) {
	// Get authenticated user context
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	// Get project ID from URL
	projectIDStr := c.Param("projectId")
	projectID, err := ulid.Parse(projectIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid project ID", err.Error())
		return
	}

	// Parse request
	var req CreateKeyPairRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request payload", err.Error())
		return
	}

	// Validate that the project ID in the URL matches the request (if provided)
	if req.ProjectID != projectIDStr {
		response.BadRequest(c, "Project ID mismatch", "URL project ID must match request project ID")
		return
	}

	// Convert request to domain request
	domainReq := &auth.CreateKeyPairRequest{
		Name:           req.Name,
		OrganizationID: authCtx.UserID, // TODO: Get from auth context when organization context is available
		ProjectID:      projectID,
		Scopes:         req.Scopes,
		RateLimitRPM:   req.RateLimitRPM,
	}

	// Parse environment ID if provided
	if req.EnvironmentID != nil {
		envID, err := ulid.Parse(*req.EnvironmentID)
		if err != nil {
			response.BadRequest(c, "Invalid environment ID", err.Error())
			return
		}
		domainReq.EnvironmentID = &envID
	}

	// Parse expiration date if provided
	if req.ExpiresAt != nil {
		// TODO: Parse time from ISO 8601 string
		// For now, skip time parsing to avoid import overhead
	}

	// Create the key pair
	createResp, err := h.keyPairService.CreateKeyPair(c.Request.Context(), authCtx.UserID, domainReq)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"user_id":    authCtx.UserID,
			"project_id": projectID,
			"name":       req.Name,
		}).Error("Failed to create key pair")
		response.InternalServerError(c, "Failed to create key pair")
		return
	}

	// Convert to response format
	resp := CreateKeyPairResponse{
		ID:        createResp.ID.String(),
		Name:      createResp.Name,
		PublicKey: createResp.PublicKey,
		SecretKey: createResp.SecretKey, // Only shown once!
		ProjectID: createResp.ProjectID.String(),
		Scopes:    createResp.Scopes,
		CreatedAt: time.Now().Format(time.RFC3339),
		ExpiresAt: formatTimePtr(createResp.ExpiresAt),
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":      authCtx.UserID,
		"key_pair_id":  createResp.ID,
		"project_id":   projectID,
		"public_key":   createResp.PublicKey,
	}).Info("Key pair created successfully")

	c.JSON(http.StatusCreated, response.SuccessResponse{
		Success: true,
		Data:    resp,
	})
}

// GetByID handles GET /projects/:projectId/key-pairs/:keyPairId
// @Summary Get key pair by ID
// @Description Retrieve a specific key pair by its ID
// @Tags Key Pairs
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID" example("01HQZMQY8PFRJQH1TQZRBQGS5Q")
// @Param keyPairId path string true "Key Pair ID" example("01HQZMQY8PFRJQH1TQZRBQGS5R")
// @Success 200 {object} response.SuccessResponse{data=KeyPairResponse} "Key pair retrieved successfully"
// @Failure 400 {object} response.ErrorResponse "Bad request - invalid project ID or key pair ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - insufficient permissions to view key pair"
// @Failure 404 {object} response.ErrorResponse "Key pair not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Security KeyPairAuth
// @Router /api/v1/projects/{projectId}/key-pairs/{keyPairId} [get]
func (h *Handler) GetByID(c *gin.Context) {
	// Get authenticated user context
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	// Parse and validate project ID
	projectIDStr := c.Param("projectId")
	projectID, err := ulid.Parse(projectIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid project ID format", err.Error())
		return
	}

	// Parse and validate key pair ID
	keyPairIDStr := c.Param("keyPairId")
	keyPairID, err := ulid.Parse(keyPairIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid key pair ID format", err.Error())
		return
	}

	// Get key pair from service
	keyPair, err := h.keyPairService.GetKeyPair(c.Request.Context(), keyPairID)
	if err != nil {
		response.Error(c, err)
		return
	}

	// Verify the key pair belongs to the requested project
	if keyPair.ProjectID != projectID {
		response.NotFound(c, "Key pair not found in this project")
		return
	}

	// Check if user has access to this key pair
	if keyPair.UserID != authCtx.UserID {
		response.Forbidden(c, "Access denied to this key pair")
		return
	}

	// Convert to response format
	keyPairResp := &KeyPairResponse{
		ID:        keyPair.ID.String(),
		Name:      keyPair.Name,
		PublicKey: keyPair.PublicKey,
		ProjectID: keyPair.ProjectID.String(),
		Scopes:    keyPair.Scopes,
		IsActive:  keyPair.IsActive,
		LastUsedAt: formatTimePtr(keyPair.LastUsedAt),
		CreatedAt: keyPair.CreatedAt.Format(time.RFC3339),
		ExpiresAt: formatTimePtr(keyPair.ExpiresAt),
	}

	response.Success(c, keyPairResp)
}

// Delete handles DELETE /projects/:projectId/key-pairs/:keyPairId
// @Summary Delete key pair
// @Description Permanently revoke and delete a key pair. This action cannot be undone and will immediately invalidate the key.
// @Tags Key Pairs
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID" example("01K4FHGHT3XX9WFM293QPZ5G9V")
// @Param keyPairId path string true "Key Pair ID" example("01K4FHGHT3XX9WFM293QPZ5G9V")
// @Success 204 "Key pair deleted successfully"
// @Failure 400 {object} response.ErrorResponse "Bad request - invalid project ID or key pair ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - insufficient permissions to delete key pairs"
// @Failure 404 {object} response.ErrorResponse "Project or key pair not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Security KeyPairAuth
// @Router /api/v1/projects/{projectId}/key-pairs/{keyPairId} [delete]
func (h *Handler) Delete(c *gin.Context) {
	// Get authenticated user context
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	// Get project ID and key pair ID from URL
	projectIDStr := c.Param("projectId")
	keyPairIDStr := c.Param("keyPairId")

	projectID, err := ulid.Parse(projectIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid project ID", err.Error())
		return
	}

	keyPairID, err := ulid.Parse(keyPairIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid key pair ID", err.Error())
		return
	}

	// Revoke the key pair
	err = h.keyPairService.RevokeKeyPair(c.Request.Context(), keyPairID)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"user_id":      authCtx.UserID,
			"project_id":   projectID,
			"key_pair_id":  keyPairID,
		}).Error("Failed to delete key pair")
		response.InternalServerError(c, "Failed to delete key pair")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":      authCtx.UserID,
		"project_id":   projectID,
		"key_pair_id":  keyPairID,
	}).Info("Key pair deleted successfully")

	c.Status(http.StatusNoContent)
}

// Update handles PATCH /projects/:projectId/key-pairs/:keyPairId
// @Summary Update key pair
// @Description Update an existing key pair's properties (name, scopes, rate limit, active status, expiration)
// @Tags Key Pairs
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID" example("01K4FHGHT3XX9WFM293QPZ5G9V")
// @Param keyPairId path string true "Key Pair ID" example("01K4FHGHT3XX9WFM293QPZ5G9V")
// @Param request body UpdateKeyPairRequest true "Key pair updates"
// @Success 200 {object} response.SuccessResponse{data=KeyPairResponse} "Key pair updated successfully"
// @Failure 400 {object} response.ErrorResponse "Bad request - invalid input"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - insufficient permissions"
// @Failure 404 {object} response.ErrorResponse "Key pair not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Security KeyPairAuth
// @Router /api/v1/projects/{projectId}/key-pairs/{keyPairId} [patch]
func (h *Handler) Update(c *gin.Context) {
	// Get authenticated user context
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	// Get IDs from URL
	keyPairIDStr := c.Param("keyPairId")
	keyPairID, err := ulid.Parse(keyPairIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid key pair ID", err.Error())
		return
	}

	// Parse request
	var req UpdateKeyPairRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request payload", err.Error())
		return
	}

	// Convert to domain request
	domainReq := &auth.UpdateKeyPairRequest{
		Name:         req.Name,
		Scopes:       req.Scopes,
		RateLimitRPM: req.RateLimitRPM,
		IsActive:     req.IsActive,
	}

	// Update the key pair
	err = h.keyPairService.UpdateKeyPair(c.Request.Context(), keyPairID, domainReq)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"user_id":     authCtx.UserID,
			"key_pair_id": keyPairID,
		}).Error("Failed to update key pair")
		response.InternalServerError(c, "Failed to update key pair")
		return
	}

	// Get updated key pair for response
	updatedKeyPair, err := h.keyPairService.GetKeyPair(c.Request.Context(), keyPairID)
	if err != nil {
		h.logger.WithError(err).WithField("key_pair_id", keyPairID).Error("Failed to get updated key pair")
		response.InternalServerError(c, "Key pair updated but failed to retrieve updated data")
		return
	}

	// Convert to response format
	resp := KeyPairResponse{
		ID:           updatedKeyPair.ID.String(),
		Name:         updatedKeyPair.Name,
		PublicKey:    updatedKeyPair.PublicKey,
		ProjectID:    updatedKeyPair.ProjectID.String(),
		EnvironmentID: stringPtr(updatedKeyPair.EnvironmentID),
		Scopes:       updatedKeyPair.Scopes,
		RateLimitRPM: updatedKeyPair.RateLimitRPM,
		IsActive:     updatedKeyPair.IsActive,
		LastUsedAt:   formatTimePtr(updatedKeyPair.LastUsedAt),
		CreatedAt:    updatedKeyPair.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		ExpiresAt:    formatTimePtr(updatedKeyPair.ExpiresAt),
	}

	response.Success(c, resp)
}

// Helper functions

func stringPtr(id *ulid.ULID) *string {
	if id == nil {
		return nil
	}
	s := id.String()
	return &s
}

func timePtr(t *string) *string {
	if t == nil {
		return nil
	}
	return t
}

// formatTimePtr formats a *time.Time to *string in RFC3339 format
func formatTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	formatted := t.Format(time.RFC3339)
	return &formatted
}