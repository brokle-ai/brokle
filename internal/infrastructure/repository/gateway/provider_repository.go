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

// ProviderRepository implements the gateway.ProviderRepository interface
type ProviderRepository struct {
	db *gorm.DB
}

// NewProviderRepository creates a new provider repository instance
func NewProviderRepository(db *gorm.DB) *ProviderRepository {
	return &ProviderRepository{
		db: db,
	}
}

// Create creates a new provider in the database
func (r *ProviderRepository) Create(ctx context.Context, provider *gateway.Provider) error {
	if provider.ID.IsZero() {
		provider.ID = ulid.New()
	}

	// Convert maps to JSON for storage
	featuresJSON, err := json.Marshal(provider.SupportedFeatures)
	if err != nil {
		return fmt.Errorf("failed to marshal supported features: %w", err)
	}

	rateLimitsJSON, err := json.Marshal(provider.RateLimits)
	if err != nil {
		return fmt.Errorf("failed to marshal rate limits: %w", err)
	}

	// Prepare SQL statement
	query := `
		INSERT INTO gateway_providers (
			id, name, type, base_url, is_enabled, default_timeout_seconds,
			max_retries, health_check_url, supported_features, rate_limits,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	now := time.Now()
	provider.CreatedAt = now
	provider.UpdatedAt = now

	err = r.db.WithContext(ctx).Exec(query,
		provider.ID,
		provider.Name,
		string(provider.Type),
		provider.BaseURL,
		provider.IsEnabled,
		provider.DefaultTimeoutSecs,
		provider.MaxRetries,
		provider.HealthCheckURL,
		string(featuresJSON),
		string(rateLimitsJSON),
		provider.CreatedAt,
		provider.UpdatedAt,
	).Error

	if err != nil {
		if isDuplicateKeyError(err) {
			return gateway.NewProviderNotFoundError(provider.Name)
		}
		return fmt.Errorf("failed to create provider: %w", err)
	}

	return nil
}

// GetByID retrieves a provider by its ID
func (r *ProviderRepository) GetByID(ctx context.Context, id ulid.ULID) (*gateway.Provider, error) {
	var provider gateway.Provider
	var providerType string
	var featuresJSON, rateLimitsJSON string

	query := `
		SELECT id, name, type, base_url, is_enabled, default_timeout_seconds,
			   max_retries, health_check_url, supported_features, rate_limits,
			   created_at, updated_at
		FROM gateway_providers
		WHERE id = $1
	`

	row := r.db.WithContext(ctx).Raw(query, id).Row()

	err := row.Scan(
		&provider.ID,
		&provider.Name,
		&providerType,
		&provider.BaseURL,
		&provider.IsEnabled,
		&provider.DefaultTimeoutSecs,
		&provider.MaxRetries,
		&provider.HealthCheckURL,
		&featuresJSON,
		&rateLimitsJSON,
		&provider.CreatedAt,
		&provider.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, gateway.NewProviderNotFoundError(id.String())
		}
		return nil, fmt.Errorf("failed to get provider by ID: %w", err)
	}

	// Parse enum and JSON fields
	provider.Type = gateway.ProviderType(providerType)

	if err := json.Unmarshal([]byte(featuresJSON), &provider.SupportedFeatures); err != nil {
		return nil, fmt.Errorf("failed to unmarshal supported features: %w", err)
	}

	if err := json.Unmarshal([]byte(rateLimitsJSON), &provider.RateLimits); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rate limits: %w", err)
	}

	return &provider, nil
}

// GetByName retrieves a provider by its name
func (r *ProviderRepository) GetByName(ctx context.Context, name string) (*gateway.Provider, error) {
	var provider gateway.Provider
	var providerType string
	var featuresJSON, rateLimitsJSON string

	query := `
		SELECT id, name, type, base_url, is_enabled, default_timeout_seconds,
			   max_retries, health_check_url, supported_features, rate_limits,
			   created_at, updated_at
		FROM gateway_providers
		WHERE name = $1
	`

	row := r.db.WithContext(ctx).Raw(query, name).Row()

	err := row.Scan(
		&provider.ID,
		&provider.Name,
		&providerType,
		&provider.BaseURL,
		&provider.IsEnabled,
		&provider.DefaultTimeoutSecs,
		&provider.MaxRetries,
		&provider.HealthCheckURL,
		&featuresJSON,
		&rateLimitsJSON,
		&provider.CreatedAt,
		&provider.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, gateway.NewProviderNotFoundError(name)
		}
		return nil, fmt.Errorf("failed to get provider by name: %w", err)
	}

	// Parse enum and JSON fields
	provider.Type = gateway.ProviderType(providerType)

	if err := json.Unmarshal([]byte(featuresJSON), &provider.SupportedFeatures); err != nil {
		return nil, fmt.Errorf("failed to unmarshal supported features: %w", err)
	}

	if err := json.Unmarshal([]byte(rateLimitsJSON), &provider.RateLimits); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rate limits: %w", err)
	}

	return &provider, nil
}

// GetByType retrieves providers by their type
func (r *ProviderRepository) GetByType(ctx context.Context, providerType gateway.ProviderType) ([]*gateway.Provider, error) {
	query := `
		SELECT id, name, type, base_url, is_enabled, default_timeout_seconds,
			   max_retries, health_check_url, supported_features, rate_limits,
			   created_at, updated_at
		FROM gateway_providers
		WHERE type = $1
		ORDER BY name
	`

	rows, err := r.db.WithContext(ctx).Raw(query, string(providerType)).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query providers by type: %w", err)
	}
	defer rows.Close()

	var providers []*gateway.Provider
	for rows.Next() {
		var provider gateway.Provider
		var typeStr string
		var featuresJSON, rateLimitsJSON string

		err := rows.Scan(
			&provider.ID,
			&provider.Name,
			&typeStr,
			&provider.BaseURL,
			&provider.IsEnabled,
			&provider.DefaultTimeoutSecs,
			&provider.MaxRetries,
			&provider.HealthCheckURL,
			&featuresJSON,
			&rateLimitsJSON,
			&provider.CreatedAt,
			&provider.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan provider row: %w", err)
		}

		// Parse enum and JSON fields
		provider.Type = gateway.ProviderType(typeStr)

		if err := json.Unmarshal([]byte(featuresJSON), &provider.SupportedFeatures); err != nil {
			return nil, fmt.Errorf("failed to unmarshal supported features: %w", err)
		}

		if err := json.Unmarshal([]byte(rateLimitsJSON), &provider.RateLimits); err != nil {
			return nil, fmt.Errorf("failed to unmarshal rate limits: %w", err)
		}

		providers = append(providers, &provider)
	}

	return providers, nil
}

// Update updates an existing provider
func (r *ProviderRepository) Update(ctx context.Context, provider *gateway.Provider) error {
	// Convert maps to JSON for storage
	featuresJSON, err := json.Marshal(provider.SupportedFeatures)
	if err != nil {
		return fmt.Errorf("failed to marshal supported features: %w", err)
	}

	rateLimitsJSON, err := json.Marshal(provider.RateLimits)
	if err != nil {
		return fmt.Errorf("failed to marshal rate limits: %w", err)
	}

	query := `
		UPDATE gateway_providers
		SET name = $2, type = $3, base_url = $4, is_enabled = $5,
			default_timeout_seconds = $6, max_retries = $7, health_check_url = $8,
			supported_features = $9, rate_limits = $10, updated_at = $11
		WHERE id = $1
	`

	provider.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Exec(query,
		provider.ID,
		provider.Name,
		string(provider.Type),
		provider.BaseURL,
		provider.IsEnabled,
		provider.DefaultTimeoutSecs,
		provider.MaxRetries,
		provider.HealthCheckURL,
		string(featuresJSON),
		string(rateLimitsJSON),
		provider.UpdatedAt,
	)

	if result.Error != nil {
		return fmt.Errorf("failed to update provider: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gateway.NewProviderNotFoundError(provider.ID.String())
	}

	return nil
}

// Delete deletes a provider by ID
func (r *ProviderRepository) Delete(ctx context.Context, id ulid.ULID) error {
	query := `DELETE FROM gateway_providers WHERE id = $1`

	result := r.db.WithContext(ctx).Exec(query, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete provider: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gateway.NewProviderNotFoundError(id.String())
	}

	return nil
}

// List retrieves providers with pagination
func (r *ProviderRepository) List(ctx context.Context, limit, offset int) ([]*gateway.Provider, error) {
	query := `
		SELECT id, name, type, base_url, is_enabled, default_timeout_seconds,
			   max_retries, health_check_url, supported_features, rate_limits,
			   created_at, updated_at
		FROM gateway_providers
		ORDER BY name
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.WithContext(ctx).Raw(query, limit, offset).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query providers: %w", err)
	}
	defer rows.Close()

	return r.scanProviders(rows)
}

