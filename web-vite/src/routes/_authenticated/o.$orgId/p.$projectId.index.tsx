import { createFileRoute, Link } from '@tanstack/react-router'
import { useSuspenseQuery } from '@tanstack/react-query'
import { BrokleError } from '@/lib/api/errors'
import { Button } from '@/components/ui/button'
import { projectMembershipQueryOptions } from '@/features/projects/queries'
import { organizationMembershipQueryOptions } from '@/features/organizations/queries'
import {
  CostByModelChart,
  RecentTracesTable,
  StatsRow,
  TopErrorsTable,
  TraceVolumeChart,
} from '@/features/overview/components'
import { overviewQueryOptions } from '@/features/overview/api/queries'

// Project home. Phase 1.5 ships the stats row + a link through to
// the traces list; chart widgets (cost-by-model, trace-volume,
// top-errors) are deferred to the second port.
export const Route = createFileRoute('/_authenticated/o/$orgId/p/$projectId/')({
  loader: ({ params, context }) =>
    context.queryClient.ensureQueryData(
      overviewQueryOptions(params.projectId, { timeRange: '24h' }),
    ),
  errorComponent: OverviewErrorBoundary,
  component: ProjectHome,
})

function OverviewErrorBoundary({ error }: { error: Error }) {
  // 4xx (auth-adjacent, not-found) renders inline; 5xx bubbles to the
  // root boundary where the global 500 UI takes over.
  if (error instanceof BrokleError && error.status >= 400 && error.status < 500) {
    return (
      <main className="mx-auto max-w-5xl p-6">
        <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-6">
          <p className="text-sm font-medium text-destructive">
            Unable to load overview
          </p>
          <p className="mt-1 text-sm text-muted-foreground">{error.message}</p>
        </div>
      </main>
    )
  }
  throw error
}

function ProjectHome() {
  const { orgId, projectId } = Route.useParams()
  const { data: org } = useSuspenseQuery(organizationMembershipQueryOptions(orgId))
  const { data: project } = useSuspenseQuery(projectMembershipQueryOptions(projectId))
  const { data: overview } = useSuspenseQuery(
    overviewQueryOptions(projectId, { timeRange: '24h' }),
  )

  return (
    <main className="mx-auto max-w-7xl p-6 space-y-6">
      <header className="flex items-baseline justify-between">
        <div className="space-y-1">
          <p className="text-xs uppercase tracking-wide text-muted-foreground">
            {org.name}
          </p>
          <h1 className="text-2xl font-semibold">{project.name}</h1>
          <p className="text-sm text-muted-foreground">Last 24 hours</p>
        </div>
        <Button asChild variant="outline" size="sm">
          <Link
            to="/o/$orgId/p/$projectId/traces"
            params={{ orgId, projectId }}
            search={{
              page: 1,
              limit: 20,
              q: undefined,
              status: undefined,
              range: 'all',
              model: undefined,
            }}
          >
            View traces →
          </Link>
        </Button>
      </header>

      <StatsRow stats={overview.stats} />

      <div className="grid gap-4 lg:grid-cols-2">
        <TraceVolumeChart data={overview.trace_volume} />
        <CostByModelChart data={overview.cost_by_model} />
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        <TopErrorsTable data={overview.top_errors} />
        <RecentTracesTable
          orgId={orgId}
          projectId={projectId}
          data={overview.recent_traces}
        />
      </div>
    </main>
  )
}
