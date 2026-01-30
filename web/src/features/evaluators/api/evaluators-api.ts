import { BrokleAPIClient } from '@/lib/api/core/client'
import type {
  Evaluator,
  CreateEvaluatorRequest,
  UpdateEvaluatorRequest,
  EvaluatorListResponse,
  EvaluatorListParams,
  EvaluatorExecution,
  EvaluatorExecutionDetail,
  ExecutionListResponse,
  ExecutionListParams,
  TriggerOptions,
  TriggerResponse,
  EvaluatorAnalyticsParams,
  EvaluatorAnalyticsResponse,
  TestEvaluatorRequest,
  TestEvaluatorResponse,
} from '../types'

const client = new BrokleAPIClient('/api')

export const evaluatorsApi = {
  /**
   * List all evaluators for a project with optional filtering
   */
  listEvaluators: async (
    projectId: string,
    params?: EvaluatorListParams
  ): Promise<EvaluatorListResponse> => {
    const queryParams = new URLSearchParams()
    if (params?.page) queryParams.set('page', String(params.page))
    if (params?.limit) queryParams.set('limit', String(params.limit))
    if (params?.status) queryParams.set('status', params.status)
    if (params?.scorer_type) queryParams.set('scorer_type', params.scorer_type)
    if (params?.search) queryParams.set('search', params.search)
    if (params?.sort_by) queryParams.set('sort_by', params.sort_by)
    if (params?.sort_dir) queryParams.set('sort_dir', params.sort_dir)

    const query = queryParams.toString()
    const url = `/v1/projects/${projectId}/evaluators${query ? `?${query}` : ''}`
    return client.get<EvaluatorListResponse>(url)
  },

  /**
   * Get a specific evaluator by ID
   */
  getEvaluator: async (projectId: string, evaluatorId: string): Promise<Evaluator> => {
    return client.get<Evaluator>(
      `/v1/projects/${projectId}/evaluators/${evaluatorId}`
    )
  },

  /**
   * Create a new evaluator
   */
  createEvaluator: async (
    projectId: string,
    data: CreateEvaluatorRequest
  ): Promise<Evaluator> => {
    return client.post<Evaluator>(
      `/v1/projects/${projectId}/evaluators`,
      data
    )
  },

  /**
   * Update an existing evaluator
   */
  updateEvaluator: async (
    projectId: string,
    evaluatorId: string,
    data: UpdateEvaluatorRequest
  ): Promise<Evaluator> => {
    return client.put<Evaluator>(
      `/v1/projects/${projectId}/evaluators/${evaluatorId}`,
      data
    )
  },

  /**
   * Delete an evaluator
   */
  deleteEvaluator: async (projectId: string, evaluatorId: string): Promise<void> => {
    await client.delete(`/v1/projects/${projectId}/evaluators/${evaluatorId}`)
  },

  /**
   * Activate an evaluator (enable automatic scoring)
   */
  activateEvaluator: async (
    projectId: string,
    evaluatorId: string
  ): Promise<{ message: string }> => {
    return client.post<{ message: string }>(
      `/v1/projects/${projectId}/evaluators/${evaluatorId}/activate`,
      {}
    )
  },

  /**
   * Deactivate an evaluator (disable automatic scoring)
   */
  deactivateEvaluator: async (
    projectId: string,
    evaluatorId: string
  ): Promise<{ message: string }> => {
    return client.post<{ message: string }>(
      `/v1/projects/${projectId}/evaluators/${evaluatorId}/deactivate`,
      {}
    )
  },

  // ============================================================================
  // Evaluator Testing API Methods
  // ============================================================================

  /**
   * Test an evaluator against sample spans without persisting scores.
   * Useful for validating evaluator configuration before activation.
   */
  testEvaluator: async (
    projectId: string,
    evaluatorId: string,
    request?: TestEvaluatorRequest
  ): Promise<TestEvaluatorResponse> => {
    return client.post<TestEvaluatorResponse>(
      `/v1/projects/${projectId}/evaluators/${evaluatorId}/test`,
      request ?? {}
    )
  },

  // ============================================================================
  // Evaluator Execution API Methods
  // ============================================================================

  /**
   * List execution history for an evaluator with optional filtering
   */
  listExecutions: async (
    projectId: string,
    evaluatorId: string,
    params?: ExecutionListParams
  ): Promise<ExecutionListResponse> => {
    const queryParams = new URLSearchParams()
    if (params?.page) queryParams.set('page', String(params.page))
    if (params?.limit) queryParams.set('limit', String(params.limit))
    if (params?.status) queryParams.set('status', params.status)
    if (params?.trigger_type) queryParams.set('trigger_type', params.trigger_type)

    const query = queryParams.toString()
    const url = `/v1/projects/${projectId}/evaluators/${evaluatorId}/executions${query ? `?${query}` : ''}`
    return client.get<ExecutionListResponse>(url)
  },

  /**
   * Get a specific execution by ID
   */
  getExecution: async (
    projectId: string,
    evaluatorId: string,
    executionId: string
  ): Promise<EvaluatorExecution> => {
    return client.get<EvaluatorExecution>(
      `/v1/projects/${projectId}/evaluators/${evaluatorId}/executions/${executionId}`
    )
  },

  /**
   * Get detailed execution info including span-level details for debugging
   */
  getExecutionDetail: async (
    projectId: string,
    evaluatorId: string,
    executionId: string
  ): Promise<EvaluatorExecutionDetail> => {
    return client.get<EvaluatorExecutionDetail>(
      `/v1/projects/${projectId}/evaluators/${evaluatorId}/executions/${executionId}/detail`
    )
  },

  /**
   * Get the latest execution for an evaluator
   */
  getLatestExecution: async (
    projectId: string,
    evaluatorId: string
  ): Promise<EvaluatorExecution> => {
    return client.get<EvaluatorExecution>(
      `/v1/projects/${projectId}/evaluators/${evaluatorId}/executions/latest`
    )
  },

  // ============================================================================
  // Manual Trigger API Methods
  // ============================================================================

  /**
   * Manually trigger an evaluator against matching spans
   * Returns 202 Accepted with execution ID for async processing
   */
  triggerEvaluator: async (
    projectId: string,
    evaluatorId: string,
    options?: TriggerOptions
  ): Promise<TriggerResponse> => {
    return client.post<TriggerResponse>(
      `/v1/projects/${projectId}/evaluators/${evaluatorId}/trigger`,
      options ?? {}
    )
  },

  // ============================================================================
  // Evaluator Analytics API Methods
  // ============================================================================

  /**
   * Get analytics for a specific evaluator
   */
  getEvaluatorAnalytics: async (
    projectId: string,
    evaluatorId: string,
    params?: EvaluatorAnalyticsParams
  ): Promise<EvaluatorAnalyticsResponse> => {
    const queryParams = new URLSearchParams()
    if (params?.period) queryParams.set('period', params.period)
    if (params?.from_timestamp) queryParams.set('from_timestamp', params.from_timestamp)
    if (params?.to_timestamp) queryParams.set('to_timestamp', params.to_timestamp)

    const query = queryParams.toString()
    const url = `/v1/projects/${projectId}/evaluators/${evaluatorId}/analytics${query ? `?${query}` : ''}`
    return client.get<EvaluatorAnalyticsResponse>(url)
  },
}
