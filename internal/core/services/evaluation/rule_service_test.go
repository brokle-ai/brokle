package evaluation

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"brokle/internal/core/domain/evaluation"
	"brokle/internal/core/domain/observability"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/pagination"
	"brokle/pkg/ulid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRuleRepository implements the RuleRepository interface for testing
type MockRuleRepository struct {
	mock.Mock
}

func (m *MockRuleRepository) Create(ctx context.Context, rule *evaluation.EvaluationRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockRuleRepository) Update(ctx context.Context, rule *evaluation.EvaluationRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockRuleRepository) Delete(ctx context.Context, id ulid.ULID, projectID ulid.ULID) error {
	args := m.Called(ctx, id, projectID)
	return args.Error(0)
}

func (m *MockRuleRepository) GetByID(ctx context.Context, id ulid.ULID, projectID ulid.ULID) (*evaluation.EvaluationRule, error) {
	args := m.Called(ctx, id, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*evaluation.EvaluationRule), args.Error(1)
}

func (m *MockRuleRepository) GetByProjectID(ctx context.Context, projectID ulid.ULID, filter *evaluation.RuleFilter, params pagination.Params) ([]*evaluation.EvaluationRule, int64, error) {
	args := m.Called(ctx, projectID, filter, params)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*evaluation.EvaluationRule), args.Get(1).(int64), args.Error(2)
}

func (m *MockRuleRepository) GetActiveByProjectID(ctx context.Context, projectID ulid.ULID) ([]*evaluation.EvaluationRule, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*evaluation.EvaluationRule), args.Error(1)
}

func (m *MockRuleRepository) ExistsByName(ctx context.Context, projectID ulid.ULID, name string) (bool, error) {
	args := m.Called(ctx, projectID, name)
	return args.Bool(0), args.Error(1)
}

// MockTraceRepository implements a minimal TraceRepository for testing
type MockTraceRepository struct {
	mock.Mock
}

