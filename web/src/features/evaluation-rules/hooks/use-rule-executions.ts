'use client'

import { useQuery } from '@tanstack/react-query'
import { evaluationRulesApi } from '../api/evaluation-rules-api'
import type { ExecutionListParams, ExecutionStatus, TriggerType } from '../types'

/**
 * Query key factory for rule executions
 */
export const ruleExecutionsKeys = {
  all: ['rule-executions'] as const,
  lists: () => [...ruleExecutionsKeys.all, 'list'] as const,
  list: (projectId: string, ruleId: string, params?: ExecutionListParams) =>
    [...ruleExecutionsKeys.lists(), projectId, ruleId, params] as const,
  details: () => [...ruleExecutionsKeys.all, 'detail'] as const,
  detail: (projectId: string, ruleId: string, executionId: string) =>
    [...ruleExecutionsKeys.details(), projectId, ruleId, executionId] as const,
  latest: (projectId: string, ruleId: string) =>
    [...ruleExecutionsKeys.all, 'latest', projectId, ruleId] as const,
}

interface UseRuleExecutionsQueryOptions {
  page?: number
  limit?: number
  status?: ExecutionStatus
  triggerType?: TriggerType
  enabled?: boolean
  refetchInterval?: number | false
}

/**
 * Hook to fetch execution history for a rule
 */
export function useRuleExecutionsQuery(
  projectId: string,
  ruleId: string,
  options: UseRuleExecutionsQueryOptions = {}
) {
  const { page = 1, limit = 25, status, triggerType, enabled = true, refetchInterval } = options

  const params: ExecutionListParams = {
    page,
    limit,
    ...(status && { status }),
    ...(triggerType && { trigger_type: triggerType }),
  }

  return useQuery({
    queryKey: ruleExecutionsKeys.list(projectId, ruleId, params),
    queryFn: () => evaluationRulesApi.listExecutions(projectId, ruleId, params),
    enabled: enabled && !!projectId && !!ruleId,
    refetchInterval,
    staleTime: 10_000, // 10 seconds - executions change frequently
  })
}

/**
 * Hook to fetch a specific execution by ID
 */
export function useRuleExecutionQuery(
  projectId: string,
  ruleId: string,
  executionId: string,
  options: { enabled?: boolean } = {}
) {
  const { enabled = true } = options

  return useQuery({
    queryKey: ruleExecutionsKeys.detail(projectId, ruleId, executionId),
    queryFn: () => evaluationRulesApi.getExecution(projectId, ruleId, executionId),
    enabled: enabled && !!projectId && !!ruleId && !!executionId,
    staleTime: 30_000, // 30 seconds
  })
}

/**
 * Hook to fetch the latest execution for a rule
 */
export function useLatestRuleExecutionQuery(
  projectId: string,
  ruleId: string,
  options: { enabled?: boolean; refetchInterval?: number | false } = {}
) {
  const { enabled = true, refetchInterval } = options

  return useQuery({
    queryKey: ruleExecutionsKeys.latest(projectId, ruleId),
    queryFn: () => evaluationRulesApi.getLatestExecution(projectId, ruleId),
    enabled: enabled && !!projectId && !!ruleId,
    refetchInterval,
    staleTime: 10_000, // 10 seconds
  })
}

/**
 * Utility to determine if we should poll for updates
 * Returns a refetch interval if any execution is in a non-terminal state
 */
export function getRefetchInterval(hasRunningExecutions: boolean): number | false {
  return hasRunningExecutions ? 5_000 : false // Poll every 5 seconds if running
}
