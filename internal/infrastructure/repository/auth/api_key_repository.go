package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	authDomain "brokle/internal/core/domain/auth"
	"brokle/pkg/ulid"
)

// apiKeyRepository implements authDomain.APIKeyRepository using GORM
type apiKeyRepository struct {
	db *gorm.DB
}

// NewAPIKeyRepository creates a new API key repository instance
func NewAPIKeyRepository(db *gorm.DB) authDomain.APIKeyRepository {
	return &apiKeyRepository{
		db: db,
	}
}

// Create creates a new API key
func (r *apiKeyRepository) Create(ctx context.Context, apiKey *authDomain.APIKey) error {
	return r.db.WithContext(ctx).Create(apiKey).Error
}

// GetByID retrieves an API key by ID
func (r *apiKeyRepository) GetByID(ctx context.Context, id ulid.ULID) (*authDomain.APIKey, error) {
	var apiKey authDomain.APIKey
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&apiKey).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("get API key by ID %s: %w", id, authDomain.ErrNotFound)
		}
		return nil, fmt.Errorf("database error getting API key by ID %s: %w", id, err)
	}
	return &apiKey, nil
}

// GetByKeyHash retrieves an API key by key hash
func (r *apiKeyRepository) GetByKeyHash(ctx context.Context, keyHash string) (*authDomain.APIKey, error) {
	var apiKey authDomain.APIKey
	err := r.db.WithContext(ctx).Where("key_hash = ? AND deleted_at IS NULL", keyHash).First(&apiKey).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("get API key by hash %s: %w", keyHash, authDomain.ErrNotFound)
		}
		return nil, fmt.Errorf("database error getting API key by hash %s: %w", keyHash, err)
	}
	return &apiKey, nil
}

// Update updates an API key
func (r *apiKeyRepository) Update(ctx context.Context, apiKey *authDomain.APIKey) error {
	return r.db.WithContext(ctx).Save(apiKey).Error
}

// Delete soft deletes an API key
func (r *apiKeyRepository) Delete(ctx context.Context, id ulid.ULID) error {
	return r.db.WithContext(ctx).Model(&authDomain.APIKey{}).Where("id = ?", id).Update("deleted_at", time.Now()).Error
}

// UpdateLastUsed updates the last used timestamp
func (r *apiKeyRepository) UpdateLastUsed(ctx context.Context, id ulid.ULID) error {
	return r.db.WithContext(ctx).
		Model(&authDomain.APIKey{}).
		Where("id = ?", id).
		Update("last_used_at", time.Now()).Error
}

// GetByUserID retrieves API keys for a user
func (r *apiKeyRepository) GetByUserID(ctx context.Context, userID ulid.ULID) ([]*authDomain.APIKey, error) {
	var apiKeys []*authDomain.APIKey
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Order("created_at DESC").
		Find(&apiKeys).Error
	return apiKeys, err
}

// GetByOrganizationID retrieves API keys for an organization
func (r *apiKeyRepository) GetByOrganizationID(ctx context.Context, orgID ulid.ULID) ([]*authDomain.APIKey, error) {
	var apiKeys []*authDomain.APIKey
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND deleted_at IS NULL", orgID).
		Order("created_at DESC").
		Find(&apiKeys).Error
	return apiKeys, err
}

// GetByProjectID retrieves API keys for a project
func (r *apiKeyRepository) GetByProjectID(ctx context.Context, projectID ulid.ULID) ([]*authDomain.APIKey, error) {
	var apiKeys []*authDomain.APIKey
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND deleted_at IS NULL", projectID).
		Order("created_at DESC").
		Find(&apiKeys).Error
	return apiKeys, err
}


// GetByFilters retrieves API keys based on filters
func (r *apiKeyRepository) GetByFilters(ctx context.Context, filters *authDomain.APIKeyFilters) ([]*authDomain.APIKey, error) {
	var apiKeys []*authDomain.APIKey
	query := r.db.WithContext(ctx).Where("deleted_at IS NULL")

	// Apply filters
	if filters.UserID != nil {
		query = query.Where("user_id = ?", *filters.UserID)
	}
	if filters.OrganizationID != nil {
		query = query.Where("organization_id = ?", *filters.OrganizationID)
	}
	if filters.ProjectID != nil {
		query = query.Where("project_id = ?", *filters.ProjectID)
	}
	if filters.Environment != nil && *filters.Environment != "" {
		query = query.Where("default_environment = ?", *filters.Environment)
	}
	if filters.IsActive != nil {
		query = query.Where("is_active = ?", *filters.IsActive)
	}
	if filters.IsExpired != nil {
		if *filters.IsExpired {
			query = query.Where("expires_at IS NOT NULL AND expires_at < ?", time.Now())
		} else {
			query = query.Where("expires_at IS NULL OR expires_at > ?", time.Now())
		}
	}

	// Apply sorting
	switch filters.SortBy {
	case "name":
		if filters.SortOrder == "desc" {
			query = query.Order("name DESC")
		} else {
			query = query.Order("name ASC")
		}
	case "created_at":
		if filters.SortOrder == "desc" {
			query = query.Order("created_at DESC")
		} else {
			query = query.Order("created_at ASC")
		}
	case "last_used_at":
		if filters.SortOrder == "desc" {
			query = query.Order("last_used_at DESC")
		} else {
			query = query.Order("last_used_at ASC")
		}
	default:
		query = query.Order("created_at DESC")
	}

	// Apply pagination
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	err := query.Find(&apiKeys).Error
	return apiKeys, err
}

// CleanupExpiredAPIKeys removes expired API keys
func (r *apiKeyRepository) CleanupExpiredAPIKeys(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at IS NOT NULL AND expires_at < ?", time.Now()).
		Delete(&authDomain.APIKey{}).Error
}

// DeactivateAPIKey deactivates an API key
func (r *apiKeyRepository) DeactivateAPIKey(ctx context.Context, id ulid.ULID) error {
	return r.db.WithContext(ctx).
		Model(&authDomain.APIKey{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}

// MarkAsUsed updates the last used timestamp for an API key
func (r *apiKeyRepository) MarkAsUsed(ctx context.Context, id ulid.ULID) error {
	return r.db.WithContext(ctx).
		Model(&authDomain.APIKey{}).
		Where("id = ?", id).
		Update("last_used_at", time.Now()).Error
}

// GetAPIKeyCount returns the total count of API keys for a user
func (r *apiKeyRepository) GetAPIKeyCount(ctx context.Context, userID ulid.ULID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&authDomain.APIKey{}).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Count(&count).Error
	return int(count), err
}

// GetActiveAPIKeyCount returns the count of active API keys for a user
func (r *apiKeyRepository) GetActiveAPIKeyCount(ctx context.Context, userID ulid.ULID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&authDomain.APIKey{}).
		Where("user_id = ? AND is_active = true AND deleted_at IS NULL", userID).
		Count(&count).Error
	return int(count), err
}