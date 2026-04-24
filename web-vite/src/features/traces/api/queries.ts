import { queryOptions } from '@tanstack/react-query'
import { rawFetch } from '@/lib/api/client'
import type { TraceListResponse } from './types'

// TkDodo-style hierarchical query keys. `list(projectId, params)`
// invalidates cleanly via `lists()` when the detail-view port adds
// mutations that should bust the list cache.
export const tracesKeys = {
  all: ['traces'] as const,
  lists: () => [...tracesKeys.all, 'list'] as const,
  list: (projectId: string, params: TraceListParams) =>
    [...tracesKeys.lists(), projectId, params] as const,
  details: () => [...tracesKeys.all, 'detail'] as const,
  detail: (traceId: string) => [...tracesKeys.details(), traceId] as const,
} as const

export interface TraceListParams {
  page: number
  limit: number
  q?: string
}

export const traceListQueryOptions = (
  projectId: string,
  params: TraceListParams,
) =>
  queryOptions({
    queryKey: tracesKeys.list(projectId, params),
    queryFn: async () => {
      const search = new URLSearchParams()
      search.set('project_id', projectId)
      search.set('page', String(params.page))
      search.set('limit', String(params.limit))
      if (params.q && params.q.length > 0) {
        // Backend accepts `search` as the free-text query param; keep
        // the user-facing alias `q` in the URL for familiarity.
        search.set('search', params.q)
      }
      const resp = await rawFetch(`/v1/traces?${search.toString()}`, {
        method: 'GET',
      })
      return (await resp.json()) as TraceListResponse
    },
    // Traces are append-only observability data — a 30s window keeps
    // in-tab navigation snappy without hiding fresh ingestion for
    // more than half a minute.
    staleTime: 30 * 1000,
  })
