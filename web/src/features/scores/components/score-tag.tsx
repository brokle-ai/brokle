'use client'

import { useState } from 'react'
import { MessageSquareMore, Trash2, X } from 'lucide-react'
import { cn } from '@/lib/utils'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import {
  HoverCard,
  HoverCardContent,
  HoverCardTrigger,
} from '@/components/ui/hover-card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import type { Score } from '../types'
import { getScoreTagClasses, getDataTypeIndicator, getSourceIndicator } from '../lib/score-colors'

interface ScoreTagProps {
  score: Score
  onDelete?: (scoreId: string) => void
  showDetails?: boolean
  compact?: boolean
  className?: string
}

/**
 * Score Tag Component (Opik pattern)
 *
 * Displays a compact score indicator with:
 * - Deterministic color based on score name
 * - Compact label + value display
 * - Hover card with full details
 * - Delete button on hover
 * - Reason tooltip with icon
 */
export function ScoreTag({
  score,
  onDelete,
  showDetails = true,
  compact = false,
  className,
}: ScoreTagProps) {
  const [isHovered, setIsHovered] = useState(false)
  const { containerClass, indicatorClass, textClass } = getScoreTagClasses(score.name)
  const { symbol: typeSymbol } = getDataTypeIndicator(score.data_type)
  const { label: sourceLabel, className: sourceClassName } = getSourceIndicator(score.source)

  // Format value based on data type
  const formattedValue = formatScoreValue(score)

  const tagContent = (
    <div
      className={cn(
        'inline-flex items-center gap-1.5 px-2 py-1 rounded-full border transition-all',
        containerClass,
        onDelete && isHovered && 'pr-1',
        className
      )}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      {/* Color indicator dot */}
      <span className={cn('w-2 h-2 rounded-full flex-shrink-0', indicatorClass)} />

      {/* Score name (truncated) */}
      <span className={cn('text-xs font-medium truncate', textClass, compact ? 'max-w-[60px]' : 'max-w-[100px]')}>
        {score.name}
      </span>

      {/* Value */}
      <span className={cn('text-xs font-semibold', textClass)}>
        {formattedValue}
      </span>

      {/* Reason indicator */}
      {score.reason && (
        <TooltipProvider delayDuration={200}>
          <Tooltip>
            <TooltipTrigger asChild>
              <MessageSquareMore className={cn('h-3 w-3 flex-shrink-0', textClass)} />
            </TooltipTrigger>
            <TooltipContent side="top" className="max-w-xs">
              <p className="text-sm">{score.reason}</p>
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      )}

      {/* Delete button (on hover) */}
      {onDelete && isHovered && (
        <Button
          variant="ghost"
          size="icon"
          className="h-4 w-4 p-0 hover:bg-transparent"
          onClick={(e) => {
            e.stopPropagation()
            onDelete(score.id)
          }}
        >
          <X className="h-3 w-3 text-muted-foreground hover:text-destructive" />
          <span className="sr-only">Delete score</span>
        </Button>
      )}
    </div>
  )

  // If showDetails is false, just return the tag
  if (!showDetails) {
    return tagContent
  }

  // Wrap with HoverCard for detailed view
  return (
    <HoverCard openDelay={300}>
      <HoverCardTrigger asChild>
        {tagContent}
      </HoverCardTrigger>
      <HoverCardContent className="w-64" align="start">
        <ScoreTagDetails score={score} onDelete={onDelete} />
      </HoverCardContent>
    </HoverCard>
  )
}

/**
 * Detailed score information shown in hover card
 */
function ScoreTagDetails({
  score,
  onDelete,
}: {
  score: Score
  onDelete?: (scoreId: string) => void
}) {
  const { indicatorClass, textClass } = getScoreTagClasses(score.name)
  const { symbol: typeSymbol, label: typeLabel } = getDataTypeIndicator(score.data_type)
  const { label: sourceLabel, className: sourceClassName } = getSourceIndicator(score.source)

  return (
    <div className="space-y-3">
      {/* Header with name and value */}
      <div className="flex items-start justify-between gap-2">
        <div className="flex items-center gap-2 min-w-0">
          <span className={cn('w-2.5 h-2.5 rounded-full flex-shrink-0', indicatorClass)} />
          <span className={cn('font-medium truncate', textClass)}>{score.name}</span>
        </div>
        {onDelete && (
          <Button
            variant="ghost"
            size="icon"
            className="h-6 w-6 flex-shrink-0"
            onClick={() => onDelete(score.id)}
          >
            <Trash2 className="h-3.5 w-3.5 text-muted-foreground hover:text-destructive" />
            <span className="sr-only">Delete score</span>
          </Button>
        )}
      </div>

      {/* Value display */}
      <div className="flex items-center gap-2">
        <span className="text-2xl font-bold">{formatScoreValue(score)}</span>
        <span className="text-sm text-muted-foreground">{typeSymbol} {typeLabel}</span>
      </div>

      {/* Source badge */}
      <div className="flex items-center gap-2">
        <span className="text-xs text-muted-foreground">Source:</span>
        <span className={cn('inline-flex items-center px-2 py-0.5 text-xs font-medium rounded', sourceClassName)}>
          {sourceLabel}
        </span>
      </div>

      {/* Reason (if present) */}
      {score.reason && (
        <div className="space-y-1">
          <span className="text-xs text-muted-foreground">Reason:</span>
          <p className="text-sm text-foreground leading-relaxed">{score.reason}</p>
        </div>
      )}

      {/* Timestamp */}
      <div className="text-xs text-muted-foreground">
        {new Date(score.timestamp).toLocaleString()}
      </div>
    </div>
  )
}

/**
 * Format score value based on data type
 */
function formatScoreValue(score: Score): string {
  switch (score.data_type) {
    case 'BOOLEAN':
      if (score.value === undefined || score.value === null) return '-'
      return score.value === 1 ? 'True' : 'False'
    case 'CATEGORICAL':
      return score.string_value ?? '-'
    case 'NUMERIC':
    default:
      if (score.value === undefined || score.value === null) return '-'
      // Show up to 4 decimal places, but trim trailing zeros
      const fixed = score.value.toFixed(4)
      return parseFloat(fixed).toString()
  }
}

/**
 * Compact score tag list for displaying multiple scores
 */
export function ScoreTagList({
  scores,
  onDelete,
  maxVisible = 3,
  className,
}: {
  scores: Score[]
  onDelete?: (scoreId: string) => void
  maxVisible?: number
  className?: string
}) {
  const visibleScores = scores.slice(0, maxVisible)
  const remainingCount = scores.length - maxVisible

  if (scores.length === 0) {
    return (
      <span className="text-sm text-muted-foreground">No scores</span>
    )
  }

  return (
    <div className={cn('flex flex-wrap items-center gap-1.5', className)}>
      {visibleScores.map((score) => (
        <ScoreTag
          key={score.id}
          score={score}
          onDelete={onDelete}
          compact
        />
      ))}
      {remainingCount > 0 && (
        <Badge variant="secondary" className="text-xs">
          +{remainingCount}
        </Badge>
      )}
    </div>
  )
}
