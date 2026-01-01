'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { extractErrorMessage } from '@/lib/api/error-utils'
import { evaluationRulesApi } from '../api/evaluation-rules-api'
import type {
  CreateEvaluationRuleRequest,
  UpdateEvaluationRuleRequest,
  EvaluationRule,
  RuleListParams,
  RuleListResponse,
  TriggerOptions,
} from '../types'
import { ruleExecutionsKeys } from './use-rule-executions'

export const evaluationRuleQueryKeys = {
  all: ['evaluation-rules'] as const,
  list: (projectId: string, params?: RuleListParams) =>
    [...evaluationRuleQueryKeys.all, 'list', projectId, params] as const,
  detail: (projectId: string, ruleId: string) =>
    [...evaluationRuleQueryKeys.all, 'detail', projectId, ruleId] as const,
}

/**
 * Query hook for listing evaluation rules with pagination and filtering
 */
export function useEvaluationRulesQuery(
  projectId: string | undefined,
  params?: RuleListParams
) {
  return useQuery({
    queryKey: evaluationRuleQueryKeys.list(projectId ?? '', params),
    queryFn: () => evaluationRulesApi.listRules(projectId!, params),
    enabled: !!projectId,
    staleTime: 30_000,
    gcTime: 5 * 60 * 1000,
  })
}

/**
 * Query hook for fetching a single evaluation rule
 */
export function useEvaluationRuleQuery(
  projectId: string | undefined,
  ruleId: string | undefined
) {
  return useQuery({
    queryKey: evaluationRuleQueryKeys.detail(projectId ?? '', ruleId ?? ''),
    queryFn: () => evaluationRulesApi.getRule(projectId!, ruleId!),
    enabled: !!projectId && !!ruleId,
    staleTime: 30_000,
    gcTime: 5 * 60 * 1000,
  })
}

/**
 * Mutation hook for creating a new evaluation rule
 */
export function useCreateEvaluationRuleMutation(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateEvaluationRuleRequest) =>
      evaluationRulesApi.createRule(projectId, data),
    onSuccess: (newRule) => {
      queryClient.invalidateQueries({
        queryKey: evaluationRuleQueryKeys.all,
      })
      toast.success('Evaluation Rule Created', {
        description: `"${newRule.name}" has been created successfully.`,
      })
    },
    onError: (error: unknown) => {
      toast.error('Failed to Create Evaluation Rule', {
        description: extractErrorMessage(
          error,
          'Could not create evaluation rule. Please try again.'
        ),
      })
    },
  })
}

/**
 * Mutation hook for updating an existing evaluation rule
 */
export function useUpdateEvaluationRuleMutation(
  projectId: string,
  ruleId: string
) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: UpdateEvaluationRuleRequest) =>
      evaluationRulesApi.updateRule(projectId, ruleId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: evaluationRuleQueryKeys.all,
      })
      toast.success('Evaluation Rule Updated', {
        description: 'Evaluation rule has been updated.',
      })
    },
    onError: (error: unknown) => {
      toast.error('Failed to Update Evaluation Rule', {
        description: extractErrorMessage(
          error,
          'Could not update evaluation rule. Please try again.'
        ),
      })
    },
  })
}

/**
 * Mutation hook for deleting an evaluation rule
 */
export function useDeleteEvaluationRuleMutation(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      ruleId,
      ruleName,
    }: {
      ruleId: string
      ruleName: string
    }) => {
      await evaluationRulesApi.deleteRule(projectId, ruleId)
      return { ruleId, ruleName }
    },
    onMutate: async ({ ruleId }) => {
      await queryClient.cancelQueries({
        queryKey: evaluationRuleQueryKeys.all,
      })

      const previousRules = queryClient.getQueryData<RuleListResponse>(
        evaluationRuleQueryKeys.list(projectId)
      )

      // Optimistic update
      if (previousRules) {
        queryClient.setQueryData<RuleListResponse>(
          evaluationRuleQueryKeys.list(projectId),
          {
            ...previousRules,
            rules: previousRules.rules.filter((r) => r.id !== ruleId),
            total: previousRules.total - 1,
          }
        )
      }

      return { previousRules }
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: evaluationRuleQueryKeys.all,
      })
      toast.success('Evaluation Rule Deleted', {
        description: `"${variables.ruleName}" has been deleted.`,
      })
    },
    onError: (error: unknown, _variables, context) => {
      if (context?.previousRules) {
        queryClient.setQueryData(
          evaluationRuleQueryKeys.list(projectId),
          context.previousRules
        )
      }
      toast.error('Failed to Delete Evaluation Rule', {
        description: extractErrorMessage(
          error,
          'Could not delete evaluation rule. Please try again.'
        ),
      })
    },
  })
}

