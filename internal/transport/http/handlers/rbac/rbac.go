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

// Handler handles RBAC-related HTTP requests (roles and permissions) - Clean version
type Handler struct {
	config                    *config.Config
	logger                    *logrus.Logger
	roleService               auth.RoleService
	permissionService         auth.PermissionService
	organizationMemberService auth.OrganizationMemberService
}

// NewHandler creates a new clean RBAC handler
func NewHandler(
	config *config.Config,
	logger *logrus.Logger,
	roleService auth.RoleService,
	permissionService auth.PermissionService,
	organizationMemberService auth.OrganizationMemberService,
) *Handler {
	return &Handler{
		config:                    config,
		logger:                    logger,
		roleService:               roleService,
		permissionService:         permissionService,
		organizationMemberService: organizationMemberService,
	}
}

// =============================================================================
// CLEAN ROLE MANAGEMENT ENDPOINTS
// =============================================================================

// CreateRole handles POST /rbac/roles
func (h *Handler) CreateRole(c *gin.Context) {
	var req auth.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid create role request")
		response.BadRequest(c, "Invalid request payload", err.Error())
		return
	}

	role, err := h.roleService.CreateRole(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"scope_type": req.ScopeType,
			"role_name":  req.Name,
		}).Error("Failed to create role")
		response.InternalServerError(c, "Failed to create role")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"role_id":    role.ID,
		"role_name":  role.Name,
		"scope_type": role.ScopeType,
	}).Info("Role created successfully")
	response.Created(c, role)
}

// GetRole handles GET /rbac/roles/{roleId}
func (h *Handler) GetRole(c *gin.Context) {
	roleIDStr := c.Param("roleId")
	roleID, err := ulid.Parse(roleIDStr)
	if err != nil {
		h.logger.WithError(err).WithField("role_id", roleIDStr).Error("Invalid role ID")
		response.BadRequest(c, "Invalid role ID", err.Error())
		return
	}

	role, err := h.roleService.GetRoleByID(c.Request.Context(), roleID)
	if err != nil {
		h.logger.WithError(err).WithField("role_id", roleID).Error("Failed to get role")
		response.NotFound(c, "Role not found")
		return
	}

	h.logger.WithField("role_id", roleID).Info("Role retrieved successfully")
	response.Success(c, role)
}

// UpdateRole handles PUT /rbac/roles/{roleId}
func (h *Handler) UpdateRole(c *gin.Context) {
	roleIDStr := c.Param("roleId")
	roleID, err := ulid.Parse(roleIDStr)
	if err != nil {
		h.logger.WithError(err).WithField("role_id", roleIDStr).Error("Invalid role ID")
		response.BadRequest(c, "Invalid role ID", err.Error())
		return
	}

	var req auth.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid update role request")
		response.BadRequest(c, "Invalid request payload", err.Error())
		return
	}

	updatedRole, err := h.roleService.UpdateRole(c.Request.Context(), roleID, &req)
	if err != nil {
		h.logger.WithError(err).WithField("role_id", roleID).Error("Failed to update role")
		response.InternalServerError(c, "Failed to update role")
		return
	}

	h.logger.WithField("role_id", roleID).Info("Role updated successfully")
	response.Success(c, updatedRole)
}

// DeleteRole handles DELETE /rbac/roles/{roleId}
func (h *Handler) DeleteRole(c *gin.Context) {
	roleIDStr := c.Param("roleId")
	roleID, err := ulid.Parse(roleIDStr)
	if err != nil {
		h.logger.WithError(err).WithField("role_id", roleIDStr).Error("Invalid role ID")
		response.BadRequest(c, "Invalid role ID", err.Error())
		return
	}

	// Get role first to check if it's a system role
	role, err := h.roleService.GetRoleByID(c.Request.Context(), roleID)
	if err != nil {
		h.logger.WithError(err).WithField("role_id", roleID).Error("Failed to get role")
		response.NotFound(c, "Role not found")
		return
	}

	// Check if it's a built-in role (cannot be deleted)
	builtinRoles := map[string]bool{
		"owner":     true,
		"admin":     true,
		"developer": true,
		"viewer":    true,
	}
	
	if builtinRoles[role.Name] {
		h.logger.WithField("role_id", roleID).Warn("Attempted to delete built-in role")
		response.Forbidden(c, "Cannot delete built-in role")
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

// ListRoles handles GET /rbac/roles
func (h *Handler) ListRoles(c *gin.Context) {
	scopeType := c.Query("scope_type")
	if scopeType == "" {
		response.BadRequest(c, "Scope type is required", "scope_type parameter cannot be empty")
		return
	}

	roles, err := h.roleService.GetRolesByScopeType(c.Request.Context(), scopeType)
	if err != nil {
		h.logger.WithError(err).WithField("scope_type", scopeType).Error("Failed to list roles")
		response.InternalServerError(c, "Failed to list roles")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"scope_type":  scopeType,
		"roles_count": len(roles),
	}).Info("Roles listed successfully")
	response.Success(c, roles)
}

// =============================================================================
// USER ROLE MANAGEMENT
// =============================================================================

// GetUserRoles handles GET /rbac/users/{userId}/roles
func (h *Handler) GetUserRoles(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := ulid.Parse(userIDStr)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userIDStr).Error("Invalid user ID")
		response.BadRequest(c, "Invalid user ID", err.Error())
		return
	}

	userMemberships, err := h.organizationMemberService.GetUserMemberships(c.Request.Context(), userID)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to get user memberships")
		response.NotFound(c, "User memberships not found")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":           userID,
		"memberships_count": len(userMemberships),
	}).Info("User memberships retrieved successfully")
	response.Success(c, userMemberships)
}

