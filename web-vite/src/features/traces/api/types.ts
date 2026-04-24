// Wire types for traces endpoints. Shape mirrors
// `internal/core/domain/observability.TraceSummary` and `.Span`.
// Timestamps are RFC 3339 strings; durations are nanoseconds (OTLP
// spec); monetary values arrive as shopspring/decimal-serialized
// strings (e.g. "0.000123"), parsed at the render boundary.

export interface TraceSummary {
  // OTEL identifiers (W3C hex)
  trace_id: string
  root_span_id: string
  project_id: string

  // Metadata
  name: string
  user_id?: string
  session_id?: string
  service_name?: string
  model_name?: string
  provider_name?: string

  // Timing (RFC 3339 strings)
  start_time: string
  end_time?: string
  duration?: number // nanoseconds

  // OTEL status
  status_code?: number // 0=UNSET, 1=OK, 2=ERROR
  has_error: boolean

  // Rollups — cost is a decimal string from shopspring/decimal
  total_cost: string
  input_tokens: number
  output_tokens: number
  total_tokens: number
  span_count: number
  error_span_count: number

  // User-managed
  tags?: string[]
  bookmarked: boolean
}

// Alias used by the list-view table/components for clarity.
export type TraceListItem = TraceSummary

export interface Pagination {
  page: number
  limit: number
  total: number
  total_pages: number
  has_next: boolean
  has_prev: boolean
}

export interface TraceListResponse {
  data: TraceSummary[]
  pagination: Pagination
}

// Full TraceSummary response is the same shape used in the list.
export type TraceDetail = TraceSummary

// Span wire type. Mirrors `observability.Span` — optional fields align
// with the `omitempty` tags on the Go struct.
export interface Span {
  trace_id: string
  span_id: string
  parent_span_id?: string
  project_id: string
  organization_id: string

  span_name: string
  span_kind: number // OTEL SpanKind enum
  status_code: number // 0=UNSET, 1=OK, 2=ERROR
  status_message?: string
  has_error: boolean

  start_time: string
  end_time?: string
  duration?: number // nanoseconds

  input?: string
  output?: string

  // Gen-AI extracted attributes (materialized by the ingestion worker)
  model_name?: string
  provider_name?: string
  span_type?: string
  level?: string
  service_name?: string

  // Attribute bags
  resource_attributes?: Record<string, string>
  span_attributes?: Record<string, string>
  scope_attributes?: Record<string, string>

  // Token / cost breakdown
  usage_details?: Record<string, number>
  cost_details?: Record<string, string>
  total_cost?: string

  version?: string
  tags?: string[]
  bookmarked?: boolean
}

// ChatML message shape — a small, permissive subset used to drive
// the IO preview's chat-view render path. Content may be absent when
// the assistant only emitted tool_calls.
export interface ChatMlToolCall {
  id?: string
  type?: 'function' | string
  function?: {
    name?: string
    arguments?: string
  }
}

export interface ChatMlMessage {
  role: 'system' | 'user' | 'assistant' | 'tool' | string
  content?: string | null
  name?: string
  tool_calls?: ChatMlToolCall[]
  tool_call_id?: string
}
