package rbac

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"brokle/internal/config"
	"brokle/internal/core/domain/auth"
	"brokle/pkg/response"
	"brokle/pkg/ulid"
)

// Handler handles RBAC-related HTTP requests (roles and permissions)
type Handler struct {
	config            *config.Config
	logger            *logrus.Logger
	roleService       auth.RoleService
	permissionService auth.PermissionService
}

// NewHandler creates a new RBAC handler
func NewHandler(
	config *config.Config,
	logger *logrus.Logger,
	roleService auth.RoleService,
	permissionService auth.PermissionService,
) *Handler {
	return &Handler{
		config:            config,
		logger:            logger,
		roleService:       roleService,
		permissionService: permissionService,
	}
}

// =============================================================================
// ROLE MANAGEMENT ENDPOINTS
// =============================================================================

// CreateRole handles POST /rbac/organizations/{orgId}/roles
// @Summary Create a new role
// @Description Create a new role for an organization (or global system role for admins)
// @Tags RBAC
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Param request body auth.CreateRoleRequest true "Role creation request"
// @Success 201 {object} response.RoleResponse "Role created successfully"
// @Failure 400 {object} response.ErrorResponse "Bad request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/rbac/organizations/{orgId}/roles [post]
func (h *Handler) CreateRole(c *gin.Context) {
	// Parse organization ID
	orgIDStr := c.Param("orgId")
	var orgID *ulid.ULID
	if orgIDStr != "" {
		parsedOrgID, err := ulid.Parse(orgIDStr)
		if err != nil {
			h.logger.WithError(err).WithField("org_id", orgIDStr).Error("Invalid organization ID")
			response.BadRequest(c, "Invalid organization ID", err.Error())
			return
		}
		orgID = &parsedOrgID
	}

	// Parse request body
	var req auth.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid create role request")
		response.BadRequest(c, "Invalid request payload", err.Error())
		return
	}

	// Set organization ID from path if not provided in body
	if req.OrganizationID == nil && orgID != nil {
		req.OrganizationID = orgID
	}

	// Create role
	role, err := h.roleService.CreateRole(c.Request.Context(), req.OrganizationID, &req)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"org_id":      req.OrganizationID,
			"role_name":   req.Name,
			"system_role": req.IsSystemRole,
		}).Error("Failed to create role")
		response.InternalServerError(c, "Failed to create role")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"role_id":     role.ID,
		"role_name":   role.Name,
		"org_id":      role.OrganizationID,
		"system_role": role.IsSystemRole,
	}).Info("Role created successfully")
	response.Created(c, role)
}

// GetRole handles GET /rbac/roles/{roleId}
// @Summary Get role details
// @Description Get detailed information about a specific role
// @Tags RBAC
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param roleId path string true "Role ID"
// @Success 200 {object} response.RoleResponse "Role details"
// @Failure 400 {object} response.ErrorResponse "Bad request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 404 {object} response.ErrorResponse "Role not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/rbac/roles/{roleId} [get]
func (h *Handler) GetRole(c *gin.Context) {
	// Parse role ID
	roleIDStr := c.Param("roleId")
	roleID, err := ulid.Parse(roleIDStr)
	if err != nil {
		h.logger.WithError(err).WithField("role_id", roleIDStr).Error("Invalid role ID")
		response.BadRequest(c, "Invalid role ID", err.Error())
		return
	}

	// Get role
	role, err := h.roleService.GetRole(c.Request.Context(), roleID)
	if err != nil {
		h.logger.WithError(err).WithField("role_id", roleID).Error("Failed to get role")
		response.NotFound(c, "Role not found")
		return
	}

	h.logger.WithField("role_id", roleID).Info("Role retrieved successfully")
	response.Success(c, role)
}

