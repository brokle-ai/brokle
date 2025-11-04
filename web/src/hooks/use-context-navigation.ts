/**
 * Context navigation convenience hook
 * 
 * Provides a clean interface for components that need to handle navigation
 * between organizations and projects with URL synchronization.
 */

import { useRouter, usePathname } from 'next/navigation'
import { useOrganization } from '@/context/org-context'
import { parsePathContext } from '@/lib/utils/slug-utils'
import type { Organization, Project } from '@/types/organization'

export interface ContextNavigationHooks {
  // Current context
  currentOrganization: Organization | null
  currentProject: Project | null
  isLoading: boolean

  // Navigation actions
  navigateToOrganization: (orgSlug: string) => Promise<void>
  navigateToProject: (orgSlug: string, projectSlug: string) => Promise<void>
  navigateToCurrentProject: (projectSlug: string) => Promise<void>
  
  // URL utilities
  getCurrentPath: () => string
  getOrganizationUrl: (orgSlug: string) => string
  getProjectUrl: (orgSlug: string, projectSlug: string) => string
  
  // Context checking
  isOnOrganizationPage: boolean
  isOnProjectPage: boolean
  urlMatchesContext: boolean
  
  // URL parsing
  urlContext: {
    orgSlug: string | null
    projectSlug: string | null
  }
}

/**
 * Hook that provides navigation utilities for organization/project context
 * 
 * Perfect for:
 * - Navigation components
 * - Breadcrumb components
 * - URL-aware context switchers
 * - Components that need to navigate programmatically
 * 
 * @example
 * ```tsx
 * function ContextBreadcrumbs() {
 *   const { 
 *     currentOrganization,
 *     currentProject,
 *     navigateToOrganization,
 *     isOnProjectPage
 *   } = useContextNavigation()
 *   
 *   return (
 *     <Breadcrumbs>
 *       <BreadcrumbItem 
 *         onClick={() => navigateToOrganization(currentOrganization.slug)}
 *       >
 *         {currentOrganization.name}
 *       </BreadcrumbItem>
 *       
 *       {isOnProjectPage && currentProject && (
 *         <BreadcrumbItem>
 *           {currentProject.name}
 *         </BreadcrumbItem>
 *       )}
 *     </Breadcrumbs>
 *   )
 * }
 * ```
 */
export function useContextNavigation(): ContextNavigationHooks {
  const context = useOrganization()
  const router = useRouter()
  const pathname = usePathname()
  const urlContext = parsePathContext(pathname)

  const navigateToOrganization = async (orgSlug: string) => {
    await context.switchOrganization(orgSlug)
  }

  const navigateToProject = async (orgSlug: string, projectSlug: string) => {
    // Switch org first if needed
    if (!context.currentOrganization || context.currentOrganization.slug !== orgSlug) {
      await context.switchOrganization(orgSlug)
    }
    // Then switch project
    await context.switchProject(projectSlug)
  }

  const navigateToCurrentProject = async (projectSlug: string) => {
    if (!context.currentOrganization) {
      throw new Error('No organization selected')
    }
    await context.switchProject(projectSlug)
  }

  const getOrganizationUrl = (orgSlug: string) => `/${orgSlug}`
  const getProjectUrl = (orgSlug: string, projectSlug: string) => `/${orgSlug}/${projectSlug}`

  // Check if current URL context matches the current context state
  const urlMatchesContext = 
    urlContext.orgSlug === context.currentOrganization?.slug &&
    urlContext.projectSlug === context.currentProject?.slug

  return {
    // Current context
    currentOrganization: context.currentOrganization,
    currentProject: context.currentProject,
    isLoading: context.isLoading,

    // Navigation actions
    navigateToOrganization,
    navigateToProject,
    navigateToCurrentProject,

    // URL utilities
    getCurrentPath: () => pathname,
    getOrganizationUrl,
    getProjectUrl,

    // Context checking
    isOnOrganizationPage: !!urlContext.orgSlug && !urlContext.projectSlug,
    isOnProjectPage: !!urlContext.orgSlug && !!urlContext.projectSlug,
    urlMatchesContext,

    // URL parsing
    urlContext: {
      orgSlug: urlContext.orgSlug || null,
      projectSlug: urlContext.projectSlug || null,
    },
  }
}

