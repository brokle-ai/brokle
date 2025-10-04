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

func (m *MockTraceRepository) GetByExternalTraceID(ctx context.Context, externalTraceID string) (*observability.Trace, error) {
	args := m.Called(ctx, externalTraceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*observability.Trace), args.Error(1)
}

func (m *MockTraceRepository) Update(ctx context.Context, trace *observability.Trace) error {
	args := m.Called(ctx, trace)
	return args.Error(0)
}

func (m *MockTraceRepository) Delete(ctx context.Context, id ulid.ULID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTraceRepository) GetByProjectID(ctx context.Context, projectID ulid.ULID, limit, offset int) ([]*observability.Trace, error) {
	args := m.Called(ctx, projectID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.Trace), args.Error(1)
}

func (m *MockTraceRepository) GetByUserID(ctx context.Context, userID ulid.ULID, limit, offset int) ([]*observability.Trace, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.Trace), args.Error(1)
}

func (m *MockTraceRepository) GetBySessionID(ctx context.Context, sessionID ulid.ULID) ([]*observability.Trace, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.Trace), args.Error(1)
}

func (m *MockTraceRepository) SearchTraces(ctx context.Context, filter *observability.TraceFilter) ([]*observability.Trace, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*observability.Trace), args.Int(1), args.Error(2)
}

func (m *MockTraceRepository) GetTraceWithObservations(ctx context.Context, id ulid.ULID) (*observability.Trace, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*observability.Trace), args.Error(1)
}

func (m *MockTraceRepository) GetTraceStats(ctx context.Context, id ulid.ULID) (*observability.TraceStats, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*observability.TraceStats), args.Error(1)
}

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

func (m *MockTraceRepository) GetTracesByTimeRange(ctx context.Context, projectID ulid.ULID, startTime, endTime time.Time, limit, offset int) ([]*observability.Trace, error) {
	args := m.Called(ctx, projectID, startTime, endTime, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.Trace), args.Error(1)
}

func (m *MockTraceRepository) CountTraces(ctx context.Context, filter *observability.TraceFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTraceRepository) GetRecentTraces(ctx context.Context, projectID ulid.ULID, limit int) ([]*observability.Trace, error) {
	args := m.Called(ctx, projectID, limit)
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

func (m *MockObservationRepository) GetByExternalObservationID(ctx context.Context, externalObservationID string) (*observability.Observation, error) {
	args := m.Called(ctx, externalObservationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*observability.Observation), args.Error(1)
}

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

func (m *MockObservationRepository) GetByParentObservationID(ctx context.Context, parentID ulid.ULID) ([]*observability.Observation, error) {
	args := m.Called(ctx, parentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.Observation), args.Error(1)
}

func (m *MockObservationRepository) GetByType(ctx context.Context, obsType observability.ObservationType, limit, offset int) ([]*observability.Observation, error) {
	args := m.Called(ctx, obsType, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.Observation), args.Error(1)
}

func (m *MockObservationRepository) GetByProvider(ctx context.Context, provider string, limit, offset int) ([]*observability.Observation, error) {
	args := m.Called(ctx, provider, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.Observation), args.Error(1)
}

func (m *MockObservationRepository) GetByModel(ctx context.Context, provider, model string, limit, offset int) ([]*observability.Observation, error) {
	args := m.Called(ctx, provider, model, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.Observation), args.Error(1)
}

func (m *MockObservationRepository) SearchObservations(ctx context.Context, filter *observability.ObservationFilter) ([]*observability.Observation, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*observability.Observation), args.Int(1), args.Error(2)
}

func (m *MockObservationRepository) GetObservationStats(ctx context.Context, filter *observability.ObservationFilter) (*observability.ObservationStats, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*observability.ObservationStats), args.Error(1)
}

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

func (m *MockObservationRepository) CompleteObservation(ctx context.Context, id ulid.ULID, endTime time.Time, output any, cost *float64) error {
	args := m.Called(ctx, id, endTime, output, cost)
	return args.Error(0)
}

func (m *MockObservationRepository) GetIncompleteObservations(ctx context.Context, projectID ulid.ULID) ([]*observability.Observation, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.Observation), args.Error(1)
}

func (m *MockObservationRepository) GetObservationsByTimeRange(ctx context.Context, filter *observability.ObservationFilter, startTime, endTime time.Time) ([]*observability.Observation, error) {
	args := m.Called(ctx, filter, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.Observation), args.Error(1)
}

