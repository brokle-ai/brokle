'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import {
  listAPIKeys,
  createAPIKey,
  updateAPIKey,
  deleteAPIKey,
} from '../api/api-keys-api'
import type {
  APIKey,
  CreateAPIKeyRequest,
  UpdateAPIKeyRequest,
  APIKeyFilters,
} from '../types/api-keys'

/**
 * Query keys for API keys
 * Structured factory pattern for cache management
 */
export const apiKeyQueryKeys = {
  all: ['api-keys'] as const,
  lists: () => [...apiKeyQueryKeys.all, 'list'] as const,
  list: (projectId: string, filters?: APIKeyFilters) =>
    [...apiKeyQueryKeys.lists(), projectId, filters] as const,
  detail: (keyId: string) => [...apiKeyQueryKeys.all, 'detail', keyId] as const,
}

/**
 * Query hook to list API keys for a project
 *
 * @param projectId - Project ULID
 * @param filters - Optional filters (status, pagination, sorting)
 * @param options - React Query options
 *
 * @example
 * ```tsx
 * const { data, isLoading, error } = useAPIKeysQuery('proj_123', {
 *   status: 'active',
 *   page: 1,
 *   limit: 50,
 * })
 * ```
 */
export function useAPIKeysQuery(
  projectId: string | undefined,
  filters?: APIKeyFilters,
  options: {
    enabled?: boolean
  } = {}
) {
  return useQuery({
    queryKey: apiKeyQueryKeys.list(projectId || '', filters),
    queryFn: () => {
      if (!projectId) {
        throw new Error('Project ID is required')
      }
      return listAPIKeys(projectId, filters)
    },
    enabled: !!projectId && (options.enabled ?? true),
    staleTime: 30000, // 30 seconds
    gcTime: 5 * 60 * 1000, // 5 minutes (formerly cacheTime)
  })
}

/**
 * Mutation hook to create a new API key
 *
 * Automatically invalidates API keys list cache on success
 * and shows toast notifications.
 *
 * @example
 * ```tsx
 * const createMutation = useCreateAPIKeyMutation('proj_123')
 *
 * const handleCreate = async () => {
 *   const newKey = await createMutation.mutateAsync({
 *     name: 'Production Key',
 *     expiry_option: '90days'
 *   })
 *   console.log(newKey.key) // Full key, only shown once
 * }
 * ```
 */
export function useCreateAPIKeyMutation(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (data: CreateAPIKeyRequest) => {
      return createAPIKey(projectId, data)
    },
    onSuccess: (newKey: APIKey) => {
      // Invalidate all API keys lists for this project
      queryClient.invalidateQueries({
        queryKey: apiKeyQueryKeys.lists()
      })

      // Show success toast
      toast.success('API Key Created!', {
        description: `${newKey.name} has been created successfully. Make sure to copy it now - you won't be able to see it again.`,
        duration: 6000, // Longer duration for important message
      })
    },
    onError: (error: unknown) => {
      const apiError = error as { message?: string }
      toast.error('Failed to Create API Key', {
        description: apiError?.message || 'Could not create API key. Please try again.',
      })
    },
  })
}

/**
 * Mutation hook to update an existing API key
 *
 * Supports optimistic updates for better UX.
 *
 * @example
 * ```tsx
 * const updateMutation = useUpdateAPIKeyMutation('proj_123')
 *
 * // Update name
 * await updateMutation.mutateAsync({
 *   keyId: 'key_456',
 *   data: { name: 'Updated Name' }
 * })
 *
 * // Revoke key
 * await updateMutation.mutateAsync({
 *   keyId: 'key_456',
 *   data: { is_active: false }
 * })
 * ```
 */
