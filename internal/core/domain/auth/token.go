package auth

import (
	"time"

	"brokle/internal/core/domain/user"
	"brokle/pkg/ulid"
)

// TokenType represents different types of JWT tokens
type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
	TokenTypeAPIKey  TokenType = "api_key"
	TokenTypeInvite  TokenType = "invite"
	TokenTypeReset   TokenType = "reset"
	TokenTypeVerify  TokenType = "verify"
)

// JWTClaims represents the standard JWT claims structure used across the platform
type JWTClaims struct {
	// Standard JWT claims
	Issuer    string `json:"iss"`
	Subject   string `json:"sub"`
	Audience  string `json:"aud,omitempty"`
	ExpiresAt int64  `json:"exp"`
	NotBefore int64  `json:"nbf"`
	IssuedAt  int64  `json:"iat"`
	JWTID     string `json:"jti"`

	// Custom claims
	TokenType TokenType  `json:"token_type"`
	UserID    ulid.ULID  `json:"user_id"`
	Email     string     `json:"email"`
	
	// Context claims
	OrganizationID *ulid.ULID `json:"organization_id,omitempty"`
	ProjectID      *ulid.ULID `json:"project_id,omitempty"`
	EnvironmentID  *ulid.ULID `json:"environment_id,omitempty"`
	
	// Permission claims
	Scopes      []string `json:"scopes,omitempty"`      // For API keys
	Permissions []string `json:"permissions,omitempty"` // User permissions in org
	Role        *string  `json:"role,omitempty"`        // User role name
	
	// API Key specific claims
	APIKeyID *ulid.ULID `json:"api_key_id,omitempty"`
	
	// Session specific claims
	SessionID *ulid.ULID `json:"session_id,omitempty"`
	
	// Security claims
	IPAddress *string `json:"ip_address,omitempty"` // For IP-bound tokens
	UserAgent *string `json:"user_agent,omitempty"` // For device-bound tokens
}

// TokenConfig represents configuration for JWT tokens
type TokenConfig struct {
	// Signing configuration
	SigningKey    string        `json:"-"` // Secret key for signing
	SigningMethod string        `json:"signing_method"`
	Issuer        string        `json:"issuer"`
	
	// Token lifetimes
	AccessTokenTTL  time.Duration `json:"access_token_ttl"`
	RefreshTokenTTL time.Duration `json:"refresh_token_ttl"`
	APIKeyTokenTTL  time.Duration `json:"api_key_token_ttl"`
	InviteTokenTTL  time.Duration `json:"invite_token_ttl"`
	ResetTokenTTL   time.Duration `json:"reset_token_ttl"`
	VerifyTokenTTL  time.Duration `json:"verify_token_ttl"`
	
	// Security settings
	RequireAudience bool `json:"require_audience"`
	AllowedAudiences []string `json:"allowed_audiences"`
	ClockSkew       time.Duration `json:"clock_skew"`
}

// DefaultTokenConfig returns the default token configuration
func DefaultTokenConfig() *TokenConfig {
	return &TokenConfig{
		SigningMethod:   "HS256",
		Issuer:         "brokle-platform",
		AccessTokenTTL: 15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour, // 7 days
		APIKeyTokenTTL:  24 * time.Hour, // API keys generate short-lived access tokens
		InviteTokenTTL:  7 * 24 * time.Hour, // 7 days
		ResetTokenTTL:   1 * time.Hour,
		VerifyTokenTTL:  24 * time.Hour,
		ClockSkew:      5 * time.Minute,
		RequireAudience: false,
	}
}

// TokenValidationResult represents the result of token validation
type TokenValidationResult struct {
	Valid       bool          `json:"valid"`
	Claims      *JWTClaims    `json:"claims,omitempty"`
	Error       string        `json:"error,omitempty"`
	ErrorCode   string        `json:"error_code,omitempty"`
	ExpiresIn   time.Duration `json:"expires_in,omitempty"` // Time until expiration
	TokenType   TokenType     `json:"token_type,omitempty"`
}

// TokenGenerationRequest represents a request to generate a new token
type TokenGenerationRequest struct {
	TokenType      TokenType              `json:"token_type"`
	UserID         ulid.ULID              `json:"user_id"`
	Email          string                 `json:"email"`
	OrganizationID *ulid.ULID             `json:"organization_id,omitempty"`
	ProjectID      *ulid.ULID             `json:"project_id,omitempty"`
	EnvironmentID  *ulid.ULID             `json:"environment_id,omitempty"`
	Scopes         []string               `json:"scopes,omitempty"`
	Permissions    []string               `json:"permissions,omitempty"`
	Role           *string                `json:"role,omitempty"`
	APIKeyID       *ulid.ULID             `json:"api_key_id,omitempty"`
	SessionID      *ulid.ULID             `json:"session_id,omitempty"`
	IPAddress      *string                `json:"ip_address,omitempty"`
	UserAgent      *string                `json:"user_agent,omitempty"`
	CustomClaims   map[string]interface{} `json:"custom_claims,omitempty"`
	TTL            *time.Duration         `json:"ttl,omitempty"` // Override default TTL
}

