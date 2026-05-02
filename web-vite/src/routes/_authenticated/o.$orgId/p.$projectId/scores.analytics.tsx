import { createFileRoute, Link } from '@tanstack/react-router'
import { z } from 'zod'
import { Button } from '@/components/ui/button'
import { ScoreAnalyticsDashboard } from '@/features/scores/components'

// Decouple from the parent /scores search-param cascade — analytics
// has its own state (primary score, comparison score, interval) which
// is held in the dashboard's local state, so the search schema accepts
// any URL params and discards them. `passthrough` keeps the inferred
// type from blocking unknown keys (which would otherwise come through
// as `[x: string]: never` and break Link-with-search type-checking).
const searchSchema = z.object({}).passthrough().catch({})

export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/scores/analytics',
)({
  validateSearch: searchSchema,
  component: ScoreAnalyticsPage,
})

function ScoreAnalyticsPage() {
  const { orgId, projectId } = Route.useParams()

  return (
    <main className="mx-auto max-w-7xl space-y-6 p-6">
      <header className="flex items-baseline justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Score Analytics</h1>
          <p className="text-sm text-muted-foreground">
            Aggregate statistics, distributions, and time-series across the
            project's recorded scores.
          </p>
        </div>
        <Button asChild variant="outline" size="sm">
          <Link
            to="/o/$orgId/p/$projectId/scores"
            params={{ orgId, projectId }}
            search={{ page: 1, limit: 20 }}
          >
            Back to scores
          </Link>
        </Button>
      </header>

      <ScoreAnalyticsDashboard projectId={projectId} />
    </main>
  )
}
