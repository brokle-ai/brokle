package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"brokle/internal/core/domain/auth"
	"brokle/internal/core/domain/user"
	"brokle/pkg/ulid"
)

// userService implements the user.UserService interface
type userService struct {
	userRepo    user.Repository
	authService auth.AuthService
	auditRepo   auth.AuditLogRepository
}

// NewUserService creates a new user service instance
func NewUserService(
	userRepo user.Repository,
	authService auth.AuthService,
	auditRepo auth.AuditLogRepository,
) user.UserService {
	return &userService{
		userRepo:    userRepo,
		authService: authService,
		auditRepo:   auditRepo,
	}
}


// GetUser retrieves user by ID
func (s *userService) GetUser(ctx context.Context, userID ulid.ULID) (*user.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

// GetUserByEmail retrieves user by email (without password)
func (s *userService) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {
	return s.userRepo.GetByEmail(ctx, email)
}

// GetUserByEmailWithPassword retrieves user by email with password for authentication
func (s *userService) GetUserByEmailWithPassword(ctx context.Context, email string) (*user.User, error) {
	return s.userRepo.GetByEmailWithPassword(ctx, email)
}

// UpdateUser updates user information
func (s *userService) UpdateUser(ctx context.Context, userID ulid.ULID, req *user.UpdateUserRequest) (*user.User, error) {
	existingUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Update fields if provided
	if req.FirstName != nil {
		existingUser.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		existingUser.LastName = *req.LastName
	}
	if req.Timezone != nil {
		existingUser.Timezone = *req.Timezone
	}
	if req.Language != nil {
		existingUser.Language = *req.Language
	}

	existingUser.UpdatedAt = time.Now()

	err = s.userRepo.Update(ctx, existingUser)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(&userID, nil, "user.updated", "user", userID.String(), "", "", ""))

	return existingUser, nil
}

