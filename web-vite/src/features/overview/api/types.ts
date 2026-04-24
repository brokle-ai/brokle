// Wire types for GET /api/v1/projects/{projectId}/overview. Shape
// comes from `analytics.OverviewResponse` in the backend
// (internal/core/domain/analytics/overview.go). The Phase-1.5 port
// only renders the stats row + checklist — the chart arrays and
// `scores_summary` are kept in the type so a follow-up chart port
// doesn't need to retype the response.

export interface OverviewStats {
  traces_count: number
  traces_trend: number
  total_cost: number
  cost_trend: number
  total_tokens: number
  tokens_trend: number
  avg_latency_ms: number
  latency_trend: number
  error_rate: number
  error_rate_trend: number
}

export interface TimeSeriesPoint {
  timestamp: string
  value: number
}

export interface CostByModel {
  model: string
  cost: number
  tokens: number
  count: number
}

export interface RecentTrace {
  trace_id: string
  name: string
  latency_ms: number
  status: 'success' | 'error'
  timestamp: string
}

export interface TopError {
  message: string
  count: number
  last_seen: string
}

export interface ScoreSummary {
  name: string
  avg_value: number
  trend: number
  sparkline: TimeSeriesPoint[]
}

export interface ChecklistStatus {
  has_project: boolean
  has_traces: boolean
  has_ai_provider: boolean
  has_evaluations: boolean
}

export interface OverviewResponse {
  stats: OverviewStats
  trace_volume: TimeSeriesPoint[]
  cost_time_series: TimeSeriesPoint[]
  token_time_series: TimeSeriesPoint[]
  error_time_series: TimeSeriesPoint[]
  cost_by_model: CostByModel[]
  recent_traces: RecentTrace[]
  top_errors: TopError[]
  // Emitted with `omitempty` on the Go side — treat as optional.
  scores_summary?: ScoreSummary[]
  checklist_status: ChecklistStatus
}

// Matches the backend enum on GetOverviewInput.TimeRange.
export type OverviewTimeRange =
  | '15m'
  | '30m'
  | '1h'
  | '3h'
  | '6h'
  | '12h'
  | '24h'
  | '7d'
  | '14d'
  | '30d'
  | 'all'
