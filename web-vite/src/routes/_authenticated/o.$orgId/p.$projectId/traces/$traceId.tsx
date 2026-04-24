import { createFileRoute, Link } from '@tanstack/react-router'
import { useSuspenseQuery } from '@tanstack/react-query'
import { z } from 'zod'
import { ArrowLeft } from 'lucide-react'
import { BrokleError } from '@/lib/api/errors'
import { Button } from '@/components/ui/button'
import { TraceDetail } from '@/features/traces/components'
import { AnnotationsDrawer } from '@/features/traces/components/annotations-drawer'
import { CommentsDrawer } from '@/features/traces/components/comments-drawer'
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
    <main className="mx-auto max-w-7xl space-y-6 p-6">
      <Button asChild variant="ghost" size="sm" className="gap-1 px-2">
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
          <ArrowLeft className="h-4 w-4" />
          Back to traces
        </Link>
      </Button>

      <TraceDetail
        trace={trace}
        spans={spans}
        headerActions={
          <>
            <AnnotationsDrawer projectId={projectId} traceId={traceId} />
            <CommentsDrawer projectId={projectId} traceId={traceId} />
          </>
        }
      />
    </main>
  )
}
