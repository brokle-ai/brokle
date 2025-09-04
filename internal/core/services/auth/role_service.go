package auth

import (
	"context"
	"fmt"
	"strings"

	"brokle/internal/core/domain/auth"
	"brokle/pkg/ulid"
)

// roleService implements auth.RoleService interface
type roleService struct {
	roleRepo       auth.RoleRepository
	permissionRepo auth.PermissionRepository
	rolePermRepo   auth.RolePermissionRepository
}

// NewRoleService creates a new role service instance
func NewRoleService(
	roleRepo auth.RoleRepository,
	permissionRepo auth.PermissionRepository,
	rolePermRepo auth.RolePermissionRepository,
) auth.RoleService {
	return &roleService{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
		rolePermRepo:   rolePermRepo,
	}
}

// Role management operations

// CreateRole creates a new role with validation
func (s *roleService) CreateRole(ctx context.Context, orgID *ulid.ULID, req *auth.CreateRoleRequest) (*auth.Role, error) {
	// Validate role name uniqueness within organization
	existing, err := s.roleRepo.GetByName(ctx, orgID, req.Name)
	if err != nil && err.Error() != "role not found" {
		return nil, fmt.Errorf("failed to check role name uniqueness: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("role with name '%s' already exists", req.Name)
	}

	// Validate permissions exist
	if len(req.PermissionIDs) > 0 {
		for _, permID := range req.PermissionIDs {
			_, err := s.permissionRepo.GetByID(ctx, permID)
			if err != nil {
				return nil, fmt.Errorf("permission %s does not exist: %w", permID, err)
			}
		}
	}

	// Create the role
	role := auth.NewRole(orgID, req.Name, req.DisplayName, req.Description, false)
	if err := s.roleRepo.Create(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	// Assign permissions to role
	if len(req.PermissionIDs) > 0 {
		if err := s.rolePermRepo.AssignPermissions(ctx, role.ID, req.PermissionIDs); err != nil {
			return nil, fmt.Errorf("failed to assign permissions to role: %w", err)
		}
	}

	// Load permissions for response
	permissions, err := s.permissionRepo.GetPermissionsByRoleID(ctx, role.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load role permissions: %w", err)
	}
	role.Permissions = make([]auth.Permission, len(permissions))
	for i, perm := range permissions {
		role.Permissions[i] = *perm
	}

	return role, nil
}

// GetRole retrieves a role by ID with permissions
func (s *roleService) GetRole(ctx context.Context, roleID ulid.ULID) (*auth.Role, error) {
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	// Load permissions
	permissions, err := s.permissionRepo.GetPermissionsByRoleID(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to load role permissions: %w", err)
	}
	role.Permissions = make([]auth.Permission, len(permissions))
	for i, perm := range permissions {
		role.Permissions[i] = *perm
	}

	return role, nil
}

// GetRoleByName retrieves a role by name and organization
func (s *roleService) GetRoleByName(ctx context.Context, orgID *ulid.ULID, name string) (*auth.Role, error) {
	role, err := s.roleRepo.GetByName(ctx, orgID, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get role by name: %w", err)
	}

	// Load permissions
	permissions, err := s.permissionRepo.GetPermissionsByRoleID(ctx, role.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load role permissions: %w", err)
	}
	role.Permissions = make([]auth.Permission, len(permissions))
	for i, perm := range permissions {
		role.Permissions[i] = *perm
	}

	return role, nil
}

// UpdateRole updates a role with validation
func (s *roleService) UpdateRole(ctx context.Context, roleID ulid.ULID, req *auth.UpdateRoleRequest) error {
	// Get existing role
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// Check if it's a system role (cannot be modified)
	if role.IsSystemRole {
		return fmt.Errorf("cannot modify system role")
	}

	// Update fields
	if req.DisplayName != nil {
		role.DisplayName = *req.DisplayName
	}
	if req.Description != nil {
		role.Description = *req.Description
	}

	// Update role
	if err := s.roleRepo.Update(ctx, role); err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	// Update permissions if provided
	if req.PermissionIDs != nil {
		// Validate permissions exist
		for _, permID := range req.PermissionIDs {
			_, err := s.permissionRepo.GetByID(ctx, permID)
			if err != nil {
				return fmt.Errorf("permission %s does not exist: %w", permID, err)
			}
		}

		// Remove all existing permissions and add new ones
		if err := s.rolePermRepo.RevokeAllPermissions(ctx, roleID); err != nil {
			return fmt.Errorf("failed to revoke existing permissions: %w", err)
		}

		if len(req.PermissionIDs) > 0 {
			if err := s.rolePermRepo.AssignPermissions(ctx, roleID, req.PermissionIDs); err != nil {
				return fmt.Errorf("failed to assign new permissions: %w", err)
			}
		}
	}

	return nil
}

// DeleteRole deletes a role with validation
func (s *roleService) DeleteRole(ctx context.Context, roleID ulid.ULID) error {
	// Get role to check if it's deletable
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// Cannot delete system roles
	if role.IsSystemRole {
		return fmt.Errorf("cannot delete system role")
	}

	// Remove all permissions first
	if err := s.rolePermRepo.RevokeAllPermissions(ctx, roleID); err != nil {
		return fmt.Errorf("failed to revoke role permissions: %w", err)
	}

	// Delete the role
	if err := s.roleRepo.Delete(ctx, roleID); err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	return nil
}

// ListRoles lists roles for an organization
func (s *roleService) ListRoles(ctx context.Context, orgID ulid.ULID) ([]*auth.Role, error) {
	roles, err := s.roleRepo.GetAllRoles(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	// Load permissions for each role
	for _, role := range roles {
		permissions, err := s.permissionRepo.GetPermissionsByRoleID(ctx, role.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load permissions for role %s: %w", role.ID, err)
		}
		role.Permissions = make([]auth.Permission, len(permissions))
		for i, perm := range permissions {
			role.Permissions[i] = *perm
		}
	}

	return roles, nil
}

// GetSystemRoles returns system-defined roles
func (s *roleService) GetSystemRoles(ctx context.Context) ([]*auth.Role, error) {
	roles, err := s.roleRepo.GetSystemRoles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get system roles: %w", err)
	}

	// Load permissions for each role
	for _, role := range roles {
		permissions, err := s.permissionRepo.GetPermissionsByRoleID(ctx, role.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load permissions for system role %s: %w", role.ID, err)
		}
		role.Permissions = make([]auth.Permission, len(permissions))
		for i, perm := range permissions {
			role.Permissions[i] = *perm
		}
	}

	return roles, nil
}

// Permission management operations

// GetPermissions returns all available permissions
func (s *roleService) GetPermissions(ctx context.Context) ([]*auth.Permission, error) {
	return s.permissionRepo.GetAllPermissions(ctx)
}

// GetPermissionsByCategory returns permissions by category
func (s *roleService) GetPermissionsByCategory(ctx context.Context, category string) ([]*auth.Permission, error) {
	return s.permissionRepo.GetByCategory(ctx, category)
}

// AssignPermissionsToRole assigns permissions to a role
func (s *roleService) AssignPermissionsToRole(ctx context.Context, roleID ulid.ULID, permissionIDs []ulid.ULID) error {
	// Check if role exists and is not system role
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}
	if role.IsSystemRole {
		return fmt.Errorf("cannot modify system role permissions")
	}

	// Validate permissions exist
	for _, permID := range permissionIDs {
		_, err := s.permissionRepo.GetByID(ctx, permID)
		if err != nil {
			return fmt.Errorf("permission %s does not exist: %w", permID, err)
		}
	}

	return s.rolePermRepo.AssignPermissions(ctx, roleID, permissionIDs)
}

// RemovePermissionsFromRole removes permissions from a role
func (s *roleService) RemovePermissionsFromRole(ctx context.Context, roleID ulid.ULID, permissionIDs []ulid.ULID) error {
	// Check if role exists and is not system role
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}
	if role.IsSystemRole {
		return fmt.Errorf("cannot modify system role permissions")
	}

	return s.rolePermRepo.RevokePermissions(ctx, roleID, permissionIDs)
}

