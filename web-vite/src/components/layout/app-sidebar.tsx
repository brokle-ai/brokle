import * as React from 'react'
import { Link } from '@tanstack/react-router'
import {
  Activity,
  BarChart3,
  CheckSquare,
  CreditCard,
  Database,
  FileText,
  FlaskConical,
  LayoutDashboard,
  MessageSquare,
  Plug,
  Settings,
  Users,
} from 'lucide-react'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
} from '@/components/ui/sidebar'
import { BrokleLogo } from '@/components/ui/brokle-logo'
import { NavMain, type NavGroupItems } from '@/components/layout/nav-main'
import { NavUser, type NavUserProfile } from '@/components/layout/nav-user'
import { SidebarSkeleton } from '@/components/layout/sidebar-skeleton'

interface AppSidebarProps extends React.ComponentProps<typeof Sidebar> {
  orgId: string
  projectId: string
  user: NavUserProfile | null
  isLoading?: boolean
}

// Builds the primary + secondary nav groups for the project chrome.
// Labels and icons mirror web/ so the two surfaces stay visually in
// sync while the route tree evolves independently.
function buildNav(orgId: string, projectId: string): {
  primary: NavGroupItems[]
  secondary: NavGroupItems[]
} {
  const p = `/o/${orgId}/p/${projectId}`
  return {
    primary: [
      {
        label: 'Observability',
        items: [
          { title: 'Overview', url: p, icon: LayoutDashboard },
          { title: 'Traces', url: `${p}/traces`, icon: Activity },
          { title: 'Sessions', url: `${p}/sessions`, icon: MessageSquare },
          { title: 'Dashboards', url: `${p}/dashboards`, icon: BarChart3 },
        ],
      },
      {
        label: 'Evaluation',
        items: [
          { title: 'Datasets', url: `${p}/datasets`, icon: Database },
          { title: 'Evaluators', url: `${p}/evaluators`, icon: CheckSquare },
          { title: 'Experiments', url: `${p}/experiments`, icon: FlaskConical },
        ],
      },
      {
        label: 'Prompts',
        items: [{ title: 'Library', url: `${p}/prompts`, icon: FileText }],
      },
    ],
    secondary: [
      {
        label: 'Settings',
        items: [
          { title: 'Members', url: `${p}/settings/members`, icon: Users },
          { title: 'AI Providers', url: `${p}/settings/ai-providers`, icon: Plug },
          { title: 'Project', url: `${p}/settings`, icon: Settings },
          { title: 'Billing', url: `/o/${orgId}/billing`, icon: CreditCard },
        ],
      },
    ],
  }
}

export function AppSidebar({
  orgId,
  projectId,
  user,
  isLoading,
  ...props
}: AppSidebarProps) {
  if (isLoading) {
    return <SidebarSkeleton />
  }

  const { primary, secondary } = buildNav(orgId, projectId)

  return (
    <Sidebar {...props} collapsible="icon" variant="sidebar">
      <SidebarHeader className="h-12 border-b mb-2">
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton
              size="lg"
              className="gap-2 hover:bg-transparent active:bg-transparent"
            >
              <Link to="/" className="flex items-center gap-2 cursor-pointer">
                <div className="flex aspect-square size-8 items-center justify-center">
                  <BrokleLogo variant="icon" size="sm" />
                </div>
                <span className="text-lg font-semibold">Brokle</span>
              </Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>

      <SidebarContent>
        <NavMain groups={primary} />
        <div className="flex-1" />

        {secondary.length > 0 && (
          <div className="border-t border-sidebar-border pt-2">
            <NavMain groups={secondary} />
          </div>
        )}
      </SidebarContent>

      <SidebarFooter>{user && <NavUser user={user} />}</SidebarFooter>
      <SidebarRail />
    </Sidebar>
  )
}
