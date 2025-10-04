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
// Test Suite for ObservationService
// ============================================================================

// TestObservationService_CreateObservation tests the CreateObservation method

// ============================================================================
// HIGH-VALUE TESTS: Complex Business Logic & Critical Paths
// ============================================================================

func TestObservationService_CompleteObservation(t *testing.T) {
	observationID := ulid.New()
	traceID := ulid.New()
	startTime := time.Now().Add(-5 * time.Second)
	endTime := time.Now()

	tests := []struct {
		name           string
		id             ulid.ULID
		completionData *observability.ObservationCompletion
		mockSetup      func(*MockObservationRepository, *MockEventPublisher)
		expectedErr    error
		checkResult    func(*testing.T, *observability.Observation)
	}{
		{
			name: "success - complete observation with all data",
			id:   observationID,
			completionData: &observability.ObservationCompletion{
				EndTime: endTime,
				Output:  map[string]interface{}{"result": "success"},
				Usage: &observability.TokenUsage{
					PromptTokens:     100,
					CompletionTokens: 50,
					TotalTokens:      150,
				},
				Cost: &observability.CostCalculation{
					InputCost:  0.001,
					OutputCost: 0.0005,
					TotalCost:  0.0015,
				},
			},
			mockSetup: func(repo *MockObservationRepository, publisher *MockEventPublisher) {
				repo.On("GetByID", mock.Anything, observationID).
					Return(&observability.Observation{
						ID:                    observationID,
						TraceID:               traceID,
						ExternalObservationID: "ext-obs-complete",
						Name:                  "Complete Me",
						Type:                  observability.ObservationTypeLLM,
						StartTime:             startTime,
						EndTime:               nil,
					}, nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*observability.Observation")).
					Return(nil)
				publisher.On("Publish", mock.Anything, mock.AnythingOfType("*observability.Event")).
					Return(nil)
			},
			expectedErr: nil,
			checkResult: func(t *testing.T, obs *observability.Observation) {
				assert.NotNil(t, obs)
				assert.NotNil(t, obs.EndTime)
				assert.NotNil(t, obs.LatencyMs)
				assert.Equal(t, 100, obs.PromptTokens)
				assert.Equal(t, 50, obs.CompletionTokens)
				assert.NotNil(t, obs.TotalCost)
			},
		},
		{
			name:           "error - empty observation ID",
			id:             ulid.ULID{},
			completionData: &observability.ObservationCompletion{EndTime: endTime},
			mockSetup: func(repo *MockObservationRepository, publisher *MockEventPublisher) {
				// No calls expected
			},
			expectedErr: observability.NewObservabilityError(
				observability.ErrCodeInvalidObservationID,
				"observation ID cannot be empty",
			),
			checkResult: nil,
		},
		{
			name:           "error - nil completion data",
			id:             observationID,
			completionData: nil,
			mockSetup: func(repo *MockObservationRepository, publisher *MockEventPublisher) {
				// No calls expected
			},
			expectedErr: observability.NewObservabilityError(
				observability.ErrCodeValidationFailed,
				"completion data cannot be nil",
			),
			checkResult: nil,
		},
		{
			name: "error - observation already completed",
			id:   observationID,
			completionData: &observability.ObservationCompletion{
				EndTime: endTime,
			},
			mockSetup: func(repo *MockObservationRepository, publisher *MockEventPublisher) {
				completedTime := time.Now().Add(-1 * time.Minute)
				repo.On("GetByID", mock.Anything, observationID).
					Return(&observability.Observation{
						ID:                    observationID,
						TraceID:               traceID,
						ExternalObservationID: "ext-obs-already-complete",
						Name:                  "Already Done",
						Type:                  observability.ObservationTypeLLM,
						StartTime:             startTime,
						EndTime:               &completedTime,
					}, nil)
			},
			expectedErr: observability.NewObservabilityError(
				observability.ErrCodeObservationAlreadyCompleted,
				"observation is already completed",
			),
			checkResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockObsRepo := new(MockObservationRepository)
			mockTraceRepo := new(MockTraceRepository)
			mockPublisher := new(MockEventPublisher)

			tt.mockSetup(mockObsRepo, mockPublisher)

			service := NewObservationService(mockObsRepo, mockTraceRepo, mockPublisher)

			result, err := service.CompleteObservation(context.Background(), tt.id, tt.completionData)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}

			mockObsRepo.AssertExpectations(t)
			mockPublisher.AssertExpectations(t)
		})
	}
}

