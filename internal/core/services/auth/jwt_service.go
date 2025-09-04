package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"brokle/internal/core/domain/auth"
	"brokle/pkg/ulid"
)

// jwtService implements the auth.JWTService interface
type jwtService struct {
	config *auth.TokenConfig
}

// NewJWTService creates a new JWT service instance
func NewJWTService(config *auth.TokenConfig) auth.JWTService {
	if config == nil {
		config = auth.DefaultTokenConfig()
	}
	return &jwtService{
		config: config,
	}
}

// GenerateAccessToken generates an access token with custom claims
func (s *jwtService) GenerateAccessToken(ctx context.Context, userID ulid.ULID, customClaims map[string]interface{}) (string, error) {
	now := time.Now()
	
	// Create JWT claims
	claims := jwt.MapClaims{
		"iss":        s.config.Issuer,
		"sub":        userID.String(),
		"iat":        now.Unix(),
		"nbf":        now.Unix(),
		"exp":        now.Add(s.config.AccessTokenTTL).Unix(),
		"jti":        ulid.New().String(),
		"token_type": string(auth.TokenTypeAccess),
		"user_id":    userID.String(),
	}

	// Add custom claims
	for key, value := range customClaims {
		claims[key] = value
	}

	// Create and sign token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config.SigningKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}

	return tokenString, nil
}

// GenerateRefreshToken generates a refresh token
func (s *jwtService) GenerateRefreshToken(ctx context.Context, userID ulid.ULID) (string, error) {
	now := time.Now()
	
	// Create JWT claims for refresh token
	claims := jwt.MapClaims{
		"iss":        s.config.Issuer,
		"sub":        userID.String(),
		"iat":        now.Unix(),
		"nbf":        now.Unix(),
		"exp":        now.Add(s.config.RefreshTokenTTL).Unix(),
		"jti":        ulid.New().String(),
		"token_type": string(auth.TokenTypeRefresh),
		"user_id":    userID.String(),
	}

	// Create and sign token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config.SigningKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return tokenString, nil
}

// GenerateAPIKeyToken generates a token for API key authentication
func (s *jwtService) GenerateAPIKeyToken(ctx context.Context, keyID ulid.ULID, scopes []string) (string, error) {
	now := time.Now()
	
	// Create JWT claims for API key token
	claims := jwt.MapClaims{
		"iss":        s.config.Issuer,
		"sub":        keyID.String(),
		"iat":        now.Unix(),
		"nbf":        now.Unix(),
		"exp":        now.Add(s.config.APIKeyTokenTTL).Unix(),
		"jti":        ulid.New().String(),
		"token_type": string(auth.TokenTypeAPIKey),
		"api_key_id": keyID.String(),
		"scopes":     scopes,
	}

	// Create and sign token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config.SigningKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign API key token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken validates any JWT token and returns claims
func (s *jwtService) ValidateToken(ctx context.Context, tokenString string) (*auth.JWTClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.SigningKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is invalid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Convert to our custom claims structure
	tokenClaims, err := s.mapClaimsToJWTClaims(claims)
	if err != nil {
		return nil, fmt.Errorf("failed to convert claims: %w", err)
	}

	// Verify issuer
	if tokenClaims.Issuer != s.config.Issuer {
		return nil, fmt.Errorf("invalid token issuer")
	}

	// Check if token is expired
	if tokenClaims.IsExpired() {
		return nil, fmt.Errorf("token is expired")
	}

	// Check not before
	if !tokenClaims.IsValidNow() {
		return nil, fmt.Errorf("token is not valid yet")
	}

	return tokenClaims, nil
}

// ValidateAccessToken validates specifically an access token
func (s *jwtService) ValidateAccessToken(ctx context.Context, tokenString string) (*auth.JWTClaims, error) {
	claims, err := s.ValidateToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != auth.TokenTypeAccess {
		return nil, fmt.Errorf("token is not an access token")
	}

	return claims, nil
}

// ValidateRefreshToken validates specifically a refresh token
func (s *jwtService) ValidateRefreshToken(ctx context.Context, tokenString string) (*auth.JWTClaims, error) {
	claims, err := s.ValidateToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != auth.TokenTypeRefresh {
		return nil, fmt.Errorf("token is not a refresh token")
	}

	return claims, nil
}

// ValidateAPIKeyToken validates specifically an API key token
func (s *jwtService) ValidateAPIKeyToken(ctx context.Context, tokenString string) (*auth.JWTClaims, error) {
	claims, err := s.ValidateToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != auth.TokenTypeAPIKey {
		return nil, fmt.Errorf("token is not an API key token")
	}

	return claims, nil
}

// ExtractClaims extracts claims without validation (for debugging)
func (s *jwtService) ExtractClaims(ctx context.Context, tokenString string) (*auth.JWTClaims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return s.mapClaimsToJWTClaims(claims)
}

// IsTokenExpired checks if token is expired without full validation
func (s *jwtService) IsTokenExpired(ctx context.Context, tokenString string) (bool, error) {
	claims, err := s.ExtractClaims(ctx, tokenString)
	if err != nil {
		return true, err
	}

	return claims.IsExpired(), nil
}

// GetTokenExpiry extracts the expiry time from a token
func (s *jwtService) GetTokenExpiry(ctx context.Context, token string) (time.Time, error) {
	claims, err := s.ExtractClaims(ctx, token)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to extract claims: %w", err)
	}
	
	return time.Unix(claims.ExpiresAt, 0), nil
}

