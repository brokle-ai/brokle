# Testing Guide

This guide provides detailed testing patterns and examples for the Brokle platform. For AI-assisted test generation, see `prompts/testing.txt`.

## Testing Philosophy

**Core Principle: Test Business Logic, Not Framework Behavior**

The Brokle platform follows a pragmatic testing approach that prioritizes high-value business logic tests over low-value granular tests.

### What We Test ✅

1. **Complex Business Logic**
   - Calculations: latency, success rates, cost computations
   - Retry mechanisms with multiple conditions
   - State machines and workflow orchestration
   - Aggregations and analytics

2. **Batch Operations**
   - Bulk processing with validation loops
   - Partial success/failure handling
   - Transaction coordination
   - Concurrent operations

3. **Multi-Step Orchestration**
   - Operations with multiple repository calls
   - Cross-service workflows
   - Event publishing with side effects
   - Compensation patterns

4. **Error Handling Patterns**
   - Domain error mapping to AppErrors
   - Error wrapping and chaining
   - Retry logic with backoff
   - Graceful degradation

5. **Analytics & Aggregations**
   - Time-based queries
   - Statistical calculations
   - Trend analysis
   - Metric computations

### What We Don't Test ❌

1. **Simple CRUD Operations** - Basic Create/Read/Update/Delete with no business logic
2. **Field Validation** - Already tested in domain layer
3. **Trivial Constructors** - Simple object creation and field assignment
4. **Framework Behavior** - ULID generation, time.Now(), errors.Is (stdlib)
5. **Static Definitions** - Constant strings, enum type checkers

## Test Coverage Guidelines

### Target Metrics

- **Service Layer**: ~1:1 test-to-code ratio (focus on business logic)
- **Domain Layer**: Minimal (only complex calculations and business rules)
- **Handler Layer**: Critical workflows only (integration tests)

### Current Coverage (Observability Domain)

- **Service Tests**: 3,485 lines (0.96:1 ratio) ✅
- **Domain Tests**: 594 lines (business logic only) ✅
- **All tests passing** with healthy coverage

### Acceptable Ratios

- ✅ **0.8:1 to 1.2:1** - Healthy coverage of business logic
- ⚠️  **< 0.5:1** - Likely missing critical test coverage
- ⚠️  **> 2:1** - Likely testing too many trivial operations

## Service Layer Tests (Primary Focus)

Test complex business logic with mock repositories using table-driven patterns.

### Complete Example: CreateTraceWithObservations

```go
// ============================================================================
// Mock Repositories (Full Interface Implementation)
// ============================================================================

type MockTraceRepository struct {
    mock.Mock
}

func (m *MockTraceRepository) Create(ctx context.Context, trace *observability.Trace) error {
    args := m.Called(ctx, trace)
    return args.Error(0)
}

func (m *MockTraceRepository) CreateBatch(ctx context.Context, traces []*observability.Trace) error {
    args := m.Called(ctx, traces)
    return args.Error(0)
}

// Implement ALL interface methods (even if not used in test)

type MockObservationRepository struct {
    mock.Mock
}

func (m *MockObservationRepository) CreateBatch(ctx context.Context, observations []observability.Observation) error {
    args := m.Called(ctx, observations)
    return args.Error(0)
}

type MockEventPublisher struct {
    mock.Mock
}

func (m *MockEventPublisher) Publish(ctx context.Context, eventType observability.EventType) error {
    args := m.Called(ctx, eventType)
    return args.Error(0)
}

// ============================================================================
// HIGH-VALUE TESTS: Complex Business Logic
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
            name: "success - trace with multiple observations",
            trace: &observability.Trace{
                Name: "Test Trace",
                Observations: []observability.Observation{
                    {Name: "Obs 1", Type: observability.ObservationTypeLLM},
                    {Name: "Obs 2", Type: observability.ObservationTypeSpan},
                },
            },
            mockSetup: func(traceRepo *MockTraceRepository, obsRepo *MockObservationRepository, publisher *MockEventPublisher) {
                // Expect trace creation
                traceRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

                // Expect batch observation creation with validation
                obsRepo.On("CreateBatch", mock.Anything, mock.MatchedBy(func(obs []observability.Observation) bool {
                    return len(obs) == 2
                })).Return(nil)

                // Expect event publishing
                publisher.On("Publish", mock.Anything, observability.EventTypeTraceCreated).Return(nil)
            },
            expectedErr: nil,
            checkResult: func(t *testing.T, trace *observability.Trace) {
                assert.NotNil(t, trace)
                assert.NotEqual(t, ulid.ULID{}, trace.ID)
                assert.Len(t, trace.Observations, 2)
                assert.NotNil(t, trace.CreatedAt)
            },
        },
        {
            name: "error - validation failure",
            trace: &observability.Trace{
                Name: "", // Invalid - empty name
            },
            mockSetup: func(traceRepo *MockTraceRepository, obsRepo *MockObservationRepository, publisher *MockEventPublisher) {
                // No calls expected - validation should fail before repository
            },
            expectedErr: appErrors.ErrValidationFailed,
            checkResult: nil,
        },
        {
            name: "error - repository failure",
            trace: &observability.Trace{
                Name: "Test Trace",
            },
            mockSetup: func(traceRepo *MockTraceRepository, obsRepo *MockObservationRepository, publisher *MockEventPublisher) {
                traceRepo.On("Create", mock.Anything, mock.Anything).
                    Return(fmt.Errorf("database error"))
            },
            expectedErr: appErrors.ErrInternalError,
            checkResult: nil,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup mocks
            mockTraceRepo := new(MockTraceRepository)
            mockObsRepo := new(MockObservationRepository)
            mockPublisher := new(MockEventPublisher)
            tt.mockSetup(mockTraceRepo, mockObsRepo, mockPublisher)

            // Create service
            service := NewTraceService(mockTraceRepo, mockObsRepo, mockPublisher)

            // Execute
            result, err := service.CreateTraceWithObservations(context.Background(), tt.trace)

            // Assert errors
            if tt.expectedErr != nil {
                assert.Error(t, err)
                assert.ErrorIs(t, err, tt.expectedErr)
            } else {
                assert.NoError(t, err)
            }

            // Assert results
            if tt.checkResult != nil {
                tt.checkResult(t, result)
            }

            // Verify all mock expectations were met
            mockTraceRepo.AssertExpectations(t)
            mockObsRepo.AssertExpectations(t)
            mockPublisher.AssertExpectations(t)
        })
    }
}
```