func (m *MockTraceRepository) InsertSpan(ctx context.Context, span *observability.Span) error {
	return nil
}
func (m *MockTraceRepository) InsertSpanBatch(ctx context.Context, spans []*observability.Span) error {
	return nil
}
func (m *MockTraceRepository) DeleteSpan(ctx context.Context, spanID string) error {
	return nil
}
func (m *MockTraceRepository) GetSpan(ctx context.Context, spanID string) (*observability.Span, error) {
	return nil, nil
}
func (m *MockTraceRepository) GetSpansByTraceID(ctx context.Context, traceID string) ([]*observability.Span, error) {
	return nil, nil
}
func (m *MockTraceRepository) GetSpanChildren(ctx context.Context, parentSpanID string) ([]*observability.Span, error) {
	return nil, nil
}
func (m *MockTraceRepository) GetSpanTree(ctx context.Context, traceID string) ([]*observability.Span, error) {
	return nil, nil
}
func (m *MockTraceRepository) GetSpansByFilter(ctx context.Context, filter *observability.SpanFilter) ([]*observability.Span, error) {
	return nil, nil
}
func (m *MockTraceRepository) CountSpansByFilter(ctx context.Context, filter *observability.SpanFilter) (int64, error) {
	return 0, nil
}
func (m *MockTraceRepository) GetRootSpan(ctx context.Context, traceID string) (*observability.Span, error) {
	return nil, nil
}
func (m *MockTraceRepository) GetRootSpanByProject(ctx context.Context, traceID string, projectID string) (*observability.Span, error) {
	return nil, nil
}
func (m *MockTraceRepository) GetSpanByProject(ctx context.Context, spanID string, projectID string) (*observability.Span, error) {
	return nil, nil
}
func (m *MockTraceRepository) GetTraceSummary(ctx context.Context, traceID string) (*observability.TraceSummary, error) {
	return nil, nil
}
func (m *MockTraceRepository) ListTraces(ctx context.Context, filter *observability.TraceFilter) ([]*observability.TraceSummary, error) {
	return nil, nil
}
func (m *MockTraceRepository) CountTraces(ctx context.Context, filter *observability.TraceFilter) (int64, error) {
	return 0, nil
}
func (m *MockTraceRepository) CountSpansInTrace(ctx context.Context, traceID string) (int64, error) {
	return 0, nil
}
func (m *MockTraceRepository) DeleteTrace(ctx context.Context, traceID string) error {
	return nil
}
func (m *MockTraceRepository) UpdateTraceTags(ctx context.Context, projectID, traceID string, tags []string) error {
	return nil
}
func (m *MockTraceRepository) UpdateTraceBookmark(ctx context.Context, projectID, traceID string, bookmarked bool) error {
	return nil
}
func (m *MockTraceRepository) GetFilterOptions(ctx context.Context, projectID string) (*observability.TraceFilterOptions, error) {
	return nil, nil
}
func (m *MockTraceRepository) GetTracesBySessionID(ctx context.Context, sessionID string) ([]*observability.TraceSummary, error) {
	return nil, nil
}
func (m *MockTraceRepository) GetTracesByUserID(ctx context.Context, userID string, filter *observability.TraceFilter) ([]*observability.TraceSummary, error) {
	return nil, nil
}
func (m *MockTraceRepository) CalculateTotalCost(ctx context.Context, traceID string) (float64, error) {
	return 0, nil
}
func (m *MockTraceRepository) CalculateTotalTokens(ctx context.Context, traceID string) (uint64, error) {
	return 0, nil
}
func (m *MockTraceRepository) QuerySpansByExpression(ctx context.Context, query string, args []interface{}) ([]*observability.Span, error) {
	return nil, nil
}
func (m *MockTraceRepository) CountSpansByExpression(ctx context.Context, query string, args []interface{}) (int64, error) {
	return 0, nil
}
func (m *MockTraceRepository) DiscoverAttributes(ctx context.Context, req *observability.AttributeDiscoveryRequest) (*observability.AttributeDiscoveryResponse, error) {
	return nil, nil
}
func (m *MockTraceRepository) ListSessions(ctx context.Context, filter *observability.SessionFilter) ([]*observability.SessionSummary, error) {
	return nil, nil
}
func (m *MockTraceRepository) CountSessions(ctx context.Context, filter *observability.SessionFilter) (int64, error) {
	return 0, nil
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestRuleService_Create(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()
	projectID := ulid.New()

	t.Run("success with valid rule", func(t *testing.T) {
		repo := new(MockRuleRepository)
		service := NewRuleService(repo, nil, new(MockTraceRepository), nil, logger)

		req := &evaluation.CreateEvaluationRuleRequest{
			Name:       "Test Rule",
			ScorerType: evaluation.ScorerTypeLLM,
			ScorerConfig: map[string]interface{}{
				"model": "gpt-4",
			},
		}

		repo.On("ExistsByName", ctx, projectID, "Test Rule").Return(false, nil)
		repo.On("Create", ctx, mock.AnythingOfType("*evaluation.EvaluationRule")).Return(nil)

		rule, err := service.Create(ctx, projectID, nil, req)

		assert.NoError(t, err)
		assert.NotNil(t, rule)
		assert.Equal(t, "Test Rule", rule.Name)
		assert.Equal(t, evaluation.ScorerTypeLLM, rule.ScorerType)
		repo.AssertExpectations(t)
	})

	t.Run("reject duplicate name", func(t *testing.T) {
		repo := new(MockRuleRepository)
		service := NewRuleService(repo, nil, new(MockTraceRepository), nil, logger)

		req := &evaluation.CreateEvaluationRuleRequest{
			Name:       "Existing Rule",
			ScorerType: evaluation.ScorerTypeBuiltin,
			ScorerConfig: map[string]interface{}{
				"scorer_name": "json_valid",
			},
		}

		repo.On("ExistsByName", ctx, projectID, "Existing Rule").Return(true, nil)

		rule, err := service.Create(ctx, projectID, nil, req)

		assert.Error(t, err)
		assert.Nil(t, rule)
		var appErr *appErrors.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, appErrors.ConflictError, appErr.Type)
		assert.Contains(t, appErr.Message, "already exists")
		repo.AssertExpectations(t)
		repo.AssertNotCalled(t, "Create")
	})

	t.Run("reject invalid name (empty)", func(t *testing.T) {
		repo := new(MockRuleRepository)
		service := NewRuleService(repo, nil, new(MockTraceRepository), nil, logger)

		req := &evaluation.CreateEvaluationRuleRequest{
			Name:       "",
			ScorerType: evaluation.ScorerTypeRegex,
			ScorerConfig: map[string]interface{}{
				"pattern": "test",
			},
		}

		rule, err := service.Create(ctx, projectID, nil, req)

		assert.Error(t, err)
		assert.Nil(t, rule)
		var appErr *appErrors.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, appErrors.ValidationError, appErr.Type)
		repo.AssertNotCalled(t, "ExistsByName")
		repo.AssertNotCalled(t, "Create")
	})

	t.Run("sets created_by when user ID provided", func(t *testing.T) {
		repo := new(MockRuleRepository)
		service := NewRuleService(repo, nil, new(MockTraceRepository), nil, logger)

		userID := ulid.New()
		req := &evaluation.CreateEvaluationRuleRequest{
			Name:       "Rule with Creator",
			ScorerType: evaluation.ScorerTypeLLM,
			ScorerConfig: map[string]interface{}{
				"model": "gpt-4",
			},
		}

		repo.On("ExistsByName", ctx, projectID, "Rule with Creator").Return(false, nil)
		repo.On("Create", ctx, mock.MatchedBy(func(rule *evaluation.EvaluationRule) bool {
			return rule.CreatedBy != nil && *rule.CreatedBy == userID.String()
		})).Return(nil)

		rule, err := service.Create(ctx, projectID, &userID, req)

		assert.NoError(t, err)
		assert.NotNil(t, rule)
		assert.NotNil(t, rule.CreatedBy)
		assert.Equal(t, userID.String(), *rule.CreatedBy)
		repo.AssertExpectations(t)
	})
}