export function useUpdateAPIKeyMutation(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      keyId,
      data
    }: {
      keyId: string
      data: UpdateAPIKeyRequest
    }) => {
      return updateAPIKey(projectId, keyId, data)
    },
    onMutate: async ({ keyId, data }) => {
      // Cancel outgoing refetches
      await queryClient.cancelQueries({
        queryKey: apiKeyQueryKeys.lists()
      })

      // Snapshot previous value
      const previousKeys = queryClient.getQueriesData({
        queryKey: apiKeyQueryKeys.lists()
      })

      // Optimistically update cache
      queryClient.setQueriesData<{ data: APIKey[] }>(
        { queryKey: apiKeyQueryKeys.lists() },
        (old) => {
          if (!old) return old

          return {
            ...old,
            data: old.data.map((key) => {
              if (key.id === keyId) {
                return {
                  ...key,
                  ...(data.name && { name: data.name }),
                  ...(data.is_active !== undefined && {
                    status: data.is_active ? 'active' as const : 'inactive' as const
                  }),
                }
              }
              return key
            }),
          }
        }
      )

      return { previousKeys }
    },
    onSuccess: (updatedKey: APIKey) => {
      // Invalidate to get fresh data from server
      queryClient.invalidateQueries({
        queryKey: apiKeyQueryKeys.lists()
      })

      // Show success toast
      const action = updatedKey.status === 'inactive' ? 'revoked' : 'updated'
      toast.success(`API Key ${action}!`, {
        description: `${updatedKey.name} has been ${action} successfully.`,
      })
    },
    onError: (error: unknown, _variables, context) => {
      // Rollback optimistic update on error
      if (context?.previousKeys) {
        context.previousKeys.forEach(([queryKey, data]) => {
          queryClient.setQueryData(queryKey, data)
        })
      }

      const apiError = error as { message?: string }
      toast.error('Failed to Update API Key', {
        description: apiError?.message || 'Could not update API key. Please try again.',
      })
    },
  })
}

/**
 * Mutation hook to delete an API key permanently
 *
 * Supports optimistic updates and confirmation dialogs.
 *
 * @example
 * ```tsx
 * const deleteMutation = useDeleteAPIKeyMutation('proj_123')
 *
 * const handleDelete = async (keyId: string, keyName: string) => {
 *   if (confirm(`Delete ${keyName}? This action cannot be undone.`)) {
 *     await deleteMutation.mutateAsync({ keyId, keyName })
 *   }
 * }
 * ```
 */
export function useDeleteAPIKeyMutation(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      keyId
    }: {
      keyId: string
      keyName: string // For display in toast
    }) => {
      return deleteAPIKey(projectId, keyId)
    },
    onMutate: async ({ keyId }) => {
      // Cancel outgoing refetches
      await queryClient.cancelQueries({
        queryKey: apiKeyQueryKeys.lists()
      })

      // Snapshot previous value
      const previousKeys = queryClient.getQueriesData({
        queryKey: apiKeyQueryKeys.lists()
      })

      // Optimistically remove from cache
      queryClient.setQueriesData<{ data: APIKey[] }>(
        { queryKey: apiKeyQueryKeys.lists() },
        (old) => {
          if (!old) return old

          return {
            ...old,
            data: old.data.filter((key) => key.id !== keyId),
          }
        }
      )

      return { previousKeys }
    },
    onSuccess: (_data, variables) => {
      // Invalidate to get fresh data from server
      queryClient.invalidateQueries({
        queryKey: apiKeyQueryKeys.lists()
      })

      // Show success toast
      toast.success('API Key Deleted!', {
        description: `${variables.keyName} has been permanently deleted.`,
      })
    },
    onError: (error: unknown, variables, context) => {
      // Rollback optimistic update on error
      if (context?.previousKeys) {
        context.previousKeys.forEach(([queryKey, data]) => {
          queryClient.setQueryData(queryKey, data)
        })
      }

      const apiError = error as { message?: string }
      toast.error('Failed to Delete API Key', {
        description: apiError?.message || 'Could not delete API key. Please try again.',
      })
    },
  })
}