### Key Patterns

1. **Complete Mock Interface Implementation**
   - Implement ALL methods from the interface
   - Use `mock.Called()` with proper return values
   - Even implement unused methods for completeness

2. **Table-Driven Tests**
   - Use `tests := []struct` pattern
   - Include name, input, mockSetup, expectedErr, checkResult
   - Test success, validation errors, and repository errors

3. **Mock Setup Functions**
   - Accept all mock repositories as parameters
   - Use `mock.MatchedBy()` for complex validation
   - Set up expected calls in logical order

4. **Result Validation Functions**
   - Accept testing.T and result as parameters
   - Verify business logic was applied correctly
   - Check calculated fields and side effects

5. **Mock Expectation Verification**
   - Always call `AssertExpectations(t)` on all mocks
   - Ensures all expected calls were made
   - Catches missing or extra repository calls

## Domain Layer Tests (Minimal)

Only test complex business logic calculations, not validation or constructors.

### Example: Business Logic Calculations

```go
// ============================================================================
// HIGH-VALUE TESTS: Business Logic Calculations
// ============================================================================

func TestObservation_CalculateLatency(t *testing.T) {
    startTime := time.Now()
    endTime := startTime.Add(150 * time.Millisecond)

    tests := []struct {
        name     string
        obs      *Observation
        expected *int
    }{
        {
            name: "with valid end time",
            obs: &Observation{
                StartTime: startTime,
                EndTime:   &endTime,
            },
            expected: func() *int { val := 150; return &val }(),
        },
        {
            name: "without end time",
            obs: &Observation{
                StartTime: startTime,
                EndTime:   nil,
            },
            expected: nil,
        },
        {
            name: "with zero start time",
            obs: &Observation{
                StartTime: time.Time{},
                EndTime:   &endTime,
            },
            expected: nil,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := tt.obs.CalculateLatency()
            if tt.expected == nil {
                assert.Nil(t, result)
            } else {
                assert.NotNil(t, result)
                assert.Equal(t, *tt.expected, *result)
            }
        })
    }
}

func TestTelemetryEvent_ShouldRetry(t *testing.T) {
    tests := []struct {
        name       string
        event      *TelemetryEvent
        maxRetries int
        expected   bool
    }{
        {
            name: "should retry - under max retries with error",
            event: &TelemetryEvent{
                RetryCount:   2,
                ErrorMessage: func() *string { s := "error"; return &s }(),
                ProcessedAt:  nil,
            },
            maxRetries: 3,
            expected:   true,
        },
        {
            name: "should not retry - at max retries",
            event: &TelemetryEvent{
                RetryCount:   3,
                ErrorMessage: func() *string { s := "error"; return &s }(),
                ProcessedAt:  nil,
            },
            maxRetries: 3,
            expected:   false,
        },
        {
            name: "should not retry - no error",
            event: &TelemetryEvent{
                RetryCount:   1,
                ErrorMessage: nil,
                ProcessedAt:  nil,
            },
            maxRetries: 3,
            expected:   false,
        },
        {
            name: "should not retry - already processed",
            event: &TelemetryEvent{
                RetryCount:   1,
                ErrorMessage: func() *string { s := "error"; return &s }(),
                ProcessedAt:  func() *time.Time { t := time.Now(); return &t }(),
            },
            maxRetries: 3,
            expected:   false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := tt.event.ShouldRetry(tt.maxRetries)
            assert.Equal(t, tt.expected, result)
        })
    }
}

func TestTelemetryBatch_CalculateSuccessRate(t *testing.T) {
    tests := []struct {
        name            string
        totalEvents     int
        processedEvents int
        expected        float64
    }{
        {"100% success", 100, 100, 100.0},
        {"95% success", 100, 95, 95.0},
        {"50% success", 100, 50, 50.0},
        {"0% success", 100, 0, 0.0},
        {"zero total events", 0, 0, 0.0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            batch := &TelemetryBatch{
                TotalEvents:     tt.totalEvents,
                ProcessedEvents: tt.processedEvents,
            }
            assert.Equal(t, tt.expected, batch.CalculateSuccessRate())
        })
    }
}
```

