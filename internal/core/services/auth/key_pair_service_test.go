package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	authDomain "brokle/internal/core/domain/auth"
	userDomain "brokle/internal/core/domain/user"
	"brokle/pkg/ulid"
	appErrors "brokle/pkg/errors"
)

// Mock repositories
type mockKeyPairRepository struct {
	mock.Mock
}

func (m *mockKeyPairRepository) Create(ctx context.Context, keyPair *authDomain.KeyPair) error {
	args := m.Called(ctx, keyPair)
	return args.Error(0)
}

func (m *mockKeyPairRepository) GetByID(ctx context.Context, id ulid.ULID) (*authDomain.KeyPair, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authDomain.KeyPair), args.Error(1)
}

func (m *mockKeyPairRepository) GetByPublicKey(ctx context.Context, publicKey string) (*authDomain.KeyPair, error) {
	args := m.Called(ctx, publicKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authDomain.KeyPair), args.Error(1)
}

func (m *mockKeyPairRepository) GetByUserID(ctx context.Context, userID ulid.ULID) ([]*authDomain.KeyPair, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*authDomain.KeyPair), args.Error(1)
}

func (m *mockKeyPairRepository) GetByOrganizationID(ctx context.Context, orgID ulid.ULID) ([]*authDomain.KeyPair, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*authDomain.KeyPair), args.Error(1)
}

func (m *mockKeyPairRepository) GetByProjectID(ctx context.Context, projectID ulid.ULID) ([]*authDomain.KeyPair, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*authDomain.KeyPair), args.Error(1)
}

func (m *mockKeyPairRepository) GetByEnvironmentID(ctx context.Context, envID ulid.ULID) ([]*authDomain.KeyPair, error) {
	args := m.Called(ctx, envID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*authDomain.KeyPair), args.Error(1)
}

func (m *mockKeyPairRepository) Update(ctx context.Context, keyPair *authDomain.KeyPair) error {
	args := m.Called(ctx, keyPair)
	return args.Error(0)
}

func (m *mockKeyPairRepository) DeactivateKeyPair(ctx context.Context, keyID ulid.ULID) error {
	args := m.Called(ctx, keyID)
	return args.Error(0)
}

func (m *mockKeyPairRepository) MarkAsUsed(ctx context.Context, keyID ulid.ULID) error {
	args := m.Called(ctx, keyID)
	return args.Error(0)
}

func (m *mockKeyPairRepository) List(ctx context.Context, filters *authDomain.KeyPairFilters) ([]*authDomain.KeyPair, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*authDomain.KeyPair), args.Error(1)
}

func (m *mockKeyPairRepository) Delete(ctx context.Context, id ulid.ULID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockKeyPairRepository) UpdateLastUsed(ctx context.Context, id ulid.ULID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockKeyPairRepository) CheckKeyPairScopes(ctx context.Context, id ulid.ULID, requiredScopes []string) (bool, error) {
	args := m.Called(ctx, id, requiredScopes)
	return args.Bool(0), args.Error(1)
}

func (m *mockKeyPairRepository) GetKeyPairCount(ctx context.Context, userID ulid.ULID) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

func (m *mockKeyPairRepository) CleanupExpiredKeyPairs(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type mockUserRepository struct {
	mock.Mock
}

func (m *mockUserRepository) GetByID(ctx context.Context, id ulid.ULID) (*userDomain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDomain.User), args.Error(1)
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*userDomain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDomain.User), args.Error(1)
}

func (m *mockUserRepository) Create(ctx context.Context, user *userDomain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepository) Update(ctx context.Context, user *userDomain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepository) Delete(ctx context.Context, id ulid.ULID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockUserRepository) UpdatePassword(ctx context.Context, userID ulid.ULID, hashedPassword string) error {
	args := m.Called(ctx, userID, hashedPassword)
	return args.Error(0)
}

func (m *mockUserRepository) UpdateEmailVerification(ctx context.Context, userID ulid.ULID, verified bool) error {
	args := m.Called(ctx, userID, verified)
	return args.Error(0)
}

func (m *mockUserRepository) UpdateLastLoginAt(ctx context.Context, userID ulid.ULID, lastLoginAt time.Time) error {
	args := m.Called(ctx, userID, lastLoginAt)
	return args.Error(0)
}