// ListEnabled retrieves all enabled providers
func (r *ProviderRepository) ListEnabled(ctx context.Context) ([]*gateway.Provider, error) {
	query := `
		SELECT id, name, type, base_url, is_enabled, default_timeout_seconds,
			   max_retries, health_check_url, supported_features, rate_limits,
			   created_at, updated_at
		FROM gateway_providers
		WHERE is_enabled = true
		ORDER BY name
	`

	rows, err := r.db.WithContext(ctx).Raw(query).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query enabled providers: %w", err)
	}
	defer rows.Close()

	return r.scanProviders(rows)
}

// ListByStatus retrieves providers by enabled status
func (r *ProviderRepository) ListByStatus(ctx context.Context, isEnabled bool, limit, offset int) ([]*gateway.Provider, error) {
	query := `
		SELECT id, name, type, base_url, is_enabled, default_timeout_seconds,
			   max_retries, health_check_url, supported_features, rate_limits,
			   created_at, updated_at
		FROM gateway_providers
		WHERE is_enabled = $1
		ORDER BY name
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.WithContext(ctx).Raw(query, isEnabled, limit, offset).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query providers by status: %w", err)
	}
	defer rows.Close()

	return r.scanProviders(rows)
}

// SearchProviders searches providers with filters
func (r *ProviderRepository) SearchProviders(ctx context.Context, filter *gateway.ProviderFilter) ([]*gateway.Provider, int, error) {
	whereClause, args := r.buildWhereClause(filter)
	
	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM gateway_providers %s", whereClause)
	var totalCount int
	err := r.db.WithContext(ctx).Raw(countQuery, args...).Scan(&totalCount).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count providers: %w", err)
	}

	// Main query
	query := fmt.Sprintf(`
		SELECT id, name, type, base_url, is_enabled, default_timeout_seconds,
			   max_retries, health_check_url, supported_features, rate_limits,
			   created_at, updated_at
		FROM gateway_providers
		%s
		ORDER BY name
	`, whereClause)

	rows, err := r.db.WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search providers: %w", err)
	}
	defer rows.Close()

	providers, err := r.scanProviders(rows)
	if err != nil {
		return nil, 0, err
	}

	return providers, totalCount, nil
}

// CountProviders counts providers with filters
func (r *ProviderRepository) CountProviders(ctx context.Context, filter *gateway.ProviderFilter) (int64, error) {
	whereClause, args := r.buildWhereClause(filter)
	query := fmt.Sprintf("SELECT COUNT(*) FROM gateway_providers %s", whereClause)
	
	var count int64
	err := r.db.WithContext(ctx).Raw(query, args...).Scan(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count providers: %w", err)
	}

	return count, nil
}

// CreateBatch creates multiple providers in a transaction
func (r *ProviderRepository) CreateBatch(ctx context.Context, providers []*gateway.Provider) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, provider := range providers {
			if err := r.createWithTx(ctx, tx, provider); err != nil {
				return err
			}
		}
		return nil
	})
}

// UpdateBatch updates multiple providers in a transaction
func (r *ProviderRepository) UpdateBatch(ctx context.Context, providers []*gateway.Provider) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, provider := range providers {
			if err := r.updateWithTx(ctx, tx, provider); err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteBatch deletes multiple providers in a transaction
func (r *ProviderRepository) DeleteBatch(ctx context.Context, ids []ulid.ULID) error {
	if len(ids) == 0 {
		return nil
	}

	query := `DELETE FROM gateway_providers WHERE id = ANY($1)`
	
	// Convert ULIDs to strings for PostgreSQL array
	stringIDs := make([]string, len(ids))
	for i, id := range ids {
		stringIDs[i] = id.String()
	}

	result := r.db.WithContext(ctx).Exec(query, pq.Array(stringIDs))
	if result.Error != nil {
		return fmt.Errorf("failed to delete providers: %w", result.Error)
	}

	return nil
}

// UpdateHealthStatus updates the health status of a provider
func (r *ProviderRepository) UpdateHealthStatus(ctx context.Context, providerID ulid.ULID, status gateway.HealthStatus) error {
	query := `UPDATE gateway_providers SET updated_at = $1 WHERE id = $2`
	
	result := r.db.WithContext(ctx).Exec(query, time.Now(), providerID)
	if result.Error != nil {
		return fmt.Errorf("failed to update provider health status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gateway.NewProviderNotFoundError(providerID.String())
	}

	return nil
}

// GetHealthyProviders retrieves providers that are enabled
func (r *ProviderRepository) GetHealthyProviders(ctx context.Context) ([]*gateway.Provider, error) {
	return r.ListEnabled(ctx)
}

// Helper methods

func (r *ProviderRepository) scanProviders(rows *sql.Rows) ([]*gateway.Provider, error) {
	var providers []*gateway.Provider

	for rows.Next() {
		var provider gateway.Provider
		var typeStr string
		var featuresJSON, rateLimitsJSON string

		err := rows.Scan(
			&provider.ID,
			&provider.Name,
			&typeStr,
			&provider.BaseURL,
			&provider.IsEnabled,
			&provider.DefaultTimeoutSecs,
			&provider.MaxRetries,
			&provider.HealthCheckURL,
			&featuresJSON,
			&rateLimitsJSON,
			&provider.CreatedAt,
			&provider.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan provider row: %w", err)
		}

		// Parse enum and JSON fields
		provider.Type = gateway.ProviderType(typeStr)

		if err := json.Unmarshal([]byte(featuresJSON), &provider.SupportedFeatures); err != nil {
			return nil, fmt.Errorf("failed to unmarshal supported features: %w", err)
		}

		if err := json.Unmarshal([]byte(rateLimitsJSON), &provider.RateLimits); err != nil {
			return nil, fmt.Errorf("failed to unmarshal rate limits: %w", err)
		}

		providers = append(providers, &provider)
	}

	return providers, nil
}

func (r *ProviderRepository) buildWhereClause(filter *gateway.ProviderFilter) (string, []interface{}) {
	if filter == nil {
		return "", nil
	}

	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.ProviderType != nil {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argIndex))
		args = append(args, string(*filter.ProviderType))
		argIndex++
	}

	if filter.IsEnabled != nil {
		conditions = append(conditions, fmt.Sprintf("is_enabled = $%d", argIndex))
		args = append(args, *filter.IsEnabled)
		argIndex++
	}

	if filter.Search != nil && *filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR type ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+*filter.Search+"%")
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

func (r *ProviderRepository) createWithTx(ctx context.Context, tx *gorm.DB, provider *gateway.Provider) error {
	if provider.ID.IsZero() {
		provider.ID = ulid.New()
	}

	// Convert maps to JSON for storage
	featuresJSON, err := json.Marshal(provider.SupportedFeatures)
	if err != nil {
		return fmt.Errorf("failed to marshal supported features: %w", err)
	}

	rateLimitsJSON, err := json.Marshal(provider.RateLimits)
	if err != nil {
		return fmt.Errorf("failed to marshal rate limits: %w", err)
	}

	query := `
		INSERT INTO gateway_providers (
			id, name, type, base_url, is_enabled, default_timeout_seconds,
			max_retries, health_check_url, supported_features, rate_limits,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	now := time.Now()
	provider.CreatedAt = now
	provider.UpdatedAt = now

	err = tx.WithContext(ctx).Exec(query,
		provider.ID,
		provider.Name,
		string(provider.Type),
		provider.BaseURL,
		provider.IsEnabled,
		provider.DefaultTimeoutSecs,
		provider.MaxRetries,
		provider.HealthCheckURL,
		string(featuresJSON),
		string(rateLimitsJSON),
		provider.CreatedAt,
		provider.UpdatedAt,
	).Error

	if err != nil {
		if isDuplicateKeyError(err) {
			return gateway.NewProviderNotFoundError(provider.Name)
		}
		return fmt.Errorf("failed to create provider: %w", err)
	}

	return nil
}

func (r *ProviderRepository) updateWithTx(ctx context.Context, tx *gorm.DB, provider *gateway.Provider) error {
	// Convert maps to JSON for storage
	featuresJSON, err := json.Marshal(provider.SupportedFeatures)
	if err != nil {
		return fmt.Errorf("failed to marshal supported features: %w", err)
	}

	rateLimitsJSON, err := json.Marshal(provider.RateLimits)
	if err != nil {
		return fmt.Errorf("failed to marshal rate limits: %w", err)
	}

	query := `
		UPDATE gateway_providers
		SET name = $2, type = $3, base_url = $4, is_enabled = $5,
			default_timeout_seconds = $6, max_retries = $7, health_check_url = $8,
			supported_features = $9, rate_limits = $10, updated_at = $11
		WHERE id = $1
	`

	provider.UpdatedAt = time.Now()

	result := tx.WithContext(ctx).Exec(query,
		provider.ID,
		provider.Name,
		string(provider.Type),
		provider.BaseURL,
		provider.IsEnabled,
		provider.DefaultTimeoutSecs,
		provider.MaxRetries,
		provider.HealthCheckURL,
		string(featuresJSON),
		string(rateLimitsJSON),
		provider.UpdatedAt,
	)

	if result.Error != nil {
		return fmt.Errorf("failed to update provider: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gateway.NewProviderNotFoundError(provider.ID.String())
	}

	return nil
}

// isDuplicateKeyError checks if the error is a duplicate key constraint violation
func isDuplicateKeyError(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505" // unique_violation
	}
	return false
}