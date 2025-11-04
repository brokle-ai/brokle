/**
 * useHasAccess Hook - Type-Safe Scope-Based Authorization
 *
 * Checks if the current user has a specific scope in the current context.
 *
 * Behavior:
 * - projectId provided: checks project-level access (org + project scopes)
 * - projectId absent: checks org-level access only (org scopes)
 * - Returns false while loading (safe default - no access until verified)
 * - Returns false if user not authenticated
 * - Returns false if organization not selected
 * - Caches results per (userId, orgId, projectId, scope) for performance
 *
 * Examples:
 * ```typescript
 * // Organization-level scope (no projectId needed)
 * const canInviteMembers = useHasAccess({ scope: "members:invite" })
 * <Button disabled={!canInviteMembers}>Invite Member</Button>
 *
 * // Project-level scope (requires projectId)
 * const canDeleteTraces = useHasAccess({ scope: "traces:delete", projectId })
 * <Button disabled={!canDeleteTraces}>Delete Trace</Button>
 * ```
 *
 * Truth Table:
 * | User | OrgID | ProjectID | Scope Level | Result |
 * |------|-------|-----------|-------------|--------|
 * | null | -     | -         | any         | false  |
 * | ✓    | null  | -         | any         | false  |
 * | ✓    | ✓     | null      | org         | check  |
 * | ✓    | ✓     | null      | project     | false  |
 * | ✓    | ✓     | ✓         | org         | check  |
 * | ✓    | ✓     | ✓         | project     | check  |
 */

import { useQuery } from '@tanstack/react-query'
import { checkUserScopes } from '@/lib/api/services/rbac'
import { useAuth } from '@/hooks/auth/use-auth'
import { useWorkspace } from '@/context/workspace-context'

// ========================================
// Type Definitions
// ========================================

/**
 * All available scopes in the system
 * This is kept in sync with backend permissions via code generation
 */
export type Scope =
  // ========================================
  // ORGANIZATION-LEVEL SCOPES (40 scopes)
  // ========================================

  // Organization Management (4)
  | 'organizations:read'
  | 'organizations:write'
  | 'organizations:delete'
  | 'organizations:admin'

  // Members Management (5)
  | 'members:read'
  | 'members:invite'
  | 'members:update'
  | 'members:remove'
  | 'members:suspend'

  // Billing Management (4)
  | 'billing:read'
  | 'billing:manage'
  | 'billing:export'
  | 'billing:admin'

  // Settings Management (6)
  | 'settings:read'
  | 'settings:write'
  | 'settings:export'
  | 'settings:import'
  | 'settings:security'
  | 'settings:admin'

  // Roles & Permissions (5)
  | 'roles:read'
  | 'roles:write'
  | 'roles:delete'
  | 'roles:assign'
  | 'permissions:read'

  // Projects Management (4)
  | 'projects:read'
  | 'projects:write'
  | 'projects:delete'
  | 'projects:admin'

  // API Keys Management (4)
  | 'api-keys:read'
  | 'api-keys:create'
  | 'api-keys:update'
  | 'api-keys:delete'

  // Integrations (3)
  | 'integrations:read'
  | 'integrations:configure'
  | 'integrations:delete'

  // Audit Logs (2)
  | 'audit-logs:read'
  | 'audit-logs:export'

  // Webhooks (4)
  | 'webhooks:read'
  | 'webhooks:create'
  | 'webhooks:update'
  | 'webhooks:delete'

  // Notifications (2)
  | 'notifications:read'
  | 'notifications:configure'

  // ========================================
  // PROJECT-LEVEL SCOPES (20 scopes)
  // ========================================

  // Traces (5)
  | 'traces:read'
  | 'traces:create'
  | 'traces:delete'
  | 'traces:export'
  | 'traces:share'

  // Analytics (4)
  | 'analytics:read'
  | 'analytics:export'
  | 'analytics:dashboards'
  | 'analytics:admin'

  // Models (3)
  | 'models:read'
  | 'models:configure'
  | 'models:admin'

  // Providers (2)
  | 'providers:read'
  | 'providers:configure'

  // Costs (2)
  | 'costs:read'
  | 'costs:export'

  // Prompts (4)
  | 'prompts:read'
  | 'prompts:create'
  | 'prompts:update'
  | 'prompts:delete'

/**
 * Scope level indicates where a scope applies
 */
export type ScopeLevel = 'organization' | 'project' | 'global'

/**
 * Mapping of scopes to their levels (for validation)
 */
export const SCOPE_LEVELS: Record<Scope, ScopeLevel> = {
  // Organization-level
  'organizations:read': 'organization',
  'organizations:write': 'organization',
  'organizations:delete': 'organization',
  'organizations:admin': 'organization',
  'members:read': 'organization',
  'members:invite': 'organization',
  'members:update': 'organization',
  'members:remove': 'organization',
  'members:suspend': 'organization',
  'billing:read': 'organization',
  'billing:manage': 'organization',
  'billing:export': 'organization',
  'billing:admin': 'organization',
  'settings:read': 'organization',
  'settings:write': 'organization',
  'settings:export': 'organization',
  'settings:import': 'organization',
  'settings:security': 'organization',
  'settings:admin': 'organization',
  'roles:read': 'organization',
  'roles:write': 'organization',
  'roles:delete': 'organization',
  'roles:assign': 'organization',
  'permissions:read': 'organization',
  'projects:read': 'organization',
  'projects:write': 'organization',
  'projects:delete': 'organization',
  'projects:admin': 'organization',
  'api-keys:read': 'organization',
  'api-keys:create': 'organization',
  'api-keys:update': 'organization',
  'api-keys:delete': 'organization',
  'integrations:read': 'organization',
  'integrations:configure': 'organization',
  'integrations:delete': 'organization',
  'audit-logs:read': 'organization',
  'audit-logs:export': 'organization',
  'webhooks:read': 'organization',
  'webhooks:create': 'organization',
  'webhooks:update': 'organization',
  'webhooks:delete': 'organization',
  'notifications:read': 'organization',
  'notifications:configure': 'organization',

  // Project-level
  'traces:read': 'project',
  'traces:create': 'project',
  'traces:delete': 'project',
  'traces:export': 'project',
  'traces:share': 'project',
  'analytics:read': 'project',
  'analytics:export': 'project',
  'analytics:dashboards': 'project',
  'analytics:admin': 'project',
  'models:read': 'project',
  'models:configure': 'project',
  'models:admin': 'project',
  'providers:read': 'project',
  'providers:configure': 'project',
  'costs:read': 'project',
  'costs:export': 'project',
  'prompts:read': 'project',
  'prompts:create': 'project',
  'prompts:update': 'project',
  'prompts:delete': 'project',
} as const

