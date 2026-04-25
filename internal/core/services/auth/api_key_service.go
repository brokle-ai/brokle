// Package auth implements authentication: login, JWT and session tokens, API keys, roles, scopes, and RBAC.
package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"

	authDomain "brokle/internal/core/domain/auth"
	orgDomain "brokle/internal/core/domain/organization"
	appErrors "brokle/pkg/errors"
)

// apiKeyLastUsedDebounce is the minimum interval between persisted
// last_used writes per API key. Updating on every validation would
// generate massive write amplification for zero UX benefit — dashboards
// render "last used X ago" at minute granularity, so sub-5-minute
// precision is invisible to users. Debouncing here eliminates ~99% of
// writes for a hot key.
const apiKeyLastUsedDebounce = 5 * time.Minute

// APIKeyService manages SDK-facing API keys: creation, validation, rotation, and revocation.
type APIKeyService struct {
	apiKeyRepo             authDomain.APIKeyRepository
	organizationMemberRepo authDomain.OrganizationMemberRepository
	projectRepo            orgDomain.ProjectRepository
	logger                 *slog.Logger
}

// NewAPIKeyService creates a new API key service instance.
func NewAPIKeyService(
	apiKeyRepo authDomain.APIKeyRepository,
	organizationMemberRepo authDomain.OrganizationMemberRepository,
	projectRepo orgDomain.ProjectRepository,
	logger *slog.Logger,
) *APIKeyService {
	return &APIKeyService{
		apiKeyRepo:             apiKeyRepo,
		organizationMemberRepo: organizationMemberRepo,
		projectRepo:            projectRepo,
		logger:                 logger,
	}
}

// CreateAPIKey creates a new industry-standard API key with pure random secret
func (s *APIKeyService) CreateAPIKey(ctx context.Context, userID uuid.UUID, req *authDomain.CreateAPIKeyRequest) (*authDomain.CreateAPIKeyResponse, error) {
	// TODO: Validate user has permission to create keys in the project
	// For now, skip membership validation - will be implemented when organization service is ready

	// Generate industry-standard pure random API key (bk_{40_char_random})
	fullKey, err := authDomain.GenerateAPIKey()
	if err != nil {
		return nil, appErrors.NewInternalError("failed to generate API key", err)
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
		// The only UNIQUE constraint on api_keys is idx_api_keys_key_hash — a
		// collision here means the newly generated key's SHA-256 matched an
		// existing row (probability ~2^-160 under random generation; effectively
		// only reachable via a PRNG bug or pre-existing duplicate data). Name
		// uniqueness is NOT enforced, so telling callers "rename and retry" is
		// misleading and actionable on the wrong field.
		if errors.Is(err, authDomain.ErrAPIKeyAlreadyExists) {
			s.logger.Error("api key hash collision on create — regeneration required",
				"user_id", userID,
				"project_id", req.ProjectID,
			)
			return nil, appErrors.NewConflictError("failed to allocate a unique api key; please retry")
		}
		return nil, appErrors.NewInternalError("failed to save API key", err)
	}

	// Return response with the full key (only shown once)
	return &authDomain.CreateAPIKeyResponse{
		ID:         apiKeyEntity.ID,
		Name:       apiKeyEntity.Name,
		Key:        fullKey, // Full key - only returned once
		KeyPreview: apiKeyEntity.KeyPreview,
		ProjectID:  apiKeyEntity.ProjectID,
		CreatedAt:  apiKeyEntity.CreatedAt,
		ExpiresAt:  apiKeyEntity.ExpiresAt,
	}, nil
}

// GetAPIKey retrieves an API key by ID
func (s *APIKeyService) GetAPIKey(ctx context.Context, keyID uuid.UUID) (*authDomain.APIKey, error) {
	apiKey, err := s.apiKeyRepo.GetByID(ctx, keyID)
	if err != nil {
		if errors.Is(err, authDomain.ErrAPIKeyNotFound) {
			return nil, appErrors.NewNotFoundError("api key", appErrors.WithParam(keyID.String()))
		}
		return nil, appErrors.NewInternalError("failed to get api key", err)
	}
	return apiKey, nil
}

// GetAPIKeys retrieves API keys based on filters
func (s *APIKeyService) GetAPIKeys(ctx context.Context, filters *authDomain.APIKeyFilters) ([]*authDomain.APIKey, error) {
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

	// Use GetByFilters for comprehensive filtering with pagination
	return s.apiKeyRepo.GetByFilters(ctx, filters)
}

// CountAPIKeys returns the total count of API keys matching the filters
func (s *APIKeyService) CountAPIKeys(ctx context.Context, filters *authDomain.APIKeyFilters) (int64, error) {
	return s.apiKeyRepo.CountByFilters(ctx, filters)
}

// DeleteAPIKey deletes (soft deletes) an API key with project ownership verification
func (s *APIKeyService) DeleteAPIKey(ctx context.Context, keyID uuid.UUID, projectID uuid.UUID) error {
	// Verify API key exists (filters out already-deleted keys)
	apiKey, err := s.apiKeyRepo.GetByID(ctx, keyID)
	if err != nil {
		if errors.Is(err, authDomain.ErrAPIKeyNotFound) {
			return appErrors.NewNotFoundError("api key", appErrors.WithParam(keyID.String()))
		}
		return appErrors.NewInternalError("failed to get api key", err)
	}

	// Verify API key belongs to specified project (security check)
	if apiKey.ProjectID != projectID {
		return appErrors.NewNotFoundError("API key not found in this project")
	}

	// Perform soft delete
	if err := s.apiKeyRepo.Delete(ctx, keyID); err != nil {
		return appErrors.NewInternalError("failed to delete API key", err)
	}

	return nil
}

