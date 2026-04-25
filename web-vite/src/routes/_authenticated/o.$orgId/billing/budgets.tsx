import { createFileRoute } from '@tanstack/react-router'
import { AlertsList, BudgetsList } from '@/features/billing/components'

export const Route = createFileRoute(
  '/_authenticated/o/$orgId/billing/budgets',
)({
  component: BudgetsTabPage,
})

function BudgetsTabPage() {
  const { orgId } = Route.useParams()
  return (
    <div className="space-y-6">
      <BudgetsList orgId={orgId} />
      <section className="space-y-3">
        <div>
          <h2 className="text-base font-semibold">Recent alerts</h2>
          <p className="text-sm text-muted-foreground">
            Threshold breaches across all budgets in this organization.
          </p>
        </div>
        <AlertsList orgId={orgId} limit={25} />
      </section>
    </div>
  )
}
