// Wire types for the scores list endpoint at
// GET /api/v1/projects/{projectId}/scores.
//
// Envelope is observability-style `{data, pagination}` where
// `pagination` is `{page, limit, total, total_pages}` (no
// has_next/has_prev — derive at the render layer). Matches the
// `listScoresResponse` DTO in
// internal/transport/http/handlers/observability/dashboard_types.go.
//
// Score value representation: the domain splits numeric vs string/
// categorical storage across two fields — `value` (float64 pointer,
// present for NUMERIC and BOOLEAN where the backend encodes true/false
// as 1.0/0.0) and `string_value` (present for CATEGORICAL). Both are
// optional; consumers select the one appropriate to `type`.

export type ScoreDataType = 'NUMERIC' | 'CATEGORICAL' | 'BOOLEAN'
export type ScoreSource = 'code' | 'llm' | 'human'

export interface ScoreListItem {
  id: string
  project_id: string
  trace_id?: string
  span_id?: string
  name: string
  // Numeric encoding; also holds 1.0 / 0.0 for BOOLEAN scores (per
  // observability.Score in internal/core/domain/observability).
  value?: number
  // Categorical label; only populated when `type === 'CATEGORICAL'`.
  string_value?: string
  type: ScoreDataType
  source: ScoreSource
  reason?: string
  experiment_id?: string
  experiment_item_id?: string
  created_by?: string
  timestamp: string
}

export interface ScoresPagination {
  page: number
  limit: number
  total: number
  total_pages: number
}

export interface ScoreListResponse {
  data: ScoreListItem[]
  pagination: ScoresPagination
}
