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
// Mock TelemetryBatchRepository
// ============================================================================

type MockTelemetryBatchRepository struct {
	mock.Mock
}

func (m *MockTelemetryBatchRepository) Create(ctx context.Context, batch *observability.TelemetryBatch) error {
	args := m.Called(ctx, batch)
	return args.Error(0)
}

func (m *MockTelemetryBatchRepository) GetByID(ctx context.Context, id ulid.ULID) (*observability.TelemetryBatch, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*observability.TelemetryBatch), args.Error(1)
}

func (m *MockTelemetryBatchRepository) Update(ctx context.Context, batch *observability.TelemetryBatch) error {
	args := m.Called(ctx, batch)
	return args.Error(0)
}

func (m *MockTelemetryBatchRepository) Delete(ctx context.Context, id ulid.ULID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTelemetryBatchRepository) GetByProjectID(ctx context.Context, projectID ulid.ULID, limit, offset int) ([]*observability.TelemetryBatch, error) {
	args := m.Called(ctx, projectID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.TelemetryBatch), args.Error(1)
}

func (m *MockTelemetryBatchRepository) GetActiveByProjectID(ctx context.Context, projectID ulid.ULID) ([]*observability.TelemetryBatch, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.TelemetryBatch), args.Error(1)
}

func (m *MockTelemetryBatchRepository) GetByStatus(ctx context.Context, status observability.BatchStatus, limit, offset int) ([]*observability.TelemetryBatch, error) {
	args := m.Called(ctx, status, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.TelemetryBatch), args.Error(1)
}

func (m *MockTelemetryBatchRepository) GetProcessingBatches(ctx context.Context) ([]*observability.TelemetryBatch, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.TelemetryBatch), args.Error(1)
}

func (m *MockTelemetryBatchRepository) GetCompletedBatches(ctx context.Context, projectID ulid.ULID, limit, offset int) ([]*observability.TelemetryBatch, error) {
	args := m.Called(ctx, projectID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.TelemetryBatch), args.Error(1)
}

func (m *MockTelemetryBatchRepository) CreateBatch(ctx context.Context, batches []*observability.TelemetryBatch) error {
	args := m.Called(ctx, batches)
	return args.Error(0)
}

func (m *MockTelemetryBatchRepository) UpdateBatch(ctx context.Context, batches []*observability.TelemetryBatch) error {
	args := m.Called(ctx, batches)
	return args.Error(0)
}

func (m *MockTelemetryBatchRepository) SearchBatches(ctx context.Context, filter *observability.TelemetryBatchFilter) ([]*observability.TelemetryBatch, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*observability.TelemetryBatch), args.Int(1), args.Error(2)
}

func (m *MockTelemetryBatchRepository) GetBatchWithEvents(ctx context.Context, id ulid.ULID) (*observability.TelemetryBatch, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*observability.TelemetryBatch), args.Error(1)
}

func (m *MockTelemetryBatchRepository) UpdateBatchStatus(ctx context.Context, batchID ulid.ULID, status observability.BatchStatus, processingTimeMs *int) error {
	args := m.Called(ctx, batchID, status, processingTimeMs)
	return args.Error(0)
}

func (m *MockTelemetryBatchRepository) GetBatchStats(ctx context.Context, id ulid.ULID) (*observability.BatchStats, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*observability.BatchStats), args.Error(1)
}

func (m *MockTelemetryBatchRepository) GetBatchProcessingMetrics(ctx context.Context, filter *observability.TelemetryBatchFilter) (*observability.BatchProcessingMetrics, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*observability.BatchProcessingMetrics), args.Error(1)
}

func (m *MockTelemetryBatchRepository) GetBatchThroughputStats(ctx context.Context, projectID ulid.ULID, timeWindow time.Duration) (*observability.BatchThroughputStats, error) {
	args := m.Called(ctx, projectID, timeWindow)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*observability.BatchThroughputStats), args.Error(1)
}

func (m *MockTelemetryBatchRepository) GetBatchesByTimeRange(ctx context.Context, projectID ulid.ULID, startTime, endTime time.Time, limit, offset int) ([]*observability.TelemetryBatch, error) {
	args := m.Called(ctx, projectID, startTime, endTime, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.TelemetryBatch), args.Error(1)
}

