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
      pages: ['traces', 'prompts', 'playground', 'settings', 'dashboards',
              'datasets', 'evaluators', 'experiments', 'scores',
              'annotation-queues', 'sessions', 'spans'] as const,
      nested: {
        // PROJECT settings
        settings: [
          'api-keys',
          'score-configs',
          'danger',
          'integrations',
          'security',
          // ACCOUNT settings
          'profile',
          'account',
          'appearance',
          'notifications',
          'display',
          // ORGANIZATION settings parent
          'organization',
        ],
        // Two-level nesting for organization settings under project
        'settings/organization': [
          'members',
          'ai-providers',
          'billing',
          'danger',
        ],
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
export type ProjectOrgSettingsPage = typeof ROUTE_CONFIG.contexts.project.nested['settings/organization'][number]

// Union types for all pages
export type OrganizationPageType = OrgPage
export type ProjectPageType = ProjectPage | `settings/${ProjectSettingsPage}` | `settings/organization/${ProjectOrgSettingsPage}`

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

  if ('nested' in config && config.nested) {
    Object.entries(config.nested).forEach(([parentPage, nestedPages]) => {
      // Add multi-level parent keys (e.g., 'settings/organization')
      if (parentPage.includes('/')) {
        pages.push(parentPage)
      }

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