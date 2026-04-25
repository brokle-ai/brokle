package user

import (
	"context"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"

	authDomain "brokle/internal/core/domain/auth"
	userDomain "brokle/internal/core/domain/user"
	authService "brokle/internal/core/services/auth"
	appErrors "brokle/pkg/errors"
)

// UserService manages identity-adjacent user state — basic profile
// fields, password lifecycle, default-organization pointer, and
// last-login bookkeeping. Profile details (bio, social links,
// completeness scoring) live in ProfileService.
type UserService struct {
	userRepo      userDomain.Repository
	auth          *authService.AuthService
	orgMemberRepo authDomain.OrganizationMemberRepository
}

// NewUserService creates a new user service instance.
func NewUserService(
	userRepo userDomain.Repository,
	auth *authService.AuthService,
	orgMemberRepo authDomain.OrganizationMemberRepository,
) *UserService {
	return &UserService{
		userRepo:      userRepo,
		auth:          auth,
		orgMemberRepo: orgMemberRepo,
	}
}

// GetUser retrieves a user by ID.
func (s *UserService) GetUser(ctx context.Context, userID uuid.UUID) (*userDomain.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

// GetUserByEmail retrieves a user by email (without the password column).
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*userDomain.User, error) {
	return s.userRepo.GetByEmail(ctx, email)
}

// UpdateUser applies a partial update to the user record. Pointer
// fields on the request mean "only update if provided".
func (s *UserService) UpdateUser(ctx context.Context, userID uuid.UUID, req *userDomain.UpdateUserRequest) (*userDomain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, userDomain.ErrNotFound) {
			return nil, appErrors.NewNotFoundError("user not found")
		}
		return nil, appErrors.NewInternalError("user lookup failed", err)
	}

	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.Timezone != nil {
		user.Timezone = *req.Timezone
	}
	if req.Language != nil {
		user.Language = *req.Language
	}
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, appErrors.NewInternalError("failed to update user", err)
	}
	return user, nil
}

// ChangePassword verifies the user's current password and updates to
// the new one. Returns 401 on current-password mismatch.
func (s *UserService) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, userDomain.ErrNotFound) {
			return appErrors.NewNotFoundError("user not found")
		}
		return appErrors.NewInternalError("user lookup failed", err)
	}

	if !user.HasPassword() {
		return appErrors.NewUnauthorizedError("user has no password set")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(currentPassword)); err != nil {
		return appErrors.NewUnauthorizedError("current password is incorrect")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return appErrors.NewInternalError("failed to hash password", err)
	}
	user.SetPassword(string(hashedPassword))

	if err := s.userRepo.Update(ctx, user); err != nil {
		return appErrors.NewInternalError("failed to update password", err)
	}
	return nil
}

// ResetPassword is the user-service hook for the password-reset flow.
// The actual reset state machine is owned by AuthService; this entry
// point is reserved for the user-domain side of any future cross-
// service coordination.
func (s *UserService) ResetPassword(ctx context.Context, token, newPassword string) error {
	// Reserved: reset workflow is currently owned end-to-end by
	// auth.AuthService.ConfirmPasswordReset.
	return nil
}

// UpdateLastLogin updates the user's last_login_at timestamp and
// increments login_count. Called from AuthService.Login on the success
// path; failures are best-effort (the login still succeeded).
func (s *UserService) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, userDomain.ErrNotFound) {
			return appErrors.NewNotFoundError("user not found")
		}
		return appErrors.NewInternalError("user lookup failed", err)
	}

	now := time.Now()
	user.LastLoginAt = &now
	user.LoginCount++
	user.UpdatedAt = now

	return s.userRepo.Update(ctx, user)
}

// SetDefaultOrganization records the user's preferred default org so
// the dashboard lands on the right context after login.
func (s *UserService) SetDefaultOrganization(ctx context.Context, userID, orgID uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, userDomain.ErrNotFound) {
			return appErrors.NewNotFoundError("user not found")
		}
		return appErrors.NewInternalError("user lookup failed", err)
	}

	user.DefaultOrganizationID = &orgID
	return s.userRepo.Update(ctx, user)
}

// ValidateUserOrgMembership reports whether user is a member of org.
// Used by the dashboard's project-scope guard before exposing data
// from a different organization.
func (s *UserService) ValidateUserOrgMembership(ctx context.Context, userID, orgID uuid.UUID) (bool, error) {
	return s.orgMemberRepo.Exists(ctx, userID, orgID)
}