func (m *MockTelemetryBatchRepository) CountBatches(ctx context.Context, filter *observability.TelemetryBatchFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTelemetryBatchRepository) GetRecentBatches(ctx context.Context, projectID ulid.ULID, limit int) ([]*observability.TelemetryBatch, error) {
	args := m.Called(ctx, projectID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.TelemetryBatch), args.Error(1)
}

// ============================================================================
// Mock TelemetryEventRepository
// ============================================================================

type MockTelemetryEventRepository struct {
	mock.Mock
}

func (m *MockTelemetryEventRepository) Create(ctx context.Context, event *observability.TelemetryEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockTelemetryEventRepository) GetByID(ctx context.Context, id ulid.ULID) (*observability.TelemetryEvent, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*observability.TelemetryEvent), args.Error(1)
}

func (m *MockTelemetryEventRepository) Update(ctx context.Context, event *observability.TelemetryEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockTelemetryEventRepository) Delete(ctx context.Context, id ulid.ULID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTelemetryEventRepository) GetByBatchID(ctx context.Context, batchID ulid.ULID) ([]*observability.TelemetryEvent, error) {
	args := m.Called(ctx, batchID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.TelemetryEvent), args.Error(1)
}

func (m *MockTelemetryEventRepository) GetUnprocessedByBatchID(ctx context.Context, batchID ulid.ULID) ([]*observability.TelemetryEvent, error) {
	args := m.Called(ctx, batchID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.TelemetryEvent), args.Error(1)
}

func (m *MockTelemetryEventRepository) GetFailedByBatchID(ctx context.Context, batchID ulid.ULID) ([]*observability.TelemetryEvent, error) {
	args := m.Called(ctx, batchID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.TelemetryEvent), args.Error(1)
}

func (m *MockTelemetryEventRepository) GetByEventType(ctx context.Context, eventType observability.TelemetryEventType, limit, offset int) ([]*observability.TelemetryEvent, error) {
	args := m.Called(ctx, eventType, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.TelemetryEvent), args.Error(1)
}

func (m *MockTelemetryEventRepository) GetUnprocessedByType(ctx context.Context, eventType observability.TelemetryEventType, limit int) ([]*observability.TelemetryEvent, error) {
	args := m.Called(ctx, eventType, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.TelemetryEvent), args.Error(1)
}

func (m *MockTelemetryEventRepository) MarkAsProcessed(ctx context.Context, eventID ulid.ULID, processedAt time.Time) error {
	args := m.Called(ctx, eventID, processedAt)
	return args.Error(0)
}

func (m *MockTelemetryEventRepository) MarkAsFailed(ctx context.Context, eventID ulid.ULID, errorMessage string) error {
	args := m.Called(ctx, eventID, errorMessage)
	return args.Error(0)
}

func (m *MockTelemetryEventRepository) CreateBatch(ctx context.Context, events []*observability.TelemetryEvent) error {
	args := m.Called(ctx, events)
	return args.Error(0)
}

func (m *MockTelemetryEventRepository) UpdateBatch(ctx context.Context, events []*observability.TelemetryEvent) error {
	args := m.Called(ctx, events)
	return args.Error(0)
}

func (m *MockTelemetryEventRepository) ProcessBatch(ctx context.Context, batchID ulid.ULID, processor func([]*observability.TelemetryEvent) error) error {
	args := m.Called(ctx, batchID, processor)
	return args.Error(0)
}

func (m *MockTelemetryEventRepository) IncrementRetryCount(ctx context.Context, id ulid.ULID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTelemetryEventRepository) GetEventsForRetry(ctx context.Context, maxRetries int, limit int) ([]*observability.TelemetryEvent, error) {
	args := m.Called(ctx, maxRetries, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.TelemetryEvent), args.Error(1)
}

func (m *MockTelemetryEventRepository) GetFailedEvents(ctx context.Context, batchID *ulid.ULID, limit, offset int) ([]*observability.TelemetryEvent, error) {
	args := m.Called(ctx, batchID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.TelemetryEvent), args.Error(1)
}

