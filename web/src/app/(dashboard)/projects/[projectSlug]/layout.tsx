import { cookies } from 'next/headers'
import { ProjectSidebar } from '@/features/projects'
import { SidebarWrapper } from '@/components/layout/sidebar-wrapper'

interface ProjectLayoutProps {
  children: React.ReactNode
  params: Promise<{ projectSlug: string }>
}

export default async function ProjectLayout({
  children,
  params,
}: ProjectLayoutProps) {
  // Read sidebar state from cookie (server-side) to avoid hydration mismatch
  const cookieStore = await cookies()
  const sidebarCookie = cookieStore.get('sidebar_state')
  const defaultOpen = sidebarCookie?.value !== 'false'

  // WorkspaceProvider in parent layout handles context detection from URL
  // Auto-detects both project and its parent organization
  return (
    <SidebarWrapper defaultOpen={defaultOpen} sidebar={<ProjectSidebar />}>
      {children}
    </SidebarWrapper>
  )
}