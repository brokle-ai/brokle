package auth

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"maps"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/google/uuid"

	"brokle/internal/config"
	authDomain "brokle/internal/core/domain/auth"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/uid"
)

// JWTService implements the auth.JWTService interface with flexible signing methods
type JWTService struct {
	cfg     *config.AuthConfig
	privateKey any // RSA private key for RS256 or []byte for HS256
	publicKey  any // RSA public key for RS256 or []byte for HS256
}

// NewJWTService creates a new JWT service instance with flexible configuration
func NewJWTService(authConfig *config.AuthConfig) (*JWTService, error) {
	if authConfig == nil {
		return nil, appErrors.NewValidationError("config", "Auth config is required")
	}

	// Validate configuration
	if err := authConfig.Validate(); err != nil {
		return nil, appErrors.NewValidationError("config", "Invalid auth config: "+err.Error())
	}

	service := &JWTService{
		cfg: authConfig,
	}

	// Load keys based on signing method
	if err := service.loadKeys(); err != nil {
		return nil, appErrors.NewInternalError("failed to load JWT keys", err)
	}

	return service, nil
}

// loadKeys loads the appropriate keys based on the signing method
func (s *JWTService) loadKeys() error {
	switch s.cfg.JWTSigningMethod {
	case "HS256":
		// SECURITY: Strict validation when creating JWT service
		if s.cfg.JWTSecret == "" {
			return appErrors.NewValidationError("JWT_SECRET", "JWT_SECRET is required for HS256 signing method - cannot create JWT service")
		}
		if len(s.cfg.JWTSecret) < 32 {
			return appErrors.NewValidationError("JWT_SECRET", fmt.Sprintf("JWT_SECRET must be at least 32 characters for security, got %d characters", len(s.cfg.JWTSecret)))
		}

		// For HMAC, use the secret as both signing and verification key
		s.privateKey = []byte(s.cfg.JWTSecret)
		s.publicKey = []byte(s.cfg.JWTSecret)
		return nil

	case "RS256":
		return s.loadRSAKeys()

	default:
		return appErrors.NewValidationError("signing_method", "Unsupported signing method: "+s.cfg.JWTSigningMethod)
	}
}

// loadRSAKeys loads RSA keys for RS256 signing
func (s *JWTService) loadRSAKeys() error {
	var privateKeyData, publicKeyData []byte
	var err error

	// Load private key (file path takes precedence over base64)
	if s.cfg.HasKeyPaths() {
		privateKeyData, err = os.ReadFile(s.cfg.JWTPrivateKeyPath)
		if err != nil {
			return appErrors.NewInternalError("failed to read private key file", err)
		}
		publicKeyData, err = os.ReadFile(s.cfg.JWTPublicKeyPath)
		if err != nil {
			return appErrors.NewInternalError("failed to read public key file", err)
		}
	} else if s.cfg.HasKeyBase64() {
		privateKeyData, err = base64.StdEncoding.DecodeString(s.cfg.JWTPrivateKeyBase64)
		if err != nil {
			return appErrors.NewInternalError("failed to decode base64 private key", err)
		}
		publicKeyData, err = base64.StdEncoding.DecodeString(s.cfg.JWTPublicKeyBase64)
		if err != nil {
			return appErrors.NewInternalError("failed to decode base64 public key", err)
		}
	} else {
		return appErrors.NewValidationError("keys", "RS256 requires either key paths or base64 encoded keys")
	}

	// Parse private key
	privateBlock, _ := pem.Decode(privateKeyData)
	if privateBlock == nil {
		return appErrors.NewInternalError("failed to decode PEM private key", nil)
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(privateBlock.Bytes)
	if err != nil {
		// Try PKCS1 format
		privateKey, err = x509.ParsePKCS1PrivateKey(privateBlock.Bytes)
		if err != nil {
			return appErrors.NewInternalError("failed to parse private key", err)
		}
	}

	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return appErrors.NewInternalError("private key is not an RSA key", nil)
	}

	// Parse public key
	publicBlock, _ := pem.Decode(publicKeyData)
	if publicBlock == nil {
		return appErrors.NewInternalError("failed to decode PEM public key", nil)
	}

	publicKey, err := x509.ParsePKIXPublicKey(publicBlock.Bytes)
	if err != nil {
		return appErrors.NewInternalError("failed to parse public key", err)
	}

	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return appErrors.NewInternalError("public key is not an RSA key", nil)
	}

	s.privateKey = rsaPrivateKey
	s.publicKey = rsaPublicKey

	return nil
}

// GenerateAccessToken generates an access token with custom claims
func (s *JWTService) GenerateAccessToken(ctx context.Context, userID uuid.UUID, customClaims map[string]any) (string, error) {
	token, _, err := s.GenerateAccessTokenWithJTI(ctx, userID, customClaims)
	return token, err
}

