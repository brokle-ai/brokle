/**
 * Context initialization utilities for organization and project context
 */

import { getUserOrganizations, getOrganizationProjects } from '@/lib/api'
import { findProjectBySlugInOrganization } from '@/lib/utils/organization-utils'
import { 
  resolveContextFromPath,
  getDefaultContext,
  canAccessContext
} from '@/lib/utils/context-resolver'
import { parsePathContext } from '@/lib/utils/slug-utils'
import type { Organization, Project } from '@/types/organization'

export interface InitializationState {
  organizations: Organization[]
  currentOrganization: Organization | null
  currentProject: Project | null
  projects: Project[]
  error: string | null
}

export interface InitializationOptions {
  userEmail: string
  pathname: string
  defaultOrganizationId?: string
  getLastContext: () => { organizationSlug: string; projectSlug?: string } | null
  updateProjectsList: (organizationId: string) => Promise<Project[]>
}

export interface InitializationResult {
  state: InitializationState
  contextToSave: { orgSlug: string; projectSlug?: string } | null
}

/**
 * Initialize context with comprehensive fallback strategy
 */
export async function initializeContext(options: InitializationOptions): Promise<InitializationResult> {
  const { userEmail, pathname, defaultOrganizationId, getLastContext, updateProjectsList } = options

  try {
    // Load user's organizations
    const userOrgs = await getUserOrganizations()

    if (userOrgs.length === 0) {
      return {
        state: {
          organizations: [],
          currentOrganization: null,
          currentProject: null,
          projects: [],
          error: 'No organizations found for user',
        },
        contextToSave: null,
      }
    }

    // Try to restore from URL first
    const { orgSlug, projectSlug } = parsePathContext(pathname)
    if (orgSlug) {
      const resolvedContext = await resolveContextFromPath({
        userEmail,
        pathname,
      })

      if (resolvedContext.hasAccess && resolvedContext.organization) {
        const projects = await updateProjectsList(resolvedContext.organization.id)
        
        return {
          state: {
            organizations: userOrgs,
            currentOrganization: resolvedContext.organization,
            currentProject: resolvedContext.project,
            projects,
            error: null,
          },
          contextToSave: {
            orgSlug: resolvedContext.organization.slug,
            projectSlug: resolvedContext.project?.slug,
          },
        }
      }
    }

    // Try to restore from persistence
    const lastContext = getLastContext()
    if (lastContext) {
      const canAccess = await canAccessContext(userEmail, lastContext.organizationSlug, lastContext.projectSlug)
      if (canAccess) {
        const org = userOrgs.find(o => o.slug === lastContext.organizationSlug)
        if (org) {
          const projects = await updateProjectsList(org.id)
          let project: Project | null = null

          if (lastContext.projectSlug) {
            project = await findProjectBySlugInOrganization(org.id, lastContext.projectSlug) || null
          }

          return {
            state: {
              organizations: userOrgs,
              currentOrganization: org,
              currentProject: project,
              projects,
              error: null,
            },
            contextToSave: null, // Already saved
          }
        }
      }
    }

    // Fallback to default context (user's default org or first org)
    const defaultContext = await getDefaultContext(userEmail, defaultOrganizationId)
    if (defaultContext.organization) {
      const projects = await updateProjectsList(defaultContext.organization.id)
      
      return {
        state: {
          organizations: userOrgs,
          currentOrganization: defaultContext.organization,
          currentProject: defaultContext.project,
          projects,
          error: null,
        },
        contextToSave: {
          orgSlug: defaultContext.organization.slug,
          projectSlug: defaultContext.project?.slug,
        },
      }
    }

    // Should never reach here, but handle gracefully
    return {
      state: {
        organizations: userOrgs,
        currentOrganization: null,
        currentProject: null,
        projects: [],
        error: 'Could not determine default context',
      },
      contextToSave: null,
    }

  } catch (error) {
    console.error('[ContextInitialization] Failed:', error)
    return {
      state: {
        organizations: [],
        currentOrganization: null,
        currentProject: null,
        projects: [],
        error: error instanceof Error ? error.message : 'Failed to initialize context',
      },
      contextToSave: null,
    }
  }
}

/**
 * Update context from URL with access validation
 */
export interface URLUpdateOptions {
  userEmail: string
  orgSlug: string
  projectSlug?: string
  updateProjectsList: (organizationId: string) => Promise<Project[]>
}

export interface URLUpdateResult {
  state: Partial<InitializationState>
  contextToSave: { orgSlug: string; projectSlug?: string } | null
}

export async function updateContextFromURL(options: URLUpdateOptions): Promise<URLUpdateResult> {
  const { userEmail, orgSlug, projectSlug, updateProjectsList } = options

  const resolvedContext = await resolveContextFromPath({
    userEmail,
    pathname: projectSlug ? `/${orgSlug}/${projectSlug}` : `/${orgSlug}`,
  })

  if (!resolvedContext.hasAccess) {
    return {
      state: {
        error: resolvedContext.error || 'Access denied',
      },
      contextToSave: null,
    }
  }

  if (resolvedContext.organization) {
    const projects = await updateProjectsList(resolvedContext.organization.id)
    
    return {
      state: {
        currentOrganization: resolvedContext.organization,
        currentProject: resolvedContext.project,
        projects,
        error: null,
      },
      contextToSave: {
        orgSlug: resolvedContext.organization.slug,
        projectSlug: resolvedContext.project?.slug,
      },
    }
  }

  return {
    state: {
      error: 'Could not resolve context from URL',
    },
    contextToSave: null,
  }
}