'use client'

import { TrendingUp, TrendingDown, Equal } from 'lucide-react'
import { cn } from '@/lib/utils'
import { classifyDiffs, type DiffDisplayConfig } from '../../lib/calculate-diff'
import type { ScoreComparisonRow } from '../../types'

interface ComparisonSummaryProps {
  scoreRows: ScoreComparisonRow[]
  baselineId: string
  config?: DiffDisplayConfig
  className?: string
}

/**
 * Summary counters showing improved/regressed/unchanged counts
 * Based on Phoenix pattern: thumbs up/down icons with color coding
 */
export function ComparisonSummary({
  scoreRows,
  baselineId,
  config,
  className,
}: ComparisonSummaryProps) {
  const classification = classifyDiffs(scoreRows, baselineId, config)
  const total = classification.improved + classification.regressed + classification.unchanged

  if (total === 0) return null

  return (
    <div className={cn('flex items-center gap-4 text-sm', className)}>
      {classification.improved > 0 && (
        <div className="inline-flex items-center gap-1.5 text-green-600 dark:text-green-400">
          <TrendingUp className="h-4 w-4" />
          <span className="font-medium">{classification.improved}</span>
          <span className="text-muted-foreground">improved</span>
        </div>
      )}

      {classification.regressed > 0 && (
        <div className="inline-flex items-center gap-1.5 text-red-600 dark:text-red-400">
          <TrendingDown className="h-4 w-4" />
          <span className="font-medium">{classification.regressed}</span>
          <span className="text-muted-foreground">regressed</span>
        </div>
      )}

      {classification.unchanged > 0 && (
        <div className="inline-flex items-center gap-1.5 text-muted-foreground">
          <Equal className="h-4 w-4" />
          <span className="font-medium">{classification.unchanged}</span>
          <span>unchanged</span>
        </div>
      )}
    </div>
  )
}
