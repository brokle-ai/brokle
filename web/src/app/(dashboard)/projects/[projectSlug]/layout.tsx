import { ProjectProvider } from '@/context/project-context'
import { ProjectSidebar } from '@/components/layout/project-sidebar'
import { SidebarWrapper } from '@/components/layout/sidebar-wrapper'

interface ProjectLayoutProps {
  children: React.ReactNode
  params: Promise<{ projectSlug: string }>
}

export default async function ProjectLayout({
  children,
  params,
}: ProjectLayoutProps) {
  const { projectSlug } = await params
  
  return (
    <ProjectProvider compositeSlug={projectSlug}>
      <SidebarWrapper sidebar={<ProjectSidebar />}>
        {children}
      </SidebarWrapper>
    </ProjectProvider>
  )
}