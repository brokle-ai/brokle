package evaluation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"brokle/internal/core/domain/evaluation"
	"brokle/internal/infrastructure/database"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/pagination"
	"brokle/pkg/ulid"
)

const (
	manualTriggerStream = "evaluation:manual-triggers"
)

type ruleService struct {
	repo             evaluation.RuleRepository
	executionService evaluation.RuleExecutionService
	redis            *database.RedisDB
	logger           *slog.Logger
}

func NewRuleService(
	repo evaluation.RuleRepository,
	executionService evaluation.RuleExecutionService,
	redis *database.RedisDB,
	logger *slog.Logger,
) evaluation.RuleService {
	return &ruleService{
		repo:             repo,
		executionService: executionService,
		redis:            redis,
		logger:           logger,
	}
}

func (s *ruleService) Create(ctx context.Context, projectID ulid.ULID, userID *ulid.ULID, req *evaluation.CreateEvaluationRuleRequest) (*evaluation.EvaluationRule, error) {
	rule := evaluation.NewEvaluationRule(projectID, req.Name, req.ScorerType, req.ScorerConfig)

	if req.Description != nil {
		rule.Description = req.Description
	}
	if req.Status != nil {
		rule.Status = *req.Status
	}
	if req.TriggerType != nil {
		rule.TriggerType = *req.TriggerType
	}
	if req.TargetScope != nil {
		rule.TargetScope = *req.TargetScope
	}
	if req.Filter != nil {
		rule.Filter = req.Filter
	}
	if req.SpanNames != nil {
		rule.SpanNames = req.SpanNames
	}
	if req.SamplingRate != nil {
		rule.SamplingRate = *req.SamplingRate
	}
	if req.VariableMapping != nil {
		rule.VariableMapping = req.VariableMapping
	}
	if userID != nil {
		id := userID.String()
		rule.CreatedBy = &id
	}

	if validationErrors := rule.Validate(); len(validationErrors) > 0 {
		return nil, appErrors.NewValidationError(validationErrors[0].Field, validationErrors[0].Message)
	}

	exists, err := s.repo.ExistsByName(ctx, projectID, req.Name)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to check name uniqueness", err)
	}
	if exists {
		return nil, appErrors.NewConflictError(fmt.Sprintf("evaluation rule '%s' already exists in this project", req.Name))
	}

	if err := s.repo.Create(ctx, rule); err != nil {
		if errors.Is(err, evaluation.ErrRuleExists) {
			return nil, appErrors.NewConflictError(fmt.Sprintf("evaluation rule '%s' already exists in this project", req.Name))
		}
		return nil, appErrors.NewInternalError("failed to create evaluation rule", err)
	}

	s.logger.Info("evaluation rule created",
		"rule_id", rule.ID,
		"project_id", projectID,
		"name", rule.Name,
		"scorer_type", rule.ScorerType,
		"status", rule.Status,
	)

	return rule, nil
}

func (s *ruleService) Update(ctx context.Context, id ulid.ULID, projectID ulid.ULID, req *evaluation.UpdateEvaluationRuleRequest) (*evaluation.EvaluationRule, error) {
	rule, err := s.repo.GetByID(ctx, id, projectID)
	if err != nil {
		if errors.Is(err, evaluation.ErrRuleNotFound) {
			return nil, appErrors.NewNotFoundError(fmt.Sprintf("evaluation rule %s", id))
		}
		return nil, appErrors.NewInternalError("failed to get evaluation rule", err)
	}

	if req.Name != nil && *req.Name != rule.Name {
		exists, err := s.repo.ExistsByName(ctx, projectID, *req.Name)
		if err != nil {
			return nil, appErrors.NewInternalError("failed to check name uniqueness", err)
		}
		if exists {
			return nil, appErrors.NewConflictError(fmt.Sprintf("evaluation rule '%s' already exists in this project", *req.Name))
		}
		rule.Name = *req.Name
	}

	if req.Description != nil {
		rule.Description = req.Description
	}
	if req.Status != nil {
		rule.Status = *req.Status
	}
	if req.TriggerType != nil {
		rule.TriggerType = *req.TriggerType
	}
	if req.TargetScope != nil {
		rule.TargetScope = *req.TargetScope
	}
	if req.Filter != nil {
		rule.Filter = req.Filter
	}
	if req.SpanNames != nil {
		rule.SpanNames = req.SpanNames
	}
	if req.SamplingRate != nil {
		rule.SamplingRate = *req.SamplingRate
	}
	if req.ScorerType != nil {
		rule.ScorerType = *req.ScorerType
	}
	if req.ScorerConfig != nil {
		rule.ScorerConfig = req.ScorerConfig
	}
	if req.VariableMapping != nil {
		rule.VariableMapping = req.VariableMapping
	}

	rule.UpdatedAt = time.Now()

	if validationErrors := rule.Validate(); len(validationErrors) > 0 {
		return nil, appErrors.NewValidationError(validationErrors[0].Field, validationErrors[0].Message)
	}

	if err := s.repo.Update(ctx, rule); err != nil {
		if errors.Is(err, evaluation.ErrRuleNotFound) {
			return nil, appErrors.NewNotFoundError(fmt.Sprintf("evaluation rule %s", id))
		}
		if errors.Is(err, evaluation.ErrRuleExists) {
			return nil, appErrors.NewConflictError(fmt.Sprintf("evaluation rule '%s' already exists in this project", rule.Name))
		}
		return nil, appErrors.NewInternalError("failed to update evaluation rule", err)
	}

	s.logger.Info("evaluation rule updated",
		"rule_id", id,
		"project_id", projectID,
	)

	return rule, nil
}

