import { createFileRoute } from '@tanstack/react-router'
import { z } from 'zod'
import { ExperimentCompareView } from '@/features/experiments/components'

// The compare view owns its own search schema, completely decoupled
// from the sibling list/detail/new routes. `.catch()` guards against
// hostile URL shapes.
const searchSchema = z.object({
  ids: z.string().optional().catch(undefined),
  baseline: z.string().optional().catch(undefined),
  view: z.enum(['card', 'table']).optional().catch(undefined),
})

export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/experiments/compare',
)({
  validateSearch: searchSchema,
  component: ExperimentComparePage,
})

function ExperimentComparePage() {
  const { orgId, projectId } = Route.useParams()
  const search = Route.useSearch()

  const experimentIds = search.ids
    ? search.ids.split(',').filter(Boolean)
    : []
  const baselineId = search.baseline
  const viewMode = search.view ?? 'card'

  return (
    <main className="mx-auto max-w-7xl p-6">
      <ExperimentCompareView
        orgId={orgId}
        projectId={projectId}
        experimentIds={experimentIds}
        baselineId={baselineId}
        viewMode={viewMode}
      />
    </main>
  )
}
