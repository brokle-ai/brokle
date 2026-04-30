package auth

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	authDomain "brokle/internal/core/domain/auth"
	appErrors "brokle/pkg/errors"
)

// floorScopePermName is the project-scope FLOOR permission. Every role
// granting any project-scoped permission MUST also grant this one, or
// users assigned the role will 403 on dashboard navigation
// (`GET /api/v1/projects/{projectId}` is gated by `projects:read`).
//
// Auto-injection in CreateCustomRole / UpdateCustomRole enforces the
// invariant at write time — matches GitHub's `metadata: read` floor-
// scope auto-include mechanic. See CLAUDE.md 2026-04-30 (project-rbac).
const floorScopePermName = "projects:read"

// RoleService manages role templates used by organizations to assign member permissions.
type RoleService struct {
	roleRepo       authDomain.RoleRepository
	rolePermRepo   authDomain.RolePermissionRepository
	permissionRepo authDomain.PermissionRepository
	logger         *slog.Logger
}

// NewRoleService creates a new clean role service instance
func NewRoleService(
	roleRepo authDomain.RoleRepository,
	rolePermRepo authDomain.RolePermissionRepository,
	permissionRepo authDomain.PermissionRepository,
	logger *slog.Logger,
) *RoleService {
	return &RoleService{
		roleRepo:       roleRepo,
		rolePermRepo:   rolePermRepo,
		permissionRepo: permissionRepo,
		logger:         logger,
	}
}

// orgProjectsTriggerNames are the org-tier project-management
// permissions that imply per-resource project access. Any role
// granting one of these MUST also carry the project-tier
// `projects:read` floor — otherwise the user can list/admin all
// projects in the org but 403s on `GET /api/v1/projects/{projectId}`
// (route gate is `projects:read` per routes.go).
//
// Why these specific verbs:
//
//   - `org_projects:list`  → "see all projects in the org" implies
//     "can read each project's details" (industry convention: list
//     implies read — GitHub, GitLab, Linear, Vercel, PostHog).
//   - `org_projects:admin` → "administer all projects in the org"
//     trivially implies "read each project."
//
// Why NOT `org_projects:create`: a create-only role gets per-project
// `projects:read` via the creator-override row written by
// ProjectService.CreateProject (round 22). Auto-injecting at role-
// creation time would broaden the user's read access to EVERY project
// in the org via inheritance — over-broad for "create-only" semantics.
// The override row correctly limits visibility to projects the user
// actually created.
var orgProjectsTriggerNames = map[string]struct{}{
	"org_projects:list":  {},
	"org_projects:admin": {},
}

// autoInjectFloorScope ensures the floor-scope invariant: any role
// granting a permission that implies per-resource project access
// MUST also carry `projects:read`. Two trigger families:
//
//  1. Any project-tier permission (scope_level = project, name !=
//     `projects:read`).
//  2. Any org-tier permission in `orgProjectsTriggerNames` —
//     `org_projects:list` and `org_projects:admin`. These let a user
//     list/admin every project in the org via the org role; without
//     the per-project floor they 403 on click-through (route gate is
//     `projects:read`).
//
// If any trigger fires AND `projects:read` is not already in the
// list, the floor permission's ID is appended. Idempotent on already-
// floor-included roles. Matches GitHub's `metadata: read` floor-
// scope auto-include mechanic.
func (s *RoleService) autoInjectFloorScope(ctx context.Context, permIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(permIDs) == 0 {
		return permIDs, nil
	}

	needsFloor := false
	hasFloor := false
	var floorID uuid.UUID

	for _, id := range permIDs {
		p, err := s.permissionRepo.GetByID(ctx, id)
		if err != nil {
			if appErrors.IsNotFound(err) {
				return nil, appErrors.InvalidParam("permission_ids",
					"unknown permission id "+id.String())
			}
			return nil, appErrors.Internal("load permission for floor-scope check", err,
				appErrors.WithOp("svc.role.auto_inject_floor_scope"))
		}
		if p.Name == floorScopePermName {
			hasFloor = true
			floorID = p.ID
			continue
		}
		if p.ScopeLevel == authDomain.ScopeLevelProject {
			needsFloor = true
			continue
		}
		if _, ok := orgProjectsTriggerNames[p.Name]; ok {
			needsFloor = true
		}
	}

	if !needsFloor || hasFloor {
		return permIDs, nil
	}

	// Need to inject — resolve floor permission ID if we didn't see it.
	if floorID == uuid.Nil {
		floor, err := s.permissionRepo.GetByName(ctx, floorScopePermName)
		if err != nil {
			return nil, appErrors.Internal("load floor permission for auto-inject", err,
				appErrors.WithOp("svc.role.auto_inject_floor_scope"))
		}
		floorID = floor.ID
	}

	if s.logger != nil {
		s.logger.InfoContext(ctx, "auto-injected projects:read floor scope",
			"floor_permission_id", floorID,
			"original_count", len(permIDs))
	}
	return append(permIDs, floorID), nil
}

