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
// Mock TelemetryEventRepository (reuse from telemetry_batch_service_test.go)
// ============================================================================

// Note: MockTelemetryEventRepository is already defined in telemetry_batch_service_test.go
// For this test file, we'll define a local version to avoid import cycles

type MockEventRepository struct {
	mock.Mock
}

func (m *MockEventRepository) Create(ctx context.Context, event *observability.TelemetryEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventRepository) GetByID(ctx context.Context, id ulid.ULID) (*observability.TelemetryEvent, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*observability.TelemetryEvent), args.Error(1)
}

func (m *MockEventRepository) Update(ctx context.Context, event *observability.TelemetryEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventRepository) Delete(ctx context.Context, id ulid.ULID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockEventRepository) GetByBatchID(ctx context.Context, batchID ulid.ULID) ([]*observability.TelemetryEvent, error) {
	args := m.Called(ctx, batchID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.TelemetryEvent), args.Error(1)
}

func (m *MockEventRepository) GetUnprocessedByBatchID(ctx context.Context, batchID ulid.ULID) ([]*observability.TelemetryEvent, error) {
	args := m.Called(ctx, batchID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.TelemetryEvent), args.Error(1)
}

func (m *MockEventRepository) GetFailedByBatchID(ctx context.Context, batchID ulid.ULID) ([]*observability.TelemetryEvent, error) {
	args := m.Called(ctx, batchID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.TelemetryEvent), args.Error(1)
}

func (m *MockEventRepository) GetByEventType(ctx context.Context, eventType observability.TelemetryEventType, limit, offset int) ([]*observability.TelemetryEvent, error) {
	args := m.Called(ctx, eventType, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.TelemetryEvent), args.Error(1)
}

func (m *MockEventRepository) GetUnprocessedByType(ctx context.Context, eventType observability.TelemetryEventType, limit int) ([]*observability.TelemetryEvent, error) {
	args := m.Called(ctx, eventType, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.TelemetryEvent), args.Error(1)
}

func (m *MockEventRepository) MarkAsProcessed(ctx context.Context, id ulid.ULID, processedAt time.Time) error {
	args := m.Called(ctx, id, processedAt)
	return args.Error(0)
}

func (m *MockEventRepository) MarkAsFailed(ctx context.Context, id ulid.ULID, errorMessage string) error {
	args := m.Called(ctx, id, errorMessage)
	return args.Error(0)
}

func (m *MockEventRepository) IncrementRetryCount(ctx context.Context, id ulid.ULID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockEventRepository) CreateBatch(ctx context.Context, events []*observability.TelemetryEvent) error {
	args := m.Called(ctx, events)
	return args.Error(0)
}

func (m *MockEventRepository) UpdateBatch(ctx context.Context, events []*observability.TelemetryEvent) error {
	args := m.Called(ctx, events)
	return args.Error(0)
}

func (m *MockEventRepository) ProcessBatch(ctx context.Context, batchID ulid.ULID, processor func([]*observability.TelemetryEvent) error) error {
	args := m.Called(ctx, batchID, processor)
	return args.Error(0)
}

func (m *MockEventRepository) GetEventsForRetry(ctx context.Context, maxRetries int, limit int) ([]*observability.TelemetryEvent, error) {
	args := m.Called(ctx, maxRetries, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.TelemetryEvent), args.Error(1)
}

func (m *MockEventRepository) GetFailedEvents(ctx context.Context, batchID *ulid.ULID, limit, offset int) ([]*observability.TelemetryEvent, error) {
	args := m.Called(ctx, batchID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.TelemetryEvent), args.Error(1)
}

func (m *MockEventRepository) DeleteFailedEvents(ctx context.Context, olderThan time.Time) (int64, error) {
	args := m.Called(ctx, olderThan)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockEventRepository) GetEventStats(ctx context.Context, filter *observability.TelemetryEventFilter) (*observability.TelemetryEventStats, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*observability.TelemetryEventStats), args.Error(1)
}

