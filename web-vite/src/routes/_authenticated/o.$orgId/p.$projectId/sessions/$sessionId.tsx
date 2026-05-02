import { createFileRoute, Link } from '@tanstack/react-router'
import { useSuspenseQuery } from '@tanstack/react-query'
import { z } from 'zod'
import { ArrowLeft } from 'lucide-react'
import { BrokleError } from '@/lib/api/errors'
import { Button } from '@/components/ui/button'
import {
  SessionDetailHeader,
} from '@/features/sessions/components'
import { sessionDetailQueryOptions } from '@/features/sessions/api/queries'
import { TracesTable } from '@/features/traces/components'
import { traceListQueryOptions } from '@/features/traces/api/queries'

// Explicit empty validateSearch decouples this leaf from the parent
// `/sessions` list search schema (page/limit/q) that would otherwise
// cascade down and force every `<Link to=".../$sessionId">` to restate
// the full filter state. Pattern mirrors traces/$traceId.tsx.
const searchSchema = z.object({}).catch({})

// Fixed page size for the session-scoped traces panel. The detail
// view isn't a paginated surface — it's a summary + recent-traces
// slice. Users go to /traces for full pagination.
const TRACES_PAGE_SIZE = 50

export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/sessions/$sessionId',
)({
  validateSearch: searchSchema,
  loader: ({ params, context }) =>
    Promise.all([
      context.queryClient.ensureQueryData(
        sessionDetailQueryOptions(params.projectId, params.sessionId),
      ),
      context.queryClient.ensureQueryData(
        traceListQueryOptions(params.projectId, {
          page: 1,
          limit: TRACES_PAGE_SIZE,
          range: 'all',
          sessionId: params.sessionId,
        }),
      ),
    ]),
  errorComponent: SessionDetailErrorBoundary,
  component: SessionDetailPage,
})

function SessionDetailErrorBoundary({ error }: { error: Error }) {
  if (
    error instanceof BrokleError &&
    error.status >= 400 &&
    error.status < 500
  ) {
    return (
      <main className="mx-auto max-w-5xl p-6">
        <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-6">
          <p className="text-sm font-medium text-destructive">
            Unable to load session
          </p>
          <p className="mt-1 text-sm text-muted-foreground">{error.message}</p>
        </div>
      </main>
    )
  }
  throw error
}

function SessionDetailPage() {
  const { orgId, projectId, sessionId } = Route.useParams()
  const { data: session } = useSuspenseQuery(
    sessionDetailQueryOptions(projectId, sessionId),
  )
  const { data: traceData } = useSuspenseQuery(
    traceListQueryOptions(projectId, {
      page: 1,
      limit: TRACES_PAGE_SIZE,
      range: 'all',
      sessionId,
    }),
  )

  return (
    <main className="mx-auto max-w-7xl space-y-6 p-6">
      <Button asChild variant="ghost" size="sm" className="gap-1 px-2">
        <Link
          to="/o/$orgId/p/$projectId/sessions"
          params={{ orgId, projectId }}
          search={{
            page: 1,
            limit: 20,
            q: undefined,
          }}
        >
          <ArrowLeft className="h-4 w-4" />
          Back to sessions
        </Link>
      </Button>

      <SessionDetailHeader session={session} />

      <section className="space-y-2">
        <h2 className="text-sm font-medium">
          Traces{' '}
          <span className="text-muted-foreground">
            ({traceData.pagination.total.toLocaleString()})
          </span>
        </h2>
        <TracesTable
          rows={traceData.data}
          renderNameLink={(trace, children) => (
            <Link
              to="/o/$orgId/p/$projectId/traces/$traceId"
              params={{ orgId, projectId, traceId: trace.trace_id }}
              className="hover:underline"
            >
              {children}
            </Link>
          )}
        />
      </section>
    </main>
  )
}