func (m *MockObservationRepository) CountObservations(ctx context.Context, filter *observability.ObservationFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
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

// ============================================================================
// CreateTrace Tests
// ============================================================================


// ============================================================================
// HIGH-VALUE TESTS: Complex Business Logic & Orchestration
// ============================================================================

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
				ProjectID:       ulid.New(),
				ExternalTraceID: "ext-trace-with-obs",
				Name:            "Trace with Observations",
				Observations: []observability.Observation{
					{
						ExternalObservationID: "obs-1",
						Type:                  observability.ObservationTypeLLM,
						Name:                  "LLM Call 1",
						StartTime:             time.Now(),
					},
					{
						ExternalObservationID: "obs-2",
						Type:                  observability.ObservationTypeSpan,
						Name:                  "Span 1",
						StartTime:             time.Now(),
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
				ProjectID:       ulid.New(),
				ExternalTraceID: "ext-trace-no-obs",
				Name:            "Trace without Observations",
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
				ProjectID:       ulid.New(),
				ExternalTraceID: "ext-trace-obs-fail",
				Name:            "Trace Obs Fail",
				Observations: []observability.Observation{
					{
						ExternalObservationID: "obs-fail",
						Type:                  observability.ObservationTypeLLM,
						Name:                  "Failing Observation",
						StartTime:             time.Now(),
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
						ID:              traceID,
						ProjectID:       ulid.New(),
						ExternalTraceID: "ext-123",
						Name:            "Test Trace",
					}, nil)

				// Then it calls GetByTraceID to get observations
				observations := []*observability.Observation{
					{
						ID:                    ulid.New(),
						TraceID:               traceID,
						ExternalObservationID: "obs-1",
						Type:                  observability.ObservationTypeLLM,
						Name:                  "LLM Call",
						StartTime:             time.Now(),
					},
				}
				obsRepo.On("GetByTraceID", mock.Anything, traceID).
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
			mockPublisher := new(MockEventPublisher)

			tt.mockSetup(mockRepo, mockObsRepo)

			service := NewTraceService(mockRepo, mockObsRepo, mockPublisher)

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
					ProjectID:       projectID,
					ExternalTraceID: "ext-1",
					Name:            "Trace 1",
				},
				{
					ProjectID:       projectID,
					ExternalTraceID: "ext-2",
					Name:            "Trace 2",
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
					ProjectID:       projectID,
					ExternalTraceID: "ext-1",
					Name:            "",  // Missing name
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
					ProjectID:       projectID,
					ExternalTraceID: "ext-1",
					Name:            "Trace 1",
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

// ============================================================================
// IngestTraceBatch Tests
// ============================================================================

func TestTraceService_IngestTraceBatch(t *testing.T) {
	projectID := ulid.New()

	tests := []struct {
		name        string
		request     *observability.BatchIngestRequest
		mockSetup   func(*MockTraceRepository, *MockEventPublisher)
		expectedErr error
		checkResult func(*testing.T, *observability.BatchIngestResult)
	}{
		{
			name: "success - ingest batch",
			request: &observability.BatchIngestRequest{
				Traces: []*observability.Trace{
					{
						ProjectID:       projectID,
						ExternalTraceID: "ext-1",
						Name:            "Trace 1",
					},
					{
						ProjectID:       projectID,
						ExternalTraceID: "ext-2",
						Name:            "Trace 2",
					},
				},
			},
			mockSetup: func(repo *MockTraceRepository, publisher *MockEventPublisher) {
				repo.On("CreateBatch", mock.Anything, mock.AnythingOfType("[]*observability.Trace")).
					Return(nil)
				publisher.On("PublishBatch", mock.Anything, mock.AnythingOfType("[]*observability.Event")).
					Return(nil)
			},
			expectedErr: nil,
			checkResult: func(t *testing.T, result *observability.BatchIngestResult) {
				assert.NotNil(t, result)
				assert.Equal(t, 2, result.ProcessedCount)
				assert.Equal(t, 0, result.FailedCount)
				assert.Len(t, result.Errors, 0)
			},
		},
		{
			name: "error - batch processing failure",
			request: &observability.BatchIngestRequest{
				Traces: []*observability.Trace{
					{
						ProjectID:       projectID,
						ExternalTraceID: "ext-1",
						Name:            "",  // Invalid
					},
				},
			},
			mockSetup: func(repo *MockTraceRepository, publisher *MockEventPublisher) {
				// No calls expected
			},
			expectedErr: observability.NewObservabilityError(
				observability.ErrCodeValidationFailed,
				"validation failed",
			),
			checkResult: func(t *testing.T, result *observability.BatchIngestResult) {
				assert.NotNil(t, result)
				assert.Equal(t, 0, result.ProcessedCount)
				assert.Equal(t, 1, result.FailedCount)
				assert.Len(t, result.Errors, 1)
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

			result, err := service.IngestTraceBatch(context.Background(), tt.request)

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

// ============================================================================
// SearchTraces Tests
// ============================================================================

func TestTraceService_GetTracesByTimeRange(t *testing.T) {
	projectID := ulid.New()
	now := time.Now()
	startTime := now.Add(-24 * time.Hour)
	endTime := now

	tests := []struct {
		name        string
		projectID   ulid.ULID
		startTime   time.Time
		endTime     time.Time
		limit       int
		offset      int
		mockSetup   func(*MockTraceRepository)
		expectedErr error
		checkResult func(*testing.T, []*observability.Trace)
	}{
		{
			name:      "success - get traces by time range",
			projectID: projectID,
			startTime: startTime,
			endTime:   endTime,
			limit:     10,
			offset:    0,
			mockSetup: func(repo *MockTraceRepository) {
				traces := []*observability.Trace{
					{ID: ulid.New(), Name: "Recent 1"},
					{ID: ulid.New(), Name: "Recent 2"},
				}
				repo.On("GetTracesByTimeRange", mock.Anything, projectID, startTime, endTime, 10, 0).
					Return(traces, nil)
			},
			expectedErr: nil,
			checkResult: func(t *testing.T, traces []*observability.Trace) {
				assert.NotNil(t, traces)
				assert.Len(t, traces, 2)
			},
		},
		{
			name:      "error - empty project ID",
			projectID: ulid.ULID{},
			startTime: startTime,
			endTime:   endTime,
			limit:     10,
			offset:    0,
			mockSetup: func(repo *MockTraceRepository) {
				// No calls expected
			},
			expectedErr: errors.New("project ID cannot be empty"),
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

			tt.mockSetup(mockRepo)

			service := NewTraceService(mockRepo, mockObsRepo, mockPublisher)

			result, err := service.GetTracesByTimeRange(
				context.Background(),
				tt.projectID,
				tt.startTime,
				tt.endTime,
				tt.limit,
				tt.offset,
			)

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
		})
	}
}

// ============================================================================
// GetRecentTraces Tests (Already exists, adding more comprehensive version)
// ============================================================================

// GetRecentTraces already has basic tests above

// ============================================================================
// GetTraceAnalytics Tests
// ============================================================================

