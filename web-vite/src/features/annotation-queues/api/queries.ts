import { queryOptions } from '@tanstack/react-query'
import { rawFetch } from '@/lib/api/client'
import type { QueueListResponse, QueueStatus } from './types'

export const queuesKeys = {
  all: ['annotation-queues'] as const,
  lists: () => [...queuesKeys.all, 'list'] as const,
  list: (projectId: string, params: QueueListParams) =>
    [...queuesKeys.lists(), projectId, params] as const,
  details: () => [...queuesKeys.all, 'detail'] as const,
  detail: (queueId: string) => [...queuesKeys.details(), queueId] as const,
} as const

export interface QueueListParams {
  page: number
  limit: number
  status?: QueueStatus
  q?: string
}

export const queueListQueryOptions = (
  projectId: string,
  params: QueueListParams,
) =>
  queryOptions({
    queryKey: queuesKeys.list(projectId, params),
    queryFn: async () => {
      const search = new URLSearchParams()
      search.set('page', String(params.page))
      search.set('limit', String(params.limit))
      if (params.status) search.set('status', params.status)
      if (params.q && params.q.length > 0) search.set('search', params.q)
      const resp = await rawFetch(
        `/api/v1/projects/${projectId}/annotation-queues?${search.toString()}`,
        { method: 'GET' },
      )
      return (await resp.json()) as QueueListResponse
    },
    staleTime: 30 * 1000,
  })
