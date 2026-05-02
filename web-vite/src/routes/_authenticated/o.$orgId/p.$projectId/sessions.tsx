import { createFileRoute, Outlet } from '@tanstack/react-router'

// Sessions shell. `sessions/index.tsx` renders the list; `$sessionId.tsx`
// renders the detail view. Each child installs its own `validateSearch`
// so the parent list's search params (page/limit/q) do not cascade into
// links targeting the detail route. Mirrors `traces.tsx`.
export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/sessions',
)({
  component: SessionsShell,
})

function SessionsShell() {
  return <Outlet />
}