// UpdateRole handles PUT /rbac/roles/{roleId}
// @Summary Update role
// @Description Update an existing role (cannot update system roles)
// @Tags RBAC
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param roleId path string true "Role ID"
// @Param request body auth.UpdateRoleRequest true "Role update request"
// @Success 200 {object} response.RoleResponse "Role updated successfully"
// @Failure 400 {object} response.ErrorResponse "Bad request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden (system role cannot be updated)"
// @Failure 404 {object} response.ErrorResponse "Role not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/rbac/roles/{roleId} [put]
func (h *Handler) UpdateRole(c *gin.Context) {
	// Parse role ID
	roleIDStr := c.Param("roleId")
	roleID, err := ulid.Parse(roleIDStr)
	if err != nil {
		h.logger.WithError(err).WithField("role_id", roleIDStr).Error("Invalid role ID")
		response.BadRequest(c, "Invalid role ID", err.Error())
		return
	}

	// Parse request body
	var req auth.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid update role request")
		response.BadRequest(c, "Invalid request payload", err.Error())
		return
	}

	// Update role
	err = h.roleService.UpdateRole(c.Request.Context(), roleID, &req)
	if err != nil {
		h.logger.WithError(err).WithField("role_id", roleID).Error("Failed to update role")
		response.InternalServerError(c, "Failed to update role")
		return
	}

	// Get updated role
	role, err := h.roleService.GetRole(c.Request.Context(), roleID)
	if err != nil {
		h.logger.WithError(err).WithField("role_id", roleID).Error("Failed to get updated role")
		response.InternalServerError(c, "Failed to retrieve updated role")
		return
	}

	h.logger.WithField("role_id", roleID).Info("Role updated successfully")
	response.Success(c, role)
}

// DeleteRole handles DELETE /rbac/roles/{roleId}
// @Summary Delete role
// @Description Delete a role (cannot delete system roles)
// @Tags RBAC
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param roleId path string true "Role ID"
// @Success 204 "Role deleted successfully"
// @Failure 400 {object} response.ErrorResponse "Bad request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden (system role cannot be deleted)"
// @Failure 404 {object} response.ErrorResponse "Role not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/rbac/roles/{roleId} [delete]
func (h *Handler) DeleteRole(c *gin.Context) {
	// Parse role ID
	roleIDStr := c.Param("roleId")
	roleID, err := ulid.Parse(roleIDStr)
	if err != nil {
		h.logger.WithError(err).WithField("role_id", roleIDStr).Error("Invalid role ID")
		response.BadRequest(c, "Invalid role ID", err.Error())
		return
	}

	// Check if role can be deleted
	canDelete, err := h.roleService.CanDeleteRole(c.Request.Context(), roleID)
	if err != nil {
		h.logger.WithError(err).WithField("role_id", roleID).Error("Failed to check if role can be deleted")
		response.InternalServerError(c, "Failed to validate role deletion")
		return
	}
	if !canDelete {
		h.logger.WithField("role_id", roleID).Warn("Attempted to delete system role")
		response.Forbidden(c, "Cannot delete system role")
		return
	}

	// Delete role
	err = h.roleService.DeleteRole(c.Request.Context(), roleID)
	if err != nil {
		h.logger.WithError(err).WithField("role_id", roleID).Error("Failed to delete role")
		response.InternalServerError(c, "Failed to delete role")
		return
	}

	h.logger.WithField("role_id", roleID).Info("Role deleted successfully")
	response.NoContent(c)
}

