package gateway

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"

	"brokle/internal/core/domain/gateway"
	"brokle/pkg/ulid"
)

// ProviderConfigRepository implements the gateway.ProviderConfigRepository interface
type ProviderConfigRepository struct {
	db *gorm.DB
}

// NewProviderConfigRepository creates a new provider config repository instance
func NewProviderConfigRepository(db *gorm.DB) *ProviderConfigRepository {
	return &ProviderConfigRepository{
		db: db,
	}
}

// Create creates a new provider configuration in the database
func (r *ProviderConfigRepository) Create(ctx context.Context, config *gateway.ProviderConfig) error {
	if config.ID.IsZero() {
		config.ID = ulid.New()
	}

	// Convert maps to JSON for storage
	rateLimitJSON, err := json.Marshal(config.RateLimitOverride)
	if err != nil {
		return fmt.Errorf("failed to marshal rate limit override: %w", err)
	}

	configurationJSON, err := json.Marshal(config.Configuration)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	query := `
		INSERT INTO gateway_provider_configs (
			id, project_id, provider_id, api_key_encrypted, is_enabled,
			custom_base_url, custom_timeout_seconds, rate_limit_override,
			priority_order, configuration, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	now := time.Now()
	config.CreatedAt = now
	config.UpdatedAt = now

	err = r.db.WithContext(ctx).Exec(query,
		config.ID,
		config.ProjectID,
		config.ProviderID,
		config.APIKeyEncrypted,
		config.IsEnabled,
		config.CustomBaseURL,
		config.CustomTimeoutSecs,
		string(rateLimitJSON),
		config.PriorityOrder,
		string(configurationJSON),
		config.CreatedAt,
		config.UpdatedAt,
	).Error

	if err != nil {
		if isDuplicateKeyError(err) {
			return gateway.NewProviderConfigNotFoundError(config.ProjectID.String(), config.ProviderID.String())
		}
		return fmt.Errorf("failed to create provider config: %w", err)
	}

	return nil
}

// GetByID retrieves a provider configuration by its ID
func (r *ProviderConfigRepository) GetByID(ctx context.Context, id ulid.ULID) (*gateway.ProviderConfig, error) {
	var config gateway.ProviderConfig
	var rateLimitJSON, configurationJSON string

	query := `
		SELECT id, project_id, provider_id, api_key_encrypted, is_enabled,
			   custom_base_url, custom_timeout_seconds, rate_limit_override,
			   priority_order, configuration, created_at, updated_at
		FROM gateway_provider_configs
		WHERE id = $1
	`

	row := r.db.WithContext(ctx).Raw(query, id).Row()

	err := row.Scan(
		&config.ID,
		&config.ProjectID,
		&config.ProviderID,
		&config.APIKeyEncrypted,
		&config.IsEnabled,
		&config.CustomBaseURL,
		&config.CustomTimeoutSecs,
		&rateLimitJSON,
		&config.PriorityOrder,
		&configurationJSON,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, gateway.NewProviderConfigNotFoundByIDError(id.String())
		}
		return nil, fmt.Errorf("failed to get provider config by ID: %w", err)
	}

	// Parse JSON fields
	if err := json.Unmarshal([]byte(rateLimitJSON), &config.RateLimitOverride); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rate limit override: %w", err)
	}

	if err := json.Unmarshal([]byte(configurationJSON), &config.Configuration); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	return &config, nil
}

// GetByProjectAndProvider retrieves a provider configuration by project ID and provider ID
func (r *ProviderConfigRepository) GetByProjectAndProvider(ctx context.Context, projectID, providerID ulid.ULID) (*gateway.ProviderConfig, error) {
	var config gateway.ProviderConfig
	var rateLimitJSON, configurationJSON string

	query := `
		SELECT id, project_id, provider_id, api_key_encrypted, is_enabled,
			   custom_base_url, custom_timeout_seconds, rate_limit_override,
			   priority_order, configuration, created_at, updated_at
		FROM gateway_provider_configs
		WHERE project_id = $1 AND provider_id = $2
		ORDER BY priority_order DESC
		LIMIT 1
	`

	row := r.db.WithContext(ctx).Raw(query, projectID, providerID).Row()

	err := row.Scan(
		&config.ID,
		&config.ProjectID,
		&config.ProviderID,
		&config.APIKeyEncrypted,
		&config.IsEnabled,
		&config.CustomBaseURL,
		&config.CustomTimeoutSecs,
		&rateLimitJSON,
		&config.PriorityOrder,
		&configurationJSON,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, gateway.NewProviderConfigNotFoundError(projectID.String(), providerID.String())
		}
		return nil, fmt.Errorf("failed to get provider config by project and provider: %w", err)
	}

	// Parse JSON fields
	if err := json.Unmarshal([]byte(rateLimitJSON), &config.RateLimitOverride); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rate limit override: %w", err)
	}

	if err := json.Unmarshal([]byte(configurationJSON), &config.Configuration); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	return &config, nil
}

// GetByProviderID retrieves provider configurations by provider ID
func (r *ProviderConfigRepository) GetByProviderID(ctx context.Context, providerID ulid.ULID) ([]*gateway.ProviderConfig, error) {
	query := `
		SELECT id, project_id, provider_id, api_key_encrypted, is_enabled,
			   custom_base_url, custom_timeout_seconds, rate_limit_override,
			   priority_order, configuration, created_at, updated_at
		FROM gateway_provider_configs
		WHERE provider_id = $1
		ORDER BY priority_order DESC, created_at DESC
	`

	rows, err := r.db.WithContext(ctx).Raw(query, providerID).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query provider configs: %w", err)
	}
	defer rows.Close()

	return r.scanProviderConfigs(rows)
}

// GetByProjectID retrieves provider configurations by project ID
func (r *ProviderConfigRepository) GetByProjectID(ctx context.Context, projectID ulid.ULID) ([]*gateway.ProviderConfig, error) {
	query := `
		SELECT id, project_id, provider_id, api_key_encrypted, is_enabled,
			   custom_base_url, custom_timeout_seconds, rate_limit_override,
			   priority_order, configuration, created_at, updated_at
		FROM gateway_provider_configs
		WHERE project_id = $1
		ORDER BY priority_order DESC, created_at DESC
	`

	rows, err := r.db.WithContext(ctx).Raw(query, projectID).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query provider configs by project: %w", err)
	}
	defer rows.Close()

	return r.scanProviderConfigs(rows)
}

// GetEnabledByProjectID retrieves enabled provider configurations by project ID
func (r *ProviderConfigRepository) GetEnabledByProjectID(ctx context.Context, projectID ulid.ULID) ([]*gateway.ProviderConfig, error) {
	query := `
		SELECT id, project_id, provider_id, api_key_encrypted, is_enabled,
			   custom_base_url, custom_timeout_seconds, rate_limit_override,
			   priority_order, configuration, created_at, updated_at
		FROM gateway_provider_configs
		WHERE project_id = $1 AND is_enabled = true
		ORDER BY priority_order DESC, created_at DESC
	`

	rows, err := r.db.WithContext(ctx).Raw(query, projectID).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query enabled provider configs by project: %w", err)
	}
	defer rows.Close()

	return r.scanProviderConfigs(rows)
}

// GetByProviderAndOrg retrieves provider configurations by provider ID and project ID (legacy compatibility)
func (r *ProviderConfigRepository) GetByProviderAndOrg(ctx context.Context, providerID, projectID ulid.ULID) ([]*gateway.ProviderConfig, error) {
	query := `
		SELECT id, project_id, provider_id, api_key_encrypted, is_enabled,
			   custom_base_url, custom_timeout_seconds, rate_limit_override,
			   priority_order, configuration, created_at, updated_at
		FROM gateway_provider_configs
		WHERE provider_id = $1 AND project_id = $2
		ORDER BY priority_order DESC, created_at DESC
	`

	rows, err := r.db.WithContext(ctx).Raw(query, providerID, projectID).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query provider configs by provider and project: %w", err)
	}
	defer rows.Close()

	return r.scanProviderConfigs(rows)
}

// GetByProviderAndProject retrieves the first enabled provider configuration by provider ID and project ID
func (r *ProviderConfigRepository) GetByProviderAndProject(ctx context.Context, providerID, projectID ulid.ULID) (*gateway.ProviderConfig, error) {
	var config gateway.ProviderConfig
	var rateLimitJSON, configurationJSON string

	query := `
		SELECT id, project_id, provider_id, api_key_encrypted, is_enabled,
			   custom_base_url, custom_timeout_seconds, rate_limit_override,
			   priority_order, configuration, created_at, updated_at
		FROM gateway_provider_configs
		WHERE provider_id = $1 AND project_id = $2 AND is_enabled = true
		ORDER BY priority_order DESC
		LIMIT 1
	`

	row := r.db.WithContext(ctx).Raw(query, providerID, projectID).Row()

	err := row.Scan(
		&config.ID,
		&config.ProjectID,
		&config.ProviderID,
		&config.APIKeyEncrypted,
		&config.IsEnabled,
		&config.CustomBaseURL,
		&config.CustomTimeoutSecs,
		&rateLimitJSON,
		&config.PriorityOrder,
		&configurationJSON,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, gateway.NewProviderConfigNotFoundByIDError(fmt.Sprintf("provider:%s,project:%s", providerID, projectID))
		}
		return nil, fmt.Errorf("failed to get provider config by provider and project: %w", err)
	}

	// Parse JSON fields
	if err := json.Unmarshal([]byte(rateLimitJSON), &config.RateLimitOverride); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rate limit override: %w", err)
	}

	if err := json.Unmarshal([]byte(configurationJSON), &config.Configuration); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	return &config, nil
}

// GetActiveByProvider retrieves enabled provider configurations by provider ID
func (r *ProviderConfigRepository) GetActiveByProvider(ctx context.Context, providerID ulid.ULID) ([]*gateway.ProviderConfig, error) {
	query := `
		SELECT id, project_id, provider_id, api_key_encrypted, is_enabled,
			   custom_base_url, custom_timeout_seconds, rate_limit_override,
			   priority_order, configuration, created_at, updated_at
		FROM gateway_provider_configs
		WHERE provider_id = $1 AND is_enabled = true
		ORDER BY priority_order DESC, created_at DESC
	`

	rows, err := r.db.WithContext(ctx).Raw(query, providerID).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query active provider configs: %w", err)
	}
	defer rows.Close()

	return r.scanProviderConfigs(rows)
}

// Update updates an existing provider configuration
func (r *ProviderConfigRepository) Update(ctx context.Context, config *gateway.ProviderConfig) error {
	// Convert maps to JSON for storage
	rateLimitJSON, err := json.Marshal(config.RateLimitOverride)
	if err != nil {
		return fmt.Errorf("failed to marshal rate limit override: %w", err)
	}

	configurationJSON, err := json.Marshal(config.Configuration)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	query := `
		UPDATE gateway_provider_configs
		SET project_id = $2, provider_id = $3, api_key_encrypted = $4,
			is_enabled = $5, custom_base_url = $6, custom_timeout_seconds = $7,
			rate_limit_override = $8, priority_order = $9, configuration = $10, updated_at = $11
		WHERE id = $1
	`

	config.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Exec(query,
		config.ID,
		config.ProjectID,
		config.ProviderID,
		config.APIKeyEncrypted,
		config.IsEnabled,
		config.CustomBaseURL,
		config.CustomTimeoutSecs,
		string(rateLimitJSON),
		config.PriorityOrder,
		string(configurationJSON),
		config.UpdatedAt,
	)

	if result.Error != nil {
		return fmt.Errorf("failed to update provider config: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gateway.NewProviderConfigNotFoundError(config.ProjectID.String(), config.ProviderID.String())
	}

	return nil
}

// Delete deletes a provider configuration by ID
func (r *ProviderConfigRepository) Delete(ctx context.Context, id ulid.ULID) error {
	query := `DELETE FROM gateway_provider_configs WHERE id = $1`

	result := r.db.WithContext(ctx).Exec(query, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete provider config: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gateway.NewProviderConfigNotFoundByIDError(id.String())
	}

	return nil
}

// List retrieves provider configurations with pagination
func (r *ProviderConfigRepository) List(ctx context.Context, limit, offset int) ([]*gateway.ProviderConfig, error) {
	query := `
		SELECT id, provider_id, organization_id, environment, priority,
			   config_data, encryption_key, is_encrypted, is_active,
			   created_at, updated_at
		FROM gateway_provider_configs
		ORDER BY priority DESC, created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.WithContext(ctx).Raw(query, limit, offset).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query provider configs: %w", err)
	}
	defer rows.Close()

	return r.scanProviderConfigs(rows)
}

