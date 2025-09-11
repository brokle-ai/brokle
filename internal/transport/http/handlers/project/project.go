package project

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	
	"brokle/internal/config"
	"brokle/internal/core/domain/organization"
	"brokle/pkg/response"
	"brokle/pkg/ulid"
)

type Handler struct {
	config              *config.Config
	logger              *logrus.Logger
	projectService      organization.ProjectService
	organizationService organization.OrganizationService
	memberService       organization.MemberService
	environmentService  organization.EnvironmentService
}

func NewHandler(
	config *config.Config, 
	logger *logrus.Logger,
	projectService organization.ProjectService,
	organizationService organization.OrganizationService,
	memberService organization.MemberService,
	environmentService organization.EnvironmentService,
) *Handler {
	return &Handler{
		config:              config,
		logger:              logger,
		projectService:      projectService,
		organizationService: organizationService,
		memberService:       memberService,
		environmentService:  environmentService,
	}
}

// Request/Response Models

// ListRequest represents the request parameters for listing projects
type ListRequest struct {
	OrganizationID string `form:"organization_id" binding:"omitempty" example:"org_1234567890" description:"Filter by organization ID"`
	Status         string `form:"status" binding:"omitempty,oneof=active paused archived" example:"active" description:"Filter by project status"`
	Page           int    `form:"page" binding:"omitempty,min=1" example:"1" description:"Page number (default: 1)"`
	Limit          int    `form:"limit" binding:"omitempty,min=1,max=100" example:"20" description:"Items per page (default: 20, max: 100)"`
	Search         string `form:"search" binding:"omitempty" example:"chatbot" description:"Search projects by name or slug"`
}