// ListRoles handles GET /rbac/organizations/{orgId}/roles
// @Summary List roles for organization
// @Description Get all roles available for an organization (system + org-specific)
// @Tags RBAC
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Param limit query int false "Limit (default: 50, max: 100)"
// @Param offset query int false "Offset (default: 0)"
// @Param search query string false "Search query"
// @Success 200 {object} response.RoleListResponse "List of roles"
// @Failure 400 {object} response.ErrorResponse "Bad request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/rbac/organizations/{orgId}/roles [get]
func (h *Handler) ListRoles(c *gin.Context) {
	// Parse organization ID
	orgIDStr := c.Param("orgId")
	var orgID *ulid.ULID
	if orgIDStr != "" && orgIDStr != "global" {
		parsedOrgID, err := ulid.Parse(orgIDStr)
		if err != nil {
			h.logger.WithError(err).WithField("org_id", orgIDStr).Error("Invalid organization ID")
			response.BadRequest(c, "Invalid organization ID", err.Error())
			return
		}
		orgID = &parsedOrgID
	}

	// Parse query parameters
	limit := parseQueryInt(c, "limit", 50)
	if limit > 100 {
		limit = 100
	}
	offset := parseQueryInt(c, "offset", 0)
	search := c.Query("search")

	// List roles
	var result *auth.RoleListResponse
	var err error

	if search != "" {
		result, err = h.roleService.SearchRoles(c.Request.Context(), orgID, search, limit, offset)
	} else {
		result, err = h.roleService.ListRoles(c.Request.Context(), orgID, limit, offset)
	}

	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"org_id": orgID,
			"limit":  limit,
			"offset": offset,
			"search": search,
		}).Error("Failed to list roles")
		response.InternalServerError(c, "Failed to list roles")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"org_id":      orgID,
		"roles_count": result.TotalCount,
		"search":      search,
	}).Info("Roles listed successfully")
	response.Success(c, result)
}

// GetRoleStatistics handles GET /rbac/organizations/{orgId}/stats
// @Summary Get role statistics
// @Description Get statistics about roles and role distribution in an organization
// @Tags RBAC
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Success 200 {object} response.RoleStatistics "Role statistics"
// @Failure 400 {object} response.ErrorResponse "Bad request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/rbac/organizations/{orgId}/stats [get]
func (h *Handler) GetRoleStatistics(c *gin.Context) {
	// Parse organization ID
	orgIDStr := c.Param("orgId")
	orgID, err := ulid.Parse(orgIDStr)
	if err != nil {
		h.logger.WithError(err).WithField("org_id", orgIDStr).Error("Invalid organization ID")
		response.BadRequest(c, "Invalid organization ID", err.Error())
		return
	}

	// Get role statistics
	stats, err := h.roleService.GetRoleStatistics(c.Request.Context(), orgID)
	if err != nil {
		h.logger.WithError(err).WithField("org_id", orgID).Error("Failed to get role statistics")
		response.InternalServerError(c, "Failed to get role statistics")
		return
	}

	h.logger.WithField("org_id", orgID).Info("Role statistics retrieved successfully")
	response.Success(c, stats)
}

// =============================================================================
// PERMISSION MANAGEMENT ENDPOINTS
// =============================================================================

// CreatePermission handles POST /rbac/permissions
// @Summary Create a new permission
// @Description Create a new permission with resource:action format
// @Tags RBAC
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body auth.CreatePermissionRequest true "Permission creation request"
// @Success 201 {object} response.PermissionResponse "Permission created successfully"
// @Failure 400 {object} response.ErrorResponse "Bad request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/rbac/permissions [post]
func (h *Handler) CreatePermission(c *gin.Context) {
	// Parse request body
	var req auth.CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid create permission request")
		response.BadRequest(c, "Invalid request payload", err.Error())
		return
	}

	// Create permission
	permission, err := h.permissionService.CreatePermission(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"resource": req.Resource,
			"action":   req.Action,
			"category": req.Category,
		}).Error("Failed to create permission")
		response.InternalServerError(c, "Failed to create permission")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"permission_id": permission.ID,
		"resource":      permission.Resource,
		"action":        permission.Action,
	}).Info("Permission created successfully")
	response.Created(c, permission)
}

// GetPermission handles GET /rbac/permissions/{permissionId}
// @Summary Get permission details
// @Description Get detailed information about a specific permission
// @Tags RBAC
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param permissionId path string true "Permission ID"
// @Success 200 {object} response.PermissionResponse "Permission details"
// @Failure 400 {object} response.ErrorResponse "Bad request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 404 {object} response.ErrorResponse "Permission not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/rbac/permissions/{permissionId} [get]
func (h *Handler) GetPermission(c *gin.Context) {
	// Parse permission ID
	permissionIDStr := c.Param("permissionId")
	permissionID, err := ulid.Parse(permissionIDStr)
	if err != nil {
		h.logger.WithError(err).WithField("permission_id", permissionIDStr).Error("Invalid permission ID")
		response.BadRequest(c, "Invalid permission ID", err.Error())
		return
	}

	// Get permission
	permission, err := h.permissionService.GetPermission(c.Request.Context(), permissionID)
	if err != nil {
		h.logger.WithError(err).WithField("permission_id", permissionID).Error("Failed to get permission")
		response.NotFound(c, "Permission not found")
		return
	}

	h.logger.WithField("permission_id", permissionID).Info("Permission retrieved successfully")
	response.Success(c, permission)
}

