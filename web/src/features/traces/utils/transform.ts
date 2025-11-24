/**
 * Response Transformers and Type Converters for Traces API
 *
 * Handles:
 * - Backend snake_case → Frontend camelCase conversion
 * - Enum conversions (UInt8 → strings)
 * - JSON attribute parsing
 * - Date conversions
 * - Safe error handling
 */

import type { Trace, Span, Score } from '../data/schema'

// ============================================================================
// Enum Converters
// ============================================================================

/**
 * Convert OTEL status code (UInt8) to string representation
 * @param code - 0 (UNSET), 1 (OK), 2 (ERROR)
 * @returns Status string
 */
export function statusCodeToString(code: number): 'unset' | 'ok' | 'error' {
  const map: Record<number, 'unset' | 'ok' | 'error'> = {
    0: 'unset',
    1: 'ok',
    2: 'error',
  }
  return map[code] ?? 'unset'
}

/**
 * Convert status string to OTEL status code (UInt8)
 * @param status - Status string
 * @returns Status code number
 */
export function stringToStatusCode(status: string): number {
  const map: Record<string, number> = {
    unset: 0,
    ok: 1,
    error: 2,
  }
  return map[status.toLowerCase()] ?? 0
}

/**
 * Convert OTEL span kind (UInt8) to string representation
 * @param kind - 0-5 enum value
 * @returns Span kind string
 */
export function spanKindToString(kind: number): string {
  const map: Record<number, string> = {
    0: 'UNSPECIFIED',
    1: 'INTERNAL',
    2: 'SERVER',
    3: 'CLIENT',
    4: 'PRODUCER',
    5: 'CONSUMER',
  }
  return map[kind] ?? 'UNSPECIFIED'
}

/**
 * Convert span kind string to OTEL span kind (UInt8)
 * @param kind - Span kind string
 * @returns Span kind number
 */
export function stringToSpanKind(kind: string): number {
  const map: Record<string, number> = {
    UNSPECIFIED: 0,
    INTERNAL: 1,
    SERVER: 2,
    CLIENT: 3,
    PRODUCER: 4,
    CONSUMER: 5,
  }
  return map[kind.toUpperCase()] ?? 0
}

// ============================================================================
// JSON Attribute Parsers
// ============================================================================

/**
 * Safely parse JSON string with error handling
 * @param jsonString - JSON string to parse
 * @param fallback - Fallback value if parsing fails
 * @returns Parsed object or fallback
 */
export function parseAttributes<T = Record<string, unknown>>(
  jsonString: string | null | undefined,
  fallback: T = {} as T
): T {
  if (!jsonString || jsonString.trim() === '') {
    return fallback
  }

  try {
    return JSON.parse(jsonString) as T
  } catch (error) {
    console.warn('[Traces Transform] Failed to parse attributes:', error)
    return fallback
  }
}

/**
 * Safely stringify object to JSON
 * @param obj - Object to stringify
 * @returns JSON string or empty object string
 */
export function stringifyAttributes(obj: unknown): string {
  try {
    return JSON.stringify(obj)
  } catch (error) {
    console.warn('[Traces Transform] Failed to stringify attributes:', error)
    return '{}'
  }
}

// ============================================================================
// Date Converters
// ============================================================================

/**
 * Convert backend timestamp to Date object
 * @param timestamp - ISO string or timestamp
 * @returns Date object or undefined
 */
export function parseTimestamp(timestamp: string | null | undefined): Date | undefined {
  if (!timestamp) return undefined

  try {
    const date = new Date(timestamp)
    if (isNaN(date.getTime())) {
      console.warn('[Traces Transform] Invalid timestamp:', timestamp)
      return undefined
    }
    return date
  } catch (error) {
    console.warn('[Traces Transform] Failed to parse timestamp:', timestamp, error)
    return undefined
  }
}

/**
 * Convert Date object to ISO string
 * @param date - Date object
 * @returns ISO string or null
 */
export function toISOString(date: Date | null | undefined): string | null {
  if (!date || !(date instanceof Date)) return null
  return date.toISOString()
}

// ============================================================================
// Trace Transformers
// ============================================================================

/**
 * Transform backend trace response to frontend Trace type
 * @param raw - Raw backend trace object
 * @returns Transformed Trace object
 */
