package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"brokle/internal/config"
	"brokle/internal/core/domain/auth"
	"brokle/internal/core/domain/user"
	"brokle/pkg/ulid"
)


// sessionService implements the auth.SessionService interface
type sessionService struct {
	authConfig  *config.AuthConfig
	sessionRepo auth.UserSessionRepository
	userRepo    user.Repository
	jwtService  auth.JWTService
	auditRepo   auth.AuditLogRepository
}

// NewSessionService creates a new session service instance
func NewSessionService(
	authConfig *config.AuthConfig,
	sessionRepo auth.UserSessionRepository,
	userRepo user.Repository,
	jwtService auth.JWTService,
	auditRepo auth.AuditLogRepository,
) auth.SessionService {
	return &sessionService{
		authConfig:  authConfig,
		sessionRepo: sessionRepo,
		userRepo:    userRepo,
		jwtService:  jwtService,
		auditRepo:   auditRepo,
	}
}


// GetSession retrieves a session by ID
func (s *sessionService) GetSession(ctx context.Context, sessionID ulid.ULID) (*auth.UserSession, error) {
	return s.sessionRepo.GetByID(ctx, sessionID)
}



// RevokeSession revokes a specific session
func (s *sessionService) RevokeSession(ctx context.Context, sessionID ulid.ULID) error {
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	err = s.sessionRepo.RevokeSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	// Log session revocation
	auditLog := auth.NewAuditLog(&session.UserID, nil, "session.revoked", "session", sessionID.String(), "", "", "")
	s.auditRepo.Create(ctx, auditLog)

	return nil
}

// GetUserSessions retrieves all sessions for a user
func (s *sessionService) GetUserSessions(ctx context.Context, userID ulid.ULID) ([]*auth.UserSession, error) {
	return s.sessionRepo.GetByUserID(ctx, userID)
}

// RevokeUserSessions revokes all sessions for a user
func (s *sessionService) RevokeUserSessions(ctx context.Context, userID ulid.ULID) error {
	err := s.sessionRepo.RevokeUserSessions(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke user sessions: %w", err)
	}

	// Log mass session revocation
	auditLog := auth.NewAuditLog(&userID, nil, "session.revoked_all", "user", userID.String(), "", "", "")
	s.auditRepo.Create(ctx, auditLog)

	return nil
}

// CleanupExpiredSessions removes expired sessions from the database
func (s *sessionService) CleanupExpiredSessions(ctx context.Context) error {
	return s.sessionRepo.CleanupExpiredSessions(ctx)
}

// GetActiveSessions retrieves only active sessions for a user
func (s *sessionService) GetActiveSessions(ctx context.Context, userID ulid.ULID) ([]*auth.UserSession, error) {
	return s.sessionRepo.GetActiveSessionsByUserID(ctx, userID)
}


// hashToken creates a SHA-256 hash of a token for secure storage
func (s *sessionService) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}