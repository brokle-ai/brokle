import { api } from '@/lib/api'
import { getOrganizationBySlug } from './organizations'
import type { Project, ProjectMetrics } from '@/types/organization'

/**
 * Projects Data API Wrappers
 * These functions provide a backward-compatible interface while using the real API
 * Maintains the same function signatures as the previous mock implementation
 */

const projectsCache: Map<string, Project[]> = new Map() // Keyed by organization ID
const cacheTimestamps: Map<string, number> = new Map() // Cache timestamps per organization
const CACHE_TTL = 5 * 60 * 1000 // 5 minutes

/**
 * Get cached projects for organization or fetch from API
 */
async function getCachedProjectsForOrganization(organizationId: string): Promise<Project[]> {
  const now = Date.now()
  const lastFetch = cacheTimestamps.get(organizationId) || 0
  
  // Return cache if valid and recent
  if (projectsCache.has(organizationId) && (now - lastFetch) < CACHE_TTL) {
    return projectsCache.get(organizationId)!
  }

  // Fetch fresh data from API
  try {
    const projects = await api.organizations.getOrganizationProjects(organizationId)
    projectsCache.set(organizationId, projects)
    cacheTimestamps.set(organizationId, now)
    return projects
  } catch (error) {
    console.error('[Projects Data] Failed to fetch projects for organization:', error)
    return []
  }
}

/**
 * Clear the projects cache (useful when switching users or after updates)
 */
export function clearProjectsCache(organizationId?: string): void {
  if (organizationId) {
    projectsCache.delete(organizationId)
    cacheTimestamps.delete(organizationId)
  } else {
    projectsCache.clear()
    cacheTimestamps.clear()
  }
}

// Utility functions with API integration

export async function getProjectBySlug(organizationId: string, slug: string): Promise<Project | undefined> {
  const projects = await getCachedProjectsForOrganization(organizationId)
  return projects.find(project => project.slug === slug)
}

export async function getProjectById(id: string): Promise<Project | undefined> {
  // For this function, we need to search across all cached organizations
  // or make a direct API call if we have the organization context
  
  // First, check all cached projects
  for (const [orgId, projects] of projectsCache.entries()) {
    const project = projects.find(p => p.id === id)
    if (project) return project
  }

  // If not found in cache and we don't know the organization,
  // we'd need to search or make assumptions
  // For now, return undefined - this could be enhanced with a project search API
  console.warn('[Projects Data] Project not found in cache, organization context needed')
  return undefined
}

export async function getProjectsByOrganization(organizationId: string): Promise<Project[]> {
  return await getCachedProjectsForOrganization(organizationId)
}

export async function getProjectsByOrganizationSlug(orgSlug: string): Promise<Project[]> {
  // Get organization by slug, then fetch its projects
  const org = await getOrganizationBySlug(orgSlug)
  if (!org) {
    console.warn(`[Projects Data] Organization '${orgSlug}' not found`)
    return []
  }
  
  return await getProjectsByOrganization(org.id)
}

export async function getActiveProjects(organizationId: string): Promise<Project[]> {
  const projects = await getProjectsByOrganization(organizationId)
  return projects.filter(project => project.status === 'active')
}

export async function getProjectMetricsSummary(organizationId: string): Promise<{
  totalRequests: number
  totalCost: number
  requestsToday: number
  costToday: number
  avgLatency: number
  avgErrorRate: number
}> {
  const projects = await getProjectsByOrganization(organizationId)
  
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
 * Get project metrics from API
 */
export async function getProjectMetrics(organizationId: string, projectId: string, environmentId?: string): Promise<ProjectMetrics> {
  try {
    const apiMetrics = await api.organizations.getProjectMetrics(organizationId, projectId, environmentId)
    
    // Map API response to frontend format
    return {
      requests_today: apiMetrics.requests_today,
      cost_today: apiMetrics.cost_today,
      avg_latency: apiMetrics.avg_latency_ms,
      error_rate: apiMetrics.error_rate,
      total_requests: apiMetrics.total_requests,
      total_cost: apiMetrics.total_cost,
    }
  } catch (error) {
    console.error('[Projects Data] Failed to fetch project metrics:', error)
    
    // Return default metrics on error
    return {
      requests_today: 0,
      cost_today: 0,
      avg_latency: 0,
      error_rate: 0,
      total_requests: 0,
      total_cost: 0,
    }
  }
}

/**
 * Create a new project
 */
export async function createProject(organizationId: string, data: {
  name: string
  slug?: string
  description?: string
}): Promise<Project> {
  const newProject = await api.organizations.createProject(organizationId, data)
  
  // Clear cache to ensure fresh data on next fetch
  clearProjectsCache(organizationId)
  
  return newProject
}

/**
 * Update project
 */
export async function updateProject(organizationId: string, projectId: string, data: Partial<{
  name: string
  description: string
}>): Promise<Project> {
  const updatedProject = await api.organizations.updateProject(organizationId, projectId, data)
  
  // Clear cache to ensure fresh data on next fetch
  clearProjectsCache(organizationId)
  
  return updatedProject
}

/**
 * Delete project
 */
export async function deleteProject(organizationId: string, projectId: string): Promise<void> {
  await api.organizations.deleteProject(organizationId, projectId)
  
  // Clear cache to ensure fresh data on next fetch
  clearProjectsCache(organizationId)
}

/**
 * Enhanced project search by slug across organizations
 */
export async function findProjectBySlug(projectSlug: string, organizationSlugs?: string[]): Promise<{
  project: Project
  organization: { id: string; slug: string; name: string }
} | undefined> {
  // If specific organization slugs provided, search only those
  if (organizationSlugs && organizationSlugs.length > 0) {
    for (const orgSlug of organizationSlugs) {
      const projects = await getProjectsByOrganizationSlug(orgSlug)
      const project = projects.find(p => p.slug === projectSlug)
      if (project) {
        const org = await getOrganizationBySlug(orgSlug)
        if (org) {
          return {
            project,
            organization: { id: org.id, slug: org.slug, name: org.name }
          }
        }
      }
    }
  }

  return undefined
}