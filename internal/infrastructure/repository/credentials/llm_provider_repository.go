// Package credentials provides repository implementations for credential storage.
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

// llmProviderCredentialRepository implements credentialsDomain.LLMProviderCredentialRepository using GORM
type llmProviderCredentialRepository struct {
	db *gorm.DB
}

// NewLLMProviderCredentialRepository creates a new repository instance
func NewLLMProviderCredentialRepository(db *gorm.DB) credentialsDomain.LLMProviderCredentialRepository {
	return &llmProviderCredentialRepository{
		db: db,
	}
}

// Create creates a new LLM provider credential.
func (r *llmProviderCredentialRepository) Create(ctx context.Context, credential *credentialsDomain.LLMProviderCredential) error {
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

// GetByID retrieves a credential by its ID.
func (r *llmProviderCredentialRepository) GetByID(ctx context.Context, id ulid.ULID) (*credentialsDomain.LLMProviderCredential, error) {
	var credential credentialsDomain.LLMProviderCredential
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&credential).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("get credential by ID %s: %w", id, credentialsDomain.ErrCredentialNotFound)
		}
		return nil, fmt.Errorf("get credential by ID: %w", err)
	}
	return &credential, nil
}

// GetByProjectAndProvider retrieves the credential for a specific project and provider.
func (r *llmProviderCredentialRepository) GetByProjectAndProvider(ctx context.Context, projectID ulid.ULID, provider credentialsDomain.LLMProvider) (*credentialsDomain.LLMProviderCredential, error) {
	var credential credentialsDomain.LLMProviderCredential
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND provider = ?", projectID, provider).
		First(&credential).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("get credential for project %s provider %s: %w", projectID, provider, credentialsDomain.ErrCredentialNotFound)
		}
		return nil, fmt.Errorf("get credential by project and provider: %w", err)
	}
	return &credential, nil
}

// ListByProject retrieves all credentials for a project.
func (r *llmProviderCredentialRepository) ListByProject(ctx context.Context, projectID ulid.ULID) ([]*credentialsDomain.LLMProviderCredential, error) {
	var credentials []*credentialsDomain.LLMProviderCredential
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("created_at DESC").
		Find(&credentials).Error
	if err != nil {
		return nil, fmt.Errorf("list credentials for project %s: %w", projectID, err)
	}
	return credentials, nil
}

// Update updates an existing credential.
func (r *llmProviderCredentialRepository) Update(ctx context.Context, credential *credentialsDomain.LLMProviderCredential) error {
	credential.UpdatedAt = time.Now()
	result := r.db.WithContext(ctx).Save(credential)
	if result.Error != nil {
		return fmt.Errorf("update credential: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("update credential: %w", credentialsDomain.ErrCredentialNotFound)
	}
	return nil
}

// Delete removes a credential by ID.
func (r *llmProviderCredentialRepository) Delete(ctx context.Context, id ulid.ULID) error {
	result := r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&credentialsDomain.LLMProviderCredential{})
	if result.Error != nil {
		return fmt.Errorf("delete credential: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("delete credential %s: %w", id, credentialsDomain.ErrCredentialNotFound)
	}
	return nil
}

// DeleteByProjectAndProvider removes a credential by project and provider.
func (r *llmProviderCredentialRepository) DeleteByProjectAndProvider(ctx context.Context, projectID ulid.ULID, provider credentialsDomain.LLMProvider) error {
	result := r.db.WithContext(ctx).
		Where("project_id = ? AND provider = ?", projectID, provider).
		Delete(&credentialsDomain.LLMProviderCredential{})
	if result.Error != nil {
		return fmt.Errorf("delete credential: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("delete credential for project %s provider %s: %w", projectID, provider, credentialsDomain.ErrCredentialNotFound)
	}
	return nil
}

// Exists checks if a credential exists for a project/provider.
func (r *llmProviderCredentialRepository) Exists(ctx context.Context, projectID ulid.ULID, provider credentialsDomain.LLMProvider) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&credentialsDomain.LLMProviderCredential{}).
		Where("project_id = ? AND provider = ?", projectID, provider).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("check credential exists: %w", err)
	}
	return count > 0, nil
}

// isUniqueViolation checks if the error is a unique constraint violation
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	// PostgreSQL unique violation error code: 23505
	errStr := err.Error()
	return strings.Contains(errStr, "23505") || strings.Contains(errStr, "unique constraint") || strings.Contains(errStr, "duplicate key")
}