// GenerateAccessTokenWithJTI generates an access token and returns both token and JTI for session tracking
func (s *JWTService) GenerateAccessTokenWithJTI(ctx context.Context, userID uuid.UUID, customClaims map[string]any) (string, string, error) {
	now := time.Now()

	// Generate JTI for this token
	jti := uid.New().String()

	// Create JWT claims
	claims := jwt.MapClaims{
		"iss":        s.cfg.JWTIssuer,
		"sub":        userID.String(),
		"iat":        now.Unix(),
		"nbf":        now.Unix(),
		"exp":        now.Add(s.cfg.AccessTokenTTL).Unix(),
		"jti":        jti,
		"token_type": string(authDomain.TokenTypeAccess),
		"user_id":    userID.String(),
	}

	// Add custom claims
	maps.Copy(claims, customClaims)

	// Create token with appropriate signing method
	var signingMethod jwt.SigningMethod
	switch s.cfg.JWTSigningMethod {
	case "HS256":
		signingMethod = jwt.SigningMethodHS256
	case "RS256":
		signingMethod = jwt.SigningMethodRS256
	default:
		return "", "", appErrors.NewValidationError("signing_method", "Unsupported signing method: "+s.cfg.JWTSigningMethod)
	}

	token := jwt.NewWithClaims(signingMethod, claims)
	tokenString, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", "", appErrors.NewInternalError("failed to sign access token", err)
	}

	return tokenString, jti, nil
}

// GenerateRefreshToken generates a refresh token
func (s *JWTService) GenerateRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	now := time.Now()

	// Create JWT claims for refresh token
	claims := jwt.MapClaims{
		"iss":        s.cfg.JWTIssuer,
		"sub":        userID.String(),
		"iat":        now.Unix(),
		"nbf":        now.Unix(),
		"exp":        now.Add(s.cfg.RefreshTokenTTL).Unix(),
		"jti":        uid.New().String(),
		"token_type": string(authDomain.TokenTypeRefresh),
		"user_id":    userID.String(),
	}

	// Create token with appropriate signing method
	var signingMethod jwt.SigningMethod
	switch s.cfg.JWTSigningMethod {
	case "HS256":
		signingMethod = jwt.SigningMethodHS256
	case "RS256":
		signingMethod = jwt.SigningMethodRS256
	default:
		return "", appErrors.NewValidationError("signing_method", "Unsupported signing method: "+s.cfg.JWTSigningMethod)
	}

	token := jwt.NewWithClaims(signingMethod, claims)
	tokenString, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", appErrors.NewInternalError("failed to sign refresh token", err)
	}

	return tokenString, nil
}

// GenerateAPIKeyToken generates a token for API key authentication
func (s *JWTService) GenerateAPIKeyToken(ctx context.Context, keyID uuid.UUID, scopes []string) (string, error) {
	now := time.Now()

	// API key tokens use access token TTL for short-lived access
	ttl := s.cfg.AccessTokenTTL

	// Create JWT claims for API key token
	claims := jwt.MapClaims{
		"iss":        s.cfg.JWTIssuer,
		"sub":        keyID.String(),
		"iat":        now.Unix(),
		"nbf":        now.Unix(),
		"exp":        now.Add(ttl).Unix(),
		"jti":        uid.New().String(),
		"token_type": string(authDomain.TokenTypeAPIKey),
		"api_key_id": keyID.String(),
		"scopes":     scopes,
	}

	// Create token with appropriate signing method
	var signingMethod jwt.SigningMethod
	switch s.cfg.JWTSigningMethod {
	case "HS256":
		signingMethod = jwt.SigningMethodHS256
	case "RS256":
		signingMethod = jwt.SigningMethodRS256
	default:
		return "", appErrors.NewValidationError("signing_method", "Unsupported signing method: "+s.cfg.JWTSigningMethod)
	}

	token := jwt.NewWithClaims(signingMethod, claims)
	tokenString, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", appErrors.NewInternalError("failed to sign API key token", err)
	}

	return tokenString, nil
}

// ValidateToken validates any JWT token and returns claims.
//
// Error contract: returns domain sentinels wrapped via fmt.Errorf("...: %w", X)
// so callers can match with errors.Is. The auth service translates these
// sentinels into *AppError at its public-method boundary (Gitea pattern).
//
//   - authDomain.ErrTokenExpired — exp/nbf failed, token decoded successfully
//   - authDomain.ErrTokenInvalid — signature, parsing, claims, or issuer failure
func (s *JWTService) ValidateToken(ctx context.Context, tokenString string) (*authDomain.JWTClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		// Verify signing method matches configuration
		switch s.cfg.JWTSigningMethod {
		case "HS256":
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: expected HMAC, got %v", token.Header["alg"])
			}
			return s.publicKey, nil
		case "RS256":
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: expected RSA, got %v", token.Header["alg"])
			}
			return s.publicKey, nil
		default:
			return nil, fmt.Errorf("unsupported signing method in config: %s", s.cfg.JWTSigningMethod)
		}
	})

	if err != nil {
		// jwt/v5 surfaces ErrTokenExpired / ErrTokenNotValidYet via the parser
		// itself (it validates exp/nbf during Parse). Map those to the domain
		// sentinel; everything else (signature, malformed, audience, issuer
		// mismatch) is a generic invalid-token failure.
		if errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, fmt.Errorf("validate token: %w", authDomain.ErrTokenExpired)
		}
		return nil, fmt.Errorf("validate token: %w", authDomain.ErrTokenInvalid)
	}

	if !token.Valid {
		return nil, fmt.Errorf("validate token: %w", authDomain.ErrTokenInvalid)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("validate token: %w", authDomain.ErrTokenInvalid)
	}

	// Convert to our custom claims structure
	tokenClaims, err := s.mapClaimsToJWTClaims(claims)
	if err != nil {
		return nil, fmt.Errorf("validate token: convert claims: %w", authDomain.ErrTokenInvalid)
	}

	// Verify issuer
	if tokenClaims.Issuer != s.cfg.JWTIssuer {
		return nil, fmt.Errorf("validate token: invalid issuer: %w", authDomain.ErrTokenInvalid)
	}

	// Defence-in-depth: jwt/v5 already enforces exp/nbf, but double-check
	// against our own clock in case a legacy parser path skips them.
	if tokenClaims.IsExpired() {
		return nil, fmt.Errorf("validate token: %w", authDomain.ErrTokenExpired)
	}

	if !tokenClaims.IsValidNow() {
		return nil, fmt.Errorf("validate token: not valid yet: %w", authDomain.ErrTokenExpired)
	}

	return tokenClaims, nil
}

