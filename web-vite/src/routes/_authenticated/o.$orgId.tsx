import { createFileRoute, Outlet, Navigate } from '@tanstack/react-router'
import { NotFoundError } from '@/lib/api/errors'
import { organizationMembershipQueryOptions } from '@/features/organizations/queries'

// Tenancy segment. Validates that the current user is a member of the
// org before letting any child route mount. NotFoundError → bounce to
// the picker. Any other error bubbles to the error component.
export const Route = createFileRoute('/_authenticated/o/$orgId')({
  loader: ({ params, context }) =>
    context.queryClient.ensureQueryData(
      organizationMembershipQueryOptions(params.orgId),
    ),
  errorComponent: ({ error }) => {
    if (error instanceof NotFoundError) return <Navigate to="/o" />
    return (
      <main className="flex min-h-screen items-center justify-center px-4">
        <p className="text-sm text-red-600">
          Unable to load organization: {error.message}
        </p>
      </main>
    )
  },
  component: () => <Outlet />,
})
