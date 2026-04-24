import { createFileRoute, Link } from '@tanstack/react-router'
import { useSuspenseQuery } from '@tanstack/react-query'
import { z } from 'zod'
import { BrokleError } from '@/lib/api/errors'
import { Button } from '@/components/ui/button'
import { PromptDetail } from '@/features/prompts/components'
import {
  promptDetailQueryOptions,
  promptVersionsQueryOptions,
} from '@/features/prompts/api/queries'

// Explicit empty validateSearch decouples this leaf from the parent
// `/prompts` search schema (page/limit/q) that would otherwise cascade
// down the nested route tree.
const searchSchema = z.object({}).catch({})

export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/prompts/$promptId/',
)({
  validateSearch: searchSchema,
  loader: ({ params, context }) =>
    Promise.all([
      context.queryClient.ensureQueryData(
        promptDetailQueryOptions(params.projectId, params.promptId),
      ),
      context.queryClient.ensureQueryData(
        promptVersionsQueryOptions(params.projectId, params.promptId),
      ),
    ]),
  errorComponent: PromptDetailErrorBoundary,
  component: PromptDetailPage,
})

function PromptDetailErrorBoundary({ error }: { error: Error }) {
  if (error instanceof BrokleError && error.status >= 400 && error.status < 500) {
    return (
      <main className="mx-auto max-w-5xl p-6">
        <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-6">
          <p className="text-sm font-medium text-destructive">
            Unable to load prompt
          </p>
          <p className="mt-1 text-sm text-muted-foreground">{error.message}</p>
        </div>
      </main>
    )
  }
  throw error
}

function PromptDetailPage() {
  const { orgId, projectId, promptId } = Route.useParams()
  const { data: prompt } = useSuspenseQuery(
    promptDetailQueryOptions(projectId, promptId),
  )
  const { data: versions } = useSuspenseQuery(
    promptVersionsQueryOptions(projectId, promptId),
  )

  return (
    <main className="mx-auto max-w-7xl p-6 space-y-4">
      <Button asChild variant="ghost" size="sm" className="-ml-2">
        <Link
          to="/o/$orgId/p/$projectId/prompts"
          params={{ orgId, projectId }}
          search={{ page: 1, limit: 20, q: undefined }}
        >
          ← Back to prompts
        </Link>
      </Button>
      <PromptDetail
        orgId={orgId}
        projectId={projectId}
        prompt={prompt}
        versions={versions}
      />
    </main>
  )
}
