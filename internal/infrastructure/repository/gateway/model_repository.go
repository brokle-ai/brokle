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

// ModelRepository implements the gateway.ModelRepository interface
type ModelRepository struct {
	db *gorm.DB
}

// NewModelRepository creates a new model repository instance
func NewModelRepository(db *gorm.DB) *ModelRepository {
	return &ModelRepository{
		db: db,
	}
}

// Create creates a new model in the database
func (r *ModelRepository) Create(ctx context.Context, model *gateway.Model) error {
	if model.ID.IsZero() {
		model.ID = ulid.New()
	}

	// Convert capabilities map to JSON
	capabilitiesJSON, err := json.Marshal(model.Capabilities)
	if err != nil {
		return fmt.Errorf("failed to marshal capabilities: %w", err)
	}

	query := `
		INSERT INTO gateway_models (
			id, provider_id, name, type, display_name, description,
			context_length, max_tokens, input_cost_per_token,
			output_cost_per_token, is_enabled, capabilities,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	now := time.Now()
	model.CreatedAt = now
	model.UpdatedAt = now

	err = r.db.WithContext(ctx).Exec(query,
		model.ID,
		model.ProviderID,
		model.Name,
		string(model.Type),
		model.DisplayName,
		model.Description,
		model.ContextLength,
		model.MaxTokens,
		model.InputCostPerToken,
		model.OutputCostPerToken,
		model.IsEnabled,
		string(capabilitiesJSON),
		model.CreatedAt,
		model.UpdatedAt,
	).Error

	if err != nil {
		if isDuplicateKeyError(err) {
			return gateway.NewModelNotFoundError(model.Name)
		}
		return fmt.Errorf("failed to create model: %w", err)
	}

	return nil
}

// GetByID retrieves a model by its ID
func (r *ModelRepository) GetByID(ctx context.Context, id ulid.ULID) (*gateway.Model, error) {
	var model gateway.Model
	var modelType string
	var capabilitiesJSON string

	query := `
		SELECT id, provider_id, name, type, display_name, description,
			   context_length, max_tokens, input_cost_per_token,
			   output_cost_per_token, is_enabled, capabilities,
			   created_at, updated_at
		FROM gateway_models
		WHERE id = $1
	`

	row := r.db.WithContext(ctx).Raw(query, id).Row()

	err := row.Scan(
		&model.ID,
		&model.ProviderID,
		&model.Name,
		&modelType,
		&model.DisplayName,
		&model.Description,
		&model.ContextLength,
		&model.MaxTokens,
		&model.InputCostPerToken,
		&model.OutputCostPerToken,
		&model.IsEnabled,
		&capabilitiesJSON,
		&model.CreatedAt,
		&model.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, gateway.NewModelNotFoundError(id.String())
		}
		return nil, fmt.Errorf("failed to get model by ID: %w", err)
	}

	// Parse enum and JSON fields
	model.Type = gateway.ModelType(modelType)

	if err := json.Unmarshal([]byte(capabilitiesJSON), &model.Capabilities); err != nil {
		return nil, fmt.Errorf("failed to unmarshal capabilities: %w", err)
	}

	return &model, nil
}

// GetByName retrieves a model by its name
func (r *ModelRepository) GetByName(ctx context.Context, name string) (*gateway.Model, error) {
	var model gateway.Model
	var modelType string
	var capabilitiesJSON string

	query := `
		SELECT id, provider_id, name, type, display_name, description,
			   context_length, max_tokens, input_cost_per_token,
			   output_cost_per_token, is_enabled, capabilities,
			   created_at, updated_at
		FROM gateway_models
		WHERE name = $1
	`

	row := r.db.WithContext(ctx).Raw(query, name).Row()

	err := row.Scan(
		&model.ID,
		&model.ProviderID,
		&model.Name,
		&modelType,
		&model.DisplayName,
		&model.Description,
		&model.ContextLength,
		&model.MaxTokens,
		&model.InputCostPerToken,
		&model.OutputCostPerToken,
		&model.IsEnabled,
		&capabilitiesJSON,
		&model.CreatedAt,
		&model.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, gateway.NewModelNotFoundError(name)
		}
		return nil, fmt.Errorf("failed to get model by name: %w", err)
	}

	// Parse enum and JSON fields
	model.Type = gateway.ModelType(modelType)

	if err := json.Unmarshal([]byte(capabilitiesJSON), &model.Capabilities); err != nil {
		return nil, fmt.Errorf("failed to unmarshal capabilities: %w", err)
	}

	return &model, nil
}

// GetByProviderAndName retrieves a model by provider ID and name
func (r *ModelRepository) GetByProviderAndName(ctx context.Context, providerID ulid.ULID, name string) (*gateway.Model, error) {
	var model gateway.Model
	var modelType string
	var capabilitiesJSON string

	query := `
		SELECT id, provider_id, name, type, display_name, description,
			   context_length, max_tokens, input_cost_per_token,
			   output_cost_per_token, is_enabled, capabilities,
			   created_at, updated_at
		FROM gateway_models
		WHERE provider_id = $1 AND name = $2
	`

	row := r.db.WithContext(ctx).Raw(query, providerID, name).Row()

	err := row.Scan(
		&model.ID,
		&model.ProviderID,
		&model.Name,
		&modelType,
		&model.DisplayName,
		&model.Description,
		&model.ContextLength,
		&model.MaxTokens,
		&model.InputCostPerToken,
		&model.OutputCostPerToken,
		&model.IsEnabled,
		&capabilitiesJSON,
		&model.CreatedAt,
		&model.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, gateway.NewModelNotFoundError(fmt.Sprintf("provider:%s,model:%s", providerID, name))
		}
		return nil, fmt.Errorf("failed to get model by provider and name: %w", err)
	}

	// Parse enum and JSON fields
	model.Type = gateway.ModelType(modelType)

	if err := json.Unmarshal([]byte(capabilitiesJSON), &model.Capabilities); err != nil {
		return nil, fmt.Errorf("failed to unmarshal capabilities: %w", err)
	}

	return &model, nil
}

// GetByProvider retrieves models by provider ID
func (r *ModelRepository) GetByProvider(ctx context.Context, providerID ulid.ULID) ([]*gateway.Model, error) {
	query := `
		SELECT id, provider_id, name, type, display_name, description,
			   context_length, max_tokens, input_cost_per_token,
			   output_cost_per_token, is_enabled, capabilities,
			   created_at, updated_at
		FROM gateway_models
		WHERE provider_id = $1
		ORDER BY name
	`

	rows, err := r.db.WithContext(ctx).Raw(query, providerID).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query models by provider: %w", err)
	}
	defer rows.Close()

	return r.scanModels(rows)
}

// GetByType retrieves models by type
func (r *ModelRepository) GetByType(ctx context.Context, modelType gateway.ModelType) ([]*gateway.Model, error) {
	query := `
		SELECT id, provider_id, name, type, display_name, description,
			   context_length, max_tokens, input_cost_per_token,
			   output_cost_per_token, is_enabled, capabilities,
			   created_at, updated_at
		FROM gateway_models
		WHERE type = $1
		ORDER BY name
	`

	rows, err := r.db.WithContext(ctx).Raw(query, string(modelType)).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query models by type: %w", err)
	}
	defer rows.Close()

	return r.scanModels(rows)
}

// Update updates an existing model
func (r *ModelRepository) Update(ctx context.Context, model *gateway.Model) error {
	// Convert capabilities map to JSON
	capabilitiesJSON, err := json.Marshal(model.Capabilities)
	if err != nil {
		return fmt.Errorf("failed to marshal capabilities: %w", err)
	}

	query := `
		UPDATE gateway_models
		SET provider_id = $2, name = $3, type = $4, display_name = $5,
			description = $6, context_length = $7, max_tokens = $8,
			input_cost_per_token = $9, output_cost_per_token = $10,
			is_enabled = $11, capabilities = $12, updated_at = $13
		WHERE id = $1
	`

	model.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Exec(query,
		model.ID,
		model.ProviderID,
		model.Name,
		string(model.Type),
		model.DisplayName,
		model.Description,
		model.ContextLength,
		model.MaxTokens,
		model.InputCostPerToken,
		model.OutputCostPerToken,
		model.IsEnabled,
		string(capabilitiesJSON),
		model.UpdatedAt,
	)

	if result.Error != nil {
		return fmt.Errorf("failed to update model: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gateway.NewModelNotFoundError(model.ID.String())
	}

	return nil
}

// Delete deletes a model by ID
func (r *ModelRepository) Delete(ctx context.Context, id ulid.ULID) error {
	query := `DELETE FROM gateway_models WHERE id = $1`

	result := r.db.WithContext(ctx).Exec(query, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete model: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gateway.NewModelNotFoundError(id.String())
	}

	return nil
}

// List retrieves models with pagination
func (r *ModelRepository) List(ctx context.Context, limit, offset int) ([]*gateway.Model, error) {
	query := `
		SELECT id, provider_id, name, type, display_name, description,
			   context_length, max_tokens, input_cost_per_token,
			   output_cost_per_token, is_enabled, capabilities,
			   created_at, updated_at
		FROM gateway_models
		ORDER BY name
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.WithContext(ctx).Raw(query, limit, offset).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query models: %w", err)
	}
	defer rows.Close()

	return r.scanModels(rows)
}