/**
 * Hook parameters
 */
export interface UseHasAccessParams {
  scope: Scope
  projectId?: string // Optional - only needed for project-level scopes
}

/**
 * Type-safe hook to check if user has a specific scope
 *
 * Returns false in these cases:
 * - User not authenticated
 * - Organization not selected
 * - Project-level scope requested without projectId
 * - Backend check returns false (user doesn't have scope)
 * - Error during check (safe default)
 * - Loading (safe default until verified)
 */
export function useHasAccess({ scope, projectId }: UseHasAccessParams): boolean {
  const { user } = useAuth()
  const { currentOrganizationId } = useWorkspace()

  // Get scope level for validation
  const scopeLevel = SCOPE_LEVELS[scope]

  // Build query key based on context
  const queryKey = ['scopes', user?.id, currentOrganizationId, projectId, scope]

  const { data: hasAccess } = useQuery({
    queryKey,
    queryFn: async () => {
      // Validation: must be authenticated
      if (!user || !currentOrganizationId) {
        return false
      }

      // Validation: project-level scopes require projectId
      if (scopeLevel === 'project' && !projectId) {
        console.warn(
          `[useHasAccess] Project-level scope "${scope}" requires projectId but none provided`
        )
        return false
      }

      // Call backend API to check scope
      try {
        const result = await checkUserScopes({
          userId: user.id,
          organizationId: currentOrganizationId,
          projectId: projectId || null,
          scopes: [scope],
        })

        return result[scope] === true
      } catch (error) {
        // Log error but return false (safe default)
        console.error(`[useHasAccess] Failed to check scope "${scope}":`, error)
        return false
      }
    },
    enabled: !!user && !!currentOrganizationId, // Only run if authenticated + org selected
    staleTime: 5 * 60 * 1000, // Cache for 5 minutes
    gcTime: 10 * 60 * 1000, // Keep in cache for 10 minutes
    retry: 1, // Retry once on failure
    initialData: false, // Safe default: no access while loading
  })

  return hasAccess ?? false
}

/**
 * Hook to check multiple scopes at once (returns map)
 *
 * Useful when you need to check multiple permissions for UI state:
 * ```typescript
 * const scopes = useHasMultipleAccess({
 *   scopes: ["traces:read", "traces:delete", "traces:export"],
 *   projectId
 * })
 *
 * <Button disabled={!scopes["traces:delete"]}>Delete</Button>
 * <Button disabled={!scopes["traces:export"]}>Export</Button>
 * ```
 */
export interface UseHasMultipleAccessParams {
  scopes: Scope[]
  projectId?: string
}

export function useHasMultipleAccess({
  scopes,
  projectId,
}: UseHasMultipleAccessParams): Record<string, boolean> {
  const { user } = useAuth()
  const { currentOrganizationId } = useWorkspace()

  const queryKey = ['scopes-multiple', user?.id, currentOrganizationId, projectId, ...scopes]

  const { data: scopeResults } = useQuery({
    queryKey,
    queryFn: async () => {
      if (!user || !currentOrganizationId) {
        // Return all false if not authenticated
        return scopes.reduce((acc, scope) => ({ ...acc, [scope]: false }), {})
      }

      try {
        return await checkUserScopes({
          userId: user.id,
          organizationId: currentOrganizationId,
          projectId: projectId || null,
          scopes,
        })
      } catch (error) {
        console.error('[useHasMultipleAccess] Failed to check scopes:', error)
        // Return all false on error (safe default)
        return scopes.reduce((acc, scope) => ({ ...acc, [scope]: false }), {})
      }
    },
    enabled: !!user && !!currentOrganizationId,
    staleTime: 5 * 60 * 1000,
    gcTime: 10 * 60 * 1000,
    retry: 1,
    initialData: scopes.reduce((acc, scope) => ({ ...acc, [scope]: false }), {}),
  })

  return scopeResults ?? scopes.reduce((acc, scope) => ({ ...acc, [scope]: false }), {})
}

/**
 * Helper to get scope level from scope name (client-side validation)
 */
export function getScopeLevel(scope: Scope): ScopeLevel {
  return SCOPE_LEVELS[scope]
}

/**
 * Helper to check if a scope is organization-level
 */
export function isOrganizationScope(scope: Scope): boolean {
  return SCOPE_LEVELS[scope] === 'organization'
}

/**
 * Helper to check if a scope is project-level
 */
export function isProjectScope(scope: Scope): boolean {
  return SCOPE_LEVELS[scope] === 'project'
}
