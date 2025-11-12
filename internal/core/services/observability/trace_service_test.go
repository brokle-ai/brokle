package observability

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"brokle/internal/core/domain/observability"
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

func (m *MockTraceRepository) GetByID(ctx context.Context, id string) (*observability.Trace, error) {
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

func (m *MockTraceRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Removed after refactor: old GetByProjectID signature removed

// Removed after refactor: old GetByUserID signature removed

func (m *MockTraceRepository) GetBySessionID(ctx context.Context, sessionID string) ([]*observability.Trace, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.Trace), args.Error(1)
}

// Removed after refactor: SearchTraces method removed

// Removed after refactor: GetTraceWithSpans method removed

// Removed after refactor: GetTraceStats method and TraceStats type removed

func (m *MockTraceRepository) CreateBatch(ctx context.Context, traces []*observability.Trace) error {
	args := m.Called(ctx, traces)
	return args.Error(0)
}

func (m *MockTraceRepository) UpdateBatch(ctx context.Context, traces []*observability.Trace) error {
	args := m.Called(ctx, traces)
	return args.Error(0)
}

func (m *MockTraceRepository) DeleteBatch(ctx context.Context, ids []string) error {
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

func (m *MockTraceRepository) GetByProjectID(ctx context.Context, projectID string, filter *observability.TraceFilter) ([]*observability.Trace, error) {
	args := m.Called(ctx, projectID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.Trace), args.Error(1)
}

func (m *MockTraceRepository) GetChildren(ctx context.Context, parentTraceID string) ([]*observability.Trace, error) {
	args := m.Called(ctx, parentTraceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.Trace), args.Error(1)
}

func (m *MockTraceRepository) GetWithSpans(ctx context.Context, id string) (*observability.Trace, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*observability.Trace), args.Error(1)
}

func (m *MockTraceRepository) GetWithScores(ctx context.Context, id string) (*observability.Trace, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*observability.Trace), args.Error(1)
}

func (m *MockTraceRepository) GetByUserID(ctx context.Context, userID string, filter *observability.TraceFilter) ([]*observability.Trace, error) {
	args := m.Called(ctx, userID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.Trace), args.Error(1)
}

type MockSpanRepository struct {
	mock.Mock
}

func (m *MockSpanRepository) Create(ctx context.Context, span *observability.Span) error {
	args := m.Called(ctx, span)
	return args.Error(0)
}

func (m *MockSpanRepository) GetByID(ctx context.Context, id string) (*observability.Span, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*observability.Span), args.Error(1)
}

// Removed after refactor: GetByExternalSpanID method removed

func (m *MockSpanRepository) Update(ctx context.Context, span *observability.Span) error {
	args := m.Called(ctx, span)
	return args.Error(0)
}

func (m *MockSpanRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSpanRepository) GetByTraceID(ctx context.Context, traceID string) ([]*observability.Span, error) {
	args := m.Called(ctx, traceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.Span), args.Error(1)
}

// Removed after refactor: GetByParentSpanID renamed to GetChildren

// Removed after refactor: GetByType method removed

// Removed after refactor: GetByProvider method removed

// Removed after refactor: GetByModel method removed

// Removed after refactor: SearchSpans method removed

// Removed after refactor: GetSpanStats method and SpanStats type removed

func (m *MockSpanRepository) CreateBatch(ctx context.Context, spans []*observability.Span) error {
	args := m.Called(ctx, spans)
	return args.Error(0)
}

func (m *MockSpanRepository) UpdateBatch(ctx context.Context, spans []*observability.Span) error {
	args := m.Called(ctx, spans)
	return args.Error(0)
}

func (m *MockSpanRepository) DeleteBatch(ctx context.Context, ids []string) error {
	args := m.Called(ctx, ids)
	return args.Error(0)
}

// Removed after refactor: CompleteSpan method removed

// Removed after refactor: GetIncompleteSpans method removed

// Removed after refactor: GetSpansByTimeRange method removed

// Removed after refactor: CountSpans method removed, now using Count

func (m *MockSpanRepository) Count(ctx context.Context, filter *observability.SpanFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockSpanRepository) GetByFilter(ctx context.Context, filter *observability.SpanFilter) ([]*observability.Span, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.Span), args.Error(1)
}

func (m *MockSpanRepository) GetChildren(ctx context.Context, parentSpanID string) ([]*observability.Span, error) {
	args := m.Called(ctx, parentSpanID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.Span), args.Error(1)
}

func (m *MockSpanRepository) GetRootSpan(ctx context.Context, traceID string) (*observability.Span, error) {
	args := m.Called(ctx, traceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*observability.Span), args.Error(1)
}

