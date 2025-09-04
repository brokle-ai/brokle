package utils

import (
	"fmt"
	"strconv"
	"time"
)

// Common time formats
const (
	RFC3339Milli = "2006-01-02T15:04:05.000Z07:00"
	ISO8601      = "2006-01-02T15:04:05Z"
	DateOnly     = "2006-01-02"
	TimeOnly     = "15:04:05"
	DateTime     = "2006-01-02 15:04:05"
	Timestamp    = "20060102150405"
)

// TimeRange represents a time range with start and end times
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// Duration represents common durations
const (
	Minute = time.Minute
	Hour   = time.Hour
	Day    = 24 * time.Hour
	Week   = 7 * Day
	Month  = 30 * Day
	Year   = 365 * Day
)

// UTC returns the current time in UTC
func UTC() time.Time {
	return time.Now().UTC()
}

// UnixMilli returns current Unix timestamp in milliseconds
func UnixMilli() int64 {
	return time.Now().UnixMilli()
}

// UnixNano returns current Unix timestamp in nanoseconds
func UnixNano() int64 {
	return time.Now().UnixNano()
}

// FormatRFC3339 formats time in RFC3339 format
func FormatRFC3339(t time.Time) string {
	return t.Format(time.RFC3339)
}

// FormatRFC3339Milli formats time in RFC3339 format with milliseconds
func FormatRFC3339Milli(t time.Time) string {
	return t.Format(RFC3339Milli)
}

// FormatISO8601 formats time in ISO8601 format
func FormatISO8601(t time.Time) string {
	return t.Format(ISO8601)
}

// ParseFlexible attempts to parse time from various common formats
func ParseFlexible(timeStr string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		RFC3339Milli,
		ISO8601,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"15:04:05",
		"2006-01-02T15:04:05.000000Z",
		"2006-01-02T15:04:05.000Z",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s", timeStr)
}

// StartOfDay returns the start of the day (00:00:00) for the given time
func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// EndOfDay returns the end of the day (23:59:59.999999999) for the given time
func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// StartOfWeek returns the start of the week (Monday 00:00:00) for the given time
func StartOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday = 7
	}
	days := weekday - 1 // Days since Monday
	return StartOfDay(t.AddDate(0, 0, -days))
}

// EndOfWeek returns the end of the week (Sunday 23:59:59.999999999) for the given time
func EndOfWeek(t time.Time) time.Time {
	return EndOfDay(StartOfWeek(t).AddDate(0, 0, 6))
}

// StartOfMonth returns the start of the month (1st day 00:00:00) for the given time
func StartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// EndOfMonth returns the end of the month (last day 23:59:59.999999999) for the given time
func EndOfMonth(t time.Time) time.Time {
	return EndOfDay(StartOfMonth(t).AddDate(0, 1, -1))
}

// StartOfYear returns the start of the year (Jan 1st 00:00:00) for the given time
func StartOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
}

// EndOfYear returns the end of the year (Dec 31st 23:59:59.999999999) for the given time
func EndOfYear(t time.Time) time.Time {
	return EndOfDay(time.Date(t.Year(), 12, 31, 0, 0, 0, 0, t.Location()))
}

// DaysAgo returns time n days ago from now
func DaysAgo(days int) time.Time {
	return time.Now().AddDate(0, 0, -days)
}

// DaysFromNow returns time n days from now
func DaysFromNow(days int) time.Time {
	return time.Now().AddDate(0, 0, days)
}

// HoursAgo returns time n hours ago from now
func HoursAgo(hours int) time.Time {
	return time.Now().Add(-time.Duration(hours) * time.Hour)
}

// HoursFromNow returns time n hours from now
func HoursFromNow(hours int) time.Time {
	return time.Now().Add(time.Duration(hours) * time.Hour)
}

// MinutesAgo returns time n minutes ago from now
func MinutesAgo(minutes int) time.Time {
	return time.Now().Add(-time.Duration(minutes) * time.Minute)
}

// MinutesFromNow returns time n minutes from now
func MinutesFromNow(minutes int) time.Time {
	return time.Now().Add(time.Duration(minutes) * time.Minute)
}

// IsToday checks if the given time is today
func IsToday(t time.Time) bool {
	now := time.Now()
	return StartOfDay(t).Equal(StartOfDay(now))
}

// IsYesterday checks if the given time is yesterday
func IsYesterday(t time.Time) bool {
	yesterday := time.Now().AddDate(0, 0, -1)
	return StartOfDay(t).Equal(StartOfDay(yesterday))
}

