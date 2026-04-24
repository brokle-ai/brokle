import { createFileRoute, Outlet, redirect } from '@tanstack/react-router'

// Pathless layout. `beforeLoad` runs before every route nested under
// `/_authenticated/…`. We do NOT hit the network here — the guard reads
// the Zustand snapshot injected into the router context via
// `context.auth.isAuthenticated`. The first page load performs the
// `/v1/users/me` fetch via `useCurrentUser`; on 401 the HTTP
// middleware's auth-retry path runs, and if refresh fails the store's
// `expireSession()` fires and the next navigation bounces here.
export const Route = createFileRoute('/_authenticated')({
  beforeLoad: ({ context, location }) => {
    if (!context.auth.isAuthenticated) {
      throw redirect({
        to: '/signin',
        search: { redirect: location.href },
      })
    }
  },
  component: () => <Outlet />,
})