// ListEnabled retrieves all enabled models
func (r *ModelRepository) ListEnabled(ctx context.Context) ([]*gateway.Model, error) {
	query := `
		SELECT id, provider_id, name, type, display_name, description,
			   context_length, max_tokens, input_cost_per_token,
			   output_cost_per_token, is_enabled, capabilities,
			   created_at, updated_at
		FROM gateway_models
		WHERE is_enabled = true
		ORDER BY name
	`

	rows, err := r.db.WithContext(ctx).Raw(query).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query enabled models: %w", err)
	}
	defer rows.Close()

	return r.scanModels(rows)
}

// ListByProviderAndStatus retrieves models by provider and enabled status
func (r *ModelRepository) ListByProviderAndStatus(ctx context.Context, providerID ulid.ULID, isEnabled bool) ([]*gateway.Model, error) {
	query := `
		SELECT id, provider_id, name, type, display_name, description,
			   context_length, max_tokens, input_cost_per_token,
			   output_cost_per_token, is_enabled, capabilities,
			   created_at, updated_at
		FROM gateway_models
		WHERE provider_id = $1 AND is_enabled = $2
		ORDER BY name
	`

	rows, err := r.db.WithContext(ctx).Raw(query, providerID, isEnabled).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query models by provider and status: %w", err)
	}
	defer rows.Close()

	return r.scanModels(rows)
}

