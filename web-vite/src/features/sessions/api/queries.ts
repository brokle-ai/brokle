import { queryOptions } from '@tanstack/react-query'
import { rawFetch } from '@/lib/api/client'
import type { SessionListResponse } from './types'

export const sessionsKeys = {
  all: ['sessions'] as const,
  lists: () => [...sessionsKeys.all, 'list'] as const,
  list: (projectId: string, params: SessionListParams) =>
    [...sessionsKeys.lists(), projectId, params] as const,
} as const

export interface SessionListParams {
  page: number
  limit: number
  q?: string
}

export const sessionListQueryOptions = (
  projectId: string,
  params: SessionListParams,
) =>
  queryOptions({
    queryKey: sessionsKeys.list(projectId, params),
    queryFn: async () => {
      const search = new URLSearchParams()
      search.set('page', String(params.page))
      search.set('limit', String(params.limit))
      if (params.q && params.q.length > 0) {
        search.set('search', params.q)
      }
      const resp = await rawFetch(
        `/api/v1/projects/${projectId}/sessions?${search.toString()}`,
        { method: 'GET' },
      )
      return (await resp.json()) as SessionListResponse
    },
    // Sessions are aggregations over append-only trace data; 30s
    // matches the traces list so cross-tab navigation feels coherent.
    staleTime: 30 * 1000,
  })
