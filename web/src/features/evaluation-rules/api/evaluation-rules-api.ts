import { BrokleAPIClient } from '@/lib/api/core/client'
import type {
  EvaluationRule,
  CreateEvaluationRuleRequest,
  UpdateEvaluationRuleRequest,
  RuleListResponse,
  RuleListParams,
  RuleExecution,
  RuleExecutionDetail,
  ExecutionListResponse,
  ExecutionListParams,
  TriggerOptions,
  TriggerResponse,
  RuleAnalyticsParams,
  RuleAnalyticsResponse,
  TestRuleRequest,
  TestRuleResponse,
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
    if (params?.sort_by) queryParams.set('sort_by', params.sort_by)
    if (params?.sort_dir) queryParams.set('sort_dir', params.sort_dir)

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
  // Rule Testing API Methods
  // ============================================================================

  /**
   * Test an evaluation rule against sample spans without persisting scores.
   * Useful for validating rule configuration before activation.
   */
  testRule: async (
    projectId: string,
    ruleId: string,
    request?: TestRuleRequest
  ): Promise<TestRuleResponse> => {
    return client.post<TestRuleResponse>(
      `/v1/projects/${projectId}/evaluations/rules/${ruleId}/test`,
      request ?? {}
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
   * Get detailed execution info including span-level details for debugging
   */
  getExecutionDetail: async (
    projectId: string,
    ruleId: string,
    executionId: string
  ): Promise<RuleExecutionDetail> => {
    return client.get<RuleExecutionDetail>(
      `/v1/projects/${projectId}/evaluations/rules/${ruleId}/executions/${executionId}/detail`
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

  // ============================================================================
  // Rule Analytics API Methods
  // ============================================================================

  /**
   * Get analytics for a specific evaluation rule
   */
  getRuleAnalytics: async (
    projectId: string,
    ruleId: string,
    params?: RuleAnalyticsParams
  ): Promise<RuleAnalyticsResponse> => {
    const queryParams = new URLSearchParams()
    if (params?.period) queryParams.set('period', params.period)
    if (params?.from_timestamp) queryParams.set('from_timestamp', params.from_timestamp)
    if (params?.to_timestamp) queryParams.set('to_timestamp', params.to_timestamp)

    const query = queryParams.toString()
    const url = `/v1/projects/${projectId}/evaluations/rules/${ruleId}/analytics${query ? `?${query}` : ''}`
    return client.get<RuleAnalyticsResponse>(url)
  },
}
