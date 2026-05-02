import { rawFetch } from '@/lib/api/client'
import type { PaginatedResponse } from '@/lib/api/core/types'
import type {
  Dataset,
  CreateDatasetRequest,
  UpdateDatasetRequest,
  DatasetItem,
  CreateDatasetItemRequest,
  DatasetListParams,
  DatasetItemListParams,
  BulkImportResult,
  ImportFromJsonRequest,
  ImportFromTracesRequest,
  ImportFromSpansRequest,
  ImportFromCsvRequest,
  DatasetVersion,
  DatasetWithVersionInfo,
  CreateDatasetVersionRequest,
  PinDatasetVersionRequest,
  DatasetWithItemCount,
} from '../types'

// Backend evaluation list envelope: `{data, total, page, limit}` flat.
// We reshape into the web/-style PaginatedResponse so hooks/components
// can be ported verbatim.
interface FlatList<T> {
  data: T[]
  total: number
  page: number
  limit: number
}

function toPaginated<T>(flat: FlatList<T>): PaginatedResponse<T> {
  const limit = flat.limit || 1
  const total = flat.total ?? 0
  const page = flat.page || 1
  const totalPages = Math.max(1, Math.ceil(total / limit))
  return {
    data: flat.data ?? [],
    pagination: {
      page,
      limit,
      total,
      totalPages,
      hasNext: page < totalPages,
      hasPrev: page > 1,
    },
  }
}

function buildQuery(params: Record<string, unknown>): string {
  const sp = new URLSearchParams()
  for (const [k, v] of Object.entries(params)) {
    if (v === undefined || v === null || v === '') continue
    sp.set(k, String(v))
  }
  const s = sp.toString()
  return s ? `?${s}` : ''
}

async function getJSON<T>(path: string): Promise<T> {
  const resp = await rawFetch(path, { method: 'GET' })
  return (await resp.json()) as T
}

async function postJSON<T>(path: string, body?: unknown): Promise<T> {
  const resp = await rawFetch(path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: body !== undefined ? JSON.stringify(body) : undefined,
  })
  const text = await resp.text()
  return (text ? JSON.parse(text) : undefined) as T
}

async function putJSON<T>(path: string, body?: unknown): Promise<T> {
  const resp = await rawFetch(path, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: body !== undefined ? JSON.stringify(body) : undefined,
  })
  const text = await resp.text()
  return (text ? JSON.parse(text) : undefined) as T
}

async function del(path: string): Promise<void> {
  await rawFetch(path, { method: 'DELETE' })
}

const base = '/api'

