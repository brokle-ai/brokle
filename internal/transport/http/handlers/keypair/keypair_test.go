package keypair

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"brokle/internal/config"
	authDomain "brokle/internal/core/domain/auth"
	"brokle/internal/transport/http/middleware"
	"brokle/pkg/response"
	"brokle/pkg/ulid"
)

// Mock KeyPairService
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

// Test helper functions
func createTestHandler() (*Handler, *mockKeyPairService) {
	mockService := &mockKeyPairService{}
	cfg := &config.Config{}
	logger := logrus.New()

	handler := &Handler{
		config:         cfg,
		logger:         logger,
		keyPairService: mockService,
	}

	return handler, mockService
}

func createTestGinContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	return ctx, recorder
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
		Scopes:         []string{"gateway:read", "analytics:read"},
	}
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
		[]string{"gateway:read", "analytics:read"},
		1000,
		nil,
	)
}

// Tests for Create handler
func TestHandler_Create(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(*mockKeyPairService)
		authSetup    func(*gin.Context)
		requestBody  interface{}
		expectStatus int
		expectError  bool
	}{
		{
			name: "successful key pair creation",
			setup: func(mockService *mockKeyPairService) {
				createResp := &authDomain.CreateKeyPairResponse{
					ID:        ulid.New(),
					Name:      "Test Key Pair",
					PublicKey: "pk_test_123456",
					SecretKey: "sk_secret_789",
					ProjectID: ulid.New(),
					Scopes:    []string{"gateway:read"},
				}
				mockService.On("CreateKeyPair", mock.Anything, mock.Anything, mock.Anything).Return(createResp, nil)
			},
			authSetup: func(ctx *gin.Context) {
				authCtx := createTestAuthContext()
				ctx.Set(middleware.AuthContextKey, authCtx)
			},
			requestBody: CreateKeyPairRequest{
				Name:         "Test Key Pair",
				ProjectID:    ulid.New().String(),
				Scopes:       []string{"gateway:read"},
				RateLimitRPM: 1000,
			},
			expectStatus: http.StatusCreated,
			expectError:  false,
		},
		{
			name:  "missing auth context",
			setup: func(mockService *mockKeyPairService) {},
			authSetup: func(ctx *gin.Context) {
				// Don't set auth context
			},
			requestBody: CreateKeyPairRequest{
				Name:         "Test Key Pair",
				ProjectID:    ulid.New().String(),
				Scopes:       []string{"gateway:read"},
				RateLimitRPM: 1000,
			},
			expectStatus: http.StatusUnauthorized,
			expectError:  true,
		},
		{
			name:  "invalid request body",
			setup: func(mockService *mockKeyPairService) {},
			authSetup: func(ctx *gin.Context) {
				authCtx := createTestAuthContext()
				ctx.Set(middleware.AuthContextKey, authCtx)
			},
			requestBody: map[string]interface{}{
				"name": "", // Invalid: empty name
			},
			expectStatus: http.StatusBadRequest,
			expectError:  true,
		},
		{
			name: "service error",
			setup: func(mockService *mockKeyPairService) {
				mockService.On("CreateKeyPair", mock.Anything, mock.Anything, mock.Anything).Return(nil, assert.AnError)
			},
			authSetup: func(ctx *gin.Context) {
				authCtx := createTestAuthContext()
				ctx.Set(middleware.AuthContextKey, authCtx)
			},
			requestBody: CreateKeyPairRequest{
				Name:         "Test Key Pair",
				ProjectID:    ulid.New().String(),
				Scopes:       []string{"gateway:read"},
				RateLimitRPM: 1000,
			},
			expectStatus: http.StatusInternalServerError,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := createTestHandler()
			ctx, recorder := createTestGinContext()

			if tt.setup != nil {
				tt.setup(mockService)
			}

			if tt.authSetup != nil {
				tt.authSetup(ctx)
			}

			// Prepare request body
			requestBody, _ := json.Marshal(tt.requestBody)
			ctx.Request = httptest.NewRequest("POST", "/key-pairs", bytes.NewBuffer(requestBody))
			ctx.Request.Header.Set("Content-Type", "application/json")

			// Call handler
			handler.Create(ctx)

			// Assert response
			assert.Equal(t, tt.expectStatus, recorder.Code)

			if !tt.expectError && recorder.Code == http.StatusCreated {
				var resp response.SuccessResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.True(t, resp.Success)
				assert.NotNil(t, resp.Data)
			}

			mockService.AssertExpectations(t)
		})
	}
}

