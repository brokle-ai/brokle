'use client'

import { cn } from '@/lib/utils'
import { calculateScorePercentile } from '../../lib/calculate-diff'

interface ScoreProgressBarProps {
  value: number
  min: number
  max: number
  className?: string
}

/**
 * Visual progress bar showing score position within min/max range
 * Based on Phoenix pattern: calculateAnnotationScorePercentile
 */
export function ScoreProgressBar({
  value,
  min,
  max,
  className,
}: ScoreProgressBarProps) {
  const percentile = calculateScorePercentile(value, min, max)

  return (
    <div
      className={cn(
        'h-1.5 w-full bg-muted rounded-full overflow-hidden',
        className
      )}
    >
      <div
        className="h-full bg-primary rounded-full transition-all duration-300"
        style={{ width: `${percentile}%` }}
      />
    </div>
  )
}
