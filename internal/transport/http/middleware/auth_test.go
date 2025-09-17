package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"brokle/internal/config"
	authDomain "brokle/internal/core/domain/auth"
	"brokle/pkg/ulid"
)

// Mock services
type mockJWTService struct {
	mock.Mock
}

func (m *mockJWTService) ValidateAccessToken(ctx context.Context, token string) (*authDomain.JWTClaims, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authDomain.JWTClaims), args.Error(1)
}

func (m *mockJWTService) ValidateRefreshToken(ctx context.Context, token string) (*authDomain.JWTClaims, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authDomain.JWTClaims), args.Error(1)
}

func (m *mockJWTService) GenerateAccessToken(ctx context.Context, userID ulid.ULID, sessionID ulid.ULID, claims map[string]interface{}) (string, error) {
	args := m.Called(ctx, userID, sessionID, claims)
	return args.String(0), args.Error(1)
}

func (m *mockJWTService) GenerateRefreshToken(ctx context.Context, userID ulid.ULID, sessionID ulid.ULID) (string, error) {
	args := m.Called(ctx, userID, sessionID)
	return args.String(0), args.Error(1)
}

func (m *mockJWTService) GetTokenExpiry(ctx context.Context, token string) (time.Time, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(time.Time), args.Error(1)
}

func (m *mockJWTService) IsTokenExpired(ctx context.Context, token string) (bool, error) {
	args := m.Called(ctx, token)
	return args.Bool(0), args.Error(1)
}

type mockKeyPairService struct {
	mock.Mock
}

func (m *mockKeyPairService) CreateKeyPair(ctx context.Context, userID ulid.ULID, req *authDomain.CreateKeyPairRequest) (*authDomain.CreateKeyPairResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authDomain.CreateKeyPairResponse), args.Error(1)
}

func (m *mockKeyPairService) ValidateKeyPair(ctx context.Context, publicKey, secretKey string) (*authDomain.KeyPair, error) {
	args := m.Called(ctx, publicKey, secretKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authDomain.KeyPair), args.Error(1)
}

func (m *mockKeyPairService) AuthenticateWithKeyPair(ctx context.Context, publicKey, secretKey string) (*authDomain.AuthContext, error) {
	args := m.Called(ctx, publicKey, secretKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authDomain.AuthContext), args.Error(1)
}

func (m *mockKeyPairService) GetKeyPair(ctx context.Context, keyID ulid.ULID) (*authDomain.KeyPair, error) {
	args := m.Called(ctx, keyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authDomain.KeyPair), args.Error(1)
}

func (m *mockKeyPairService) GetKeyPairs(ctx context.Context, filters *authDomain.KeyPairFilters) ([]*authDomain.KeyPair, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*authDomain.KeyPair), args.Error(1)
}

func (m *mockKeyPairService) UpdateKeyPair(ctx context.Context, keyID ulid.ULID, req *authDomain.UpdateKeyPairRequest) error {
	args := m.Called(ctx, keyID, req)
	return args.Error(0)
}

func (m *mockKeyPairService) RevokeKeyPair(ctx context.Context, keyID ulid.ULID) error {
	args := m.Called(ctx, keyID)
	return args.Error(0)
}

func (m *mockKeyPairService) UpdateLastUsed(ctx context.Context, keyID ulid.ULID) error {
	args := m.Called(ctx, keyID)
	return args.Error(0)
}

func (m *mockKeyPairService) CheckRateLimit(ctx context.Context, keyID ulid.ULID) (bool, error) {
	args := m.Called(ctx, keyID)
	return args.Bool(0), args.Error(1)
}

func (m *mockKeyPairService) GetKeyPairContext(ctx context.Context, keyID ulid.ULID) (*authDomain.AuthContext, error) {
	args := m.Called(ctx, keyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authDomain.AuthContext), args.Error(1)
}

func (m *mockKeyPairService) CanKeyPairAccessResource(ctx context.Context, keyID ulid.ULID, resource string) (bool, error) {
	args := m.Called(ctx, keyID, resource)
	return args.Bool(0), args.Error(1)
}

func (m *mockKeyPairService) CheckKeyPairScopes(ctx context.Context, keyID ulid.ULID, requiredScopes []string) (bool, error) {
	args := m.Called(ctx, keyID, requiredScopes)
	return args.Bool(0), args.Error(1)
}