func (m *mockUserRepository) CompleteOnboarding(ctx context.Context, userID ulid.ULID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *mockUserRepository) CreateOnboardingQuestion(ctx context.Context, question *userDomain.OnboardingQuestion) error {
	args := m.Called(ctx, question)
	return args.Error(0)
}

func (m *mockUserRepository) GetActiveOnboardingQuestions(ctx context.Context) ([]*userDomain.OnboardingQuestion, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*userDomain.OnboardingQuestion), args.Error(1)
}

func (m *mockUserRepository) GetOnboardingQuestionByID(ctx context.Context, id ulid.ULID) (*userDomain.OnboardingQuestion, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDomain.OnboardingQuestion), args.Error(1)
}

func (m *mockUserRepository) GetActiveOnboardingQuestionCount(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *mockUserRepository) CreateProfile(ctx context.Context, profile *userDomain.UserProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *mockUserRepository) GetProfile(ctx context.Context, userID ulid.ULID) (*userDomain.UserProfile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDomain.UserProfile), args.Error(1)
}

func (m *mockUserRepository) UpdateProfile(ctx context.Context, profile *userDomain.UserProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

// Test helper functions
func createTestKeyPairService() (*keyPairService, *mockKeyPairRepository, *mockUserRepository) {
	mockKeyPairRepo := &mockKeyPairRepository{}
	mockUserRepo := &mockUserRepository{}
	service := &keyPairService{
		keyPairRepo: mockKeyPairRepo,
		userRepo:    mockUserRepo,
	}
	return service, mockKeyPairRepo, mockUserRepo
}

func createTestUser() *userDomain.User {
	userID := ulid.New()
	return &userDomain.User{
		ID:        userID,
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}
}

func createTestKeyPair() *authDomain.KeyPair {
	userID := ulid.New()
	orgID := ulid.New()
	projectID := ulid.New()

	keyPair := authDomain.NewKeyPair(
		userID,
		orgID,
		projectID,
		"Test Key Pair",
		"pk_"+projectID.String()+"_abcdef123456",
		"$2a$10$hashedSecretKey",
		[]string{string(string(authDomain.ScopeGatewayRead)), string(string(authDomain.ScopeAnalyticsRead))},
		1000,
		nil,
	)

	return keyPair
}

// Tests for CreateKeyPair
func TestKeyPairService_CreateKeyPair(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*mockKeyPairRepository, *mockUserRepository)
		userID  ulid.ULID
		request *authDomain.CreateKeyPairRequest
		wantErr bool
		errType string
	}{
		{
			name: "successful key pair creation",
			setup: func(keyPairRepo *mockKeyPairRepository, userRepo *mockUserRepository) {
				user := createTestUser()
				userRepo.On("GetByID", mock.Anything, mock.Anything).Return(user, nil)
				keyPairRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
			userID: ulid.New(),
			request: &authDomain.CreateKeyPairRequest{
				OrganizationID: ulid.New(),
				ProjectID:      ulid.New(),
				Name:           "Test Key Pair",
				Scopes:         []string{string(authDomain.ScopeGatewayRead)},
				RateLimitRPM:   1000,
			},
			wantErr: false,
		},
		{
			name: "user not found",
			setup: func(keyPairRepo *mockKeyPairRepository, userRepo *mockUserRepository) {
				userRepo.On("GetByID", mock.Anything, mock.Anything).Return(nil, userDomain.ErrNotFound)
			},
			userID: ulid.New(),
			request: &authDomain.CreateKeyPairRequest{
				OrganizationID: ulid.New(),
				ProjectID:      ulid.New(),
				Name:           "Test Key Pair",
				Scopes:         []string{string(authDomain.ScopeGatewayRead)},
				RateLimitRPM:   1000,
			},
			wantErr: true,
			errType: "not_found",
		},
		{
			name: "repository creation fails",
			setup: func(keyPairRepo *mockKeyPairRepository, userRepo *mockUserRepository) {
				user := createTestUser()
				userRepo.On("GetByID", mock.Anything, mock.Anything).Return(user, nil)
				keyPairRepo.On("Create", mock.Anything, mock.Anything).Return(assert.AnError)
			},
			userID: ulid.New(),
			request: &authDomain.CreateKeyPairRequest{
				OrganizationID: ulid.New(),
				ProjectID:      ulid.New(),
				Name:           "Test Key Pair",
				Scopes:         []string{string(authDomain.ScopeGatewayRead)},
				RateLimitRPM:   1000,
			},
			wantErr: true,
			errType: "internal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockKeyPairRepo, mockUserRepo := createTestKeyPairService()

			if tt.setup != nil {
				tt.setup(mockKeyPairRepo, mockUserRepo)
			}

			ctx := context.Background()
			resp, err := service.CreateKeyPair(ctx, tt.userID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)

				// Check error type if specified
				if tt.errType != "" {
					switch tt.errType {
					case "not_found":
						appErr, ok := err.(*appErrors.AppError)
						assert.True(t, ok)
						assert.Equal(t, appErrors.NotFoundError, appErr.Type)
					case "internal":
						appErr, ok := err.(*appErrors.AppError)
						assert.True(t, ok)
						assert.Equal(t, appErrors.InternalError, appErr.Type)
					}
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotEmpty(t, resp.PublicKey)
				assert.NotEmpty(t, resp.SecretKey)
				assert.True(t, len(resp.PublicKey) > 10)
				assert.True(t, len(resp.SecretKey) > 10)
				assert.Equal(t, tt.request.Name, resp.Name)
				assert.Equal(t, tt.request.Scopes, resp.Scopes)
			}

			mockKeyPairRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

// Tests for ValidateKeyPair
func TestKeyPairService_ValidateKeyPair(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*mockKeyPairRepository, *mockUserRepository)
		publicKey string
		secretKey string
		wantErr   bool
		errType   string
	}{
		{
			name: "successful validation",
			setup: func(keyPairRepo *mockKeyPairRepository, userRepo *mockUserRepository) {
				keyPair := createTestKeyPair()
				keyPair.IsActive = true
				// Set a known secret key hash for testing
				keyPair.SecretKeyHash = "$2a$10$N2P.vMGNXmfXhcJM8HvTDeXK6K7G8C3W8h/Wz.Gl.YQy1Lrl3ht5W" // hash of "test-secret"
				keyPairRepo.On("GetByPublicKey", mock.Anything, mock.Anything).Return(keyPair, nil)
				keyPairRepo.On("MarkAsUsed", mock.Anything, keyPair.ID).Return(nil)
			},
			publicKey: "pk_test_abcdef123456",
			secretKey: "test-secret",
			wantErr:   false,
		},
		{
			name: "public key not found",
			setup: func(keyPairRepo *mockKeyPairRepository, userRepo *mockUserRepository) {
				keyPairRepo.On("GetByPublicKey", mock.Anything, mock.Anything).Return(nil, assert.AnError)
			},
			publicKey: "pk_nonexistent_123456",
			secretKey: "test-secret",
			wantErr:   true,
			errType:   "unauthorized",
		},
		{
			name: "inactive key pair",
			setup: func(keyPairRepo *mockKeyPairRepository, userRepo *mockUserRepository) {
				keyPair := createTestKeyPair()
				keyPair.IsActive = false
				keyPairRepo.On("GetByPublicKey", mock.Anything, mock.Anything).Return(keyPair, nil)
			},
			publicKey: "pk_test_abcdef123456",
			secretKey: "test-secret",
			wantErr:   true,
			errType:   "unauthorized",
		},
		{
			name: "wrong secret key",
			setup: func(keyPairRepo *mockKeyPairRepository, userRepo *mockUserRepository) {
				keyPair := createTestKeyPair()
				keyPair.IsActive = true
				keyPair.SecretKeyHash = "$2a$10$N2P.vMGNXmfXhcJM8HvTDeXK6K7G8C3W8h/Wz.Gl.YQy1Lrl3ht5W" // hash of "test-secret"
				keyPairRepo.On("GetByPublicKey", mock.Anything, mock.Anything).Return(keyPair, nil)
			},
			publicKey: "pk_test_abcdef123456",
			secretKey: "wrong-secret",
			wantErr:   true,
			errType:   "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockKeyPairRepo, mockUserRepo := createTestKeyPairService()

			if tt.setup != nil {
				tt.setup(mockKeyPairRepo, mockUserRepo)
			}

			ctx := context.Background()
			keyPair, err := service.ValidateKeyPair(ctx, tt.publicKey, tt.secretKey)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, keyPair)

				if tt.errType == "unauthorized" {
					appErr, ok := err.(*appErrors.AppError)
					assert.True(t, ok)
					assert.Equal(t, appErrors.UnauthorizedError, appErr.Type)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, keyPair)
				assert.True(t, keyPair.IsActive)
			}

			mockKeyPairRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

// Tests for GenerateKeyPair
func TestKeyPairService_GenerateKeyPair(t *testing.T) {
	service, _, _ := createTestKeyPairService()
	projectID := ulid.New()

	ctx := context.Background()
	publicKey, secretKey, err := service.GenerateKeyPair(ctx, projectID)

	require.NoError(t, err)
	assert.NotEmpty(t, publicKey)
	assert.NotEmpty(t, secretKey)

	// Verify public key format: pk_projectId_random
	assert.True(t, len(publicKey) > 30)
	assert.Contains(t, publicKey, "pk_")
	assert.Contains(t, publicKey, projectID.String())

	// Verify secret key format: sk_random
	assert.True(t, len(secretKey) > 10)
	assert.True(t, len(secretKey) > 3 && secretKey[:3] == "sk_")

	// Verify keys are different on multiple calls
	publicKey2, secretKey2, err2 := service.GenerateKeyPair(ctx, projectID)
	require.NoError(t, err2)
	assert.NotEqual(t, publicKey, publicKey2)
	assert.NotEqual(t, secretKey, secretKey2)
}

// Tests for AuthenticateWithKeyPair
func TestKeyPairService_AuthenticateWithKeyPair(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*mockKeyPairRepository, *mockUserRepository)
		publicKey string
		secretKey string
		wantErr   bool
	}{
		{
			name: "successful authentication",
			setup: func(keyPairRepo *mockKeyPairRepository, userRepo *mockUserRepository) {
				keyPair := createTestKeyPair()
				keyPair.IsActive = true
				keyPair.SecretKeyHash = "$2a$10$N2P.vMGNXmfXhcJM8HvTDeXK6K7G8C3W8h/Wz.Gl.YQy1Lrl3ht5W" // hash of "test-secret"
				keyPairRepo.On("GetByPublicKey", mock.Anything, mock.Anything).Return(keyPair, nil)
				keyPairRepo.On("MarkAsUsed", mock.Anything, keyPair.ID).Return(nil)
			},
			publicKey: "pk_test_abcdef123456",
			secretKey: "test-secret",
			wantErr:   false,
		},
		{
			name: "validation fails",
			setup: func(keyPairRepo *mockKeyPairRepository, userRepo *mockUserRepository) {
				keyPairRepo.On("GetByPublicKey", mock.Anything, mock.Anything).Return(nil, assert.AnError)
			},
			publicKey: "pk_invalid_123456",
			secretKey: "test-secret",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockKeyPairRepo, mockUserRepo := createTestKeyPairService()

			if tt.setup != nil {
				tt.setup(mockKeyPairRepo, mockUserRepo)
			}

			ctx := context.Background()
			authCtx, err := service.AuthenticateWithKeyPair(ctx, tt.publicKey, tt.secretKey)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, authCtx)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, authCtx)
				assert.NotNil(t, authCtx.KeyPairID)
				assert.NotNil(t, authCtx.OrganizationID)
				assert.NotNil(t, authCtx.ProjectID)
				assert.NotEmpty(t, authCtx.Scopes)
			}

			mockKeyPairRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

