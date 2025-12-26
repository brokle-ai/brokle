import { BrokleAPIClient } from '@/lib/api/core/client'
import type {
  Dataset,
  CreateDatasetRequest,
  UpdateDatasetRequest,
  DatasetItem,
  CreateDatasetItemRequest,
  DatasetItemListResponse,
} from '../types'

const client = new BrokleAPIClient('/api')

export const datasetsApi = {
  // Datasets
  listDatasets: async (projectId: string): Promise<Dataset[]> => {
    return client.get<Dataset[]>(`/v1/projects/${projectId}/datasets`)
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
    limit = 50,
    offset = 0
  ): Promise<DatasetItemListResponse> => {
    return client.get<DatasetItemListResponse>(
      `/v1/projects/${projectId}/datasets/${datasetId}/items?limit=${limit}&offset=${offset}`
    )
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
}
