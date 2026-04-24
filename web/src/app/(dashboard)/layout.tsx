import { cookies } from 'next/headers'

import { getCurrentUser } from '@/lib/auth/dal'
import { DashboardLayoutClient } from './dashboard-layout-client'

/**
 * Dashboard root layout — async Server Component.
 *
 * Runs once per cold load on the Next.js server (Node runtime).
 * Fetches the authenticated user + organization tree from the Go
 * backend BEFORE the HTML is sent to the browser, so the dashboard
 * shell renders fully populated. No client-side bootstrap fetch,
 * no "Loading workspace…" spinner, no race between auth-store
 * initialization and WorkspaceProvider's useQuery.
 *
 * Auth invariants:
 *   - proxy.ts (route guard) redirects cookieless requests to /signin
 *     before they reach this layout.
 *   - getCurrentUser() (DAL) re-redirects on backend 401 (cookie
 *     present but invalid/expired — devtools clear, server-side
 *     rotation, etc.).
 *   - If both pass, the layout is reached only with a valid session;
 *     the client never has to render an unauthenticated state.
 */
export default async function DashboardLayout({
  children,
}: {
  children: React.ReactNode
}) {
  // Run the sidebar-state cookie read and the DAL fetch in parallel.
  // The cookie read is local (no I/O); pairing it with Promise.all
  // costs nothing and reads more naturally than sequential awaits.
  const [cookieStore, profile] = await Promise.all([
    cookies(),
    getCurrentUser(),
  ])

  const sidebarCookie = cookieStore.get('sidebar_state')
  const defaultOpen = sidebarCookie?.value !== 'false'

  return (
    <DashboardLayoutClient
      defaultOpen={defaultOpen}
      initialUser={profile.user}
      initialOrganizations={profile.organizations}
    >
      {children}
    </DashboardLayoutClient>
  )
}