func (m *mockKeyPairService) GetKeyPairsByUser(ctx context.Context, userID ulid.ULID) ([]*authDomain.KeyPair, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*authDomain.KeyPair), args.Error(1)
}

func (m *mockKeyPairService) GetKeyPairsByOrganization(ctx context.Context, orgID ulid.ULID) ([]*authDomain.KeyPair, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*authDomain.KeyPair), args.Error(1)
}

func (m *mockKeyPairService) GetKeyPairsByProject(ctx context.Context, projectID ulid.ULID) ([]*authDomain.KeyPair, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*authDomain.KeyPair), args.Error(1)
}

func (m *mockKeyPairService) GetKeyPairsByEnvironment(ctx context.Context, envID ulid.ULID) ([]*authDomain.KeyPair, error) {
	args := m.Called(ctx, envID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*authDomain.KeyPair), args.Error(1)
}

func (m *mockKeyPairService) GenerateKeyPair(ctx context.Context, projectID ulid.ULID) (string, string, error) {
	args := m.Called(ctx, projectID)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *mockKeyPairService) ValidatePublicKeyFormat(ctx context.Context, publicKey string) error {
	args := m.Called(ctx, publicKey)
	return args.Error(0)
}

func (m *mockKeyPairService) ExtractProjectIDFromPublicKey(ctx context.Context, publicKey string) (ulid.ULID, error) {
	args := m.Called(ctx, publicKey)
	return args.Get(0).(ulid.ULID), args.Error(1)
}

type mockBlacklistedTokenService struct {
	mock.Mock
}

func (m *mockBlacklistedTokenService) IsTokenBlacklisted(ctx context.Context, jti string) (bool, error) {
	args := m.Called(ctx, jti)
	return args.Bool(0), args.Error(1)
}

func (m *mockBlacklistedTokenService) BlacklistToken(ctx context.Context, jti string, userID ulid.ULID, expiresAt time.Time, reason string) error {
	args := m.Called(ctx, jti, userID, expiresAt, reason)
	return args.Error(0)
}

func (m *mockBlacklistedTokenService) GetBlacklistedTokens(ctx context.Context, userID ulid.ULID) ([]*authDomain.BlacklistedToken, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*authDomain.BlacklistedToken), args.Error(1)
}

func (m *mockBlacklistedTokenService) CleanupExpiredTokens(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type mockOrganizationMemberService struct {
	mock.Mock
}

func (m *mockOrganizationMemberService) GetMember(ctx context.Context, orgID, userID ulid.ULID) (*authDomain.OrganizationMember, error) {
	args := m.Called(ctx, orgID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authDomain.OrganizationMember), args.Error(1)
}

func (m *mockOrganizationMemberService) GetMembers(ctx context.Context, orgID ulid.ULID) ([]*authDomain.OrganizationMember, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*authDomain.OrganizationMember), args.Error(1)
}

func (m *mockOrganizationMemberService) AddMember(ctx context.Context, orgID, userID ulid.ULID, roleID string) error {
	args := m.Called(ctx, orgID, userID, roleID)
	return args.Error(0)
}

func (m *mockOrganizationMemberService) UpdateMemberRole(ctx context.Context, orgID, userID ulid.ULID, roleID string) error {
	args := m.Called(ctx, orgID, userID, roleID)
	return args.Error(0)
}

func (m *mockOrganizationMemberService) RemoveMember(ctx context.Context, orgID, userID ulid.ULID) error {
	args := m.Called(ctx, orgID, userID)
	return args.Error(0)
}

func (m *mockOrganizationMemberService) HasPermission(ctx context.Context, orgID, userID ulid.ULID, permission string) (bool, error) {
	args := m.Called(ctx, orgID, userID, permission)
	return args.Bool(0), args.Error(1)
}

func (m *mockOrganizationMemberService) GetUserRoles(ctx context.Context, orgID, userID ulid.ULID) ([]*authDomain.Role, error) {
	args := m.Called(ctx, orgID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*authDomain.Role), args.Error(1)
}

func (m *mockOrganizationMemberService) GetUserPermissions(ctx context.Context, orgID, userID ulid.ULID) ([]*authDomain.Permission, error) {
	args := m.Called(ctx, orgID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*authDomain.Permission), args.Error(1)
}

