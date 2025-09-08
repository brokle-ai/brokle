package auth

import (
	"context"
	"fmt"

	"brokle/internal/core/domain/auth"
	"brokle/pkg/ulid"
)

// roleService implements clean auth.RoleService interface
type roleService struct {
	roleRepo       auth.RoleRepository
	userRoleRepo   auth.UserRoleRepository
	permissionRepo auth.PermissionRepository
	rolePermRepo   auth.RolePermissionRepository
}

// NewRoleService creates a new clean role service instance
func NewRoleService(
	roleRepo auth.RoleRepository,
	userRoleRepo auth.UserRoleRepository,
	permissionRepo auth.PermissionRepository,
	rolePermRepo auth.RolePermissionRepository,
) auth.RoleService {
	return &roleService{
		roleRepo:       roleRepo,
		userRoleRepo:   userRoleRepo,
		permissionRepo: permissionRepo,
		rolePermRepo:   rolePermRepo,
	}
}

// Clean role management operations

// CreateRole creates a new scoped role with validation
func (s *roleService) CreateRole(ctx context.Context, req *auth.CreateRoleRequest) (*auth.Role, error) {
	// Resolve scope ID based on scope type
	var scopeID *ulid.ULID
	if req.ScopeType != auth.ScopeSystem {
		if req.ScopeID == nil {
			return nil, fmt.Errorf("scope_id is required for %s scoped roles", req.ScopeType)
		}
		scopeID = req.ScopeID
	}

	// Validate role name uniqueness within scope
	existing, err := s.roleRepo.GetByScopedName(ctx, req.ScopeType, scopeID, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check role name uniqueness: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("role with name '%s' already exists in %s scope", req.Name, req.ScopeType)
	}

	// Validate permissions exist
	if len(req.PermissionIDs) > 0 {
		for _, permID := range req.PermissionIDs {
			_, err := s.permissionRepo.GetByID(ctx, permID)
			if err != nil {
				return nil, fmt.Errorf("permission %s not found", permID.String())
			}
		}
	}

	// Create role
	role := auth.NewRole(req.ScopeType, scopeID, req.Name, req.DisplayName, req.Description)
	if err := s.roleRepo.Create(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	// Assign permissions
	if len(req.PermissionIDs) > 0 {
		if err := s.roleRepo.AssignRolePermissions(ctx, role.ID, req.PermissionIDs); err != nil {
			// Rollback role creation on permission assignment failure
			s.roleRepo.Delete(ctx, role.ID)
			return nil, fmt.Errorf("failed to assign permissions to role: %w", err)
		}
	}

	return role, nil
}

// GetRoleByID retrieves a role by its ID
func (s *roleService) GetRoleByID(ctx context.Context, roleID ulid.ULID) (*auth.Role, error) {
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	if role == nil {
		return nil, fmt.Errorf("role not found")
	}
	return role, nil
}

// GetRolesByScope retrieves roles within a specific scope
func (s *roleService) GetRolesByScope(ctx context.Context, scopeType string, scopeID *ulid.ULID) ([]*auth.Role, error) {
	roles, err := s.roleRepo.GetByScope(ctx, scopeType, scopeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles by scope: %w", err)
	}
	return roles, nil
}

// GetSystemRoles retrieves all system roles
func (s *roleService) GetSystemRoles(ctx context.Context) ([]*auth.Role, error) {
	return s.roleRepo.GetSystemRoles(ctx)
}

// GetOrganizationRoles retrieves roles for a specific organization
func (s *roleService) GetOrganizationRoles(ctx context.Context, orgID ulid.ULID) ([]*auth.Role, error) {
	return s.roleRepo.GetOrganizationRoles(ctx, orgID)
}

// UpdateRole updates an existing role
func (s *roleService) UpdateRole(ctx context.Context, roleID ulid.ULID, req *auth.UpdateRoleRequest) (*auth.Role, error) {
	// Get existing role
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	if role == nil {
		return nil, fmt.Errorf("role not found")
	}

	// Check if system role (cannot be modified)
	if role.IsSystemRole() {
		return nil, fmt.Errorf("system roles cannot be modified")
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
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	// Update permissions if provided
	if len(req.PermissionIDs) > 0 {
		if err := s.roleRepo.UpdateRolePermissions(ctx, roleID, req.PermissionIDs); err != nil {
			return nil, fmt.Errorf("failed to update role permissions: %w", err)
		}
	}

	return role, nil
}

// DeleteRole deletes a role
func (s *roleService) DeleteRole(ctx context.Context, roleID ulid.ULID) error {
	// Check if role can be deleted
	canDelete, err := s.roleRepo.CanDeleteRole(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to check if role can be deleted: %w", err)
	}
	if !canDelete {
		return fmt.Errorf("system roles cannot be deleted")
	}

	// Check if role has any users assigned
	userCount, err := s.userRoleRepo.GetRoleUserCount(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to check role usage: %w", err)
	}
	if userCount > 0 {
		return fmt.Errorf("cannot delete role: %d users are assigned to this role", userCount)
	}

	// Delete role
	return s.roleRepo.Delete(ctx, roleID)
}

// Clean user role management

// AssignUserRole assigns a role to a user
func (s *roleService) AssignUserRole(ctx context.Context, userID, roleID ulid.ULID) error {
	// Check if role exists
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}
	if role == nil {
		return fmt.Errorf("role not found")
	}

	// Check if assignment already exists
	exists, err := s.userRoleRepo.Exists(ctx, userID, roleID)
	if err != nil {
		return fmt.Errorf("failed to check existing assignment: %w", err)
	}
	if exists {
		return fmt.Errorf("user already has this role")
	}

	// Create assignment
	userRole := auth.NewUserRole(userID, roleID)
	return s.userRoleRepo.Create(ctx, userRole)
}

// RevokeUserRole removes a role from a user
func (s *roleService) RevokeUserRole(ctx context.Context, userID, roleID ulid.ULID) error {
	// Check if assignment exists
	exists, err := s.userRoleRepo.Exists(ctx, userID, roleID)
	if err != nil {
		return fmt.Errorf("failed to check existing assignment: %w", err)
	}
	if !exists {
		return fmt.Errorf("user does not have this role")
	}

	return s.userRoleRepo.Delete(ctx, userID, roleID)
}

// GetUserRoles retrieves all roles for a user
func (s *roleService) GetUserRoles(ctx context.Context, userID ulid.ULID) ([]*auth.Role, error) {
	return s.roleRepo.GetUserRoles(ctx, userID)
}

// GetUserEffectivePermissions gets all effective permissions for a user (union across all scopes)
func (s *roleService) GetUserEffectivePermissions(ctx context.Context, userID ulid.ULID) ([]string, error) {
	return s.roleRepo.GetUserEffectivePermissions(ctx, userID)
}

// CheckUserPermission checks if a user has a specific permission
func (s *roleService) CheckUserPermission(ctx context.Context, userID ulid.ULID, permission string) (bool, error) {
	return s.roleRepo.HasUserPermission(ctx, userID, permission)
}

// Clean role permission management

// GetRolePermissions retrieves permissions for a role
func (s *roleService) GetRolePermissions(ctx context.Context, roleID ulid.ULID) ([]*auth.Permission, error) {
	return s.roleRepo.GetRolePermissions(ctx, roleID)
}

// AssignRolePermissions assigns permissions to a role
func (s *roleService) AssignRolePermissions(ctx context.Context, roleID ulid.ULID, permissionIDs []ulid.ULID) error {
	// Validate role exists and is not system role
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}
	if role == nil {
		return fmt.Errorf("role not found")
	}
	if role.IsSystemRole() {
		return fmt.Errorf("cannot modify permissions for system roles")
	}

	return s.roleRepo.AssignRolePermissions(ctx, roleID, permissionIDs)
}