func TestRuleService_Update(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()
	projectID := ulid.New()
	ruleID := ulid.New()

	t.Run("success with partial update", func(t *testing.T) {
		repo := new(MockRuleRepository)
		service := NewRuleService(repo, nil, new(MockTraceRepository), nil, logger)

		existingRule := &evaluation.EvaluationRule{
			ID:          ruleID,
			ProjectID:   projectID,
			Name:        "Original Name",
			ScorerType:  evaluation.ScorerTypeLLM,
			TriggerType: evaluation.RuleTriggerOnSpanComplete,
			TargetScope: evaluation.TargetScopeSpan,
			ScorerConfig: map[string]interface{}{
				"model": "gpt-4",
			},
			Status: evaluation.RuleStatusInactive,
		}

		newDescription := "Updated description"
		req := &evaluation.UpdateEvaluationRuleRequest{
			Description: &newDescription,
		}

		repo.On("GetByID", ctx, ruleID, projectID).Return(existingRule, nil)
		repo.On("Update", ctx, mock.AnythingOfType("*evaluation.EvaluationRule")).Return(nil)

		rule, err := service.Update(ctx, ruleID, projectID, req)

		assert.NoError(t, err)
		assert.NotNil(t, rule)
		assert.Equal(t, "Original Name", rule.Name)
		assert.Equal(t, &newDescription, rule.Description)
		repo.AssertExpectations(t)
	})

	t.Run("reject name conflict on rename", func(t *testing.T) {
		repo := new(MockRuleRepository)
		service := NewRuleService(repo, nil, new(MockTraceRepository), nil, logger)

		existingRule := &evaluation.EvaluationRule{
			ID:          ruleID,
			ProjectID:   projectID,
			Name:        "Original Name",
			ScorerType:  evaluation.ScorerTypeLLM,
			TriggerType: evaluation.RuleTriggerOnSpanComplete,
			TargetScope: evaluation.TargetScopeSpan,
			ScorerConfig: map[string]interface{}{
				"model": "gpt-4",
			},
			Status: evaluation.RuleStatusInactive,
		}

		newName := "Conflicting Name"
		req := &evaluation.UpdateEvaluationRuleRequest{
			Name: &newName,
		}

		repo.On("GetByID", ctx, ruleID, projectID).Return(existingRule, nil)
		repo.On("ExistsByName", ctx, projectID, "Conflicting Name").Return(true, nil)

		rule, err := service.Update(ctx, ruleID, projectID, req)

		assert.Error(t, err)
		assert.Nil(t, rule)
		var appErr *appErrors.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, appErrors.ConflictError, appErr.Type)
		assert.Contains(t, appErr.Message, "already exists")
		repo.AssertExpectations(t)
		repo.AssertNotCalled(t, "Update")
	})

	t.Run("not found error", func(t *testing.T) {
		repo := new(MockRuleRepository)
		service := NewRuleService(repo, nil, new(MockTraceRepository), nil, logger)

		newName := "New Name"
		req := &evaluation.UpdateEvaluationRuleRequest{
			Name: &newName,
		}

		repo.On("GetByID", ctx, ruleID, projectID).Return(nil, evaluation.ErrRuleNotFound)

		rule, err := service.Update(ctx, ruleID, projectID, req)

		assert.Error(t, err)
		assert.Nil(t, rule)
		var appErr *appErrors.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, appErrors.NotFoundError, appErr.Type)
		repo.AssertExpectations(t)
	})

	t.Run("allow rename to same name", func(t *testing.T) {
		repo := new(MockRuleRepository)
		service := NewRuleService(repo, nil, new(MockTraceRepository), nil, logger)

		existingRule := &evaluation.EvaluationRule{
			ID:          ruleID,
			ProjectID:   projectID,
			Name:        "Same Name",
			ScorerType:  evaluation.ScorerTypeLLM,
			TriggerType: evaluation.RuleTriggerOnSpanComplete,
			TargetScope: evaluation.TargetScopeSpan,
			ScorerConfig: map[string]interface{}{
				"model": "gpt-4",
			},
			Status: evaluation.RuleStatusInactive,
		}

		sameName := "Same Name"
		req := &evaluation.UpdateEvaluationRuleRequest{
			Name: &sameName,
		}

		repo.On("GetByID", ctx, ruleID, projectID).Return(existingRule, nil)
		repo.On("Update", ctx, mock.AnythingOfType("*evaluation.EvaluationRule")).Return(nil)

		rule, err := service.Update(ctx, ruleID, projectID, req)

		assert.NoError(t, err)
		assert.NotNil(t, rule)
		assert.Equal(t, "Same Name", rule.Name)
		repo.AssertExpectations(t)
		repo.AssertNotCalled(t, "ExistsByName")
	})
}

