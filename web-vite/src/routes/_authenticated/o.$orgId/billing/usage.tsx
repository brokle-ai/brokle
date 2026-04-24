import { createFileRoute } from '@tanstack/react-router'
import { useSuspenseQuery } from '@tanstack/react-query'
import { UsageTable } from '@/features/billing/components'
import { usageQueryOptions } from '@/features/billing/api/queries'

export const Route = createFileRoute(
  '/_authenticated/o/$orgId/billing/usage',
)({
  loader: ({ params, context }) =>
    context.queryClient.ensureQueryData(usageQueryOptions(params.orgId)),
  component: UsagePage,
})

function UsagePage() {
  const { orgId } = Route.useParams()
  const { data: overview } = useSuspenseQuery(usageQueryOptions(orgId))
  return <UsageTable overview={overview} />
}
