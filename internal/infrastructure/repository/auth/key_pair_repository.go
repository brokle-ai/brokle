package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	authDomain "brokle/internal/core/domain/auth"
	"brokle/pkg/ulid"
)

// keyPairRepository implements authDomain.KeyPairRepository using GORM
type keyPairRepository struct {
	db *gorm.DB
}

// NewKeyPairRepository creates a new key pair repository instance
func NewKeyPairRepository(db *gorm.DB) authDomain.KeyPairRepository {
	return &keyPairRepository{
		db: db,
	}
}

// Create creates a new key pair
func (r *keyPairRepository) Create(ctx context.Context, keyPair *authDomain.KeyPair) error {
	return r.db.WithContext(ctx).Create(keyPair).Error
}

// GetByID retrieves a key pair by ID
func (r *keyPairRepository) GetByID(ctx context.Context, id ulid.ULID) (*authDomain.KeyPair, error) {
	var keyPair authDomain.KeyPair
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&keyPair).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("get key pair by ID %s: %w", id, authDomain.ErrNotFound)
		}
		return nil, fmt.Errorf("database error getting key pair by ID %s: %w", id, err)
	}
	return &keyPair, nil
}

// GetByPublicKey retrieves a key pair by public key
func (r *keyPairRepository) GetByPublicKey(ctx context.Context, publicKey string) (*authDomain.KeyPair, error) {
	var keyPair authDomain.KeyPair
	err := r.db.WithContext(ctx).Where("public_key = ? AND deleted_at IS NULL", publicKey).First(&keyPair).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("get key pair by public key: %w", authDomain.ErrNotFound)
		}
		return nil, fmt.Errorf("database error getting key pair by public key: %w", err)
	}
	return &keyPair, nil
}

// GetBySecretKeyHash retrieves a key pair by secret key hash
func (r *keyPairRepository) GetBySecretKeyHash(ctx context.Context, secretKeyHash string) (*authDomain.KeyPair, error) {
	var keyPair authDomain.KeyPair
	err := r.db.WithContext(ctx).Where("secret_key_hash = ? AND deleted_at IS NULL", secretKeyHash).First(&keyPair).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("get key pair by secret key hash: %w", authDomain.ErrNotFound)
		}
		return nil, fmt.Errorf("database error getting key pair by secret key hash: %w", err)
	}
	return &keyPair, nil
}

// Update updates an existing key pair
func (r *keyPairRepository) Update(ctx context.Context, keyPair *authDomain.KeyPair) error {
	keyPair.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(keyPair).Error
}

// Delete soft deletes a key pair
func (r *keyPairRepository) Delete(ctx context.Context, id ulid.ULID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&authDomain.KeyPair{}).Error
}

// ValidateKeyPair validates a public+secret key pair for authentication
func (r *keyPairRepository) ValidateKeyPair(ctx context.Context, publicKey, secretKey string) (*authDomain.KeyPair, error) {
	// Get key pair by public key
	keyPair, err := r.GetByPublicKey(ctx, publicKey)
	if err != nil {
		return nil, fmt.Errorf("invalid key pair: %w", err)
	}

	// Check if key pair is valid (active and not expired)
	if !keyPair.IsValid() {
		return nil, fmt.Errorf("key pair is inactive or expired")
	}

	// Validate secret key against stored hash
	err = bcrypt.CompareHashAndPassword([]byte(keyPair.SecretKeyHash), []byte(secretKey))
	if err != nil {
		return nil, fmt.Errorf("invalid key pair: secret key mismatch")
	}

	return keyPair, nil
}

// GetByUserID retrieves all key pairs for a user
func (r *keyPairRepository) GetByUserID(ctx context.Context, userID ulid.ULID) ([]*authDomain.KeyPair, error) {
	var keyPairs []*authDomain.KeyPair
	err := r.db.WithContext(ctx).Where("user_id = ? AND deleted_at IS NULL", userID).Find(&keyPairs).Error
	if err != nil {
		return nil, fmt.Errorf("database error getting key pairs by user ID %s: %w", userID, err)
	}
	return keyPairs, nil
}

// GetByOrganizationID retrieves all key pairs for an organization
func (r *keyPairRepository) GetByOrganizationID(ctx context.Context, orgID ulid.ULID) ([]*authDomain.KeyPair, error) {
	var keyPairs []*authDomain.KeyPair
	err := r.db.WithContext(ctx).Where("organization_id = ? AND deleted_at IS NULL", orgID).Find(&keyPairs).Error
	if err != nil {
		return nil, fmt.Errorf("database error getting key pairs by organization ID %s: %w", orgID, err)
	}
	return keyPairs, nil
}

