import { createFileRoute, redirect } from '@tanstack/react-router'
import { useSuspenseQuery } from '@tanstack/react-query'
import { projectListQueryOptions } from '@/features/projects/queries'
import { organizationMembershipQueryOptions } from '@/features/organizations/queries'
import { ProjectGrid } from '@/features/organizations/components/project-grid'

// If the org has exactly one project, skip the picker and land the
// user directly in it. Otherwise render the project picker.
export const Route = createFileRoute('/_authenticated/o/$orgId/')({
  loader: async ({ params, context }) => {
    const [list] = await Promise.all([
      context.queryClient.ensureQueryData(
        projectListQueryOptions(params.orgId),
      ),
      context.queryClient.ensureQueryData(
        organizationMembershipQueryOptions(params.orgId),
      ),
    ])
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
  const { data: projects } = useSuspenseQuery(projectListQueryOptions(orgId))

  return (
    <main className="mx-auto w-full max-w-7xl p-6">
      <ProjectGrid orgId={orgId} projects={projects.data} />
    </main>
  )
}
