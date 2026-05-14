package auth

import (
	"time"

	"github.com/google/uuid"

	"brokle/pkg/uid"
)

// UserSession represents an active user session with secure token management.
// SECURITY: Access tokens are NOT stored — only session metadata and hashed
// refresh tokens. The current JTI is stored to support immediate revocation
// of in-flight access tokens via the blacklist.
type UserSession struct {
	ExpiresAt           time.Time      `json:"expires_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	CreatedAt           time.Time      `json:"created_at"`
	RefreshExpiresAt    time.Time      `json:"refresh_expires_at"`
	DeviceInfo          map[string]any `json:"device_info,omitempty"`
	IPAddress           *string        `json:"ip_address,omitempty"`
	UserAgent           *string        `json:"user_agent,omitempty"`
	LastUsedAt          *time.Time     `json:"last_used_at,omitempty"`
	RevokedAt           *time.Time     `json:"revoked_at,omitempty"`
	CurrentJTI          string         `json:"-"`
	RefreshTokenHash    string         `json:"-"`
	RefreshTokenVersion int            `json:"refresh_token_version"`
	ID                  uuid.UUID      `json:"id"`
	UserID              uuid.UUID      `json:"user_id"`
	IsActive            bool           `json:"is_active"`
}

// IsExpired reports whether the session's access-token window has elapsed.
func (s *UserSession) IsExpired() bool { return time.Now().After(s.ExpiresAt) }

// IsRefreshExpired reports whether the session's refresh-token window has elapsed.
func (s *UserSession) IsRefreshExpired() bool { return time.Now().After(s.RefreshExpiresAt) }

// IsValid reports whether the session is currently usable.
func (s *UserSession) IsValid() bool {
	return s.IsActive && !s.IsExpired() && s.RevokedAt == nil
}

// MarkAsUsed updates the last-used timestamp.
func (s *UserSession) MarkAsUsed() {
	now := time.Now()
	s.LastUsedAt = &now
	s.UpdatedAt = now
}

// Revoke deactivates the session and timestamps the revocation.
func (s *UserSession) Revoke() {
	now := time.Now()
	s.RevokedAt = &now
	s.IsActive = false
	s.UpdatedAt = now
}

// Deactivate flags the session as inactive without recording a revocation.
func (s *UserSession) Deactivate() {
	s.IsActive = false
	s.UpdatedAt = time.Now()
}

// NewUserSession constructs a session record for the given identity and tokens.
func NewUserSession(userID uuid.UUID, refreshTokenHash string, currentJTI string, expiresAt, refreshExpiresAt time.Time, ipAddress, userAgent *string, deviceInfo map[string]any) *UserSession {
	return &UserSession{
		ID:                  uid.New(),
		UserID:              userID,
		RefreshTokenHash:    refreshTokenHash,
		RefreshTokenVersion: 1,
		CurrentJTI:          currentJTI,
		ExpiresAt:           expiresAt,
		RefreshExpiresAt:    refreshExpiresAt,
		IPAddress:           ipAddress,
		UserAgent:           userAgent,
		DeviceInfo:          deviceInfo,
		IsActive:            true,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
}

// BlacklistedToken represents a revoked access token for immediate
// revocation. Two flavours: per-JTI ("individual") and user-wide
// timestamp-based ("user_wide_timestamp"); the latter is used for
// GDPR/SOC2 "revoke all sessions for this user" operations.
type BlacklistedToken struct {
	ExpiresAt          time.Time `json:"expires_at"`
	RevokedAt          time.Time `json:"revoked_at"`
	CreatedAt          time.Time `json:"created_at"`
	BlacklistTimestamp *int64    `json:"blacklist_timestamp,omitempty"`
	JTI                string    `json:"jti"`
	Reason             string    `json:"reason"`
	TokenType          string    `json:"token_type"`
	UserID             uuid.UUID `json:"user_id"`
}

// Blacklisted-token type discriminators.
const (
	TokenTypeIndividual    = "individual"          // Individual JTI-based blacklisting (default)
	TokenTypeUserTimestamp = "user_wide_timestamp" // User-wide timestamp blacklisting (GDPR/SOC2)
)

// NewBlacklistedToken constructs an individual JTI blacklist entry.
func NewBlacklistedToken(jti string, userID uuid.UUID, expiresAt time.Time, reason string) *BlacklistedToken {
	return &BlacklistedToken{
		JTI:       jti,
		UserID:    userID,
		ExpiresAt: expiresAt,
		RevokedAt: time.Now(),
		Reason:    reason,
		TokenType: TokenTypeIndividual,
		CreatedAt: time.Now(),
	}
}

// NewUserTimestampBlacklistedToken creates a user-wide timestamp
// blacklist entry for GDPR/SOC2 compliance. ExpiresAt is set 24 hours
// past the blacklist timestamp to ensure cleanup eventually drops the
// entry; tokens issued before the timestamp are rejected on validation.
func NewUserTimestampBlacklistedToken(userID uuid.UUID, blacklistTimestamp int64, reason string) *BlacklistedToken {
	userWideJTI := uid.New()
	farFutureExpiry := time.Unix(blacklistTimestamp, 0).Add(24 * time.Hour)

	return &BlacklistedToken{
		JTI:                userWideJTI.String(),
		UserID:             userID,
		ExpiresAt:          farFutureExpiry,
		RevokedAt:          time.Now(),
		Reason:             reason,
		TokenType:          TokenTypeUserTimestamp,
		BlacklistTimestamp: &blacklistTimestamp,
		CreatedAt:          time.Now(),
	}
}

// SessionStats represents aggregate session statistics surfaced by the
// platform health dashboard.
type SessionStats struct {
	ActiveSessions   int64 `json:"active_sessions"`
	ExpiredSessions  int64 `json:"expired_sessions"`
	TotalSessions    int64 `json:"total_sessions"`
	SessionsToday    int64 `json:"sessions_today"`
	SessionsThisWeek int64 `json:"sessions_this_week"`
	AvgSessionLength int64 `json:"avg_session_length_minutes"`
}

// PasswordResetToken represents a password reset token. The plaintext
// token is JSON-omitted so it never round-trips to clients on read.
type PasswordResetToken struct {
	ExpiresAt time.Time  `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	Token     string     `json:"-"`
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
}