// TestObservationService_DeleteObservation tests the DeleteObservation method
func TestObservationService_CreateObservationsBatch(t *testing.T) {
	traceID := ulid.New()

	tests := []struct {
		name         string
		observations []*observability.Observation
		mockSetup    func(*MockObservationRepository, *MockEventPublisher)
		expectedErr  error
		checkResult  func(*testing.T, []*observability.Observation)
	}{
		{
			name: "success - create batch of observations",
			observations: []*observability.Observation{
				{
					TraceID:               traceID,
					ExternalObservationID: "ext-obs-batch-1",
					Name:                  "Batch Obs 1",
					Type:                  observability.ObservationTypeLLM,
					StartTime:             time.Now(),
				},
				{
					TraceID:               traceID,
					ExternalObservationID: "ext-obs-batch-2",
					Name:                  "Batch Obs 2",
					Type:                  observability.ObservationTypeSpan,
					StartTime:             time.Now(),
				},
			},
			mockSetup: func(repo *MockObservationRepository, publisher *MockEventPublisher) {
				repo.On("CreateBatch", mock.Anything, mock.AnythingOfType("[]*observability.Observation")).
					Return(nil)
				publisher.On("PublishBatch", mock.Anything, mock.AnythingOfType("[]*observability.Event")).
					Return(nil)
			},
			expectedErr: nil,
			checkResult: func(t *testing.T, obs []*observability.Observation) {
				assert.Len(t, obs, 2)
				assert.NotEqual(t, ulid.ULID{}, obs[0].ID)
				assert.NotEqual(t, ulid.ULID{}, obs[1].ID)
				assert.NotZero(t, obs[0].CreatedAt)
			},
		},
		{
			name:         "success - empty batch",
			observations: []*observability.Observation{},
			mockSetup: func(repo *MockObservationRepository, publisher *MockEventPublisher) {
				// No calls expected
			},
			expectedErr: nil,
			checkResult: func(t *testing.T, obs []*observability.Observation) {
				assert.Len(t, obs, 0)
			},
		},
		{
			name: "error - validation failure in batch",
			observations: []*observability.Observation{
				{
					TraceID:               traceID,
					ExternalObservationID: "ext-obs-invalid",
					Name:                  "", // Invalid - missing name
					Type:                  observability.ObservationTypeLLM,
					StartTime:             time.Now(),
				},
			},
			mockSetup: func(repo *MockObservationRepository, publisher *MockEventPublisher) {
				// No calls expected
			},
			expectedErr: observability.NewValidationError("name", "observation name is required"),
			checkResult: nil,
		},
		{
			name: "error - repository batch failure",
			observations: []*observability.Observation{
				{
					TraceID:               traceID,
					ExternalObservationID: "ext-obs-fail",
					Name:                  "Fail Obs",
					Type:                  observability.ObservationTypeLLM,
					StartTime:             time.Now(),
				},
			},
			mockSetup: func(repo *MockObservationRepository, publisher *MockEventPublisher) {
				repo.On("CreateBatch", mock.Anything, mock.AnythingOfType("[]*observability.Observation")).
					Return(assert.AnError)
			},
			expectedErr: assert.AnError,
			checkResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockObsRepo := new(MockObservationRepository)
			mockTraceRepo := new(MockTraceRepository)
			mockPublisher := new(MockEventPublisher)

			tt.mockSetup(mockObsRepo, mockPublisher)

			service := NewObservationService(mockObsRepo, mockTraceRepo, mockPublisher)

			result, err := service.CreateObservationsBatch(context.Background(), tt.observations)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}

			mockObsRepo.AssertExpectations(t)
			mockPublisher.AssertExpectations(t)
		})
	}
}

