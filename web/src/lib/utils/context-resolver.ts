/**
 * Context resolution utilities for organization and project routing
 */

import { getOrganizationBySlug, getUserOrganizations, checkUserHasAccessToOrganization } from '@/lib/data/organizations'
import { getProjectBySlug, getProjectsByOrganizationSlug } from '@/lib/data/projects'
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
export function resolveContextFromPath(
  options: ContextResolutionOptions
): ResolvedContext {
  const { userEmail, pathname, fallbackToFirstOrg = false } = options
  const { orgSlug, projectSlug } = parsePathContext(pathname)

  // No organization in path
  if (!orgSlug) {
    if (fallbackToFirstOrg) {
      const userOrgs = getUserOrganizations(userEmail)
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
  const organization = getOrganizationBySlug(orgSlug)
  if (!organization) {
    return {
      organization: null,
      project: null,
      hasAccess: false,
      userRole: null,
      error: `Organization "${orgSlug}" not found`,
    }
  }

  const hasOrgAccess = checkUserHasAccessToOrganization(userEmail, orgSlug)
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
  const project = getProjectBySlug(organization.id, projectSlug)
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
export function getAvailableOrganizations(userEmail: string): Organization[] {
  return getUserOrganizations(userEmail)
}

/**
 * Get available projects for a user in an organization
 */
export function getAvailableProjects(userEmail: string, orgSlug: string): Project[] {
  const hasAccess = checkUserHasAccessToOrganization(userEmail, orgSlug)
  if (!hasAccess) return []
  
  return getProjectsByOrganizationSlug(orgSlug)
}

/**
 * Get default context for a user (their default organization or first organization)
 */
export function getDefaultContext(userEmail: string, defaultOrgId?: string): {
  organization: Organization | null
  project: Project | null
} {
  const userOrgs = getUserOrganizations(userEmail)
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
  const orgProjects = getProjectsByOrganization(selectedOrg.id)
  const firstProject = orgProjects.find(p => p.status === 'active') || orgProjects[0] || null

  return {
    organization: selectedOrg,
    project: firstProject,
  }
}

/**
 * Check if user can access a specific context
 */
export function canAccessContext(
  userEmail: string,
  orgSlug: string,
  projectSlug?: string
): boolean {
  const context = resolveContextFromPath({
    userEmail,
    pathname: projectSlug ? `/${orgSlug}/${projectSlug}` : `/${orgSlug}`,
  })

  return context.hasAccess && !context.error
}

/**
 * Get user's role in organization
 */
export function getUserRole(userEmail: string, orgSlug: string): OrganizationRole | null {
  const organization = getOrganizationBySlug(orgSlug)
  if (!organization) return null

  const member = organization.members.find(m => m.email === userEmail)
  return member ? (member.role as OrganizationRole) : null
}

// Helper to import the function we need
function getProjectsByOrganization(organizationId: string): Project[] {
  // Import here to avoid circular dependencies
  const { getProjectsByOrganization: getProjects } = require('@/lib/data/projects')
  return getProjects(organizationId)
}