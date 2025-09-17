package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	authDomain "brokle/internal/core/domain/auth"
	userDomain "brokle/internal/core/domain/user"
	"brokle/pkg/ulid"
	appErrors "brokle/pkg/errors"
)

// keyPairService implements the authDomain.KeyPairService interface
type keyPairService struct {
	keyPairRepo authDomain.KeyPairRepository
	userRepo    userDomain.Repository
}

// NewKeyPairService creates a new key pair service instance
func NewKeyPairService(
	keyPairRepo authDomain.KeyPairRepository,
	userRepo userDomain.Repository,
) authDomain.KeyPairService {
	return &keyPairService{
		keyPairRepo: keyPairRepo,
		userRepo:    userRepo,
	}
}

// CreateKeyPair creates a new key pair for the specified user and project
func (s *keyPairService) CreateKeyPair(ctx context.Context, userID ulid.ULID, req *authDomain.CreateKeyPairRequest) (*authDomain.CreateKeyPairResponse, error) {
	// Validate user exists
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == userDomain.ErrNotFound {
			return nil, appErrors.NewNotFoundError("User not found")
		}
		return nil, appErrors.NewInternalError("Failed to validate user", err)
	}

	// Generate public and secret keys
	publicKey, secretKey, err := s.GenerateKeyPair(ctx, req.ProjectID)
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to generate key pair", err)
	}

	// Hash the secret key for storage
	secretKeyHash, err := s.hashSecretKey(secretKey)
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to hash secret key", err)
	}

	// Create the key pair entity
	keyPair := authDomain.NewKeyPair(
		userID,
		req.OrganizationID,
		req.ProjectID,
		req.Name,
		publicKey,
		secretKeyHash,
		req.Scopes,
		req.RateLimitRPM,
		req.ExpiresAt,
	)

	// Validate the key pair
	if err := keyPair.ValidatePublicKeyFormat(); err != nil {
		return nil, appErrors.NewBadRequestError("Invalid public key format", err.Error())
	}

	if err := keyPair.ValidateSecretKeyPrefix(); err != nil {
		return nil, appErrors.NewBadRequestError("Invalid secret key prefix", err.Error())
	}

	// Save to database
	if err := s.keyPairRepo.Create(ctx, keyPair); err != nil {
		return nil, appErrors.NewInternalError("Failed to create key pair", err)
	}

	// Return response with secret key (only shown once)
	return &authDomain.CreateKeyPairResponse{
		ID:        keyPair.ID,
		Name:      keyPair.Name,
		PublicKey: keyPair.PublicKey,
		SecretKey: secretKey, // Only returned on creation
		ProjectID: keyPair.ProjectID,
		Scopes:    keyPair.Scopes,
		ExpiresAt: keyPair.ExpiresAt,
	}, nil
}

// GetKeyPair retrieves a key pair by ID
func (s *keyPairService) GetKeyPair(ctx context.Context, keyID ulid.ULID) (*authDomain.KeyPair, error) {
	keyPair, err := s.keyPairRepo.GetByID(ctx, keyID)
	if err != nil {
		return nil, appErrors.NewNotFoundError("Key pair not found")
	}
	return keyPair, nil
}

// GetKeyPairs retrieves key pairs based on filters
func (s *keyPairService) GetKeyPairs(ctx context.Context, filters *authDomain.KeyPairFilters) ([]*authDomain.KeyPair, error) {
	// TODO: Implement repository method for filtered queries
	// For now, return by user ID if specified
	if filters.UserID != nil {
		return s.keyPairRepo.GetByUserID(ctx, *filters.UserID)
	}
	if filters.OrganizationID != nil {
		return s.keyPairRepo.GetByOrganizationID(ctx, *filters.OrganizationID)
	}
	if filters.ProjectID != nil {
		return s.keyPairRepo.GetByProjectID(ctx, *filters.ProjectID)
	}

	return nil, appErrors.NewBadRequestError("At least one filter must be specified", "Provide userID, organizationID, or projectID")
}

// UpdateKeyPair updates an existing key pair
func (s *keyPairService) UpdateKeyPair(ctx context.Context, keyID ulid.ULID, req *authDomain.UpdateKeyPairRequest) error {
	keyPair, err := s.keyPairRepo.GetByID(ctx, keyID)
	if err != nil {
		return appErrors.NewNotFoundError("Key pair not found")
	}

	// Update fields if provided
	if req.Name != nil {
		keyPair.Name = *req.Name
	}
	if req.Scopes != nil {
		keyPair.Scopes = req.Scopes
	}
	if req.RateLimitRPM != nil {
		keyPair.RateLimitRPM = *req.RateLimitRPM
	}
	if req.ExpiresAt != nil {
		keyPair.ExpiresAt = req.ExpiresAt
	}
	if req.IsActive != nil {
		keyPair.IsActive = *req.IsActive
	}

	keyPair.UpdatedAt = time.Now()

	return s.keyPairRepo.Update(ctx, keyPair)
}