// Tests for scope checking
func TestKeyPairService_CheckKeyPairScopes(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(*mockKeyPairRepository)
		keyID          ulid.ULID
		requiredScopes []string
		expectAllowed  bool
		wantErr        bool
	}{
		{
			name: "has admin scope - allows everything",
			setup: func(keyPairRepo *mockKeyPairRepository) {
				keyPair := createTestKeyPair()
				keyPair.Scopes = []string{string(authDomain.ScopeAdmin)}
				keyPairRepo.On("GetByID", mock.Anything, mock.Anything).Return(keyPair, nil)
			},
			keyID:          ulid.New(),
			requiredScopes: []string{string(authDomain.ScopeGatewayWrite), string(authDomain.ScopeAnalyticsRead)},
			expectAllowed:  true,
			wantErr:        false,
		},
		{
			name: "has required scopes",
			setup: func(keyPairRepo *mockKeyPairRepository) {
				keyPair := createTestKeyPair()
				keyPair.Scopes = []string{string(authDomain.ScopeGatewayRead), string(authDomain.ScopeAnalyticsRead), string(authDomain.ScopeConfigRead)}
				keyPairRepo.On("GetByID", mock.Anything, mock.Anything).Return(keyPair, nil)
			},
			keyID:          ulid.New(),
			requiredScopes: []string{string(authDomain.ScopeGatewayRead), string(authDomain.ScopeAnalyticsRead)},
			expectAllowed:  true,
			wantErr:        false,
		},
		{
			name: "missing required scope",
			setup: func(keyPairRepo *mockKeyPairRepository) {
				keyPair := createTestKeyPair()
				keyPair.Scopes = []string{string(authDomain.ScopeGatewayRead)}
				keyPairRepo.On("GetByID", mock.Anything, mock.Anything).Return(keyPair, nil)
			},
			keyID:          ulid.New(),
			requiredScopes: []string{string(authDomain.ScopeGatewayRead), string(authDomain.ScopeGatewayWrite)},
			expectAllowed:  false,
			wantErr:        false,
		},
		{
			name: "key pair not found",
			setup: func(keyPairRepo *mockKeyPairRepository) {
				keyPairRepo.On("GetByID", mock.Anything, mock.Anything).Return(nil, assert.AnError)
			},
			keyID:          ulid.New(),
			requiredScopes: []string{string(authDomain.ScopeGatewayRead)},
			expectAllowed:  false,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockKeyPairRepo, _ := createTestKeyPairService()

			if tt.setup != nil {
				tt.setup(mockKeyPairRepo)
			}

			ctx := context.Background()
			allowed, err := service.CheckKeyPairScopes(ctx, tt.keyID, tt.requiredScopes)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectAllowed, allowed)
			}

			mockKeyPairRepo.AssertExpectations(t)
		})
	}
}