## Integration Tests (Critical Workflows)

Test complete workflows with real databases.

### Example: Full Trace Lifecycle

```go
func TestTraceService_Integration(t *testing.T) {
    // Skip if not running integration tests
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    // Setup test database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)

    // Setup real dependencies
    traceRepo := repository.NewTraceRepository(db)
    obsRepo := repository.NewObservationRepository(db)
    publisher := events.NewEventPublisher()
    service := services.NewTraceService(traceRepo, obsRepo, publisher)

    // Test complete workflow
    trace := &observability.Trace{
        Name:     "Integration Test Trace",
        UserID:   stringPtr("user_123"),
        Tags:     map[string]interface{}{"env": "test"},
        Metadata: map[string]interface{}{"version": "1.0"},
        Observations: []observability.Observation{
            {
                Name:      "Test Observation",
                Type:      observability.ObservationTypeLLM,
                StartTime: time.Now(),
            },
        },
    }

    // Create trace with observations
    result, err := service.CreateTraceWithObservations(context.Background(), trace)
    require.NoError(t, err)
    require.NotEqual(t, ulid.ULID{}, result.ID)

    // Verify data was actually persisted
    retrieved, err := service.GetTraceWithObservations(context.Background(), result.ID)
    require.NoError(t, err)
    assert.Equal(t, result.ID, retrieved.ID)
    assert.Equal(t, "Integration Test Trace", retrieved.Name)
    assert.Len(t, retrieved.Observations, 1)
    assert.Equal(t, "Test Observation", retrieved.Observations[0].Name)

    // Test update workflow
    retrieved.Tags["updated"] = true
    err = service.UpdateTrace(context.Background(), retrieved)
    require.NoError(t, err)

    // Verify update was persisted
    updated, err := service.GetTrace(context.Background(), result.ID)
    require.NoError(t, err)
    assert.Equal(t, true, updated.Tags["updated"])
}

// Helper functions
func setupTestDB(t *testing.T) *gorm.DB {
    // Setup PostgreSQL test database
    db, err := gorm.Open(postgres.Open(os.Getenv("TEST_DATABASE_URL")), &gorm.Config{})
    require.NoError(t, err)

    // Run migrations
    err = db.AutoMigrate(&observability.Trace{}, &observability.Observation{})
    require.NoError(t, err)

    return db
}

func cleanupTestDB(t *testing.T, db *gorm.DB) {
    // Clean up test data
    db.Exec("TRUNCATE TABLE traces CASCADE")
    db.Exec("TRUNCATE TABLE observations CASCADE")
}
```

## Running Tests

```bash
# Run all tests
make test

# Run unit tests only (excludes integration)
make test-unit

# Run integration tests with real databases
make test-integration

# Run with coverage report
make test-coverage

# Run specific package tests
go test ./internal/services/observability -v

# Run specific test
go test ./internal/services/observability -run TestTraceService_CreateTraceWithObservations -v

# Run tests with race detection
go test -race ./...

# Run integration tests only
go test -short=false ./...

# Skip integration tests
go test -short ./...
```

## Test Quality Checklist

Before committing tests:

### Must Have ✅
- Uses table-driven test pattern
- Mocks implement full repository interfaces
- Focuses on business logic, not framework behavior
- Verifies mock expectations with `AssertExpectations()`
- Tests batch operations if service has batch methods
- Tests error handling (domain → AppError mapping)
- Maintains healthy test-to-code ratio (~1:1)

### Must Not Have ❌
- No tests for simple CRUD operations
- No tests for validation rules (already in domain)
- No tests for trivial constructors
- No tests for framework behavior (ULID, time, errors)

## Reference Implementation

See the observability domain for complete examples:

- **Service Tests**: `internal/services/observability/*_test.go` (3,485 lines)
  - `trace_service_test.go` (831 lines)
  - `observation_service_test.go` (442 lines)
  - `telemetry_event_service_test.go` (848 lines)
  - `telemetry_batch_service_test.go` (670 lines)
  - `quality_service_test.go` (516 lines)

- **Domain Tests**: `internal/core/domain/observability/*_test.go` (594 lines)
  - `entity_test.go` (258 lines) - Business logic calculations
  - `errors_test.go` (66 lines) - Error wrapping/chaining
  - `events_test.go` (270 lines) - Event creation with calculations

These tests demonstrate the complete testing philosophy with real-world examples.

## AI-Assisted Test Generation

For detailed patterns and AI prompt guidance, see:
- **`prompts/testing.txt`** - Complete AI testing prompt (900+ lines)
- Covers all testing patterns in detail
- Includes before/after refactoring examples
- Provides comprehensive checklists and workflows
