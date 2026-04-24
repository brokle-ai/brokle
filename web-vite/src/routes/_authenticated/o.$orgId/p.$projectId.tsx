import { createFileRoute, Outlet, Navigate } from '@tanstack/react-router'
import { NotFoundError } from '@/lib/api/errors'
import { projectMembershipQueryOptions } from '@/features/projects/queries'
import { AppSidebar } from '@/components/layout/app-sidebar'
import { AppHeader } from '@/components/layout/app-header'

// Project tenancy segment — also owns the dashboard chrome (sidebar +
// header) so every nested route renders under the same shell.
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
  component: ProjectShell,
})

function ProjectShell() {
  return (
    <div className="flex min-h-screen">
      <AppSidebar />
      <div className="flex flex-1 flex-col">
        <AppHeader />
        <main className="flex-1 overflow-auto">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
