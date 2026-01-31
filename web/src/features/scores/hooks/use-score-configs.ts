'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { scoresApi, type ScoreConfigsResponse } from '../api/scores-api'
import type {
  CreateScoreConfigRequest,
  UpdateScoreConfigRequest,
  ScoreConfig,
  ScoreConfigListParams,
} from '../types'

export const scoreConfigQueryKeys = {
  all: ['score-configs'] as const,
  list: (projectId: string) => [...scoreConfigQueryKeys.all, 'list', projectId] as const,
  detail: (projectId: string, configId: string) =>
    [...scoreConfigQueryKeys.all, 'detail', projectId, configId] as const,
}

export function useScoreConfigsQuery(
  projectId: string | undefined,
  params?: ScoreConfigListParams
) {
  return useQuery({
    queryKey: [
      ...scoreConfigQueryKeys.list(projectId ?? ''),
      params?.page,
      params?.limit,
    ],
    queryFn: () => scoresApi.listScoreConfigs(projectId!, params),
    enabled: !!projectId,
    staleTime: 30_000,
    gcTime: 5 * 60 * 1000,
  })
}

export function useScoreConfigQuery(
  projectId: string | undefined,
  configId: string | undefined
) {
  return useQuery({
    queryKey: scoreConfigQueryKeys.detail(projectId ?? '', configId ?? ''),
    queryFn: () => scoresApi.getScoreConfig(projectId!, configId!),
    enabled: !!projectId && !!configId,
    staleTime: 30_000,
    gcTime: 5 * 60 * 1000,
  })
}

export function useCreateScoreConfigMutation(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateScoreConfigRequest) =>
      scoresApi.createScoreConfig(projectId, data),
    onSuccess: (newConfig) => {
      queryClient.invalidateQueries({
        queryKey: scoreConfigQueryKeys.list(projectId),
      })
      toast.success('Score Config Created', {
        description: `"${newConfig.name}" has been created successfully.`,
      })
    },
    onError: (error: unknown) => {
      const apiError = error as { message?: string }
      toast.error('Failed to Create Score Config', {
        description: apiError?.message || 'Could not create score config. Please try again.',
      })
    },
  })
}

export function useUpdateScoreConfigMutation(projectId: string, configId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: UpdateScoreConfigRequest) =>
      scoresApi.updateScoreConfig(projectId, configId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: scoreConfigQueryKeys.all,
      })
      toast.success('Score Config Updated', {
        description: 'Score config has been updated.',
      })
    },
    onError: (error: unknown) => {
      const apiError = error as { message?: string }
      toast.error('Failed to Update Score Config', {
        description: apiError?.message || 'Could not update score config. Please try again.',
      })
    },
  })
}

/**
 * Fetch multiple score configs by their IDs.
 * Used by annotation queues to load configs for scoring forms.
 */
export function useScoreConfigsByIdsQuery(
  projectId: string | undefined,
  configIds: string[]
) {
  return useQuery({
    queryKey: [...scoreConfigQueryKeys.list(projectId ?? ''), 'byIds', configIds],
    queryFn: async () => {
      if (!projectId || configIds.length === 0) return []
      const response = await scoresApi.listScoreConfigs(projectId)
      return response.configs.filter((c) => configIds.includes(c.id))
    },
    enabled: !!projectId && configIds.length > 0,
    staleTime: 30_000,
    gcTime: 5 * 60 * 1000,
  })
}

export function useDeleteScoreConfigMutation(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ configId, configName }: { configId: string; configName: string }) => {
      await scoresApi.deleteScoreConfig(projectId, configId)
      return { configId, configName }
    },
    onMutate: async ({ configId }) => {
      await queryClient.cancelQueries({
        queryKey: scoreConfigQueryKeys.list(projectId),
      })

      // Get ALL matching queries (prefix match for paginated queries)
      const previousQueries = queryClient.getQueriesData<ScoreConfigsResponse>({
        queryKey: scoreConfigQueryKeys.list(projectId),
      })

      // Optimistic update - update ALL matching queries
      queryClient.setQueriesData<ScoreConfigsResponse>(
        { queryKey: scoreConfigQueryKeys.list(projectId) },
        (old) => old ? {
          ...old,
          configs: old.configs.filter((c) => c.id !== configId),
          totalCount: old.totalCount - 1,
        } : old
      )

      return { previousQueries }
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: scoreConfigQueryKeys.list(projectId),
      })
      toast.success('Score Config Deleted', {
        description: `"${variables.configName}" has been deleted.`,
      })
    },
    onError: (error: unknown, _variables, context) => {
      // Rollback ALL affected queries
      context?.previousQueries?.forEach(([queryKey, data]) => {
        queryClient.setQueryData(queryKey, data)
      })
      const apiError = error as { message?: string }
      toast.error('Failed to Delete Score Config', {
        description: apiError?.message || 'Could not delete score config. Please try again.',
      })
    },
  })
}
