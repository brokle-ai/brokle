import type { LucideIcon } from 'lucide-react'
import { Link, useRouterState } from '@tanstack/react-router'
import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from '@/components/ui/sidebar'

export interface NavRoute {
  title: string
  url: string
  icon?: LucideIcon
  badge?: string
}

export interface NavGroupItems {
  label: string
  items: NavRoute[]
}

interface NavMainProps {
  groups: NavGroupItems[]
  ungrouped?: NavRoute[]
}

export function NavMain({ groups, ungrouped = [] }: NavMainProps) {
  const { setOpenMobile } = useSidebar()
  const { location } = useRouterState()
  const pathname = location.pathname

  const isActive = (url: string) => pathname === url || pathname.startsWith(`${url}/`)

  return (
    <>
      {ungrouped.length > 0 && (
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu>
              {ungrouped.map((route) => (
                <NavItem
                  key={route.url}
                  route={route}
                  active={isActive(route.url)}
                  onNavigate={() => setOpenMobile(false)}
                />
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      )}

      {groups.map((group) => (
        <SidebarGroup key={group.label}>
          <SidebarGroupLabel>{group.label}</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {group.items.map((route) => (
                <NavItem
                  key={route.url}
                  route={route}
                  active={isActive(route.url)}
                  onNavigate={() => setOpenMobile(false)}
                />
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      ))}
    </>
  )
}

function NavItem({
  route,
  active,
  onNavigate,
}: {
  route: NavRoute
  active: boolean
  onNavigate: () => void
}) {
  const Icon = route.icon
  return (
    <SidebarMenuItem>
      <SidebarMenuButton asChild isActive={active} tooltip={route.title}>
        <Link to={route.url} onClick={onNavigate}>
          {Icon ? <Icon /> : null}
          <span>{route.title}</span>
          {route.badge ? (
            <div className="ml-auto flex h-4 min-w-[1.5rem] items-center justify-center rounded bg-secondary px-1 py-0 text-center text-xs text-secondary-foreground">
              {route.badge}
            </div>
          ) : null}
        </Link>
      </SidebarMenuButton>
    </SidebarMenuItem>
  )
}