func (m *MockSpanRepository) GetTreeByTraceID(ctx context.Context, traceID string) ([]*observability.Span, error) {
	args := m.Called(ctx, traceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.Span), args.Error(1)
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

// MockScoreRepository is defined in span_service_test.go

// ============================================================================
// CreateTrace Tests
// ============================================================================

// ============================================================================
// HIGH-VALUE TESTS: Complex Business Logic & Orchestration
// ============================================================================

// Removed after refactor: CreateTraceWithSpans method no longer exists
/*
func TestTraceService_CreateTraceWithSpans(t *testing.T) {
	tests := []struct {
		name        string
		trace       *observability.Trace
		mockSetup   func(*MockTraceRepository, *MockSpanRepository, *MockEventPublisher)
		expectedErr error
		checkResult func(*testing.T, *observability.Trace)
	}{
		{
			name: "success - trace with spans",
			trace: &observability.Trace{
				ProjectID: "11111111111111111111111111111111",
				Name:      "Trace with Spans",
				Spans: []*observability.Span{
					{
						Type:      observability.SpanTypeGeneration,
						Name:      "LLM Call 1",
						StartTime: time.Now(),
					},
					{
						Type:      observability.SpanTypeSpan,
						Name:      "Span 1",
						StartTime: time.Now(),
					},
				},
			},
			mockSetup: func(traceRepo *MockTraceRepository, obsRepo *MockSpanRepository, publisher *MockEventPublisher) {
				traceRepo.On("Create", mock.Anything, mock.AnythingOfType("*observability.Trace")).
					Return(nil)
				obsRepo.On("CreateBatch", mock.Anything, mock.AnythingOfType("[]*observability.Span")).
					Return(nil)
				publisher.On("Publish", mock.Anything, mock.AnythingOfType("*observability.Event")).
					Return(nil)
			},
			expectedErr: nil,
			checkResult: func(t *testing.T, trace *observability.Trace) {
				assert.NotNil(t, trace)
				// The returned trace doesn't include spans
				// But the original trace object has been modified with IDs
			},
		},
		{
			name: "success - trace without spans",
			trace: &observability.Trace{
				ProjectID: "22222222222222222222222222222222",
				Name:      "Trace without Spans",
			},
			mockSetup: func(traceRepo *MockTraceRepository, obsRepo *MockSpanRepository, publisher *MockEventPublisher) {
				traceRepo.On("Create", mock.Anything, mock.AnythingOfType("*observability.Trace")).
					Return(nil)
				publisher.On("Publish", mock.Anything, mock.AnythingOfType("*observability.Event")).
					Return(nil)
				// No span repo call expected
			},
			expectedErr: nil,
			checkResult: func(t *testing.T, trace *observability.Trace) {
				assert.NotNil(t, trace)
			},
		},
		{
			name:  "error - nil trace",
			trace: nil,
			mockSetup: func(traceRepo *MockTraceRepository, obsRepo *MockSpanRepository, publisher *MockEventPublisher) {
				// No calls expected
			},
			expectedErr: &observability.ObservabilityError{},
			checkResult: func(t *testing.T, trace *observability.Trace) {
				assert.Nil(t, trace)
			},
		},
		{
			name: "error - span creation failure",
			trace: &observability.Trace{
				ProjectID: "33333333333333333333333333333333",
				Name:      "Trace Obs Fail",
				Spans: []*observability.Span{
					{
						Type:      observability.SpanTypeGeneration,
						Name:      "Failing Span",
						StartTime: time.Now(),
					},
				},
			},
			mockSetup: func(traceRepo *MockTraceRepository, obsRepo *MockSpanRepository, publisher *MockEventPublisher) {
				traceRepo.On("Create", mock.Anything, mock.AnythingOfType("*observability.Trace")).
					Return(nil)
				obsRepo.On("CreateBatch", mock.Anything, mock.AnythingOfType("[]*observability.Span")).
					Return(errors.New("span creation failed"))
				publisher.On("Publish", mock.Anything, mock.AnythingOfType("*observability.Event")).
					Return(nil)
			},
			expectedErr: errors.New("failed to create spans"),
			checkResult: func(t *testing.T, trace *observability.Trace) {
				assert.Nil(t, trace)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTraceRepo := new(MockTraceRepository)
			mockObsRepo := new(MockSpanRepository)
			mockPublisher := new(MockEventPublisher)

			tt.mockSetup(mockTraceRepo, mockObsRepo, mockPublisher)

			service := NewTraceService(mockTraceRepo, mockObsRepo, mockPublisher)

			result, err := service.CreateTraceWithSpans(context.Background(), tt.trace)

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

func TestTraceService_GetTraceWithSpans(t *testing.T) {
	traceID := "12345678901234567890123456789012"

	tests := []struct {
		name        string
		id          string
		mockSetup   func(*MockTraceRepository, *MockSpanRepository)
		expectedErr error
		checkResult func(*testing.T, *observability.Trace)
	}{
		{
			name: "success - trace with spans found",
			id:   traceID,
			mockSetup: func(traceRepo *MockTraceRepository, obsRepo *MockSpanRepository) {
				// GetTraceWithSpans calls GetTrace which calls GetByID
				traceRepo.On("GetByID", mock.Anything, traceID).
					Return(&observability.Trace{
						ID:        traceID,
						ProjectID: "98765432109876543210987654321098",
						Name:      "Test Trace",
					}, nil)

				// Then it calls GetByTraceID to get spans
				spans := []*observability.Span{
					{
						ID:        "abcdef12345678901234567890123456",
						TraceID:   traceID,
						Type:      observability.SpanTypeGeneration,
						Name:      "LLM Call",
						StartTime: time.Now(),
					},
				}
				obsRepo.On("GetByTraceID", mock.Anything, traceID).
					Return(spans, nil)
			},
			expectedErr: nil,
			checkResult: func(t *testing.T, trace *observability.Trace) {
				assert.NotNil(t, trace)
				assert.Len(t, trace.Spans, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockTraceRepository)
			mockObsRepo := new(MockSpanRepository)

			tt.mockSetup(mockRepo, mockObsRepo)

			mockScoreRepo := &MockScoreRepository{}
			service := NewTraceService(mockRepo, mockObsRepo, mockScoreRepo, logrus.New())

			result, err := service.GetTraceWithSpans(context.Background(), tt.id)

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
	projectID := "44444444444444444444444444444444"

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
			mockObsRepo := new(MockSpanRepository)
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
