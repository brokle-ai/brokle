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

// invitationService implements the organization.InvitationService interface
type invitationService struct {
	inviteRepo  organization.InvitationRepository
	orgRepo     organization.OrganizationRepository
	memberRepo  organization.MemberRepository
	userRepo    user.Repository
	roleService auth.RoleService
	auditRepo   auth.AuditLogRepository
}

// NewInvitationService creates a new invitation service instance
func NewInvitationService(
	inviteRepo organization.InvitationRepository,
	orgRepo organization.OrganizationRepository,
	memberRepo organization.MemberRepository,
	userRepo user.Repository,
	roleService auth.RoleService,
	auditRepo auth.AuditLogRepository,
) organization.InvitationService {
	return &invitationService{
		inviteRepo:  inviteRepo,
		orgRepo:     orgRepo,
		memberRepo:  memberRepo,
		userRepo:    userRepo,
		roleService: roleService,
		auditRepo:   auditRepo,
	}
}

// InviteUser creates an invitation for a user to join an organization
func (s *invitationService) InviteUser(ctx context.Context, orgID ulid.ULID, inviterID ulid.ULID, req *organization.InviteUserRequest) (*organization.Invitation, error) {
	// Verify organization exists
	_, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	// Verify role exists
	role, err := s.roleService.GetRole(ctx, req.RoleID)
	if err != nil {
		return nil, fmt.Errorf("role not found: %w", err)
	}

	// Check for existing pending invitation
	if req.Email != "" {
		existing, _ := s.inviteRepo.GetPendingByEmail(ctx, orgID, req.Email)
		if existing != nil {
			return nil, errors.New("invitation already exists for this email")
		}
	}

	// Create invitation
	token := ulid.New().String()
	expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days
	invitation := organization.NewInvitation(orgID, req.RoleID, inviterID, req.Email, token, expiresAt)
	err = s.inviteRepo.Create(ctx, invitation)
	if err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(&inviterID, &orgID, "invitation.created", "invitation", invitation.ID.String(),
		fmt.Sprintf(`{"email": "%s", "role": "%s"}`, req.Email, role.Name), "", ""))

	return invitation, nil
}

// AcceptInvitation accepts an invitation and adds the user to the organization
func (s *invitationService) AcceptInvitation(ctx context.Context, token string, userID ulid.ULID) error {
	// Get invitation by token
	invitation, err := s.inviteRepo.GetByToken(ctx, token)
	if err != nil {
		return fmt.Errorf("invitation not found: %w", err)
	}

	if invitation.Status != organization.InvitationStatusPending {
		return errors.New("invitation is not pending")
	}

	if invitation.ExpiresAt.Before(time.Now()) {
		return errors.New("invitation has expired")
	}

	// Check if user is already a member
	isMember, err := s.memberRepo.IsMember(ctx, userID, invitation.OrganizationID)
	if err != nil {
		return fmt.Errorf("failed to check membership: %w", err)
	}
	if isMember {
		return errors.New("user is already a member of this organization")
	}

	// Add user as member
	member := organization.NewMember(invitation.OrganizationID, userID, invitation.RoleID)
	err = s.memberRepo.Create(ctx, member)
	if err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}

	// Mark invitation as accepted
	invitation.Status = organization.InvitationStatusAccepted
	invitation.AcceptedAt = &time.Time{}
	*invitation.AcceptedAt = time.Now()
	invitation.UpdatedAt = time.Now()
	err = s.inviteRepo.Update(ctx, invitation)
	if err != nil {
		return fmt.Errorf("failed to update invitation: %w", err)
	}

	// Set as default organization if user doesn't have one
	user, _ := s.userRepo.GetByID(ctx, userID)
	if user != nil && user.DefaultOrganizationID == nil {
		err = s.userRepo.SetDefaultOrganization(ctx, userID, invitation.OrganizationID)
		if err != nil {
			fmt.Printf("Failed to set default organization: %v\n", err)
		}
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(&userID, &invitation.OrganizationID, "invitation.accepted", "invitation", invitation.ID.String(),
		fmt.Sprintf(`{"user_id": "%s"}`, userID.String()), "", ""))

	return nil
}

