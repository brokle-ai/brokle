import { queryOptions } from '@tanstack/react-query'
import { rawFetch } from '@/lib/api/client'
import type { DashboardDetail, DashboardListResponse } from './types'

// Hierarchical keys so the deferred editor/duplicate/delete mutations can
// bust the list cache cleanly via `lists()`.
export const dashboardsKeys = {
  all: ['dashboards'] as const,
  lists: () => [...dashboardsKeys.all, 'list'] as const,
  list: (projectId: string, params: DashboardListParams) =>
    [...dashboardsKeys.lists(), projectId, params] as const,
  details: () => [...dashboardsKeys.all, 'detail'] as const,
  detail: (dashboardId: string) =>
    [...dashboardsKeys.details(), dashboardId] as const,
} as const

export interface DashboardListParams {
  page: number
  limit: number
  q?: string
}

export const dashboardListQueryOptions = (
  projectId: string,
  params: DashboardListParams,
) =>
  queryOptions({
    queryKey: dashboardsKeys.list(projectId, params),
    queryFn: async () => {
      // Backend uses offset pagination. Translate page/limit → offset at
      // the edge so the route layer stays in page-based UX terms.
      const offset = Math.max(0, (params.page - 1) * params.limit)
      const search = new URLSearchParams()
      search.set('limit', String(params.limit))
      search.set('offset', String(offset))
      if (params.q && params.q.length > 0) {
        search.set('name', params.q)
      }
      const resp = await rawFetch(
        `/api/v1/projects/${projectId}/dashboards?${search.toString()}`,
        { method: 'GET' },
      )
      return (await resp.json()) as DashboardListResponse
    },
    staleTime: 30 * 1000,
  })

// GET /api/v1/projects/{projectId}/dashboards/{dashboardId}
// Returns the full Dashboard entity (config.widgets + layout). The
// static viewer renders widgets in `config.widgets` order; the
// drag-drop grid editor that consumes `layout` is a next-port concern.
export const dashboardDetailQueryOptions = (
  projectId: string,
  dashboardId: string,
) =>
  queryOptions({
    queryKey: dashboardsKeys.detail(dashboardId),
    queryFn: async () => {
      const resp = await rawFetch(
        `/api/v1/projects/${projectId}/dashboards/${encodeURIComponent(dashboardId)}`,
        { method: 'GET' },
      )
      return (await resp.json()) as DashboardDetail
    },
    staleTime: 30 * 1000,
  })
