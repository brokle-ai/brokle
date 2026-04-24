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

// Score-config endpoint — GET /api/v1/projects/{projectId}/score-configs.
// Envelope is the evaluation-package flat pageList — `{data, total,
// page, limit}`. `min_value`/`max_value` are only present for NUMERIC;
// `categories` only for CATEGORICAL. The annotation form + review UI
// dispatch on `type` to render the right input widget.
export interface ScoreConfig {
  id: string
  project_id: string
  name: string
  description?: string
  type: ScoreDataType
  min_value?: number
  max_value?: number
  categories?: string[]
  metadata?: Record<string, unknown>
  created_at: string
  updated_at: string
}

export interface ScoreConfigListResponse {
  data: ScoreConfig[]
  total: number
  page: number
  limit: number
}
