import { Header } from '@/components/layout/header'
import { ContextNavbar } from '@/components/layout/context-navbar'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { cn } from '@/lib/utils'

interface DashboardHeaderProps {
  className?: string
}

/**
 * DashboardHeader - Reusable header for org and project dashboard pages
 *
 * Features:
 * - ContextNavbar automatically shows:
 *   - Organization selector only (on /organizations/* routes)
 *   - Organization + Project selectors (on /projects/* routes)
 * - Mobile responsive (Theme/Profile hidden on small screens)
 * - Search component (icon on mobile, full bar on desktop)
 *
 * Usage:
 * ```tsx
 * <DashboardHeader />
 * ```
 */
export function DashboardHeader({ className }: DashboardHeaderProps) {
  return (
    <Header className={className}>
      <ContextNavbar />
      <div className='ml-auto flex items-center space-x-4'>
        <Search />
        <ThemeSwitch className="hidden sm:flex" />
        <ProfileDropdown className="hidden sm:flex" />
      </div>
    </Header>
  )
}