// ListPermissions handles GET /rbac/permissions
// @Summary List all permissions
// @Description Get all available permissions in the system
// @Tags RBAC
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit (default: 50, max: 100)"
// @Param offset query int false "Offset (default: 0)"
// @Param category query string false "Filter by category"
// @Param resource query string false "Filter by resource"
// @Param search query string false "Search query"
// @Success 200 {object} response.PermissionListResponse "List of permissions"
// @Failure 400 {object} response.ErrorResponse "Bad request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/rbac/permissions [get]
func (h *Handler) ListPermissions(c *gin.Context) {
	// Parse query parameters
	limit := parseQueryInt(c, "limit", 50)
	if limit > 100 {
		limit = 100
	}
	offset := parseQueryInt(c, "offset", 0)
	category := c.Query("category")
	resource := c.Query("resource")
	search := c.Query("search")

	// List permissions based on filters
	var permissions []*auth.Permission
	var totalCount int
	var err error

	if search != "" {
		result, searchErr := h.permissionService.SearchPermissions(c.Request.Context(), search, limit, offset)
		if searchErr != nil {
			err = searchErr
		} else {
			permissions = result.Permissions
			totalCount = result.TotalCount
		}
	} else if category != "" {
		permissions, err = h.permissionService.GetPermissionsByCategory(c.Request.Context(), category)
		if err == nil {
			totalCount = len(permissions)
		}
	} else if resource != "" {
		permissions, err = h.permissionService.GetPermissionsByResource(c.Request.Context(), resource)
		if err == nil {
			totalCount = len(permissions)
		}
	} else {
		result, listErr := h.permissionService.ListPermissions(c.Request.Context(), limit, offset)
		if listErr != nil {
			err = listErr
		} else {
			permissions = result.Permissions
			totalCount = result.TotalCount
		}
	}

	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"limit":    limit,
			"offset":   offset,
			"category": category,
			"resource": resource,
			"search":   search,
		}).Error("Failed to list permissions")
		response.InternalServerError(c, "Failed to list permissions")
		return
	}

	result := &auth.PermissionListResponse{
		Permissions: permissions,
		TotalCount:  totalCount,
		Page:        offset/limit + 1,
		PageSize:    limit,
	}

	h.logger.WithFields(logrus.Fields{
		"permissions_count": totalCount,
		"category":          category,
		"resource":          resource,
		"search":            search,
	}).Info("Permissions listed successfully")
	response.Success(c, result)
}

// GetAvailableResources handles GET /rbac/permissions/resources
// @Summary Get available resources
// @Description Get all distinct resources that have permissions defined
// @Tags RBAC
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} string "List of available resources"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/rbac/permissions/resources [get]
func (h *Handler) GetAvailableResources(c *gin.Context) {
	// Get available resources
	resources, err := h.permissionService.GetAvailableResources(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get available resources")
		response.InternalServerError(c, "Failed to get available resources")
		return
	}

	h.logger.WithField("resources_count", len(resources)).Info("Available resources retrieved successfully")
	response.Success(c, resources)
}

