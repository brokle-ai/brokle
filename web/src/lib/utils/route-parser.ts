/**
 * Route parsing utilities - Dynamic parsing using central configuration
 * No hardcoded patterns, all driven by ROUTE_CONFIG
 */

import { ROUTE_CONFIG, RouteContext, getContextConfig } from '@/lib/config/routes'

export interface ParsedRoute {
  context: RouteContext
  slug: string
  pageType?: string
  fullPath: string
  isHomePage: boolean
  isNestedPage: boolean
  parentPage?: string
  nestedPage?: string
}

/**
 * Parse current pathname using route configuration
 * Returns null if path doesn't match any known patterns
 */
export function parseCurrentRoute(pathname: string): ParsedRoute | null {
  // Try each context pattern from configuration
  for (const [contextName, config] of Object.entries(ROUTE_CONFIG.contexts)) {
    const context = contextName as RouteContext
    
    if (matchesContextPattern(pathname, context)) {
      const slug = extractSlug(pathname, context)
      const pageInfo = extractPageInfo(pathname, context)
      
      return {
        context,
        slug,
        pageType: pageInfo.pageType,
        fullPath: pathname,
        isHomePage: !pageInfo.pageType,
        isNestedPage: pageInfo.isNested,
        parentPage: pageInfo.parentPage,
        nestedPage: pageInfo.nestedPage
      }
    }
  }
  
  return null
}

/**
 * Check if pathname matches a specific context pattern
 */
export function matchesContextPattern(pathname: string, context: RouteContext): boolean {
  const config = getContextConfig(context)
  const pattern = config.pattern
  
  // Convert Next.js pattern to regex
  // /organizations/[orgSlug] → /organizations/([^/]+)
  const regexPattern = pattern
    .replace(/\[([^\]]+)\]/g, '([^/]+)')
    .replace(/\//g, '\\/')
  
  const regex = new RegExp(`^${regexPattern}(?:/(.*))?$`)
  return regex.test(pathname)
}

/**
 * Extract slug from pathname for given context
 */
export function extractSlug(pathname: string, context: RouteContext): string {
  const config = getContextConfig(context)
  const pattern = config.pattern
  
  // Convert pattern to regex and extract slug
  const regexPattern = pattern
    .replace(/\[([^\]]+)\]/g, '([^/]+)')
    .replace(/\//g, '\\/')
  
  const regex = new RegExp(`^${regexPattern}`)
  const match = pathname.match(regex)
  
  return match?.[1] || ''
}

/**
 * Extract page information from pathname
 * Supports up to three-level nesting (e.g., settings/organization/billing)
 */
function extractPageInfo(pathname: string, context: RouteContext): {
  pageType?: string
  isNested: boolean
  parentPage?: string
  nestedPage?: string
} {
  const config = getContextConfig(context)
  const pattern = config.pattern

  // Extract everything after the base pattern
  const regexPattern = pattern
    .replace(/\[([^\]]+)\]/g, '([^/]+)')
    .replace(/\//g, '\\/')

  const regex = new RegExp(`^${regexPattern}(?:/(.*))?$`)
  const match = pathname.match(regex)
  const pagesPart = match?.[2] // Everything after slug

  if (!pagesPart) {
    return { isNested: false }
  }

  const pathSegments = pagesPart.split('/')
  const firstSegment = pathSegments[0]

  // Check for THREE-level nesting first (e.g., settings/organization/billing)
  // This must be checked before two-level to avoid partial matches
  if (pathSegments.length >= 3 && config.nested) {
    const twoLevelKey = `${pathSegments[0]}/${pathSegments[1]}`
    const nestedConfig = config.nested[twoLevelKey as keyof typeof config.nested]

    if (Array.isArray(nestedConfig) && nestedConfig.includes(pathSegments[2] as never)) {
      return {
        pageType: `${twoLevelKey}/${pathSegments[2]}`,
        isNested: true,
        parentPage: twoLevelKey,
        nestedPage: pathSegments[2]
      }
    }
  }

  // Check for TWO-level nesting (e.g., settings/api-keys)
  if (pathSegments.length >= 2 && config.nested?.[firstSegment]) {
    const parentPage = firstSegment
    const nestedPage = pathSegments[1]
    const nestedConfig = config.nested[parentPage as keyof typeof config.nested]

    // Only treat as nested if the child page is in the config
    if (Array.isArray(nestedConfig) && nestedConfig.includes(nestedPage as never)) {
      return {
        pageType: `${parentPage}/${nestedPage}`,
        isNested: true,
        parentPage,
        nestedPage
      }
    }
  }

  // Regular page (single segment)
  return {
    pageType: firstSegment,
    isNested: false
  }
}

/**
 * Check if a route context matches a given pattern
 */
export function isRouteContext(value: string): value is RouteContext {
  return value in ROUTE_CONFIG.contexts
}

/**
 * Get the base URL for a context (without any page)
 */
export function getContextBasePattern(context: RouteContext): string {
  return getContextConfig(context).pattern
}