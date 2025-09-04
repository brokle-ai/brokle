import { api } from '@/lib/api'
import type { Organization, OrganizationMember, OrganizationRole } from '@/types/organization'

/**
 * Organization Data API Wrappers
 * These functions provide a backward-compatible interface while using the real API
 * Maintains the same function signatures as the previous mock implementation
 */

let organizationsCache: Organization[] | null = null
let cacheTimestamp: number = 0
const CACHE_TTL = 5 * 60 * 1000 // 5 minutes

/**
 * Get cached organizations or fetch from API
 */
async function getCachedOrganizations(): Promise<Organization[]> {
  const now = Date.now()
  
  // Return cache if valid and recent
  if (organizationsCache && (now - cacheTimestamp) < CACHE_TTL) {
    return organizationsCache
  }

  // Fetch fresh data from API
  try {
    organizationsCache = await api.organizations.getUserOrganizations()
    cacheTimestamp = now
    return organizationsCache
  } catch (error) {
    console.error('[Organizations Data] Failed to fetch organizations:', error)
    
    // Return empty array if API fails
    return []
  }
}

/**
 * Clear the organizations cache (useful when switching users or after updates)
 */
export function clearOrganizationsCache(): void {
  organizationsCache = null
  cacheTimestamp = 0
}

// Utility functions with API integration

export async function getOrganizationBySlug(slug: string): Promise<Organization | undefined> {
  const organizations = await getCachedOrganizations()
  return organizations.find(org => org.slug === slug)
}

export async function getOrganizationById(id: string): Promise<Organization | undefined> {
  const organizations = await getCachedOrganizations()
  const org = organizations.find(org => org.id === id)
  
  // If not in cache, try to fetch directly from API
  if (!org) {
    try {
      return await api.organizations.getOrganization(id)
    } catch (error) {
      console.warn('[Organizations Data] Failed to fetch organization by ID:', error)
      return undefined
    }
  }
  
  return org
}

export async function getUserOrganizations(userEmail: string): Promise<Organization[]> {
  // This function now gets organizations from the API directly
  // The userEmail parameter is kept for backward compatibility but not used
  // since the API uses authentication context
  return await getCachedOrganizations()
}

export async function getUserRoleInOrganization(userEmail: string, orgSlug: string): Promise<OrganizationRole | null> {
  const org = await getOrganizationBySlug(orgSlug)
  if (!org) return null
  
  // Try to get organization members if not already loaded
  if (!org.members || org.members.length === 0) {
    try {
      const members = await api.organizations.getOrganizationMembers(org.id)
      org.members = members
    } catch (error) {
      console.warn('[Organizations Data] Failed to fetch organization members:', error)
      // Assume owner role if we can't fetch members but user has access to org
      return 'owner'
    }
  }
  
  const member = org.members.find(member => member.email === userEmail)
  return member ? member.role : null
}

export async function checkUserHasAccessToOrganization(userEmail: string, orgSlug: string): Promise<boolean> {
  const role = await getUserRoleInOrganization(userEmail, orgSlug)
  return role !== null
}

/**
 * Get organization members
 */
export async function getOrganizationMembers(organizationId: string): Promise<OrganizationMember[]> {
  try {
    return await api.organizations.getOrganizationMembers(organizationId)
  } catch (error) {
    console.error('[Organizations Data] Failed to fetch organization members:', error)
    return []
  }
}

/**
 * Create a new organization
 */
export async function createOrganization(data: {
  name: string
  slug?: string
  billing_email: string
  plan?: 'free' | 'pro' | 'business' | 'enterprise'
}): Promise<Organization> {
  const newOrg = await api.organizations.createOrganization(data)
  
  // Clear cache to ensure fresh data on next fetch
  clearOrganizationsCache()
  
  return newOrg
}

/**
 * Update organization
 */
export async function updateOrganization(organizationId: string, data: Partial<{
  name: string
  billing_email: string
  subscription_plan: 'free' | 'pro' | 'business' | 'enterprise'
}>): Promise<Organization> {
  const updatedOrg = await api.organizations.updateOrganization(organizationId, data)
  
  // Clear cache to ensure fresh data on next fetch
  clearOrganizationsCache()
  
  return updatedOrg
}

/**
 * Invite user to organization
 */
export async function inviteUserToOrganization(organizationId: string, email: string, role: 'admin' | 'developer' | 'viewer'): Promise<void> {
  await api.organizations.inviteUser(organizationId, email, role)
  
  // Clear cache to ensure fresh data on next fetch
  clearOrganizationsCache()
}

/**
 * Remove user from organization
 */
export async function removeUserFromOrganization(organizationId: string, userId: string): Promise<void> {
  await api.organizations.removeUser(organizationId, userId)
  
  // Clear cache to ensure fresh data on next fetch
  clearOrganizationsCache()
}

/**
 * Update user role in organization
 */
export async function updateUserRoleInOrganization(organizationId: string, userId: string, role: 'admin' | 'developer' | 'viewer'): Promise<void> {
  await api.organizations.updateUserRole(organizationId, userId, role)
  
  // Clear cache to ensure fresh data on next fetch
  clearOrganizationsCache()
}