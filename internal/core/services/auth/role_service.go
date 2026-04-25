package auth

import (
	"context"
	"errors"

	"github.com/google/uuid"

	authDomain "brokle/internal/core/domain/auth"
	appErrors "brokle/pkg/errors"
)

// RoleService manages role templates used by organizations to assign member permissions.
type RoleService struct {
	roleRepo     authDomain.RoleRepository
	rolePermRepo authDomain.RolePermissionRepository
}

// NewRoleService creates a new clean role service instance
func NewRoleService(
	roleRepo authDomain.RoleRepository,
	rolePermRepo authDomain.RolePermissionRepository,
) *RoleService {
	return &RoleService{
		roleRepo:     roleRepo,
		rolePermRepo: rolePermRepo,
	}
}

// CreateRole creates a new template role
func (s *RoleService) CreateRole(ctx context.Context, req *authDomain.CreateRoleRequest) (*authDomain.Role, error) {
	// Validate request
	if req.Name == "" {
		return nil, appErrors.NewValidationError("name", "Role name is required")
	}
	if req.ScopeType == "" {
		return nil, appErrors.NewValidationError("scope_type", "Scope type is required")
	}

	// Check if role already exists with this name and scope
	existing, err := s.roleRepo.GetByNameAndScope(ctx, req.Name, req.ScopeType)
	if err == nil && existing != nil {
		return nil, appErrors.NewConflictError("role with name " + req.Name + " and scope " + req.ScopeType + " already exists")
	}

	// Create new role
	role := authDomain.NewRole(req.Name, req.ScopeType, req.Description)

	err = s.roleRepo.Create(ctx, role)
	if err != nil {
		if errors.Is(err, authDomain.ErrRoleAlreadyExists) {
			return nil, appErrors.NewConflictError("role with name " + role.Name + " and scope " + role.ScopeType + " already exists")
		}
		return nil, appErrors.NewInternalError("failed to create role", err)
	}

	return role, nil
}

// GetRoleByID gets a role by ID
func (s *RoleService) GetRoleByID(ctx context.Context, roleID uuid.UUID) (*authDomain.Role, error) {
	return s.roleRepo.GetByID(ctx, roleID)
}

// GetRoleByNameAndScope gets a role by name and scope type
func (s *RoleService) GetRoleByNameAndScope(ctx context.Context, name, scopeType string) (*authDomain.Role, error) {
	return s.roleRepo.GetByNameAndScope(ctx, name, scopeType)
}

// UpdateRole updates a role
func (s *RoleService) UpdateRole(ctx context.Context, roleID uuid.UUID, req *authDomain.UpdateRoleRequest) (*authDomain.Role, error) {
	// Get existing role
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return nil, appErrors.NewNotFoundError("role not found")
	}

	// Update fields
	if req.Description != nil {
		role.Description = req.Description
	}

	// Save changes
	err = s.roleRepo.Update(ctx, role)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to update role", err)
	}

	return role, nil
}

// DeleteRole deletes a role
func (s *RoleService) DeleteRole(ctx context.Context, roleID uuid.UUID) error {
	// Get role to check if it exists
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return appErrors.NewNotFoundError("role not found")
	}

	// Built-in role names that cannot be deleted
	builtinRoles := map[string]bool{
		"owner":     true,
		"admin":     true,
		"developer": true,
		"viewer":    true,
	}

	if builtinRoles[role.Name] {
		return appErrors.NewForbiddenError("cannot delete built-in role: " + role.Name)
	}

	return s.roleRepo.Delete(ctx, roleID)
}

// GetRolesByScopeType gets all roles for a specific scope type
func (s *RoleService) GetRolesByScopeType(ctx context.Context, scopeType string) ([]*authDomain.Role, error) {
	return s.roleRepo.GetByScopeType(ctx, scopeType)
}

// ListRoles gets all template roles
func (s *RoleService) ListRoles(ctx context.Context) ([]*authDomain.Role, error) {
	return s.roleRepo.ListRoles(ctx)
}

// GetRolePermissions gets all permissions assigned to a role
func (s *RoleService) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*authDomain.Permission, error) {
	return s.roleRepo.GetRolePermissions(ctx, roleID)
}

