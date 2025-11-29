package observability

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"brokle/internal/core/domain/observability"
)

// ============================================================================
// Mock Repositories - OTEL-Native
// ============================================================================

type MockTraceRepository struct {
	mock.Mock
}

func (m *MockTraceRepository) GetRootSpan(ctx context.Context, traceID string) (*observability.Span, error) {
	args := m.Called(ctx, traceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*observability.Span), args.Error(1)
}

func (m *MockTraceRepository) GetTraceMetrics(ctx context.Context, traceID string) (*observability.TraceMetrics, error) {
	args := m.Called(ctx, traceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*observability.TraceMetrics), args.Error(1)
}

func (m *MockTraceRepository) CalculateTotalCost(ctx context.Context, traceID string) (float64, error) {
	args := m.Called(ctx, traceID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockTraceRepository) CountSpans(ctx context.Context, traceID string) (int64, error) {
	args := m.Called(ctx, traceID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTraceRepository) ListTraces(ctx context.Context, filter *observability.TraceFilter) ([]*observability.TraceMetrics, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.TraceMetrics), args.Error(1)
}

func (m *MockTraceRepository) GetBySessionID(ctx context.Context, sessionID string) ([]*observability.TraceMetrics, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.TraceMetrics), args.Error(1)
}

func (m *MockTraceRepository) GetByUserID(ctx context.Context, userID string, filter *observability.TraceFilter) ([]*observability.TraceMetrics, error) {
	args := m.Called(ctx, userID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.TraceMetrics), args.Error(1)
}

func (m *MockTraceRepository) Count(ctx context.Context, filter *observability.TraceFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

type MockSpanRepository struct {
	mock.Mock
}

func (m *MockSpanRepository) Create(ctx context.Context, span *observability.Span) error {
	args := m.Called(ctx, span)
	return args.Error(0)
}

func (m *MockSpanRepository) CreateBatch(ctx context.Context, spans []*observability.Span) error {
	args := m.Called(ctx, spans)
	return args.Error(0)
}

func (m *MockSpanRepository) GetByID(ctx context.Context, id string) (*observability.Span, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*observability.Span), args.Error(1)
}

func (m *MockSpanRepository) GetByTraceID(ctx context.Context, traceID string) ([]*observability.Span, error) {
	args := m.Called(ctx, traceID)
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

func (m *MockSpanRepository) GetChildren(ctx context.Context, parentSpanID string) ([]*observability.Span, error) {
	args := m.Called(ctx, parentSpanID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.Span), args.Error(1)
}

func (m *MockSpanRepository) GetTreeByTraceID(ctx context.Context, traceID string) ([]*observability.Span, error) {
	args := m.Called(ctx, traceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.Span), args.Error(1)
}

func (m *MockSpanRepository) GetByFilter(ctx context.Context, filter *observability.SpanFilter) ([]*observability.Span, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.Span), args.Error(1)
}

func (m *MockSpanRepository) Update(ctx context.Context, span *observability.Span) error {
	args := m.Called(ctx, span)
	return args.Error(0)
}

