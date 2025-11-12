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

	// Convert metadata map to JSON
	metadataJSON, err := json.Marshal(model.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO gateway_models (
			id, provider_id, model_name, display_name, model_type,
			input_cost_per_1k_tokens, output_cost_per_1k_tokens, max_context_tokens,
			supports_streaming, supports_functions, supports_vision,
			quality_score, speed_score, metadata, is_enabled,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`

	now := time.Now()
	model.CreatedAt = now
	model.UpdatedAt = now

	err = r.db.WithContext(ctx).Exec(query,
		model.ID,
		model.ProviderID,
		model.ModelName,
		model.DisplayName,
		string(model.ModelType),
		model.InputCostPer1kTokens,
		model.OutputCostPer1kTokens,
		model.MaxContextTokens,
		model.SupportsStreaming,
		model.SupportsFunctions,
		model.SupportsVision,
		model.QualityScore,
		model.SpeedScore,
		string(metadataJSON),
		model.IsEnabled,
		model.CreatedAt,
		model.UpdatedAt,
	).Error

	if err != nil {
		if isDuplicateKeyError(err) {
			return gateway.NewModelNotFoundError(model.ModelName)
		}
		return fmt.Errorf("failed to create model: %w", err)
	}

	return nil
}

// GetByID retrieves a model by its ID
func (r *ModelRepository) GetByID(ctx context.Context, id ulid.ULID) (*gateway.Model, error) {
	var model gateway.Model
	var modelType string
	var metadataJSON string

	query := `
		SELECT id, provider_id, model_name, display_name, model_type,
			   input_cost_per_1k_tokens, output_cost_per_1k_tokens, max_context_tokens,
			   supports_streaming, supports_functions, supports_vision,
			   quality_score, speed_score, metadata, is_enabled,
			   created_at, updated_at
		FROM gateway_models
		WHERE id = $1
	`

	row := r.db.WithContext(ctx).Raw(query, id).Row()

	err := row.Scan(
		&model.ID,
		&model.ProviderID,
		&model.ModelName,
		&model.DisplayName,
		&modelType,
		&model.InputCostPer1kTokens,
		&model.OutputCostPer1kTokens,
		&model.MaxContextTokens,
		&model.SupportsStreaming,
		&model.SupportsFunctions,
		&model.SupportsVision,
		&model.QualityScore,
		&model.SpeedScore,
		&metadataJSON,
		&model.IsEnabled,
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
	model.ModelType = gateway.ModelType(modelType)

	if err := json.Unmarshal([]byte(metadataJSON), &model.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &model, nil
}

// GetByName retrieves a model by its name
func (r *ModelRepository) GetByName(ctx context.Context, name string) (*gateway.Model, error) {
	var model gateway.Model
	var modelType string
	var metadataJSON string

	query := `
		SELECT id, provider_id, model_name, display_name, model_type,
			   input_cost_per_1k_tokens, output_cost_per_1k_tokens, max_context_tokens,
			   supports_streaming, supports_functions, supports_vision,
			   quality_score, speed_score, metadata, is_enabled,
			   created_at, updated_at
		FROM gateway_models
		WHERE model_name = $1
	`

	row := r.db.WithContext(ctx).Raw(query, name).Row()

	err := row.Scan(
		&model.ID,
		&model.ProviderID,
		&model.ModelName,
		&model.DisplayName,
		&modelType,
		&model.InputCostPer1kTokens,
		&model.OutputCostPer1kTokens,
		&model.MaxContextTokens,
		&model.SupportsStreaming,
		&model.SupportsFunctions,
		&model.SupportsVision,
		&model.QualityScore,
		&model.SpeedScore,
		&metadataJSON,
		&model.IsEnabled,
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
	model.ModelType = gateway.ModelType(modelType)

	if err := json.Unmarshal([]byte(metadataJSON), &model.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &model, nil
}

// GetByProviderAndName retrieves a model by provider ID and name
func (r *ModelRepository) GetByProviderAndName(ctx context.Context, providerID ulid.ULID, name string) (*gateway.Model, error) {
	var model gateway.Model
	var modelType string
	var metadataJSON string

	query := `
		SELECT id, provider_id, model_name, display_name, model_type,
			   input_cost_per_1k_tokens, output_cost_per_1k_tokens, max_context_tokens,
			   supports_streaming, supports_functions, supports_vision,
			   quality_score, speed_score, metadata, is_enabled,
			   created_at, updated_at
		FROM gateway_models
		WHERE provider_id = $1 AND model_name = $2
	`

	row := r.db.WithContext(ctx).Raw(query, providerID, name).Row()

	err := row.Scan(
		&model.ID,
		&model.ProviderID,
		&model.ModelName,
		&model.DisplayName,
		&modelType,
		&model.InputCostPer1kTokens,
		&model.OutputCostPer1kTokens,
		&model.MaxContextTokens,
		&model.SupportsStreaming,
		&model.SupportsFunctions,
		&model.SupportsVision,
		&model.QualityScore,
		&model.SpeedScore,
		&metadataJSON,
		&model.IsEnabled,
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
	model.ModelType = gateway.ModelType(modelType)

	if err := json.Unmarshal([]byte(metadataJSON), &model.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &model, nil
}

// GetByProvider retrieves models by provider ID
func (r *ModelRepository) GetByProvider(ctx context.Context, providerID ulid.ULID) ([]*gateway.Model, error) {
	query := `
		SELECT id, provider_id, model_name, display_name, model_type,
			   input_cost_per_1k_tokens, output_cost_per_1k_tokens, max_context_tokens,
			   supports_streaming, supports_functions, supports_vision,
			   quality_score, speed_score, metadata, is_enabled,
			   created_at, updated_at
		FROM gateway_models
		WHERE provider_id = $1
		ORDER BY model_name
	`

	rows, err := r.db.WithContext(ctx).Raw(query, providerID).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query models by provider: %w", err)
	}
	defer rows.Close()

	return r.scanModels(rows)
}

// GetEnabledByProviderID retrieves enabled models for a specific provider
func (r *ModelRepository) GetEnabledByProviderID(ctx context.Context, providerID ulid.ULID) ([]*gateway.Model, error) {
	query := `
		SELECT id, provider_id, model_name, display_name, model_type,
			   input_cost_per_1k_tokens, output_cost_per_1k_tokens, max_context_tokens,
			   supports_streaming, supports_functions, supports_vision,
			   quality_score, speed_score, metadata, is_enabled,
			   created_at, updated_at
		FROM gateway_models
		WHERE provider_id = $1 AND is_enabled = true
		ORDER BY model_name ASC
	`

	rows, err := r.db.WithContext(ctx).Raw(query, providerID).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query enabled models by provider ID: %w", err)
	}
	defer rows.Close()

	return r.scanModels(rows)
}

// GetByType retrieves models by type
func (r *ModelRepository) GetByType(ctx context.Context, modelType gateway.ModelType) ([]*gateway.Model, error) {
	query := `
		SELECT id, provider_id, model_name, display_name, model_type,
			   input_cost_per_1k_tokens, output_cost_per_1k_tokens, max_context_tokens,
			   supports_streaming, supports_functions, supports_vision,
			   quality_score, speed_score, metadata, is_enabled,
			   created_at, updated_at
		FROM gateway_models
		WHERE model_type = $1
		ORDER BY model_name
	`

	rows, err := r.db.WithContext(ctx).Raw(query, string(modelType)).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query models by type: %w", err)
	}
	defer rows.Close()

	return r.scanModels(rows)
}

// GetByModelName retrieves a model by its model name
func (r *ModelRepository) GetByModelName(ctx context.Context, modelName string) (*gateway.Model, error) {
	// This is the same as GetByName - delegate to that method
	return r.GetByName(ctx, modelName)
}

// GetByModelType retrieves models by their model type with limit and offset
func (r *ModelRepository) GetByModelType(ctx context.Context, modelType gateway.ModelType, limit, offset int) ([]*gateway.Model, error) {
	// Delegate to the existing GetByType method for now
	return r.GetByType(ctx, modelType)
}

// GetByProviderAndModel retrieves a model by provider ID and model name
func (r *ModelRepository) GetByProviderAndModel(ctx context.Context, providerID ulid.ULID, modelName string) (*gateway.Model, error) {
	// Delegate to the existing GetByProviderAndName method
	return r.GetByProviderAndName(ctx, providerID, modelName)
}

// Update updates an existing model
func (r *ModelRepository) Update(ctx context.Context, model *gateway.Model) error {
	// Convert metadata map to JSON
	metadataJSON, err := json.Marshal(model.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		UPDATE gateway_models
		SET provider_id = $2, model_name = $3, model_type = $4, display_name = $5,
			input_cost_per_1k_tokens = $6, output_cost_per_1k_tokens = $7,
			max_context_tokens = $8, supports_streaming = $9, supports_functions = $10,
			supports_vision = $11, quality_score = $12, speed_score = $13,
			metadata = $14, is_enabled = $15, updated_at = $16
		WHERE id = $1
	`

	model.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Exec(query,
		model.ID,
		model.ProviderID,
		model.ModelName,
		string(model.ModelType),
		model.DisplayName,
		model.InputCostPer1kTokens,
		model.OutputCostPer1kTokens,
		model.MaxContextTokens,
		model.SupportsStreaming,
		model.SupportsFunctions,
		model.SupportsVision,
		model.QualityScore,
		model.SpeedScore,
		string(metadataJSON),
		model.IsEnabled,
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
		SELECT id, provider_id, model_name, display_name, model_type,
			   input_cost_per_1k_tokens, output_cost_per_1k_tokens, max_context_tokens,
			   supports_streaming, supports_functions, supports_vision,
			   quality_score, speed_score, metadata, is_enabled,
			   created_at, updated_at
		FROM gateway_models
		ORDER BY model_name
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
		SELECT id, provider_id, model_name, display_name, model_type,
			   input_cost_per_1k_tokens, output_cost_per_1k_tokens, max_context_tokens,
			   supports_streaming, supports_functions, supports_vision,
			   quality_score, speed_score, metadata, is_enabled,
			   created_at, updated_at
		FROM gateway_models
		WHERE is_enabled = true
		ORDER BY model_name
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
		SELECT id, provider_id, model_name, display_name, model_type,
			   input_cost_per_1k_tokens, output_cost_per_1k_tokens, max_context_tokens,
			   supports_streaming, supports_functions, supports_vision,
			   quality_score, speed_score, metadata, is_enabled,
			   created_at, updated_at
		FROM gateway_models
		WHERE provider_id = $1 AND is_enabled = $2
		ORDER BY model_name
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
	// ModelFilter doesn't have Limit/Offset fields, they're passed as separate parameters
	// This method might need to be updated to match the interface signature

	query := fmt.Sprintf(`
		SELECT id, provider_id, model_name, display_name, model_type,
			   input_cost_per_1k_tokens, output_cost_per_1k_tokens, max_context_tokens,
			   supports_streaming, supports_functions, supports_vision,
			   quality_score, speed_score, metadata, is_enabled,
			   created_at, updated_at
		FROM gateway_models
		%s
		ORDER BY model_name%s
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
		SELECT id, provider_id, model_name, display_name, model_type,
			   input_cost_per_1k_tokens, output_cost_per_1k_tokens, max_context_tokens,
			   supports_streaming, supports_functions, supports_vision,
			   quality_score, speed_score, metadata, is_enabled,
			   created_at, updated_at
		FROM gateway_models
		WHERE metadata ? $1 AND is_enabled = true
		ORDER BY model_name
	`

	rows, err := r.db.WithContext(ctx).Raw(query, capability).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query models by capability: %w", err)
	}
	defer rows.Close()

	return r.scanModels(rows)
}

// GetAvailableModelsForProject retrieves models available for a specific project
func (r *ModelRepository) GetAvailableModelsForProject(ctx context.Context, projectID ulid.ULID) ([]*gateway.Model, error) {
	// For now, return all enabled models - in a real implementation,
	// this would join with provider configs to check which providers
	// are configured for this project
	query := `
		SELECT id, provider_id, model_name, display_name, model_type,
			   input_cost_per_1k_tokens, output_cost_per_1k_tokens, max_context_tokens,
			   supports_streaming, supports_functions, supports_vision,
			   quality_score, speed_score, metadata, is_enabled,
			   created_at, updated_at
		FROM gateway_models
		WHERE is_enabled = true
		ORDER BY model_name
	`

	rows, err := r.db.WithContext(ctx).Raw(query).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query models for project: %w", err)
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
		var metadataJSON string

		err := rows.Scan(
			&model.ID,
			&model.ProviderID,
			&model.ModelName,
			&model.DisplayName,
			&typeStr,
			&model.InputCostPer1kTokens,
			&model.OutputCostPer1kTokens,
			&model.MaxContextTokens,
			&model.SupportsStreaming,
			&model.SupportsFunctions,
			&model.SupportsVision,
			&model.QualityScore,
			&model.SpeedScore,
			&metadataJSON,
			&model.IsEnabled,
			&model.CreatedAt,
			&model.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan model row: %w", err)
		}

		// Parse enum and JSON fields
		model.ModelType = gateway.ModelType(typeStr)

		if err := json.Unmarshal([]byte(metadataJSON), &model.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
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
		conditions = append(conditions, fmt.Sprintf("model_type = $%d", argIndex))
		args = append(args, string(*filter.ModelType))
		argIndex++
	}

	if filter.IsEnabled != nil {
		conditions = append(conditions, fmt.Sprintf("is_enabled = $%d", argIndex))
		args = append(args, *filter.IsEnabled)
		argIndex++
	}

	if filter.Search != nil && *filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(model_name ILIKE $%d OR display_name ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+*filter.Search+"%")
		argIndex++
	}

	if filter.MinContextTokens != nil {
		conditions = append(conditions, fmt.Sprintf("max_context_tokens >= $%d", argIndex))
		args = append(args, *filter.MinContextTokens)
		argIndex++
	}

	if filter.MaxContextTokens != nil {
		conditions = append(conditions, fmt.Sprintf("max_context_tokens <= $%d", argIndex))
		args = append(args, *filter.MaxContextTokens)
		argIndex++
	}

	if filter.MinCostPer1k != nil {
		conditions = append(conditions, fmt.Sprintf("input_cost_per_1k_tokens >= $%d", argIndex))
		args = append(args, *filter.MinCostPer1k)
		argIndex++
	}

	if filter.MaxCostPer1k != nil {
		conditions = append(conditions, fmt.Sprintf("input_cost_per_1k_tokens <= $%d", argIndex))
		args = append(args, *filter.MaxCostPer1k)
		argIndex++
	}

	// CreatedAfter and CreatedBefore fields are not in the ModelFilter struct
	// They can be added if needed in the future

	if len(conditions) == 0 {
		return "", args
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}

func (r *ModelRepository) createWithTx(ctx context.Context, tx *gorm.DB, model *gateway.Model) error {
	if model.ID.IsZero() {
		model.ID = ulid.New()
	}

	// Convert metadata map to JSON
	metadataJSON, err := json.Marshal(model.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO gateway_models (
			id, provider_id, model_name, display_name, model_type,
			input_cost_per_1k_tokens, output_cost_per_1k_tokens, max_context_tokens,
			supports_streaming, supports_functions, supports_vision,
			quality_score, speed_score, metadata, is_enabled,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`

	now := time.Now()
	model.CreatedAt = now
	model.UpdatedAt = now

	err = tx.WithContext(ctx).Exec(query,
		model.ID,
		model.ProviderID,
		model.ModelName,
		model.DisplayName,
		string(model.ModelType),
		model.InputCostPer1kTokens,
		model.OutputCostPer1kTokens,
		model.MaxContextTokens,
		model.SupportsStreaming,
		model.SupportsFunctions,
		model.SupportsVision,
		model.QualityScore,
		model.SpeedScore,
		string(metadataJSON),
		model.IsEnabled,
		model.CreatedAt,
		model.UpdatedAt,
	).Error

	if err != nil {
		if isDuplicateKeyError(err) {
			return gateway.NewModelNotFoundError(model.ModelName)
		}
		return fmt.Errorf("failed to create model: %w", err)
	}

	return nil
}

func (r *ModelRepository) updateWithTx(ctx context.Context, tx *gorm.DB, model *gateway.Model) error {
	// Convert metadata map to JSON
	metadataJSON, err := json.Marshal(model.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		UPDATE gateway_models
		SET provider_id = $2, model_name = $3, model_type = $4, display_name = $5,
			input_cost_per_1k_tokens = $6, output_cost_per_1k_tokens = $7,
			max_context_tokens = $8, supports_streaming = $9, supports_functions = $10,
			supports_vision = $11, quality_score = $12, speed_score = $13,
			metadata = $14, is_enabled = $15, updated_at = $16
		WHERE id = $1
	`

	model.UpdatedAt = time.Now()

	result := tx.WithContext(ctx).Exec(query,
		model.ID,
		model.ProviderID,
		model.ModelName,
		string(model.ModelType),
		model.DisplayName,
		model.InputCostPer1kTokens,
		model.OutputCostPer1kTokens,
		model.MaxContextTokens,
		model.SupportsStreaming,
		model.SupportsFunctions,
		model.SupportsVision,
		model.QualityScore,
		model.SpeedScore,
		string(metadataJSON),
		model.IsEnabled,
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

// GetByProviderID retrieves all models for a specific provider
func (r *ModelRepository) GetByProviderID(ctx context.Context, providerID ulid.ULID) ([]*gateway.Model, error) {
	query := `
		SELECT id, provider_id, model_name, display_name, model_type,
			   input_cost_per_1k_tokens, output_cost_per_1k_tokens, max_context_tokens,
			   supports_streaming, supports_functions, supports_vision,
			   quality_score, speed_score, metadata, is_enabled,
			   created_at, updated_at
		FROM gateway_models
		WHERE provider_id = $1
		ORDER BY model_name ASC
	`

	rows, err := r.db.WithContext(ctx).Raw(query, providerID).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query models by provider ID: %w", err)
	}
	defer rows.Close()

	return r.scanModels(rows)
}

// GetCheapestModels retrieves the cheapest models for a given model type
func (r *ModelRepository) GetCheapestModels(ctx context.Context, modelType gateway.ModelType, limit int) ([]*gateway.Model, error) {
	query := `
		SELECT id, provider_id, model_name, display_name, model_type,
			   input_cost_per_1k_tokens, output_cost_per_1k_tokens, max_context_tokens,
			   supports_streaming, supports_functions, supports_vision,
			   quality_score, speed_score, metadata, is_enabled,
			   created_at, updated_at
		FROM gateway_models
		WHERE model_type = $1 AND is_enabled = true
		ORDER BY (input_cost_per_1k_tokens + output_cost_per_1k_tokens) ASC
		LIMIT $2
	`

	rows, err := r.db.WithContext(ctx).Raw(query, string(modelType), limit).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query cheapest models: %w", err)
	}
	defer rows.Close()

	return r.scanModels(rows)
}

// GetFastestModels retrieves the fastest models for a given model type
func (r *ModelRepository) GetFastestModels(ctx context.Context, modelType gateway.ModelType, limit int) ([]*gateway.Model, error) {
	query := `
		SELECT id, provider_id, model_name, display_name, model_type,
			   input_cost_per_1k_tokens, output_cost_per_1k_tokens, max_context_tokens,
			   supports_streaming, supports_functions, supports_vision,
			   quality_score, speed_score, metadata, is_enabled,
			   created_at, updated_at
		FROM gateway_models
		WHERE model_type = $1 AND is_enabled = true
		ORDER BY speed_score DESC
		LIMIT $2
	`

	rows, err := r.db.WithContext(ctx).Raw(query, string(modelType), limit).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query fastest models: %w", err)
	}
	defer rows.Close()

	return r.scanModels(rows)
}

// GetFunctionModels retrieves models that support function calling
func (r *ModelRepository) GetFunctionModels(ctx context.Context) ([]*gateway.Model, error) {
	query := `
		SELECT id, provider_id, model_name, display_name, model_type,
			   input_cost_per_1k_tokens, output_cost_per_1k_tokens, max_context_tokens,
			   supports_streaming, supports_functions, supports_vision,
			   quality_score, speed_score, metadata, is_enabled,
			   created_at, updated_at
		FROM gateway_models
		WHERE supports_functions = true AND is_enabled = true
		ORDER BY quality_score DESC, model_name ASC
	`

	rows, err := r.db.WithContext(ctx).Raw(query).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query function models: %w", err)
	}
	defer rows.Close()

	return r.scanModels(rows)
}

// GetModelsByCostRange retrieves models within a specific cost range
func (r *ModelRepository) GetModelsByCostRange(ctx context.Context, minCost, maxCost float64) ([]*gateway.Model, error) {
	query := `
		SELECT id, provider_id, model_name, display_name, model_type,
			   input_cost_per_1k_tokens, output_cost_per_1k_tokens, max_context_tokens,
			   supports_streaming, supports_functions, supports_vision,
			   quality_score, speed_score, metadata, is_enabled,
			   created_at, updated_at
		FROM gateway_models
		WHERE is_enabled = true 
		  AND (input_cost_per_1k_tokens + output_cost_per_1k_tokens) >= $1 
		  AND (input_cost_per_1k_tokens + output_cost_per_1k_tokens) <= $2
		ORDER BY (input_cost_per_1k_tokens + output_cost_per_1k_tokens) ASC, model_name ASC
	`

	rows, err := r.db.WithContext(ctx).Raw(query, minCost, maxCost).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query models by cost range: %w", err)
	}
	defer rows.Close()

	return r.scanModels(rows)
}

// GetModelsByQualityRange retrieves models within a specific quality score range
func (r *ModelRepository) GetModelsByQualityRange(ctx context.Context, minQuality, maxQuality float64) ([]*gateway.Model, error) {
	query := `
		SELECT id, provider_id, model_name, display_name, model_type,
			   input_cost_per_1k_tokens, output_cost_per_1k_tokens, max_context_tokens,
			   supports_streaming, supports_functions, supports_vision,
			   quality_score, speed_score, metadata, is_enabled,
			   created_at, updated_at
		FROM gateway_models
		WHERE is_enabled = true 
		  AND quality_score >= $1 
		  AND quality_score <= $2
		ORDER BY quality_score DESC, model_name ASC
	`

	rows, err := r.db.WithContext(ctx).Raw(query, minQuality, maxQuality).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query models by quality range: %w", err)
	}
	defer rows.Close()

	return r.scanModels(rows)
}

