import { queryOptions } from '@tanstack/react-query'
import { rawFetch } from '@/lib/api/client'
import type {
  CreateExperimentRequest,
  ExperimentDetail,
  ExperimentItemListResponse,
  ExperimentListResponse,
  ExperimentMetricsResponse,
  RerunExperimentRequest,
} from './types'

export const experimentsKeys = {
  all: ['experiments'] as const,
  lists: () => [...experimentsKeys.all, 'list'] as const,
  list: (projectId: string, params: ExperimentListParams) =>
    [...experimentsKeys.lists(), projectId, params] as const,
  details: () => [...experimentsKeys.all, 'detail'] as const,
  detail: (experimentId: string) =>
    [...experimentsKeys.details(), experimentId] as const,
} as const

export interface ExperimentListParams {
  page: number
  limit: number
  q?: string
}

export const experimentListQueryOptions = (
  projectId: string,
  params: ExperimentListParams,
) =>
  queryOptions({
    queryKey: experimentsKeys.list(projectId, params),
    queryFn: async () => {
      const search = new URLSearchParams()
      search.set('page', String(params.page))
      search.set('limit', String(params.limit))
      if (params.q && params.q.length > 0) {
        search.set('search', params.q)
      }
      const resp = await rawFetch(
        `/api/v1/projects/${projectId}/experiments?${search.toString()}`,
        { method: 'GET' },
      )
      return (await resp.json()) as ExperimentListResponse
    },
    // Experiments transition (running → completed) — keep window short
    // so the list reflects progress without aggressive refetch.
    staleTime: 15 * 1000,
  })

export const experimentDetailQueryOptions = (
  projectId: string,
  experimentId: string,
) =>
  queryOptions({
    queryKey: experimentsKeys.detail(experimentId),
    queryFn: async () => {
      const resp = await rawFetch(
        `/api/v1/projects/${projectId}/experiments/${experimentId}`,
        { method: 'GET' },
      )
      return (await resp.json()) as ExperimentDetail
    },
    staleTime: 15 * 1000,
  })

// Metrics endpoint — separate query key so periodic refetch on progress
// doesn't bust the detail-card cache.
export const experimentMetricsKey = (experimentId: string) =>
  [...experimentsKeys.detail(experimentId), 'metrics'] as const

export const experimentMetricsQueryOptions = (
  projectId: string,
  experimentId: string,
) =>
  queryOptions({
    queryKey: experimentMetricsKey(experimentId),
    queryFn: async () => {
      const resp = await rawFetch(
        `/api/v1/projects/${projectId}/experiments/${experimentId}/metrics`,
        { method: 'GET' },
      )
      return (await resp.json()) as ExperimentMetricsResponse
    },
    staleTime: 10 * 1000,
  })

export interface ExperimentItemsListParams {
  limit: number
  offset: number
}

// Experiment items use limit/offset (predates the pageList wrapper used
// elsewhere in evaluation). Derive page-boundary state at the render
// layer the same way.
export const experimentItemsKey = (
  experimentId: string,
  params: ExperimentItemsListParams,
) => [...experimentsKeys.detail(experimentId), 'items', params] as const

export const experimentItemsListQueryOptions = (
  projectId: string,
  experimentId: string,
  params: ExperimentItemsListParams,
) =>
  queryOptions({
    queryKey: experimentItemsKey(experimentId, params),
    queryFn: async () => {
      const search = new URLSearchParams()
      search.set('limit', String(params.limit))
      search.set('offset', String(params.offset))
      const resp = await rawFetch(
        `/api/v1/projects/${projectId}/experiments/${experimentId}/items?${search.toString()}`,
        { method: 'GET' },
      )
      return (await resp.json()) as ExperimentItemListResponse
    },
    staleTime: 10 * 1000,
  })

// Mutation functions. Only create + delete + rerun are wired at the
// dashboard plane; update is also exposed on the backend but the v1
// form has no fields to surface it (name-only editing isn't a workflow
// the product has today).
export async function createExperiment(
  projectId: string,
  data: CreateExperimentRequest,
): Promise<ExperimentDetail> {
  const resp = await rawFetch(`/api/v1/projects/${projectId}/experiments`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  return (await resp.json()) as ExperimentDetail
}

export async function deleteExperiment(
  projectId: string,
  experimentId: string,
): Promise<void> {
  await rawFetch(
    `/api/v1/projects/${projectId}/experiments/${experimentId}`,
    { method: 'DELETE' },
  )
}

// Rerun clones the experiment against the same dataset and returns the
// new experiment (the backend creates a fresh row and queues it for
// execution). The empty body is valid — all fields are optional.
export async function rerunExperiment(
  projectId: string,
  experimentId: string,
  data: RerunExperimentRequest = {},
): Promise<ExperimentDetail> {
  const resp = await rawFetch(
    `/api/v1/projects/${projectId}/experiments/${experimentId}/rerun`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    },
  )
  return (await resp.json()) as ExperimentDetail
}
