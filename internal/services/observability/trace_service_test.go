package observability

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"brokle/internal/core/domain/observability"
	"brokle/pkg/ulid"
)

// ============================================================================
// Mock Repositories
// ============================================================================

type MockTraceRepository struct {
	mock.Mock
}

func (m *MockTraceRepository) Create(ctx context.Context, trace *observability.Trace) error {
	args := m.Called(ctx, trace)
	return args.Error(0)
}

func (m *MockTraceRepository) GetByID(ctx context.Context, id ulid.ULID) (*observability.Trace, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*observability.Trace), args.Error(1)
}

// Removed after refactor: GetByExternalTraceID method removed

func (m *MockTraceRepository) Update(ctx context.Context, trace *observability.Trace) error {
	args := m.Called(ctx, trace)
	return args.Error(0)
}

func (m *MockTraceRepository) Delete(ctx context.Context, id ulid.ULID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Removed after refactor: old GetByProjectID signature removed

// Removed after refactor: old GetByUserID signature removed

func (m *MockTraceRepository) GetBySessionID(ctx context.Context, sessionID ulid.ULID) ([]*observability.Trace, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.Trace), args.Error(1)
}

// Removed after refactor: SearchTraces method removed

// Removed after refactor: GetTraceWithObservations method removed

// Removed after refactor: GetTraceStats method and TraceStats type removed

func (m *MockTraceRepository) CreateBatch(ctx context.Context, traces []*observability.Trace) error {
	args := m.Called(ctx, traces)
	return args.Error(0)
}

func (m *MockTraceRepository) UpdateBatch(ctx context.Context, traces []*observability.Trace) error {
	args := m.Called(ctx, traces)
	return args.Error(0)
}

func (m *MockTraceRepository) DeleteBatch(ctx context.Context, ids []ulid.ULID) error {
	args := m.Called(ctx, ids)
	return args.Error(0)
}

// Removed after refactor: GetTracesByTimeRange method removed

// Removed after refactor: CountTraces method removed, now using Count

// Removed after refactor: GetRecentTraces method removed

func (m *MockTraceRepository) Count(ctx context.Context, filter *observability.TraceFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTraceRepository) GetByProjectID(ctx context.Context, projectID ulid.ULID, filter *observability.TraceFilter) ([]*observability.Trace, error) {
	args := m.Called(ctx, projectID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.Trace), args.Error(1)
}

func (m *MockTraceRepository) GetChildren(ctx context.Context, parentTraceID ulid.ULID) ([]*observability.Trace, error) {
	args := m.Called(ctx, parentTraceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.Trace), args.Error(1)
}

func (m *MockTraceRepository) GetWithObservations(ctx context.Context, id ulid.ULID) (*observability.Trace, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*observability.Trace), args.Error(1)
}

func (m *MockTraceRepository) GetWithScores(ctx context.Context, id ulid.ULID) (*observability.Trace, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*observability.Trace), args.Error(1)
}

func (m *MockTraceRepository) GetByUserID(ctx context.Context, userID ulid.ULID, filter *observability.TraceFilter) ([]*observability.Trace, error) {
	args := m.Called(ctx, userID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.Trace), args.Error(1)
}

type MockObservationRepository struct {
	mock.Mock
}

func (m *MockObservationRepository) Create(ctx context.Context, observation *observability.Observation) error {
	args := m.Called(ctx, observation)
	return args.Error(0)
}

func (m *MockObservationRepository) GetByID(ctx context.Context, id ulid.ULID) (*observability.Observation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*observability.Observation), args.Error(1)
}

// Removed after refactor: GetByExternalObservationID method removed

func (m *MockObservationRepository) Update(ctx context.Context, observation *observability.Observation) error {
	args := m.Called(ctx, observation)
	return args.Error(0)
}

func (m *MockObservationRepository) Delete(ctx context.Context, id ulid.ULID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockObservationRepository) GetByTraceID(ctx context.Context, traceID ulid.ULID) ([]*observability.Observation, error) {
	args := m.Called(ctx, traceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.Observation), args.Error(1)
}

