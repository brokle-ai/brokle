'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { experimentsApi } from '../api/experiments-api'
import type { PaginatedResponse } from '@/lib/api/core/types'
import type {
  CreateExperimentRequest,
  UpdateExperimentRequest,
  RerunExperimentRequest,
  Experiment,
  ExperimentListParams,
} from '../types'

export const experimentQueryKeys = {
  all: ['experiments'] as const,
  list: (projectId: string) =>
    [...experimentQueryKeys.all, 'list', projectId] as const,
  detail: (projectId: string, experimentId: string) =>
    [...experimentQueryKeys.all, 'detail', projectId, experimentId] as const,
  items: (projectId: string, experimentId: string) =>
    [...experimentQueryKeys.all, 'items', projectId, experimentId] as const,
}

export function useExperimentsQuery(
  projectId: string | undefined,
  params?: ExperimentListParams
) {
  return useQuery({
    queryKey: [
      ...experimentQueryKeys.list(projectId ?? ''),
      params?.search,
      params?.page,
      params?.limit,
      params?.dataset_id,
      params?.status,
      params?.ids,
    ],
    queryFn: () => experimentsApi.listExperiments(projectId!, params),
    enabled: !!projectId,
    staleTime: 30_000,
    gcTime: 5 * 60 * 1000,
  })
}

export function useExperimentsByIdsQuery(
  projectId: string | undefined,
  ids: string[]
) {
  return useQuery({
    queryKey: ['experiments', 'byIds', projectId, ids.sort().join(',')],
    queryFn: () =>
      experimentsApi.listExperiments(projectId!, {
        ids: ids.join(','),
        limit: ids.length,
      }),
    enabled: !!projectId && ids.length > 0,
    staleTime: 5 * 60 * 1000, // 5 minutes - these are specific fetches
    gcTime: 10 * 60 * 1000,
  })
}

export function useExperimentQuery(
  projectId: string | undefined,
  experimentId: string | undefined
) {
  return useQuery({
    queryKey: experimentQueryKeys.detail(projectId ?? '', experimentId ?? ''),
    queryFn: () => experimentsApi.getExperiment(projectId!, experimentId!),
    enabled: !!projectId && !!experimentId,
    staleTime: 30_000,
    gcTime: 5 * 60 * 1000,
  })
}

export function useExperimentItemsQuery(
  projectId: string | undefined,
  experimentId: string | undefined,
  limit = 50,
  offset = 0
) {
  return useQuery({
    queryKey: [
      ...experimentQueryKeys.items(projectId ?? '', experimentId ?? ''),
      limit,
      offset,
    ],
    queryFn: () =>
      experimentsApi.listExperimentItems(projectId!, experimentId!, limit, offset),
    enabled: !!projectId && !!experimentId,
    staleTime: 30_000,
    gcTime: 5 * 60 * 1000,
  })
}

export function useCreateExperimentMutation(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateExperimentRequest) =>
      experimentsApi.createExperiment(projectId, data),
    onSuccess: (newExperiment) => {
      queryClient.invalidateQueries({
        queryKey: experimentQueryKeys.all,
      })
      toast.success('Experiment Created', {
        description: `"${newExperiment.name}" has been created successfully.`,
      })
    },
    onError: (error: unknown) => {
      const apiError = error as { message?: string }
      toast.error('Failed to Create Experiment', {
        description:
          apiError?.message || 'Could not create experiment. Please try again.',
      })
    },
  })
}

export function useUpdateExperimentMutation(
  projectId: string,
  experimentId: string
) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: UpdateExperimentRequest) =>
      experimentsApi.updateExperiment(projectId, experimentId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: experimentQueryKeys.all,
      })
      toast.success('Experiment Updated', {
        description: 'Experiment has been updated.',
      })
    },
    onError: (error: unknown) => {
      const apiError = error as { message?: string }
      toast.error('Failed to Update Experiment', {
        description:
          apiError?.message || 'Could not update experiment. Please try again.',
      })
    },
  })
}

export function useDeleteExperimentMutation(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      experimentId,
      experimentName,
    }: {
      experimentId: string
      experimentName: string
    }) => {
      await experimentsApi.deleteExperiment(projectId, experimentId)
      return { experimentId, experimentName }
    },
    onMutate: async ({ experimentId }) => {
      await queryClient.cancelQueries({
        queryKey: experimentQueryKeys.list(projectId),
      })

      // Get ALL matching queries (prefix match for paginated queries)
      const previousQueries = queryClient.getQueriesData<PaginatedResponse<Experiment>>({
        queryKey: experimentQueryKeys.list(projectId),
      })

      // Optimistic update - update ALL matching queries
      queryClient.setQueriesData<PaginatedResponse<Experiment>>(
        { queryKey: experimentQueryKeys.list(projectId) },
        (old) => old ? {
          data: old.data.filter((e) => e.id !== experimentId),
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
        queryKey: experimentQueryKeys.list(projectId),
      })
      toast.success('Experiment Deleted', {
        description: `"${variables.experimentName}" has been deleted.`,
      })
    },
    onError: (error: unknown, _variables, context) => {
      // Rollback ALL affected queries
      context?.previousQueries?.forEach(([queryKey, data]) => {
        queryClient.setQueryData(queryKey, data)
      })
      const apiError = error as { message?: string }
      toast.error('Failed to Delete Experiment', {
        description:
          apiError?.message || 'Could not delete experiment. Please try again.',
      })
    },
  })
}

export function useRerunExperimentMutation(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({
      experimentId,
      data,
    }: {
      experimentId: string
      data?: RerunExperimentRequest
    }) => experimentsApi.rerunExperiment(projectId, experimentId, data),
    onSuccess: (newExperiment) => {
      queryClient.invalidateQueries({
        queryKey: experimentQueryKeys.all,
      })
      toast.success('Experiment Re-run Created', {
        description: `"${newExperiment.name}" has been created. Run your evaluation task to populate results.`,
      })
    },
    onError: (error: unknown) => {
      const apiError = error as { message?: string }
      toast.error('Failed to Re-run Experiment', {
        description:
          apiError?.message || 'Could not create re-run. Please try again.',
      })
    },
  })
}