func (m *MockEventRepository) GetEventTypeDistribution(ctx context.Context, batchID *ulid.ULID) (map[observability.TelemetryEventType]int, error) {
	args := m.Called(ctx, batchID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[observability.TelemetryEventType]int), args.Error(1)
}

func (m *MockEventRepository) CountEvents(ctx context.Context, filter *observability.TelemetryEventFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

// ============================================================================
// Mock TelemetryBatchRepository
// ============================================================================

type MockBatchRepository struct {
	mock.Mock
}

func (m *MockBatchRepository) Create(ctx context.Context, batch *observability.TelemetryBatch) error {
	args := m.Called(ctx, batch)
	return args.Error(0)
}

func (m *MockBatchRepository) GetByID(ctx context.Context, id ulid.ULID) (*observability.TelemetryBatch, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*observability.TelemetryBatch), args.Error(1)
}

func (m *MockBatchRepository) Update(ctx context.Context, batch *observability.TelemetryBatch) error {
	args := m.Called(ctx, batch)
	return args.Error(0)
}

func (m *MockBatchRepository) Delete(ctx context.Context, id ulid.ULID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockBatchRepository) GetByProjectID(ctx context.Context, projectID ulid.ULID, limit, offset int) ([]*observability.TelemetryBatch, error) {
	args := m.Called(ctx, projectID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.TelemetryBatch), args.Error(1)
}

func (m *MockBatchRepository) GetActiveByProjectID(ctx context.Context, projectID ulid.ULID) ([]*observability.TelemetryBatch, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.TelemetryBatch), args.Error(1)
}

func (m *MockBatchRepository) GetByStatus(ctx context.Context, status observability.BatchStatus, limit, offset int) ([]*observability.TelemetryBatch, error) {
	args := m.Called(ctx, status, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.TelemetryBatch), args.Error(1)
}

func (m *MockBatchRepository) GetProcessingBatches(ctx context.Context) ([]*observability.TelemetryBatch, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.TelemetryBatch), args.Error(1)
}

func (m *MockBatchRepository) GetCompletedBatches(ctx context.Context, projectID ulid.ULID, limit, offset int) ([]*observability.TelemetryBatch, error) {
	args := m.Called(ctx, projectID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.TelemetryBatch), args.Error(1)
}

func (m *MockBatchRepository) CreateBatch(ctx context.Context, batches []*observability.TelemetryBatch) error {
	args := m.Called(ctx, batches)
	return args.Error(0)
}

func (m *MockBatchRepository) UpdateBatch(ctx context.Context, batches []*observability.TelemetryBatch) error {
	args := m.Called(ctx, batches)
	return args.Error(0)
}

func (m *MockBatchRepository) SearchBatches(ctx context.Context, filter *observability.TelemetryBatchFilter) ([]*observability.TelemetryBatch, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*observability.TelemetryBatch), args.Int(1), args.Error(2)
}

func (m *MockBatchRepository) GetBatchWithEvents(ctx context.Context, id ulid.ULID) (*observability.TelemetryBatch, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*observability.TelemetryBatch), args.Error(1)
}

func (m *MockBatchRepository) UpdateBatchStatus(ctx context.Context, batchID ulid.ULID, status observability.BatchStatus, processingTimeMs *int) error {
	args := m.Called(ctx, batchID, status, processingTimeMs)
	return args.Error(0)
}

func (m *MockBatchRepository) GetBatchStats(ctx context.Context, id ulid.ULID) (*observability.BatchStats, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*observability.BatchStats), args.Error(1)
}

func (m *MockBatchRepository) GetBatchProcessingMetrics(ctx context.Context, filter *observability.TelemetryBatchFilter) (*observability.BatchProcessingMetrics, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*observability.BatchProcessingMetrics), args.Error(1)
}

