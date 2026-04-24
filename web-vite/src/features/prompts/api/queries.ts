import { queryOptions } from '@tanstack/react-query'
import { rawFetch } from '@/lib/api/client'
import type {
  CreatePromptRequest,
  CreateVersionRequest,
  Prompt,
  PromptListResponse,
  PromptVersion,
} from './types'

// TkDodo-style hierarchical query keys. `list(projectId, params)`
// invalidates cleanly via `lists()`; `detail(promptId)` busts when a
// new version is posted.
export const promptsKeys = {
  all: ['prompts'] as const,
  lists: () => [...promptsKeys.all, 'list'] as const,
  list: (projectId: string, params: PromptListParams) =>
    [...promptsKeys.lists(), projectId, params] as const,
  details: () => [...promptsKeys.all, 'detail'] as const,
  detail: (projectId: string, promptId: string) =>
    [...promptsKeys.details(), projectId, promptId] as const,
  versions: (projectId: string, promptId: string) =>
    [...promptsKeys.all, 'versions', projectId, promptId] as const,
} as const

export interface PromptListParams {
  page: number
  limit: number
  q?: string
}

export const promptListQueryOptions = (
  projectId: string,
  params: PromptListParams,
) =>
  queryOptions({
    queryKey: promptsKeys.list(projectId, params),
    queryFn: async () => {
      const search = new URLSearchParams()
      search.set('page', String(params.page))
      search.set('limit', String(params.limit))
      if (params.q && params.q.length > 0) {
        search.set('search', params.q)
      }
      const resp = await rawFetch(
        `/api/v1/projects/${projectId}/prompts?${search.toString()}`,
        { method: 'GET' },
      )
      return (await resp.json()) as PromptListResponse
    },
    staleTime: 30 * 1000,
  })

export const promptDetailQueryOptions = (
  projectId: string,
  promptId: string,
) =>
  queryOptions({
    queryKey: promptsKeys.detail(projectId, promptId),
    queryFn: async () => {
      const resp = await rawFetch(
        `/api/v1/projects/${projectId}/prompts/${promptId}`,
        { method: 'GET' },
      )
      return (await resp.json()) as Prompt
    },
    staleTime: 30 * 1000,
  })

export const promptVersionsQueryOptions = (
  projectId: string,
  promptId: string,
) =>
  queryOptions({
    queryKey: promptsKeys.versions(projectId, promptId),
    queryFn: async () => {
      const resp = await rawFetch(
        `/api/v1/projects/${projectId}/prompts/${promptId}/versions`,
        { method: 'GET' },
      )
      return (await resp.json()) as PromptVersion[]
    },
    staleTime: 30 * 1000,
  })

// Mutation functions — intentionally left unwrapped so callers can
// compose them with useMutation + their own onSuccess/navigate logic.
export async function createPrompt(
  projectId: string,
  data: CreatePromptRequest,
): Promise<Prompt> {
  const resp = await rawFetch(`/api/v1/projects/${projectId}/prompts`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  return (await resp.json()) as Prompt
}

export async function createPromptVersion(
  projectId: string,
  promptId: string,
  data: CreateVersionRequest,
): Promise<PromptVersion> {
  const resp = await rawFetch(
    `/api/v1/projects/${projectId}/prompts/${promptId}/versions`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    },
  )
  return (await resp.json()) as PromptVersion
}
