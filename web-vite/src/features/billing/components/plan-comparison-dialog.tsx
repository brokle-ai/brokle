import { Check } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'

interface PlanComparisonDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  currentPlan: string | undefined
}

interface PlanTier {
  name: string
  label: string
  price: string
  blurb: string
  features: string[]
  cta: 'current' | 'upgrade' | 'downgrade' | 'contact'
}

// Static plan catalogue. The pricing copy here matches what the
// backend's `/effective-pricing` endpoint reports for the public Plan
// rows — keep them in sync with the seed pricing in
// `migrations/postgres/seeds/pricing.sql`. Self-serve checkout is
// deferred (Stripe integration), so the CTAs are descriptive rather
// than transactional today.
const PLANS: PlanTier[] = [
  {
    name: 'free',
    label: 'Free',
    price: '$0',
    blurb: 'Get started with generous monthly limits.',
    features: [
      '1M spans / month',
      '1 GB data processed',
      '10K scores / month',
      'Community support',
    ],
    cta: 'downgrade',
  },
  {
    name: 'pro',
    label: 'Pro',
    price: 'Usage-based',
    blurb: 'Scales with your traffic, no per-seat fees.',
    features: [
      '$0.50 / 100K spans',
      '$3.00 / GB data',
      '$0.20 / 1K scores',
      'Email support',
    ],
    cta: 'upgrade',
  },
  {
    name: 'enterprise',
    label: 'Enterprise',
    price: 'Custom',
    blurb: 'Volume tiers, contracts, SSO and SLAs.',
    features: [
      'Negotiated volume pricing',
      'SAML SSO + RBAC',
      'Dedicated support + SLA',
      'Annual contract',
    ],
    cta: 'contact',
  },
]

function ctaLabel(plan: PlanTier, currentPlan: string | undefined): string {
  if (plan.name === currentPlan) return 'Current plan'
  switch (plan.cta) {
    case 'upgrade':
      return 'Upgrade'
    case 'downgrade':
      return 'Downgrade'
    case 'contact':
      return 'Contact sales'
    default:
      return 'Switch'
  }
}

export function PlanComparisonDialog({
  open,
  onOpenChange,
  currentPlan,
}: PlanComparisonDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-3xl">
        <DialogHeader>
          <DialogTitle>Choose a plan</DialogTitle>
          <DialogDescription>
            Compare what each tier includes. Self-serve checkout is
            coming soon — for now, contact sales to switch tiers.
          </DialogDescription>
        </DialogHeader>

        <div className="grid grid-cols-1 gap-4 py-4 md:grid-cols-3">
          {PLANS.map((plan) => {
            const isCurrent = plan.name === currentPlan
            return (
              <div
                key={plan.name}
                className={cn(
                  'flex flex-col rounded-lg border p-4',
                  isCurrent && 'border-primary bg-primary/5',
                )}
              >
                <div className="flex items-baseline justify-between">
                  <h3 className="text-base font-semibold">{plan.label}</h3>
                  {isCurrent && (
                    <Badge variant="default" className="text-xs">
                      Current
                    </Badge>
                  )}
                </div>
                <p className="mt-1 text-2xl font-bold">{plan.price}</p>
                <p className="mt-1 text-sm text-muted-foreground">
                  {plan.blurb}
                </p>
                <ul className="mt-4 space-y-2 text-sm">
                  {plan.features.map((feat) => (
                    <li key={feat} className="flex items-start gap-2">
                      <Check className="mt-0.5 h-4 w-4 shrink-0 text-primary" />
                      <span>{feat}</span>
                    </li>
                  ))}
                </ul>
                <div className="mt-4 flex-1" />
                <Button
                  className="mt-4"
                  variant={isCurrent ? 'outline' : 'default'}
                  disabled={isCurrent}
                  onClick={() => {
                    if (plan.cta === 'contact') {
                      window.open('mailto:sales@brokle.com', '_blank')
                    } else {
                      // Self-serve plan switching is deferred — surface
                      // a clear next-step instead of pretending to act.
                      window.open('mailto:sales@brokle.com', '_blank')
                    }
                  }}
                >
                  {ctaLabel(plan, currentPlan)}
                </Button>
              </div>
            )
          })}
        </div>

        <DialogFooter>
          <Button variant="ghost" onClick={() => onOpenChange(false)}>
            Close
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
