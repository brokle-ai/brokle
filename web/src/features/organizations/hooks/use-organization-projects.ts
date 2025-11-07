'use client'

import { useQuery } from '@tanstack/react-query'
import { useWorkspace } from '@/context/workspace-context'
import { getOrganizationProjects } from '../api/organizations-api'

/**
 * Hook to fetch projects for the current organization
 */
export function useOrganizationProjects() {
  const { currentOrganization } = useWorkspace()

  return useQuery({
    queryKey: ['organizations', currentOrganization?.id, 'projects'],
    queryFn: () => {
      if (!currentOrganization?.id) {
        throw new Error('No organization selected')
      }
      return getOrganizationProjects(currentOrganization.id)
    },
    enabled: !!currentOrganization?.id,
    staleTime: 60 * 1000, // 1 minute
  })
}