// AssignRolePermissions assigns permissions to a role
func (s *RoleService) AssignRolePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID, grantedBy *uuid.UUID) error {
	// Verify role exists
	_, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return appErrors.NewNotFoundError("role not found")
	}

	return s.roleRepo.AssignRolePermissions(ctx, roleID, permissionIDs, grantedBy)
}

// RevokeRolePermissions revokes permissions from a role
func (s *RoleService) RevokeRolePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	// Verify role exists
	_, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return appErrors.NewNotFoundError("role not found")
	}

	return s.roleRepo.RevokeRolePermissions(ctx, roleID, permissionIDs)
}

// GetRoleStatistics gets role usage statistics
func (s *RoleService) GetRoleStatistics(ctx context.Context) (*authDomain.RoleStatistics, error) {
	return s.roleRepo.GetRoleStatistics(ctx)
}

// System template role methods

func (s *RoleService) GetSystemRoles(ctx context.Context) ([]*authDomain.Role, error) {
	return s.roleRepo.GetSystemRoles(ctx)
}

// Custom scoped role management

func (s *RoleService) CreateCustomRole(ctx context.Context, scopeType string, scopeID uuid.UUID, req *authDomain.CreateRoleRequest) (*authDomain.Role, error) {
	// Validate request
	if req.Name == "" {
		return nil, appErrors.NewValidationError("name", "Role name is required")
	}
	if scopeType == "" {
		return nil, appErrors.NewValidationError("scope_type", "Scope type is required")
	}

	// Check if custom role already exists with this name and scope
	existing, err := s.roleRepo.GetByNameScopeAndID(ctx, req.Name, scopeType, &scopeID)
	if err == nil && existing != nil {
		return nil, appErrors.NewConflictError("custom role with name " + req.Name + " already exists in this scope")
	}

	// Create new custom role
	role := authDomain.NewCustomRole(req.Name, scopeType, req.Description, scopeID)

	err = s.roleRepo.Create(ctx, role)
	if err != nil {
		if errors.Is(err, authDomain.ErrRoleAlreadyExists) {
			return nil, appErrors.NewConflictError("custom role with name " + req.Name + " already exists in this scope")
		}
		return nil, appErrors.NewInternalError("failed to create custom role", err)
	}

	// Assign permissions if provided
	if len(req.PermissionIDs) > 0 {
		err = s.roleRepo.AssignRolePermissions(ctx, role.ID, req.PermissionIDs, nil)
		if err != nil {
			return nil, appErrors.NewInternalError("failed to assign permissions to custom role", err)
		}
	}

	return role, nil
}

func (s *RoleService) GetCustomRolesByOrganization(ctx context.Context, organizationID uuid.UUID) ([]*authDomain.Role, error) {
	return s.roleRepo.GetCustomRolesByOrganization(ctx, organizationID)
}

func (s *RoleService) UpdateCustomRole(ctx context.Context, roleID uuid.UUID, req *authDomain.UpdateRoleRequest) (*authDomain.Role, error) {
	// Get existing role and verify it's a custom role
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return nil, appErrors.NewNotFoundError("custom role not found")
	}

	if role.IsSystemRole() {
		return nil, appErrors.NewForbiddenError("cannot update system role")
	}

	// Update fields
	if req.Description != nil {
		role.Description = req.Description
	}

	// Save changes
	err = s.roleRepo.Update(ctx, role)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to update custom role", err)
	}

	// Update permissions if provided
	if req.PermissionIDs != nil {
		err = s.roleRepo.UpdateRolePermissions(ctx, role.ID, req.PermissionIDs, nil)
		if err != nil {
			return nil, appErrors.NewInternalError("failed to update role permissions", err)
		}
	}

	return role, nil
}

func (s *RoleService) DeleteCustomRole(ctx context.Context, roleID uuid.UUID) error {
	// Get role to check if it exists and is a custom role
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return appErrors.NewNotFoundError("custom role not found")
	}

	if role.IsSystemRole() {
		return appErrors.NewForbiddenError("cannot delete system role")
	}

	// TODO: Add check if role is in use by organization members
	// This would require checking the organization_members table

	return s.roleRepo.Delete(ctx, roleID)
}
