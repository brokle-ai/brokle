'use client'

import Link from 'next/link'
import { usePathname, useParams } from 'next/navigation'
import { 
  Users, 
  CreditCard, 
  Settings, 
  Building2
} from 'lucide-react'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
  SidebarGroup,
  SidebarGroupLabel,
  SidebarSeparator,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from '@/components/ui/sidebar'
import { NavUser } from '@/components/layout/nav-user'
import { BrokleLogo } from '@/assets/brokle-logo'

const organizationNavItems = [
  {
    title: 'Overview',
    url: (orgSlug: string) => `/organizations/${orgSlug}`,
    icon: Building2,
  },
  {
    title: 'Members',
    url: (orgSlug: string) => `/organizations/${orgSlug}/members`,
    icon: Users,
  },
  {
    title: 'Billing',
    url: (orgSlug: string) => `/organizations/${orgSlug}/billing`,
    icon: CreditCard,
  },
  {
    title: 'Settings',
    url: (orgSlug: string) => `/organizations/${orgSlug}/settings`,
    icon: Settings,
  },
]

export function OrgSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const pathname = usePathname()
  const params = useParams()
  const { setOpenMobile } = useSidebar()
  
  const orgSlug = params?.orgSlug as string

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
        <SidebarGroup>
          <SidebarGroupLabel>Organization</SidebarGroupLabel>
          <SidebarMenu>
            {organizationNavItems.map((item) => {
              const href = item.url(orgSlug)
              const isActive = pathname === href
              const Icon = item.icon

              return (
                <SidebarMenuItem key={item.title}>
                  <SidebarMenuButton
                    asChild
                    isActive={isActive}
                    tooltip={item.title}
                  >
                    <Link href={href} onClick={() => setOpenMobile(false)}>
                      <Icon />
                      <span>{item.title}</span>
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              )
            })}
          </SidebarMenu>
        </SidebarGroup>
      </SidebarContent>
      
      <SidebarFooter>
        <NavUser 
          user={{
            name: 'AI Engineer',
            email: 'engineer@company.com',
            avatar: '/avatars/user.jpg',
          }}
        />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  )
}