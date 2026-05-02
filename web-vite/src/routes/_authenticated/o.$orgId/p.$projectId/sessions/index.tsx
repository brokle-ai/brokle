import { createFileRoute, Link } from '@tanstack/react-router'
import { useSuspenseQuery } from '@tanstack/react-query'
import { z } from 'zod'
import { BrokleError } from '@/lib/api/errors'
import { Button } from '@/components/ui/button'
import { SessionsTable } from '@/features/sessions/components'
import { sessionListQueryOptions } from '@/features/sessions/api/queries'

// Zod-validated search params. `.catch` keeps a hostile URL from
// throwing the whole route — invalid values fall back to the default.
// Backend caps limit at 100 (SessionFilter.SetDefaults, session_entity.go).
const searchSchema = z.object({
  page: z.number().int().min(1).catch(1),
  limit: z.number().int().min(1).max(100).catch(20),
  q: z.string().optional().catch(undefined),
})

export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/sessions/',
)({
  validateSearch: searchSchema,
  loaderDeps: ({ search }) => ({
    page: search.page,
    limit: search.limit,
    q: search.q,
  }),
  loader: ({ params, context, deps }) =>
    context.queryClient.ensureQueryData(
      sessionListQueryOptions(params.projectId, {
        page: deps.page,
        limit: deps.limit,
        q: deps.q,
      }),
    ),
  errorComponent: SessionsErrorBoundary,
  component: SessionsPage,
})

function SessionsErrorBoundary({ error }: { error: Error }) {
  if (error instanceof BrokleError && error.status >= 400 && error.status < 500) {
    return (
      <main className="mx-auto max-w-5xl p-6">
        <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-6">
          <p className="text-sm font-medium text-destructive">
            Unable to load sessions
          </p>
          <p className="mt-1 text-sm text-muted-foreground">{error.message}</p>
        </div>
      </main>
    )
  }
  throw error
}

function SessionsPage() {
  const { orgId, projectId } = Route.useParams()
  const search = Route.useSearch()
  const { data } = useSuspenseQuery(
    sessionListQueryOptions(projectId, {
      page: search.page,
      limit: search.limit,
      q: search.q,
    }),
  )

  const { data: rows, pagination } = data

  return (
    <main className="mx-auto max-w-7xl p-6 space-y-4">
      <header className="flex items-baseline justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Sessions</h1>
          <p className="text-sm text-muted-foreground">
            {pagination.total.toLocaleString()} total
          </p>
        </div>
      </header>

      <SessionsTable
        rows={rows}
        projectId={projectId}
        renderSessionLink={(session, children) => (
          <Link
            to="/o/$orgId/p/$projectId/sessions/$sessionId"
            params={{ orgId, projectId, sessionId: session.session_id }}
            className="hover:underline"
            onClick={(e) => e.stopPropagation()}
          >
            {children}
          </Link>
        )}
        renderDetailLink={(session, children) => (
          <Link
            to="/o/$orgId/p/$projectId/sessions/$sessionId"
            params={{ orgId, projectId, sessionId: session.session_id }}
            className="inline-flex"
            onClick={(e) => e.stopPropagation()}
            aria-label="Open session detail"
          >
            {children}
          </Link>
        )}
        renderTraceLink={(traceId, children) => (
          <Link
            to="/o/$orgId/p/$projectId/traces/$traceId"
            params={{ orgId, projectId, traceId }}
            className="hover:underline"
          >
            {children}
          </Link>
        )}
      />

      <nav className="flex items-center justify-between">
        <p className="text-xs text-muted-foreground">
          Page {pagination.page} of {Math.max(1, pagination.total_pages)}
        </p>
        <div className="flex gap-2">
          <Button asChild variant="outline" size="sm" disabled={!pagination.has_prev}>
            <Link
              to="/o/$orgId/p/$projectId/sessions"
              params={{ orgId, projectId }}
              search={{
                page: Math.max(1, search.page - 1),
                limit: search.limit,
                q: search.q,
              }}
            >
              Previous
            </Link>
          </Button>
          <Button asChild variant="outline" size="sm" disabled={!pagination.has_next}>
            <Link
              to="/o/$orgId/p/$projectId/sessions"
              params={{ orgId, projectId }}
              search={{
                page: search.page + 1,
                limit: search.limit,
                q: search.q,
              }}
            >
              Next
            </Link>
          </Button>
        </div>
      </nav>
    </main>
  )
}
