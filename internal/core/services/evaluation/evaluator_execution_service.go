package evaluation

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"brokle/internal/core/domain/evaluation"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/pagination"
	"brokle/pkg/ulid"
)

type evaluatorExecutionService struct {
	repo   evaluation.EvaluatorExecutionRepository
	logger *slog.Logger
}

func NewEvaluatorExecutionService(
	repo evaluation.EvaluatorExecutionRepository,
	logger *slog.Logger,
) evaluation.EvaluatorExecutionService {
	return &evaluatorExecutionService{
		repo:   repo,
		logger: logger,
	}
}

func (s *evaluatorExecutionService) StartExecution(
	ctx context.Context,
	evaluatorID ulid.ULID,
	projectID ulid.ULID,
	triggerType evaluation.TriggerType,
) (*evaluation.EvaluatorExecution, error) {
	execution := evaluation.NewEvaluatorExecution(evaluatorID, projectID, triggerType)
	execution.Start()

	if err := s.repo.Create(ctx, execution); err != nil {
		return nil, appErrors.NewInternalError("failed to create rule execution", err)
	}

	s.logger.Info("evaluator execution started",
		"execution_id", execution.ID,
		"evaluator_id", evaluatorID,
		"project_id", projectID,
		"trigger_type", triggerType,
	)

	return execution, nil
}

func (s *evaluatorExecutionService) CompleteExecution(
	ctx context.Context,
	executionID ulid.ULID,
	projectID ulid.ULID,
	spansMatched, spansScored, errorsCount int,
) error {
	execution, err := s.repo.GetByID(ctx, executionID, projectID)
	if err != nil {
		if errors.Is(err, evaluation.ErrExecutionNotFound) {
			return appErrors.NewNotFoundError(fmt.Sprintf("evaluator execution %s", executionID))
		}
		return appErrors.NewInternalError("failed to get rule execution", err)
	}

	if execution.IsTerminal() {
		return appErrors.NewConflictError("execution is already in a terminal state")
	}

	execution.Complete(spansMatched, spansScored, errorsCount)

	if err := s.repo.Update(ctx, execution); err != nil {
		return appErrors.NewInternalError("failed to update rule execution", err)
	}

	s.logger.Info("evaluator execution completed",
		"execution_id", executionID,
		"evaluator_id", execution.EvaluatorID,
		"project_id", projectID,
		"spans_matched", spansMatched,
		"spans_scored", spansScored,
		"errors_count", errorsCount,
		"duration_ms", execution.DurationMs,
	)

	return nil
}

func (s *evaluatorExecutionService) FailExecution(
	ctx context.Context,
	executionID ulid.ULID,
	projectID ulid.ULID,
	errorMessage string,
) error {
	execution, err := s.repo.GetByID(ctx, executionID, projectID)
	if err != nil {
		if errors.Is(err, evaluation.ErrExecutionNotFound) {
			return appErrors.NewNotFoundError(fmt.Sprintf("evaluator execution %s", executionID))
		}
		return appErrors.NewInternalError("failed to get rule execution", err)
	}

	if execution.IsTerminal() {
		return appErrors.NewConflictError("execution is already in a terminal state")
	}

	execution.Fail(errorMessage)

	if err := s.repo.Update(ctx, execution); err != nil {
		return appErrors.NewInternalError("failed to update rule execution", err)
	}

	s.logger.Error("evaluator execution failed",
		"execution_id", executionID,
		"evaluator_id", execution.EvaluatorID,
		"project_id", projectID,
		"error_message", errorMessage,
		"duration_ms", execution.DurationMs,
	)

	return nil
}

func (s *evaluatorExecutionService) CancelExecution(
	ctx context.Context,
	executionID ulid.ULID,
	projectID ulid.ULID,
) error {
	execution, err := s.repo.GetByID(ctx, executionID, projectID)
	if err != nil {
		if errors.Is(err, evaluation.ErrExecutionNotFound) {
			return appErrors.NewNotFoundError(fmt.Sprintf("evaluator execution %s", executionID))
		}
		return appErrors.NewInternalError("failed to get rule execution", err)
	}

	if execution.IsTerminal() {
		return appErrors.NewConflictError("execution is already in a terminal state")
	}

	execution.Cancel()

	if err := s.repo.Update(ctx, execution); err != nil {
		return appErrors.NewInternalError("failed to update rule execution", err)
	}

	s.logger.Info("evaluator execution cancelled",
		"execution_id", executionID,
		"evaluator_id", execution.EvaluatorID,
		"project_id", projectID,
	)

	return nil
}

func (s *evaluatorExecutionService) GetByID(
	ctx context.Context,
	id ulid.ULID,
	projectID ulid.ULID,
) (*evaluation.EvaluatorExecution, error) {
	execution, err := s.repo.GetByID(ctx, id, projectID)
	if err != nil {
		if errors.Is(err, evaluation.ErrExecutionNotFound) {
			return nil, appErrors.NewNotFoundError(fmt.Sprintf("evaluator execution %s", id))
		}
		return nil, appErrors.NewInternalError("failed to get rule execution", err)
	}
	return execution, nil
}