func (s *ruleService) Delete(ctx context.Context, id ulid.ULID, projectID ulid.ULID) error {
	rule, err := s.repo.GetByID(ctx, id, projectID)
	if err != nil {
		if errors.Is(err, evaluation.ErrRuleNotFound) {
			return appErrors.NewNotFoundError(fmt.Sprintf("evaluation rule %s", id))
		}
		return appErrors.NewInternalError("failed to get evaluation rule", err)
	}

	if err := s.repo.Delete(ctx, id, projectID); err != nil {
		if errors.Is(err, evaluation.ErrRuleNotFound) {
			return appErrors.NewNotFoundError(fmt.Sprintf("evaluation rule %s", id))
		}
		return appErrors.NewInternalError("failed to delete evaluation rule", err)
	}

	s.logger.Info("evaluation rule deleted",
		"rule_id", id,
		"project_id", projectID,
		"name", rule.Name,
	)

	return nil
}

func (s *ruleService) GetByID(ctx context.Context, id ulid.ULID, projectID ulid.ULID) (*evaluation.EvaluationRule, error) {
	rule, err := s.repo.GetByID(ctx, id, projectID)
	if err != nil {
		if errors.Is(err, evaluation.ErrRuleNotFound) {
			return nil, appErrors.NewNotFoundError(fmt.Sprintf("evaluation rule %s", id))
		}
		return nil, appErrors.NewInternalError("failed to get evaluation rule", err)
	}
	return rule, nil
}

func (s *ruleService) List(ctx context.Context, projectID ulid.ULID, filter *evaluation.RuleFilter, params pagination.Params) ([]*evaluation.EvaluationRule, int64, error) {
	rules, total, err := s.repo.GetByProjectID(ctx, projectID, filter, params)
	if err != nil {
		return nil, 0, appErrors.NewInternalError("failed to list evaluation rules", err)
	}
	return rules, total, nil
}

func (s *ruleService) Activate(ctx context.Context, id ulid.ULID, projectID ulid.ULID) error {
	rule, err := s.repo.GetByID(ctx, id, projectID)
	if err != nil {
		if errors.Is(err, evaluation.ErrRuleNotFound) {
			return appErrors.NewNotFoundError(fmt.Sprintf("evaluation rule %s", id))
		}
		return appErrors.NewInternalError("failed to get evaluation rule", err)
	}

	if rule.Status == evaluation.RuleStatusActive {
		return nil
	}

	rule.Status = evaluation.RuleStatusActive
	rule.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, rule); err != nil {
		return appErrors.NewInternalError("failed to activate evaluation rule", err)
	}

	s.logger.Info("evaluation rule activated",
		"rule_id", id,
		"project_id", projectID,
		"name", rule.Name,
	)

	return nil
}

func (s *ruleService) Deactivate(ctx context.Context, id ulid.ULID, projectID ulid.ULID) error {
	rule, err := s.repo.GetByID(ctx, id, projectID)
	if err != nil {
		if errors.Is(err, evaluation.ErrRuleNotFound) {
			return appErrors.NewNotFoundError(fmt.Sprintf("evaluation rule %s", id))
		}
		return appErrors.NewInternalError("failed to get evaluation rule", err)
	}

	if rule.Status == evaluation.RuleStatusInactive {
		return nil
	}

	rule.Status = evaluation.RuleStatusInactive
	rule.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, rule); err != nil {
		return appErrors.NewInternalError("failed to deactivate evaluation rule", err)
	}

	s.logger.Info("evaluation rule deactivated",
		"rule_id", id,
		"project_id", projectID,
		"name", rule.Name,
	)

	return nil
}

