'use client'

import {
  Check,
  AlertCircle,
  DollarSign,
  CreditCard,
  Plus,
  FileText,
  ExternalLink,
  Sparkles,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { cn } from '@/lib/utils'

import { useUsageOverviewQuery } from '../hooks'

interface BillingPageProps {
  organizationId: string
  className?: string
}

function formatCurrency(value: number): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(value)
}

function formatPeriodDate(dateStr: string | undefined): string {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('en-US', {
    month: 'long',
    day: 'numeric',
    year: 'numeric',
  })
}

export function BillingPage({ organizationId, className }: BillingPageProps) {
  const {
    data: overview,
    isLoading,
    error,
  } = useUsageOverviewQuery(organizationId)

  const errorMessage = error
    ? typeof error === 'object' && 'message' in error
      ? (error.message as string)
      : String(error)
    : null

  // Calculate if all usage is within free tier
  const allWithinFree =
    overview &&
    (overview.spans ?? 0) <= (overview.free_spans_total ?? 0) &&
    (overview.bytes ?? 0) <= (overview.free_bytes_total ?? 0) &&
    (overview.scores ?? 0) <= (overview.free_scores_total ?? 0)

  // Pricing tiers (per 100K spans, per GB, per 1K scores)
  const pricingTiers = [
    {
      dimension: 'Spans',
      freeTier: '1M spans/month',
      overage: '$0.50 per 100K',
    },
    {
      dimension: 'Data Processed',
      freeTier: '1 GB/month',
      overage: '$3.00 per GB',
    },
    {
      dimension: 'Scores',
      freeTier: '10K/month',
      overage: '$0.20 per 1K',
    },
  ]

  return (
    <div className={cn('space-y-6', className)}>
      {errorMessage ? (
        <Card>
          <CardContent className="py-6">
            <div className="flex items-center gap-2 text-destructive">
              <AlertCircle className="h-4 w-4" />
              <p className="text-sm">{errorMessage}</p>
            </div>
          </CardContent>
        </Card>
      ) : (
        <>
          {/* Current Period Cost */}
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle>Current Billing Period</CardTitle>
                  <CardDescription>
                    {formatPeriodDate(overview?.period_start)} -{' '}
                    {formatPeriodDate(overview?.period_end)}
                  </CardDescription>
                </div>
                <div className="flex items-center gap-2">
                  {isLoading ? (
                    <Skeleton className="h-6 w-16" />
                  ) : (
                    <>
                      <Badge variant="outline">Free Plan</Badge>
                      <Button size="sm">
                        <Sparkles className="mr-2 h-4 w-4" />
                        Upgrade Plan
                      </Button>
                    </>
                  )}
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <div className="flex items-center gap-4">
                <div className="flex h-14 w-14 items-center justify-center rounded-full bg-primary/10">
                  <DollarSign className="h-7 w-7 text-primary" />
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Estimated Cost</p>
                  {isLoading ? (
                    <Skeleton className="h-9 w-28 mt-1" />
                  ) : (
                    <p className="text-3xl font-bold">
                      {formatCurrency(overview?.estimated_cost ?? 0)}
                    </p>
                  )}
                </div>
              </div>

              {/* Free Tier Status */}
              {!isLoading && allWithinFree && (
                <div className="flex items-center gap-2 mt-6 p-3 bg-green-50 dark:bg-green-950/20 border border-green-200 dark:border-green-900 rounded-lg">
                  <Check className="h-4 w-4 text-green-600" />
                  <span className="text-sm text-green-700 dark:text-green-400">
                    All usage is within your free tier — no charges this period
                  </span>
                </div>
              )}
            </CardContent>
          </Card>

          {/* Pricing Tiers */}
          <Card>
            <CardHeader>
              <CardTitle>Pricing</CardTitle>
              <CardDescription>
                Usage-based billing: Spans + Data + Scores
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Dimension</TableHead>
                    <TableHead>Free Tier</TableHead>
                    <TableHead>After Free Tier</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {pricingTiers.map((tier) => (
                    <TableRow key={tier.dimension}>
                      <TableCell className="font-medium">
                        {tier.dimension}
                      </TableCell>
                      <TableCell className="text-muted-foreground">
                        {tier.freeTier}
                      </TableCell>
                      <TableCell>{tier.overage}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>

              <div className="mt-4 p-3 bg-muted/50 rounded-lg">
                <p className="text-sm text-muted-foreground">
                  <strong>Example:</strong> 2M spans + 3GB + 50K scores ={' '}
                  <span className="font-mono">(10×$0.50) + (2×$3) + (40×$0.20) = $19/month</span>
                </p>
              </div>
            </CardContent>
          </Card>

          {/* Payment Method */}
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle>Payment Method</CardTitle>
                  <CardDescription>
                    Manage your payment methods
                  </CardDescription>
                </div>
                <Button variant="outline" size="sm">
                  <Plus className="mr-2 h-4 w-4" />
                  Add Card
                </Button>
              </div>
            </CardHeader>
            <CardContent>
              {/* Placeholder - No payment method */}
              <div className="flex items-center justify-center py-8 border-2 border-dashed rounded-lg">
                <div className="text-center">
                  <CreditCard className="h-10 w-10 text-muted-foreground mx-auto mb-3" />
                  <p className="text-sm text-muted-foreground">
                    No payment method on file
                  </p>
                  <p className="text-xs text-muted-foreground mt-1">
                    Add a card to enable paid features
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Invoice History */}
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle>Invoice History</CardTitle>
                  <CardDescription>
                    View and download past invoices
                  </CardDescription>
                </div>
                <Button variant="ghost" size="sm">
                  View All
                  <ExternalLink className="ml-2 h-4 w-4" />
                </Button>
              </div>
            </CardHeader>
            <CardContent>
              {/* Placeholder - No invoices */}
              <div className="flex items-center justify-center py-8 border-2 border-dashed rounded-lg">
                <div className="text-center">
                  <FileText className="h-10 w-10 text-muted-foreground mx-auto mb-3" />
                  <p className="text-sm text-muted-foreground">
                    No invoices yet
                  </p>
                  <p className="text-xs text-muted-foreground mt-1">
                    Invoices will appear here after your first billing cycle
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>
        </>
      )}
    </div>
  )
}
