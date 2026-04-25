package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"

	"brokle/internal/config"
	authDomain "brokle/internal/core/domain/auth"
	userDomain "brokle/internal/core/domain/user"
	appErrors "brokle/pkg/errors"
)

// AuthService implements user authentication flows (login, refresh, logout,
// password change/reset, OAuth session helpers). Audit events are recorded
// inline via recordAudit; no decorator layer.
type AuthService struct {
	cfg               *config.AuthConfig
	userRepo          userDomain.Repository
	sessionRepo       authDomain.UserSessionRepository
	auditRepo         authDomain.AuditLogRepository
	passwordResetRepo authDomain.PasswordResetTokenRepository
	jwt               *JWTService
	roles             *RoleService
	blacklist         *BlacklistedTokenService
	redis             *redis.Client // OAuth session storage
	logger            *slog.Logger
}

// NewAuthService wires an AuthService with its dependencies.
func NewAuthService(
	cfg *config.AuthConfig,
	userRepo userDomain.Repository,
	sessionRepo authDomain.UserSessionRepository,
	auditRepo authDomain.AuditLogRepository,
	jwt *JWTService,
	roles *RoleService,
	passwordResetRepo authDomain.PasswordResetTokenRepository,
	blacklist *BlacklistedTokenService,
	redisClient *redis.Client,
	logger *slog.Logger,
) *AuthService {
	return &AuthService{
		cfg:               cfg,
		userRepo:          userRepo,
		sessionRepo:       sessionRepo,
		auditRepo:         auditRepo,
		jwt:               jwt,
		roles:             roles,
		passwordResetRepo: passwordResetRepo,
		blacklist:         blacklist,
		redis:             redisClient,
		logger:            logger,
	}
}

// recordAudit writes an audit log best-effort; audit failures never propagate.
func (s *AuthService) recordAudit(ctx context.Context, userID *uuid.UUID, action, resource, resourceID string, metadata map[string]any) {
	if s.auditRepo == nil {
		return
	}
	auditLog := authDomain.NewAuditLog(userID, nil, action, resource, resourceID, metadata, "", "")
	if err := s.auditRepo.Create(ctx, auditLog); err != nil && s.logger != nil {
		s.logger.Error("failed to write audit log", "action", action, "error", err)
	}
}

// classifyAuthFailure maps a pre-classified AppError to a stable audit reason
// string. Unknown / non-AppError paths fall through to "system_error".
func classifyAuthFailure(err error) string {
	appErr, ok := appErrors.IsAppError(err)
	if !ok {
		return "system_error"
	}
	switch appErr.Type {
	case appErrors.TypeAuthentication:
		return "invalid_credentials"
	case appErrors.TypePermission:
		return "account_inactive"
	}
	return "system_error"
}

// classifyRefreshFailure categorises refresh-token rejections.
func classifyRefreshFailure(err error) string {
	appErr, ok := appErrors.IsAppError(err)
	if !ok {
		return "system_error"
	}
	if appErr.Type == appErrors.TypeAuthentication {
		return "invalid_token"
	}
	return "system_error"
}

