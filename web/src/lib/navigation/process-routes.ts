import { type Route, type ProcessedRoute, type NavigationContext, RouteSection, type RouteGroup } from './types'
import { isNavItemActive } from '../utils/navigation'
import { hasRouteAccess } from '@/hooks/rbac/route-rbac-utils'

export function processNavigation(params: {
  routes: Route[]
  context: NavigationContext
  permissions: Record<string, boolean>
  featureFlags: Record<string, boolean>
  isPermissionsLoading?: boolean
}): {
  mainNavigation: {
    grouped: Partial<Record<RouteGroup, ProcessedRoute[]>>
    ungrouped: ProcessedRoute[]
  }
  secondaryNavigation: {
    grouped: Partial<Record<RouteGroup, ProcessedRoute[]>>
    ungrouped: ProcessedRoute[]
  }
  flatNavigation: ProcessedRoute[]
} {
  const { routes, context, permissions, featureFlags, isPermissionsLoading } = params

  // Don't filter during permission load (prevents empty sidebar flicker)
  if (isPermissionsLoading) {
    return {
      mainNavigation: { grouped: {}, ungrouped: [] },
      secondaryNavigation: { grouped: {}, ungrouped: [] },
      flatNavigation: [],
    }
  }

  const processRoute = (route: Route): ProcessedRoute | null => {
    const hasOrgSlug = route.pathname.includes('[orgSlug]')
    const hasProjectSlug = route.pathname.includes('[projectSlug]')

    if (hasOrgSlug && !context.currentOrgSlug) return null
    if (hasProjectSlug && !context.currentProjectSlug) return null

    if (route.featureFlag && !featureFlags[route.featureFlag]) return null

    // TODO: This will filter properly once real RBAC is implemented
    if (route.rbacScope && !hasRouteAccess(route.rbacScope, permissions)) {
      return null
    }

    if (route.show && !route.show(context)) return null

    const url = route.pathname
      .replace('[orgSlug]', context.currentOrgSlug ?? '')
      .replace('[projectSlug]', context.currentProjectSlug ?? '')

    const isActive = isNavItemActive(context.pathname, url)

    return { ...route, url, isActive }
  }

  const processed = routes
    .map(processRoute)
    .filter((r): r is ProcessedRoute => r !== null)

  const main = processed.filter(r => r.section === RouteSection.Main)
  const secondary = processed.filter(r => r.section === RouteSection.Secondary)

  const groupRoutes = (routes: ProcessedRoute[]) => {
    const ungrouped = routes.filter(r => !r.group)
    const grouped: Partial<Record<RouteGroup, ProcessedRoute[]>> = {}

    routes.forEach(route => {
      if (route.group) {
        if (!grouped[route.group]) grouped[route.group] = []
        grouped[route.group]!.push(route)
      }
    })

    return { ungrouped, grouped }
  }

  return {
    mainNavigation: groupRoutes(main),
    secondaryNavigation: groupRoutes(secondary),
    flatNavigation: processed,
  }
}