func (m *MockBatchRepository) GetBatchThroughputStats(ctx context.Context, projectID ulid.ULID, timeWindow time.Duration) (*observability.BatchThroughputStats, error) {
	args := m.Called(ctx, projectID, timeWindow)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*observability.BatchThroughputStats), args.Error(1)
}

func (m *MockBatchRepository) GetBatchesByTimeRange(ctx context.Context, projectID ulid.ULID, startTime, endTime time.Time, limit, offset int) ([]*observability.TelemetryBatch, error) {
	args := m.Called(ctx, projectID, startTime, endTime, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.TelemetryBatch), args.Error(1)
}

func (m *MockBatchRepository) CountBatches(ctx context.Context, filter *observability.TelemetryBatchFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockBatchRepository) GetRecentBatches(ctx context.Context, projectID ulid.ULID, limit int) ([]*observability.TelemetryBatch, error) {
	args := m.Called(ctx, projectID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.TelemetryBatch), args.Error(1)
}

// ============================================================================
// Core CRUD Tests
// ============================================================================


// ============================================================================
// HIGH-VALUE TESTS: Event Processing, Retry Logic, Batch Operations
// ============================================================================

func TestTelemetryEventService_CreateEventsBatch(t *testing.T) {
	batchID := ulid.New()

	tests := []struct {
		name        string
		events      []*observability.TelemetryEvent
		mockSetup   func(*MockEventRepository)
		expectedErr string
	}{
		{
			name: "success - create batch of events",
			events: []*observability.TelemetryEvent{
				{
					BatchID:      batchID,
					EventType:    observability.TelemetryEventTypeTraceCreate,
					EventPayload: map[string]interface{}{"trace_id": "test1"},
				},
				{
					BatchID:      batchID,
					EventType:    observability.TelemetryEventTypeObservationCreate,
					EventPayload: map[string]interface{}{"observation_id": "test2"},
				},
			},
			mockSetup: func(repo *MockEventRepository) {
				repo.On("CreateBatch", mock.Anything, mock.AnythingOfType("[]*observability.TelemetryEvent")).
					Return(nil)
			},
			expectedErr: "",
		},
		{
			name:   "success - empty batch",
			events: []*observability.TelemetryEvent{},
			mockSetup: func(repo *MockEventRepository) {
				// No calls expected
			},
			expectedErr: "",
		},
		{
			name: "error - nil event in batch",
			events: []*observability.TelemetryEvent{
				{
					BatchID:      batchID,
					EventType:    observability.TelemetryEventTypeTraceCreate,
					EventPayload: map[string]interface{}{"data": "value"},
				},
				nil,
			},
			mockSetup: func(repo *MockEventRepository) {
				// No calls expected
			},
			expectedErr: "event at index 1 cannot be nil",
		},
		{
			name: "error - validation failure in batch",
			events: []*observability.TelemetryEvent{
				{
					BatchID:      batchID,
					EventType:    observability.TelemetryEventTypeTraceCreate,
					EventPayload: map[string]interface{}{"data": "value"},
				},
				{
					BatchID:      batchID,
					EventType:    "", // Invalid
					EventPayload: map[string]interface{}{"data": "value"},
				},
			},
			mockSetup: func(repo *MockEventRepository) {
				// No calls expected
			},
			expectedErr: "event validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEventRepo := new(MockEventRepository)
			mockBatchRepo := new(MockBatchRepository)

			tt.mockSetup(mockEventRepo)

			service := NewTelemetryEventService(mockEventRepo, mockBatchRepo)

			err := service.CreateEventsBatch(context.Background(), tt.events)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			mockEventRepo.AssertExpectations(t)
		})
	}
}

