package observability

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// HIGH-VALUE TESTS: Business Logic Calculations
// ============================================================================

// TestObservation_CalculateLatency tests the latency calculation logic
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

// TestTelemetryBatch_CalculateSuccessRate tests batch success rate calculation
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

// TestTelemetryBatch_CalculateProcessingTime tests batch processing time calculation
func TestTelemetryBatch_CalculateProcessingTime(t *testing.T) {
	createdAt := time.Now()
	completedAt := createdAt.Add(500 * time.Millisecond)

	tests := []struct {
		name     string
		batch    *TelemetryBatch
		expected *int
	}{
		{
			name: "with completed time",
			batch: &TelemetryBatch{
				CreatedAt:   createdAt,
				CompletedAt: &completedAt,
			},
			expected: func() *int { val := 500; return &val }(),
		},
		{
			name: "without completed time",
			batch: &TelemetryBatch{
				CreatedAt:   createdAt,
				CompletedAt: nil,
			},
			expected: nil,
		},
		{
			name: "with zero created time",
			batch: &TelemetryBatch{
				CreatedAt:   time.Time{},
				CompletedAt: &completedAt,
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.batch.CalculateProcessingTime()
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, *tt.expected, *result)
			}
		})
	}
}

// TestTelemetryEvent_ShouldRetry tests complex retry decision logic
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

// TestTelemetryEventDeduplication_TimeUntilExpiry tests expiry time calculation
func TestTelemetryEventDeduplication_TimeUntilExpiry(t *testing.T) {
	now := time.Now()
	past := now.Add(-1 * time.Hour)
	future := now.Add(1 * time.Hour)

	tests := []struct {
		name           string
		dedup          *TelemetryEventDeduplication
		expectZero     bool
		expectPositive bool
	}{
		{
			name: "already expired",
			dedup: &TelemetryEventDeduplication{
				ExpiresAt: past,
			},
			expectZero:     true,
			expectPositive: false,
		},
		{
			name: "not expired",
			dedup: &TelemetryEventDeduplication{
				ExpiresAt: future,
			},
			expectZero:     false,
			expectPositive: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.dedup.TimeUntilExpiry()
			if tt.expectZero {
				assert.Equal(t, time.Duration(0), result)
			}
			if tt.expectPositive {
				assert.Greater(t, result, time.Duration(0))
			}
		})
	}
}

// TestTelemetryBatch_IsCompleted tests batch completion status logic
func TestTelemetryBatch_IsCompleted(t *testing.T) {
	tests := []struct {
		name     string
		status   BatchStatus
		expected bool
	}{
		{"completed status", BatchStatusCompleted, true},
		{"failed status", BatchStatusFailed, true},
		{"partial status", BatchStatusPartial, true},
		{"processing status", BatchStatusProcessing, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batch := &TelemetryBatch{Status: tt.status}
			assert.Equal(t, tt.expected, batch.IsCompleted())
		})
	}
}
