/**
 * Context coordination utilities for managing cross-concern synchronization
 */

import { getOrganizationProjects } from '@/lib/api'
import { canAccessContext, getUserRole } from '@/lib/utils/context-resolver'
import { findOrganizationBySlug } from '@/lib/utils/organization-utils'
import type { Organization, Project, OrganizationRole } from '@/types/organization'
import type { User } from '@/types/auth'

/**
 * Coordination helpers that manage state synchronization
 * These helpers ensure that coupled concerns stay coordinated
 */

/**
 * Update projects list for an organization and return the projects
 * This ensures projects state stays in sync with the current organization
 */
export async function updateProjectsList(organizationId: string): Promise<Project[]> {
  return await getOrganizationProjects(organizationId)
}

/**
 * Check if user has access to a specific context
 * Centralizes access control logic
 */
export async function hasContextAccess(
  userEmail: string | undefined,
  orgSlug: string,
  projectSlug?: string
): Promise<boolean> {
  if (!userEmail) return false
  return await canAccessContext(userEmail, orgSlug, projectSlug)
}

/**
 * Get user's role in an organization
 * Centralizes role resolution logic
 */
export async function getUserContextRole(
  userEmail: string | undefined,
  orgSlug: string
): Promise<OrganizationRole | null> {
  if (!userEmail) return null
  return await getUserRole(userEmail, orgSlug)
}

/**
 * Get projects for an organization by slug
 * Provides consistent project retrieval
 */
export async function getOrganizationProjectsBySlug(orgSlug: string): Promise<Project[]> {
  const organization = await findOrganizationBySlug(orgSlug)
  if (!organization) return []
  return await getOrganizationProjects(organization.id)
}

/**
 * Validate context state consistency
 * Ensures that the context state is internally consistent
 */
export interface ContextValidationResult {
  isValid: boolean
  errors: string[]
  warnings: string[]
}

export function validateContextState(
  currentOrganization: Organization | null,
  currentProject: Project | null,
  projects: Project[]
): ContextValidationResult {
  const errors: string[] = []
  const warnings: string[] = []

  // Check if project belongs to current organization
  if (currentProject && currentOrganization) {
    if (currentProject.organizationId !== currentOrganization.id) {
      errors.push(
        `Project "${currentProject.slug}" does not belong to organization "${currentOrganization.slug}"`
      )
    }
  }

  // Check if project exists in projects list
  if (currentProject && projects.length > 0) {
    const projectExists = projects.some(p => p.id === currentProject.id)
    if (!projectExists) {
      warnings.push(
        `Current project "${currentProject.slug}" is not in the projects list`
      )
    }
  }

  // Check if we have projects but no organization
  if (projects.length > 0 && !currentOrganization) {
    warnings.push('Projects exist but no organization is selected')
  }

  return {
    isValid: errors.length === 0,
    errors,
    warnings,
  }
}

/**
 * Check if context initialization dependencies are ready
 * Prevents premature initialization
 */
export function areInitializationDependenciesReady(
  authLoading: boolean,
  isAuthenticated: boolean,
  persistenceLoaded: boolean,
  userEmail?: string
): boolean {
  return !authLoading && isAuthenticated && persistenceLoaded && !!userEmail
}

/**
 * Check if URL handling dependencies are ready
 * Prevents premature URL processing
 */
export function areURLHandlingDependenciesReady(
  isAuthenticated: boolean,
  isLoading: boolean,
  userEmail?: string
): boolean {
  return isAuthenticated && !isLoading && !!userEmail
}

/**
 * Determine if context should be cleared
 * Centralizes the logic for when to clear context
 */
export function shouldClearContext(
  isAuthenticated: boolean,
  userEmail?: string
): boolean {
  return !isAuthenticated || !userEmail
}

/**
 * Coordination state interface for tracking dependencies
 */
export interface CoordinationState {
  authReady: boolean
  persistenceReady: boolean
  urlReady: boolean
  canInitialize: boolean
  canHandleURL: boolean
  shouldClear: boolean
}

/**
 * Get coordination state summary for debugging and state management
 */
export function getCoordinationState(
  authLoading: boolean,
  isAuthenticated: boolean,
  persistenceLoaded: boolean,
  isLoading: boolean,
  userEmail?: string
): CoordinationState {
  const authReady = !authLoading && isAuthenticated && !!userEmail
  const persistenceReady = persistenceLoaded
  const urlReady = !isLoading

  return {
    authReady,
    persistenceReady,
    urlReady,
    canInitialize: areInitializationDependenciesReady(
      authLoading,
      isAuthenticated,
      persistenceLoaded,
      userEmail
    ),
    canHandleURL: areURLHandlingDependenciesReady(
      isAuthenticated,
      isLoading,
      userEmail
    ),
    shouldClear: shouldClearContext(isAuthenticated, userEmail),
  }
}