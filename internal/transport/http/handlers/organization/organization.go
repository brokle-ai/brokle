package organization

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"brokle/internal/config"
	"brokle/internal/core/domain/organization"
	"brokle/pkg/response"
)

// Handler handles organization endpoints
type Handler struct {
	config              *config.Config
	logger              *logrus.Logger
	organizationService organization.OrganizationService
	memberService       organization.MemberService
	projectService      organization.ProjectService
	environmentService  organization.EnvironmentService
	invitationService   organization.InvitationService
	Settings            *SettingsHandler // Embedded settings handler
}

// Request/Response Models

// Organization represents an organization entity
type Organization struct {
	ID          string    `json:"id" example:"org_1234567890" description:"Unique organization identifier"`
	Name        string    `json:"name" example:"Acme Corporation" description:"Organization name"`
	Slug        string    `json:"slug" example:"acme-corp" description:"URL-friendly organization identifier"`
	Description string    `json:"description,omitempty" example:"Leading AI solutions provider" description:"Optional organization description"`
	Plan        string    `json:"plan" example:"pro" description:"Subscription plan (free, pro, business, enterprise)"`
	Status      string    `json:"status" example:"active" description:"Organization status (active, suspended, deleted)"`
	CreatedAt   time.Time `json:"created_at" example:"2024-01-01T00:00:00Z" description:"Creation timestamp"`
	UpdatedAt   time.Time `json:"updated_at" example:"2024-01-01T00:00:00Z" description:"Last update timestamp"`
	OwnerID     string    `json:"owner_id" example:"usr_1234567890" description:"Organization owner user ID"`
}

// CreateOrganizationRequest represents the request to create an organization
type CreateOrganizationRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100" example:"Acme Corporation" description:"Organization name (2-100 characters)"`
	Slug        string `json:"slug,omitempty" binding:"omitempty,min=2,max=50" example:"acme-corp" description:"Optional URL-friendly identifier (auto-generated if not provided)"`
	Description string `json:"description,omitempty" binding:"omitempty,max=500" example:"Leading AI solutions provider" description:"Optional description (max 500 characters)"`
}

// UpdateOrganizationRequest represents the request to update an organization
type UpdateOrganizationRequest struct {
	Name        string `json:"name,omitempty" binding:"omitempty,min=2,max=100" example:"Acme Corporation" description:"Organization name (2-100 characters)"`
	Description string `json:"description,omitempty" binding:"omitempty,max=500" example:"Leading AI solutions provider" description:"Description (max 500 characters)"`
}

// OrganizationMember represents a member of an organization
type OrganizationMember struct {
	ID         string    `json:"id" example:"usr_1234567890" description:"User ID"`
	Email      string    `json:"email" example:"john@acme.com" description:"User email address"`
	FirstName  string    `json:"first_name" example:"John" description:"User first name"`
	LastName   string    `json:"last_name" example:"Doe" description:"User last name"`
	Role       string    `json:"role" example:"admin" description:"Role in organization (owner, admin, developer, viewer)"`
	Status     string    `json:"status" example:"active" description:"Member status (active, invited, suspended)"`
	JoinedAt   time.Time `json:"joined_at" example:"2024-01-01T00:00:00Z" description:"When user joined organization"`
	InvitedAt  time.Time `json:"invited_at,omitempty" example:"2024-01-01T00:00:00Z" description:"When user was invited (if not yet accepted)"`
	InvitedBy  string    `json:"invited_by,omitempty" example:"usr_0987654321" description:"ID of user who sent invitation"`
}

// InviteMemberRequest represents the request to invite a member to an organization
type InviteMemberRequest struct {
	Email string `json:"email" binding:"required,email" example:"john@acme.com" description:"Email address of user to invite"`
	Role  string `json:"role" binding:"required,oneof=admin developer viewer" example:"developer" description:"Role to assign (admin, developer, viewer)"`
}

