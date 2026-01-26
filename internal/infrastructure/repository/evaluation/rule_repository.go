package evaluation

import (
	"context"
	"errors"

	"brokle/internal/core/domain/evaluation"
	"brokle/pkg/pagination"
	"brokle/pkg/ulid"

	"gorm.io/gorm"
)

type RuleRepository struct {
	db *gorm.DB
}

func NewRuleRepository(db *gorm.DB) *RuleRepository {
	return &RuleRepository{db: db}
}

func (r *RuleRepository) Create(ctx context.Context, rule *evaluation.EvaluationRule) error {
	result := r.db.WithContext(ctx).Create(rule)
	if result.Error != nil {
		if isUniqueViolation(result.Error) {
			return evaluation.ErrRuleExists
		}
		return result.Error
	}
	return nil
}

func (r *RuleRepository) GetByID(ctx context.Context, id ulid.ULID, projectID ulid.ULID) (*evaluation.EvaluationRule, error) {
	var rule evaluation.EvaluationRule
	result := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ?", id.String(), projectID.String()).
		First(&rule)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, evaluation.ErrRuleNotFound
		}
		return nil, result.Error
	}
	return &rule, nil
}

func (r *RuleRepository) GetByProjectID(ctx context.Context, projectID ulid.ULID, filter *evaluation.RuleFilter, params pagination.Params) ([]*evaluation.EvaluationRule, int64, error) {
	var rules []*evaluation.EvaluationRule
	var total int64

	query := r.db.WithContext(ctx).
		Model(&evaluation.EvaluationRule{}).
		Where("project_id = ?", projectID.String())

	if filter != nil {
		if filter.Status != nil {
			query = query.Where("status = ?", *filter.Status)
		}
		if filter.ScorerType != nil {
			query = query.Where("scorer_type = ?", *filter.ScorerType)
		}
		if filter.Search != nil && *filter.Search != "" {
			searchTerm := "%" + *filter.Search + "%"
			query = query.Where("name ILIKE ?", searchTerm)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (params.Page - 1) * params.Limit
	result := query.
		Order(params.GetSortOrder(params.SortBy, "id")).
		Limit(params.Limit).
		Offset(offset).
		Find(&rules)

	if result.Error != nil {
		return nil, 0, result.Error
	}
	return rules, total, nil
}

func (r *RuleRepository) GetActiveByProjectID(ctx context.Context, projectID ulid.ULID) ([]*evaluation.EvaluationRule, error) {
	var rules []*evaluation.EvaluationRule
	result := r.db.WithContext(ctx).
		Where("project_id = ? AND status = ?", projectID.String(), evaluation.RuleStatusActive).
		Order("created_at DESC").
		Find(&rules)

	if result.Error != nil {
		return nil, result.Error
	}
	return rules, nil
}

func (r *RuleRepository) Update(ctx context.Context, rule *evaluation.EvaluationRule) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ?", rule.ID.String(), rule.ProjectID.String()).
		Save(rule)

	if result.Error != nil {
		if isUniqueViolation(result.Error) {
			return evaluation.ErrRuleExists
		}
		return result.Error
	}

	if result.RowsAffected == 0 {
		return evaluation.ErrRuleNotFound
	}
	return nil
}

func (r *RuleRepository) Delete(ctx context.Context, id ulid.ULID, projectID ulid.ULID) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ?", id.String(), projectID.String()).
		Delete(&evaluation.EvaluationRule{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return evaluation.ErrRuleNotFound
	}
	return nil
}

func (r *RuleRepository) ExistsByName(ctx context.Context, projectID ulid.ULID, name string) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&evaluation.EvaluationRule{}).
		Where("project_id = ? AND name = ?", projectID.String(), name).
		Count(&count)

	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