// RevokeKeyPair deactivates a key pair
func (s *keyPairService) RevokeKeyPair(ctx context.Context, keyID ulid.ULID) error {
	return s.keyPairRepo.DeactivateKeyPair(ctx, keyID)
}

// ValidateKeyPair validates a public+secret key pair for authentication
func (s *keyPairService) ValidateKeyPair(ctx context.Context, publicKey, secretKey string) (*authDomain.KeyPair, error) {
	// Get key pair by public key
	keyPair, err := s.keyPairRepo.GetByPublicKey(ctx, publicKey)
	if err != nil {
		return nil, appErrors.NewUnauthorizedError("Invalid key pair")
	}

	// Check if key pair is valid (active and not expired)
	if !keyPair.IsValid() {
		return nil, appErrors.NewUnauthorizedError("Key pair is inactive or expired")
	}

	// Validate secret key
	if !s.validateSecretKey(secretKey, keyPair.SecretKeyHash) {
		return nil, appErrors.NewUnauthorizedError("Invalid key pair")
	}

	// Update last used timestamp
	if err := s.keyPairRepo.MarkAsUsed(ctx, keyPair.ID); err != nil {
		// Log error but don't fail authentication
		// TODO: Add proper logging
	}

	return keyPair, nil
}

// AuthenticateWithKeyPair authenticates using public+secret key pair and returns auth context
func (s *keyPairService) AuthenticateWithKeyPair(ctx context.Context, publicKey, secretKey string) (*authDomain.AuthContext, error) {
	keyPair, err := s.ValidateKeyPair(ctx, publicKey, secretKey)
	if err != nil {
		return nil, err
	}

	// Create authentication context
	authCtx := &authDomain.AuthContext{
		UserID:         keyPair.UserID,
		KeyPairID:      &keyPair.ID,
		OrganizationID: &keyPair.OrganizationID,
		ProjectID:      &keyPair.ProjectID,
		EnvironmentID:  keyPair.EnvironmentID,
		Scopes:         keyPair.Scopes,
	}

	return authCtx, nil
}

// UpdateLastUsed updates the last used timestamp for a key pair
func (s *keyPairService) UpdateLastUsed(ctx context.Context, keyID ulid.ULID) error {
	return s.keyPairRepo.MarkAsUsed(ctx, keyID)
}

// CheckRateLimit checks if the key pair has exceeded its rate limit
func (s *keyPairService) CheckRateLimit(ctx context.Context, keyID ulid.ULID) (bool, error) {
	// TODO: Implement rate limiting logic
	// This would typically involve checking request counts in Redis or similar
	// For now, return true (allowed)
	return true, nil
}

// GetKeyPairContext retrieves the authentication context for a key pair
func (s *keyPairService) GetKeyPairContext(ctx context.Context, keyID ulid.ULID) (*authDomain.AuthContext, error) {
	keyPair, err := s.keyPairRepo.GetByID(ctx, keyID)
	if err != nil {
		return nil, appErrors.NewNotFoundError("Key pair not found")
	}

	return &authDomain.AuthContext{
		UserID:         keyPair.UserID,
		KeyPairID:      &keyPair.ID,
		OrganizationID: &keyPair.OrganizationID,
		ProjectID:      &keyPair.ProjectID,
		EnvironmentID:  keyPair.EnvironmentID,
		Scopes:         keyPair.Scopes,
	}, nil
}

// CanKeyPairAccessResource checks if a key pair can access a specific resource
func (s *keyPairService) CanKeyPairAccessResource(ctx context.Context, keyID ulid.ULID, resource string) (bool, error) {
	keyPair, err := s.keyPairRepo.GetByID(ctx, keyID)
	if err != nil {
		return false, appErrors.NewNotFoundError("Key pair not found")
	}

	// Check if key pair has admin scope (grants all access)
	if keyPair.HasScope(authDomain.ScopeAdmin) {
		return true, nil
	}

	// Check resource-specific scopes
	switch {
	case strings.HasPrefix(resource, "gateway"):
		return keyPair.HasScope(authDomain.ScopeGatewayRead) || keyPair.HasScope(authDomain.ScopeGatewayWrite), nil
	case strings.HasPrefix(resource, "analytics"):
		return keyPair.HasScope(authDomain.ScopeAnalyticsRead), nil
	case strings.HasPrefix(resource, "config"):
		return keyPair.HasScope(authDomain.ScopeConfigRead) || keyPair.HasScope(authDomain.ScopeConfigWrite), nil
	default:
		return false, nil
	}
}