/**
 * Mutation hook for activating an evaluation rule
 */
export function useActivateEvaluationRuleMutation(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      ruleId,
      ruleName,
    }: {
      ruleId: string
      ruleName: string
    }) => {
      await evaluationRulesApi.activateRule(projectId, ruleId)
      return { ruleId, ruleName }
    },
    onMutate: async ({ ruleId }) => {
      await queryClient.cancelQueries({
        queryKey: evaluationRuleQueryKeys.all,
      })

      // Optimistic update for the detail query
      const previousRule = queryClient.getQueryData<EvaluationRule>(
        evaluationRuleQueryKeys.detail(projectId, ruleId)
      )

      if (previousRule) {
        queryClient.setQueryData<EvaluationRule>(
          evaluationRuleQueryKeys.detail(projectId, ruleId),
          { ...previousRule, status: 'active' }
        )
      }

      return { previousRule }
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: evaluationRuleQueryKeys.all,
      })
      toast.success('Evaluation Rule Activated', {
        description: `"${variables.ruleName}" is now active and will automatically score matching spans.`,
      })
    },
    onError: (error: unknown, variables, context) => {
      if (context?.previousRule) {
        queryClient.setQueryData(
          evaluationRuleQueryKeys.detail(projectId, variables.ruleId),
          context.previousRule
        )
      }
      toast.error('Failed to Activate Evaluation Rule', {
        description: extractErrorMessage(
          error,
          'Could not activate evaluation rule. Please try again.'
        ),
      })
    },
  })
}

/**
 * Mutation hook for deactivating an evaluation rule
 */
export function useDeactivateEvaluationRuleMutation(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      ruleId,
      ruleName,
    }: {
      ruleId: string
      ruleName: string
    }) => {
      await evaluationRulesApi.deactivateRule(projectId, ruleId)
      return { ruleId, ruleName }
    },
    onMutate: async ({ ruleId }) => {
      await queryClient.cancelQueries({
        queryKey: evaluationRuleQueryKeys.all,
      })

      // Optimistic update for the detail query
      const previousRule = queryClient.getQueryData<EvaluationRule>(
        evaluationRuleQueryKeys.detail(projectId, ruleId)
      )

      if (previousRule) {
        queryClient.setQueryData<EvaluationRule>(
          evaluationRuleQueryKeys.detail(projectId, ruleId),
          { ...previousRule, status: 'inactive' }
        )
      }

      return { previousRule }
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: evaluationRuleQueryKeys.all,
      })
      toast.success('Evaluation Rule Deactivated', {
        description: `"${variables.ruleName}" is now inactive and will no longer score spans.`,
      })
    },
    onError: (error: unknown, variables, context) => {
      if (context?.previousRule) {
        queryClient.setQueryData(
          evaluationRuleQueryKeys.detail(projectId, variables.ruleId),
          context.previousRule
        )
      }
      toast.error('Failed to Deactivate Evaluation Rule', {
        description: extractErrorMessage(
          error,
          'Could not deactivate evaluation rule. Please try again.'
        ),
      })
    },
  })
}

/**
 * Mutation hook for manually triggering an evaluation rule
 * Starts async evaluation against matching spans
 */
export function useTriggerEvaluationRuleMutation(
  projectId: string,
  ruleId: string
) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (options?: TriggerOptions) =>
      evaluationRulesApi.triggerRule(projectId, ruleId, options),
    onSuccess: (response) => {
      // Invalidate executions to show the new pending execution
      queryClient.invalidateQueries({
        queryKey: ruleExecutionsKeys.all,
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
