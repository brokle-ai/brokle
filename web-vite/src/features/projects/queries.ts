import { queryOptions } from '@tanstack/react-query'
import { rawFetch } from '@/lib/api/client'
import type { ListResponse } from '@/features/organizations/queries'

export interface Project {
  id: string
  organization_id: string
  name: string
  slug: string
  created_at: string
  updated_at: string
}

export const projectKeys = {
  all: ['projects'] as const,
  lists: () => [...projectKeys.all, 'list'] as const,
  listForOrg: (orgId: string) => [...projectKeys.lists(), orgId] as const,
  details: () => [...projectKeys.all, 'detail'] as const,
  detail: (id: string) => [...projectKeys.details(), id] as const,
} as const

export const projectListQueryOptions = (orgId: string) =>
  queryOptions({
    queryKey: projectKeys.listForOrg(orgId),
    queryFn: async () => {
      const resp = await rawFetch(`/v1/organizations/${orgId}/projects`, {
        method: 'GET',
      })
      return (await resp.json()) as ListResponse<Project>
    },
    staleTime: 5 * 60 * 1000,
  })

export const projectMembershipQueryOptions = (projectId: string) =>
  queryOptions({
    queryKey: projectKeys.detail(projectId),
    queryFn: async () => {
      const resp = await rawFetch(`/v1/projects/${projectId}`, { method: 'GET' })
      return (await resp.json()) as Project
    },
    staleTime: 5 * 60 * 1000,
    retry: false,
  })