// DeclineInvitation declines an invitation
func (s *invitationService) DeclineInvitation(ctx context.Context, token string) error {
	// Get invitation by token
	invitation, err := s.inviteRepo.GetByToken(ctx, token)
	if err != nil {
		return fmt.Errorf("invitation not found: %w", err)
	}

	if invitation.Status != organization.InvitationStatusPending {
		return errors.New("invitation is not pending")
	}

	// Mark invitation as declined
	invitation.Status = organization.InvitationStatusRevoked
	invitation.UpdatedAt = time.Now()
	err = s.inviteRepo.Update(ctx, invitation)
	if err != nil {
		return fmt.Errorf("failed to update invitation: %w", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(nil, &invitation.OrganizationID, "invitation.declined", "invitation", invitation.ID.String(), "", "", ""))

	return nil
}

// RevokeInvitation revokes a pending invitation
func (s *invitationService) RevokeInvitation(ctx context.Context, invitationID ulid.ULID, revokedByID ulid.ULID) error {
	// Get invitation
	invitation, err := s.inviteRepo.GetByID(ctx, invitationID)
	if err != nil {
		return fmt.Errorf("invitation not found: %w", err)
	}

	if invitation.Status != organization.InvitationStatusPending {
		return errors.New("invitation is not pending")
	}

	// Mark invitation as revoked
	invitation.Status = organization.InvitationStatusRevoked
	invitation.UpdatedAt = time.Now()
	err = s.inviteRepo.Update(ctx, invitation)
	if err != nil {
		return fmt.Errorf("failed to update invitation: %w", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(&revokedByID, &invitation.OrganizationID, "invitation.revoked", "invitation", invitationID.String(), "", "", ""))

	return nil
}

// ResendInvitation resends a pending invitation
func (s *invitationService) ResendInvitation(ctx context.Context, invitationID ulid.ULID, resentByID ulid.ULID) error {
	// Get invitation
	invitation, err := s.inviteRepo.GetByID(ctx, invitationID)
	if err != nil {
		return fmt.Errorf("invitation not found: %w", err)
	}

	if invitation.Status != organization.InvitationStatusPending {
		return errors.New("invitation is not pending")
	}

	// Update expiration time
	invitation.ExpiresAt = time.Now().Add(7 * 24 * time.Hour) // 7 days from now
	invitation.UpdatedAt = time.Now()
	err = s.inviteRepo.Update(ctx, invitation)
	if err != nil {
		return fmt.Errorf("failed to update invitation: %w", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(&resentByID, &invitation.OrganizationID, "invitation.resent", "invitation", invitationID.String(), "", "", ""))

	return nil
}

// GetInvitation retrieves an invitation by ID
func (s *invitationService) GetInvitation(ctx context.Context, invitationID ulid.ULID) (*organization.Invitation, error) {
	return s.inviteRepo.GetByID(ctx, invitationID)
}

// GetInvitationByToken retrieves an invitation by token
func (s *invitationService) GetInvitationByToken(ctx context.Context, token string) (*organization.Invitation, error) {
	return s.inviteRepo.GetByToken(ctx, token)
}

// GetPendingInvitations retrieves all pending invitations for an organization
func (s *invitationService) GetPendingInvitations(ctx context.Context, orgID ulid.ULID) ([]*organization.Invitation, error) {
	allInvitations, err := s.inviteRepo.GetByOrganizationID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	var pendingInvitations []*organization.Invitation
	for _, invitation := range allInvitations {
		if invitation.Status == organization.InvitationStatusPending {
			pendingInvitations = append(pendingInvitations, invitation)
		}
	}

	return pendingInvitations, nil
}

// GetUserInvitations retrieves all invitations for a user by email
func (s *invitationService) GetUserInvitations(ctx context.Context, email string) ([]*organization.Invitation, error) {
	return s.inviteRepo.GetByEmail(ctx, email)
}

// ValidateInvitationToken validates an invitation token and returns the invitation
func (s *invitationService) ValidateInvitationToken(ctx context.Context, token string) (*organization.Invitation, error) {
	invitation, err := s.inviteRepo.GetByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("invitation not found: %w", err)
	}

	if invitation.Status != organization.InvitationStatusPending {
		return nil, errors.New("invitation is not pending")
	}

	if invitation.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("invitation has expired")
	}

	return invitation, nil
}

// IsEmailAlreadyInvited checks if an email already has a pending invitation for an organization
func (s *invitationService) IsEmailAlreadyInvited(ctx context.Context, email string, orgID ulid.ULID) (bool, error) {
	existing, err := s.inviteRepo.GetPendingByEmail(ctx, orgID, email)
	if err != nil {
		return false, err
	}
	return existing != nil, nil
}

// CleanupExpiredInvitations removes expired invitations
func (s *invitationService) CleanupExpiredInvitations(ctx context.Context) error {
	// This would typically be called by a background job
	// For now, return nil as the repository doesn't have a bulk delete method
	// This could be implemented by getting all invitations and filtering/deleting expired ones
	return nil
}