import { createFileRoute, Outlet } from '@tanstack/react-router'

// Traces shell. The index route (`traces/index.tsx`) renders the list
// + filter bar; `$traceId.index.tsx` renders the detail view. Each
// child installs its own `validateSearch` so the parent list's search
// params (page/limit/q/status/range/model) do not cascade into links
// targeting the detail route.
export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/traces',
)({
  component: TracesShell,
})

function TracesShell() {
  return <Outlet />
}