func (m *MockTelemetryEventRepository) DeleteFailedEvents(ctx context.Context, olderThan time.Time) (int64, error) {
	args := m.Called(ctx, olderThan)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTelemetryEventRepository) GetEventStats(ctx context.Context, filter *observability.TelemetryEventFilter) (*observability.TelemetryEventStats, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*observability.TelemetryEventStats), args.Error(1)
}

func (m *MockTelemetryEventRepository) GetEventTypeDistribution(ctx context.Context, batchID *ulid.ULID) (map[observability.TelemetryEventType]int, error) {
	args := m.Called(ctx, batchID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[observability.TelemetryEventType]int), args.Error(1)
}

func (m *MockTelemetryEventRepository) CountEvents(ctx context.Context, filter *observability.TelemetryEventFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

// ============================================================================
// Mock TelemetryDeduplicationRepository
// ============================================================================

type MockTelemetryDeduplicationRepository struct {
	mock.Mock
}

func (m *MockTelemetryDeduplicationRepository) Create(ctx context.Context, dedup *observability.TelemetryEventDeduplication) error {
	args := m.Called(ctx, dedup)
	return args.Error(0)
}

func (m *MockTelemetryDeduplicationRepository) GetByEventID(ctx context.Context, eventID ulid.ULID) (*observability.TelemetryEventDeduplication, error) {
	args := m.Called(ctx, eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*observability.TelemetryEventDeduplication), args.Error(1)
}

func (m *MockTelemetryDeduplicationRepository) Delete(ctx context.Context, eventID ulid.ULID) error {
	args := m.Called(ctx, eventID)
	return args.Error(0)
}

func (m *MockTelemetryDeduplicationRepository) Exists(ctx context.Context, eventID ulid.ULID) (bool, error) {
	args := m.Called(ctx, eventID)
	return args.Bool(0), args.Error(1)
}

func (m *MockTelemetryDeduplicationRepository) ExistsInBatch(ctx context.Context, eventID ulid.ULID, batchID ulid.ULID) (bool, error) {
	args := m.Called(ctx, eventID, batchID)
	return args.Bool(0), args.Error(1)
}

func (m *MockTelemetryDeduplicationRepository) ExistsWithRedisCheck(ctx context.Context, eventID ulid.ULID) (bool, bool, error) {
	args := m.Called(ctx, eventID)
	return args.Bool(0), args.Bool(1), args.Error(2)
}

func (m *MockTelemetryDeduplicationRepository) StoreInRedis(ctx context.Context, eventID ulid.ULID, batchID ulid.ULID, ttl time.Duration) error {
	args := m.Called(ctx, eventID, batchID, ttl)
	return args.Error(0)
}

func (m *MockTelemetryDeduplicationRepository) GetFromRedis(ctx context.Context, eventID ulid.ULID) (*ulid.ULID, error) {
	args := m.Called(ctx, eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ulid.ULID), args.Error(1)
}

func (m *MockTelemetryDeduplicationRepository) CreateBatch(ctx context.Context, entries []*observability.TelemetryEventDeduplication) error {
	args := m.Called(ctx, entries)
	return args.Error(0)
}

func (m *MockTelemetryDeduplicationRepository) DeleteByBatchID(ctx context.Context, batchID ulid.ULID) error {
	args := m.Called(ctx, batchID)
	return args.Error(0)
}

func (m *MockTelemetryDeduplicationRepository) CleanupExpired(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTelemetryDeduplicationRepository) GetExpiredEntries(ctx context.Context, limit int) ([]*observability.TelemetryEventDeduplication, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.TelemetryEventDeduplication), args.Error(1)
}

func (m *MockTelemetryDeduplicationRepository) BatchCleanup(ctx context.Context, olderThan time.Time, batchSize int) (int64, error) {
	args := m.Called(ctx, olderThan, batchSize)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTelemetryDeduplicationRepository) GetByProjectID(ctx context.Context, projectID ulid.ULID, limit, offset int) ([]*observability.TelemetryEventDeduplication, error) {
	args := m.Called(ctx, projectID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.TelemetryEventDeduplication), args.Error(1)
}

