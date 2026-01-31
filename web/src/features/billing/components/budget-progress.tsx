'use client'

import * as React from 'react'
import { AlertCircle, CheckCircle, TrendingUp } from 'lucide-react'
import { Progress } from '@/components/ui/progress'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'
import type { UsageBudget, AlertDimension } from '../types'
import { formatBytes, formatNumber } from '../types'

interface BudgetProgressProps {
  budget: UsageBudget
  className?: string
  compact?: boolean
}

interface DimensionProgressProps {
  label: string
  dimension: AlertDimension
  current: number
  limit: number | undefined
  formatValue: (value: number) => string
  compact?: boolean
}

function getProgressColor(percentage: number): string {
  if (percentage >= 100) return 'bg-destructive'
  if (percentage >= 80) return 'bg-orange-500'
  if (percentage >= 50) return 'bg-yellow-500'
  return 'bg-primary'
}

function getProgressTextColor(percentage: number): string {
  if (percentage >= 100) return 'text-destructive'
  if (percentage >= 80) return 'text-orange-500'
  if (percentage >= 50) return 'text-yellow-500'
  return 'text-muted-foreground'
}

function DimensionProgress({
  label,
  current,
  limit,
  formatValue,
  compact,
}: DimensionProgressProps) {
  if (!limit) return null

  const percentage = Math.min((current / limit) * 100, 100)
  const isOverLimit = current >= limit

  return (
    <div className={cn('space-y-1.5', compact && 'space-y-1')}>
      <div className="flex items-center justify-between">
        <span className={cn('text-sm font-medium', compact && 'text-xs')}>
          {label}
        </span>
        <span className={cn('text-sm', compact && 'text-xs', getProgressTextColor(percentage))}>
          {formatValue(current)} / {formatValue(limit)}
        </span>
      </div>
      <div className="relative">
        <Progress
          value={percentage}
          className={cn('h-2', compact && 'h-1.5')}
        />
        <div
          className={cn(
            'absolute inset-0 rounded-full',
            getProgressColor(percentage)
          )}
          style={{ width: `${percentage}%`, opacity: 0.2 }}
        />
      </div>
      {!compact && (
        <div className="flex items-center justify-between text-xs text-muted-foreground">
          <span>{percentage.toFixed(1)}% used</span>
          {isOverLimit && (
            <Badge variant="destructive" className="h-5 text-xs">
              Limit reached
            </Badge>
          )}
        </div>
      )}
    </div>
  )
}

function formatCurrency(value: number | string): string {
  // Convert string to number (decimal fields come as strings from backend)
  const numValue = typeof value === 'string' ? parseFloat(value) : value

  if (numValue < 0.01) return '$0.00'
  if (numValue < 100) return `$${numValue.toFixed(2)}`
  if (numValue < 1000) return `$${numValue.toFixed(1)}`
  return `$${(numValue / 1000).toFixed(1)}k`
}

export function BudgetProgress({ budget, className, compact }: BudgetProgressProps) {
  const dimensions = [
    {
      label: 'Spans',
      dimension: 'spans' as AlertDimension,
      current: budget.current_spans,
      limit: budget.span_limit,
      formatValue: formatNumber,
    },
    {
      label: 'Data',
      dimension: 'bytes' as AlertDimension,
      current: budget.current_bytes,
      limit: budget.bytes_limit,
      formatValue: formatBytes,
    },
    {
      label: 'Scores',
      dimension: 'scores' as AlertDimension,
      current: budget.current_scores,
      limit: budget.score_limit,
      formatValue: formatNumber,
    },
    {
      label: 'Cost',
      dimension: 'cost' as AlertDimension,
      current: parseFloat(budget.current_cost),
      limit: budget.cost_limit ? parseFloat(budget.cost_limit) : undefined,
      formatValue: formatCurrency,
    },
  ].filter((d) => d.limit !== undefined && d.limit !== null)

  if (dimensions.length === 0) {
    return (
      <div className={cn('text-sm text-muted-foreground', className)}>
        No limits configured
      </div>
    )
  }

  // Calculate overall status
  const maxPercentage = Math.max(
    ...dimensions.map((d) => (d.current / (d.limit ?? 1)) * 100)
  )

  const StatusIcon = maxPercentage >= 100
    ? AlertCircle
    : maxPercentage >= 80
    ? TrendingUp
    : CheckCircle

  const statusColor = maxPercentage >= 100
    ? 'text-destructive'
    : maxPercentage >= 80
    ? 'text-orange-500'
    : 'text-green-500'

  return (
    <div className={cn('space-y-4', compact && 'space-y-2', className)}>
      {!compact && (
        <div className="flex items-center gap-2">
          <StatusIcon className={cn('h-4 w-4', statusColor)} />
          <span className={cn('text-sm font-medium', statusColor)}>
            {maxPercentage >= 100
              ? 'Budget exceeded'
              : maxPercentage >= 80
              ? 'Approaching limit'
              : 'Within budget'}
          </span>
        </div>
      )}
      {dimensions.map((dimension) => (
        <DimensionProgress
          key={dimension.dimension}
          {...dimension}
          compact={compact}
        />
      ))}
    </div>
  )
}
