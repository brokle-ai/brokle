import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import { budgetsQueryOptions } from '../api/queries'
import { BudgetFormDialog } from './budget-form-dialog'
import type { UsageBudget } from '../api/types'

interface BudgetCardProps {
  orgId: string
}

function formatMoney(decimalStr: string | undefined): string {
  if (!decimalStr) return '—'
  const n = Number(decimalStr)
  if (!Number.isFinite(n)) return decimalStr
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(n)
}

// First org-level budget (`project_id == null`) drives the card. If an
// org runs multiple project budgets they show up as rows in a future
// list view — not in this landing summary.
function pickOrgBudget(budgets: UsageBudget[]): UsageBudget | null {
  const orgLevel = budgets.find((b) => !b.project_id)
  return orgLevel ?? budgets[0] ?? null
}

// Landing card on /o/$orgId/billing. Shows current-period spend against
// the configured limit; when no budget exists it shifts to a CTA that
// opens the same form dialog in "create" mode.
export function BudgetCard({ orgId }: BudgetCardProps) {
  const [formOpen, setFormOpen] = useState(false)
  const { data, isLoading, isError, error } = useQuery(
    budgetsQueryOptions(orgId),
  )

  const budget = data ? pickOrgBudget(data) : null

  if (isError) {
    return (
      <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-6">
        <p className="text-sm font-medium text-destructive">
          Unable to load budget
        </p>
        <p className="mt-1 text-sm text-muted-foreground">
          {error instanceof Error ? error.message : 'Unknown error'}
        </p>
      </div>
    )
  }

  if (isLoading) {
    return (
      <div className="rounded-lg border p-6">
        <p className="text-sm text-muted-foreground">Loading budget…</p>
      </div>
    )
  }

  return (
    <>
      {budget ? (
        <BudgetPanel budget={budget} onEdit={() => setFormOpen(true)} />
      ) : (
        <EmptyBudgetPanel onCreate={() => setFormOpen(true)} />
      )}

      <BudgetFormDialog
        orgId={orgId}
        budget={budget}
        open={formOpen}
        onOpenChange={setFormOpen}
      />
    </>
  )
}

function BudgetPanel({
  budget,
  onEdit,
}: {
  budget: UsageBudget
  onEdit: () => void
}) {
  const limit = budget.cost_limit ? Number(budget.cost_limit) : 0
  const current = Number(budget.current_cost)
  const pct =
    limit > 0 && Number.isFinite(current)
      ? Math.min(100, Math.max(0, (current / limit) * 100))
      : 0

  return (
    <div className="rounded-lg border p-6 space-y-4">
      <div className="flex items-start justify-between gap-4">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <h3 className="text-base font-semibold">{budget.name}</h3>
            <Badge variant="outline" className="capitalize">
              {budget.budget_type}
            </Badge>
            {!budget.is_active && <Badge variant="secondary">Paused</Badge>}
          </div>
          <p className="text-sm text-muted-foreground">
            {formatMoney(budget.current_cost)} of{' '}
            {formatMoney(budget.cost_limit)} this period
          </p>
        </div>
        <Button variant="outline" size="sm" onClick={onEdit}>
          Edit budget
        </Button>
      </div>

      <div className="space-y-1">
        <Progress value={pct} />
        <p className="text-xs text-muted-foreground">
          {pct.toFixed(1)}% used
          {budget.alert_thresholds.length > 0 &&
            ` · alerts at ${budget.alert_thresholds.join(', ')}%`}
        </p>
      </div>
    </div>
  )
}

function EmptyBudgetPanel({ onCreate }: { onCreate: () => void }) {
  return (
    <div className="rounded-lg border border-dashed p-6 flex items-start justify-between gap-4">
      <div className="space-y-1">
        <h3 className="text-base font-semibold">No budget set</h3>
        <p className="text-sm text-muted-foreground">
          Set a monthly or weekly spend cap and get alerts before you hit it.
        </p>
      </div>
      <Button onClick={onCreate}>Create budget</Button>
    </div>
  )
}