// Removed after refactor: GetByParentObservationID renamed to GetChildren

// Removed after refactor: GetByType method removed

// Removed after refactor: GetByProvider method removed

// Removed after refactor: GetByModel method removed

// Removed after refactor: SearchObservations method removed

// Removed after refactor: GetObservationStats method and ObservationStats type removed

func (m *MockObservationRepository) CreateBatch(ctx context.Context, observations []*observability.Observation) error {
	args := m.Called(ctx, observations)
	return args.Error(0)
}

func (m *MockObservationRepository) UpdateBatch(ctx context.Context, observations []*observability.Observation) error {
	args := m.Called(ctx, observations)
	return args.Error(0)
}

func (m *MockObservationRepository) DeleteBatch(ctx context.Context, ids []ulid.ULID) error {
	args := m.Called(ctx, ids)
	return args.Error(0)
}

// Removed after refactor: CompleteObservation method removed

// Removed after refactor: GetIncompleteObservations method removed

// Removed after refactor: GetObservationsByTimeRange method removed

// Removed after refactor: CountObservations method removed, now using Count

func (m *MockObservationRepository) Count(ctx context.Context, filter *observability.ObservationFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockObservationRepository) GetByFilter(ctx context.Context, filter *observability.ObservationFilter) ([]*observability.Observation, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.Observation), args.Error(1)
}

func (m *MockObservationRepository) GetChildren(ctx context.Context, parentObservationID ulid.ULID) ([]*observability.Observation, error) {
	args := m.Called(ctx, parentObservationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.Observation), args.Error(1)
}

func (m *MockObservationRepository) GetTreeByTraceID(ctx context.Context, traceID ulid.ULID) ([]*observability.Observation, error) {
	args := m.Called(ctx, traceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.Observation), args.Error(1)
}

type MockEventPublisher struct {
	mock.Mock
}

