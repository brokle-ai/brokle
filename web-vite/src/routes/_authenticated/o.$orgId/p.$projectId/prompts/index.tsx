import { createFileRoute } from '@tanstack/react-router'
import { z } from 'zod'
import { BrokleError } from '@/lib/api/errors'
import { Prompts } from '@/features/prompts'

// Zod-validated search params. `.catch` keeps a hostile URL from
// throwing the whole route — invalid values fall back to the default.
// Kept open (passthrough for unknown keys) because the nuqs-backed
// table state writes `search`, `types`, `sortBy`, etc. into the URL.
const searchSchema = z
  .object({
    page: z.number().int().min(1).optional().catch(undefined),
    pageSize: z.number().int().min(1).max(100).optional().catch(undefined),
    search: z.string().optional().catch(undefined),
    types: z.string().optional().catch(undefined),
    sortBy: z.string().optional().catch(undefined),
    sortOrder: z.string().optional().catch(undefined),
  })
  .catch({})

export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/prompts/',
)({
  validateSearch: searchSchema,
  errorComponent: PromptsErrorBoundary,
  component: PromptsPage,
})

function PromptsErrorBoundary({ error }: { error: Error }) {
  if (
    error instanceof BrokleError &&
    error.status >= 400 &&
    error.status < 500
  ) {
    return (
      <main className="mx-auto max-w-5xl p-6">
        <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-6">
          <p className="text-sm font-medium text-destructive">
            Unable to load prompts
          </p>
          <p className="mt-1 text-sm text-muted-foreground">{error.message}</p>
        </div>
      </main>
    )
  }
  throw error
}

function PromptsPage() {
  const { orgId, projectId } = Route.useParams()
  return (
    <div className="flex h-full flex-col">
      <Prompts orgId={orgId} projectId={projectId} />
    </div>
  )
}
