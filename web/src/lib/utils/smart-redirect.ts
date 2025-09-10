/**
 * Smart redirect utilities v2 - Config-driven, no hardcoding
 * Complete rewrite using central configuration system
 */

import { parseCurrentRoute } from './route-parser'
import { isValidPage, shouldPreservePage, pageExistsInContext } from './route-validation'
import { RouteContext } from '@/lib/config/routes'
import { buildOrgUrl, buildProjectUrl } from './slug-utils'

export interface RedirectContext {
  currentPath: string
  targetOrgSlug?: string
  targetOrgId?: string  
  targetOrgName?: string
  targetProjectSlug?: string
  targetProjectId?: string
  targetProjectName?: string
}

// Legacy interface for backward compatibility
export type SmartRedirectParams = RedirectContext

/**
 * Determines the smart redirect URL using configuration-driven logic
 * Completely eliminates hardcoded patterns and page types
 */
export function getSmartRedirectUrl(params: SmartRedirectParams): string {
  // Parse current route using configuration
  const currentRoute = parseCurrentRoute(params.currentPath)
  const targetContext = determineTargetContext(params)
  
  // If we can't parse the current route, go to target home
  if (!currentRoute) {
    return buildTargetUrl(targetContext, params)
  }
  
  // Check if this is same-context switching (org→org or project→project)
  if (shouldPreservePage(currentRoute.context, targetContext)) {
    // Try to preserve current page in target context
    if (currentRoute.pageType && pageExistsInContext(currentRoute.pageType, targetContext)) {
      return buildTargetUrl(targetContext, params, currentRoute.pageType)
    }
  }
  
  // Cross-context switching or page doesn't exist → go to target home
  return buildTargetUrl(targetContext, params)
}

/**
 * Determine which context we're switching to based on parameters
 */
function determineTargetContext(params: SmartRedirectParams): RouteContext {
  if (params.targetProjectSlug || params.targetProjectId) {
    return 'project'
  }
  if (params.targetOrgSlug || params.targetOrgId) {
    return 'organization'
  }
  
  // Default fallback (shouldn't happen in normal usage)
  return 'organization'
}

/**
 * Build the target URL for the given context and optional page
 */
function buildTargetUrl(
  context: RouteContext, 
  params: SmartRedirectParams, 
  pageType?: string
): string {
  const baseUrl = context === 'organization' 
    ? buildOrgUrl(params.targetOrgName!, params.targetOrgId!)
    : buildProjectUrl(params.targetProjectName!, params.targetProjectId!)
  
  // Add page type if specified and valid
  if (pageType && isValidPage(pageType, context)) {
    return `${baseUrl}/${pageType}`
  }
  
  return baseUrl
}

/**
 * Legacy compatibility function
 */
export function isValidPageForContext(pageType: string, context: 'organization' | 'project'): boolean {
  return isValidPage(pageType, context)
}

/**
 * Debug helper: Explain redirect decision
 */
export function explainRedirectDecision(params: SmartRedirectParams): {
  currentRoute: ReturnType<typeof parseCurrentRoute>
  targetContext: RouteContext
  decision: string
  resultUrl: string
} {
  const currentRoute = parseCurrentRoute(params.currentPath)
  const targetContext = determineTargetContext(params)
  const resultUrl = getSmartRedirectUrl(params)
  
  let decision = 'Unknown'
  
  if (!currentRoute) {
    decision = 'Could not parse current route → go to target home'
  } else if (shouldPreservePage(currentRoute.context, targetContext)) {
    if (currentRoute.pageType && pageExistsInContext(currentRoute.pageType, targetContext)) {
      decision = `Same context switch: preserve page "${currentRoute.pageType}"`
    } else {
      decision = `Same context switch: page "${currentRoute.pageType}" doesn't exist in target → home`
    }
  } else {
    decision = `Cross-context switch: ${currentRoute.context} → ${targetContext} → go to target home`
  }
  
  return {
    currentRoute,
    targetContext,
    decision,
    resultUrl
  }
}