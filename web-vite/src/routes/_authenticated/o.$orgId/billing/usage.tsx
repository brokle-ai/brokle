import { createFileRoute } from '@tanstack/react-router'
import { useSuspenseQuery } from '@tanstack/react-query'
import { z } from 'zod'
import { UsageByProjectTable, UsageTable } from '@/features/billing/components'
import {
  type UsagePeriod,
  type UsagePeriodRelative,
  usageQueryOptions,
} from '@/features/billing/api/queries'

// URL-backed period state: either a `range` preset OR a `from`/`to`
// custom window (RFC 3339). The legacy single `range` query param is
// preserved so existing bookmarks keep resolving to a sensible view.
const PERIOD_PRESETS = [
  'current',
  'previous',
  '7d',
  '14d',
  '30d',
] as const satisfies readonly UsagePeriodRelative[]

const searchSchema = z.object({
  range: z.enum(PERIOD_PRESETS).optional().catch(undefined),
  from: z.string().optional().catch(undefined),
  to: z.string().optional().catch(undefined),
})

export const Route = createFileRoute(
  '/_authenticated/o/$orgId/billing/usage',
)({
  validateSearch: searchSchema,
  loader: ({ params, context }) =>
    context.queryClient.ensureQueryData(usageQueryOptions(params.orgId)),
  component: UsagePage,
})

function periodFromSearch(search: {
  range?: UsagePeriodRelative
  from?: string
  to?: string
}): UsagePeriod {
  if (search.from && search.to) {
    return { kind: 'custom', from: search.from, to: search.to }
  }
  return { kind: 'relative', relative: search.range ?? 'current' }
}

function UsagePage() {
  const { orgId } = Route.useParams()
  const search = Route.useSearch()
  const navigate = Route.useNavigate()
  const { data: overview } = useSuspenseQuery(usageQueryOptions(orgId))
  const period = periodFromSearch(search)

  const setPeriod = (next: UsagePeriod) => {
    if (next.kind === 'custom') {
      void navigate({
        search: { from: next.from, to: next.to, range: undefined },
        replace: true,
      })
    } else {
      void navigate({
        search: { range: next.relative, from: undefined, to: undefined },
        replace: true,
      })
    }
  }

  return (
    <div className="space-y-8">
      <UsageTable overview={overview} />
      <UsageByProjectTable
        orgId={orgId}
        period={period}
        onPeriodChange={setPeriod}
      />
    </div>
  )
}
