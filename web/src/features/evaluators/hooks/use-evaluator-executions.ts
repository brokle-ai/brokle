'use client'

import { useQuery } from '@tanstack/react-query'
import { evaluatorsApi } from '../api/evaluators-api'
import type { ExecutionListParams, ExecutionStatus, TriggerType } from '../types'

/**
 * Query key factory for evaluator executions
 */
export const evaluatorExecutionsKeys = {
  all: ['evaluator-executions'] as const,
  lists: () => [...evaluatorExecutionsKeys.all, 'list'] as const,
  list: (projectId: string, evaluatorId: string, params?: ExecutionListParams) =>
    [...evaluatorExecutionsKeys.lists(), projectId, evaluatorId, params] as const,
  details: () => [...evaluatorExecutionsKeys.all, 'detail'] as const,
  detail: (projectId: string, evaluatorId: string, executionId: string) =>
    [...evaluatorExecutionsKeys.details(), projectId, evaluatorId, executionId] as const,
  latest: (projectId: string, evaluatorId: string) =>
    [...evaluatorExecutionsKeys.all, 'latest', projectId, evaluatorId] as const,
}

interface UseEvaluatorExecutionsQueryOptions {
  page?: number
  limit?: number
  status?: ExecutionStatus
  triggerType?: TriggerType
  enabled?: boolean
  refetchInterval?: number | false
}

/**
 * Hook to fetch execution history for an evaluator
 */
export function useEvaluatorExecutionsQuery(
  projectId: string,
  evaluatorId: string,
  options: UseEvaluatorExecutionsQueryOptions = {}
) {
  const { page = 1, limit = 25, status, triggerType, enabled = true, refetchInterval } = options

  const params: ExecutionListParams = {
    page,
    limit,
    ...(status && { status }),
    ...(triggerType && { trigger_type: triggerType }),
  }

  return useQuery({
    queryKey: evaluatorExecutionsKeys.list(projectId, evaluatorId, params),
    queryFn: () => evaluatorsApi.listExecutions(projectId, evaluatorId, params),
    enabled: enabled && !!projectId && !!evaluatorId,
    refetchInterval,
    staleTime: 10_000, // 10 seconds - executions change frequently
  })
}

/**
 * Hook to fetch a specific execution by ID
 */
export function useEvaluatorExecutionQuery(
  projectId: string,
  evaluatorId: string,
  executionId: string,
  options: { enabled?: boolean } = {}
) {
  const { enabled = true } = options

  return useQuery({
    queryKey: evaluatorExecutionsKeys.detail(projectId, evaluatorId, executionId),
    queryFn: () => evaluatorsApi.getExecution(projectId, evaluatorId, executionId),
    enabled: enabled && !!projectId && !!evaluatorId && !!executionId,
    staleTime: 30_000, // 30 seconds
  })
}

/**
 * Hook to fetch detailed execution info including span-level debugging data
 */
export function useEvaluatorExecutionDetailQuery(
  projectId: string,
  evaluatorId: string,
  executionId: string,
  options: { enabled?: boolean } = {}
) {
  const { enabled = true } = options

  return useQuery({
    queryKey: [...evaluatorExecutionsKeys.detail(projectId, evaluatorId, executionId), 'full'] as const,
    queryFn: () => evaluatorsApi.getExecutionDetail(projectId, evaluatorId, executionId),
    enabled: enabled && !!projectId && !!evaluatorId && !!executionId,
    staleTime: 60_000, // 1 minute - detailed data doesn't change
  })
}

/**
 * Hook to fetch the latest execution for an evaluator
 */
export function useLatestEvaluatorExecutionQuery(
  projectId: string,
  evaluatorId: string,
  options: { enabled?: boolean; refetchInterval?: number | false } = {}
) {
  const { enabled = true, refetchInterval } = options

  return useQuery({
    queryKey: evaluatorExecutionsKeys.latest(projectId, evaluatorId),
    queryFn: () => evaluatorsApi.getLatestExecution(projectId, evaluatorId),
    enabled: enabled && !!projectId && !!evaluatorId,
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
