import { createFileRoute, Outlet, Navigate } from '@tanstack/react-router'
import { NotFoundError } from '@/lib/api/errors'
import { projectMembershipQueryOptions } from '@/features/projects/queries'

// Project tenancy segment. Validates project access; NotFoundError
// bounces back to the project picker for the org.
export const Route = createFileRoute('/_authenticated/o/$orgId/p/$projectId')({
  loader: ({ params, context }) =>
    context.queryClient.ensureQueryData(
      projectMembershipQueryOptions(params.projectId),
    ),
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
  component: () => <Outlet />,
})
