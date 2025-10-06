package observability

import (
	"context"
	"testing"
	"time"

	"brokle/internal/core/domain/observability"
	"brokle/pkg/ulid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ============================================================================
// Mock QualityScoreRepository
// ============================================================================

type MockQualityScoreRepository struct {
	mock.Mock
}

func (m *MockQualityScoreRepository) Create(ctx context.Context, score *observability.QualityScore) error {
	args := m.Called(ctx, score)
	return args.Error(0)
}

func (m *MockQualityScoreRepository) GetByID(ctx context.Context, id ulid.ULID) (*observability.QualityScore, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*observability.QualityScore), args.Error(1)
}

func (m *MockQualityScoreRepository) Update(ctx context.Context, score *observability.QualityScore) error {
	args := m.Called(ctx, score)
	return args.Error(0)
}

func (m *MockQualityScoreRepository) Delete(ctx context.Context, id ulid.ULID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockQualityScoreRepository) GetByTraceID(ctx context.Context, traceID ulid.ULID) ([]*observability.QualityScore, error) {
	args := m.Called(ctx, traceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.QualityScore), args.Error(1)
}

func (m *MockQualityScoreRepository) GetByObservationID(ctx context.Context, observationID ulid.ULID) ([]*observability.QualityScore, error) {
	args := m.Called(ctx, observationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.QualityScore), args.Error(1)
}

func (m *MockQualityScoreRepository) GetByScoreName(ctx context.Context, scoreName string, limit, offset int) ([]*observability.QualityScore, error) {
	args := m.Called(ctx, scoreName, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.QualityScore), args.Error(1)
}

func (m *MockQualityScoreRepository) CreateBatch(ctx context.Context, scores []*observability.QualityScore) error {
	args := m.Called(ctx, scores)
	return args.Error(0)
}

func (m *MockQualityScoreRepository) UpdateBatch(ctx context.Context, scores []*observability.QualityScore) error {
	args := m.Called(ctx, scores)
	return args.Error(0)
}

func (m *MockQualityScoreRepository) DeleteBatch(ctx context.Context, ids []ulid.ULID) error {
	args := m.Called(ctx, ids)
	return args.Error(0)
}

func (m *MockQualityScoreRepository) GetAverageScoreByName(ctx context.Context, scoreName string, filter *observability.QualityScoreFilter) (float64, error) {
	args := m.Called(ctx, scoreName, filter)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockQualityScoreRepository) GetScoreDistribution(ctx context.Context, scoreName string, filter *observability.QualityScoreFilter) (map[string]int, error) {
	args := m.Called(ctx, scoreName, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int), args.Error(1)
}

func (m *MockQualityScoreRepository) GetScoresByTimeRange(ctx context.Context, filter *observability.QualityScoreFilter, startTime, endTime time.Time) ([]*observability.QualityScore, error) {
	args := m.Called(ctx, filter, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.QualityScore), args.Error(1)
}

func (m *MockQualityScoreRepository) GetBySource(ctx context.Context, source observability.ScoreSource, limit, offset int) ([]*observability.QualityScore, error) {
	args := m.Called(ctx, source, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.QualityScore), args.Error(1)
}

func (m *MockQualityScoreRepository) GetByEvaluator(ctx context.Context, evaluatorName string, limit, offset int) ([]*observability.QualityScore, error) {
	args := m.Called(ctx, evaluatorName, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.QualityScore), args.Error(1)
}

func (m *MockQualityScoreRepository) GetByTraceAndScoreName(ctx context.Context, traceID ulid.ULID, scoreName string) (*observability.QualityScore, error) {
	args := m.Called(ctx, traceID, scoreName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*observability.QualityScore), args.Error(1)
}

func (m *MockQualityScoreRepository) GetByObservationAndScoreName(ctx context.Context, observationID ulid.ULID, scoreName string) (*observability.QualityScore, error) {
	args := m.Called(ctx, observationID, scoreName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*observability.QualityScore), args.Error(1)
}

// ============================================================================
// Mock QualityEvaluator
// ============================================================================

type MockQualityEvaluator struct {
	mock.Mock
}

func (m *MockQualityEvaluator) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockQualityEvaluator) Version() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockQualityEvaluator) Description() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockQualityEvaluator) SupportedTypes() []observability.ObservationType {
	args := m.Called()
	return args.Get(0).([]observability.ObservationType)
}

func (m *MockQualityEvaluator) ValidateInput(input *observability.EvaluationInput) error {
	args := m.Called(input)
	return args.Error(0)
}

func (m *MockQualityEvaluator) Evaluate(ctx context.Context, input *observability.EvaluationInput) (*observability.QualityScore, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*observability.QualityScore), args.Error(1)
}

// ============================================================================
// Core CRUD Tests
// ============================================================================

// TestQualityService_CreateQualityScore tests the CreateQualityScore method

// ============================================================================
// HIGH-VALUE TESTS: Evaluator Management & Analytics
// ============================================================================

