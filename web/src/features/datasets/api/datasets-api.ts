import { BrokleAPIClient } from '@/lib/api/core/client'
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

const client = new BrokleAPIClient('/api')

/**
 * Response type for datasets list
 */
export interface DatasetsResponse {
  datasets: DatasetWithItemCount[]
  totalCount: number
  page: number
  pageSize: number
  totalPages: number
  hasNext: boolean
  hasPrev: boolean
}

/**
 * Response type for dataset items list
 */
export interface DatasetItemsResponse {
  items: DatasetItem[]
  totalCount: number
  page: number
  pageSize: number
  totalPages: number
  hasNext: boolean
  hasPrev: boolean
}

export const datasetsApi = {
  // Datasets
  listDatasets: async (
    projectId: string,
    params: DatasetListParams = {}
  ): Promise<DatasetsResponse> => {
    const { search, page = 1, limit = 50, sortBy = 'updated_at', sortDir = 'desc' } = params

    const queryParams: Record<string, string | number> = {
      page,
      limit,
      sort_by: sortBy,
      sort_dir: sortDir,
    }

    if (search) {
      queryParams.search = search
    }

    const response = await client.getPaginated<DatasetWithItemCount>(
      `/v1/projects/${projectId}/datasets`,
      queryParams
    )

    return {
      datasets: response.data,
      totalCount: response.pagination.total,
      page: response.pagination.page,
      pageSize: response.pagination.limit,
      totalPages: response.pagination.totalPages,
      hasNext: response.pagination.hasNext,
      hasPrev: response.pagination.hasPrev,
    }
  },

  getDataset: async (projectId: string, datasetId: string): Promise<Dataset> => {
    return client.get<Dataset>(`/v1/projects/${projectId}/datasets/${datasetId}`)
  },

  createDataset: async (
    projectId: string,
    data: CreateDatasetRequest
  ): Promise<Dataset> => {
    return client.post<Dataset>(`/v1/projects/${projectId}/datasets`, data)
  },

  updateDataset: async (
    projectId: string,
    datasetId: string,
    data: UpdateDatasetRequest
  ): Promise<Dataset> => {
    return client.put<Dataset>(
      `/v1/projects/${projectId}/datasets/${datasetId}`,
      data
    )
  },

  deleteDataset: async (projectId: string, datasetId: string): Promise<void> => {
    await client.delete(`/v1/projects/${projectId}/datasets/${datasetId}`)
  },

  // Dataset Items
  listDatasetItems: async (
    projectId: string,
    datasetId: string,
    params?: DatasetItemListParams
  ): Promise<DatasetItemsResponse> => {
    const queryParams: Record<string, string | number> = {}
    if (params?.page) queryParams.page = params.page
    if (params?.limit) queryParams.limit = params.limit

    const response = await client.getPaginated<DatasetItem>(
      `/v1/projects/${projectId}/datasets/${datasetId}/items`,
      queryParams
    )

    return {
      items: response.data,
      totalCount: response.pagination.total,
      page: response.pagination.page,
      pageSize: response.pagination.limit,
      totalPages: response.pagination.totalPages,
      hasNext: response.pagination.hasNext,
      hasPrev: response.pagination.hasPrev,
    }
  },

  createDatasetItem: async (
    projectId: string,
    datasetId: string,
    data: CreateDatasetItemRequest
  ): Promise<DatasetItem> => {
    return client.post<DatasetItem>(
      `/v1/projects/${projectId}/datasets/${datasetId}/items`,
      data
    )
  },

  deleteDatasetItem: async (
    projectId: string,
    datasetId: string,
    itemId: string
  ): Promise<void> => {
    await client.delete(
      `/v1/projects/${projectId}/datasets/${datasetId}/items/${itemId}`
    )
  },

  // Import Methods
  importFromJson: async (
    projectId: string,
    datasetId: string,
    data: ImportFromJsonRequest
  ): Promise<BulkImportResult> => {
    return client.post<BulkImportResult>(
      `/v1/projects/${projectId}/datasets/${datasetId}/items/import-json`,
      data
    )
  },

  importFromTraces: async (
    projectId: string,
    datasetId: string,
    data: ImportFromTracesRequest
  ): Promise<BulkImportResult> => {
    return client.post<BulkImportResult>(
      `/v1/projects/${projectId}/datasets/${datasetId}/items/from-traces`,
      data
    )
  },

  importFromSpans: async (
    projectId: string,
    datasetId: string,
    data: ImportFromSpansRequest
  ): Promise<BulkImportResult> => {
    return client.post<BulkImportResult>(
      `/v1/projects/${projectId}/datasets/${datasetId}/items/from-spans`,
      data
    )
  },

  importFromCsv: async (
    projectId: string,
    datasetId: string,
    data: ImportFromCsvRequest
  ): Promise<BulkImportResult> => {
    return client.post<BulkImportResult>(
      `/v1/projects/${projectId}/datasets/${datasetId}/items/import-csv`,
      data
    )
  },

  // Export Methods
  exportDataset: async (
    projectId: string,
    datasetId: string
  ): Promise<DatasetItem[]> => {
    return client.get<DatasetItem[]>(
      `/v1/projects/${projectId}/datasets/${datasetId}/items/export`
    )
  },

  // Dataset Versioning
  getDatasetWithVersionInfo: async (
    projectId: string,
    datasetId: string
  ): Promise<DatasetWithVersionInfo> => {
    return client.get<DatasetWithVersionInfo>(
      `/v1/projects/${projectId}/datasets/${datasetId}/info`
    )
  },

  listVersions: async (
    projectId: string,
    datasetId: string
  ): Promise<DatasetVersion[]> => {
    return client.get<DatasetVersion[]>(
      `/v1/projects/${projectId}/datasets/${datasetId}/versions`
    )
  },

  getVersion: async (
    projectId: string,
    datasetId: string,
    versionId: string
  ): Promise<DatasetVersion> => {
    return client.get<DatasetVersion>(
      `/v1/projects/${projectId}/datasets/${datasetId}/versions/${versionId}`
    )
  },

  createVersion: async (
    projectId: string,
    datasetId: string,
    data?: CreateDatasetVersionRequest
  ): Promise<DatasetVersion> => {
    return client.post<DatasetVersion>(
      `/v1/projects/${projectId}/datasets/${datasetId}/versions`,
      data ?? {}
    )
  },

  getVersionItems: async (
    projectId: string,
    datasetId: string,
    versionId: string,
    params?: DatasetItemListParams
  ): Promise<DatasetItemsResponse> => {
    const queryParams: Record<string, string | number> = {}
    if (params?.page) queryParams.page = params.page
    if (params?.limit) queryParams.limit = params.limit

    const response = await client.getPaginated<DatasetItem>(
      `/v1/projects/${projectId}/datasets/${datasetId}/versions/${versionId}/items`,
      queryParams
    )

    return {
      items: response.data,
      totalCount: response.pagination.total,
      page: response.pagination.page,
      pageSize: response.pagination.limit,
      totalPages: response.pagination.totalPages,
      hasNext: response.pagination.hasNext,
      hasPrev: response.pagination.hasPrev,
    }
  },

  pinVersion: async (
    projectId: string,
    datasetId: string,
    data: PinDatasetVersionRequest
  ): Promise<Dataset> => {
    return client.post<Dataset>(
      `/v1/projects/${projectId}/datasets/${datasetId}/pin`,
      data
    )
  },

  unpinVersion: async (
    projectId: string,
    datasetId: string
  ): Promise<Dataset> => {
    return client.post<Dataset>(
      `/v1/projects/${projectId}/datasets/${datasetId}/pin`,
      { version_id: null }
    )
  },
}
