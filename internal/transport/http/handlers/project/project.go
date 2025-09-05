package project

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

// Project represents a project entity
type Project struct {
	ID             string    `json:"id" example:"proj_1234567890" description:"Unique project identifier"`
	Name           string    `json:"name" example:"AI Chatbot" description:"Project name"`
	Slug           string    `json:"slug" example:"ai-chatbot" description:"URL-friendly project identifier"`
	Description    string    `json:"description,omitempty" example:"Customer support AI chatbot" description:"Optional project description"`
	OrganizationID string    `json:"organization_id" example:"org_1234567890" description:"Organization ID this project belongs to"`
	Status         string    `json:"status" example:"active" description:"Project status (active, paused, archived)"`
	CreatedAt      time.Time `json:"created_at" example:"2024-01-01T00:00:00Z" description:"Creation timestamp"`
	UpdatedAt      time.Time `json:"updated_at" example:"2024-01-01T00:00:00Z" description:"Last update timestamp"`
	OwnerID        string    `json:"owner_id" example:"usr_1234567890" description:"Project owner user ID"`
	Environments   int       `json:"environments_count" example:"3" description:"Number of environments in this project"`
}

// CreateProjectRequest represents the request to create a project
type CreateProjectRequest struct {
	Name           string `json:"name" binding:"required,min=2,max=100" example:"AI Chatbot" description:"Project name (2-100 characters)"`
	Slug           string `json:"slug,omitempty" binding:"omitempty,min=2,max=50" example:"ai-chatbot" description:"Optional URL-friendly identifier (auto-generated if not provided)"`
	Description    string `json:"description,omitempty" binding:"omitempty,max=500" example:"Customer support AI chatbot" description:"Optional description (max 500 characters)"`
	OrganizationID string `json:"organization_id" binding:"required" example:"org_1234567890" description:"Organization ID this project belongs to"`
}

// UpdateProjectRequest represents the request to update a project
type UpdateProjectRequest struct {
	Name        string `json:"name,omitempty" binding:"omitempty,min=2,max=100" example:"AI Chatbot" description:"Project name (2-100 characters)"`
	Description string `json:"description,omitempty" binding:"omitempty,max=500" example:"Customer support AI chatbot" description:"Description (max 500 characters)"`
	Status      string `json:"status,omitempty" binding:"omitempty,oneof=active paused archived" example:"active" description:"Project status (active, paused, archived)"`
}

// ListProjectsResponse represents the response when listing projects
type ListProjectsResponse struct {
	Projects []Project `json:"projects" description:"List of projects"`
	Total    int       `json:"total" example:"15" description:"Total number of projects"`
	Page     int       `json:"page" example:"1" description:"Current page number"`
	Limit    int       `json:"limit" example:"20" description:"Items per page"`
}

// List handles GET /projects
// @Summary List projects
// @Description Get a paginated list of projects accessible to the authenticated user
// @Tags Projects
// @Accept json
// @Produce json
// @Param organization_id query string false "Filter by organization ID" example("org_1234567890")
// @Param status query string false "Filter by project status" Enums(active,paused,archived)
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Items per page" default(20) minimum(1) maximum(100)
// @Param search query string false "Search projects by name or slug"
// @Success 200 {object} response.SuccessResponse{data=ListProjectsResponse} "List of projects"
// @Failure 400 {object} response.ErrorResponse "Bad request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/projects [get]
func (h *Handler) List(c *gin.Context) { response.Success(c, gin.H{"message": "List projects - TODO"}) }
// Create handles POST /projects
// @Summary Create project
// @Description Create a new project within an organization. User must have appropriate permissions in the organization.
// @Tags Projects
// @Accept json
// @Produce json
// @Param request body CreateProjectRequest true "Project details"
// @Success 201 {object} response.SuccessResponse{data=Project} "Project created successfully"
// @Failure 400 {object} response.ErrorResponse "Bad request - invalid input or validation errors"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - insufficient permissions in organization"
// @Failure 404 {object} response.ErrorResponse "Organization not found"
// @Failure 409 {object} response.ErrorResponse "Conflict - project slug already exists in organization"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/projects [post]
func (h *Handler) Create(c *gin.Context) { response.Success(c, gin.H{"message": "Create project - TODO"}) }
// Get handles GET /projects/:projectId
// @Summary Get project details
// @Description Get detailed information about a specific project
// @Tags Projects
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID" example("proj_1234567890")
// @Success 200 {object} response.SuccessResponse{data=Project} "Project details"
// @Failure 400 {object} response.ErrorResponse "Bad request - invalid project ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - insufficient permissions"
// @Failure 404 {object} response.ErrorResponse "Project not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/projects/{projectId} [get]
func (h *Handler) Get(c *gin.Context) { response.Success(c, gin.H{"message": "Get project - TODO"}) }
// Update handles PUT /projects/:projectId
// @Summary Update project
// @Description Update project details. Requires appropriate permissions within the project organization.
// @Tags Projects
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID" example("proj_1234567890")
// @Param request body UpdateProjectRequest true "Updated project details"
// @Success 200 {object} response.SuccessResponse{data=Project} "Project updated successfully"
// @Failure 400 {object} response.ErrorResponse "Bad request - invalid input or validation errors"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - insufficient permissions"
// @Failure 404 {object} response.ErrorResponse "Project not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/projects/{projectId} [put]
func (h *Handler) Update(c *gin.Context) { response.Success(c, gin.H{"message": "Update project - TODO"}) }
// Delete handles DELETE /projects/:projectId
// @Summary Delete project
// @Description Permanently delete a project and all associated environments and data. This action cannot be undone.
// @Tags Projects
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID" example("proj_1234567890")
// @Success 204 "Project deleted successfully"
// @Failure 400 {object} response.ErrorResponse "Bad request - invalid project ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - insufficient permissions (requires admin or owner role)"
// @Failure 404 {object} response.ErrorResponse "Project not found"
// @Failure 409 {object} response.ErrorResponse "Conflict - cannot delete project with active environments or API usage"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/projects/{projectId} [delete]
func (h *Handler) Delete(c *gin.Context) { response.Success(c, gin.H{"message": "Delete project - TODO"}) }

// ListEnvironments handles GET /projects/:projectId/environments  
// @Summary List project environments
// @Description Get all environments that belong to a specific project
// @Tags Projects
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID" example("proj_1234567890")
// @Success 200 {object} response.SuccessResponse "List of project environments"
// @Failure 400 {object} response.ErrorResponse "Bad request - invalid project ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - insufficient permissions"
// @Failure 404 {object} response.ErrorResponse "Project not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/projects/{projectId}/environments [get]
func (h *Handler) ListEnvironments(c *gin.Context) { 
	response.Success(c, gin.H{"message": "List project environments - TODO"}) 
}