func TestTelemetryEventService_UpdateEventsBatch(t *testing.T) {
	batchID := ulid.New()
	eventID1 := ulid.New()
	eventID2 := ulid.New()

	tests := []struct {
		name        string
		events      []*observability.TelemetryEvent
		mockSetup   func(*MockEventRepository)
		expectedErr string
	}{
		{
			name: "success - update batch of events",
			events: []*observability.TelemetryEvent{
				{
					ID:           eventID1,
					BatchID:      batchID,
					EventType:    observability.TelemetryEventTypeTraceUpdate,
					EventPayload: map[string]interface{}{"updated": true},
				},
				{
					ID:           eventID2,
					BatchID:      batchID,
					EventType:    observability.TelemetryEventTypeObservationUpdate,
					EventPayload: map[string]interface{}{"updated": true},
				},
			},
			mockSetup: func(repo *MockEventRepository) {
				repo.On("UpdateBatch", mock.Anything, mock.AnythingOfType("[]*observability.TelemetryEvent")).
					Return(nil)
			},
			expectedErr: "",
		},
		{
			name:   "success - empty batch",
			events: []*observability.TelemetryEvent{},
			mockSetup: func(repo *MockEventRepository) {
				// No calls expected
			},
			expectedErr: "",
		},
		{
			name: "error - nil event in batch",
			events: []*observability.TelemetryEvent{
				{
					ID:           eventID1,
					BatchID:      batchID,
					EventType:    observability.TelemetryEventTypeTraceUpdate,
					EventPayload: map[string]interface{}{"data": "value"},
				},
				nil,
			},
			mockSetup: func(repo *MockEventRepository) {
				// No calls expected
			},
			expectedErr: "event at index 1 cannot be nil",
		},
		{
			name: "error - event without ID in batch",
			events: []*observability.TelemetryEvent{
				{
					ID:           ulid.ULID{},
					BatchID:      batchID,
					EventType:    observability.TelemetryEventTypeTraceUpdate,
					EventPayload: map[string]interface{}{"data": "value"},
				},
			},
			mockSetup: func(repo *MockEventRepository) {
				// No calls expected
			},
			expectedErr: "event at index 0 must have a valid ID for update",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEventRepo := new(MockEventRepository)
			mockBatchRepo := new(MockBatchRepository)

			tt.mockSetup(mockEventRepo)

			service := NewTelemetryEventService(mockEventRepo, mockBatchRepo)

			err := service.UpdateEventsBatch(context.Background(), tt.events)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			mockEventRepo.AssertExpectations(t)
		})
	}
}

func TestTelemetryEventService_ProcessEventsBatch(t *testing.T) {
	batchID := ulid.New()

	tests := []struct {
		name        string
		events      []*observability.TelemetryEvent
		mockSetup   func(*MockEventRepository)
		expectedErr string
		checkResult func(*testing.T, *observability.EventProcessingResult)
	}{
		{
			name: "success - process all events",
			events: []*observability.TelemetryEvent{
				{
					ID:           ulid.New(),
					BatchID:      batchID,
					EventType:    observability.TelemetryEventTypeTraceCreate,
					EventPayload: map[string]interface{}{"trace_id": "test1"},
				},
				{
					ID:           ulid.New(),
					BatchID:      batchID,
					EventType:    observability.TelemetryEventTypeObservationCreate,
					EventPayload: map[string]interface{}{"observation_id": "test2"},
				},
			},
			mockSetup: func(repo *MockEventRepository) {
				repo.On("MarkAsProcessed", mock.Anything, mock.AnythingOfType("ulid.ULID"), mock.AnythingOfType("time.Time")).
					Return(nil).Times(2)
			},
			expectedErr: "",
			checkResult: func(t *testing.T, result *observability.EventProcessingResult) {
				assert.NotNil(t, result)
				assert.Equal(t, 2, result.ProcessedCount)
				assert.Equal(t, 0, result.FailedCount)
				assert.Equal(t, 100.0, result.SuccessRate)
			},
		},
		{
			name:   "success - empty batch",
			events: []*observability.TelemetryEvent{},
			mockSetup: func(repo *MockEventRepository) {
				// No calls expected
			},
			expectedErr: "",
			checkResult: func(t *testing.T, result *observability.EventProcessingResult) {
				assert.NotNil(t, result)
				assert.Equal(t, 0, result.ProcessedCount)
				assert.Equal(t, 100.0, result.SuccessRate)
			},
		},
		{
			name: "partial success - some events fail validation",
			events: []*observability.TelemetryEvent{
				{
					ID:           ulid.New(),
					BatchID:      batchID,
					EventType:    observability.TelemetryEventTypeTraceCreate,
					EventPayload: map[string]interface{}{"data": "value"},
				},
				nil, // Will be skipped
			},
			mockSetup: func(repo *MockEventRepository) {
				repo.On("MarkAsProcessed", mock.Anything, mock.AnythingOfType("ulid.ULID"), mock.AnythingOfType("time.Time")).
					Return(nil).Once()
			},
			expectedErr: "",
			checkResult: func(t *testing.T, result *observability.EventProcessingResult) {
				assert.NotNil(t, result)
				assert.Equal(t, 1, result.ProcessedCount)
				assert.Equal(t, 1, result.NotProcessedCount)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEventRepo := new(MockEventRepository)
			mockBatchRepo := new(MockBatchRepository)

			tt.mockSetup(mockEventRepo)

			service := NewTelemetryEventService(mockEventRepo, mockBatchRepo)

			result, err := service.ProcessEventsBatch(context.Background(), tt.events)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}

			mockEventRepo.AssertExpectations(t)
		})
	}
}

