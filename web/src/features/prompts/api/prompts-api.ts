import { BrokleAPIClient } from '@/lib/api/core/client'
import type { PaginatedResponse } from '@/lib/api/core/types'
import type {
  Prompt,
  PromptListItem,
  PromptVersion,
  VersionDiff,
  UpsertResponse,
  CreatePromptRequest,
  UpdatePromptRequest,
  CreateVersionRequest,
  GetPromptsParams,
  PromptType,
  TemplateDialect,
  TextTemplate,
  ChatTemplate,
} from '../types'

// Template validation types
export interface SyntaxError {
  line: number
  column: number
  message: string
  code: string
}

export interface SyntaxWarning {
  line: number
  column: number
  message: string
  code: string
}

export interface ValidateTemplateRequest {
  template: TextTemplate | ChatTemplate
  type: PromptType
  dialect?: TemplateDialect
}

export interface ValidateTemplateResponse {
  valid: boolean
  dialect: TemplateDialect
  variables: string[]
  errors: SyntaxError[]
  warnings: SyntaxWarning[]
}

export interface PreviewTemplateRequest {
  template: TextTemplate | ChatTemplate
  type: PromptType
  variables: Record<string, unknown>
  dialect?: TemplateDialect
}

export interface PreviewTemplateResponse {
  compiled: TextTemplate | ChatTemplate
  dialect: TemplateDialect
}

export interface DetectDialectRequest {
  template: TextTemplate | ChatTemplate
  type: PromptType
}

export interface DetectDialectResponse {
  dialect: TemplateDialect
}

const client = new BrokleAPIClient('/api')

export const getPrompts = async (params: GetPromptsParams): Promise<PaginatedResponse<PromptListItem>> => {
  const { projectId, type, tags, search, page = 1, limit = 50 } = params

  const queryParams: Record<string, any> = {
    page,
    limit,
  }

  if (type) queryParams.type = type
  if (tags && tags.length > 0) queryParams.tags = tags.join(',')
  if (search) queryParams.search = search

  return client.getPaginated<PromptListItem>(
    `/v1/projects/${projectId}/prompts`,
    queryParams
  )
}

export const getPromptById = async (
  projectId: string,
  promptId: string
): Promise<Prompt> => {
  return client.get<Prompt>(`/v1/projects/${projectId}/prompts/${promptId}`)
}

export const createPrompt = async (
  projectId: string,
  data: CreatePromptRequest
): Promise<Prompt> => {
  return client.post<Prompt>(`/v1/projects/${projectId}/prompts`, data)
}

export const updatePrompt = async (
  projectId: string,
  promptId: string,
  data: UpdatePromptRequest
): Promise<void> => {
  await client.put(`/v1/projects/${projectId}/prompts/${promptId}`, data)
}

export const deletePrompt = async (
  projectId: string,
  promptId: string
): Promise<void> => {
  await client.delete(`/v1/projects/${projectId}/prompts/${promptId}`)
}

export const getVersions = async (
  projectId: string,
  promptId: string
): Promise<PromptVersion[]> => {
  return client.get<PromptVersion[]>(
    `/v1/projects/${projectId}/prompts/${promptId}/versions`
  )
}

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

export const getVersion = async (
  projectId: string,
  promptId: string,
  versionId: string
): Promise<PromptVersion> => {
  return client.get<PromptVersion>(
    `/v1/projects/${projectId}/prompts/${promptId}/versions/${versionId}`
  )
}

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

export const getProtectedLabels = async (
  projectId: string
): Promise<string[]> => {
  const response = await client.get<{ protected_labels: string[] }>(
    `/v1/projects/${projectId}/prompts/settings/protected-labels`
  )
  return response.protected_labels
}

export const setProtectedLabels = async (
  projectId: string,
  labels: string[]
): Promise<void> => {
  await client.put(`/v1/projects/${projectId}/prompts/settings/protected-labels`, {
    protected_labels: labels,
  })
}

export const validateTemplate = async (
  projectId: string,
  data: ValidateTemplateRequest
): Promise<ValidateTemplateResponse> => {
  return client.post<ValidateTemplateResponse>(
    `/v1/projects/${projectId}/prompts/validate-template`,
    data
  )
}

export const previewTemplate = async (
  projectId: string,
  data: PreviewTemplateRequest
): Promise<PreviewTemplateResponse> => {
  return client.post<PreviewTemplateResponse>(
    `/v1/projects/${projectId}/prompts/preview-template`,
    data
  )
}

export const detectDialect = async (
  projectId: string,
  data: DetectDialectRequest
): Promise<DetectDialectResponse> => {
  return client.post<DetectDialectResponse>(
    `/v1/projects/${projectId}/prompts/detect-dialect`,
    data
  )
}