// Tests for List handler
func TestHandler_List(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(*mockKeyPairService)
		authSetup    func(*gin.Context)
		queryParams  map[string]string
		expectStatus int
		expectError  bool
	}{
		{
			name: "successful list by user",
			setup: func(mockService *mockKeyPairService) {
				keyPairs := []*authDomain.KeyPair{createTestKeyPair()}
				mockService.On("GetKeyPairsByUser", mock.Anything, mock.Anything).Return(keyPairs, nil)
			},
			authSetup: func(ctx *gin.Context) {
				authCtx := createTestAuthContext()
				ctx.Set(middleware.AuthContextKey, authCtx)
			},
			queryParams: map[string]string{
				"filter": "user",
			},
			expectStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name: "successful list by project",
			setup: func(mockService *mockKeyPairService) {
				keyPairs := []*authDomain.KeyPair{createTestKeyPair()}
				mockService.On("GetKeyPairsByProject", mock.Anything, mock.Anything).Return(keyPairs, nil)
			},
			authSetup: func(ctx *gin.Context) {
				authCtx := createTestAuthContext()
				ctx.Set(middleware.AuthContextKey, authCtx)
			},
			queryParams: map[string]string{
				"filter": "project",
			},
			expectStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:  "missing auth context",
			setup: func(mockService *mockKeyPairService) {},
			authSetup: func(ctx *gin.Context) {
				// Don't set auth context
			},
			queryParams: map[string]string{
				"filter": "user",
			},
			expectStatus: http.StatusUnauthorized,
			expectError:  true,
		},
		{
			name: "service error",
			setup: func(mockService *mockKeyPairService) {
				mockService.On("GetKeyPairsByUser", mock.Anything, mock.Anything).Return(nil, assert.AnError)
			},
			authSetup: func(ctx *gin.Context) {
				authCtx := createTestAuthContext()
				ctx.Set(middleware.AuthContextKey, authCtx)
			},
			queryParams: map[string]string{
				"filter": "user",
			},
			expectStatus: http.StatusInternalServerError,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := createTestHandler()
			ctx, recorder := createTestGinContext()

			if tt.setup != nil {
				tt.setup(mockService)
			}

			if tt.authSetup != nil {
				tt.authSetup(ctx)
			}

			// Set query parameters
			ctx.Request = httptest.NewRequest("GET", "/key-pairs", nil)
			q := ctx.Request.URL.Query()
			for key, value := range tt.queryParams {
				q.Add(key, value)
			}
			ctx.Request.URL.RawQuery = q.Encode()

			// Call handler
			handler.List(ctx)

			// Assert response
			assert.Equal(t, tt.expectStatus, recorder.Code)

			if !tt.expectError && recorder.Code == http.StatusOK {
				var resp response.SuccessResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.True(t, resp.Success)
			}

			mockService.AssertExpectations(t)
		})
	}
}

// Tests for GetByID handler
func TestHandler_GetByID(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(*mockKeyPairService)
		authSetup    func(*gin.Context)
		pathParams   map[string]string
		expectStatus int
		expectError  bool
	}{
		{
			name: "successful get by ID",
			setup: func(mockService *mockKeyPairService) {
				keyPair := createTestKeyPair()
				mockService.On("GetKeyPair", mock.Anything, mock.Anything).Return(keyPair, nil)
			},
			authSetup: func(ctx *gin.Context) {
				authCtx := createTestAuthContext()
				ctx.Set(middleware.AuthContextKey, authCtx)
			},
			pathParams: map[string]string{
				"projectId": ulid.New().String(),
				"keyPairId": ulid.New().String(),
			},
			expectStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:  "missing auth context",
			setup: func(mockService *mockKeyPairService) {},
			authSetup: func(ctx *gin.Context) {
				// Don't set auth context
			},
			pathParams: map[string]string{
				"projectId": ulid.New().String(),
				"keyPairId": ulid.New().String(),
			},
			expectStatus: http.StatusUnauthorized,
			expectError:  true,
		},
		{
			name:  "invalid key pair ID",
			setup: func(mockService *mockKeyPairService) {},
			authSetup: func(ctx *gin.Context) {
				authCtx := createTestAuthContext()
				ctx.Set(middleware.AuthContextKey, authCtx)
			},
			pathParams: map[string]string{
				"projectId": ulid.New().String(),
				"keyPairId": "invalid-id",
			},
			expectStatus: http.StatusBadRequest,
			expectError:  true,
		},
		{
			name: "key pair not found",
			setup: func(mockService *mockKeyPairService) {
				mockService.On("GetKeyPair", mock.Anything, mock.Anything).Return(nil, assert.AnError)
			},
			authSetup: func(ctx *gin.Context) {
				authCtx := createTestAuthContext()
				ctx.Set(middleware.AuthContextKey, authCtx)
			},
			pathParams: map[string]string{
				"projectId": ulid.New().String(),
				"keyPairId": ulid.New().String(),
			},
			expectStatus: http.StatusInternalServerError,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := createTestHandler()
			ctx, recorder := createTestGinContext()

			if tt.setup != nil {
				tt.setup(mockService)
			}

			if tt.authSetup != nil {
				tt.authSetup(ctx)
			}

			// Set path parameters
			for key, value := range tt.pathParams {
				ctx.Params = append(ctx.Params, gin.Param{Key: key, Value: value})
			}

			ctx.Request = httptest.NewRequest("GET", "/key-pairs/"+tt.pathParams["keyPairId"], nil)

			// Call handler
			handler.GetByID(ctx)

			// Assert response
			assert.Equal(t, tt.expectStatus, recorder.Code)

			if !tt.expectError && recorder.Code == http.StatusOK {
				var resp response.SuccessResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.True(t, resp.Success)
				assert.NotNil(t, resp.Data)
			}

			mockService.AssertExpectations(t)
		})
	}
}

