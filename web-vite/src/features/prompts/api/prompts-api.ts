import { rawFetch } from '@/lib/api/client'
import type {
  Prompt,
  PromptVersion,
  VersionDiff,
  CreatePromptRequest,
  UpdatePromptRequest,
  CreateVersionRequest,
  PromptType,
  TemplateDialect,
  PromptTemplate,
  PromptListResponse,
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
  template: PromptTemplate
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
  template: PromptTemplate
  type: PromptType
  variables: Record<string, unknown>
  dialect?: TemplateDialect
}

export interface PreviewTemplateResponse {
  compiled: PromptTemplate
  dialect: TemplateDialect
}

export interface DetectDialectRequest {
  template: PromptTemplate
  type: PromptType
}

export interface DetectDialectResponse {
  dialect: TemplateDialect
}

export interface GetPromptsParams {
  projectId: string
  type?: PromptType
  tags?: string[]
  search?: string
  page?: number
  limit?: number
}

async function get<T>(path: string, query?: Record<string, unknown>): Promise<T> {
  const qs = query
    ? '?' +
      Object.entries(query)
        .filter(([, v]) => v !== undefined && v !== null && v !== '')
        .map(
          ([k, v]) =>
            `${encodeURIComponent(k)}=${encodeURIComponent(
              Array.isArray(v) ? v.join(',') : String(v),
            )}`,
        )
        .join('&')
    : ''
  const resp = await rawFetch(`${path}${qs}`, { method: 'GET' })
  return (await resp.json()) as T
}

async function send<T>(
  method: 'POST' | 'PUT' | 'PATCH' | 'DELETE',
  path: string,
  body?: unknown,
): Promise<T> {
  const init: RequestInit = { method }
  if (body !== undefined) {
    init.headers = { 'Content-Type': 'application/json' }
    init.body = JSON.stringify(body)
  }
  const resp = await rawFetch(path, init)
  // Some endpoints return no content.
  if (resp.status === 204) return undefined as T
  const text = await resp.text()
  if (!text) return undefined as T
  return JSON.parse(text) as T
}

export async function getPrompts(params: GetPromptsParams): Promise<PromptListResponse> {
  const { projectId, type, tags, search, page = 1, limit = 50 } = params
  const query: Record<string, unknown> = { page, limit }
  if (type) query.type = type
  if (tags && tags.length > 0) query.tags = tags.join(',')
  if (search) query.search = search
  return get<PromptListResponse>(`/api/v1/projects/${projectId}/prompts`, query)
}

export function getPromptById(projectId: string, promptId: string): Promise<Prompt> {
  return get<Prompt>(`/api/v1/projects/${projectId}/prompts/${promptId}`)
}

export function createPrompt(
  projectId: string,
  data: CreatePromptRequest,
): Promise<Prompt> {
  return send<Prompt>('POST', `/api/v1/projects/${projectId}/prompts`, data)
}

export function updatePrompt(
  projectId: string,
  promptId: string,
  data: UpdatePromptRequest,
): Promise<void> {
  return send<void>('PUT', `/api/v1/projects/${projectId}/prompts/${promptId}`, data)
}

export function deletePrompt(projectId: string, promptId: string): Promise<void> {
  return send<void>('DELETE', `/api/v1/projects/${projectId}/prompts/${promptId}`)
}

export function getVersions(projectId: string, promptId: string): Promise<PromptVersion[]> {
  return get<PromptVersion[]>(
    `/api/v1/projects/${projectId}/prompts/${promptId}/versions`,
  )
}

export function createVersion(
  projectId: string,
  promptId: string,
  data: CreateVersionRequest,
): Promise<PromptVersion> {
  return send<PromptVersion>(
    'POST',
    `/api/v1/projects/${projectId}/prompts/${promptId}/versions`,
    data,
  )
}

export function getVersion(
  projectId: string,
  promptId: string,
  versionId: string,
): Promise<PromptVersion> {
  return get<PromptVersion>(
    `/api/v1/projects/${projectId}/prompts/${promptId}/versions/${versionId}`,
  )
}

export function setLabels(
  projectId: string,
  promptId: string,
  versionId: string,
  labels: string[],
): Promise<void> {
  return send<void>(
    'PATCH',
    `/api/v1/projects/${projectId}/prompts/${promptId}/versions/${versionId}/labels`,
    { labels },
  )
}

export function getVersionDiff(
  projectId: string,
  promptId: string,
  fromVersion: number,
  toVersion: number,
): Promise<VersionDiff> {
  return get<VersionDiff>(
    `/api/v1/projects/${projectId}/prompts/${promptId}/diff`,
    { from: fromVersion, to: toVersion },
  )
}

export async function getProtectedLabels(projectId: string): Promise<string[]> {
  const resp = await get<{ protected_labels: string[] }>(
    `/api/v1/projects/${projectId}/prompts/settings/protected-labels`,
  )
  return resp.protected_labels
}

export function setProtectedLabels(projectId: string, labels: string[]): Promise<void> {
  return send<void>(
    'PUT',
    `/api/v1/projects/${projectId}/prompts/settings/protected-labels`,
    { protected_labels: labels },
  )
}

export function validateTemplate(
  projectId: string,
  data: ValidateTemplateRequest,
): Promise<ValidateTemplateResponse> {
  return send<ValidateTemplateResponse>(
    'POST',
    `/api/v1/projects/${projectId}/prompts/validate-template`,
    data,
  )
}

export function previewTemplate(
  projectId: string,
  data: PreviewTemplateRequest,
): Promise<PreviewTemplateResponse> {
  return send<PreviewTemplateResponse>(
    'POST',
    `/api/v1/projects/${projectId}/prompts/preview-template`,
    data,
  )
}

export function detectDialect(
  projectId: string,
  data: DetectDialectRequest,
): Promise<DetectDialectResponse> {
  return send<DetectDialectResponse>(
    'POST',
    `/api/v1/projects/${projectId}/prompts/detect-dialect`,
    data,
  )
}
