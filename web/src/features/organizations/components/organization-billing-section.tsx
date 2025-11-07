'use client'

import { useState } from 'react'
import { CreditCard, TrendingUp, DollarSign, Zap, ExternalLink } from 'lucide-react'
import { useWorkspace } from '@/context/workspace-context'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { toast } from 'sonner'

export function OrganizationBillingSection() {
  const { currentOrganization } = useWorkspace()
  const [billingEmail, setBillingEmail] = useState(currentOrganization?.billing_email || '')
  const [isLoading, setIsLoading] = useState(false)

  if (!currentOrganization) {
    return null
  }

  const handleUpdateBillingEmail = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsLoading(true)

    try {
      // TODO: Implement API call
      await new Promise(resolve => setTimeout(resolve, 1000))
      toast.success('Billing email updated successfully')
    } catch (error) {
      console.error('Failed to update billing email:', error)
      toast.error('Failed to update billing email. Please try again.')
    } finally {
      setIsLoading(false)
    }
  }

  const getPlanColor = (plan: string) => {
    switch (plan) {
      case 'enterprise':
        return 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-300'
      case 'business':
        return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300'
      case 'pro':
        return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300'
      case 'free':
        return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300'
      default:
        return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300'
    }
  }

  const getPlanFeatures = (plan: string) => {
    const features: Record<string, string[]> = {
      free: ['10,000 requests/month', 'Community support', 'Basic analytics', '7-day data retention'],
      pro: ['100,000 requests/month', 'Email support', 'Advanced analytics', '30-day data retention', 'Custom integrations'],
      business: ['1M requests/month', 'Priority support', 'Team collaboration', '90-day data retention', 'SSO (coming soon)', 'Advanced security'],
      enterprise: ['Unlimited requests', '24/7 phone support', 'Dedicated account manager', 'Custom retention', 'SSO & SAML', 'SLA guarantee']
    }
    return features[plan] || features.free
  }

  return (
    <div className="space-y-8">
      {/* Current Plan */}
      <div className="rounded-lg border p-4 space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <div className="text-sm font-medium text-muted-foreground mb-1">Current Plan</div>
            <div className="flex items-center gap-2">
              <Badge className={getPlanColor(currentOrganization.plan)} className="text-lg px-3 py-1">
                {currentOrganization.plan.charAt(0).toUpperCase() + currentOrganization.plan.slice(1)}
              </Badge>
            </div>
          </div>
          {currentOrganization.plan !== 'enterprise' && (
            <Button>
              <Zap className="mr-2 h-4 w-4" />
              Upgrade Plan
            </Button>
          )}
        </div>

        <Separator />

        <div>
          <div className="text-sm font-medium mb-3">Plan Features</div>
          <ul className="space-y-2">
            {getPlanFeatures(currentOrganization.plan).map((feature, index) => (
              <li key={index} className="flex items-center gap-2 text-sm">
                <div className="h-1.5 w-1.5 rounded-full bg-primary" />
                {feature}
              </li>
            ))}
          </ul>
        </div>
      </div>

      {/* Usage Statistics */}
      <div className="space-y-4">
        <h3 className="text-lg font-medium">Usage This Month</h3>

        <div className="grid grid-cols-3 gap-4">
          <div className="rounded-lg border p-4">
            <div className="flex items-center gap-2 mb-2">
              <TrendingUp className="h-4 w-4 text-muted-foreground" />
              <div className="text-sm font-medium text-muted-foreground">Requests</div>
            </div>
            <div className="text-2xl font-bold">12,847</div>
            <div className="text-xs text-muted-foreground mt-1">of 100,000 limit</div>
          </div>

          <div className="rounded-lg border p-4">
            <div className="flex items-center gap-2 mb-2">
              <DollarSign className="h-4 w-4 text-muted-foreground" />
              <div className="text-sm font-medium text-muted-foreground">Estimated Cost</div>
            </div>
            <div className="text-2xl font-bold">$24.50</div>
            <div className="text-xs text-muted-foreground mt-1">Current billing cycle</div>
          </div>

          <div className="rounded-lg border p-4">
            <div className="flex items-center gap-2 mb-2">
              <Zap className="h-4 w-4 text-muted-foreground" />
              <div className="text-sm font-medium text-muted-foreground">Models</div>
            </div>
            <div className="text-2xl font-bold">5</div>
            <div className="text-xs text-muted-foreground mt-1">Active AI models</div>
          </div>
        </div>
      </div>

      {/* Billing Contact */}
      <form onSubmit={handleUpdateBillingEmail} className="space-y-4">
        <h3 className="text-lg font-medium">Billing Contact</h3>

        <div className="space-y-2">
          <Label htmlFor="billingEmail">Billing Email</Label>
          <Input
            id="billingEmail"
            type="email"
            value={billingEmail}
            onChange={(e) => setBillingEmail(e.target.value)}
            placeholder="billing@example.com"
          />
          <p className="text-xs text-muted-foreground">
            Invoices and billing notifications will be sent to this email address
          </p>
        </div>

        <Button type="submit" disabled={isLoading}>
          {isLoading ? 'Saving...' : 'Update Billing Email'}
        </Button>
      </form>

      {/* Payment Method */}
      <div className="space-y-4">
        <h3 className="text-lg font-medium">Payment Method</h3>

        <div className="rounded-lg border p-4 space-y-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-muted rounded-lg">
                <CreditCard className="h-5 w-5" />
              </div>
              <div>
                <div className="font-medium">Visa ending in 4242</div>
                <div className="text-sm text-muted-foreground">Expires 12/2025</div>
              </div>
            </div>
            <Button variant="outline" size="sm">
              Update
            </Button>
          </div>
        </div>

        <Button variant="ghost" className="w-full" asChild>
          <a href="#" target="_blank" rel="noopener noreferrer">
            <ExternalLink className="mr-2 h-4 w-4" />
            Manage Payment Methods in Stripe
          </a>
        </Button>
      </div>

      {/* Invoice History */}
      <div className="space-y-4">
        <h3 className="text-lg font-medium">Recent Invoices</h3>

        <div className="rounded-lg border p-4 text-center py-8 text-muted-foreground">
          <CreditCard className="mx-auto h-8 w-8 mb-2 opacity-50" />
          <div className="text-sm">No invoices yet</div>
          <div className="text-xs">Invoices will appear here once billing starts</div>
        </div>
      </div>
    </div>
  )
}
