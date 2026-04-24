import { queryOptions } from '@tanstack/react-query'
import { rawFetch } from '@/lib/api/client'
import type { DatasetListResponse } from './types'

// TkDodo-style hierarchical query keys. `list(projectId, params)`
// invalidates cleanly via `lists()` once the detail-view port adds
// mutations that should bust the list cache.
export const datasetsKeys = {
  all: ['datasets'] as const,
  lists: () => [...datasetsKeys.all, 'list'] as const,
  list: (projectId: string, params: DatasetListParams) =>
    [...datasetsKeys.lists(), projectId, params] as const,
  details: () => [...datasetsKeys.all, 'detail'] as const,
  detail: (datasetId: string) => [...datasetsKeys.details(), datasetId] as const,
} as const

export interface DatasetListParams {
  page: number
  limit: number
  q?: string
}

export const datasetListQueryOptions = (
  projectId: string,
  params: DatasetListParams,
) =>
  queryOptions({
    queryKey: datasetsKeys.list(projectId, params),
    queryFn: async () => {
      const search = new URLSearchParams()
      search.set('page', String(params.page))
      search.set('limit', String(params.limit))
      if (params.q && params.q.length > 0) {
        // Backend accepts `search` as the free-text query param; keep
        // the user-facing alias `q` in the URL for familiarity.
        search.set('search', params.q)
      }
      const resp = await rawFetch(
        `/api/v1/projects/${projectId}/datasets?${search.toString()}`,
        { method: 'GET' },
      )
      return (await resp.json()) as DatasetListResponse
    },
    // Datasets change infrequently (manual create/edit) — 30s keeps
    // in-tab navigation snappy without masking fresh changes.
    staleTime: 30 * 1000,
  })
