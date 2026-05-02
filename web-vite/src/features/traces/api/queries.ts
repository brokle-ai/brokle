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
  detail: (projectId: string, traceId: string) =>
    [...tracesKeys.details(), projectId, traceId] as const,
  spansOf: (projectId: string, traceId: string) =>
    [...tracesKeys.detail(projectId, traceId), 'spans'] as const,
  scoresOf: (traceId: string, projectId: string) =>
    [...tracesKeys.detail(projectId, traceId), 'scores'] as const,
  annotations: (traceId: string, projectId: string) =>
    [...tracesKeys.detail(projectId, traceId), 'annotations'] as const,
  comments: (traceId: string, projectId: string) =>
    [...tracesKeys.detail(projectId, traceId), 'comments'] as const,
} as const

// Relative time ranges surfaced in the filter bar. Translated to a
// concrete `start_time` (unix seconds) in the query fetcher; `all`
// means "do not send the bound" and lets the backend window run.
export type TraceRange = '24h' | '7d' | '30d' | 'all'
export type TraceStatus = 'ok' | 'error' | 'unset'
export type TraceSortKey =
  | 'duration'
  | 'model_name'
  | 'total_cost'
  | 'total_tokens'
  | 'span_count'
  | 'start_time'
export type TraceSortDir = 'asc' | 'desc'

export interface TraceListParams {
  page: number
  limit: number
  q?: string
  status?: TraceStatus
  range: TraceRange
  model?: string
  // Server-side sort. Backend `ListTracesInput` accepts `sort_by` +
  // `sort_dir`; we forward both when set so the table column header
  // sort state survives reloads / shared URLs. Both must be set
  // together; either missing falls back to the default `start_time
  // desc`.
  sortBy?: TraceSortKey
  sortDir?: TraceSortDir
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
      if (params.sortBy && params.sortDir) {
        search.set('sort_by', params.sortBy)
        search.set('sort_dir', params.sortDir)
      }
      const resp = await rawFetch(
        `/api/v1/projects/${encodeURIComponent(projectId)}/traces?${search.toString()}`,
        { method: 'GET' },
      )
      return (await resp.json()) as TraceListResponse
    },
    staleTime: 30 * 1000,
  })

export const traceDetailQueryOptions = (projectId: string, traceId: string) =>
  queryOptions({
    queryKey: tracesKeys.detail(projectId, traceId),
    queryFn: async () => {
      const resp = await rawFetch(
        `/api/v1/projects/${encodeURIComponent(projectId)}/traces/${encodeURIComponent(traceId)}`,
        { method: 'GET' },
      )
      return (await resp.json()) as TraceDetail
    },
    staleTime: 30 * 1000,
  })

export const traceSpansQueryOptions = (projectId: string, traceId: string) =>
  queryOptions({
    queryKey: tracesKeys.spansOf(projectId, traceId),
    queryFn: async () => {
      const resp = await rawFetch(
        `/api/v1/projects/${encodeURIComponent(projectId)}/traces/${encodeURIComponent(traceId)}/spans`,
        { method: 'GET' },
      )
      return (await resp.json()) as Span[]
    },
    staleTime: 30 * 1000,
  })

// ---------------- annotations + comments ------------------------------
//
// Both endpoints are nested under
// `/api/v1/projects/{projectId}/traces/{traceId}` (tenant scoping is
// implicit in the path — the trace lookup verifies the trace belongs
// to the addressed project). Mutation functions are kept thin so
// callers compose them with `useMutation` + their own invalidation.
// The drawer components attach the invalidation logic.