func (m *MockEventPublisher) Publish(ctx context.Context, event *observability.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishBatch(ctx context.Context, events []*observability.Event) error {
	args := m.Called(ctx, events)
	return args.Error(0)
}

// MockScoreRepository is defined in observation_service_test.go

// ============================================================================
// CreateTrace Tests
// ============================================================================


// ============================================================================
// HIGH-VALUE TESTS: Complex Business Logic & Orchestration
// ============================================================================

// Removed after refactor: CreateTraceWithObservations method no longer exists
/*
func TestTraceService_CreateTraceWithObservations(t *testing.T) {
	tests := []struct {
		name        string
		trace       *observability.Trace
		mockSetup   func(*MockTraceRepository, *MockObservationRepository, *MockEventPublisher)
		expectedErr error
		checkResult func(*testing.T, *observability.Trace)
	}{
		{
			name: "success - trace with observations",
			trace: &observability.Trace{
				ProjectID: ulid.New(),
				Name:      "Trace with Observations",
				Observations: []*observability.Observation{
					{
						Type:      observability.ObservationTypeLLM,
						Name:      "LLM Call 1",
						StartTime: time.Now(),
					},
					{
						Type:      observability.ObservationTypeSpan,
						Name:      "Span 1",
						StartTime: time.Now(),
					},
				},
			},
			mockSetup: func(traceRepo *MockTraceRepository, obsRepo *MockObservationRepository, publisher *MockEventPublisher) {
				traceRepo.On("Create", mock.Anything, mock.AnythingOfType("*observability.Trace")).
					Return(nil)
				obsRepo.On("CreateBatch", mock.Anything, mock.AnythingOfType("[]*observability.Observation")).
					Return(nil)
				publisher.On("Publish", mock.Anything, mock.AnythingOfType("*observability.Event")).
					Return(nil)
			},
			expectedErr: nil,
			checkResult: func(t *testing.T, trace *observability.Trace) {
				assert.NotNil(t, trace)
				// The returned trace doesn't include observations
				// But the original trace object has been modified with IDs
			},
		},
		{
			name: "success - trace without observations",
			trace: &observability.Trace{
				ProjectID: ulid.New(),
				Name:      "Trace without Observations",
			},
			mockSetup: func(traceRepo *MockTraceRepository, obsRepo *MockObservationRepository, publisher *MockEventPublisher) {
				traceRepo.On("Create", mock.Anything, mock.AnythingOfType("*observability.Trace")).
					Return(nil)
				publisher.On("Publish", mock.Anything, mock.AnythingOfType("*observability.Event")).
					Return(nil)
				// No observation repo call expected
			},
			expectedErr: nil,
			checkResult: func(t *testing.T, trace *observability.Trace) {
				assert.NotNil(t, trace)
			},
		},
		{
			name:  "error - nil trace",
			trace: nil,
			mockSetup: func(traceRepo *MockTraceRepository, obsRepo *MockObservationRepository, publisher *MockEventPublisher) {
				// No calls expected
			},
			expectedErr: &observability.ObservabilityError{},
			checkResult: func(t *testing.T, trace *observability.Trace) {
				assert.Nil(t, trace)
			},
		},
		{
			name: "error - observation creation failure",
			trace: &observability.Trace{
				ProjectID: ulid.New(),
				Name:      "Trace Obs Fail",
				Observations: []*observability.Observation{
					{
						Type:      observability.ObservationTypeLLM,
						Name:      "Failing Observation",
						StartTime: time.Now(),
					},
				},
			},
			mockSetup: func(traceRepo *MockTraceRepository, obsRepo *MockObservationRepository, publisher *MockEventPublisher) {
				traceRepo.On("Create", mock.Anything, mock.AnythingOfType("*observability.Trace")).
					Return(nil)
				obsRepo.On("CreateBatch", mock.Anything, mock.AnythingOfType("[]*observability.Observation")).
					Return(errors.New("observation creation failed"))
				publisher.On("Publish", mock.Anything, mock.AnythingOfType("*observability.Event")).
					Return(nil)
			},
			expectedErr: errors.New("failed to create observations"),
			checkResult: func(t *testing.T, trace *observability.Trace) {
				assert.Nil(t, trace)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTraceRepo := new(MockTraceRepository)
			mockObsRepo := new(MockObservationRepository)
			mockPublisher := new(MockEventPublisher)

			tt.mockSetup(mockTraceRepo, mockObsRepo, mockPublisher)

			service := NewTraceService(mockTraceRepo, mockObsRepo, mockPublisher)

			result, err := service.CreateTraceWithObservations(context.Background(), tt.trace)

			if tt.expectedErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}

			mockTraceRepo.AssertExpectations(t)
			mockObsRepo.AssertExpectations(t)
			mockPublisher.AssertExpectations(t)
		})
	}
}

// ============================================================================
// GetTrace Tests
// ============================================================================

*/

func TestTraceService_GetTraceWithObservations(t *testing.T) {
	traceID := ulid.New()

	tests := []struct {
		name        string
		id          ulid.ULID
		mockSetup   func(*MockTraceRepository, *MockObservationRepository)
		expectedErr error
		checkResult func(*testing.T, *observability.Trace)
	}{
		{
			name: "success - trace with observations found",
			id:   traceID,
			mockSetup: func(traceRepo *MockTraceRepository, obsRepo *MockObservationRepository) {
				// GetTraceWithObservations calls GetTrace which calls GetByID
				traceRepo.On("GetByID", mock.Anything, traceID).
				Return(&observability.Trace{
					ID:        traceID,
					ProjectID: ulid.New(),
					Name:      "Test Trace",
				}, nil)

				// Then it calls GetByTraceID to get observations
				observations := []*observability.Observation{
					{
						ID:        ulid.New(),
						TraceID:   traceID,
						Type:      observability.ObservationTypeLLM,
						Name:      "LLM Call",
						StartTime: time.Now(),
					},
				}
			obsRepo.On("GetTreeByTraceID", mock.Anything, traceID).
				Return(observations, nil)
			},
			expectedErr: nil,
			checkResult: func(t *testing.T, trace *observability.Trace) {
				assert.NotNil(t, trace)
				assert.Len(t, trace.Observations, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockTraceRepository)
		mockObsRepo := new(MockObservationRepository)

		tt.mockSetup(mockRepo, mockObsRepo)

			mockScoreRepo := &MockScoreRepository{}
			service := NewTraceService(mockRepo, mockObsRepo, mockScoreRepo)

			result, err := service.GetTraceWithObservations(context.Background(), tt.id)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedErr))
			} else {
				assert.NoError(t, err)
			}

			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}

			mockRepo.AssertExpectations(t)
			mockObsRepo.AssertExpectations(t)
		})
	}
}

