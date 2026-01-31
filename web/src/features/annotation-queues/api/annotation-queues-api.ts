import { BrokleAPIClient } from '@/lib/api/core/client'
import type {
  AnnotationQueue,
  QueueWithStats,
  QueueStats,
  QueueItem,
  QueueAssignment,
  CreateQueueRequest,
  UpdateQueueRequest,
  AddItemsBatchRequest,
  ClaimNextRequest,
  CompleteItemRequest,
  SkipItemRequest,
  AssignUserRequest,
  BatchAddItemsResponse,
  QueueListFilter,
  ItemListFilter,
} from '../types'

const client = new BrokleAPIClient('/api')

/**
 * Response type for queues list
 */
export interface QueuesResponse {
  queues: QueueWithStats[]
  totalCount: number
  page: number
  pageSize: number
  totalPages: number
  hasNext: boolean
  hasPrev: boolean
}

/**
 * Response type for queue items list
 */
export interface QueueItemsResponse {
  items: QueueItem[]
  totalCount: number
  page: number
  pageSize: number
  totalPages: number
  hasNext: boolean
  hasPrev: boolean
}

export const annotationQueuesApi = {
  // Queues

  listQueues: async (
    projectId: string,
    params?: QueueListFilter
  ): Promise<QueuesResponse> => {
    const queryParams: Record<string, string | number> = {}
    if (params?.status) queryParams.status = params.status
    if (params?.page) queryParams.page = params.page
    if (params?.limit) queryParams.limit = params.limit
    if (params?.search) queryParams.search = params.search

    const response = await client.getPaginated<QueueWithStats>(
      `/v1/projects/${projectId}/annotation-queues`,
      queryParams
    )

    return {
      queues: response.data,
      totalCount: response.pagination.total,
      page: response.pagination.page,
      pageSize: response.pagination.limit,
      totalPages: response.pagination.totalPages,
      hasNext: response.pagination.hasNext,
      hasPrev: response.pagination.hasPrev,
    }
  },

  getQueue: async (
    projectId: string,
    queueId: string
  ): Promise<AnnotationQueue> => {
    return client.get<AnnotationQueue>(
      `/v1/projects/${projectId}/annotation-queues/${queueId}`
    )
  },

  createQueue: async (
    projectId: string,
    data: CreateQueueRequest
  ): Promise<AnnotationQueue> => {
    return client.post<AnnotationQueue>(
      `/v1/projects/${projectId}/annotation-queues`,
      data
    )
  },

  updateQueue: async (
    projectId: string,
    queueId: string,
    data: UpdateQueueRequest
  ): Promise<AnnotationQueue> => {
    return client.put<AnnotationQueue>(
      `/v1/projects/${projectId}/annotation-queues/${queueId}`,
      data
    )
  },

  deleteQueue: async (projectId: string, queueId: string): Promise<void> => {
    await client.delete(`/v1/projects/${projectId}/annotation-queues/${queueId}`)
  },

  // Queue Stats

  getQueueStats: async (
    projectId: string,
    queueId: string
  ): Promise<QueueStats> => {
    return client.get<QueueStats>(
      `/v1/projects/${projectId}/annotation-queues/${queueId}/stats`
    )
  },

  // Queue Items

  listItems: async (
    projectId: string,
    queueId: string,
    filter?: ItemListFilter
  ): Promise<QueueItemsResponse> => {
    const params: Record<string, string | number> = {}
    if (filter?.status) params.status = filter.status
    if (filter?.page) params.page = filter.page
    if (filter?.limit) params.limit = filter.limit

    const response = await client.getPaginated<QueueItem>(
      `/v1/projects/${projectId}/annotation-queues/${queueId}/items`,
      params
    )

    return {
      items: response.data,
      totalCount: response.pagination.total,
      page: response.pagination.page,
      pageSize: response.pagination.limit,
      totalPages: response.pagination.totalPages,
      hasNext: response.pagination.hasNext,
      hasPrev: response.pagination.hasPrev,
    }
  },

  addItems: async (
    projectId: string,
    queueId: string,
    data: AddItemsBatchRequest
  ): Promise<BatchAddItemsResponse> => {
    return client.post<BatchAddItemsResponse>(
      `/v1/projects/${projectId}/annotation-queues/${queueId}/items`,
      data
    )
  },

  claimNext: async (
    projectId: string,
    queueId: string,
    data?: ClaimNextRequest
  ): Promise<QueueItem> => {
    return client.post<QueueItem>(
      `/v1/projects/${projectId}/annotation-queues/${queueId}/items/claim`,
      data ?? {}
    )
  },

  completeItem: async (
    projectId: string,
    queueId: string,
    itemId: string,
    data?: CompleteItemRequest
  ): Promise<QueueItem> => {
    return client.post<QueueItem>(
      `/v1/projects/${projectId}/annotation-queues/${queueId}/items/${itemId}/complete`,
      data ?? {}
    )
  },

  skipItem: async (
    projectId: string,
    queueId: string,
    itemId: string,
    data?: SkipItemRequest
  ): Promise<QueueItem> => {
    return client.post<QueueItem>(
      `/v1/projects/${projectId}/annotation-queues/${queueId}/items/${itemId}/skip`,
      data ?? {}
    )
  },

  releaseItem: async (
    projectId: string,
    queueId: string,
    itemId: string
  ): Promise<void> => {
    await client.post(
      `/v1/projects/${projectId}/annotation-queues/${queueId}/items/${itemId}/release`,
      {}
    )
  },

  // Assignments

  listAssignments: async (
    projectId: string,
    queueId: string
  ): Promise<QueueAssignment[]> => {
    return client.get<QueueAssignment[]>(
      `/v1/projects/${projectId}/annotation-queues/${queueId}/assignments`
    )
  },

  assignUser: async (
    projectId: string,
    queueId: string,
    data: AssignUserRequest
  ): Promise<QueueAssignment> => {
    return client.post<QueueAssignment>(
      `/v1/projects/${projectId}/annotation-queues/${queueId}/assignments`,
      data
    )
  },

  unassignUser: async (
    projectId: string,
    queueId: string,
    userId: string
  ): Promise<void> => {
    await client.delete(
      `/v1/projects/${projectId}/annotation-queues/${queueId}/assignments/${userId}`
    )
  },
}
