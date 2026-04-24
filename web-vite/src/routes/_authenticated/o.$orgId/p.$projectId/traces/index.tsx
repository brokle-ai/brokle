import { useEffect, useState } from 'react'
import { createFileRoute, Link, useNavigate } from '@tanstack/react-router'
import { useSuspenseQuery } from '@tanstack/react-query'
import { z } from 'zod'
import { BrokleError } from '@/lib/api/errors'
import {
  TracesPeekView,
  TracesTable,
  type TracesFilterValue,
  type TracesSortKey,
} from '@/features/traces/components'
import {
  TracesProvider,
  useTraces,
} from '@/features/traces/context/traces-context'
import { AddTraceToDatasetDialog } from '@/features/datasets'
import { traceListQueryOptions } from '@/features/traces/api/queries'
import type { TraceListItem } from '@/features/traces/api/types'

/**
 * Zod-validated search params. `.catch(...)` keeps a hostile URL from
 * throwing the whole route — invalid values fall back to the default.
 * Limit is clamped to the band the backend honours (see
 * `internal/transport/http/handlers/observability` — the traces list
 * endpoint caps at 100).
 *
 * The schema grew with the table port: sort keys and client-side facet
 * arrays are now URL-persisted so back/forward restores both the
 * server-side sort and the user's narrowed faceted selection.
 */
const sortKeySchema = z.enum([
  'duration',
  'model_name',
  'total_cost',
  'total_tokens',
  'span_count',
  'start_time',
])

const searchSchema = z.object({
  page: z.number().int().min(1).catch(1),
  limit: z.number().int().min(1).max(100).catch(20),
  q: z.string().optional().catch(undefined),
  status: z.enum(['ok', 'error', 'unset']).optional().catch(undefined),
  range: z.enum(['24h', '7d', '30d', 'all']).catch('all'),
  model: z.string().optional().catch(undefined),
  sort_by: sortKeySchema.optional().catch(undefined),
  sort_dir: z.enum(['asc', 'desc']).optional().catch(undefined),
  // Peek-sheet URL state — both optional, invalid values drop. The
  // peek sheet mounts only while `peek` is set; `peekTab` is a soft
  // hint for which tab to open (defaults to 'io' in the hook).
  peek: z.string().optional().catch(undefined),
  peekTab: z
    .enum(['io', 'tree', 'metadata', 'scores', 'comments'])
    .optional()
    .catch(undefined),
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
    sort_by: search.sort_by,
    sort_dir: search.sort_dir,
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
  return (
    <TracesProvider>
      <TracesPageInner />
    </TracesProvider>
  )
}

function TracesPageInner() {
  const { orgId, projectId } = Route.useParams()
  const search = Route.useSearch()
  const navigate = useNavigate({ from: Route.fullPath })
  const { setCurrentPageTraceIds, setCurrentPageTraces } = useTraces()

  const { data, isFetching } = useSuspenseQuery(
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

  const [addToDatasetTrace, setAddToDatasetTrace] =
    useState<TraceListItem | null>(null)

  const handleFilterChange = (next: TracesFilterValue) => {
    navigate({
      search: (prev) => ({
        ...prev,
        page: 1,
        limit: search.limit,
        q: next.q,
        status: next.status,
        range: next.range,
        model: next.model,
      }),
    })
  }

  const handlePaginationChange = (page: number, limit: number) => {
    navigate({
      search: (prev) => ({ ...prev, page, limit }),
    })
  }

  const handleSortChange = (
    sortBy: TracesSortKey | null,
    sortDir: 'asc' | 'desc' | null,
  ) => {
    navigate({
      search: (prev) => ({
        ...prev,
        sort_by: sortBy ?? undefined,
        sort_dir: sortDir ?? undefined,
      }),
    })
  }

  // Full-page navigation — used by the name Link and the row-actions
  // "View" menu item. The row click handler below opens the peek sheet
  // instead, matching the "peek by default, full page opt-in" UX.
  const handleViewDetail = (trace: TraceListItem) => {
    void navigate({
      to: '/o/$orgId/p/$projectId/traces/$traceId',
      params: { orgId, projectId, traceId: trace.trace_id },
    })
  }

  // Row click → open peek in the overlay. The peek hooks read the
  // `peek` search param from the same route search schema we declared
  // above; the navigate here only touches `peek`/`peekTab`.
  const handleOpenPeek = (trace: TraceListItem) => {
    void navigate({
      search: (prev) => ({ ...prev, peek: trace.trace_id, peekTab: undefined }),
    })
  }

  // Keep the peek navigator's `currentPageTraceIds` in sync with the
  // rows currently rendered — feeds prev/next chevrons in the header.
  useEffect(() => {
    setCurrentPageTraceIds(rows.map((r) => r.trace_id))
    setCurrentPageTraces(rows)
  }, [rows, setCurrentPageTraceIds, setCurrentPageTraces])

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
        server={{
          pagination,
          filterValue,
          onFilterChange: handleFilterChange,
          onPaginationChange: handlePaginationChange,
          sortBy: search.sort_by ?? null,
          sortDir: search.sort_dir ?? null,
          onSortChange: handleSortChange,
          isFetching,
          onViewDetail: handleViewDetail,
          onAddToDataset: setAddToDatasetTrace,
          onRowClick: handleOpenPeek,
        }}
      />

      <TracesPeekView orgId={orgId} projectId={projectId} />

      {addToDatasetTrace && (
        <AddTraceToDatasetDialog
          projectId={projectId}
          traceId={addToDatasetTrace.trace_id}
          traceName={addToDatasetTrace.name}
          open={addToDatasetTrace !== null}
          onOpenChange={(open) => {
            if (!open) setAddToDatasetTrace(null)
          }}
        />
      )}
    </main>
  )
}
