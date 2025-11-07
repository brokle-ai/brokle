import { useWorkspace } from '@/context/workspace-context'
import { usePathname } from 'next/navigation'
import { useRoutePermissions, extractRequiredScopes } from './rbac/route-rbac-utils'
import { useFeatureFlagMap } from '@/lib/feature-flags'
import { ROUTES } from '@/lib/navigation/routes'
import { type NavigationContext } from '@/lib/navigation/types'
import { useMemo } from 'react'

export function useNavigationContext() {
  const workspace = useWorkspace()
  const pathname = usePathname()

  const allScopes = useMemo(() => extractRequiredScopes(ROUTES), [])

  const { permissions, isLoading: isPermissionsLoading } = useRoutePermissions(
    allScopes,
    workspace.currentProject?.id ?? null
  )

  const featureFlags = useFeatureFlagMap()

  const context: NavigationContext = useMemo(() => ({
    currentOrganizationId: workspace.currentOrganization?.id ?? null,
    currentProjectId: workspace.currentProject?.id ?? null,
    currentOrgSlug: workspace.currentOrganization?.compositeSlug ?? null,
    currentProjectSlug: workspace.currentProject?.compositeSlug ?? null,
    pathname,
    currentProject: workspace.currentProject,
    currentOrganization: workspace.currentOrganization,
  }), [workspace.currentOrganization, workspace.currentProject, pathname])

  return {
    context,
    permissions,
    featureFlags,
    isLoading: workspace.isLoading || isPermissionsLoading,
    isPermissionsLoading,
    user: workspace.user ? {
      name: workspace.user.name ?? '',
      email: workspace.user.email ?? '',
      avatar: workspace.user.avatar,
    } : null,
  }
}
