package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"brokle/internal/core/domain/auth"
	"brokle/internal/core/domain/user"
	"brokle/pkg/ulid"
)

// ptrToString safely converts *string to string
func ptrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// sessionService implements the auth.SessionService interface
type sessionService struct {
	sessionRepo auth.SessionRepository
	userRepo    user.Repository
	jwtService  auth.JWTService
	auditRepo   auth.AuditLogRepository
}

// NewSessionService creates a new session service instance
func NewSessionService(
	sessionRepo auth.SessionRepository,
	userRepo user.Repository,
	jwtService auth.JWTService,
	auditRepo auth.AuditLogRepository,
) auth.SessionService {
	return &sessionService{
		sessionRepo: sessionRepo,
		userRepo:    userRepo,
		jwtService:  jwtService,
		auditRepo:   auditRepo,
	}
}

// CreateSession creates a new user session
func (s *sessionService) CreateSession(ctx context.Context, userID ulid.ULID, req *auth.CreateSessionRequest) (*auth.Session, error) {
	// Verify user exists and is active
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if !user.IsActive {
		return nil, errors.New("user account is inactive")
	}

	// Get user permissions for token generation (placeholder - would be handled by role service)
	var permissions []string

	// Generate access and refresh tokens
	accessToken, err := s.jwtService.GenerateAccessToken(ctx, userID, map[string]interface{}{
		"email":          user.Email,
		"organization_id": user.DefaultOrganizationID,
		"permissions":     permissions,
		"ip_address":      req.IPAddress,
		"user_agent":      req.UserAgent,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Calculate expiration times
	config := auth.DefaultTokenConfig()
	expiresAt := time.Now().Add(config.AccessTokenTTL)
	refreshExpiresAt := time.Now().Add(config.RefreshTokenTTL)
	
	// Extend refresh token if "remember me" is checked
	if req.Remember {
		refreshExpiresAt = time.Now().Add(30 * 24 * time.Hour) // 30 days
	}

	// Create session
	session := &auth.Session{
		ID:               ulid.New(),
		UserID:           userID,
		Token:            accessToken,
		RefreshToken:     refreshToken,
		ExpiresAt:        expiresAt,
		RefreshExpiresAt: refreshExpiresAt,
		IPAddress:        req.IPAddress,
		UserAgent:        req.UserAgent,
		IsActive:         true,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	err = s.sessionRepo.Create(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Log session creation
	s.auditRepo.Create(ctx, &auth.AuditLog{
		ID:             ulid.New(),
		UserID:         &userID,
		OrganizationID: user.DefaultOrganizationID,
		Action:         "session.created",
		Resource:       "session",
		ResourceID:     session.ID.String(),
		IPAddress:      ptrToString(req.IPAddress),
		UserAgent:      ptrToString(req.UserAgent),
		Metadata:       fmt.Sprintf(`{"remember": %t}`, req.Remember),
		CreatedAt:      time.Now(),
	})

	return session, nil
}

// GetSession retrieves a session by ID
func (s *sessionService) GetSession(ctx context.Context, sessionID ulid.ULID) (*auth.Session, error) {
	return s.sessionRepo.GetByID(ctx, sessionID)
}

// GetSessionByToken retrieves a session by access token
func (s *sessionService) GetSessionByToken(ctx context.Context, token string) (*auth.Session, error) {
	return s.sessionRepo.GetByToken(ctx, token)
}

// ValidateSession validates a session token and returns the session if valid
func (s *sessionService) ValidateSession(ctx context.Context, token string) (*auth.Session, error) {
	// First validate the JWT token
	claims, err := s.jwtService.ValidateAccessToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Get session from database
	session, err := s.sessionRepo.GetByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// Check if session is active
	if !session.IsActive {
		return nil, errors.New("session is inactive")
	}

	// Check if session is expired
	if session.IsExpired() {
		// Automatically deactivate expired session
		session.IsActive = false
		s.sessionRepo.Update(ctx, session)
		return nil, errors.New("session is expired")
	}

	// Verify session belongs to the user in the token
	if session.UserID != claims.UserID {
		return nil, errors.New("session does not match token user")
	}

	// Update last used timestamp
	session.LastUsedAt = &time.Time{}
	*session.LastUsedAt = time.Now()
	session.UpdatedAt = time.Now()
	
	// Update session (don't fail if this fails)
	s.sessionRepo.Update(ctx, session)

	return session, nil
}

// RefreshSession generates a new access token using refresh token
func (s *sessionService) RefreshSession(ctx context.Context, refreshToken string) (*auth.Session, error) {
	// Validate refresh token
	claims, err := s.jwtService.ValidateRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Get session by refresh token
	session, err := s.sessionRepo.GetByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// Check if session is active
	if !session.IsActive {
		return nil, errors.New("session is inactive")
	}

	// Check if refresh token is expired
	if session.IsRefreshExpired() {
		// Automatically deactivate expired session
		session.IsActive = false
		s.sessionRepo.Update(ctx, session)
		return nil, errors.New("refresh token is expired")
	}

	// Get user to include in new token
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if !user.IsActive {
		return nil, errors.New("user account is inactive")
	}

	// Get user permissions
	var permissions []string

	// Generate new access token
	accessToken, err := s.jwtService.GenerateAccessToken(ctx, user.ID, map[string]interface{}{
		"email":          user.Email,
		"organization_id": user.DefaultOrganizationID,
		"permissions":     permissions,
		"session_id":      session.ID.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Update session with new access token
	config := auth.DefaultTokenConfig()
	session.Token = accessToken
	session.ExpiresAt = time.Now().Add(config.AccessTokenTTL)
	session.UpdatedAt = time.Now()

	err = s.sessionRepo.Update(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	// Log token refresh
	s.auditRepo.Create(ctx, &auth.AuditLog{
		ID:             ulid.New(),
		UserID:         &user.ID,
		OrganizationID: user.DefaultOrganizationID,
		Action:         "session.refreshed",
		Resource:       "session",
		ResourceID:     session.ID.String(),
		CreatedAt:      time.Now(),
	})

	return session, nil
}

// RevokeSession revokes a specific session
func (s *sessionService) RevokeSession(ctx context.Context, sessionID ulid.ULID) error {
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	err = s.sessionRepo.DeactivateSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	// Log session revocation
	s.auditRepo.Create(ctx, &auth.AuditLog{
		ID:        ulid.New(),
		UserID:    &session.UserID,
		Action:    "session.revoked",
		Resource:  "session",
		ResourceID: sessionID.String(),
		CreatedAt: time.Now(),
	})

	return nil
}

// GetUserSessions retrieves all sessions for a user
func (s *sessionService) GetUserSessions(ctx context.Context, userID ulid.ULID) ([]*auth.Session, error) {
	return s.sessionRepo.GetByUserID(ctx, userID)
}

// RevokeUserSessions revokes all sessions for a user
func (s *sessionService) RevokeUserSessions(ctx context.Context, userID ulid.ULID) error {
	err := s.sessionRepo.DeactivateUserSessions(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke user sessions: %w", err)
	}

	// Log mass session revocation
	s.auditRepo.Create(ctx, &auth.AuditLog{
		ID:        ulid.New(),
		UserID:    &userID,
		Action:    "session.revoked_all",
		Resource:  "user",
		ResourceID: userID.String(),
		CreatedAt: time.Now(),
	})

	return nil
}

// CleanupExpiredSessions removes expired sessions from the database
func (s *sessionService) CleanupExpiredSessions(ctx context.Context) error {
	return s.sessionRepo.CleanupExpiredSessions(ctx)
}

// GetActiveSessions retrieves only active sessions for a user
func (s *sessionService) GetActiveSessions(ctx context.Context, userID ulid.ULID) ([]*auth.Session, error) {
	return s.sessionRepo.GetActiveSessionsByUserID(ctx, userID)
}

// MarkSessionAsUsed updates the session's last used timestamp
func (s *sessionService) MarkSessionAsUsed(ctx context.Context, sessionID ulid.ULID) error {
	return s.sessionRepo.MarkAsUsed(ctx, sessionID)
}