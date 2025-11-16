package organization

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"brokle/internal/config"
	"brokle/internal/core/domain/auth"
	"brokle/internal/core/domain/organization"
	"brokle/internal/core/domain/user"
	"brokle/pkg/response"
	"brokle/pkg/ulid"
)

// Handler handles organization endpoints
type Handler struct {
	config              *config.Config
	logger              *logrus.Logger
	organizationService organization.OrganizationService
	memberService       organization.MemberService
	projectService      organization.ProjectService
	invitationService   organization.InvitationService
	userService         user.UserService
	roleService         auth.RoleService
	Settings            *SettingsHandler // Embedded settings handler
}

// Request/Response Models

// Organization represents an organization entity
type Organization struct {
	CreatedAt   time.Time `json:"created_at" example:"2024-01-01T00:00:00Z" description:"Creation timestamp"`
	UpdatedAt   time.Time `json:"updated_at" example:"2024-01-01T00:00:00Z" description:"Last update timestamp"`
	ID          string    `json:"id" example:"org_1234567890" description:"Unique organization identifier"`
	Name        string    `json:"name" example:"Acme Corporation" description:"Organization name"`
	Description string    `json:"description,omitempty" example:"Leading AI solutions provider" description:"Optional organization description"`
	Plan        string    `json:"plan" example:"pro" description:"Subscription plan (free, pro, business, enterprise)"`
	Status      string    `json:"status" example:"active" description:"Organization status (active, suspended, deleted)"`
}

// CreateOrganizationRequest represents the request to create an organization
type CreateOrganizationRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100" example:"Acme Corporation" description:"Organization name (2-100 characters)"`
	Description string `json:"description,omitempty" binding:"omitempty,max=500" example:"Leading AI solutions provider" description:"Optional description (max 500 characters)"`
}

// UpdateOrganizationRequest represents the request to update an organization
type UpdateOrganizationRequest struct {
	Name        string `json:"name,omitempty" binding:"omitempty,min=2,max=100" example:"Acme Corporation" description:"Organization name (2-100 characters)"`
	Description string `json:"description,omitempty" binding:"omitempty,max=500" example:"Leading AI solutions provider" description:"Description (max 500 characters)"`
}

// OrganizationMember represents a member of an organization
type OrganizationMember struct {
	JoinedAt  time.Time `json:"joined_at" example:"2024-01-01T00:00:00Z" description:"When user joined organization"`
	CreatedAt time.Time `json:"created_at" example:"2024-01-01T00:00:00Z" description:"When membership was created"`
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-01T00:00:00Z" description:"When membership was last updated"`
	InvitedBy *string   `json:"invited_by,omitempty" example:"john@inviter.com" description:"Email of user who sent invitation"`
	UserID    string    `json:"user_id" example:"usr_1234567890" description:"User ID"`
	Email     string    `json:"email" example:"john@acme.com" description:"User email address"`
	FirstName string    `json:"first_name" example:"John" description:"User first name"`
	LastName  string    `json:"last_name" example:"Doe" description:"User last name"`
	Role      string    `json:"role" example:"admin" description:"Role name in organization"`
	Status    string    `json:"status" example:"active" description:"Member status (active, invited, suspended)"`
}

// ListMembersRequest represents request parameters for listing members
type ListMembersRequest struct {
	OrgID  string `uri:"orgId" binding:"required" example:"01FXYZ123456789ABCDEFGHIJK0" description:"Organization ID"`
	Status string `form:"status" example:"active" description:"Filter by member status"`
	Role   string `form:"role" example:"admin" description:"Filter by member role"`
}

// InviteMemberRequest represents the request to invite a member to an organization
type InviteMemberRequest struct {
	OrgID string `uri:"orgId" binding:"required" example:"01FXYZ123456789ABCDEFGHIJK0" description:"Organization ID"`
	Email string `json:"email" binding:"required,email" example:"john@acme.com" description:"Email address of user to invite"`
	Role  string `json:"role" binding:"required,oneof=admin developer viewer" example:"developer" description:"Role to assign (admin, developer, viewer)"`
}

