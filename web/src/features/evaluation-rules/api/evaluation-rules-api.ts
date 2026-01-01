import { BrokleAPIClient } from '@/lib/api/core/client'
import type {
  EvaluationRule,
  CreateEvaluationRuleRequest,
  UpdateEvaluationRuleRequest,
  RuleListResponse,
  RuleListParams,
  RuleExecution,
  ExecutionListResponse,
  ExecutionListParams,
  TriggerOptions,
  TriggerResponse,
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

  // ============================================================================
  // Rule Execution API Methods
  // ============================================================================

  /**
   * List execution history for a rule with optional filtering
   */
  listExecutions: async (
    projectId: string,
    ruleId: string,
    params?: ExecutionListParams
  ): Promise<ExecutionListResponse> => {
    const queryParams = new URLSearchParams()
    if (params?.page) queryParams.set('page', String(params.page))
    if (params?.limit) queryParams.set('limit', String(params.limit))
    if (params?.status) queryParams.set('status', params.status)
    if (params?.trigger_type) queryParams.set('trigger_type', params.trigger_type)

    const query = queryParams.toString()
    const url = `/v1/projects/${projectId}/evaluations/rules/${ruleId}/executions${query ? `?${query}` : ''}`
    return client.get<ExecutionListResponse>(url)
  },

  /**
   * Get a specific execution by ID
   */
  getExecution: async (
    projectId: string,
    ruleId: string,
    executionId: string
  ): Promise<RuleExecution> => {
    return client.get<RuleExecution>(
      `/v1/projects/${projectId}/evaluations/rules/${ruleId}/executions/${executionId}`
    )
  },

  /**
   * Get the latest execution for a rule
   */
  getLatestExecution: async (
    projectId: string,
    ruleId: string
  ): Promise<RuleExecution> => {
    return client.get<RuleExecution>(
      `/v1/projects/${projectId}/evaluations/rules/${ruleId}/executions/latest`
    )
  },

  // ============================================================================
  // Manual Trigger API Methods
  // ============================================================================

  /**
   * Manually trigger an evaluation rule against matching spans
   * Returns 202 Accepted with execution ID for async processing
   */
  triggerRule: async (
    projectId: string,
    ruleId: string,
    options?: TriggerOptions
  ): Promise<TriggerResponse> => {
    return client.post<TriggerResponse>(
      `/v1/projects/${projectId}/evaluations/rules/${ruleId}/trigger`,
      options ?? {}
    )
  },
}
