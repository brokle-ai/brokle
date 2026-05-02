// Wire types for the trace-sessions list endpoint. Shape matches
// `observability.TraceSessionSummary` in the backend (session_entity.go)
// and the JSON emitted by `GET /api/v1/projects/{projectId}/sessions`.
//
// Sessions are aggregations over traces (grouped by `session_id`), not
// a first-class entity — the list view renders the rollup columns and
// links back to `/traces?session_id=…` for the per-trace detail.

export interface SessionListItem {
  session_id: string
  trace_count: number
  first_trace: string // RFC 3339
  last_trace: string // RFC 3339
  total_duration: number // nanoseconds
  total_tokens: number
  // shopspring/decimal's default JSON encoding emits a number literal
  // (not a quoted string); matches web/'s Next.js client typing.
  total_cost: number
  error_count: number
  user_ids: string[]
}

export interface Pagination {
  page: number
  limit: number
  total: number
  total_pages: number
  has_next: boolean
  has_prev: boolean
}

export interface SessionListResponse {
  data: SessionListItem[]
  pagination: Pagination
}

// There is no dedicated "get session by id" endpoint — sessions are a
// ClickHouse aggregation over the traces table. The detail view
// derives its rollup row by calling the sessions list with a
// `search={sessionId}` filter, which yields the matching summary row.
// `SessionDetail` is just an alias to keep callers expressive.
export type SessionDetail = SessionListItem