// GetUserPermissions handles GET /rbac/users/{userId}/permissions
func (h *Handler) GetUserPermissions(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := ulid.Parse(userIDStr)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userIDStr).Error("Invalid user ID")
		response.BadRequest(c, "Invalid user ID", err.Error())
		return
	}

	permissions, err := h.organizationMemberService.GetUserEffectivePermissions(c.Request.Context(), userID)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to get user permissions")
		response.NotFound(c, "User permissions not found")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":           userID,
		"permissions_count": len(permissions),
	}).Info("User permissions retrieved successfully")
	response.Success(c, permissions)
}

// AssignOrganizationRole handles POST /rbac/users/{userId}/organizations/{orgId}/roles
func (h *Handler) AssignOrganizationRole(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := ulid.Parse(userIDStr)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userIDStr).Error("Invalid user ID")
		response.BadRequest(c, "Invalid user ID", err.Error())
		return
	}

	orgIDStr := c.Param("orgId")
	orgID, err := ulid.Parse(orgIDStr)
	if err != nil {
		h.logger.WithError(err).WithField("org_id", orgIDStr).Error("Invalid organization ID")
		response.BadRequest(c, "Invalid organization ID", err.Error())
		return
	}

	var req auth.AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid assign role request")
		response.BadRequest(c, "Invalid request payload", err.Error())
		return
	}

	// Create organization membership with role
	member, err := h.organizationMemberService.AddMember(c.Request.Context(), userID, orgID, req.RoleID, nil)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"user_id": userID,
			"org_id":  orgID,
			"role_id": req.RoleID,
		}).Error("Failed to assign role to user in organization")
		response.InternalServerError(c, "Failed to assign role to user in organization")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"org_id":  orgID,
		"role_id": req.RoleID,
	}).Info("Role assigned to user in organization successfully")
	response.Created(c, member)
}

// RemoveOrganizationMember handles DELETE /rbac/users/{userId}/organizations/{orgId}
func (h *Handler) RemoveOrganizationMember(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := ulid.Parse(userIDStr)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userIDStr).Error("Invalid user ID")
		response.BadRequest(c, "Invalid user ID", err.Error())
		return
	}

	orgIDStr := c.Param("orgId")
	orgID, err := ulid.Parse(orgIDStr)
	if err != nil {
		h.logger.WithError(err).WithField("org_id", orgIDStr).Error("Invalid organization ID")
		response.BadRequest(c, "Invalid organization ID", err.Error())
		return
	}

	err = h.organizationMemberService.RemoveMember(c.Request.Context(), userID, orgID)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"user_id": userID,
			"org_id":  orgID,
		}).Error("Failed to remove user from organization")
		response.InternalServerError(c, "Failed to remove user from organization")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"org_id":  orgID,
	}).Info("User removed from organization successfully")
	response.NoContent(c)
}

// CheckUserPermissions handles POST /rbac/users/{userId}/permissions/check
func (h *Handler) CheckUserPermissions(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := ulid.Parse(userIDStr)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userIDStr).Error("Invalid user ID")
		response.BadRequest(c, "Invalid user ID", err.Error())
		return
	}

	var req auth.CheckPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid check permissions request")
		response.BadRequest(c, "Invalid request payload", err.Error())
		return
	}

	result, err := h.organizationMemberService.CheckUserPermissions(c.Request.Context(), userID, req.ResourceActions)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"user_id":           userID,
			"permissions_count": len(req.ResourceActions),
		}).Error("Failed to check user permissions")
		response.InternalServerError(c, "Failed to check permissions")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":           userID,
		"permissions_count": len(req.ResourceActions),
	}).Info("User permissions checked successfully")
	response.Success(c, result)
}