// DeactivateUser deactivates a user account
func (s *userService) DeactivateUser(ctx context.Context, userID ulid.ULID) error {
	existingUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	existingUser.IsActive = false
	existingUser.UpdatedAt = time.Now()

	err = s.userRepo.Update(ctx, existingUser)
	if err != nil {
		return fmt.Errorf("failed to deactivate user: %w", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(&userID, nil, "user.deactivated", "user", userID.String(), "", "", ""))

	return nil
}

// ReactivateUser reactivates a deactivated user account
func (s *userService) ReactivateUser(ctx context.Context, userID ulid.ULID) error {
	existingUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	existingUser.IsActive = true
	existingUser.UpdatedAt = time.Now()

	err = s.userRepo.Update(ctx, existingUser)
	if err != nil {
		return fmt.Errorf("failed to reactivate user: %w", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(&userID, nil, "user.reactivated", "user", userID.String(), "", "", ""))

	return nil
}

// DeleteUser soft deletes a user account
func (s *userService) DeleteUser(ctx context.Context, userID ulid.ULID) error {
	existingUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	err = s.userRepo.Delete(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(&userID, nil, "user.deleted", "user", userID.String(),
		fmt.Sprintf(`{"email": "%s"}`, existingUser.Email), "", ""))

	return nil
}

// ListUsers retrieves users with pagination and filters
func (s *userService) ListUsers(ctx context.Context, filters *user.ListFilters) ([]*user.User, int, error) {
	users, total, err := s.userRepo.List(ctx, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, total, nil
}

// SearchUsers searches for users by query  
func (s *userService) SearchUsers(ctx context.Context, query string, limit, offset int) ([]*user.User, int, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []*user.User{}, 0, nil
	}

	// For now, use List with basic filters since Search may not be implemented
	filters := &user.ListFilters{
		Limit:  limit,
		Offset: offset,
	}
	users, total, err := s.userRepo.List(ctx, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search users: %w", err)
	}

	return users, total, nil
}

// GetUsersByIDs retrieves multiple users by their IDs
func (s *userService) GetUsersByIDs(ctx context.Context, userIDs []ulid.ULID) ([]*user.User, error) {
	if len(userIDs) == 0 {
		return []*user.User{}, nil
	}

	return s.userRepo.GetByIDs(ctx, userIDs)
}

// GetPublicUsers retrieves public user information by IDs
func (s *userService) GetPublicUsers(ctx context.Context, userIDs []ulid.ULID) ([]*user.PublicUser, error) {
	users, err := s.GetUsersByIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	publicUsers := make([]*user.PublicUser, len(users))
	for i, u := range users {
		publicUsers[i] = u.ToPublic()
	}

	return publicUsers, nil
}

// VerifyEmail verifies user's email with token
func (s *userService) VerifyEmail(ctx context.Context, userID ulid.ULID, token string) error {
	// This would typically validate the token and mark email as verified
	existingUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	now := time.Now()
	existingUser.IsEmailVerified = true
	existingUser.EmailVerifiedAt = &now
	existingUser.UpdatedAt = now

	err = s.userRepo.Update(ctx, existingUser)
	if err != nil {
		return fmt.Errorf("failed to verify email: %w", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(&userID, nil, "user.email_verified", "user", userID.String(), "", "", ""))

	return nil
}

// MarkEmailAsVerified directly marks user's email as verified
func (s *userService) MarkEmailAsVerified(ctx context.Context, userID ulid.ULID) error {
	existingUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	now := time.Now()
	existingUser.IsEmailVerified = true
	existingUser.EmailVerifiedAt = &now
	existingUser.UpdatedAt = now

	return s.userRepo.Update(ctx, existingUser)
}

// SendVerificationEmail sends email verification email
func (s *userService) SendVerificationEmail(ctx context.Context, userID ulid.ULID) error {
	// This would integrate with email service to send verification email
	// For now, just create audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(&userID, nil, "user.verification_email_sent", "user", userID.String(), "", "", ""))
	return nil
}

// RequestPasswordReset initiates password reset process
func (s *userService) RequestPasswordReset(ctx context.Context, email string) error {
	existingUser, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Don't reveal if email exists or not for security
		return nil
	}

	// This would generate reset token and send email
	// For now, just create audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(&existingUser.ID, nil, "user.password_reset_requested", "user", existingUser.ID.String(), "", "", ""))
	return nil
}

// ResetPassword resets user password with token
func (s *userService) ResetPassword(ctx context.Context, token, newPassword string) error {
	// This would validate token and update password
	// Implementation would need token validation logic
	return nil
}

// ChangePassword changes user password
func (s *userService) ChangePassword(ctx context.Context, userID ulid.ULID, currentPassword, newPassword string) error {
	existingUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Verify current password
	err = bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(currentPassword))
	if err != nil {
		return errors.New("current password is incorrect")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	existingUser.Password = string(hashedPassword)
	existingUser.UpdatedAt = time.Now()

	err = s.userRepo.Update(ctx, existingUser)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(&userID, nil, "user.password_changed", "user", userID.String(), "", "", ""))

	return nil
}

// UpdateLastLogin updates user's last login time
func (s *userService) UpdateLastLogin(ctx context.Context, userID ulid.ULID) error {
	existingUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	now := time.Now()
	existingUser.LastLoginAt = &now
	existingUser.LoginCount++
	existingUser.UpdatedAt = now

	return s.userRepo.Update(ctx, existingUser)
}

// GetUserActivity retrieves user activity metrics
func (s *userService) GetUserActivity(ctx context.Context, userID ulid.ULID) (*user.UserActivity, error) {
	// This would aggregate activity data from various sources
	// For now, return basic activity
	existingUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	activity := &user.UserActivity{
		UserID:           userID,
		TotalLogins:      0, // Would be calculated from sessions
		DashboardViews:   0, // Would be calculated from analytics
		APIRequestsCount: 0, // Would be calculated from API logs
		CreatedProjects:  0, // Would be calculated from projects
		JoinedOrgs:       0, // Would be calculated from organization memberships
	}

	if existingUser.LastLoginAt != nil {
		lastLogin := existingUser.LastLoginAt.Format(time.RFC3339)
		activity.LastLoginAt = &lastLogin
	}

	return activity, nil
}

// SetDefaultOrganization sets user's default organization
func (s *userService) SetDefaultOrganization(ctx context.Context, userID, orgID ulid.ULID) error {
	existingUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	existingUser.DefaultOrganizationID = &orgID
	return s.userRepo.Update(ctx, existingUser)
}

// GetDefaultOrganization gets user's default organization
func (s *userService) GetDefaultOrganization(ctx context.Context, userID ulid.ULID) (*ulid.ULID, error) {
	existingUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return existingUser.DefaultOrganizationID, nil
}

// GetUserStats retrieves aggregate user statistics
func (s *userService) GetUserStats(ctx context.Context) (*user.UserStats, error) {
	// This would aggregate statistics from the database
	// For now, return basic stats structure
	return &user.UserStats{
		TotalUsers:        0, // Would be calculated
		ActiveUsers:       0, // Would be calculated
		VerifiedUsers:     0, // Would be calculated
		NewUsersToday:     0, // Would be calculated
		NewUsersThisWeek:  0, // Would be calculated
		NewUsersThisMonth: 0, // Would be calculated
	}, nil
}