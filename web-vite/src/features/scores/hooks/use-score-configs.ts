import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { BrokleError } from '@/lib/api/errors'
import {
  createScoreConfig,
  deleteScoreConfig,
  scoreConfigsPagedQueryOptions,
  scoreConfigsQueryOptions,
  scoresKeys,
  updateScoreConfig,
} from '../api/queries'
import type {
  CreateScoreConfigRequest,
  UpdateScoreConfigRequest,
} from '../api/types'

/**
 * Read hook for the project score-config catalog (high-cap, no
 * pagination). Used by annotation widgets that need the full set.
 */
export function useScoreConfigsQuery(projectId: string | undefined) {
  return useQuery({
    ...scoreConfigsQueryOptions(projectId ?? ''),
    enabled: !!projectId,
  })
}

/**
 * Paginated score-configs query for the management surface. Mirrors
 * the wire envelope `{data, total, page, limit}`.
 */
export function useScoreConfigsPagedQuery(
  projectId: string | undefined,
  page: number,
  limit: number,
) {
  return useQuery({
    ...scoreConfigsPagedQueryOptions(projectId ?? '', page, limit),
    enabled: !!projectId,
  })
}

function describeError(err: unknown): string | undefined {
  if (err instanceof BrokleError) return err.message
  if (err instanceof Error) return err.message
  return undefined
}

export function useCreateScoreConfigMutation(projectId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: CreateScoreConfigRequest) =>
      createScoreConfig(projectId, data),
    onSuccess: (created) => {
      queryClient.invalidateQueries({ queryKey: scoresKeys.configs() })
      toast.success('Score config created', {
        description: `"${created.name}" is ready to use.`,
      })
    },
    onError: (err) => {
      toast.error('Failed to create score config', {
        description: describeError(err) ?? 'Please try again.',
      })
    },
  })
}

export function useUpdateScoreConfigMutation(
  projectId: string,
  configId: string,
) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: UpdateScoreConfigRequest) =>
      updateScoreConfig(projectId, configId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: scoresKeys.configs() })
      toast.success('Score config updated')
    },
    onError: (err) => {
      toast.error('Failed to update score config', {
        description: describeError(err) ?? 'Please try again.',
      })
    },
  })
}

export function useDeleteScoreConfigMutation(projectId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({
      configId,
    }: {
      configId: string
      configName: string
    }) => {
      await deleteScoreConfig(projectId, configId)
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: scoresKeys.configs() })
      toast.success('Score config deleted', {
        description: `"${variables.configName}" has been removed.`,
      })
    },
    onError: (err) => {
      toast.error('Failed to delete score config', {
        description: describeError(err) ?? 'Please try again.',
      })
    },
  })
}