// ============================================================================
// CreateTracesBatch Tests
// ============================================================================

// Removed after refactor: CreateTracesBatch method no longer exists
/*
func TestTraceService_CreateTracesBatch(t *testing.T) {
	projectID := ulid.New()

	tests := []struct {
		name        string
		traces      []*observability.Trace
		mockSetup   func(*MockTraceRepository, *MockEventPublisher)
		expectedErr error
		checkResult func(*testing.T, []*observability.Trace)
	}{
		{
			name: "success - create batch of traces",
			traces: []*observability.Trace{
				{
					ProjectID: projectID,
					Name:      "Trace 1",
				},
				{
					ProjectID: projectID,
					Name:      "Trace 2",
				},
			},
			mockSetup: func(repo *MockTraceRepository, publisher *MockEventPublisher) {
				repo.On("CreateBatch", mock.Anything, mock.AnythingOfType("[]*observability.Trace")).
					Return(nil)
				publisher.On("PublishBatch", mock.Anything, mock.AnythingOfType("[]*observability.Event")).
					Return(nil)
			},
			expectedErr: nil,
			checkResult: func(t *testing.T, traces []*observability.Trace) {
				assert.NotNil(t, traces)
				assert.Len(t, traces, 2)
				// Check IDs were generated
				for _, trace := range traces {
					assert.NotEqual(t, ulid.ULID{}, trace.ID)
					assert.False(t, trace.CreatedAt.IsZero())
				}
			},
		},
		{
			name:   "success - empty batch",
			traces: []*observability.Trace{},
			mockSetup: func(repo *MockTraceRepository, publisher *MockEventPublisher) {
				// No calls expected for empty batch
			},
			expectedErr: nil,
			checkResult: func(t *testing.T, traces []*observability.Trace) {
				assert.NotNil(t, traces)
				assert.Len(t, traces, 0)
			},
		},
		{
			name: "error - validation failure in batch",
			traces: []*observability.Trace{
				{
					ProjectID: projectID,
					Name:      "",  // Missing name
				},
			},
			mockSetup: func(repo *MockTraceRepository, publisher *MockEventPublisher) {
				// No calls expected due to validation failure
			},
			expectedErr: errors.New("validation failed"),
			checkResult: func(t *testing.T, traces []*observability.Trace) {
				assert.Nil(t, traces)
			},
		},
		{
			name: "error - repository batch failure",
			traces: []*observability.Trace{
				{
					ProjectID: projectID,
					Name:      "Trace 1",
				},
			},
			mockSetup: func(repo *MockTraceRepository, publisher *MockEventPublisher) {
				repo.On("CreateBatch", mock.Anything, mock.AnythingOfType("[]*observability.Trace")).
					Return(errors.New("database error"))
			},
			expectedErr: errors.New("failed to create traces batch"),
			checkResult: func(t *testing.T, traces []*observability.Trace) {
				assert.Nil(t, traces)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockTraceRepository)
			mockObsRepo := new(MockObservationRepository)
			mockPublisher := new(MockEventPublisher)

			tt.mockSetup(mockRepo, mockPublisher)

			service := NewTraceService(mockRepo, mockObsRepo, mockPublisher)

			result, err := service.CreateTracesBatch(context.Background(), tt.traces)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}

			mockRepo.AssertExpectations(t)
			mockPublisher.AssertExpectations(t)
		})
	}
}

// Removed after refactor: IngestTraceBatch tests removed - BatchIngestRequest and BatchIngestResult types no longer exist

*/

// Removed after refactor: GetTracesByTimeRange, GetRecentTraces, and GetTraceAnalytics methods no longer exist