// ListOrganizationsResponse represents the response when listing organizations
type ListOrganizationsResponse struct {
	Organizations []Organization `json:"organizations" description:"List of organizations"`
	Total         int            `json:"total" example:"5" description:"Total number of organizations"`
	Page          int            `json:"page" example:"1" description:"Current page number"`
	Limit         int            `json:"limit" example:"20" description:"Items per page"`
}

// ListMembersResponse represents the response when listing organization members
type ListMembersResponse struct {
	Members []OrganizationMember `json:"members" description:"List of organization members"`
	Total   int                  `json:"total" example:"10" description:"Total number of members"`
}

// NewHandler creates a new organization handler
func NewHandler(
	config *config.Config,
	logger *logrus.Logger,
	organizationService organization.OrganizationService,
	memberService organization.MemberService,
	projectService organization.ProjectService,
	environmentService organization.EnvironmentService,
	invitationService organization.InvitationService,
	settingsService organization.OrganizationSettingsService,
) *Handler {
	return &Handler{
		config:              config,
		logger:              logger,
		organizationService: organizationService,
		memberService:       memberService,
		projectService:      projectService,
		environmentService:  environmentService,
		invitationService:   invitationService,
		Settings:            NewSettingsHandler(config, logger, settingsService),
	}
}

// List handles GET /organizations
// @Summary List organizations
// @Description Get a paginated list of organizations for the authenticated user
// @Tags Organizations
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Items per page" default(20) minimum(1) maximum(100)
// @Param search query string false "Search organizations by name or slug"
// @Success 200 {object} response.SuccessResponse{data=ListOrganizationsResponse} "List of organizations"
// @Failure 400 {object} response.ErrorResponse "Bad request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/organizations [get]
func (h *Handler) List(c *gin.Context) {
	response.Success(c, gin.H{"message": "List organizations - TODO"})
}

// Create handles POST /organizations
// @Summary Create organization
// @Description Create a new organization. The authenticated user becomes the organization owner.
// @Tags Organizations
// @Accept json
// @Produce json
// @Param request body CreateOrganizationRequest true "Organization details"
// @Success 201 {object} response.SuccessResponse{data=Organization} "Organization created successfully"
// @Failure 400 {object} response.ErrorResponse "Bad request - invalid input or validation errors"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 409 {object} response.ErrorResponse "Conflict - organization slug already exists"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/organizations [post]
func (h *Handler) Create(c *gin.Context) {
	response.Success(c, gin.H{"message": "Create organization - TODO"})
}

// Get handles GET /organizations/:orgId
// @Summary Get organization details
// @Description Get detailed information about a specific organization
// @Tags Organizations
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID" example("org_1234567890")
// @Success 200 {object} response.SuccessResponse{data=Organization} "Organization details"
// @Failure 400 {object} response.ErrorResponse "Bad request - invalid organization ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - insufficient permissions"
// @Failure 404 {object} response.ErrorResponse "Organization not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/organizations/{orgId} [get]
func (h *Handler) Get(c *gin.Context) {
	response.Success(c, gin.H{"message": "Get organization - TODO"})
}

// Update handles PUT /organizations/:orgId
// @Summary Update organization
// @Description Update organization details. Only owners and admins can update organization settings.
// @Tags Organizations
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID" example("org_1234567890")
// @Param request body UpdateOrganizationRequest true "Updated organization details"
// @Success 200 {object} response.SuccessResponse{data=Organization} "Organization updated successfully"
// @Failure 400 {object} response.ErrorResponse "Bad request - invalid input or validation errors"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - insufficient permissions (requires owner or admin role)"
// @Failure 404 {object} response.ErrorResponse "Organization not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/organizations/{orgId} [put]
func (h *Handler) Update(c *gin.Context) {
	response.Success(c, gin.H{"message": "Update organization - TODO"})
}

