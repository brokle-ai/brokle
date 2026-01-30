import { format, formatDistanceToNow } from 'date-fns'

/**
 * Safely format a date with fallback to '-' for invalid dates
 *
 * @param date - The date to format (can be Date, null, or undefined)
 * @param formatStr - The format string (e.g., 'PPpp', 'yyyy-MM-dd')
 * @returns Formatted date string or '-' if invalid
 */
export function safeFormat(
  date: Date | undefined | null,
  formatStr: string
): string {
  if (!date || !(date instanceof Date) || isNaN(date.getTime())) {
    return '-'
  }
  return format(date, formatStr)
}

/**
 * Safely format relative time distance with fallback to '-' for invalid dates
 *
 * @param date - The date to format (can be Date, null, or undefined)
 * @returns Relative time string (e.g., '2 hours ago') or '-' if invalid
 */
export function safeFormatDistance(
  date: Date | undefined | null
): string {
  if (!date || !(date instanceof Date) || isNaN(date.getTime())) {
    return '-'
  }
  return formatDistanceToNow(date, { addSuffix: true })
}

/**
 * Safely format duration in milliseconds to human-readable string
 *
 * @param durationMs - Duration in milliseconds
 * @returns Formatted duration string (e.g., '1.23s', '123ms') or '-' if invalid
 */
export function safeFormatDuration(
  durationMs: number | undefined | null
): string {
  if (durationMs === undefined || durationMs === null || isNaN(durationMs)) {
    return '-'
  }

  if (durationMs < 1000) {
    return `${durationMs.toFixed(0)}ms`
  }

  return `${(durationMs / 1000).toFixed(2)}s`
}

/**
 * Format duration from nanoseconds to adaptive human-readable string
 * Uses industry-standard adaptive formatting (Datadog, Jaeger, Google Cloud Trace)
 *
 * @param nanos - Duration in nanoseconds
 * @returns Formatted duration string (e.g., '500ns', '250µs', '45.3ms', '2.50s')
 */
export function formatDuration(nanos: number | undefined | null): string {
  if (nanos == null) return '-'

  const ms = nanos / 1_000_000
  const us = nanos / 1_000

  if (nanos < 1_000) return `${nanos}ns`
  if (nanos < 1_000_000) return `${Math.round(us)}µs`
  if (ms < 100) return `${ms.toFixed(1)}ms`
  if (ms < 1000) return `${Math.round(ms)}ms`
  if (ms < 10000) return `${(ms / 1000).toFixed(2)}s`
  return `${(ms / 1000).toFixed(1)}s`
}

/**
 * Format cost value to currency string
 * Handles both string (from ClickHouse Decimal) and number types
 *
 * @param cost - Cost value (can be string from DB or number)
 * @returns Formatted cost string (e.g., '$0.0012') or '-' if invalid
 */
export function formatCost(cost: number | string | undefined | null): string {
  if (cost === undefined || cost === null || cost === '') return '-'
  const numCost = typeof cost === 'string' ? parseFloat(cost) : cost
  if (isNaN(numCost)) return '-'
  return `$${numCost.toFixed(4)}`
}

/**
 * Format cost with more precision (for detailed views)
 * Handles both string (from ClickHouse Decimal) and number types
 *
 * @param cost - Cost value (can be string from DB or number)
 * @returns Formatted cost string with 6 decimal places
 */
export function formatCostDetailed(cost: number | string | undefined | null): string {
  if (cost === undefined || cost === null || cost === '') return '-'
  const numCost = typeof cost === 'string' ? parseFloat(cost) : cost
  if (isNaN(numCost)) return '-'
  return `$${numCost.toFixed(6)}`
}
