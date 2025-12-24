import { Header } from '@/components/layout/header'
import { ContextNavbar } from '@/components/layout/context-navbar'
import { Search } from '@/components/search'

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
 * - Search component (icon on mobile, full bar on desktop)
 *
 * Note: Theme and Profile settings are in the sidebar user dropdown (NavUser)
 */
export function DashboardHeader({ className }: DashboardHeaderProps) {
  return (
    <Header className={className}>
      <ContextNavbar />
      <div className='ml-auto flex items-center space-x-4'>
        <Search />
      </div>
    </Header>
  )
}