// SearchModels searches models with filters
func (r *ModelRepository) SearchModels(ctx context.Context, filter *gateway.ModelFilter) ([]*gateway.Model, int, error) {
	whereClause, args := r.buildWhereClause(filter)
	
	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM gateway_models %s", whereClause)
	var totalCount int
	err := r.db.WithContext(ctx).Raw(countQuery, args...).Scan(&totalCount).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count models: %w", err)
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
		SELECT id, provider_id, name, type, display_name, description,
			   context_length, max_tokens, input_cost_per_token,
			   output_cost_per_token, is_enabled, capabilities,
			   created_at, updated_at
		FROM gateway_models
		%s
		ORDER BY name%s
	`, whereClause, limitClause)

	rows, err := r.db.WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search models: %w", err)
	}
	defer rows.Close()

	models, err := r.scanModels(rows)
	if err != nil {
		return nil, 0, err
	}

	return models, totalCount, nil
}

// CountModels counts models with filters
func (r *ModelRepository) CountModels(ctx context.Context, filter *gateway.ModelFilter) (int64, error) {
	whereClause, args := r.buildWhereClause(filter)
	query := fmt.Sprintf("SELECT COUNT(*) FROM gateway_models %s", whereClause)
	
	var count int64
	err := r.db.WithContext(ctx).Raw(query, args...).Scan(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count models: %w", err)
	}

	return count, nil
}

// CreateBatch creates multiple models in a transaction
func (r *ModelRepository) CreateBatch(ctx context.Context, models []*gateway.Model) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, model := range models {
			if err := r.createWithTx(ctx, tx, model); err != nil {
				return err
			}
		}
		return nil
	})
}

// UpdateBatch updates multiple models in a transaction
func (r *ModelRepository) UpdateBatch(ctx context.Context, models []*gateway.Model) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, model := range models {
			if err := r.updateWithTx(ctx, tx, model); err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteBatch deletes multiple models in a transaction
func (r *ModelRepository) DeleteBatch(ctx context.Context, ids []ulid.ULID) error {
	if len(ids) == 0 {
		return nil
	}

	query := `DELETE FROM gateway_models WHERE id = ANY($1)`
	
	// Convert ULIDs to strings for PostgreSQL array
	stringIDs := make([]string, len(ids))
	for i, id := range ids {
		stringIDs[i] = id.String()
	}

	result := r.db.WithContext(ctx).Exec(query, pq.Array(stringIDs))
	if result.Error != nil {
		return fmt.Errorf("failed to delete models: %w", result.Error)
	}

	return nil
}

// SyncProviderModels synchronizes models for a provider (used for model discovery)
func (r *ModelRepository) SyncProviderModels(ctx context.Context, providerID ulid.ULID, models []*gateway.Model) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete existing models for the provider
		deleteQuery := `DELETE FROM gateway_models WHERE provider_id = $1`
		if err := tx.WithContext(ctx).Exec(deleteQuery, providerID).Error; err != nil {
			return fmt.Errorf("failed to delete existing models: %w", err)
		}

		// Insert new models
		for _, model := range models {
			model.ProviderID = providerID // Ensure provider ID is set
			if err := r.createWithTx(ctx, tx, model); err != nil {
				return err
			}
		}

		return nil
	})
}

// GetModelsWithCapability retrieves models that have a specific capability
func (r *ModelRepository) GetModelsWithCapability(ctx context.Context, capability string) ([]*gateway.Model, error) {
	query := `
		SELECT id, provider_id, name, type, display_name, description,
			   context_length, max_tokens, input_cost_per_token,
			   output_cost_per_token, is_enabled, capabilities,
			   created_at, updated_at
		FROM gateway_models
		WHERE capabilities ? $1 AND is_enabled = true
		ORDER BY name
	`

	rows, err := r.db.WithContext(ctx).Raw(query, capability).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query models by capability: %w", err)
	}
	defer rows.Close()

	return r.scanModels(rows)
}

// Helper methods

func (r *ModelRepository) scanModels(rows *sql.Rows) ([]*gateway.Model, error) {
	var models []*gateway.Model

	for rows.Next() {
		var model gateway.Model
		var typeStr string
		var capabilitiesJSON string

		err := rows.Scan(
			&model.ID,
			&model.ProviderID,
			&model.Name,
			&typeStr,
			&model.DisplayName,
			&model.Description,
			&model.ContextLength,
			&model.MaxTokens,
			&model.InputCostPerToken,
			&model.OutputCostPerToken,
			&model.IsEnabled,
			&capabilitiesJSON,
			&model.CreatedAt,
			&model.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan model row: %w", err)
		}

		// Parse enum and JSON fields
		model.Type = gateway.ModelType(typeStr)

		if err := json.Unmarshal([]byte(capabilitiesJSON), &model.Capabilities); err != nil {
			return nil, fmt.Errorf("failed to unmarshal capabilities: %w", err)
		}

		models = append(models, &model)
	}

	return models, nil
}

func (r *ModelRepository) buildWhereClause(filter *gateway.ModelFilter) (string, []interface{}) {
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

	if filter.ModelType != nil {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argIndex))
		args = append(args, string(*filter.ModelType))
		argIndex++
	}

	if filter.IsEnabled != nil {
		conditions = append(conditions, fmt.Sprintf("is_enabled = $%d", argIndex))
		args = append(args, *filter.IsEnabled)
		argIndex++
	}

	if filter.Search != nil && *filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR display_name ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex, argIndex))
		args = append(args, "%"+*filter.Search+"%")
		argIndex++
	}

	if filter.MinContextLength != nil {
		conditions = append(conditions, fmt.Sprintf("context_length >= $%d", argIndex))
		args = append(args, *filter.MinContextLength)
		argIndex++
	}

	if filter.MaxContextLength != nil {
		conditions = append(conditions, fmt.Sprintf("context_length <= $%d", argIndex))
		args = append(args, *filter.MaxContextLength)
		argIndex++
	}

	if filter.MaxInputCostPerToken != nil {
		conditions = append(conditions, fmt.Sprintf("input_cost_per_token <= $%d", argIndex))
		args = append(args, *filter.MaxInputCostPerToken)
		argIndex++
	}

	if filter.Capability != nil && *filter.Capability != "" {
		conditions = append(conditions, fmt.Sprintf("capabilities ? $%d", argIndex))
		args = append(args, *filter.Capability)
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

func (r *ModelRepository) createWithTx(ctx context.Context, tx *gorm.DB, model *gateway.Model) error {
	if model.ID.IsZero() {
		model.ID = ulid.New()
	}

	// Convert capabilities map to JSON
	capabilitiesJSON, err := json.Marshal(model.Capabilities)
	if err != nil {
		return fmt.Errorf("failed to marshal capabilities: %w", err)
	}

	query := `
		INSERT INTO gateway_models (
			id, provider_id, name, type, display_name, description,
			context_length, max_tokens, input_cost_per_token,
			output_cost_per_token, is_enabled, capabilities,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	now := time.Now()
	model.CreatedAt = now
	model.UpdatedAt = now

	err = tx.WithContext(ctx).Exec(query,
		model.ID,
		model.ProviderID,
		model.Name,
		string(model.Type),
		model.DisplayName,
		model.Description,
		model.ContextLength,
		model.MaxTokens,
		model.InputCostPerToken,
		model.OutputCostPerToken,
		model.IsEnabled,
		string(capabilitiesJSON),
		model.CreatedAt,
		model.UpdatedAt,
	).Error

	if err != nil {
		if isDuplicateKeyError(err) {
			return gateway.NewModelNotFoundError(model.Name)
		}
		return fmt.Errorf("failed to create model: %w", err)
	}

	return nil
}

func (r *ModelRepository) updateWithTx(ctx context.Context, tx *gorm.DB, model *gateway.Model) error {
	// Convert capabilities map to JSON
	capabilitiesJSON, err := json.Marshal(model.Capabilities)
	if err != nil {
		return fmt.Errorf("failed to marshal capabilities: %w", err)
	}

	query := `
		UPDATE gateway_models
		SET provider_id = $2, name = $3, type = $4, display_name = $5,
			description = $6, context_length = $7, max_tokens = $8,
			input_cost_per_token = $9, output_cost_per_token = $10,
			is_enabled = $11, capabilities = $12, updated_at = $13
		WHERE id = $1
	`

	model.UpdatedAt = time.Now()

	result := tx.WithContext(ctx).Exec(query,
		model.ID,
		model.ProviderID,
		model.Name,
		string(model.Type),
		model.DisplayName,
		model.Description,
		model.ContextLength,
		model.MaxTokens,
		model.InputCostPerToken,
		model.OutputCostPerToken,
		model.IsEnabled,
		string(capabilitiesJSON),
		model.UpdatedAt,
	)

	if result.Error != nil {
		return fmt.Errorf("failed to update model: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gateway.NewModelNotFoundError(model.ID.String())
	}

	return nil
}