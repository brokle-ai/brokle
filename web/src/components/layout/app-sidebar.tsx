'use client'

import Link from 'next/link'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
  SidebarMenu,
  SidebarMenuItem,
  SidebarMenuButton,
} from '@/components/ui/sidebar'
import { NavMain } from '@/components/layout/nav-main'
import { NavUser } from '@/components/layout/nav-user'
import { BrokleLogo } from '@/components/ui/brokle-logo'
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
      <SidebarHeader className="h-12 border-b mb-2">
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton
              size="lg"
              className="gap-2 hover:bg-transparent active:bg-transparent"
            >
              <Link href="/" className="flex items-center gap-2 cursor-pointer">
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
        <NavMain items={mainNavigation} />
        <div className="flex-1" />

        {/* Secondary navigation with separator */}
        {(secondaryNavigation.ungrouped.length > 0 ||
          Object.keys(secondaryNavigation.grouped).length > 0) && (
          <div className="border-t border-sidebar-border pt-2">
            <NavMain items={secondaryNavigation} />
          </div>
        )}
      </SidebarContent>

      <SidebarFooter>
        {user && <NavUser user={user} />}
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  )
}
