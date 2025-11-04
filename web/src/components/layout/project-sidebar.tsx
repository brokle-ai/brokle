'use client'

import Link from 'next/link'
import { usePathname, useParams } from 'next/navigation'
import { 
  BarChart3, 
  DollarSign, 
  Cpu, 
  Settings, 
  FolderOpen,
  Key,
  Puzzle,
  Shield,
  AlertTriangle
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

const projectNavItems = [
  {
    title: 'Overview',
    url: (projectSlug: string) => `/projects/${projectSlug}`,
    icon: FolderOpen,
  },
  {
    title: 'Analytics',
    url: (projectSlug: string) => `/projects/${projectSlug}/analytics`,
    icon: BarChart3,
  },
  {
    title: 'Costs',
    url: (projectSlug: string) => `/projects/${projectSlug}/costs`,
    icon: DollarSign,
    badge: '$1.2k',
  },
  {
    title: 'Models',
    url: (projectSlug: string) => `/projects/${projectSlug}/models`,
    icon: Cpu,
  },
]

const projectSettingsItems = [
  {
    title: 'General',
    url: (projectSlug: string) => `/projects/${projectSlug}/settings`,
    icon: Settings,
  },
  {
    title: 'API Keys',
    url: (projectSlug: string) => `/projects/${projectSlug}/settings/api-keys`,
    icon: Key,
  },
  {
    title: 'Integrations',
    url: (projectSlug: string) => `/projects/${projectSlug}/settings/integrations`,
    icon: Puzzle,
  },
  {
    title: 'Security',
    url: (projectSlug: string) => `/projects/${projectSlug}/settings/security`,
    icon: Shield,
  },
  {
    title: 'Danger Zone',
    url: (projectSlug: string) => `/projects/${projectSlug}/settings/danger`,
    icon: AlertTriangle,
  },
]

export function ProjectSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const pathname = usePathname()
  const params = useParams()
  const { setOpenMobile } = useSidebar()
  
  const projectSlug = params?.projectSlug as string

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
        {/* Main Navigation */}
        <SidebarGroup>
          <SidebarGroupLabel>Project</SidebarGroupLabel>
          <SidebarMenu>
            {projectNavItems.map((item) => {
              const href = item.url(projectSlug)
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
                      {item.badge && (
                        <div className="ml-auto text-xs bg-secondary text-secondary-foreground px-1 py-0 rounded text-center min-w-[1.5rem] h-4 flex items-center justify-center">
                          {item.badge}
                        </div>
                      )}
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              )
            })}
          </SidebarMenu>
        </SidebarGroup>

        {/* Settings Section */}
        <SidebarGroup>
          <SidebarGroupLabel>Settings</SidebarGroupLabel>
          <SidebarMenu>
            {projectSettingsItems.map((item) => {
              const href = item.url(projectSlug)
              const isActive = pathname === href
              const Icon = item.icon
              const isDangerZone = item.title === 'Danger Zone'

              return (
                <SidebarMenuItem key={item.title}>
                  <SidebarMenuButton
                    asChild
                    isActive={isActive}
                    tooltip={item.title}
                    className={isDangerZone ? 'hover:bg-destructive/10 hover:text-destructive' : ''}
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