// RemoveMemberRequest represents the request to remove a member from an organization
type RemoveMemberRequest struct {
	OrgID  string `uri:"orgId" binding:"required" example:"01FXYZ123456789ABCDEFGHIJK0" description:"Organization ID"`
	UserID string `uri:"userId" binding:"required" example:"01FXYZ123456789ABCDEFGHIJK0" description:"User ID to remove"`
}

// ListRequest represents request parameters for listing organizations
type ListRequest struct {
	Search string `form:"search" example:"acme" description:"Search organizations by name or slug"`
	Page   int    `form:"page,default=1" binding:"min=1" example:"1" description:"Page number"`
	Limit  int    `form:"limit,default=20" binding:"min=1,max=100" example:"20" description:"Items per page"`
}

// GetRequest represents request parameters for getting an organization
type GetRequest struct {
	OrgID string `uri:"orgId" binding:"required" example:"01FXYZ123456789ABCDEFGHIJK0" description:"Organization ID"`
}

// NewHandler creates a new organization handler
func NewHandler(
	config *config.Config,
	logger *logrus.Logger,
	organizationService organization.OrganizationService,
	memberService organization.MemberService,
	projectService organization.ProjectService,
	invitationService organization.InvitationService,
	settingsService organization.OrganizationSettingsService,
	userService user.UserService,
	roleService auth.RoleService,
) *Handler {
	return &Handler{
		config:              config,
		logger:              logger,
		organizationService: organizationService,
		memberService:       memberService,
		projectService:      projectService,
		invitationService:   invitationService,
		userService:         userService,
		roleService:         roleService,
		Settings:            NewSettingsHandler(config, logger, settingsService),
	}
}

