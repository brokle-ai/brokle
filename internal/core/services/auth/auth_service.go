package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"brokle/internal/config"
	"brokle/internal/core/domain/auth"
	"brokle/internal/core/domain/user"
	"brokle/pkg/ulid"
)

// authService implements the auth.AuthService interface
type authService struct {
	authConfig        *config.AuthConfig
	userRepo          user.Repository
	sessionRepo       auth.UserSessionRepository
	auditRepo         auth.AuditLogRepository
	jwtService        auth.JWTService
	roleService       auth.RoleService
	passwordResetRepo auth.PasswordResetTokenRepository
	blacklistedTokens auth.BlacklistedTokenService
}

// NewAuthService creates a new auth service instance
func NewAuthService(
	authConfig *config.AuthConfig,
	userRepo user.Repository,
	sessionRepo auth.UserSessionRepository,
	auditRepo auth.AuditLogRepository,
	jwtService auth.JWTService,
	roleService auth.RoleService,
	passwordResetRepo auth.PasswordResetTokenRepository,
	blacklistedTokens auth.BlacklistedTokenService,
) auth.AuthService {
	return &authService{
		authConfig:        authConfig,
		userRepo:          userRepo,
		sessionRepo:       sessionRepo,
		auditRepo:         auditRepo,
		jwtService:        jwtService,
		roleService:       roleService,
		passwordResetRepo: passwordResetRepo,
		blacklistedTokens: blacklistedTokens,
	}
}

// Login authenticates a user and returns a login response
func (s *authService) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	// Get user with password
	user, err := s.userRepo.GetByEmailWithPassword(ctx, req.Email)
	if err != nil {
		// Log failed login attempt
		auditLog := auth.NewAuditLog(nil, nil, "auth.login.failed", "user", "", fmt.Sprintf(`{"email": "%s", "reason": "user_not_found"}`, req.Email), "", "")
		s.auditRepo.Create(ctx, auditLog)
		return nil, errors.New("invalid credentials")
	}

	// Check if user is active
	if !user.IsActive {
		auditLog := auth.NewAuditLog(&user.ID, nil, "auth.login.failed", "user", user.ID.String(), `{"reason": "user_inactive"}`, "", "")
		s.auditRepo.Create(ctx, auditLog)
		return nil, errors.New("account is inactive")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		auditLog := auth.NewAuditLog(&user.ID, nil, "auth.login.failed", "user", user.ID.String(), `{"reason": "invalid_password"}`, "", "")
		s.auditRepo.Create(ctx, auditLog)
		return nil, errors.New("invalid credentials")
	}

	// Get user permissions for default organization
	var permissions []string
	if user.DefaultOrganizationID != nil {
		permissions, _ = s.roleService.GetUserPermissionStrings(ctx, user.ID, *user.DefaultOrganizationID)
	}

	// Generate access token with JTI for session tracking
	accessToken, jti, err := s.jwtService.GenerateAccessTokenWithJTI(ctx, user.ID, map[string]interface{}{
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

	// Use configurable token TTLs from AuthConfig
	expiresAt := time.Now().Add(s.authConfig.AccessTokenTTL)
	refreshExpiresAt := time.Now().Add(s.authConfig.RefreshTokenTTL)

	// Hash the refresh token for secure storage
	refreshTokenHash := s.hashToken(refreshToken)

	// Extract IP address and user agent from request context (if available)
	var ipAddress, userAgent *string
	// TODO: Extract from request context when available
	
	// Create secure session (NO ACCESS TOKEN STORED)
	session := auth.NewUserSession(user.ID, refreshTokenHash, jti, expiresAt, refreshExpiresAt, ipAddress, userAgent, req.DeviceInfo)
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
	auditLog := auth.NewAuditLog(&user.ID, user.DefaultOrganizationID, "auth.login.success", "user", user.ID.String(), fmt.Sprintf(`{"session_id": "%s"}`, session.ID.String()), "", "")
	s.auditRepo.Create(ctx, auditLog)

	return &auth.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.authConfig.AccessTokenTTL.Seconds()),
	}, nil
}

