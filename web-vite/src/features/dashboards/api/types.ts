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