func (s *evaluatorExecutionService) ListByEvaluatorID(
	ctx context.Context,
	evaluatorID ulid.ULID,
	projectID ulid.ULID,
	filter *evaluation.ExecutionFilter,
	params pagination.Params,
) ([]*evaluation.EvaluatorExecution, int64, error) {
	executions, total, err := s.repo.GetByEvaluatorID(ctx, evaluatorID, projectID, filter, params)
	if err != nil {
		return nil, 0, appErrors.NewInternalError("failed to list rule executions", err)
	}
	return executions, total, nil
}

func (s *evaluatorExecutionService) GetLatestByEvaluatorID(
	ctx context.Context,
	evaluatorID ulid.ULID,
	projectID ulid.ULID,
) (*evaluation.EvaluatorExecution, error) {
	execution, err := s.repo.GetLatestByEvaluatorID(ctx, evaluatorID, projectID)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get latest rule execution", err)
	}
	return execution, nil
}

func (s *evaluatorExecutionService) IncrementCounters(
	ctx context.Context,
	executionID string,
	projectID ulid.ULID,
	spansScored, errorsCount int,
) error {
	id, err := ulid.Parse(executionID)
	if err != nil {
		return appErrors.NewValidationError("invalid execution ID format", executionID)
	}

	if err := s.repo.IncrementCounters(ctx, id, projectID, spansScored, errorsCount); err != nil {
		if errors.Is(err, evaluation.ErrExecutionNotFound) {
			// Don't fail on not found - execution may have been deleted
			s.logger.Warn("execution not found for counter increment",
				"execution_id", executionID,
				"project_id", projectID,
				"spans_scored", spansScored,
				"errors_count", errorsCount,
			)
			return nil
		}
		return appErrors.NewInternalError("failed to increment execution counters", err)
	}

	return nil
}

func (s *evaluatorExecutionService) StartExecutionWithCount(
	ctx context.Context,
	evaluatorID ulid.ULID,
	projectID ulid.ULID,
	triggerType evaluation.TriggerType,
	spansMatched int,
) (*evaluation.EvaluatorExecution, error) {
	execution := evaluation.NewEvaluatorExecution(evaluatorID, projectID, triggerType)
	execution.SpansMatched = spansMatched
	execution.Start()

	if err := s.repo.Create(ctx, execution); err != nil {
		return nil, appErrors.NewInternalError("failed to create rule execution", err)
	}

	s.logger.Info("evaluator execution started with count",
		"execution_id", execution.ID,
		"evaluator_id", evaluatorID,
		"project_id", projectID,
		"trigger_type", triggerType,
		"spans_matched", spansMatched,
	)

	return execution, nil
}

func (s *evaluatorExecutionService) IncrementAndCheckCompletion(
	ctx context.Context,
	executionID ulid.ULID,
	projectID ulid.ULID,
	spansScored, errorsCount int,
) (bool, error) {
	completed, err := s.repo.IncrementCountersAndComplete(ctx, executionID, projectID, spansScored, errorsCount)
	if err != nil {
		if errors.Is(err, evaluation.ErrExecutionNotFound) {
			// Don't fail on not found - execution may have been deleted
			s.logger.Warn("execution not found for counter increment",
				"execution_id", executionID,
				"spans_scored", spansScored,
				"errors_count", errorsCount,
			)
			return false, nil
		}
		return false, appErrors.NewInternalError("failed to increment and check completion", err)
	}

	if completed {
		s.logger.Info("evaluator execution auto-completed",
			"execution_id", executionID,
			"project_id", projectID,
		)
	}

	return completed, nil
}

func (s *evaluatorExecutionService) UpdateSpansMatched(
	ctx context.Context,
	executionID ulid.ULID,
	projectID ulid.ULID,
	spansMatched int,
) error {
	if err := s.repo.UpdateSpansMatched(ctx, executionID, projectID, spansMatched); err != nil {
		if errors.Is(err, evaluation.ErrExecutionNotFound) {
			return appErrors.NewNotFoundError(fmt.Sprintf("evaluator execution %s", executionID))
		}
		return appErrors.NewInternalError("failed to update spans_matched", err)
	}
	return nil
}

func (s *evaluatorExecutionService) GetExecutionDetail(
	ctx context.Context,
	executionID ulid.ULID,
	projectID ulid.ULID,
	evaluatorID ulid.ULID,
) (*evaluation.ExecutionDetailResponse, error) {
	// Get the execution record
	execution, err := s.repo.GetByID(ctx, executionID, projectID)
	if err != nil {
		if errors.Is(err, evaluation.ErrExecutionNotFound) {
			return nil, appErrors.NewNotFoundError(fmt.Sprintf("evaluator execution %s", executionID))
		}
		return nil, appErrors.NewInternalError("failed to get rule execution", err)
	}

	// Validate execution belongs to the specified rule
	if execution.EvaluatorID != evaluatorID {
		return nil, appErrors.NewNotFoundError(fmt.Sprintf("execution %s not found for rule %s", executionID, evaluatorID))
	}

	// For now, return the execution with empty span details.
	// Full implementation would query ClickHouse for span-level execution data.
	// This provides the foundation to be enhanced with detailed span results.
	response := &evaluation.ExecutionDetailResponse{
		Execution: execution.ToResponse(),
		Spans:     []evaluation.SpanExecutionDetail{},
		// RuleSnapshot would be populated if we stored rule config at execution time
	}

	s.logger.Info("execution detail retrieved",
		"execution_id", executionID,
		"project_id", projectID,
		"evaluator_id", execution.EvaluatorID,
	)

	return response, nil
}
