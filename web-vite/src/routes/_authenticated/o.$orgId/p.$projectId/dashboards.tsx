import { createFileRoute, Outlet } from '@tanstack/react-router'

// Dashboards shell. `dashboards/index.tsx` renders the list;
// `$dashboardId.tsx` renders the static widget viewer. Each child
// installs its own `validateSearch` so the parent list's search
// params (page/limit/q) do not cascade into links targeting the
// detail route. Mirrors `traces.tsx` / `sessions.tsx`.
export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/dashboards',
)({
  component: DashboardsShell,
})

function DashboardsShell() {
  return <Outlet />
}
