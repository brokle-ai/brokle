'use client'

import { extractIdFromCompositeSlug, isValidCompositeSlug } from '@/lib/utils/slug-utils'
import type { BrokleAPIError } from '@/lib/api/core/types'

/**
 * URL-Based Context Manager - Clean, multi-tab friendly context management
 * Derives organization and project IDs from URL pathname using smart caching
 */

export interface ContextHeaders {
  'X-Org-ID'?: string
  'X-Project-ID'?: string
  'X-Environment-ID'?: string
}

export interface ContextOptions {
  includeOrgContext?: boolean
  includeProjectContext?: boolean
  includeEnvironmentContext?: boolean
  customOrgId?: string
  customProjectId?: string
  customEnvironmentId?: string
}

interface CachedContext {
  orgId: string
  orgSlug: string
  projectId?: string
  projectSlug?: string
  environmentId?: string
  timestamp: number
}

export class URLContextManager {
  private cache = new Map<string, CachedContext>()
  private readonly CACHE_TTL = 10 * 60 * 1000 // 10 minutes
  private readonly DEBUG = process.env.NODE_ENV === 'development'
  
  /**
   * Get headers for API requests based on URL pathname
   * This is the main method that replaces the old context manager
   */
  async getHeadersFromURL(pathname: string, options: ContextOptions = {}): Promise<ContextHeaders> {
    try {
      // Parse URL to extract composite slugs
      const pathSegments = pathname.split('/').filter(Boolean)
      let orgCompositeSlug: string | null = null
      let projectCompositeSlug: string | null = null
      
      // Look for organization composite slug (pattern: /organizations/[orgSlug])
      const orgIndex = pathSegments.indexOf('organizations')
      if (orgIndex !== -1 && pathSegments[orgIndex + 1]) {
        orgCompositeSlug = pathSegments[orgIndex + 1]
      }
      
      // Look for project composite slug (pattern: /projects/[projectSlug])
      const projectIndex = pathSegments.indexOf('projects')
      if (projectIndex !== -1 && pathSegments[projectIndex + 1]) {
        projectCompositeSlug = pathSegments[projectIndex + 1]
      }
      
      if (!orgCompositeSlug) {
        if (this.DEBUG) console.debug('[URLContextManager] No organization composite slug in pathname:', pathname)
        return {} // No org in URL, no headers needed
      }
      
      if (this.DEBUG) console.debug(`[URLContextManager] Getting headers for pathname: ${pathname}, orgSlug: ${orgCompositeSlug}, projectSlug: ${projectCompositeSlug}`)
      
      // Get or resolve context using composite slugs
      const context = await this.getResolvedContext(pathname, orgCompositeSlug, projectCompositeSlug)
      
      if (!context) {
        console.warn(`[URLContextManager] Could not resolve context for pathname: ${pathname}`)
        return {}
      }
      
      // Build headers based on options
      const headers = this.buildHeaders(context, options)
      
      if (this.DEBUG) console.debug('[URLContextManager] Generated headers:', headers)
      return headers
      
    } catch (error) {
      console.error('[URLContextManager] Failed to get headers for:', pathname, error)
      return {}
    }
  }
  
  /**
   * Get cached context or resolve from API
   */
  private async getResolvedContext(
    pathname: string, 
    orgCompositeSlug: string, 
    projectCompositeSlug?: string
  ): Promise<CachedContext | null> {
    // Check cache first
    const cached = this.getCachedContext(pathname)
    if (cached && this.isCacheValid(cached)) {
      return cached
    }
    
    // Resolve composite slugs to IDs via direct extraction
    const resolved = await this.resolveContext(orgCompositeSlug, projectCompositeSlug)
    
    if (resolved) {
      // Cache the resolved context
      this.cacheContext(pathname, resolved)
    }
    
    return resolved
  }
  
  /**
   * Resolve organization and project composite slugs to IDs using direct ID extraction
   * Uses composite slug format with embedded IDs for direct lookup
   */
  private async resolveContext(
    orgCompositeSlug: string, 
    projectCompositeSlug?: string
  ): Promise<CachedContext | null> {
    try {
      // Import specific functions to avoid circular dependencies
      const { getOrganizationById, getProjectById } = await import('@/lib/api')
      
      // Validate and extract organization ID from composite slug
      if (!isValidCompositeSlug(orgCompositeSlug)) {
        console.warn(`[URLContextManager] Invalid organization composite slug format: ${orgCompositeSlug}`)
        return null
      }
      
      const orgId = extractIdFromCompositeSlug(orgCompositeSlug)
      if (this.DEBUG) console.debug(`[URLContextManager] Resolving organization ID: ${orgId}`)
      
      const org = await getOrganizationById(orgId)
      
      let project = null
      if (projectCompositeSlug) {
        try {
          // Validate and extract project ID from composite slug
          if (!isValidCompositeSlug(projectCompositeSlug)) {
            console.warn(`[URLContextManager] Invalid project composite slug format: ${projectCompositeSlug}`)
          } else {
            const projectId = extractIdFromCompositeSlug(projectCompositeSlug)
            if (this.DEBUG) console.debug(`[URLContextManager] Resolving project ID: ${projectId}`)
            
            project = await getProjectById(projectId)
          }
        } catch (error) {
          // Handle different error types appropriately
          if (error && typeof error === 'object' && 'statusCode' in error) {
            const apiError = error as any
            if (apiError.statusCode === 404) {
              console.warn(`[URLContextManager] Project '${projectCompositeSlug}' not found`)
            } else if (apiError.statusCode === 403) {
              console.warn(`[URLContextManager] Access denied to project '${projectCompositeSlug}'`)
            } else {
              console.error(`[URLContextManager] API error resolving project '${projectCompositeSlug}':`, apiError.message)
            }
          } else {
            console.warn('[URLContextManager] Unexpected error resolving project:', projectCompositeSlug, error)
          }
          // Don't fail the entire context resolution if project is not found
          // Return organization context without project
        }
      }
      
      const resolvedContext = {
        orgId: org.id,
        orgSlug: orgCompositeSlug, // Store composite slug from URL
        projectId: project?.id,
        projectSlug: projectCompositeSlug, // Store composite slug from URL
        timestamp: Date.now(),
      }

      if (this.DEBUG) console.debug('[URLContextManager] Successfully resolved context:', {
        orgSlug: resolvedContext.orgSlug,
        orgId: resolvedContext.orgId,
        projectSlug: resolvedContext.projectSlug,
        projectId: resolvedContext.projectId
      })

      return resolvedContext
      
    } catch (error) {
      // Handle organization resolution errors
      if (error && typeof error === 'object' && 'statusCode' in error) {
        const apiError = error as any
        if (apiError.statusCode === 404) {
          console.warn(`[URLContextManager] Organization '${orgCompositeSlug}' not found`)
        } else if (apiError.statusCode === 403) {
          console.warn(`[URLContextManager] Access denied to organization '${orgCompositeSlug}'`)
        } else {
          console.error(`[URLContextManager] API error resolving organization '${orgCompositeSlug}':`, apiError.message)
        }
      } else {
        console.error('[URLContextManager] Unexpected error resolving context:', error)
      }
      
      return null
    }
  }
  