// RevokeRolePermissions removes permissions from a role
func (s *roleService) RevokeRolePermissions(ctx context.Context, roleID ulid.ULID, permissionIDs []ulid.ULID) error {
	// Validate role exists and is not system role
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}
	if role == nil {
		return fmt.Errorf("role not found")
	}
	if role.IsSystemRole() {
		return fmt.Errorf("cannot modify permissions for system roles")
	}

	return s.roleRepo.RevokeRolePermissions(ctx, roleID, permissionIDs)
}

// GetRoleStatistics gets statistics about roles
func (s *roleService) GetRoleStatistics(ctx context.Context, scopeType string, scopeID *ulid.ULID) (*auth.RoleStatistics, error) {
	return s.roleRepo.GetRoleStatistics(ctx)
}

// CheckUserPermissions checks if a user has specific permissions
func (s *roleService) CheckUserPermissions(ctx context.Context, userID ulid.ULID, permissions []string) (map[string]bool, error) {
	// Get user's effective permissions
	userPermissions, err := s.GetUserEffectivePermissions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user effective permissions: %w", err)
	}
	
	// Create a map for fast lookup
	permissionSet := make(map[string]bool)
	for _, perm := range userPermissions {
		permissionSet[perm] = true
	}
	
	// Check each requested permission
	result := make(map[string]bool)
	for _, permission := range permissions {
		result[permission] = permissionSet[permission]
	}
	
	return result, nil
}

// GetRoleByNameAndScope gets a role by name and scope
func (s *roleService) GetRoleByNameAndScope(ctx context.Context, name, scopeType string, scopeID *ulid.ULID) (*auth.Role, error) {
	return s.roleRepo.GetByScopedName(ctx, scopeType, scopeID, name)
}