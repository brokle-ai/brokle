package environment

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

// Environment represents an environment entity
type Environment struct {
	ID          string    `json:"id" example:"env_1234567890" description:"Unique environment identifier"`
	Name        string    `json:"name" example:"Production" description:"Environment name"`
	Slug        string    `json:"slug" example:"production" description:"URL-friendly environment identifier"`
	Description string    `json:"description,omitempty" example:"Production environment for live traffic" description:"Optional environment description"`
	ProjectID   string    `json:"project_id" example:"proj_1234567890" description:"Project ID this environment belongs to"`
	Type        string    `json:"type" example:"production" description:"Environment type (development, staging, production)"`
	Status      string    `json:"status" example:"active" description:"Environment status (active, paused, archived)"`
	CreatedAt   time.Time `json:"created_at" example:"2024-01-01T00:00:00Z" description:"Creation timestamp"`
	UpdatedAt   time.Time `json:"updated_at" example:"2024-01-01T00:00:00Z" description:"Last update timestamp"`
	APIKeys     int       `json:"api_keys_count" example:"2" description:"Number of API keys in this environment"`
}

// CreateEnvironmentRequest represents the request to create an environment
type CreateEnvironmentRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100" example:"Production" description:"Environment name (2-100 characters)"`
	Slug        string `json:"slug,omitempty" binding:"omitempty,min=2,max=50" example:"production" description:"Optional URL-friendly identifier (auto-generated if not provided)"`
	Description string `json:"description,omitempty" binding:"omitempty,max=500" example:"Production environment for live traffic" description:"Optional description (max 500 characters)"`
	ProjectID   string `json:"project_id" binding:"required" example:"proj_1234567890" description:"Project ID this environment belongs to"`
	Type        string `json:"type" binding:"required,oneof=development staging production" example:"production" description:"Environment type (development, staging, production)"`
}

// UpdateEnvironmentRequest represents the request to update an environment
type UpdateEnvironmentRequest struct {
	Name        string `json:"name,omitempty" binding:"omitempty,min=2,max=100" example:"Production" description:"Environment name (2-100 characters)"`
	Description string `json:"description,omitempty" binding:"omitempty,max=500" example:"Production environment for live traffic" description:"Description (max 500 characters)"`
	Status      string `json:"status,omitempty" binding:"omitempty,oneof=active paused archived" example:"active" description:"Environment status (active, paused, archived)"`
}

// ListEnvironmentsResponse represents the response when listing environments
type ListEnvironmentsResponse struct {
	Environments []Environment `json:"environments" description:"List of environments"`
	Total        int           `json:"total" example:"3" description:"Total number of environments"`
	Page         int           `json:"page" example:"1" description:"Current page number"`
	Limit        int           `json:"limit" example:"20" description:"Items per page"`
}

// List handles GET /environments
// @Summary List environments
// @Description Get a paginated list of environments accessible to the authenticated user
// @Tags Environments
// @Accept json
// @Produce json
// @Param project_id query string false "Filter by project ID" example("proj_1234567890")
// @Param type query string false "Filter by environment type" Enums(development,staging,production)
// @Param status query string false "Filter by environment status" Enums(active,paused,archived)
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Items per page" default(20) minimum(1) maximum(100)
// @Success 200 {object} response.SuccessResponse{data=ListEnvironmentsResponse} "List of environments"
// @Failure 400 {object} response.ErrorResponse "Bad request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/environments [get]
func (h *Handler) List(c *gin.Context) { response.Success(c, gin.H{"message": "List environments - TODO"}) }
// Create handles POST /environments
// @Summary Create environment
// @Description Create a new environment within a project. User must have appropriate permissions in the project.
// @Tags Environments
// @Accept json
// @Produce json
// @Param request body CreateEnvironmentRequest true "Environment details"
// @Success 201 {object} response.SuccessResponse{data=Environment} "Environment created successfully"
// @Failure 400 {object} response.ErrorResponse "Bad request - invalid input or validation errors"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - insufficient permissions in project"
// @Failure 404 {object} response.ErrorResponse "Project not found"
// @Failure 409 {object} response.ErrorResponse "Conflict - environment slug already exists in project"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/environments [post]
func (h *Handler) Create(c *gin.Context) { response.Success(c, gin.H{"message": "Create environment - TODO"}) }
// Get handles GET /environments/:envId
// @Summary Get environment details
// @Description Get detailed information about a specific environment
// @Tags Environments
// @Accept json
// @Produce json
// @Param envId path string true "Environment ID" example("env_1234567890")
// @Success 200 {object} response.SuccessResponse{data=Environment} "Environment details"
// @Failure 400 {object} response.ErrorResponse "Bad request - invalid environment ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - insufficient permissions"
// @Failure 404 {object} response.ErrorResponse "Environment not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/environments/{envId} [get]
func (h *Handler) Get(c *gin.Context) { response.Success(c, gin.H{"message": "Get environment - TODO"}) }
// Update handles PUT /environments/:envId
// @Summary Update environment
// @Description Update environment details. Requires appropriate permissions within the project.
// @Tags Environments
// @Accept json
// @Produce json
// @Param envId path string true "Environment ID" example("env_1234567890")
// @Param request body UpdateEnvironmentRequest true "Updated environment details"
// @Success 200 {object} response.SuccessResponse{data=Environment} "Environment updated successfully"
// @Failure 400 {object} response.ErrorResponse "Bad request - invalid input or validation errors"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - insufficient permissions"
// @Failure 404 {object} response.ErrorResponse "Environment not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/environments/{envId} [put]
func (h *Handler) Update(c *gin.Context) { response.Success(c, gin.H{"message": "Update environment - TODO"}) }
// Delete handles DELETE /environments/:envId
// @Summary Delete environment
// @Description Permanently delete an environment and all associated API keys and data. This action cannot be undone.
// @Tags Environments
// @Accept json
// @Produce json
// @Param envId path string true "Environment ID" example("env_1234567890")
// @Success 204 "Environment deleted successfully"
// @Failure 400 {object} response.ErrorResponse "Bad request - invalid environment ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - insufficient permissions (requires admin or owner role)"
// @Failure 404 {object} response.ErrorResponse "Environment not found"
// @Failure 409 {object} response.ErrorResponse "Conflict - cannot delete environment with active API keys or recent usage"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/environments/{envId} [delete]
func (h *Handler) Delete(c *gin.Context) { response.Success(c, gin.H{"message": "Delete environment - TODO"}) }