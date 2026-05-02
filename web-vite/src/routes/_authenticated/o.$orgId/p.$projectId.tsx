import { createFileRoute, Navigate } from '@tanstack/react-router'
import { NotFoundError } from '@/lib/api/errors'
import { projectMembershipQueryOptions } from '@/features/projects/queries'
import { organizationListQueryOptions } from '@/features/organizations/queries'
import { AuthenticatedLayout } from '@/components/layout/authenticated-layout'

// Project tenancy segment — also owns the dashboard chrome (sidebar +
// header) so every nested route renders under the same shell. The org
// list is prefetched here so the sidebar's org switcher renders
// without its own suspense boundary.
export const Route = createFileRoute('/_authenticated/o/$orgId/p/$projectId')({
  loader: ({ params, context }) =>
    Promise.all([
      context.queryClient.ensureQueryData(
        projectMembershipQueryOptions(params.projectId),
      ),
      context.queryClient.ensureQueryData(organizationListQueryOptions()),
    ]),
  errorComponent: ({ error }) => {
    if (error instanceof NotFoundError) {
      return <Navigate to="/o/$orgId" params={{ orgId: Route.useParams().orgId }} />
    }
    return (
      <main className="flex min-h-screen items-center justify-center px-4">
        <p className="text-sm text-red-600">Unable to load project: {error.message}</p>
      </main>
    )
  },
  component: AuthenticatedLayout,
})
