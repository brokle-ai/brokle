import { cookies } from 'next/headers'
import { DashboardLayoutClient } from './dashboard-layout-client'

export default async function DashboardLayout({
  children,
}: {
  children: React.ReactNode
}) {
  // Read sidebar state from cookie (server-side) to avoid hydration mismatch
  const cookieStore = await cookies()
  const sidebarCookie = cookieStore.get('sidebar_state')
  const defaultOpen = sidebarCookie?.value !== 'false'

  return <DashboardLayoutClient defaultOpen={defaultOpen}>{children}</DashboardLayoutClient>
}
