package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"brokle/internal/core/domain/auth"
	"brokle/internal/core/domain/user"
	"brokle/pkg/ulid"
)

// authService implements the auth.AuthService interface
type authService struct {
	userRepo    user.Repository
	sessionRepo auth.SessionRepository
	auditRepo   auth.AuditLogRepository
	jwtService  auth.JWTService
	roleService auth.RoleService
}

// NewAuthService creates a new auth service instance
func NewAuthService(
	userRepo user.Repository,
	sessionRepo auth.SessionRepository,
	auditRepo auth.AuditLogRepository,
	jwtService auth.JWTService,
	roleService auth.RoleService,
) auth.AuthService {
	return &authService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		auditRepo:   auditRepo,
		jwtService:  jwtService,
		roleService: roleService,
	}
}

// Login authenticates a user and returns a login response
func (s *authService) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	// Get user with password
	user, err := s.userRepo.GetByEmailWithPassword(ctx, req.Email)
	if err != nil {
		// Log failed login attempt
		s.auditRepo.Create(ctx, &auth.AuditLog{
			ID:        ulid.New(),
			Action:    "auth.login.failed",
			Resource:  "user",
			Metadata:  fmt.Sprintf(`{"email": "%s", "reason": "user_not_found"}`, req.Email),
			CreatedAt: time.Now(),
		})
		return nil, errors.New("invalid credentials")
	}

	// Check if user is active
	if !user.IsActive {
		s.auditRepo.Create(ctx, &auth.AuditLog{
			ID:        ulid.New(),
			UserID:    &user.ID,
			Action:    "auth.login.failed",
			Resource:  "user",
			ResourceID: user.ID.String(),
			Metadata:  `{"reason": "user_inactive"}`,
			CreatedAt: time.Now(),
		})
		return nil, errors.New("account is inactive")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		s.auditRepo.Create(ctx, &auth.AuditLog{
			ID:        ulid.New(),
			UserID:    &user.ID,
			Action:    "auth.login.failed",
			Resource:  "user",
			ResourceID: user.ID.String(),
			Metadata:  `{"reason": "invalid_password"}`,
			CreatedAt: time.Now(),
		})
		return nil, errors.New("invalid credentials")
	}

	// Get user permissions for default organization
	var permissions []string
	if user.DefaultOrganizationID != nil {
		permissions, _ = s.roleService.GetUserPermissions(ctx, user.ID, *user.DefaultOrganizationID)
	}

	// Generate access token
	accessToken, err := s.jwtService.GenerateAccessToken(ctx, user.ID, map[string]interface{}{
		"email":          user.Email,
		"organization_id": user.DefaultOrganizationID,
		"permissions":     permissions,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Calculate token expiration - use default 24h for access, 7d for refresh
	accessTokenTTL := 24 * time.Hour
	refreshTokenTTL := 7 * 24 * time.Hour
	expiresAt := time.Now().Add(accessTokenTTL)
	refreshExpiresAt := time.Now().Add(refreshTokenTTL)

	// Create session
	session := auth.NewSession(user.ID, accessToken, refreshToken, expiresAt, refreshExpiresAt, nil, nil)
	err = s.sessionRepo.Create(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Update last login
	err = s.userRepo.UpdateLastLogin(ctx, user.ID)
	if err != nil {
		// Log but don't fail login
		fmt.Printf("Failed to update last login: %v\n", err)
	}

	// Log successful login
	s.auditRepo.Create(ctx, &auth.AuditLog{
		ID:             ulid.New(),
		UserID:         &user.ID,
		OrganizationID: user.DefaultOrganizationID,
		Action:         "auth.login.success",
		Resource:       "user",
		ResourceID:     user.ID.String(),
		Metadata:       fmt.Sprintf(`{"session_id": "%s"}`, session.ID.String()),
		CreatedAt:      time.Now(),
	})

	// Prepare response
	authUser := &auth.AuthUser{
		ID:         user.ID,
		Email:      user.Email,
		Name:                  user.GetFullName(),
		AvatarURL:  &user.AvatarURL,
		IsEmailVerified:       user.IsEmailVerified,
		OnboardingCompleted:   user.OnboardingCompleted,
		DefaultOrganizationID: user.DefaultOrganizationID,
	}

	return &auth.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(accessTokenTTL.Seconds()),
		User:         authUser,
	}, nil
}

// Register creates a new user account
func (s *authService) Register(ctx context.Context, req *auth.RegisterRequest) (*user.User, error) {
	// Check if user already exists
	existingUser, _ := s.userRepo.GetByEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, errors.New("user already exists with this email")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user using constructor
	user := user.NewUser(req.Email, req.FirstName, req.LastName)
	user.SetPassword(string(hashedPassword))

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Log registration
	s.auditRepo.Create(ctx, &auth.AuditLog{
		ID:        ulid.New(),
		UserID:    &user.ID,
		Action:    "auth.register.success",
		Resource:  "user",
		ResourceID: user.ID.String(),
		Metadata:  fmt.Sprintf(`{"email": "%s"}`, user.Email),
		CreatedAt: time.Now(),
	})

	return user, nil
}

// Logout invalidates a user session
func (s *authService) Logout(ctx context.Context, sessionID ulid.ULID) error {
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	err = s.sessionRepo.DeactivateSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	// Log logout
	s.auditRepo.Create(ctx, &auth.AuditLog{
		ID:        ulid.New(),
		UserID:    &session.UserID,
		Action:    "auth.logout.success",
		Resource:  "session",
		ResourceID: sessionID.String(),
		CreatedAt: time.Now(),
	})

	return nil
}

// RefreshToken generates new access token using refresh token
func (s *authService) RefreshToken(ctx context.Context, req *auth.RefreshTokenRequest) (*auth.LoginResponse, error) {
	// Validate refresh token
	claims, err := s.jwtService.ValidateRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Get session by refresh token
	session, err := s.sessionRepo.GetByRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if !session.IsActive {
		return nil, errors.New("session is inactive")
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if !user.IsActive {
		return nil, errors.New("user is inactive")
	}

	// Get user permissions
	var permissions []string
	if user.DefaultOrganizationID != nil {
		permissions, _ = s.roleService.GetUserPermissions(ctx, user.ID, *user.DefaultOrganizationID)
	}

	// Generate new access token
	accessToken, err := s.jwtService.GenerateAccessToken(ctx, user.ID, map[string]interface{}{
		"email":          user.Email,
		"organization_id": user.DefaultOrganizationID,
		"permissions":     permissions,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Update session with new access token
	accessTokenTTL := 24 * time.Hour
	session.Token = accessToken
	session.ExpiresAt = time.Now().Add(accessTokenTTL)
	session.UpdatedAt = time.Now()

	err = s.sessionRepo.Update(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	// Prepare response
	authUser := &auth.AuthUser{
		ID:         user.ID,
		Email:      user.Email,
		Name:                  user.GetFullName(),
		AvatarURL:  &user.AvatarURL,
		IsEmailVerified:       user.IsEmailVerified,
		OnboardingCompleted:   user.OnboardingCompleted,
		DefaultOrganizationID: user.DefaultOrganizationID,
	}

	return &auth.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: req.RefreshToken, // Keep same refresh token
		TokenType:    "Bearer",
		ExpiresIn:    int64(accessTokenTTL.Seconds()),
		User:         authUser,
	}, nil
}

// ChangePassword changes a user's password
func (s *authService) ChangePassword(ctx context.Context, userID ulid.ULID, currentPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Verify current password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword))
	if err != nil {
		return errors.New("current password is incorrect")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update password
	err = s.userRepo.UpdatePassword(ctx, userID, string(hashedPassword))
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Revoke all user sessions (force re-login)
	err = s.sessionRepo.DeactivateUserSessions(ctx, userID)
	if err != nil {
		// Log but don't fail
		fmt.Printf("Failed to revoke user sessions: %v\n", err)
	}

	// Log password change
	s.auditRepo.Create(ctx, &auth.AuditLog{
		ID:        ulid.New(),
		UserID:    &userID,
		Action:    "auth.password.changed",
		Resource:  "user",
		ResourceID: userID.String(),
		CreatedAt: time.Now(),
	})

	return nil
}

// ResetPassword initiates password reset process
func (s *authService) ResetPassword(ctx context.Context, email string) error {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Don't reveal if user exists or not
		return nil
	}

	// TODO: Generate reset token and send email
	// For now, just log the action
	s.auditRepo.Create(ctx, &auth.AuditLog{
		ID:        ulid.New(),
		UserID:    &user.ID,
		Action:    "auth.password.reset_requested",
		Resource:  "user",
		ResourceID: user.ID.String(),
		Metadata:  fmt.Sprintf(`{"email": "%s"}`, email),
		CreatedAt: time.Now(),
	})

	return nil
}

// ConfirmPasswordReset completes password reset process
func (s *authService) ConfirmPasswordReset(ctx context.Context, token, newPassword string) error {
	// TODO: Implement password reset confirmation
	return errors.New("not implemented")
}

// SendEmailVerification sends email verification
func (s *authService) SendEmailVerification(ctx context.Context, userID ulid.ULID) error {
	// TODO: Generate verification token and send email
	s.auditRepo.Create(ctx, &auth.AuditLog{
		ID:        ulid.New(),
		UserID:    &userID,
		Action:    "auth.email_verification.sent",
		Resource:  "user",
		ResourceID: userID.String(),
		CreatedAt: time.Now(),
	})

	return nil
}

// VerifyEmail verifies user's email
func (s *authService) VerifyEmail(ctx context.Context, token string) error {
	// TODO: Implement email verification
	return errors.New("not implemented")
}

// GetCurrentUser returns current user information
func (s *authService) GetCurrentUser(ctx context.Context, userID ulid.ULID) (*user.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

// UpdateProfile updates user profile
func (s *authService) UpdateProfile(ctx context.Context, userID ulid.ULID, req *auth.UpdateProfileRequest) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Update fields if provided
	if req.FirstName != nil && *req.FirstName != "" {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil && *req.LastName != "" {
		user.LastName = *req.LastName
	}
	if req.AvatarURL != nil {
		user.AvatarURL = *req.AvatarURL
	}
	if req.Phone != nil {
		user.Phone = *req.Phone
	}
	if req.Timezone != nil {
		user.Timezone = *req.Timezone
	}
	if req.Language != nil {
		user.Language = *req.Language
	}

	user.UpdatedAt = time.Now()

	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Log profile update
	s.auditRepo.Create(ctx, &auth.AuditLog{
		ID:        ulid.New(),
		UserID:    &userID,
		Action:    "user.profile.updated",
		Resource:  "user",
		ResourceID: userID.String(),
		CreatedAt: time.Now(),
	})

	return nil
}

// GetUserSessions returns user's active sessions
func (s *authService) GetUserSessions(ctx context.Context, userID ulid.ULID) ([]*auth.Session, error) {
	return s.sessionRepo.GetActiveSessionsByUserID(ctx, userID)
}

// RevokeSession revokes a specific user session
func (s *authService) RevokeSession(ctx context.Context, userID, sessionID ulid.ULID) error {
	// Verify session belongs to user
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	if session.UserID != userID {
		return errors.New("session does not belong to user")
	}

	return s.sessionRepo.DeactivateSession(ctx, sessionID)
}

// RevokeAllSessions revokes all user sessions
func (s *authService) RevokeAllSessions(ctx context.Context, userID ulid.ULID) error {
	return s.sessionRepo.DeactivateUserSessions(ctx, userID)
}

// GetAuthContext returns authentication context from token
func (s *authService) GetAuthContext(ctx context.Context, token string) (*auth.AuthContext, error) {
	// Validate JWT token
	claims, err := s.jwtService.ValidateAccessToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Get user permissions if organization is specified
	var permissions []string
	if claims.OrganizationID != nil {
		permissions, _ = s.roleService.GetUserPermissions(ctx, claims.UserID, *claims.OrganizationID)
	}

	return &auth.AuthContext{
		UserID:         claims.UserID,
		OrganizationID: claims.OrganizationID,
		Permissions:    permissions,
		SessionID:      nil, // Will be set by session validation
	}, nil
}

// ValidateAuthToken validates token and returns auth context
func (s *authService) ValidateAuthToken(ctx context.Context, token string) (*auth.AuthContext, error) {
	return s.GetAuthContext(ctx, token)
}