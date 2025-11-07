'use client'

import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
  SidebarSeparator,
  SidebarMenu,
  SidebarMenuItem,
  SidebarMenuButton,
} from '@/components/ui/sidebar'
import { NavMain } from '@/components/layout/nav-main'
import { NavUser } from '@/components/layout/nav-user'
import { BrokleLogo } from '@/assets/brokle-logo'
import { SidebarSkeleton } from '@/components/layout/sidebar-skeleton'
import { type ProcessedRoute, type RouteGroup } from '@/lib/navigation/types'

interface AppSidebarProps extends React.ComponentProps<typeof Sidebar> {
  mainNavigation: {
    grouped: Partial<Record<RouteGroup, ProcessedRoute[]>>
    ungrouped: ProcessedRoute[]
  }
  secondaryNavigation: {
    grouped: Partial<Record<RouteGroup, ProcessedRoute[]>>
    ungrouped: ProcessedRoute[]
  }
  user: {
    name: string
    email: string
    avatar?: string
  } | null
  isLoading?: boolean
}

export function AppSidebar({
  mainNavigation,
  secondaryNavigation,
  user,
  isLoading,
  ...props
}: AppSidebarProps) {
  if (isLoading) {
    return <SidebarSkeleton />
  }

  return (
    <Sidebar {...props} collapsible="icon" variant="sidebar">
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton
              size="lg"
              className="gap-2 hover:bg-transparent active:bg-transparent"
            >
              <div className="flex aspect-square size-8 items-center justify-center">
                <BrokleLogo className="size-6" />
              </div>
              <span className="text-lg font-semibold">Brokle</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
        <SidebarSeparator />
      </SidebarHeader>

      <SidebarContent>
        <NavMain items={mainNavigation} />
        <div className="flex-1" />
        <NavMain items={secondaryNavigation} />
      </SidebarContent>

      <SidebarFooter>
        {user && <NavUser user={user} />}
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  )
}
