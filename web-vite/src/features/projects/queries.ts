import { queryOptions } from '@tanstack/react-query'
import { rawFetch } from '@/lib/api/client'
import type { ListResponse } from '@/features/organizations/queries'

export interface Project {
  id: string
  organization_id: string
  name: string
  slug: string
  description?: string
  status?: 'active' | 'archived'
  created_at: string
  updated_at: string
}

export interface UpdateProjectRequest {
  name?: string
  description?: string
}

export interface CreateProjectRequest {
  organization_id: string
  name: string
  description?: string
}

export async function createProject(
  data: CreateProjectRequest,
): Promise<Project> {
  const resp = await rawFetch('/api/v1/projects', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  return (await resp.json()) as Project
}

export async function updateProject(
  projectId: string,
  data: UpdateProjectRequest,
): Promise<Project> {
  const resp = await rawFetch(`/api/v1/projects/${projectId}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  return (await resp.json()) as Project
}

// DELETE returns 204 — soft-delete on the backend. The project is
// unrecoverable from the UI; the user loses access immediately and the
// route redirects to the org landing page on success.
export async function deleteProject(projectId: string): Promise<void> {
  await rawFetch(`/api/v1/projects/${projectId}`, { method: 'DELETE' })
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
      // Dashboard-plane list endpoint is flat `/api/v1/projects` with an
      // `organization_id` filter — NOT `/organizations/:id/projects`.
      const params = new URLSearchParams({ organization_id: orgId, limit: '100' })
      const resp = await rawFetch(`/api/v1/projects?${params.toString()}`, {
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
      const resp = await rawFetch(`/api/v1/projects/${projectId}`, { method: 'GET' })
      return (await resp.json()) as Project
    },
    staleTime: 5 * 60 * 1000,
    retry: false,
  })
