/**
 * Context action utilities for organization and project management
 */

import { canAccessContext } from '@/lib/utils/context-resolver'
import { findOrganizationBySlug, findProjectBySlugInOrganization } from '@/lib/utils/organization-utils'
import { setDefaultOrganization } from '@/lib/api'
import type { 
  Organization, 
  Project, 
  CreateOrganizationData, 
  CreateProjectData
} from '@/types/organization'
import type { User } from '@/types/auth'

export interface ActionDependencies {
  user: User
  router: {
    push: (url: string) => void
  }
  saveLastContext: (orgSlug: string, projectSlug?: string) => void
  updateProjectsList: (organizationId: string) => Promise<void>
}

export interface ActionResult<T = void> {
  success: boolean
  data?: T
  error?: string
}

/**
 * Switch to a different organization
 */
export async function switchOrganization(
  orgSlug: string,
  deps: ActionDependencies
): Promise<ActionResult<Organization>> {
  const { user, router, saveLastContext, updateProjectsList } = deps

  if (!user?.email) {
    return { success: false, error: 'User not authenticated' }
  }

  const canAccess = canAccessContext(user.email, orgSlug)
  if (!canAccess) {
    return { success: false, error: `Access denied to organization "${orgSlug}"` }
  }

  const organization = await findOrganizationBySlug(orgSlug)
  if (!organization) {
    return { success: false, error: `Organization "${orgSlug}" not found` }
  }

  // Update state through callbacks
  await updateProjectsList(organization.id)
  saveLastContext(organization.slug)

  // Update user's default organization in backend (async, don't block navigation)
  try {
    await setDefaultOrganization(organization.id)
  } catch (error) {
    // Log error but don't fail the organization switch
    console.warn('[SwitchOrganization] Failed to update default organization:', error)
  }

  // Navigate to organization
  router.push(`/${orgSlug}`)

  return { success: true, data: organization }
}

/**
 * Switch to a different project
 */
export async function switchProject(
  projectSlug: string,
  currentOrganization: Organization | null,
  deps: Pick<ActionDependencies, 'user' | 'router' | 'saveLastContext'>
): Promise<ActionResult<Project>> {
  const { user, router, saveLastContext } = deps

  if (!currentOrganization) {
    return { success: false, error: 'No organization selected' }
  }

  if (!user?.email) {
    return { success: false, error: 'User not authenticated' }
  }

  const project = await findProjectBySlugInOrganization(currentOrganization.id, projectSlug)
  if (!project) {
    return { success: false, error: `Project "${projectSlug}" not found` }
  }

  // Update persistence
  saveLastContext(currentOrganization.slug, project.slug)

  // Navigate to project
  router.push(`/${currentOrganization.slug}/${projectSlug}`)

  return { success: true, data: project }
}

/**
 * Create a new organization
 */
export async function createOrganization(
  data: CreateOrganizationData,
  user: User
): Promise<ActionResult<Organization>> {
  if (!user?.email) {
    return { success: false, error: 'User not authenticated' }
  }

  // In a real app, this would call the API
  const newOrg: Organization = {
    id: `org-${Date.now()}`,
    name: data.name,
    slug: data.slug || data.name.toLowerCase().replace(/\s+/g, '-'),
    plan: data.plan || 'free',
    billing_email: data.billing_email,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    members: [{
      id: user.id,
      email: user.email,
      role: 'owner',
      name: user.name || 'User',
      joined_at: new Date().toISOString(),
    }],
  }

  return { success: true, data: newOrg }
}

/**
 * Create a new project
 */
export async function createProject(
  data: CreateProjectData,
  user: User,
  currentOrganization: Organization | null
): Promise<ActionResult<Project>> {
  if (!user?.email) {
    return { success: false, error: 'User not authenticated' }
  }

  if (!currentOrganization) {
    return { success: false, error: 'No organization selected' }
  }

  // In a real app, this would call the API
  const newProject: Project = {
    id: `proj-${Date.now()}`,
    name: data.name,
    slug: data.slug || data.name.toLowerCase().replace(/\s+/g, '-'),
    organizationId: data.organizationId,
    description: data.description,
    status: 'active',
    environment: data.environment || 'development',
    metrics: {
      requests_today: 0,
      cost_today: 0,
      avg_latency: 0,
      error_rate: 0,
    },
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  }

  return { success: true, data: newProject }
}

/**
 * Utility function to clear all context state
 */
export interface ClearContextState {
  organizations: Organization[]
  currentOrganization: Organization | null
  currentProject: Project | null
  projects: Project[]
  error: string | null
  isLoading: boolean
}

export function getClearedContextState(): ClearContextState {
  return {
    organizations: [],
    currentOrganization: null,
    currentProject: null,
    projects: [],
    error: null,
    isLoading: false,
  }
}