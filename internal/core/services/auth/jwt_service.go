package auth

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"brokle/internal/config"
	"brokle/internal/core/domain/auth"
	"brokle/pkg/ulid"
)

// jwtService implements the auth.JWTService interface with flexible signing methods
type jwtService struct {
	config     *config.AuthConfig
	privateKey interface{} // RSA private key for RS256 or []byte for HS256
	publicKey  interface{} // RSA public key for RS256 or []byte for HS256
}

// NewJWTService creates a new JWT service instance with flexible configuration
func NewJWTService(authConfig *config.AuthConfig) (auth.JWTService, error) {
	if authConfig == nil {
		return nil, fmt.Errorf("auth config is required")
	}

	// Validate configuration
	if err := authConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid auth config: %w", err)
	}

	service := &jwtService{
		config: authConfig,
	}

	// Load keys based on signing method
	if err := service.loadKeys(); err != nil {
		return nil, fmt.Errorf("failed to load JWT keys: %w", err)
	}

	return service, nil
}

// loadKeys loads the appropriate keys based on the signing method
func (s *jwtService) loadKeys() error {
	switch s.config.JWTSigningMethod {
	case "HS256":
		// For HMAC, use the secret as both signing and verification key
		s.privateKey = []byte(s.config.JWTSecret)
		s.publicKey = []byte(s.config.JWTSecret)
		return nil

	case "RS256":
		return s.loadRSAKeys()

	default:
		return fmt.Errorf("unsupported signing method: %s", s.config.JWTSigningMethod)
	}
}

// loadRSAKeys loads RSA keys for RS256 signing
func (s *jwtService) loadRSAKeys() error {
	var privateKeyData, publicKeyData []byte
	var err error

	// Load private key (file path takes precedence over base64)
	if s.config.HasKeyPaths() {
		privateKeyData, err = ioutil.ReadFile(s.config.JWTPrivateKeyPath)
		if err != nil {
			return fmt.Errorf("failed to read private key file: %w", err)
		}
		publicKeyData, err = ioutil.ReadFile(s.config.JWTPublicKeyPath)
		if err != nil {
			return fmt.Errorf("failed to read public key file: %w", err)
		}
	} else if s.config.HasKeyBase64() {
		privateKeyData, err = base64.StdEncoding.DecodeString(s.config.JWTPrivateKeyBase64)
		if err != nil {
			return fmt.Errorf("failed to decode base64 private key: %w", err)
		}
		publicKeyData, err = base64.StdEncoding.DecodeString(s.config.JWTPublicKeyBase64)
		if err != nil {
			return fmt.Errorf("failed to decode base64 public key: %w", err)
		}
	} else {
		return fmt.Errorf("RS256 requires either key paths or base64 encoded keys")
	}

	// Parse private key
	privateBlock, _ := pem.Decode(privateKeyData)
	if privateBlock == nil {
		return fmt.Errorf("failed to decode PEM private key")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(privateBlock.Bytes)
	if err != nil {
		// Try PKCS1 format
		privateKey, err = x509.ParsePKCS1PrivateKey(privateBlock.Bytes)
		if err != nil {
			return fmt.Errorf("failed to parse private key: %w", err)
		}
	}

	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return fmt.Errorf("private key is not an RSA key")
	}

	// Parse public key
	publicBlock, _ := pem.Decode(publicKeyData)
	if publicBlock == nil {
		return fmt.Errorf("failed to decode PEM public key")
	}

	publicKey, err := x509.ParsePKIXPublicKey(publicBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("public key is not an RSA key")
	}

	s.privateKey = rsaPrivateKey
	s.publicKey = rsaPublicKey

	return nil
}

// GenerateAccessToken generates an access token with custom claims
func (s *jwtService) GenerateAccessToken(ctx context.Context, userID ulid.ULID, customClaims map[string]interface{}) (string, error) {
	token, _, err := s.GenerateAccessTokenWithJTI(ctx, userID, customClaims)
	return token, err
}

