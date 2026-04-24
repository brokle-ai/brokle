import { queryOptions } from '@tanstack/react-query'
import { rawFetch } from '@/lib/api/client'
import type { EvaluatorListResponse } from './types'

export const evaluatorsKeys = {
  all: ['evaluators'] as const,
  lists: () => [...evaluatorsKeys.all, 'list'] as const,
  list: (projectId: string, params: EvaluatorListParams) =>
    [...evaluatorsKeys.lists(), projectId, params] as const,
  details: () => [...evaluatorsKeys.all, 'detail'] as const,
  detail: (evaluatorId: string) =>
    [...evaluatorsKeys.details(), evaluatorId] as const,
} as const

export interface EvaluatorListParams {
  page: number
  limit: number
  q?: string
}

export const evaluatorListQueryOptions = (
  projectId: string,
  params: EvaluatorListParams,
) =>
  queryOptions({
    queryKey: evaluatorsKeys.list(projectId, params),
    queryFn: async () => {
      const search = new URLSearchParams()
      search.set('page', String(params.page))
      search.set('limit', String(params.limit))
      if (params.q && params.q.length > 0) {
        search.set('search', params.q)
      }
      const resp = await rawFetch(
        `/api/v1/projects/${projectId}/evaluators?${search.toString()}`,
        { method: 'GET' },
      )
      return (await resp.json()) as EvaluatorListResponse
    },
    staleTime: 30 * 1000,
  })
