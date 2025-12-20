package credentials

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	credentialsDomain "brokle/internal/core/domain/credentials"
	"brokle/pkg/ulid"
)

type providerCredentialRepository struct {
	db *gorm.DB
}

func NewProviderCredentialRepository(db *gorm.DB) credentialsDomain.ProviderCredentialRepository {
	return &providerCredentialRepository{
		db: db,
	}
}

func (r *providerCredentialRepository) Create(ctx context.Context, credential *credentialsDomain.ProviderCredential) error {
	err := r.db.WithContext(ctx).Create(credential).Error
	if err != nil {
		// Check for unique constraint violation
		if isUniqueViolation(err) {
			return fmt.Errorf("create credential: %w", credentialsDomain.ErrCredentialExists)
		}
		return fmt.Errorf("create credential: %w", err)
	}
	return nil
}

// GetByID retrieves a credential by its ID within a specific project.
// Returns ErrCredentialNotFound if not found or belongs to different project.
func (r *providerCredentialRepository) GetByID(ctx context.Context, id ulid.ULID, projectID ulid.ULID) (*credentialsDomain.ProviderCredential, error) {
	var credential credentialsDomain.ProviderCredential
	err := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ?", id, projectID).
		First(&credential).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("get credential by ID %s: %w", id, credentialsDomain.ErrCredentialNotFound)
		}
		return nil, fmt.Errorf("get credential by ID: %w", err)
	}
	return &credential, nil
}

// GetByProjectAndName retrieves the credential for a specific project and name.
// Returns nil if not found.
func (r *providerCredentialRepository) GetByProjectAndName(ctx context.Context, projectID ulid.ULID, name string) (*credentialsDomain.ProviderCredential, error) {
	var credential credentialsDomain.ProviderCredential
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND name = ?", projectID, name).
		First(&credential).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil for uniqueness checks (not found = available)
		}
		return nil, fmt.Errorf("get credential by name: %w", err)
	}
	return &credential, nil
}

// GetByProjectAndAdapter retrieves all credentials for a specific project and adapter type.
// Returns empty slice if none found.
func (r *providerCredentialRepository) GetByProjectAndAdapter(ctx context.Context, projectID ulid.ULID, adapter credentialsDomain.Provider) ([]*credentialsDomain.ProviderCredential, error) {
	var credentials []*credentialsDomain.ProviderCredential
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND adapter = ?", projectID, adapter).
		Order("created_at DESC").
		Find(&credentials).Error
	if err != nil {
		return nil, fmt.Errorf("get credentials by adapter: %w", err)
	}
	return credentials, nil
}

func (r *providerCredentialRepository) ListByProject(ctx context.Context, projectID ulid.ULID) ([]*credentialsDomain.ProviderCredential, error) {
	var credentials []*credentialsDomain.ProviderCredential
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("created_at DESC").
		Find(&credentials).Error
	if err != nil {
		return nil, fmt.Errorf("list credentials for project %s: %w", projectID, err)
	}
	return credentials, nil
}

// Update updates an existing credential within a specific project.
// Returns ErrCredentialNotFound if not found or belongs to different project.
func (r *providerCredentialRepository) Update(ctx context.Context, credential *credentialsDomain.ProviderCredential, projectID ulid.ULID) error {
	credential.UpdatedAt = time.Now()

	// Use project-scoped update to prevent cross-project modification
	result := r.db.WithContext(ctx).
		Model(&credentialsDomain.ProviderCredential{}).
		Where("id = ? AND project_id = ?", credential.ID, projectID).
		Updates(map[string]interface{}{
			"name":          credential.Name,
			"encrypted_key": credential.EncryptedKey,
			"key_preview":   credential.KeyPreview,
			"base_url":      credential.BaseURL,
			"config":        credential.Config,
			"custom_models": credential.CustomModels,
			"headers":       credential.Headers,
			"updated_at":    credential.UpdatedAt,
		})

	if result.Error != nil {
		// Check for unique constraint violation (name conflict)
		if isUniqueViolation(result.Error) {
			return fmt.Errorf("update credential: %w", credentialsDomain.ErrCredentialExists)
		}
		return fmt.Errorf("update credential: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("update credential: %w", credentialsDomain.ErrCredentialNotFound)
	}
	return nil
}

// Delete removes a credential by ID within a specific project.
// Returns ErrCredentialNotFound if not found or belongs to different project.
func (r *providerCredentialRepository) Delete(ctx context.Context, id ulid.ULID, projectID ulid.ULID) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ?", id, projectID).
		Delete(&credentialsDomain.ProviderCredential{})
	if result.Error != nil {
		return fmt.Errorf("delete credential: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("delete credential %s: %w", id, credentialsDomain.ErrCredentialNotFound)
	}
	return nil
}

func (r *providerCredentialRepository) ExistsByProjectAndName(ctx context.Context, projectID ulid.ULID, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&credentialsDomain.ProviderCredential{}).
		Where("project_id = ? AND name = ?", projectID, name).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("check credential exists: %w", err)
	}
	return count > 0, nil
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	// PostgreSQL unique violation error code: 23505
	errStr := err.Error()
	return strings.Contains(errStr, "23505") || strings.Contains(errStr, "unique constraint") || strings.Contains(errStr, "duplicate key")
}
