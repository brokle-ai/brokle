'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { AuthenticatedLayout } from "@/components/layout/authenticated-layout"
import { WorkspaceProvider, useWorkspace } from '@/context/workspace-context'
import { useAuthStore } from '@/features/authentication'
import { SidebarProvider, SidebarInset } from '@/components/ui/sidebar'
import { AppSidebar } from '@/components/layout/app-sidebar'
import { useNavigationContext } from '@/hooks/use-navigation-context'
import { processNavigation } from '@/lib/navigation/process-routes'
import { ROUTES as NAV_ROUTES } from '@/lib/navigation/routes'
import { ROUTES } from '@/lib/routes'
import { WorkspaceErrorPage } from '@/components/errors/workspace-error-page'
import type { User, Organization, OrganizationWithProjects } from '@/features/authentication'

interface DashboardLayoutClientProps {
  children: React.ReactNode
  defaultOpen: boolean
  // Server-fetched bootstrap (DAL → /v1/users/me).
  // Always present in production: layout.tsx redirects via DAL on
  // 401 before this component renders. Required (not optional) so
  // the type system enforces the server fetch.
  initialUser: User
  initialOrganizations: OrganizationWithProjects[]
}

export function DashboardLayoutClient({
  children,
  defaultOpen,
  initialUser,
  initialOrganizations,
}: DashboardLayoutClientProps) {
  const isAuthenticated = useAuthStore(state => state.isAuthenticated)
  const router = useRouter()

  // Hydrate Zustand from server-fetched data on first render.
  //
  // useState's lazy initializer fires synchronously during the first
  // render, BEFORE children mount, so any descendant reading from
  // the store sees the hydrated state on its first render too.
  // The guard inside hydrate() makes the call idempotent across
  // StrictMode double-renders and HMR.
  useState(() => {
    // Pick a sensible default organization to seed the auth-store's
    // legacy `organization` field (some components still read this
    // directly). Workspace context owns the richer org/project tree.
    const defaultOrg: Organization | null = pickDefaultOrganization(
      initialUser,
      initialOrganizations,
    )

    useAuthStore.getState().hydrate({
      user: initialUser,
      organization: defaultOrg,
    })
    return true
  })

  // Belt-and-suspenders for in-app auth-state changes (logout in
  // another tab via storage event, cookie cleared in devtools, refresh
  // failed mid-session). proxy.ts (route guard) handles cold-load
  // redirect; this covers state changes after mount.
  useEffect(() => {
    if (!isAuthenticated) {
      router.replace(ROUTES.SIGNIN)
    }
  }, [isAuthenticated, router])

  // If isAuthenticated flipped to false after mount (logout, cookie
  // cleared), render nothing while the redirect runs. Server-side
  // verification guarantees we always start authenticated; this branch
  // only ever hits on client-side state change.
  if (!isAuthenticated) {
    return null
  }

  return (
    <WorkspaceProvider initialData={{ user: initialUser, organizations: initialOrganizations }}>
      <AuthenticatedLayout>
        <DashboardLayoutContent defaultOpen={defaultOpen}>
          {children}
        </DashboardLayoutContent>
      </AuthenticatedLayout>
    </WorkspaceProvider>
  )
}

// pickDefaultOrganization — selects the org that should populate the
// auth-store's legacy single-organization slot. Honours
// `defaultOrganizationId` when present, falls back to the first.
function pickDefaultOrganization(
  user: User,
  organizations: OrganizationWithProjects[],
): Organization | null {
  if (organizations.length === 0) return null
  const preferred = user.defaultOrganizationId
    ? organizations.find(o => o.id === user.defaultOrganizationId)
    : undefined
  const chosen = preferred ?? organizations[0]
  return {
    id: chosen.id,
    name: chosen.name,
    plan: chosen.plan,
    members: [],
    usage: {
      traces_this_month: 0,
      observed_cost_this_month: 0,
      models_observed: 0,
    },
    createdAt: chosen.createdAt,
    updatedAt: chosen.updatedAt,
  }
}

function DashboardLayoutContent({
  children,
  defaultOpen
}: {
  children: React.ReactNode
  defaultOpen: boolean
}) {
  const { error } = useWorkspace()
  const navigationContext = useNavigationContext()

  // Show full-page error (no sidebar, no header chrome) for any
  // workspace error that survives the server-side bootstrap (rare —
  // server pre-fetched, so error is only set on a stale revalidation
  // failure).
  if (error) {
    return <WorkspaceErrorPage error={error} />
  }

  // Server-side bootstrap means workspace data is always available
  // on first render. No loading spinner here.
  const { mainNavigation, secondaryNavigation } = processNavigation({
    routes: NAV_ROUTES,
    context: navigationContext.context,
    permissions: navigationContext.permissions,
    featureFlags: navigationContext.featureFlags,
    isPermissionsLoading: navigationContext.isPermissionsLoading,
  })

  return (
    <SidebarProvider defaultOpen={defaultOpen}>
      <AppSidebar
        mainNavigation={mainNavigation}
        secondaryNavigation={secondaryNavigation}
        user={navigationContext.user}
        isLoading={navigationContext.isLoading}
      />
      <SidebarInset>
        {children}
      </SidebarInset>
    </SidebarProvider>
  )
}
