import { queryOptions } from '@tanstack/react-query'
import { rawFetch } from '@/lib/api/client'
import type {
  CreateEvaluatorRequest,
  EvaluatorDetail,
  EvaluatorListResponse,
  TestEvaluatorRequest,
  TestEvaluatorResponse,
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

// Activate/deactivate hit dedicated POST endpoints rather than mutating
// `status` via PUT. The backend emits the updated evaluator in the
// response envelope just like the update path.
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
// writing scores. Request is always POST — the backend accepts an
// empty body for "use defaults" so the UI can keep the form optional.
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
