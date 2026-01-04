/**
 * Dashboard Types
 *
 * TypeScript type definitions for the dashboard feature,
 * matching the Go backend domain types.
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

// Filter operators
export type FilterOperator =
  | 'eq'
  | 'neq'
  | 'gt'
  | 'lt'
  | 'gte'
  | 'lte'
  | 'contains'
  | 'in'

// Query filter for widget queries
export interface QueryFilter {
  field: string
  operator: FilterOperator
  value: unknown
}

// Import and re-export time range types from shared component
import type { TimeRange, RelativeTimeRange } from '@/components/shared/time-range-picker'
export type { TimeRange, RelativeTimeRange }

// Widget query configuration
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

// Widget configuration
export interface Widget {
  id: string
  type: WidgetType
  title: string
  description?: string
  query: WidgetQuery
  config?: Record<string, unknown>
}

// Layout item for widget positioning
export interface LayoutItem {
  widget_id: string
  x: number
  y: number
  w: number
  h: number
}

// Variable type enum
export type VariableType = 'string' | 'number' | 'select' | 'query'

// Query config for dynamic variable options
export interface VariableQueryConfig {
  view: WidgetViewType
  dimension: string
  limit?: number
}

// Variable for dashboard-level filters
export interface Variable {
  name: string
  type: VariableType
  label?: string // Display label (defaults to name)
  default?: unknown
  options?: string[] // For 'select' type - static options
  query_config?: VariableQueryConfig // For 'query' type - dynamic options
  multi?: boolean // Allow multiple values
}

// Variable values map (variable name -> current value)
export type VariableValues = Record<string, unknown>

// Dashboard configuration
export interface DashboardConfig {
  widgets: Widget[]
  refresh_rate?: number
  time_range?: TimeRange
  variables?: Variable[]
}

// Main Dashboard entity
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

// Dashboard list item (same as full dashboard for now)
export type DashboardListItem = Dashboard

// Create dashboard request
export interface CreateDashboardRequest {
  name: string
  description?: string
  config?: DashboardConfig
  layout?: LayoutItem[]
}

// Update dashboard request
export interface UpdateDashboardRequest {
  name?: string
  description?: string
  config?: DashboardConfig
  layout?: LayoutItem[]
}

// Dashboard list response
export interface DashboardListResponse {
  dashboards: Dashboard[]
  total: number
  limit: number
  offset: number
}

// Dashboard filter parameters
export interface DashboardFilter {
  name?: string
  limit?: number
  offset?: number
}

// Get dashboards params
export interface GetDashboardsParams {
  projectId: string
  filter?: DashboardFilter
}

// ============================================
// Widget Query Execution Types
// ============================================

// Query execution request parameters
export interface QueryExecutionParams {
  time_range?: TimeRange
  force_refresh?: boolean
  variable_values?: VariableValues
}

// Query metadata from execution
export interface QueryMetadata {
  executed_at: string
  duration_ms: number
  row_count: number
  cached: boolean
}

// Single widget query result
export interface WidgetQueryResult {
  widget_id: string
  data: Record<string, unknown>[] | null
  error?: string
  metadata?: QueryMetadata
}

// Dashboard query results (all widgets)
export interface DashboardQueryResults {
  dashboard_id: string
  results: Record<string, WidgetQueryResult>
  executed_at: string
}

// ============================================
// View Definition Types (for Query Builder)
// ============================================

// Measure definition for query builder UI
export interface MeasureDefinition {
  id: string
  label: string
  description?: string
  type: 'count' | 'sum' | 'avg' | 'min' | 'max' | 'p50' | 'p95' | 'p99' | 'rate'
  unit?: 'count' | 'ms' | 'tokens' | 'USD' | 'percentage'
}

// Dimension definition for query builder UI
export interface DimensionDefinition {
  id: string
  label: string
  description?: string
  column_type: string
  bucketable: boolean
}

// View definition for a data source
export interface ViewDefinition {
  name: string
  description: string
  measures: MeasureDefinition[]
  dimensions: DimensionDefinition[]
}

// View definitions response from API
export interface ViewDefinitionsResponse {
  views: Record<WidgetViewType, ViewDefinition>
}

// ============================================
// Dashboard Template Types
// ============================================

// Template category enum
export type TemplateCategory = 'llm-overview' | 'cost-analytics' | 'quality-scores'

// Dashboard template entity
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

// Create dashboard from template request
export interface CreateFromTemplateRequest {
  template_id: string
  name: string
}

// Duplicate dashboard request
export interface DuplicateDashboardRequest {
  name: string
}

// ============================================
// Dashboard Export/Import Types
// ============================================

// Dashboard export for sharing or backup
export interface DashboardExport {
  version: string
  exported_at: string
  name: string
  description?: string
  config: DashboardConfig
  layout: LayoutItem[]
}

// Dashboard import request
export interface DashboardImportRequest {
  data: DashboardExport
  name?: string
}

// ============================================
// React Grid Layout Types
// ============================================

// React-grid-layout compatible layout item
export interface ReactGridLayoutItem {
  i: string // widget_id
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
