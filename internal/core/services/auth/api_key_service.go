package auth

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

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

// CreateAPIKey creates a new project-scoped API key
func (s *apiKeyService) CreateAPIKey(ctx context.Context, userID ulid.ULID, req *authDomain.CreateAPIKeyRequest) (*authDomain.CreateAPIKeyResponse, error) {
	// TODO: Validate user has permission to create keys in the project
	// For now, skip membership validation - will be implemented when organization service is ready

	// Generate project-scoped API key
	fullKey, _, _, err := authDomain.GenerateProjectScopedAPIKey(req.ProjectID)
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to generate API key", err)
	}

	// Hash the FULL key (not just secret) for secure storage
	keyHash, err := bcrypt.GenerateFromPassword([]byte(fullKey), bcrypt.DefaultCost)
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to hash API key", err)
	}

	// Create key preview for display
	keyPreview := authDomain.CreateKeyPreview(fullKey)

	// Create API key entity
	apiKeyEntity := authDomain.NewAPIKey(
		userID,
		req.ProjectID,
		req.Name,
		string(keyHash), // Hash of full key
		keyPreview,
		req.ExpiresAt,
	)

	// Save to database
	if err := s.apiKeyRepo.Create(ctx, apiKeyEntity); err != nil {
		return nil, appErrors.NewInternalError("Failed to save API key", err)
	}

	// Return response with the full key (only shown once)
	return &authDomain.CreateAPIKeyResponse{
		ID:         apiKeyEntity.ID.String(),
		Name:       apiKeyEntity.Name,
		Key:        fullKey, // Full key - only returned once
		KeyPreview: apiKeyEntity.KeyPreview,
		ProjectID:  apiKeyEntity.ProjectID.String(),
		CreatedAt:  apiKeyEntity.CreatedAt,
		ExpiresAt:  apiKeyEntity.ExpiresAt,
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

// ValidateAPIKey validates a project-scoped API key by comparing hash
func (s *apiKeyService) ValidateAPIKey(ctx context.Context, fullKey string) (*authDomain.ValidateAPIKeyResponse, error) {
	// Extract project ID from the key format
	projectID, err := authDomain.ExtractProjectIDFromFullKey(fullKey)
	if err != nil {
		return nil, appErrors.NewUnauthorizedError("Invalid API key format")
	}

	// Get all API keys for this project
	apiKeys, err := s.apiKeyRepo.GetByProjectID(ctx, projectID)
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to validate API key", err)
	}

	// Find the matching key by comparing hashes
	var matchedKey *authDomain.APIKey
	for _, key := range apiKeys {
		if bcrypt.CompareHashAndPassword([]byte(key.KeyHash), []byte(fullKey)) == nil {
			matchedKey = key
			break
		}
	}

	if matchedKey == nil {
		return nil, appErrors.NewUnauthorizedError("Invalid API key")
	}

	// Check if key is active
	if !matchedKey.IsActive {
		return nil, appErrors.NewUnauthorizedError("API key is inactive")
	}

	// Check expiration
	if matchedKey.ExpiresAt != nil && time.Now().After(*matchedKey.ExpiresAt) {
		return nil, appErrors.NewUnauthorizedError("API key has expired")
	}

	// Create auth context
	authContext := &authDomain.AuthContext{
		UserID:   matchedKey.UserID,
		APIKeyID: &matchedKey.ID,
	}

	// Update last used timestamp (async)
	go func() {
		ctx := context.Background()
		if err := s.apiKeyRepo.UpdateLastUsed(ctx, matchedKey.ID); err != nil {
			// Log error but don't fail validation
		}
	}()

	return &authDomain.ValidateAPIKeyResponse{
		APIKey:      matchedKey,
		ProjectID:   projectID,
		Valid:       true,
		AuthContext: authContext,
	}, nil
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
// Note: With scopes removed, this now checks if the key is active
// Access control should be handled at the organization RBAC level
func (s *apiKeyService) CanAPIKeyAccessResource(ctx context.Context, keyID ulid.ULID, resource string) (bool, error) {
	apiKey, err := s.apiKeyRepo.GetByID(ctx, keyID)
	if err != nil {
		return false, fmt.Errorf("get API key: %w", err)
	}

	// All active API keys have full access to their project
	// Fine-grained permissions handled by organization RBAC
	return apiKey.IsActive, nil
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