func (m *MockTelemetryDeduplicationRepository) CountByProjectID(ctx context.Context, projectID ulid.ULID) (int64, error) {
	args := m.Called(ctx, projectID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTelemetryDeduplicationRepository) CleanupByProjectID(ctx context.Context, projectID ulid.ULID, olderThan time.Time) (int64, error) {
	args := m.Called(ctx, projectID, olderThan)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTelemetryDeduplicationRepository) CheckBatchDuplicates(ctx context.Context, eventIDs []ulid.ULID) ([]ulid.ULID, error) {
	args := m.Called(ctx, eventIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]ulid.ULID), args.Error(1)
}

// ============================================================================
// Core CRUD Tests
// ============================================================================

// TestTelemetryBatchService_CreateBatch tests the CreateBatch method

// ============================================================================
// HIGH-VALUE TESTS: Batch Processing & Analytics
// ============================================================================

func TestTelemetryBatchService_ProcessEventsBatch(t *testing.T) {
	batchID := ulid.New()

	tests := []struct {
		name        string
		events      []*observability.TelemetryEvent
		mockSetup   func(*MockTelemetryBatchRepository, *MockTelemetryEventRepository, *MockTelemetryDeduplicationRepository)
		expectedErr string
		checkResult func(*testing.T, *observability.BatchProcessingResult)
	}{
		{
			name: "success - process valid events",
			events: []*observability.TelemetryEvent{
				{
					ID:           ulid.New(),
					BatchID:      batchID,
					EventType:    observability.TelemetryEventTypeTraceCreate,
					EventPayload: map[string]any{"trace_id": "test"},
				},
				{
					ID:           ulid.New(),
					BatchID:      batchID,
					EventType:    observability.TelemetryEventTypeObservationCreate,
					EventPayload: map[string]any{"observation_id": "test"},
				},
			},
			mockSetup:   func(batchRepo *MockTelemetryBatchRepository, eventRepo *MockTelemetryEventRepository, dedupRepo *MockTelemetryDeduplicationRepository) {},
			expectedErr: "",
			checkResult: func(t *testing.T, result *observability.BatchProcessingResult) {
				assert.NotNil(t, result)
				assert.Equal(t, 2, result.TotalEvents)
				assert.Equal(t, 2, result.ProcessedEvents)
				assert.Equal(t, 0, result.FailedEvents)
				assert.Equal(t, 100.0, result.SuccessRate)
			},
		},
		{
			name:        "success - empty batch",
			events:      []*observability.TelemetryEvent{},
			mockSetup:   func(batchRepo *MockTelemetryBatchRepository, eventRepo *MockTelemetryEventRepository, dedupRepo *MockTelemetryDeduplicationRepository) {},
			expectedErr: "",
			checkResult: func(t *testing.T, result *observability.BatchProcessingResult) {
				assert.NotNil(t, result)
				assert.Equal(t, 0, result.TotalEvents)
				assert.Equal(t, 100.0, result.SuccessRate)
			},
		},
		{
			name: "partial success - some events invalid",
			events: []*observability.TelemetryEvent{
				{
					ID:           ulid.New(),
					BatchID:      batchID,
					EventType:    observability.TelemetryEventTypeTraceCreate,
					EventPayload: map[string]any{"trace_id": "test"},
				},
				nil, // nil event
				{
					ID:           ulid.New(),
					BatchID:      batchID,
					EventType:    "", // Invalid - missing event type
					EventPayload: map[string]any{},
				},
			},
			mockSetup:   func(batchRepo *MockTelemetryBatchRepository, eventRepo *MockTelemetryEventRepository, dedupRepo *MockTelemetryDeduplicationRepository) {},
			expectedErr: "",
			checkResult: func(t *testing.T, result *observability.BatchProcessingResult) {
				assert.NotNil(t, result)
				assert.Equal(t, 3, result.TotalEvents)
				assert.Equal(t, 1, result.ProcessedEvents)
				assert.Equal(t, 2, result.FailedEvents)
				assert.Len(t, result.Errors, 1) // Only validation errors are recorded
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBatchRepo := new(MockTelemetryBatchRepository)
			mockEventRepo := new(MockTelemetryEventRepository)
			mockDedupRepo := new(MockTelemetryDeduplicationRepository)

			tt.mockSetup(mockBatchRepo, mockEventRepo, mockDedupRepo)

			service := NewTelemetryBatchService(mockBatchRepo, mockEventRepo, mockDedupRepo)

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
		})
	}
}