export const traceAnnotationsQueryOptions = (
  projectId: string,
  traceId: string,
) =>
  queryOptions({
    queryKey: tracesKeys.annotations(traceId, projectId),
    queryFn: async () => {
      const resp = await rawFetch(
        `/api/v1/projects/${encodeURIComponent(projectId)}/traces/${encodeURIComponent(traceId)}/scores`,
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
  const resp = await rawFetch(
    `/api/v1/projects/${encodeURIComponent(projectId)}/traces/${encodeURIComponent(traceId)}/scores`,
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
  await rawFetch(
    `/api/v1/projects/${encodeURIComponent(projectId)}/traces/${encodeURIComponent(traceId)}/scores/${encodeURIComponent(scoreId)}`,
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
      const resp = await rawFetch(
        `/api/v1/projects/${encodeURIComponent(projectId)}/traces/${encodeURIComponent(traceId)}/comments`,
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
  const resp = await rawFetch(
    `/api/v1/projects/${encodeURIComponent(projectId)}/traces/${encodeURIComponent(traceId)}/comments`,
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
  const resp = await rawFetch(
    `/api/v1/projects/${encodeURIComponent(projectId)}/traces/${encodeURIComponent(traceId)}/comments/${encodeURIComponent(parentId)}/replies`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    },
  )
  return (await resp.json()) as TraceComment
}

// ---------------- trace tags / bookmark -------------------------------
//
// Shapes mirror observability.UpdateTraceTagsRequest + the inline body
// on UpdateTraceBookmarkInput. Both endpoints return the updated
// field echoed back — we coerce to the raw value here because the
// caller only needs the new state for cache reconciliation.

export async function updateTraceTags(
  projectId: string,
  traceId: string,
  tags: string[],
): Promise<string[]> {
  const resp = await rawFetch(
    `/api/v1/projects/${encodeURIComponent(projectId)}/traces/${encodeURIComponent(traceId)}/tags`,
    {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ tags }),
    },
  )
  const result = (await resp.json()) as { tags?: string[] }
  return Array.isArray(result.tags) ? result.tags : tags
}

export async function updateTraceBookmark(
  projectId: string,
  traceId: string,
  bookmarked: boolean,
): Promise<boolean> {
  const resp = await rawFetch(
    `/api/v1/projects/${encodeURIComponent(projectId)}/traces/${encodeURIComponent(traceId)}/bookmark`,
    {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ bookmarked }),
    },
  )
  const result = (await resp.json()) as { bookmarked?: boolean }
  return typeof result.bookmarked === 'boolean' ? result.bookmarked : bookmarked
}

export async function toggleTraceCommentReaction(
  projectId: string,
  traceId: string,
  commentId: string,
  data: ToggleReactionRequest,
): Promise<TraceReaction[]> {
  const resp = await rawFetch(
    `/api/v1/projects/${encodeURIComponent(projectId)}/traces/${encodeURIComponent(traceId)}/comments/${encodeURIComponent(commentId)}/reactions`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    },
  )
  const result = (await resp.json()) as TraceReaction[] | null
  return Array.isArray(result) ? result : []
}

// ---------------- per-trace scores (for span-row badges) --------------
//
// Lightweight query that pulls all scores for a trace via the scores
// list endpoint scoped by `trace_id`. Used to drive the inline
// `<ScoreTagList>` on each span row in the span tree without firing a
// per-row request. Backend endpoint: `GET /api/v1/projects/{projectId}/scores`.
export interface TraceScoreItem {
  id: string
  project_id: string
  trace_id?: string
  span_id?: string
  name: string
  value?: number
  string_value?: string
  type: 'NUMERIC' | 'CATEGORICAL' | 'BOOLEAN'
  source: string
  reason?: string
  created_by?: string
  timestamp: string
}

interface ScoresListResponse {
  data: TraceScoreItem[]
  pagination: { page: number; limit: number; total: number; total_pages: number; has_next: boolean; has_prev: boolean }
}

export const traceScoresQueryOptions = (projectId: string, traceId: string) =>
  queryOptions({
    queryKey: tracesKeys.scoresOf(traceId, projectId),
    queryFn: async () => {
      const search = new URLSearchParams({
        page: '1',
        limit: '200',
        trace_id: traceId,
      })
      const resp = await rawFetch(
        `/api/v1/projects/${encodeURIComponent(projectId)}/scores?${search.toString()}`,
        { method: 'GET' },
      )
      const body = (await resp.json()) as ScoresListResponse
      return Array.isArray(body.data) ? body.data : []
    },
    staleTime: 30 * 1000,
  })

// ---------------- filter presets --------------------------------------
//
// Backed by the dashboard `filter-presets` Huma operations under
// `/api/v1/projects/{projectId}/filter-presets`. The drawer UI uses
// this API to persist saved filters across sessions; the local-storage
// fallback path activates only when the request fails (e.g. during
// initial dev with the endpoint disabled). Filters here are the same
// `FilterCondition` shape the shared FilterBuilder emits.

import type { FilterCondition } from '@/components/shared/filter-builder'

export interface FilterPreset {
  id: string
  project_id: string
  name: string
  description?: string
  table_name: 'traces' | 'spans'
  filters: FilterCondition[]
  search_query?: string
  search_types?: string[]
  is_public: boolean
  created_by?: string
  created_at: string
  updated_at: string
}

export interface CreateFilterPresetRequest {
  name: string
  description?: string
  table_name: 'traces' | 'spans'
  filters: FilterCondition[]
  search_query?: string
  search_types?: string[]
  is_public?: boolean
}

export interface UpdateFilterPresetRequest {
  name?: string
  description?: string
  filters?: FilterCondition[]
  search_query?: string
  search_types?: string[]
  is_public?: boolean
}

export const filterPresetsKeys = {
  all: ['filter-presets'] as const,
  list: (projectId: string, tableName: 'traces' | 'spans') =>
    [...filterPresetsKeys.all, projectId, tableName] as const,
} as const

export const filterPresetsQueryOptions = (
  projectId: string,
  tableName: 'traces' | 'spans',
) =>
  queryOptions({
    queryKey: filterPresetsKeys.list(projectId, tableName),
    queryFn: async () => {
      const search = new URLSearchParams({
        project_id: projectId,
        table_name: tableName,
        include_public: 'true',
      })
      const resp = await rawFetch(
        `/api/v1/projects/${encodeURIComponent(projectId)}/filter-presets?${search.toString()}`,
        { method: 'GET' },
      )
      const body = (await resp.json()) as FilterPreset[] | null
      return Array.isArray(body) ? body : []
    },
    staleTime: 60 * 1000,
  })

export async function createFilterPreset(
  projectId: string,
  data: CreateFilterPresetRequest,
): Promise<FilterPreset> {
  const resp = await rawFetch(
    `/api/v1/projects/${encodeURIComponent(projectId)}/filter-presets`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    },
  )
  return (await resp.json()) as FilterPreset
}

