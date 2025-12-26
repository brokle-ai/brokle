'use client'

import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { getDiffDisplay } from '../lib/calculate-diff'
import type { ExperimentScoreDiff } from '../types'

interface DiffLabelProps {
  diff: ExperimentScoreDiff | undefined
  preferNegativeDiff?: boolean
  className?: string
  variant?: 'badge' | 'inline'
}

export function DiffLabel({
  diff,
  preferNegativeDiff = false,
  className,
  variant = 'badge',
}: DiffLabelProps) {
  const display = getDiffDisplay(diff, { preferNegativeDiff })

  if (!display) return null

  const styleClasses = {
    positive:
      'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
    negative: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400',
    neutral: 'bg-muted text-muted-foreground',
  }

  if (variant === 'inline') {
    return (
      <span
        className={cn(
          'text-xs font-medium',
          display.style === 'positive' && 'text-green-600 dark:text-green-400',
          display.style === 'negative' && 'text-red-600 dark:text-red-400',
          display.style === 'neutral' && 'text-muted-foreground',
          className
        )}
      >
        {display.label}
      </span>
    )
  }

  return (
    <Badge
      variant="secondary"
      className={cn(
        'text-xs font-medium px-1.5 py-0.5',
        styleClasses[display.style],
        className
      )}
    >
      {display.label}
    </Badge>
  )
}