// GetActionsForResource handles GET /rbac/permissions/resources/{resource}/actions
// @Summary Get actions for resource
// @Description Get all actions available for a specific resource
// @Tags RBAC
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param resource path string true "Resource name"
// @Success 200 {array} string "List of actions for the resource"
// @Failure 400 {object} response.ErrorResponse "Bad request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/rbac/permissions/resources/{resource}/actions [get]
func (h *Handler) GetActionsForResource(c *gin.Context) {
	resource := c.Param("resource")
	if resource == "" {
		response.BadRequest(c, "Resource parameter is required", "resource parameter cannot be empty")
		return
	}

	// Get actions for resource
	actions, err := h.permissionService.GetActionsForResource(c.Request.Context(), resource)
	if err != nil {
		h.logger.WithError(err).WithField("resource", resource).Error("Failed to get actions for resource")
		response.InternalServerError(c, "Failed to get actions for resource")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"resource":      resource,
		"actions_count": len(actions),
	}).Info("Actions for resource retrieved successfully")
	response.Success(c, actions)
}

// =============================================================================
// USER ROLE AND PERMISSION ENDPOINTS
// =============================================================================

// GetUserRole handles GET /rbac/users/{userId}/organizations/{orgId}/role
// @Summary Get user's role in organization
// @Description Get the role assigned to a user in a specific organization
// @Tags RBAC
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param userId path string true "User ID"
// @Param orgId path string true "Organization ID"
// @Success 200 {object} response.RoleResponse "User's role"
// @Failure 400 {object} response.ErrorResponse "Bad request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 404 {object} response.ErrorResponse "User role not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/rbac/users/{userId}/organizations/{orgId}/role [get]
func (h *Handler) GetUserRole(c *gin.Context) {
	// Parse user ID
	userIDStr := c.Param("userId")
	userID, err := ulid.Parse(userIDStr)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userIDStr).Error("Invalid user ID")
		response.BadRequest(c, "Invalid user ID", err.Error())
		return
	}

	// Parse organization ID
	orgIDStr := c.Param("orgId")
	orgID, err := ulid.Parse(orgIDStr)
	if err != nil {
		h.logger.WithError(err).WithField("org_id", orgIDStr).Error("Invalid organization ID")
		response.BadRequest(c, "Invalid organization ID", err.Error())
		return
	}

	// Get user role
	role, err := h.roleService.GetUserRole(c.Request.Context(), userID, orgID)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"user_id": userID,
			"org_id":  orgID,
		}).Error("Failed to get user role")
		response.NotFound(c, "User role not found")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"org_id":  orgID,
		"role_id": role.ID,
	}).Info("User role retrieved successfully")
	response.Success(c, role)
}

// GetUserPermissions handles GET /rbac/users/{userId}/organizations/{orgId}/permissions
// @Summary Get user's effective permissions
// @Description Get all effective permissions for a user in an organization
// @Tags RBAC
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param userId path string true "User ID"
// @Param orgId path string true "Organization ID"
// @Success 200 {object} response.UserPermissionsResponse "User's effective permissions"
// @Failure 400 {object} response.ErrorResponse "Bad request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 404 {object} response.ErrorResponse "User permissions not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/rbac/users/{userId}/organizations/{orgId}/permissions [get]
func (h *Handler) GetUserPermissions(c *gin.Context) {
	// Parse user ID
	userIDStr := c.Param("userId")
	userID, err := ulid.Parse(userIDStr)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userIDStr).Error("Invalid user ID")
		response.BadRequest(c, "Invalid user ID", err.Error())
		return
	}

	// Parse organization ID
	orgIDStr := c.Param("orgId")
	orgID, err := ulid.Parse(orgIDStr)
	if err != nil {
		h.logger.WithError(err).WithField("org_id", orgIDStr).Error("Invalid organization ID")
		response.BadRequest(c, "Invalid organization ID", err.Error())
		return
	}

	// Get user permissions
	permissions, err := h.roleService.GetUserPermissions(c.Request.Context(), userID, orgID)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"user_id": userID,
			"org_id":  orgID,
		}).Error("Failed to get user permissions")
		response.NotFound(c, "User permissions not found")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":            userID,
		"org_id":             orgID,
		"permissions_count":  len(permissions.Permissions),
	}).Info("User permissions retrieved successfully")
	response.Success(c, permissions)
}