  /**
   * Build headers object based on resolved context and options
   */
  private buildHeaders(context: CachedContext, options: ContextOptions): ContextHeaders {
    const headers: ContextHeaders = {}
    
    // Add organization header if requested
    if (options.includeOrgContext) {
      headers['X-Org-ID'] = options.customOrgId || context.orgId
    }
    
    // Add project header if requested and available
    if (options.includeProjectContext && (context.projectId || options.customProjectId)) {
      headers['X-Project-ID'] = options.customProjectId || context.projectId
    }
    
    // Add environment header if requested and available
    if (options.includeEnvironmentContext && (context.environmentId || options.customEnvironmentId)) {
      headers['X-Environment-ID'] = options.customEnvironmentId || context.environmentId
    }
    
    return headers
  }
  
  /**
   * Get cached context for pathname
   */
  private getCachedContext(pathname: string): CachedContext | null {
    return this.cache.get(pathname) || null
  }
  
  /**
   * Check if cached context is still valid
   */
  private isCacheValid(context: CachedContext): boolean {
    return (Date.now() - context.timestamp) < this.CACHE_TTL
  }
  
  /**
   * Cache resolved context
   */
  private cacheContext(pathname: string, context: CachedContext): void {
    this.cache.set(pathname, context)
    
    // Clean up old cache entries periodically
    this.cleanupCache()
  }
  
  /**
   * Remove expired cache entries
   */
  private cleanupCache(): void {
    const now = Date.now()
    const toDelete: string[] = []
    
    for (const [key, context] of this.cache.entries()) {
      if ((now - context.timestamp) > this.CACHE_TTL) {
        toDelete.push(key)
      }
    }
    
    toDelete.forEach(key => this.cache.delete(key))
  }
  
  /**
   * Clear all cached contexts
   */
  public clearCache(): void {
    this.cache.clear()
  }
  
  /**
   * Get current context info for debugging
   */
  public getCurrentContext(pathname: string): CachedContext | null {
    return this.getCachedContext(pathname)
  }
  
  /**
   * Preload context for a URL (useful for navigation)
   */
  async preloadContext(pathname: string): Promise<void> {
    // Parse URL to extract composite slugs
    const pathSegments = pathname.split('/').filter(Boolean)
    let orgCompositeSlug: string | null = null
    let projectCompositeSlug: string | null = null
    
    // Look for organization composite slug
    const orgIndex = pathSegments.indexOf('organizations')
    if (orgIndex !== -1 && pathSegments[orgIndex + 1]) {
      orgCompositeSlug = pathSegments[orgIndex + 1]
    }
    
    // Look for project composite slug
    const projectIndex = pathSegments.indexOf('projects')
    if (projectIndex !== -1 && pathSegments[projectIndex + 1]) {
      projectCompositeSlug = pathSegments[projectIndex + 1]
    }
    
    if (orgCompositeSlug) {
      await this.getResolvedContext(pathname, orgCompositeSlug, projectCompositeSlug)
    }
  }
  
  /**
   * Debug helper - log current cache state
   */
  public debug(): void {
    if (process.env.NODE_ENV !== 'development') return
    
    console.group('🌐 URLContextManager Debug')
    console.log('Cache size:', this.cache.size)
    console.log('Cached paths:', Array.from(this.cache.keys()))
    for (const [path, context] of this.cache.entries()) {
      const age = Date.now() - context.timestamp
      const valid = this.isCacheValid(context)
      console.log(`${path}:`, {
        org: `${context.orgSlug} (${context.orgId})`,
        project: context.projectSlug ? `${context.projectSlug} (${context.projectId})` : 'None',
        age: `${Math.round(age / 1000)}s`,
        valid
      })
    }
    console.groupEnd()
  }
}

// Singleton instance
export const urlContextManager = new URLContextManager()

// Convenience helper functions
export function getAPIHeaders(pathname: string, options?: ContextOptions): Promise<ContextHeaders> {
  return urlContextManager.getHeadersFromURL(pathname, options)
}

export function getCurrentContextHeaders(options?: ContextOptions): Promise<ContextHeaders> {
  if (typeof window === 'undefined') {
    return Promise.resolve({})
  }
  return urlContextManager.getHeadersFromURL(window.location.pathname, options)
}