package auth

import (
	"context"

	"github.com/google/uuid"

	authDomain "brokle/internal/core/domain/auth"
	"brokle/internal/core/domain/common"
	appErrors "brokle/pkg/errors"
)

// OrganizationMemberService assigns roles to organization members and checks access.
type OrganizationMemberService struct {
	orgMemberRepo     authDomain.OrganizationMemberRepository
	projectMemberRepo authDomain.ProjectMemberRepository
	roleRepo          authDomain.RoleRepository
	tx                common.Transactor
}

// NewOrganizationMemberService creates a new organization member service instance
func NewOrganizationMemberService(
	orgMemberRepo authDomain.OrganizationMemberRepository,
	projectMemberRepo authDomain.ProjectMemberRepository,
	roleRepo authDomain.RoleRepository,
	tx common.Transactor,
) *OrganizationMemberService {
	return &OrganizationMemberService{
		orgMemberRepo:     orgMemberRepo,
		projectMemberRepo: projectMemberRepo,
		roleRepo:          roleRepo,
		tx:                tx,
	}
}

// AddMember adds a user to an organization with specified role
func (s *OrganizationMemberService) AddMember(ctx context.Context, userID, orgID, roleID uuid.UUID, invitedBy *uuid.UUID) (*authDomain.OrganizationMember, error) {
	// Verify role exists
	_, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return nil, appErrors.NotFound("role")
	}

	// Check if user is already a member
	exists, err := s.orgMemberRepo.Exists(ctx, userID, orgID)
	if err != nil {
		return nil, appErrors.Internal("failed to check membership", err)
	}
	if exists {
		return nil, appErrors.Conflict("", "user is already a member of this organization")
	}

	// Create new membership
	member := authDomain.NewOrganizationMember(userID, orgID, roleID, invitedBy)

	err = s.orgMemberRepo.Create(ctx, member)
	if err != nil {
		return nil, appErrors.Internal("failed to create membership", err)
	}

	return member, nil
}

// RemoveMember removes a user from an organization. The two writes
// — cascade-delete every project_members row in this org's projects,
// then soft-delete the organization_members row — must commit
// together or roll back together. project_members has no FK cascade
// to organization_members, so without service-level atomicity a
// transient DB error between the two writes would leave the user
// holding org membership but missing every project grant (or
// vice-versa). WithinTransaction is reentrant (FLATTEN), so callers
// already inside an outer tx reuse it.
func (s *OrganizationMemberService) RemoveMember(ctx context.Context, userID, orgID uuid.UUID) error {
	// Check if user is a member
	exists, err := s.orgMemberRepo.Exists(ctx, userID, orgID)
	if err != nil {
		return appErrors.Internal("failed to check membership", err)
	}
	if !exists {
		return appErrors.NotFound("member", appErrors.WithMessage("user is not a member of this organization"))
	}

	return s.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := s.projectMemberRepo.DeleteAllInOrgForUser(ctx, userID, orgID); err != nil {
			return appErrors.Internal("failed to cascade-delete project_members", err)
		}
		if err := s.orgMemberRepo.Delete(ctx, userID, orgID); err != nil {
			return appErrors.Internal("failed to remove organization member", err)
		}
		return nil
	})
}

// UpdateMemberRole updates a member's role in an organization
func (s *OrganizationMemberService) UpdateMemberRole(ctx context.Context, userID, orgID, roleID uuid.UUID) error {
	// Verify role exists
	_, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return appErrors.NotFound("role")
	}

	// Check if user is a member
	exists, err := s.orgMemberRepo.Exists(ctx, userID, orgID)
	if err != nil {
		return appErrors.Internal("failed to check membership", err)
	}
	if !exists {
		return appErrors.NotFound("member", appErrors.WithMessage("user is not a member of this organization"))
	}

	return s.orgMemberRepo.UpdateMemberRole(ctx, userID, orgID, roleID)
}

// GetMember gets a specific organization membership
func (s *OrganizationMemberService) GetMember(ctx context.Context, userID, orgID uuid.UUID) (*authDomain.OrganizationMember, error) {
	return s.orgMemberRepo.GetByUserAndOrganization(ctx, userID, orgID)
}

// GetUserMemberships gets all organization memberships for a user
func (s *OrganizationMemberService) GetUserMemberships(ctx context.Context, userID uuid.UUID) ([]*authDomain.OrganizationMember, error) {
	return s.orgMemberRepo.GetByUserID(ctx, userID)
}

// GetOrganizationMembers gets all members of an organization
func (s *OrganizationMemberService) GetOrganizationMembers(ctx context.Context, orgID uuid.UUID) ([]*authDomain.OrganizationMember, error) {
	return s.orgMemberRepo.GetByOrganizationID(ctx, orgID)
}

// GetMembersByRole gets all members with a specific role
func (s *OrganizationMemberService) GetMembersByRole(ctx context.Context, roleID uuid.UUID) ([]*authDomain.OrganizationMember, error) {
	return s.orgMemberRepo.GetByRole(ctx, roleID)
}

// IsMember checks if a user is a member of an organization
func (s *OrganizationMemberService) IsMember(ctx context.Context, userID, orgID uuid.UUID) (bool, error) {
	return s.orgMemberRepo.Exists(ctx, userID, orgID)
}

// GetUserEffectivePermissions gets all effective permissions for a user across all organizations
func (s *OrganizationMemberService) GetUserEffectivePermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return s.orgMemberRepo.GetUserEffectivePermissions(ctx, userID)
}

// GetUserPermissionsInOrganization gets user permissions within a specific organization
func (s *OrganizationMemberService) GetUserPermissionsInOrganization(ctx context.Context, userID, orgID uuid.UUID) ([]string, error) {
	return s.orgMemberRepo.GetUserPermissionsInOrganization(ctx, userID, orgID)
}

// CheckUserPermissions checks multiple permissions for a user across every
// org they belong to. Used by the org-agnostic /api/v1/rbac/users/{userId}/
// permissions/check endpoint. The scope-aware permission middleware uses
// ProjectMemberService.CheckUserPermissionsInScope instead.
func (s *OrganizationMemberService) CheckUserPermissions(ctx context.Context, userID uuid.UUID, permissions []string) (map[string]bool, error) {
	return s.orgMemberRepo.CheckUserPermissions(ctx, userID, permissions)
}

// GetActiveMembers gets all active members of an organization
func (s *OrganizationMemberService) GetActiveMembers(ctx context.Context, orgID uuid.UUID) ([]*authDomain.OrganizationMember, error) {
	return s.orgMemberRepo.GetActiveMembers(ctx, orgID)
}

// GetMemberCount gets the count of members in an organization
func (s *OrganizationMemberService) GetMemberCount(ctx context.Context, orgID uuid.UUID) (int, error) {
	return s.orgMemberRepo.GetMemberCount(ctx, orgID)
}

// GetMembersByRoleCount gets member counts by role in an organization
func (s *OrganizationMemberService) GetMembersByRoleCount(ctx context.Context, orgID uuid.UUID) (map[string]int, error) {
	return s.orgMemberRepo.GetMembersByRole(ctx, orgID)
}
