'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { extractErrorMessage } from '@/lib/api/error-utils'
import { evaluatorsApi, type EvaluatorsResponse } from '../api/evaluators-api'
import type {
  CreateEvaluatorRequest,
  UpdateEvaluatorRequest,
  Evaluator,
  EvaluatorListParams,
  TriggerOptions,
} from '../types'
import { evaluatorExecutionsKeys } from './use-evaluator-executions'

export const evaluatorQueryKeys = {
  all: ['evaluators'] as const,
  list: (projectId: string, params?: EvaluatorListParams) =>
    [...evaluatorQueryKeys.all, 'list', projectId, params] as const,
  detail: (projectId: string, evaluatorId: string) =>
    [...evaluatorQueryKeys.all, 'detail', projectId, evaluatorId] as const,
}

/**
 * Query hook for listing evaluators with pagination and filtering
 */
export function useEvaluatorsQuery(
  projectId: string | undefined,
  params?: EvaluatorListParams
) {
  return useQuery({
    queryKey: evaluatorQueryKeys.list(projectId ?? '', params),
    queryFn: () => evaluatorsApi.listEvaluators(projectId!, params),
    enabled: !!projectId,
    staleTime: 30_000,
    gcTime: 5 * 60 * 1000,
  })
}

/**
 * Query hook for fetching a single evaluator
 */
export function useEvaluatorQuery(
  projectId: string | undefined,
  evaluatorId: string | undefined
) {
  return useQuery({
    queryKey: evaluatorQueryKeys.detail(projectId ?? '', evaluatorId ?? ''),
    queryFn: () => evaluatorsApi.getEvaluator(projectId!, evaluatorId!),
    enabled: !!projectId && !!evaluatorId,
    staleTime: 30_000,
    gcTime: 5 * 60 * 1000,
  })
}

/**
 * Mutation hook for creating a new evaluator
 */
export function useCreateEvaluatorMutation(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateEvaluatorRequest) =>
      evaluatorsApi.createEvaluator(projectId, data),
    onSuccess: (newEvaluator) => {
      queryClient.invalidateQueries({
        queryKey: evaluatorQueryKeys.all,
      })
      toast.success('Evaluator Created', {
        description: `"${newEvaluator.name}" has been created successfully.`,
      })
    },
    onError: (error: unknown) => {
      toast.error('Failed to Create Evaluator', {
        description: extractErrorMessage(
          error,
          'Could not create evaluator. Please try again.'
        ),
      })
    },
  })
}

/**
 * Mutation hook for updating an existing evaluator
 */
export function useUpdateEvaluatorMutation(
  projectId: string,
  evaluatorId: string
) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: UpdateEvaluatorRequest) =>
      evaluatorsApi.updateEvaluator(projectId, evaluatorId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: evaluatorQueryKeys.all,
      })
      toast.success('Evaluator Updated', {
        description: 'Evaluator has been updated.',
      })
    },
    onError: (error: unknown) => {
      toast.error('Failed to Update Evaluator', {
        description: extractErrorMessage(
          error,
          'Could not update evaluator. Please try again.'
        ),
      })
    },
  })
}

/**
 * Mutation hook for deleting an evaluator
 */
export function useDeleteEvaluatorMutation(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      evaluatorId,
      evaluatorName,
    }: {
      evaluatorId: string
      evaluatorName: string
    }) => {
      await evaluatorsApi.deleteEvaluator(projectId, evaluatorId)
      return { evaluatorId, evaluatorName }
    },
    onMutate: async ({ evaluatorId }) => {
      await queryClient.cancelQueries({
        queryKey: evaluatorQueryKeys.all,
      })

      const previousEvaluators = queryClient.getQueryData<EvaluatorsResponse>(
        evaluatorQueryKeys.list(projectId)
      )

      // Optimistic update
      if (previousEvaluators) {
        queryClient.setQueryData<EvaluatorsResponse>(
          evaluatorQueryKeys.list(projectId),
          {
            ...previousEvaluators,
            evaluators: previousEvaluators.evaluators.filter((e) => e.id !== evaluatorId),
            totalCount: previousEvaluators.totalCount - 1,
          }
        )
      }

      return { previousEvaluators }
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: evaluatorQueryKeys.all,
      })
      toast.success('Evaluator Deleted', {
        description: `"${variables.evaluatorName}" has been deleted.`,
      })
    },
    onError: (error: unknown, _variables, context) => {
      if (context?.previousEvaluators) {
        queryClient.setQueryData(
          evaluatorQueryKeys.list(projectId),
          context.previousEvaluators
        )
      }
      toast.error('Failed to Delete Evaluator', {
        description: extractErrorMessage(
          error,
          'Could not delete evaluator. Please try again.'
        ),
      })
    },
  })
}

