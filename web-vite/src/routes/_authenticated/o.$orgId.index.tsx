import { createFileRoute, Link, redirect } from '@tanstack/react-router'
import { useSuspenseQuery } from '@tanstack/react-query'
import { projectListQueryOptions } from '@/features/projects/queries'

// If the org has exactly one project, skip the picker and land the
// user directly in it. Otherwise render the project picker.
export const Route = createFileRoute('/_authenticated/o/$orgId/')({
  loader: async ({ params, context }) => {
    const list = await context.queryClient.ensureQueryData(
      projectListQueryOptions(params.orgId),
    )
    if (list.data.length === 1) {
      throw redirect({
        to: '/o/$orgId/p/$projectId',
        params: { orgId: params.orgId, projectId: list.data[0]!.id },
      })
    }
    return list
  },
  component: ProjectPicker,
})

function ProjectPicker() {
  const { orgId } = Route.useParams()
  const { data } = useSuspenseQuery(projectListQueryOptions(orgId))

  if (data.data.length === 0) {
    return (
      <main className="flex min-h-screen items-center justify-center px-4">
        <div className="max-w-md space-y-3 rounded-lg border p-6 text-center">
          <h1 className="text-lg font-semibold">No projects yet</h1>
          <p className="text-sm text-muted-foreground">
            This organization has no projects. Create one from the admin area.
          </p>
        </div>
      </main>
    )
  }

  return (
    <main className="mx-auto max-w-3xl p-6 space-y-4">
      <h1 className="text-xl font-semibold">Select a project</h1>
      <ul className="space-y-2">
        {data.data.map((project) => (
          <li key={project.id}>
            <Link
              to="/o/$orgId/p/$projectId"
              params={{ orgId, projectId: project.id }}
              className="block rounded border p-4 hover:bg-muted"
            >
              <div className="font-medium">{project.name}</div>
              <div className="text-xs text-muted-foreground">{project.slug}</div>
            </Link>
          </li>
        ))}
      </ul>
    </main>
  )
}
