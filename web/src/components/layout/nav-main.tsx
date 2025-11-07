'use client'

import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from '@/components/ui/sidebar'
import Link from 'next/link'
import { type ProcessedRoute, type RouteGroup, type BadgeConfig } from '@/lib/navigation/types'

interface NavMainProps {
  items: {
    grouped: Partial<Record<RouteGroup, ProcessedRoute[]>>
    ungrouped: ProcessedRoute[]
  }
}

export function NavMain({ items }: NavMainProps) {
  const { setOpenMobile } = useSidebar()

  return (
    <>
      {/* Ungrouped items */}
      {items.ungrouped.length > 0 && (
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu>
              {items.ungrouped.map(route => (
                <NavItem
                  key={route.pathname}
                  route={route}
                  onNavigate={() => setOpenMobile(false)}
                />
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      )}

      {/* Grouped items */}
      {Object.entries(items.grouped).map(([group, routes]) => (
        <SidebarGroup key={group}>
          <SidebarGroupLabel>{group}</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {routes.map(route => (
                <NavItem
                  key={route.pathname}
                  route={route}
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

function NavItem({ route, onNavigate }: { route: ProcessedRoute, onNavigate: () => void }) {
  // Handle custom menu nodes
  if (route.menuNode) {
    return (
      <SidebarMenuItem>
        {route.menuNode}
      </SidebarMenuItem>
    )
  }

  return (
    <SidebarMenuItem>
      <SidebarMenuButton
        asChild
        isActive={route.isActive}
        tooltip={route.title}
        className={
          route.title === 'Danger Zone'
            ? 'hover:bg-destructive/10 hover:text-destructive'
            : ''
        }
      >
        <Link
          href={route.url}
          target={route.newTab ? '_blank' : undefined}
          onClick={onNavigate}
        >
          {route.icon && <route.icon />}
          <span>{route.title}</span>
          {route.badge && <RouteBadge config={route.badge} />}
        </Link>
      </SidebarMenuButton>
    </SidebarMenuItem>
  )
}

function RouteBadge({ config }: { config: BadgeConfig }) {
  if (config.type === 'static') {
    return (
      <div className="ml-auto text-xs bg-secondary text-secondary-foreground px-1 py-0 rounded text-center min-w-[1.5rem] h-4 flex items-center justify-center">
        {config.value}
      </div>
    )
  }

  // TODO: Implement dynamic badge fetching once API is available
  // For now, show placeholder
  return (
    <div className="ml-auto text-xs bg-secondary text-secondary-foreground px-1 py-0 rounded text-center min-w-[1.5rem] h-4 flex items-center justify-center">
      --
    </div>
  )
}
