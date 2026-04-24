// Wire types for the traces list endpoint. Narrow shape — only the
// fields the list view actually renders. A detail-view port will add
// Span[] / Score[] / per-span materialized columns against a richer
// type in a follow-up commit.
//
// Shape comes from the `observability.traces` ClickHouse table (see
// migrations/clickhouse/20251112000001_create_otel_traces.up.sql) and
// matches the JSON emitted by `GET /v1/traces`. The legacy Next.js
// feature at web/src/features/traces/data/schema.ts ships a richer Zod
// schema with Date-valued fields; the backend emits RFC 3339 strings,
// so we keep the wire type string-typed and let components format on
// render.

export interface TraceListItem {
  // OTEL identifiers
  trace_id: string
  project_id: string

  // Metadata
  name: string
  user_id?: string
  session_id?: string

  // Timing (RFC 3339 strings on the wire)
  start_time: string
  end_time?: string
  duration?: number // nanoseconds

  // OTEL status
  status_code: number // 0=UNSET, 1=OK, 2=ERROR
  status_message?: string
  has_error: boolean

  // I/O previews (may be absent on the list endpoint; available on
  // detail)
  input?: string
  output?: string

  // Tags and extracted attrs
  tags?: string[]
  environment?: string
  service_name?: string
  service_version?: string
  release?: string

  // Model / provider (lifted from the root span)
  model_name?: string
  provider_name?: string

  // Flags
  bookmarked: boolean
  public: boolean

  // Versioning
  version?: string

  // Computed rollups
  cost?: number
  tokens?: number
  spanCount?: number

  // Record timestamps
  created_at?: string
  updated_at?: string
}

export interface Pagination {
  page: number
  limit: number
  total: number
  total_pages: number
  has_next: boolean
  has_prev: boolean
}

export interface TraceListResponse {
  data: TraceListItem[]
  pagination: Pagination
}
