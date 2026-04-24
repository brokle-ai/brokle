import { queryOptions } from '@tanstack/react-query'
import { rawFetch } from '@/lib/api/client'
import type {
  CreateTraceAnnotationRequest,
  CreateTraceCommentRequest,
  Span,
  ToggleReactionRequest,
  TraceAnnotation,
  TraceComment,
  TraceCommentsListResponse,
  TraceDetail,
  TraceListResponse,
  TraceReaction,
} from './types'

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
  annotations: (traceId: string, projectId: string) =>
    [...tracesKeys.detail(traceId), 'annotations', projectId] as const,
  comments: (traceId: string, projectId: string) =>
    [...tracesKeys.detail(traceId), 'comments', projectId] as const,
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
  // Optional session scope. When present, the backend filters to traces
  // whose `session_id` attribute equals this value (see
  // `internal/transport/http/handlers/observability/dashboard_types.go`
  // `ListTracesInput.SessionID`). Used by the session-detail route to
  // reuse <TracesTable> without duplicating filter state.
  sessionId?: string
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
      if (params.sessionId && params.sessionId.length > 0) {
        search.set('session_id', params.sessionId)
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

// ---------------- annotations + comments ------------------------------
//
// Both endpoints live under `/api/v1/traces/{traceId}` and require
// `project_id` as a query param (tenant scoping — the trace lookup
// also verifies `project_id` matches the trace's project). Mutation
// functions are kept thin so callers compose them with `useMutation`
// + their own invalidation. The drawer components attach the
// invalidation logic.

export const traceAnnotationsQueryOptions = (
  projectId: string,
  traceId: string,
) =>
  queryOptions({
    queryKey: tracesKeys.annotations(traceId, projectId),
    queryFn: async () => {
      const params = new URLSearchParams({ project_id: projectId })
      const resp = await rawFetch(
        `/api/v1/traces/${encodeURIComponent(traceId)}/scores?${params.toString()}`,
        { method: 'GET' },
      )
      const data = (await resp.json()) as TraceAnnotation[] | null
      return Array.isArray(data) ? data : []
    },
    staleTime: 20 * 1000,
  })

export async function createTraceAnnotation(
  projectId: string,
  traceId: string,
  data: CreateTraceAnnotationRequest,
): Promise<TraceAnnotation> {
  const params = new URLSearchParams({ project_id: projectId })
  const resp = await rawFetch(
    `/api/v1/traces/${encodeURIComponent(traceId)}/scores?${params.toString()}`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    },
  )
  return (await resp.json()) as TraceAnnotation
}

export async function deleteTraceAnnotation(
  projectId: string,
  traceId: string,
  scoreId: string,
): Promise<void> {
  const params = new URLSearchParams({ project_id: projectId })
  await rawFetch(
    `/api/v1/traces/${encodeURIComponent(traceId)}/scores/${scoreId}?${params.toString()}`,
    { method: 'DELETE' },
  )
}

export const traceCommentsQueryOptions = (
  projectId: string,
  traceId: string,
) =>
  queryOptions({
    queryKey: tracesKeys.comments(traceId, projectId),
    queryFn: async () => {
      const params = new URLSearchParams({ project_id: projectId })
      const resp = await rawFetch(
        `/api/v1/traces/${encodeURIComponent(traceId)}/comments?${params.toString()}`,
        { method: 'GET' },
      )
      return (await resp.json()) as TraceCommentsListResponse
    },
    staleTime: 15 * 1000,
  })

export async function createTraceComment(
  projectId: string,
  traceId: string,
  data: CreateTraceCommentRequest,
): Promise<TraceComment> {
  const params = new URLSearchParams({ project_id: projectId })
  const resp = await rawFetch(
    `/api/v1/traces/${encodeURIComponent(traceId)}/comments?${params.toString()}`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    },
  )
  return (await resp.json()) as TraceComment
}

export async function createTraceCommentReply(
  projectId: string,
  traceId: string,
  parentId: string,
  data: CreateTraceCommentRequest,
): Promise<TraceComment> {
  const params = new URLSearchParams({ project_id: projectId })
  const resp = await rawFetch(
    `/api/v1/traces/${encodeURIComponent(traceId)}/comments/${parentId}/replies?${params.toString()}`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    },
  )
  return (await resp.json()) as TraceComment
}

export async function toggleTraceCommentReaction(
  projectId: string,
  traceId: string,
  commentId: string,
  data: ToggleReactionRequest,
): Promise<TraceReaction[]> {
  const params = new URLSearchParams({ project_id: projectId })
  const resp = await rawFetch(
    `/api/v1/traces/${encodeURIComponent(traceId)}/comments/${commentId}/reactions?${params.toString()}`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    },
  )
  const result = (await resp.json()) as TraceReaction[] | null
  return Array.isArray(result) ? result : []
}
