import { queryOptions } from '@tanstack/react-query'
import { rawFetch } from '@/lib/api/client'
import type { PromptListResponse } from './types'

// TkDodo-style hierarchical query keys. `list(projectId, params)`
// invalidates cleanly via `lists()` when the detail/mutation ports
// land and need to bust the list cache after create/update/delete.
export const promptsKeys = {
  all: ['prompts'] as const,
  lists: () => [...promptsKeys.all, 'list'] as const,
  list: (projectId: string, params: PromptListParams) =>
    [...promptsKeys.lists(), projectId, params] as const,
  details: () => [...promptsKeys.all, 'detail'] as const,
  detail: (promptId: string) => [...promptsKeys.details(), promptId] as const,
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
        // Backend accepts `search` as the free-text query param; keep
        // the user-facing alias `q` in the URL for familiarity.
        search.set('search', params.q)
      }
      const resp = await rawFetch(
        `/api/v1/projects/${projectId}/prompts?${search.toString()}`,
        { method: 'GET' },
      )
      return (await resp.json()) as PromptListResponse
    },
    // Prompts are user-edited resources — keep a short staleness window
    // so that navigating back to the list after a mutation reflects the
    // change without a forced refetch.
    staleTime: 30 * 1000,
  })
