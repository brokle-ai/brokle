import { createFileRoute, Link, useNavigate } from '@tanstack/react-router'
import { useSuspenseQuery } from '@tanstack/react-query'
import { z } from 'zod'
import { BrokleError } from '@/lib/api/errors'
import { Button } from '@/components/ui/button'
import {
  ScoresFilterBar,
  ScoresTable,
  type ScoresFilterValue,
} from '@/features/scores/components'
import { scoreListQueryOptions } from '@/features/scores/api/queries'

// Zod-validated search params. `.catch` keeps a hostile URL from
// throwing the whole route — invalid values fall back to the default.
// Backend param names (checked against
// `internal/transport/http/handlers/observability/dashboard_types.go`
// scoreFilterQuery): name, source, type, trace_id, span_id. On the
// URL we keep `traceId`/`spanId` camelCase to match React search
// conventions; the query fetcher translates to snake_case at the
// HTTP boundary.
const searchSchema = z.object({
  page: z.number().int().min(1).catch(1),
  limit: z.number().int().min(1).max(100).catch(20),
  name: z.string().optional().catch(undefined),
  source: z.enum(['code', 'llm', 'human']).optional().catch(undefined),
  type: z
    .enum(['NUMERIC', 'CATEGORICAL', 'BOOLEAN'])
    .optional()
    .catch(undefined),
  traceId: z.string().optional().catch(undefined),
  spanId: z.string().optional().catch(undefined),
})

export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/scores',
)({
  validateSearch: searchSchema,
  loaderDeps: ({ search }) => ({
    page: search.page,
    limit: search.limit,
    name: search.name,
    source: search.source,
    type: search.type,
    traceId: search.traceId,
    spanId: search.spanId,
  }),
  loader: ({ params, context, deps }) =>
    context.queryClient.ensureQueryData(
      scoreListQueryOptions(params.projectId, {
        page: deps.page,
        limit: deps.limit,
        name: deps.name,
        source: deps.source,
        type: deps.type,
        traceId: deps.traceId,
        spanId: deps.spanId,
      }),
    ),
  errorComponent: ScoresErrorBoundary,
  component: ScoresPage,
})

function ScoresErrorBoundary({ error }: { error: Error }) {
  if (
    error instanceof BrokleError &&
    error.status >= 400 &&
    error.status < 500
  ) {
    return (
      <main className="mx-auto max-w-5xl p-6">
        <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-6">
          <p className="text-sm font-medium text-destructive">
            Unable to load scores
          </p>
          <p className="mt-1 text-sm text-muted-foreground">{error.message}</p>
        </div>
      </main>
    )
  }
  throw error
}

function ScoresPage() {
  const { orgId, projectId } = Route.useParams()
  const search = Route.useSearch()
  const navigate = useNavigate({ from: Route.fullPath })

  const { data } = useSuspenseQuery(
    scoreListQueryOptions(projectId, {
      page: search.page,
      limit: search.limit,
      name: search.name,
      source: search.source,
      type: search.type,
      traceId: search.traceId,
      spanId: search.spanId,
    }),
  )

  // Observability-style envelope: `{data, pagination: {page, limit,
  // total, total_pages}}`. No has_next/has_prev on the wire — derive.
  const { data: rows, pagination } = data
  const hasPrev = pagination.page > 1
  const hasNext = pagination.page < pagination.total_pages

  const filterValue: ScoresFilterValue = {
    name: search.name,
    source: search.source,
    type: search.type,
    traceId: search.traceId,
    spanId: search.spanId,
  }

  const handleFilterChange = (next: ScoresFilterValue) => {
    // Any filter change resets the pager to page 1.
    navigate({
      search: {
        page: 1,
        limit: search.limit,
        name: next.name,
        source: next.source,
        type: next.type,
        traceId: next.traceId,
        spanId: next.spanId,
      },
    })
  }

  return (
    <main className="mx-auto max-w-7xl p-6 space-y-4">
      <header className="flex items-baseline justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Scores</h1>
          <p className="text-sm text-muted-foreground">
            {pagination.total.toLocaleString()} total
          </p>
        </div>
      </header>

      <ScoresFilterBar value={filterValue} onChange={handleFilterChange} />

      <ScoresTable rows={rows} />

      <nav className="flex items-center justify-between">
        <p className="text-xs text-muted-foreground">
          Page {pagination.page} of {Math.max(1, pagination.total_pages)}
        </p>
        <div className="flex gap-2">
          <Button asChild variant="outline" size="sm" disabled={!hasPrev}>
            <Link
              to="/o/$orgId/p/$projectId/scores"
              params={{ orgId, projectId }}
              search={{
                page: Math.max(1, search.page - 1),
                limit: search.limit,
                name: search.name,
                source: search.source,
                type: search.type,
                traceId: search.traceId,
                spanId: search.spanId,
              }}
            >
              Previous
            </Link>
          </Button>
          <Button asChild variant="outline" size="sm" disabled={!hasNext}>
            <Link
              to="/o/$orgId/p/$projectId/scores"
              params={{ orgId, projectId }}
              search={{
                page: search.page + 1,
                limit: search.limit,
                name: search.name,
                source: search.source,
                type: search.type,
                traceId: search.traceId,
                spanId: search.spanId,
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
