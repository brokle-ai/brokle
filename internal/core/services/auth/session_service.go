package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/google/uuid"

	"brokle/internal/config"
	authDomain "brokle/internal/core/domain/auth"
	userDomain "brokle/internal/core/domain/user"
	appErrors "brokle/pkg/errors"
)

// SessionService manages user sessions: lookup, revocation, and cleanup.
type SessionService struct {
	cfg         *config.AuthConfig
	sessionRepo authDomain.UserSessionRepository
	userRepo    userDomain.Repository
	jwt         *JWTService
}

// NewSessionService wires a SessionService with its dependencies.
func NewSessionService(
	cfg *config.AuthConfig,
	sessionRepo authDomain.UserSessionRepository,
	userRepo userDomain.Repository,
	jwt *JWTService,
) *SessionService {
	return &SessionService{
		cfg:         cfg,
		sessionRepo: sessionRepo,
		userRepo:    userRepo,
		jwt:         jwt,
	}
}

// GetSession retrieves a session by ID
func (s *SessionService) GetSession(ctx context.Context, sessionID uuid.UUID) (*authDomain.UserSession, error) {
	return s.sessionRepo.GetByID(ctx, sessionID)
}

// RevokeSession revokes a specific session
func (s *SessionService) RevokeSession(ctx context.Context, sessionID uuid.UUID) error {
	_, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return appErrors.NewNotFoundError("session not found")
	}

	err = s.sessionRepo.RevokeSession(ctx, sessionID)
	if err != nil {
		return appErrors.NewInternalError("failed to revoke session", err)
	}

	return nil
}

// GetUserSessions retrieves all sessions for a user
func (s *SessionService) GetUserSessions(ctx context.Context, userID uuid.UUID) ([]*authDomain.UserSession, error) {
	return s.sessionRepo.GetByUserID(ctx, userID)
}

// RevokeUserSessions revokes all sessions for a user
func (s *SessionService) RevokeUserSessions(ctx context.Context, userID uuid.UUID) error {
	err := s.sessionRepo.RevokeUserSessions(ctx, userID)
	if err != nil {
		return appErrors.NewInternalError("failed to revoke user sessions", err)
	}

	return nil
}

// CleanupExpiredSessions removes expired sessions from the database
func (s *SessionService) CleanupExpiredSessions(ctx context.Context) error {
	return s.sessionRepo.CleanupExpiredSessions(ctx)
}

// GetActiveSessions retrieves only active sessions for a user
func (s *SessionService) GetActiveSessions(ctx context.Context, userID uuid.UUID) ([]*authDomain.UserSession, error) {
	return s.sessionRepo.GetActiveSessionsByUserID(ctx, userID)
}

// hashToken creates a SHA-256 hash of a token for secure storage
func (s *SessionService) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
