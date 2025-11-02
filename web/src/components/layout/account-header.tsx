import { Header } from '@/components/layout/header'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { cn } from '@/lib/utils'

interface AccountHeaderProps {
  className?: string
  title?: string
}

/**
 * AccountHeader - Reusable header for account/settings pages
 *
 * Features:
 * - NO organization/project switchers
 * - Optional title display
 * - Mobile responsive (Theme/Profile hidden on small screens)
 * - Search component (icon on mobile, full bar on desktop)
 *
 * Usage:
 * ```tsx
 * <AccountHeader title="Account Settings" />
 * <AccountHeader title="Profile" />
 * <AccountHeader />  // No title
 * ```
 */
export function AccountHeader({ className, title }: AccountHeaderProps) {
  return (
    <Header className={className}>
      {title && <h2 className="text-lg font-semibold">{title}</h2>}
      <div className='ml-auto flex items-center space-x-4'>
        <Search />
        <ThemeSwitch className="hidden sm:flex" />
        <ProfileDropdown className="hidden sm:flex" />
      </div>
    </Header>
  )
}
