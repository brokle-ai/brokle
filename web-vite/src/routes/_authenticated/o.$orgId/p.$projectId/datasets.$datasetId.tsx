import { createFileRoute } from '@tanstack/react-router'
import { z } from 'zod'
import { BrokleError } from '@/lib/api/errors'
import { DatasetDetail } from '@/features/datasets/components'
import {
  datasetDetailQueryOptions,
  datasetItemsListQueryOptions,
} from '@/features/datasets/api/queries'

// Detail route. File-based routing nests this under `p.$projectId`
// SIBLING to `datasets.tsx` (not child-of), so no validateSearch
// cascades in from the list route — we declare our own schema.
const searchSchema = z.object({
  page: z.number().int().min(1).catch(1),
  limit: z.number().int().min(1).max(100).catch(20),
})

export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/datasets/$datasetId',
)({
  validateSearch: searchSchema,
  loaderDeps: ({ search }) => ({ page: search.page, limit: search.limit }),
  loader: async ({ params, context, deps }) => {
    // Parallel prefetch — detail + first page of items — so the
    // suspense boundary renders once rather than flickering through
    // two waterfalls.
    await Promise.all([
      context.queryClient.ensureQueryData(
        datasetDetailQueryOptions(params.projectId, params.datasetId),
      ),
      context.queryClient.ensureQueryData(
        datasetItemsListQueryOptions(params.projectId, params.datasetId, {
          page: deps.page,
          limit: deps.limit,
        }),
      ),
    ])
  },
  errorComponent: DatasetDetailErrorBoundary,
  component: DatasetDetailPage,
})

function DatasetDetailErrorBoundary({ error }: { error: Error }) {
  if (error instanceof BrokleError && error.status >= 400 && error.status < 500) {
    return (
      <main className="mx-auto max-w-5xl p-6">
        <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-6">
          <p className="text-sm font-medium text-destructive">
            Unable to load dataset
          </p>
          <p className="mt-1 text-sm text-muted-foreground">{error.message}</p>
        </div>
      </main>
    )
  }
  throw error
}

function DatasetDetailPage() {
  const { orgId, projectId, datasetId } = Route.useParams()
  const search = Route.useSearch()
  return (
    <DatasetDetail
      orgId={orgId}
      projectId={projectId}
      datasetId={datasetId}
      page={search.page}
      limit={search.limit}
    />
  )
}
