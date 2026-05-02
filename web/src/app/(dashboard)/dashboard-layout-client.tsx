'use client'

import { useState } from 'react'
import { AuthenticatedLayout } from "@/components/layout/authenticated-layout"
import { WorkspaceProvider, useWorkspace } from '@/context/workspace-context'
import { useAuthStore } from '@/features/authentication'
import { SidebarProvider, SidebarInset } from '@/components/ui/sidebar'
import { AppSidebar } from '@/components/layout/app-sidebar'
import { useNavigationContext } from '@/hooks/use-navigation-context'
import { processNavigation } from '@/lib/navigation/process-routes'
import { ROUTES as NAV_ROUTES } from '@/lib/navigation/routes'
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
  // Seed the legacy auth-store from server-fetched data so components
  // that still read from it (auth-status, etc.) see a populated user
  // on first render. The workspace context below owns the richer
  // org/project tree and is the canonical source for new code.
  //
  // useState's lazy initializer runs synchronously during the first
  // render and before any descendant mounts. The guard inside
  // hydrate() keeps the call idempotent across StrictMode
  // double-renders and HMR.
  //
  // No client-side cold-load redirect guard here — auth is enforced
  // by proxy.ts (cookie-presence gate) + lib/auth/dal.ts (server-side
  // round-trip). Layout.tsx only reaches this component with a valid
  // session. Cross-tab logout / runtime session expiry are handled
  // reactively by the `auth:session-expired` and `storage` listeners
  // in components/providers.tsx — those react to actual events, not
  // a stale comparison against an un-hydrated store snapshot. The
  // earlier `useEffect(() => router.replace(SIGNIN))` here was a
  // closure-stale anti-pattern that the Next.js 16 auth guide
  // explicitly warns against.
  useState(() => {
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
