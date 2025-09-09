'use client'

import React, { createContext, useContext, useState, useEffect, useCallback, ReactNode } from 'react'
import { useRouter, usePathname } from 'next/navigation'
import { useAuth } from '@/context/auth-context'
import { 
  getUserOrganizations,
  getOrganizationProjects,
  createOrganization as apiCreateOrganization,
  createProject as apiCreateProject
} from '@/lib/api'
import { parsePathContext } from '@/lib/utils/slug-utils'
import type { 
  Organization, 
  Project, 
  OrganizationContext, 
  CreateOrganizationData, 
  CreateProjectData,
  OrganizationRole
} from '@/types/organization'

const OrganizationContextValue = createContext<OrganizationContext | undefined>(undefined)

interface OrganizationProviderProps {
  children: ReactNode
}

export function OrganizationProvider({ children }: OrganizationProviderProps) {
  // ===================================================================
  // STATE MANAGEMENT
  // ===================================================================
  
  const [organizations, setOrganizations] = useState<Organization[]>([])
  const [currentOrganization, setCurrentOrganization] = useState<Organization | null>(null)
  const [currentProject, setCurrentProject] = useState<Project | null>(null)
  const [projects, setProjects] = useState<Project[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // ===================================================================
  // DEPENDENCIES
  // ===================================================================
  
  const { user, isAuthenticated, isLoading: authLoading } = useAuth()
  const router = useRouter()
  const pathname = usePathname()

  // ===================================================================
  // INITIALIZATION & URL HANDLING
  // ===================================================================

  const initializeContext = useCallback(async () => {
    if (!user?.email) {
      console.log('[OrganizationContext] No user email, skipping initialization')
      return
    }

    try {
      setIsLoading(true)
      setError(null)

      console.log('[OrganizationContext] Starting initialization for user:', user.email)

      // Simply load user's organizations for UI display
      const userOrgs = await getUserOrganizations()
      setOrganizations(userOrgs)

      console.log('[OrganizationContext] Successfully loaded', userOrgs.length, 'organizations')

    } catch (error) {
      console.error('[OrganizationContext] Failed to load organizations:', error)
      
      // Check if it's an auth error specifically
      if (error && typeof error === 'object' && 'statusCode' in error && error.statusCode === 401) {
        console.warn('[OrganizationContext] Authentication required - user may not be fully authenticated yet')
        setError('Authentication required. Please sign in.')
      } else {
        setError(error instanceof Error ? error.message : 'Failed to load organizations')
      }
    } finally {
      setIsLoading(false)
    }
  }, [user?.email])

  const updateContextFromUrl = useCallback(async (orgSlug: string, projectSlug?: string) => {
    if (!organizations.length) return

    try {
      // Set loading state during URL context update
      setIsLoading(true)
      setError(null)
      
      // Find organization by slug for UI state
      const org = organizations.find(o => o.slug === orgSlug)
      if (!org) {
        setError(`Organization '${orgSlug}' not found`)
        setIsLoading(false)
        return
      }

      // Update UI state only (no context manager)
      if (!currentOrganization || currentOrganization.id !== org.id) {
        console.log('[OrganizationContext] Switching to organization:', org.name)
        setCurrentOrganization(org)
        
        // Load projects for this organization
        const orgProjects = await getOrganizationProjects(org.id)
        setProjects(orgProjects)
        
        setCurrentProject(null) // Reset project when switching orgs
        
        // Handle project selection AFTER projects are loaded
        if (projectSlug) {
          const project = orgProjects.find(p => p.slug === projectSlug)
          if (project) {
            console.log('[OrganizationContext] Setting current project:', project.name)
            setCurrentProject(project)
          } else {
            setError(`Project '${projectSlug}' not found in organization '${orgSlug}'`)
          }
        }
      } else {
        // Same organization, just handle project change
        if (projectSlug) {
          const project = projects.find(p => p.slug === projectSlug)
          if (project && (!currentProject || currentProject.id !== project.id)) {
            console.log('[OrganizationContext] Setting current project:', project.name)
            setCurrentProject(project)
          } else if (!project) {
            setError(`Project '${projectSlug}' not found in organization '${orgSlug}'`)
          }
        } else {
          // No project slug, clear current project
          setCurrentProject(null)
        }
      }

    } catch (error) {
      console.error('[OrganizationContext] URL update failed:', error)
      setError(error instanceof Error ? error.message : 'Failed to update from URL')
    } finally {
      // Always clear loading state after URL update completes
      setIsLoading(false)
    }
  }, [organizations, currentOrganization, projects, currentProject])

  const clearContext = useCallback(() => {
    setOrganizations([])
    setCurrentOrganization(null)
    setCurrentProject(null)
    setProjects([])
    setError(null)
    setIsLoading(false)
  }, [])

  // ===================================================================
  // ORGANIZATION ACTIONS
  // ===================================================================

  const switchOrganization = useCallback(async (orgSlug: string) => {
    if (!user) {
      throw new Error('User not authenticated')
    }

    try {
      // Find organization by slug
      const org = organizations.find(o => o.slug === orgSlug)
      if (!org) {
        throw new Error(`Organization '${orgSlug}' not found`)
      }

      // Update UI state only  
      setCurrentOrganization(org)
      setCurrentProject(null) // Clear project when switching orgs

      // Load projects for new organization
      const orgProjects = await getOrganizationProjects(org.id)
      setProjects(orgProjects)

      // Navigate to organization dashboard
      router.push(`/${orgSlug}`)
      setError(null)

      console.log('[OrganizationContext] Switched to organization:', org.name)

    } catch (error) {
      console.error('[OrganizationContext] Organization switch failed:', error)
      throw error
    }
  }, [user, organizations, router])

  const createOrganization = useCallback(async (data: CreateOrganizationData): Promise<Organization> => {
    if (!user) {
      throw new Error('User not authenticated')
    }

    try {
      // Create organization via API
      const newOrg = await apiCreateOrganization({
        name: data.name,
        slug: data.slug,
        billing_email: data.billing_email,
        subscription_plan: data.plan,
      })

      // Update local state
      setOrganizations(prev => [...prev, newOrg])

      console.log('[OrganizationContext] Created organization:', newOrg.name)
      return newOrg

    } catch (error) {
      console.error('[OrganizationContext] Organization creation failed:', error)
      throw error
    }
  }, [user])

  // ===================================================================
  // PROJECT ACTIONS
  // ===================================================================

  const switchProject = useCallback(async (projectSlug: string) => {
    if (!user) {
      throw new Error('User not authenticated')
    }

    if (!currentOrganization) {
      throw new Error('No organization selected')
    }

    try {
      // Find project by slug
      const project = projects.find(p => p.slug === projectSlug)
      if (!project) {
        throw new Error(`Project '${projectSlug}' not found`)
      }

      // Update UI state only
      setCurrentProject(project)

      // Navigate to project dashboard
      router.push(`/${currentOrganization.slug}/${projectSlug}`)
      setError(null)

      console.log('[OrganizationContext] Switched to project:', project.name)

    } catch (error) {
      console.error('[OrganizationContext] Project switch failed:', error)
      throw error
    }
  }, [user, currentOrganization, projects, router])

  const createProject = useCallback(async (data: CreateProjectData): Promise<Project> => {
    if (!user) {
      throw new Error('User not authenticated')
    }

    if (!currentOrganization) {
      throw new Error('No organization selected')
    }

    try {
      // Create project via API
      const newProject = await apiCreateProject(currentOrganization.id, {
        name: data.name,
        slug: data.slug,
        description: data.description,
      })

      // Update local state
      setProjects(prev => [...prev, newProject])

      console.log('[OrganizationContext] Created project:', newProject.name)
      return newProject

    } catch (error) {
      console.error('[OrganizationContext] Project creation failed:', error)
      throw error
    }
  }, [user, currentOrganization])

  // ===================================================================
  // COORDINATION EFFECTS
  // ===================================================================

  // Initialize context on auth state change
  useEffect(() => {
    // Wait for auth to fully initialize before doing anything
    if (authLoading) {
      console.log('[OrganizationContext] Auth still loading, waiting...')
      return
    }

    if (isAuthenticated && user?.email) {
      console.log('[OrganizationContext] Auth confirmed, initializing context for user:', user.email)
      initializeContext()
    } else {
      console.log('[OrganizationContext] Not authenticated or no user, clearing context', {
        isAuthenticated,
        hasUser: !!user?.email
      })
      clearContext()
      
      // Set a helpful error message if user is trying to access org routes without auth
      if (typeof window !== 'undefined') {
        const pathname = window.location.pathname
        if (pathname !== '/' && pathname !== '/auth/signin' && pathname !== '/auth/signup') {
          setError('Please sign in to access this page')
        }
      }
    }
  }, [isAuthenticated, user?.email, authLoading, initializeContext, clearContext])

  // Handle URL changes - only after context is initialized and auth is complete
  useEffect(() => {
    // Triple check: auth complete, user exists, and context not loading
    if (!isAuthenticated || !user?.email || authLoading || isLoading) {
      console.log('[OrganizationContext] Skipping URL update:', {
        isAuthenticated,
        hasUser: !!user?.email,
        authLoading,
        contextLoading: isLoading
      })
      return
    }

    // Only proceed if we have organizations loaded
    if (organizations.length === 0) {
      console.log('[OrganizationContext] No organizations loaded yet, skipping URL update')
      return
    }

    const { orgSlug, projectSlug } = parsePathContext(pathname)
    
    // Only update if we have an org slug and it's different from current
    if (orgSlug && (!currentOrganization || currentOrganization.slug !== orgSlug || 
        (projectSlug && (!currentProject || currentProject.slug !== projectSlug)))) {
      console.log('[OrganizationContext] Updating context from URL:', { 
        orgSlug, 
        projectSlug, 
        currentOrgSlug: currentOrganization?.slug,
        currentProjectSlug: currentProject?.slug
      })
      updateContextFromUrl(orgSlug, projectSlug)
    }
  }, [pathname, isAuthenticated, user?.email, authLoading, isLoading, organizations.length, currentOrganization?.slug, currentProject?.slug, updateContextFromUrl])

  // ===================================================================
  // UTILITY FUNCTIONS
  // ===================================================================

  const hasAccess = useCallback((orgSlug: string, projectSlug?: string): boolean => {
    if (!user?.email) return false
    
    const org = organizations.find(o => o.slug === orgSlug)
    if (!org) return false
    
    if (!projectSlug) return true
    
    const project = projects.find(p => p.slug === projectSlug && p.organizationId === org.id)
    return !!project
  }, [user?.email, organizations, projects])

  const getUserRole = useCallback((orgSlug: string): OrganizationRole | null => {
    if (!user?.email) return null
    
    const org = organizations.find(o => o.slug === orgSlug)
    if (!org) return null
    
    // For now, assume user has owner role if they have access
    // This would be enhanced with proper member role checking from API
    return 'owner'
  }, [user?.email, organizations])

  const getProjectsByOrg = useCallback((orgSlug: string): Project[] => {
    const org = organizations.find(o => o.slug === orgSlug)
    if (!org) return []
    
    return projects.filter(p => p.organizationId === org.id)
  }, [organizations, projects])

  // ===================================================================
  // CONTEXT PROVIDER
  // ===================================================================

  const value: OrganizationContext = {
    // State
    organizations,
    currentOrganization,
    currentProject,
    projects,
    isLoading: isLoading || authLoading,
    error: authLoading ? null : error, // Don't show errors while auth is loading

    // Actions
    switchOrganization,
    switchProject,
    createOrganization,
    createProject,

    // Utils
    hasAccess,
    getUserRole,
    getProjectsByOrg,
  }

  return (
    <OrganizationContextValue.Provider value={value}>
      {children}
    </OrganizationContextValue.Provider>
  )
}

// Hook to use the organization context
export function useOrganization(): OrganizationContext {
  const context = useContext(OrganizationContextValue)
  if (context === undefined) {
    throw new Error('useOrganization must be used within an OrganizationProvider')
  }
  return context
}

// Hook to use current project
export function useProject() {
  const { currentProject, projects, currentOrganization } = useOrganization()

  return {
    currentProject,
    projects,
    organization: currentOrganization,
    hasProject: currentProject !== null,
    projectsInCurrentOrg: projects,
  }
}

// Export for convenience
export { OrganizationContextValue as OrganizationContext }