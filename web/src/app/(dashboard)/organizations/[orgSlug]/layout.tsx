import { OrgProvider } from '@/context/org-context'
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
  const { orgSlug } = await params
  
  return (
    <OrgProvider compositeSlug={orgSlug}>
      <SidebarWrapper sidebar={<OrgSidebar />}>
        {children}
      </SidebarWrapper>
    </OrgProvider>
  )
}