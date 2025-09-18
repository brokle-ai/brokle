package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	authDomain "brokle/internal/core/domain/auth"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/ulid"
)

// apiKeyService implements the authDomain.APIKeyService interface
type apiKeyService struct {
	apiKeyRepo             authDomain.APIKeyRepository
	organizationMemberRepo authDomain.OrganizationMemberRepository
}

// NewAPIKeyService creates a new API key service instance
func NewAPIKeyService(
	apiKeyRepo authDomain.APIKeyRepository,
	organizationMemberRepo authDomain.OrganizationMemberRepository,
) authDomain.APIKeyService {
	return &apiKeyService{
		apiKeyRepo:             apiKeyRepo,
		organizationMemberRepo: organizationMemberRepo,
	}
}

// Environment constants for key prefixes
const (
	EnvProduction  = "production"
	EnvTesting     = "testing"
	EnvStaging     = "staging"
	EnvDevelopment = "development"
	EnvLocal       = "local"
)

// Key prefix mapping
var envPrefixMap = map[string]string{
	EnvProduction:  "bk_live_",
	EnvTesting:     "bk_test_",
	EnvStaging:     "bk_test_",
	EnvDevelopment: "bk_dev_",
	EnvLocal:       "bk_dev_",
}

// CreateAPIKey creates a new API key
func (s *apiKeyService) CreateAPIKey(ctx context.Context, userID ulid.ULID, req *authDomain.CreateAPIKeyRequest) (*authDomain.CreateAPIKeyResponse, error) {
	// TODO: Validate user has permission to create keys in the organization
	// For now, skip membership validation - will be implemented when organization service is ready

	// Validate and set default environment
	defaultEnv := req.DefaultEnvironment
	if defaultEnv == "" {
		defaultEnv = authDomain.DefaultEnvironmentName // "default"
	}

	// Validate environment name according to rules
	if err := authDomain.ValidateEnvironmentName(defaultEnv); err != nil {
		return nil, appErrors.NewBadRequestError("Invalid environment name", err.Error())
	}

	// Determine environment type for key prefix (default to development for safety)
	envType := EnvDevelopment
	if defaultEnv == "production" || defaultEnv == "prod" {
		envType = EnvProduction
	} else if defaultEnv == "staging" || defaultEnv == "stage" {
		envType = EnvStaging
	}

	// Generate API key
	apiKey, err := s.generateAPIKey(envType)
	if err != nil {
		return nil, fmt.Errorf("generate API key: %w", err)
	}

	// Hash the key for storage
	keyHash := s.hashAPIKey(apiKey)
	keyPrefix := s.extractKeyPrefix(apiKey)

	// Create API key entity with new schema
	apiKeyEntity := authDomain.NewAPIKey(
		userID,
		req.OrganizationID,
		req.ProjectID, // Now required
		req.Name,
		keyPrefix,
		keyHash,
		defaultEnv, // Default environment
		req.Scopes,
		req.RateLimitRPM,
		req.ExpiresAt,
	)

	// Save to database
	if err := s.apiKeyRepo.Create(ctx, apiKeyEntity); err != nil {
		return nil, fmt.Errorf("create API key: %w", err)
	}

	// Return response with the actual key (only shown once)
	return &authDomain.CreateAPIKeyResponse{
		ID:        apiKeyEntity.ID,
		Name:      apiKeyEntity.Name,
		Key:       apiKey, // Full key - only returned once
		KeyPrefix: keyPrefix,
		Scopes:    apiKeyEntity.Scopes,
		ExpiresAt: apiKeyEntity.ExpiresAt,
	}, nil
}

// GetAPIKey retrieves an API key by ID
func (s *apiKeyService) GetAPIKey(ctx context.Context, keyID ulid.ULID) (*authDomain.APIKey, error) {
	apiKey, err := s.apiKeyRepo.GetByID(ctx, keyID)
	if err != nil {
		return nil, fmt.Errorf("get API key: %w", err)
	}
	return apiKey, nil
}

// GetAPIKeys retrieves API keys based on filters
func (s *apiKeyService) GetAPIKeys(ctx context.Context, filters *authDomain.APIKeyFilters) ([]*authDomain.APIKey, error) {
	// Use existing repository methods based on filters
	if filters.ProjectID != nil {
		return s.apiKeyRepo.GetByProjectID(ctx, *filters.ProjectID)
	}
	if filters.OrganizationID != nil {
		return s.apiKeyRepo.GetByOrganizationID(ctx, *filters.OrganizationID)
	}
	if filters.UserID != nil {
		return s.apiKeyRepo.GetByUserID(ctx, *filters.UserID)
	}

	// If no specific filters, return empty array for now
	// TODO: Implement comprehensive filtering in repository including environment tags
	return []*authDomain.APIKey{}, nil
}

// UpdateAPIKey updates an existing API key
func (s *apiKeyService) UpdateAPIKey(ctx context.Context, keyID ulid.ULID, req *authDomain.UpdateAPIKeyRequest) error {
	// Get existing key
	apiKey, err := s.apiKeyRepo.GetByID(ctx, keyID)
	if err != nil {
		return fmt.Errorf("get API key: %w", err)
	}

	// Update fields
	if req.Name != nil {
		apiKey.Name = *req.Name
	}
	if req.DefaultEnvironment != nil {
		// Validate environment name
		if err := authDomain.ValidateEnvironmentName(*req.DefaultEnvironment); err != nil {
			return appErrors.NewBadRequestError("Invalid environment name", err.Error())
		}
		apiKey.DefaultEnvironment = *req.DefaultEnvironment
	}
	if req.Scopes != nil {
		apiKey.Scopes = req.Scopes
	}
	if req.RateLimitRPM != nil {
		apiKey.RateLimitRPM = *req.RateLimitRPM
	}
	if req.IsActive != nil {
		apiKey.IsActive = *req.IsActive
	}

	apiKey.UpdatedAt = time.Now()

	// Save changes
	if err := s.apiKeyRepo.Update(ctx, apiKey); err != nil {
		return fmt.Errorf("update API key: %w", err)
	}

	return nil
}

