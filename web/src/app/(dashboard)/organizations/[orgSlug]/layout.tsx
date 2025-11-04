import { cookies } from 'next/headers'
import { OrgSidebar } from '@/components/layout/org-sidebar'
import { SidebarWrapper } from '@/components/layout/sidebar-wrapper'

interface OrganizationLayoutProps {
  children: React.ReactNode
  params: Promise<{ orgSlug: string }>
}

export default async function OrganizationLayout({
  children,
  params,
}: OrganizationLayoutProps) {
  // Read sidebar state from cookie (server-side) to avoid hydration mismatch
  const cookieStore = await cookies()
  const sidebarCookie = cookieStore.get('sidebar_state')
  const defaultOpen = sidebarCookie?.value !== 'false'

  // WorkspaceProvider in parent layout handles context detection from URL
  return (
    <SidebarWrapper defaultOpen={defaultOpen} sidebar={<OrgSidebar />}>
      {children}
    </SidebarWrapper>
  )
}