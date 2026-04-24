import { createRootRouteWithContext, Outlet } from '@tanstack/react-router'
import { TanStackRouterDevtools } from '@tanstack/react-router-devtools'
import type { QueryClient } from '@tanstack/react-query'
import { Suspense } from 'react'

// Router context passed into every route via `context` on
// createRouter(). `beforeLoad` guards read this at navigation time
// without touching the network.
export interface RouterContext {
  queryClient: QueryClient
  auth: {
    readonly isAuthenticated: boolean
    readonly userId: string | null
  }
}

export const Route = createRootRouteWithContext<RouterContext>()({
  component: RootLayout,
})

function RootLayout() {
  return (
    <>
      <Outlet />
      {import.meta.env.DEV && (
        <Suspense fallback={null}>
          <TanStackRouterDevtools position="bottom-right" />
        </Suspense>
      )}
    </>
  )
}
