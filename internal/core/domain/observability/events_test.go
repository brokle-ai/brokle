package observability

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"brokle/pkg/ulid"
)

// ============================================================================
// HIGH-VALUE TESTS: Event Creation with Business Logic
// ============================================================================

// TestNewTraceCreatedEvent tests trace event creation with observation counting
func TestNewTraceCreatedEvent(t *testing.T) {
	projectID := ulid.New()
	userID := ulid.New()
	sessionID := ulid.New()

	tests := []struct {
		name           string
		trace          *Trace
		userID         *ulid.ULID
		validateFields func(*testing.T, *Event)
	}{
		{
			name: "trace with multiple observations",
			trace: &Trace{
				ID:              ulid.New(),
				ProjectID:       projectID,
				ExternalTraceID: "ext-trace-123",
				Name:            "Test Trace",
				SessionID:       &sessionID,
				UserID:          &userID,
				Tags:            map[string]interface{}{"env": "test"},
				Metadata:        map[string]interface{}{"version": "1.0"},
				Observations:    []Observation{{}, {}}, // 2 observations
			},
			userID: &userID,
			validateFields: func(t *testing.T, event *Event) {
				assert.Equal(t, EventTypeTraceCreated, event.Type)
				assert.Equal(t, projectID, event.ProjectID)
				assert.Equal(t, 2, event.Data["observation_count"])
				assert.NotNil(t, event.Correlation)
			},
		},
		{
			name: "trace without observations",
			trace: &Trace{
				ID:              ulid.New(),
				ProjectID:       projectID,
				ExternalTraceID: "ext-trace-456",
				Name:            "Minimal Trace",
			},
			userID: nil,
			validateFields: func(t *testing.T, event *Event) {
				assert.Equal(t, 0, event.Data["observation_count"])
				assert.Nil(t, event.UserID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := NewTraceCreatedEvent(tt.trace, tt.userID)

			assert.NotNil(t, event)
			assert.NotEqual(t, ulid.ULID{}, event.ID)
			assert.Equal(t, tt.trace.ID.String(), event.Subject)
			assert.False(t, event.Timestamp.IsZero())

			if tt.validateFields != nil {
				tt.validateFields(t, event)
			}
		})
	}
}

// TestNewObservationCompletedEvent tests observation completion with latency calculation
func TestNewObservationCompletedEvent(t *testing.T) {
	traceID := ulid.New()
	userID := ulid.New()
	provider := "anthropic"
	model := "claude-3"

	tests := []struct {
		name        string
		observation *Observation
		userID      *ulid.ULID
		hasLatency  bool
	}{
		{
			name: "completed with latency and metrics",
			observation: &Observation{
				ID:           ulid.New(),
				TraceID:      traceID,
				Type:         ObservationTypeLLM,
				Provider:     &provider,
				Model:        &model,
				StartTime:    time.Now().Add(-100 * time.Millisecond),
				EndTime:      func() *time.Time { t := time.Now(); return &t }(),
				TotalTokens:  1500,
				TotalCost:    func() *float64 { c := 0.05; return &c }(),
				QualityScore: func() *float64 { q := 0.95; return &q }(),
			},
			userID:     &userID,
			hasLatency: true,
		},
		{
			name: "completed without end time",
			observation: &Observation{
				ID:        ulid.New(),
				TraceID:   traceID,
				Type:      ObservationTypeSpan,
				StartTime: time.Now(),
				EndTime:   nil,
			},
			userID:     nil,
			hasLatency: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := NewObservationCompletedEvent(tt.observation, tt.userID)

			assert.NotNil(t, event)
			assert.Equal(t, EventTypeObservationCompleted, event.Type)

			if tt.hasLatency {
				assert.NotNil(t, event.Data["latency_ms"])
			}

			assert.Equal(t, tt.observation.TotalTokens, event.Data["total_tokens"])
			assert.Equal(t, tt.observation.TotalCost, event.Data["total_cost"])
		})
	}
}

// TestNewBatchIngestionCompletedEvent tests batch completion with success rate calculation
func TestNewBatchIngestionCompletedEvent(t *testing.T) {
	projectID := ulid.New()
	userID := ulid.New()

	tests := []struct {
		name   string
		result *BatchIngestResult
		userID *ulid.ULID
	}{
		{
			name: "partial success with errors",
			result: &BatchIngestResult{
				ProcessedCount: 95,
				FailedCount:    5,
				Errors:         []BatchIngestionError{{}, {}},
				Duration:       500 * time.Millisecond,
				JobID:          func() *string { s := "job-123"; return &s }(),
			},
			userID: &userID,
		},
		{
			name: "complete failure",
			result: &BatchIngestResult{
				ProcessedCount: 0,
				FailedCount:    100,
				Errors:         []BatchIngestionError{{}, {}, {}},
				Duration:       200 * time.Millisecond,
			},
			userID: nil,
		},
		{
			name: "perfect success",
			result: &BatchIngestResult{
				ProcessedCount: 100,
				FailedCount:    0,
				Errors:         []BatchIngestionError{},
				Duration:       300 * time.Millisecond,
			},
			userID: &userID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := NewBatchIngestionCompletedEvent(projectID, tt.result, tt.userID)

			assert.NotNil(t, event)
			assert.Equal(t, EventTypeBatchIngestionCompleted, event.Type)
			assert.Equal(t, tt.result.ProcessedCount, event.Data["processed_count"])
			assert.Equal(t, tt.result.FailedCount, event.Data["failed_count"])
			assert.Equal(t, tt.result.Duration, event.Data["duration"])
			assert.Equal(t, len(tt.result.Errors), event.Data["errors_count"])

			if tt.result.JobID != nil {
				assert.Equal(t, tt.result.JobID, event.Metadata["job_id"])
			}
		})
	}
}

// TestNewThresholdExceededEvent tests threshold exceeded calculation
func TestNewThresholdExceededEvent(t *testing.T) {
	projectID := ulid.New()

	tests := []struct {
		name           string
		metric         string
		threshold      string
		currentValue   float64
		thresholdValue float64
		metadata       map[string]any
	}{
		{
			name:           "latency threshold exceeded",
			metric:         "avg_latency_ms",
			threshold:      "p95_latency",
			currentValue:   5500,
			thresholdValue: 5000,
			metadata: map[string]any{
				"window": "5m",
			},
		},
		{
			name:           "cost threshold exceeded",
			metric:         "daily_cost",
			threshold:      "budget_limit",
			currentValue:   150.0,
			thresholdValue: 100.0,
			metadata:       nil,
		},
		{
			name:           "error rate threshold",
			metric:         "error_rate",
			threshold:      "max_error_rate",
			currentValue:   0.15,
			thresholdValue: 0.05,
			metadata: map[string]any{
				"interval": "1h",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := NewThresholdExceededEvent(
				projectID,
				tt.metric,
				tt.threshold,
				tt.currentValue,
				tt.thresholdValue,
				tt.metadata,
			)

			assert.NotNil(t, event)
			assert.Equal(t, EventTypeThresholdExceeded, event.Type)

			// Check calculated fields
			exceededBy := tt.currentValue - tt.thresholdValue
			exceededByPct := (exceededBy / tt.thresholdValue) * 100
			assert.Equal(t, exceededBy, event.Data["exceeded_by"])
			assert.Equal(t, exceededByPct, event.Data["exceeded_by_pct"])

			if tt.metadata != nil {
				assert.Equal(t, tt.metadata, event.Metadata)
			}
		})
	}
}
