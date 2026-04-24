import { createFileRoute, Link } from '@tanstack/react-router'
import { useSuspenseQuery } from '@tanstack/react-query'
import { z } from 'zod'
import { BrokleError } from '@/lib/api/errors'
import { Button } from '@/components/ui/button'
import { EvaluatorsTable } from '@/features/evaluators/components'
import { evaluatorListQueryOptions } from '@/features/evaluators/api/queries'

const searchSchema = z.object({
  page: z.number().int().min(1).catch(1),
  limit: z.number().int().min(1).max(100).catch(20),
  q: z.string().optional().catch(undefined),
})

export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/evaluators',
)({
  validateSearch: searchSchema,
  loaderDeps: ({ search }) => ({
    page: search.page,
    limit: search.limit,
    q: search.q,
  }),
  loader: ({ params, context, deps }) =>
    context.queryClient.ensureQueryData(
      evaluatorListQueryOptions(params.projectId, {
        page: deps.page,
        limit: deps.limit,
        q: deps.q,
      }),
    ),
  errorComponent: EvaluatorsErrorBoundary,
  component: EvaluatorsPage,
})

function EvaluatorsErrorBoundary({ error }: { error: Error }) {
  if (error instanceof BrokleError && error.status >= 400 && error.status < 500) {
    return (
      <main className="mx-auto max-w-5xl p-6">
        <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-6">
          <p className="text-sm font-medium text-destructive">
            Unable to load evaluators
          </p>
          <p className="mt-1 text-sm text-muted-foreground">{error.message}</p>
        </div>
      </main>
    )
  }
  throw error
}

function EvaluatorsPage() {
  const { orgId, projectId } = Route.useParams()
  const search = Route.useSearch()
  const { data } = useSuspenseQuery(
    evaluatorListQueryOptions(projectId, {
      page: search.page,
      limit: search.limit,
      q: search.q,
    }),
  )

  const { data: rows, total, page, limit } = data
  const totalPages = Math.max(1, Math.ceil(total / Math.max(1, limit)))
  const hasPrev = page > 1
  const hasNext = page < totalPages

  return (
    <main className="mx-auto max-w-7xl p-6 space-y-4">
      <header className="flex items-baseline justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Evaluators</h1>
          <p className="text-sm text-muted-foreground">
            {total.toLocaleString()} total
          </p>
        </div>
      </header>

      <EvaluatorsTable rows={rows} />

      <nav className="flex items-center justify-between">
        <p className="text-xs text-muted-foreground">
          Page {page} of {totalPages}
        </p>
        <div className="flex gap-2">
          <Button asChild variant="outline" size="sm" disabled={!hasPrev}>
            <Link
              to="/o/$orgId/p/$projectId/evaluators"
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
          <Button asChild variant="outline" size="sm" disabled={!hasNext}>
            <Link
              to="/o/$orgId/p/$projectId/evaluators"
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
