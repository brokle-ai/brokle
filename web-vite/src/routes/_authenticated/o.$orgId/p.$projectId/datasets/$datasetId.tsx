import { createFileRoute, Outlet } from '@tanstack/react-router'
import { BrokleError } from '@/lib/api/errors'
import { DatasetDetailLayout } from '@/features/datasets'

// Parent route for dataset-detail surfaces. Renders the shared layout
// (header, tabs, dialogs) and delegates the body to the matched child
// route via <Outlet/>. Three children: index (items), versions,
// settings. Each declares its own `validateSearch` (or leaves it
// unset) — no parent search cascade.
export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/datasets/$datasetId',
)({
  errorComponent: DatasetDetailErrorBoundary,
  component: DatasetDetailShell,
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

function DatasetDetailShell() {
  const { orgId, projectId, datasetId } = Route.useParams()
  return (
    <DatasetDetailLayout
      projectSlug={`${orgId}/${projectId}`}
      datasetId={datasetId}
    >
      <Outlet />
    </DatasetDetailLayout>
  )
}
