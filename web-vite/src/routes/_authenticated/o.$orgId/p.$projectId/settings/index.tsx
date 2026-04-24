import { createFileRoute } from '@tanstack/react-router'
import { useSuspenseQuery } from '@tanstack/react-query'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { projectMembershipQueryOptions } from '@/features/projects/queries'

export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/settings/',
)({
  loader: ({ params, context }) =>
    context.queryClient.ensureQueryData(
      projectMembershipQueryOptions(params.projectId),
    ),
  component: ProjectSettingsIndex,
})

function formatTimestamp(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString()
}

function ProjectSettingsIndex() {
  const { projectId } = Route.useParams()
  const { data: project } = useSuspenseQuery(
    projectMembershipQueryOptions(projectId),
  )

  return (
    <Card>
      <CardHeader>
        <CardTitle>{project.name}</CardTitle>
        <CardDescription>
          Read-only view. Rename, transfer, and delete operations ship in a
          follow-up port.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <dl className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <Field label="Name" value={project.name} />
          <Field label="Slug" value={project.slug} mono />
          <Field label="Project ID" value={project.id} mono />
          <Field
            label="Organization ID"
            value={project.organization_id}
            mono
          />
          <Field label="Created" value={formatTimestamp(project.created_at)} />
          <Field label="Updated" value={formatTimestamp(project.updated_at)} />
        </dl>
      </CardContent>
    </Card>
  )
}

function Field({
  label,
  value,
  mono = false,
}: {
  label: string
  value: string
  mono?: boolean
}) {
  return (
    <div>
      <dt className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
        {label}
      </dt>
      <dd
        className={
          mono
            ? 'mt-1 font-mono text-xs text-foreground'
            : 'mt-1 text-sm text-foreground'
        }
      >
        {value}
      </dd>
    </div>
  )
}
