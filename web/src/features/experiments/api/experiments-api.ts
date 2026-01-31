import { BrokleAPIClient } from '@/lib/api/core/client'
import type {
  Experiment,
  ExperimentProgress,
  CreateExperimentRequest,
  UpdateExperimentRequest,
  RerunExperimentRequest,
  ExperimentListParams,
  ExperimentItemListResponse,
  ExperimentComparisonResponse,
  CompareExperimentsRequest,
  CreateExperimentFromWizardRequest,
  ValidateStepRequest,
  ValidateStepResponse,
  EstimateCostRequest,
  EstimateCostResponse,
  DatasetFieldsResponse,
  ExperimentConfig,
} from '../types'

const client = new BrokleAPIClient('/api')

/**
 * Response type for experiments list
 */
export interface ExperimentsResponse {
  experiments: Experiment[]
  totalCount: number
  page: number
  pageSize: number
  totalPages: number
  hasNext: boolean
  hasPrev: boolean
}

export const experimentsApi = {
  // Experiments
  listExperiments: async (
    projectId: string,
    params?: ExperimentListParams
  ): Promise<ExperimentsResponse> => {
    const queryParams: Record<string, string | number> = {}
    if (params?.page) queryParams.page = params.page
    if (params?.limit) queryParams.limit = params.limit
    if (params?.dataset_id) queryParams.dataset_id = params.dataset_id
    if (params?.status) queryParams.status = params.status
    if (params?.search) queryParams.search = params.search
    if (params?.ids) queryParams.ids = params.ids

    const response = await client.getPaginated<Experiment>(
      `/v1/projects/${projectId}/experiments`,
      queryParams
    )

    return {
      experiments: response.data,
      totalCount: response.pagination.total,
      page: response.pagination.page,
      pageSize: response.pagination.limit,
      totalPages: response.pagination.totalPages,
      hasNext: response.pagination.hasNext,
      hasPrev: response.pagination.hasPrev,
    }
  },

  getExperiment: async (
    projectId: string,
    experimentId: string
  ): Promise<Experiment> => {
    return client.get<Experiment>(
      `/v1/projects/${projectId}/experiments/${experimentId}`
    )
  },

  getExperimentProgress: async (
    projectId: string,
    experimentId: string
  ): Promise<ExperimentProgress> => {
    return client.get<ExperimentProgress>(
      `/v1/projects/${projectId}/experiments/${experimentId}/progress`
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

  compareExperiments: async (
    projectId: string,
    experimentIds: string[],
    baselineId?: string
  ): Promise<ExperimentComparisonResponse> => {
    const payload: CompareExperimentsRequest = {
      experiment_ids: experimentIds,
      ...(baselineId && { baseline_id: baselineId }),
    }
    return client.post<ExperimentComparisonResponse>(
      `/v1/projects/${projectId}/experiments/compare`,
      payload
    )
  },

  rerunExperiment: async (
    projectId: string,
    experimentId: string,
    data?: RerunExperimentRequest
  ): Promise<Experiment> => {
    return client.post<Experiment>(
      `/v1/projects/${projectId}/experiments/${experimentId}/rerun`,
      data ?? {}
    )
  },

  // ============================================================================
  // Experiment Wizard (Dashboard-only)
  // ============================================================================

  createFromWizard: async (
    projectId: string,
    data: CreateExperimentFromWizardRequest
  ): Promise<Experiment> => {
    return client.post<Experiment>(
      `/v1/projects/${projectId}/experiments/wizard`,
      data
    )
  },

  validateWizardStep: async (
    projectId: string,
    data: ValidateStepRequest
  ): Promise<ValidateStepResponse> => {
    return client.post<ValidateStepResponse>(
      `/v1/projects/${projectId}/experiments/wizard/validate`,
      data
    )
  },

  estimateCost: async (
    projectId: string,
    data: EstimateCostRequest
  ): Promise<EstimateCostResponse> => {
    return client.post<EstimateCostResponse>(
      `/v1/projects/${projectId}/experiments/wizard/estimate`,
      data
    )
  },

  getDatasetFields: async (
    projectId: string,
    datasetId: string
  ): Promise<DatasetFieldsResponse> => {
    return client.get<DatasetFieldsResponse>(
      `/v1/projects/${projectId}/datasets/${datasetId}/fields`
    )
  },

  getExperimentConfig: async (
    projectId: string,
    experimentId: string
  ): Promise<ExperimentConfig> => {
    return client.get<ExperimentConfig>(
      `/v1/projects/${projectId}/experiments/${experimentId}/config`
    )
  },
}