// Tests for Delete handler
func TestHandler_Delete(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(*mockKeyPairService)
		authSetup    func(*gin.Context)
		pathParams   map[string]string
		expectStatus int
		expectError  bool
	}{
		{
			name: "successful deletion",
			setup: func(mockService *mockKeyPairService) {
				keyPair := createTestKeyPair()
				mockService.On("GetKeyPair", mock.Anything, mock.Anything).Return(keyPair, nil)
				mockService.On("RevokeKeyPair", mock.Anything, mock.Anything).Return(nil)
			},
			authSetup: func(ctx *gin.Context) {
				authCtx := createTestAuthContext()
				ctx.Set(middleware.AuthContextKey, authCtx)
			},
			pathParams: map[string]string{
				"projectId": ulid.New().String(),
				"keyPairId": ulid.New().String(),
			},
			expectStatus: http.StatusNoContent,
			expectError:  false,
		},
		{
			name:  "missing auth context",
			setup: func(mockService *mockKeyPairService) {},
			authSetup: func(ctx *gin.Context) {
				// Don't set auth context
			},
			pathParams: map[string]string{
				"projectId": ulid.New().String(),
				"keyPairId": ulid.New().String(),
			},
			expectStatus: http.StatusUnauthorized,
			expectError:  true,
		},
		{
			name: "key pair not found",
			setup: func(mockService *mockKeyPairService) {
				mockService.On("GetKeyPair", mock.Anything, mock.Anything).Return(nil, assert.AnError)
			},
			authSetup: func(ctx *gin.Context) {
				authCtx := createTestAuthContext()
				ctx.Set(middleware.AuthContextKey, authCtx)
			},
			pathParams: map[string]string{
				"projectId": ulid.New().String(),
				"keyPairId": ulid.New().String(),
			},
			expectStatus: http.StatusInternalServerError,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := createTestHandler()
			ctx, recorder := createTestGinContext()

			if tt.setup != nil {
				tt.setup(mockService)
			}

			if tt.authSetup != nil {
				tt.authSetup(ctx)
			}

			// Set path parameters
			for key, value := range tt.pathParams {
				ctx.Params = append(ctx.Params, gin.Param{Key: key, Value: value})
			}

			ctx.Request = httptest.NewRequest("DELETE", "/key-pairs/"+tt.pathParams["keyPairId"], nil)

			// Call handler
			handler.Delete(ctx)

			// Assert response
			assert.Equal(t, tt.expectStatus, recorder.Code)

			mockService.AssertExpectations(t)
		})
	}
}

// Integration test for complete handler flow
func TestHandler_IntegrationFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	handler, mockService := createTestHandler()

	// Setup mock expectations for integration test
	createResp := &authDomain.CreateKeyPairResponse{
		ID:        ulid.New(),
		Name:      "Integration Test Key",
		PublicKey: "pk_test_integration",
		SecretKey: "sk_secret_integration",
		ProjectID: ulid.New(),
		Scopes:    []string{"gateway:read"},
	}

	keyPairs := []*authDomain.KeyPair{createTestKeyPair()}

	mockService.On("CreateKeyPair", mock.Anything, mock.Anything, mock.Anything).Return(createResp, nil)
	mockService.On("GetKeyPairsByUser", mock.Anything, mock.Anything).Return(keyPairs, nil)

	// Setup middleware to add auth context
	router.Use(func(c *gin.Context) {
		authCtx := createTestAuthContext()
		c.Set(middleware.AuthContextKey, authCtx)
		c.Next()
	})

	// Setup routes
	router.POST("/key-pairs", handler.Create)
	router.GET("/key-pairs", handler.List)

	// Test Create
	t.Run("create key pair", func(t *testing.T) {
		requestBody := CreateKeyPairRequest{
			Name:         "Integration Test Key",
			ProjectID:    ulid.New().String(),
			Scopes:       []string{"gateway:read"},
			RateLimitRPM: 1000,
		}

		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/key-pairs", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusCreated, recorder.Code)

		var resp response.SuccessResponse
		err := json.Unmarshal(recorder.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.True(t, resp.Success)
	})

	// Test List
	t.Run("list key pairs", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/key-pairs?filter=user", nil)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var resp response.SuccessResponse
		err := json.Unmarshal(recorder.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.True(t, resp.Success)
	})

	mockService.AssertExpectations(t)
}