// RevokeAPIKey revokes an API key
func (s *apiKeyService) RevokeAPIKey(ctx context.Context, keyID ulid.ULID) error {
	if err := s.apiKeyRepo.DeactivateAPIKey(ctx, keyID); err != nil {
		return fmt.Errorf("revoke API key: %w", err)
	}
	return nil
}

// ValidateAPIKey validates an API key and returns the key entity
func (s *apiKeyService) ValidateAPIKey(ctx context.Context, apiKey string) (*authDomain.APIKey, error) {
	// Validate key format
	if !s.isValidKeyFormat(apiKey) {
		return nil, appErrors.NewUnauthorizedError("Invalid API key format")
	}

	// Hash the key for lookup
	keyHash := s.hashAPIKey(apiKey)

	// Get key from database
	key, err := s.apiKeyRepo.GetByKeyHash(ctx, keyHash)
	if err != nil {
		return nil, appErrors.NewUnauthorizedError("Invalid API key")
	}

	// Check if key is active
	if !key.IsActive {
		return nil, appErrors.NewUnauthorizedError("API key is inactive")
	}

	// Check if key is expired
	if key.IsExpired() {
		return nil, appErrors.NewUnauthorizedError("API key has expired")
	}

	// Update last used timestamp asynchronously
	go func() {
		bgCtx := context.Background()
		s.apiKeyRepo.MarkAsUsed(bgCtx, key.ID)
	}()

	return key, nil
}

// UpdateLastUsed updates the last used timestamp
func (s *apiKeyService) UpdateLastUsed(ctx context.Context, keyID ulid.ULID) error {
	return s.apiKeyRepo.MarkAsUsed(ctx, keyID)
}

// CheckRateLimit checks if the API key has exceeded rate limits
func (s *apiKeyService) CheckRateLimit(ctx context.Context, keyID ulid.ULID) (bool, error) {
	// TODO: Implement rate limiting logic with Redis
	// For now, always allow requests
	return true, nil
}

// GetAPIKeyContext creates an AuthContext from an API key
func (s *apiKeyService) GetAPIKeyContext(ctx context.Context, keyID ulid.ULID) (*authDomain.AuthContext, error) {
	apiKey, err := s.apiKeyRepo.GetByID(ctx, keyID)
	if err != nil {
		return nil, fmt.Errorf("get API key: %w", err)
	}

	return &authDomain.AuthContext{
		UserID:   apiKey.UserID,
		APIKeyID: &apiKey.ID,
	}, nil
}

// CanAPIKeyAccessResource checks if an API key can access a specific resource
func (s *apiKeyService) CanAPIKeyAccessResource(ctx context.Context, keyID ulid.ULID, resource string) (bool, error) {
	apiKey, err := s.apiKeyRepo.GetByID(ctx, keyID)
	if err != nil {
		return false, fmt.Errorf("get API key: %w", err)
	}

	// Check if resource is in scopes
	for _, scope := range apiKey.Scopes {
		if scope == "*" || scope == resource {
			return true, nil
		}
	}

	return false, nil
}

// Scoped access methods
func (s *apiKeyService) GetAPIKeysByUser(ctx context.Context, userID ulid.ULID) ([]*authDomain.APIKey, error) {
	return s.apiKeyRepo.GetByUserID(ctx, userID)
}

func (s *apiKeyService) GetAPIKeysByOrganization(ctx context.Context, orgID ulid.ULID) ([]*authDomain.APIKey, error) {
	return s.apiKeyRepo.GetByOrganizationID(ctx, orgID)
}

func (s *apiKeyService) GetAPIKeysByProject(ctx context.Context, projectID ulid.ULID) ([]*authDomain.APIKey, error) {
	return s.apiKeyRepo.GetByProjectID(ctx, projectID)
}

// Private helper methods

// generateAPIKey generates a new API key with environment prefix
func (s *apiKeyService) generateAPIKey(envType string) (string, error) {
	// Get prefix for environment
	prefix, exists := envPrefixMap[envType]
	if !exists {
		prefix = envPrefixMap[EnvDevelopment] // Default to dev
	}

	// Generate 24 random characters for the key suffix
	randomBytes := make([]byte, 18) // 18 bytes = 24 hex chars
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}

	// Convert to hex string
	randomHex := hex.EncodeToString(randomBytes)

	return prefix + randomHex, nil
}

// hashAPIKey creates a SHA-256 hash of the API key
func (s *apiKeyService) hashAPIKey(apiKey string) string {
	hash := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(hash[:])
}

// extractKeyPrefix extracts the display prefix from an API key
func (s *apiKeyService) extractKeyPrefix(apiKey string) string {
	if len(apiKey) < 8 {
		return apiKey
	}
	return apiKey[:8]
}

// isValidKeyFormat validates the API key format
func (s *apiKeyService) isValidKeyFormat(apiKey string) bool {
	// Check minimum length (prefix + random part)
	if len(apiKey) < 16 {
		return false
	}

	// Check if it starts with a valid prefix
	for _, prefix := range envPrefixMap {
		if strings.HasPrefix(apiKey, prefix) {
			return true
		}
	}

	return false
}

// detectEnvironmentFromKey detects environment type from key prefix
func (s *apiKeyService) detectEnvironmentFromKey(apiKey string) string {
	for env, prefix := range envPrefixMap {
		if strings.HasPrefix(apiKey, prefix) {
			return env
		}
	}
	return EnvDevelopment // Default
}