// ValidateAPIKey validates an industry-standard API key using direct SHA-256 hash lookup
// This is O(1) with unique index on key_hash column (GitHub/Stripe pattern)
func (s *APIKeyService) ValidateAPIKey(ctx context.Context, fullKey string) (*authDomain.ValidateAPIKeyResponse, error) {
	// Validate API key format (bk_{40_chars})
	if err := authDomain.ValidateAPIKeyFormat(fullKey); err != nil {
		return nil, appErrors.NewUnauthorizedError("invalid API key format")
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
		if errors.Is(err, authDomain.ErrAPIKeyNotFound) {
			// Don't expose whether key exists or not (security best practice)
			return nil, appErrors.NewUnauthorizedError("invalid API key")
		}
		// Infrastructure error (DB connection, migration issue, etc.) - return 500
		return nil, appErrors.NewInternalError("failed to validate API key", err)
	}

	// Check expiration (deleted keys filtered by GORM soft delete)
	if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
		return nil, appErrors.NewUnauthorizedError("API key has expired")
	}

	// Create auth context
	authContext := &authDomain.AuthContext{
		UserID:   apiKey.UserID,
		APIKeyID: &apiKey.ID,
	}

	// Update last-used timestamp asynchronously. Debounce on the cached
	// LastUsedAt — skip the write if it was persisted within the debounce
	// window. Uses a detached context with a short timeout so a slow DB
	// write cannot leak the goroutine; failures are logged best-effort and
	// never propagated (validation already succeeded).
	if apiKey.LastUsedAt == nil || time.Since(*apiKey.LastUsedAt) > apiKeyLastUsedDebounce {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := s.apiKeyRepo.UpdateLastUsed(ctx, apiKey.ID); err != nil {
				s.logger.Warn("failed to update api key last_used timestamp",
					"api_key_id", apiKey.ID,
					"error", err,
				)
			}
		}()
	}

	// Look up project to get OrganizationID for billing aggregation
	project, err := s.projectRepo.GetByID(ctx, apiKey.ProjectID)
	if err != nil {
		// Project lookup failure is critical - shouldn't happen for valid API keys
		return nil, appErrors.NewInternalError("failed to look up project for API key", err)
	}

	// Return validation response with project_id and organization_id from database
	return &authDomain.ValidateAPIKeyResponse{
		APIKey:         apiKey,
		ProjectID:      apiKey.ProjectID,       // Retrieved from database, not extracted from key
		OrganizationID: project.OrganizationID, // From project lookup for billing aggregation
		Valid:          true,
		AuthContext:    authContext,
	}, nil
}

// CheckRateLimit checks if the API key has exceeded rate limits
func (s *APIKeyService) CheckRateLimit(ctx context.Context, keyID uuid.UUID) (bool, error) {
	// TODO: Implement rate limiting logic with Redis
	// For now, always allow requests
	return true, nil
}

// GetAPIKeyContext creates an AuthContext from an API key
func (s *APIKeyService) GetAPIKeyContext(ctx context.Context, keyID uuid.UUID) (*authDomain.AuthContext, error) {
	apiKey, err := s.apiKeyRepo.GetByID(ctx, keyID)
	if err != nil {
		if errors.Is(err, authDomain.ErrAPIKeyNotFound) {
			return nil, appErrors.NewNotFoundError("api key", appErrors.WithParam(keyID.String()))
		}
		return nil, appErrors.NewInternalError("failed to get api key", err)
	}

	return &authDomain.AuthContext{
		UserID:   apiKey.UserID,
		APIKeyID: &apiKey.ID,
	}, nil
}

// CanAPIKeyAccessResource checks if an API key can access a specific resource
// Note: All non-deleted, non-expired API keys have full access to their project
// Access control should be handled at the organization RBAC level
func (s *APIKeyService) CanAPIKeyAccessResource(ctx context.Context, keyID uuid.UUID, resource string) (bool, error) {
	apiKey, err := s.apiKeyRepo.GetByID(ctx, keyID)
	if err != nil {
		if errors.Is(err, authDomain.ErrAPIKeyNotFound) {
			return false, appErrors.NewNotFoundError("api key", appErrors.WithParam(keyID.String()))
		}
		return false, appErrors.NewInternalError("failed to get api key", err)
	}

	// All API keys have full access to their project (deleted keys filtered by GORM)
	// Fine-grained permissions handled by organization RBAC
	return !apiKey.IsExpired(), nil
}

// Scoped access methods
func (s *APIKeyService) GetAPIKeysByUser(ctx context.Context, userID uuid.UUID) ([]*authDomain.APIKey, error) {
	return s.apiKeyRepo.GetByUserID(ctx, userID)
}

func (s *APIKeyService) GetAPIKeysByOrganization(ctx context.Context, orgID uuid.UUID) ([]*authDomain.APIKey, error) {
	return s.apiKeyRepo.GetByOrganizationID(ctx, orgID)
}

func (s *APIKeyService) GetAPIKeysByProject(ctx context.Context, projectID uuid.UUID) ([]*authDomain.APIKey, error) {
	return s.apiKeyRepo.GetByProjectID(ctx, projectID)
}
