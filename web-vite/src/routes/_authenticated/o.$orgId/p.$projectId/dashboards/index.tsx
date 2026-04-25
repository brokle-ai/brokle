import { createFileRoute, Link, useNavigate } from '@tanstack/react-router'
import { useSuspenseQuery } from '@tanstack/react-query'
import { z } from 'zod'
import { LayoutGrid, Table as TableIcon } from 'lucide-react'
import { BrokleError } from '@/lib/api/errors'
import { Button } from '@/components/ui/button'
import {
  DashboardsTable,
  DashboardsCardList,
  DashboardsProvider,
  CreateDashboardDialog,
} from '@/features/dashboards/components'
import { dashboardListQueryOptions } from '@/features/dashboards/api/queries'
import { useLocalStorage } from '@/hooks/use-local-storage'

const searchSchema = z.object({
  page: z.number().int().min(1).catch(1),
  limit: z.number().int().min(1).max(100).catch(20),
  q: z.string().optional().catch(undefined),
})

export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/dashboards/',
)({
  validateSearch: searchSchema,
  loaderDeps: ({ search }) => ({
    page: search.page,
    limit: search.limit,
    q: search.q,
  }),
  loader: ({ params, context, deps }) =>
    context.queryClient.ensureQueryData(
      dashboardListQueryOptions(params.projectId, {
        page: deps.page,
        limit: deps.limit,
        q: deps.q,
      }),
    ),
  errorComponent: DashboardsErrorBoundary,
  component: DashboardsPage,
})

function DashboardsErrorBoundary({ error }: { error: Error }) {
  if (
    error instanceof BrokleError &&
    error.status >= 400 &&
    error.status < 500
  ) {
    return (
      <main className="mx-auto max-w-5xl p-6">
        <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-6">
          <p className="text-sm font-medium text-destructive">
            Unable to load dashboards
          </p>
          <p className="mt-1 text-sm text-muted-foreground">{error.message}</p>
        </div>
      </main>
    )
  }
  throw error
}

type DashboardsView = 'table' | 'cards'

function isDashboardsView(value: unknown): value is DashboardsView {
  return value === 'table' || value === 'cards'
}

function DashboardsPage() {
  const { orgId, projectId } = Route.useParams()
  const search = Route.useSearch()
  const navigate = useNavigate({ from: Route.fullPath })
  const [storedView, setStoredView] = useLocalStorage<DashboardsView>(
    'dashboards.view',
    'cards',
  )
  const view: DashboardsView = isDashboardsView(storedView)
    ? storedView
    : 'cards'

  const { data, isFetching } = useSuspenseQuery(
    dashboardListQueryOptions(projectId, {
      page: search.page,
      limit: search.limit,
      q: search.q,
    }),
  )

  // Backend envelope is offset-based: `{dashboards, total, limit, offset}`.
  // Project it onto the same page/hasPrev/hasNext shape the route layer
  // expects.
  const rows = data.dashboards
  const { total, limit } = data
  const totalPages = Math.max(1, Math.ceil(total / Math.max(1, limit)))
  const page = search.page
  const hasPrev = page > 1
  const hasNext = page < totalPages

  const handleSearchChange = (value: string) => {
    void navigate({
      search: (prev) => ({
        ...prev,
        page: 1,
        q: value.length > 0 ? value : undefined,
      }),
      replace: true,
    })
  }

  return (
    <DashboardsProvider projectId={projectId} orgId={orgId}>
      <main className="mx-auto max-w-7xl p-6 space-y-4">
        <header className="flex items-baseline justify-between">
          <div>
            <h1 className="text-2xl font-semibold">Dashboards</h1>
            <p className="text-sm text-muted-foreground">
              {total.toLocaleString()} total
            </p>
          </div>
          <div className="flex items-center gap-2">
            <div className="inline-flex rounded-md border bg-background p-0.5">
              <Button
                type="button"
                variant={view === 'cards' ? 'secondary' : 'ghost'}
                size="sm"
                className="h-7 px-2"
                onClick={() => setStoredView('cards')}
                aria-label="Card view"
                aria-pressed={view === 'cards'}
              >
                <LayoutGrid className="h-4 w-4" />
              </Button>
              <Button
                type="button"
                variant={view === 'table' ? 'secondary' : 'ghost'}
                size="sm"
                className="h-7 px-2"
                onClick={() => setStoredView('table')}
                aria-label="Table view"
                aria-pressed={view === 'table'}
              >
                <TableIcon className="h-4 w-4" />
              </Button>
            </div>
            <CreateDashboardDialog projectId={projectId} />
          </div>
        </header>

        {view === 'table' ? (
          <DashboardsTable
            rows={rows}
            renderNameLink={(dashboard, children) => (
              <Link
                to="/o/$orgId/p/$projectId/dashboards/$dashboardId"
                params={{ orgId, projectId, dashboardId: dashboard.id }}
                className="hover:underline"
              >
                {children}
              </Link>
            )}
          />
        ) : (
          <DashboardsCardList
            data={rows}
            totalCount={total}
            searchValue={search.q ?? ''}
            onSearchChange={handleSearchChange}
            isFetching={isFetching}
          />
        )}

        <nav className="flex items-center justify-between">
          <p className="text-xs text-muted-foreground">
            Page {page} of {totalPages}
          </p>
          <div className="flex gap-2">
            <Button asChild variant="outline" size="sm" disabled={!hasPrev}>
              <Link
                to="/o/$orgId/p/$projectId/dashboards"
                params={{ orgId, projectId }}
                search={{
                  page: Math.max(1, page - 1),
                  limit: search.limit,
                  q: search.q,
                }}
              >
                Previous
              </Link>
            </Button>
            <Button asChild variant="outline" size="sm" disabled={!hasNext}>
              <Link
                to="/o/$orgId/p/$projectId/dashboards"
                params={{ orgId, projectId }}
                search={{
                  page: page + 1,
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
    </DashboardsProvider>
  )
}
