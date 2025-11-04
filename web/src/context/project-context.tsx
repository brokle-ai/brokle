'use client'

import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import { useAuth } from '@/hooks/auth/use-auth'
import { useOrganization } from '@/context/org-context'
import { 
  getProjectById,
  createProject as apiCreateProject
} from '@/lib/api'
import { 
  extractIdFromCompositeSlug, 
  isValidCompositeSlug 
} from '@/lib/utils/slug-utils'
import type { 
  Project, 
  CreateProjectData
} from '@/types/organization'

interface ProjectContextValue {
  // State
  currentProject: Project | null
  isLoading: boolean
  error: string | null

  // Actions
  loadProjectByCompositeSlug: (compositeSlug: string) => Promise<void>
  createProject: (organizationId: string, data: CreateProjectData) => Promise<Project>
  setCurrentProject: (project: Project | null) => void  // For external context clearing
  clearError: () => void
}

const ProjectContext = createContext<ProjectContextValue | undefined>(undefined)

interface ProjectProviderProps {
  children: ReactNode
  compositeSlug?: string // Current composite slug from URL (e.g., "analytics-platform-01k4mzr4f36gmrf5r21v5fkxvj")
}

export function ProjectProvider({ children, compositeSlug }: ProjectProviderProps) {
  const [currentProject, setCurrentProject] = useState<Project | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const { user } = useAuth()
  const { organizations, setCurrentOrganizationId } = useOrganization()

  // Load project when composite slug changes
  useEffect(() => {
    if (!user) {
      // No user, clear state
      setCurrentProject(null)
      setIsLoading(false)
      return
    }

    if (!compositeSlug) {
      setCurrentProject(null)
      setIsLoading(false)
      return
    }

    loadProjectByCompositeSlug(compositeSlug)
  }, [user, compositeSlug])

  // Sync organization ID from current project with race condition guards
  useEffect(() => {
    // Guard: Only sync when both currentProject and organizations are loaded
    if (!currentProject || !organizations.length) {
      return
    }

    const projectOrgId = currentProject.organizationId
    
    // Development warning: Check if project's organizationId exists in organizations array
    if (process.env.NODE_ENV === 'development') {
      const orgExists = organizations.find(org => org.id === projectOrgId)
      if (!orgExists) {
        console.warn('[ProjectProvider] Project organizationId not found in organizations array:', {
          projectId: currentProject.id,
          projectName: currentProject.name,
          organizationId: projectOrgId,
          availableOrgs: organizations.map(org => ({ id: org.id, name: org.name }))
        })
      }
    }

    // Sync organization ID to ensure organization selector shows correct org
    setCurrentOrganizationId(projectOrgId)
    
    console.log('[ProjectContext] Synced organization from project:', {
      projectName: currentProject.name,
      organizationId: projectOrgId
    })
  }, [currentProject, organizations, setCurrentOrganizationId])

  const loadProjectByCompositeSlug = async (slug: string) => {
    try {
      setIsLoading(true)
      setError(null)

      // Validate composite slug format
      if (!isValidCompositeSlug(slug)) {
        throw new Error(`Invalid composite slug format: ${slug}`)
      }

      // Extract ID from composite slug
      const projectId = extractIdFromCompositeSlug(slug)
      
      // Load project directly by ID
      const project = await getProjectById(projectId)
      setCurrentProject(project)

      console.log('[ProjectContext] Loaded project:', project.name)
    } catch (error) {
      console.error('[ProjectContext] Failed to load project:', error)
      setError(error instanceof Error ? error.message : 'Failed to load project')
      setCurrentProject(null)
    } finally {
      setIsLoading(false)
    }
  }

  const createProject = async (
    organizationId: string, 
    data: CreateProjectData
  ): Promise<Project> => {
    if (!user) {
      throw new Error('User not authenticated')
    }

    try {
      const newProject = await apiCreateProject(organizationId, {
        name: data.name,
        slug: data.slug,
        description: data.description,
      })

      console.log('[ProjectContext] Created project:', newProject.name)
      
      return newProject
    } catch (error) {
      console.error('[ProjectContext] Project creation failed:', error)
      throw error
    }
  }

  const clearError = () => {
    setError(null)
  }

  const value: ProjectContextValue = {
    // State
    currentProject,
    isLoading,
    error,

    // Actions
    loadProjectByCompositeSlug,
    createProject,
    setCurrentProject,
    clearError,
  }

  return (
    <ProjectContext.Provider value={value}>
      {children}
    </ProjectContext.Provider>
  )
}

// Hook to use the project context
export function useProject(): ProjectContextValue {
  const context = useContext(ProjectContext)
  if (context === undefined) {
    throw new Error('useProject must be used within a ProjectProvider')
  }
  return context
}

// Export for convenience
export { ProjectContext }