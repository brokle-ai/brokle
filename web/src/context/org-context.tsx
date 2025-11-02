'use client'

import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import { useAuth } from '@/hooks/auth/use-auth'
import { 
  getUserOrganizations,
  getOrganizationById,
  getOrganizationProjects,
  createOrganization as apiCreateOrganization
} from '@/lib/api'
import {
  extractIdFromCompositeSlug,
  generateCompositeSlug,
  isValidCompositeSlug
} from '@/lib/utils/slug-utils'
import type { 
  Organization, 
  CreateOrganizationData,
  Project
} from '@/types/organization'

/**
 * TODO: BACKEND INTEGRATION REQUIRED - Permission-Based Access Control
 * 
 * Current Status: Frontend access control has been cleaned up and role-based
 * dependencies removed. The following needs to be implemented with backend:
 *
 * 1. Permission Calculation:
 *    - Backend should calculate user permissions for current organization
 *    - Include permissions in auth response: user.permissions: Permission[]
 *    - Update permissions when switching organizations
 *
 * 2. Access Control Functions (to be added to this interface):
 *    - hasPermission(permission: Permission): boolean
 *    - hasAnyPermission(permissions: Permission[]): boolean  
 *    - hasAllPermissions(permissions: Permission[]): boolean
 *    - getUserPermissions(): Permission[]
 *
 * 3. Integration Points:
 *    - Update useAuth hook to include permission methods
 *    - Create PermissionGuard component using backend-provided permissions
 *    - Replace all TODO comments in pages with actual permission checks
 *
 * 4. Backend API Changes Needed:
 *    - GET /auth/me -> include user.permissions for current org context
 *    - PUT /auth/organization/{orgId}/switch -> recalculate permissions
 *    - RBAC system should map roles to permissions server-side
 *
 * 5. Security Model:
 *    - All permission checking on frontend is for UI/UX only
 *    - Backend MUST verify permissions on all API endpoints
 *    - Never trust frontend permission state for security decisions
 */
interface OrgContextValue {
  // State
  organizations: Organization[]
  currentOrganization: Organization | null  // Computed from currentOrganizationId
  currentOrganizationId: string | null
  projects: Project[]
  isLoading: boolean
  isLoadingProjects: boolean
  isOrgReady: boolean  // True when organizations loaded AND currentOrganizationId resolved
  error: string | null

  // Actions
  loadOrganizationByCompositeSlug: (compositeSlug: string) => Promise<void>
  switchOrganization: (compositeSlug: string) => Promise<void>
  switchProject: (compositeSlug: string) => Promise<void>
  setCurrentOrganizationId: (organizationId: string | null) => void
  createOrganization: (data: CreateOrganizationData) => Promise<Organization>
  clearError: () => void
}

const OrgContext = createContext<OrgContextValue | undefined>(undefined)

interface OrgProviderProps {
  children: ReactNode
  compositeSlug?: string // Current composite slug from URL (e.g., "brokle-tech-01k4mzr3zexw0qe66df8dkbez3")
}