// List handles GET /organizations
// @Summary List organizations
// @Description Get a paginated list of organizations for the authenticated user
// @Tags Organizations
// @Accept json
// @Produce json
// @Param cursor query string false "Pagination cursor" example("eyJjcmVhdGVkX2F0IjoiMjAyNC0wMS0wMVQxMjowMDowMFoiLCJpZCI6IjAxSDJYM1k0WjUifQ==")
// @Param page_size query int false "Items per page" Enums(10,20,30,40,50) default(50)
// @Param sort_by query string false "Sort field" Enums(created_at,name) default("created_at")
// @Param sort_dir query string false "Sort direction" Enums(asc,desc) default("desc")
// @Param search query string false "Search organizations by name or slug"
// @Success 200 {object} response.APIResponse{data=[]Organization,meta=response.Meta{pagination=response.Pagination}} "List of organizations with cursor pagination"
// @Failure 400 {object} response.ErrorResponse "Bad request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/organizations [get]
func (h *Handler) List(c *gin.Context) {
	var req ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.WithError(err).Error("Invalid list organizations request")
		response.BadRequest(c, "Invalid request parameters", err.Error())
		return
	}

	// Get user ID from context
	userIDValue, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		response.Unauthorized(c, "Authentication required")
		return
	}

	userID, ok := userIDValue.(ulid.ULID)
	if !ok {
		h.logger.Error("Invalid user ID type in context")
		response.InternalServerError(c, "Internal error")
		return
	}

	// Get user organizations
	organizations, err := h.organizationService.GetUserOrganizations(c.Request.Context(), userID)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to get user organizations")
		response.InternalServerError(c, "Failed to retrieve organizations")
		return
	}

	// Convert to response format and apply filtering/pagination
	var filteredOrgs []Organization
	for _, org := range organizations {
		// Apply search filter if provided
		if req.Search != "" {
			if !strings.Contains(strings.ToLower(org.Name), strings.ToLower(req.Search)) {
				continue
			}
		}

		filteredOrgs = append(filteredOrgs, Organization{
			ID:        org.ID.String(),
			Name:      org.Name,
			Plan:      org.Plan,
			Status:    org.SubscriptionStatus,
			CreatedAt: org.CreatedAt,
			UpdatedAt: org.UpdatedAt,
		})
	}

	// Parse offset pagination parameters
	params := response.ParsePaginationParams(
		c.Query("page"),
		c.Query("limit"),
		c.Query("sort_by"),
		c.Query("sort_dir"),
	)

	total := len(filteredOrgs)

	// Sort organizations for stable ordering
	sort.Slice(filteredOrgs, func(i, j int) bool {
		if params.SortDir == "asc" {
			return filteredOrgs[i].CreatedAt.Before(filteredOrgs[j].CreatedAt)
		}
		return filteredOrgs[i].CreatedAt.After(filteredOrgs[j].CreatedAt)
	})

	// Apply offset pagination
	offset := params.GetOffset()
	limit := params.Limit

	// Calculate end index for slicing
	end := offset + limit
	if end > len(filteredOrgs) {
		end = len(filteredOrgs)
	}

	// Apply pagination slice
	if offset < len(filteredOrgs) {
		filteredOrgs = filteredOrgs[offset:end]
	} else {
		filteredOrgs = []Organization{}
	}

	// Create offset pagination
	pag := response.NewPagination(params.Page, params.Limit, int64(total))

	h.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"count":   len(filteredOrgs),
		"total":   total,
	}).Info("Organizations listed successfully")

	response.SuccessWithPagination(c, filteredOrgs, pag)
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
	var req CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid create organization request")
		response.BadRequest(c, "Invalid request payload", err.Error())
		return
	}

	// Get user ID from context
	userIDValue, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		response.Unauthorized(c, "Authentication required")
		return
	}

	userID, ok := userIDValue.(ulid.ULID)
	if !ok {
		h.logger.Error("Invalid user ID type in context")
		response.InternalServerError(c, "Internal error")
		return
	}

	// Create organization request
	createReq := &organization.CreateOrganizationRequest{
		Name:         req.Name,
		BillingEmail: "", // Will be set from user email or provided
	}

	// Create organization
	org, err := h.organizationService.CreateOrganization(c.Request.Context(), userID, createReq)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"user_id": userID,
			"name":    req.Name,
		}).Error("Failed to create organization")

		// Handle specific errors
		if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "duplicate") {
			response.Conflict(c, "Organization name already exists")
			return
		}

		response.InternalServerError(c, "Failed to create organization")
		return
	}

	// Convert to response format
	responseData := Organization{
		ID:        org.ID.String(),
		Name:      org.Name,
		Plan:      org.Plan,
		Status:    org.SubscriptionStatus,
		CreatedAt: org.CreatedAt,
		UpdatedAt: org.UpdatedAt,
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":         userID,
		"organization_id": org.ID,
		"name":            org.Name,
	}).Info("Organization created successfully")

	c.Header("Location", fmt.Sprintf("/api/v1/organizations/%s", org.ID))
	response.Success(c, responseData)
	c.Status(201) // Created
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
	var req GetRequest
	if err := c.ShouldBindUri(&req); err != nil {
		h.logger.WithError(err).Error("Invalid get organization request")
		response.BadRequest(c, "Invalid organization ID", err.Error())
		return
	}

	// Get user ID from context
	userIDValue, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		response.Unauthorized(c, "Authentication required")
		return
	}

	userID, ok := userIDValue.(ulid.ULID)
	if !ok {
		h.logger.Error("Invalid user ID type in context")
		response.InternalServerError(c, "Internal error")
		return
	}

	// Parse organization ID
	orgID, err := ulid.Parse(req.OrgID)
	if err != nil {
		h.logger.WithError(err).WithField("org_id", req.OrgID).Error("Invalid organization ID format")
		response.BadRequest(c, "Invalid organization ID format", err.Error())
		return
	}

	// Check if user can access this organization
	canAccess, err := h.memberService.CanUserAccessOrganization(c.Request.Context(), userID, orgID)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"user_id": userID,
			"org_id":  orgID,
		}).Error("Failed to check organization access")
		response.InternalServerError(c, "Failed to check organization access")
		return
	}

	if !canAccess {
		h.logger.WithFields(logrus.Fields{
			"user_id": userID,
			"org_id":  orgID,
		}).Warn("User attempted to access organization without permission")
		response.Forbidden(c, "Insufficient permissions to access this organization")
		return
	}

	// Get organization
	org, err := h.organizationService.GetOrganization(c.Request.Context(), orgID)
	if err != nil {
		h.logger.WithError(err).WithField("org_id", orgID).Error("Failed to get organization")
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "Organization")
			return
		}
		response.InternalServerError(c, "Failed to retrieve organization")
		return
	}

	// Convert to response format
	responseData := Organization{
		ID:        org.ID.String(),
		Name:      org.Name,
		Plan:      org.Plan,
		Status:    org.SubscriptionStatus,
		CreatedAt: org.CreatedAt,
		UpdatedAt: org.UpdatedAt,
	}

	h.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"org_id":  orgID,
	}).Info("Organization retrieved successfully")

	response.Success(c, responseData)
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
	var uriReq GetRequest
	if err := c.ShouldBindUri(&uriReq); err != nil {
		h.logger.WithError(err).Error("Invalid update organization URI request")
		response.BadRequest(c, "Invalid organization ID", err.Error())
		return
	}

	var req UpdateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid update organization request")
		response.BadRequest(c, "Invalid request payload", err.Error())
		return
	}

	// Get user ID from context
	userIDValue, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		response.Unauthorized(c, "Authentication required")
		return
	}

	userID, ok := userIDValue.(ulid.ULID)
	if !ok {
		h.logger.Error("Invalid user ID type in context")
		response.InternalServerError(c, "Internal error")
		return
	}

	// Parse organization ID
	orgID, err := ulid.Parse(uriReq.OrgID)
	if err != nil {
		h.logger.WithError(err).WithField("org_id", uriReq.OrgID).Error("Invalid organization ID format")
		response.BadRequest(c, "Invalid organization ID format", err.Error())
		return
	}

	// Check if user can access this organization
	canAccess, err := h.memberService.CanUserAccessOrganization(c.Request.Context(), userID, orgID)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"user_id": userID,
			"org_id":  orgID,
		}).Error("Failed to check organization access")
		response.InternalServerError(c, "Failed to check organization access")
		return
	}

	if !canAccess {
		h.logger.WithFields(logrus.Fields{
			"user_id": userID,
			"org_id":  orgID,
		}).Warn("User attempted to update organization without permission")
		response.Forbidden(c, "Insufficient permissions to update this organization")
		return
	}

	// Create update request
	updateReq := &organization.UpdateOrganizationRequest{
		Name:         &req.Name,
		BillingEmail: &req.Description, // Using description field temporarily
	}

	// Update organization
	err = h.organizationService.UpdateOrganization(c.Request.Context(), orgID, updateReq)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"user_id": userID,
			"org_id":  orgID,
		}).Error("Failed to update organization")
		response.InternalServerError(c, "Failed to update organization")
		return
	}

	// Get updated organization
	org, err := h.organizationService.GetOrganization(c.Request.Context(), orgID)
	if err != nil {
		h.logger.WithError(err).WithField("org_id", orgID).Error("Failed to get updated organization")
		response.InternalServerError(c, "Failed to retrieve updated organization")
		return
	}

	// Convert to response format
	responseData := Organization{
		ID:        org.ID.String(),
		Name:      org.Name,
		Plan:      org.Plan,
		Status:    org.SubscriptionStatus,
		CreatedAt: org.CreatedAt,
		UpdatedAt: org.UpdatedAt,
	}

	h.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"org_id":  orgID,
	}).Info("Organization updated successfully")

	response.Success(c, responseData)
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
	var req GetRequest
	if err := c.ShouldBindUri(&req); err != nil {
		h.logger.WithError(err).Error("Invalid delete organization request")
		response.BadRequest(c, "Invalid organization ID", err.Error())
		return
	}

	// Get user ID from context
	userIDValue, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		response.Unauthorized(c, "Authentication required")
		return
	}

	userID, ok := userIDValue.(ulid.ULID)
	if !ok {
		h.logger.Error("Invalid user ID type in context")
		response.InternalServerError(c, "Internal error")
		return
	}

	// Parse organization ID
	orgID, err := ulid.Parse(req.OrgID)
	if err != nil {
		h.logger.WithError(err).WithField("org_id", req.OrgID).Error("Invalid organization ID format")
		response.BadRequest(c, "Invalid organization ID format", err.Error())
		return
	}

	// Check if user can access this organization
	canAccess, err := h.memberService.CanUserAccessOrganization(c.Request.Context(), userID, orgID)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"user_id": userID,
			"org_id":  orgID,
		}).Error("Failed to check organization access")
		response.InternalServerError(c, "Failed to check organization access")
		return
	}

	if !canAccess {
		h.logger.WithFields(logrus.Fields{
			"user_id": userID,
			"org_id":  orgID,
		}).Warn("User attempted to delete organization without permission")
		response.Forbidden(c, "Insufficient permissions to delete this organization")
		return
	}

	// Delete organization
	err = h.organizationService.DeleteOrganization(c.Request.Context(), orgID)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"user_id": userID,
			"org_id":  orgID,
		}).Error("Failed to delete organization")

		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "Organization")
			return
		}

		response.InternalServerError(c, "Failed to delete organization")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"org_id":  orgID,
	}).Info("Organization deleted successfully")

	c.Status(204) // No Content
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
// @Success 200 {object} response.APIResponse{data=[]OrganizationMember,meta=response.Meta{pagination=response.Pagination}} "List of organization members with cursor pagination"
// @Failure 400 {object} response.ErrorResponse "Bad request - invalid organization ID or query parameters"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - insufficient permissions to view members"
// @Failure 404 {object} response.ErrorResponse "Organization not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/organizations/{orgId}/members [get]
func (h *Handler) ListMembers(c *gin.Context) {
	var req ListMembersRequest
	if err := c.ShouldBindUri(&req); err != nil {
		h.logger.WithError(err).Error("Invalid list members URI request")
		response.BadRequest(c, "Invalid organization ID", err.Error())
		return
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.WithError(err).Error("Invalid list members query request")
		response.BadRequest(c, "Invalid query parameters", err.Error())
		return
	}

	// Get user ID from context
	userIDValue, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		response.Unauthorized(c, "Authentication required")
		return
	}

	userID, ok := userIDValue.(ulid.ULID)
	if !ok {
		h.logger.Error("Invalid user ID type in context")
		response.InternalServerError(c, "Internal error")
		return
	}

	// Parse organization ID
	orgID, err := ulid.Parse(req.OrgID)
	if err != nil {
		h.logger.WithError(err).WithField("org_id", req.OrgID).Error("Invalid organization ID format")
		response.BadRequest(c, "Invalid organization ID format", err.Error())
		return
	}

	// Check if user can access this organization
	canAccess, err := h.memberService.CanUserAccessOrganization(c.Request.Context(), userID, orgID)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"user_id": userID,
			"org_id":  orgID,
		}).Error("Failed to check organization access")
		response.InternalServerError(c, "Failed to check organization access")
		return
	}

	if !canAccess {
		h.logger.WithFields(logrus.Fields{
			"user_id": userID,
			"org_id":  orgID,
		}).Warn("User attempted to list members without permission")
		response.Forbidden(c, "Insufficient permissions to view organization members")
		return
	}

	// Get organization members
	members, err := h.memberService.GetMembers(c.Request.Context(), orgID)
	if err != nil {
		h.logger.WithError(err).WithField("org_id", orgID).Error("Failed to get organization members")
		response.InternalServerError(c, "Failed to retrieve organization members")
		return
	}

	// Convert to response format with user and role resolution
	var memberList []OrganizationMember
	for _, member := range members {
		// Resolve user details - this should always succeed if data integrity is maintained
		userDetails, err := h.userService.GetUser(c.Request.Context(), member.UserID)
		if err != nil {
			h.logger.WithError(err).WithFields(logrus.Fields{
				"user_id":         member.UserID,
				"organization_id": orgID,
				"member_status":   member.Status,
			}).Error("CRITICAL: Member references non-existent user - data integrity violation")
			response.InternalServerError(c, "Data integrity error - please contact support")
			return
		}

		// Resolve role details - this should always succeed if data integrity is maintained
		role, err := h.roleService.GetRoleByID(c.Request.Context(), member.RoleID)
		if err != nil {
			h.logger.WithError(err).WithFields(logrus.Fields{
				"role_id":         member.RoleID,
				"organization_id": orgID,
				"user_id":         member.UserID,
			}).Error("CRITICAL: Member references non-existent role - data integrity violation")
			response.InternalServerError(c, "Data integrity error - please contact support")
			return
		}

		// Resolve inviter details if present
		var invitedByEmail *string
		if member.InvitedBy != nil {
			inviter, err := h.userService.GetUser(c.Request.Context(), *member.InvitedBy)
			if err != nil {
				h.logger.WithError(err).WithField("inviter_id", *member.InvitedBy).Warn("Failed to get inviter details - continuing without inviter info")
				// For inviter, we can continue without this data as it's not critical
			} else {
				invitedByEmail = &inviter.Email
			}
		}

		memberList = append(memberList, OrganizationMember{
			UserID:    member.UserID.String(),
			Email:     userDetails.Email,
			FirstName: userDetails.FirstName,
			LastName:  userDetails.LastName,
			Role:      role.Name,
			Status:    member.Status,
			JoinedAt:  member.JoinedAt,
			InvitedBy: invitedByEmail,
			CreatedAt: member.CreatedAt,
			UpdatedAt: member.UpdatedAt,
		})
	}

	// Parse offset pagination parameters
	params := response.ParsePaginationParams(
		c.Query("page"),
		c.Query("limit"),
		c.Query("sort_by"),
		c.Query("sort_dir"),
	)

	total := len(memberList)

	// Sort members for stable ordering
	sort.Slice(memberList, func(i, j int) bool {
		if params.SortDir == "asc" {
			return memberList[i].CreatedAt.Before(memberList[j].CreatedAt)
		}
		return memberList[i].CreatedAt.After(memberList[j].CreatedAt)
	})

	// Apply offset pagination
	offset := params.GetOffset()
	limit := params.Limit

	// Calculate end index for slicing
	end := offset + limit
	if end > len(memberList) {
		end = len(memberList)
	}

	// Apply pagination slice
	if offset < len(memberList) {
		memberList = memberList[offset:end]
	} else {
		memberList = []OrganizationMember{}
	}

	// Create offset pagination
	pag := response.NewPagination(params.Page, params.Limit, int64(total))

	h.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"org_id":  orgID,
		"count":   len(memberList),
	}).Info("Organization members listed successfully")

	response.SuccessWithPagination(c, memberList, pag)
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
	var uriReq InviteMemberRequest
	if err := c.ShouldBindUri(&uriReq); err != nil {
		h.logger.WithError(err).Error("Invalid invite member URI request")
		response.BadRequest(c, "Invalid organization ID", err.Error())
		return
	}

	var req InviteMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid invite member request")
		response.BadRequest(c, "Invalid request payload", err.Error())
		return
	}

	// Copy org ID from URI
	req.OrgID = uriReq.OrgID

	// Get user ID from context
	userIDValue, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		response.Unauthorized(c, "Authentication required")
		return
	}

	userID, ok := userIDValue.(ulid.ULID)
	if !ok {
		h.logger.Error("Invalid user ID type in context")
		response.InternalServerError(c, "Internal error")
		return
	}

	// Parse organization ID
	orgID, err := ulid.Parse(req.OrgID)
	if err != nil {
		h.logger.WithError(err).WithField("org_id", req.OrgID).Error("Invalid organization ID format")
		response.BadRequest(c, "Invalid organization ID format", err.Error())
		return
	}

	// Check if user can access this organization
	canAccess, err := h.memberService.CanUserAccessOrganization(c.Request.Context(), userID, orgID)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"user_id": userID,
			"org_id":  orgID,
		}).Error("Failed to check organization access")
		response.InternalServerError(c, "Failed to check organization access")
		return
	}

	if !canAccess {
		h.logger.WithFields(logrus.Fields{
			"user_id": userID,
			"org_id":  orgID,
		}).Warn("User attempted to invite member without permission")
		response.Forbidden(c, "Insufficient permissions to invite members to this organization")
		return
	}

	// For now, we'll create a basic invitation request
	// TODO: Need to get actual role ID based on role name
	roleID := ulid.New() // This should be resolved from role name

	inviteReq := &organization.InviteUserRequest{
		Email:  req.Email,
		RoleID: roleID,
	}

	// Send invitation
	invitation, err := h.invitationService.InviteUser(c.Request.Context(), orgID, userID, inviteReq)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"user_id":      userID,
			"org_id":       orgID,
			"invite_email": req.Email,
		}).Error("Failed to invite user")

		if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "already member") {
			response.Conflict(c, "User is already a member or has pending invitation")
			return
		}

		response.InternalServerError(c, "Failed to invite user")
		return
	}

	// Convert to response format
	inviterEmail := userID.String() // Fallback to userID string
	inviterUser, err := h.userService.GetUser(c.Request.Context(), userID)
	if err == nil {
		inviterEmail = inviterUser.Email
	}

	responseData := OrganizationMember{
		Email:     req.Email,
		Role:      req.Role,
		Status:    "invited",
		JoinedAt:  invitation.CreatedAt,
		InvitedBy: &inviterEmail,
		CreatedAt: invitation.CreatedAt,
		UpdatedAt: invitation.UpdatedAt,
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":       userID,
		"org_id":        orgID,
		"invite_email":  req.Email,
		"invitation_id": invitation.ID,
	}).Info("User invited successfully")

	c.Header("Location", fmt.Sprintf("/api/v1/organizations/%s/members", orgID))
	response.Success(c, responseData)
	c.Status(201) // Created
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
	var req RemoveMemberRequest
	if err := c.ShouldBindUri(&req); err != nil {
		h.logger.WithError(err).Error("Invalid remove member request")
		response.BadRequest(c, "Invalid organization ID or user ID", err.Error())
		return
	}

	// Get user ID from context (the user making the request)
	currentUserIDValue, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		response.Unauthorized(c, "Authentication required")
		return
	}

	currentUserID, ok := currentUserIDValue.(ulid.ULID)
	if !ok {
		h.logger.Error("Invalid user ID type in context")
		response.InternalServerError(c, "Internal error")
		return
	}

	// Parse organization ID
	orgID, err := ulid.Parse(req.OrgID)
	if err != nil {
		h.logger.WithError(err).WithField("org_id", req.OrgID).Error("Invalid organization ID format")
		response.BadRequest(c, "Invalid organization ID format", err.Error())
		return
	}

	// Parse user ID to remove
	userToRemoveID, err := ulid.Parse(req.UserID)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", req.UserID).Error("Invalid user ID format")
		response.BadRequest(c, "Invalid user ID format", err.Error())
		return
	}

	// Check if current user can access this organization
	canAccess, err := h.memberService.CanUserAccessOrganization(c.Request.Context(), currentUserID, orgID)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"current_user_id": currentUserID,
			"org_id":          orgID,
		}).Error("Failed to check organization access")
		response.InternalServerError(c, "Failed to check organization access")
		return
	}

	if !canAccess {
		h.logger.WithFields(logrus.Fields{
			"current_user_id": currentUserID,
			"org_id":          orgID,
		}).Warn("User attempted to remove member without permission")
		response.Forbidden(c, "Insufficient permissions to remove members from this organization")
		return
	}

	// Remove member
	err = h.memberService.RemoveMember(c.Request.Context(), orgID, userToRemoveID, currentUserID)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"current_user_id":   currentUserID,
			"user_to_remove_id": userToRemoveID,
			"org_id":            orgID,
		}).Error("Failed to remove member")

		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "Member")
			return
		}

		response.InternalServerError(c, "Failed to remove member")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"current_user_id":   currentUserID,
		"user_to_remove_id": userToRemoveID,
		"org_id":            orgID,
	}).Info("Member removed successfully")

	c.Status(204) // No Content
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

