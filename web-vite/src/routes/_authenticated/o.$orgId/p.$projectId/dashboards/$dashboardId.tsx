import { createFileRoute, Link } from '@tanstack/react-router'
import { useSuspenseQuery } from '@tanstack/react-query'
import { z } from 'zod'
import { ArrowLeft } from 'lucide-react'
import { BrokleError } from '@/lib/api/errors'
import { Button } from '@/components/ui/button'
import { DashboardDetail } from '@/features/dashboards/components'
import { dashboardDetailQueryOptions } from '@/features/dashboards/api/queries'

// Empty validateSearch decouples this leaf from the parent
// `/dashboards` list search schema (page/limit/q) so links targeting
// the detail route need not restate filter state. Mirrors
// `traces/$traceId.tsx` / `sessions/$sessionId.tsx`.
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
  const { data: dashboard } = useSuspenseQuery(
    dashboardDetailQueryOptions(projectId, dashboardId),
  )

  return (
    <main className="mx-auto max-w-7xl space-y-6 p-6">
      <Button asChild variant="ghost" size="sm" className="gap-1 px-2">
        <Link
          to="/o/$orgId/p/$projectId/dashboards"
          params={{ orgId, projectId }}
          search={{
            page: 1,
            limit: 20,
            q: undefined,
          }}
        >
          <ArrowLeft className="h-4 w-4" />
          Back to dashboards
        </Link>
      </Button>

      <DashboardDetail dashboard={dashboard} />
    </main>
  )
}
