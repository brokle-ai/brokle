'use client'

import { Badge } from '@/components/ui/badge'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { MessageSquare } from 'lucide-react'
import { cn } from '@/lib/utils'
import type { Score, ScoreDataType, ScoreSource } from '../types'

interface ScoreValueCellProps {
  score: Score
  showName?: boolean
  className?: string
}

/**
 * Formats the score value based on data type.
 * - Numeric: 4 decimal places
 * - Boolean: "True" / "False"
 * - Categorical: String value as-is
 */
function formatValue(score: Score): string {
  switch (score.data_type) {
    case 'BOOLEAN':
      return score.value === 1 ? 'True' : 'False'
    case 'CATEGORICAL':
      return score.string_value ?? '-'
    case 'NUMERIC':
    default:
      return score.value?.toFixed(4) ?? '-'
  }
}

/**
 * Get badge variant based on data type and value.
 * - Boolean: green for True, red for False
 * - Numeric: color gradient based on value
 * - Categorical: neutral outline
 */
function getValueVariant(
  dataType: ScoreDataType,
  value: number | undefined
): 'default' | 'secondary' | 'destructive' | 'outline' {
  if (dataType === 'BOOLEAN') {
    return value === 1 ? 'default' : 'destructive'
  }
  if (dataType === 'CATEGORICAL') {
    return 'outline'
  }
  // Numeric: gradient based on 0-1 range
  if (value !== undefined) {
    if (value >= 0.8) return 'default'
    if (value >= 0.5) return 'secondary'
    return 'destructive'
  }
  return 'outline'
}

/**
 * Get source badge styling.
 * - human: blue (manual annotation)
 * - code: green (automated/SDK)
 * - llm: purple (AI-generated)
 */
function getSourceStyles(source: ScoreSource): { bg: string; text: string } {
  switch (source) {
    case 'human':
      return { bg: 'bg-blue-100 dark:bg-blue-900/30', text: 'text-blue-700 dark:text-blue-300' }
    case 'code':
      return { bg: 'bg-green-100 dark:bg-green-900/30', text: 'text-green-700 dark:text-green-300' }
    case 'llm':
      return { bg: 'bg-purple-100 dark:bg-purple-900/30', text: 'text-purple-700 dark:text-purple-300' }
    default:
      return { bg: 'bg-muted', text: 'text-muted-foreground' }
  }
}

function getSourceLabel(source: ScoreSource): string {
  switch (source) {
    case 'human':
      return 'Human'
    case 'code':
      return 'SDK'
    case 'llm':
      return 'LLM'
    default:
      return source
  }
}

export function ScoreValueCell({ score, showName = false, className }: ScoreValueCellProps) {
  const valueVariant = getValueVariant(score.data_type, score.value)
  const sourceStyles = getSourceStyles(score.source)
  const formattedValue = formatValue(score)

  return (
    <div className={cn('flex items-center gap-2', className)}>
      {showName && (
        <span className="text-sm text-muted-foreground truncate max-w-[120px]">
          {score.name}
        </span>
      )}

      {/* Value badge */}
      <Badge variant={valueVariant} size="sm">
        {formattedValue}
      </Badge>

      {/* Source badge */}
      <span
        className={cn(
          'inline-flex items-center px-1.5 py-0.5 text-xs font-medium rounded',
          sourceStyles.bg,
          sourceStyles.text
        )}
      >
        {getSourceLabel(score.source)}
      </span>

      {/* Reason indicator */}
      {score.reason && (
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger asChild>
              <MessageSquare className="h-3.5 w-3.5 text-muted-foreground cursor-help" />
            </TooltipTrigger>
            <TooltipContent side="top" className="max-w-xs">
              <p className="text-sm">{score.reason}</p>
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      )}
    </div>
  )
}

// Export individual formatters for use in tables
export { formatValue, getValueVariant, getSourceStyles, getSourceLabel }
