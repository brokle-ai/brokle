import { queryOptions } from '@tanstack/react-query'
import { rawFetch } from '@/lib/api/client'
import type {
  ClaimNextRequest,
  CompleteItemRequest,
  QueueDetail,
  QueueItem,
  QueueItemStatus,
  QueueItemsListResponse,
  QueueListResponse,
  QueueStatus,
  SkipItemRequest,
} from './types'

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

// Use the `/stats` variant so detail view has item counts without a
// second fetch — matches the list view's QueueWithStats shape.
export const queueDetailQueryOptions = (projectId: string, queueId: string) =>
  queryOptions({
    queryKey: queuesKeys.detail(queueId),
    queryFn: async () => {
      const resp = await rawFetch(
        `/api/v1/projects/${projectId}/annotation-queues/${queueId}/stats`,
        { method: 'GET' },
      )
      return (await resp.json()) as QueueDetail
    },
    staleTime: 15 * 1000,
  })

export interface QueueItemsListParams {
  page: number
  limit: number
  status?: QueueItemStatus
}

export const queueItemsKey = (queueId: string, params: QueueItemsListParams) =>
  [...queuesKeys.detail(queueId), 'items', params] as const

export const queueItemsListQueryOptions = (
  projectId: string,
  queueId: string,
  params: QueueItemsListParams,
) =>
  queryOptions({
    queryKey: queueItemsKey(queueId, params),
    queryFn: async () => {
      const search = new URLSearchParams()
      search.set('page', String(params.page))
      search.set('limit', String(params.limit))
      if (params.status) search.set('status', params.status)
      const resp = await rawFetch(
        `/api/v1/projects/${projectId}/annotation-queues/${queueId}/items?${search.toString()}`,
        { method: 'GET' },
      )
      return (await resp.json()) as QueueItemsListResponse
    },
    staleTime: 10 * 1000,
  })

// Item lifecycle mutations — kept thin so callers can compose them
// with useMutation + their own onSuccess/invalidation logic. Each hits
// a 2xx-body POST endpoint that returns the post-state QueueItem (or
// 404 when no claimable item is left).
export async function claimNextItem(
  projectId: string,
  queueId: string,
  data: ClaimNextRequest = {},
): Promise<QueueItem> {
  const resp = await rawFetch(
    `/api/v1/projects/${projectId}/annotation-queues/${queueId}/items/claim`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    },
  )
  return (await resp.json()) as QueueItem
}

export async function completeItem(
  projectId: string,
  queueId: string,
  itemId: string,
  data: CompleteItemRequest = {},
): Promise<QueueItem> {
  const resp = await rawFetch(
    `/api/v1/projects/${projectId}/annotation-queues/${queueId}/items/${itemId}/complete`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    },
  )
  return (await resp.json()) as QueueItem
}

export async function skipItem(
  projectId: string,
  queueId: string,
  itemId: string,
  data: SkipItemRequest = {},
): Promise<QueueItem> {
  const resp = await rawFetch(
    `/api/v1/projects/${projectId}/annotation-queues/${queueId}/items/${itemId}/skip`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    },
  )
  return (await resp.json()) as QueueItem
}