/**
 * Hook for breadcrumb navigation
 * 
 * Provides utilities specifically for building breadcrumb navigation
 * 
 * @example
 * ```tsx
 * function NavigationBreadcrumbs() {
 *   const { breadcrumbs, navigateToBreadcrumb } = useBreadcrumbNavigation()
 *   
 *   return (
 *     <nav>
 *       {breadcrumbs.map((crumb, index) => (
 *         <span key={index}>
 *           {crumb.isClickable ? (
 *             <button onClick={() => navigateToBreadcrumb(crumb)}>
 *               {crumb.label}
 *             </button>
 *           ) : (
 *             <span>{crumb.label}</span>
 *           )}
 *           {index < breadcrumbs.length - 1 && <span> / </span>}
 *         </span>
 *       ))}
 *     </nav>
 *   )
 * }
 * ```
 */
export function useBreadcrumbNavigation() {
  const navigation = useContextNavigation()

  interface Breadcrumb {
    label: string
    url: string
    isClickable: boolean
    type: 'organization' | 'project'
    slug: string
  }

  const breadcrumbs: Breadcrumb[] = []

  // Add organization breadcrumb if we have one
  if (navigation.currentOrganization) {
    breadcrumbs.push({
      label: navigation.currentOrganization.name,
      url: navigation.getOrganizationUrl(navigation.currentOrganization.slug),
      isClickable: navigation.isOnProjectPage, // Can click if we're currently on project page
      type: 'organization',
      slug: navigation.currentOrganization.slug,
    })
  }

  // Add project breadcrumb if we have one
  if (navigation.currentProject && navigation.currentOrganization) {
    breadcrumbs.push({
      label: navigation.currentProject.name,
      url: navigation.getProjectUrl(
        navigation.currentOrganization.slug, 
        navigation.currentProject.slug
      ),
      isClickable: false, // Current page, not clickable
      type: 'project',
      slug: navigation.currentProject.slug,
    })
  }

  const navigateToBreadcrumb = async (breadcrumb: Breadcrumb) => {
    if (!breadcrumb.isClickable) return

    if (breadcrumb.type === 'organization') {
      await navigation.navigateToOrganization(breadcrumb.slug)
    }
    // Projects are never clickable in breadcrumbs (they're the current page)
  }

  return {
    breadcrumbs,
    navigateToBreadcrumb,
    hasBreadcrumbs: breadcrumbs.length > 0,
    hasMultipleLevels: breadcrumbs.length > 1,
  }
}

/**
 * Hook for context-aware routing
 * 
 * Provides utilities for building context-aware routes and handling navigation
 * 
 * @example
 * ```tsx
 * function ContextAwareLink({ href, children }) {
 *   const { buildContextAwareUrl, isValidContextRoute } = useContextAwareRouting()
 *   
 *   const fullUrl = buildContextAwareUrl(href)
 *   
 *   if (!isValidContextRoute(fullUrl)) {
 *     return <span className="disabled">{children}</span>
 *   }
 *   
 *   return <Link href={fullUrl}>{children}</Link>
 * }
 * ```
 */
export function useContextAwareRouting() {
  const navigation = useContextNavigation()
  const router = useRouter()

  /**
   * Build a context-aware URL by prepending current org/project context
   */
  const buildContextAwareUrl = (relativePath: string): string => {
    if (relativePath.startsWith('/')) {
      // Absolute path, return as-is
      return relativePath
    }

    // Build context-aware path
    if (navigation.currentProject && navigation.currentOrganization) {
      return `/${navigation.currentOrganization.slug}/${navigation.currentProject.slug}/${relativePath}`
    } else if (navigation.currentOrganization) {
      return `/${navigation.currentOrganization.slug}/${relativePath}`
    } else {
      return `/${relativePath}`
    }
  }

  /**
   * Check if a route is accessible in the current context
   * TODO: Implement with backend permission checking
   */
  const isValidContextRoute = (url: string): boolean => {
    const { orgSlug, projectSlug } = parsePathContext(url)
    
    if (!orgSlug) return true // Root routes are always valid
    
    // TODO: Replace with backend permission checking
    // For now, assume all routes are accessible if user is authenticated
    return true
  }

  /**
   * Navigate to a relative path within the current context
   */
  const navigateWithinContext = (relativePath: string) => {
    const fullUrl = buildContextAwareUrl(relativePath)
    router.push(fullUrl)
  }

  return {
    buildContextAwareUrl,
    isValidContextRoute,
    navigateWithinContext,
    
    // Context info for building routes
    currentOrgSlug: navigation.currentOrganization?.slug || null,
    currentProjectSlug: navigation.currentProject?.slug || null,
    hasFullContext: !!(navigation.currentOrganization && navigation.currentProject),
    hasOrgContext: !!navigation.currentOrganization,
  }
}