// Test helper functions
func createTestAuthMiddleware() (*AuthMiddleware, *mockJWTService, *mockKeyPairService, *mockBlacklistedTokenService, *mockOrganizationMemberService) {
	mockJWT := &mockJWTService{}
	mockKeyPair := &mockKeyPairService{}
	mockBlacklist := &mockBlacklistedTokenService{}
	mockOrgMember := &mockOrganizationMemberService{}

	cfg := &config.Config{}
	logger := logrus.New()

	middleware := NewAuthMiddleware(cfg, logger, mockJWT, mockKeyPair, mockBlacklist, mockOrgMember)

	return middleware, mockJWT, mockKeyPair, mockBlacklist, mockOrgMember
}

func createTestGinContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest("GET", "/test", nil)
	return ctx, recorder
}

func createTestKeyPair() *authDomain.KeyPair {
	userID := ulid.New()
	orgID := ulid.New()
	projectID := ulid.New()

	return authDomain.NewKeyPair(
		userID,
		orgID,
		projectID,
		"Test Key Pair",
		"pk_"+projectID.String()+"_abcdef123456",
		"$2a$10$hashedSecretKey",
		[]string{authDomain.ScopeGatewayRead, authDomain.ScopeAnalyticsRead},
		1000,
		nil,
	)
}

func createTestAuthContext() *authDomain.AuthContext {
	userID := ulid.New()
	keyPairID := ulid.New()
	orgID := ulid.New()
	projectID := ulid.New()

	return &authDomain.AuthContext{
		UserID:         userID,
		KeyPairID:      &keyPairID,
		OrganizationID: &orgID,
		ProjectID:      &projectID,
		Scopes:         []string{authDomain.ScopeGatewayRead, authDomain.ScopeAnalyticsRead},
	}
}

// Tests for RequireKeyPair middleware
func TestAuthMiddleware_RequireKeyPair(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(*mockKeyPairService)
		headers      map[string]string
		expectStatus int
		expectAuth   bool
	}{
		{
			name: "successful key pair authentication - Authorization header",
			setup: func(mockKeyPair *mockKeyPairService) {
				authCtx := createTestAuthContext()
				mockKeyPair.On("AuthenticateWithKeyPair", mock.Anything, "pk_test_123", "sk_secret_456").Return(authCtx, nil)
			},
			headers: map[string]string{
				"Authorization": "Bearer pk_test_123:sk_secret_456",
			},
			expectStatus: http.StatusOK,
			expectAuth:   true,
		},
		{
			name: "successful key pair authentication - X-API-Key header",
			setup: func(mockKeyPair *mockKeyPairService) {
				authCtx := createTestAuthContext()
				mockKeyPair.On("AuthenticateWithKeyPair", mock.Anything, "pk_test_123", "sk_secret_456").Return(authCtx, nil)
			},
			headers: map[string]string{
				"X-API-Key": "pk_test_123:sk_secret_456",
			},
			expectStatus: http.StatusOK,
			expectAuth:   true,
		},
		{
			name:  "missing authentication header",
			setup: func(mockKeyPair *mockKeyPairService) {},
			headers: map[string]string{
				// No auth headers
			},
			expectStatus: http.StatusUnauthorized,
			expectAuth:   false,
		},
		{
			name:  "invalid key pair format - missing colon",
			setup: func(mockKeyPair *mockKeyPairService) {},
			headers: map[string]string{
				"Authorization": "Bearer pk_test_123_sk_secret_456",
			},
			expectStatus: http.StatusUnauthorized,
			expectAuth:   false,
		},
		{
			name: "authentication service error",
			setup: func(mockKeyPair *mockKeyPairService) {
				mockKeyPair.On("AuthenticateWithKeyPair", mock.Anything, "pk_test_123", "sk_secret_456").Return(nil, assert.AnError)
			},
			headers: map[string]string{
				"Authorization": "Bearer pk_test_123:sk_secret_456",
			},
			expectStatus: http.StatusUnauthorized,
			expectAuth:   false,
		},
		{
			name:  "empty key pair",
			setup: func(mockKeyPair *mockKeyPairService) {},
			headers: map[string]string{
				"Authorization": "Bearer :",
			},
			expectStatus: http.StatusUnauthorized,
			expectAuth:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware, _, mockKeyPair, _, _ := createTestAuthMiddleware()
			ctx, recorder := createTestGinContext()

			if tt.setup != nil {
				tt.setup(mockKeyPair)
			}

			// Set headers
			for key, value := range tt.headers {
				ctx.Request.Header.Set(key, value)
			}

			// Create a test handler that checks auth context
			testHandler := func(c *gin.Context) {
				if tt.expectAuth {
					authCtx, exists := GetAuthContext(c)
					assert.True(t, exists)
					assert.NotNil(t, authCtx)
					assert.NotNil(t, authCtx.KeyPairID)
				}
				c.Status(http.StatusOK)
			}

			// Apply middleware and run handler
			middlewareFunc := middleware.RequireKeyPair()
			middlewareFunc(ctx)

			if ctx.IsAborted() {
				assert.Equal(t, tt.expectStatus, recorder.Code)
			} else {
				testHandler(ctx)
				assert.Equal(t, tt.expectStatus, recorder.Code)
			}

			mockKeyPair.AssertExpectations(t)
		})
	}
}

