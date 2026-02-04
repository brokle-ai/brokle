'use client'

import Link from 'next/link'
import { LogOut } from 'lucide-react'
import { cn } from '@/lib/utils'
import { getUserInitials } from '@/lib/utils/user-utils'
import { buildProjectUrl } from '@/lib/utils/slug-utils'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuShortcut,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { useAuth, useLogoutMutation } from '@/features/authentication'
import type { ProjectSummary } from '@/features/authentication'

interface ProfileDropdownProps {
  className?: string
  /** Current project for building settings links. Pass null when outside WorkspaceProvider. */
  currentProject?: ProjectSummary | null
}

/**
 * ProfileDropdown - User profile menu with settings links
 *
 * This component is context-agnostic. The parent is responsible for providing
 * currentProject from useWorkspace() when inside WorkspaceProvider, or null when outside.
 *
 * @example
 * ```tsx
 * // Inside dashboard (with WorkspaceProvider)
 * const { currentProject } = useWorkspace()
 * <ProfileDropdown currentProject={currentProject} />
 *
 * // Outside dashboard (no WorkspaceProvider)
 * <ProfileDropdown currentProject={null} />
 * ```
 */
export function ProfileDropdown({ className, currentProject = null }: ProfileDropdownProps) {
  const { user } = useAuth()
  const { mutate: handleLogout, isPending: isLoggingOut } = useLogoutMutation()

  // Only render on authenticated pages
  if (!user) return null

  // Compute user display values
  const initials = getUserInitials({
    firstName: user.firstName,
    lastName: user.lastName,
    email: user.email
  })
  const displayName = `${user.firstName || ''} ${user.lastName || ''}`.trim() || user.email || 'User'

  return (
    <DropdownMenu modal={false}>
      <DropdownMenuTrigger asChild>
        <Button variant='ghost' className={cn('relative h-8 w-8 rounded-full', className)}>
          <Avatar className='h-8 w-8'>
            <AvatarImage src={user.avatar || undefined} alt={displayName} />
            <AvatarFallback>{initials}</AvatarFallback>
          </Avatar>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent className='w-56' align='end' forceMount>
        <DropdownMenuLabel className='font-normal'>
          <div className='flex flex-col space-y-1'>
            <p className='text-sm leading-none font-medium'>{displayName}</p>
            <p className='text-muted-foreground text-xs leading-none'>
              {user.email}
            </p>
          </div>
        </DropdownMenuLabel>
        <DropdownMenuSeparator />
        {currentProject && (
          <DropdownMenuGroup>
            <DropdownMenuItem asChild>
              <Link href={buildProjectUrl(currentProject.name, currentProject.id, 'settings/profile')}>
                Profile
                <DropdownMenuShortcut>⇧⌘P</DropdownMenuShortcut>
              </Link>
            </DropdownMenuItem>
            <DropdownMenuItem asChild>
              <Link href={buildProjectUrl(currentProject.name, currentProject.id, 'settings/organization/billing')}>
                Billing
                <DropdownMenuShortcut>⌘B</DropdownMenuShortcut>
              </Link>
            </DropdownMenuItem>
            <DropdownMenuItem asChild>
              <Link href={buildProjectUrl(currentProject.name, currentProject.id, 'settings')}>
                Settings
                <DropdownMenuShortcut>⌘S</DropdownMenuShortcut>
              </Link>
            </DropdownMenuItem>
            <DropdownMenuItem>New Team</DropdownMenuItem>
          </DropdownMenuGroup>
        )}
        <DropdownMenuSeparator />
        <DropdownMenuItem
          onClick={() => handleLogout()}
          disabled={isLoggingOut}
        >
          <LogOut className="mr-2 h-4 w-4" />
          {isLoggingOut ? 'Logging out...' : 'Log out'}
          <DropdownMenuShortcut>⇧⌘Q</DropdownMenuShortcut>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
