package auth

import (
	"context"
	"errors"

	"github.com/google/uuid"

	authDomain "brokle/internal/core/domain/auth"
	appErrors "brokle/pkg/errors"
)

// ProjectMemberService manages project-level role overrides. A project
// membership is OPTIONAL — users access projects through their org-level
// role by default; a project_members row is created only when an admin
// wants to elevate (or in future, scope-restrict) a specific user's role
// inside a single project. Effective permissions follow Langfuse MAX
// semantics — see CheckUserPermissionsInScope.
type ProjectMemberService struct {
	projMemberRepo authDomain.ProjectMemberRepository
	orgMemberRepo  authDomain.OrganizationMemberRepository
	roleRepo       authDomain.RoleRepository
}

// NewProjectMemberService constructs the service.
func NewProjectMemberService(
	projMemberRepo authDomain.ProjectMemberRepository,
	orgMemberRepo authDomain.OrganizationMemberRepository,
	roleRepo authDomain.RoleRepository,
) *ProjectMemberService {
	return &ProjectMemberService{
		projMemberRepo: projMemberRepo,
		orgMemberRepo:  orgMemberRepo,
		roleRepo:       roleRepo,
	}
}

// AddMember elevates an existing org member with a project-specific role.
// The user MUST already be a member of the project's organization — project
// membership is an override on the org role, not a way to grant access to
// users outside the org.
func (s *ProjectMemberService) AddMember(ctx context.Context, userID, projectID, orgID, roleID uuid.UUID) (*authDomain.ProjectMember, error) {
	isOrgMember, err := s.orgMemberRepo.Exists(ctx, userID, orgID)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to check organization membership", err)
	}
	if !isOrgMember {
		return nil, appErrors.NewValidationError("User is not a member of the project's organization", "user must join the organization before being assigned a project role")
	}

	if _, err := s.roleRepo.GetByID(ctx, roleID); err != nil {
		if errors.Is(err, authDomain.ErrNotFound) {
			return nil, appErrors.NewNotFoundError("role")
		}
		return nil, appErrors.NewInternalError("failed to load role", err)
	}

	exists, err := s.projMemberRepo.IsMember(ctx, userID, projectID)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to check project membership", err)
	}
	if exists {
		return nil, appErrors.NewConflictError("user already has a project-level role; use UpdateMemberRole instead")
	}

	member := authDomain.NewProjectMember(userID, projectID, roleID)
	if err := s.projMemberRepo.Create(ctx, member); err != nil {
		return nil, appErrors.NewInternalError("failed to create project membership", err)
	}
	return member, nil
}

// UpdateMemberRole changes the project-level role of an existing project
// member. Returns 404 if no project_members row exists for (user, project).
func (s *ProjectMemberService) UpdateMemberRole(ctx context.Context, userID, projectID, roleID uuid.UUID) error {
	if _, err := s.roleRepo.GetByID(ctx, roleID); err != nil {
		if errors.Is(err, authDomain.ErrNotFound) {
			return appErrors.NewNotFoundError("role")
		}
		return appErrors.NewInternalError("failed to load role", err)
	}

	exists, err := s.projMemberRepo.IsMember(ctx, userID, projectID)
	if err != nil {
		return appErrors.NewInternalError("failed to check project membership", err)
	}
	if !exists {
		return appErrors.NewNotFoundError("project member")
	}

	if err := s.projMemberRepo.UpdateRole(ctx, userID, projectID, roleID); err != nil {
		return appErrors.NewInternalError("failed to update project member role", err)
	}
	return nil
}

// RemoveMember deletes the project-level role override; the user reverts
// to whatever access their org-level role grants them.
func (s *ProjectMemberService) RemoveMember(ctx context.Context, userID, projectID uuid.UUID) error {
	exists, err := s.projMemberRepo.IsMember(ctx, userID, projectID)
	if err != nil {
		return appErrors.NewInternalError("failed to check project membership", err)
	}
	if !exists {
		return appErrors.NewNotFoundError("project member")
	}

	if err := s.projMemberRepo.Delete(ctx, userID, projectID); err != nil {
		return appErrors.NewInternalError("failed to remove project membership", err)
	}
	return nil
}

// GetMember returns a single project membership row, or NotFound.
func (s *ProjectMemberService) GetMember(ctx context.Context, userID, projectID uuid.UUID) (*authDomain.ProjectMember, error) {
	m, err := s.projMemberRepo.GetByUserAndProject(ctx, userID, projectID)
	if err != nil {
		if errors.Is(err, authDomain.ErrNotFound) {
			return nil, appErrors.NewNotFoundError("project member")
		}
		return nil, appErrors.NewInternalError("failed to get project member", err)
	}
	return m, nil
}

// ListProjectMembers returns every project_members row for a project (the
// users with explicit role overrides). Org-level-only members do NOT appear
// here — they're surfaced via OrganizationMemberService.
func (s *ProjectMemberService) ListProjectMembers(ctx context.Context, projectID uuid.UUID) ([]*authDomain.ProjectMember, error) {
	members, err := s.projMemberRepo.ListByProject(ctx, projectID)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to list project members", err)
	}
	return members, nil
}

// IsMember checks for a project_members row. Note: this does NOT mean
// "can the user access the project" — that's RequireProjectAccess
// (which checks org membership too).
func (s *ProjectMemberService) IsMember(ctx context.Context, userID, projectID uuid.UUID) (bool, error) {
	return s.projMemberRepo.IsMember(ctx, userID, projectID)
}

// CheckUserPermissionsInScope resolves the user's effective permissions
// for (orgID, optional projectID) using Langfuse MAX semantics: the
// effective permission set is the union of org-role perms and (if any)
// project-role perms. Pass uuid.Nil as projectID to resolve org-only.
//
// Returns map[permission]granted for each requested permission.
func (s *ProjectMemberService) CheckUserPermissionsInScope(
	ctx context.Context,
	userID, orgID, projectID uuid.UUID,
	permissions []string,
) (map[string]bool, error) {
	if len(permissions) == 0 {
		return map[string]bool{}, nil
	}
	granted, err := s.projMemberRepo.ListUserEffectivePermissionsInScope(ctx, userID, orgID, projectID)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to resolve effective permissions", err)
	}
	grantedSet := make(map[string]struct{}, len(granted))
	for _, p := range granted {
		grantedSet[p] = struct{}{}
	}
	result := make(map[string]bool, len(permissions))
	for _, p := range permissions {
		_, ok := grantedSet[p]
		result[p] = ok
	}
	return result, nil
}
