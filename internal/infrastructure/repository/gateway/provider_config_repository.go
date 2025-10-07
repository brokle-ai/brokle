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
	configDataJSON, err := json.Marshal(config.ConfigData)
	if err != nil {
		return fmt.Errorf("failed to marshal config data: %w", err)
	}

	encryptionKeyJSON, err := json.Marshal(config.EncryptionKey)
	if err != nil {
		return fmt.Errorf("failed to marshal encryption key: %w", err)
	}

	query := `
		INSERT INTO gateway_provider_configs (
			id, provider_id, organization_id, environment, priority,
			config_data, encryption_key, is_encrypted, is_active,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	now := time.Now()
	config.CreatedAt = now
	config.UpdatedAt = now

	err = r.db.WithContext(ctx).Exec(query,
		config.ID,
		config.ProviderID,
		config.OrganizationID,
		string(config.Environment),
		config.Priority,
		string(configDataJSON),
		string(encryptionKeyJSON),
		config.IsEncrypted,
		config.IsActive,
		config.CreatedAt,
		config.UpdatedAt,
	).Error

	if err != nil {
		if isDuplicateKeyError(err) {
			return gateway.NewProviderConfigNotFoundError(config.ProviderID.String())
		}
		return fmt.Errorf("failed to create provider config: %w", err)
	}

	return nil
}

// GetByID retrieves a provider configuration by its ID
func (r *ProviderConfigRepository) GetByID(ctx context.Context, id ulid.ULID) (*gateway.ProviderConfig, error) {
	var config gateway.ProviderConfig
	var environment string
	var configDataJSON, encryptionKeyJSON string

	query := `
		SELECT id, provider_id, organization_id, environment, priority,
			   config_data, encryption_key, is_encrypted, is_active,
			   created_at, updated_at
		FROM gateway_provider_configs
		WHERE id = $1
	`

	row := r.db.WithContext(ctx).Raw(query, id).Row()

	err := row.Scan(
		&config.ID,
		&config.ProviderID,
		&config.OrganizationID,
		&environment,
		&config.Priority,
		&configDataJSON,
		&encryptionKeyJSON,
		&config.IsEncrypted,
		&config.IsActive,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, gateway.NewProviderConfigNotFoundError(id.String())
		}
		return nil, fmt.Errorf("failed to get provider config by ID: %w", err)
	}

	// Parse enum and JSON fields
	config.Environment = gateway.Environment(environment)

	if err := json.Unmarshal([]byte(configDataJSON), &config.ConfigData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config data: %w", err)
	}

	if err := json.Unmarshal([]byte(encryptionKeyJSON), &config.EncryptionKey); err != nil {
		return nil, fmt.Errorf("failed to unmarshal encryption key: %w", err)
	}

	return &config, nil
}

// GetByProvider retrieves provider configurations by provider ID
func (r *ProviderConfigRepository) GetByProvider(ctx context.Context, providerID ulid.ULID) ([]*gateway.ProviderConfig, error) {
	query := `
		SELECT id, provider_id, organization_id, environment, priority,
			   config_data, encryption_key, is_encrypted, is_active,
			   created_at, updated_at
		FROM gateway_provider_configs
		WHERE provider_id = $1
		ORDER BY priority DESC, created_at DESC
	`

	rows, err := r.db.WithContext(ctx).Raw(query, providerID).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query provider configs: %w", err)
	}
	defer rows.Close()

	return r.scanProviderConfigs(rows)
}

// GetByProviderAndOrg retrieves provider configurations by provider ID and organization ID
func (r *ProviderConfigRepository) GetByProviderAndOrg(ctx context.Context, providerID, orgID ulid.ULID) ([]*gateway.ProviderConfig, error) {
	query := `
		SELECT id, provider_id, organization_id, environment, priority,
			   config_data, encryption_key, is_encrypted, is_active,
			   created_at, updated_at
		FROM gateway_provider_configs
		WHERE provider_id = $1 AND organization_id = $2
		ORDER BY priority DESC, created_at DESC
	`

	rows, err := r.db.WithContext(ctx).Raw(query, providerID, orgID).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query provider configs by provider and org: %w", err)
	}
	defer rows.Close()

	return r.scanProviderConfigs(rows)
}

// GetByProviderOrgAndEnv retrieves a specific provider configuration by provider, organization, and environment
func (r *ProviderConfigRepository) GetByProviderOrgAndEnv(ctx context.Context, providerID, orgID ulid.ULID, env gateway.Environment) (*gateway.ProviderConfig, error) {
	var config gateway.ProviderConfig
	var environment string
	var configDataJSON, encryptionKeyJSON string

	query := `
		SELECT id, provider_id, organization_id, environment, priority,
			   config_data, encryption_key, is_encrypted, is_active,
			   created_at, updated_at
		FROM gateway_provider_configs
		WHERE provider_id = $1 AND organization_id = $2 AND environment = $3 AND is_active = true
		ORDER BY priority DESC
		LIMIT 1
	`

	row := r.db.WithContext(ctx).Raw(query, providerID, orgID, string(env)).Row()

	err := row.Scan(
		&config.ID,
		&config.ProviderID,
		&config.OrganizationID,
		&environment,
		&config.Priority,
		&configDataJSON,
		&encryptionKeyJSON,
		&config.IsEncrypted,
		&config.IsActive,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, gateway.NewProviderConfigNotFoundError(fmt.Sprintf("provider:%s,org:%s,env:%s", providerID, orgID, env))
		}
		return nil, fmt.Errorf("failed to get provider config by provider, org, and env: %w", err)
	}

	// Parse enum and JSON fields
	config.Environment = gateway.Environment(environment)

	if err := json.Unmarshal([]byte(configDataJSON), &config.ConfigData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config data: %w", err)
	}

	if err := json.Unmarshal([]byte(encryptionKeyJSON), &config.EncryptionKey); err != nil {
		return nil, fmt.Errorf("failed to unmarshal encryption key: %w", err)
	}

	return &config, nil
}

// GetActiveByProvider retrieves active provider configurations by provider ID
func (r *ProviderConfigRepository) GetActiveByProvider(ctx context.Context, providerID ulid.ULID) ([]*gateway.ProviderConfig, error) {
	query := `
		SELECT id, provider_id, organization_id, environment, priority,
			   config_data, encryption_key, is_encrypted, is_active,
			   created_at, updated_at
		FROM gateway_provider_configs
		WHERE provider_id = $1 AND is_active = true
		ORDER BY priority DESC, created_at DESC
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
	configDataJSON, err := json.Marshal(config.ConfigData)
	if err != nil {
		return fmt.Errorf("failed to marshal config data: %w", err)
	}

	encryptionKeyJSON, err := json.Marshal(config.EncryptionKey)
	if err != nil {
		return fmt.Errorf("failed to marshal encryption key: %w", err)
	}

	query := `
		UPDATE gateway_provider_configs
		SET provider_id = $2, organization_id = $3, environment = $4,
			priority = $5, config_data = $6, encryption_key = $7,
			is_encrypted = $8, is_active = $9, updated_at = $10
		WHERE id = $1
	`

	config.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Exec(query,
		config.ID,
		config.ProviderID,
		config.OrganizationID,
		string(config.Environment),
		config.Priority,
		string(configDataJSON),
		string(encryptionKeyJSON),
		config.IsEncrypted,
		config.IsActive,
		config.UpdatedAt,
	)

	if result.Error != nil {
		return fmt.Errorf("failed to update provider config: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gateway.NewProviderConfigNotFoundError(config.ID.String())
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
		return gateway.NewProviderConfigNotFoundError(id.String())
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

	// Main query with pagination
	limitClause := ""
	if filter != nil {
		if filter.Limit > 0 {
			args = append(args, filter.Limit)
			limitClause = fmt.Sprintf(" LIMIT $%d", len(args))
		}
		if filter.Offset > 0 {
			args = append(args, filter.Offset)
			limitClause += fmt.Sprintf(" OFFSET $%d", len(args))
		}
	}

	query := fmt.Sprintf(`
		SELECT id, provider_id, organization_id, environment, priority,
			   config_data, encryption_key, is_encrypted, is_active,
			   created_at, updated_at
		FROM gateway_provider_configs
		%s
		ORDER BY priority DESC, created_at DESC%s
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
		return gateway.NewProviderConfigNotFoundError(configID.String())
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
		return gateway.NewProviderConfigNotFoundError(configID.String())
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
		return gateway.NewProviderConfigNotFoundError(configID.String())
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
		return gateway.NewProviderConfigNotFoundError(configID.String())
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
		return gateway.NewProviderConfigNotFoundError(configID.String())
	}

	return nil
}

// Helper methods

func (r *ProviderConfigRepository) scanProviderConfigs(rows *sql.Rows) ([]*gateway.ProviderConfig, error) {
	var configs []*gateway.ProviderConfig

	for rows.Next() {
		var config gateway.ProviderConfig
		var environment string
		var configDataJSON, encryptionKeyJSON string

		err := rows.Scan(
			&config.ID,
			&config.ProviderID,
			&config.OrganizationID,
			&environment,
			&config.Priority,
			&configDataJSON,
			&encryptionKeyJSON,
			&config.IsEncrypted,
			&config.IsActive,
			&config.CreatedAt,
			&config.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan provider config row: %w", err)
		}

		// Parse enum and JSON fields
		config.Environment = gateway.Environment(environment)

		if err := json.Unmarshal([]byte(configDataJSON), &config.ConfigData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config data: %w", err)
		}

		if err := json.Unmarshal([]byte(encryptionKeyJSON), &config.EncryptionKey); err != nil {
			return nil, fmt.Errorf("failed to unmarshal encryption key: %w", err)
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

	if filter.OrganizationID != nil {
		conditions = append(conditions, fmt.Sprintf("organization_id = $%d", argIndex))
		args = append(args, *filter.OrganizationID)
		argIndex++
	}

	if filter.Environment != nil {
		conditions = append(conditions, fmt.Sprintf("environment = $%d", argIndex))
		args = append(args, string(*filter.Environment))
		argIndex++
	}

	if filter.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *filter.IsActive)
		argIndex++
	}

	if filter.IsEncrypted != nil {
		conditions = append(conditions, fmt.Sprintf("is_encrypted = $%d", argIndex))
		args = append(args, *filter.IsEncrypted)
		argIndex++
	}

	if filter.MinPriority != nil {
		conditions = append(conditions, fmt.Sprintf("priority >= $%d", argIndex))
		args = append(args, *filter.MinPriority)
		argIndex++
	}

	if filter.MaxPriority != nil {
		conditions = append(conditions, fmt.Sprintf("priority <= $%d", argIndex))
		args = append(args, *filter.MaxPriority)
		argIndex++
	}

	if filter.CreatedAfter != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, *filter.CreatedAfter)
		argIndex++
	}

	if filter.CreatedBefore != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, *filter.CreatedBefore)
		argIndex++
	}

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
	configDataJSON, err := json.Marshal(config.ConfigData)
	if err != nil {
		return fmt.Errorf("failed to marshal config data: %w", err)
	}

	encryptionKeyJSON, err := json.Marshal(config.EncryptionKey)
	if err != nil {
		return fmt.Errorf("failed to marshal encryption key: %w", err)
	}

	query := `
		INSERT INTO gateway_provider_configs (
			id, provider_id, organization_id, environment, priority,
			config_data, encryption_key, is_encrypted, is_active,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	now := time.Now()
	config.CreatedAt = now
	config.UpdatedAt = now

	err = tx.WithContext(ctx).Exec(query,
		config.ID,
		config.ProviderID,
		config.OrganizationID,
		string(config.Environment),
		config.Priority,
		string(configDataJSON),
		string(encryptionKeyJSON),
		config.IsEncrypted,
		config.IsActive,
		config.CreatedAt,
		config.UpdatedAt,
	).Error

	if err != nil {
		if isDuplicateKeyError(err) {
			return gateway.NewProviderConfigNotFoundError(config.ProviderID.String())
		}
		return fmt.Errorf("failed to create provider config: %w", err)
	}

	return nil
}

func (r *ProviderConfigRepository) updateWithTx(ctx context.Context, tx *gorm.DB, config *gateway.ProviderConfig) error {
	// Convert maps to JSON for storage
	configDataJSON, err := json.Marshal(config.ConfigData)
	if err != nil {
		return fmt.Errorf("failed to marshal config data: %w", err)
	}

	encryptionKeyJSON, err := json.Marshal(config.EncryptionKey)
	if err != nil {
		return fmt.Errorf("failed to marshal encryption key: %w", err)
	}

	query := `
		UPDATE gateway_provider_configs
		SET provider_id = $2, organization_id = $3, environment = $4,
			priority = $5, config_data = $6, encryption_key = $7,
			is_encrypted = $8, is_active = $9, updated_at = $10
		WHERE id = $1
	`

	config.UpdatedAt = time.Now()

	result := tx.WithContext(ctx).Exec(query,
		config.ID,
		config.ProviderID,
		config.OrganizationID,
		string(config.Environment),
		config.Priority,
		string(configDataJSON),
		string(encryptionKeyJSON),
		config.IsEncrypted,
		config.IsActive,
		config.UpdatedAt,
	)

	if result.Error != nil {
		return fmt.Errorf("failed to update provider config: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gateway.NewProviderConfigNotFoundError(config.ID.String())
	}

	return nil
}