// NewJWTClaims creates a new JWT claims structure with default values
func NewJWTClaims(req *TokenGenerationRequest) *JWTClaims {
	now := time.Now()
	
	return &JWTClaims{
		Issuer:         "brokle-platform",
		Subject:        req.UserID.String(),
		JWTID:          ulid.New().String(),
		IssuedAt:       now.Unix(),
		NotBefore:      now.Unix(),
		TokenType:      req.TokenType,
		UserID:         req.UserID,
		Email:          req.Email,
		OrganizationID: req.OrganizationID,
		ProjectID:      req.ProjectID,
		EnvironmentID:  req.EnvironmentID,
		Scopes:         req.Scopes,
		Permissions:    req.Permissions,
		Role:           req.Role,
		APIKeyID:       req.APIKeyID,
		SessionID:      req.SessionID,
		IPAddress:      req.IPAddress,
		UserAgent:      req.UserAgent,
	}
}

// IsExpired checks if the token is expired
func (c *JWTClaims) IsExpired() bool {
	return time.Now().Unix() > c.ExpiresAt
}

// IsValidNow checks if the token is valid at the current time (not expired, not before)
func (c *JWTClaims) IsValidNow() bool {
	now := time.Now().Unix()
	return now >= c.NotBefore && now < c.ExpiresAt
}

// TimeUntilExpiry returns the duration until the token expires
func (c *JWTClaims) TimeUntilExpiry() time.Duration {
	return time.Until(time.Unix(c.ExpiresAt, 0))
}

// GetUserContext returns the user context from the token claims
func (c *JWTClaims) GetUserContext() *AuthContext {
	return &AuthContext{
		UserID:         c.UserID,
		OrganizationID: c.OrganizationID,
		Role:           c.Role,
		Permissions:    c.Permissions,
		APIKeyID:       c.APIKeyID,
		SessionID:      c.SessionID,
	}
}

// HasScope checks if the token has a specific scope (for API keys)
func (c *JWTClaims) HasScope(scope string) bool {
	for _, s := range c.Scopes {
		if s == scope || s == "*" {
			return true
		}
		// Check wildcard scopes (e.g., "users.*" matches "users.read")
		if len(s) > 0 && s[len(s)-1] == '*' {
			prefix := s[:len(s)-1]
			if len(scope) > len(prefix) && scope[:len(prefix)] == prefix {
				return true
			}
		}
	}
	return false
}

// HasPermission checks if the token has a specific permission
func (c *JWTClaims) HasPermission(permission string) bool {
	for _, p := range c.Permissions {
		if p == permission || p == "*" {
			return true
		}
		// Check wildcard permissions
		if len(p) > 0 && p[len(p)-1] == '*' {
			prefix := p[:len(p)-1]
			if len(permission) > len(prefix) && permission[:len(prefix)] == prefix {
				return true
			}
		}
	}
	return false
}

// PasswordResetToken represents a password reset token
type PasswordResetToken struct {
	ID        ulid.ULID  `json:"id" gorm:"type:char(26);primaryKey"`
	UserID    ulid.ULID  `json:"user_id" gorm:"type:char(26);not null"`
	Token     string     `json:"token" gorm:"size:255;not null;uniqueIndex"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"` // Replaced Used bool with UsedAt timestamp
	
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	// Relations
	User user.User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// EmailVerificationToken represents an email verification token
type EmailVerificationToken struct {
	ID        ulid.ULID `json:"id" gorm:"type:char(26);primaryKey"`
	UserID    ulid.ULID `json:"user_id" gorm:"type:char(26);not null"`
	Token     string    `json:"token" gorm:"size:255;not null;uniqueIndex"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used" gorm:"default:false"`
	
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	// Relations
	User user.User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// Token creation helpers
func NewPasswordResetToken(userID ulid.ULID, token string, expiresAt time.Time) *PasswordResetToken {
	return &PasswordResetToken{
		ID:        ulid.New(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		UsedAt:    nil, // Not used initially
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func NewEmailVerificationToken(userID ulid.ULID, token string, expiresAt time.Time) *EmailVerificationToken {
	return &EmailVerificationToken{
		ID:        ulid.New(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		Used:      false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Token validation methods
func (t *PasswordResetToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

func (t *PasswordResetToken) IsValid() bool {
	return t.UsedAt == nil && !t.IsExpired()
}

func (t *PasswordResetToken) MarkAsUsed() {
	now := time.Now()
	t.UsedAt = &now
	t.UpdatedAt = now
}

func (t *PasswordResetToken) IsUsed() bool {
	return t.UsedAt != nil
}

func (t *EmailVerificationToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

func (t *EmailVerificationToken) IsValid() bool {
	return !t.Used && !t.IsExpired()
}

// Table name methods
func (PasswordResetToken) TableName() string       { return "password_reset_tokens" }
func (EmailVerificationToken) TableName() string   { return "email_verification_tokens" }