// GetRequest represents the request parameters for getting a project
type GetRequest struct {
	ProjectID string `uri:"projectId" binding:"required" example:"proj_1234567890" description:"Project ID"`
}

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
// @Success 200 {object} response.APIResponse{data=[]Project,meta=response.Meta{pagination=response.Pagination}} "List of projects with pagination"
// @Failure 400 {object} response.ErrorResponse "Bad request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/projects [get]
func (h *Handler) List(c *gin.Context) {
	// Extract user ID from JWT
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.WithField("endpoint", "List").Error("User not authenticated")
		response.Unauthorized(c, "User not authenticated")
		return
	}
	
	userULID, ok := userID.(ulid.ULID)
	if !ok {
		h.logger.WithFields(logrus.Fields{
			"endpoint": "List",
			"user_id":  userID,
		}).Error("Invalid user ID type")
		response.BadRequest(c, "Invalid user ID", "")
		return
	}

	// Bind and validate query parameters
	var req ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.WithFields(logrus.Fields{
			"endpoint": "List",
			"user_id":  userULID.String(),
			"error":    err.Error(),
		}).Error("Invalid request parameters")
		response.BadRequest(c, "Invalid request parameters", "")
		return
	}

	// Set default pagination values
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}

	ctx := c.Request.Context()
	var projects []*organization.Project
	var total int

	// If organization_id is provided, filter by organization and validate access
	if req.OrganizationID != "" {
		orgULID, err := ulid.Parse(req.OrganizationID)
		if err != nil {
			h.logger.WithFields(logrus.Fields{
				"endpoint":        "List",
				"user_id":         userULID.String(),
				"organization_id": req.OrganizationID,
				"error":           err.Error(),
			}).Error("Invalid organization ID format")
			response.BadRequest(c, "Invalid organization ID", "")
			return
		}

		// Validate user is member of organization
		isMember, err := h.memberService.IsMember(ctx, userULID, orgULID)
		if err != nil {
			h.logger.WithFields(logrus.Fields{
				"endpoint":        "List",
				"user_id":         userULID.String(),
				"organization_id": req.OrganizationID,
				"error":           err.Error(),
			}).Error("Failed to check organization membership")
			response.InternalServerError(c, "Failed to verify organization access")
			return
		}

		if !isMember {
			h.logger.WithFields(logrus.Fields{
				"endpoint":        "List",
				"user_id":         userULID.String(),
				"organization_id": req.OrganizationID,
			}).Warn("User attempted to access organization projects without membership")
			response.Forbidden(c, "You don't have access to this organization")
			return
		}

		// Get projects for the specific organization
		orgProjects, err := h.projectService.GetProjectsByOrganization(ctx, orgULID)
		if err != nil {
			h.logger.WithFields(logrus.Fields{
				"endpoint":        "List",
				"user_id":         userULID.String(),
				"organization_id": req.OrganizationID,
				"error":           err.Error(),
			}).Error("Failed to get organization projects")
			response.InternalServerError(c, "Failed to retrieve projects")
			return
		}

		projects = orgProjects
		total = len(projects)
	} else {
		// Get projects from all user's organizations
		userOrgs, err := h.organizationService.GetUserOrganizations(ctx, userULID)
		if err != nil {
			h.logger.WithFields(logrus.Fields{
				"endpoint": "List",
				"user_id":  userULID.String(),
				"error":    err.Error(),
			}).Error("Failed to get user organizations")
			response.InternalServerError(c, "Failed to retrieve projects")
			return
		}

		// Collect projects from all organizations
		var allProjects []*organization.Project
		for _, org := range userOrgs {
			orgProjects, err := h.projectService.GetProjectsByOrganization(ctx, org.ID)
			if err != nil {
				h.logger.WithFields(logrus.Fields{
					"endpoint":        "List",
					"user_id":         userULID.String(),
					"organization_id": org.ID.String(),
					"error":           err.Error(),
				}).Error("Failed to get projects for organization")
				continue // Skip this organization but continue with others
			}
			allProjects = append(allProjects, orgProjects...)
		}

		projects = allProjects
		total = len(projects)
	}

	// Apply filtering
	var filteredProjects []*organization.Project
	for _, project := range projects {
		// Status filter
		if req.Status != "" && req.Status != "active" {
			// For now, all projects are considered "active" - extend this when status field is added
			continue
		}

		// Search filter
		if req.Search != "" {
			searchLower := strings.ToLower(req.Search)
			if !strings.Contains(strings.ToLower(project.Name), searchLower) &&
				!strings.Contains(strings.ToLower(project.Slug), searchLower) {
				continue
			}
		}

		filteredProjects = append(filteredProjects, project)
	}

	// Update total after filtering
	total = len(filteredProjects)

	// Apply pagination
	offset := (req.Page - 1) * req.Limit
	end := offset + req.Limit
	if offset > total {
		offset = total
	}
	if end > total {
		end = total
	}

	var paginatedProjects []*organization.Project
	if offset < total {
		paginatedProjects = filteredProjects[offset:end]
	}

	// Convert to response format
	responseProjects := make([]Project, len(paginatedProjects))
	for i, proj := range paginatedProjects {
		envCount, _ := h.environmentService.GetEnvironmentCount(ctx, proj.ID)
		responseProjects[i] = Project{
			ID:             proj.ID.String(),
			Name:           proj.Name,
			Slug:           proj.Slug,
			Description:    proj.Description,
			OrganizationID: proj.OrganizationID.String(),
			Status:         "active", // Default status - extend when status field is added
			CreatedAt:      proj.CreatedAt,
			UpdatedAt:      proj.UpdatedAt,
			OwnerID:        "", // TODO: Add owner concept when needed
			Environments:   envCount,
		}
	}

	// Create pagination
	pagination := response.NewPagination(req.Page, req.Limit, int64(total))

	response.SuccessWithPagination(c, responseProjects, pagination)

	h.logger.WithFields(logrus.Fields{
		"endpoint":        "List",
		"user_id":         userULID.String(),
		"organization_id": req.OrganizationID,
		"total_projects":  total,
		"returned":        len(responseProjects),
		"page":            req.Page,
		"limit":           req.Limit,
	}).Info("Projects listed successfully")
}
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
func (h *Handler) Create(c *gin.Context) {
	// Extract user ID from JWT
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.WithField("endpoint", "Create").Error("User not authenticated")
		response.Unauthorized(c, "User not authenticated")
		return
	}

	userULID, ok := userID.(ulid.ULID)
	if !ok {
		h.logger.WithFields(logrus.Fields{
			"endpoint": "Create",
			"user_id":  userID,
		}).Error("Invalid user ID type")
		response.BadRequest(c, "Invalid user ID", "")
		return
	}

	// Bind and validate request body
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithFields(logrus.Fields{
			"endpoint": "Create",
			"user_id":  userULID.String(),
			"error":    err.Error(),
		}).Error("Invalid request body")
		response.BadRequest(c, "Invalid request body", "")
		return
	}

	ctx := c.Request.Context()

	// Parse and validate organization ID
	orgULID, err := ulid.Parse(req.OrganizationID)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"endpoint":        "Create",
			"user_id":         userULID.String(),
			"organization_id": req.OrganizationID,
			"error":           err.Error(),
		}).Error("Invalid organization ID format")
		response.BadRequest(c, "Invalid organization ID", "")
		return
	}

	// Validate user is member of organization
	isMember, err := h.memberService.IsMember(ctx, userULID, orgULID)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"endpoint":        "Create",
			"user_id":         userULID.String(),
			"organization_id": req.OrganizationID,
			"error":           err.Error(),
		}).Error("Failed to check organization membership")
		response.InternalServerError(c, "Failed to verify organization access")
		return
	}

	if !isMember {
		h.logger.WithFields(logrus.Fields{
			"endpoint":        "Create",
			"user_id":         userULID.String(),
			"organization_id": req.OrganizationID,
		}).Warn("User attempted to create project in organization without membership")
		response.Forbidden(c, "You don't have permission to create projects in this organization")
		return
	}

	// Auto-generate slug if not provided
	slug := req.Slug
	if slug == "" {
		// Generate slug from name (simplified version)
		slug = strings.ToLower(strings.ReplaceAll(req.Name, " ", "-"))
		slug = strings.ReplaceAll(slug, "_", "-")
		// Remove special characters (basic implementation)
		validChars := "abcdefghijklmnopqrstuvwxyz0123456789-"
		var slugBuilder strings.Builder
		for _, char := range slug {
			if strings.ContainsRune(validChars, char) {
				slugBuilder.WriteRune(char)
			}
		}
		slug = slugBuilder.String()
		
		// Ensure slug is not empty and not too long
		if slug == "" {
			slug = "project"
		}
		if len(slug) > 50 {
			slug = slug[:50]
		}
	}

	// Create project via service
	createReq := &organization.CreateProjectRequest{
		Name:        req.Name,
		Slug:        slug,
		Description: req.Description,
	}

	project, err := h.projectService.CreateProject(ctx, orgULID, createReq)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			h.logger.WithFields(logrus.Fields{
				"endpoint":        "Create",
				"user_id":         userULID.String(),
				"organization_id": req.OrganizationID,
				"slug":            slug,
				"error":           err.Error(),
			}).Warn("Project slug already exists")
			response.Conflict(c, "Project with this slug already exists in organization")
			return
		}

		if strings.Contains(err.Error(), "not found") {
			h.logger.WithFields(logrus.Fields{
				"endpoint":        "Create",
				"user_id":         userULID.String(),
				"organization_id": req.OrganizationID,
				"error":           err.Error(),
			}).Error("Organization not found")
			response.NotFound(c, "Organization")
			return
		}

		h.logger.WithFields(logrus.Fields{
			"endpoint":        "Create",
			"user_id":         userULID.String(),
			"organization_id": req.OrganizationID,
			"project_name":    req.Name,
			"error":           err.Error(),
		}).Error("Failed to create project")
		response.InternalServerError(c, "Failed to create project")
		return
	}

	// Get environment count for response
	envCount, _ := h.environmentService.GetEnvironmentCount(ctx, project.ID)

	// Convert to response format
	responseProject := Project{
		ID:             project.ID.String(),
		Name:           project.Name,
		Slug:           project.Slug,
		Description:    project.Description,
		OrganizationID: project.OrganizationID.String(),
		Status:         "active", // Default status
		CreatedAt:      project.CreatedAt,
		UpdatedAt:      project.UpdatedAt,
		OwnerID:        userULID.String(), // Set creator as owner
		Environments:   envCount,
	}

	response.Created(c, responseProject)

	h.logger.WithFields(logrus.Fields{
		"endpoint":        "Create",
		"user_id":         userULID.String(),
		"organization_id": req.OrganizationID,
		"project_id":      project.ID.String(),
		"project_name":    project.Name,
		"project_slug":    project.Slug,
	}).Info("Project created successfully")
}
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
func (h *Handler) Get(c *gin.Context) {
	// Extract user ID from JWT
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.WithField("endpoint", "Get").Error("User not authenticated")
		response.Unauthorized(c, "User not authenticated")
		return
	}

	userULID, ok := userID.(ulid.ULID)
	if !ok {
		h.logger.WithFields(logrus.Fields{
			"endpoint": "Get",
			"user_id":  userID,
		}).Error("Invalid user ID type")
		response.BadRequest(c, "Invalid user ID", "")
		return
	}

	// Parse and validate project ID from path parameter
	projectIDStr := c.Param("projectId")
	if projectIDStr == "" {
		h.logger.WithFields(logrus.Fields{
			"endpoint": "Get",
			"user_id":  userULID.String(),
		}).Error("Project ID parameter missing")
		response.BadRequest(c, "Project ID is required", "")
		return
	}

	projectID, err := ulid.Parse(projectIDStr)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"endpoint":   "Get",
			"user_id":    userULID.String(),
			"project_id": projectIDStr,
			"error":      err.Error(),
		}).Error("Invalid project ID format")
		response.BadRequest(c, "Invalid project ID", "")
		return
	}

	ctx := c.Request.Context()

	// Validate user can access this project
	err = h.projectService.ValidateProjectAccess(ctx, userULID, projectID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.logger.WithFields(logrus.Fields{
				"endpoint":   "Get",
				"user_id":    userULID.String(),
				"project_id": projectIDStr,
				"error":      err.Error(),
			}).Warn("Project not found")
			response.NotFound(c, "Project")
			return
		}

		if strings.Contains(err.Error(), "access") {
			h.logger.WithFields(logrus.Fields{
				"endpoint":   "Get",
				"user_id":    userULID.String(),
				"project_id": projectIDStr,
				"error":      err.Error(),
			}).Warn("User attempted to access project without permission")
			response.Forbidden(c, "You don't have access to this project")
			return
		}

		h.logger.WithFields(logrus.Fields{
			"endpoint":   "Get",
			"user_id":    userULID.String(),
			"project_id": projectIDStr,
			"error":      err.Error(),
		}).Error("Failed to validate project access")
		response.InternalServerError(c, "Failed to validate project access")
		return
	}

	// Get project details
	project, err := h.projectService.GetProject(ctx, projectID)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"endpoint":   "Get",
			"user_id":    userULID.String(),
			"project_id": projectIDStr,
			"error":      err.Error(),
		}).Error("Failed to get project")
		response.InternalServerError(c, "Failed to retrieve project")
		return
	}

	// Get environment count for response
	envCount, _ := h.environmentService.GetEnvironmentCount(ctx, project.ID)

	// Convert to response format
	responseProject := Project{
		ID:             project.ID.String(),
		Name:           project.Name,
		Slug:           project.Slug,
		Description:    project.Description,
		OrganizationID: project.OrganizationID.String(),
		Status:         "active", // Default status
		CreatedAt:      project.CreatedAt,
		UpdatedAt:      project.UpdatedAt,
		OwnerID:        "", // TODO: Add owner concept when needed
		Environments:   envCount,
	}

	response.Success(c, responseProject)

	h.logger.WithFields(logrus.Fields{
		"endpoint":     "Get",
		"user_id":      userULID.String(),
		"project_id":   project.ID.String(),
		"project_name": project.Name,
		"environments": envCount,
	}).Info("Project retrieved successfully")
}
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
func (h *Handler) Update(c *gin.Context) {
	// Extract user ID from JWT
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.WithField("endpoint", "Update").Error("User not authenticated")
		response.Unauthorized(c, "User not authenticated")
		return
	}

	userULID, ok := userID.(ulid.ULID)
	if !ok {
		h.logger.WithFields(logrus.Fields{
			"endpoint": "Update",
			"user_id":  userID,
		}).Error("Invalid user ID type")
		response.BadRequest(c, "Invalid user ID", "")
		return
	}

	// Parse and validate project ID from path parameter
	projectIDStr := c.Param("projectId")
	if projectIDStr == "" {
		h.logger.WithFields(logrus.Fields{
			"endpoint": "Update",
			"user_id":  userULID.String(),
		}).Error("Project ID parameter missing")
		response.BadRequest(c, "Project ID is required", "")
		return
	}

	projectID, err := ulid.Parse(projectIDStr)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"endpoint":   "Update",
			"user_id":    userULID.String(),
			"project_id": projectIDStr,
			"error":      err.Error(),
		}).Error("Invalid project ID format")
		response.BadRequest(c, "Invalid project ID", "")
		return
	}

	// Bind and validate request body
	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithFields(logrus.Fields{
			"endpoint":   "Update",
			"user_id":    userULID.String(),
			"project_id": projectIDStr,
			"error":      err.Error(),
		}).Error("Invalid request body")
		response.BadRequest(c, "Invalid request body", "")
		return
	}

	ctx := c.Request.Context()

	// Validate user can access this project
	err = h.projectService.ValidateProjectAccess(ctx, userULID, projectID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.logger.WithFields(logrus.Fields{
				"endpoint":   "Update",
				"user_id":    userULID.String(),
				"project_id": projectIDStr,
				"error":      err.Error(),
			}).Warn("Project not found")
			response.NotFound(c, "Project")
			return
		}

		if strings.Contains(err.Error(), "access") {
			h.logger.WithFields(logrus.Fields{
				"endpoint":   "Update",
				"user_id":    userULID.String(),
				"project_id": projectIDStr,
				"error":      err.Error(),
			}).Warn("User attempted to update project without permission")
			response.Forbidden(c, "You don't have permission to update this project")
			return
		}

		h.logger.WithFields(logrus.Fields{
			"endpoint":   "Update",
			"user_id":    userULID.String(),
			"project_id": projectIDStr,
			"error":      err.Error(),
		}).Error("Failed to validate project access")
		response.InternalServerError(c, "Failed to validate project access")
		return
	}

	// Update project via service
	updateReq := &organization.UpdateProjectRequest{}
	if req.Name != "" {
		updateReq.Name = &req.Name
	}
	if req.Description != "" {
		updateReq.Description = &req.Description
	}

	err = h.projectService.UpdateProject(ctx, projectID, updateReq)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"endpoint":   "Update",
			"user_id":    userULID.String(),
			"project_id": projectIDStr,
			"error":      err.Error(),
		}).Error("Failed to update project")
		response.InternalServerError(c, "Failed to update project")
		return
	}

	// Get updated project details
	project, err := h.projectService.GetProject(ctx, projectID)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"endpoint":   "Update",
			"user_id":    userULID.String(),
			"project_id": projectIDStr,
			"error":      err.Error(),
		}).Error("Failed to get updated project")
		response.InternalServerError(c, "Failed to retrieve updated project")
		return
	}

	// Get environment count for response
	envCount, _ := h.environmentService.GetEnvironmentCount(ctx, project.ID)

	// Convert to response format
	responseProject := Project{
		ID:             project.ID.String(),
		Name:           project.Name,
		Slug:           project.Slug,
		Description:    project.Description,
		OrganizationID: project.OrganizationID.String(),
		Status:         req.Status,
		CreatedAt:      project.CreatedAt,
		UpdatedAt:      project.UpdatedAt,
		OwnerID:        "", // TODO: Add owner concept when needed
		Environments:   envCount,
	}

	// Handle status field from request if provided
	if req.Status != "" {
		responseProject.Status = req.Status
	} else {
		responseProject.Status = "active"
	}

	response.Success(c, responseProject)

	h.logger.WithFields(logrus.Fields{
		"endpoint":     "Update",
		"user_id":      userULID.String(),
		"project_id":   project.ID.String(),
		"project_name": project.Name,
		"changes":      fmt.Sprintf("name=%v, description=%v, status=%v", req.Name != "", req.Description != "", req.Status != ""),
	}).Info("Project updated successfully")
}
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
func (h *Handler) Delete(c *gin.Context) {
	// Extract user ID from JWT
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.WithField("endpoint", "Delete").Error("User not authenticated")
		response.Unauthorized(c, "User not authenticated")
		return
	}

	userULID, ok := userID.(ulid.ULID)
	if !ok {
		h.logger.WithFields(logrus.Fields{
			"endpoint": "Delete",
			"user_id":  userID,
		}).Error("Invalid user ID type")
		response.BadRequest(c, "Invalid user ID", "")
		return
	}

	// Parse and validate project ID from path parameter
	projectIDStr := c.Param("projectId")
	if projectIDStr == "" {
		h.logger.WithFields(logrus.Fields{
			"endpoint": "Delete",
			"user_id":  userULID.String(),
		}).Error("Project ID parameter missing")
		response.BadRequest(c, "Project ID is required", "")
		return
	}

	projectID, err := ulid.Parse(projectIDStr)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"endpoint":   "Delete",
			"user_id":    userULID.String(),
			"project_id": projectIDStr,
			"error":      err.Error(),
		}).Error("Invalid project ID format")
		response.BadRequest(c, "Invalid project ID", "")
		return
	}

	ctx := c.Request.Context()

	// Validate user can access this project
	err = h.projectService.ValidateProjectAccess(ctx, userULID, projectID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.logger.WithFields(logrus.Fields{
				"endpoint":   "Delete",
				"user_id":    userULID.String(),
				"project_id": projectIDStr,
				"error":      err.Error(),
			}).Warn("Project not found")
			response.NotFound(c, "Project")
			return
		}

		if strings.Contains(err.Error(), "access") {
			h.logger.WithFields(logrus.Fields{
				"endpoint":   "Delete",
				"user_id":    userULID.String(),
				"project_id": projectIDStr,
				"error":      err.Error(),
			}).Warn("User attempted to delete project without permission")
			response.Forbidden(c, "You don't have permission to delete this project")
			return
		}

		h.logger.WithFields(logrus.Fields{
			"endpoint":   "Delete",
			"user_id":    userULID.String(),
			"project_id": projectIDStr,
			"error":      err.Error(),
		}).Error("Failed to validate project access")
		response.InternalServerError(c, "Failed to validate project access")
		return
	}

	// Get project details before deletion for logging
	project, err := h.projectService.GetProject(ctx, projectID)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"endpoint":   "Delete",
			"user_id":    userULID.String(),
			"project_id": projectIDStr,
			"error":      err.Error(),
		}).Error("Failed to get project for deletion")
		response.InternalServerError(c, "Failed to get project")
		return
	}

	// TODO: Add additional validation for admin/owner permissions
	// For now, we allow any organization member to delete projects

	// TODO: Check if project has active environments or API usage
	// For now, we allow deletion regardless of environments

	// Delete project via service (soft delete)
	err = h.projectService.DeleteProject(ctx, projectID)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"endpoint":     "Delete",
			"user_id":      userULID.String(),
			"project_id":   projectIDStr,
			"project_name": project.Name,
			"error":        err.Error(),
		}).Error("Failed to delete project")
		response.InternalServerError(c, "Failed to delete project")
		return
	}

	response.NoContent(c)

	h.logger.WithFields(logrus.Fields{
		"endpoint":        "Delete",
		"user_id":         userULID.String(),
		"project_id":      project.ID.String(),
		"project_name":    project.Name,
		"organization_id": project.OrganizationID.String(),
	}).Info("Project deleted successfully")
}

