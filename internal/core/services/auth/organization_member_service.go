package auth

import (
	"context"
	"fmt"

	"brokle/internal/core/domain/auth"
	"brokle/pkg/ulid"
)

// organizationMemberService implements the auth.OrganizationMemberService interface
type organizationMemberService struct {
	orgMemberRepo auth.OrganizationMemberRepository
	roleRepo      auth.RoleRepository
	auditRepo     auth.AuditLogRepository
}

// NewOrganizationMemberService creates a new organization member service instance
func NewOrganizationMemberService(
	orgMemberRepo auth.OrganizationMemberRepository,
	roleRepo auth.RoleRepository,
	auditRepo auth.AuditLogRepository,
) auth.OrganizationMemberService {
	return &organizationMemberService{
		orgMemberRepo: orgMemberRepo,
		roleRepo:      roleRepo,
		auditRepo:     auditRepo,
	}
}

// AddMember adds a user to an organization with specified role
func (s *organizationMemberService) AddMember(ctx context.Context, userID, orgID, roleID ulid.ULID, invitedBy *ulid.ULID) (*auth.OrganizationMember, error) {
	// Verify role exists
	_, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("role not found: %w", err)
	}

	// Check if user is already a member
	exists, err := s.orgMemberRepo.Exists(ctx, userID, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("user is already a member of this organization")
	}

	// Create new membership
	member := auth.NewOrganizationMember(userID, orgID, roleID, invitedBy)
	
	err = s.orgMemberRepo.Create(ctx, member)
	if err != nil {
		return nil, fmt.Errorf("failed to create membership: %w", err)
	}

	return member, nil
}

// RemoveMember removes a user from an organization
func (s *organizationMemberService) RemoveMember(ctx context.Context, userID, orgID ulid.ULID) error {
	// Check if user is a member
	exists, err := s.orgMemberRepo.Exists(ctx, userID, orgID)
	if err != nil {
		return fmt.Errorf("failed to check membership: %w", err)
	}
	if !exists {
		return fmt.Errorf("user is not a member of this organization")
	}

	return s.orgMemberRepo.Delete(ctx, userID, orgID)
}

// UpdateMemberRole updates a member's role in an organization
func (s *organizationMemberService) UpdateMemberRole(ctx context.Context, userID, orgID, roleID ulid.ULID) error {
	// Verify role exists
	_, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// Check if user is a member
	exists, err := s.orgMemberRepo.Exists(ctx, userID, orgID)
	if err != nil {
		return fmt.Errorf("failed to check membership: %w", err)
	}
	if !exists {
		return fmt.Errorf("user is not a member of this organization")
	}

	return s.orgMemberRepo.UpdateMemberRole(ctx, userID, orgID, roleID)
}

// GetMember gets a specific organization membership
func (s *organizationMemberService) GetMember(ctx context.Context, userID, orgID ulid.ULID) (*auth.OrganizationMember, error) {
	return s.orgMemberRepo.GetByUserAndOrganization(ctx, userID, orgID)
}

// GetUserMemberships gets all organization memberships for a user
func (s *organizationMemberService) GetUserMemberships(ctx context.Context, userID ulid.ULID) ([]*auth.OrganizationMember, error) {
	return s.orgMemberRepo.GetByUserID(ctx, userID)
}

// GetOrganizationMembers gets all members of an organization
func (s *organizationMemberService) GetOrganizationMembers(ctx context.Context, orgID ulid.ULID) ([]*auth.OrganizationMember, error) {
	return s.orgMemberRepo.GetByOrganizationID(ctx, orgID)
}

// GetMembersByRole gets all members with a specific role
func (s *organizationMemberService) GetMembersByRole(ctx context.Context, roleID ulid.ULID) ([]*auth.OrganizationMember, error) {
	return s.orgMemberRepo.GetByRole(ctx, roleID)
}

// IsMember checks if a user is a member of an organization
func (s *organizationMemberService) IsMember(ctx context.Context, userID, orgID ulid.ULID) (bool, error) {
	return s.orgMemberRepo.Exists(ctx, userID, orgID)
}

// GetUserEffectivePermissions gets all effective permissions for a user across all organizations
func (s *organizationMemberService) GetUserEffectivePermissions(ctx context.Context, userID ulid.ULID) ([]string, error) {
	return s.orgMemberRepo.GetUserEffectivePermissions(ctx, userID)
}

// GetUserPermissionsInOrganization gets user permissions within a specific organization
func (s *organizationMemberService) GetUserPermissionsInOrganization(ctx context.Context, userID, orgID ulid.ULID) ([]string, error) {
	return s.orgMemberRepo.GetUserPermissionsInOrganization(ctx, userID, orgID)
}

// CheckUserPermission checks if a user has a specific permission
func (s *organizationMemberService) CheckUserPermission(ctx context.Context, userID ulid.ULID, permission string) (bool, error) {
	return s.orgMemberRepo.HasUserPermission(ctx, userID, permission)
}

// CheckUserPermissions checks multiple permissions for a user
func (s *organizationMemberService) CheckUserPermissions(ctx context.Context, userID ulid.ULID, permissions []string) (map[string]bool, error) {
	return s.orgMemberRepo.CheckUserPermissions(ctx, userID, permissions)
}

// ActivateMember activates a member in an organization
func (s *organizationMemberService) ActivateMember(ctx context.Context, userID, orgID ulid.ULID) error {
	return s.orgMemberRepo.ActivateMember(ctx, userID, orgID)
}

// SuspendMember suspends a member in an organization
func (s *organizationMemberService) SuspendMember(ctx context.Context, userID, orgID ulid.ULID) error {
	return s.orgMemberRepo.SuspendMember(ctx, userID, orgID)
}

// GetActiveMembers gets all active members of an organization
func (s *organizationMemberService) GetActiveMembers(ctx context.Context, orgID ulid.ULID) ([]*auth.OrganizationMember, error) {
	return s.orgMemberRepo.GetActiveMembers(ctx, orgID)
}

// GetMemberCount gets the count of members in an organization
func (s *organizationMemberService) GetMemberCount(ctx context.Context, orgID ulid.ULID) (int, error) {
	return s.orgMemberRepo.GetMemberCount(ctx, orgID)
}

// GetMembersByRoleCount gets member counts by role in an organization
func (s *organizationMemberService) GetMembersByRoleCount(ctx context.Context, orgID ulid.ULID) (map[string]int, error) {
	return s.orgMemberRepo.GetMembersByRole(ctx, orgID)
}