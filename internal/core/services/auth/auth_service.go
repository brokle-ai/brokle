package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"brokle/internal/config"
	"brokle/internal/core/domain/auth"
	"brokle/internal/core/domain/user"
	"brokle/pkg/ulid"
	appErrors "brokle/pkg/errors"
)

// authService implements the auth.AuthService interface
type authService struct {
	authConfig        *config.AuthConfig
	userRepo          user.Repository
	sessionRepo       auth.UserSessionRepository
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
	jwtService auth.JWTService,
	roleService auth.RoleService,
	passwordResetRepo auth.PasswordResetTokenRepository,
	blacklistedTokens auth.BlacklistedTokenService,
) auth.AuthService {
	return &authService{
		authConfig:        authConfig,
		userRepo:          userRepo,
		sessionRepo:       sessionRepo,
		jwtService:        jwtService,
		roleService:       roleService,
		passwordResetRepo: passwordResetRepo,
		blacklistedTokens: blacklistedTokens,
	}
}

// Login authenticates a user and returns a login response
func (s *authService) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	// Get user with password
	foundUser, err := s.userRepo.GetByEmailWithPassword(ctx, req.Email)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return nil, appErrors.NewUnauthorizedError("Invalid email or password")
		}
		return nil, appErrors.NewInternalError("Authentication service unavailable", err)
	}

	// Check if user is active
	if !foundUser.IsActive {
		return nil, appErrors.NewForbiddenError("Account is inactive")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(req.Password))
	if err != nil {
		return nil, appErrors.NewUnauthorizedError("Invalid email or password")
	}

	// Get user effective permissions across all scopes
	// Note: Permissions are now handled by OrganizationMemberService
	permissions := []string{}

	// Generate access token with JTI for session tracking
	accessToken, jti, err := s.jwtService.GenerateAccessTokenWithJTI(ctx, foundUser.ID, map[string]interface{}{
		"email":          foundUser.Email,
		"organization_id": foundUser.DefaultOrganizationID,
		"permissions":     permissions,
	})
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to generate access token", err)
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(ctx, foundUser.ID)
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to generate refresh token", err)
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
	session := auth.NewUserSession(foundUser.ID, refreshTokenHash, jti, expiresAt, refreshExpiresAt, ipAddress, userAgent, req.DeviceInfo)
	err = s.sessionRepo.Create(ctx, session)
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to create session", err)
	}

	// Update last login
	err = s.userRepo.UpdateLastLogin(ctx, foundUser.ID)
	if err != nil {
		// Non-critical error, continue with login
	}

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
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, user.ErrNotFound) {
		return nil, appErrors.NewInternalError("User lookup failed", err)
	}
	if existingUser != nil {
		return nil, appErrors.NewConflictError("Email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to hash password", err)
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
			return appErrors.NewInternalError("Failed to create user", err)
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
			return appErrors.NewInternalError("Failed to create user profile", err)
		}
		
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Auto-login: Generate tokens for the new user - get effective permissions
	// Note: Permissions are now handled by OrganizationMemberService
	permissions := []string{}

	// Generate access token with JTI for session tracking
	accessToken, jti, err := s.jwtService.GenerateAccessTokenWithJTI(ctx, newUser.ID, map[string]interface{}{
		"email":          newUser.Email,
		"organization_id": newUser.DefaultOrganizationID,
		"permissions":     permissions,
	})
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to generate access token", err)
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(ctx, newUser.ID)
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to generate refresh token", err)
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
		return nil, appErrors.NewInternalError("Failed to create session", err)
	}

	// Update last login (for the auto-login)
	err = s.userRepo.UpdateLastLogin(ctx, newUser.ID)
	if err != nil {
		// Non-critical error, continue with registration
	}

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
		return appErrors.NewInternalError("Failed to blacklist token", err)
	}
	

	return nil
}

