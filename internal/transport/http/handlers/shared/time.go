package shared

import (
	"time"

	"brokle/internal/core/domain/analytics"
	appErrors "brokle/pkg/errors"
)

// ParseTimeRange parses time range from HTTP request parameters.
//
// Parameters:
//   - from, to: Custom range in RFC3339 format (both required if either provided)
//   - timeRangeStr: Preset like "24h", "7d", "30d"
//   - defaultRange: Fallback when no parameters provided
//
// Returns (from, to, error) - error is AppError ready for response.Error()
func ParseTimeRange(from, to, timeRangeStr string, defaultRange analytics.TimeRange) (time.Time, time.Time, error) {
	var fromTime, toTime time.Time

	if from != "" && to != "" {
		// Custom date range provided
		var parseErr error
		fromTime, parseErr = time.Parse(time.RFC3339, from)
		if parseErr != nil {
			return time.Time{}, time.Time{}, appErrors.NewValidationError(
				"Invalid 'from' date format",
				"from must be in ISO 8601 format (e.g., 2024-01-01T00:00:00Z)",
			)
		}
		toTime, parseErr = time.Parse(time.RFC3339, to)
		if parseErr != nil {
			return time.Time{}, time.Time{}, appErrors.NewValidationError(
				"Invalid 'to' date format",
				"to must be in ISO 8601 format (e.g., 2024-01-02T00:00:00Z)",
			)
		}

		// Validate range order
		if toTime.Before(fromTime) {
			return time.Time{}, time.Time{}, appErrors.NewValidationError(
				"Invalid date range",
				"'to' must be after 'from'",
			)
		}

		return fromTime.UTC(), toTime.UTC(), nil
	}

	if from != "" || to != "" {
		// Incomplete date range
		return time.Time{}, time.Time{}, appErrors.NewValidationError(
			"Incomplete date range",
			"both 'from' and 'to' are required for custom date range",
		)
	}

	// Use predefined time range
	timeRange := analytics.ParseTimeRange(timeRangeStr)
	if timeRangeStr == "" {
		timeRange = defaultRange
	}
	toTime = time.Now().UTC()
	fromTime = toTime.Add(-timeRange.Duration())

	return fromTime, toTime, nil
}