func (s *ruleService) GetActiveByProjectID(ctx context.Context, projectID ulid.ULID) ([]*evaluation.EvaluationRule, error) {
	rules, err := s.repo.GetActiveByProjectID(ctx, projectID)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get active evaluation rules", err)
	}
	return rules, nil
}

func (s *ruleService) TriggerRule(ctx context.Context, ruleID ulid.ULID, projectID ulid.ULID, opts *evaluation.TriggerOptions) (*evaluation.TriggerResponse, error) {
	// Validate rule exists (can trigger inactive rules for testing)
	rule, err := s.repo.GetByID(ctx, ruleID, projectID)
	if err != nil {
		if errors.Is(err, evaluation.ErrRuleNotFound) {
			return nil, appErrors.NewNotFoundError(fmt.Sprintf("evaluation rule %s", ruleID))
		}
		return nil, appErrors.NewInternalError("failed to get evaluation rule", err)
	}

	execution, err := s.executionService.StartExecution(ctx, ruleID, projectID, evaluation.TriggerTypeManual)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to create execution record", err)
	}

	triggerMsg := ManualTriggerMessage{
		ExecutionID:    execution.ID,
		RuleID:         ruleID,
		ProjectID:      projectID,
		ScorerType:     rule.ScorerType,
		ScorerConfig:   rule.ScorerConfig,
		TargetScope:    rule.TargetScope,
		Filter:         rule.Filter,
		SpanNames:      rule.SpanNames,
		SamplingRate:   rule.SamplingRate,
		VariableMapping: rule.VariableMapping,
		CreatedAt:      time.Now(),
	}

	if opts != nil {
		triggerMsg.TimeRangeStart = opts.TimeRangeStart
		triggerMsg.TimeRangeEnd = opts.TimeRangeEnd
		triggerMsg.SpanIDs = opts.SpanIDs
		if opts.SampleLimit > 0 {
			triggerMsg.SampleLimit = opts.SampleLimit
		} else {
			triggerMsg.SampleLimit = 1000 // Default limit
		}
	} else {
		triggerMsg.SampleLimit = 1000
	}

	msgData, err := json.Marshal(triggerMsg)
	if err != nil {
		// Fail the execution since we can't publish
		_ = s.executionService.FailExecution(ctx, execution.ID, projectID, "failed to serialize trigger message")
		return nil, appErrors.NewInternalError("failed to serialize trigger message", err)
	}

	_, err = s.redis.Client.XAdd(ctx, &redis.XAddArgs{
		Stream: manualTriggerStream,
		Values: map[string]interface{}{
			"data": string(msgData),
		},
	}).Result()
	if err != nil {
		// Fail the execution since we can't publish
		_ = s.executionService.FailExecution(ctx, execution.ID, projectID, "failed to queue trigger job")
		return nil, appErrors.NewInternalError("failed to queue manual trigger job", err)
	}

	s.logger.Info("manual evaluation triggered",
		"rule_id", ruleID,
		"project_id", projectID,
		"execution_id", execution.ID,
		"rule_name", rule.Name,
	)

	return &evaluation.TriggerResponse{
		ExecutionID: execution.ID.String(),
		SpansQueued: 0, // Will be updated by worker when it starts processing
		Message:     "Manual evaluation queued successfully",
	}, nil
}

// ManualTriggerMessage is the message format for the manual trigger stream
type ManualTriggerMessage struct {
	ExecutionID     ulid.ULID                 `json:"execution_id"`
	RuleID          ulid.ULID                 `json:"rule_id"`
	ProjectID       ulid.ULID                 `json:"project_id"`
	ScorerType      evaluation.ScorerType     `json:"scorer_type"`
	ScorerConfig    map[string]any            `json:"scorer_config"`
	TargetScope     evaluation.TargetScope    `json:"target_scope"`
	Filter          []evaluation.FilterClause `json:"filter,omitempty"`
	SpanNames       []string                  `json:"span_names,omitempty"`
	SamplingRate    float64                   `json:"sampling_rate"`
	VariableMapping []evaluation.VariableMap  `json:"variable_mapping,omitempty"`
	TimeRangeStart  *time.Time                `json:"time_range_start,omitempty"`
	TimeRangeEnd    *time.Time                `json:"time_range_end,omitempty"`
	SpanIDs         []string                  `json:"span_ids,omitempty"`
	SampleLimit     int                       `json:"sample_limit"`
	CreatedAt       time.Time                 `json:"created_at"`
}
