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
  ItemListResponse,
  BatchAddItemsResponse,
  ItemListFilter,
} from '../types'

const client = new BrokleAPIClient('/api')

export const annotationQueuesApi = {
  // Queues

  listQueues: async (projectId: string): Promise<QueueWithStats[]> => {
    return client.get<QueueWithStats[]>(
      `/v1/projects/${projectId}/annotation-queues`
    )
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
  ): Promise<ItemListResponse> => {
    const params = new URLSearchParams()
    if (filter?.status) params.append('status', filter.status)
    if (filter?.limit) params.append('limit', String(filter.limit))
    if (filter?.offset) params.append('offset', String(filter.offset))

    const queryString = params.toString()
    const url = `/v1/projects/${projectId}/annotation-queues/${queueId}/items${queryString ? `?${queryString}` : ''}`

    return client.get<ItemListResponse>(url)
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
