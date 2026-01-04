import { subHours } from 'date-fns'
import type { TimeRange } from './types'
import { RELATIVE_OPTIONS } from './types'

/**
 * Get the actual Date objects for a TimeRange
 * Converts relative presets to absolute dates based on current time
 */
export function getTimeRangeDates(range: TimeRange): { from: Date; to: Date } {
  const now = new Date()

  // Custom range with explicit from/to dates
  if (range.relative === 'custom' && range.from && range.to) {
    return {
      from: new Date(range.from),
      to: new Date(range.to),
    }
  }

  // Find the matching relative option
  const option = RELATIVE_OPTIONS.find((o) => o.value === range.relative)
  if (option) {
    return {
      from: new Date(now.getTime() - option.duration),
      to: now,
    }
  }

  // Default to 24 hours
  return {
    from: subHours(now, 24),
    to: now,
  }
}

/**
 * Get browser timezone abbreviation (e.g., "EST", "PST", "UTC")
 */
export function getTimezoneAbbr(): string {
  return (
    new Date()
      .toLocaleTimeString('en-US', { timeZoneName: 'short' })
      .split(' ')
      .pop() || 'UTC'
  )
}

/**
 * Format a TimeRange for display
 */
export function formatTimeRangeLabel(range: TimeRange): string {
  if (range.relative && range.relative !== 'custom') {
    const option = RELATIVE_OPTIONS.find((o) => o.value === range.relative)
    return option?.label || range.relative
  }

  if (range.relative === 'custom' && range.from && range.to) {
    const from = new Date(range.from)
    const to = new Date(range.to)
    const formatOpts: Intl.DateTimeFormatOptions = {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    }
    return `${from.toLocaleDateString('en-US', formatOpts)} - ${to.toLocaleDateString('en-US', formatOpts)}`
  }

  return 'Select time range'
}

/**
 * Convert TimeRange to API query parameters
 */
export function timeRangeToApiParams(
  range: TimeRange
): { time_range?: string; from?: string; to?: string } {
  if (range.relative === 'custom' && range.from && range.to) {
    return {
      from: range.from,
      to: range.to,
    }
  }

  if (range.relative && range.relative !== 'custom') {
    return {
      time_range: range.relative,
    }
  }

  // Default to 24h
  return {
    time_range: '24h',
  }
}
