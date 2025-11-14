/**
 * Central route configuration - Single source of truth for all route patterns
 * Type-safe, maintainable, and easily extensible
 */

export const ROUTE_CONFIG = {
  contexts: {
    organization: {
      pattern: '/organizations/[orgSlug]',
      pages: ['billing', 'members', 'settings'] as const,
      preserveContext: true,
      homeRoute: '/'
    },
    project: {
      pattern: '/projects/[projectSlug]',
      pages: ['tasks', 'traces', 'settings'] as const,
      nested: {
        settings: ['integrations', 'api-keys', 'security', 'danger']
      } as const,
      preserveContext: true,
      homeRoute: '/'
    }
  }
} as const

// Auto-generated types from configuration
export type RouteContext = keyof typeof ROUTE_CONFIG.contexts
export type OrgPage = typeof ROUTE_CONFIG.contexts.organization.pages[number]  
export type ProjectPage = typeof ROUTE_CONFIG.contexts.project.pages[number]
export type ProjectSettingsPage = typeof ROUTE_CONFIG.contexts.project.nested.settings[number]

// Union types for all pages
export type OrganizationPageType = OrgPage
export type ProjectPageType = ProjectPage | `settings/${ProjectSettingsPage}`

// Route configuration type
export interface RouteConfig {
  pattern: string
  pages: readonly string[]
  nested?: Record<string, readonly string[]>
  preserveContext: boolean
  homeRoute: string
}

// Base configuration type without nested
export interface BaseRouteConfig {
  pattern: string
  pages: readonly string[]
  preserveContext: boolean
  homeRoute: string
}

// Configuration type with nested
export interface NestedRouteConfig extends BaseRouteConfig {
  nested: Record<string, readonly string[]>
}

// Helper to get configuration for a context
export function getContextConfig(context: RouteContext): RouteConfig {
  return ROUTE_CONFIG.contexts[context]
}

// Helper to get all valid pages for a context (including nested)
export function getAllPages(context: RouteContext): string[] {
  const config = ROUTE_CONFIG.contexts[context]
  const pages = [...config.pages] as string[]
  
  if (config.nested) {
    Object.entries(config.nested).forEach(([parentPage, nestedPages]) => {
      (nestedPages as readonly string[]).forEach((nestedPage: string) => {
        pages.push(`${parentPage}/${nestedPage}`)
      })
    })
  }
  
  return pages
}

// Helper to check if context supports preserving page on switch
export function shouldPreserveContext(context: RouteContext): boolean {
  return ROUTE_CONFIG.contexts[context].preserveContext
}