package evaluation

import (
	"context"
	"time"

	"github.com/google/uuid"

	evalDomain "brokle/internal/core/domain/evaluation"
	"brokle/internal/infrastructure/db"
	"brokle/internal/infrastructure/db/gen"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/pagination"
)

type evaluatorExecutionRepository struct {
	tm *db.TxManager
}

func NewEvaluatorExecutionRepository(tm *db.TxManager) evalDomain.EvaluatorExecutionRepository {
	return &evaluatorExecutionRepository{tm: tm}
}

func (r *evaluatorExecutionRepository) Create(ctx context.Context, e *evalDomain.EvaluatorExecution) error {
	meta, err := marshalEvalJSON(e.Metadata)
	if err != nil {
		return err
	}
	var durationMs *int32
	if e.DurationMs != nil {
		d := int32(*e.DurationMs)
		durationMs = &d
	}
	return r.tm.Queries(ctx).CreateEvaluatorExecution(ctx, gen.CreateEvaluatorExecutionParams{
		ID:           e.ID,
		EvaluatorID:  e.EvaluatorID,
		ProjectID:    e.ProjectID,
		Status:       string(e.Status),
		TriggerType:  string(e.TriggerType),
		SpansMatched: int32(e.SpansMatched),
		SpansScored:  int32(e.SpansScored),
		ErrorsCount:  int32(e.ErrorsCount),
		ErrorMessage: e.ErrorMessage,
		StartedAt:    e.StartedAt,
		CompletedAt:  e.CompletedAt,
		DurationMs:   durationMs,
		Metadata:     meta,
	})
}

func (r *evaluatorExecutionRepository) Update(ctx context.Context, e *evalDomain.EvaluatorExecution) error {
	meta, err := marshalEvalJSON(e.Metadata)
	if err != nil {
		return err
	}
	var durationMs *int32
	if e.DurationMs != nil {
		d := int32(*e.DurationMs)
		durationMs = &d
	}
	n, err := r.tm.Queries(ctx).UpdateEvaluatorExecution(ctx, gen.UpdateEvaluatorExecutionParams{
		ID:           e.ID,
		ProjectID:    e.ProjectID,
		Status:       string(e.Status),
		SpansMatched: int32(e.SpansMatched),
		SpansScored:  int32(e.SpansScored),
		ErrorsCount:  int32(e.ErrorsCount),
		ErrorMessage: e.ErrorMessage,
		StartedAt:    e.StartedAt,
		CompletedAt:  e.CompletedAt,
		DurationMs:   durationMs,
		Metadata:     meta,
	})
	if err != nil {
		return appErrors.Internal("update execution", err, appErrors.WithOp("repo.execution.update"))
	}
	if n == 0 {
		return appErrors.NotFound("execution", appErrors.WithOp("repo.execution.update"))
	}
	return nil
}

func (r *evaluatorExecutionRepository) GetByID(ctx context.Context, id, projectID uuid.UUID) (*evalDomain.EvaluatorExecution, error) {
	row, err := r.tm.Queries(ctx).GetEvaluatorExecutionByID(ctx, gen.GetEvaluatorExecutionByIDParams{
		ID:        id,
		ProjectID: projectID,
	})
	if err != nil {
		if db.IsNoRows(err) {
			return nil, appErrors.NotFound("execution", appErrors.WithOp("repo.execution.get_by_id"))
		}
		return nil, appErrors.Internal("get execution", err, appErrors.WithOp("repo.execution.get_by_id"))
	}
	return executionFromRow(&row)
}

