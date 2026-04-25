import { Outlet, useParams } from '@tanstack/react-router'
import { SidebarInset, SidebarProvider } from '@/components/ui/sidebar'
import { AppSidebar } from '@/components/layout/app-sidebar'
import { Header } from '@/components/layout/header'
import { Breadcrumbs } from '@/components/layout/breadcrumbs'
import { OrganizationSelector } from '@/components/layout/organization-selector'
import { useAuthStore } from '@/stores/auth-store'

// Full dashboard chrome: collapsible sidebar + scroll-aware header with
// breadcrumbs and org switcher + outlet for the page content. Mounted
// from the `/o/$orgId/p/$projectId` route so every nested route shares
// the same shell and the same query-primed tenancy data.
export function AuthenticatedLayout() {
  const { orgId, projectId } = useParams({ strict: false }) as {
    orgId?: string
    projectId?: string
  }
  const sessionUser = useAuthStore((s) => s.user)

  if (!orgId || !projectId) return <Outlet />

  const user = sessionUser
    ? {
        name:
          [sessionUser.first_name, sessionUser.last_name]
            .filter(Boolean)
            .join(' ')
            .trim() || sessionUser.email,
        email: sessionUser.email,
      }
    : null

  return (
    <SidebarProvider defaultOpen>
      <AppSidebar orgId={orgId} projectId={projectId} user={user} />
      <SidebarInset>
        <Header>
          <div className="flex flex-1 items-center gap-3">
            <OrganizationSelector currentOrgId={orgId} />
            <span className="text-muted-foreground">/</span>
            <Breadcrumbs />
          </div>
        </Header>
        <Outlet />
      </SidebarInset>
    </SidebarProvider>
  )
}
