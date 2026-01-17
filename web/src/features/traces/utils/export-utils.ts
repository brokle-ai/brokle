import type { Trace, Span } from '../data/schema'

/**
 * Format duration from nanoseconds to human-readable string
 */
function formatDuration(durationNs?: number): string {
  if (!durationNs) return ''
  const ms = durationNs / 1_000_000
  if (ms < 1000) return `${ms.toFixed(2)}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

/**
 * Escape a value for CSV format
 * Wraps in quotes if contains comma, quotes, or newlines
 */
function escapeCSV(val: unknown): string {
  const str = String(val ?? '')
  if (str.includes(',') || str.includes('"') || str.includes('\n')) {
    return `"${str.replace(/"/g, '""')}"`
  }
  return str
}

/**
 * Convert trace + spans to CSV string
 * Format: One row per span with trace context
 */
export function traceToCSV(trace: Trace, spans: Span[]): string {
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
    'input_tokens',
    'output_tokens',
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
    formatDuration(span.duration),
    span.model_name || '',
    span.provider_name || '',
    span.gen_ai_usage_input_tokens || '',
    span.gen_ai_usage_output_tokens || '',
    span.total_cost || '',
    JSON.stringify(span.input || ''),
    JSON.stringify(span.output || ''),
  ])

  const csvRows = [
    headers.join(','),
    ...rows.map((row) => row.map(escapeCSV).join(',')),
  ]

  return csvRows.join('\n')
}

/**
 * Convert trace + spans to JSON export format
 */
export function traceToJSON(trace: Trace, spans: Span[]): string {
  const exportData = {
    trace: {
      trace_id: trace.trace_id,
      name: trace.name,
      status_code: trace.status_code,
      status_message: trace.status_message,
      start_time: trace.start_time,
      end_time: trace.end_time,
      duration: trace.duration,
      has_error: trace.has_error,
      user_id: trace.user_id,
      session_id: trace.session_id,
      model_name: trace.model_name,
      provider_name: trace.provider_name,
      cost: trace.cost,
      tokens: trace.tokens,
      tags: trace.tags,
      bookmarked: trace.bookmarked,
      input: trace.input,
      output: trace.output,
      resource_attributes: trace.resource_attributes,
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
      input_tokens: span.gen_ai_usage_input_tokens,
      output_tokens: span.gen_ai_usage_output_tokens,
      total_cost: span.total_cost,
      input: span.input,
      output: span.output,
      attributes: span.attributes,
      metadata: span.metadata,
    })),
    exported_at: new Date().toISOString(),
    span_count: spans.length,
  }

  return JSON.stringify(exportData, null, 2)
}

/**
 * Trigger file download in browser
 */
export function downloadFile(
  content: string,
  filename: string,
  mimeType: string
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
