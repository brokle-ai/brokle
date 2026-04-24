import { queryOptions } from '@tanstack/react-query'
import { rawFetch } from '@/lib/api/client'
import type { ApiKeyListResponse } from './types'

// TkDodo-style hierarchical query keys. `list(projectId, params)`
// invalidates cleanly via `lists()` when the create / revoke port
// adds mutations that should bust the list cache.
export const apiKeyKeys = {
  all: ['api-keys'] as const,
  lists: () => [...apiKeyKeys.all, 'list'] as const,
  list: (projectId: string, params: ApiKeyListParams) =>
    [...apiKeyKeys.lists(), projectId, params] as const,
} as const

export interface ApiKeyListParams {
  page: number
  limit: number
}

export const apiKeyListQueryOptions = (
  projectId: string,
  params: ApiKeyListParams,
) =>
  queryOptions({
    queryKey: apiKeyKeys.list(projectId, params),
    queryFn: async () => {
      const search = new URLSearchParams()
      search.set('page', String(params.page))
      search.set('limit', String(params.limit))
      const resp = await rawFetch(
        `/api/v1/projects/${projectId}/api-keys?${search.toString()}`,
        { method: 'GET' },
      )
      return (await resp.json()) as ApiKeyListResponse
    },
    staleTime: 60 * 1000,
  })
