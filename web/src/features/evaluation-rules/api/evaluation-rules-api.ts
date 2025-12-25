import { BrokleAPIClient } from '@/lib/api/core/client'
import type {
  EvaluationRule,
  CreateEvaluationRuleRequest,
  UpdateEvaluationRuleRequest,
  RuleListResponse,
  RuleListParams,
} from '../types'

const client = new BrokleAPIClient('/api')

export const evaluationRulesApi = {
  /**
   * List all evaluation rules for a project with optional filtering
   */
  listRules: async (
    projectId: string,
    params?: RuleListParams
  ): Promise<RuleListResponse> => {
    const queryParams = new URLSearchParams()
    if (params?.page) queryParams.set('page', String(params.page))
    if (params?.limit) queryParams.set('limit', String(params.limit))
    if (params?.status) queryParams.set('status', params.status)
    if (params?.scorer_type) queryParams.set('scorer_type', params.scorer_type)
    if (params?.search) queryParams.set('search', params.search)

    const query = queryParams.toString()
    const url = `/v1/projects/${projectId}/evaluations/rules${query ? `?${query}` : ''}`
    return client.get<RuleListResponse>(url)
  },

  /**
   * Get a specific evaluation rule by ID
   */
  getRule: async (projectId: string, ruleId: string): Promise<EvaluationRule> => {
    return client.get<EvaluationRule>(
      `/v1/projects/${projectId}/evaluations/rules/${ruleId}`
    )
  },

  /**
   * Create a new evaluation rule
   */
  createRule: async (
    projectId: string,
    data: CreateEvaluationRuleRequest
  ): Promise<EvaluationRule> => {
    return client.post<EvaluationRule>(
      `/v1/projects/${projectId}/evaluations/rules`,
      data
    )
  },

  /**
   * Update an existing evaluation rule
   */
  updateRule: async (
    projectId: string,
    ruleId: string,
    data: UpdateEvaluationRuleRequest
  ): Promise<EvaluationRule> => {
    return client.put<EvaluationRule>(
      `/v1/projects/${projectId}/evaluations/rules/${ruleId}`,
      data
    )
  },

  /**
   * Delete an evaluation rule
   */
  deleteRule: async (projectId: string, ruleId: string): Promise<void> => {
    await client.delete(`/v1/projects/${projectId}/evaluations/rules/${ruleId}`)
  },

  /**
   * Activate an evaluation rule (enable automatic scoring)
   */
  activateRule: async (
    projectId: string,
    ruleId: string
  ): Promise<{ message: string }> => {
    return client.post<{ message: string }>(
      `/v1/projects/${projectId}/evaluations/rules/${ruleId}/activate`,
      {}
    )
  },

  /**
   * Deactivate an evaluation rule (disable automatic scoring)
   */
  deactivateRule: async (
    projectId: string,
    ruleId: string
  ): Promise<{ message: string }> => {
    return client.post<{ message: string }>(
      `/v1/projects/${projectId}/evaluations/rules/${ruleId}/deactivate`,
      {}
    )
  },
}