// GenerateAccessTokenWithJTI generates an access token and returns both token and JTI for session tracking
func (s *jwtService) GenerateAccessTokenWithJTI(ctx context.Context, userID ulid.ULID, customClaims map[string]interface{}) (string, string, error) {
	now := time.Now()
	
	// Generate JTI for this token
	jti := ulid.New().String()
	
	// Create JWT claims
	claims := jwt.MapClaims{
		"iss":        s.config.JWTIssuer,
		"sub":        userID.String(),
		"iat":        now.Unix(),
		"nbf":        now.Unix(),
		"exp":        now.Add(s.config.AccessTokenTTL).Unix(),
		"jti":        jti,
		"token_type": string(auth.TokenTypeAccess),
		"user_id":    userID.String(),
	}

	// Add custom claims
	for key, value := range customClaims {
		claims[key] = value
	}

	// Create token with appropriate signing method
	var signingMethod jwt.SigningMethod
	switch s.config.JWTSigningMethod {
	case "HS256":
		signingMethod = jwt.SigningMethodHS256
	case "RS256":
		signingMethod = jwt.SigningMethodRS256
	default:
		return "", "", fmt.Errorf("unsupported signing method: %s", s.config.JWTSigningMethod)
	}

	token := jwt.NewWithClaims(signingMethod, claims)
	tokenString, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign access token: %w", err)
	}

	return tokenString, jti, nil
}

// GenerateRefreshToken generates a refresh token
func (s *jwtService) GenerateRefreshToken(ctx context.Context, userID ulid.ULID) (string, error) {
	now := time.Now()
	
	// Create JWT claims for refresh token
	claims := jwt.MapClaims{
		"iss":        s.config.JWTIssuer,
		"sub":        userID.String(),
		"iat":        now.Unix(),
		"nbf":        now.Unix(),
		"exp":        now.Add(s.config.RefreshTokenTTL).Unix(),
		"jti":        ulid.New().String(),
		"token_type": string(auth.TokenTypeRefresh),
		"user_id":    userID.String(),
	}

	// Create token with appropriate signing method
	var signingMethod jwt.SigningMethod
	switch s.config.JWTSigningMethod {
	case "HS256":
		signingMethod = jwt.SigningMethodHS256
	case "RS256":
		signingMethod = jwt.SigningMethodRS256
	default:
		return "", fmt.Errorf("unsupported signing method: %s", s.config.JWTSigningMethod)
	}

	token := jwt.NewWithClaims(signingMethod, claims)
	tokenString, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return tokenString, nil
}

// GenerateAPIKeyToken generates a token for API key authentication
func (s *jwtService) GenerateAPIKeyToken(ctx context.Context, keyID ulid.ULID, scopes []string) (string, error) {
	now := time.Now()
	
	// API key tokens use access token TTL for short-lived access
	ttl := s.config.AccessTokenTTL
	
	// Create JWT claims for API key token
	claims := jwt.MapClaims{
		"iss":        s.config.JWTIssuer,
		"sub":        keyID.String(),
		"iat":        now.Unix(),
		"nbf":        now.Unix(),
		"exp":        now.Add(ttl).Unix(),
		"jti":        ulid.New().String(),
		"token_type": string(auth.TokenTypeAPIKey),
		"api_key_id": keyID.String(),
		"scopes":     scopes,
	}

	// Create token with appropriate signing method
	var signingMethod jwt.SigningMethod
	switch s.config.JWTSigningMethod {
	case "HS256":
		signingMethod = jwt.SigningMethodHS256
	case "RS256":
		signingMethod = jwt.SigningMethodRS256
	default:
		return "", fmt.Errorf("unsupported signing method: %s", s.config.JWTSigningMethod)
	}

	token := jwt.NewWithClaims(signingMethod, claims)
	tokenString, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign API key token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken validates any JWT token and returns claims
func (s *jwtService) ValidateToken(ctx context.Context, tokenString string) (*auth.JWTClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method matches configuration
		switch s.config.JWTSigningMethod {
		case "HS256":
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v (expected HMAC)", token.Header["alg"])
			}
			return s.publicKey, nil
		case "RS256":
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v (expected RSA)", token.Header["alg"])
			}
			return s.publicKey, nil
		default:
			return nil, fmt.Errorf("unsupported signing method in config: %s", s.config.JWTSigningMethod)
		}
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
	if tokenClaims.Issuer != s.config.JWTIssuer {
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