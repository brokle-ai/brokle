// Shared types for the time range picker used on the overview and
// future dashboard pages. Matches the backend `time_range` enum on
// `GetOverviewInput` — see overview.go in the analytics domain. The
// `custom` sentinel routes from/to ISO strings instead of a preset.

export type RelativeTimeRange =
  | '15m'
  | '30m'
  | '1h'
  | '3h'
  | '6h'
  | '12h'
  | '24h'
  | '7d'
  | '14d'
  | '30d'
  | 'all'
  | 'custom'

export interface TimeRange {
  from?: string
  to?: string
  relative?: RelativeTimeRange
}

export interface TimeRangePickerProps {
  value: TimeRange
  onChange: (range: TimeRange) => void
  className?: string
}

export interface RelativeOption {
  value: RelativeTimeRange
  label: string
  duration: number
}

export const RELATIVE_OPTIONS: RelativeOption[] = [
  { value: '15m', label: 'Last 15 minutes', duration: 1000 * 60 * 15 },
  { value: '30m', label: 'Last 30 minutes', duration: 1000 * 60 * 30 },
  { value: '1h', label: 'Last 1 hour', duration: 1000 * 60 * 60 },
  { value: '3h', label: 'Last 3 hours', duration: 1000 * 60 * 60 * 3 },
  { value: '6h', label: 'Last 6 hours', duration: 1000 * 60 * 60 * 6 },
  { value: '12h', label: 'Last 12 hours', duration: 1000 * 60 * 60 * 12 },
  { value: '24h', label: 'Last 24 hours', duration: 1000 * 60 * 60 * 24 },
  { value: '7d', label: 'Last 7 days', duration: 1000 * 60 * 60 * 24 * 7 },
  { value: '14d', label: 'Last 14 days', duration: 1000 * 60 * 60 * 24 * 14 },
  { value: '30d', label: 'Last 30 days', duration: 1000 * 60 * 60 * 24 * 30 },
  { value: 'all', label: 'All time', duration: 1000 * 60 * 60 * 24 * 365 },
]
