'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { annotationQueuesApi } from '../api/annotation-queues-api'
import type { PaginatedResponse } from '@/lib/api/core/types'
import type {
  CreateQueueRequest,
  UpdateQueueRequest,
  AddItemsBatchRequest,
  ClaimNextRequest,
  CompleteItemRequest,
  SkipItemRequest,
  AssignUserRequest,
  QueueWithStats,
  QueueItem,
  QueueListFilter,
  ItemListFilter,
} from '../types'

// Query Keys
export const annotationQueueQueryKeys = {
  all: ['annotation-queues'] as const,
  list: (projectId: string) =>
    [...annotationQueueQueryKeys.all, 'list', projectId] as const,
  detail: (projectId: string, queueId: string) =>
    [...annotationQueueQueryKeys.all, 'detail', projectId, queueId] as const,
  stats: (projectId: string, queueId: string) =>
    [...annotationQueueQueryKeys.all, 'stats', projectId, queueId] as const,
  items: (projectId: string, queueId: string) =>
    [...annotationQueueQueryKeys.all, 'items', projectId, queueId] as const,
  assignments: (projectId: string, queueId: string) =>
    [...annotationQueueQueryKeys.all, 'assignments', projectId, queueId] as const,
}

// ============================================================================
// Queue Queries
// ============================================================================

export function useAnnotationQueuesQuery(
  projectId: string | undefined,
  params?: QueueListFilter
) {
  return useQuery({
    queryKey: [
      ...annotationQueueQueryKeys.list(projectId ?? ''),
      params?.search,
      params?.status,
      params?.page,
      params?.limit,
    ],
    queryFn: () => annotationQueuesApi.listQueues(projectId!, params),
    enabled: !!projectId,
    staleTime: 30_000,
    gcTime: 5 * 60 * 1000,
  })
}

export function useAnnotationQueueQuery(
  projectId: string | undefined,
  queueId: string | undefined
) {
  return useQuery({
    queryKey: annotationQueueQueryKeys.detail(projectId ?? '', queueId ?? ''),
    queryFn: () => annotationQueuesApi.getQueue(projectId!, queueId!),
    enabled: !!projectId && !!queueId,
    staleTime: 30_000,
    gcTime: 5 * 60 * 1000,
  })
}

export function useQueueStatsQuery(
  projectId: string | undefined,
  queueId: string | undefined
) {
  return useQuery({
    queryKey: annotationQueueQueryKeys.stats(projectId ?? '', queueId ?? ''),
    queryFn: () => annotationQueuesApi.getQueueStats(projectId!, queueId!),
    enabled: !!projectId && !!queueId,
    staleTime: 10_000, // Shorter stale time for stats
    gcTime: 60 * 1000,
  })
}

// ============================================================================
// Queue Mutations
// ============================================================================

export function useCreateQueueMutation(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateQueueRequest) =>
      annotationQueuesApi.createQueue(projectId, data),
    onSuccess: (newQueue) => {
      queryClient.invalidateQueries({
        queryKey: annotationQueueQueryKeys.list(projectId),
      })
      toast.success('Queue Created', {
        description: `"${newQueue.name}" has been created successfully.`,
      })
    },
    onError: (error: unknown) => {
      const apiError = error as { message?: string }
      toast.error('Failed to Create Queue', {
        description: apiError?.message || 'Could not create annotation queue. Please try again.',
      })
    },
  })
}

export function useUpdateQueueMutation(projectId: string, queueId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: UpdateQueueRequest) =>
      annotationQueuesApi.updateQueue(projectId, queueId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: annotationQueueQueryKeys.all,
      })
      toast.success('Queue Updated', {
        description: 'Annotation queue has been updated.',
      })
    },
    onError: (error: unknown) => {
      const apiError = error as { message?: string }
      toast.error('Failed to Update Queue', {
        description: apiError?.message || 'Could not update annotation queue. Please try again.',
      })
    },
  })
}