export function transformTrace(raw: any): Trace {
  return {
    trace_id: raw.trace_id || raw.traceId || '',
    project_id: raw.project_id || raw.projectId || '',
    name: raw.name || 'Unnamed Trace',
    user_id: raw.user_id || raw.userId || undefined,
    session_id: raw.session_id || raw.sessionId || undefined,

    // Timestamps
    start_time: parseTimestamp(raw.start_time || raw.startTime) || new Date(),
    end_time: parseTimestamp(raw.end_time || raw.endTime),
    duration: raw.duration || undefined, // Nanoseconds (OTLP spec)

    // Status
    status_code: raw.status_code ?? raw.statusCode ?? 0,
    status_message: raw.status_message || raw.statusMessage || undefined,
    has_error: raw.has_error ?? (raw.status_code === 2 || raw.statusCode === 2),

    // Attributes (parse JSON strings)
    resource_attributes: typeof raw.resource_attributes === 'string'
      ? parseAttributes(raw.resource_attributes)
      : (raw.resource_attributes || raw.resourceAttributes || {}),

    // I/O data
    input: raw.input || undefined,
    output: raw.output || undefined,

    // Tags
    tags: Array.isArray(raw.tags) ? raw.tags : [],

    // Extracted attributes
    environment: raw.environment || '',
    service_name: raw.service_name || raw.serviceName || undefined,
    service_version: raw.service_version || raw.serviceVersion || undefined,
    release: raw.release || undefined,

    // Flags
    bookmarked: raw.bookmarked ?? false,
    public: raw.public ?? false,

    // Versioning
    version: raw.version || undefined,

    // Computed fields
    cost: raw.cost || undefined,
    tokens: raw.tokens || undefined,
    spanCount: raw.span_count || raw.spanCount || 0,

    // Timestamps (duplicates for compatibility)
    created_at: parseTimestamp(raw.created_at || raw.createdAt) || new Date(),
    updated_at: parseTimestamp(raw.updated_at || raw.updatedAt),

    // Relationships (optional)
    spans: raw.spans ? raw.spans.map(transformSpan) : undefined,
    scores: raw.scores ? raw.scores.map(transformScore) : undefined,
  }
}

/**
 * Transform frontend Trace to backend request format
 * @param trace - Frontend Trace object
 * @returns Backend-compatible object
 */
export function serializeTrace(trace: Partial<Trace>): any {
  return {
    trace_id: trace.trace_id,
    project_id: trace.project_id,
    name: trace.name,
    user_id: trace.user_id,
    session_id: trace.session_id,
    start_time: toISOString(trace.start_time),
    end_time: toISOString(trace.end_time),
    duration_ms: trace.duration_ms,
    status_code: trace.status_code,
    status_message: trace.status_message,
    resource_attributes: stringifyAttributes(trace.resource_attributes),
    input: trace.input,
    output: trace.output,
    tags: trace.tags,
    environment: trace.environment,
    service_name: trace.service_name,
    service_version: trace.service_version,
    release: trace.release,
    bookmarked: trace.bookmarked,
    public: trace.public,
    version: trace.version,
  }
}

// ============================================================================
// Span Transformers
// ============================================================================

/**
 * Transform backend span response to frontend Span type
 * @param raw - Raw backend span object
 * @returns Transformed Span object
 */
export function transformSpan(raw: any): Span {
  return {
    span_id: raw.span_id || raw.spanId || '',
    trace_id: raw.trace_id || raw.traceId || '',
    parent_span_id: raw.parent_span_id || raw.parentSpanId || undefined,
    project_id: raw.project_id || raw.projectId || '',

    // Metadata
    span_name: raw.span_name || raw.spanName || 'Unnamed Span',
    span_kind: raw.span_kind ?? raw.spanKind ?? 0,

    // Timestamps
    start_time: parseTimestamp(raw.start_time || raw.startTime) || new Date(),
    end_time: parseTimestamp(raw.end_time || raw.endTime),
    duration: raw.duration || undefined, // Nanoseconds (OTLP spec)

    // Status
    status_code: raw.status_code ?? raw.statusCode ?? 0,
    status_message: raw.status_message || raw.statusMessage || undefined,
    has_error: raw.has_error ?? (raw.status_code === 2 || raw.statusCode === 2),

    // Attributes (new schema names - parse JSON strings if needed)
    attributes: typeof raw.attributes === 'string'
      ? parseAttributes(raw.attributes)
      : (raw.attributes || {}),
    metadata: typeof raw.metadata === 'string'
      ? parseAttributes(raw.metadata)
      : (raw.metadata || {}),

    // I/O data
    input: raw.input || undefined,
    output: raw.output || undefined,

    // OTEL Events/Links (arrays)
    events_timestamp: raw.events_timestamp || raw.eventsTimestamp || [],
    events_name: raw.events_name || raw.eventsName || [],
    events_attributes: raw.events_attributes || raw.eventsAttributes || [],
    links_trace_id: raw.links_trace_id || raw.linksTraceId || [],
    links_span_id: raw.links_span_id || raw.linksSpanId || [],
    links_attributes: raw.links_attributes || raw.linksAttributes || [],

    // Materialized Columns (16 total)
    gen_ai_operation_name: raw.gen_ai_operation_name || raw.genAiOperationName || undefined,
    gen_ai_provider_name: raw.gen_ai_provider_name || raw.genAiProviderName || undefined,
    gen_ai_request_model: raw.gen_ai_request_model || raw.genAiRequestModel || undefined,
    gen_ai_request_max_tokens: raw.gen_ai_request_max_tokens || raw.genAiRequestMaxTokens || undefined,
    gen_ai_request_temperature: raw.gen_ai_request_temperature || raw.genAiRequestTemperature || undefined,
    gen_ai_request_top_p: raw.gen_ai_request_top_p || raw.genAiRequestTopP || undefined,
    gen_ai_usage_input_tokens: raw.gen_ai_usage_input_tokens || raw.genAiUsageInputTokens || undefined,
    gen_ai_usage_output_tokens: raw.gen_ai_usage_output_tokens || raw.genAiUsageOutputTokens || undefined,

    // Materialized columns (new schema)
    model_name: raw.model_name || raw.modelName || undefined,
    provider_name: raw.provider_name || raw.providerName || undefined,
    span_type: raw.span_type || raw.spanType || undefined,
    version: raw.version || undefined,
    level: raw.level || undefined,

    // Usage & Cost Maps
    usage_details: raw.usage_details || raw.usageDetails || undefined,
    cost_details: raw.cost_details || raw.costDetails || undefined,
    pricing_snapshot: raw.pricing_snapshot || raw.pricingSnapshot || undefined,
    total_cost: raw.total_cost || raw.totalCost || undefined,

    // Timestamps
    created_at: parseTimestamp(raw.created_at || raw.createdAt) || new Date(),

    // Relationships (optional)
    scores: raw.scores ? raw.scores.map(transformScore) : undefined,
    child_spans: raw.child_spans ? raw.child_spans.map(transformSpan) : undefined,
  }
}

