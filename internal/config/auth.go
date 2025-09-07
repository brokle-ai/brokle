package config

import (
	"errors"
	"time"
)

// AuthConfig contains authentication and JWT token configuration.
type AuthConfig struct {
	// Token Lifetimes (Go duration strings)
	AccessTokenTTL       time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL      time.Duration `mapstructure:"refresh_token_ttl"`
	TokenRotationEnabled bool          `mapstructure:"token_rotation_enabled"`

	// Rate Limiting
	RateLimitEnabled bool          `mapstructure:"rate_limit_enabled"`
	RateLimitPerIP   int           `mapstructure:"rate_limit_per_ip"`
	RateLimitPerUser int           `mapstructure:"rate_limit_per_user"`
	RateLimitWindow  time.Duration `mapstructure:"rate_limit_window"`

	// JWT Security (Smart defaults: HS256 for dev, RS256 for prod)
	JWTSigningMethod string `mapstructure:"jwt_signing_method"`
	JWTIssuer        string `mapstructure:"jwt_issuer"`

	// HS256 (Simple/Development)
	JWTSecret string `mapstructure:"jwt_secret"`

	// RS256 (Production/Enterprise)
	JWTPrivateKeyPath   string `mapstructure:"jwt_private_key_path"`
	JWTPublicKeyPath    string `mapstructure:"jwt_public_key_path"`
	JWTPrivateKeyBase64 string `mapstructure:"jwt_private_key_base64"`
	JWTPublicKeyBase64  string `mapstructure:"jwt_public_key_base64"`
}

// Validate ensures the auth configuration is valid and complete.
func (c *AuthConfig) Validate() error {
	// Validate token TTLs
	if c.AccessTokenTTL <= 0 {
		return errors.New("access_token_ttl must be greater than 0")
	}
	if c.RefreshTokenTTL <= 0 {
		return errors.New("refresh_token_ttl must be greater than 0")
	}
	if c.AccessTokenTTL >= c.RefreshTokenTTL {
		return errors.New("refresh_token_ttl must be longer than access_token_ttl")
	}

	// Validate JWT signing method and keys
	switch c.JWTSigningMethod {
	case "HS256":
		if c.JWTSecret == "" {
			return errors.New("JWT_SECRET required for HS256 signing method")
		}
		if len(c.JWTSecret) < 32 {
			return errors.New("JWT_SECRET must be at least 32 characters for security")
		}
	case "RS256":
		hasPath := c.JWTPrivateKeyPath != "" && c.JWTPublicKeyPath != ""
		hasBase64 := c.JWTPrivateKeyBase64 != "" && c.JWTPublicKeyBase64 != ""
		if !hasPath && !hasBase64 {
			return errors.New("RS256 requires either key paths or base64 encoded keys")
		}
	default:
		return errors.New("unsupported JWT signing method, use HS256 or RS256")
	}

	// Validate issuer
	if c.JWTIssuer == "" {
		return errors.New("jwt_issuer is required")
	}

	// Validate rate limiting settings if enabled
	if c.RateLimitEnabled {
		if c.RateLimitPerIP <= 0 {
			return errors.New("rate_limit_per_ip must be greater than 0 when rate limiting is enabled")
		}
		if c.RateLimitPerUser <= 0 {
			return errors.New("rate_limit_per_user must be greater than 0 when rate limiting is enabled")
		}
		if c.RateLimitWindow <= 0 {
			return errors.New("rate_limit_window must be greater than 0 when rate limiting is enabled")
		}
	}

	return nil
}

// IsHS256 returns true if using HMAC SHA-256 signing method.
func (c *AuthConfig) IsHS256() bool {
	return c.JWTSigningMethod == "HS256"
}

// IsRS256 returns true if using RSA SHA-256 signing method.
func (c *AuthConfig) IsRS256() bool {
	return c.JWTSigningMethod == "RS256"
}

// HasKeyPaths returns true if RSA key file paths are configured.
func (c *AuthConfig) HasKeyPaths() bool {
	return c.JWTPrivateKeyPath != "" && c.JWTPublicKeyPath != ""
}

// HasKeyBase64 returns true if base64 encoded RSA keys are configured.
func (c *AuthConfig) HasKeyBase64() bool {
	return c.JWTPrivateKeyBase64 != "" && c.JWTPublicKeyBase64 != ""
}

// DefaultAuthConfig returns the default authentication configuration.
func DefaultAuthConfig() AuthConfig {
	return AuthConfig{
		// Token defaults (15 minutes access, 7 days refresh)
		AccessTokenTTL:       15 * time.Minute,
		RefreshTokenTTL:      7 * 24 * time.Hour,
		TokenRotationEnabled: true,

		// Rate limiting defaults (disabled by default)
		RateLimitEnabled: false,
		RateLimitPerIP:   100,
		RateLimitPerUser: 1000,
		RateLimitWindow:  1 * time.Hour,

		// JWT defaults (HS256 for development ease)
		JWTSigningMethod: "HS256",
		JWTIssuer:        "brokle",
		JWTSecret:        "", // Must be set in environment
	}
}