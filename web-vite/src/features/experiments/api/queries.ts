import { queryOptions } from '@tanstack/react-query'
import { rawFetch } from '@/lib/api/client'
import type { ExperimentListResponse } from './types'

export const experimentsKeys = {
  all: ['experiments'] as const,
  lists: () => [...experimentsKeys.all, 'list'] as const,
  list: (projectId: string, params: ExperimentListParams) =>
    [...experimentsKeys.lists(), projectId, params] as const,
  details: () => [...experimentsKeys.all, 'detail'] as const,
  detail: (experimentId: string) =>
    [...experimentsKeys.details(), experimentId] as const,
} as const

export interface ExperimentListParams {
  page: number
  limit: number
  q?: string
}

export const experimentListQueryOptions = (
  projectId: string,
  params: ExperimentListParams,
) =>
  queryOptions({
    queryKey: experimentsKeys.list(projectId, params),
    queryFn: async () => {
      const search = new URLSearchParams()
      search.set('page', String(params.page))
      search.set('limit', String(params.limit))
      if (params.q && params.q.length > 0) {
        search.set('search', params.q)
      }
      const resp = await rawFetch(
        `/api/v1/projects/${projectId}/experiments?${search.toString()}`,
        { method: 'GET' },
      )
      return (await resp.json()) as ExperimentListResponse
    },
    // Experiments transition (running → completed) — keep window short
    // so the list reflects progress without aggressive refetch.
    staleTime: 15 * 1000,
  })
