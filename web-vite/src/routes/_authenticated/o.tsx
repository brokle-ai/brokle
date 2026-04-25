import { createFileRoute } from '@tanstack/react-router'
import { useSuspenseQuery } from '@tanstack/react-query'
import { organizationListQueryOptions } from '@/features/organizations/queries'
import { OrganizationGrid } from '@/features/organizations/components/organization-grid'

// Org picker. Lands here when the user has no default organization
// set, or deliberately navigates to /o.
export const Route = createFileRoute('/_authenticated/o')({
  loader: ({ context }) =>
    context.queryClient.ensureQueryData(organizationListQueryOptions()),
  component: OrgPicker,
})

function OrgPicker() {
  const { data } = useSuspenseQuery(organizationListQueryOptions())

  return (
    <main className="mx-auto w-full max-w-7xl p-6">
      <OrganizationGrid organizations={data.data} />
    </main>
  )
}
