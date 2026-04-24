import { createFileRoute, Link, useNavigate } from '@tanstack/react-router'
import { useSuspenseQuery } from '@tanstack/react-query'
import { z } from 'zod'
import { BrokleError } from '@/lib/api/errors'
import { Button } from '@/components/ui/button'
import {
  TracesTable,
  TracesFilterBar,
  type TracesFilterValue,
} from '@/features/traces/components'
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
  status: z.enum(['ok', 'error', 'unset']).optional().catch(undefined),
  // `range` is required — the filter bar always has a concrete value.
  // `.catch('all')` both seeds the default and guards hostile URLs.
  range: z.enum(['24h', '7d', '30d', 'all']).catch('all'),
  model: z.string().optional().catch(undefined),
})

export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/traces/',
)({
  validateSearch: searchSchema,
  loaderDeps: ({ search }) => ({
    page: search.page,
    limit: search.limit,
    q: search.q,
    status: search.status,
    range: search.range,
    model: search.model,
  }),
  loader: ({ params, context, deps }) =>
    context.queryClient.ensureQueryData(
      traceListQueryOptions(params.projectId, {
        page: deps.page,
        limit: deps.limit,
        q: deps.q,
        status: deps.status,
        range: deps.range,
        model: deps.model,
      }),
    ),
  errorComponent: TracesErrorBoundary,
  component: TracesPage,
})

function TracesErrorBoundary({ error }: { error: Error }) {
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
  const navigate = useNavigate({ from: Route.fullPath })

  const { data } = useSuspenseQuery(
    traceListQueryOptions(projectId, {
      page: search.page,
      limit: search.limit,
      q: search.q,
      status: search.status,
      range: search.range,
      model: search.model,
    }),
  )

  const { data: rows, pagination } = data

  const filterValue: TracesFilterValue = {
    q: search.q,
    status: search.status,
    range: search.range,
    model: search.model,
  }

  const handleFilterChange = (next: TracesFilterValue) => {
    // Any filter change resets the pager to page 1.
    navigate({
      search: {
        page: 1,
        limit: search.limit,
        q: next.q,
        status: next.status,
        range: next.range,
        model: next.model,
      },
    })
  }

  return (
    <main className="mx-auto max-w-7xl space-y-4 p-6">
      <header className="flex items-baseline justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Traces</h1>
          <p className="text-sm text-muted-foreground">
            {pagination.total.toLocaleString()} total
          </p>
        </div>
      </header>

      <TracesFilterBar value={filterValue} onChange={handleFilterChange} />

      <TracesTable
        rows={rows}
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
                status: search.status,
                range: search.range,
                model: search.model,
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
                status: search.status,
                range: search.range,
                model: search.model,
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
