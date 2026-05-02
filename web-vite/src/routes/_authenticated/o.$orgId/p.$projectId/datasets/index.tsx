import { createFileRoute } from '@tanstack/react-router'
import { BrokleError } from '@/lib/api/errors'
import { Datasets } from '@/features/datasets'

// List route. The ported `<Datasets>` component owns its own query via
// `useProjectDatasets` which reads search params off the route via
// `useSearch({ strict: false })`; we keep the validateSearch schema
// permissive so legacy URLs from web/ don't throw.
export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/datasets/',
)({
  errorComponent: DatasetsErrorBoundary,
  component: DatasetsPage,
})

function DatasetsErrorBoundary({ error }: { error: Error }) {
  if (error instanceof BrokleError && error.status >= 400 && error.status < 500) {
    return (
      <main className="mx-auto max-w-5xl p-6">
        <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-6">
          <p className="text-sm font-medium text-destructive">
            Unable to load datasets
          </p>
          <p className="mt-1 text-sm text-muted-foreground">{error.message}</p>
        </div>
      </main>
    )
  }
  throw error
}

function DatasetsPage() {
  const { orgId, projectId } = Route.useParams()
  // `projectSlug` in the ported feature is the web/-style composite
  // slug. In web-vite routes it's just the two URL segments joined so
  // the value is stable across re-renders — nothing interpolates it
  // into a URL anymore.
  return <Datasets projectSlug={`${orgId}/${projectId}`} />
}
