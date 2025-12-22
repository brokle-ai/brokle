import { BrokleAPIClient } from '@/lib/api/core/client'
import type {
  ScoreConfig,
  CreateScoreConfigRequest,
  UpdateScoreConfigRequest,
} from '../types'

const client = new BrokleAPIClient('/api')

export const scoresApi = {
  // Score Configs
  listScoreConfigs: async (projectId: string): Promise<ScoreConfig[]> => {
    return client.get<ScoreConfig[]>(`/v1/projects/${projectId}/score-configs`)
  },

  getScoreConfig: async (
    projectId: string,
    configId: string
  ): Promise<ScoreConfig> => {
    return client.get<ScoreConfig>(
      `/v1/projects/${projectId}/score-configs/${configId}`
    )
  },

  createScoreConfig: async (
    projectId: string,
    data: CreateScoreConfigRequest
  ): Promise<ScoreConfig> => {
    return client.post<ScoreConfig>(
      `/v1/projects/${projectId}/score-configs`,
      data
    )
  },

  updateScoreConfig: async (
    projectId: string,
    configId: string,
    data: UpdateScoreConfigRequest
  ): Promise<ScoreConfig> => {
    return client.put<ScoreConfig>(
      `/v1/projects/${projectId}/score-configs/${configId}`,
      data
    )
  },

  deleteScoreConfig: async (
    projectId: string,
    configId: string
  ): Promise<void> => {
    await client.delete(`/v1/projects/${projectId}/score-configs/${configId}`)
  },
}
