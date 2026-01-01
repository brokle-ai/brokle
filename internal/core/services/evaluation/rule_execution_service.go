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

type ruleExecutionService struct {
	repo   evaluation.RuleExecutionRepository
	logger *slog.Logger
}

func NewRuleExecutionService(
	repo evaluation.RuleExecutionRepository,
	logger *slog.Logger,
) evaluation.RuleExecutionService {
	return &ruleExecutionService{
		repo:   repo,
		logger: logger,
	}
}

func (s *ruleExecutionService) StartExecution(
	ctx context.Context,
	ruleID ulid.ULID,
	projectID ulid.ULID,
	triggerType evaluation.TriggerType,
) (*evaluation.RuleExecution, error) {
	execution := evaluation.NewRuleExecution(ruleID, projectID, triggerType)
	execution.Start()

	if err := s.repo.Create(ctx, execution); err != nil {
		return nil, appErrors.NewInternalError("failed to create rule execution", err)
	}

	s.logger.Info("rule execution started",
		"execution_id", execution.ID,
		"rule_id", ruleID,
		"project_id", projectID,
		"trigger_type", triggerType,
	)

	return execution, nil
}

func (s *ruleExecutionService) CompleteExecution(
	ctx context.Context,
	executionID ulid.ULID,
	projectID ulid.ULID,
	spansMatched, spansScored, errorsCount int,
) error {
	execution, err := s.repo.GetByID(ctx, executionID, projectID)
	if err != nil {
		if errors.Is(err, evaluation.ErrExecutionNotFound) {
			return appErrors.NewNotFoundError(fmt.Sprintf("rule execution %s", executionID))
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

	s.logger.Info("rule execution completed",
		"execution_id", executionID,
		"rule_id", execution.RuleID,
		"project_id", projectID,
		"spans_matched", spansMatched,
		"spans_scored", spansScored,
		"errors_count", errorsCount,
		"duration_ms", execution.DurationMs,
	)

	return nil
}

func (s *ruleExecutionService) FailExecution(
	ctx context.Context,
	executionID ulid.ULID,
	projectID ulid.ULID,
	errorMessage string,
) error {
	execution, err := s.repo.GetByID(ctx, executionID, projectID)
	if err != nil {
		if errors.Is(err, evaluation.ErrExecutionNotFound) {
			return appErrors.NewNotFoundError(fmt.Sprintf("rule execution %s", executionID))
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

	s.logger.Error("rule execution failed",
		"execution_id", executionID,
		"rule_id", execution.RuleID,
		"project_id", projectID,
		"error_message", errorMessage,
		"duration_ms", execution.DurationMs,
	)

	return nil
}

func (s *ruleExecutionService) CancelExecution(
	ctx context.Context,
	executionID ulid.ULID,
	projectID ulid.ULID,
) error {
	execution, err := s.repo.GetByID(ctx, executionID, projectID)
	if err != nil {
		if errors.Is(err, evaluation.ErrExecutionNotFound) {
			return appErrors.NewNotFoundError(fmt.Sprintf("rule execution %s", executionID))
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

	s.logger.Info("rule execution cancelled",
		"execution_id", executionID,
		"rule_id", execution.RuleID,
		"project_id", projectID,
	)

	return nil
}

func (s *ruleExecutionService) GetByID(
	ctx context.Context,
	id ulid.ULID,
	projectID ulid.ULID,
) (*evaluation.RuleExecution, error) {
	execution, err := s.repo.GetByID(ctx, id, projectID)
	if err != nil {
		if errors.Is(err, evaluation.ErrExecutionNotFound) {
			return nil, appErrors.NewNotFoundError(fmt.Sprintf("rule execution %s", id))
		}
		return nil, appErrors.NewInternalError("failed to get rule execution", err)
	}
	return execution, nil
}

func (s *ruleExecutionService) ListByRuleID(
	ctx context.Context,
	ruleID ulid.ULID,
	projectID ulid.ULID,
	filter *evaluation.ExecutionFilter,
	params pagination.Params,
) ([]*evaluation.RuleExecution, int64, error) {
	executions, total, err := s.repo.GetByRuleID(ctx, ruleID, projectID, filter, params)
	if err != nil {
		return nil, 0, appErrors.NewInternalError("failed to list rule executions", err)
	}
	return executions, total, nil
}

func (s *ruleExecutionService) GetLatestByRuleID(
	ctx context.Context,
	ruleID ulid.ULID,
	projectID ulid.ULID,
) (*evaluation.RuleExecution, error) {
	execution, err := s.repo.GetLatestByRuleID(ctx, ruleID, projectID)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get latest rule execution", err)
	}
	return execution, nil
}

func (s *ruleExecutionService) IncrementCounters(
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

func (s *ruleExecutionService) StartExecutionWithCount(
	ctx context.Context,
	ruleID ulid.ULID,
	projectID ulid.ULID,
	triggerType evaluation.TriggerType,
	spansMatched int,
) (*evaluation.RuleExecution, error) {
	execution := evaluation.NewRuleExecution(ruleID, projectID, triggerType)
	execution.SpansMatched = spansMatched
	execution.Start()

	if err := s.repo.Create(ctx, execution); err != nil {
		return nil, appErrors.NewInternalError("failed to create rule execution", err)
	}

	s.logger.Info("rule execution started with count",
		"execution_id", execution.ID,
		"rule_id", ruleID,
		"project_id", projectID,
		"trigger_type", triggerType,
		"spans_matched", spansMatched,
	)

	return execution, nil
}

func (s *ruleExecutionService) IncrementAndCheckCompletion(
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
		s.logger.Info("rule execution auto-completed",
			"execution_id", executionID,
			"project_id", projectID,
		)
	}

	return completed, nil
}

func (s *ruleExecutionService) UpdateSpansMatched(
	ctx context.Context,
	executionID ulid.ULID,
	projectID ulid.ULID,
	spansMatched int,
) error {
	if err := s.repo.UpdateSpansMatched(ctx, executionID, projectID, spansMatched); err != nil {
		if errors.Is(err, evaluation.ErrExecutionNotFound) {
			return appErrors.NewNotFoundError(fmt.Sprintf("rule execution %s", executionID))
		}
		return appErrors.NewInternalError("failed to update spans_matched", err)
	}
	return nil
}