// Login authenticates a user and returns a login response
func (s *AuthService) Login(ctx context.Context, req *authDomain.LoginRequest) (resp *authDomain.LoginResponse, err error) {
	var authUserID *uuid.UUID
	defer func() {
		if err != nil {
			s.recordAudit(ctx, authUserID, "auth.login.failed", "user", "", map[string]any{
				"email":  req.Email,
				"reason": classifyAuthFailure(err),
			})
			return
		}
		s.recordAudit(ctx, authUserID, "auth.login.success", "user", "", map[string]any{
			"email": req.Email,
		})
	}()

	// Get user with password
	user, err := s.userRepo.GetByEmailWithPassword(ctx, req.Email)
	if err != nil {
		if errors.Is(err, userDomain.ErrNotFound) {
			return nil, appErrors.NewUnauthorizedError("invalid email or password")
		}
		return nil, appErrors.NewInternalError("authentication service unavailable", err)
	}
	authUserID = &user.ID

	// Check if user is active
	if !user.IsActive {
		return nil, appErrors.NewForbiddenError("account is inactive")
	}

	// Block OAuth users from password login
	if user.AuthMethod == "oauth" {
		providerName := "OAuth"
		if user.OAuthProvider != nil {
			providerName = *user.OAuthProvider
		}
		return nil, appErrors.NewUnauthorizedError("this account uses " + providerName + " login - please sign in with " + providerName)
	}

	// Verify password (only for password-based accounts)
	if user.AuthMethod == "password" {
		if !user.HasPassword() {
			return nil, appErrors.NewUnauthorizedError("invalid email or password")
		}

		err = bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(req.Password))
		if err != nil {
			return nil, appErrors.NewUnauthorizedError("invalid email or password")
		}
	}

	// Get user effective permissions across all scopes
	// Note: Permissions are now handled by OrganizationMemberService
	permissions := []string{}

	// Generate access token with JTI for session tracking
	accessToken, jti, err := s.jwt.GenerateAccessTokenWithJTI(ctx, user.ID, map[string]any{
		"email":           user.Email,
		"organization_id": user.DefaultOrganizationID,
		"permissions":     permissions,
	})
	if err != nil {
		return nil, appErrors.NewInternalError("failed to generate access token", err)
	}

	refreshToken, err := s.jwt.GenerateRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to generate refresh token", err)
	}

	// Use configurable token TTLs from AuthConfig
	expiresAt := time.Now().Add(s.cfg.AccessTokenTTL)
	refreshExpiresAt := time.Now().Add(s.cfg.RefreshTokenTTL)

	// Hash the refresh token for secure storage
	refreshTokenHash := s.hashToken(refreshToken)

	// Extract IP address and user agent from request context (if available)
	var ipAddress, userAgent *string
	// TODO: Extract from request context when available

	// Create secure session (NO ACCESS TOKEN STORED)
	session := authDomain.NewUserSession(user.ID, refreshTokenHash, jti, expiresAt, refreshExpiresAt, ipAddress, userAgent, req.DeviceInfo)
	err = s.sessionRepo.Create(ctx, session)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to create session", err)
	}

	// Update last_login. Best-effort — login succeeds regardless.
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		s.logger.Warn("failed to update last_login after successful login",
			"error", err, "user_id", user.ID)
	}

	return &authDomain.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.cfg.AccessTokenTTL.Seconds()),
	}, nil
}

// GenerateTokensForUser generates login tokens for a user without password validation.
// Used for OAuth signup, email verification, trusted authentication flows, and existing OAuth user login.
func (s *AuthService) GenerateTokensForUser(ctx context.Context, userID uuid.UUID) (*authDomain.LoginResponse, error) {
	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, userDomain.ErrNotFound) {
			return nil, appErrors.NewNotFoundError("user not found")
		}
		return nil, appErrors.NewInternalError("user lookup failed", err)
	}

	// Check if user is active
	if !user.IsActive {
		return nil, appErrors.NewForbiddenError("account is inactive")
	}

	// Get user effective permissions
	permissions := []string{}

	// Generate access token with JTI for session tracking
	accessToken, jti, err := s.jwt.GenerateAccessTokenWithJTI(ctx, user.ID, map[string]any{
		"email":           user.Email,
		"organization_id": user.DefaultOrganizationID,
		"permissions":     permissions,
	})
	if err != nil {
		return nil, appErrors.NewInternalError("failed to generate access token", err)
	}

	refreshToken, err := s.jwt.GenerateRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to generate refresh token", err)
	}

	// Use configurable token TTLs from AuthConfig
	expiresAt := time.Now().Add(s.cfg.AccessTokenTTL)
	refreshExpiresAt := time.Now().Add(s.cfg.RefreshTokenTTL)

	// Hash the refresh token for secure storage
	refreshTokenHash := s.hashToken(refreshToken)

	// Create secure session
	session := authDomain.NewUserSession(user.ID, refreshTokenHash, jti, expiresAt, refreshExpiresAt, nil, nil, nil)
	err = s.sessionRepo.Create(ctx, session)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to create session", err)
	}

	// Update last login
	_ = s.userRepo.UpdateLastLogin(ctx, user.ID)

	return &authDomain.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.cfg.AccessTokenTTL.Seconds()),
	}, nil
}

// Logout invalidates a user access token via JTI blacklisting
func (s *AuthService) Logout(ctx context.Context, jti string, userID uuid.UUID) (err error) {
	defer func() {
		action := "auth.logout.success"
		if err != nil {
			action = "auth.logout.failed"
		}
		s.recordAudit(ctx, &userID, action, "user", userID.String(), map[string]any{"jti": jti})
	}()

	// Blacklist the current access token immediately
	expiry := time.Now().Add(s.cfg.AccessTokenTTL) // Blacklist until token would expire
	if err = s.blacklist.BlacklistToken(ctx, jti, userID, expiry, "user_logout"); err != nil {
		return appErrors.NewInternalError("failed to blacklist token", err)
	}
	return nil
}

