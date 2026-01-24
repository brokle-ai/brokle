'use client'

import { TrendingUp, TrendingDown, HelpCircle } from 'lucide-react'
import { cn } from '@/lib/utils'
import { calculateDiffPercentage } from '../../lib/calculate-diff'

interface DeltaPercentageProps {
  baseline: number
  current: number
  preferNegativeDiff?: boolean
  className?: string
  showIcon?: boolean
}

/**
 * Display percentage change with arrow indicator
 * Based on Phoenix pattern: +8.5% with arrow
 *
 * Handles edge cases:
 * - baseline=0, current=0 → renders nothing (no change)
 * - baseline=0, current≠0 → renders "N/A" (undefined percentage)
 */
export function DeltaPercentage({
  baseline,
  current,
  preferNegativeDiff = false,
  className,
  showIcon = true,
}: DeltaPercentageProps) {
  const result = calculateDiffPercentage(baseline, current)

  // Both values are 0 → no change to display
  if (!result) {
    return null
  }

  // Undefined percentage (baseline was 0 but current is not)
  if (result.isUndefined) {
    const isPositiveDirection = result.direction === '+'
    const isImprovement = preferNegativeDiff ? !isPositiveDirection : isPositiveDirection
    const colorClass = isImprovement
      ? 'text-green-600 dark:text-green-400'
      : 'text-red-600 dark:text-red-400'

    return (
      <span className={cn('inline-flex items-center gap-1 text-xs font-medium', colorClass, className)}>
        {showIcon && <HelpCircle className="h-3 w-3" />}
        <span>N/A</span>
      </span>
    )
  }

  const { percentage, direction } = result

  // Determine if this is an improvement
  const isPositiveDirection = direction === '+'
  const isImprovement = preferNegativeDiff ? !isPositiveDirection : isPositiveDirection

  // Format percentage
  const formattedPercent =
    percentage < 0.1 ? percentage.toFixed(2) : percentage.toFixed(1)

  // Choose styling based on improvement
  const colorClass = isImprovement
    ? 'text-green-600 dark:text-green-400'
    : 'text-red-600 dark:text-red-400'

  const Icon = isPositiveDirection ? TrendingUp : TrendingDown

  return (
    <span className={cn('inline-flex items-center gap-1 text-xs font-medium', colorClass, className)}>
      {showIcon && <Icon className="h-3 w-3" />}
      <span>
        {direction}
        {formattedPercent}%
      </span>
    </span>
  )
}