func (r *evaluatorExecutionRepository) GetByEvaluatorID(
	ctx context.Context,
	evaluatorID, projectID uuid.UUID,
	filter *evalDomain.ExecutionFilter,
	params pagination.Params,
) ([]*evalDomain.EvaluatorExecution, int64, error) {
	// Simple WHERE + optional filters. Squirrel isn't needed; two
	// boolean filters against a small table.
	base := "WHERE evaluator_id = $1 AND project_id = $2"
	args := []any{evaluatorID, projectID}
	idx := 3
	if filter != nil {
		if filter.Status != nil {
			base += " AND status = $" + itoa(idx)
			args = append(args, string(*filter.Status))
			idx++
		}
		if filter.TriggerType != nil {
			base += " AND trigger_type = $" + itoa(idx)
			args = append(args, string(*filter.TriggerType))
			idx++
		}
	}
	var total int64
	if err := r.tm.DB(ctx).QueryRow(ctx, "SELECT COUNT(*) FROM evaluator_executions "+base, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	limitIdx := idx
	offsetIdx := idx + 1
	query := "SELECT id, evaluator_id, project_id, status, trigger_type, spans_matched, spans_scored, errors_count, error_message, started_at, completed_at, duration_ms, metadata, created_at FROM evaluator_executions " + base + " ORDER BY created_at DESC LIMIT $" + itoa(limitIdx) + " OFFSET $" + itoa(offsetIdx)
	args = append(args, int32(params.Limit), int32((params.Page-1)*params.Limit))
	rows, err := r.tm.DB(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := make([]*evalDomain.EvaluatorExecution, 0)
	for rows.Next() {
		var r2 gen.EvaluatorExecution
		if err := rows.Scan(
			&r2.ID, &r2.EvaluatorID, &r2.ProjectID, &r2.Status, &r2.TriggerType,
			&r2.SpansMatched, &r2.SpansScored, &r2.ErrorsCount, &r2.ErrorMessage,
			&r2.StartedAt, &r2.CompletedAt, &r2.DurationMs, &r2.Metadata, &r2.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		e, err := executionFromRow(&r2)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, e)
	}
	return out, total, rows.Err()
}

func (r *evaluatorExecutionRepository) GetLatestByEvaluatorID(ctx context.Context, evaluatorID, projectID uuid.UUID) (*evalDomain.EvaluatorExecution, error) {
	row, err := r.tm.Queries(ctx).GetLatestEvaluatorExecution(ctx, gen.GetLatestEvaluatorExecutionParams{
		EvaluatorID: evaluatorID,
		ProjectID:   projectID,
	})
	if err != nil {
		if db.IsNoRows(err) {
			return nil, nil
		}
		return nil, err
	}
	return executionFromRow(&row)
}

func (r *evaluatorExecutionRepository) IncrementCounters(ctx context.Context, id, projectID uuid.UUID, spansScored, errorsCount int) error {
	n, err := r.tm.Queries(ctx).IncrementEvaluatorExecutionCounters(ctx, gen.IncrementEvaluatorExecutionCountersParams{
		ID:           id,
		ProjectID:    projectID,
		SpansScored:  int32(spansScored),
		ErrorsCount:  int32(errorsCount),
	})
	if err != nil {
		return appErrors.Internal("increment execution counters", err, appErrors.WithOp("repo.execution.increment_counters"))
	}
	if n == 0 {
		return appErrors.NotFound("execution", appErrors.WithOp("repo.execution.increment_counters"))
	}
	return nil
}

func (r *evaluatorExecutionRepository) UpdateSpansMatched(ctx context.Context, id, projectID uuid.UUID, spansMatched int) error {
	n, err := r.tm.Queries(ctx).UpdateEvaluatorExecutionSpansMatched(ctx, gen.UpdateEvaluatorExecutionSpansMatchedParams{
		ID:           id,
		ProjectID:    projectID,
		SpansMatched: int32(spansMatched),
	})
	if err != nil {
		return appErrors.Internal("update execution spans matched", err, appErrors.WithOp("repo.execution.update_spans_matched"))
	}
	if n == 0 {
		return appErrors.NotFound("execution", appErrors.WithOp("repo.execution.update_spans_matched"))
	}
	return nil
}

// IncrementCountersAndComplete locks the execution row, increments
// spans_scored + errors_count, and flips status to completed once all
// matched spans have been processed.
func (r *evaluatorExecutionRepository) IncrementCountersAndComplete(ctx context.Context, id, projectID uuid.UUID, spansScored, errorsCount int) (bool, error) {
	var completed bool
	err := r.tm.WithinTransaction(ctx, func(ctx context.Context) error {
		row, err := r.tm.Queries(ctx).LockEvaluatorExecutionForUpdate(ctx, gen.LockEvaluatorExecutionForUpdateParams{
			ID:        id,
			ProjectID: projectID,
		})
		if err != nil {
			if db.IsNoRows(err) {
				return appErrors.NotFound("execution", appErrors.WithOp("repo.execution.lock_for_update"))
			}
			return appErrors.Internal("lock execution", err, appErrors.WithOp("repo.execution.lock_for_update"))
		}
		row.SpansScored += int32(spansScored)
		row.ErrorsCount += int32(errorsCount)
		if row.SpansScored+row.ErrorsCount >= row.SpansMatched && row.Status == string(evalDomain.ExecutionStatusRunning) {
			row.Status = string(evalDomain.ExecutionStatusCompleted)
			now := time.Now()
			row.CompletedAt = &now
			if row.StartedAt != nil {
				d := int32(now.Sub(*row.StartedAt).Milliseconds())
				row.DurationMs = &d
			}
			completed = true
		}
		_, err = r.tm.Queries(ctx).ApplyEvaluatorExecutionCompletion(ctx, gen.ApplyEvaluatorExecutionCompletionParams{
			ID:          row.ID,
			ProjectID:   row.ProjectID,
			SpansScored: row.SpansScored,
			ErrorsCount: row.ErrorsCount,
			Status:      row.Status,
			CompletedAt: row.CompletedAt,
			DurationMs:  row.DurationMs,
		})
		return err
	})
	return completed, err
}

func executionFromRow(row *gen.EvaluatorExecution) (*evalDomain.EvaluatorExecution, error) {
	e := &evalDomain.EvaluatorExecution{
		ID:           row.ID,
		EvaluatorID:  row.EvaluatorID,
		ProjectID:    row.ProjectID,
		Status:       evalDomain.ExecutionStatus(row.Status),
		TriggerType:  evalDomain.TriggerType(row.TriggerType),
		SpansMatched: int(row.SpansMatched),
		SpansScored:  int(row.SpansScored),
		ErrorsCount:  int(row.ErrorsCount),
		ErrorMessage: row.ErrorMessage,
		StartedAt:    row.StartedAt,
		CompletedAt:  row.CompletedAt,
		CreatedAt:    row.CreatedAt,
	}
	if row.DurationMs != nil {
		d := int(*row.DurationMs)
		e.DurationMs = &d
	}
	if err := unmarshalEvalJSON(row.Metadata, &e.Metadata); err != nil {
		return nil, err
	}
	return e, nil
}

// itoa (local, avoids strconv import for a single 1-digit index usage).
func itoa(i int) string {
	if i < 10 {
		return string(rune('0' + i))
	}
	// Two-digit fallback (shouldn't happen with the small arg list here).
	return string(rune('0'+i/10)) + string(rune('0'+i%10))
}
