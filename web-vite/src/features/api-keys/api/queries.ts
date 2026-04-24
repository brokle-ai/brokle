import { queryOptions } from '@tanstack/react-query'
import { rawFetch } from '@/lib/api/client'
import type {
  ApiKey,
  ApiKeyListResponse,
  CreateApiKeyRequest,
} from './types'

// TkDodo-style hierarchical query keys. `lists()` busts the cache on
// any create / revoke mutation; `all` is reserved for future nesting.
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

// The full plaintext `key` is only present in this response; the
// caller is responsible for getting it on the user's clipboard before
// the dialog closes.
export async function createApiKey(
  projectId: string,
  data: CreateApiKeyRequest,
): Promise<ApiKey> {
  const resp = await rawFetch(`/api/v1/projects/${projectId}/api-keys`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  return (await resp.json()) as ApiKey
}

// Revoke / delete an API key — 204 No Content. SDK clients using the
// revoked key start failing authentication on their next request.
export async function revokeApiKey(
  projectId: string,
  keyId: string,
): Promise<void> {
  await rawFetch(`/api/v1/projects/${projectId}/api-keys/${keyId}`, {
    method: 'DELETE',
  })
}
