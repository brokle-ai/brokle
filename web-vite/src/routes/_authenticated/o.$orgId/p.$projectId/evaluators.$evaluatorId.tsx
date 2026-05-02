import { createFileRoute } from '@tanstack/react-router'
import { z } from 'zod'
import { BrokleError } from '@/lib/api/errors'
import { EvaluatorDetail } from '@/features/evaluators/components'
import { evaluatorDetailQueryOptions } from '@/features/evaluators/api/queries'

// Detail route inherits the parent list route's `validateSearch`
// (TanStack Router resolves nested routes through prefix matching, so
// `evaluators.tsx` is the parent of `evaluators.$evaluatorId.tsx`).
// The detail view doesn't use page/limit/q itself, but callers linking
// here must still satisfy the merged type — we keep the parent fields
// on the schema so the Link type-check matches.
const searchSchema = z.object({
  page: z.number().int().min(1).catch(1),
  limit: z.number().int().min(1).max(100).catch(20),
  q: z.string().optional().catch(undefined),
})

export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/evaluators/$evaluatorId',
)({
  validateSearch: searchSchema,
  loader: ({ params, context }) =>
    context.queryClient.ensureQueryData(
      evaluatorDetailQueryOptions(params.projectId, params.evaluatorId),
    ),
  errorComponent: EvaluatorDetailErrorBoundary,
  component: EvaluatorDetailPage,
})

function EvaluatorDetailErrorBoundary({ error }: { error: Error }) {
  if (error instanceof BrokleError && error.status >= 400 && error.status < 500) {
    return (
      <main className="mx-auto max-w-5xl p-6">
        <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-6">
          <p className="text-sm font-medium text-destructive">
            Unable to load evaluator
          </p>
          <p className="mt-1 text-sm text-muted-foreground">{error.message}</p>
        </div>
      </main>
    )
  }
  throw error
}

function EvaluatorDetailPage() {
  const { orgId, projectId, evaluatorId } = Route.useParams()
  return (
    <EvaluatorDetail
      orgId={orgId}
      projectId={projectId}
      evaluatorId={evaluatorId}
    />
  )
}
