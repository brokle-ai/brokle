import { queryOptions } from '@tanstack/react-query'
import { rawFetch } from '@/lib/api/client'
import { NotFoundError } from '@/lib/api/errors'
import type { SessionDetail, SessionListResponse } from './types'

export const sessionsKeys = {
  all: ['sessions'] as const,
  lists: () => [...sessionsKeys.all, 'list'] as const,
  list: (projectId: string, params: SessionListParams) =>
    [...sessionsKeys.lists(), projectId, params] as const,
  details: () => [...sessionsKeys.all, 'detail'] as const,
  detail: (projectId: string, sessionId: string) =>
    [...sessionsKeys.details(), projectId, sessionId] as const,
} as const

export interface SessionListParams {
  page: number
  limit: number
  q?: string
}

export const sessionListQueryOptions = (
  projectId: string,
  params: SessionListParams,
) =>
  queryOptions({
    queryKey: sessionsKeys.list(projectId, params),
    queryFn: async () => {
      const search = new URLSearchParams()
      search.set('page', String(params.page))
      search.set('limit', String(params.limit))
      if (params.q && params.q.length > 0) {
        search.set('search', params.q)
      }
      const resp = await rawFetch(
        `/api/v1/projects/${projectId}/sessions?${search.toString()}`,
        { method: 'GET' },
      )
      return (await resp.json()) as SessionListResponse
    },
    // Sessions are aggregations over append-only trace data; 30s
    // matches the traces list so cross-tab navigation feels coherent.
    staleTime: 30 * 1000,
  })

// Session detail is derived: the backend has no dedicated
// `/sessions/{id}` endpoint — sessions are a ClickHouse aggregation,
// not a stored entity. We call the list endpoint with a
// `search={sessionId}` filter (partial match on session_id) and pick
// the exact row. 404 if the aggregation returns no matching row.
export const sessionDetailQueryOptions = (
  projectId: string,
  sessionId: string,
) =>
  queryOptions({
    queryKey: sessionsKeys.detail(projectId, sessionId),
    queryFn: async (): Promise<SessionDetail> => {
      const search = new URLSearchParams()
      search.set('page', '1')
      search.set('limit', '10')
      search.set('search', sessionId)
      const resp = await rawFetch(
        `/api/v1/projects/${projectId}/sessions?${search.toString()}`,
        { method: 'GET' },
      )
      const body = (await resp.json()) as SessionListResponse
      const hit = body.data.find((row) => row.session_id === sessionId)
      if (!hit) {
        throw new NotFoundError(404, {
          error: {
            type: 'not_found',
            message: `Session ${sessionId} not found`,
          },
        })
      }
      return hit
    },
    staleTime: 30 * 1000,
  })
