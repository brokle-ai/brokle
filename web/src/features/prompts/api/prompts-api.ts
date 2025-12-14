import { BrokleAPIClient } from '@/lib/api/core/client'
import type {
  Prompt,
  PromptListItem,
  PromptVersion,
  VersionDiff,
  ExecutePromptResponse,
  UpsertResponse,
  CreatePromptRequest,
  UpdatePromptRequest,
  CreateVersionRequest,
  SetLabelsRequest,
  ExecutePromptRequest,
  GetPromptsParams,
  GetPromptParams,
} from '../types'

const client = new BrokleAPIClient('/api')

// ============================================================================
// Prompts API
// ============================================================================

/**
 * Get all prompts for a project with filtering and pagination
 *
 * Backend endpoint: GET /api/v1/projects/:projectId/prompts
 */
export const getPrompts = async (params: GetPromptsParams): Promise<{
  prompts: PromptListItem[]
  totalCount: number
  page: number
  pageSize: number
  totalPages: number
}> => {
  const { projectId, type, tags, search, page = 1, limit = 50 } = params

  const queryParams: Record<string, any> = {
    page,
    limit,
  }

  if (type) queryParams.type = type
  if (tags && tags.length > 0) queryParams.tags = tags.join(',')
  if (search) queryParams.search = search

  const response = await client.getPaginated<PromptListItem>(
    `/v1/projects/${projectId}/prompts`,
    queryParams
  )

  return {
    prompts: response.data,
    totalCount: response.pagination.total,
    page: response.pagination.page,
    pageSize: response.pagination.limit,
    totalPages: response.pagination.totalPages,
  }
}

/**
 * Get a prompt by ID
 *
 * Backend endpoint: GET /api/v1/projects/:projectId/prompts/:promptId
 */
export const getPromptById = async (
  projectId: string,
  promptId: string
): Promise<Prompt> => {
  return client.get<Prompt>(`/v1/projects/${projectId}/prompts/${promptId}`)
}

/**
 * Create a new prompt
 *
 * Backend endpoint: POST /api/v1/projects/:projectId/prompts
 */
export const createPrompt = async (
  projectId: string,
  data: CreatePromptRequest
): Promise<Prompt> => {
  return client.post<Prompt>(`/v1/projects/${projectId}/prompts`, data)
}

/**
 * Update a prompt's metadata
 *
 * Backend endpoint: PUT /api/v1/projects/:projectId/prompts/:promptId
 */
export const updatePrompt = async (
  projectId: string,
  promptId: string,
  data: UpdatePromptRequest
): Promise<void> => {
  await client.put(`/v1/projects/${projectId}/prompts/${promptId}`, data)
}

/**
 * Delete a prompt
 *
 * Backend endpoint: DELETE /api/v1/projects/:projectId/prompts/:promptId
 */
export const deletePrompt = async (
  projectId: string,
  promptId: string
): Promise<void> => {
  await client.delete(`/v1/projects/${projectId}/prompts/${promptId}`)
}

// ============================================================================
// Version API
// ============================================================================

/**
 * Get all versions of a prompt
 *
 * Backend endpoint: GET /api/v1/projects/:projectId/prompts/:promptId/versions
 */
export const getVersions = async (
  projectId: string,
  promptId: string
): Promise<PromptVersion[]> => {
  return client.get<PromptVersion[]>(
    `/v1/projects/${projectId}/prompts/${promptId}/versions`
  )
}

/**
 * Create a new version
 *
 * Backend endpoint: POST /api/v1/projects/:projectId/prompts/:promptId/versions
 */
export const createVersion = async (
  projectId: string,
  promptId: string,
  data: CreateVersionRequest
): Promise<PromptVersion> => {
  return client.post<PromptVersion>(
    `/v1/projects/${projectId}/prompts/${promptId}/versions`,
    data
  )
}

/**
 * Get a specific version
 *
 * Backend endpoint: GET /api/v1/projects/:projectId/prompts/:promptId/versions/:versionId
 */
export const getVersion = async (
  projectId: string,
  promptId: string,
  versionId: string
): Promise<PromptVersion> => {
  return client.get<PromptVersion>(
    `/v1/projects/${projectId}/prompts/${promptId}/versions/${versionId}`
  )
}

/**
 * Set labels on a version
 *
 * Backend endpoint: PATCH /api/v1/projects/:projectId/prompts/:promptId/versions/:versionId/labels
 */
export const setLabels = async (
  projectId: string,
  promptId: string,
  versionId: string,
  labels: string[]
): Promise<void> => {
  await client.patch(
    `/v1/projects/${projectId}/prompts/${promptId}/versions/${versionId}/labels`,
    { labels }
  )
}

/**
 * Compare two versions
 *
 * Backend endpoint: GET /api/v1/projects/:projectId/prompts/:promptId/diff
 */
export const getVersionDiff = async (
  projectId: string,
  promptId: string,
  fromVersion: number,
  toVersion: number
): Promise<VersionDiff> => {
  return client.get<VersionDiff>(
    `/v1/projects/${projectId}/prompts/${promptId}/diff`,
    { from: fromVersion, to: toVersion }
  )
}

// ============================================================================
// Execution API
// ============================================================================

/**
 * Execute a prompt version with LLM
 *
 * Backend endpoint: POST /api/v1/projects/:projectId/prompts/:promptId/versions/:versionId/execute
 */
export const executePrompt = async (
  projectId: string,
  promptId: string,
  versionId: string,
  data: ExecutePromptRequest
): Promise<ExecutePromptResponse> => {
  return client.post<ExecutePromptResponse>(
    `/v1/projects/${projectId}/prompts/${promptId}/versions/${versionId}/execute`,
    data
  )
}

// ============================================================================
// Protected Labels API
// ============================================================================

/**
 * Get protected labels for a project
 *
 * Backend endpoint: GET /api/v1/projects/:projectId/prompts/settings/protected-labels
 */
export const getProtectedLabels = async (
  projectId: string
): Promise<string[]> => {
  const response = await client.get<{ protected_labels: string[] }>(
    `/v1/projects/${projectId}/prompts/settings/protected-labels`
  )
  return response.protected_labels
}

/**
 * Set protected labels for a project
 *
 * Backend endpoint: PUT /api/v1/projects/:projectId/prompts/settings/protected-labels
 */
export const setProtectedLabels = async (
  projectId: string,
  labels: string[]
): Promise<void> => {
  await client.put(`/v1/projects/${projectId}/prompts/settings/protected-labels`, {
    protected_labels: labels,
  })
}
