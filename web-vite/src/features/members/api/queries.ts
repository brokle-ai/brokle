import { queryOptions } from '@tanstack/react-query'
import { rawFetch } from '@/lib/api/client'
import type {
  InviteMemberRequest,
  Invitation,
  InvitationListResponse,
  MemberListResponse,
  RoleListResponse,
} from './types'

// TkDodo-style hierarchical query keys. Parameters exist in the key
// for forward-compat with future search / role filter; today the
// backend ignores `page` / `limit` / `q` on this endpoint.
export const memberKeys = {
  all: ['members'] as const,
  lists: () => [...memberKeys.all, 'list'] as const,
  list: (orgId: string, params: MemberListParams) =>
    [...memberKeys.lists(), orgId, params] as const,
  invitations: () => [...memberKeys.all, 'invitations'] as const,
  invitationList: (orgId: string) =>
    [...memberKeys.invitations(), orgId] as const,
  roles: () => [...memberKeys.all, 'roles'] as const,
  roleList: (scopeType: string) => [...memberKeys.roles(), scopeType] as const,
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

export const pendingInvitationsQueryOptions = (orgId: string) =>
  queryOptions({
    queryKey: memberKeys.invitationList(orgId),
    queryFn: async () => {
      const resp = await rawFetch(
        `/api/v1/organizations/${orgId}/invitations`,
        { method: 'GET' },
      )
      return (await resp.json()) as InvitationListResponse
    },
    staleTime: 30 * 1000,
  })

// Roles for the invite dialog's role select — `scope_type=organization`
// returns the org-scoped templates (owner / admin / developer / viewer).
// Longer staleTime because the system-template list doesn't mutate in
// a normal session.
export const organizationRolesQueryOptions = () =>
  queryOptions({
    queryKey: memberKeys.roleList('organization'),
    queryFn: async () => {
      const resp = await rawFetch(
        `/api/v1/rbac/roles?scope_type=organization`,
        { method: 'GET' },
      )
      return (await resp.json()) as RoleListResponse
    },
    staleTime: 5 * 60 * 1000,
  })

export async function inviteMember(
  orgId: string,
  data: InviteMemberRequest,
): Promise<Invitation> {
  const resp = await rawFetch(`/api/v1/organizations/${orgId}/invitations`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  return (await resp.json()) as Invitation
}

// Revoke a pending invitation — 204 No Content. No body parse.
export async function revokeInvitation(
  orgId: string,
  invitationId: string,
): Promise<void> {
  await rawFetch(
    `/api/v1/organizations/${orgId}/invitations/${invitationId}`,
    { method: 'DELETE' },
  )
}

// Remove an accepted member from the organization — 204 No Content.
// The backend route is `/members/{userId}` (the task brief said
// `{memberId}` but the Huma operation uses `userId`).
export async function removeMember(
  orgId: string,
  userId: string,
): Promise<void> {
  await rawFetch(`/api/v1/organizations/${orgId}/members/${userId}`, {
    method: 'DELETE',
  })
}