func (m *MockSpanRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSpanRepository) Count(ctx context.Context, filter *observability.SpanFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

// ============================================================================
// OTEL-Native TraceService Tests
// ============================================================================

func TestTraceService_GetRootSpan(t *testing.T) {
	tests := []struct {
		name        string
		traceID     string
		mockSetup   func(*MockTraceRepository)
		expectedErr bool
	}{
		{
			name:    "success - valid trace_id",
			traceID: "12345678901234567890123456789012",
			mockSetup: func(repo *MockTraceRepository) {
				rootSpan := &observability.Span{
					SpanID:       "span123",
					TraceID:      "12345678901234567890123456789012",
					ParentSpanID: nil,
					SpanName:     "root-span",
					ProjectID:    "proj123",
				}
				repo.On("GetRootSpan", mock.Anything, "12345678901234567890123456789012").
					Return(rootSpan, nil)
			},
			expectedErr: false,
		},
		{
			name:        "error - invalid trace_id length",
			traceID:     "invalid",
			mockSetup:   func(repo *MockTraceRepository) {},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockTraceRepository)
			mockSpanRepo := new(MockSpanRepository)
			tt.mockSetup(mockRepo)

			service := NewTraceService(mockRepo, mockSpanRepo, logrus.New())
			result, err := service.GetRootSpan(context.Background(), tt.traceID)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestTraceService_GetTraceMetrics(t *testing.T) {
	tests := []struct {
		name        string
		traceID     string
		mockSetup   func(*MockTraceRepository)
		expectedErr bool
		checkResult func(*testing.T, *observability.TraceMetrics)
	}{
		{
			name:    "success - valid aggregation",
			traceID: "12345678901234567890123456789012",
			mockSetup: func(repo *MockTraceRepository) {
				metrics := &observability.TraceMetrics{
					TraceID:      "12345678901234567890123456789012",
					RootSpanID:   "root123",
					ProjectID:    "proj123",
					TotalCost:    decimal.NewFromFloat(0.05),
					TotalTokens:  1000,
					SpanCount:    5,
					HasError:     false,
				}
				repo.On("GetTraceMetrics", mock.Anything, "12345678901234567890123456789012").
					Return(metrics, nil)
			},
			expectedErr: false,
			checkResult: func(t *testing.T, metrics *observability.TraceMetrics) {
				assert.Equal(t, int64(5), metrics.SpanCount)
				assert.Equal(t, uint64(1000), metrics.TotalTokens)
				assert.True(t, metrics.TotalCost.GreaterThan(decimal.Zero))
			},
		},
		{
			name:        "error - invalid trace_id",
			traceID:     "short",
			mockSetup:   func(repo *MockTraceRepository) {},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockTraceRepository)
			mockSpanRepo := new(MockSpanRepository)
			tt.mockSetup(mockRepo)

			service := NewTraceService(mockRepo, mockSpanRepo, logrus.New())
			result, err := service.GetTraceMetrics(context.Background(), tt.traceID)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.checkResult != nil {
					tt.checkResult(t, result)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestTraceService_ListTraces(t *testing.T) {
	projectID := "proj123"
	now := time.Now()

	tests := []struct {
		name        string
		filter      *observability.TraceFilter
		mockSetup   func(*MockTraceRepository)
		expectedErr bool
		checkResult func(*testing.T, []*observability.TraceMetrics)
	}{
		{
			name: "success - list traces with filter",
			filter: &observability.TraceFilter{
				ProjectID: projectID,
			},
			mockSetup: func(repo *MockTraceRepository) {
				traces := []*observability.TraceMetrics{
					{
						TraceID:    "trace1",
						ProjectID:  projectID,
						SpanCount:  3,
						StartTime:  now,
						EndTime:    now.Add(time.Second),
					},
					{
						TraceID:    "trace2",
						ProjectID:  projectID,
						SpanCount:  5,
						StartTime:  now.Add(-time.Hour),
						EndTime:    now.Add(-time.Hour).Add(2 * time.Second),
					},
				}
				repo.On("ListTraces", mock.Anything, mock.AnythingOfType("*observability.TraceFilter")).
					Return(traces, nil)
			},
			expectedErr: false,
			checkResult: func(t *testing.T, traces []*observability.TraceMetrics) {
				assert.Len(t, traces, 2)
				assert.Equal(t, int64(3), traces[0].SpanCount)
				assert.Equal(t, int64(5), traces[1].SpanCount)
			},
		},
		{
			name:        "error - nil filter",
			filter:      nil,
			mockSetup:   func(repo *MockTraceRepository) {},
			expectedErr: true,
		},
		{
			name: "error - empty project_id",
			filter: &observability.TraceFilter{
				ProjectID: "",
			},
			mockSetup:   func(repo *MockTraceRepository) {},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockTraceRepository)
			mockSpanRepo := new(MockSpanRepository)
			tt.mockSetup(mockRepo)

			service := NewTraceService(mockRepo, mockSpanRepo, logrus.New())
			result, err := service.ListTraces(context.Background(), tt.filter)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.checkResult != nil {
					tt.checkResult(t, result)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestTraceService_GetTraceWithAllSpans(t *testing.T) {
	tests := []struct {
		name        string
		traceID     string
		mockSetup   func(*MockSpanRepository)
		expectedErr bool
		checkResult func(*testing.T, []*observability.Span)
	}{
		{
			name:    "success - get all spans",
			traceID: "12345678901234567890123456789012",
			mockSetup: func(repo *MockSpanRepository) {
				spans := []*observability.Span{
					{SpanID: "span1", TraceID: "12345678901234567890123456789012"},
					{SpanID: "span2", TraceID: "12345678901234567890123456789012"},
					{SpanID: "span3", TraceID: "12345678901234567890123456789012"},
				}
				repo.On("GetByTraceID", mock.Anything, "12345678901234567890123456789012").
					Return(spans, nil)
			},
			expectedErr: false,
			checkResult: func(t *testing.T, spans []*observability.Span) {
				assert.Len(t, spans, 3)
			},
		},
		{
			name:        "error - invalid trace_id",
			traceID:     "invalid",
			mockSetup:   func(repo *MockSpanRepository) {},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTraceRepo := new(MockTraceRepository)
			mockSpanRepo := new(MockSpanRepository)
			tt.mockSetup(mockSpanRepo)

			service := NewTraceService(mockTraceRepo, mockSpanRepo, logrus.New())
			result, err := service.GetTraceWithAllSpans(context.Background(), tt.traceID)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.checkResult != nil {
					tt.checkResult(t, result)
				}
			}

			mockSpanRepo.AssertExpectations(t)
		})
	}
}

func TestTraceService_CalculateTraceCost(t *testing.T) {
	tests := []struct {
		name         string
		traceID      string
		mockSetup    func(*MockTraceRepository)
		expectedCost float64
		expectedErr  bool
	}{
		{
			name:    "success - calculate cost",
			traceID: "12345678901234567890123456789012",
			mockSetup: func(repo *MockTraceRepository) {
				repo.On("CalculateTotalCost", mock.Anything, "12345678901234567890123456789012").
					Return(0.05, nil)
			},
			expectedCost: 0.05,
			expectedErr:  false,
		},
		{
			name:        "error - invalid trace_id",
			traceID:     "invalid",
			mockSetup:   func(repo *MockTraceRepository) {},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockTraceRepository)
			mockSpanRepo := new(MockSpanRepository)
			tt.mockSetup(mockRepo)

			service := NewTraceService(mockRepo, mockSpanRepo, logrus.New())
			result, err := service.CalculateTraceCost(context.Background(), tt.traceID)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCost, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
