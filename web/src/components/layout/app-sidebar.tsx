'use client'

import * as React from 'react'
import { useParams } from 'next/navigation'

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
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton
              size='lg'
              className='gap-2 hover:bg-transparent active:bg-transparent'
            >
              <div className='flex aspect-square size-8 items-center justify-center'>
                <BrokleLogo className='size-6' />
              </div>
              <span className='text-lg font-semibold'>Brokle</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
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