// ============================================================================
// Score Transformers
// ============================================================================

/**
 * Transform backend score response to frontend Score type
 * @param raw - Raw backend score object
 * @returns Transformed Score object
 */
export function transformScore(raw: any): Score {
  return {
    id: raw.id || '',
    project_id: raw.project_id || raw.projectId || '',
    trace_id: raw.trace_id || raw.traceId || '',
    span_id: raw.span_id || raw.spanId || '',

    // Score data
    name: raw.name || '',
    value: raw.value ?? undefined,
    string_value: raw.string_value || raw.stringValue || undefined,
    data_type: raw.data_type || raw.dataType || 'NUMERIC',

    // Metadata
    source: raw.source || 'API',
    comment: raw.comment || undefined,

    // Evaluator info
    evaluator_name: raw.evaluator_name || raw.evaluatorName || undefined,
    evaluator_version: raw.evaluator_version || raw.evaluatorVersion || undefined,
    evaluator_config: raw.evaluator_config || raw.evaluatorConfig || undefined,
    author_user_id: raw.author_user_id || raw.authorUserId || undefined,

    // Timestamps
    timestamp: parseTimestamp(raw.timestamp) || new Date(),
    version: raw.version || undefined,
  }
}

// ============================================================================
// Pagination Transformer
// ============================================================================

/**
 * Transform backend pagination response
 * @param raw - Raw backend pagination object
 * @returns Transformed pagination object
 */
export function transformPagination(raw: any) {
  return {
    page: raw.page || 1,
    limit: raw.limit || 20,
    total: raw.total || 0,
    totalPages: raw.total_pages || raw.totalPages || 0,
  }
}

// ============================================================================
// API Response Transformers
// ============================================================================

/**
 * Transform traces list API response
 * @param response - Raw API response
 * @returns Transformed traces and pagination
 */
export function transformTracesResponse(response: any): {
  traces: Trace[]
  pagination: ReturnType<typeof transformPagination>
} {
  return {
    traces: Array.isArray(response.data) ? response.data.map(transformTrace) : [],
    // Backend returns pagination in response.meta.pagination
    pagination: transformPagination(response.meta?.pagination || response.pagination || {}),
  }
}

/**
 * Transform single trace API response
 * @param response - Raw API response
 * @returns Transformed trace
 */
export function transformTraceResponse(response: any): Trace {
  return transformTrace(response.data || response)
}

/**
 * Transform spans list API response
 * @param response - Raw API response
 * @returns Transformed spans and pagination
 */
export function transformSpansResponse(response: any): {
  spans: Span[]
  pagination: ReturnType<typeof transformPagination>
} {
  return {
    spans: Array.isArray(response.data) ? response.data.map(transformSpan) : [],
    // Backend returns pagination in response.meta.pagination
    pagination: transformPagination(response.meta?.pagination || response.pagination || {}),
  }
}
