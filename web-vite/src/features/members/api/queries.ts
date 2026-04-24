import { queryOptions } from '@tanstack/react-query'
import { rawFetch } from '@/lib/api/client'
import type { MemberListResponse } from './types'

// TkDodo-style hierarchical query keys. Parameters exist in the key
// for forward-compat with future search / role filter; today the
// backend ignores `page` / `limit` / `q` on this endpoint.
export const memberKeys = {
  all: ['members'] as const,
  lists: () => [...memberKeys.all, 'list'] as const,
  list: (orgId: string, params: MemberListParams) =>
    [...memberKeys.lists(), orgId, params] as const,
} as const

export interface MemberListParams {
  page: number
  limit: number
  q?: string
}

export const memberListQueryOptions = (
  orgId: string,
  params: MemberListParams,
) =>
  queryOptions({
    queryKey: memberKeys.list(orgId, params),
    queryFn: async () => {
      // The list-members operation at
      // /api/v1/organizations/{orgId}/members returns the full
      // members array — no pagination — so we don't currently wire
      // page/limit/q into the query string.
      const resp = await rawFetch(`/api/v1/organizations/${orgId}/members`, {
        method: 'GET',
      })
      return (await resp.json()) as MemberListResponse
    },
    staleTime: 60 * 1000,
  })
