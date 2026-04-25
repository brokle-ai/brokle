import { queryOptions } from '@tanstack/react-query'
import { rawFetch } from '@/lib/api/client'
import type {
  CreateEvaluatorRequest,
  EvaluatorAnalyticsParams,
  EvaluatorAnalyticsResponse,
  EvaluatorDetail,
  EvaluatorExecution,
  EvaluatorExecutionDetail,
  EvaluatorListResponse,
  ExecutionListParams,
  ExecutionListResponse,
  TestEvaluatorRequest,
  TestEvaluatorResponse,
  TriggerOptions,
  TriggerResponse,
  UpdateEvaluatorRequest,
} from './types'

export const evaluatorsKeys = {
  all: ['evaluators'] as const,
  lists: () => [...evaluatorsKeys.all, 'list'] as const,
  list: (projectId: string, params: EvaluatorListParams) =>
    [...evaluatorsKeys.lists(), projectId, params] as const,
  details: () => [...evaluatorsKeys.all, 'detail'] as const,
  detail: (evaluatorId: string) =>
    [...evaluatorsKeys.details(), evaluatorId] as const,
  // Execution sub-namespace lives under the evaluator key so the parent
  // detail invalidation also drops execution data when desired.
  executions: () => [...evaluatorsKeys.all, 'executions'] as const,
  executionList: (
    projectId: string,
    evaluatorId: string,
    params: ExecutionListParams,
  ) =>
    [
      ...evaluatorsKeys.executions(),
      'list',
      projectId,
      evaluatorId,
      params,
    ] as const,
  executionDetail: (
    projectId: string,
    evaluatorId: string,
    executionId: string,
  ) =>
    [
      ...evaluatorsKeys.executions(),
      'detail',
      projectId,
      evaluatorId,
      executionId,
    ] as const,
  executionLatest: (projectId: string, evaluatorId: string) =>
    [...evaluatorsKeys.executions(), 'latest', projectId, evaluatorId] as const,
  analytics: (
    projectId: string,
    evaluatorId: string,
    params: EvaluatorAnalyticsParams,
  ) =>
    [
      ...evaluatorsKeys.all,
      'analytics',
      projectId,
      evaluatorId,
      params,
    ] as const,
} as const

export interface EvaluatorListParams {
  page: number
  limit: number
  q?: string
}

export const evaluatorListQueryOptions = (
  projectId: string,
  params: EvaluatorListParams,
) =>
  queryOptions({
    queryKey: evaluatorsKeys.list(projectId, params),
    queryFn: async () => {
      const search = new URLSearchParams()
      search.set('page', String(params.page))
      search.set('limit', String(params.limit))
      if (params.q && params.q.length > 0) {
        search.set('search', params.q)
      }
      const resp = await rawFetch(
        `/api/v1/projects/${projectId}/evaluators?${search.toString()}`,
        { method: 'GET' },
      )
      return (await resp.json()) as EvaluatorListResponse
    },
    staleTime: 30 * 1000,
  })

export const evaluatorDetailQueryOptions = (
  projectId: string,
  evaluatorId: string,
) =>
  queryOptions({
    queryKey: evaluatorsKeys.detail(evaluatorId),
    queryFn: async () => {
      const resp = await rawFetch(
        `/api/v1/projects/${projectId}/evaluators/${evaluatorId}`,
        { method: 'GET' },
      )
      return (await resp.json()) as EvaluatorDetail
    },
    staleTime: 30 * 1000,
  })

