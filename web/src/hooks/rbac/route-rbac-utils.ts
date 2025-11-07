import { useMemo } from 'react'
import { type Scope } from './use-has-access'
import { type Route } from '@/lib/navigation/types'

export function hasRouteAccess(
  requiredScopes: Scope | Scope[] | undefined,
  permissionMap: Record<string, boolean>,
): boolean {
  if (!requiredScopes) return true
  const scopes = Array.isArray(requiredScopes) ? requiredScopes : [requiredScopes]
  return scopes.some(scope => permissionMap[scope] === true)
}

export function extractRequiredScopes(routes: Route[]): Scope[] {
  const scopes = new Set<Scope>()
  routes.forEach(route => {
    if (route.rbacScope) {
      if (Array.isArray(route.rbacScope)) {
        route.rbacScope.forEach(s => scopes.add(s))
      } else {
        scopes.add(route.rbacScope)
      }
    }
  })
  return Array.from(scopes)
}

/**
 * RBAC Permission Hook - Currently Stubbed
 *
 * TODO: BLOCKER - Replace with real RBAC once permission data is available
 *
 * This hook currently returns all permissions as `true` for UI development.
 * Once RBAC is implemented, replace the stub with one of these approaches:
 *
 * Option 1: API-based with useQueries
 * ```typescript
 * const queries = useQueries({
 *   queries: scopes.map(scope => ({
 *     queryKey: ['rbac', 'permission', scope, projectId],
 *     queryFn: async () => {
 *       const response = await fetch('/api/v1/rbac/check', {
 *         method: 'POST',
 *         body: JSON.stringify({ scope, projectId }),
 *       })
 *       const data = await response.json()
 *       return data.hasAccess === true
 *     },
 *   })),
 * })
 * const isLoading = queries.some(q => q.isLoading || q.isPending || q.isFetching)
 * ```
 *
 * Option 2: Session-based (if permissions are preloaded)
 * ```typescript
 * const session = useSession()
 * const isLoading = session.status === 'loading'
 * const permissions = useMemo(() => {
 *   const map: Record<string, boolean> = {}
 *   scopes.forEach(scope => {
 *     map[scope] = session?.user?.permissions?.includes(scope) ?? false
 *   })
 *   return map
 * }, [scopes, session])
 * ```
 */
export function useRoutePermissions(
  scopes: Scope[],
  _projectId?: string | null
): {
  permissions: Record<string, boolean>
  isLoading: boolean
} {
  const permissions = useMemo(() => {
    const map: Record<string, boolean> = {}

    // ⚠️ STUB: All permissions return true for UI development
    // Replace this block once RBAC API/session data is available
    scopes.forEach(scope => {
      map[scope] = true  // Always grants access
    })

    return map
  }, [scopes])

  return {
    permissions,
    isLoading: false,  // No loading since not fetching
  }
}
