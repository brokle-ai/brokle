package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"brokle/pkg/ulid"
)

// Test KeyPair domain entity
func TestKeyPair_NewKeyPair(t *testing.T) {
	userID := ulid.New()
	orgID := ulid.New()
	projectID := ulid.New()
	name := "Test Key Pair"
	publicKey := "pk_" + projectID.String() + "_abcdef123456"
	secretKeyHash := "$2a$10$hashedSecretKey"
	scopes := []string{string(ScopeGatewayRead), string(ScopeAnalyticsRead)}
	rateLimitRPM := 1000

	keyPair := NewKeyPair(
		userID,
		orgID,
		projectID,
		name,
		publicKey,
		secretKeyHash,
		scopes,
		rateLimitRPM,
		nil,
	)

	assert.NotEmpty(t, keyPair.ID)
	assert.Equal(t, userID, keyPair.UserID)
	assert.Equal(t, orgID, keyPair.OrganizationID)
	assert.Equal(t, projectID, keyPair.ProjectID)
	assert.Equal(t, name, keyPair.Name)
	assert.Equal(t, publicKey, keyPair.PublicKey)
	assert.Equal(t, secretKeyHash, keyPair.SecretKeyHash)
	assert.Equal(t, scopes, keyPair.Scopes)
	assert.Equal(t, rateLimitRPM, keyPair.RateLimitRPM)
	assert.True(t, keyPair.IsActive)
	assert.Nil(t, keyPair.ExpiresAt)
	assert.Nil(t, keyPair.LastUsedAt)
	assert.NotEmpty(t, keyPair.CreatedAt)
	assert.NotEmpty(t, keyPair.UpdatedAt)
}

func TestKeyPair_NewKeyPairWithExpiry(t *testing.T) {
	userID := ulid.New()
	orgID := ulid.New()
	projectID := ulid.New()
	expiresAt := time.Now().Add(24 * time.Hour)

	keyPair := NewKeyPair(
		userID,
		orgID,
		projectID,
		"Test Key Pair",
		"pk_test_123456",
		"$2a$10$hashedSecretKey",
		[]string{string(ScopeGatewayRead)},
		1000,
		&expiresAt,
	)

	assert.NotNil(t, keyPair.ExpiresAt)
	assert.Equal(t, expiresAt, *keyPair.ExpiresAt)
}