// ============================================================================
// Query Tests
// ============================================================================

func TestTelemetryEventService_GetEventsForRetry(t *testing.T) {
	tests := []struct {
		name        string
		maxRetries  int
		limit       int
		mockSetup   func(*MockEventRepository)
		expectedErr string
		checkResult func(*testing.T, []*observability.TelemetryEvent)
	}{
		{
			name:       "success - get events for retry",
			maxRetries: 3,
			limit:      10,
			mockSetup: func(repo *MockEventRepository) {
				events := []*observability.TelemetryEvent{
					{ID: ulid.New(), RetryCount: 1},
					{ID: ulid.New(), RetryCount: 2},
				}
				repo.On("GetEventsForRetry", mock.Anything, 3, 10).
					Return(events, nil)
			},
			expectedErr: "",
			checkResult: func(t *testing.T, events []*observability.TelemetryEvent) {
				assert.Len(t, events, 2)
			},
		},
		{
			name:       "success - apply default limits",
			maxRetries: 0,
			limit:      0,
			mockSetup: func(repo *MockEventRepository) {
				repo.On("GetEventsForRetry", mock.Anything, 3, 100).
					Return([]*observability.TelemetryEvent{}, nil)
			},
			expectedErr: "",
			checkResult: func(t *testing.T, events []*observability.TelemetryEvent) {
				assert.NotNil(t, events)
			},
		},
		{
			name:       "success - cap limit at 1000",
			maxRetries: 5,
			limit:      5000,
			mockSetup: func(repo *MockEventRepository) {
				repo.On("GetEventsForRetry", mock.Anything, 5, 1000).
					Return([]*observability.TelemetryEvent{}, nil)
			},
			expectedErr: "",
			checkResult: func(t *testing.T, events []*observability.TelemetryEvent) {
				assert.NotNil(t, events)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEventRepo := new(MockEventRepository)
			mockBatchRepo := new(MockBatchRepository)

			tt.mockSetup(mockEventRepo)

			service := NewTelemetryEventService(mockEventRepo, mockBatchRepo)

			events, err := service.GetEventsForRetry(context.Background(), tt.maxRetries, tt.limit)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			if tt.checkResult != nil {
				tt.checkResult(t, events)
			}

			mockEventRepo.AssertExpectations(t)
		})
	}
}

