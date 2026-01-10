'use client'

import * as React from 'react'
import { Database, HardDrive, CheckCircle, DollarSign } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'
import type { UsageOverview } from '../types'
import { formatBytes, formatNumber } from '../types'

interface UsageSummaryCardsProps {
  usage: UsageOverview | null
  isLoading?: boolean
  error?: string | null
  className?: string
}

function formatCurrency(value: number): string {
  if (value < 0.01) return '$0.00'
  if (value < 1) return `$${value.toFixed(2)}`
  if (value < 100) return `$${value.toFixed(2)}`
  if (value < 1000) return `$${value.toFixed(1)}`
  if (value < 10000) return `$${(value / 1000).toFixed(1)}k`
  return `$${(value / 1000).toFixed(0)}k`
}

interface UsageCardProps {
  title: string
  value: string
  used: number
  total: number
  freeRemaining: string
  icon: React.ElementType
  loading?: boolean
}

function UsageCard({
  title,
  value,
  used,
  total,
  freeRemaining,
  icon: Icon,
  loading,
}: UsageCardProps) {
  const percentage = total > 0 ? Math.min((used / total) * 100, 100) : 0

  if (loading) {
    return (
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <Skeleton className="h-4 w-[100px]" />
          <Skeleton className="h-4 w-4 rounded" />
        </CardHeader>
        <CardContent>
          <Skeleton className="h-8 w-[120px] mb-4" />
          <Skeleton className="h-2 w-full mb-2" />
          <Skeleton className="h-4 w-[140px]" />
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">{title}</CardTitle>
        <Icon className="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">{value}</div>
        <div className="mt-4 space-y-2">
          <Progress value={percentage} className="h-2" />
          <p className="text-xs text-muted-foreground">
            {freeRemaining} free remaining ({percentage.toFixed(1)}% used)
          </p>
        </div>
      </CardContent>
    </Card>
  )
}

export function UsageSummaryCards({
  usage,
  isLoading,
  error,
  className,
}: UsageSummaryCardsProps) {
  if (error) {
    return (
      <div className={cn('grid gap-4 md:grid-cols-2 lg:grid-cols-4', className)}>
        {[1, 2, 3, 4].map((i) => (
          <Card key={i}>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium text-destructive">
                Error
              </CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-xs text-muted-foreground">{error}</p>
            </CardContent>
          </Card>
        ))}
      </div>
    )
  }

  const spansUsed = usage?.spans ?? 0
  const spansTotal = usage?.free_spans_total ?? 1_000_000
  const spansRemaining = usage?.free_spans_remaining ?? spansTotal

  const bytesUsed = usage?.bytes ?? 0
  const bytesTotal = usage?.free_bytes_total ?? 1073741824 // 1 GB
  const bytesRemaining = usage?.free_bytes_remaining ?? bytesTotal

  const scoresUsed = usage?.scores ?? 0
  const scoresTotal = usage?.free_scores_total ?? 10_000
  const scoresRemaining = usage?.free_scores_remaining ?? scoresTotal

  return (
    <div className={cn('grid gap-4 md:grid-cols-2 lg:grid-cols-4', className)}>
      <UsageCard
        title="Spans"
        value={formatNumber(spansUsed)}
        used={spansUsed}
        total={spansTotal}
        freeRemaining={formatNumber(spansRemaining)}
        icon={Database}
        loading={isLoading}
      />
      <UsageCard
        title="Data Processed"
        value={formatBytes(bytesUsed)}
        used={bytesUsed}
        total={bytesTotal}
        freeRemaining={formatBytes(bytesRemaining)}
        icon={HardDrive}
        loading={isLoading}
      />
      <UsageCard
        title="Scores"
        value={formatNumber(scoresUsed)}
        used={scoresUsed}
        total={scoresTotal}
        freeRemaining={formatNumber(scoresRemaining)}
        icon={CheckCircle}
        loading={isLoading}
      />
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Estimated Cost</CardTitle>
          <DollarSign className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <>
              <Skeleton className="h-8 w-[100px] mb-4" />
              <Skeleton className="h-4 w-[160px]" />
            </>
          ) : (
            <>
              <div className="text-2xl font-bold">
                {formatCurrency(usage?.estimated_cost ?? 0)}
              </div>
              <p className="text-xs text-muted-foreground mt-4">
                Current billing period
              </p>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
