/**
 * Overview Types
 *
 * TypeScript type definitions for the overview feature,
 * matching the Go backend domain types in internal/core/domain/analytics/overview.go
 */

// Re-export time range types from shared component
export type { TimeRange, RelativeTimeRange } from '@/components/shared/time-range-picker'

// Legacy type alias for backward compatibility (deprecated, use TimeRange instead)
export type OverviewTimeRange = '15m' | '30m' | '1h' | '3h' | '6h' | '12h' | '24h' | '7d' | '14d' | '30d'

// Stats row metrics with trend indicators
export interface OverviewStats {
  traces_count: number
  traces_trend: number
  total_cost: number
  cost_trend: number
  avg_latency_ms: number
  latency_trend: number
  error_rate: number
  error_rate_trend: number
}

// Time series data point for charts
export interface TimeSeriesPoint {
  timestamp: string
  value: number
}

// Cost breakdown by model
export interface CostByModel {
  model: string
  cost: number
}

// Recent trace summary
export interface RecentTrace {
  trace_id: string
  name: string
  latency_ms: number
  status: 'success' | 'error'
  timestamp: string
}

// Top error summary
export interface TopError {
  message: string
  count: number
  last_seen: string
}

// Score summary with sparkline data
export interface ScoreSummary {
  name: string
  avg_value: number
  trend: number
  sparkline: TimeSeriesPoint[]
}

// Onboarding checklist status
export interface ChecklistStatus {
  has_project: boolean
  has_traces: boolean
  has_ai_provider: boolean
  has_evaluations: boolean
}

// Complete overview response
export interface OverviewResponse {
  stats: OverviewStats
  trace_volume: TimeSeriesPoint[]
  cost_by_model: CostByModel[]
  recent_traces: RecentTrace[]
  top_errors: TopError[]
  scores_summary: ScoreSummary[] | null
  checklist_status: ChecklistStatus
}

// Overview request parameters
export interface OverviewRequest {
  time_range?: OverviewTimeRange
}