export const datasetsApi = {
  listDatasets: async (
    projectId: string,
    params: DatasetListParams = {},
  ): Promise<PaginatedResponse<DatasetWithItemCount>> => {
    const { search, page = 1, limit = 50, sortBy = 'updated_at', sortDir = 'desc' } = params
    const q = buildQuery({ page, limit, sort_by: sortBy, sort_dir: sortDir, search })
    const flat = await getJSON<FlatList<DatasetWithItemCount>>(
      `${base}/v1/projects/${projectId}/datasets${q}`,
    )
    return toPaginated(flat)
  },

  getDataset: (projectId: string, datasetId: string): Promise<Dataset> =>
    getJSON<Dataset>(`${base}/v1/projects/${projectId}/datasets/${datasetId}`),

  createDataset: (projectId: string, data: CreateDatasetRequest): Promise<Dataset> =>
    postJSON<Dataset>(`${base}/v1/projects/${projectId}/datasets`, data),

  updateDataset: (
    projectId: string,
    datasetId: string,
    data: UpdateDatasetRequest,
  ): Promise<Dataset> =>
    putJSON<Dataset>(`${base}/v1/projects/${projectId}/datasets/${datasetId}`, data),

  deleteDataset: (projectId: string, datasetId: string): Promise<void> =>
    del(`${base}/v1/projects/${projectId}/datasets/${datasetId}`),

  listDatasetItems: async (
    projectId: string,
    datasetId: string,
    params?: DatasetItemListParams,
  ): Promise<PaginatedResponse<DatasetItem>> => {
    const q = buildQuery({ page: params?.page, limit: params?.limit })
    const flat = await getJSON<FlatList<DatasetItem>>(
      `${base}/v1/projects/${projectId}/datasets/${datasetId}/items${q}`,
    )
    return toPaginated(flat)
  },

  createDatasetItem: (
    projectId: string,
    datasetId: string,
    data: CreateDatasetItemRequest,
  ): Promise<DatasetItem> =>
    postJSON<DatasetItem>(
      `${base}/v1/projects/${projectId}/datasets/${datasetId}/items`,
      data,
    ),

  deleteDatasetItem: (
    projectId: string,
    datasetId: string,
    itemId: string,
  ): Promise<void> =>
    del(`${base}/v1/projects/${projectId}/datasets/${datasetId}/items/${itemId}`),

  importFromJson: (
    projectId: string,
    datasetId: string,
    data: ImportFromJsonRequest,
  ): Promise<BulkImportResult> =>
    postJSON<BulkImportResult>(
      `${base}/v1/projects/${projectId}/datasets/${datasetId}/items/import-json`,
      data,
    ),

  importFromTraces: (
    projectId: string,
    datasetId: string,
    data: ImportFromTracesRequest,
  ): Promise<BulkImportResult> =>
    postJSON<BulkImportResult>(
      `${base}/v1/projects/${projectId}/datasets/${datasetId}/items/from-traces`,
      data,
    ),

  importFromSpans: (
    projectId: string,
    datasetId: string,
    data: ImportFromSpansRequest,
  ): Promise<BulkImportResult> =>
    postJSON<BulkImportResult>(
      `${base}/v1/projects/${projectId}/datasets/${datasetId}/items/from-spans`,
      data,
    ),

  importFromCsv: (
    projectId: string,
    datasetId: string,
    data: ImportFromCsvRequest,
  ): Promise<BulkImportResult> =>
    postJSON<BulkImportResult>(
      `${base}/v1/projects/${projectId}/datasets/${datasetId}/items/import-csv`,
      data,
    ),

  exportDataset: (projectId: string, datasetId: string): Promise<DatasetItem[]> =>
    getJSON<DatasetItem[]>(
      `${base}/v1/projects/${projectId}/datasets/${datasetId}/items/export`,
    ),

  getDatasetWithVersionInfo: (
    projectId: string,
    datasetId: string,
  ): Promise<DatasetWithVersionInfo> =>
    getJSON<DatasetWithVersionInfo>(
      `${base}/v1/projects/${projectId}/datasets/${datasetId}/info`,
    ),

  listVersions: (projectId: string, datasetId: string): Promise<DatasetVersion[]> =>
    getJSON<DatasetVersion[]>(
      `${base}/v1/projects/${projectId}/datasets/${datasetId}/versions`,
    ),

  getVersion: (
    projectId: string,
    datasetId: string,
    versionId: string,
  ): Promise<DatasetVersion> =>
    getJSON<DatasetVersion>(
      `${base}/v1/projects/${projectId}/datasets/${datasetId}/versions/${versionId}`,
    ),

  createVersion: (
    projectId: string,
    datasetId: string,
    data?: CreateDatasetVersionRequest,
  ): Promise<DatasetVersion> =>
    postJSON<DatasetVersion>(
      `${base}/v1/projects/${projectId}/datasets/${datasetId}/versions`,
      data ?? {},
    ),

  getVersionItems: async (
    projectId: string,
    datasetId: string,
    versionId: string,
    params?: DatasetItemListParams,
  ): Promise<PaginatedResponse<DatasetItem>> => {
    const q = buildQuery({ page: params?.page, limit: params?.limit })
    const flat = await getJSON<FlatList<DatasetItem>>(
      `${base}/v1/projects/${projectId}/datasets/${datasetId}/versions/${versionId}/items${q}`,
    )
    return toPaginated(flat)
  },

  pinVersion: (
    projectId: string,
    datasetId: string,
    data: PinDatasetVersionRequest,
  ): Promise<Dataset> =>
    postJSON<Dataset>(
      `${base}/v1/projects/${projectId}/datasets/${datasetId}/pin`,
      data,
    ),

  unpinVersion: (projectId: string, datasetId: string): Promise<Dataset> =>
    postJSON<Dataset>(`${base}/v1/projects/${projectId}/datasets/${datasetId}/pin`, {
      version_id: null,
    }),
}