// InvitationDetailsResponse represents invitation details for validation
type InvitationDetailsResponse struct {
	ExpiresAt        time.Time `json:"expires_at"`
	OrganizationName string    `json:"organization_name" example:"Acme Corp"`
	OrganizationID   string    `json:"organization_id" example:"01HX..."`
	InviterName      string    `json:"inviter_name" example:"John"`
	Role             string    `json:"role" example:"developer"`
	Email            string    `json:"email" example:"user@example.com"`
	IsExpired        bool      `json:"is_expired"`
}

// ValidateInvitationToken validates an invitation token (PUBLIC endpoint)
// @Summary Validate invitation token
// @Description Validate an invitation token and return invitation details
// @Tags Invitations
// @Produce json
// @Param token path string true "Invitation token"
// @Success 200 {object} InvitationDetailsResponse "Valid invitation"
// @Failure 404 {object} response.ErrorResponse "Invalid or not found"
// @Failure 410 {object} response.ErrorResponse "Invitation expired"
// @Router /api/v1/invitations/validate/{token} [get]
func (h *Handler) ValidateInvitationToken(c *gin.Context) {
	token := c.Param("token")

	// Get invitation by token
	invitation, err := h.invitationService.GetInvitationByToken(c.Request.Context(), token)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get invitation")
		response.Error(c, err)
		return
	}

	// Check if expired or not pending
	isExpired := time.Now().After(invitation.ExpiresAt) ||
		invitation.Status != organization.InvitationStatusPending

	if isExpired {
		response.ErrorWithStatus(c, 410, "invitation_expired", "Invitation has expired or is no longer valid", "")
		return
	}

	// Get organization details
	org, err := h.organizationService.GetOrganization(c.Request.Context(), invitation.OrganizationID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get organization")
		response.Error(c, err)
		return
	}

	// Get inviter details
	inviter, err := h.userService.GetUser(c.Request.Context(), invitation.InvitedByID)
	if err != nil {
		h.logger.WithError(err).Warn("Failed to get inviter details")
		// Non-critical - continue with unknown inviter
	}

	// Get role details
	role, err := h.roleService.GetRoleByID(c.Request.Context(), invitation.RoleID)
	if err != nil {
		h.logger.WithError(err).Warn("Failed to get role details")
	}

	inviterName := "Unknown"
	if inviter != nil {
		inviterName = inviter.FirstName // Only first name for privacy
	}

	roleName := "Member"
	if role != nil {
		roleName = role.Name
	}

	resp := InvitationDetailsResponse{
		OrganizationName: org.Name,
		OrganizationID:   org.ID.String(),
		InviterName:      inviterName,
		Role:             roleName,
		Email:            invitation.Email,
		ExpiresAt:        invitation.ExpiresAt,
		IsExpired:        isExpired,
	}

	response.Success(c, resp)
}
