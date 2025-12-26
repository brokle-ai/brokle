'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { scoresApi } from '../api/scores-api'
import type { CreateScoreConfigRequest, UpdateScoreConfigRequest, ScoreConfig } from '../types'

export const scoreConfigQueryKeys = {
  all: ['score-configs'] as const,
  list: (projectId: string) => [...scoreConfigQueryKeys.all, 'list', projectId] as const,
  detail: (projectId: string, configId: string) =>
    [...scoreConfigQueryKeys.all, 'detail', projectId, configId] as const,
}

export function useScoreConfigsQuery(projectId: string | undefined) {
  return useQuery({
    queryKey: scoreConfigQueryKeys.list(projectId ?? ''),
    queryFn: () => scoresApi.listScoreConfigs(projectId!),
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

      const previousConfigs = queryClient.getQueryData<ScoreConfig[]>(
        scoreConfigQueryKeys.list(projectId)
      )

      // Optimistic update
      queryClient.setQueryData<ScoreConfig[]>(
        scoreConfigQueryKeys.list(projectId),
        (old) => old?.filter((c) => c.id !== configId) ?? []
      )

      return { previousConfigs }
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
      if (context?.previousConfigs) {
        queryClient.setQueryData(
          scoreConfigQueryKeys.list(projectId),
          context.previousConfigs
        )
      }
      const apiError = error as { message?: string }
      toast.error('Failed to Delete Score Config', {
        description: apiError?.message || 'Could not delete score config. Please try again.',
      })
    },
  })
}