// Tests for RequireEitherAuth middleware
func TestAuthMiddleware_RequireEitherAuth(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(*mockJWTService, *mockKeyPairService, *mockBlacklistedTokenService)
		headers      map[string]string
		expectStatus int
		expectAuth   bool
		authType     string // "jwt" or "keypair"
	}{
		{
			name: "successful JWT authentication",
			setup: func(mockJWT *mockJWTService, mockKeyPair *mockKeyPairService, mockBlacklist *mockBlacklistedTokenService) {
				claims := &authDomain.JWTClaims{
					Subject:   ulid.New().String(),
					JTI:       "test-jti",
					TokenType: authDomain.TokenTypeAccess,
				}
				mockJWT.On("ValidateAccessToken", mock.Anything, "valid-jwt-token").Return(claims, nil)
				mockBlacklist.On("IsTokenBlacklisted", mock.Anything, "test-jti").Return(false, nil)
			},
			headers: map[string]string{
				"Authorization": "Bearer valid-jwt-token",
			},
			expectStatus: http.StatusOK,
			expectAuth:   true,
			authType:     "jwt",
		},
		{
			name: "successful key pair authentication",
			setup: func(mockJWT *mockJWTService, mockKeyPair *mockKeyPairService, mockBlacklist *mockBlacklistedTokenService) {
				authCtx := createTestAuthContext()
				// JWT validation should fail
				mockJWT.On("ValidateAccessToken", mock.Anything, "pk_test_123:sk_secret_456").Return(nil, assert.AnError)
				// Key pair validation should succeed
				mockKeyPair.On("AuthenticateWithKeyPair", mock.Anything, "pk_test_123", "sk_secret_456").Return(authCtx, nil)
			},
			headers: map[string]string{
				"Authorization": "Bearer pk_test_123:sk_secret_456",
			},
			expectStatus: http.StatusOK,
			expectAuth:   true,
			authType:     "keypair",
		},
		{
			name: "both authentications fail",
			setup: func(mockJWT *mockJWTService, mockKeyPair *mockKeyPairService, mockBlacklist *mockBlacklistedTokenService) {
				mockJWT.On("ValidateAccessToken", mock.Anything, "invalid-token").Return(nil, assert.AnError)
				mockKeyPair.On("AuthenticateWithKeyPair", mock.Anything, "invalid", "token").Return(nil, assert.AnError)
			},
			headers: map[string]string{
				"Authorization": "Bearer invalid-token",
			},
			expectStatus: http.StatusUnauthorized,
			expectAuth:   false,
		},
		{
			name:  "missing authorization header",
			setup: func(mockJWT *mockJWTService, mockKeyPair *mockKeyPairService, mockBlacklist *mockBlacklistedTokenService) {},
			headers: map[string]string{
				// No auth headers
			},
			expectStatus: http.StatusUnauthorized,
			expectAuth:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware, mockJWT, mockKeyPair, mockBlacklist, _ := createTestAuthMiddleware()
			ctx, recorder := createTestGinContext()

			if tt.setup != nil {
				tt.setup(mockJWT, mockKeyPair, mockBlacklist)
			}

			// Set headers
			for key, value := range tt.headers {
				ctx.Request.Header.Set(key, value)
			}

			// Create a test handler that checks auth context
			testHandler := func(c *gin.Context) {
				if tt.expectAuth {
					authCtx, exists := GetAuthContext(c)
					assert.True(t, exists)
					assert.NotNil(t, authCtx)

					if tt.authType == "keypair" {
						assert.NotNil(t, authCtx.KeyPairID)
					}
				}
				c.Status(http.StatusOK)
			}

			// Apply middleware and run handler
			middlewareFunc := middleware.RequireEitherAuth()
			middlewareFunc(ctx)

			if ctx.IsAborted() {
				assert.Equal(t, tt.expectStatus, recorder.Code)
			} else {
				testHandler(ctx)
				assert.Equal(t, tt.expectStatus, recorder.Code)
			}

			mockJWT.AssertExpectations(t)
			mockKeyPair.AssertExpectations(t)
			mockBlacklist.AssertExpectations(t)
		})
	}
}