// Test key pair validation methods
func TestKeyPairService_ValidatePublicKeyFormat(t *testing.T) {
	service, _, _ := createTestKeyPairService()
	ctx := context.Background()

	tests := []struct {
		name      string
		publicKey string
		wantErr   bool
	}{
		{
			name:      "valid public key format",
			publicKey: "pk_01HQZMQY8PFRJQH1TQZRBQGS5Q_abcdef123456789",
			wantErr:   false,
		},
		{
			name:      "missing pk_ prefix",
			publicKey: "01HQZMQY8PFRJQH1TQZRBQGS5Q_abcdef123456789",
			wantErr:   true,
		},
		{
			name:      "invalid ULID in project ID",
			publicKey: "pk_invalid_ulid_abcdef123456789",
			wantErr:   true,
		},
		{
			name:      "missing random suffix",
			publicKey: "pk_01HQZMQY8PFRJQH1TQZRBQGS5Q",
			wantErr:   true,
		},
		{
			name:      "wrong ULID length",
			publicKey: "pk_01HQZMQY8PFRJQH1TQZR_abcdef123456789",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidatePublicKeyFormat(ctx, tt.publicKey)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test project ID extraction
func TestKeyPairService_ExtractProjectIDFromPublicKey(t *testing.T) {
	service, _, _ := createTestKeyPairService()
	ctx := context.Background()

	projectID := ulid.New()
	publicKey := "pk_" + projectID.String() + "_abcdef123456789"

	extractedID, err := service.ExtractProjectIDFromPublicKey(ctx, publicKey)
	require.NoError(t, err)
	assert.Equal(t, projectID, extractedID)

	// Test with invalid format
	_, err = service.ExtractProjectIDFromPublicKey(ctx, "invalid_key")
	assert.Error(t, err)
}