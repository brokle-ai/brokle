import { createFileRoute, Link } from '@tanstack/react-router'
import { useSuspenseQuery } from '@tanstack/react-query'
import { z } from 'zod'
import { BrokleError } from '@/lib/api/errors'
import { Button } from '@/components/ui/button'
import { ScoresTable } from '@/features/scores/components'
import { scoreListQueryOptions } from '@/features/scores/api/queries'

const searchSchema = z.object({
  page: z.number().int().min(1).catch(1),
  limit: z.number().int().min(1).max(100).catch(20),
  name: z.string().optional().catch(undefined),
  source: z.enum(['code', 'llm', 'human']).optional().catch(undefined),
  type: z
    .enum(['NUMERIC', 'CATEGORICAL', 'BOOLEAN'])
    .optional()
    .catch(undefined),
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
  }),
  loader: ({ params, context, deps }) =>
    context.queryClient.ensureQueryData(
      scoreListQueryOptions(params.projectId, {
        page: deps.page,
        limit: deps.limit,
        name: deps.name,
        source: deps.source,
        type: deps.type,
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
  const { data } = useSuspenseQuery(
    scoreListQueryOptions(projectId, {
      page: search.page,
      limit: search.limit,
      name: search.name,
      source: search.source,
      type: search.type,
    }),
  )

  // Observability-style envelope: `{data, pagination: {page, limit,
  // total, total_pages}}`. No has_next/has_prev on the wire — derive.
  const { data: rows, pagination } = data
  const hasPrev = pagination.page > 1
  const hasNext = pagination.page < pagination.total_pages

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
