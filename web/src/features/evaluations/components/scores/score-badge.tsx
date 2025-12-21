'use client'

import { Badge } from '@/components/ui/badge'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import type { Score } from '../../types'

interface ScoreBadgeProps {
  score: Score
  showReason?: boolean
}

export function ScoreBadge({ score, showReason = false }: ScoreBadgeProps) {
  const getVariant = () => {
    if (score.data_type === 'BOOLEAN') {
      return score.value === 1 ? 'default' : 'destructive'
    }
    if (score.value !== undefined) {
      if (score.value >= 0.8) return 'default'
      if (score.value >= 0.5) return 'secondary'
      return 'destructive'
    }
    return 'outline'
  }

  const displayValue = () => {
    if (score.data_type === 'BOOLEAN') {
      return score.value === 1 ? 'Yes' : 'No'
    }
    if (score.data_type === 'CATEGORICAL') {
      return score.string_value
    }
    return score.value?.toFixed(2)
  }

  const badge = (
    <Badge variant={getVariant()}>
      {score.name}: {displayValue()}
    </Badge>
  )

  if (showReason && score.reason) {
    return (
      <Tooltip>
        <TooltipTrigger asChild>{badge}</TooltipTrigger>
        <TooltipContent>
          <p className="max-w-xs">{score.reason}</p>
        </TooltipContent>
      </Tooltip>
    )
  }

  return badge
}
