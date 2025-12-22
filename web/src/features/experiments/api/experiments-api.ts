import { BrokleAPIClient } from '@/lib/api/core/client'
import type {
  Experiment,
  CreateExperimentRequest,
  UpdateExperimentRequest,
  ExperimentItemListResponse,
} from '../types'

const client = new BrokleAPIClient('/api')

export const experimentsApi = {
  // Experiments
  listExperiments: async (
    projectId: string,
    params?: { dataset_id?: string; status?: string }
  ): Promise<Experiment[]> => {
    const queryParams = new URLSearchParams()
    if (params?.dataset_id) queryParams.set('dataset_id', params.dataset_id)
    if (params?.status) queryParams.set('status', params.status)
    const query = queryParams.toString()
    const url = `/v1/projects/${projectId}/experiments${query ? `?${query}` : ''}`
    return client.get<Experiment[]>(url)
  },

  getExperiment: async (
    projectId: string,
    experimentId: string
  ): Promise<Experiment> => {
    return client.get<Experiment>(
      `/v1/projects/${projectId}/experiments/${experimentId}`
    )
  },

  createExperiment: async (
    projectId: string,
    data: CreateExperimentRequest
  ): Promise<Experiment> => {
    return client.post<Experiment>(
      `/v1/projects/${projectId}/experiments`,
      data
    )
  },

  updateExperiment: async (
    projectId: string,
    experimentId: string,
    data: UpdateExperimentRequest
  ): Promise<Experiment> => {
    return client.put<Experiment>(
      `/v1/projects/${projectId}/experiments/${experimentId}`,
      data
    )
  },

  deleteExperiment: async (
    projectId: string,
    experimentId: string
  ): Promise<void> => {
    await client.delete(`/v1/projects/${projectId}/experiments/${experimentId}`)
  },

  // Experiment Items (read-only from dashboard)
  listExperimentItems: async (
    projectId: string,
    experimentId: string,
    limit = 50,
    offset = 0
  ): Promise<ExperimentItemListResponse> => {
    return client.get<ExperimentItemListResponse>(
      `/v1/projects/${projectId}/experiments/${experimentId}/items?limit=${limit}&offset=${offset}`
    )
  },
}
