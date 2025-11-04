'use client'

import * as React from 'react'
import { useParams } from 'next/navigation'

import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
  SidebarGroup,
  SidebarSeparator,
} from '@/components/ui/sidebar'
import { NavGroup } from '@/components/layout/nav-group'
import { NavUser } from '@/components/layout/nav-user'
import { BrokleLogo } from '@/assets/brokle-logo'
import { getSidebarData } from './data/sidebar-data'

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const params = useParams()
  
  // Extract context from URL params for navigation
  const orgSlug = params?.orgSlug as string
  const projectSlug = params?.projectSlug as string
  
  // Generate context-aware navigation
  const sidebarData = getSidebarData(orgSlug, projectSlug)

  return (
    <Sidebar {...props} collapsible="icon" variant="sidebar">
      <SidebarHeader>
        <SidebarGroup className="py-2">
          <div className="flex items-center gap-2 px-2 py-1">
            <BrokleLogo className="h-6 w-6" />
            <span className="text-lg font-semibold group-data-[collapsible=icon]:hidden">
              Brokle
            </span>
          </div>
        </SidebarGroup>
        <SidebarSeparator />
      </SidebarHeader>
      <SidebarContent>
        {sidebarData.navGroups.map((navGroup) => (
          <NavGroup key={navGroup.title} {...navGroup} />
        ))}
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={sidebarData.user} />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  )
}
