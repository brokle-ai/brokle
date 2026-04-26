import { createRootRouteWithContext, Outlet } from '@tanstack/react-router'
import { TanStackRouterDevtools } from '@tanstack/react-router-devtools'
import type { QueryClient } from '@tanstack/react-query'
import { NuqsAdapter } from 'nuqs/adapters/tanstack-router'
import { Suspense } from 'react'
import { RootErrorFallback } from '@/components/layout/root-error-fallback'

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
  // Global error boundary mirrored from Opik / SigNoz. Catches any
  // beforeLoad / loader / render error from anywhere in the tree
  // that isn't handled by a closer per-route `errorComponent`.
  errorComponent: ({ error, reset }) => (
    <RootErrorFallback error={error} onRetry={reset} />
  ),
})

function RootLayout() {
  // NuqsAdapter mounts inside the router (not in main.tsx) because
  // `nuqs/adapters/tanstack-router` uses `useLocation`/`useRouter`
  // hooks that require the <RouterProvider> context. Wrapping the
  // Outlet here puts every descendant route inside the adapter so
  // every `useQueryStates`/`useQueryState` consumer (dashboards,
  // datasets, prompts) resolves without per-feature setup.
  return (
    <NuqsAdapter>
      <Outlet />
      {import.meta.env.DEV && (
        <Suspense fallback={null}>
          <TanStackRouterDevtools position="bottom-right" />
        </Suspense>
      )}
    </NuqsAdapter>
  )
}