// CreateRole creates a new template role
func (s *RoleService) CreateRole(ctx context.Context, req *authDomain.CreateRoleRequest) (*authDomain.Role, error) {
	// Validate request
	if req.Name == "" {
		return nil, appErrors.InvalidParam("name", "Role name is required")
	}
	if req.ScopeType == "" {
		return nil, appErrors.InvalidParam("scope_type", "Scope type is required")
	}

	// Check if role already exists with this name and scope
	existing, err := s.roleRepo.GetByNameAndScope(ctx, req.Name, req.ScopeType)
	if err == nil && existing != nil {
		return nil, appErrors.Conflict("role", "role with name " + req.Name + " and scope " + req.ScopeType + " already exists")
	}

	// Create new role
	role := authDomain.NewRole(req.Name, req.ScopeType, req.Description)

	err = s.roleRepo.Create(ctx, role)
	if err != nil {
		if appErrors.IsAlreadyExists(err) {
			return nil, appErrors.Conflict("role", "role with name " + role.Name + " and scope " + role.ScopeType + " already exists")
		}
		return nil, appErrors.Internal("failed to create role", err)
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
		return nil, appErrors.NotFound("role")
	}

	// Update fields
	if req.Description != nil {
		role.Description = req.Description
	}

	// Save changes
	err = s.roleRepo.Update(ctx, role)
	if err != nil {
		return nil, appErrors.Internal("failed to update role", err)
	}

	return role, nil
}

// DeleteRole deletes a role
func (s *RoleService) DeleteRole(ctx context.Context, roleID uuid.UUID) error {
	// Get role to check if it exists
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return appErrors.NotFound("role")
	}

	// Built-in role names that cannot be deleted
	builtinRoles := map[string]bool{
		"owner":     true,
		"admin":     true,
		"developer": true,
		"viewer":    true,
	}

	if builtinRoles[role.Name] {
		return appErrors.PermissionDenied("", "cannot delete built-in role: " + role.Name)
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
		return appErrors.NotFound("role")
	}

	return s.roleRepo.AssignRolePermissions(ctx, roleID, permissionIDs, grantedBy)
}

// RevokeRolePermissions revokes permissions from a role
func (s *RoleService) RevokeRolePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	// Verify role exists
	_, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return appErrors.NotFound("role")
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
		return nil, appErrors.InvalidParam("name", "Role name is required")
	}
	if scopeType == "" {
		return nil, appErrors.InvalidParam("scope_type", "Scope type is required")
	}

	// Check if custom role already exists with this name and scope
	existing, err := s.roleRepo.GetByNameScopeAndID(ctx, req.Name, scopeType, &scopeID)
	if err == nil && existing != nil {
		return nil, appErrors.Conflict("role", "custom role with name " + req.Name + " already exists in this scope")
	}

	// Create new custom role
	role := authDomain.NewCustomRole(req.Name, scopeType, req.Description, scopeID)

	err = s.roleRepo.Create(ctx, role)
	if err != nil {
		if appErrors.IsAlreadyExists(err) {
			return nil, appErrors.Conflict("role", "custom role with name " + req.Name + " already exists in this scope")
		}
		return nil, appErrors.Internal("failed to create custom role", err)
	}

	// Assign permissions if provided. Auto-inject the floor-scope
	// permission (`projects:read`) if any project-scoped perm is in the
	// list and the floor is missing — see autoInjectFloorScope docstring.
	if len(req.PermissionIDs) > 0 {
		permIDs, err := s.autoInjectFloorScope(ctx, req.PermissionIDs)
		if err != nil {
			return nil, err
		}
		err = s.roleRepo.AssignRolePermissions(ctx, role.ID, permIDs, nil)
		if err != nil {
			return nil, appErrors.Internal("failed to assign permissions to custom role", err)
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
		return nil, appErrors.NotFound("role")
	}

	if role.IsSystemRole() {
		return nil, appErrors.PermissionDenied("", "cannot update system role")
	}

	// Update fields
	if req.Description != nil {
		role.Description = req.Description
	}

	// Save changes
	err = s.roleRepo.Update(ctx, role)
	if err != nil {
		return nil, appErrors.Internal("failed to update custom role", err)
	}

	// Update permissions if provided. Same auto-injection as Create
	// so updates can't drop a role into a floor-scope-violating state.
	if req.PermissionIDs != nil {
		permIDs, err := s.autoInjectFloorScope(ctx, req.PermissionIDs)
		if err != nil {
			return nil, err
		}
		err = s.roleRepo.UpdateRolePermissions(ctx, role.ID, permIDs, nil)
		if err != nil {
			return nil, appErrors.Internal("failed to update role permissions", err)
		}
	}

	return role, nil
}

func (s *RoleService) DeleteCustomRole(ctx context.Context, roleID uuid.UUID) error {
	// Get role to check if it exists and is a custom role
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return appErrors.NotFound("role")
	}

	if role.IsSystemRole() {
		return appErrors.PermissionDenied("", "cannot delete system role")
	}

	// TODO: Add check if role is in use by organization members
	// This would require checking the organization_members table

	return s.roleRepo.Delete(ctx, roleID)
}
