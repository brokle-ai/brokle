import { createFileRoute, Link } from '@tanstack/react-router'
import { useSuspenseQuery } from '@tanstack/react-query'
import { z } from 'zod'
import { BrokleError } from '@/lib/api/errors'
import { Button } from '@/components/ui/button'
import { ApiKeysTable } from '@/features/api-keys/components'
import { apiKeyListQueryOptions } from '@/features/api-keys/api/queries'

// Backend caps limit at 100 (see handlers/apikey/handlers.go
// parsePagination / pkg/pagination.IsValidPageSize).
const searchSchema = z.object({
  page: z.number().int().min(1).catch(1),
  limit: z.number().int().min(1).max(100).catch(25),
})

export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/settings/api-keys',
)({
  validateSearch: searchSchema,
  loaderDeps: ({ search }) => ({ page: search.page, limit: search.limit }),
  loader: ({ params, context, deps }) =>
    context.queryClient.ensureQueryData(
      apiKeyListQueryOptions(params.projectId, {
        page: deps.page,
        limit: deps.limit,
      }),
    ),
  errorComponent: ApiKeysErrorBoundary,
  component: ApiKeysPage,
})

function ApiKeysErrorBoundary({ error }: { error: Error }) {
  if (
    error instanceof BrokleError &&
    error.status >= 400 &&
    error.status < 500
  ) {
    return (
      <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-6">
        <p className="text-sm font-medium text-destructive">
          Unable to load API keys
        </p>
        <p className="mt-1 text-sm text-muted-foreground">{error.message}</p>
      </div>
    )
  }
  throw error
}

function ApiKeysPage() {
  const { orgId, projectId } = Route.useParams()
  const search = Route.useSearch()
  const { data } = useSuspenseQuery(
    apiKeyListQueryOptions(projectId, {
      page: search.page,
      limit: search.limit,
    }),
  )

  const { data: rows, pagination } = data

  return (
    <section className="space-y-4">
      <div>
        <h2 className="text-lg font-semibold">API keys</h2>
        <p className="text-sm text-muted-foreground">
          {pagination.total.toLocaleString()} total
        </p>
      </div>

      <ApiKeysTable rows={rows} />

      <nav className="flex items-center justify-between">
        <p className="text-xs text-muted-foreground">
          Page {pagination.page} of {Math.max(1, pagination.total_pages)}
        </p>
        <div className="flex gap-2">
          <Button
            asChild
            variant="outline"
            size="sm"
            disabled={!pagination.has_prev}
          >
            <Link
              to="/o/$orgId/p/$projectId/settings/api-keys"
              params={{ orgId, projectId }}
              search={{
                page: Math.max(1, search.page - 1),
                limit: search.limit,
              }}
            >
              Previous
            </Link>
          </Button>
          <Button
            asChild
            variant="outline"
            size="sm"
            disabled={!pagination.has_next}
          >
            <Link
              to="/o/$orgId/p/$projectId/settings/api-keys"
              params={{ orgId, projectId }}
              search={{ page: search.page + 1, limit: search.limit }}
            >
              Next
            </Link>
          </Button>
        </div>
      </nav>
    </section>
  )
}