// TestObservationService_UpdateObservationsBatch tests the UpdateObservationsBatch method
func TestObservationService_UpdateObservationsBatch(t *testing.T) {
	traceID := ulid.New()
	obs1ID := ulid.New()
	obs2ID := ulid.New()

	tests := []struct {
		name         string
		observations []*observability.Observation
		mockSetup    func(*MockObservationRepository)
		expectedErr  error
		checkResult  func(*testing.T, []*observability.Observation)
	}{
		{
			name: "success - update batch of observations",
			observations: []*observability.Observation{
				{
					ID:                    obs1ID,
					TraceID:               traceID,
					ExternalObservationID: "ext-obs-update-1",
					Name:                  "Updated Obs 1",
					Type:                  observability.ObservationTypeLLM,
					StartTime:             time.Now(),
				},
				{
					ID:                    obs2ID,
					TraceID:               traceID,
					ExternalObservationID: "ext-obs-update-2",
					Name:                  "Updated Obs 2",
					Type:                  observability.ObservationTypeSpan,
					StartTime:             time.Now(),
				},
			},
			mockSetup: func(repo *MockObservationRepository) {
				repo.On("UpdateBatch", mock.Anything, mock.AnythingOfType("[]*observability.Observation")).
					Return(nil)
			},
			expectedErr: nil,
			checkResult: func(t *testing.T, obs []*observability.Observation) {
				assert.Len(t, obs, 2)
				assert.NotZero(t, obs[0].UpdatedAt)
			},
		},
		{
			name:         "success - empty batch",
			observations: []*observability.Observation{},
			mockSetup: func(repo *MockObservationRepository) {
				// No calls expected
			},
			expectedErr: nil,
			checkResult: func(t *testing.T, obs []*observability.Observation) {
				assert.Len(t, obs, 0)
			},
		},
		{
			name: "error - observation without ID",
			observations: []*observability.Observation{
				{
					TraceID:               traceID,
					ExternalObservationID: "ext-obs-no-id",
					Name:                  "No ID Obs",
					Type:                  observability.ObservationTypeLLM,
					StartTime:             time.Now(),
				},
			},
			mockSetup: func(repo *MockObservationRepository) {
				// No calls expected
			},
			expectedErr: observability.NewObservabilityError(
				observability.ErrCodeInvalidObservationID,
				"observation ID cannot be empty",
			),
			checkResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockObsRepo := new(MockObservationRepository)
			mockTraceRepo := new(MockTraceRepository)
			mockPublisher := new(MockEventPublisher)

			tt.mockSetup(mockObsRepo)

			service := NewObservationService(mockObsRepo, mockTraceRepo, mockPublisher)

			result, err := service.UpdateObservationsBatch(context.Background(), tt.observations)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}

			mockObsRepo.AssertExpectations(t)
		})
	}
}

// ============================================================================
// Analytics Tests
// ============================================================================

// TestObservationService_GetObservationStats tests the GetObservationStats method
func TestObservationService_GetObservationAnalytics(t *testing.T) {
	projectID := ulid.New()

	tests := []struct {
		name        string
		filter      *observability.AnalyticsFilter
		mockSetup   func()
		expectedErr error
		checkResult func(*testing.T, *observability.ObservationAnalytics)
	}{
		{
			name: "success - get observation analytics",
			filter: &observability.AnalyticsFilter{
				ProjectID: projectID,
				StartTime: time.Now().Add(-24 * time.Hour),
				EndTime:   time.Now(),
			},
			mockSetup:   func() {},
			expectedErr: nil,
			checkResult: func(t *testing.T, analytics *observability.ObservationAnalytics) {
				assert.NotNil(t, analytics)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockObsRepo := new(MockObservationRepository)
			mockTraceRepo := new(MockTraceRepository)
			mockPublisher := new(MockEventPublisher)

			tt.mockSetup()

			service := NewObservationService(mockObsRepo, mockTraceRepo, mockPublisher)

			result, err := service.GetObservationAnalytics(context.Background(), tt.filter)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}