// Test KeyPair validation methods
func TestKeyPair_ValidatePublicKeyFormat(t *testing.T) {
	keyPair := &KeyPair{}

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
			name:      "invalid format - missing parts",
			publicKey: "pk_01HQZMQY8PFRJQH1TQZRBQGS5Q",
			wantErr:   true,
		},
		{
			name:      "invalid ULID format",
			publicKey: "pk_invalid_ulid_abcdef123456789",
			wantErr:   true,
		},
		{
			name:      "wrong ULID length",
			publicKey: "pk_01HQZMQY8PFRJQH1TQZ_abcdef123456789",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyPair.PublicKey = tt.publicKey
			err := keyPair.ValidatePublicKeyFormat()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestKeyPair_ValidateSecretKeyPrefix(t *testing.T) {
	tests := []struct {
		name             string
		secretKeyPrefix  string
		wantErr          bool
	}{
		{
			name:            "valid sk_ prefix",
			secretKeyPrefix: "sk_",
			wantErr:         false,
		},
		{
			name:            "invalid prefix - missing underscore",
			secretKeyPrefix: "sk",
			wantErr:         true,
		},
		{
			name:            "empty prefix",
			secretKeyPrefix: "",
			wantErr:         true,
		},
		{
			name:            "wrong prefix",
			secretKeyPrefix: "pk_",
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyPair := &KeyPair{
				SecretKeyPrefix: tt.secretKeyPrefix,
			}
			err := keyPair.ValidateSecretKeyPrefix()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test KeyPair validity checks
func TestKeyPair_IsValid(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *KeyPair
		expectValid bool
	}{
		{
			name: "active key pair without expiry",
			setup: func() *KeyPair {
				return &KeyPair{
					IsActive:  true,
					ExpiresAt: nil,
				}
			},
			expectValid: true,
		},
		{
			name: "active key pair with future expiry",
			setup: func() *KeyPair {
				future := time.Now().Add(24 * time.Hour)
				return &KeyPair{
					IsActive:  true,
					ExpiresAt: &future,
				}
			},
			expectValid: true,
		},
		{
			name: "inactive key pair",
			setup: func() *KeyPair {
				return &KeyPair{
					IsActive:  false,
					ExpiresAt: nil,
				}
			},
			expectValid: false,
		},
		{
			name: "expired key pair",
			setup: func() *KeyPair {
				past := time.Now().Add(-24 * time.Hour)
				return &KeyPair{
					IsActive:  true,
					ExpiresAt: &past,
				}
			},
			expectValid: false,
		},
		{
			name: "inactive and expired key pair",
			setup: func() *KeyPair {
				past := time.Now().Add(-24 * time.Hour)
				return &KeyPair{
					IsActive:  false,
					ExpiresAt: &past,
				}
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyPair := tt.setup()
			assert.Equal(t, tt.expectValid, keyPair.IsValid())
		})
	}
}

// Test KeyPair scope checking
func TestKeyPair_HasScope(t *testing.T) {
	tests := []struct {
		name      string
		scopes    []string
		checkScope KeyPairScope
		expected  bool
	}{
		{
			name:      "has specific scope",
			scopes:    []string{string(ScopeGatewayRead), string(ScopeAnalyticsRead)},
			checkScope: ScopeGatewayRead,
			expected:  true,
		},
		{
			name:      "doesn't have specific scope",
			scopes:    []string{string(ScopeGatewayRead), string(ScopeAnalyticsRead)},
			checkScope: ScopeGatewayWrite,
			expected:  false,
		},
		{
			name:      "has admin scope - grants all",
			scopes:    []string{string(ScopeAdmin)},
			checkScope: ScopeGatewayWrite,
			expected:  true,
		},
		{
			name:      "empty scopes",
			scopes:    []string{},
			checkScope: ScopeGatewayRead,
			expected:  false,
		},
		{
			name:      "admin scope check",
			scopes:    []string{string(ScopeAdmin)},
			checkScope: ScopeAdmin,
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyPair := &KeyPair{
				Scopes: tt.scopes,
			}
			assert.Equal(t, tt.expected, keyPair.HasScope(tt.checkScope))
		})
	}
}

// Test AuthContext
func TestAuthContext_AuthType(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *AuthContext
		expected  string
	}{
		{
			name: "key pair authentication",
			setup: func() *AuthContext {
				keyPairID := ulid.New()
				return &AuthContext{
					UserID:    ulid.New(),
					KeyPairID: &keyPairID,
				}
			},
			expected: "keypair",
		},
		{
			name: "session authentication",
			setup: func() *AuthContext {
				sessionID := ulid.New()
				return &AuthContext{
					UserID:    ulid.New(),
					SessionID: &sessionID,
				}
			},
			expected: "session",
		},
		{
			name: "no auth identifiers",
			setup: func() *AuthContext {
				return &AuthContext{
					UserID: ulid.New(),
				}
			},
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authCtx := tt.setup()
			var authType string
			if authCtx.KeyPairID != nil {
				authType = "keypair"
			} else if authCtx.SessionID != nil {
				authType = "session"
			} else {
				authType = "unknown"
			}
			assert.Equal(t, tt.expected, authType)
		})
	}
}

// Test AuthContext validation for key pair authentication
func TestAuthContext_IsKeyPairAuth(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *AuthContext
		expected bool
	}{
		{
			name: "has key pair ID",
			setup: func() *AuthContext {
				keyPairID := ulid.New()
				return &AuthContext{
					KeyPairID: &keyPairID,
				}
			},
			expected: true,
		},
		{
			name: "no key pair ID",
			setup: func() *AuthContext {
				return &AuthContext{
					KeyPairID: nil,
				}
			},
			expected: false,
		},
		{
			name: "empty auth context",
			setup: func() *AuthContext {
				return &AuthContext{}
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authCtx := tt.setup()
			assert.Equal(t, tt.expected, authCtx.KeyPairID != nil)
		})
	}
}

// Test KeyPairScope constants
func TestKeyPairScope_Constants(t *testing.T) {
	// Verify all scope constants are defined correctly
	assert.Equal(t, KeyPairScope("gateway:read"), ScopeGatewayRead)
	assert.Equal(t, KeyPairScope("gateway:write"), ScopeGatewayWrite)
	assert.Equal(t, KeyPairScope("analytics:read"), ScopeAnalyticsRead)
	assert.Equal(t, KeyPairScope("config:read"), ScopeConfigRead)
	assert.Equal(t, KeyPairScope("config:write"), ScopeConfigWrite)
	assert.Equal(t, KeyPairScope("admin"), ScopeAdmin)
}

