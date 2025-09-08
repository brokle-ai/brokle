package auth

import (
	"context"
	"fmt"

	"brokle/internal/core/domain/auth"
	"brokle/pkg/ulid"
)

// roleService implements clean auth.RoleService interface (template roles only)
type roleService struct {
	roleRepo     auth.RoleRepository
	rolePermRepo auth.RolePermissionRepository
}

// NewRoleService creates a new clean role service instance
func NewRoleService(
	roleRepo auth.RoleRepository,
	rolePermRepo auth.RolePermissionRepository,
) auth.RoleService {
	return &roleService{
		roleRepo:     roleRepo,
		rolePermRepo: rolePermRepo,
	}
}

// CreateRole creates a new template role
func (s *roleService) CreateRole(ctx context.Context, req *auth.CreateRoleRequest) (*auth.Role, error) {
	// Validate request
	if req.Name == "" {
		return nil, fmt.Errorf("role name is required")
	}
	if req.ScopeType == "" {
		return nil, fmt.Errorf("scope type is required")
	}

	// Check if role already exists with this name and scope
	existing, err := s.roleRepo.GetByNameAndScope(ctx, req.Name, req.ScopeType)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("role with name %s and scope %s already exists", req.Name, req.ScopeType)
	}

	// Create new role
	role := auth.NewRole(req.Name, req.ScopeType, req.Description)
	
	err = s.roleRepo.Create(ctx, role)
	if err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	return role, nil
}

// GetRoleByID gets a role by ID
func (s *roleService) GetRoleByID(ctx context.Context, roleID ulid.ULID) (*auth.Role, error) {
	return s.roleRepo.GetByID(ctx, roleID)
}

// GetRoleByNameAndScope gets a role by name and scope type
func (s *roleService) GetRoleByNameAndScope(ctx context.Context, name, scopeType string) (*auth.Role, error) {
	return s.roleRepo.GetByNameAndScope(ctx, name, scopeType)
}

// UpdateRole updates a role
func (s *roleService) UpdateRole(ctx context.Context, roleID ulid.ULID, req *auth.UpdateRoleRequest) (*auth.Role, error) {
	// Get existing role
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("role not found: %w", err)
	}

	// Update fields
	if req.Description != nil {
		role.Description = *req.Description
	}

	// Save changes
	err = s.roleRepo.Update(ctx, role)
	if err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	return role, nil
}

// DeleteRole deletes a role
func (s *roleService) DeleteRole(ctx context.Context, roleID ulid.ULID) error {
	// Get role to check if it exists
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// Built-in role names that cannot be deleted
	builtinRoles := map[string]bool{
		"owner":     true,
		"admin":     true,
		"developer": true,
		"viewer":    true,
	}

	if builtinRoles[role.Name] {
		return fmt.Errorf("cannot delete built-in role: %s", role.Name)
	}

	return s.roleRepo.Delete(ctx, roleID)
}

// GetRolesByScopeType gets all roles for a specific scope type
func (s *roleService) GetRolesByScopeType(ctx context.Context, scopeType string) ([]*auth.Role, error) {
	return s.roleRepo.GetByScopeType(ctx, scopeType)
}

// GetAllRoles gets all template roles
func (s *roleService) GetAllRoles(ctx context.Context) ([]*auth.Role, error) {
	return s.roleRepo.GetAllRoles(ctx)
}

// GetRolePermissions gets all permissions assigned to a role
func (s *roleService) GetRolePermissions(ctx context.Context, roleID ulid.ULID) ([]*auth.Permission, error) {
	return s.roleRepo.GetRolePermissions(ctx, roleID)
}

// AssignRolePermissions assigns permissions to a role
func (s *roleService) AssignRolePermissions(ctx context.Context, roleID ulid.ULID, permissionIDs []ulid.ULID, grantedBy *ulid.ULID) error {
	// Verify role exists
	_, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	return s.roleRepo.AssignRolePermissions(ctx, roleID, permissionIDs, grantedBy)
}

// RevokeRolePermissions revokes permissions from a role
func (s *roleService) RevokeRolePermissions(ctx context.Context, roleID ulid.ULID, permissionIDs []ulid.ULID) error {
	// Verify role exists
	_, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	return s.roleRepo.RevokeRolePermissions(ctx, roleID, permissionIDs)
}

// GetRoleStatistics gets role usage statistics
func (s *roleService) GetRoleStatistics(ctx context.Context) (*auth.RoleStatistics, error) {
	return s.roleRepo.GetRoleStatistics(ctx)
}

// System template role methods

func (s *roleService) GetSystemRoles(ctx context.Context) ([]*auth.Role, error) {
	return s.roleRepo.GetSystemRoles(ctx)
}

// Custom scoped role management

func (s *roleService) CreateCustomRole(ctx context.Context, scopeType string, scopeID ulid.ULID, req *auth.CreateRoleRequest) (*auth.Role, error) {
	// Validate request
	if req.Name == "" {
		return nil, fmt.Errorf("role name is required")
	}
	if scopeType == "" {
		return nil, fmt.Errorf("scope type is required")
	}

	// Check if custom role already exists with this name and scope
	existing, err := s.roleRepo.GetByNameScopeAndID(ctx, req.Name, scopeType, &scopeID)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("custom role with name %s already exists in this scope", req.Name)
	}

	// Create new custom role
	role := auth.NewCustomRole(req.Name, scopeType, req.Description, scopeID)
	
	err = s.roleRepo.Create(ctx, role)
	if err != nil {
		return nil, fmt.Errorf("failed to create custom role: %w", err)
	}

	// Assign permissions if provided
	if len(req.PermissionIDs) > 0 {
		err = s.roleRepo.AssignRolePermissions(ctx, role.ID, req.PermissionIDs, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to assign permissions to custom role: %w", err)
		}
	}

	return role, nil
}

func (s *roleService) GetCustomRolesByOrganization(ctx context.Context, organizationID ulid.ULID) ([]*auth.Role, error) {
	return s.roleRepo.GetCustomRolesByOrganization(ctx, organizationID)
}

func (s *roleService) UpdateCustomRole(ctx context.Context, roleID ulid.ULID, req *auth.UpdateRoleRequest) (*auth.Role, error) {
	// Get existing role and verify it's a custom role
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("custom role not found: %w", err)
	}

	if role.IsSystemRole() {
		return nil, fmt.Errorf("cannot update system role")
	}

	// Update fields
	if req.Description != nil {
		role.Description = *req.Description
	}

	// Save changes
	err = s.roleRepo.Update(ctx, role)
	if err != nil {
		return nil, fmt.Errorf("failed to update custom role: %w", err)
	}

	// Update permissions if provided
	if req.PermissionIDs != nil {
		err = s.roleRepo.UpdateRolePermissions(ctx, role.ID, req.PermissionIDs, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to update role permissions: %w", err)
		}
	}

	return role, nil
}

func (s *roleService) DeleteCustomRole(ctx context.Context, roleID ulid.ULID) error {
	// Get role to check if it exists and is a custom role
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("custom role not found: %w", err)
	}

	if role.IsSystemRole() {
		return fmt.Errorf("cannot delete system role")
	}

	// TODO: Add check if role is in use by organization members
	// This would require checking the organization_members table

	return s.roleRepo.Delete(ctx, roleID)
}