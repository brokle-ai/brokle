import { createFileRoute } from '@tanstack/react-router'
import { useSuspenseQuery } from '@tanstack/react-query'
import { projectMembershipQueryOptions } from '@/features/projects/queries'
import { organizationMembershipQueryOptions } from '@/features/organizations/queries'

// Project home / dashboard entry point. Phase 1.5 replaces this with
// the actual observability overview when the first feature group
// ports. Keeps the router tree valid in the meantime.
export const Route = createFileRoute('/_authenticated/o/$orgId/p/$projectId/')({
  component: ProjectHome,
})

function ProjectHome() {
  const { orgId, projectId } = Route.useParams()
  const { data: org } = useSuspenseQuery(organizationMembershipQueryOptions(orgId))
  const { data: project } = useSuspenseQuery(projectMembershipQueryOptions(projectId))

  return (
    <main className="mx-auto max-w-5xl p-6 space-y-6">
      <header className="space-y-1">
        <p className="text-xs uppercase tracking-wide text-muted-foreground">
          {org.name}
        </p>
        <h1 className="text-2xl font-semibold">{project.name}</h1>
      </header>
      <section className="rounded-lg border p-6">
        <h2 className="font-medium">Project home</h2>
        <p className="mt-1 text-sm text-muted-foreground">
          Observability overview ports here in Phase 1.5. Tenancy scaffold is live
          (org {org.id}, project {project.id}).
        </p>
      </section>
    </main>
  )
}
