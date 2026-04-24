import { queryOptions } from '@tanstack/react-query'
import { rawFetch } from '@/lib/api/client'
import type {
  ScoreConfigListResponse,
  ScoreDataType,
  ScoreListResponse,
  ScoreSource,
} from './types'

export const scoresKeys = {
  all: ['scores'] as const,
  lists: () => [...scoresKeys.all, 'list'] as const,
  list: (projectId: string, params: ScoreListParams) =>
    [...scoresKeys.lists(), projectId, params] as const,
  details: () => [...scoresKeys.all, 'detail'] as const,
  detail: (scoreId: string) => [...scoresKeys.details(), scoreId] as const,
  configs: () => [...scoresKeys.all, 'configs'] as const,
  configsList: (projectId: string) =>
    [...scoresKeys.configs(), projectId] as const,
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

// Fetches the full score-config catalog for a project. The queue
// review UI and the trace annotations drawer both render score inputs
// driven by this list — NUMERIC fields respect `min_value`/`max_value`,
// CATEGORICAL dispatches to a select of `categories`, BOOLEAN is a
// yes/no toggle. Limit is hard-coded high because projects rarely
// exceed a handful of configs; if that ever becomes a real concern we
// can paginate at the call site.
export const scoreConfigsQueryOptions = (projectId: string) =>
  queryOptions({
    queryKey: scoresKeys.configsList(projectId),
    queryFn: async () => {
      const resp = await rawFetch(
        `/api/v1/projects/${projectId}/score-configs?page=1&limit=200`,
        { method: 'GET' },
      )
      return (await resp.json()) as ScoreConfigListResponse
    },
    staleTime: 60 * 1000,
  })
