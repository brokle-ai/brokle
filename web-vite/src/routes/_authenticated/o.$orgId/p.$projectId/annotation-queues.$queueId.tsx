import { createFileRoute } from '@tanstack/react-router'
import { z } from 'zod'
import { BrokleError } from '@/lib/api/errors'
import { AnnotationQueueDetail } from '@/features/annotation-queues/components'
import {
  queueDetailQueryOptions,
  queueItemsListQueryOptions,
} from '@/features/annotation-queues/api/queries'

// Detail route inherits the parent list route's `validateSearch`
// (`page`/`limit`/`status: 'active'|'paused'|'archived'|undefined`/`q`) —
// TanStack Router resolves nested routes through prefix matching, not
// directory nesting, so `annotation-queues.tsx` is the parent. We reuse
// the parent's page/limit, don't carry `q` into the detail URL, and
// rename our item-level status filter to `itemStatus` so it doesn't
// collide with the parent's queue-level `status` field.
const searchSchema = z.object({
  page: z.number().int().min(1).catch(1),
  limit: z.number().int().min(1).max(100).catch(20),
  itemStatus: z
    .enum(['pending', 'completed', 'skipped'])
    .optional()
    .catch(undefined),
})

export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/annotation-queues/$queueId',
)({
  validateSearch: searchSchema,
  loaderDeps: ({ search }) => ({
    page: search.page,
    limit: search.limit,
    itemStatus: search.itemStatus,
  }),
  loader: async ({ params, context, deps }) => {
    await Promise.all([
      context.queryClient.ensureQueryData(
        queueDetailQueryOptions(params.projectId, params.queueId),
      ),
      context.queryClient.ensureQueryData(
        queueItemsListQueryOptions(params.projectId, params.queueId, {
          page: deps.page,
          limit: deps.limit,
          status: deps.itemStatus,
        }),
      ),
    ])
  },
  errorComponent: QueueDetailErrorBoundary,
  component: QueueDetailPage,
})

function QueueDetailErrorBoundary({ error }: { error: Error }) {
  if (error instanceof BrokleError && error.status >= 400 && error.status < 500) {
    return (
      <main className="mx-auto max-w-5xl p-6">
        <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-6">
          <p className="text-sm font-medium text-destructive">
            Unable to load annotation queue
          </p>
          <p className="mt-1 text-sm text-muted-foreground">{error.message}</p>
        </div>
      </main>
    )
  }
  throw error
}

function QueueDetailPage() {
  const { orgId, projectId, queueId } = Route.useParams()
  const search = Route.useSearch()
  return (
    <AnnotationQueueDetail
      orgId={orgId}
      projectId={projectId}
      queueId={queueId}
      page={search.page}
      limit={search.limit}
      status={search.itemStatus}
    />
  )
}
