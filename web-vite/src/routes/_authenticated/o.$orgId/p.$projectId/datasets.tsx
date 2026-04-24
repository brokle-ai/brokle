import { createFileRoute, Link } from '@tanstack/react-router'
import { useSuspenseQuery } from '@tanstack/react-query'
import { z } from 'zod'
import { BrokleError } from '@/lib/api/errors'
import { Button } from '@/components/ui/button'
import { DatasetsTable } from '@/features/datasets/components'
import { datasetListQueryOptions } from '@/features/datasets/api/queries'

// Zod-validated search params. `.catch` keeps a hostile URL from
// throwing the whole route — invalid values fall back to the default.
// Backend honours a 1–100 limit band on the pageList endpoints.
const searchSchema = z.object({
  page: z.number().int().min(1).catch(1),
  limit: z.number().int().min(1).max(100).catch(20),
  q: z.string().optional().catch(undefined),
})

export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/datasets',
)({
  validateSearch: searchSchema,
  loaderDeps: ({ search }) => ({
    page: search.page,
    limit: search.limit,
    q: search.q,
  }),
  loader: ({ params, context, deps }) =>
    context.queryClient.ensureQueryData(
      datasetListQueryOptions(params.projectId, {
        page: deps.page,
        limit: deps.limit,
        q: deps.q,
      }),
    ),
  errorComponent: DatasetsErrorBoundary,
  component: DatasetsPage,
})

function DatasetsErrorBoundary({ error }: { error: Error }) {
  if (error instanceof BrokleError && error.status >= 400 && error.status < 500) {
    return (
      <main className="mx-auto max-w-5xl p-6">
        <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-6">
          <p className="text-sm font-medium text-destructive">
            Unable to load datasets
          </p>
          <p className="mt-1 text-sm text-muted-foreground">{error.message}</p>
        </div>
      </main>
    )
  }
  throw error
}

function DatasetsPage() {
  const { orgId, projectId } = Route.useParams()
  const search = Route.useSearch()
  const { data } = useSuspenseQuery(
    datasetListQueryOptions(projectId, {
      page: search.page,
      limit: search.limit,
      q: search.q,
    }),
  )

  const { data: rows, total, page, limit } = data
  // Evaluation list envelope is flat `{data, total, page, limit}` — unlike
  // observability's `pagination: {...}`. Derive page-boundary state here.
  const totalPages = Math.max(1, Math.ceil(total / Math.max(1, limit)))
  const hasPrev = page > 1
  const hasNext = page < totalPages

  return (
    <main className="mx-auto max-w-7xl p-6 space-y-4">
      <header className="flex items-baseline justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Datasets</h1>
          <p className="text-sm text-muted-foreground">
            {total.toLocaleString()} total
          </p>
        </div>
      </header>

      <DatasetsTable rows={rows} />

      <nav className="flex items-center justify-between">
        <p className="text-xs text-muted-foreground">
          Page {page} of {totalPages}
        </p>
        <div className="flex gap-2">
          <Button asChild variant="outline" size="sm" disabled={!hasPrev}>
            <Link
              to="/o/$orgId/p/$projectId/datasets"
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
              to="/o/$orgId/p/$projectId/datasets"
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
