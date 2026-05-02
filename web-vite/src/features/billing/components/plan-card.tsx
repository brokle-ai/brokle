import { useState } from 'react'
import { Sparkles } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import type { EffectivePricing } from '../api/types'
import { PlanComparisonDialog } from './plan-comparison-dialog'

interface PlanCardProps {
  pricing: EffectivePricing
}

// decimal strings arrive with arbitrary precision (shopspring/decimal).
// Parse with `Number` for display only — never for arithmetic the user
// sees. Zero/invalid values fall back to a dash so the card never
// shows NaN.
function formatMoney(decimalStr: string | undefined, currency = 'USD'): string {
  if (!decimalStr) return '—'
  const n = Number(decimalStr)
  if (!Number.isFinite(n)) return decimalStr
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency,
    minimumFractionDigits: 2,
    maximumFractionDigits: 4,
  }).format(n)
}

function formatInt(n: number | undefined): string {
  if (n === undefined || n === null) return '—'
  return n.toLocaleString()
}

function formatGB(decimalStr: string | undefined): string {
  if (!decimalStr) return '—'
  const n = Number(decimalStr)
  if (!Number.isFinite(n)) return decimalStr
  return `${n.toLocaleString(undefined, { maximumFractionDigits: 2 })} GB`
}

function formatDate(iso: string | undefined): string {
  if (!iso) return '—'
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

function planTierLabel(name: string | undefined): string {
  if (!name) return 'Unknown'
  return name.charAt(0).toUpperCase() + name.slice(1)
}

export function PlanCard({ pricing }: PlanCardProps) {
  const plan = pricing.base_plan
  const contract = pricing.contract
  const currency = contract?.currency ?? 'USD'
  const [comparisonOpen, setComparisonOpen] = useState(false)

  return (
    <div className="rounded-lg border bg-card">
      <div className="flex items-start justify-between border-b p-6">
        <div>
          <p className="text-xs uppercase tracking-wide text-muted-foreground">
            Current plan
          </p>
          <div className="mt-1 flex items-baseline gap-3">
            <h2 className="text-2xl font-semibold">
              {planTierLabel(plan?.name)}
            </h2>
            {contract ? (
              <Badge variant="secondary">Enterprise contract</Badge>
            ) : (
              <Badge variant="outline">Standard</Badge>
            )}
          </div>
        </div>
        <div className="flex flex-col items-end gap-2">
          {contract?.end_date ? (
            <div className="text-right">
              <p className="text-xs uppercase tracking-wide text-muted-foreground">
                Renews / expires
              </p>
              <p className="mt-1 text-sm font-medium">
                {formatDate(contract.end_date)}
              </p>
            </div>
          ) : null}
          <Button
            size="sm"
            variant={plan?.name === 'enterprise' ? 'outline' : 'default'}
            onClick={() => setComparisonOpen(true)}
          >
            <Sparkles className="mr-2 h-4 w-4" />
            {plan?.name === 'free' ? 'Upgrade plan' : 'Change plan'}
          </Button>
        </div>
      </div>

      <dl className="grid grid-cols-1 gap-6 p-6 sm:grid-cols-3">
        <div>
          <dt className="text-xs uppercase tracking-wide text-muted-foreground">
            Spans included
          </dt>
          <dd className="mt-1 text-lg font-semibold">
            {formatInt(pricing.free_spans)}
          </dd>
          <dd className="text-xs text-muted-foreground">
            then {formatMoney(pricing.price_per_100k_spans, currency)} /100K
          </dd>
        </div>
        <div>
          <dt className="text-xs uppercase tracking-wide text-muted-foreground">
            Data included
          </dt>
          <dd className="mt-1 text-lg font-semibold">
            {formatGB(pricing.free_gb)}
          </dd>
          <dd className="text-xs text-muted-foreground">
            then {formatMoney(pricing.price_per_gb, currency)} /GB
          </dd>
        </div>
        <div>
          <dt className="text-xs uppercase tracking-wide text-muted-foreground">
            Scores included
          </dt>
          <dd className="mt-1 text-lg font-semibold">
            {formatInt(pricing.free_scores)}
          </dd>
          <dd className="text-xs text-muted-foreground">
            then {formatMoney(pricing.price_per_1k_scores, currency)} /1K
          </dd>
        </div>
      </dl>

      {contract ? (
        <div className="border-t bg-muted/40 px-6 py-3 text-xs text-muted-foreground">
          Contract <span className="font-mono">{contract.contract_number}</span>
          {contract.minimum_commit_amount
            ? ` · min commit ${formatMoney(contract.minimum_commit_amount, currency)}`
            : null}
          {pricing.has_volume_tiers ? ' · volume tiers active' : null}
        </div>
      ) : null}

      <PlanComparisonDialog
        open={comparisonOpen}
        onOpenChange={setComparisonOpen}
        currentPlan={plan?.name}
      />
    </div>
  )
}