// RefreshToken generates new access token using refresh token
func (s *authService) RefreshToken(ctx context.Context, req *auth.RefreshTokenRequest) (*auth.LoginResponse, error) {
	// Validate refresh token
	claims, err := s.jwtService.ValidateRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		if errors.Is(err, auth.ErrTokenExpired) {
			return nil, appErrors.NewUnauthorizedError("Refresh token expired")
		}
		if errors.Is(err, auth.ErrTokenInvalid) {
			return nil, appErrors.NewUnauthorizedError("Invalid refresh token")
		}
		return nil, appErrors.NewInternalError("Token validation failed", err)
	}

	// Get session by refresh token hash
	refreshTokenHash := s.hashToken(req.RefreshToken)
	session, err := s.sessionRepo.GetByRefreshTokenHash(ctx, refreshTokenHash)
	if err != nil {
		if errors.Is(err, auth.ErrSessionNotFound) {
			return nil, appErrors.NewUnauthorizedError("Session not found")
		}
		return nil, appErrors.NewInternalError("Session lookup failed", err)
	}

	if !session.IsActive {
		return nil, appErrors.NewUnauthorizedError("Session is inactive")
	}

	// Get user
	foundUser, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return nil, appErrors.NewUnauthorizedError("User not found")
		}
		return nil, appErrors.NewInternalError("User lookup failed", err)
	}

	if !foundUser.IsActive {
		return nil, appErrors.NewForbiddenError("User is inactive")
	}

	// Get user effective permissions across all scopes
	// Note: Permissions are now handled by OrganizationMemberService
	permissions := []string{}

	// Generate new access token with JTI for session tracking
	accessToken, jti, err := s.jwtService.GenerateAccessTokenWithJTI(ctx, foundUser.ID, map[string]interface{}{
		"email":          foundUser.Email,
		"organization_id": foundUser.DefaultOrganizationID,
		"permissions":     permissions,
	})
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to generate access token", err)
	}

	// Implement token rotation if enabled
	var newRefreshToken string
	if s.authConfig.TokenRotationEnabled {
		// Generate new refresh token
		newRefreshToken, err = s.jwtService.GenerateRefreshToken(ctx, foundUser.ID)
		if err != nil {
			return nil, appErrors.NewInternalError("Failed to generate new refresh token", err)
		}

		// Blacklist the old refresh token to prevent reuse
		oldRefreshClaims, err := s.jwtService.ValidateRefreshToken(ctx, req.RefreshToken)
		if err == nil && oldRefreshClaims.JWTID != "" {
			// Add old refresh token to blacklist
			err = s.blacklistedTokens.BlacklistToken(
				ctx,
				oldRefreshClaims.JWTID,
				foundUser.ID,
				time.Now().Add(s.authConfig.RefreshTokenTTL), // Keep in blacklist until natural expiry
				"token_rotation",
			)
			if err != nil {
				// Non-critical error, continue with token rotation
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
		return nil, appErrors.NewInternalError("Failed to update session", err)
	}


	return &auth.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.authConfig.AccessTokenTTL.Seconds()),
	}, nil
}

// ChangePassword changes a user's password
func (s *authService) ChangePassword(ctx context.Context, userID ulid.ULID, currentPassword, newPassword string) error {
	foundUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return appErrors.NewNotFoundError("User not found")
		}
		return appErrors.NewInternalError("User lookup failed", err)
	}

	// Verify current password
	err = bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(currentPassword))
	if err != nil {
		return appErrors.NewUnauthorizedError("Current password is incorrect")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return appErrors.NewInternalError("Failed to hash new password", err)
	}

	// Update password
	err = s.userRepo.UpdatePassword(ctx, userID, string(hashedPassword))
	if err != nil {
		return appErrors.NewInternalError("Failed to update password", err)
	}

	// Revoke all user sessions (force re-login)
	err = s.sessionRepo.RevokeUserSessions(ctx, userID)
	if err != nil {
		// Non-critical error, password change succeeded
	}


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
		// Non-critical error, continue with reset process
	}

	// Generate secure reset token
	tokenBytes := make([]byte, 32)
	_, err = rand.Read(tokenBytes)
	if err != nil {
		return appErrors.NewInternalError("Failed to generate reset token", err)
	}
	tokenString := hex.EncodeToString(tokenBytes)

	// Create password reset token (expires in 1 hour)
	resetToken := auth.NewPasswordResetToken(user.ID, tokenString, time.Now().Add(1*time.Hour))
	err = s.passwordResetRepo.Create(ctx, resetToken)
	if err != nil {
		return appErrors.NewInternalError("Failed to create password reset token", err)
	}

	// TODO: Send email with reset link containing tokenString
	// The email would contain a link like: https://app.brokle.com/reset-password?token=tokenString


	return nil
}