// GetRolePermissions retrieves permissions for a role
func (s *roleService) GetRolePermissions(ctx context.Context, roleID ulid.ULID) ([]*auth.Permission, error) {
	return s.permissionRepo.GetPermissionsByRoleID(ctx, roleID)
}

// RBAC Permission checking - Core authorization methods

// HasPermission checks if a user has a specific permission in an organization
func (s *roleService) HasPermission(ctx context.Context, userID, orgID ulid.ULID, permission string) (bool, error) {
	userPermissions, err := s.permissionRepo.GetUserPermissions(ctx, userID, orgID)
	if err != nil {
		return false, fmt.Errorf("failed to get user permissions: %w", err)
	}

	return s.checkPermission(permission, userPermissions), nil
}

// HasPermissions checks if a user has all specified permissions in an organization
func (s *roleService) HasPermissions(ctx context.Context, userID, orgID ulid.ULID, permissions []string) (bool, error) {
	userPermissions, err := s.permissionRepo.GetUserPermissions(ctx, userID, orgID)
	if err != nil {
		return false, fmt.Errorf("failed to get user permissions: %w", err)
	}

	for _, permission := range permissions {
		if !s.checkPermission(permission, userPermissions) {
			return false, nil
		}
	}
	return true, nil
}

// HasAnyPermission checks if a user has any of the specified permissions in an organization
func (s *roleService) HasAnyPermission(ctx context.Context, userID, orgID ulid.ULID, permissions []string) (bool, error) {
	userPermissions, err := s.permissionRepo.GetUserPermissions(ctx, userID, orgID)
	if err != nil {
		return false, fmt.Errorf("failed to get user permissions: %w", err)
	}

	for _, permission := range permissions {
		if s.checkPermission(permission, userPermissions) {
			return true, nil
		}
	}
	return false, nil
}

