import { createFileRoute } from '@tanstack/react-router'
import { z } from 'zod'
import { BrokleError } from '@/lib/api/errors'
import { ReviewItem } from '@/features/annotation-queues/components'
import { queueDetailQueryOptions } from '@/features/annotation-queues/api/queries'
import { scoreConfigsQueryOptions } from '@/features/scores/api/queries'

// Decouples the review leaf from the parent detail route's
// `page`/`limit`/`itemStatus` search schema. Reviewers arriving via
// `<Link to=".../review">` shouldn't need to restate the parent's
// pagination state, and the review page has no search filters of its
// own.
//
// Why `z.record(z.string(), z.unknown())` instead of the more obvious
// `z.object({}).catch({})`: TanStack Router unions the leaf's search
// shape with every search schema above it in the tree. When the leaf
// is a strictly-typed empty object, the inferred target type becomes
// `{[x:string]: never; page: number; ...}` — a type no concrete value
// satisfies (named keys collide with the never index signature). The
// record-of-unknown here keeps the index signature permissive so
// `<Link>`/`navigate({ to })` can omit the parent's fields without a
// type error. Runtime behaviour is identical: the router discards any
// fields not declared in this schema, so navigating here always
// produces an empty search bag.
const searchSchema = z.record(z.string(), z.unknown()).catch({})

export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/annotation-queues/$queueId/review',
)({
  validateSearch: searchSchema,
  loader: async ({ params, context }) => {
    // Loader fetches queue metadata + the project-wide score-config
    // catalog. The actual "claim next item" call is a side-effecting
    // POST and runs from the component's mount effect (see
    // `review-item.tsx`) — route loaders must be idempotent.
    await Promise.all([
      context.queryClient.ensureQueryData(
        queueDetailQueryOptions(params.projectId, params.queueId),
      ),
      context.queryClient.ensureQueryData(
        scoreConfigsQueryOptions(params.projectId),
      ),
    ])
  },
  errorComponent: ReviewErrorBoundary,
  component: ReviewPage,
})

function ReviewErrorBoundary({ error }: { error: Error }) {
  if (
    error instanceof BrokleError &&
    error.status >= 400 &&
    error.status < 500
  ) {
    return (
      <main className="mx-auto max-w-5xl p-6">
        <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-6">
          <p className="text-sm font-medium text-destructive">
            Unable to start review
          </p>
          <p className="mt-1 text-sm text-muted-foreground">
            {error.message}
          </p>
        </div>
      </main>
    )
  }
  throw error
}

function ReviewPage() {
  const { orgId, projectId, queueId } = Route.useParams()
  return (
    <ReviewItem orgId={orgId} projectId={projectId} queueId={queueId} />
  )
}
