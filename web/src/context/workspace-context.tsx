'use client'

import { createContext, useContext, ReactNode, useMemo, useState, useEffect, useCallback } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { usePathname } from 'next/navigation'
import { BrokleAPIClient } from '@/lib/api/core/client'
import { extractIdFromCompositeSlug, isValidCompositeSlug } from '@/lib/utils/slug-utils'
import { setDefaultOrganization } from '@/features/authentication/api/auth-api'
import {
  createWorkspaceError,
  classifySlugError,
  classifyAPIError,
  WorkspaceErrorCode,
  type WorkspaceError,
} from './workspace-errors'
import type {
  User,
  OrganizationWithProjects,
  ProjectSummary,
  SubscriptionPlan,
  OrganizationRole,
  ProjectStatus,
  OrganizationMember,
} from '@/features/authentication'
import type {
  EnhancedUserProfileResponse,
  BackendOrganizationWithProjects,
  BackendProjectSummary
} from '@/types/api-responses'

const client = new BrokleAPIClient('/api')

/**
 * WorkspaceContext value interface
 *
 * Provides centralized state management for organizations, projects, and workspace operations.
 * Includes granular loading states and error classification for optimal UX.
 *
 * @example
 * ```tsx
 * const {
 *   currentOrganization,
 *   loadingState,
 *   canInteract,
 *   switchOrganization
 * } = useWorkspace()
 * ```
 */
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
  error: WorkspaceError | null

  // Granular loading states
  loadingState: {
    isInitializing: boolean       // First-time load
    isRefreshing: boolean         // Background refresh
    isSwitchingOrg: boolean       // Organization switch in progress
    isSwitchingProject: boolean   // Project switch in progress
  }

  // Computed convenience flags
  canInteract: boolean            // Can user interact with switchers?

  // Actions
  refresh: () => Promise<void>
  switchOrganization: (compositeSlug: string) => Promise<string>  // Returns org ID
  switchProject: (compositeSlug: string) => Promise<string>        // Returns project ID
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
  const [urlError, setUrlError] = useState<WorkspaceError | null>(null)

  // Centralized loading state
  const [loadingState, setLoadingState] = useState({
    isInitializing: true,
    isRefreshing: false,
    isSwitchingOrg: false,
    isSwitchingProject: false,
  })

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
          status: proj.status as ProjectStatus,
          createdAt: proj.created_at,
          updatedAt: proj.updated_at,
          metrics: {
            traces_collected: 0,
            observed_cost: 0,
            active_rules: 0,
            running_experiments: 0,
          },
        })),
        members: [] as OrganizationMember[], // Will be populated from API when needed
        usage: {
          traces_this_month: 0,
          observed_cost_this_month: 0,
          models_observed: 0,
        },
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

  // Update loading state when query state changes
  useEffect(() => {
    setLoadingState(prev => ({
      ...prev,
      isInitializing: isLoading && !data,
      isRefreshing: isLoading && !!data,
    }))
  }, [isLoading, data])

  // Organization switch handler
  const switchOrganization = useCallback(async (compositeSlug: string): Promise<string> => {
    setLoadingState(prev => ({ ...prev, isSwitchingOrg: true }))

    try {
      const targetOrgId = extractIdFromCompositeSlug(compositeSlug)

      // Update backend default org
      await setDefaultOrganization(targetOrgId)

      // Refresh workspace data
      await queryClient.refetchQueries({ queryKey: ['workspace'] })

      return targetOrgId
    } catch (error) {
      throw classifyAPIError(error)
    } finally {
      setLoadingState(prev => ({ ...prev, isSwitchingOrg: false }))
    }
  }, [queryClient])

  // Project switch handler
  const switchProject = useCallback(async (compositeSlug: string): Promise<string> => {
    setLoadingState(prev => ({ ...prev, isSwitchingProject: true }))

    try {
      const projectId = extractIdFromCompositeSlug(compositeSlug)

      // No API call needed - just navigate
      // Loading state cleared after navigation completes
      return projectId
    } catch {
      throw classifySlugError(compositeSlug, 'project')
    } finally {
      setLoadingState(prev => ({ ...prev, isSwitchingProject: false }))
    }
  }, [])

  // Auto-detect context from URL with proper error handling
  const { currentOrganization, currentProject, detectedUrlError } = useMemo(() => {
    if (!data) return {
      currentOrganization: null,
      currentProject: null,
      detectedUrlError: null
    }

    let urlError: WorkspaceError | null = null

    // Try to detect project from /projects/[projectSlug] and auto-infer parent org
    const projectMatch = pathname.match(/\/projects\/([^/]+)/)
    if (projectMatch) {
      const compositeSlug = projectMatch[1]

      if (!isValidCompositeSlug(compositeSlug)) {
        urlError = classifySlugError(compositeSlug, 'project')
      } else {
        try {
          const projectId = extractIdFromCompositeSlug(compositeSlug)

          // Find project and its parent organization
          for (const org of data.organizations) {
            const project = org.projects.find(p => p.id === projectId)
            if (project) {
              if (process.env.NODE_ENV === 'development') {
                console.log('[Workspace] Detected project from URL:', project.name, 'in org:', org.name)
              }
              return {
                currentOrganization: org,
                currentProject: project,
                detectedUrlError: null
              }
            }
          }

          // Project not found in any organization
          urlError = createWorkspaceError(
            WorkspaceErrorCode.PROJECT_NOT_FOUND,
            { slug: compositeSlug, projectId }
          )
        } catch {
          urlError = classifySlugError(compositeSlug, 'project')
        }
      }
    }

    // Default organization selection with fallback (PostHog pattern)
    // Priority: 1) User's default org 2) First org in list 3) null (no orgs)
    if (!urlError) {
      // Start with first organization as fallback
      let selectedOrg = data.organizations.length > 0 ? data.organizations[0] : null

      // If user has a default organization preference, try to use it
      if (data.user.defaultOrganizationId) {
        const defaultOrg = data.organizations.find(o => o.id === data.user.defaultOrganizationId)
        if (defaultOrg) {
          selectedOrg = defaultOrg
        }
        // If defaultOrganizationId points to invalid org, we fall back to first org
      }

      if (selectedOrg) {
        if (process.env.NODE_ENV === 'development') {
          console.log('[Workspace] Using org:', selectedOrg.name,
            data.user.defaultOrganizationId === selectedOrg.id ? '(user default)' : '(fallback to first)')
        }
        return {
          currentOrganization: selectedOrg,
          currentProject: null,
          detectedUrlError: null
        }
      }
    }

    return {
      currentOrganization: null,
      currentProject: null,
      detectedUrlError: urlError
    }
  }, [data, pathname])

  // Set URL error from useMemo detection
  useEffect(() => {
    setUrlError(detectedUrlError)
  }, [detectedUrlError])

  const value: WorkspaceContextValue = {
    user: data?.user || null,
    organizations: data?.organizations || [],
    currentOrganization,
    currentProject,
    isLoading: isLoading && !data,  // Only loading if no data yet (prevents shimmer during refetch)
    isInitialized: !!data,
    error: queryError ? classifyAPIError(queryError) : urlError,

    // Granular loading states
    loadingState,

    // Computed convenience flags
    canInteract: !loadingState.isSwitchingOrg && !loadingState.isSwitchingProject,

    // Actions
    refresh: async () => {
      await queryClient.refetchQueries({ queryKey: ['workspace'] })
    },
    switchOrganization,
    switchProject,
    clearError: () => setUrlError(null),
  }

  return (
    <WorkspaceContext.Provider value={value}>
      {children}
    </WorkspaceContext.Provider>
  )
}

/**
 * Hook to access workspace context
 *
 * Provides access to current organization, project, loading states, and switch methods.
 * Must be used within a WorkspaceProvider.
 *
 * @throws {Error} If used outside of WorkspaceProvider
 *
 * @example
 * ```tsx
 * function MyComponent() {
 *   const {
 *     currentOrganization,
 *     currentProject,
 *     loadingState,
 *     canInteract,
 *     switchOrganization,
 *     error
 *   } = useWorkspace()
 *
 *   if (error) {
 *     return <ErrorMessage>{error.userMessage}</ErrorMessage>
 *   }
 *
 *   return <div>...</div>
 * }
 * ```
 */
export function useWorkspace(): WorkspaceContextValue {
  const context = useContext(WorkspaceContext)
  if (context === undefined) {
    throw new Error('useWorkspace must be used within WorkspaceProvider')
  }
  return context
}
