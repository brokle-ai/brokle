package observability

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// HIGH-VALUE TESTS: Business Logic Calculations
// ============================================================================

// TestObservation_CalculateLatencyMs tests the latency calculation logic
func TestObservation_CalculateLatencyMs(t *testing.T) {
	startTime := time.Now()
	endTime := startTime.Add(150 * time.Millisecond)

	tests := []struct {
		name     string
		obs      *Observation
		expected *uint32
	}{
		{
			name: "with valid end time",
			obs: &Observation{
				StartTime: startTime,
				EndTime:   &endTime,
			},
			expected: func() *uint32 { val := uint32(150); return &val }(),
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
			expected: func() *uint32 { val := uint32(endTime.Sub(time.Time{}).Milliseconds()); return &val }(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.obs.CalculateLatencyMs()
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, *tt.expected, *result)
			}
		})
	}
}

// Removed after refactor: CalculateSuccessRate method was removed from TelemetryBatch
// The success rate calculation is now done in services layer

// Removed after refactor: CalculateProcessingTime method was removed from TelemetryBatch
// Processing time is now calculated in the services layer using ProcessingTimeMs field

// Removed after refactor: ShouldRetry method was removed from TelemetryEvent
// Retry logic is now handled in the services layer

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
			assert.LessOrEqual(t, result, time.Duration(0))
		}
			if tt.expectPositive {
				assert.Greater(t, result, time.Duration(0))
			}
		})
	}
}

// Removed after refactor: IsCompleted method was removed from TelemetryBatch
// The BatchStatusPartial constant no longer exists. Completion status is checked via Status field directly.