// GetStreamingModels retrieves models that support streaming
func (r *ModelRepository) GetStreamingModels(ctx context.Context) ([]*gateway.Model, error) {
	query := `
		SELECT id, provider_id, model_name, display_name, model_type,
			   input_cost_per_1k_tokens, output_cost_per_1k_tokens, max_context_tokens,
			   supports_streaming, supports_functions, supports_vision,
			   quality_score, speed_score, metadata, is_enabled,
			   created_at, updated_at
		FROM gateway_models
		WHERE supports_streaming = true AND is_enabled = true
		ORDER BY speed_score DESC, model_name ASC
	`

	rows, err := r.db.WithContext(ctx).Raw(query).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query streaming models: %w", err)
	}
	defer rows.Close()

	return r.scanModels(rows)
}

// GetVisionModels retrieves models that support vision capabilities
func (r *ModelRepository) GetVisionModels(ctx context.Context) ([]*gateway.Model, error) {
	query := `
		SELECT id, provider_id, model_name, display_name, model_type,
			   input_cost_per_1k_tokens, output_cost_per_1k_tokens, max_context_tokens,
			   supports_streaming, supports_functions, supports_vision,
			   quality_score, speed_score, metadata, is_enabled,
			   created_at, updated_at
		FROM gateway_models
		WHERE supports_vision = true AND is_enabled = true
		ORDER BY quality_score DESC, model_name ASC
	`

	rows, err := r.db.WithContext(ctx).Raw(query).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query vision models: %w", err)
	}
	defer rows.Close()

	return r.scanModels(rows)
}

