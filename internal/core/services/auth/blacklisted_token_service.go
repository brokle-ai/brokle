package auth

import (
	"context"
	"time"

	authDomain "brokle/internal/core/domain/auth"
	"brokle/pkg/ulid"
	appErrors "brokle/pkg/errors"
)

// blacklistedTokenService implements the auth.BlacklistedTokenService interface
type blacklistedTokenService struct {
	blacklistedTokenRepo authDomain.BlacklistedTokenRepository
}

// NewBlacklistedTokenService creates a new blacklisted token service instance
func NewBlacklistedTokenService(
	blacklistedTokenRepo authDomain.BlacklistedTokenRepository,
) authDomain.BlacklistedTokenService {
	return &blacklistedTokenService{
		blacklistedTokenRepo: blacklistedTokenRepo,
	}
}

// BlacklistToken adds a token to the blacklist for immediate revocation
func (s *blacklistedTokenService) BlacklistToken(ctx context.Context, jti string, userID ulid.ULID, expiresAt time.Time, reason string) error {
	// Check if token is already blacklisted
	isBlacklisted, err := s.blacklistedTokenRepo.IsTokenBlacklisted(ctx, jti)
	if err != nil {
		return appErrors.NewInternalError("Failed to check token blacklist status", err)
	}
	
	if isBlacklisted {
		// Token is already blacklisted, no need to add again
		return nil
	}

	// Create blacklisted token entry
	blacklistedToken := authDomain.NewBlacklistedToken(jti, userID, expiresAt, reason)
	
	// Add to blacklist
	err = s.blacklistedTokenRepo.Create(ctx, blacklistedToken)
	if err != nil {
		return appErrors.NewInternalError("Failed to blacklist token", err)
	}


	return nil
}

// IsTokenBlacklisted checks if a token is blacklisted (optimized for fast lookup)
func (s *blacklistedTokenService) IsTokenBlacklisted(ctx context.Context, jti string) (bool, error) {
	return s.blacklistedTokenRepo.IsTokenBlacklisted(ctx, jti)
}

// GetBlacklistedToken retrieves a blacklisted token by JTI
func (s *blacklistedTokenService) GetBlacklistedToken(ctx context.Context, jti string) (*authDomain.BlacklistedToken, error) {
	return s.blacklistedTokenRepo.GetByJTI(ctx, jti)
}

// CreateUserTimestampBlacklist creates a user-wide timestamp blacklist for GDPR/SOC2 compliance
func (s *blacklistedTokenService) CreateUserTimestampBlacklist(ctx context.Context, userID ulid.ULID, reason string) error {
	blacklistTimestamp := time.Now().Unix()
	
	err := s.blacklistedTokenRepo.CreateUserTimestampBlacklist(ctx, userID, blacklistTimestamp, reason)
	if err != nil {
		return appErrors.NewInternalError("Failed to create user timestamp blacklist", err)
	}


	return nil
}

// IsUserBlacklistedAfterTimestamp checks if a user is blacklisted after a specific timestamp
func (s *blacklistedTokenService) IsUserBlacklistedAfterTimestamp(ctx context.Context, userID ulid.ULID, tokenIssuedAt int64) (bool, error) {
	return s.blacklistedTokenRepo.IsUserBlacklistedAfterTimestamp(ctx, userID, tokenIssuedAt)
}

// GetUserBlacklistTimestamp gets the latest blacklist timestamp for a user
func (s *blacklistedTokenService) GetUserBlacklistTimestamp(ctx context.Context, userID ulid.ULID) (*int64, error) {
	return s.blacklistedTokenRepo.GetUserBlacklistTimestamp(ctx, userID)
}

// BlacklistUserTokens blacklists all active tokens for a user (now uses timestamp approach for GDPR/SOC2 compliance)
func (s *blacklistedTokenService) BlacklistUserTokens(ctx context.Context, userID ulid.ULID, reason string) error {
	// Use the new timestamp approach for comprehensive GDPR/SOC2 compliance
	return s.CreateUserTimestampBlacklist(ctx, userID, reason)
}

// GetUserBlacklistedTokens retrieves blacklisted tokens for a specific user
func (s *blacklistedTokenService) GetUserBlacklistedTokens(ctx context.Context, userID ulid.ULID, limit, offset int) ([]*authDomain.BlacklistedToken, error) {
	return s.blacklistedTokenRepo.GetBlacklistedTokensByUser(ctx, userID, limit, offset)
}

// CleanupExpiredTokens removes naturally expired tokens from blacklist
func (s *blacklistedTokenService) CleanupExpiredTokens(ctx context.Context) error {
	err := s.blacklistedTokenRepo.CleanupExpiredTokens(ctx)
	if err != nil {
		return appErrors.NewInternalError("Failed to cleanup expired tokens", err)
	}


	return nil
}

// CleanupOldTokens removes tokens older than specified time
func (s *blacklistedTokenService) CleanupOldTokens(ctx context.Context, olderThan time.Time) error {
	err := s.blacklistedTokenRepo.CleanupTokensOlderThan(ctx, olderThan)
	if err != nil {
		return appErrors.NewInternalError("Failed to cleanup old tokens", err)
	}


	return nil
}

// GetBlacklistedTokensCount returns total count of blacklisted tokens
func (s *blacklistedTokenService) GetBlacklistedTokensCount(ctx context.Context) (int64, error) {
	return s.blacklistedTokenRepo.GetBlacklistedTokensCount(ctx)
}

// GetTokensByReason retrieves tokens blacklisted for a specific reason
func (s *blacklistedTokenService) GetTokensByReason(ctx context.Context, reason string) ([]*authDomain.BlacklistedToken, error) {
	return s.blacklistedTokenRepo.GetBlacklistedTokensByReason(ctx, reason)
}