func TestRuleService_Activate(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()
	projectID := ulid.New()
	ruleID := ulid.New()

	t.Run("success", func(t *testing.T) {
		repo := new(MockRuleRepository)
		service := NewRuleService(repo, nil, new(MockTraceRepository), nil, logger)

		existingRule := &evaluation.EvaluationRule{
			ID:          ruleID,
			ProjectID:   projectID,
			Name:        "Test Rule",
			ScorerType:  evaluation.ScorerTypeLLM,
			TriggerType: evaluation.RuleTriggerOnSpanComplete,
			TargetScope: evaluation.TargetScopeSpan,
			ScorerConfig: map[string]interface{}{
				"model": "gpt-4",
			},
			Status: evaluation.RuleStatusInactive,
		}

		repo.On("GetByID", ctx, ruleID, projectID).Return(existingRule, nil)
		repo.On("Update", ctx, mock.MatchedBy(func(rule *evaluation.EvaluationRule) bool {
			return rule.Status == evaluation.RuleStatusActive
		})).Return(nil)

		err := service.Activate(ctx, ruleID, projectID)

		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("idempotent - already active", func(t *testing.T) {
		repo := new(MockRuleRepository)
		service := NewRuleService(repo, nil, new(MockTraceRepository), nil, logger)

		existingRule := &evaluation.EvaluationRule{
			ID:          ruleID,
			ProjectID:   projectID,
			Name:        "Test Rule",
			ScorerType:  evaluation.ScorerTypeLLM,
			TriggerType: evaluation.RuleTriggerOnSpanComplete,
			TargetScope: evaluation.TargetScopeSpan,
			ScorerConfig: map[string]interface{}{
				"model": "gpt-4",
			},
			Status: evaluation.RuleStatusActive,
		}

		repo.On("GetByID", ctx, ruleID, projectID).Return(existingRule, nil)

		err := service.Activate(ctx, ruleID, projectID)

		assert.NoError(t, err)
		repo.AssertExpectations(t)
		repo.AssertNotCalled(t, "Update")
	})

	t.Run("not found error", func(t *testing.T) {
		repo := new(MockRuleRepository)
		service := NewRuleService(repo, nil, new(MockTraceRepository), nil, logger)

		repo.On("GetByID", ctx, ruleID, projectID).Return(nil, evaluation.ErrRuleNotFound)

		err := service.Activate(ctx, ruleID, projectID)

		assert.Error(t, err)
		var appErr *appErrors.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, appErrors.NotFoundError, appErr.Type)
		repo.AssertExpectations(t)
	})
}

func TestRuleService_Deactivate(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()
	projectID := ulid.New()
	ruleID := ulid.New()

	t.Run("success", func(t *testing.T) {
		repo := new(MockRuleRepository)
		service := NewRuleService(repo, nil, new(MockTraceRepository), nil, logger)

		existingRule := &evaluation.EvaluationRule{
			ID:          ruleID,
			ProjectID:   projectID,
			Name:        "Test Rule",
			ScorerType:  evaluation.ScorerTypeLLM,
			TriggerType: evaluation.RuleTriggerOnSpanComplete,
			TargetScope: evaluation.TargetScopeSpan,
			ScorerConfig: map[string]interface{}{
				"model": "gpt-4",
			},
			Status: evaluation.RuleStatusActive,
		}

		repo.On("GetByID", ctx, ruleID, projectID).Return(existingRule, nil)
		repo.On("Update", ctx, mock.MatchedBy(func(rule *evaluation.EvaluationRule) bool {
			return rule.Status == evaluation.RuleStatusInactive
		})).Return(nil)

		err := service.Deactivate(ctx, ruleID, projectID)

		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("idempotent - already inactive", func(t *testing.T) {
		repo := new(MockRuleRepository)
		service := NewRuleService(repo, nil, new(MockTraceRepository), nil, logger)

		existingRule := &evaluation.EvaluationRule{
			ID:          ruleID,
			ProjectID:   projectID,
			Name:        "Test Rule",
			ScorerType:  evaluation.ScorerTypeLLM,
			TriggerType: evaluation.RuleTriggerOnSpanComplete,
			TargetScope: evaluation.TargetScopeSpan,
			ScorerConfig: map[string]interface{}{
				"model": "gpt-4",
			},
			Status: evaluation.RuleStatusInactive,
		}

		repo.On("GetByID", ctx, ruleID, projectID).Return(existingRule, nil)

		err := service.Deactivate(ctx, ruleID, projectID)

		assert.NoError(t, err)
		repo.AssertExpectations(t)
		repo.AssertNotCalled(t, "Update")
	})

	t.Run("not found error", func(t *testing.T) {
		repo := new(MockRuleRepository)
		service := NewRuleService(repo, nil, new(MockTraceRepository), nil, logger)

		repo.On("GetByID", ctx, ruleID, projectID).Return(nil, evaluation.ErrRuleNotFound)

		err := service.Deactivate(ctx, ruleID, projectID)

		assert.Error(t, err)
		var appErr *appErrors.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, appErrors.NotFoundError, appErr.Type)
		repo.AssertExpectations(t)
	})
}

func TestRuleService_Delete(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()
	projectID := ulid.New()
	ruleID := ulid.New()

	t.Run("success", func(t *testing.T) {
		repo := new(MockRuleRepository)
		service := NewRuleService(repo, nil, new(MockTraceRepository), nil, logger)

		existingRule := &evaluation.EvaluationRule{
			ID:          ruleID,
			ProjectID:   projectID,
			Name:        "Rule to Delete",
			ScorerType:  evaluation.ScorerTypeLLM,
			TriggerType: evaluation.RuleTriggerOnSpanComplete,
			TargetScope: evaluation.TargetScopeSpan,
			ScorerConfig: map[string]interface{}{
				"model": "gpt-4",
			},
		}

		repo.On("GetByID", ctx, ruleID, projectID).Return(existingRule, nil)
		repo.On("Delete", ctx, ruleID, projectID).Return(nil)

		err := service.Delete(ctx, ruleID, projectID)

		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("not found error on GetByID", func(t *testing.T) {
		repo := new(MockRuleRepository)
		service := NewRuleService(repo, nil, new(MockTraceRepository), nil, logger)

		repo.On("GetByID", ctx, ruleID, projectID).Return(nil, evaluation.ErrRuleNotFound)

		err := service.Delete(ctx, ruleID, projectID)

		assert.Error(t, err)
		var appErr *appErrors.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, appErrors.NotFoundError, appErr.Type)
		repo.AssertExpectations(t)
		repo.AssertNotCalled(t, "Delete")
	})
}

func TestRuleService_GetByID(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()
	projectID := ulid.New()
	ruleID := ulid.New()

	t.Run("success", func(t *testing.T) {
		repo := new(MockRuleRepository)
		service := NewRuleService(repo, nil, new(MockTraceRepository), nil, logger)

		existingRule := &evaluation.EvaluationRule{
			ID:          ruleID,
			ProjectID:   projectID,
			Name:        "Test Rule",
			ScorerType:  evaluation.ScorerTypeLLM,
			TriggerType: evaluation.RuleTriggerOnSpanComplete,
			TargetScope: evaluation.TargetScopeSpan,
			ScorerConfig: map[string]interface{}{
				"model": "gpt-4",
			},
		}

		repo.On("GetByID", ctx, ruleID, projectID).Return(existingRule, nil)

		rule, err := service.GetByID(ctx, ruleID, projectID)

		assert.NoError(t, err)
		assert.NotNil(t, rule)
		assert.Equal(t, ruleID, rule.ID)
		assert.Equal(t, "Test Rule", rule.Name)
		repo.AssertExpectations(t)
	})

	t.Run("not found error", func(t *testing.T) {
		repo := new(MockRuleRepository)
		service := NewRuleService(repo, nil, new(MockTraceRepository), nil, logger)

		repo.On("GetByID", ctx, ruleID, projectID).Return(nil, evaluation.ErrRuleNotFound)

		rule, err := service.GetByID(ctx, ruleID, projectID)

		assert.Error(t, err)
		assert.Nil(t, rule)
		var appErr *appErrors.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, appErrors.NotFoundError, appErr.Type)
		repo.AssertExpectations(t)
	})
}

func TestRuleService_List(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()
	projectID := ulid.New()

	t.Run("success with results", func(t *testing.T) {
		repo := new(MockRuleRepository)
		service := NewRuleService(repo, nil, new(MockTraceRepository), nil, logger)

		expectedRules := []*evaluation.EvaluationRule{
			{
				ID:         ulid.New(),
				ProjectID:  projectID,
				Name:       "Rule 1",
				ScorerType: evaluation.ScorerTypeLLM,
			},
			{
				ID:         ulid.New(),
				ProjectID:  projectID,
				Name:       "Rule 2",
				ScorerType: evaluation.ScorerTypeBuiltin,
			},
		}

		params := pagination.Params{Page: 1, Limit: 10}
		repo.On("GetByProjectID", ctx, projectID, (*evaluation.RuleFilter)(nil), params).Return(expectedRules, int64(2), nil)

		rules, total, err := service.List(ctx, projectID, nil, params)

		assert.NoError(t, err)
		assert.Len(t, rules, 2)
		assert.Equal(t, int64(2), total)
		repo.AssertExpectations(t)
	})

	t.Run("success with filter", func(t *testing.T) {
		repo := new(MockRuleRepository)
		service := NewRuleService(repo, nil, new(MockTraceRepository), nil, logger)

		status := evaluation.RuleStatusActive
		filter := &evaluation.RuleFilter{
			Status: &status,
		}

		expectedRules := []*evaluation.EvaluationRule{
			{
				ID:         ulid.New(),
				ProjectID:  projectID,
				Name:       "Active Rule",
				ScorerType: evaluation.ScorerTypeLLM,
				Status:     evaluation.RuleStatusActive,
			},
		}

		params := pagination.Params{Page: 1, Limit: 10}
		repo.On("GetByProjectID", ctx, projectID, filter, params).Return(expectedRules, int64(1), nil)

		rules, total, err := service.List(ctx, projectID, filter, params)

		assert.NoError(t, err)
		assert.Len(t, rules, 1)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, evaluation.RuleStatusActive, rules[0].Status)
		repo.AssertExpectations(t)
	})

	t.Run("success with empty results", func(t *testing.T) {
		repo := new(MockRuleRepository)
		service := NewRuleService(repo, nil, new(MockTraceRepository), nil, logger)

		params := pagination.Params{Page: 1, Limit: 10}
		repo.On("GetByProjectID", ctx, projectID, (*evaluation.RuleFilter)(nil), params).Return([]*evaluation.EvaluationRule{}, int64(0), nil)

		rules, total, err := service.List(ctx, projectID, nil, params)

		assert.NoError(t, err)
		assert.Empty(t, rules)
		assert.Equal(t, int64(0), total)
		repo.AssertExpectations(t)
	})
}