export async function updateFilterPreset(
  projectId: string,
  presetId: string,
  data: UpdateFilterPresetRequest,
): Promise<FilterPreset> {
  const resp = await rawFetch(
    `/api/v1/projects/${encodeURIComponent(projectId)}/filter-presets/${encodeURIComponent(presetId)}`,
    {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    },
  )
  return (await resp.json()) as FilterPreset
}

export async function deleteFilterPreset(
  projectId: string,
  presetId: string,
): Promise<void> {
  await rawFetch(
    `/api/v1/projects/${encodeURIComponent(projectId)}/filter-presets/${encodeURIComponent(presetId)}`,
    { method: 'DELETE' },
  )
}

// ---------------- trace deletion --------------------------------------
//
// Single-row delete is wired against the `delete-trace` Huma op
// (`DELETE /api/v1/projects/{projectId}/traces/{id}`). The endpoint
// returns HTTP 204 — no body to parse. Bulk delete is intentionally
// not implemented here: the backend has no batch-delete operation, so
// the list view's bulk-delete dialog stays in its "deferred" state
// until a batch-delete operation lands.
export async function deleteTrace(
  projectId: string,
  traceId: string,
): Promise<void> {
  await rawFetch(
    `/api/v1/projects/${encodeURIComponent(projectId)}/traces/${encodeURIComponent(traceId)}`,
    { method: 'DELETE' },
  )
}
