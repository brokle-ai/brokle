import { BrokleAPIClient } from '@/lib/api/core/client'
import type { PaginatedResponse, QueryParams } from '@/lib/api/core/types'
import type {
  Score,
  ScoreConfig,
  ScoreListParams,
  CreateScoreConfigRequest,
  UpdateScoreConfigRequest,
  ScoreAnalyticsParams,
  ScoreAnalyticsData,
} from '../types'

const client = new BrokleAPIClient('/api')

/**
 * Builds type-safe query params from score list params.
 * Avoids unsafe type assertions while maintaining API compatibility.
 */
function buildScoreListQueryParams(params?: ScoreListParams): QueryParams | undefined {
  if (!params) return undefined
  const queryParams: QueryParams = {}
  if (params.trace_id) queryParams.trace_id = params.trace_id
  if (params.span_id) queryParams.span_id = params.span_id
  if (params.name) queryParams.name = params.name
  if (params.source) queryParams.source = params.source
  if (params.type) queryParams.type = params.type
  if (params.page) queryParams.page = params.page
  if (params.limit) queryParams.limit = params.limit
  if (params.sort_by) queryParams.sort_by = params.sort_by
  if (params.sort_dir) queryParams.sort_dir = params.sort_dir
  return queryParams
}

/**
 * Builds type-safe query params from analytics params.
 * Avoids unsafe type assertions while maintaining API compatibility.
 */
function buildAnalyticsQueryParams(params: ScoreAnalyticsParams): QueryParams {
  const queryParams: QueryParams = {
    score_name: params.score_name,
  }

  if (params.compare_score_name) {
    queryParams.compare_score_name = params.compare_score_name
  }
  if (params.from_timestamp) {
    queryParams.from_timestamp = params.from_timestamp
  }
  if (params.to_timestamp) {
    queryParams.to_timestamp = params.to_timestamp
  }
  if (params.interval) {
    queryParams.interval = params.interval
  }

  return queryParams
}

export const scoresApi = {
  // Scores (quality scores)
  listScores: async (
    projectId: string,
    params?: ScoreListParams
  ): Promise<PaginatedResponse<Score>> => {
    return client.getPaginated<Score>(
      `/v1/projects/${projectId}/scores`,
      buildScoreListQueryParams(params)
    )
  },

  getScore: async (projectId: string, scoreId: string): Promise<Score> => {
    return client.get<Score>(`/v1/projects/${projectId}/scores/${scoreId}`)
  },

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

  // Analytics
  getScoreAnalytics: async (
    projectId: string,
    params: ScoreAnalyticsParams
  ): Promise<ScoreAnalyticsData> => {
    return client.get<ScoreAnalyticsData>(
      `/v1/projects/${projectId}/scores/analytics`,
      buildAnalyticsQueryParams(params)
    )
  },

  getScoreNames: async (projectId: string): Promise<string[]> => {
    return client.get<string[]>(`/v1/projects/${projectId}/scores/names`)
  },
}
