import { queryOptions } from '@tanstack/react-query'
import { rawFetch } from '@/lib/api/client'
import type {
  CreateScoreConfigRequest,
  ScoreAnalyticsData,
  ScoreAnalyticsParams,
  ScoreConfig,
  ScoreConfigListResponse,
  ScoreDataType,
  ScoreListResponse,
  ScoreSource,
  UpdateScoreConfigRequest,
} from './types'

// Deletes a single score by project + trace + score id. Mirrors the
// backend DELETE /api/v1/projects/{projectId}/traces/{id}/scores/{scoreId}
// (observability handler, delete-trace-score). Returns nothing on
// success (204).
export async function deleteTraceScore(
  projectId: string,
  traceId: string,
  scoreId: string,
): Promise<void> {
  await rawFetch(
    `/api/v1/projects/${encodeURIComponent(projectId)}/traces/${encodeURIComponent(traceId)}/scores/${encodeURIComponent(scoreId)}`,
    { method: 'DELETE' },
  )
}

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
  configsListPaged: (projectId: string, page: number, limit: number) =>
    [...scoresKeys.configs(), projectId, page, limit] as const,
  analytics: () => [...scoresKeys.all, 'analytics'] as const,
  analyticsData: (projectId: string, params: ScoreAnalyticsParams) =>
    [...scoresKeys.analytics(), 'data', projectId, params] as const,
  analyticsNames: (projectId: string) =>
    [...scoresKeys.analytics(), 'names', projectId] as const,
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

// Paginated score-configs query for the configs management surface.
// `scoreConfigsQueryOptions` deliberately fetches a high cap for
// embedded annotation widgets; the configs CRUD page wants real
// pagination so we keep a separate keyed query.
export const scoreConfigsPagedQueryOptions = (
  projectId: string,
  page: number,
  limit: number,
) =>
  queryOptions({
    queryKey: scoresKeys.configsListPaged(projectId, page, limit),
    queryFn: async () => {
      const resp = await rawFetch(
        `/api/v1/projects/${projectId}/score-configs?page=${page}&limit=${limit}`,
        { method: 'GET' },
      )
      return (await resp.json()) as ScoreConfigListResponse
    },
    staleTime: 30 * 1000,
  })

export async function createScoreConfig(
  projectId: string,
  data: CreateScoreConfigRequest,
): Promise<ScoreConfig> {
  const resp = await rawFetch(`/api/v1/projects/${projectId}/score-configs`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  return (await resp.json()) as ScoreConfig
}

export async function updateScoreConfig(
  projectId: string,
  configId: string,
  data: UpdateScoreConfigRequest,
): Promise<ScoreConfig> {
  const resp = await rawFetch(
    `/api/v1/projects/${projectId}/score-configs/${configId}`,
    {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    },
  )
  return (await resp.json()) as ScoreConfig
}

export async function deleteScoreConfig(
  projectId: string,
  configId: string,
): Promise<void> {
  await rawFetch(`/api/v1/projects/${projectId}/score-configs/${configId}`, {
    method: 'DELETE',
  })
}

// Score analytics — primary aggregation surface for the analytics
// dashboard. The handler computes statistics, time series, and (when
// `compare_score_name` is supplied) a paired heatmap + comparison
// metrics in one round trip; we mirror that single-call shape.
export const scoreAnalyticsQueryOptions = (
  projectId: string,
  params: ScoreAnalyticsParams,
) =>
  queryOptions({
    queryKey: scoresKeys.analyticsData(projectId, params),
    queryFn: async () => {
      const search = new URLSearchParams()
      search.set('score_name', params.score_name)
      if (params.compare_score_name)
        search.set('compare_score_name', params.compare_score_name)
      if (params.from_timestamp)
        search.set('from_timestamp', params.from_timestamp)
      if (params.to_timestamp) search.set('to_timestamp', params.to_timestamp)
      if (params.interval) search.set('interval', params.interval)
      const resp = await rawFetch(
        `/api/v1/projects/${projectId}/scores/analytics?${search.toString()}`,
        { method: 'GET' },
      )
      return (await resp.json()) as ScoreAnalyticsData
    },
    staleTime: 60 * 1000,
  })

// The names endpoint feeds the score selector on the analytics page.
// Returns a flat string array of distinct score names recorded in the
// project. Cached longer than analytics — the catalog churns less than
// the underlying values.
export const scoreNamesQueryOptions = (projectId: string) =>
  queryOptions({
    queryKey: scoresKeys.analyticsNames(projectId),
    queryFn: async () => {
      const resp = await rawFetch(
        `/api/v1/projects/${projectId}/scores/names`,
        { method: 'GET' },
      )
      return (await resp.json()) as string[]
    },
    staleTime: 60 * 1000,
  })
