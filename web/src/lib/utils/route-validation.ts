/**
 * Route validation utilities - Config-driven validation helpers
 * All validation logic driven by central configuration
 */

import { RouteContext, getContextConfig, getAllPages, ROUTE_CONFIG } from '@/lib/config/routes'

/**
 * Check if a page type is valid for the given context
 */
export function isValidPage(pageType: string | undefined, context: RouteContext): boolean {
  if (!pageType) return true // Home page is always valid
  
  const validPages = getAllPages(context)
  return validPages.includes(pageType)
}

/**
 * Get all valid pages for a context (including nested)
 */
export function getValidPages(context: RouteContext): string[] {
  return getAllPages(context)
}

/**
 * Check if a page is a nested page (e.g., "settings/integrations")
 */
export function isNestedPage(pageType: string | undefined, context: RouteContext): boolean {
  if (!pageType) return false
  
  const config = getContextConfig(context)
  if (!config.nested) return false
  
  // Check if pageType follows the pattern "parent/child"
  const parts = pageType.split('/')
  if (parts.length !== 2) return false
  
  const [parentPage, childPage] = parts
  return config.nested[parentPage]?.includes(childPage as any) || false
}

/**
 * Get the parent page for a nested page
 */
export function getParentPage(pageType: string): string | null {
  const parts = pageType.split('/')
  return parts.length > 1 ? parts[0] : null
}

/**
 * Get the nested page name for a nested page
 */
export function getNestedPageName(pageType: string): string | null {
  const parts = pageType.split('/')
  return parts.length > 1 ? parts[1] : null
}

/**
 * Check if two contexts are the same (same-context switching)
 */
export function isSameContext(currentContext: RouteContext | undefined, targetContext: RouteContext): boolean {
  return currentContext === targetContext
}

/**
 * Check if context switching should preserve the current page
 */
export function shouldPreservePage(
  currentContext: RouteContext | undefined, 
  targetContext: RouteContext
): boolean {
  if (!currentContext) return false
  
  const currentConfig = getContextConfig(currentContext)
  const targetConfig = getContextConfig(targetContext)
  
  // Both contexts must support page preservation
  return currentConfig.preserveContext && targetConfig.preserveContext && isSameContext(currentContext, targetContext)
}

/**
 * Check if a page exists in the target context
 */
export function pageExistsInContext(pageType: string | undefined, context: RouteContext): boolean {
  return isValidPage(pageType, context)
}

/**
 * Get the home route for a context
 */
export function getHomeRoute(context: RouteContext): string {
  return getContextConfig(context).homeRoute
}

/**
 * Validate that a context exists in our configuration
 */
export function isValidContext(context: string): context is RouteContext {
  return context in ROUTE_CONFIG.contexts
}

/**
 * Get all available contexts
 */
export function getAllContexts(): RouteContext[] {
  return Object.keys(ROUTE_CONFIG.contexts) as RouteContext[]
}

/**
 * Debug helper: Get configuration summary for a context
 */
export function getContextSummary(context: RouteContext) {
  const config = getContextConfig(context)
  const allPages = getAllPages(context)
  
  return {
    context,
    pattern: config.pattern,
    totalPages: allPages.length,
    regularPages: config.pages.length,
    nestedPages: allPages.length - config.pages.length,
    preserveContext: config.preserveContext,
    pages: allPages
  }
}