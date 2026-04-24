'use client'

import { createContext, useContext, ReactNode, useMemo, useState, useEffect, useCallback } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { usePathname } from 'next/navigation'
import { BrokleAPIClient } from '@/lib/api/core/client'
import { extractIdFromCompositeSlug, isValidCompositeSlug } from '@/lib/utils/slug-utils'
import { setDefaultOrganization } from '@/features/authentication/api/auth-api'
import { mapEnhancedUserProfile } from '@/lib/auth/profile-mapper'
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

interface WorkspaceProviderProps {
  children: ReactNode
  /**
   * Server-fetched workspace data. When provided, the React Query
   * cache is seeded with this payload and no client-side fetch fires
   * on mount — the dashboard renders fully populated from the
   * Server Component DAL bootstrap (lib/auth/dal.ts).
   *
   * Optional only to keep this component reusable for non-dashboard
   * mount points (Storybook, isolated tests). Production dashboard
   * paths always pass it.
   */
  initialData?: WorkspaceData
}

export function WorkspaceProvider({ children, initialData }: WorkspaceProviderProps) {
  const pathname = usePathname()
  const queryClient = useQueryClient()
  const [urlError, setUrlError] = useState<WorkspaceError | null>(null)

  // Centralized loading state. Initialised from server-fetched data
  // when present (no spinner on cold load); otherwise `isInitializing`
  // starts true until the first useQuery fetch resolves.
  const [loadingState, setLoadingState] = useState({
    isInitializing: !initialData,
    isRefreshing: false,
    isSwitchingOrg: false,
    isSwitchingProject: false,
  })

  // Fetch workspace data with React Query.
  //
  // When `initialData` is provided (production dashboard path), the
  // query is seeded with the server-fetched payload and treated as
  // fresh for `staleTime`, so no fetch fires on mount. Refetches
  // (window focus, reconnect, manual `refresh()`) still go through
  // the queryFn for stale revalidation. The mapping logic uses the
  // shared profile-mapper module so server and client stay in lockstep.
  const { data, isLoading, error: queryError } = useQuery<WorkspaceData>({
    queryKey: ['workspace'],
    queryFn: async () => {
      const response = await client.get<EnhancedUserProfileResponse>('/v1/users/me')
      const mapped = mapEnhancedUserProfile(response)
      if (process.env.NODE_ENV === 'development') {
        console.log('[Workspace] Refetched data:', {
          user: mapped.user.email,
          orgCount: mapped.organizations.length,
          projectCount: mapped.organizations.reduce((sum, org) => sum + org.projects.length, 0),
        })
      }
      return mapped
    },
    initialData,
    initialDataUpdatedAt: initialData ? Date.now() : undefined,
    staleTime: 5 * 60 * 1000,       // 5 minutes
    gcTime: 10 * 60 * 1000,         // 10 minutes
    refetchOnWindowFocus: true,
    refetchOnReconnect: true,
    // Retry policy is inherited from the global QueryClient default in
    // components/providers.tsx, which correctly skips retries on 4xx.
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