export function OrgProvider({ children, compositeSlug }: OrgProviderProps) {
  const [organizations, setOrganizations] = useState<Organization[]>([])
  const [currentOrganizationId, setCurrentOrganizationId] = useState<string | null>(null)
  const [projects, setProjects] = useState<Project[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [isLoadingProjects, setIsLoadingProjects] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Computed properties
  const currentOrganization = currentOrganizationId 
    ? organizations.find(org => org.id === currentOrganizationId) || null
    : null
    
  const isOrgReady = !isLoading && organizations.length > 0 && (
    currentOrganizationId === null || currentOrganization !== null
  )

  const { user } = useAuth()

  // Load user organizations when user is available
  useEffect(() => {
    if (!user) {
      setOrganizations([])
      setCurrentOrganizationId(null)
      setProjects([])
      setIsLoading(false)
      return
    }

    loadUserOrganizations()
  }, [user])

  // Load organization when composite slug changes
  useEffect(() => {
    if (!user) {
      // No user, clear state
      setCurrentOrganizationId(null)
      setProjects([])
      setIsLoading(false)
      return
    }

    if (!compositeSlug) {
      setCurrentOrganizationId(null)
      setProjects([])
      setIsLoading(false)
      return
    }

    loadOrganizationByCompositeSlug(compositeSlug)
  }, [user, compositeSlug])

  // Load projects when current organization ID changes
  useEffect(() => {
    if (!currentOrganizationId) {
      setProjects([])
      return
    }

    loadOrganizationProjects(currentOrganizationId)
  }, [currentOrganizationId])

  const loadOrganizationByCompositeSlug = async (slug: string) => {
    try {
      setIsLoading(true)
      setError(null)

      // Validate composite slug format
      if (!isValidCompositeSlug(slug)) {
        throw new Error(`Invalid composite slug format: ${slug}`)
      }

      // Extract ID from composite slug
      const organizationId = extractIdFromCompositeSlug(slug)
      
      // Load organization directly by ID
      const organization = await getOrganizationById(organizationId)
      setCurrentOrganizationId(organization.id)

      console.log('[OrgContext] Loaded organization:', organization.name)
    } catch (error) {
      console.error('[OrgContext] Failed to load organization:', error)
      setError(error instanceof Error ? error.message : 'Failed to load organization')
      setCurrentOrganizationId(null)
    } finally {
      setIsLoading(false)
    }
  }

  const loadOrganizationProjects = async (organizationId: string) => {
    try {
      setIsLoadingProjects(true)

      const projectsData = await getOrganizationProjects(organizationId)
      setProjects(projectsData)
    } catch (error) {
      console.error('[OrgContext] Failed to load projects:', error)
      // Don't set main error for project loading failures, just log
      setProjects([])
    } finally {
      setIsLoadingProjects(false)
    }
  }

  const loadUserOrganizations = async () => {
    try {
      setIsLoading(true)
      setError(null)

      const orgData = await getUserOrganizations()
      setOrganizations(orgData.data) // getUserOrganizations returns PaginatedResponse
      
      console.log('[OrgContext] Loaded', orgData.data.length, 'organizations')
    } catch (error) {
      console.error('[OrgContext] Failed to load organizations:', error)
      setError(error instanceof Error ? error.message : 'Failed to load organizations')
      setOrganizations([])
    } finally {
      setIsLoading(false)
    }
  }

  const switchOrganization = async (compositeSlug: string, onProjectClear?: () => void) => {
    try {
      setIsLoading(true)
      setError(null)

      // Extract ID from composite slug
      const orgId = extractIdFromCompositeSlug(compositeSlug)

      // Find organization by ID
      const org = organizations.find(o => o.id === orgId)
      if (!org) {
        throw new Error(`Organization not found for composite slug: ${compositeSlug}`)
      }

      // If it's the same organization, no need to switch
      if (currentOrganizationId === org.id) {
        setIsLoading(false)
        return
      }

      // Clear project context if callback provided (when switching from project page)
      if (onProjectClear) {
        onProjectClear()
        console.log('[OrgContext] Cleared project context before organization switch')
      }

      setCurrentOrganizationId(org.id)
      console.log('[OrgContext] Switched to organization:', org.name)
    } catch (error) {
      console.error('[OrgContext] Failed to switch organization:', error)
      setError(error instanceof Error ? error.message : 'Failed to switch organization')
    } finally {
      setIsLoading(false)
    }
  }

  const createOrganization = async (data: CreateOrganizationData): Promise<Organization> => {
    if (!user) {
      throw new Error('User not authenticated')
    }

    try {
      const newOrg = await apiCreateOrganization({
        name: data.name,
        description: data.description,
      })

      console.log('[OrgContext] Created organization:', newOrg.name)

      return newOrg
    } catch (error) {
      console.error('[OrgContext] Organization creation failed:', error)
      throw error
    }
  }

  const switchProject = async (compositeSlug: string) => {
    try {
      // Extract project ID from composite slug
      const projectId = extractIdFromCompositeSlug(compositeSlug)

      // Verify project exists in current projects list
      const project = projects.find(p => p.id === projectId)
      if (!project) {
        throw new Error(`Project not found for composite slug: ${compositeSlug}`)
      }

      // Verify project belongs to current organization
      if (!currentOrganization) {
        throw new Error('Cannot switch project: no organization selected')
      }

      if (project.organizationId !== currentOrganization.id) {
        throw new Error('Project does not belong to current organization')
      }

      // Navigate to project page (ProjectContext will handle loading project state)
      const orgCompositeSlug = generateCompositeSlug(currentOrganization.name, currentOrganization.id)
      const projectUrl = `/organizations/${orgCompositeSlug}/projects/${compositeSlug}`

      console.log('[OrgContext] Switching to project:', project.name)

      // Use Next.js router for navigation (imported at top of file if not already)
      if (typeof window !== 'undefined') {
        window.location.href = projectUrl
      }
    } catch (error) {
      console.error('[OrgContext] Failed to switch project:', error)
      throw error
    }
  }

  const clearError = () => {
    setError(null)
  }

  const value: OrgContextValue = {
    // State
    organizations,
    currentOrganization,
    currentOrganizationId,
    projects,
    isLoading,
    isLoadingProjects,
    isOrgReady,
    error,

    // Actions
    loadOrganizationByCompositeSlug,
    switchOrganization,
    switchProject,
    setCurrentOrganizationId,
    createOrganization,
    clearError,
  }

  return (
    <OrgContext.Provider value={value}>
      {children}
    </OrgContext.Provider>
  )
}

// Hook to use the org context
export function useOrganization(): OrgContextValue {
  const context = useContext(OrgContext)
  if (context === undefined) {
    throw new Error('useOrganization must be used within an OrgProvider')
  }
  return context
}

// Export for convenience
export { OrgContext }