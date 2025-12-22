'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { experimentsApi } from '../api/experiments-api'
import type {
  CreateExperimentRequest,
  UpdateExperimentRequest,
  Experiment,
} from '../types'

export const experimentQueryKeys = {
  all: ['experiments'] as const,
  list: (projectId: string, filters?: { dataset_id?: string; status?: string }) =>
    [...experimentQueryKeys.all, 'list', projectId, filters] as const,
  detail: (projectId: string, experimentId: string) =>
    [...experimentQueryKeys.all, 'detail', projectId, experimentId] as const,
  items: (projectId: string, experimentId: string) =>
    [...experimentQueryKeys.all, 'items', projectId, experimentId] as const,
}

export function useExperimentsQuery(
  projectId: string | undefined,
  filters?: { dataset_id?: string; status?: string }
) {
  return useQuery({
    queryKey: experimentQueryKeys.list(projectId ?? '', filters),
    queryFn: () => experimentsApi.listExperiments(projectId!, filters),
    enabled: !!projectId,
    staleTime: 30_000,
    gcTime: 5 * 60 * 1000,
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
        queryKey: experimentQueryKeys.all,
      })

      const previousExperiments = queryClient.getQueryData<Experiment[]>(
        experimentQueryKeys.list(projectId)
      )

      // Optimistic update
      queryClient.setQueryData<Experiment[]>(
        experimentQueryKeys.list(projectId),
        (old) => old?.filter((e) => e.id !== experimentId) ?? []
      )

      return { previousExperiments }
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: experimentQueryKeys.all,
      })
      toast.success('Experiment Deleted', {
        description: `"${variables.experimentName}" has been deleted.`,
      })
    },
    onError: (error: unknown, _variables, context) => {
      if (context?.previousExperiments) {
        queryClient.setQueryData(
          experimentQueryKeys.list(projectId),
          context.previousExperiments
        )
      }
      const apiError = error as { message?: string }
      toast.error('Failed to Delete Experiment', {
        description:
          apiError?.message || 'Could not delete experiment. Please try again.',
      })
    },
  })
}
