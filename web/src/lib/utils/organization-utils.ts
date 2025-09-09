/**
 * Organization utility functions
 * Extracted from data layer - utilities without caching
 */

import { 
  getUserOrganizations,
  getOrganizationMembers,
  getOrganizationProjects,
  getProjectMetrics
} from '@/lib/api'
import type { 
  Organization, 
  Project, 
  ProjectMetrics, 
  OrganizationRole 
} from '@/types/organization'

/**
 * Get project metrics summary for an organization
 */
export async function getProjectMetricsSummary(organizationId: string): Promise<{
  totalRequests: number
  totalCost: number
  requestsToday: number
  costToday: number
  avgLatency: number
  avgErrorRate: number
}> {
  const projects = await getOrganizationProjects(organizationId)
  
  if (projects.length === 0) {
    return {
      totalRequests: 0,
      totalCost: 0,
      requestsToday: 0,
      costToday: 0,
      avgLatency: 0,
      avgErrorRate: 0,
    }
  }

  return projects.reduce((summary, project) => ({
    totalRequests: summary.totalRequests + (project.metrics.total_requests || 0),
    totalCost: summary.totalCost + (project.metrics.total_cost || 0),
    requestsToday: summary.requestsToday + project.metrics.requests_today,
    costToday: summary.costToday + project.metrics.cost_today,
    avgLatency: (summary.avgLatency + project.metrics.avg_latency) / 2,
    avgErrorRate: (summary.avgErrorRate + project.metrics.error_rate) / 2,
  }), {
    totalRequests: 0,
    totalCost: 0,
    requestsToday: 0,
    costToday: 0,
    avgLatency: 0,
    avgErrorRate: 0,
  })
}

/**
 * Enhanced project search by slug across organizations
 */
export async function findProjectBySlug(
  projectSlug: string, 
  organizationSlugs?: string[]
): Promise<{
  project: Project
  organization: { id: string; slug: string; name: string }
} | undefined> {
  // Get user organizations
  const userOrgs = await getUserOrganizations()
  
  // Filter to specific organizations if provided
  const targetOrgs = organizationSlugs 
    ? userOrgs.filter(org => organizationSlugs.includes(org.slug))
    : userOrgs
  
  // Search across organizations
  for (const org of targetOrgs) {
    const projects = await getOrganizationProjects(org.id)
    const project = projects.find(p => p.slug === projectSlug)
    
    if (project) {
      return {
        project,
        organization: { id: org.id, slug: org.slug, name: org.name }
      }
    }
  }
  
  return undefined
}

/**
 * Get user's role in organization
 */
export async function getUserRoleInOrganization(
  userEmail: string, 
  organizationId: string
): Promise<OrganizationRole | null> {
  try {
    const members = await getOrganizationMembers(organizationId)
    const member = members.find(member => member.email === userEmail)
    return member ? member.role : null
  } catch (error) {
    console.warn('[OrganizationUtils] Failed to get user role:', error)
    return null
  }
}

/**
 * Check if user has access to organization
 */
export async function checkUserHasAccessToOrganization(
  userEmail: string, 
  organizationId: string
): Promise<boolean> {
  const role = await getUserRoleInOrganization(userEmail, organizationId)
  return role !== null
}

/**
 * Get active projects for an organization
 */
export async function getActiveProjects(organizationId: string): Promise<Project[]> {
  const projects = await getOrganizationProjects(organizationId)
  return projects.filter(project => project.status === 'active')
}

/**
 * Find organization by slug from user's organizations
 */
export async function findOrganizationBySlug(slug: string): Promise<Organization | undefined> {
  const organizations = await getUserOrganizations()
  return organizations.find(org => org.slug === slug)
}

/**
 * Find project by slug within a specific organization
 */
export async function findProjectBySlugInOrganization(
  organizationId: string, 
  slug: string
): Promise<Project | undefined> {
  const projects = await getOrganizationProjects(organizationId)
  return projects.find(project => project.slug === slug)
}