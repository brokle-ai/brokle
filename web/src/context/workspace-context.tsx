'use client'

import { createContext, useContext, ReactNode, useMemo, useState, useEffect } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { usePathname } from 'next/navigation'
import { BrokleAPIClient } from '@/lib/api/core/client'
import { extractIdFromCompositeSlug, isValidCompositeSlug } from '@/lib/utils/slug-utils'
import type {
  User,
  OrganizationWithProjects,
  ProjectSummary,
  SubscriptionPlan,
  OrganizationRole,
} from '@/types/auth'
import type {
  EnhancedUserProfileResponse,
  BackendOrganizationWithProjects,
  BackendProjectSummary
} from '@/types/api-responses'

const client = new BrokleAPIClient('/api')

interface WorkspaceContextValue {
  // Data
  user: User | null
  organizations: OrganizationWithProjects[]

  // Current context (auto-selected from URL)
  currentOrganization: OrganizationWithProjects | null
  currentProject: ProjectSummary | null

  // State
  isLoading: boolean
  isInitialized: boolean
  error: string | null

  // Actions
  refresh: () => void
  clearError: () => void
}

const WorkspaceContext = createContext<WorkspaceContextValue | undefined>(undefined)

interface WorkspaceData {
  user: User
  organizations: OrganizationWithProjects[]
}

export function WorkspaceProvider({ children }: { children: ReactNode }) {
  const pathname = usePathname()
  const queryClient = useQueryClient()
  const [urlError, setUrlError] = useState<string | null>(null)

  // Fetch workspace data with React Query
  const { data, isLoading, error: queryError } = useQuery<WorkspaceData>({
    queryKey: ['workspace'],
    queryFn: async () => {
      const response = await client.get<EnhancedUserProfileResponse>('/v1/users/me')

      // Map user (all existing fields preserved)
      const user: User = {
        id: response.id,
        email: response.email,
        firstName: response.first_name,
        lastName: response.last_name,
        name: `${response.first_name} ${response.last_name}`.trim(),
        role: 'user',
        organizationId: '',
        defaultOrganizationId: response.default_organization_id ?? undefined,
        projects: [],
        createdAt: response.created_at,
        updatedAt: response.updated_at,
        isEmailVerified: response.is_email_verified,
        onboardingCompletedAt: response.onboarding_completed_at ?? undefined,
        organizations: [], // Will be set below
      }

      // Map organizations with nested projects
      const organizations: OrganizationWithProjects[] = (response.organizations || []).map((org: BackendOrganizationWithProjects) => ({
        id: org.id,
        name: org.name,
        compositeSlug: org.composite_slug,
        plan: org.plan as SubscriptionPlan,
        role: org.role as OrganizationRole,
        createdAt: org.created_at,
        updatedAt: org.updated_at,
        projects: (org.projects || []).map((proj: BackendProjectSummary) => ({
          id: proj.id,
          name: proj.name,
          compositeSlug: proj.composite_slug,
          description: proj.description || '',
          organizationId: proj.organization_id,
          createdAt: proj.created_at,
          updatedAt: proj.updated_at,
        })),
      }))

      user.organizations = organizations

      // Development logging only
      if (process.env.NODE_ENV === 'development') {
        console.log('[Workspace] Loaded data:', {
          user: user.email,
          orgCount: organizations.length,
          projectCount: organizations.reduce((sum, org) => sum + org.projects.length, 0),
        })
      }

      return { user, organizations }
    },
    staleTime: 5 * 60 * 1000,       // 5 minutes
    gcTime: 10 * 60 * 1000,         // 10 minutes
    refetchOnWindowFocus: true,
    refetchOnReconnect: true,
    retry: 3,                       // Retry failed requests 3 times
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),  // Exponential backoff
  })

  // Auto-detect context from URL (no side effects in useMemo)
  const { currentOrganization, currentProject } = useMemo(() => {
    if (!data) return { currentOrganization: null, currentProject: null }

    // Try to detect organization from /organizations/[orgSlug]
    const orgMatch = pathname.match(/\/organizations\/([^/]+)/)
    if (orgMatch) {
      const compositeSlug = orgMatch[1]
      if (isValidCompositeSlug(compositeSlug)) {
        try {
          const orgId = extractIdFromCompositeSlug(compositeSlug)
          const org = data.organizations.find(o => o.id === orgId)
          if (org) {
            if (process.env.NODE_ENV === 'development') {
              console.log('[Workspace] Detected org from URL:', org.name)
            }
            return { currentOrganization: org, currentProject: null }
          }
        } catch (err) {
          // Invalid slug - will be handled in useEffect
          console.warn('[Workspace] Failed to extract org ID from slug:', compositeSlug)
        }
      }
    }

    // Try to detect project from /projects/[projectSlug] and auto-infer parent org
    const projectMatch = pathname.match(/\/projects\/([^/]+)/)
    if (projectMatch) {
      const compositeSlug = projectMatch[1]
      if (isValidCompositeSlug(compositeSlug)) {
        try {
          const projectId = extractIdFromCompositeSlug(compositeSlug)

          // Find project and its parent organization
          for (const org of data.organizations) {
            const project = org.projects.find(p => p.id === projectId)
            if (project) {
              if (process.env.NODE_ENV === 'development') {
                console.log('[Workspace] Detected project from URL:', project.name, 'in org:', org.name)
              }
              return { currentOrganization: org, currentProject: project }
            }
          }
        } catch (err) {
          // Invalid slug - will be handled in useEffect
          console.warn('[Workspace] Failed to extract project ID from slug:', compositeSlug)
        }
      }
    }

    // Default: use user's default organization
    if (data.user.defaultOrganizationId) {
      const defaultOrg = data.organizations.find(o => o.id === data.user.defaultOrganizationId)
      if (defaultOrg) {
        if (process.env.NODE_ENV === 'development') {
          console.log('[Workspace] Using default org:', defaultOrg.name)
        }
        return { currentOrganization: defaultOrg, currentProject: null }
      }
    }

    return { currentOrganization: null, currentProject: null }
  }, [data, pathname])

  // Handle URL errors in useEffect (side effects belong here, not useMemo)
  useEffect(() => {
    const hasOrgOrProjectInURL = pathname.match(/\/(organizations|projects)\//)
    if (hasOrgOrProjectInURL && !currentOrganization && !isLoading) {
      setUrlError('Organization or project not found')
    } else {
      setUrlError(null)
    }
  }, [currentOrganization, currentProject, pathname, isLoading])

  const value: WorkspaceContextValue = {
    user: data?.user || null,
    organizations: data?.organizations || [],
    currentOrganization,
    currentProject,
    isLoading: isLoading && !data,  // Only loading if no data yet (prevents shimmer during refetch)
    isInitialized: !!data,
    error: queryError?.message || urlError,
    refresh: () => queryClient.invalidateQueries({ queryKey: ['workspace'] }),
    clearError: () => setUrlError(null),
  }

  return (
    <WorkspaceContext.Provider value={value}>
      {children}
    </WorkspaceContext.Provider>
  )
}

// Export useWorkspace hook
export function useWorkspace(): WorkspaceContextValue {
  const context = useContext(WorkspaceContext)
  if (context === undefined) {
    throw new Error('useWorkspace must be used within WorkspaceProvider')
  }
  return context
}