func TestQualityService_RegisterEvaluator(t *testing.T) {
	tests := []struct {
		name        string
		evaluator   observability.QualityEvaluator
		mockSetup   func(*MockQualityEvaluator)
		expectedErr error
	}{
		{
			name: "success - register evaluator",
			mockSetup: func(evaluator *MockQualityEvaluator) {
				evaluator.On("Name").Return("test-evaluator")
			},
			expectedErr: nil,
		},
		{
			name:      "error - nil evaluator",
			evaluator: nil,
			mockSetup: func(evaluator *MockQualityEvaluator) {
				// No calls expected
			},
			expectedErr: observability.NewObservabilityError(
				observability.ErrCodeValidationFailed,
				"evaluator cannot be nil",
			),
		},
		{
			name: "error - empty evaluator name",
			mockSetup: func(evaluator *MockQualityEvaluator) {
				evaluator.On("Name").Return("")
			},
			expectedErr: observability.NewObservabilityError(
				observability.ErrCodeValidationFailed,
				"evaluator name cannot be empty",
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockQSRepo := new(MockQualityScoreRepository)
			mockTraceRepo := new(MockTraceRepository)
			mockObsRepo := new(MockObservationRepository)
			mockPublisher := new(MockEventPublisher)
			mockEvaluator := new(MockQualityEvaluator)

			if tt.evaluator == nil && tt.name != "error - nil evaluator" {
				// For non-nil test cases, set up the mock evaluator
				tt.mockSetup(mockEvaluator)
			} else if tt.name != "error - nil evaluator" {
				tt.mockSetup(mockEvaluator)
			}

			service := NewQualityService(mockQSRepo, mockTraceRepo, mockObsRepo, mockPublisher)

			var err error
			if tt.name == "error - nil evaluator" {
				err = service.RegisterEvaluator(context.Background(), nil)
			} else {
				err = service.RegisterEvaluator(context.Background(), mockEvaluator)
			}

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestQualityService_GetEvaluator tests the GetEvaluator method
func TestQualityService_GetEvaluator(t *testing.T) {
	tests := []struct {
		name          string
		evaluatorName string
		setupService  func(observability.QualityService)
		expectedErr   error
		checkResult   func(*testing.T, observability.QualityEvaluator)
	}{
		{
			name:          "success - evaluator found",
			evaluatorName: "test-evaluator",
			setupService: func(svc observability.QualityService) {
				mockEval := new(MockQualityEvaluator)
				mockEval.On("Name").Return("test-evaluator")
				_ = svc.RegisterEvaluator(context.Background(), mockEval)
			},
			expectedErr: nil,
			checkResult: func(t *testing.T, evaluator observability.QualityEvaluator) {
				assert.NotNil(t, evaluator)
			},
		},
		{
			name:          "error - empty name",
			evaluatorName: "",
			setupService:  func(svc observability.QualityService) {},
			expectedErr: observability.NewObservabilityError(
				observability.ErrCodeValidationFailed,
				"evaluator name cannot be empty",
			),
			checkResult: nil,
		},
		{
			name:          "error - evaluator not found",
			evaluatorName: "non-existent",
			setupService:  func(svc observability.QualityService) {},
			expectedErr: observability.NewObservabilityError(
				observability.ErrCodeEvaluatorNotFound,
				"evaluator not found",
			),
			checkResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockQSRepo := new(MockQualityScoreRepository)
			mockTraceRepo := new(MockTraceRepository)
			mockObsRepo := new(MockObservationRepository)
			mockPublisher := new(MockEventPublisher)

			service := NewQualityService(mockQSRepo, mockTraceRepo, mockObsRepo, mockPublisher)
			tt.setupService(service)

			evaluator, err := service.GetEvaluator(context.Background(), tt.evaluatorName)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			if tt.checkResult != nil {
				tt.checkResult(t, evaluator)
			}
		})
	}
}

// TestQualityService_ListEvaluators tests the ListEvaluators method
func TestQualityService_ListEvaluators(t *testing.T) {
	tests := []struct {
		name         string
		setupService func(observability.QualityService)
		expectedErr  error
		checkResult  func(*testing.T, []observability.QualityEvaluatorInfo)
	}{
		{
			name: "success - list evaluators",
			setupService: func(svc observability.QualityService) {
				mockEval1 := new(MockQualityEvaluator)
				mockEval1.On("Name").Return("evaluator-1")
				mockEval1.On("Version").Return("1.0.0")
				mockEval1.On("Description").Return("Test evaluator 1")
				mockEval1.On("SupportedTypes").Return([]observability.ObservationType{observability.ObservationTypeLLM})

				mockEval2 := new(MockQualityEvaluator)
				mockEval2.On("Name").Return("evaluator-2")
				mockEval2.On("Version").Return("2.0.0")
				mockEval2.On("Description").Return("Test evaluator 2")
				mockEval2.On("SupportedTypes").Return([]observability.ObservationType{observability.ObservationTypeSpan})

				_ = svc.RegisterEvaluator(context.Background(), mockEval1)
				_ = svc.RegisterEvaluator(context.Background(), mockEval2)
			},
			expectedErr: nil,
			checkResult: func(t *testing.T, evaluators []observability.QualityEvaluatorInfo) {
				assert.Len(t, evaluators, 2)
			},
		},
		{
			name:         "success - empty list",
			setupService: func(svc observability.QualityService) {},
			expectedErr:  nil,
			checkResult: func(t *testing.T, evaluators []observability.QualityEvaluatorInfo) {
				assert.Len(t, evaluators, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockQSRepo := new(MockQualityScoreRepository)
			mockTraceRepo := new(MockTraceRepository)
			mockObsRepo := new(MockObservationRepository)
			mockPublisher := new(MockEventPublisher)

			service := NewQualityService(mockQSRepo, mockTraceRepo, mockObsRepo, mockPublisher)
			tt.setupService(service)

			evaluators, err := service.ListEvaluators(context.Background())

			if tt.expectedErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.checkResult != nil {
				tt.checkResult(t, evaluators)
			}
		})
	}
}

// ============================================================================
// Analytics Tests
// ============================================================================

// TestQualityService_GetQualityAnalytics tests the GetQualityAnalytics method
func TestQualityService_GetQualityAnalytics(t *testing.T) {
	projectID := ulid.New()

	tests := []struct {
		name        string
		filter      *observability.AnalyticsFilter
		mockSetup   func()
		expectedErr error
		checkResult func(*testing.T, *observability.QualityAnalytics)
	}{
		{
			name: "success - get quality analytics",
			filter: &observability.AnalyticsFilter{
				ProjectID: projectID,
				StartTime: time.Now().Add(-24 * time.Hour),
				EndTime:   time.Now(),
			},
			mockSetup:   func() {},
			expectedErr: nil,
			checkResult: func(t *testing.T, analytics *observability.QualityAnalytics) {
				assert.NotNil(t, analytics)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockQSRepo := new(MockQualityScoreRepository)
			mockTraceRepo := new(MockTraceRepository)
			mockObsRepo := new(MockObservationRepository)
			mockPublisher := new(MockEventPublisher)

			tt.mockSetup()

			service := NewQualityService(mockQSRepo, mockTraceRepo, mockObsRepo, mockPublisher)

			result, err := service.GetQualityAnalytics(context.Background(), tt.filter)

			if tt.expectedErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}

// TestQualityService_GetQualityTrends tests the GetQualityTrends method
func TestQualityService_GetQualityTrends(t *testing.T) {
	projectID := ulid.New()

	tests := []struct {
		name        string
		filter      *observability.AnalyticsFilter
		interval    string
		mockSetup   func()
		expectedErr error
		checkResult func(*testing.T, []*observability.QualityTrendPoint)
	}{
		{
			name: "success - get quality trends",
			filter: &observability.AnalyticsFilter{
				ProjectID: projectID,
				StartTime: time.Now().Add(-7 * 24 * time.Hour),
				EndTime:   time.Now(),
			},
			interval:    "day",
			mockSetup:   func() {},
			expectedErr: nil,
			checkResult: func(t *testing.T, trends []*observability.QualityTrendPoint) {
				assert.NotNil(t, trends)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockQSRepo := new(MockQualityScoreRepository)
			mockTraceRepo := new(MockTraceRepository)
			mockObsRepo := new(MockObservationRepository)
			mockPublisher := new(MockEventPublisher)

			tt.mockSetup()

			service := NewQualityService(mockQSRepo, mockTraceRepo, mockObsRepo, mockPublisher)

			result, err := service.GetQualityTrends(context.Background(), tt.filter, tt.interval)

			if tt.expectedErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}

// TestQualityService_GetScoreDistribution tests the GetScoreDistribution method
func TestQualityService_GetScoreDistribution(t *testing.T) {
	tests := []struct {
		name        string
		scoreName   string
		filter      *observability.QualityScoreFilter
		mockSetup   func()
		expectedErr error
		checkResult func(*testing.T, map[string]int)
	}{
		{
			name:        "success - get score distribution",
			scoreName:   "accuracy",
			filter:      &observability.QualityScoreFilter{},
			mockSetup:   func() {},
			expectedErr: nil,
			checkResult: func(t *testing.T, distribution map[string]int) {
				assert.NotNil(t, distribution)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockQSRepo := new(MockQualityScoreRepository)
			mockTraceRepo := new(MockTraceRepository)
			mockObsRepo := new(MockObservationRepository)
			mockPublisher := new(MockEventPublisher)

			tt.mockSetup()

			service := NewQualityService(mockQSRepo, mockTraceRepo, mockObsRepo, mockPublisher)

			result, err := service.GetScoreDistribution(context.Background(), tt.scoreName, tt.filter)

			if tt.expectedErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}