// RefreshToken generates new access token using refresh token
func (s *AuthService) RefreshToken(ctx context.Context, req *authDomain.RefreshTokenRequest) (resp *authDomain.LoginResponse, err error) {
	var refreshUserID *uuid.UUID
	defer func() {
		if err != nil {
			s.recordAudit(ctx, refreshUserID, "auth.refresh_token.failed", "token", "", map[string]any{
				"reason": classifyRefreshFailure(err),
			})
			return
		}
		s.recordAudit(ctx, refreshUserID, "auth.refresh_token.success", "token", "", nil)
	}()

	// Validate refresh token
	claims, err := s.jwt.ValidateRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		if errors.Is(err, authDomain.ErrTokenExpired) {
			return nil, appErrors.NewUnauthorizedError("refresh token expired")
		}
		if errors.Is(err, authDomain.ErrTokenInvalid) {
			return nil, appErrors.NewUnauthorizedError("invalid refresh token")
		}
		return nil, appErrors.NewInternalError("token validation failed", err)
	}

	// Get session by refresh token hash
	refreshTokenHash := s.hashToken(req.RefreshToken)
	session, err := s.sessionRepo.GetByRefreshTokenHash(ctx, refreshTokenHash)
	if err != nil {
		if errors.Is(err, authDomain.ErrSessionNotFound) {
			return nil, appErrors.NewUnauthorizedError("session not found")
		}
		return nil, appErrors.NewInternalError("session lookup failed", err)
	}

	if !session.IsActive {
		return nil, appErrors.NewUnauthorizedError("session is inactive")
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, userDomain.ErrNotFound) {
			return nil, appErrors.NewUnauthorizedError("user not found")
		}
		return nil, appErrors.NewInternalError("user lookup failed", err)
	}
	refreshUserID = &user.ID

	if !user.IsActive {
		return nil, appErrors.NewForbiddenError("user is inactive")
	}

	// Get user effective permissions across all scopes
	// Note: Permissions are now handled by OrganizationMemberService
	permissions := []string{}

	// Generate new access token with JTI for session tracking
	accessToken, jti, err := s.jwt.GenerateAccessTokenWithJTI(ctx, user.ID, map[string]any{
		"email":           user.Email,
		"organization_id": user.DefaultOrganizationID,
		"permissions":     permissions,
	})
	if err != nil {
		return nil, appErrors.NewInternalError("failed to generate access token", err)
	}

	// Implement token rotation if enabled
	var newRefreshToken string
	if s.cfg.TokenRotationEnabled {
		// Generate new refresh token
		newRefreshToken, err = s.jwt.GenerateRefreshToken(ctx, user.ID)
		if err != nil {
			return nil, appErrors.NewInternalError("failed to generate new refresh token", err)
		}

		// Blacklist the old refresh token to prevent reuse
		oldRefreshClaims, err := s.jwt.ValidateRefreshToken(ctx, req.RefreshToken)
		if err == nil && oldRefreshClaims.JWTID != "" {
			// Add old refresh token to blacklist. Best-effort — if this
			// fails the old token would still be honoured until natural
			// expiry, which weakens rotation but doesn't block the new
			// token from working.
			if err := s.blacklist.BlacklistToken(
				ctx,
				oldRefreshClaims.JWTID,
				user.ID,
				time.Now().Add(s.cfg.RefreshTokenTTL),
				"token_rotation",
			); err != nil {
				s.logger.Warn("failed to blacklist old refresh token during rotation",
					"error", err, "jti", oldRefreshClaims.JWTID, "user_id", user.ID)
			}
		}

		// Update session with new refresh token hash
		session.RefreshTokenHash = s.hashToken(newRefreshToken)
	} else {
		newRefreshToken = req.RefreshToken // Keep same refresh token
	}

	// Update session with new JTI and expiry (NO ACCESS TOKEN STORED)
	session.CurrentJTI = jti
	session.ExpiresAt = time.Now().Add(s.cfg.AccessTokenTTL)
	session.UpdatedAt = time.Now()
	session.MarkAsUsed() // Update last used timestamp

	err = s.sessionRepo.Update(ctx, session)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to update session", err)
	}

	return &authDomain.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.cfg.AccessTokenTTL.Seconds()),
	}, nil
}