// GetUserPermissions returns all permissions for a user in an organization
func (s *roleService) GetUserPermissions(ctx context.Context, userID, orgID ulid.ULID) ([]string, error) {
	return s.permissionRepo.GetUserPermissions(ctx, userID, orgID)
}

// GetUserRole returns the user's role in an organization (assumes single role per user per org)
func (s *roleService) GetUserRole(ctx context.Context, userID, orgID ulid.ULID) (*auth.Role, error) {
	// This would typically require a user-role repository to get the user's assigned role
	// For now, this is a placeholder that would need to be implemented based on your
	// organization membership system
	return nil, fmt.Errorf("GetUserRole not yet implemented - requires user-organization membership integration")
}

// Role seeding and defaults

// SeedSystemRoles creates the standard system roles
func (s *roleService) SeedSystemRoles(ctx context.Context) error {
	for roleName, permissions := range auth.SystemRoles {
		// Check if role already exists
		existing, err := s.roleRepo.GetByName(ctx, nil, roleName)
		if err == nil && existing != nil {
			continue // Role already exists
		}

		// Create system role
		role := auth.NewRole(nil, roleName, strings.Title(roleName), fmt.Sprintf("System-defined %s role", roleName), true)
		if err := s.roleRepo.Create(ctx, role); err != nil {
			return fmt.Errorf("failed to create system role %s: %w", roleName, err)
		}

		// Get permission IDs
		permissionIDs := make([]ulid.ULID, 0, len(permissions))
		for _, permName := range permissions {
			perm, err := s.permissionRepo.GetByName(ctx, permName)
			if err != nil {
				return fmt.Errorf("permission %s not found for role %s: %w", permName, roleName, err)
			}
			permissionIDs = append(permissionIDs, perm.ID)
		}

		// Assign permissions
		if len(permissionIDs) > 0 {
			if err := s.rolePermRepo.AssignPermissions(ctx, role.ID, permissionIDs); err != nil {
				return fmt.Errorf("failed to assign permissions to role %s: %w", roleName, err)
			}
		}
	}

	return nil
}

// SeedSystemPermissions creates the standard system permissions
func (s *roleService) SeedSystemPermissions(ctx context.Context) error {
	for _, permName := range auth.StandardPermissions {
		// Check if permission already exists
		existing, err := s.permissionRepo.GetByName(ctx, permName)
		if err == nil && existing != nil {
			continue // Permission already exists
		}

		// Parse permission to extract category and display name
		parts := strings.Split(permName, ".")
		if len(parts) != 2 {
			return fmt.Errorf("invalid permission format: %s", permName)
		}

		category := parts[0]
		action := parts[1]
		displayName := fmt.Sprintf("%s %s", strings.Title(action), strings.Title(category))
		description := fmt.Sprintf("Permission to %s %s", action, category)

		// Create permission
		perm := auth.NewPermission(permName, displayName, description, category)
		if err := s.permissionRepo.Create(ctx, perm); err != nil {
			return fmt.Errorf("failed to create permission %s: %w", permName, err)
		}
	}

	return nil
}

// EnsureDefaultRoles ensures default roles exist for an organization
func (s *roleService) EnsureDefaultRoles(ctx context.Context, orgID ulid.ULID) error {
	// For now, this just ensures system roles are available
	// In a full implementation, you might want to create organization-specific default roles
	return s.SeedSystemRoles(ctx)
}

// Helper methods

// checkPermission checks if a permission is granted, supporting wildcard permissions
func (s *roleService) checkPermission(requestedPermission string, userPermissions []string) bool {
	for _, userPerm := range userPermissions {
		// Exact match
		if userPerm == requestedPermission {
			return true
		}
		
		// Wildcard match (e.g., "users.*" matches "users.read")
		if userPerm == "*" {
			return true // Super admin permission
		}
		
		// Category wildcard (e.g., "users.*" matches "users.read")
		if strings.HasSuffix(userPerm, ".*") {
			category := strings.TrimSuffix(userPerm, ".*")
			if strings.HasPrefix(requestedPermission, category+".") {
				return true
			}
		}
	}
	
	return false
}