func TestTelemetryEventService_GetFailedEvents(t *testing.T) {
	batchID := ulid.New()

	tests := []struct {
		name        string
		batchID     *ulid.ULID
		limit       int
		offset      int
		mockSetup   func(*MockEventRepository)
		expectedErr string
		checkResult func(*testing.T, []*observability.TelemetryEvent)
	}{
		{
			name:    "success - get failed events with batch filter",
			batchID: &batchID,
			limit:   10,
			offset:  0,
			mockSetup: func(repo *MockEventRepository) {
				events := []*observability.TelemetryEvent{
					{ID: ulid.New(), BatchID: batchID, ErrorMessage: stringPtr("error1")},
					{ID: ulid.New(), BatchID: batchID, ErrorMessage: stringPtr("error2")},
				}
				repo.On("GetFailedEvents", mock.Anything, &batchID, 10, 0).
					Return(events, nil)
			},
			expectedErr: "",
			checkResult: func(t *testing.T, events []*observability.TelemetryEvent) {
				assert.Len(t, events, 2)
			},
		},
		{
			name:    "success - get failed events without batch filter",
			batchID: nil,
			limit:   20,
			offset:  10,
			mockSetup: func(repo *MockEventRepository) {
				events := []*observability.TelemetryEvent{
					{ID: ulid.New(), ErrorMessage: stringPtr("error")},
				}
				repo.On("GetFailedEvents", mock.Anything, (*ulid.ULID)(nil), 20, 10).
					Return(events, nil)
			},
			expectedErr: "",
			checkResult: func(t *testing.T, events []*observability.TelemetryEvent) {
				assert.Len(t, events, 1)
			},
		},
		{
			name:    "success - apply default limits",
			batchID: nil,
			limit:   0,
			offset:  -5,
			mockSetup: func(repo *MockEventRepository) {
				repo.On("GetFailedEvents", mock.Anything, (*ulid.ULID)(nil), 50, 0).
					Return([]*observability.TelemetryEvent{}, nil)
			},
			expectedErr: "",
			checkResult: func(t *testing.T, events []*observability.TelemetryEvent) {
				assert.NotNil(t, events)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEventRepo := new(MockEventRepository)
			mockBatchRepo := new(MockBatchRepository)

			tt.mockSetup(mockEventRepo)

			service := NewTelemetryEventService(mockEventRepo, mockBatchRepo)

			events, err := service.GetFailedEvents(context.Background(), tt.batchID, tt.limit, tt.offset)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			if tt.checkResult != nil {
				tt.checkResult(t, events)
			}

			mockEventRepo.AssertExpectations(t)
		})
	}
}

// ============================================================================
// Batch Operation Tests
// ============================================================================

func TestTelemetryEventService_CleanupFailedEvents(t *testing.T) {
	olderThan := time.Now().Add(-24 * time.Hour)

	tests := []struct {
		name        string
		olderThan   time.Time
		mockSetup   func(*MockEventRepository)
		expectedErr string
		checkResult func(*testing.T, int64)
	}{
		{
			name:      "success - cleanup failed events",
			olderThan: olderThan,
			mockSetup: func(repo *MockEventRepository) {
				repo.On("DeleteFailedEvents", mock.Anything, olderThan).
					Return(int64(42), nil)
			},
			expectedErr: "",
			checkResult: func(t *testing.T, deletedCount int64) {
				assert.Equal(t, int64(42), deletedCount)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEventRepo := new(MockEventRepository)
			mockBatchRepo := new(MockBatchRepository)

			tt.mockSetup(mockEventRepo)

			service := NewTelemetryEventService(mockEventRepo, mockBatchRepo)

			deletedCount, err := service.CleanupFailedEvents(context.Background(), tt.olderThan)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			if tt.checkResult != nil {
				tt.checkResult(t, deletedCount)
			}

			mockEventRepo.AssertExpectations(t)
		})
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

func stringPtr(s string) *string {
	return &s
}