// Mutation functions. Backend uses PUT for update (not PATCH) — see
// internal/transport/http/handlers/evaluation/evaluator.go. Activate /
// deactivate are separate POST endpoints, not PATCH on status.
export async function createEvaluator(
  projectId: string,
  data: CreateEvaluatorRequest,
): Promise<EvaluatorDetail> {
  const resp = await rawFetch(`/api/v1/projects/${projectId}/evaluators`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  return (await resp.json()) as EvaluatorDetail
}

export async function updateEvaluator(
  projectId: string,
  evaluatorId: string,
  data: UpdateEvaluatorRequest,
): Promise<EvaluatorDetail> {
  const resp = await rawFetch(
    `/api/v1/projects/${projectId}/evaluators/${evaluatorId}`,
    {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    },
  )
  return (await resp.json()) as EvaluatorDetail
}

export async function deleteEvaluator(
  projectId: string,
  evaluatorId: string,
): Promise<void> {
  await rawFetch(`/api/v1/projects/${projectId}/evaluators/${evaluatorId}`, {
    method: 'DELETE',
  })
}

export async function activateEvaluator(
  projectId: string,
  evaluatorId: string,
): Promise<EvaluatorDetail> {
  const resp = await rawFetch(
    `/api/v1/projects/${projectId}/evaluators/${evaluatorId}/activate`,
    { method: 'POST' },
  )
  return (await resp.json()) as EvaluatorDetail
}

export async function deactivateEvaluator(
  projectId: string,
  evaluatorId: string,
): Promise<EvaluatorDetail> {
  const resp = await rawFetch(
    `/api/v1/projects/${projectId}/evaluators/${evaluatorId}/deactivate`,
    { method: 'POST' },
  )
  return (await resp.json()) as EvaluatorDetail
}

// Test runs evaluate the configuration against real spans without
// writing scores.
export async function testEvaluator(
  projectId: string,
  evaluatorId: string,
  data: TestEvaluatorRequest,
): Promise<TestEvaluatorResponse> {
  const resp = await rawFetch(
    `/api/v1/projects/${projectId}/evaluators/${evaluatorId}/test`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    },
  )
  return (await resp.json()) as TestEvaluatorResponse
}

// ============================================================================
// Executions
// ============================================================================

export const executionListQueryOptions = (
  projectId: string,
  evaluatorId: string,
  params: ExecutionListParams,
  refetchInterval?: number | false,
) =>
  queryOptions({
    queryKey: evaluatorsKeys.executionList(projectId, evaluatorId, params),
    queryFn: async () => {
      const search = new URLSearchParams()
      if (params.page) search.set('page', String(params.page))
      if (params.limit) search.set('limit', String(params.limit))
      if (params.status) search.set('status', params.status)
      if (params.trigger_type) search.set('trigger_type', params.trigger_type)
      const qs = search.toString()
      const resp = await rawFetch(
        `/api/v1/projects/${projectId}/evaluators/${evaluatorId}/executions${
          qs ? `?${qs}` : ''
        }`,
        { method: 'GET' },
      )
      return (await resp.json()) as ExecutionListResponse
    },
    staleTime: 10 * 1000,
    refetchInterval,
  })

export const executionDetailQueryOptions = (
  projectId: string,
  evaluatorId: string,
  executionId: string,
) =>
  queryOptions({
    queryKey: evaluatorsKeys.executionDetail(
      projectId,
      evaluatorId,
      executionId,
    ),
    queryFn: async () => {
      const resp = await rawFetch(
        `/api/v1/projects/${projectId}/evaluators/${evaluatorId}/executions/${executionId}/detail`,
        { method: 'GET' },
      )
      return (await resp.json()) as EvaluatorExecutionDetail
    },
    staleTime: 60 * 1000,
  })

export const executionLatestQueryOptions = (
  projectId: string,
  evaluatorId: string,
) =>
  queryOptions({
    queryKey: evaluatorsKeys.executionLatest(projectId, evaluatorId),
    queryFn: async () => {
      const resp = await rawFetch(
        `/api/v1/projects/${projectId}/evaluators/${evaluatorId}/executions/latest`,
        { method: 'GET' },
      )
      return (await resp.json()) as EvaluatorExecution
    },
    staleTime: 10 * 1000,
  })

// ============================================================================
// Analytics
// ============================================================================

export const evaluatorAnalyticsQueryOptions = (
  projectId: string,
  evaluatorId: string,
  params: EvaluatorAnalyticsParams,
) =>
  queryOptions({
    queryKey: evaluatorsKeys.analytics(projectId, evaluatorId, params),
    queryFn: async () => {
      const search = new URLSearchParams()
      if (params.period) search.set('period', params.period)
      if (params.from_timestamp)
        search.set('from_timestamp', params.from_timestamp)
      if (params.to_timestamp) search.set('to_timestamp', params.to_timestamp)
      const qs = search.toString()
      const resp = await rawFetch(
        `/api/v1/projects/${projectId}/evaluators/${evaluatorId}/analytics${
          qs ? `?${qs}` : ''
        }`,
        { method: 'GET' },
      )
      return (await resp.json()) as EvaluatorAnalyticsResponse
    },
    staleTime: 30 * 1000,
  })

// ============================================================================
// Manual trigger
// ============================================================================

export async function triggerEvaluator(
  projectId: string,
  evaluatorId: string,
  options: TriggerOptions = {},
): Promise<TriggerResponse> {
  const resp = await rawFetch(
    `/api/v1/projects/${projectId}/evaluators/${evaluatorId}/trigger`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(options),
    },
  )
  return (await resp.json()) as TriggerResponse
}
