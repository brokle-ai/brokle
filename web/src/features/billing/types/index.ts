/**
 * Billing Types
 *
 * TypeScript type definitions for the billing feature,
 * matching the Go backend domain types in internal/core/domain/billing
 *
 * Usage-based billing: Spans + GB Processed + Scores
 */

// Re-export time range types from shared component
export type { TimeRange, RelativeTimeRange } from '@/components/shared/time-range-picker'

// ============================================================================
// Usage Overview Types (3 Dimensions: Spans + GB + Scores)
// ============================================================================

export interface UsageOverview {
  organization_id: string
  period_start: string
  period_end: string

  // Current usage (3 dimensions)
  spans: number
  bytes: number
  scores: number

  // Free tier remaining
  free_spans_remaining: number
  free_bytes_remaining: number
  free_scores_remaining: number

  // Free tier totals (for progress display)
  free_spans_total: number
  free_bytes_total: number
  free_scores_total: number

  // Calculated cost
  estimated_cost: number
}

// ============================================================================
// Time Series & Project Breakdown
// ============================================================================

export interface BillableUsage {
  organization_id: string
  project_id: string
  bucket_time: string

  // Billable dimensions
  span_count: number
  bytes_processed: number
  score_count: number

  // AI provider costs (informational, not billable)
  ai_provider_cost: number
}

export interface BillableUsageSummary {
  organization_id: string
  project_id?: string
  project_name?: string

  total_spans: number
  total_bytes: number
  total_scores: number
  total_ai_provider_cost: number
}

// ============================================================================
// Budget Types
// ============================================================================

export type BudgetType = 'monthly' | 'weekly'

export interface UsageBudget {
  id: string
  organization_id: string
  project_id?: string
  name: string
  budget_type: BudgetType

  // Limits (any can be set, null = no limit)
  span_limit?: number
  bytes_limit?: number
  score_limit?: number
  cost_limit?: number

  // Current usage
  current_spans: number
  current_bytes: number
  current_scores: number
  current_cost: number

  // Alert thresholds (flexible array of percentages, e.g., [50, 80, 100])
  alert_thresholds: number[]

  is_active: boolean
  created_at: string
  updated_at: string
}

export interface CreateBudgetRequest {
  name: string
  project_id?: string
  budget_type: BudgetType
  span_limit?: number
  bytes_limit?: number
  score_limit?: number
  cost_limit?: number
  alert_thresholds?: number[] // e.g., [50, 80, 100]
}

export interface UpdateBudgetRequest extends Partial<CreateBudgetRequest> {
  is_active?: boolean
}

// ============================================================================
// Alert Types
// ============================================================================

export type AlertSeverity = 'info' | 'warning' | 'critical'
export type AlertStatus = 'triggered' | 'acknowledged' | 'resolved'
export type AlertDimension = 'spans' | 'bytes' | 'scores' | 'cost'

export interface UsageAlert {
  id: string
  budget_id?: string
  organization_id: string
  project_id?: string
  alert_threshold: number // The threshold percentage (1-100)
  dimension: AlertDimension
  severity: AlertSeverity
  threshold_value: number
  actual_value: number
  percent_used: number
  status: AlertStatus
  triggered_at: string
  acknowledged_at?: string
  notification_sent: boolean
}

// ============================================================================
// Pricing Types
// ============================================================================

export interface PricingConfig {
  id: string
  name: string

  // Span pricing (per 100K)
  free_spans: number
  price_per_100k_spans?: number

  // Data volume pricing (per GB)
  free_gb: number
  price_per_gb?: number

  // Score pricing (per 1K)
  free_scores: number
  price_per_1k_scores?: number

  is_active: boolean
}

// ============================================================================
// Helper Types
// ============================================================================

export interface UsageTimeSeriesParams {
  from: string
  to: string
  granularity?: 'hourly' | 'daily'
}

export interface UsageByProjectParams {
  from: string
  to: string
}

// Utility function to format bytes to human readable
export const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`
}

// Utility function to format large numbers
export const formatNumber = (num: number): string => {
  if (num >= 1_000_000) return `${(num / 1_000_000).toFixed(1)}M`
  if (num >= 1_000) return `${(num / 1_000).toFixed(1)}K`
  return num.toString()
}
