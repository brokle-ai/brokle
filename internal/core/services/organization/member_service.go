package organization

import (
	"context"
	"errors"
	"fmt"
	"time"

	"brokle/internal/core/domain/auth"
	"brokle/internal/core/domain/organization"
	"brokle/internal/core/domain/user"
	"brokle/pkg/ulid"
)

// memberService implements the organization.MemberService interface
type memberService struct {
	memberRepo  organization.MemberRepository
	orgRepo     organization.OrganizationRepository
	userRepo    user.Repository
	roleService auth.RoleService
	auditRepo   auth.AuditLogRepository
}

// NewMemberService creates a new member service instance
func NewMemberService(
	memberRepo organization.MemberRepository,
	orgRepo organization.OrganizationRepository,
	userRepo user.Repository,
	roleService auth.RoleService,
	auditRepo auth.AuditLogRepository,
) organization.MemberService {
	return &memberService{
		memberRepo:  memberRepo,
		orgRepo:     orgRepo,
		userRepo:    userRepo,
		roleService: roleService,
		auditRepo:   auditRepo,
	}
}

// AddMember adds a user to an organization with specified role
func (s *memberService) AddMember(ctx context.Context, orgID, userID, roleID ulid.ULID, addedByID ulid.ULID) error {
	// Verify organization exists
	_, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return fmt.Errorf("organization not found: %w", err)
	}

	// Verify user exists
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Verify role exists
	role, err := s.roleService.GetRoleByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// Check if user is already a member
	isMember, err := s.memberRepo.IsMember(ctx, userID, orgID)
	if err != nil {
		return fmt.Errorf("failed to check membership: %w", err)
	}
	if isMember {
		return errors.New("user is already a member of this organization")
	}

	// Create member
	member := organization.NewMember(orgID, userID, roleID)
	err = s.memberRepo.Create(ctx, member)
	if err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(&addedByID, &orgID, "member.added", "organization", orgID.String(),
		fmt.Sprintf(`{"user_email": "%s", "role": "%s"}`, user.Email, role.Name), "", ""))

	return nil
}

// RemoveMember removes a user from an organization
func (s *memberService) RemoveMember(ctx context.Context, orgID, userID ulid.ULID, removedByID ulid.ULID) error {
	// Verify membership exists
	member, err := s.memberRepo.GetByUserAndOrganization(ctx, userID, orgID)
	if err != nil {
		return fmt.Errorf("member not found: %w", err)
	}

	// Check if this is the only owner
	ownerRole, err := s.roleService.GetRoleByNameAndScope(ctx, "owner", auth.ScopeOrganization, &orgID)
	if err != nil {
		return fmt.Errorf("failed to get owner role: %w", err)
	}

	if member.RoleID == ownerRole.ID {
		ownerCount, err := s.memberRepo.CountByOrganizationAndRole(ctx, orgID, ownerRole.ID)
		if err != nil {
			return fmt.Errorf("failed to count owners: %w", err)
		}
		if ownerCount <= 1 {
			return errors.New("cannot remove the last owner of the organization")
		}
	}

	// Remove member
	err = s.memberRepo.Delete(ctx, member.ID)
	if err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	// If this was their default organization, clear it
	user, err := s.userRepo.GetByID(ctx, userID)
	if err == nil && user.DefaultOrganizationID != nil && *user.DefaultOrganizationID == orgID {
		err = s.userRepo.SetDefaultOrganization(ctx, userID, ulid.ULID{})
		if err != nil {
			fmt.Printf("Failed to clear default organization: %v\n", err)
		}
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(&removedByID, &orgID, "member.removed", "organization", orgID.String(),
		fmt.Sprintf(`{"user_id": "%s"}`, userID.String()), "", ""))

	return nil
}

// UpdateMemberRole updates a member's role in an organization
func (s *memberService) UpdateMemberRole(ctx context.Context, orgID, userID, newRoleID ulid.ULID, updatedByID ulid.ULID) error {
	// Verify membership exists
	member, err := s.memberRepo.GetByUserAndOrganization(ctx, userID, orgID)
	if err != nil {
		return fmt.Errorf("member not found: %w", err)
	}

	// Verify new role exists
	newRole, err := s.roleService.GetRoleByID(ctx, newRoleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// Check if demoting the last owner
	ownerRole, err := s.roleService.GetRoleByNameAndScope(ctx, "owner", auth.ScopeOrganization, &orgID)
	if err != nil {
		return fmt.Errorf("failed to get owner role: %w", err)
	}

	if member.RoleID == ownerRole.ID && newRoleID != ownerRole.ID {
		ownerCount, err := s.memberRepo.CountByOrganizationAndRole(ctx, orgID, ownerRole.ID)
		if err != nil {
			return fmt.Errorf("failed to count owners: %w", err)
		}
		if ownerCount <= 1 {
			return errors.New("cannot demote the last owner of the organization")
		}
	}

	// Update role
	member.RoleID = newRoleID
	member.UpdatedAt = time.Now()
	err = s.memberRepo.Update(ctx, member)
	if err != nil {
		return fmt.Errorf("failed to update member role: %w", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(&updatedByID, &orgID, "member.role_updated", "organization", orgID.String(),
		fmt.Sprintf(`{"user_id": "%s", "new_role": "%s"}`, userID.String(), newRole.Name), "", ""))

	return nil
}

// GetMember retrieves a specific member
func (s *memberService) GetMember(ctx context.Context, orgID, userID ulid.ULID) (*organization.Member, error) {
	return s.memberRepo.GetByUserAndOrganization(ctx, userID, orgID)
}

// GetMembers retrieves all members of an organization
func (s *memberService) GetMembers(ctx context.Context, orgID ulid.ULID) ([]*organization.Member, error) {
	return s.memberRepo.GetByOrganizationID(ctx, orgID)
}

// IsMember checks if a user is a member of an organization
func (s *memberService) IsMember(ctx context.Context, userID, orgID ulid.ULID) (bool, error) {
	return s.memberRepo.IsMember(ctx, userID, orgID)
}

// CanUserAccessOrganization checks if user can access organization
func (s *memberService) CanUserAccessOrganization(ctx context.Context, userID, orgID ulid.ULID) (bool, error) {
	return s.memberRepo.IsMember(ctx, userID, orgID)
}

// GetUserRole returns a user's role ID in an organization
func (s *memberService) GetUserRole(ctx context.Context, userID, orgID ulid.ULID) (ulid.ULID, error) {
	member, err := s.memberRepo.GetByUserAndOrganization(ctx, userID, orgID)
	if err != nil {
		return ulid.ULID{}, fmt.Errorf("member not found: %w", err)
	}

	return member.RoleID, nil
}

// GetMemberCount returns the number of members in an organization
func (s *memberService) GetMemberCount(ctx context.Context, orgID ulid.ULID) (int, error) {
	members, err := s.memberRepo.GetByOrganizationID(ctx, orgID)
	if err != nil {
		return 0, fmt.Errorf("failed to get members: %w", err)
	}
	return len(members), nil
}

// GetMembersByRole returns all members with a specific role
func (s *memberService) GetMembersByRole(ctx context.Context, orgID, roleID ulid.ULID) ([]*organization.Member, error) {
	allMembers, err := s.memberRepo.GetByOrganizationID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get members: %w", err)
	}

	var membersWithRole []*organization.Member
	for _, member := range allMembers {
		if member.RoleID == roleID {
			membersWithRole = append(membersWithRole, member)
		}
	}

	return membersWithRole, nil
}