// ChangePassword changes a user's password
func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) (err error) {
	defer func() {
		if err == nil {
			s.recordAudit(ctx, &userID, "auth.password.changed", "user", userID.String(), nil)
		}
	}()

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, userDomain.ErrNotFound) {
			return appErrors.NewNotFoundError("user not found")
		}
		return appErrors.NewInternalError("user lookup failed", err)
	}

	// Verify current password
	if !user.HasPassword() {
		return appErrors.NewUnauthorizedError("user has no password set")
	}
	err = bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(currentPassword))
	if err != nil {
		return appErrors.NewUnauthorizedError("current password is incorrect")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return appErrors.NewInternalError("failed to hash new password", err)
	}

	// Update password
	err = s.userRepo.UpdatePassword(ctx, userID, string(hashedPassword))
	if err != nil {
		return appErrors.NewInternalError("failed to update password", err)
	}

	// Revoke all user sessions (force re-login). Best-effort — if this
	// fails the password change still succeeded, but we want operators
	// to know that stale sessions may be lingering for the user.
	if err := s.sessionRepo.RevokeUserSessions(ctx, userID); err != nil {
		s.logger.Warn("failed to revoke sessions after password change",
			"error", err, "user_id", userID)
	}

	return nil
}

// ResetPassword initiates password reset process
func (s *AuthService) ResetPassword(ctx context.Context, email string) error {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Don't reveal if user exists or not
		return nil
	}
	s.recordAudit(ctx, &user.ID, "auth.password.reset_requested", "user", user.ID.String(), map[string]any{
		"email": email,
	})

	// Invalidate any existing password reset tokens for this user.
	// Best-effort — older outstanding tokens become "long-lived" if
	// this fails, but the new token will still work and ConfirmPasswordReset
	// validates against the per-token used flag.
	if err := s.passwordResetRepo.InvalidateAllUserTokens(ctx, user.ID); err != nil {
		s.logger.Warn("failed to invalidate prior password reset tokens",
			"error", err, "user_id", user.ID)
	}

	// Generate secure reset token
	tokenBytes := make([]byte, 32)
	_, err = rand.Read(tokenBytes)
	if err != nil {
		return appErrors.NewInternalError("failed to generate reset token", err)
	}
	tokenString := hex.EncodeToString(tokenBytes)

	// Create password reset token (expires in 1 hour)
	resetToken := authDomain.NewPasswordResetToken(user.ID, tokenString, time.Now().Add(1*time.Hour))
	err = s.passwordResetRepo.Create(ctx, resetToken)
	if err != nil {
		return appErrors.NewInternalError("failed to create password reset token", err)
	}

	// TODO: Send email with reset link containing tokenString
	// The email would contain a link like: https://app.brokle.com/reset-password?token=tokenString

	return nil
}

// ConfirmPasswordReset completes password reset process
func (s *AuthService) ConfirmPasswordReset(ctx context.Context, token, newPassword string) error {
	// Find and validate password reset token
	resetToken, err := s.passwordResetRepo.GetByToken(ctx, token)
	if err != nil {
		return appErrors.NewUnauthorizedError("invalid or expired password reset token")
	}

	// Check if token is valid (not used and not expired)
	isValid, err := s.passwordResetRepo.IsValid(ctx, resetToken.ID)
	if err != nil {
		return appErrors.NewInternalError("failed to validate password reset token", err)
	}
	if !isValid {
		return appErrors.NewUnauthorizedError("password reset token is invalid or expired")
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, resetToken.UserID)
	if err != nil {
		if errors.Is(err, userDomain.ErrNotFound) {
			return appErrors.NewNotFoundError("user not found")
		}
		return appErrors.NewInternalError("user lookup failed", err)
	}

	if !user.IsActive {
		return appErrors.NewForbiddenError("user account is inactive")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return appErrors.NewInternalError("failed to hash new password", err)
	}

	// Update password
	err = s.userRepo.UpdatePassword(ctx, user.ID, string(hashedPassword))
	if err != nil {
		return appErrors.NewInternalError("failed to update password", err)
	}

	// Mark token as used. Best-effort — the password update has already
	// landed; if this fails the token would be re-usable until expiry,
	// which is a degraded state worth alerting on but not aborting on.
	if err := s.passwordResetRepo.MarkAsUsed(ctx, resetToken.ID); err != nil {
		s.logger.Warn("failed to mark password reset token as used",
			"error", err, "token_id", resetToken.ID, "user_id", user.ID)
	}

	// Revoke all user sessions (force re-login with new password).
	// Best-effort — same reasoning as ChangePassword above.
	if err := s.sessionRepo.RevokeUserSessions(ctx, user.ID); err != nil {
		s.logger.Warn("failed to revoke sessions after password reset",
			"error", err, "user_id", user.ID)
	}

	s.recordAudit(ctx, &user.ID, "auth.password.reset", "user", user.ID.String(), nil)
	return nil
}