// CheckKeyPairScopes checks if a key pair has all required scopes
func (s *keyPairService) CheckKeyPairScopes(ctx context.Context, keyID ulid.ULID, requiredScopes []string) (bool, error) {
	keyPair, err := s.keyPairRepo.GetByID(ctx, keyID)
	if err != nil {
		return false, appErrors.NewNotFoundError("Key pair not found")
	}

	// Check if key pair has admin scope (grants all permissions)
	if keyPair.HasScope(authDomain.ScopeAdmin) {
		return true, nil
	}

	// Check each required scope
	for _, requiredScope := range requiredScopes {
		scope := authDomain.KeyPairScope(requiredScope)
		if !keyPair.HasScope(scope) {
			return false, nil
		}
	}

	return true, nil
}

// GetKeyPairsByUser retrieves all key pairs for a user
func (s *keyPairService) GetKeyPairsByUser(ctx context.Context, userID ulid.ULID) ([]*authDomain.KeyPair, error) {
	return s.keyPairRepo.GetByUserID(ctx, userID)
}

// GetKeyPairsByOrganization retrieves all key pairs for an organization
func (s *keyPairService) GetKeyPairsByOrganization(ctx context.Context, orgID ulid.ULID) ([]*authDomain.KeyPair, error) {
	return s.keyPairRepo.GetByOrganizationID(ctx, orgID)
}

// GetKeyPairsByProject retrieves all key pairs for a project
func (s *keyPairService) GetKeyPairsByProject(ctx context.Context, projectID ulid.ULID) ([]*authDomain.KeyPair, error) {
	return s.keyPairRepo.GetByProjectID(ctx, projectID)
}

// GetKeyPairsByEnvironment retrieves all key pairs for an environment
func (s *keyPairService) GetKeyPairsByEnvironment(ctx context.Context, envID ulid.ULID) ([]*authDomain.KeyPair, error) {
	return s.keyPairRepo.GetByEnvironmentID(ctx, envID)
}

// GenerateKeyPair generates a new public+secret key pair for a project
func (s *keyPairService) GenerateKeyPair(ctx context.Context, projectID ulid.ULID) (publicKey, secretKey string, err error) {
	// Generate random suffix for keys
	randomBytes := make([]byte, 16) // 32 hex characters
	if _, err := rand.Read(randomBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	randomSuffix := hex.EncodeToString(randomBytes)

	// Generate public key: pk_projectId_random
	publicKey = fmt.Sprintf("pk_%s_%s", projectID.String(), randomSuffix)

	// Generate secret key: sk_random (different random suffix)
	secretRandomBytes := make([]byte, 20) // 40 hex characters
	if _, err := rand.Read(secretRandomBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate secret random bytes: %w", err)
	}
	secretRandomSuffix := hex.EncodeToString(secretRandomBytes)
	secretKey = fmt.Sprintf("sk_%s", secretRandomSuffix)

	return publicKey, secretKey, nil
}

// ValidatePublicKeyFormat validates the format of a public key
func (s *keyPairService) ValidatePublicKeyFormat(ctx context.Context, publicKey string) error {
	if !strings.HasPrefix(publicKey, "pk_") {
		return fmt.Errorf("public key must start with 'pk_', got: %s", publicKey)
	}

	parts := strings.Split(publicKey, "_")
	if len(parts) < 3 {
		return fmt.Errorf("public key must be in format pk_projectId_random, got: %s", publicKey)
	}

	projectIDPart := parts[1]
	if len(projectIDPart) != 26 {
		return fmt.Errorf("project ID in public key must be 26 characters (ULID), got: %d characters", len(projectIDPart))
	}

	// Validate ULID format
	if _, err := ulid.Parse(projectIDPart); err != nil {
		return fmt.Errorf("invalid project ID format in public key: %w", err)
	}

	return nil
}

// ExtractProjectIDFromPublicKey extracts the project ID from a public key
func (s *keyPairService) ExtractProjectIDFromPublicKey(ctx context.Context, publicKey string) (ulid.ULID, error) {
	if err := s.ValidatePublicKeyFormat(ctx, publicKey); err != nil {
		return ulid.ULID{}, err
	}

	parts := strings.Split(publicKey, "_")
	projectIDStr := parts[1]

	projectID, err := ulid.Parse(projectIDStr)
	if err != nil {
		return ulid.ULID{}, fmt.Errorf("invalid project ID in public key: %w", err)
	}

	return projectID, nil
}

// hashSecretKey hashes a secret key using bcrypt for secure storage
func (s *keyPairService) hashSecretKey(secretKey string) (string, error) {
	// Use bcrypt for consistency with password hashing
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(secretKey), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash secret key: %w", err)
	}
	return string(hashedBytes), nil
}

// validateSecretKey validates a secret key against its stored hash
func (s *keyPairService) validateSecretKey(secretKey, secretKeyHash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(secretKeyHash), []byte(secretKey))
	return err == nil
}