// GetByProjectID retrieves all key pairs for a project
func (r *keyPairRepository) GetByProjectID(ctx context.Context, projectID ulid.ULID) ([]*authDomain.KeyPair, error) {
	var keyPairs []*authDomain.KeyPair
	err := r.db.WithContext(ctx).Where("project_id = ? AND deleted_at IS NULL", projectID).Find(&keyPairs).Error
	if err != nil {
		return nil, fmt.Errorf("database error getting key pairs by project ID %s: %w", projectID, err)
	}
	return keyPairs, nil
}

// GetByEnvironmentID retrieves all key pairs for an environment
func (r *keyPairRepository) GetByEnvironmentID(ctx context.Context, envID ulid.ULID) ([]*authDomain.KeyPair, error) {
	var keyPairs []*authDomain.KeyPair
	err := r.db.WithContext(ctx).Where("environment_id = ? AND deleted_at IS NULL", envID).Find(&keyPairs).Error
	if err != nil {
		return nil, fmt.Errorf("database error getting key pairs by environment ID %s: %w", envID, err)
	}
	return keyPairs, nil
}

// DeactivateKeyPair deactivates a key pair
func (r *keyPairRepository) DeactivateKeyPair(ctx context.Context, id ulid.ULID) error {
	return r.db.WithContext(ctx).Model(&authDomain.KeyPair{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_active":  false,
			"updated_at": time.Now(),
		}).Error
}

// MarkAsUsed updates the last used timestamp for a key pair
func (r *keyPairRepository) MarkAsUsed(ctx context.Context, id ulid.ULID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&authDomain.KeyPair{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_used_at": &now,
			"updated_at":   now,
		}).Error
}

// CleanupExpiredKeyPairs removes expired key pairs
func (r *keyPairRepository) CleanupExpiredKeyPairs(ctx context.Context) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Where("expires_at IS NOT NULL AND expires_at < ?", now).
		Delete(&authDomain.KeyPair{}).Error
}

// ValidatePublicKeyFormat validates the format of a public key
func (r *keyPairRepository) ValidatePublicKeyFormat(ctx context.Context, publicKey string) error {
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
func (r *keyPairRepository) ExtractProjectIDFromPublicKey(ctx context.Context, publicKey string) (ulid.ULID, error) {
	if err := r.ValidatePublicKeyFormat(ctx, publicKey); err != nil {
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

// CheckKeyPairScopes checks if a key pair has all required scopes
func (r *keyPairRepository) CheckKeyPairScopes(ctx context.Context, id ulid.ULID, requiredScopes []string) (bool, error) {
	keyPair, err := r.GetByID(ctx, id)
	if err != nil {
		return false, err
	}

	// Check if key pair has admin scope (grants all permissions)
	for _, scope := range keyPair.Scopes {
		if scope == string(authDomain.ScopeAdmin) {
			return true, nil
		}
	}

	// Check each required scope
	keyPairScopesMap := make(map[string]bool)
	for _, scope := range keyPair.Scopes {
		keyPairScopesMap[scope] = true
	}

	for _, requiredScope := range requiredScopes {
		if !keyPairScopesMap[requiredScope] {
			return false, nil
		}
	}

	return true, nil
}

// GetKeyPairCount returns the total number of key pairs for a user
func (r *keyPairRepository) GetKeyPairCount(ctx context.Context, userID ulid.ULID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&authDomain.KeyPair{}).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("database error counting key pairs for user %s: %w", userID, err)
	}
	return int(count), nil
}

// GetActiveKeyPairCount returns the number of active key pairs for a user
func (r *keyPairRepository) GetActiveKeyPairCount(ctx context.Context, userID ulid.ULID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&authDomain.KeyPair{}).
		Where("user_id = ? AND is_active = ? AND deleted_at IS NULL", userID, true).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("database error counting active key pairs for user %s: %w", userID, err)
	}
	return int(count), nil
}

// GetKeyPairCountByProject returns the number of key pairs for a project
func (r *keyPairRepository) GetKeyPairCountByProject(ctx context.Context, projectID ulid.ULID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&authDomain.KeyPair{}).
		Where("project_id = ? AND deleted_at IS NULL", projectID).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("database error counting key pairs for project %s: %w", projectID, err)
	}
	return int(count), nil
}