// AssignUserRole handles POST /rbac/users/{userId}/organizations/{orgId}/role
// @Summary Assign role to user
// @Description Assign a role to a user in an organization
// @Tags RBAC
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param userId path string true "User ID"
// @Param orgId path string true "Organization ID"
// @Param request body auth.AssignRoleRequest true "Role assignment request"
// @Success 200 {object} response.RoleResponse "Assigned role"
// @Failure 400 {object} response.ErrorResponse "Bad request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/rbac/users/{userId}/organizations/{orgId}/role [post]
func (h *Handler) AssignUserRole(c *gin.Context) {
	// Parse user ID
	userIDStr := c.Param("userId")
	userID, err := ulid.Parse(userIDStr)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userIDStr).Error("Invalid user ID")
		response.BadRequest(c, "Invalid user ID", err.Error())
		return
	}

	// Parse organization ID
	orgIDStr := c.Param("orgId")
	orgID, err := ulid.Parse(orgIDStr)
	if err != nil {
		h.logger.WithError(err).WithField("org_id", orgIDStr).Error("Invalid organization ID")
		response.BadRequest(c, "Invalid organization ID", err.Error())
		return
	}

	// Parse request body
	var req auth.AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid assign role request")
		response.BadRequest(c, "Invalid request payload", err.Error())
		return
	}

	// Assign role to user
	err = h.roleService.AssignUserRole(c.Request.Context(), userID, orgID, req.RoleID)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"user_id": userID,
			"org_id":  orgID,
			"role_id": req.RoleID,
		}).Error("Failed to assign role to user")
		response.InternalServerError(c, "Failed to assign role to user")
		return
	}

	// Get assigned role for response
	role, err := h.roleService.GetRole(c.Request.Context(), req.RoleID)
	if err != nil {
		h.logger.WithError(err).WithField("role_id", req.RoleID).Error("Failed to get assigned role")
		response.InternalServerError(c, "Role assigned but failed to retrieve details")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"org_id":  orgID,
		"role_id": req.RoleID,
	}).Info("Role assigned to user successfully")
	response.Success(c, role)
}

// CheckUserPermissions handles POST /rbac/users/{userId}/organizations/{orgId}/permissions/check
// @Summary Check user permissions
// @Description Check if a user has specific permissions in an organization
// @Tags RBAC
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param userId path string true "User ID"
// @Param orgId path string true "Organization ID"
// @Param request body auth.CheckPermissionsRequest true "Permissions to check"
// @Success 200 {object} response.CheckPermissionsResponse "Permission check results"
// @Failure 400 {object} response.ErrorResponse "Bad request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/rbac/users/{userId}/organizations/{orgId}/permissions/check [post]
func (h *Handler) CheckUserPermissions(c *gin.Context) {
	// Parse user ID
	userIDStr := c.Param("userId")
	userID, err := ulid.Parse(userIDStr)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userIDStr).Error("Invalid user ID")
		response.BadRequest(c, "Invalid user ID", err.Error())
		return
	}

	// Parse organization ID
	orgIDStr := c.Param("orgId")
	orgID, err := ulid.Parse(orgIDStr)
	if err != nil {
		h.logger.WithError(err).WithField("org_id", orgIDStr).Error("Invalid organization ID")
		response.BadRequest(c, "Invalid organization ID", err.Error())
		return
	}

	// Parse request body
	var req auth.CheckPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid check permissions request")
		response.BadRequest(c, "Invalid request payload", err.Error())
		return
	}

	// Check permissions
	result, err := h.roleService.CheckPermissions(c.Request.Context(), userID, orgID, req.ResourceActions)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"user_id":            userID,
			"org_id":             orgID,
			"permissions_count":  len(req.ResourceActions),
		}).Error("Failed to check user permissions")
		response.InternalServerError(c, "Failed to check permissions")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":           userID,
		"org_id":            orgID,
		"permissions_count": len(req.ResourceActions),
	}).Info("User permissions checked successfully")
	response.Success(c, result)
}

// =============================================================================
// UTILITY FUNCTIONS
// =============================================================================

// parseQueryInt parses an integer query parameter with a default value
func parseQueryInt(c *gin.Context, key string, defaultValue int) int {
	if value := c.Query(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}