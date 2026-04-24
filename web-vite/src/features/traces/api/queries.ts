import { queryOptions } from '@tanstack/react-query'
import { rawFetch } from '@/lib/api/client'
import type { Span, TraceDetail, TraceListResponse } from './types'

// TkDodo-style hierarchical query keys.
export const tracesKeys = {
  all: ['traces'] as const,
  lists: () => [...tracesKeys.all, 'list'] as const,
  list: (projectId: string, params: TraceListParams) =>
    [...tracesKeys.lists(), projectId, params] as const,
  details: () => [...tracesKeys.all, 'detail'] as const,
  detail: (traceId: string) => [...tracesKeys.details(), traceId] as const,
  spansOf: (traceId: string) =>
    [...tracesKeys.detail(traceId), 'spans'] as const,
} as const

// Relative time ranges surfaced in the filter bar. Translated to a
// concrete `start_time` (unix seconds) in the query fetcher; `all`
// means "do not send the bound" and lets the backend window run.
export type TraceRange = '24h' | '7d' | '30d' | 'all'
export type TraceStatus = 'ok' | 'error' | 'unset'

export interface TraceListParams {
  page: number
  limit: number
  q?: string
  status?: TraceStatus
  range: TraceRange
  model?: string
}

// Convert a relative range to a unix-seconds lower bound. Upper bound
// is unbounded (backend defaults to now).
function rangeToStartTimeSec(range: TraceRange): number | undefined {
  if (range === 'all') return undefined
  const nowMs = Date.now()
  const windowMs =
    range === '24h'
      ? 24 * 60 * 60 * 1000
      : range === '7d'
        ? 7 * 24 * 60 * 60 * 1000
        : 30 * 24 * 60 * 60 * 1000
  return Math.floor((nowMs - windowMs) / 1000)
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
        search.set('search', params.q)
      }
      if (params.status) {
        search.set('status', params.status)
      }
      const startSec = rangeToStartTimeSec(params.range)
      if (startSec !== undefined) {
        search.set('start_time', String(startSec))
      }
      if (params.model && params.model.length > 0) {
        search.set('model_name', params.model)
      }
      const resp = await rawFetch(`/api/v1/traces?${search.toString()}`, {
        method: 'GET',
      })
      return (await resp.json()) as TraceListResponse
    },
    staleTime: 30 * 1000,
  })

export const traceDetailQueryOptions = (traceId: string) =>
  queryOptions({
    queryKey: tracesKeys.detail(traceId),
    queryFn: async () => {
      const resp = await rawFetch(
        `/api/v1/traces/${encodeURIComponent(traceId)}`,
        { method: 'GET' },
      )
      return (await resp.json()) as TraceDetail
    },
    staleTime: 30 * 1000,
  })

export const traceSpansQueryOptions = (traceId: string) =>
  queryOptions({
    queryKey: tracesKeys.spansOf(traceId),
    queryFn: async () => {
      const resp = await rawFetch(
        `/api/v1/traces/${encodeURIComponent(traceId)}/spans`,
        { method: 'GET' },
      )
      return (await resp.json()) as Span[]
    },
    staleTime: 30 * 1000,
  })
