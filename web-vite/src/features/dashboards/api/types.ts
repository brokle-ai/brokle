// Wire types for the dashboards list endpoint at
// GET /api/v1/projects/{projectId}/dashboards. Narrow shape — only the
// fields the list view renders. Widget/layout/config shapes live in
// the full Dashboard domain (see internal/core/domain/dashboard/entities.go)
// and are deferred to the dashboard-editor port.
//
// Envelope: the backend returns offset-based pagination —
// `{dashboards: Dashboard[], total, limit, offset}` — NOT the
// observability `{data, pagination}` shape. Derive page numbers at the
// render layer.

export interface DashboardListItem {
  id: string
  project_id: string
  name: string
  description?: string
  is_locked: boolean
  created_by?: string
  created_at: string
  updated_at: string
}

export interface DashboardListResponse {
  dashboards: DashboardListItem[]
  total: number
  limit: number
  offset: number
}

// Widget/layout/config shapes for the detail view. Wire types mirror
// `internal/core/domain/dashboard/entities.go`:
//   - Dashboard has a single DashboardConfig with widgets[], refresh rate,
//     optional time range, variables[]
//   - Layout is a flat list of {widget_id, x, y, w, h} items; the grid
//     editor maps widget IDs → placement. For the static viewer we
//     render widgets in config.widgets order, ignoring layout.
//
// `WidgetType` is the enum from the backend — we keep the string
// literal union in sync. An unrecognised type is still renderable
// (it's just a badge + TODO panel) so we tolerate string at the edge.

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

export interface Widget {
  id: string
  type: WidgetType | string
  title: string
  description?: string
  // `query` and `config` are widget-specific; the static view doesn't
  // execute them, so we keep them as opaque shapes. Typed editing is
  // the next-port scope.
  query: Record<string, unknown>
  config?: Record<string, unknown>
}

export interface LayoutItem {
  widget_id: string
  x: number
  y: number
  w: number
  h: number
}

export interface DashboardConfig {
  widgets: Widget[]
  refresh_rate?: number
  time_range?: Record<string, unknown>
  variables?: Record<string, unknown>[]
}

export interface DashboardDetail {
  id: string
  project_id: string
  name: string
  description?: string
  config: DashboardConfig
  layout: LayoutItem[]
  is_locked: boolean
  created_by?: string
  created_at: string // RFC 3339
  updated_at: string // RFC 3339
}
