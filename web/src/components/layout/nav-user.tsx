'use client'

import Link from 'next/link'
import {
  ChevronsUpDown,
  CreditCard,
  LogOut,
  Monitor,
  Moon,
  Sparkles,
  Sun,
} from 'lucide-react'
import { getUserInitials } from '@/lib/utils/user-utils'
import { buildProjectUrl } from '@/lib/utils/slug-utils'
import { useWorkspace } from '@/context/workspace-context'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from '@/components/ui/sidebar'
import { useLogoutMutation } from '@/features/authentication'
import { useTheme } from '@/context/theme-context'

export function NavUser({
  user,
}: {
  user: {
    name: string
    email: string
    avatar?: string
  }
}) {
  const { isMobile } = useSidebar()
  const { mutate: handleLogout, isPending: isLoggingOut } = useLogoutMutation()
  const { theme, setTheme } = useTheme()
  const { currentProject } = useWorkspace()

  // Compute user initials from full name
  const initials = getUserInitials({ name: user.name, email: user.email })

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <SidebarMenuButton
              size='lg'
              className='data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground'
            >
              <Avatar className='h-8 w-8 rounded-lg'>
                <AvatarImage src={user.avatar} alt={user.name} />
                <AvatarFallback className='rounded-lg'>{initials}</AvatarFallback>
              </Avatar>
              <div className='grid flex-1 text-left text-sm leading-tight'>
                <span className='truncate font-semibold'>{user.name}</span>
                <span className='truncate text-xs'>{user.email}</span>
              </div>
              <ChevronsUpDown className='ml-auto size-4' />
            </SidebarMenuButton>
          </DropdownMenuTrigger>
          <DropdownMenuContent
            className='w-(--radix-dropdown-menu-trigger-width) min-w-56 rounded-lg'
            side={isMobile ? 'bottom' : 'right'}
            align='end'
            sideOffset={4}
          >
            <DropdownMenuLabel className='p-0 font-normal'>
              <div className='flex items-center gap-2 px-1 py-1.5 text-left text-sm'>
                <Avatar className='h-8 w-8 rounded-lg'>
                  <AvatarImage src={user.avatar} alt={user.name} />
                  <AvatarFallback className='rounded-lg'>{initials}</AvatarFallback>
                </Avatar>
                <div className='grid flex-1 text-left text-sm leading-tight'>
                  <span className='truncate font-semibold'>{user.name}</span>
                  <span className='truncate text-xs'>{user.email}</span>
                </div>
              </div>
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuGroup>
              <DropdownMenuItem>
                <Sparkles />
                Upgrade to Pro
              </DropdownMenuItem>
            </DropdownMenuGroup>
            <DropdownMenuSeparator />
            {currentProject && (
              <DropdownMenuGroup>
                <DropdownMenuItem asChild>
                  <Link href={buildProjectUrl(currentProject.name, currentProject.id, 'settings/organization/billing')}>
                    <CreditCard />
                    Billing
                  </Link>
                </DropdownMenuItem>
              </DropdownMenuGroup>
            )}
            <DropdownMenuSeparator />
            <DropdownMenuGroup>
              <DropdownMenuLabel className='text-xs text-muted-foreground px-2 py-1.5'>
                Theme
              </DropdownMenuLabel>
              <div className='flex items-center gap-1 px-2 pb-1.5 w-full'>
                <Button
                  variant={theme === 'light' ? 'secondary' : 'ghost'}
                  size='icon'
                  className='h-8 flex-1'
                  onClick={() => setTheme('light')}
                >
                  <Sun className='h-4 w-4' />
                </Button>
                <Button
                  variant={theme === 'dark' ? 'secondary' : 'ghost'}
                  size='icon'
                  className='h-8 flex-1'
                  onClick={() => setTheme('dark')}
                >
                  <Moon className='h-4 w-4' />
                </Button>
                <Button
                  variant={theme === 'system' ? 'secondary' : 'ghost'}
                  size='icon'
                  className='h-8 flex-1'
                  onClick={() => setTheme('system')}
                >
                  <Monitor className='h-4 w-4' />
                </Button>
              </div>
            </DropdownMenuGroup>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onClick={() => handleLogout()}
              disabled={isLoggingOut}
            >
              <LogOut />
              {isLoggingOut ? 'Logging out...' : 'Log out'}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  )
}