export function useDeleteQueueMutation(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ queueId, queueName }: { queueId: string; queueName: string }) => {
      await annotationQueuesApi.deleteQueue(projectId, queueId)
      return { queueId, queueName }
    },
    onMutate: async ({ queueId }) => {
      await queryClient.cancelQueries({
        queryKey: annotationQueueQueryKeys.list(projectId),
      })

      // Get ALL matching queries (prefix match for paginated queries)
      const previousQueries = queryClient.getQueriesData<PaginatedResponse<QueueWithStats>>({
        queryKey: annotationQueueQueryKeys.list(projectId),
      })

      // Optimistic update - update ALL matching queries
      queryClient.setQueriesData<PaginatedResponse<QueueWithStats>>(
        { queryKey: annotationQueueQueryKeys.list(projectId) },
        (old) => old ? {
          data: old.data.filter((q) => q.queue.id !== queueId),
          pagination: {
            ...old.pagination,
            total: old.pagination.total - 1,
          },
        } : old
      )

      return { previousQueries }
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: annotationQueueQueryKeys.list(projectId),
      })
      toast.success('Queue Deleted', {
        description: `"${variables.queueName}" has been deleted.`,
      })
    },
    onError: (error: unknown, _variables, context) => {
      // Rollback ALL affected queries
      context?.previousQueries?.forEach(([queryKey, data]) => {
        queryClient.setQueryData(queryKey, data)
      })
      const apiError = error as { message?: string }
      toast.error('Failed to Delete Queue', {
        description: apiError?.message || 'Could not delete annotation queue. Please try again.',
      })
    },
  })
}

// ============================================================================
// Item Queries
// ============================================================================

export function useQueueItemsQuery(
  projectId: string | undefined,
  queueId: string | undefined,
  filter?: ItemListFilter
) {
  return useQuery({
    queryKey: [
      ...annotationQueueQueryKeys.items(projectId ?? '', queueId ?? ''),
      filter?.status,
      filter?.page,
      filter?.limit,
    ],
    queryFn: () => annotationQueuesApi.listItems(projectId!, queueId!, filter),
    enabled: !!projectId && !!queueId,
    staleTime: 15_000, // Shorter stale time for items that change frequently
    gcTime: 60 * 1000,
  })
}

// ============================================================================
// Item Mutations
// ============================================================================

export function useAddItemsMutation(projectId: string, queueId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: AddItemsBatchRequest) =>
      annotationQueuesApi.addItems(projectId, queueId, data),
    onSuccess: (result) => {
      queryClient.invalidateQueries({
        queryKey: annotationQueueQueryKeys.items(projectId, queueId),
      })
      queryClient.invalidateQueries({
        queryKey: annotationQueueQueryKeys.stats(projectId, queueId),
      })
      queryClient.invalidateQueries({
        queryKey: annotationQueueQueryKeys.list(projectId),
      })
      toast.success('Items Added', {
        description: `${result.created} item(s) added to the queue.`,
      })
    },
    onError: (error: unknown) => {
      const apiError = error as { message?: string }
      toast.error('Failed to Add Items', {
        description: apiError?.message || 'Could not add items to queue. Please try again.',
      })
    },
  })
}

export function useClaimNextItemMutation(projectId: string, queueId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data?: ClaimNextRequest) =>
      annotationQueuesApi.claimNext(projectId, queueId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: annotationQueueQueryKeys.items(projectId, queueId),
      })
      queryClient.invalidateQueries({
        queryKey: annotationQueueQueryKeys.stats(projectId, queueId),
      })
    },
    onError: (error: unknown) => {
      const apiError = error as { message?: string }
      // Don't show toast for "no items available" - this is expected behavior
      if (!apiError?.message?.toLowerCase().includes('no items available')) {
        toast.error('Failed to Claim Item', {
          description: apiError?.message || 'Could not claim next item. Please try again.',
        })
      }
    },
  })
}

