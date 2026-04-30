/**
 * useHasProjectAccess — type-safe project-tier authorization hook.
 *
 * Pure synchronous in-memory lookup against the workspace bootstrap
 * data (`/api/v1/users/me`, cached by `useWorkspace`). NO per-check
 * API round trip. Mirrors Langfuse's session-bootstrap RBAC pattern
 * (competitors/langfuse/web/src/features/rbac/utils/checkProjectAccess.ts).
 *
 * Project-tier scopes resolve additively: the user's org-role
 * project-tier projection UNIONs with any project_members grant
 * (computed server-side in the `project_scope_perms` CTE — see
 * internal/infrastructure/db/queries/project_member.sql and
 * docs/adr/0001-rbac-additive-semantics.md).
 *
 * `projectId` is REQUIRED at the type level — no short-circuit possible.
 *
 * Examples:
 * ```typescript
 * const { projectId } = useParams<{ projectId: string }>()
 * const canDeleteTraces = useHasProjectAccess({ scope: 'traces:delete', projectId })
 * const canWriteDashboards = useHasProjectAccess({ scope: 'dashboards:write', projectId })
 * ```
 */

import { useMemo } from 'react'
import { useWorkspace } from '@/context/workspace-context'
import type { ProjectScope } from '@/generated/permissions'
import type { ProjectSummary, OrganizationWithProjects } from '@/features/authentication'

export type { ProjectScope }

export interface UseHasProjectAccessParams {
  scope: ProjectScope
  /**
   * Project ID. Required — there is no fallback. Read from route
   * params (`useParams()` in App Router pages).
   */
  projectId: string
}

function findProject(
  organizations: OrganizationWithProjects[],
  projectId: string,
): ProjectSummary | undefined {
  for (const org of organizations) {
    const proj = org.projects.find((p) => p.id === projectId)
    if (proj) return proj
  }
  return undefined
}

export function useHasProjectAccess({
  scope,
  projectId,
}: UseHasProjectAccessParams): boolean {
  const { organizations } = useWorkspace()
  const project = useMemo(
    () => findProject(organizations, projectId),
    [organizations, projectId],
  )
  return project?.scopes.includes(scope) ?? false
}

export interface UseHasMultipleProjectAccessParams {
  scopes: ProjectScope[]
  projectId: string
}

/**
 * Bulk variant returning a `Record<ProjectScope, boolean>` — useful
 * when a single component renders multiple guarded controls and wants
 * one stable identity for memoisation.
 */
export function useHasMultipleProjectAccess({
  scopes,
  projectId,
}: UseHasMultipleProjectAccessParams): Record<ProjectScope, boolean> {
  const { organizations } = useWorkspace()
  const project = useMemo(
    () => findProject(organizations, projectId),
    [organizations, projectId],
  )
  const granted = new Set(project?.scopes ?? [])
  return scopes.reduce(
    (acc, s) => {
      acc[s] = granted.has(s)
      return acc
    },
    {} as Record<ProjectScope, boolean>,
  )
}