// Environment represents an environment entity for responses
type Environment struct {
	ID        string    `json:"id" example:"env_1234567890" description:"Unique environment identifier"`
	Name      string    `json:"name" example:"Production" description:"Environment name"`
	Slug      string    `json:"slug" example:"prod" description:"URL-friendly environment identifier"`
	ProjectID string    `json:"project_id" example:"proj_1234567890" description:"Project ID this environment belongs to"`
	CreatedAt time.Time `json:"created_at" example:"2024-01-01T00:00:00Z" description:"Creation timestamp"`
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-01T00:00:00Z" description:"Last update timestamp"`
}

// ListEnvironmentsResponse represents the response when listing project environments
type ListEnvironmentsResponse struct {
	Environments []Environment `json:"environments" description:"List of environments"`
	ProjectID    string        `json:"project_id" example:"proj_1234567890" description:"Project ID"`
	ProjectName  string        `json:"project_name" example:"AI Chatbot" description:"Project name"`
}

// ListEnvironments handles GET /projects/:projectId/environments  
// @Summary List project environments
// @Description Get all environments that belong to a specific project
// @Tags Projects
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID" example("proj_1234567890")
// @Success 200 {object} response.SuccessResponse{data=ListEnvironmentsResponse} "List of project environments"
// @Failure 400 {object} response.ErrorResponse "Bad request - invalid project ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - insufficient permissions"
// @Failure 404 {object} response.ErrorResponse "Project not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/projects/{projectId}/environments [get]
func (h *Handler) ListEnvironments(c *gin.Context) {
	// Extract user ID from JWT
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.WithField("endpoint", "ListEnvironments").Error("User not authenticated")
		response.Unauthorized(c, "User not authenticated")
		return
	}

	userULID, ok := userID.(ulid.ULID)
	if !ok {
		h.logger.WithFields(logrus.Fields{
			"endpoint": "ListEnvironments",
			"user_id":  userID,
		}).Error("Invalid user ID type")
		response.BadRequest(c, "Invalid user ID", "")
		return
	}

	// Parse and validate project ID from path parameter
	projectIDStr := c.Param("projectId")
	if projectIDStr == "" {
		h.logger.WithFields(logrus.Fields{
			"endpoint": "ListEnvironments",
			"user_id":  userULID.String(),
		}).Error("Project ID parameter missing")
		response.BadRequest(c, "Project ID is required", "")
		return
	}

	projectID, err := ulid.Parse(projectIDStr)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"endpoint":   "ListEnvironments",
			"user_id":    userULID.String(),
			"project_id": projectIDStr,
			"error":      err.Error(),
		}).Error("Invalid project ID format")
		response.BadRequest(c, "Invalid project ID", "")
		return
	}

	ctx := c.Request.Context()

	// Validate user can access this project
	err = h.projectService.ValidateProjectAccess(ctx, userULID, projectID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.logger.WithFields(logrus.Fields{
				"endpoint":   "ListEnvironments",
				"user_id":    userULID.String(),
				"project_id": projectIDStr,
				"error":      err.Error(),
			}).Warn("Project not found")
			response.NotFound(c, "Project")
			return
		}

		if strings.Contains(err.Error(), "access") {
			h.logger.WithFields(logrus.Fields{
				"endpoint":   "ListEnvironments",
				"user_id":    userULID.String(),
				"project_id": projectIDStr,
				"error":      err.Error(),
			}).Warn("User attempted to access project environments without permission")
			response.Forbidden(c, "You don't have access to this project")
			return
		}

		h.logger.WithFields(logrus.Fields{
			"endpoint":   "ListEnvironments",
			"user_id":    userULID.String(),
			"project_id": projectIDStr,
			"error":      err.Error(),
		}).Error("Failed to validate project access")
		response.InternalServerError(c, "Failed to validate project access")
		return
	}

	// Get project details for response metadata
	project, err := h.projectService.GetProject(ctx, projectID)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"endpoint":   "ListEnvironments",
			"user_id":    userULID.String(),
			"project_id": projectIDStr,
			"error":      err.Error(),
		}).Error("Failed to get project")
		response.InternalServerError(c, "Failed to get project")
		return
	}

	// Get environments for the project
	environments, err := h.environmentService.GetEnvironmentsByProject(ctx, projectID)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"endpoint":     "ListEnvironments",
			"user_id":      userULID.String(),
			"project_id":   projectIDStr,
			"project_name": project.Name,
			"error":        err.Error(),
		}).Error("Failed to get project environments")
		response.InternalServerError(c, "Failed to retrieve environments")
		return
	}

	// Convert to response format
	responseEnvironments := make([]Environment, len(environments))
	for i, env := range environments {
		responseEnvironments[i] = Environment{
			ID:        env.ID.String(),
			Name:      env.Name,
			Slug:      env.Slug,
			ProjectID: env.ProjectID.String(),
			CreatedAt: env.CreatedAt,
			UpdatedAt: env.UpdatedAt,
		}
	}

	responseData := ListEnvironmentsResponse{
		Environments: responseEnvironments,
		ProjectID:    project.ID.String(),
		ProjectName:  project.Name,
	}

	response.Success(c, responseData)

	h.logger.WithFields(logrus.Fields{
		"endpoint":        "ListEnvironments",
		"user_id":         userULID.String(),
		"project_id":      project.ID.String(),
		"project_name":    project.Name,
		"total_envs":      len(environments),
		"organization_id": project.OrganizationID.String(),
	}).Info("Project environments listed successfully")
}