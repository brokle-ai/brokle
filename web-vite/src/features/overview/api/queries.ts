import { queryOptions } from '@tanstack/react-query'
import { rawFetch } from '@/lib/api/client'
import type { OverviewResponse, OverviewTimeRange } from './types'

export const overviewKeys = {
  all: ['overview'] as const,
  project: (projectId: string, timeRange: OverviewTimeRange) =>
    [...overviewKeys.all, projectId, timeRange] as const,
} as const

export interface OverviewQueryParams {
  timeRange?: OverviewTimeRange
}

const DEFAULT_TIME_RANGE: OverviewTimeRange = '24h'

export const overviewQueryOptions = (
  projectId: string,
  params: OverviewQueryParams = {},
) => {
  const timeRange = params.timeRange ?? DEFAULT_TIME_RANGE
  return queryOptions({
    queryKey: overviewKeys.project(projectId, timeRange),
    queryFn: async () => {
      const search = new URLSearchParams()
      search.set('time_range', timeRange)
      const resp = await rawFetch(
        `/api/v1/projects/${projectId}/overview?${search.toString()}`,
        { method: 'GET' },
      )
      return (await resp.json()) as OverviewResponse
    },
    // The backend computes trends over the previous period, so short
    // staleTime keeps the homepage responsive without hammering
    // ClickHouse on every tab focus.
    staleTime: 30 * 1000,
  })
}