// Delete handles DELETE /organizations/:orgId
// @Summary Delete organization
// @Description Permanently delete an organization. Only organization owners can delete organizations. This action cannot be undone.
// @Tags Organizations
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID" example("org_1234567890")
// @Success 204 "Organization deleted successfully"
// @Failure 400 {object} response.ErrorResponse "Bad request - invalid organization ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - only organization owners can delete organizations"
// @Failure 404 {object} response.ErrorResponse "Organization not found"
// @Failure 409 {object} response.ErrorResponse "Conflict - cannot delete organization with active projects or subscriptions"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/organizations/{orgId} [delete]
func (h *Handler) Delete(c *gin.Context) {
	response.Success(c, gin.H{"message": "Delete organization - TODO"})
}

// ListMembers handles GET /organizations/:orgId/members
// @Summary List organization members
// @Description Get a list of all members in an organization, including their roles and status
// @Tags Organizations
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID" example("org_1234567890")
// @Param status query string false "Filter by member status" Enums(active,invited,suspended)
// @Param role query string false "Filter by member role" Enums(owner,admin,developer,viewer)
// @Success 200 {object} response.SuccessResponse{data=ListMembersResponse} "List of organization members"
// @Failure 400 {object} response.ErrorResponse "Bad request - invalid organization ID or query parameters"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - insufficient permissions to view members"
// @Failure 404 {object} response.ErrorResponse "Organization not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/organizations/{orgId}/members [get]
func (h *Handler) ListMembers(c *gin.Context) {
	response.Success(c, gin.H{"message": "List organization members - TODO"})
}

// InviteMember handles POST /organizations/:orgId/members
// @Summary Invite member to organization
// @Description Send an invitation to join the organization. Only owners and admins can invite new members.
// @Tags Organizations
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID" example("org_1234567890")
// @Param request body InviteMemberRequest true "Member invitation details"
// @Success 201 {object} response.SuccessResponse{data=OrganizationMember} "Member invitation sent successfully"
// @Failure 400 {object} response.ErrorResponse "Bad request - invalid input or validation errors"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - insufficient permissions (requires owner or admin role)"
// @Failure 404 {object} response.ErrorResponse "Organization not found"
// @Failure 409 {object} response.ErrorResponse "Conflict - user is already a member or has pending invitation"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/organizations/{orgId}/members [post]
func (h *Handler) InviteMember(c *gin.Context) {
	response.Success(c, gin.H{"message": "Invite organization member - TODO"})
}

// RemoveMember handles DELETE /organizations/:orgId/members/:userId
// @Summary Remove member from organization
// @Description Remove a member from the organization or revoke their invitation. Owners and admins can remove members.
// @Tags Organizations
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID" example("org_1234567890")
// @Param userId path string true "User ID to remove" example("usr_1234567890")
// @Success 204 "Member removed successfully"
// @Failure 400 {object} response.ErrorResponse "Bad request - invalid organization ID or user ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - insufficient permissions or cannot remove organization owner"
// @Failure 404 {object} response.ErrorResponse "Organization or member not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/organizations/{orgId}/members/{userId} [delete]
func (h *Handler) RemoveMember(c *gin.Context) {
	response.Success(c, gin.H{"message": "Remove organization member - TODO"})
}

// Settings delegation methods
func (h *Handler) GetSettings(c *gin.Context)        { h.Settings.GetAllSettings(c) }
func (h *Handler) CreateSetting(c *gin.Context)      { h.Settings.CreateSetting(c) }
func (h *Handler) GetSetting(c *gin.Context)         { h.Settings.GetSetting(c) }
func (h *Handler) UpdateSetting(c *gin.Context)      { h.Settings.UpdateSetting(c) }
func (h *Handler) DeleteSetting(c *gin.Context)      { h.Settings.DeleteSetting(c) }
func (h *Handler) BulkCreateSettings(c *gin.Context) { h.Settings.BulkCreateSettings(c) }
func (h *Handler) ExportSettings(c *gin.Context)     { h.Settings.ExportSettings(c) }
func (h *Handler) ImportSettings(c *gin.Context)     { h.Settings.ImportSettings(c) }
func (h *Handler) ResetToDefaults(c *gin.Context)    { h.Settings.ResetToDefaults(c) }