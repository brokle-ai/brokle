package auth

import (
	"context"

	"github.com/google/uuid"

	authDomain "brokle/internal/core/domain/auth"
	appErrors "brokle/pkg/errors"
)

// OrganizationMemberService assigns roles to organization members and checks access.
type OrganizationMemberService struct {
	orgMemberRepo authDomain.OrganizationMemberRepository
	roleRepo      authDomain.RoleRepository
}

// NewOrganizationMemberService creates a new organization member service instance
func NewOrganizationMemberService(
	orgMemberRepo authDomain.OrganizationMemberRepository,
	roleRepo authDomain.RoleRepository,
) *OrganizationMemberService {
	return &OrganizationMemberService{
		orgMemberRepo: orgMemberRepo,
		roleRepo:      roleRepo,
	}
}

// AddMember adds a user to an organization with specified role
func (s *OrganizationMemberService) AddMember(ctx context.Context, userID, orgID, roleID uuid.UUID, invitedBy *uuid.UUID) (*authDomain.OrganizationMember, error) {
	// Verify role exists
	_, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return nil, appErrors.NewNotFoundError("role not found")
	}

	// Check if user is already a member
	exists, err := s.orgMemberRepo.Exists(ctx, userID, orgID)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to check membership", err)
	}
	if exists {
		return nil, appErrors.NewConflictError("user is already a member of this organization")
	}

	// Create new membership
	member := authDomain.NewOrganizationMember(userID, orgID, roleID, invitedBy)

	err = s.orgMemberRepo.Create(ctx, member)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to create membership", err)
	}

	return member, nil
}

// RemoveMember removes a user from an organization
func (s *OrganizationMemberService) RemoveMember(ctx context.Context, userID, orgID uuid.UUID) error {
	// Check if user is a member
	exists, err := s.orgMemberRepo.Exists(ctx, userID, orgID)
	if err != nil {
		return appErrors.NewInternalError("failed to check membership", err)
	}
	if !exists {
		return appErrors.NewNotFoundError("user is not a member of this organization")
	}

	return s.orgMemberRepo.Delete(ctx, userID, orgID)
}

// UpdateMemberRole updates a member's role in an organization
func (s *OrganizationMemberService) UpdateMemberRole(ctx context.Context, userID, orgID, roleID uuid.UUID) error {
	// Verify role exists
	_, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return appErrors.NewNotFoundError("role not found")
	}

	// Check if user is a member
	exists, err := s.orgMemberRepo.Exists(ctx, userID, orgID)
	if err != nil {
		return appErrors.NewInternalError("failed to check membership", err)
	}
	if !exists {
		return appErrors.NewNotFoundError("user is not a member of this organization")
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

// ActivateMember activates a member in an organization
func (s *OrganizationMemberService) ActivateMember(ctx context.Context, userID, orgID uuid.UUID) error {
	return s.orgMemberRepo.ActivateMember(ctx, userID, orgID)
}

// SuspendMember suspends a member in an organization
func (s *OrganizationMemberService) SuspendMember(ctx context.Context, userID, orgID uuid.UUID) error {
	return s.orgMemberRepo.SuspendMember(ctx, userID, orgID)
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