// GetRoleStatistics handles GET /rbac/roles/statistics
func (h *Handler) GetRoleStatistics(c *gin.Context) {
	stats, err := h.roleService.GetRoleStatistics(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get role statistics")
		response.InternalServerError(c, "Failed to get role statistics")
		return
	}

	h.logger.Info("Role statistics retrieved successfully")
	response.Success(c, stats)
}

// =============================================================================
// PERMISSION MANAGEMENT ENDPOINTS
// =============================================================================

// ListPermissions handles GET /rbac/permissions
func (h *Handler) ListPermissions(c *gin.Context) {
	limit := parseQueryInt(c, "limit", 50)
	if limit > 100 {
		limit = 100
	}
	offset := parseQueryInt(c, "offset", 0)

	result, err := h.permissionService.ListPermissions(c.Request.Context(), limit, offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list permissions")
		response.InternalServerError(c, "Failed to list permissions")
		return
	}

	h.logger.WithField("permissions_count", result.TotalCount).Info("Permissions listed successfully")
	response.Success(c, result)
}

// CreatePermission handles POST /rbac/permissions
func (h *Handler) CreatePermission(c *gin.Context) {
	var req auth.CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid create permission request")
		response.BadRequest(c, "Invalid request payload", err.Error())
		return
	}

	permission, err := h.permissionService.CreatePermission(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"resource": req.Resource,
			"action":   req.Action,
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
func (h *Handler) GetPermission(c *gin.Context) {
	permissionIDStr := c.Param("permissionId")
	permissionID, err := ulid.Parse(permissionIDStr)
	if err != nil {
		h.logger.WithError(err).WithField("permission_id", permissionIDStr).Error("Invalid permission ID")
		response.BadRequest(c, "Invalid permission ID", err.Error())
		return
	}

	permission, err := h.permissionService.GetPermission(c.Request.Context(), permissionID)
	if err != nil {
		h.logger.WithError(err).WithField("permission_id", permissionID).Error("Failed to get permission")
		response.NotFound(c, "Permission not found")
		return
	}

	h.logger.WithField("permission_id", permissionID).Info("Permission retrieved successfully")
	response.Success(c, permission)
}

// GetAvailableResources handles GET /rbac/permissions/resources
func (h *Handler) GetAvailableResources(c *gin.Context) {
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
func (h *Handler) GetActionsForResource(c *gin.Context) {
	resource := c.Param("resource")
	if resource == "" {
		response.BadRequest(c, "Resource parameter is required", "resource parameter cannot be empty")
		return
	}

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

// Legacy method for backward compatibility
func (h *Handler) GetUserRole(c *gin.Context) {
	// This is a legacy endpoint that should redirect to the new GetUserRoles endpoint
	userIDStr := c.Param("userId")
	userID, err := ulid.Parse(userIDStr)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userIDStr).Error("Invalid user ID")
		response.BadRequest(c, "Invalid user ID", err.Error())
		return
	}

	// Get all user memberships instead of direct roles
	userMemberships, err := h.organizationMemberService.GetUserMemberships(c.Request.Context(), userID)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to get user memberships")
		response.NotFound(c, "User memberships not found")
		return
	}

	// Return first membership for backward compatibility
	if len(userMemberships) > 0 {
		response.Success(c, userMemberships[0])
	} else {
		response.NotFound(c, "User has no organization memberships")
	}
}

// Custom Role Management Handlers

// CreateCustomRole creates a custom role for an organization
func (h *Handler) CreateCustomRole(c *gin.Context) {
	// Get organization ID from URL
	orgIDStr := c.Param("orgId")
	orgID, err := ulid.Parse(orgIDStr)
	if err != nil {
		h.logger.WithError(err).WithField("org_id", orgIDStr).Error("Invalid organization ID")
		response.BadRequest(c, "Invalid organization ID", err.Error())
		return
	}

	var req auth.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to bind custom role request")
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	// Validate scope type for custom roles
	if req.ScopeType != auth.ScopeOrganization {
		response.BadRequest(c, "Custom roles must have organization scope type", "scope_type must be 'organization' for custom roles")
		return
	}

	role, err := h.roleService.CreateCustomRole(c.Request.Context(), auth.ScopeOrganization, orgID, &req)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"org_id":     orgID,
			"role_name":  req.Name,
			"scope_type": req.ScopeType,
		}).Error("Failed to create custom role")
		
		if err.Error() == "custom role with name "+req.Name+" already exists in this scope" {
			response.Conflict(c, err.Error())
			return
		}
		
		response.InternalServerError(c, "Failed to create custom role")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"role_id":    role.ID,
		"role_name":  role.Name,
		"org_id":     orgID,
	}).Info("Custom role created successfully")

	response.Success(c, role)
}

