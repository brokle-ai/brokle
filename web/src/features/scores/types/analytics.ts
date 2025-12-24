/**
 * Score Analytics Types
 * TypeScript types matching Go domain types in internal/core/domain/observability/repository.go
 */

export interface ScoreAnalyticsParams {
  score_name: string
  compare_score_name?: string
  from_timestamp?: string
  to_timestamp?: string
  interval?: 'hour' | 'day' | 'week'
}

export interface ScoreStatistics {
  count: number
  mean: number
  std_dev: number
  min: number
  max: number
  median: number
  mode?: string
  mode_percent?: number
}

export interface TimeSeriesPoint {
  timestamp: string
  avg_value: number
  count: number
}

export interface DistributionBin {
  bin_start: number
  bin_end: number
  count: number
}

export interface HeatmapCell {
  row: number
  col: number
  value: number
  row_label: string
  col_label: string
}

export interface ComparisonMetrics {
  matched_count: number
  pearson_correlation: number
  spearman_correlation: number
  mae: number
  rmse: number
  cohens_kappa?: number
  overall_agreement?: number
}

export interface ScoreAnalyticsData {
  statistics: ScoreStatistics
  time_series: TimeSeriesPoint[]
  distribution: DistributionBin[]
  heatmap?: HeatmapCell[]
  comparison?: ComparisonMetrics
  compare_statistics?: ScoreStatistics
  compare_time_series?: TimeSeriesPoint[]
  compare_distribution?: DistributionBin[]
}

// UI-specific types for rendering
export interface InterpretationResult {
  strength: 'Very Strong' | 'Strong' | 'Moderate' | 'Weak' | 'Very Weak' | 'None'
  color: 'green' | 'blue' | 'yellow' | 'orange' | 'red' | 'gray'
  description: string
}

export interface ChartDataPoint {
  date: string
  value: number
  count?: number
  label?: string
}

export interface DistributionChartData {
  label: string
  value: number
  percentage: number
}
