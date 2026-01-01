package evaluation

import (
	"context"
	"errors"
	"time"

	"brokle/internal/core/domain/evaluation"
	"brokle/pkg/pagination"
	"brokle/pkg/ulid"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RuleExecutionRepository struct {
	db *gorm.DB
}

func NewRuleExecutionRepository(db *gorm.DB) *RuleExecutionRepository {
	return &RuleExecutionRepository{db: db}
}

func (r *RuleExecutionRepository) Create(ctx context.Context, execution *evaluation.RuleExecution) error {
	result := r.db.WithContext(ctx).Create(execution)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *RuleExecutionRepository) Update(ctx context.Context, execution *evaluation.RuleExecution) error {
	result := r.db.WithContext(ctx).
		Model(&evaluation.RuleExecution{}).
		Where("id = ? AND project_id = ?", execution.ID.String(), execution.ProjectID.String()).
		Updates(map[string]interface{}{
			"status":        execution.Status,
			"spans_matched": execution.SpansMatched,
			"spans_scored":  execution.SpansScored,
			"errors_count":  execution.ErrorsCount,
			"error_message": execution.ErrorMessage,
			"started_at":    execution.StartedAt,
			"completed_at":  execution.CompletedAt,
			"duration_ms":   execution.DurationMs,
			"metadata":      execution.Metadata,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return evaluation.ErrExecutionNotFound
	}
	return nil
}

func (r *RuleExecutionRepository) GetByID(ctx context.Context, id ulid.ULID, projectID ulid.ULID) (*evaluation.RuleExecution, error) {
	var execution evaluation.RuleExecution
	result := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ?", id.String(), projectID.String()).
		First(&execution)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, evaluation.ErrExecutionNotFound
		}
		return nil, result.Error
	}
	return &execution, nil
}

func (r *RuleExecutionRepository) GetByRuleID(
	ctx context.Context,
	ruleID ulid.ULID,
	projectID ulid.ULID,
	filter *evaluation.ExecutionFilter,
	params pagination.Params,
) ([]*evaluation.RuleExecution, int64, error) {
	var executions []*evaluation.RuleExecution
	var total int64

	query := r.db.WithContext(ctx).
		Model(&evaluation.RuleExecution{}).
		Where("rule_id = ? AND project_id = ?", ruleID.String(), projectID.String())

	if filter != nil {
		if filter.Status != nil {
			query = query.Where("status = ?", *filter.Status)
		}
		if filter.TriggerType != nil {
			query = query.Where("trigger_type = ?", *filter.TriggerType)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (params.Page - 1) * params.Limit
	result := query.
		Order("created_at DESC").
		Limit(params.Limit).
		Offset(offset).
		Find(&executions)

	if result.Error != nil {
		return nil, 0, result.Error
	}
	return executions, total, nil
}

func (r *RuleExecutionRepository) GetLatestByRuleID(ctx context.Context, ruleID ulid.ULID, projectID ulid.ULID) (*evaluation.RuleExecution, error) {
	var execution evaluation.RuleExecution
	result := r.db.WithContext(ctx).
		Where("rule_id = ? AND project_id = ?", ruleID.String(), projectID.String()).
		Order("created_at DESC").
		First(&execution)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // No execution found, return nil without error
		}
		return nil, result.Error
	}
	return &execution, nil
}

func (r *RuleExecutionRepository) IncrementCounters(
	ctx context.Context,
	id ulid.ULID,
	projectID ulid.ULID,
	spansScored, errorsCount int,
) error {
	result := r.db.WithContext(ctx).
		Model(&evaluation.RuleExecution{}).
		Where("id = ? AND project_id = ?", id.String(), projectID.String()).
		Updates(map[string]interface{}{
			"spans_scored": gorm.Expr("spans_scored + ?", spansScored),
			"errors_count": gorm.Expr("errors_count + ?", errorsCount),
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return evaluation.ErrExecutionNotFound
	}
	return nil
}

func (r *RuleExecutionRepository) UpdateSpansMatched(
	ctx context.Context,
	id ulid.ULID,
	projectID ulid.ULID,
	spansMatched int,
) error {
	result := r.db.WithContext(ctx).
		Model(&evaluation.RuleExecution{}).
		Where("id = ? AND project_id = ?", id.String(), projectID.String()).
		Update("spans_matched", spansMatched)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return evaluation.ErrExecutionNotFound
	}
	return nil
}

func (r *RuleExecutionRepository) IncrementCountersAndComplete(
	ctx context.Context,
	id ulid.ULID,
	projectID ulid.ULID,
	spansScored, errorsCount int,
) (bool, error) {
	var completed bool

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Lock the row for update
		var exec evaluation.RuleExecution
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND project_id = ?", id.String(), projectID.String()).
			First(&exec).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return evaluation.ErrExecutionNotFound
			}
			return err
		}

		// Increment counters
		exec.SpansScored += spansScored
		exec.ErrorsCount += errorsCount

		// Check if execution is complete (all spans processed)
		if exec.SpansScored+exec.ErrorsCount >= exec.SpansMatched && exec.Status == evaluation.ExecutionStatusRunning {
			exec.Status = evaluation.ExecutionStatusCompleted
			now := time.Now()
			exec.CompletedAt = &now
			if exec.StartedAt != nil {
				durationMs := int(now.Sub(*exec.StartedAt).Milliseconds())
				exec.DurationMs = &durationMs
			}
			completed = true
		}

		return tx.Model(&evaluation.RuleExecution{}).
			Where("id = ?", id.String()).
			Updates(map[string]interface{}{
				"spans_scored": exec.SpansScored,
				"errors_count": exec.ErrorsCount,
				"status":       exec.Status,
				"completed_at": exec.CompletedAt,
				"duration_ms":  exec.DurationMs,
			}).Error
	})

	return completed, err
}
