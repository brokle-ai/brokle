import { createFileRoute, Link } from '@tanstack/react-router'
import { useSuspenseQuery } from '@tanstack/react-query'
import { z } from 'zod'
import { BrokleError } from '@/lib/api/errors'
import { Button } from '@/components/ui/button'
import { TracesTable } from '@/features/traces/components'
import { traceListQueryOptions } from '@/features/traces/api/queries'

// Zod-validated search params. `.catch` keeps a hostile URL from
// throwing the whole route — invalid values fall back to the default.
// Limit is clamped to the band the backend honours (see
// internal/transport/http/handlers/observability — the traces list
// endpoint caps at 100).
const searchSchema = z.object({
  page: z.number().int().min(1).catch(1),
  limit: z.number().int().min(1).max(100).catch(20),
  q: z.string().optional().catch(undefined),
})

export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/traces',
)({
  validateSearch: searchSchema,
  // `loaderDeps` teaches TanStack Router which search params feed the
  // loader — changing `page` / `limit` / `q` re-runs the loader
  // (and therefore the prefetch), changing unrelated params does not.
  loaderDeps: ({ search }) => ({
    page: search.page,
    limit: search.limit,
    q: search.q,
  }),
  loader: ({ params, context, deps }) =>
    context.queryClient.ensureQueryData(
      traceListQueryOptions(params.projectId, {
        page: deps.page,
        limit: deps.limit,
        q: deps.q,
      }),
    ),
  errorComponent: TracesErrorBoundary,
  component: TracesPage,
})

function TracesErrorBoundary({ error }: { error: Error }) {
  // Render inline for client-remediable 4xx (validation, not-found).
  // Let 401/403 and 5xx bubble — the auth store handles 401, the root
  // error boundary handles the rest.
  if (error instanceof BrokleError && error.status >= 400 && error.status < 500) {
    return (
      <main className="mx-auto max-w-5xl p-6">
        <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-6">
          <p className="text-sm font-medium text-destructive">
            Unable to load traces
          </p>
          <p className="mt-1 text-sm text-muted-foreground">{error.message}</p>
        </div>
      </main>
    )
  }
  throw error
}

function TracesPage() {
  const { orgId, projectId } = Route.useParams()
  const search = Route.useSearch()
  const { data } = useSuspenseQuery(
    traceListQueryOptions(projectId, {
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
          <h1 className="text-2xl font-semibold">Traces</h1>
          <p className="text-sm text-muted-foreground">
            {pagination.total.toLocaleString()} total
          </p>
        </div>
      </header>

      <TracesTable rows={rows} />

      <nav className="flex items-center justify-between">
        <p className="text-xs text-muted-foreground">
          Page {pagination.page} of {Math.max(1, pagination.total_pages)}
        </p>
        <div className="flex gap-2">
          <Button asChild variant="outline" size="sm" disabled={!pagination.has_prev}>
            <Link
              to="/o/$orgId/p/$projectId/traces"
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
              to="/o/$orgId/p/$projectId/traces"
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
