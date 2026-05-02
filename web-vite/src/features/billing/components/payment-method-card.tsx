import { CreditCard, Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'

// Payment method UI placeholder. Stripe Elements integration is
// deferred — when the backend grows a `/billing/payment-methods`
// endpoint plus a Setup Intent flow, swap the empty-state for a list
// of cards + the Stripe `<PaymentElement>` mounted inside a dialog.
// Today the card is purely informational so the billing page reads
// the same shape as Stripe's / Vercel's billing surface.
export function PaymentMethodCard() {
  return (
    <div className="rounded-lg border bg-card">
      <div className="flex items-center justify-between border-b p-6">
        <div>
          <h2 className="text-base font-semibold">Payment method</h2>
          <p className="text-sm text-muted-foreground">
            Add a card to enable paid features.
          </p>
        </div>
        <Button variant="outline" size="sm" disabled>
          <Plus className="mr-2 h-4 w-4" />
          Add card
        </Button>
      </div>
      <div className="p-6">
        <div className="flex flex-col items-center justify-center gap-2 rounded-md border-2 border-dashed py-8 text-center">
          <CreditCard className="h-10 w-10 text-muted-foreground" />
          <p className="text-sm text-muted-foreground">
            No payment method on file
          </p>
          <p className="text-xs text-muted-foreground">
            Stripe checkout integration coming soon.
          </p>
        </div>
      </div>
    </div>
  )
}
