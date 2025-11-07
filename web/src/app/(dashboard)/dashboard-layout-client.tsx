'use client'

import { useEffect } from 'react'
import { AuthenticatedLayout } from "@/components/layout/authenticated-layout"
import { WorkspaceProvider } from '@/context/workspace-context'
import { useAuthStore } from '@/features/authentication'
import { SidebarProvider, SidebarInset } from '@/components/ui/sidebar'
import { AppSidebar } from '@/components/layout/app-sidebar'
import { useNavigationContext } from '@/hooks/use-navigation-context'
import { processNavigation } from '@/lib/navigation/process-routes'
import { ROUTES } from '@/lib/navigation/routes'

interface DashboardLayoutClientProps {
  children: React.ReactNode
  defaultOpen: boolean
}

export function DashboardLayoutClient({ children, defaultOpen }: DashboardLayoutClientProps) {
  const initializeAuth = useAuthStore(state => state.initializeAuth)
  const isLoading = useAuthStore(state => state.isLoading)

  useEffect(() => {
    initializeAuth()
  }, [initializeAuth])

  if (isLoading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="text-center">
          <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent mx-auto mb-4" />
          <p className="text-sm text-muted-foreground">Loading...</p>
        </div>
      </div>
    )
  }

  return (
    <WorkspaceProvider>
      <AuthenticatedLayout>
        <DashboardLayoutContent defaultOpen={defaultOpen}>
          {children}
        </DashboardLayoutContent>
      </AuthenticatedLayout>
    </WorkspaceProvider>
  )
}

function DashboardLayoutContent({
  children,
  defaultOpen
}: {
  children: React.ReactNode
  defaultOpen: boolean
}) {
  const navigationContext = useNavigationContext()

  const { mainNavigation, secondaryNavigation } = processNavigation({
    routes: ROUTES,
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