// NewPasswordResetToken constructs a reset-token record. The caller is
// responsible for generating the token string (typically random + base32).
func NewPasswordResetToken(userID uuid.UUID, token string, expiresAt time.Time) *PasswordResetToken {
	return &PasswordResetToken{
		ID:        uid.New(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// ----- Wire DTOs for the auth flow ----------------------------------

// LoginRequest is the body of POST /api/v1/auth/login.
type LoginRequest struct {
	DeviceInfo map[string]any `json:"device_info,omitempty"`
	Email      string         `json:"email" validate:"required,email"`
	Password   string         `json:"password" validate:"required"`
}

// LoginResponse is the success body of POST /api/v1/auth/login and the
// refresh-token rotation endpoint.
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"` // Always "Bearer"
	ExpiresIn    int64  `json:"expires_in"` // Seconds until access-token expiration
}

// AuthUser is the lightweight identity shape returned alongside
// LoginResponse for the dashboard's bootstrap.
type AuthUser struct {
	AvatarURL             *string    `json:"avatar_url,omitempty"`
	DefaultOrganizationID *uuid.UUID `json:"default_organization_id,omitempty"`
	Email                 string     `json:"email"`
	Name                  string     `json:"name"`
	ID                    uuid.UUID  `json:"id"`
	IsEmailVerified       bool       `json:"is_email_verified"`
}

// RefreshTokenRequest is the body of POST /api/v1/auth/refresh.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// AuthContext represents the authenticated principal carried via
// request context (httpctx) — UserID is always present; APIKeyID is
// non-nil only on SDK-plane requests; SessionID is non-nil only on
// dashboard-plane requests.
type AuthContext struct {
	APIKeyID  *uuid.UUID `json:"api_key_id,omitempty"`
	SessionID *uuid.UUID `json:"session_id,omitempty"`
	UserID    uuid.UUID  `json:"user_id"`
}

// OAuthSession holds the temporary state of an in-progress OAuth
// signup. Persisted in Redis for ~15 minutes between provider callback
// and the frontend confirmation step. Not a database entity.
type OAuthSession struct {
	ExpiresAt       time.Time `json:"expires_at"`
	InvitationToken *string   `json:"invitation_token,omitempty"`
	Email           string    `json:"email"`
	FirstName       string    `json:"first_name"`
	LastName        string    `json:"last_name"`
	Provider        string    `json:"provider"`
	ProviderID      string    `json:"provider_id"`
}

// LoginTokenSession holds a one-time-use redirect payload returned to
// the frontend after an OAuth login completes. The session is deleted
// on first read to prevent replay.
type LoginTokenSession struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int64     `json:"expires_in"`
	UserID       uuid.UUID `json:"user_id"`
	CreatedAt    time.Time `json:"created_at"`
}
