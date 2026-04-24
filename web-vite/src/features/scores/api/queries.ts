import { queryOptions } from '@tanstack/react-query'
import { rawFetch } from '@/lib/api/client'
import type { ScoreDataType, ScoreListResponse, ScoreSource } from './types'

export const scoresKeys = {
  all: ['scores'] as const,
  lists: () => [...scoresKeys.all, 'list'] as const,
  list: (projectId: string, params: ScoreListParams) =>
    [...scoresKeys.lists(), projectId, params] as const,
  details: () => [...scoresKeys.all, 'detail'] as const,
  detail: (scoreId: string) => [...scoresKeys.details(), scoreId] as const,
} as const

export interface ScoreListParams {
  page: number
  limit: number
  name?: string
  source?: ScoreSource
  type?: ScoreDataType
  traceId?: string
  spanId?: string
}

export const scoreListQueryOptions = (
  projectId: string,
  params: ScoreListParams,
) =>
  queryOptions({
    queryKey: scoresKeys.list(projectId, params),
    queryFn: async () => {
      const search = new URLSearchParams()
      search.set('page', String(params.page))
      search.set('limit', String(params.limit))
      if (params.name) search.set('name', params.name)
      if (params.source) search.set('source', params.source)
      if (params.type) search.set('type', params.type)
      if (params.traceId) search.set('trace_id', params.traceId)
      if (params.spanId) search.set('span_id', params.spanId)
      const resp = await rawFetch(
        `/api/v1/projects/${projectId}/scores?${search.toString()}`,
        { method: 'GET' },
      )
      return (await resp.json()) as ScoreListResponse
    },
    staleTime: 30 * 1000,
  })
