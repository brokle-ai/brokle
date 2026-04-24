import { format, formatDistanceToNow } from 'date-fns'

/**
 * Safely format a date with fallback to '-' for invalid dates
 */
export function safeFormat(
  date: Date | string | undefined | null,
  formatStr: string,
): string {
  if (!date) return '-'
  const d = date instanceof Date ? date : new Date(date)
  if (isNaN(d.getTime())) return '-'
  return format(d, formatStr)
}

/**
 * Safely format relative time distance with fallback to '-' for invalid dates
 */
export function safeFormatDistance(
  date: Date | string | undefined | null,
): string {
  if (!date) return '-'
  const d = date instanceof Date ? date : new Date(date)
  if (isNaN(d.getTime())) return '-'
  return formatDistanceToNow(d, { addSuffix: true })
}

/**
 * Format duration from nanoseconds to adaptive human-readable string.
 * Uses industry-standard adaptive formatting (Datadog, Jaeger, Google Cloud Trace).
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
 * Format cost value to currency string.
 * Handles both string (from ClickHouse/Decimal) and number types.
 */
export function formatCost(cost: number | string | undefined | null): string {
  if (cost === undefined || cost === null || cost === '') return '-'
  const numCost = typeof cost === 'string' ? parseFloat(cost) : cost
  if (!Number.isFinite(numCost)) return '-'
  if (numCost === 0) return '$0.00'
  if (numCost < 0.01) return `$${numCost.toFixed(6)}`
  return `$${numCost.toFixed(4)}`
}

/**
 * Format cost with more precision (for detailed views).
 */
export function formatCostDetailed(
  cost: number | string | undefined | null,
): string {
  if (cost === undefined || cost === null || cost === '') return '-'
  const numCost = typeof cost === 'string' ? parseFloat(cost) : cost
  if (!Number.isFinite(numCost)) return '-'
  return `$${numCost.toFixed(6)}`
}

/**
 * Format a token count with locale-aware thousands separators.
 */
export function formatTokens(tokens: number | undefined | null): string {
  if (tokens === undefined || tokens === null || !Number.isFinite(tokens))
    return '-'
  return tokens.toLocaleString()
}
