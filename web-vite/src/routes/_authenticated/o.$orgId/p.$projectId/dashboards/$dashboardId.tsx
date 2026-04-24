import { createFileRoute } from '@tanstack/react-router'
import { z } from 'zod'
import { BrokleError } from '@/lib/api/errors'
import { DashboardDetail } from '@/features/dashboards/components'
import { dashboardDetailQueryOptions } from '@/features/dashboards/api/queries'

// Empty validateSearch decouples this leaf from the parent
// `/dashboards` list search schema (page/limit/q) so links targeting
// the detail route need not restate filter state. The dashboard
// component manages its own time-range + variable URL params via
// nuqs — keep `catch({})` permissive so those don't fail validation.
const searchSchema = z.object({}).catch({})

export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/dashboards/$dashboardId',
)({
  validateSearch: searchSchema,
  loader: ({ params, context }) =>
    context.queryClient.ensureQueryData(
      dashboardDetailQueryOptions(params.projectId, params.dashboardId),
    ),
  errorComponent: DashboardDetailErrorBoundary,
  component: DashboardDetailPage,
})

function DashboardDetailErrorBoundary({ error }: { error: Error }) {
  if (
    error instanceof BrokleError &&
    error.status >= 400 &&
    error.status < 500
  ) {
    return (
      <main className="mx-auto max-w-5xl p-6">
        <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-6">
          <p className="text-sm font-medium text-destructive">
            Unable to load dashboard
          </p>
          <p className="mt-1 text-sm text-muted-foreground">{error.message}</p>
        </div>
      </main>
    )
  }
  throw error
}

function DashboardDetailPage() {
  const { orgId, projectId, dashboardId } = Route.useParams()

  return (
    <main className="flex flex-col">
      <DashboardDetail
        orgId={orgId}
        projectId={projectId}
        dashboardId={dashboardId}
      />
    </main>
  )
}