// ConfirmPasswordReset completes password reset process
func (s *authService) ConfirmPasswordReset(ctx context.Context, token, newPassword string) error {
	// Find and validate password reset token
	resetToken, err := s.passwordResetRepo.GetByToken(ctx, token)
	if err != nil {
		return appErrors.NewUnauthorizedError("Invalid or expired password reset token")
	}

	// Check if token is valid (not used and not expired)
	isValid, err := s.passwordResetRepo.IsValid(ctx, resetToken.ID)
	if err != nil {
		return appErrors.NewInternalError("Failed to validate password reset token", err)
	}
	if !isValid {
		return appErrors.NewUnauthorizedError("Password reset token is invalid or expired")
	}

	// Get user
	foundUser, err := s.userRepo.GetByID(ctx, resetToken.UserID)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return appErrors.NewNotFoundError("User not found")
		}
		return appErrors.NewInternalError("User lookup failed", err)
	}

	if !foundUser.IsActive {
		return appErrors.NewForbiddenError("User account is inactive")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return appErrors.NewInternalError("Failed to hash new password", err)
	}

	// Update password
	err = s.userRepo.UpdatePassword(ctx, foundUser.ID, string(hashedPassword))
	if err != nil {
		return appErrors.NewInternalError("Failed to update password", err)
	}

	// Mark token as used
	err = s.passwordResetRepo.MarkAsUsed(ctx, resetToken.ID)
	if err != nil {
		// Non-critical error - password was already updated
	}

	// Revoke all user sessions (force re-login with new password)
	err = s.sessionRepo.RevokeUserSessions(ctx, foundUser.ID)
	if err != nil {
		// Non-critical error, password reset succeeded
	}


	return nil
}

// SendEmailVerification sends email verification
func (s *authService) SendEmailVerification(ctx context.Context, userID ulid.ULID) error {
	// TODO: Generate verification token and send email

	return nil
}

// VerifyEmail verifies user's email
func (s *authService) VerifyEmail(ctx context.Context, token string) error {
	// TODO: Implement email verification
	return appErrors.NewNotImplementedError("Email verification not implemented")
}

// GetCurrentUser returns current user information
func (s *authService) GetCurrentUser(ctx context.Context, userID ulid.ULID) (*user.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

// UpdateProfile updates user profile
func (s *authService) UpdateProfile(ctx context.Context, userID ulid.ULID, req *auth.UpdateProfileRequest) error {
	foundUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return appErrors.NewNotFoundError("User not found")
		}
		return appErrors.NewInternalError("User lookup failed", err)
	}

	// Update fields if provided
	if req.FirstName != nil && *req.FirstName != "" {
		foundUser.FirstName = *req.FirstName
	}
	if req.LastName != nil && *req.LastName != "" {
		foundUser.LastName = *req.LastName
	}
	if req.Timezone != nil {
		foundUser.Timezone = *req.Timezone
	}
	if req.Language != nil {
		foundUser.Language = *req.Language
	}

	foundUser.UpdatedAt = time.Now()

	err = s.userRepo.Update(ctx, foundUser)
	if err != nil {
		return appErrors.NewInternalError("Failed to update user", err)
	}


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
		if errors.Is(err, auth.ErrSessionNotFound) {
			return appErrors.NewNotFoundError("Session not found")
		}
		return appErrors.NewInternalError("Session lookup failed", err)
	}

	if session.UserID != userID {
		return appErrors.NewForbiddenError("Session does not belong to user")
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
		if errors.Is(err, auth.ErrTokenExpired) {
			return nil, appErrors.NewUnauthorizedError("Token expired")
		}
		if errors.Is(err, auth.ErrTokenInvalid) {
			return nil, appErrors.NewUnauthorizedError("Invalid token")
		}
		return nil, appErrors.NewInternalError("Token validation failed", err)
	}

	// Return clean auth context - permissions resolved dynamically when needed
	return claims.GetUserContext(), nil
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
		return appErrors.NewInternalError("Failed to revoke access token", err)
	}


	return nil
}

// RevokeUserAccessTokens revokes all active access tokens for a user
func (s *authService) RevokeUserAccessTokens(ctx context.Context, userID ulid.ULID, reason string) error {
	// Blacklist all user tokens
	err := s.blacklistedTokens.BlacklistUserTokens(ctx, userID, reason)
	if err != nil {
		return appErrors.NewInternalError("Failed to revoke user access tokens", err)
	}


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