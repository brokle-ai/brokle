package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
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

// CreateAPIKey creates a new industry-standard API key with pure random secret
func (s *apiKeyService) CreateAPIKey(ctx context.Context, userID ulid.ULID, req *authDomain.CreateAPIKeyRequest) (*authDomain.CreateAPIKeyResponse, error) {
	// TODO: Validate user has permission to create keys in the project
	// For now, skip membership validation - will be implemented when organization service is ready

	// Generate industry-standard pure random API key (bk_{40_char_random})
	fullKey, err := authDomain.GenerateAPIKey()
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to generate API key", err)
	}

	// Hash the full key for secure storage using SHA-256 (industry standard for API keys)
	// Note: SHA-256 is deterministic (same input = same output), enabling O(1) lookup
	// This is different from bcrypt (used for passwords) which is non-deterministic
	hash := sha256.Sum256([]byte(fullKey))
	keyHash := hex.EncodeToString(hash[:])

	// Create key preview for display (bk_...xyz)
	keyPreview := authDomain.CreateKeyPreview(fullKey)

	// Create API key entity (project_id stored in database, not in key)
	apiKeyEntity := authDomain.NewAPIKey(
		userID,
		req.ProjectID,
		req.Name,
		keyHash, // SHA-256 hash of full key (deterministic, enables O(1) lookup)
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

// ValidateAPIKey validates an industry-standard API key using direct SHA-256 hash lookup
// This is O(1) with unique index on key_hash column (GitHub/Stripe pattern)
func (s *apiKeyService) ValidateAPIKey(ctx context.Context, fullKey string) (*authDomain.ValidateAPIKeyResponse, error) {
	// Validate API key format (bk_{40_chars})
	if err := authDomain.ValidateAPIKeyFormat(fullKey); err != nil {
		return nil, appErrors.NewUnauthorizedError("Invalid API key format")
	}

	// Hash the incoming key using SHA-256 for O(1) lookup
	// SHA-256 is deterministic (same input = same hash), enabling direct database lookup
	// This is the industry standard for API keys (GitHub, Stripe, OpenAI all use this)
	hash := sha256.Sum256([]byte(fullKey))
	keyHash := hex.EncodeToString(hash[:])

	// Direct lookup by hash (O(1) with unique index on key_hash)
	apiKey, err := s.apiKeyRepo.GetByKeyHash(ctx, keyHash)
	if err != nil {
		// Distinguish between not-found (401) and infrastructure errors (500)
		if errors.Is(err, authDomain.ErrNotFound) {
			// Don't expose whether key exists or not (security best practice)
			return nil, appErrors.NewUnauthorizedError("Invalid API key")
		}
		// Infrastructure error (DB connection, migration issue, etc.) - return 500
		return nil, appErrors.NewInternalError("Failed to validate API key", err)
	}

	// Check if key is active
	if !apiKey.IsActive {
		return nil, appErrors.NewUnauthorizedError("API key is inactive")
	}

	// Check expiration
	if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
		return nil, appErrors.NewUnauthorizedError("API key has expired")
	}

	// Create auth context
	authContext := &authDomain.AuthContext{
		UserID:   apiKey.UserID,
		APIKeyID: &apiKey.ID,
	}

	// Update last used timestamp (async, don't block validation)
	go func() {
		ctx := context.Background()
		if err := s.apiKeyRepo.UpdateLastUsed(ctx, apiKey.ID); err != nil {
			// Log error but don't fail validation
			// TODO: Add proper logging when logger is available
		}
	}()

	// Return validation response with project_id from database
	return &authDomain.ValidateAPIKeyResponse{
		APIKey:      apiKey,
		ProjectID:   apiKey.ProjectID, // Retrieved from database, not extracted from key
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

