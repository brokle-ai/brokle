/**
 * Context resolution utilities for organization and project routing
 */

import { getUserOrganizations, getOrganizationProjects } from '@/lib/api'
import { 
  findOrganizationBySlug, 
  findProjectBySlugInOrganization, 
  checkUserHasAccessToOrganization 
} from '@/lib/utils/organization-utils'
import { parsePathContext } from './slug-utils'
import type { Organization, Project, OrganizationRole } from '@/types/organization'

export interface ResolvedContext {
  organization: Organization | null
  project: Project | null
  hasAccess: boolean
  userRole: OrganizationRole | null
  error: string | null
}

export interface ContextResolutionOptions {
  userEmail: string
  pathname: string
  fallbackToFirstOrg?: boolean
}

/**
 * Resolve context from URL pathname and user information
 */
export async function resolveContextFromPath(
  options: ContextResolutionOptions
): Promise<ResolvedContext> {
  const { userEmail, pathname, fallbackToFirstOrg = false } = options
  const { orgSlug, projectSlug } = parsePathContext(pathname)

  // No organization in path
  if (!orgSlug) {
    if (fallbackToFirstOrg) {
      const userOrgs = await getUserOrganizations()
      if (userOrgs.length > 0) {
        return {
          organization: userOrgs[0],
          project: null,
          hasAccess: true,
          userRole: userOrgs[0].members.find(m => m.email === userEmail)?.role as OrganizationRole || null,
          error: null,
        }
      }
    }
    
    return {
      organization: null,
      project: null,
      hasAccess: false,
      userRole: null,
      error: null,
    }
  }

  // Check organization access
  const organization = await findOrganizationBySlug(orgSlug)
  if (!organization) {
    return {
      organization: null,
      project: null,
      hasAccess: false,
      userRole: null,
      error: `Organization "${orgSlug}" not found`,
    }
  }

  const hasOrgAccess = await checkUserHasAccessToOrganization(userEmail, organization.id)
  if (!hasOrgAccess) {
    return {
      organization,
      project: null,
      hasAccess: false,
      userRole: null,
      error: `Access denied to organization "${orgSlug}"`,
    }
  }

  const userRole = organization.members.find(m => m.email === userEmail)?.role as OrganizationRole

  // No project in path - return org context
  if (!projectSlug) {
    return {
      organization,
      project: null,
      hasAccess: true,
      userRole,
      error: null,
    }
  }

  // Resolve project context
  const project = await findProjectBySlugInOrganization(organization.id, projectSlug)
  if (!project) {
    return {
      organization,
      project: null,
      hasAccess: true,
      userRole,
      error: `Project "${projectSlug}" not found in organization "${orgSlug}"`,
    }
  }

  return {
    organization,
    project,
    hasAccess: true,
    userRole,
    error: null,
  }
}

/**
 * Get available organizations for a user
 */
export async function getAvailableOrganizations(userEmail: string): Promise<Organization[]> {
  return await getUserOrganizations()
}

/**
 * Get available projects for a user in an organization
 */
export async function getAvailableProjects(userEmail: string, orgSlug: string): Promise<Project[]> {
  const organization = await findOrganizationBySlug(orgSlug)
  if (!organization) return []
  
  const hasAccess = await checkUserHasAccessToOrganization(userEmail, organization.id)
  if (!hasAccess) return []
  
  return await getOrganizationProjects(organization.id)
}

/**
 * Get default context for a user (their default organization or first organization)
 */
export async function getDefaultContext(userEmail: string, defaultOrgId?: string): Promise<{
  organization: Organization | null
  project: Project | null
}> {
  const userOrgs = await getUserOrganizations()
  if (userOrgs.length === 0) {
    return { organization: null, project: null }
  }

  // Try to find the user's default organization first
  let defaultOrg = null
  if (defaultOrgId) {
    defaultOrg = userOrgs.find(org => org.id === defaultOrgId)
  }

  // Fall back to first organization if no default or default not found
  const selectedOrg = defaultOrg || userOrgs[0]
  const orgProjects = await getOrganizationProjects(selectedOrg.id)
  const firstProject = orgProjects.find(p => p.status === 'active') || orgProjects[0] || null

  return {
    organization: selectedOrg,
    project: firstProject,
  }
}

/**
 * Check if user can access a specific context
 */
export async function canAccessContext(
  userEmail: string,
  orgSlug: string,
  projectSlug?: string
): Promise<boolean> {
  const context = await resolveContextFromPath({
    userEmail,
    pathname: projectSlug ? `/${orgSlug}/${projectSlug}` : `/${orgSlug}`,
  })

  return context.hasAccess && !context.error
}

/**
 * Get user's role in organization
 */
export async function getUserRole(userEmail: string, orgSlug: string): Promise<OrganizationRole | null> {
  const organization = await findOrganizationBySlug(orgSlug)
  if (!organization) return null

  // Use the utility function to get user role
  const { getUserRoleInOrganization } = await import('@/lib/utils/organization-utils')
  return await getUserRoleInOrganization(userEmail, organization.id)
}

// This helper is no longer needed - using direct API calls