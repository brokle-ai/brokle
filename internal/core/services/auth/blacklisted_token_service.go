package auth

import (
	"context"
	"fmt"
	"time"

	"brokle/internal/core/domain/auth"
	"brokle/pkg/ulid"
)

// blacklistedTokenService implements the auth.BlacklistedTokenService interface
type blacklistedTokenService struct {
	blacklistedTokenRepo auth.BlacklistedTokenRepository
	auditRepo            auth.AuditLogRepository
}

// NewBlacklistedTokenService creates a new blacklisted token service instance
func NewBlacklistedTokenService(
	blacklistedTokenRepo auth.BlacklistedTokenRepository,
	auditRepo auth.AuditLogRepository,
) auth.BlacklistedTokenService {
	return &blacklistedTokenService{
		blacklistedTokenRepo: blacklistedTokenRepo,
		auditRepo:            auditRepo,
	}
}

// BlacklistToken adds a token to the blacklist for immediate revocation
func (s *blacklistedTokenService) BlacklistToken(ctx context.Context, jti string, userID ulid.ULID, expiresAt time.Time, reason string) error {
	// Check if token is already blacklisted
	isBlacklisted, err := s.blacklistedTokenRepo.IsTokenBlacklisted(ctx, jti)
	if err != nil {
		return fmt.Errorf("failed to check token blacklist status: %w", err)
	}
	
	if isBlacklisted {
		// Token is already blacklisted, no need to add again
		return nil
	}

	// Create blacklisted token entry
	blacklistedToken := auth.NewBlacklistedToken(jti, userID, expiresAt, reason)
	
	// Add to blacklist
	err = s.blacklistedTokenRepo.Create(ctx, blacklistedToken)
	if err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	// Log the blacklisting action
	auditLog := auth.NewAuditLog(
		&userID, 
		nil, 
		"token.blacklisted", 
		"token", 
		jti, 
		fmt.Sprintf(`{"reason": "%s", "expires_at": "%s"}`, reason, expiresAt.Format(time.RFC3339)),
		"", 
		"",
	)
	s.auditRepo.Create(ctx, auditLog) // Don't fail if audit logging fails

	return nil
}

// IsTokenBlacklisted checks if a token is blacklisted (optimized for fast lookup)
func (s *blacklistedTokenService) IsTokenBlacklisted(ctx context.Context, jti string) (bool, error) {
	return s.blacklistedTokenRepo.IsTokenBlacklisted(ctx, jti)
}

// GetBlacklistedToken retrieves a blacklisted token by JTI
func (s *blacklistedTokenService) GetBlacklistedToken(ctx context.Context, jti string) (*auth.BlacklistedToken, error) {
	return s.blacklistedTokenRepo.GetByJTI(ctx, jti)
}

// BlacklistUserTokens blacklists all active tokens for a user (emergency revocation)
func (s *blacklistedTokenService) BlacklistUserTokens(ctx context.Context, userID ulid.ULID, reason string) error {
	// Use repository method to blacklist all user tokens
	err := s.blacklistedTokenRepo.BlacklistUserTokens(ctx, userID, reason)
	if err != nil {
		return fmt.Errorf("failed to blacklist user tokens: %w", err)
	}

	// Log the bulk blacklisting action
	auditLog := auth.NewAuditLog(
		&userID, 
		nil, 
		"token.bulk_blacklisted", 
		"user", 
		userID.String(), 
		fmt.Sprintf(`{"reason": "%s", "action": "bulk_revocation"}`, reason),
		"", 
		"",
	)
	s.auditRepo.Create(ctx, auditLog) // Don't fail if audit logging fails

	return nil
}

// GetUserBlacklistedTokens retrieves blacklisted tokens for a specific user
func (s *blacklistedTokenService) GetUserBlacklistedTokens(ctx context.Context, userID ulid.ULID, limit, offset int) ([]*auth.BlacklistedToken, error) {
	return s.blacklistedTokenRepo.GetBlacklistedTokensByUser(ctx, userID, limit, offset)
}

// CleanupExpiredTokens removes naturally expired tokens from blacklist
func (s *blacklistedTokenService) CleanupExpiredTokens(ctx context.Context) error {
	err := s.blacklistedTokenRepo.CleanupExpiredTokens(ctx)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired tokens: %w", err)
	}

	// Log cleanup action
	auditLog := auth.NewAuditLog(
		nil, 
		nil, 
		"token.cleanup_expired", 
		"system", 
		"blacklisted_tokens", 
		`{"action": "cleanup", "type": "expired"}`,
		"", 
		"",
	)
	s.auditRepo.Create(ctx, auditLog) // Don't fail if audit logging fails

	return nil
}

// CleanupOldTokens removes tokens older than specified time
func (s *blacklistedTokenService) CleanupOldTokens(ctx context.Context, olderThan time.Time) error {
	err := s.blacklistedTokenRepo.CleanupTokensOlderThan(ctx, olderThan)
	if err != nil {
		return fmt.Errorf("failed to cleanup old tokens: %w", err)
	}

	// Log cleanup action
	auditLog := auth.NewAuditLog(
		nil, 
		nil, 
		"token.cleanup_old", 
		"system", 
		"blacklisted_tokens", 
		fmt.Sprintf(`{"action": "cleanup", "type": "old", "older_than": "%s"}`, olderThan.Format(time.RFC3339)),
		"", 
		"",
	)
	s.auditRepo.Create(ctx, auditLog) // Don't fail if audit logging fails

	return nil
}

// GetBlacklistedTokensCount returns total count of blacklisted tokens
func (s *blacklistedTokenService) GetBlacklistedTokensCount(ctx context.Context) (int64, error) {
	return s.blacklistedTokenRepo.GetBlacklistedTokensCount(ctx)
}

// GetTokensByReason retrieves tokens blacklisted for a specific reason
func (s *blacklistedTokenService) GetTokensByReason(ctx context.Context, reason string) ([]*auth.BlacklistedToken, error) {
	return s.blacklistedTokenRepo.GetBlacklistedTokensByReason(ctx, reason)
}