// GetCustomRoles lists all custom roles for an organization
func (h *Handler) GetCustomRoles(c *gin.Context) {
	// Get organization ID from URL
	orgIDStr := c.Param("orgId")
	orgID, err := ulid.Parse(orgIDStr)
	if err != nil {
		h.logger.WithError(err).WithField("org_id", orgIDStr).Error("Invalid organization ID")
		response.BadRequest(c, "Invalid organization ID", err.Error())
		return
	}

	roles, err := h.roleService.GetCustomRolesByOrganization(c.Request.Context(), orgID)
	if err != nil {
		h.logger.WithError(err).WithField("org_id", orgID).Error("Failed to get custom roles")
		response.InternalServerError(c, "Failed to retrieve custom roles")
		return
	}

	response.Success(c, gin.H{
		"roles":       roles,
		"total_count": len(roles),
	})
}

// GetCustomRole retrieves a specific custom role
func (h *Handler) GetCustomRole(c *gin.Context) {
	roleIDStr := c.Param("roleId")
	roleID, err := ulid.Parse(roleIDStr)
	if err != nil {
		h.logger.WithError(err).WithField("role_id", roleIDStr).Error("Invalid role ID")
		response.BadRequest(c, "Invalid role ID", err.Error())
		return
	}

	role, err := h.roleService.GetRoleByID(c.Request.Context(), roleID)
	if err != nil {
		h.logger.WithError(err).WithField("role_id", roleID).Error("Failed to get custom role")
		response.NotFound(c, "Custom role not found")
		return
	}

	// Verify it's a custom role (not system role)
	if role.IsSystemRole() {
		response.BadRequest(c, "Cannot access system role through custom role endpoint", "use /rbac/roles endpoint for system roles")
		return
	}

	response.Success(c, role)
}

// UpdateCustomRole updates a custom role
func (h *Handler) UpdateCustomRole(c *gin.Context) {
	roleIDStr := c.Param("roleId")
	roleID, err := ulid.Parse(roleIDStr)
	if err != nil {
		h.logger.WithError(err).WithField("role_id", roleIDStr).Error("Invalid role ID")
		response.BadRequest(c, "Invalid role ID", err.Error())
		return
	}

	var req auth.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to bind update custom role request")
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	role, err := h.roleService.UpdateCustomRole(c.Request.Context(), roleID, &req)
	if err != nil {
		h.logger.WithError(err).WithField("role_id", roleID).Error("Failed to update custom role")
		
		if err.Error() == "cannot update system role" {
			response.Forbidden(c, err.Error())
			return
		}
		
		response.InternalServerError(c, "Failed to update custom role")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"role_id":   role.ID,
		"role_name": role.Name,
	}).Info("Custom role updated successfully")

	response.Success(c, role)
}

// DeleteCustomRole deletes a custom role
func (h *Handler) DeleteCustomRole(c *gin.Context) {
	roleIDStr := c.Param("roleId")
	roleID, err := ulid.Parse(roleIDStr)
	if err != nil {
		h.logger.WithError(err).WithField("role_id", roleIDStr).Error("Invalid role ID")
		response.BadRequest(c, "Invalid role ID", err.Error())
		return
	}

	err = h.roleService.DeleteCustomRole(c.Request.Context(), roleID)
	if err != nil {
		h.logger.WithError(err).WithField("role_id", roleID).Error("Failed to delete custom role")
		
		if err.Error() == "cannot delete system role" {
			response.Forbidden(c, err.Error())
			return
		}
		
		response.InternalServerError(c, "Failed to delete custom role")
		return
	}

	h.logger.WithField("role_id", roleID).Info("Custom role deleted successfully")
	response.Success(c, gin.H{"message": "Custom role deleted successfully"})
}

// parseQueryInt parses an integer query parameter with a default value
func parseQueryInt(c *gin.Context, key string, defaultValue int) int {
	if value := c.Query(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}