// Tests for RequireScope middleware
func TestAuthMiddleware_RequireScope(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(*mockKeyPairService)
		authSetup      func(*gin.Context)
		requiredScopes []string
		expectStatus   int
		expectPass     bool
	}{
		{
			name: "successful scope validation",
			setup: func(mockKeyPair *mockKeyPairService) {
				mockKeyPair.On("CheckKeyPairScopes", mock.Anything, mock.Anything, []string{authDomain.ScopeGatewayRead}).Return(true, nil)
			},
			authSetup: func(ctx *gin.Context) {
				authCtx := createTestAuthContext()
				ctx.Set(AuthContextKey, authCtx)
			},
			requiredScopes: []string{authDomain.ScopeGatewayRead},
			expectStatus:   http.StatusOK,
			expectPass:     true,
		},
		{
			name: "insufficient scopes",
			setup: func(mockKeyPair *mockKeyPairService) {
				mockKeyPair.On("CheckKeyPairScopes", mock.Anything, mock.Anything, []string{authDomain.ScopeGatewayWrite}).Return(false, nil)
			},
			authSetup: func(ctx *gin.Context) {
				authCtx := createTestAuthContext()
				ctx.Set(AuthContextKey, authCtx)
			},
			requiredScopes: []string{authDomain.ScopeGatewayWrite},
			expectStatus:   http.StatusForbidden,
			expectPass:     false,
		},
		{
			name:  "missing auth context",
			setup: func(mockKeyPair *mockKeyPairService) {},
			authSetup: func(ctx *gin.Context) {
				// Don't set auth context
			},
			requiredScopes: []string{authDomain.ScopeGatewayRead},
			expectStatus:   http.StatusUnauthorized,
			expectPass:     false,
		},
		{
			name:  "JWT auth without key pair ID",
			setup: func(mockKeyPair *mockKeyPairService) {},
			authSetup: func(ctx *gin.Context) {
				authCtx := &authDomain.AuthContext{
					UserID: ulid.New(),
					// No KeyPairID - this is JWT auth
				}
				ctx.Set(AuthContextKey, authCtx)
			},
			requiredScopes: []string{authDomain.ScopeGatewayRead},
			expectStatus:   http.StatusForbidden,
			expectPass:     false,
		},
		{
			name: "scope check service error",
			setup: func(mockKeyPair *mockKeyPairService) {
				mockKeyPair.On("CheckKeyPairScopes", mock.Anything, mock.Anything, []string{authDomain.ScopeGatewayRead}).Return(false, assert.AnError)
			},
			authSetup: func(ctx *gin.Context) {
				authCtx := createTestAuthContext()
				ctx.Set(AuthContextKey, authCtx)
			},
			requiredScopes: []string{authDomain.ScopeGatewayRead},
			expectStatus:   http.StatusInternalServerError,
			expectPass:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware, _, mockKeyPair, _, _ := createTestAuthMiddleware()
			ctx, recorder := createTestGinContext()

			if tt.setup != nil {
				tt.setup(mockKeyPair)
			}

			if tt.authSetup != nil {
				tt.authSetup(ctx)
			}

			// Create a test handler
			testHandler := func(c *gin.Context) {
				c.Status(http.StatusOK)
			}

			// Apply middleware and run handler
			middlewareFunc := middleware.RequireScope(tt.requiredScopes...)
			middlewareFunc(ctx)

			if ctx.IsAborted() {
				assert.Equal(t, tt.expectStatus, recorder.Code)
			} else {
				testHandler(ctx)
				assert.Equal(t, tt.expectStatus, recorder.Code)
			}

			mockKeyPair.AssertExpectations(t)
		})
	}
}