// ParseTokenClaims extracts token claims without validation (used for inspection)
func (s *jwtService) ParseTokenClaims(ctx context.Context, tokenString string) (*auth.JWTClaims, error) {
	return s.ExtractClaims(ctx, tokenString)
}

// GetTokenTTL returns remaining time until token expires
func (s *jwtService) GetTokenTTL(ctx context.Context, tokenString string) (time.Duration, error) {
	claims, err := s.ExtractClaims(ctx, tokenString)
	if err != nil {
		return 0, err
	}

	if claims.IsExpired() {
		return 0, nil
	}

	return claims.TimeUntilExpiry(), nil
}

// mapClaimsToJWTClaims converts jwt.MapClaims to our JWTClaims structure
func (s *jwtService) mapClaimsToJWTClaims(claims jwt.MapClaims) (*auth.JWTClaims, error) {
	jwtClaims := &auth.JWTClaims{}

	// Helper function to safely extract string claims
	getString := func(key string) string {
		if val, ok := claims[key].(string); ok {
			return val
		}
		return ""
	}

	// Helper function to safely extract int64 claims
	getInt64 := func(key string) int64 {
		if val, ok := claims[key].(float64); ok {
			return int64(val)
		}
		return 0
	}

	// Helper function to safely extract ULID claims
	getULID := func(key string) *ulid.ULID {
		if str := getString(key); str != "" {
			if id, err := ulid.Parse(str); err == nil {
				return &id
			}
		}
		return nil
	}

	// Helper function to safely extract string array claims
	getStringArray := func(key string) []string {
		if val, ok := claims[key].([]interface{}); ok {
			result := make([]string, 0, len(val))
			for _, item := range val {
				if str, ok := item.(string); ok {
					result = append(result, str)
				}
			}
			return result
		}
		return nil
	}

	// Standard JWT claims
	jwtClaims.Issuer = getString("iss")
	jwtClaims.Subject = getString("sub")
	jwtClaims.Audience = getString("aud")
	jwtClaims.ExpiresAt = getInt64("exp")
	jwtClaims.NotBefore = getInt64("nbf")
	jwtClaims.IssuedAt = getInt64("iat")
	jwtClaims.JWTID = getString("jti")

	// Custom claims
	jwtClaims.TokenType = auth.TokenType(getString("token_type"))
	jwtClaims.Email = getString("email")

	// Parse UserID
	if userIDStr := getString("user_id"); userIDStr != "" {
		if userID, err := ulid.Parse(userIDStr); err == nil {
			jwtClaims.UserID = userID
		}
	}

	// Context claims
	jwtClaims.OrganizationID = getULID("organization_id")
	jwtClaims.ProjectID = getULID("project_id")
	jwtClaims.EnvironmentID = getULID("environment_id")

	// Permission claims
	jwtClaims.Scopes = getStringArray("scopes")
	jwtClaims.Permissions = getStringArray("permissions")
	if role := getString("role"); role != "" {
		jwtClaims.Role = &role
	}

	// API Key and session claims
	jwtClaims.APIKeyID = getULID("api_key_id")
	jwtClaims.SessionID = getULID("session_id")

	// Security claims
	if ipAddress := getString("ip_address"); ipAddress != "" {
		jwtClaims.IPAddress = &ipAddress
	}
	if userAgent := getString("user_agent"); userAgent != "" {
		jwtClaims.UserAgent = &userAgent
	}

	return jwtClaims, nil
}