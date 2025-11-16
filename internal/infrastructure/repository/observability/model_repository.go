package observability

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"brokle/internal/core/domain/observability"
	"brokle/pkg/pagination"
)

// ModelRepository implements observability.ModelRepository
type ModelRepository struct {
	db *gorm.DB
}

// NewModelRepository creates a new model pricing repository
func NewModelRepository(db *gorm.DB) *ModelRepository {
	return &ModelRepository{db: db}
}

// FindByModelName finds pricing for a model using regex pattern matching
// Implements Langfuse pattern: project-scoped pricing with global fallback
// Priority: project-specific > global > NULL (continue without costs)
func (r *ModelRepository) FindByModelName(
	ctx context.Context,
	modelName, projectID string,
) (*observability.Model, error) {
	var pricing observability.Model

	// Query with regex matching and temporal validity
	err := r.db.WithContext(ctx).
		Where("(project_id = ? OR project_id IS NULL)", projectID).
		Where("? ~ match_pattern", modelName).
		Where("(end_date IS NULL OR end_date > CURRENT_TIMESTAMP)").
		Where("is_deprecated = ?", false).
		Order("project_id ASC NULLS LAST"). // Project-specific first
		Order("start_date DESC").           // Latest pricing first
		First(&pricing).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// No pricing found - wrap with domain error
			return nil, fmt.Errorf("find model %s in project %s: %w", modelName, projectID, observability.ErrModelNotFound)
		}
		return nil, fmt.Errorf("query model pricing: %w", err)
	}

	return &pricing, nil
}

// GetByID retrieves pricing by ID
func (r *ModelRepository) GetByID(ctx context.Context, id string) (*observability.Model, error) {
	var pricing observability.Model

	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&pricing).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("get model by id %s: %w", id, observability.ErrModelNotFound)
		}
		return nil, fmt.Errorf("query model by id: %w", err)
	}

	return &pricing, nil
}

// Create creates a new model pricing entry
func (r *ModelRepository) Create(ctx context.Context, pricing *observability.Model) error {
	pricing.CreatedAt = time.Now()
	pricing.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Create(pricing).Error
}

// Update updates an existing model pricing entry
func (r *ModelRepository) Update(ctx context.Context, pricing *observability.Model) error {
	pricing.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(pricing).Error
}

// Delete soft deletes a model pricing entry by setting end_date to now
func (r *ModelRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&observability.Model{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"end_date":   time.Now(),
			"updated_at": time.Now(),
		}).Error
}

// List retrieves model pricing entries with optional filters
func (r *ModelRepository) List(
	ctx context.Context,
	filter *observability.ModelFilter,
) ([]*observability.Model, error) {
	var pricings []*observability.Model

	query := r.db.WithContext(ctx)

	if filter.ProjectID != nil {
		query = query.Where("project_id = ?", *filter.ProjectID)
	}

	if filter.Provider != nil {
		query = query.Where("provider = ?", *filter.Provider)
	}

	if filter.ModelName != nil {
		query = query.Where("model_name = ?", *filter.ModelName)
	}

	if filter.IsActive != nil && *filter.IsActive {
		query = query.Where("(end_date IS NULL OR end_date > ?)", time.Now())
	}

	if filter.IsDeprecated != nil {
		query = query.Where("is_deprecated = ?", *filter.IsDeprecated)
	}

	// Determine sort field and direction with SQL injection protection
	allowedSortFields := []string{"created_at", "updated_at", "provider", "model_name", "input_cost_per_million", "output_cost_per_million", "id"}
	sortField := "created_at" // default
	sortDir := "DESC"
	limit := pagination.DefaultPageSize
	offset := 0

	if filter != nil {
		// Validate sort field against whitelist
		if filter.Params.SortBy != "" {
			validated, err := pagination.ValidateSortField(filter.Params.SortBy, allowedSortFields)
			if err != nil {
				return nil, err
			}
			if validated != "" {
				sortField = validated
			}
		}
		if filter.Params.SortDir == "asc" {
			sortDir = "ASC"
		}
		if filter.Params.Limit > 0 {
			limit = filter.Params.Limit
		}
		offset = filter.Params.GetOffset()
	}

	// Apply sorting with secondary sort on id for stable ordering
	query = query.Order(fmt.Sprintf("%s %s, id %s", sortField, sortDir, sortDir))

	// Apply limit and offset for pagination
	query = query.Limit(limit).Offset(offset)

	err := query.Find(&pricings).Error
	if err != nil {
		return nil, err
	}

	return pricings, nil
}

// Count returns the count of model pricing entries matching the filter
func (r *ModelRepository) Count(
	ctx context.Context,
	filter *observability.ModelFilter,
) (int64, error) {
	var count int64

	query := r.db.WithContext(ctx).Model(&observability.Model{})

	if filter.ProjectID != nil {
		query = query.Where("project_id = ?", *filter.ProjectID)
	}

	if filter.Provider != nil {
		query = query.Where("provider = ?", *filter.Provider)
	}

	if filter.ModelName != nil {
		query = query.Where("model_name = ?", *filter.ModelName)
	}

	if filter.IsActive != nil && *filter.IsActive {
		query = query.Where("(end_date IS NULL OR end_date > ?)", time.Now())
	}

	if filter.IsDeprecated != nil {
		query = query.Where("is_deprecated = ?", *filter.IsDeprecated)
	}

	err := query.Count(&count).Error
	return count, err
}

// FindHistoricalPricing finds pricing valid at a specific timestamp
func (r *ModelRepository) FindHistoricalPricing(
	ctx context.Context,
	modelName, projectID string,
	timestamp time.Time,
) (*observability.Model, error) {
	var pricing observability.Model

	err := r.db.WithContext(ctx).
		Where("(project_id = ? OR project_id IS NULL)", projectID).
		Where("? ~ match_pattern", modelName).
		Where("(start_date IS NULL OR start_date <= ?)", timestamp).
		Where("(end_date IS NULL OR end_date > ?)", timestamp).
		Order("project_id ASC NULLS LAST").
		Order("start_date DESC").
		First(&pricing).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("find historical pricing for %s at %v: %w", modelName, timestamp, observability.ErrModelNotFound)
		}
		return nil, fmt.Errorf("query historical pricing: %w", err)
	}

	return &pricing, nil
}