// Test GetAuthContext helper function
func TestGetAuthContext(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*gin.Context)
		expectAuth bool
	}{
		{
			name: "auth context exists",
			setup: func(ctx *gin.Context) {
				authCtx := createTestAuthContext()
				ctx.Set(AuthContextKey, authCtx)
			},
			expectAuth: true,
		},
		{
			name: "auth context missing",
			setup: func(ctx *gin.Context) {
				// Don't set auth context
			},
			expectAuth: false,
		},
		{
			name: "wrong type in context",
			setup: func(ctx *gin.Context) {
				ctx.Set(AuthContextKey, "wrong-type")
			},
			expectAuth: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, _ := createTestGinContext()

			if tt.setup != nil {
				tt.setup(ctx)
			}

			authCtx, exists := GetAuthContext(ctx)

			if tt.expectAuth {
				assert.True(t, exists)
				assert.NotNil(t, authCtx)
				assert.NotEmpty(t, authCtx.UserID)
			} else {
				assert.False(t, exists)
				assert.Nil(t, authCtx)
			}
		})
	}
}

// Test extractKeyPair helper function
func TestExtractKeyPair(t *testing.T) {
	tests := []struct {
		name        string
		authValue   string
		expectValid bool
		expectPub   string
		expectSec   string
	}{
		{
			name:        "valid key pair format",
			authValue:   "pk_test_123:sk_secret_456",
			expectValid: true,
			expectPub:   "pk_test_123",
			expectSec:   "sk_secret_456",
		},
		{
			name:        "valid with Bearer prefix",
			authValue:   "Bearer pk_test_123:sk_secret_456",
			expectValid: true,
			expectPub:   "pk_test_123",
			expectSec:   "sk_secret_456",
		},
		{
			name:        "missing colon separator",
			authValue:   "pk_test_123_sk_secret_456",
			expectValid: false,
		},
		{
			name:        "empty public key",
			authValue:   ":sk_secret_456",
			expectValid: false,
		},
		{
			name:        "empty secret key",
			authValue:   "pk_test_123:",
			expectValid: false,
		},
		{
			name:        "empty value",
			authValue:   "",
			expectValid: false,
		},
		{
			name:        "only Bearer prefix",
			authValue:   "Bearer",
			expectValid: false,
		},
		{
			name:        "multiple colons",
			authValue:   "pk_test:123:sk_secret:456",
			expectValid: true,
			expectPub:   "pk_test",
			expectSec:   "123:sk_secret:456", // Everything after first colon
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			publicKey, secretKey, valid := extractKeyPair(tt.authValue)

			assert.Equal(t, tt.expectValid, valid)
			if tt.expectValid {
				assert.Equal(t, tt.expectPub, publicKey)
				assert.Equal(t, tt.expectSec, secretKey)
			}
		})
	}
}

// Integration test for authentication flow
func TestAuthMiddleware_IntegrationFlow(t *testing.T) {
	middleware, _, mockKeyPair, _, _ := createTestAuthMiddleware()

	// Setup mock expectations
	authCtx := createTestAuthContext()
	mockKeyPair.On("AuthenticateWithKeyPair", mock.Anything, "pk_test_123", "sk_secret_456").Return(authCtx, nil)
	mockKeyPair.On("CheckKeyPairScopes", mock.Anything, authCtx.KeyPairID, []string{authDomain.ScopeGatewayRead}).Return(true, nil)

	// Create test route
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Apply middlewares in order
	router.Use(middleware.RequireKeyPair())
	router.Use(middleware.RequireScope(authDomain.ScopeGatewayRead))

	// Add test endpoint
	router.GET("/test", func(c *gin.Context) {
		authCtx, exists := GetAuthContext(c)
		assert.True(t, exists)
		assert.NotNil(t, authCtx)
		assert.NotNil(t, authCtx.KeyPairID)

		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// Make request
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer pk_test_123:sk_secret_456")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	// Assert successful flow
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]bool
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response["success"])

	mockKeyPair.AssertExpectations(t)
}