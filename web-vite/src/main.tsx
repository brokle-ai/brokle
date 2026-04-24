import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { QueryClientProvider } from '@tanstack/react-query'
import { RouterProvider, createRouter } from '@tanstack/react-router'

import { routeTree } from './routeTree.gen'
import { createAppQueryClient } from '@/lib/api/query-client'
import { useAuthStore } from '@/stores/auth-store'
import { Toaster } from '@/components/ui/sonner'
import './index.css'

const queryClient = createAppQueryClient()

const router = createRouter({
  routeTree,
  context: {
    queryClient,
    // Filled in by <RouterWrapper> via router.update before RouterProvider
    // mounts so `beforeLoad` sees the current snapshot on first load.
    auth: { isAuthenticated: false, userId: null },
  },
  defaultPreload: 'intent',
  defaultPreloadStaleTime: 0,
})

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}

// Subscribe the router's context to the auth store so
// `beforeLoad({ context })` reads fresh values without re-creating the
// router instance on every state change.
useAuthStore.subscribe((state) => {
  router.update({
    context: {
      queryClient,
      auth: {
        isAuthenticated: state.isAuthenticated,
        userId: state.user?.id ?? null,
      },
    },
  })
})

const rootEl = document.getElementById('root')
if (!rootEl) throw new Error('#root missing from index.html')

createRoot(rootEl).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
      <Toaster position="top-right" richColors closeButton />
    </QueryClientProvider>
  </StrictMode>,
)
