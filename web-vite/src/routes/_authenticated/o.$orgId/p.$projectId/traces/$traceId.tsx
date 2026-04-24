import { createFileRoute } from '@tanstack/react-router'
import { useSuspenseQuery } from '@tanstack/react-query'
import { z } from 'zod'
import { BrokleError } from '@/lib/api/errors'
import { TraceDetail } from '@/features/traces/components'
import {
  traceDetailQueryOptions,
  traceSpansQueryOptions,
} from '@/features/traces/api/queries'

// Explicit empty validateSearch decouples this leaf from the parent
// `/traces` list search schema (page/limit/q/status/range/model) that
// would otherwise cascade down and force every `<Link to=".../$traceId">`
// to restate the full filter state.
const searchSchema = z.object({}).catch({})

export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/traces/$traceId',
)({
  validateSearch: searchSchema,
  loader: ({ params, context }) =>
    Promise.all([
      context.queryClient.ensureQueryData(
        traceDetailQueryOptions(params.traceId),
      ),
      context.queryClient.ensureQueryData(
        traceSpansQueryOptions(params.traceId),
      ),
    ]),
  errorComponent: TraceDetailErrorBoundary,
  component: TraceDetailPage,
})

function TraceDetailErrorBoundary({ error }: { error: Error }) {
  if (error instanceof BrokleError && error.status >= 400 && error.status < 500) {
    return (
      <main className="mx-auto max-w-5xl p-6">
        <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-6">
          <p className="text-sm font-medium text-destructive">
            Unable to load trace
          </p>
          <p className="mt-1 text-sm text-muted-foreground">{error.message}</p>
        </div>
      </main>
    )
  }
  throw error
}

function TraceDetailPage() {
  const { orgId, projectId, traceId } = Route.useParams()
  const { data: trace } = useSuspenseQuery(traceDetailQueryOptions(traceId))
  const { data: spans } = useSuspenseQuery(traceSpansQueryOptions(traceId))

  return (
    <TraceDetail
      trace={trace}
      spans={spans}
      orgId={orgId}
      projectId={projectId}
    />
  )
}
