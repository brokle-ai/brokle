import { useState } from 'react'
import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { useSuspenseQuery } from '@tanstack/react-query'
import { z } from 'zod'
import { RefreshCw } from 'lucide-react'
import { BrokleError } from '@/lib/api/errors'
import { Button } from '@/components/ui/button'
import { projectMembershipQueryOptions } from '@/features/projects/queries'
import { organizationMembershipQueryOptions } from '@/features/organizations/queries'
import {
  CostByModelChart,
  OnboardingChecklist,
  RecentTracesTable,
  ScoreOverview,
  StatsRow,
  TopErrorsTable,
  TraceVolumeChart,
} from '@/features/overview/components'
import { overviewQueryOptions } from '@/features/overview/api/queries'
import type { OverviewTimeRange } from '@/features/overview/api/types'
import { TimeRangePicker } from '@/components/shared/time-range-picker'
import type { TimeRange } from '@/components/shared/time-range-picker'
import { cn } from '@/lib/utils'

// Valid `time_range` presets on the backend. `custom` is present in
// the picker shape but the overview API only accepts the enum below;
// when the user picks "custom" we fall back to the default for now.
const TIME_RANGE_VALUES = [
  '15m',
  '30m',
  '1h',
  '3h',
  '6h',
  '12h',
  '24h',
  '7d',
  '14d',
  '30d',
  'all',
] as const satisfies readonly OverviewTimeRange[]

const searchSchema = z
  .object({
    range: z.enum(TIME_RANGE_VALUES).optional().catch(undefined),
  })
  .catch({})

export const Route = createFileRoute('/_authenticated/o/$orgId/p/$projectId/')({
  validateSearch: searchSchema,
  loaderDeps: ({ search }) => ({ range: search.range ?? '24h' }),
  loader: ({ params, context, deps }) =>
    context.queryClient.ensureQueryData(
      overviewQueryOptions(params.projectId, { timeRange: deps.range }),
    ),
  errorComponent: OverviewErrorBoundary,
  component: ProjectHome,
})

function OverviewErrorBoundary({ error }: { error: Error }) {
  if (
    error instanceof BrokleError &&
    error.status >= 400 &&
    error.status < 500
  ) {
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
  const search = Route.useSearch()
  const range: OverviewTimeRange = search.range ?? '24h'
  const navigate = useNavigate({ from: Route.fullPath })
  const { data: org } = useSuspenseQuery(
    organizationMembershipQueryOptions(orgId),
  )
  const { data: project } = useSuspenseQuery(
    projectMembershipQueryOptions(projectId),
  )
  const { data: overview, isFetching, refetch } = useSuspenseQuery(
    overviewQueryOptions(projectId, { timeRange: range }),
  )

  const [checklistDismissed, setChecklistDismissed] = useState(false)

  const handleTimeRangeChange = (next: TimeRange) => {
    // Custom ranges aren't plumbed through the overview API yet —
    // collapse to the default preset. The picker still round-trips
    // preset selections correctly for dashboards.
    const rel = next.relative
    const nextRange: OverviewTimeRange =
      rel && rel !== 'custom' && (TIME_RANGE_VALUES as readonly string[]).includes(rel)
        ? (rel as OverviewTimeRange)
        : '24h'
    navigate({ search: { range: nextRange } })
  }

  return (
    <main className="mx-auto max-w-7xl p-6 space-y-6">
      <header className="flex items-start justify-between gap-4">
        <div className="space-y-1">
          <p className="text-xs uppercase tracking-wide text-muted-foreground">
            {org.name}
          </p>
          <h1 className="text-2xl font-semibold">{project.name}</h1>
          <p className="text-sm text-muted-foreground">Overview</p>
        </div>
        <div className="flex items-center gap-2">
          <TimeRangePicker
            value={{ relative: range }}
            onChange={handleTimeRangeChange}
          />
          <Button
            variant="outline"
            size="icon"
            onClick={() => refetch()}
            disabled={isFetching}
          >
            <RefreshCw
              className={cn('h-4 w-4', isFetching && 'animate-spin')}
            />
          </Button>
        </div>
      </header>

      {!checklistDismissed && (
        <OnboardingChecklist
          checklistStatus={overview.checklist_status}
          orgId={orgId}
          projectId={projectId}
          onDismiss={() => setChecklistDismissed(true)}
        />
      )}

      <StatsRow stats={overview.stats} />

      <div className="grid gap-4 lg:grid-cols-2">
        <TraceVolumeChart data={overview.trace_volume} />
        <CostByModelChart data={overview.cost_by_model} />
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        <RecentTracesTable
          orgId={orgId}
          projectId={projectId}
          data={overview.recent_traces}
        />
        <TopErrorsTable data={overview.top_errors} />
      </div>

      <ScoreOverview
        data={overview.scores_summary}
        orgId={orgId}
        projectId={projectId}
      />
    </main>
  )
}
