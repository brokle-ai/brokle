import { queryOptions } from '@tanstack/react-query'
import { rawFetch } from '@/lib/api/client'
import type {
  CreateDatasetRequest,
  DatasetDetail,
  DatasetItemsListResponse,
  DatasetListResponse,
  UpdateDatasetRequest,
} from './types'

// TkDodo-style hierarchical query keys. `list(projectId, params)`
// invalidates cleanly via `lists()` once the detail-view port adds
// mutations that should bust the list cache.
export const datasetsKeys = {
  all: ['datasets'] as const,
  lists: () => [...datasetsKeys.all, 'list'] as const,
  list: (projectId: string, params: DatasetListParams) =>
    [...datasetsKeys.lists(), projectId, params] as const,
  details: () => [...datasetsKeys.all, 'detail'] as const,
  detail: (datasetId: string) => [...datasetsKeys.details(), datasetId] as const,
} as const

export interface DatasetListParams {
  page: number
  limit: number
  q?: string
}

export const datasetListQueryOptions = (
  projectId: string,
  params: DatasetListParams,
) =>
  queryOptions({
    queryKey: datasetsKeys.list(projectId, params),
    queryFn: async () => {
      const search = new URLSearchParams()
      search.set('page', String(params.page))
      search.set('limit', String(params.limit))
      if (params.q && params.q.length > 0) {
        // Backend accepts `search` as the free-text query param; keep
        // the user-facing alias `q` in the URL for familiarity.
        search.set('search', params.q)
      }
      const resp = await rawFetch(
        `/api/v1/projects/${projectId}/datasets?${search.toString()}`,
        { method: 'GET' },
      )
      return (await resp.json()) as DatasetListResponse
    },
    // Datasets change infrequently (manual create/edit) — 30s keeps
    // in-tab navigation snappy without masking fresh changes.
    staleTime: 30 * 1000,
  })

export const datasetDetailQueryOptions = (
  projectId: string,
  datasetId: string,
) =>
  queryOptions({
    queryKey: datasetsKeys.detail(datasetId),
    queryFn: async () => {
      const resp = await rawFetch(
        `/api/v1/projects/${projectId}/datasets/${datasetId}`,
        { method: 'GET' },
      )
      return (await resp.json()) as DatasetDetail
    },
    staleTime: 30 * 1000,
  })

export interface DatasetItemsListParams {
  page: number
  limit: number
}

// Dataset items sub-key — nested under `detail(datasetId)` so cache
// invalidation on a dataset also buckets items alongside the detail.
export const datasetItemsKey = (datasetId: string, params: DatasetItemsListParams) =>
  [...datasetsKeys.detail(datasetId), 'items', params] as const

export const datasetItemsListQueryOptions = (
  projectId: string,
  datasetId: string,
  params: DatasetItemsListParams,
) =>
  queryOptions({
    queryKey: datasetItemsKey(datasetId, params),
    queryFn: async () => {
      const search = new URLSearchParams()
      search.set('page', String(params.page))
      search.set('limit', String(params.limit))
      const resp = await rawFetch(
        `/api/v1/projects/${projectId}/datasets/${datasetId}/items?${search.toString()}`,
        { method: 'GET' },
      )
      return (await resp.json()) as DatasetItemsListResponse
    },
    staleTime: 15 * 1000,
  })

// Mutation functions — kept thin so callers can compose useMutation +
// onSuccess/navigate/invalidate in their own component.
export async function createDataset(
  projectId: string,
  data: CreateDatasetRequest,
): Promise<DatasetDetail> {
  const resp = await rawFetch(`/api/v1/projects/${projectId}/datasets`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  return (await resp.json()) as DatasetDetail
}

// Backend accepts PUT (not PATCH) — matches the update-dataset Huma
// operation at internal/transport/http/handlers/evaluation/dataset.go.
// Semantically still a partial update since the handler reads
// `UpdateDatasetRequest` with `omitempty` pointer fields.
export async function updateDataset(
  projectId: string,
  datasetId: string,
  data: UpdateDatasetRequest,
): Promise<DatasetDetail> {
  const resp = await rawFetch(
    `/api/v1/projects/${projectId}/datasets/${datasetId}`,
    {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    },
  )
  return (await resp.json()) as DatasetDetail
}

export async function deleteDataset(
  projectId: string,
  datasetId: string,
): Promise<void> {
  await rawFetch(`/api/v1/projects/${projectId}/datasets/${datasetId}`, {
    method: 'DELETE',
  })
}