// ValidateAccessToken validates specifically an access token.
// Wraps authDomain.ErrTokenInvalid when the token type does not match.
func (s *JWTService) ValidateAccessToken(ctx context.Context, tokenString string) (*authDomain.JWTClaims, error) {
	claims, err := s.ValidateToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != authDomain.TokenTypeAccess {
		return nil, fmt.Errorf("validate access token: wrong type %q: %w", claims.TokenType, authDomain.ErrTokenInvalid)
	}

	return claims, nil
}

// ValidateRefreshToken validates specifically a refresh token.
// Wraps authDomain.ErrTokenInvalid when the token type does not match.
func (s *JWTService) ValidateRefreshToken(ctx context.Context, tokenString string) (*authDomain.JWTClaims, error) {
	claims, err := s.ValidateToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != authDomain.TokenTypeRefresh {
		return nil, fmt.Errorf("validate refresh token: wrong type %q: %w", claims.TokenType, authDomain.ErrTokenInvalid)
	}

	return claims, nil
}

// ValidateAPIKeyToken validates specifically an API key token.
// Wraps authDomain.ErrTokenInvalid when the token type does not match.
func (s *JWTService) ValidateAPIKeyToken(ctx context.Context, tokenString string) (*authDomain.JWTClaims, error) {
	claims, err := s.ValidateToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != authDomain.TokenTypeAPIKey {
		return nil, fmt.Errorf("validate api key token: wrong type %q: %w", claims.TokenType, authDomain.ErrTokenInvalid)
	}

	return claims, nil
}

// ExtractClaims extracts claims without validation (for debugging).
// Wraps authDomain.ErrTokenInvalid on any parse/claim-shape failure.
func (s *JWTService) ExtractClaims(ctx context.Context, tokenString string) (*authDomain.JWTClaims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("extract claims: %w", authDomain.ErrTokenInvalid)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("extract claims: %w", authDomain.ErrTokenInvalid)
	}

	return s.mapClaimsToJWTClaims(claims)
}

// IsTokenExpired checks if token is expired without full validation
func (s *JWTService) IsTokenExpired(ctx context.Context, tokenString string) (bool, error) {
	claims, err := s.ExtractClaims(ctx, tokenString)
	if err != nil {
		return true, err
	}

	return claims.IsExpired(), nil
}

// GetTokenExpiry extracts the expiry time from a token.
// Forwards the wrapped authDomain.ErrTokenInvalid sentinel from ExtractClaims.
func (s *JWTService) GetTokenExpiry(ctx context.Context, token string) (time.Time, error) {
	claims, err := s.ExtractClaims(ctx, token)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(claims.ExpiresAt, 0), nil
}

// GetTokenTTL returns remaining time until token expires
func (s *JWTService) GetTokenTTL(ctx context.Context, tokenString string) (time.Duration, error) {
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
func (s *JWTService) mapClaimsToJWTClaims(claims jwt.MapClaims) (*authDomain.JWTClaims, error) {
	jwtClaims := &authDomain.JWTClaims{}

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

	// Helper function to safely extract UUID claims
	getUUID := func(key string) *uuid.UUID {
		if str := getString(key); str != "" {
			if id, err := uuid.Parse(str); err == nil {
				return &id
			}
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
	jwtClaims.TokenType = authDomain.TokenType(getString("token_type"))
	jwtClaims.Email = getString("email")

	// Parse UserID
	if userIDStr := getString("user_id"); userIDStr != "" {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			jwtClaims.UserID = userID
		}
	}

	// Clean JWT structure - no context or permission claims stored in JWT

	// API Key and session claims
	jwtClaims.APIKeyID = getUUID("api_key_id")
	jwtClaims.SessionID = getUUID("session_id")

	// Clean JWT structure - no IP or UserAgent tracking

	return jwtClaims, nil
}