// TestTelemetryBatchService_ListBatches tests the ListBatches method
func TestTelemetryBatchService_GetBatchStats(t *testing.T) {
	batchID := ulid.New()

	tests := []struct {
		name        string
		batchID     ulid.ULID
		mockSetup   func(*MockTelemetryBatchRepository)
		expectedErr string
		checkResult func(*testing.T, *observability.BatchStats)
	}{
		{
			name:    "success - stats retrieved",
			batchID: batchID,
			mockSetup: func(repo *MockTelemetryBatchRepository) {
				stats := &observability.BatchStats{
					BatchID:         batchID,
					TotalEvents:     100,
					ProcessedEvents: 95,
					FailedEvents:    5,
					SuccessRate:     95.0,
				}
				repo.On("GetBatchStats", mock.Anything, batchID).
					Return(stats, nil)
			},
			expectedErr: "",
			checkResult: func(t *testing.T, stats *observability.BatchStats) {
				assert.NotNil(t, stats)
				assert.Equal(t, batchID, stats.BatchID)
				assert.Equal(t, 95.0, stats.SuccessRate)
			},
		},
		{
			name:    "error - zero batch ID",
			batchID: ulid.ULID{},
			mockSetup: func(repo *MockTelemetryBatchRepository) {
				// No calls expected
			},
			expectedErr: "batch ID cannot be zero",
			checkResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBatchRepo := new(MockTelemetryBatchRepository)
			mockEventRepo := new(MockTelemetryEventRepository)
			mockDedupRepo := new(MockTelemetryDeduplicationRepository)

			tt.mockSetup(mockBatchRepo)

			service := NewTelemetryBatchService(mockBatchRepo, mockEventRepo, mockDedupRepo)

			stats, err := service.GetBatchStats(context.Background(), tt.batchID)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			if tt.checkResult != nil {
				tt.checkResult(t, stats)
			}

			mockBatchRepo.AssertExpectations(t)
		})
	}
}

// TestTelemetryBatchService_GetThroughputStats tests the GetThroughputStats method
func TestTelemetryBatchService_GetThroughputStats(t *testing.T) {
	projectID := ulid.New()

	tests := []struct {
		name        string
		projectID   ulid.ULID
		timeWindow  time.Duration
		mockSetup   func(*MockTelemetryBatchRepository)
		expectedErr string
		checkResult func(*testing.T, *observability.BatchThroughputStats)
	}{
		{
			name:       "success - throughput stats retrieved",
			projectID:  projectID,
			timeWindow: time.Hour,
			mockSetup: func(repo *MockTelemetryBatchRepository) {
				stats := &observability.BatchThroughputStats{
					BatchesPerMinute:      0.83,
					EventsPerMinute:       83.33,
					AverageEventsPerBatch: 100.0,
					PeakThroughput:        150.0,
					ThroughputTrend:       "stable",
					TimeWindow:            time.Hour,
					LastCalculated:        time.Now(),
				}
				repo.On("GetBatchThroughputStats", mock.Anything, projectID, time.Hour).
					Return(stats, nil)
			},
			expectedErr: "",
			checkResult: func(t *testing.T, stats *observability.BatchThroughputStats) {
				assert.NotNil(t, stats)
				assert.Equal(t, 0.83, stats.BatchesPerMinute)
				assert.Equal(t, "stable", stats.ThroughputTrend)
			},
		},
		{
			name:       "error - zero project ID",
			projectID:  ulid.ULID{},
			timeWindow: time.Hour,
			mockSetup: func(repo *MockTelemetryBatchRepository) {
				// No calls expected
			},
			expectedErr: "project ID cannot be zero",
			checkResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBatchRepo := new(MockTelemetryBatchRepository)
			mockEventRepo := new(MockTelemetryEventRepository)
			mockDedupRepo := new(MockTelemetryDeduplicationRepository)

			tt.mockSetup(mockBatchRepo)

			service := NewTelemetryBatchService(mockBatchRepo, mockEventRepo, mockDedupRepo)

			stats, err := service.GetThroughputStats(context.Background(), tt.projectID, tt.timeWindow)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			if tt.checkResult != nil {
				tt.checkResult(t, stats)
			}

			mockBatchRepo.AssertExpectations(t)
		})
	}
}
