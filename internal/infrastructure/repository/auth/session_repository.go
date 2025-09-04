package auth

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"brokle/internal/core/domain/auth"
	"brokle/pkg/ulid"
)

// sessionRepository implements auth.SessionRepository using GORM
type sessionRepository struct {
	db *gorm.DB
}

// NewSessionRepository creates a new session repository instance
func NewSessionRepository(db *gorm.DB) auth.SessionRepository {
	return &sessionRepository{
		db: db,
	}
}

// Create creates a new session
func (r *sessionRepository) Create(ctx context.Context, session *auth.Session) error {
	return r.db.WithContext(ctx).Create(session).Error
}

// GetByID retrieves a session by ID
func (r *sessionRepository) GetByID(ctx context.Context, id ulid.ULID) (*auth.Session, error) {
	var session auth.Session
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("session not found")
		}
		return nil, err
	}
	return &session, nil
}

// GetByToken retrieves a session by access token
func (r *sessionRepository) GetByToken(ctx context.Context, token string) (*auth.Session, error) {
	var session auth.Session
	err := r.db.WithContext(ctx).Where("token = ?", token).First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("session not found")
		}
		return nil, err
	}
	return &session, nil
}

// GetByRefreshToken retrieves a session by refresh token
func (r *sessionRepository) GetByRefreshToken(ctx context.Context, refreshToken string) (*auth.Session, error) {
	var session auth.Session
	err := r.db.WithContext(ctx).Where("refresh_token = ?", refreshToken).First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("session not found")
		}
		return nil, err
	}
	return &session, nil
}

// Update updates a session
func (r *sessionRepository) Update(ctx context.Context, session *auth.Session) error {
	return r.db.WithContext(ctx).Save(session).Error
}

// Delete permanently deletes a session
func (r *sessionRepository) Delete(ctx context.Context, id ulid.ULID) error {
	return r.db.WithContext(ctx).Delete(&auth.Session{}, "id = ?", id.String()).Error
}

// DeactivateSession deactivates a session
func (r *sessionRepository) DeactivateSession(ctx context.Context, sessionID ulid.ULID) error {
	return r.db.WithContext(ctx).
		Model(&auth.Session{}).
		Where("id = ?", sessionID).
		Update("is_active", false).Error
}

// DeactivateUserSessions deactivates all sessions for a user
func (r *sessionRepository) DeactivateUserSessions(ctx context.Context, userID ulid.ULID) error {
	return r.db.WithContext(ctx).
		Model(&auth.Session{}).
		Where("user_id = ?", userID).
		Update("is_active", false).Error
}

// GetByUserID retrieves all sessions for a user
func (r *sessionRepository) GetByUserID(ctx context.Context, userID ulid.ULID) ([]*auth.Session, error) {
	var sessions []*auth.Session
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&sessions).Error
	return sessions, err
}

// GetActiveSessionsByUserID retrieves all active sessions for a user
func (r *sessionRepository) GetActiveSessionsByUserID(ctx context.Context, userID ulid.ULID) ([]*auth.Session, error) {
	var sessions []*auth.Session
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_active = true AND expires_at > ?", userID, time.Now()).
		Order("created_at DESC").
		Find(&sessions).Error
	return sessions, err
}

// CleanupExpiredSessions removes expired sessions
func (r *sessionRepository) CleanupExpiredSessions(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&auth.Session{}).Error
}

// GetSessionStats returns session statistics
func (r *sessionRepository) GetSessionStats(ctx context.Context) (*auth.SessionStats, error) {
	stats := &auth.SessionStats{}

	// Total active sessions
	err := r.db.WithContext(ctx).
		Model(&auth.Session{}).
		Where("is_active = true AND expires_at > ?", time.Now()).
		Count(&stats.ActiveSessions).Error
	if err != nil {
		return nil, err
	}

	// Sessions created today
	today := time.Now().Truncate(24 * time.Hour)
	err = r.db.WithContext(ctx).
		Model(&auth.Session{}).
		Where("created_at >= ?", today).
		Count(&stats.SessionsToday).Error
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// MarkAsUsed updates the last used timestamp for a session
func (r *sessionRepository) MarkAsUsed(ctx context.Context, id ulid.ULID) error {
	return r.db.WithContext(ctx).
		Model(&auth.Session{}).
		Where("id = ?", id).
		Update("last_used_at", time.Now()).Error
}