import { createFileRoute } from '@tanstack/react-router'
import { z } from 'zod'
import { BrokleError } from '@/lib/api/errors'
import { ExecutionDetailView } from '@/features/evaluators/components'
import { executionDetailQueryOptions } from '@/features/evaluators/api/queries'

// Decouple the leaf from the parent search-param cascade — `.catch({})`
// accepts any URL search and ignores it. Even though the parent
// `evaluators.tsx` declares `page`/`limit`/`q`, TanStack Router's
// type machinery resolves the leaf's own schema independently when
// `validateSearch` is explicit, so callers can `<Link to="…executions/$executionId" search={{}} />`.
const searchSchema = z.object({}).passthrough().catch({})

export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/evaluators/$evaluatorId/executions/$executionId',
)({
  validateSearch: searchSchema,
  loader: ({ params, context }) =>
    context.queryClient.ensureQueryData(
      executionDetailQueryOptions(
        params.projectId,
        params.evaluatorId,
        params.executionId,
      ),
    ),
  errorComponent: ExecutionDetailErrorBoundary,
  component: ExecutionDetailPage,
})

function ExecutionDetailErrorBoundary({ error }: { error: Error }) {
  if (error instanceof BrokleError && error.status >= 400 && error.status < 500) {
    return (
      <main className="mx-auto max-w-5xl p-6">
        <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-6">
          <p className="text-sm font-medium text-destructive">
            Unable to load execution
          </p>
          <p className="mt-1 text-sm text-muted-foreground">{error.message}</p>
        </div>
      </main>
    )
  }
  throw error
}

function ExecutionDetailPage() {
  const { orgId, projectId, evaluatorId, executionId } = Route.useParams()
  return (
    <ExecutionDetailView
      orgId={orgId}
      projectId={projectId}
      evaluatorId={evaluatorId}
      executionId={executionId}
    />
  )
}