// SendEmailVerification sends email verification
func (s *AuthService) SendEmailVerification(ctx context.Context, userID uuid.UUID) error {
	// TODO: Generate verification token and send email

	return nil
}

// VerifyEmail verifies user's email
func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	// TODO: Implement email verification
	return appErrors.NewNotImplementedError("email verification not implemented")
}

// GetCurrentUser returns current user information
func (s *AuthService) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*userDomain.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

// UpdateProfile updates user profile
func (s *AuthService) UpdateProfile(ctx context.Context, userID uuid.UUID, req *authDomain.UpdateAuthProfileRequest) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, userDomain.ErrNotFound) {
			return appErrors.NewNotFoundError("user not found")
		}
		return appErrors.NewInternalError("user lookup failed", err)
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
		return appErrors.NewInternalError("failed to update user", err)
	}

	return nil
}

// GetUserSessions returns user's active sessions
func (s *AuthService) GetUserSessions(ctx context.Context, userID uuid.UUID) ([]*authDomain.UserSession, error) {
	return s.sessionRepo.GetActiveSessionsByUserID(ctx, userID)
}

// RevokeSession revokes a specific user session
func (s *AuthService) RevokeSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	// Verify session belongs to user
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, authDomain.ErrSessionNotFound) {
			return appErrors.NewNotFoundError("session not found")
		}
		return appErrors.NewInternalError("session lookup failed", err)
	}

	if session.UserID != userID {
		return appErrors.NewForbiddenError("session does not belong to user")
	}

	return s.sessionRepo.RevokeSession(ctx, sessionID)
}

// RevokeAllSessions revokes all user sessions
func (s *AuthService) RevokeAllSessions(ctx context.Context, userID uuid.UUID) error {
	return s.sessionRepo.RevokeUserSessions(ctx, userID)
}

// GetAuthContext returns authentication context from token
func (s *AuthService) GetAuthContext(ctx context.Context, token string) (*authDomain.AuthContext, error) {
	// Validate JWT token
	claims, err := s.jwt.ValidateAccessToken(ctx, token)
	if err != nil {
		if errors.Is(err, authDomain.ErrTokenExpired) {
			return nil, appErrors.NewUnauthorizedError("token expired")
		}
		if errors.Is(err, authDomain.ErrTokenInvalid) {
			return nil, appErrors.NewUnauthorizedError("invalid token")
		}
		return nil, appErrors.NewInternalError("token validation failed", err)
	}

	// Return clean auth context - permissions resolved dynamically when needed
	return claims.GetUserContext(), nil
}

// ValidateAuthToken validates token and returns auth context
func (s *AuthService) ValidateAuthToken(ctx context.Context, token string) (*authDomain.AuthContext, error) {
	return s.GetAuthContext(ctx, token)
}

// RevokeAccessToken immediately revokes an access token by adding it to blacklist
func (s *AuthService) RevokeAccessToken(ctx context.Context, jti string, userID uuid.UUID, reason string) error {
	// Parse JTI to get token expiration time
	// We need the expiration time to know when to cleanup the blacklisted token
	// For now, we'll use a default expiration time based on config
	expiresAt := time.Now().Add(s.cfg.AccessTokenTTL)

	// Add token to blacklist
	err := s.blacklist.BlacklistToken(ctx, jti, userID, expiresAt, reason)
	if err != nil {
		return appErrors.NewInternalError("failed to revoke access token", err)
	}

	return nil
}

// RevokeUserAccessTokens revokes all active access tokens for a user
func (s *AuthService) RevokeUserAccessTokens(ctx context.Context, userID uuid.UUID, reason string) error {
	// Blacklist all user tokens
	err := s.blacklist.BlacklistUserTokens(ctx, userID, reason)
	if err != nil {
		return appErrors.NewInternalError("failed to revoke user access tokens", err)
	}

	return nil
}

// IsTokenRevoked checks if an access token has been revoked
func (s *AuthService) IsTokenRevoked(ctx context.Context, jti string) (bool, error) {
	return s.blacklist.IsTokenBlacklisted(ctx, jti)
}

// hashToken creates a SHA-256 hash of a token for secure storage
func (s *AuthService) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
