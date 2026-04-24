import { queryOptions } from '@tanstack/react-query'
import { getDashboards, getDashboardById } from './dashboards-api'
import type { Dashboard, DashboardListResponse } from '../types'

// Hierarchical query keys shared with the mutation hooks. Kept
// separate from `dashboardQueryKeys` in hooks/ to avoid a cyclic
// import from routes → queries → hooks. The mutation invalidations
// use `['dashboards', 'list']` as a prefix so both shapes stay in
// lockstep.
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
    queryFn: async (): Promise<DashboardListResponse> => {
      const offset = Math.max(0, (params.page - 1) * params.limit)
      return getDashboards(projectId, {
        limit: params.limit,
        offset,
        name: params.q && params.q.length > 0 ? params.q : undefined,
      })
    },
    staleTime: 30 * 1000,
  })

export const dashboardDetailQueryOptions = (
  projectId: string,
  dashboardId: string,
) =>
  queryOptions({
    queryKey: dashboardsKeys.detail(dashboardId),
    queryFn: (): Promise<Dashboard> =>
      getDashboardById(projectId, dashboardId),
    staleTime: 30 * 1000,
  })
