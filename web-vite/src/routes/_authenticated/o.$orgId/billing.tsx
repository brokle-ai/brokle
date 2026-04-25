import { createFileRoute, Link, Outlet, useMatches } from '@tanstack/react-router'
import { useSuspenseQuery } from '@tanstack/react-query'
import { BrokleError } from '@/lib/api/errors'
import {
  AlertsList,
  BudgetCard,
  PaymentMethodCard,
  PlanCard,
} from '@/features/billing/components'
import { planQueryOptions } from '@/features/billing/api/queries'

// The billing surface is three tabs over one layout — plan (this
// route's own content), usage, invoices. A match on the child path is
// what flips us into "show the Outlet instead of the plan card" mode
// so the shared header + tab strip render once regardless of which tab
// is active.
export const Route = createFileRoute('/_authenticated/o/$orgId/billing')({
  loader: ({ params, context }) =>
    context.queryClient.ensureQueryData(planQueryOptions(params.orgId)),
  errorComponent: BillingErrorBoundary,
  component: BillingLayout,
})

function BillingErrorBoundary({ error }: { error: Error }) {
  if (error instanceof BrokleError && error.status >= 400 && error.status < 500) {
    return (
      <main className="mx-auto max-w-5xl p-6">
        <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-6">
          <p className="text-sm font-medium text-destructive">
            Unable to load billing
          </p>
          <p className="mt-1 text-sm text-muted-foreground">{error.message}</p>
        </div>
      </main>
    )
  }
  throw error
}

function BillingLayout() {
  const { orgId } = Route.useParams()
  const matches = useMatches()
  // Match the leaf route id — if any child of the billing layout is
  // active, defer to <Outlet />. Otherwise render the plan card as the
  // landing content.
  const isChild = matches.some((m) =>
    m.routeId.startsWith('/_authenticated/o/$orgId/billing/'),
  )

  return (
    <main className="mx-auto max-w-5xl space-y-6 p-6">
      <header>
        <h1 className="text-2xl font-semibold">Billing</h1>
        <p className="text-sm text-muted-foreground">
          Manage your plan, usage, and invoices for this organization.
        </p>
      </header>

      <nav className="flex gap-1 border-b">
        <TabLink to="/o/$orgId/billing" params={{ orgId }} label="Plan" exact />
        <TabLink
          to="/o/$orgId/billing/usage"
          params={{ orgId }}
          label="Usage"
        />
        <TabLink
          to="/o/$orgId/billing/budgets"
          params={{ orgId }}
          label="Budgets"
        />
        <TabLink
          to="/o/$orgId/billing/invoices"
          params={{ orgId }}
          label="Invoices"
        />
      </nav>

      {isChild ? <Outlet /> : <PlanLandingContent orgId={orgId} />}
    </main>
  )
}

function PlanLandingContent({ orgId }: { orgId: string }) {
  const { data: pricing } = useSuspenseQuery(planQueryOptions(orgId))
  return (
    <div className="space-y-6">
      <PlanCard pricing={pricing} />
      <BudgetCard orgId={orgId} />
      <section className="space-y-3">
        <div>
          <h2 className="text-base font-semibold">Active alerts</h2>
          <p className="text-sm text-muted-foreground">
            Threshold breaches that haven&apos;t been acknowledged.
          </p>
        </div>
        <AlertsList orgId={orgId} limit={10} triggeredOnly />
      </section>
      <PaymentMethodCard />
    </div>
  )
}

// Thin wrapper so the active tab picks up an underline. TanStack
// Router's `Link.activeProps` would work too, but this keeps the
// "exact" match semantics explicit for the parent (/billing) tab.
interface TabLinkProps {
  to:
    | '/o/$orgId/billing'
    | '/o/$orgId/billing/usage'
    | '/o/$orgId/billing/invoices'
    | '/o/$orgId/billing/budgets'
  params: { orgId: string }
  label: string
  exact?: boolean
}

function TabLink({ to, params, label, exact }: TabLinkProps) {
  return (
    <Link
      to={to}
      params={params}
      activeOptions={{ exact: !!exact }}
      className="px-4 py-2 text-sm text-muted-foreground hover:text-foreground"
      activeProps={{
        className:
          'px-4 py-2 text-sm font-medium text-foreground border-b-2 border-primary -mb-px',
      }}
    >
      {label}
    </Link>
  )
}