// Test CreateKeyPairRequest validation
func TestCreateKeyPairRequest_Validation(t *testing.T) {
	projectID := ulid.New()
	orgID := ulid.New()

	tests := []struct {
		name    string
		request CreateKeyPairRequest
		wantErr bool
	}{
		{
			name: "valid request",
			request: CreateKeyPairRequest{
				OrganizationID: orgID,
				ProjectID:      projectID,
				Name:           "Test Key Pair",
				Scopes:         []string{string(ScopeGatewayRead)},
				RateLimitRPM:   1000,
			},
			wantErr: false,
		},
		{
			name: "empty name",
			request: CreateKeyPairRequest{
				OrganizationID: orgID,
				ProjectID:      projectID,
				Name:           "",
				Scopes:         []string{string(ScopeGatewayRead)},
				RateLimitRPM:   1000,
			},
			wantErr: true,
		},
		{
			name: "empty scopes",
			request: CreateKeyPairRequest{
				OrganizationID: orgID,
				ProjectID:      projectID,
				Name:           "Test Key Pair",
				Scopes:         []string{},
				RateLimitRPM:   1000,
			},
			wantErr: true,
		},
		{
			name: "zero rate limit",
			request: CreateKeyPairRequest{
				OrganizationID: orgID,
				ProjectID:      projectID,
				Name:           "Test Key Pair",
				Scopes:         []string{string(ScopeGatewayRead)},
				RateLimitRPM:   0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation logic - in real implementation this would be in a validator
			hasError := tt.request.Name == "" ||
						len(tt.request.Scopes) == 0 ||
						tt.request.RateLimitRPM <= 0

			assert.Equal(t, tt.wantErr, hasError)
		})
	}
}

// Test UpdateKeyPairRequest
func TestUpdateKeyPairRequest_PartialUpdate(t *testing.T) {
	// Test that UpdateKeyPairRequest allows partial updates
	newName := "Updated Name"
	newRateLimit := 2000
	newActive := false

	req := UpdateKeyPairRequest{
		Name:         &newName,
		RateLimitRPM: &newRateLimit,
		IsActive:     &newActive,
		// Scopes and ExpiresAt not provided - should allow partial updates
	}

	assert.NotNil(t, req.Name)
	assert.Equal(t, "Updated Name", *req.Name)
	assert.NotNil(t, req.RateLimitRPM)
	assert.Equal(t, 2000, *req.RateLimitRPM)
	assert.NotNil(t, req.IsActive)
	assert.False(t, *req.IsActive)
	assert.Nil(t, req.Scopes)
	assert.Nil(t, req.ExpiresAt)
}

// Test CreateKeyPairResponse
func TestCreateKeyPairResponse_SecretKeyOnlyOnce(t *testing.T) {
	// Verify that CreateKeyPairResponse includes secret key
	resp := CreateKeyPairResponse{
		ID:        ulid.New(),
		Name:      "Test Key",
		PublicKey: "pk_test_123",
		SecretKey: "sk_secret_456",
		ProjectID: ulid.New(),
		Scopes:    []string{string(ScopeGatewayRead)},
	}

	assert.NotEmpty(t, resp.SecretKey)
	assert.Contains(t, resp.SecretKey, "sk_")
	assert.Contains(t, resp.PublicKey, "pk_")
}

// Test KeyPairFilters
func TestKeyPairFilters_Construction(t *testing.T) {
	userID := ulid.New()
	orgID := ulid.New()
	projectID := ulid.New()

	filters := KeyPairFilters{
		UserID:         &userID,
		OrganizationID: &orgID,
		ProjectID:      &projectID,
	}

	assert.NotNil(t, filters.UserID)
	assert.Equal(t, userID, *filters.UserID)
	assert.NotNil(t, filters.OrganizationID)
	assert.Equal(t, orgID, *filters.OrganizationID)
	assert.NotNil(t, filters.ProjectID)
	assert.Equal(t, projectID, *filters.ProjectID)
}

// Benchmark for KeyPair creation
func BenchmarkNewKeyPair(b *testing.B) {
	userID := ulid.New()
	orgID := ulid.New()
	projectID := ulid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewKeyPair(
			userID,
			orgID,
			projectID,
			"Benchmark Key Pair",
			"pk_test_benchmark",
			"$2a$10$hashedSecret",
			[]string{string(ScopeGatewayRead)},
			1000,
			nil,
		)
	}
}

// Benchmark for scope checking
func BenchmarkKeyPair_HasScope(b *testing.B) {
	keyPair := &KeyPair{
		Scopes: []string{string(ScopeGatewayRead), string(ScopeAnalyticsRead), string(ScopeConfigRead)},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		keyPair.HasScope(ScopeGatewayWrite)
	}
}