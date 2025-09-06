package auth

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"

	"brokle/internal/core/domain/auth"
	"brokle/pkg/ulid"
)

// userSessionRepository implements auth.UserSessionRepository using GORM
type userSessionRepository struct {
	db *gorm.DB
}

// NewUserSessionRepository creates a new user session repository instance
func NewUserSessionRepository(db *gorm.DB) auth.UserSessionRepository {
	return &userSessionRepository{
		db: db,
	}
}

// Create creates a new user session
func (r *userSessionRepository) Create(ctx context.Context, session *auth.UserSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

// GetByID retrieves a user session by ID
func (r *userSessionRepository) GetByID(ctx context.Context, id ulid.ULID) (*auth.UserSession, error) {
	var session auth.UserSession
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user session not found")
		}
		return nil, err
	}
	return &session, nil
}

// GetByToken retrieves a user session by access token
func (r *userSessionRepository) GetByToken(ctx context.Context, token string) (*auth.UserSession, error) {
	var session auth.UserSession
	err := r.db.WithContext(ctx).Where("token = ? AND revoked_at IS NULL", token).First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user session not found")
		}
		return nil, err
	}
	return &session, nil
}

// GetByRefreshToken retrieves a user session by refresh token
func (r *userSessionRepository) GetByRefreshToken(ctx context.Context, refreshToken string) (*auth.UserSession, error) {
	var session auth.UserSession
	err := r.db.WithContext(ctx).Where("refresh_token = ? AND revoked_at IS NULL", refreshToken).First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user session not found")
		}
		return nil, err
	}
	return &session, nil
}

// Update updates an existing user session
func (r *userSessionRepository) Update(ctx context.Context, session *auth.UserSession) error {
	return r.db.WithContext(ctx).Save(session).Error
}

// Delete deletes a user session by ID (hard delete)
func (r *userSessionRepository) Delete(ctx context.Context, id ulid.ULID) error {
	return r.db.WithContext(ctx).Delete(&auth.UserSession{}, "id = ?", id).Error
}

// GetByUserID retrieves all sessions for a user
func (r *userSessionRepository) GetByUserID(ctx context.Context, userID ulid.ULID) ([]*auth.UserSession, error) {
	var sessions []*auth.UserSession
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

// GetActiveSessionsByUserID retrieves active sessions for a user
func (r *userSessionRepository) GetActiveSessionsByUserID(ctx context.Context, userID ulid.ULID) ([]*auth.UserSession, error) {
	var sessions []*auth.UserSession
	err := r.db.WithContext(ctx).Where("user_id = ? AND is_active = ? AND revoked_at IS NULL AND expires_at > ?", userID, true, time.Now()).Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

// DeactivateSession deactivates a session without revoking it
func (r *userSessionRepository) DeactivateSession(ctx context.Context, id ulid.ULID) error {
	return r.db.WithContext(ctx).Model(&auth.UserSession{}).Where("id = ?", id).Update("is_active", false).Error
}

// DeactivateUserSessions deactivates all sessions for a user
func (r *userSessionRepository) DeactivateUserSessions(ctx context.Context, userID ulid.ULID) error {
	return r.db.WithContext(ctx).Model(&auth.UserSession{}).Where("user_id = ?", userID).Update("is_active", false).Error
}

// RevokeSession revokes a session (sets revoked_at timestamp)
func (r *userSessionRepository) RevokeSession(ctx context.Context, id ulid.ULID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&auth.UserSession{}).Where("id = ?", id).Updates(map[string]interface{}{
		"revoked_at": now,
		"is_active":  false,
		"updated_at": now,
	}).Error
}

// RevokeUserSessions revokes all sessions for a user
func (r *userSessionRepository) RevokeUserSessions(ctx context.Context, userID ulid.ULID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&auth.UserSession{}).Where("user_id = ? AND revoked_at IS NULL", userID).Updates(map[string]interface{}{
		"revoked_at": now,
		"is_active":  false,
		"updated_at": now,
	}).Error
}

// CleanupExpiredSessions removes expired sessions
func (r *userSessionRepository) CleanupExpiredSessions(ctx context.Context) error {
	return r.db.WithContext(ctx).Delete(&auth.UserSession{}, "expires_at < ?", time.Now()).Error
}

// CleanupRevokedSessions removes revoked sessions older than 30 days
func (r *userSessionRepository) CleanupRevokedSessions(ctx context.Context) error {
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	return r.db.WithContext(ctx).Delete(&auth.UserSession{}, "revoked_at IS NOT NULL AND revoked_at < ?", thirtyDaysAgo).Error
}

// MarkAsUsed updates the last_used_at timestamp
func (r *userSessionRepository) MarkAsUsed(ctx context.Context, id ulid.ULID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&auth.UserSession{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_used_at": now,
		"updated_at":   now,
	}).Error
}

// GetByDeviceInfo retrieves sessions by user ID and device info
func (r *userSessionRepository) GetByDeviceInfo(ctx context.Context, userID ulid.ULID, deviceInfo interface{}) ([]*auth.UserSession, error) {
	var sessions []*auth.UserSession
	
	// Convert device info to JSON for comparison
	deviceJSON, err := json.Marshal(deviceInfo)
	if err != nil {
		return nil, err
	}

	err = r.db.WithContext(ctx).Where("user_id = ? AND device_info::jsonb = ?::jsonb", userID, string(deviceJSON)).Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

// GetActiveSessionsCount returns the count of active sessions for a user
func (r *userSessionRepository) GetActiveSessionsCount(ctx context.Context, userID ulid.ULID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&auth.UserSession{}).Where("user_id = ? AND is_active = ? AND revoked_at IS NULL AND expires_at > ?", userID, true, time.Now()).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return int(count), nil
}