// Register creates a new user account and auto-login (returns login tokens)
func (s *authService) Register(ctx context.Context, req *auth.RegisterRequest) (*auth.LoginResponse, error) {
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
	newUser := user.NewUser(req.Email, req.FirstName, req.LastName)
	newUser.SetPassword(string(hashedPassword))

	// Set timezone and language from request if provided
	if req.Timezone != "" {
		newUser.Timezone = req.Timezone
	}
	if req.Language != "" {
		newUser.Language = req.Language
	}

	// Create user and profile in a transaction to ensure atomicity
	err = s.userRepo.Transaction(func(tx user.Repository) error {
		// Create user
		if err := tx.Create(ctx, newUser); err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		// Create user profile with default values (fixes registration bug)
		profile := user.NewUserProfile(newUser.ID)
		if req.Timezone != "" {
			profile.Timezone = req.Timezone
		}
		if req.Language != "" {
			profile.Language = req.Language
		}
		
		if err := tx.CreateProfile(ctx, profile); err != nil {
			return fmt.Errorf("failed to create user profile: %w", err)
		}
		
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Log registration
	auditLog := auth.NewAuditLog(&newUser.ID, nil, "auth.register.success", "user", newUser.ID.String(), fmt.Sprintf(`{"email": "%s"}`, newUser.Email), "", "")
	s.auditRepo.Create(ctx, auditLog)

	// Auto-login: Generate tokens for the new user
	var permissions []string
	if newUser.DefaultOrganizationID != nil {
		permissions, _ = s.roleService.GetUserPermissionStrings(ctx, newUser.ID, *newUser.DefaultOrganizationID)
	}

	// Generate access token with JTI for session tracking
	accessToken, jti, err := s.jwtService.GenerateAccessTokenWithJTI(ctx, newUser.ID, map[string]interface{}{
		"email":          newUser.Email,
		"organization_id": newUser.DefaultOrganizationID,
		"permissions":     permissions,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(ctx, newUser.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Use configurable token TTLs from AuthConfig
	expiresAt := time.Now().Add(s.authConfig.AccessTokenTTL)
	refreshExpiresAt := time.Now().Add(s.authConfig.RefreshTokenTTL)

	// Hash the refresh token for secure storage
	refreshTokenHash := s.hashToken(refreshToken)

	// Create secure session (NO ACCESS TOKEN STORED)
	session := auth.NewUserSession(newUser.ID, refreshTokenHash, jti, expiresAt, refreshExpiresAt, nil, nil, nil)
	err = s.sessionRepo.Create(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Update last login (for the auto-login)
	err = s.userRepo.UpdateLastLogin(ctx, newUser.ID)
	if err != nil {
		// Log but don't fail registration
		fmt.Printf("Failed to update last login after registration: %v\n", err)
	}

	// Log auto-login after registration
	auditLog = auth.NewAuditLog(&newUser.ID, newUser.DefaultOrganizationID, "auth.register.auto_login", "user", newUser.ID.String(), fmt.Sprintf(`{"session_id": "%s"}`, session.ID.String()), "", "")
	s.auditRepo.Create(ctx, auditLog)

	// Return login tokens (same as login response)
	return &auth.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.authConfig.AccessTokenTTL.Seconds()),
	}, nil
}

// Logout invalidates a user access token via JTI blacklisting
func (s *authService) Logout(ctx context.Context, jti string, userID ulid.ULID) error {
	// Blacklist the current access token immediately
	expiry := time.Now().Add(s.authConfig.AccessTokenTTL) // Blacklist until token would expire
	err := s.blacklistedTokens.BlacklistToken(ctx, jti, userID, expiry, "user_logout")
	if err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}
	
	// Log logout
	auditLog := auth.NewAuditLog(&userID, nil, "auth.logout.success", "token", jti, "", "", "")
	s.auditRepo.Create(ctx, auditLog)

	return nil
}

// RefreshToken generates new access token using refresh token
func (s *authService) RefreshToken(ctx context.Context, req *auth.RefreshTokenRequest) (*auth.LoginResponse, error) {
	// Validate refresh token
	claims, err := s.jwtService.ValidateRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Get session by refresh token hash
	refreshTokenHash := s.hashToken(req.RefreshToken)
	session, err := s.sessionRepo.GetByRefreshTokenHash(ctx, refreshTokenHash)
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
		permissions, _ = s.roleService.GetUserPermissionStrings(ctx, user.ID, *user.DefaultOrganizationID)
	}

	// Generate new access token with JTI for session tracking
	accessToken, jti, err := s.jwtService.GenerateAccessTokenWithJTI(ctx, user.ID, map[string]interface{}{
		"email":          user.Email,
		"organization_id": user.DefaultOrganizationID,
		"permissions":     permissions,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Implement token rotation if enabled
	var newRefreshToken string
	if s.authConfig.TokenRotationEnabled {
		// Generate new refresh token
		newRefreshToken, err = s.jwtService.GenerateRefreshToken(ctx, user.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to generate new refresh token: %w", err)
		}

		// Blacklist the old refresh token to prevent reuse
		oldRefreshClaims, err := s.jwtService.ValidateRefreshToken(ctx, req.RefreshToken)
		if err == nil && oldRefreshClaims.JWTID != "" {
			// Add old refresh token to blacklist
			err = s.blacklistedTokens.BlacklistToken(
				ctx,
				oldRefreshClaims.JWTID,
				user.ID,
				time.Now().Add(s.authConfig.RefreshTokenTTL), // Keep in blacklist until natural expiry
				"token_rotation",
			)
			if err != nil {
				// Log error but don't fail the rotation
				fmt.Printf("Warning: Failed to blacklist old refresh token: %v\n", err)
			}
		}

		// Update session with new refresh token hash
		session.RefreshTokenHash = s.hashToken(newRefreshToken)
	} else {
		newRefreshToken = req.RefreshToken // Keep same refresh token
	}

	// Update session with new JTI and expiry (NO ACCESS TOKEN STORED)
	session.CurrentJTI = jti
	session.ExpiresAt = time.Now().Add(s.authConfig.AccessTokenTTL)
	session.UpdatedAt = time.Now()
	session.MarkAsUsed() // Update last used timestamp

	err = s.sessionRepo.Update(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	// Log token refresh (with rotation info)
	rotationInfo := ""
	if s.authConfig.TokenRotationEnabled {
		rotationInfo = "with_rotation"
	}
	auditLog := auth.NewAuditLog(&user.ID, user.DefaultOrganizationID, "auth.token.refresh", "session", session.ID.String(), fmt.Sprintf(`{"jti": "%s", "rotation": "%s"}`, jti, rotationInfo), "", "")
	s.auditRepo.Create(ctx, auditLog)

	return &auth.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.authConfig.AccessTokenTTL.Seconds()),
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
	err = s.sessionRepo.RevokeUserSessions(ctx, userID)
	if err != nil {
		// Log but don't fail
		fmt.Printf("Failed to revoke user sessions: %v\n", err)
	}

	// Log password change
	auditLog := auth.NewAuditLog(&userID, nil, "auth.password.changed", "user", userID.String(), "", "", "")
	s.auditRepo.Create(ctx, auditLog)

	return nil
}

// ResetPassword initiates password reset process
func (s *authService) ResetPassword(ctx context.Context, email string) error {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Don't reveal if user exists or not
		return nil
	}

	// Invalidate any existing password reset tokens for this user
	err = s.passwordResetRepo.InvalidateAllUserTokens(ctx, user.ID)
	if err != nil {
		// Log but continue
		fmt.Printf("Failed to invalidate existing password reset tokens: %v\n", err)
	}

	// Generate secure reset token
	tokenBytes := make([]byte, 32)
	_, err = rand.Read(tokenBytes)
	if err != nil {
		return fmt.Errorf("failed to generate reset token: %w", err)
	}
	tokenString := hex.EncodeToString(tokenBytes)

	// Create password reset token (expires in 1 hour)
	resetToken := auth.NewPasswordResetToken(user.ID, tokenString, time.Now().Add(1*time.Hour))
	err = s.passwordResetRepo.Create(ctx, resetToken)
	if err != nil {
		return fmt.Errorf("failed to create password reset token: %w", err)
	}

	// TODO: Send email with reset link containing tokenString
	// The email would contain a link like: https://app.brokle.com/reset-password?token=tokenString

	// Log password reset request
	auditLog := auth.NewAuditLog(&user.ID, nil, "auth.password.reset_requested", "user", user.ID.String(), fmt.Sprintf(`{"email": "%s", "token_id": "%s"}`, email, resetToken.ID.String()), "", "")
	s.auditRepo.Create(ctx, auditLog)

	return nil
}

// ConfirmPasswordReset completes password reset process
func (s *authService) ConfirmPasswordReset(ctx context.Context, token, newPassword string) error {
	// Find and validate password reset token
	resetToken, err := s.passwordResetRepo.GetByToken(ctx, token)
	if err != nil {
		return errors.New("invalid or expired password reset token")
	}

	// Check if token is valid (not used and not expired)
	isValid, err := s.passwordResetRepo.IsValid(ctx, resetToken.ID)
	if err != nil {
		return fmt.Errorf("failed to validate password reset token: %w", err)
	}
	if !isValid {
		return errors.New("password reset token is invalid or expired")
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, resetToken.UserID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if !user.IsActive {
		return errors.New("user account is inactive")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update password
	err = s.userRepo.UpdatePassword(ctx, user.ID, string(hashedPassword))
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Mark token as used
	err = s.passwordResetRepo.MarkAsUsed(ctx, resetToken.ID)
	if err != nil {
		// Log but don't fail - password was already updated
		fmt.Printf("Failed to mark password reset token as used: %v\n", err)
	}

	// Revoke all user sessions (force re-login with new password)
	err = s.sessionRepo.RevokeUserSessions(ctx, user.ID)
	if err != nil {
		// Log but don't fail
		fmt.Printf("Failed to revoke user sessions after password reset: %v\n", err)
	}

	// Log successful password reset
	auditLog := auth.NewAuditLog(&user.ID, nil, "auth.password.reset_completed", "user", user.ID.String(), fmt.Sprintf(`{"token_id": "%s"}`, resetToken.ID.String()), "", "")
	s.auditRepo.Create(ctx, auditLog)

	return nil
}

// SendEmailVerification sends email verification
func (s *authService) SendEmailVerification(ctx context.Context, userID ulid.ULID) error {
	// TODO: Generate verification token and send email
	auditLog := auth.NewAuditLog(&userID, nil, "auth.email_verification.sent", "user", userID.String(), "", "", "")
	s.auditRepo.Create(ctx, auditLog)

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
	auditLog := auth.NewAuditLog(&userID, nil, "user.profile.updated", "user", userID.String(), "", "", "")
	s.auditRepo.Create(ctx, auditLog)

	return nil
}

// GetUserSessions returns user's active sessions
func (s *authService) GetUserSessions(ctx context.Context, userID ulid.ULID) ([]*auth.UserSession, error) {
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

	return s.sessionRepo.RevokeSession(ctx, sessionID)
}

// RevokeAllSessions revokes all user sessions
func (s *authService) RevokeAllSessions(ctx context.Context, userID ulid.ULID) error {
	return s.sessionRepo.RevokeUserSessions(ctx, userID)
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
		permissions, _ = s.roleService.GetUserPermissionStrings(ctx, claims.UserID, *claims.OrganizationID)
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

// RevokeAccessToken immediately revokes an access token by adding it to blacklist
func (s *authService) RevokeAccessToken(ctx context.Context, jti string, userID ulid.ULID, reason string) error {
	// Parse JTI to get token expiration time
	// We need the expiration time to know when to cleanup the blacklisted token
	// For now, we'll use a default expiration time based on config
	expiresAt := time.Now().Add(s.authConfig.AccessTokenTTL)
	
	// Add token to blacklist
	err := s.blacklistedTokens.BlacklistToken(ctx, jti, userID, expiresAt, reason)
	if err != nil {
		return fmt.Errorf("failed to revoke access token: %w", err)
	}

	// Log the revocation
	auditLog := auth.NewAuditLog(&userID, nil, "auth.access_token.revoked", "token", jti, 
		fmt.Sprintf(`{"reason": "%s"}`, reason), "", "")
	s.auditRepo.Create(ctx, auditLog)

	return nil
}

// RevokeUserAccessTokens revokes all active access tokens for a user
func (s *authService) RevokeUserAccessTokens(ctx context.Context, userID ulid.ULID, reason string) error {
	// Blacklist all user tokens
	err := s.blacklistedTokens.BlacklistUserTokens(ctx, userID, reason)
	if err != nil {
		return fmt.Errorf("failed to revoke user access tokens: %w", err)
	}

	// Log the bulk revocation
	auditLog := auth.NewAuditLog(&userID, nil, "auth.user_tokens.revoked", "user", userID.String(), 
		fmt.Sprintf(`{"reason": "%s", "action": "bulk_revocation"}`, reason), "", "")
	s.auditRepo.Create(ctx, auditLog)

	return nil
}

// IsTokenRevoked checks if an access token has been revoked
func (s *authService) IsTokenRevoked(ctx context.Context, jti string) (bool, error) {
	return s.blacklistedTokens.IsTokenBlacklisted(ctx, jti)
}

// hashToken creates a SHA-256 hash of a token for secure storage
func (s *authService) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}