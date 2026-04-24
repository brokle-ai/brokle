import { createFileRoute, Link } from '@tanstack/react-router'
import { useSuspenseQuery } from '@tanstack/react-query'
import { organizationListQueryOptions } from '@/features/organizations/queries'

// Org picker. Lands here when the user has no default organization
// set, or deliberately navigates to /o.
export const Route = createFileRoute('/_authenticated/o')({
  loader: ({ context }) =>
    context.queryClient.ensureQueryData(organizationListQueryOptions()),
  component: OrgPicker,
})

function OrgPicker() {
  const { data } = useSuspenseQuery(organizationListQueryOptions())

  if (data.data.length === 0) {
    return (
      <main className="flex min-h-screen items-center justify-center px-4">
        <div className="max-w-md space-y-3 rounded-lg border p-6 text-center">
          <h1 className="text-lg font-semibold">No organizations yet</h1>
          <p className="text-sm text-muted-foreground">
            You&apos;re signed in, but you&apos;re not a member of any organization. Ask
            your admin to invite you, or create one if you&apos;re getting started.
          </p>
        </div>
      </main>
    )
  }

  return (
    <main className="mx-auto max-w-3xl p-6 space-y-4">
      <h1 className="text-xl font-semibold">Select an organization</h1>
      <ul className="space-y-2">
        {data.data.map((org) => (
          <li key={org.id}>
            <Link
              to="/o/$orgId"
              params={{ orgId: org.id }}
              className="block rounded border p-4 hover:bg-muted"
            >
              <div className="font-medium">{org.name}</div>
              <div className="text-xs text-muted-foreground">
                {org.subscription_plan} · {org.slug}
              </div>
            </Link>
          </li>
        ))}
      </ul>
    </main>
  )
}