// ListWithProvider retrieves models with provider information
func (r *ModelRepository) ListWithProvider(ctx context.Context, limit, offset int) ([]*gateway.Model, error) {
	// For now, just return the regular list - in a full implementation,
	// this would join with the provider table to include provider details
	return r.List(ctx, limit, offset)
}

// GetCompatibleModels retrieves models compatible with the given requirements
func (r *ModelRepository) GetCompatibleModels(ctx context.Context, requirements *gateway.ModelRequirements) ([]*gateway.Model, error) {
	query := `
		SELECT id, provider_id, model_name, display_name, model_type,
			   input_cost_per_1k_tokens, output_cost_per_1k_tokens, max_context_tokens,
			   supports_streaming, supports_functions, supports_vision,
			   quality_score, speed_score, metadata, is_enabled,
			   created_at, updated_at
		FROM gateway_models
		WHERE is_enabled = true
	`

	var args []interface{}
	argIndex := 1

	// ModelType is not a pointer, so we always filter by it
	query += fmt.Sprintf(" AND model_type = $%d", argIndex)
	args = append(args, string(requirements.ModelType))
	argIndex++

	if requirements.MaxCostPer1k != nil {
		query += fmt.Sprintf(" AND (input_cost_per_1k_tokens + output_cost_per_1k_tokens) <= $%d", argIndex)
		args = append(args, *requirements.MaxCostPer1k)
		argIndex++
	}

	if requirements.MinContextTokens != nil {
		query += fmt.Sprintf(" AND max_context_tokens >= $%d", argIndex)
		args = append(args, *requirements.MinContextTokens)
		argIndex++
	}

	if requirements.SupportsStreaming != nil && *requirements.SupportsStreaming {
		query += " AND supports_streaming = true"
	}

	if requirements.SupportsFunctions != nil && *requirements.SupportsFunctions {
		query += " AND supports_functions = true"
	}

	if requirements.SupportsVision != nil && *requirements.SupportsVision {
		query += " AND supports_vision = true"
	}

	if requirements.MinQualityScore != nil {
		query += fmt.Sprintf(" AND quality_score >= $%d", argIndex)
		args = append(args, *requirements.MinQualityScore)
		argIndex++
	}

	query += " ORDER BY quality_score DESC, speed_score DESC"

	rows, err := r.db.WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query compatible models: %w", err)
	}
	defer rows.Close()

	return r.scanModels(rows)
}
