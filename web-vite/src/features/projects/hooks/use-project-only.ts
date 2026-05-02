// Port-shim of web/'s useProjectOnly hook. The canonical web/ impl
// reads from a Workspace context; in web-vite we derive the current
// project from the TanStack Router `/o/$orgId/p/$projectId` route
// params. That gives us projectId + compositeSlug for free without
// needing a global workspace store.

import { useParams } from '@tanstack/react-router'

export interface ProjectSummary {
  id: string
  name: string
  description: string
  status: 'active' | 'inactive'
  compositeSlug: string
  createdAt: string
  updatedAt: string
}

export interface ProjectOnlyContext {
  currentProject: ProjectSummary | null
  projects: ProjectSummary[]
  organization: null
  isLoading: boolean
  error: null
  switchProject: (compositeSlug: string) => void
  hasProject: boolean
  projectCount: number
  hasMultipleProjects: boolean
  projectsInCurrentOrg: ProjectSummary[]
}

// strict: false so we can call from any descendant of the
// `/_authenticated/o/$orgId/p/$projectId` tree without the type
// system insisting on a specific leaf route.
export function useProjectOnly(): ProjectOnlyContext {
  const params = useParams({ strict: false }) as {
    orgId?: string
    projectId?: string
  }

  const projectId = params.projectId ?? ''
  const orgId = params.orgId ?? ''

  const currentProject: ProjectSummary | null = projectId
    ? {
        id: projectId,
        name: '',
        description: '',
        status: 'active',
        compositeSlug: `${orgId}/${projectId}`,
        createdAt: '',
        updatedAt: '',
      }
    : null

  return {
    currentProject,
    projects: currentProject ? [currentProject] : [],
    organization: null,
    isLoading: false,
    error: null,
    switchProject: () => {
      // Not used by the datasets feature — leave as a no-op shim.
    },
    hasProject: currentProject !== null,
    projectCount: currentProject ? 1 : 0,
    hasMultipleProjects: false,
    projectsInCurrentOrg: currentProject ? [currentProject] : [],
  }
}
