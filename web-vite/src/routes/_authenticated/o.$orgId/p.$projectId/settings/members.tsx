import { createFileRoute } from '@tanstack/react-router'
import { useSuspenseQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { z } from 'zod'
import { BrokleError } from '@/lib/api/errors'
import { Button } from '@/components/ui/button'
import {
  InviteMemberDialog,
  MembersTable,
  PendingInvitationsTable,
} from '@/features/members/components'
import {
  memberListQueryOptions,
  pendingInvitationsQueryOptions,
} from '@/features/members/api/queries'

// Search params exist for forward-compat (future pagination / search).
// The current list-members endpoint returns all members in a single
// response, so these values are stored in the URL but not yet sent
// to the backend — see members/api/queries.ts.
const searchSchema = z.object({
  page: z.number().int().min(1).catch(1),
  limit: z.number().int().min(1).max(100).catch(50),
  q: z.string().optional().catch(undefined),
})

export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/settings/members',
)({
  validateSearch: searchSchema,
  loaderDeps: ({ search }) => ({
    page: search.page,
    limit: search.limit,
    q: search.q,
  }),
  loader: ({ params, context, deps }) =>
    // Fire both prefetches in parallel — members and pending
    // invitations are independent reads and the route needs both.
    Promise.all([
      context.queryClient.ensureQueryData(
        memberListQueryOptions(params.orgId, {
          page: deps.page,
          limit: deps.limit,
          q: deps.q,
        }),
      ),
      context.queryClient.ensureQueryData(
        pendingInvitationsQueryOptions(params.orgId),
      ),
    ]),
  errorComponent: MembersErrorBoundary,
  component: MembersPage,
})

function MembersErrorBoundary({ error }: { error: Error }) {
  if (
    error instanceof BrokleError &&
    error.status >= 400 &&
    error.status < 500
  ) {
    return (
      <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-6">
        <p className="text-sm font-medium text-destructive">
          Unable to load members
        </p>
        <p className="mt-1 text-sm text-muted-foreground">{error.message}</p>
      </div>
    )
  }
  throw error
}

function MembersPage() {
  const { orgId } = Route.useParams()
  const search = Route.useSearch()
  const [inviteOpen, setInviteOpen] = useState(false)
  const { data } = useSuspenseQuery(
    memberListQueryOptions(orgId, {
      page: search.page,
      limit: search.limit,
      q: search.q,
    }),
  )
  const { data: invitations } = useSuspenseQuery(
    pendingInvitationsQueryOptions(orgId),
  )

  // Filter out non-pending rows for the "pending" table — accepted /
  // revoked / expired invitations live on the same list endpoint but
  // this view is intentionally scoped to actionable ones.
  const pendingRows = invitations.data.filter(
    (inv) => inv.status === 'pending',
  )

  return (
    <section className="space-y-8">
      <div className="space-y-4">
        <div className="flex items-start justify-between gap-4">
          <div>
            <h2 className="text-lg font-semibold">Members</h2>
            <p className="text-sm text-muted-foreground">
              {data.data.length.toLocaleString()} total
            </p>
          </div>
          <Button onClick={() => setInviteOpen(true)}>Invite member</Button>
        </div>

        <MembersTable orgId={orgId} rows={data.data} />
      </div>

      <div className="space-y-3">
        <div>
          <h3 className="text-base font-semibold">Pending invitations</h3>
          <p className="text-sm text-muted-foreground">
            {pendingRows.length.toLocaleString()} awaiting acceptance
          </p>
        </div>
        <PendingInvitationsTable orgId={orgId} rows={pendingRows} />
      </div>

      <InviteMemberDialog
        orgId={orgId}
        open={inviteOpen}
        onOpenChange={setInviteOpen}
      />
    </section>
  )
}