export function useCompleteItemMutation(projectId: string, queueId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ itemId, data }: { itemId: string; data?: CompleteItemRequest }) =>
      annotationQueuesApi.completeItem(projectId, queueId, itemId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: annotationQueueQueryKeys.items(projectId, queueId),
      })
      queryClient.invalidateQueries({
        queryKey: annotationQueueQueryKeys.stats(projectId, queueId),
      })
      queryClient.invalidateQueries({
        queryKey: annotationQueueQueryKeys.list(projectId),
      })
      toast.success('Item Completed', {
        description: 'Annotation has been submitted.',
      })
    },
    onError: (error: unknown) => {
      const apiError = error as { message?: string }
      toast.error('Failed to Complete Item', {
        description: apiError?.message || 'Could not complete item. Please try again.',
      })
    },
  })
}

export function useSkipItemMutation(projectId: string, queueId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ itemId, data }: { itemId: string; data?: SkipItemRequest }) =>
      annotationQueuesApi.skipItem(projectId, queueId, itemId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: annotationQueueQueryKeys.items(projectId, queueId),
      })
      queryClient.invalidateQueries({
        queryKey: annotationQueueQueryKeys.stats(projectId, queueId),
      })
      queryClient.invalidateQueries({
        queryKey: annotationQueueQueryKeys.list(projectId),
      })
      toast.info('Item Skipped', {
        description: 'Moving to next item.',
      })
    },
    onError: (error: unknown) => {
      const apiError = error as { message?: string }
      toast.error('Failed to Skip Item', {
        description: apiError?.message || 'Could not skip item. Please try again.',
      })
    },
  })
}

export function useReleaseItemMutation(projectId: string, queueId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (itemId: string) =>
      annotationQueuesApi.releaseItem(projectId, queueId, itemId),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: annotationQueueQueryKeys.items(projectId, queueId),
      })
      queryClient.invalidateQueries({
        queryKey: annotationQueueQueryKeys.stats(projectId, queueId),
      })
      toast.info('Item Released', {
        description: 'Item lock has been released.',
      })
    },
    onError: (error: unknown) => {
      const apiError = error as { message?: string }
      toast.error('Failed to Release Item', {
        description: apiError?.message || 'Could not release item. Please try again.',
      })
    },
  })
}

// ============================================================================
// Assignment Queries
// ============================================================================

export function useQueueAssignmentsQuery(
  projectId: string | undefined,
  queueId: string | undefined
) {
  return useQuery({
    queryKey: annotationQueueQueryKeys.assignments(projectId ?? '', queueId ?? ''),
    queryFn: () => annotationQueuesApi.listAssignments(projectId!, queueId!),
    enabled: !!projectId && !!queueId,
    staleTime: 60_000,
    gcTime: 5 * 60 * 1000,
  })
}

// ============================================================================
// Assignment Mutations
// ============================================================================

export function useAssignUserMutation(projectId: string, queueId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: AssignUserRequest) =>
      annotationQueuesApi.assignUser(projectId, queueId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: annotationQueueQueryKeys.assignments(projectId, queueId),
      })
      toast.success('User Assigned', {
        description: 'User has been assigned to the queue.',
      })
    },
    onError: (error: unknown) => {
      const apiError = error as { message?: string }
      toast.error('Failed to Assign User', {
        description: apiError?.message || 'Could not assign user to queue. Please try again.',
      })
    },
  })
}

export function useUnassignUserMutation(projectId: string, queueId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (userId: string) =>
      annotationQueuesApi.unassignUser(projectId, queueId, userId),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: annotationQueueQueryKeys.assignments(projectId, queueId),
      })
      toast.success('User Removed', {
        description: 'User has been removed from the queue.',
      })
    },
    onError: (error: unknown) => {
      const apiError = error as { message?: string }
      toast.error('Failed to Remove User', {
        description: apiError?.message || 'Could not remove user from queue. Please try again.',
      })
    },
  })
}