// IsTomorrow checks if the given time is tomorrow
func IsTomorrow(t time.Time) bool {
	tomorrow := time.Now().AddDate(0, 0, 1)
	return StartOfDay(t).Equal(StartOfDay(tomorrow))
}

// IsWeekend checks if the given time falls on a weekend
func IsWeekend(t time.Time) bool {
	weekday := t.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

// IsWeekday checks if the given time falls on a weekday
func IsWeekday(t time.Time) bool {
	return !IsWeekend(t)
}

// DurationHuman returns a human-readable duration string
func DurationHuman(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d seconds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	}
	if d < Day {
		return fmt.Sprintf("%d hours", int(d.Hours()))
	}
	days := int(d.Hours() / 24)
	if days == 1 {
		return "1 day"
	}
	return fmt.Sprintf("%d days", days)
}

// TimeAgo returns a human-readable string representing how long ago the time was
func TimeAgo(t time.Time) string {
	now := time.Now()
	if t.After(now) {
		return "in the future"
	}

	diff := now.Sub(t)
	return DurationHuman(diff) + " ago"
}

// TimeUntil returns a human-readable string representing how long until the time
func TimeUntil(t time.Time) string {
	now := time.Now()
	if t.Before(now) {
		return "in the past"
	}

	diff := t.Sub(now)
	return "in " + DurationHuman(diff)
}

// ParseUnixTimestamp parses Unix timestamp (seconds, milliseconds, or nanoseconds)
func ParseUnixTimestamp(timestamp string) (time.Time, error) {
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timestamp: %w", err)
	}

	// Determine if it's seconds, milliseconds, or nanoseconds based on length
	switch len(timestamp) {
	case 10: // seconds
		return time.Unix(ts, 0), nil
	case 13: // milliseconds
		return time.Unix(ts/1000, (ts%1000)*int64(time.Millisecond)), nil
	case 19: // nanoseconds
		return time.Unix(0, ts), nil
	default:
		// Try to guess based on reasonable ranges
		if ts < 1e10 {
			return time.Unix(ts, 0), nil // seconds
		} else if ts < 1e13 {
			return time.Unix(ts/1000, (ts%1000)*int64(time.Millisecond)), nil // milliseconds
		} else {
			return time.Unix(0, ts), nil // nanoseconds
		}
	}
}

// GetTimeRange creates a time range for common periods
func GetTimeRange(period string) (TimeRange, error) {
	now := time.Now()
	
	switch period {
	case "today":
		return TimeRange{
			Start: StartOfDay(now),
			End:   EndOfDay(now),
		}, nil
	case "yesterday":
		yesterday := now.AddDate(0, 0, -1)
		return TimeRange{
			Start: StartOfDay(yesterday),
			End:   EndOfDay(yesterday),
		}, nil
	case "this_week":
		return TimeRange{
			Start: StartOfWeek(now),
			End:   EndOfWeek(now),
		}, nil
	case "last_week":
		lastWeek := now.AddDate(0, 0, -7)
		return TimeRange{
			Start: StartOfWeek(lastWeek),
			End:   EndOfWeek(lastWeek),
		}, nil
	case "this_month":
		return TimeRange{
			Start: StartOfMonth(now),
			End:   EndOfMonth(now),
		}, nil
	case "last_month":
		lastMonth := now.AddDate(0, -1, 0)
		return TimeRange{
			Start: StartOfMonth(lastMonth),
			End:   EndOfMonth(lastMonth),
		}, nil
	case "this_year":
		return TimeRange{
			Start: StartOfYear(now),
			End:   EndOfYear(now),
		}, nil
	case "last_year":
		lastYear := now.AddDate(-1, 0, 0)
		return TimeRange{
			Start: StartOfYear(lastYear),
			End:   EndOfYear(lastYear),
		}, nil
	case "last_7_days":
		return TimeRange{
			Start: StartOfDay(now.AddDate(0, 0, -6)),
			End:   EndOfDay(now),
		}, nil
	case "last_30_days":
		return TimeRange{
			Start: StartOfDay(now.AddDate(0, 0, -29)),
			End:   EndOfDay(now),
		}, nil
	case "last_90_days":
		return TimeRange{
			Start: StartOfDay(now.AddDate(0, 0, -89)),
			End:   EndOfDay(now),
		}, nil
	default:
		return TimeRange{}, fmt.Errorf("unknown time period: %s", period)
	}
}

// Contains checks if a time falls within the time range
func (tr TimeRange) Contains(t time.Time) bool {
	return t.After(tr.Start) && t.Before(tr.End) || t.Equal(tr.Start) || t.Equal(tr.End)
}

// Duration returns the duration of the time range
func (tr TimeRange) Duration() time.Duration {
	return tr.End.Sub(tr.Start)
}