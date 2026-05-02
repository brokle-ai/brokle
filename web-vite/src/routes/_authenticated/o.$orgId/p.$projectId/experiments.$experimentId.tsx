import { createFileRoute } from '@tanstack/react-router'
import { z } from 'zod'
import { BrokleError } from '@/lib/api/errors'
import { ExperimentDetail } from '@/features/experiments/components'
import {
  experimentDetailQueryOptions,
  experimentItemsListQueryOptions,
  experimentMetricsQueryOptions,
} from '@/features/experiments/api/queries'

// Experiment items use limit/offset on the backend (predates pageList),
// so we mirror that on the search params. `.catch` keeps hostile URLs
// from throwing.
const searchSchema = z.object({
  limit: z.number().int().min(1).max(100).catch(20),
  offset: z.number().int().min(0).catch(0),
})

export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/experiments/$experimentId',
)({
  validateSearch: searchSchema,
  loaderDeps: ({ search }) => ({
    limit: search.limit,
    offset: search.offset,
  }),
  loader: async ({ params, context, deps }) => {
    // Parallelise detail + metrics + first page of items so the view
    // renders once on suspense resolution.
    await Promise.all([
      context.queryClient.ensureQueryData(
        experimentDetailQueryOptions(params.projectId, params.experimentId),
      ),
      context.queryClient.ensureQueryData(
        experimentMetricsQueryOptions(params.projectId, params.experimentId),
      ),
      context.queryClient.ensureQueryData(
        experimentItemsListQueryOptions(
          params.projectId,
          params.experimentId,
          { limit: deps.limit, offset: deps.offset },
        ),
      ),
    ])
  },
  errorComponent: ExperimentDetailErrorBoundary,
  component: ExperimentDetailPage,
})

function ExperimentDetailErrorBoundary({ error }: { error: Error }) {
  if (error instanceof BrokleError && error.status >= 400 && error.status < 500) {
    return (
      <main className="mx-auto max-w-5xl p-6">
        <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-6">
          <p className="text-sm font-medium text-destructive">
            Unable to load experiment
          </p>
          <p className="mt-1 text-sm text-muted-foreground">{error.message}</p>
        </div>
      </main>
    )
  }
  throw error
}

function ExperimentDetailPage() {
  const { orgId, projectId, experimentId } = Route.useParams()
  const search = Route.useSearch()
  return (
    <ExperimentDetail
      orgId={orgId}
      projectId={projectId}
      experimentId={experimentId}
      limit={search.limit}
      offset={search.offset}
    />
  )
}