// ListByOrganization retrieves provider configurations by organization ID
func (r *ProviderConfigRepository) ListByOrganization(ctx context.Context, orgID ulid.ULID) ([]*gateway.ProviderConfig, error) {
	query := `
		SELECT id, provider_id, organization_id, environment, priority,
			   config_data, encryption_key, is_encrypted, is_active,
			   created_at, updated_at
		FROM gateway_provider_configs
		WHERE organization_id = $1
		ORDER BY priority DESC, created_at DESC
	`

	rows, err := r.db.WithContext(ctx).Raw(query, orgID).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query provider configs by organization: %w", err)
	}
	defer rows.Close()

	return r.scanProviderConfigs(rows)
}

// SearchConfigs searches provider configurations with filters
func (r *ProviderConfigRepository) SearchConfigs(ctx context.Context, filter *gateway.ProviderConfigFilter) ([]*gateway.ProviderConfig, int, error) {
	whereClause, args := r.buildWhereClause(filter)

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM gateway_provider_configs %s", whereClause)
	var totalCount int
	err := r.db.WithContext(ctx).Raw(countQuery, args...).Scan(&totalCount).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count provider configs: %w", err)
	}

	// Main query with pagination (using fixed limit/offset for now)
	limitClause := " LIMIT 100" // Default limit

	query := fmt.Sprintf(`
		SELECT id, project_id, provider_id, api_key_encrypted, is_enabled,
			   custom_base_url, custom_timeout_seconds, rate_limit_override,
			   priority_order, configuration, created_at, updated_at
		FROM gateway_provider_configs
		%s
		ORDER BY priority_order DESC, created_at DESC%s
	`, whereClause, limitClause)

	rows, err := r.db.WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search provider configs: %w", err)
	}
	defer rows.Close()

	configs, err := r.scanProviderConfigs(rows)
	if err != nil {
		return nil, 0, err
	}

	return configs, totalCount, nil
}

