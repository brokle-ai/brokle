/**
 * useHasOrganizationAccess — type-safe org-tier authorization hook.
 *
 * Pure synchronous in-memory lookup against the workspace bootstrap
 * data (`/api/v1/users/me`, cached by `useWorkspace`). NO per-check
 * API round trip. Mirrors Langfuse's session-bootstrap RBAC pattern
 * (competitors/langfuse/web/src/features/rbac/utils/checkOrganizationAccess.ts).
 *
 * Org-tier scopes resolve against the user's organization role only.
 * project_members grants do NOT affect org-tier permissions
 * (backend: scope-partitioned resolver in
 * internal/infrastructure/db/queries/project_member.sql
 * `org_scope_perms` CTE — `WHERE p.scope_level = 'organization'`).
 *
 * Examples:
 * ```typescript
 * const canCreateProject = useHasOrganizationAccess({ scope: 'org_projects:create' })
 * const canInviteMembers = useHasOrganizationAccess({ scope: 'org_members:invite' })
 *
 * // Wrong-tier scope is a TypeScript error:
 * // useHasOrganizationAccess({ scope: 'traces:delete' }) // ❌ ProjectScope, not OrganizationScope
 * ```
 *
 * Returns false when:
 * - Workspace data not yet loaded (safe default during bootstrap)
 * - No organization context (no current org and none provided as override)
 * - User's role for the org does not grant the scope
 */

import { useWorkspace } from '@/context/workspace-context'
// `OrganizationScope` is generated from seeds/permissions.yaml — the
// backend permission catalog is the single source of truth. Run
// `make gen-frontend-permissions` after editing the YAML; the drift-
// guard test in internal/seeder/frontend_permissions_test.go fails CI
// if the generated file is stale. See CLAUDE.md 2026-04-30 (project-rbac).
import type { OrganizationScope } from '@/generated/permissions'

export type { OrganizationScope }

export interface UseHasOrganizationAccessParams {
  scope: OrganizationScope
  /**
   * Organization ID override. Defaults to the current workspace's
   * organization. Pass explicitly for cross-org checks (e.g., role
   * editor showing permissions for a different org).
   */
  organizationId?: string
}

export function useHasOrganizationAccess({
  scope,
  organizationId,
}: UseHasOrganizationAccessParams): boolean {
  const { organizations, currentOrganization } = useWorkspace()
  const orgId = organizationId ?? currentOrganization?.id
  if (!orgId) return false
  const org = organizations.find((o) => o.id === orgId)
  return org?.scopes.includes(scope) ?? false
}

export interface UseHasMultipleOrganizationAccessParams {
  scopes: OrganizationScope[]
  organizationId?: string
}

/**
 * Bulk variant returning a `Record<OrganizationScope, boolean>` — useful
 * when a single component renders multiple guarded controls and wants
 * one stable identity for memoisation.
 */
export function useHasMultipleOrganizationAccess({
  scopes,
  organizationId,
}: UseHasMultipleOrganizationAccessParams): Record<OrganizationScope, boolean> {
  const { organizations, currentOrganization } = useWorkspace()
  const orgId = organizationId ?? currentOrganization?.id
  const org = orgId ? organizations.find((o) => o.id === orgId) : undefined
  const granted = new Set(org?.scopes ?? [])
  return scopes.reduce(
    (acc, s) => {
      acc[s] = granted.has(s)
      return acc
    },
    {} as Record<OrganizationScope, boolean>,
  )
}