/**
 * Mutation hook for activating an evaluator
 */
export function useActivateEvaluatorMutation(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      evaluatorId,
      evaluatorName,
    }: {
      evaluatorId: string
      evaluatorName: string
    }) => {
      await evaluatorsApi.activateEvaluator(projectId, evaluatorId)
      return { evaluatorId, evaluatorName }
    },
    onMutate: async ({ evaluatorId }) => {
      await queryClient.cancelQueries({
        queryKey: evaluatorQueryKeys.all,
      })

      // Optimistic update for the detail query
      const previousEvaluator = queryClient.getQueryData<Evaluator>(
        evaluatorQueryKeys.detail(projectId, evaluatorId)
      )

      if (previousEvaluator) {
        queryClient.setQueryData<Evaluator>(
          evaluatorQueryKeys.detail(projectId, evaluatorId),
          { ...previousEvaluator, status: 'active' }
        )
      }

      return { previousEvaluator }
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: evaluatorQueryKeys.all,
      })
      toast.success('Evaluator Activated', {
        description: `"${variables.evaluatorName}" is now active and will automatically score matching spans.`,
      })
    },
    onError: (error: unknown, variables, context) => {
      if (context?.previousEvaluator) {
        queryClient.setQueryData(
          evaluatorQueryKeys.detail(projectId, variables.evaluatorId),
          context.previousEvaluator
        )
      }
      toast.error('Failed to Activate Evaluator', {
        description: extractErrorMessage(
          error,
          'Could not activate evaluator. Please try again.'
        ),
      })
    },
  })
}

/**
 * Mutation hook for deactivating an evaluator
 */
export function useDeactivateEvaluatorMutation(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      evaluatorId,
      evaluatorName,
    }: {
      evaluatorId: string
      evaluatorName: string
    }) => {
      await evaluatorsApi.deactivateEvaluator(projectId, evaluatorId)
      return { evaluatorId, evaluatorName }
    },
    onMutate: async ({ evaluatorId }) => {
      await queryClient.cancelQueries({
        queryKey: evaluatorQueryKeys.all,
      })

      // Optimistic update for the detail query
      const previousEvaluator = queryClient.getQueryData<Evaluator>(
        evaluatorQueryKeys.detail(projectId, evaluatorId)
      )

      if (previousEvaluator) {
        queryClient.setQueryData<Evaluator>(
          evaluatorQueryKeys.detail(projectId, evaluatorId),
          { ...previousEvaluator, status: 'inactive' }
        )
      }

      return { previousEvaluator }
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: evaluatorQueryKeys.all,
      })
      toast.success('Evaluator Deactivated', {
        description: `"${variables.evaluatorName}" is now inactive and will no longer score spans.`,
      })
    },
    onError: (error: unknown, variables, context) => {
      if (context?.previousEvaluator) {
        queryClient.setQueryData(
          evaluatorQueryKeys.detail(projectId, variables.evaluatorId),
          context.previousEvaluator
        )
      }
      toast.error('Failed to Deactivate Evaluator', {
        description: extractErrorMessage(
          error,
          'Could not deactivate evaluator. Please try again.'
        ),
      })
    },
  })
}

/**
 * Mutation hook for manually triggering an evaluator
 * Starts async evaluation against matching spans
 */
export function useTriggerEvaluatorMutation(
  projectId: string,
  evaluatorId: string
) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (options?: TriggerOptions) =>
      evaluatorsApi.triggerEvaluator(projectId, evaluatorId, options),
    onSuccess: (response) => {
      // Invalidate executions to show the new pending execution
      queryClient.invalidateQueries({
        queryKey: evaluatorExecutionsKeys.all,
      })
      toast.success('Evaluation Triggered', {
        description: response.message || 'Evaluation has been queued for processing.',
      })
    },
    onError: (error: unknown) => {
      toast.error('Failed to Trigger Evaluation', {
        description: extractErrorMessage(
          error,
          'Could not trigger evaluation. Please try again.'
        ),
      })
    },
  })
}
