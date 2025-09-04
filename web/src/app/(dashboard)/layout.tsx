import { cookies } from 'next/headers'
import { AuthenticatedLayout } from "@/components/layout/authenticated-layout"
import { getServerSession } from '@/lib/auth/server-auth'

const SIDEBAR_COOKIE_NAME = 'sidebar_state'

export default async function DashboardLayout({
  children,
}: {
  children: React.ReactNode
}) {
  // Get server session (middleware already handled auth redirect)
  const session = await getServerSession()
  
  // Read sidebar state from cookie server-side
  const cookieStore = await cookies()
  const sidebarState = cookieStore.get(SIDEBAR_COOKIE_NAME)
  const defaultSidebarOpen = sidebarState?.value !== 'false'

  return (
    <AuthenticatedLayout 
      defaultSidebarOpen={defaultSidebarOpen}
      serverUser={session.user}
    >
      {children}
    </AuthenticatedLayout>
  )
}