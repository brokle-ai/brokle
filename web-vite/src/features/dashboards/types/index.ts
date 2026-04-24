/**
 * Dashboard Types
 *
 * Mirrors `internal/core/domain/dashboard/entities.go`. Kept close to
 * the backend wire shape so the API layer is a thin transform.
 */

// Widget type enum
export type WidgetType =
  | 'stat'
  | 'time_series'
  | 'table'
  | 'bar'
  | 'pie'
  | 'heatmap'
  | 'histogram'
  | 'trace_list'
  | 'text'

// View type for widget queries
export type WidgetViewType = 'traces' | 'spans' | 'scores'

// Filter operators (dashboard-local — distinct from the shared
// filter-builder's string operators; see query-builder/filter-builder)
export type FilterOperator =
  | 'eq'
  | 'neq'
  | 'gt'
  | 'lt'
  | 'gte'
  | 'lte'
  | 'contains'
  | 'in'

export interface QueryFilter {
  field: string
  operator: FilterOperator
  value: unknown
}

// Re-export shared TimeRange types
import type {
  TimeRange,
  RelativeTimeRange,
} from '@/components/shared/time-range-picker'
export type { TimeRange, RelativeTimeRange }

export interface WidgetQuery {
  view: WidgetViewType
  measures: string[]
  dimensions?: string[]
  filters?: QueryFilter[]
  time_range?: TimeRange
  limit?: number
  order_by?: string
  order_dir?: 'asc' | 'desc'
}

export interface Widget {
  id: string
  type: WidgetType
  title: string
  description?: string
  query: WidgetQuery
  config?: Record<string, unknown>
}

export interface LayoutItem {
  widget_id: string
  x: number
  y: number
  w: number
  h: number
}

export type VariableType = 'string' | 'number' | 'select' | 'query'

export interface VariableQueryConfig {
  view: WidgetViewType
  dimension: string
  limit?: number
}

export interface Variable {
  name: string
  type: VariableType
  label?: string
  default?: unknown
  options?: string[]
  query_config?: VariableQueryConfig
  multi?: boolean
}

export type VariableValues = Record<string, unknown>

export interface DashboardConfig {
  widgets: Widget[]
  refresh_rate?: number
  time_range?: TimeRange
  variables?: Variable[]
}

export interface Dashboard {
  id: string
  project_id: string
  name: string
  description?: string
  config: DashboardConfig
  layout: LayoutItem[]
  is_locked: boolean
  created_by?: string
  created_at: string
  updated_at: string
}

export type DashboardListItem = Dashboard

export interface CreateDashboardRequest {
  name: string
  description?: string
  config?: DashboardConfig
  layout?: LayoutItem[]
}

export interface UpdateDashboardRequest {
  name?: string
  description?: string
  config?: DashboardConfig
  layout?: LayoutItem[]
}

export interface DashboardListResponse {
  dashboards: Dashboard[]
  total: number
  limit: number
  offset: number
}

export interface DashboardFilter {
  name?: string
  limit?: number
  offset?: number
}

// ============================================
// Widget Query Execution Types
// ============================================

export interface QueryExecutionParams {
  time_range?: TimeRange
  force_refresh?: boolean
  variable_values?: VariableValues
}

export interface QueryMetadata {
  executed_at: string
  duration_ms: number
  row_count: number
  cached: boolean
}

export interface WidgetQueryResult {
  widget_id: string
  data: Record<string, unknown>[] | null
  error?: string
  metadata?: QueryMetadata
}

export interface DashboardQueryResults {
  dashboard_id: string
  results: Record<string, WidgetQueryResult>
  executed_at: string
}

// ============================================
// View Definition Types
// ============================================

export interface MeasureDefinition {
  id: string
  label: string
  description?: string
  type: 'count' | 'sum' | 'avg' | 'min' | 'max' | 'p50' | 'p95' | 'p99' | 'rate'
  unit?: 'count' | 'ms' | 'tokens' | 'USD' | 'percentage'
}

export interface DimensionDefinition {
  id: string
  label: string
  description?: string
  column_type: string
  bucketable: boolean
}

export interface ViewDefinition {
  name: string
  description: string
  measures: MeasureDefinition[]
  dimensions: DimensionDefinition[]
}

export interface ViewDefinitionsResponse {
  views: Record<WidgetViewType, ViewDefinition>
}

// ============================================
// Templates
// ============================================

export type TemplateCategory =
  | 'llm-overview'
  | 'cost-analytics'
  | 'quality-scores'

export interface DashboardTemplate {
  id: string
  name: string
  description: string
  category: TemplateCategory
  config: DashboardConfig
  layout: LayoutItem[]
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface CreateFromTemplateRequest {
  template_id: string
  name: string
}

export interface DuplicateDashboardRequest {
  name: string
}

// ============================================
// Export / Import
// ============================================

export interface DashboardExport {
  version: string
  exported_at: string
  name: string
  description?: string
  config: DashboardConfig
  layout: LayoutItem[]
}

export interface DashboardImportRequest {
  data: DashboardExport
  name?: string
}

// ============================================
// React Grid Layout
// ============================================

export interface ReactGridLayoutItem {
  i: string
  x: number
  y: number
  w: number
  h: number
  minW?: number
  minH?: number
  maxW?: number
  maxH?: number
  static?: boolean
  isDraggable?: boolean
  isResizable?: boolean
}
