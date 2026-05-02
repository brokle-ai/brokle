import { type ReactNode } from 'react'
import { act, render, type RenderResult } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import {
  RouterProvider,
  createRootRoute,
  createRoute,
  createRouter,
  createMemoryHistory,
  Outlet,
} from '@tanstack/react-router'

// Lightweight test harness for components that depend on TanStack
// Router `<Link>` + React Query mutations. Mounts the component on a
// dummy `/` route inside a memory-history router; sibling routes for
// `/forgot-password`, `/signup`, etc. are stubbed so any `<Link>`
// resolves without errors.

export async function renderWithProviders(
  ui: ReactNode,
): Promise<RenderResult> {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  })

  const rootRoute = createRootRoute({ component: () => <Outlet /> })
  const indexRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/',
    component: () => <>{ui}</>,
  })
  const stubs = [
    '/signin',
    '/signup',
    '/forgot-password',
    '/reset-password',
    '/verify-email',
  ].map((path) =>
    createRoute({
      getParentRoute: () => rootRoute,
      path,
      component: () => null,
    }),
  )

  const routeTree = rootRoute.addChildren([indexRoute, ...stubs])
  const router = createRouter({
    routeTree,
    history: createMemoryHistory({ initialEntries: ['/'] }),
    // Mirror production wiring (main.tsx) — components reach the
    // shared QueryClient via `router.options.context.queryClient`.
    context: { queryClient, auth: { isAuthenticated: false, userId: null } },
  })

  // Resolve the initial match before render so the index route's
  // component (= the UI under test) is available synchronously after
  // the first render commit.
  await router.load()

  let result!: RenderResult
  await act(async () => {
    result = render(
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
      </QueryClientProvider>,
    )
  })
  return result
}
