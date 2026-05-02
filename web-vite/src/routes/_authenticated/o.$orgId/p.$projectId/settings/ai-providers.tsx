import { createFileRoute } from '@tanstack/react-router'
import { useSuspenseQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { z } from 'zod'
import { BrokleError } from '@/lib/api/errors'
import { Button } from '@/components/ui/button'
import { AIProvidersTable } from '@/features/ai-providers/components'
import { aiProvidersListQueryOptions } from '@/features/ai-providers/api/queries'

// Empty search schema — decouples this route from any parent search
// cascade while leaving headroom for future filtering.
const searchSchema = z.object({}).catch({})

export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/settings/ai-providers',
)({
  validateSearch: searchSchema,
  // Credentials are org-scoped (shared across the org's projects), so
  // we prefetch by orgId, not projectId.
  loader: ({ params, context }) =>
    context.queryClient.ensureQueryData(
      aiProvidersListQueryOptions(params.orgId),
    ),
  errorComponent: AIProvidersErrorBoundary,
  component: AIProvidersPage,
})

function AIProvidersErrorBoundary({ error }: { error: Error }) {
  if (
    error instanceof BrokleError &&
    error.status >= 400 &&
    error.status < 500
  ) {
    return (
      <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-6">
        <p className="text-sm font-medium text-destructive">
          Unable to load AI providers
        </p>
        <p className="mt-1 text-sm text-muted-foreground">{error.message}</p>
      </div>
    )
  }
  throw error
}

function AIProvidersPage() {
  const { orgId } = Route.useParams()
  const [addOpen, setAddOpen] = useState(false)
  const { data } = useSuspenseQuery(aiProvidersListQueryOptions(orgId))

  return (
    <section className="space-y-4">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h2 className="text-lg font-semibold">AI providers</h2>
          <p className="text-sm text-muted-foreground">
            {data.length.toLocaleString()} configured. Shared across all
            projects in this organization.
          </p>
        </div>
        <Button onClick={() => setAddOpen(true)}>Add provider</Button>
      </div>

      <AIProvidersTable
        orgId={orgId}
        addDialogOpen={addOpen}
        onAddDialogOpenChange={setAddOpen}
      />
    </section>
  )
}
