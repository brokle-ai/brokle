import type { Span, TraceDetail } from '../api/types'

/**
 * Format duration from nanoseconds to human-readable string. Local
 * helper kept independent from `format-helpers.formatDuration` because
 * the export format (e.g. CSV) prefers the simple ms/s units rather
 * than the adaptive ns/µs/ms/s ladder used in the UI.
 */
function formatDurationForExport(durationNs?: number): string {
  if (!durationNs) return ''
  const ms = durationNs / 1_000_000
  if (ms < 1000) return `${ms.toFixed(2)}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

/**
 * Escape a value for CSV format. Wraps in double-quotes when the value
 * contains a comma, quote, or newline; embedded quotes are doubled per
 * RFC 4180.
 */
function escapeCSV(val: unknown): string {
  const str = String(val ?? '')
  if (str.includes(',') || str.includes('"') || str.includes('\n')) {
    return `"${str.replace(/"/g, '""')}"`
  }
  return str
}

/**
 * Convert trace + spans to a CSV string. One row per span with the
 * parent trace's name attached as a denormalised column for easier
 * analysis in spreadsheets.
 */
export function traceToCSV(trace: TraceDetail, spans: Span[]): string {
  const headers = [
    'trace_id',
    'trace_name',
    'span_id',
    'parent_span_id',
    'span_name',
    'span_kind',
    'status_code',
    'status_message',
    'start_time',
    'end_time',
    'duration',
    'model_name',
    'provider_name',
    'total_cost',
    'input',
    'output',
  ]

  const rows = spans.map((span) => [
    trace.trace_id,
    trace.name,
    span.span_id,
    span.parent_span_id || '',
    span.span_name,
    span.span_kind,
    span.status_code,
    span.status_message || '',
    span.start_time ? new Date(span.start_time).toISOString() : '',
    span.end_time ? new Date(span.end_time).toISOString() : '',
    formatDurationForExport(span.duration),
    span.model_name || '',
    span.provider_name || '',
    span.total_cost || '',
    JSON.stringify(span.input || ''),
    JSON.stringify(span.output || ''),
  ])

  return [
    headers.join(','),
    ...rows.map((row) => row.map(escapeCSV).join(',')),
  ].join('\n')
}

/**
 * Convert trace + spans to a JSON export shape. Includes the full
 * span attribute bags so a downstream debugger has everything needed
 * to reconstruct the trace context.
 */
export function traceToJSON(trace: TraceDetail, spans: Span[]): string {
  const exportData = {
    trace: {
      trace_id: trace.trace_id,
      name: trace.name,
      status_code: trace.status_code,
      start_time: trace.start_time,
      end_time: trace.end_time,
      duration: trace.duration,
      has_error: trace.has_error,
      user_id: trace.user_id,
      session_id: trace.session_id,
      model_name: trace.model_name,
      provider_name: trace.provider_name,
      total_cost: trace.total_cost,
      input_tokens: trace.input_tokens,
      output_tokens: trace.output_tokens,
      total_tokens: trace.total_tokens,
      tags: trace.tags,
      bookmarked: trace.bookmarked,
    },
    spans: spans.map((span) => ({
      span_id: span.span_id,
      parent_span_id: span.parent_span_id,
      span_name: span.span_name,
      span_kind: span.span_kind,
      status_code: span.status_code,
      status_message: span.status_message,
      start_time: span.start_time,
      end_time: span.end_time,
      duration: span.duration,
      model_name: span.model_name,
      provider_name: span.provider_name,
      total_cost: span.total_cost,
      usage_details: span.usage_details,
      cost_details: span.cost_details,
      input: span.input,
      output: span.output,
      span_attributes: span.span_attributes,
      resource_attributes: span.resource_attributes,
      scope_attributes: span.scope_attributes,
    })),
    exported_at: new Date().toISOString(),
    span_count: spans.length,
  }

  return JSON.stringify(exportData, null, 2)
}

/**
 * Trigger a browser download for the given content. Uses an
 * object-URL + transient anchor so we don't have to leak DOM nodes.
 */
export function downloadFile(
  content: string,
  filename: string,
  mimeType: string,
): void {
  const blob = new Blob([content], { type: mimeType })
  const url = window.URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  window.URL.revokeObjectURL(url)
}