// CountConfigs counts provider configurations with filters
func (r *ProviderConfigRepository) CountConfigs(ctx context.Context, filter *gateway.ProviderConfigFilter) (int64, error) {
	whereClause, args := r.buildWhereClause(filter)
	query := fmt.Sprintf("SELECT COUNT(*) FROM gateway_provider_configs %s", whereClause)

	var count int64
	err := r.db.WithContext(ctx).Raw(query, args...).Scan(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count provider configs: %w", err)
	}

	return count, nil
}

// CreateBatch creates multiple provider configurations in a transaction
func (r *ProviderConfigRepository) CreateBatch(ctx context.Context, configs []*gateway.ProviderConfig) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, config := range configs {
			if err := r.createWithTx(ctx, tx, config); err != nil {
				return err
			}
		}
		return nil
	})
}

// UpdateBatch updates multiple provider configurations in a transaction
func (r *ProviderConfigRepository) UpdateBatch(ctx context.Context, configs []*gateway.ProviderConfig) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, config := range configs {
			if err := r.updateWithTx(ctx, tx, config); err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteBatch deletes multiple provider configurations in a transaction
func (r *ProviderConfigRepository) DeleteBatch(ctx context.Context, ids []ulid.ULID) error {
	if len(ids) == 0 {
		return nil
	}

	query := `DELETE FROM gateway_provider_configs WHERE id = ANY($1)`

	// Convert ULIDs to strings for PostgreSQL array
	stringIDs := make([]string, len(ids))
	for i, id := range ids {
		stringIDs[i] = id.String()
	}

	result := r.db.WithContext(ctx).Exec(query, pq.Array(stringIDs))
	if result.Error != nil {
		return fmt.Errorf("failed to delete provider configs: %w", result.Error)
	}

	return nil
}

// EncryptConfig encrypts a provider configuration's sensitive data
func (r *ProviderConfigRepository) EncryptConfig(ctx context.Context, configID ulid.ULID, encryptedData map[string]interface{}, encryptionKey map[string]string) error {
	// Convert maps to JSON for storage
	configDataJSON, err := json.Marshal(encryptedData)
	if err != nil {
		return fmt.Errorf("failed to marshal encrypted config data: %w", err)
	}

	encryptionKeyJSON, err := json.Marshal(encryptionKey)
	if err != nil {
		return fmt.Errorf("failed to marshal encryption key: %w", err)
	}

	query := `
		UPDATE gateway_provider_configs
		SET config_data = $2, encryption_key = $3, is_encrypted = true, updated_at = $4
		WHERE id = $1
	`

	result := r.db.WithContext(ctx).Exec(query, configID, string(configDataJSON), string(encryptionKeyJSON), time.Now())
	if result.Error != nil {
		return fmt.Errorf("failed to encrypt provider config: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gateway.NewProviderConfigNotFoundByIDError(configID.String())
	}

	return nil
}

// DecryptConfig decrypts a provider configuration's sensitive data
func (r *ProviderConfigRepository) DecryptConfig(ctx context.Context, configID ulid.ULID, decryptedData map[string]interface{}) error {
	// Convert map to JSON for storage
	configDataJSON, err := json.Marshal(decryptedData)
	if err != nil {
		return fmt.Errorf("failed to marshal decrypted config data: %w", err)
	}

	query := `
		UPDATE gateway_provider_configs
		SET config_data = $2, encryption_key = $3, is_encrypted = false, updated_at = $4
		WHERE id = $1
	`

	result := r.db.WithContext(ctx).Exec(query, configID, string(configDataJSON), "{}", time.Now())
	if result.Error != nil {
		return fmt.Errorf("failed to decrypt provider config: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gateway.NewProviderConfigNotFoundByIDError(configID.String())
	}

	return nil
}

// ValidateConfig validates a provider configuration against the provider schema
func (r *ProviderConfigRepository) ValidateConfig(ctx context.Context, configID ulid.ULID) error {
	// Implementation depends on validation logic - for now, just check if config exists
	_, err := r.GetByID(ctx, configID)
	if err != nil {
		return fmt.Errorf("failed to validate provider config: %w", err)
	}

	// TODO: Add actual validation logic here
	// This could involve checking against provider-specific schemas,
	// validating required fields, testing connection, etc.

	return nil
}

// UpdateConfigPriority updates the priority of a provider configuration
func (r *ProviderConfigRepository) UpdateConfigPriority(ctx context.Context, configID ulid.ULID, priority int) error {
	query := `
		UPDATE gateway_provider_configs
		SET priority = $2, updated_at = $3
		WHERE id = $1
	`

	result := r.db.WithContext(ctx).Exec(query, configID, priority, time.Now())
	if result.Error != nil {
		return fmt.Errorf("failed to update config priority: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gateway.NewProviderConfigNotFoundByIDError(configID.String())
	}

	return nil
}

// ActivateConfig activates a provider configuration
func (r *ProviderConfigRepository) ActivateConfig(ctx context.Context, configID ulid.ULID) error {
	query := `
		UPDATE gateway_provider_configs
		SET is_active = true, updated_at = $2
		WHERE id = $1
	`

	result := r.db.WithContext(ctx).Exec(query, configID, time.Now())
	if result.Error != nil {
		return fmt.Errorf("failed to activate provider config: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gateway.NewProviderConfigNotFoundByIDError(configID.String())
	}

	return nil
}

// DeactivateConfig deactivates a provider configuration
func (r *ProviderConfigRepository) DeactivateConfig(ctx context.Context, configID ulid.ULID) error {
	query := `
		UPDATE gateway_provider_configs
		SET is_active = false, updated_at = $2
		WHERE id = $1
	`

	result := r.db.WithContext(ctx).Exec(query, configID, time.Now())
	if result.Error != nil {
		return fmt.Errorf("failed to deactivate provider config: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gateway.NewProviderConfigNotFoundByIDError(configID.String())
	}

	return nil
}

// Helper methods

func (r *ProviderConfigRepository) scanProviderConfigs(rows *sql.Rows) ([]*gateway.ProviderConfig, error) {
	var configs []*gateway.ProviderConfig

	for rows.Next() {
		var config gateway.ProviderConfig
		var rateLimitJSON, configurationJSON string

		err := rows.Scan(
			&config.ID,
			&config.ProjectID,
			&config.ProviderID,
			&config.APIKeyEncrypted,
			&config.IsEnabled,
			&config.CustomBaseURL,
			&config.CustomTimeoutSecs,
			&rateLimitJSON,
			&config.PriorityOrder,
			&configurationJSON,
			&config.CreatedAt,
			&config.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan provider config row: %w", err)
		}

		// Parse JSON fields
		if err := json.Unmarshal([]byte(rateLimitJSON), &config.RateLimitOverride); err != nil {
			return nil, fmt.Errorf("failed to unmarshal rate limit override: %w", err)
		}

		if err := json.Unmarshal([]byte(configurationJSON), &config.Configuration); err != nil {
			return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
		}

		configs = append(configs, &config)
	}

	return configs, nil
}

func (r *ProviderConfigRepository) buildWhereClause(filter *gateway.ProviderConfigFilter) (string, []interface{}) {
	if filter == nil {
		return "", nil
	}

	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.ProviderID != nil {
		conditions = append(conditions, fmt.Sprintf("provider_id = $%d", argIndex))
		args = append(args, *filter.ProviderID)
		argIndex++
	}

	if filter.ProjectID != nil {
		conditions = append(conditions, fmt.Sprintf("project_id = $%d", argIndex))
		args = append(args, *filter.ProjectID)
		argIndex++
	}

	if filter.IsEnabled != nil {
		conditions = append(conditions, fmt.Sprintf("is_enabled = $%d", argIndex))
		args = append(args, *filter.IsEnabled)
		argIndex++
	}

	if filter.MinPriority != nil {
		conditions = append(conditions, fmt.Sprintf("priority_order >= $%d", argIndex))
		args = append(args, *filter.MinPriority)
		argIndex++
	}

	if filter.MaxPriority != nil {
		conditions = append(conditions, fmt.Sprintf("priority_order <= $%d", argIndex))
		args = append(args, *filter.MaxPriority)
		argIndex++
	}

	// CreatedAfter and CreatedBefore fields are not available in ProviderConfigFilter

	if len(conditions) == 0 {
		return "", args
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}

func (r *ProviderConfigRepository) createWithTx(ctx context.Context, tx *gorm.DB, config *gateway.ProviderConfig) error {
	if config.ID.IsZero() {
		config.ID = ulid.New()
	}

	// Convert maps to JSON for storage
	rateLimitJSON, err := json.Marshal(config.RateLimitOverride)
	if err != nil {
		return fmt.Errorf("failed to marshal rate limit override: %w", err)
	}

	configurationJSON, err := json.Marshal(config.Configuration)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	query := `
		INSERT INTO gateway_provider_configs (
			id, project_id, provider_id, api_key_encrypted, is_enabled,
			custom_base_url, custom_timeout_seconds, rate_limit_override,
			priority_order, configuration, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	now := time.Now()
	config.CreatedAt = now
	config.UpdatedAt = now

	err = tx.WithContext(ctx).Exec(query,
		config.ID,
		config.ProjectID,
		config.ProviderID,
		config.APIKeyEncrypted,
		config.IsEnabled,
		config.CustomBaseURL,
		config.CustomTimeoutSecs,
		string(rateLimitJSON),
		config.PriorityOrder,
		string(configurationJSON),
		config.CreatedAt,
		config.UpdatedAt,
	).Error

	if err != nil {
		if isDuplicateKeyError(err) {
			return gateway.NewProviderConfigNotFoundByIDError(config.ID.String())
		}
		return fmt.Errorf("failed to create provider config: %w", err)
	}

	return nil
}

func (r *ProviderConfigRepository) updateWithTx(ctx context.Context, tx *gorm.DB, config *gateway.ProviderConfig) error {
	// Convert maps to JSON for storage
	rateLimitJSON, err := json.Marshal(config.RateLimitOverride)
	if err != nil {
		return fmt.Errorf("failed to marshal rate limit override: %w", err)
	}

	configurationJSON, err := json.Marshal(config.Configuration)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	query := `
		UPDATE gateway_provider_configs
		SET project_id = $2, provider_id = $3, api_key_encrypted = $4,
			is_enabled = $5, custom_base_url = $6, custom_timeout_seconds = $7,
			rate_limit_override = $8, priority_order = $9, configuration = $10, updated_at = $11
		WHERE id = $1
	`

	config.UpdatedAt = time.Now()

	result := tx.WithContext(ctx).Exec(query,
		config.ID,
		config.ProjectID,
		config.ProviderID,
		config.APIKeyEncrypted,
		config.IsEnabled,
		config.CustomBaseURL,
		config.CustomTimeoutSecs,
		string(rateLimitJSON),
		config.PriorityOrder,
		string(configurationJSON),
		config.UpdatedAt,
	)

	if result.Error != nil {
		return fmt.Errorf("failed to update provider config: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gateway.NewProviderConfigNotFoundByIDError(config.ID.String())
	}

	return nil
}

// CountProjectsForProvider counts the number of projects that have configurations for a specific provider
func (r *ProviderConfigRepository) CountProjectsForProvider(ctx context.Context, providerID ulid.ULID) (int64, error) {
	query := `
		SELECT COUNT(DISTINCT project_id) 
		FROM gateway_provider_configs 
		WHERE provider_id = $1
	`

	var count int64
	err := r.db.WithContext(ctx).Raw(query, providerID).Scan(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count projects for provider: %w", err)
	}

	return count, nil
}

// EncryptAPIKey encrypts an API key (placeholder implementation)
func (r *ProviderConfigRepository) EncryptAPIKey(ctx context.Context, plaintext string) (string, error) {
	// TODO: Implement proper encryption
	// For now, return the plaintext (NOT SECURE - FOR DEVELOPMENT ONLY)
	return plaintext, nil
}

// DecryptAPIKey decrypts an API key (placeholder implementation)
func (r *ProviderConfigRepository) DecryptAPIKey(ctx context.Context, encrypted string) (string, error) {
	// TODO: Implement proper decryption
	// For now, return the encrypted value as-is (NOT SECURE - FOR DEVELOPMENT ONLY)
	return encrypted, nil
}

// ListEnabled retrieves all enabled provider configurations
func (r *ProviderConfigRepository) ListEnabled(ctx context.Context) ([]*gateway.ProviderConfig, error) {
	query := `
		SELECT id, project_id, provider_id, api_key_encrypted, is_enabled,
			   custom_base_url, custom_timeout_seconds, rate_limit_override,
			   priority_order, configuration, created_at, updated_at
		FROM gateway_provider_configs
		WHERE is_enabled = true
		ORDER BY priority_order ASC
	`

	rows, err := r.db.WithContext(ctx).Raw(query).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query enabled provider configs: %w", err)
	}
	defer rows.Close()

	return r.scanProviderConfigs(rows)
}

// GetByProjectIDWithProvider retrieves provider configs for a project with provider details
func (r *ProviderConfigRepository) GetByProjectIDWithProvider(ctx context.Context, projectID ulid.ULID) ([]*gateway.ProviderConfig, error) {
	// For now, delegate to existing GetByProjectID method
	// In a full implementation, this would join with provider table to include provider details
	return r.GetByProjectID(ctx, projectID)
}

// GetOrderedByPriority retrieves provider configs for a project ordered by priority
func (r *ProviderConfigRepository) GetOrderedByPriority(ctx context.Context, projectID ulid.ULID) ([]*gateway.ProviderConfig, error) {
	query := `
		SELECT id, project_id, provider_id, api_key_encrypted, is_enabled,
			   custom_base_url, custom_timeout_seconds, rate_limit_override,
			   priority_order, configuration, created_at, updated_at
		FROM gateway_provider_configs
		WHERE project_id = $1
		ORDER BY priority_order ASC
	`

	rows, err := r.db.WithContext(ctx).Raw(query, projectID).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query provider configs by priority: %w", err)
	}
	defer rows.Close()

	return r.scanProviderConfigs(rows)
}

// TestProviderConnection tests the connection to a provider using the given configuration
func (r *ProviderConfigRepository) TestProviderConnection(ctx context.Context, config *gateway.ProviderConfig) error {
	// TODO: Implement actual connection testing logic
	// This would involve making a test request to the provider's API
	// to verify that the configuration is valid and the connection works

	// For now, just validate that the config has required fields
	if config.APIKeyEncrypted == "" {
		return fmt.Errorf("API key is required for provider connection test")
	}

	// In a real implementation, this would:
	// 1. Decrypt the API key
	// 2. Make a test request to the provider's API
	// 3. Return success/failure based on the response

	return nil // Placeholder - always succeeds
}

// UpdatePriority updates the priority of a provider configuration
func (r *ProviderConfigRepository) UpdatePriority(ctx context.Context, configID ulid.ULID, priority int) error {
	// Delegate to existing UpdateConfigPriority method
	return r.UpdateConfigPriority(ctx, configID, priority)
}

// ValidateConfiguration validates a provider configuration
func (r *ProviderConfigRepository) ValidateConfiguration(ctx context.Context, config *gateway.ProviderConfig) error {
	// TODO: Implement actual validation logic
	// This would involve validating the provider configuration schema,
	// checking required fields, validating API key format, etc.

	// Basic validation - check required fields
	if config.ProviderID.IsZero() {
		return fmt.Errorf("provider ID is required")
	}

	if config.ProjectID.IsZero() {
		return fmt.Errorf("project ID is required")
	}

	if config.APIKeyEncrypted == "" {
		return fmt.Errorf("API key is required")
	}

	// In a full implementation, this would:
	// 1. Validate against provider-specific schemas
	// 2. Check API key format
	// 3. Validate configuration parameters
	// 4. Test connection to provider

	return nil // Placeholder - always succeeds after basic validation
}
