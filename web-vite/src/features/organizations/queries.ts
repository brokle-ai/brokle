import { queryOptions } from '@tanstack/react-query'
import { rawFetch } from '@/lib/api/client'

export interface Organization {
  id: string
  name: string
  slug: string
  billing_email: string
  subscription_plan: 'free' | 'pro' | 'business' | 'enterprise'
  created_at: string
  updated_at: string
}

export interface ListResponse<T> {
  data: T[]
  pagination: {
    page: number
    limit: number
    total: number
    total_pages: number
    has_next: boolean
    has_prev: boolean
  }
}

export const organizationKeys = {
  all: ['organizations'] as const,
  lists: () => [...organizationKeys.all, 'list'] as const,
  list: () => [...organizationKeys.lists()] as const,
  details: () => [...organizationKeys.all, 'detail'] as const,
  detail: (id: string) => [...organizationKeys.details(), id] as const,
} as const

export const organizationListQueryOptions = () =>
  queryOptions({
    queryKey: organizationKeys.list(),
    queryFn: async () => {
      const resp = await rawFetch('/v1/organizations', { method: 'GET' })
      return (await resp.json()) as ListResponse<Organization>
    },
    staleTime: 5 * 60 * 1000,
  })

export const organizationMembershipQueryOptions = (orgId: string) =>
  queryOptions({
    queryKey: organizationKeys.detail(orgId),
    queryFn: async () => {
      const resp = await rawFetch(`/v1/organizations/${orgId}`, { method: 'GET' })
      return (await resp.json()) as Organization
    },
    staleTime: 5 * 60 * 1000,
    retry: false,
  })
