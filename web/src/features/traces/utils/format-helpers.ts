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
