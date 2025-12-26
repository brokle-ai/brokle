package observability

import (
	"context"

	"brokle/internal/core/domain/observability"

	"github.com/stretchr/testify/mock"
)

// MockScoreRepository is a mock implementation of ScoreRepository
type MockScoreRepository struct {
	mock.Mock
}

func (m *MockScoreRepository) Create(ctx context.Context, score *observability.Score) error {
	args := m.Called(ctx, score)
	return args.Error(0)
}

func (m *MockScoreRepository) Update(ctx context.Context, score *observability.Score) error {
	args := m.Called(ctx, score)
	return args.Error(0)
}

func (m *MockScoreRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockScoreRepository) GetByID(ctx context.Context, id string) (*observability.Score, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*observability.Score), args.Error(1)
}

func (m *MockScoreRepository) GetByTraceID(ctx context.Context, traceID string) ([]*observability.Score, error) {
	args := m.Called(ctx, traceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.Score), args.Error(1)
}

func (m *MockScoreRepository) GetBySpanID(ctx context.Context, spanID string) ([]*observability.Score, error) {
	args := m.Called(ctx, spanID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.Score), args.Error(1)
}

func (m *MockScoreRepository) GetBySessionID(ctx context.Context, sessionID string) ([]*observability.Score, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.Score), args.Error(1)
}

func (m *MockScoreRepository) GetByFilter(ctx context.Context, filter *observability.ScoreFilter) ([]*observability.Score, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*observability.Score), args.Error(1)
}

func (m *MockScoreRepository) CreateBatch(ctx context.Context, scores []*observability.Score) error {
	args := m.Called(ctx, scores)
	return args.Error(0)
}

func (m *MockScoreRepository) Count(ctx context.Context, filter *observability.ScoreFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockScoreRepository) ExistsByConfigName(ctx context.Context, projectID, configName string) (bool, error) {
	args := m.Called(ctx, projectID, configName)
	return args.Bool(0), args.Error(1)
}
