import { createFileRoute } from '@tanstack/react-router'
import { useSuspenseQuery } from '@tanstack/react-query'
import { organizationListQueryOptions } from '@/features/organizations/queries'
import { OrganizationGrid } from '@/features/organizations/components/organization-grid'

// Org picker. Lands here when the user has no default organization
// set, or deliberately navigates to /o.
//
// Lives in `o.index.tsx` (not `o.tsx`) so it ONLY renders for the
// exact path `/o`. The sibling `o.tsx` is a layout-only route with
// just <Outlet />, so child routes (`/o/$orgId`, `/o/$orgId/p/...`)
// can render their own content. Without this split, the picker would
// render for every `/o/*` URL and override the matched child.
export const Route = createFileRoute('/_authenticated/o/')({
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
