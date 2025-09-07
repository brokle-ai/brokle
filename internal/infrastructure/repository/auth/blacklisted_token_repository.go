package auth

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"brokle/internal/core/domain/auth"
	"brokle/pkg/ulid"
)

// blacklistedTokenRepository implements auth.BlacklistedTokenRepository using GORM
type blacklistedTokenRepository struct {
	db *gorm.DB
}

// NewBlacklistedTokenRepository creates a new blacklisted token repository instance
func NewBlacklistedTokenRepository(db *gorm.DB) auth.BlacklistedTokenRepository {
	return &blacklistedTokenRepository{
		db: db,
	}
}

// Create adds a new token to the blacklist
func (r *blacklistedTokenRepository) Create(ctx context.Context, blacklistedToken *auth.BlacklistedToken) error {
	return r.db.WithContext(ctx).Create(blacklistedToken).Error
}

// GetByJTI retrieves a blacklisted token by JWT ID
func (r *blacklistedTokenRepository) GetByJTI(ctx context.Context, jti string) (*auth.BlacklistedToken, error) {
	var token auth.BlacklistedToken
	err := r.db.WithContext(ctx).Where("jti = ?", jti).First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("blacklisted token not found")
		}
		return nil, err
	}
	return &token, nil
}

// IsTokenBlacklisted checks if a token is blacklisted (optimized for fast lookup)
func (r *blacklistedTokenRepository) IsTokenBlacklisted(ctx context.Context, jti string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&auth.BlacklistedToken{}).
		Where("jti = ? AND expires_at > ?", jti, time.Now()).
		Count(&count).Error
	
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}

// CleanupExpiredTokens removes tokens that have naturally expired
func (r *blacklistedTokenRepository) CleanupExpiredTokens(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Delete(&auth.BlacklistedToken{}, "expires_at <= ?", time.Now()).
		Error
}

// CleanupTokensOlderThan removes tokens older than specified time
func (r *blacklistedTokenRepository) CleanupTokensOlderThan(ctx context.Context, olderThan time.Time) error {
	return r.db.WithContext(ctx).
		Delete(&auth.BlacklistedToken{}, "created_at < ?", olderThan).
		Error
}

// CreateUserTimestampBlacklist creates a user-wide timestamp blacklist entry for GDPR/SOC2 compliance
func (r *blacklistedTokenRepository) CreateUserTimestampBlacklist(ctx context.Context, userID ulid.ULID, blacklistTimestamp int64, reason string) error {
	blacklistedToken := auth.NewUserTimestampBlacklistedToken(userID, blacklistTimestamp, reason)
	return r.db.WithContext(ctx).Create(blacklistedToken).Error
}

// IsUserBlacklistedAfterTimestamp checks if a user is blacklisted after a specific timestamp
func (r *blacklistedTokenRepository) IsUserBlacklistedAfterTimestamp(ctx context.Context, userID ulid.ULID, tokenIssuedAt int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&auth.BlacklistedToken{}).
		Where("user_id = ? AND token_type = ? AND blacklist_timestamp IS NOT NULL AND ? < blacklist_timestamp AND expires_at > ?", 
			userID, auth.TokenTypeUserTimestamp, tokenIssuedAt, time.Now()).
		Count(&count).Error
	
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}

// GetUserBlacklistTimestamp gets the latest blacklist timestamp for a user
func (r *blacklistedTokenRepository) GetUserBlacklistTimestamp(ctx context.Context, userID ulid.ULID) (*int64, error) {
	var token auth.BlacklistedToken
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND token_type = ? AND blacklist_timestamp IS NOT NULL AND expires_at > ?", 
			userID, auth.TokenTypeUserTimestamp, time.Now()).
		Order("blacklist_timestamp DESC").
		First(&token).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // No user-wide blacklist found
		}
		return nil, err
	}
	
	return token.BlacklistTimestamp, nil
}

// BlacklistUserTokens adds all active tokens for a user to blacklist (legacy method, now replaced by timestamp approach)
func (r *blacklistedTokenRepository) BlacklistUserTokens(ctx context.Context, userID ulid.ULID, reason string) error {
	// DEPRECATED: This method is now replaced by CreateUserTimestampBlacklist for GDPR/SOC2 compliance
	// Create a user-wide timestamp blacklist instead of trying to track individual JTIs
	blacklistTimestamp := time.Now().Unix()
	return r.CreateUserTimestampBlacklist(ctx, userID, blacklistTimestamp, reason)
}

// GetBlacklistedTokensByUser retrieves blacklisted tokens for a specific user
func (r *blacklistedTokenRepository) GetBlacklistedTokensByUser(ctx context.Context, userID ulid.ULID, limit, offset int) ([]*auth.BlacklistedToken, error) {
	var tokens []*auth.BlacklistedToken
	
	query := r.db.WithContext(ctx).Where("user_id = ?", userID).
		Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	err := query.Find(&tokens).Error
	if err != nil {
		return nil, err
	}
	
	return tokens, nil
}

// GetBlacklistedTokensCount returns the total count of blacklisted tokens
func (r *blacklistedTokenRepository) GetBlacklistedTokensCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&auth.BlacklistedToken{}).Count(&count).Error
	return count, err
}

// GetBlacklistedTokensByReason retrieves tokens blacklisted for a specific reason
func (r *blacklistedTokenRepository) GetBlacklistedTokensByReason(ctx context.Context, reason string) ([]*auth.BlacklistedToken, error) {
	var tokens []*auth.BlacklistedToken
	err := r.db.WithContext(ctx).Where("reason = ?", reason).
		Order("created_at DESC").Find(&tokens).Error
	
	if err != nil {
		return nil, err
	}
	
	return tokens, nil
}