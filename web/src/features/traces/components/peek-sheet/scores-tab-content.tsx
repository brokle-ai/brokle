'use client'

import * as React from 'react'
import { Badge } from '@/components/ui/badge'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import {
  HoverCard,
  HoverCardContent,
  HoverCardTrigger,
} from '@/components/ui/hover-card'
import { Star, AlertCircle, MessageSquare, Braces } from 'lucide-react'
import { cn } from '@/lib/utils'
import { CollapsibleSection } from './collapsible-section'
import { useTraceScoresQuery } from '../../hooks/use-trace-scores'
import type { Score, Span } from '../../data/schema'
import { format, formatDistanceToNow } from 'date-fns'

// ============================================================================
// Types
// ============================================================================

interface ScoresTabContentProps {
  projectId: string
  traceId: string
  spans?: Span[]
}

interface GroupedScores {
  traceScores: Score[]
  spanScores: Array<{
    spanId: string
    spanName: string
    scores: Score[]
  }>
}

// ============================================================================
// Score Display Helpers
// ============================================================================

/**
 * Get badge variant based on score value and type
 * Following patterns:
 * - BOOLEAN: Green for true, red for false
 * - NUMERIC: Green ≥0.8, yellow ≥0.5, red <0.5
 * - CATEGORICAL: Neutral outline
 */
function getScoreBadgeVariant(
  score: Score
): 'default' | 'secondary' | 'destructive' | 'outline' {
  if (score.type === 'BOOLEAN') {
    return score.value === 1 ? 'default' : 'destructive'
  }

  if (score.type === 'NUMERIC' && score.value !== undefined) {
    if (score.value >= 0.8) return 'default'
    if (score.value >= 0.5) return 'secondary'
    return 'destructive'
  }

  return 'outline'
}

/**
 * Get display value for score based on type
 */
function getScoreDisplayValue(score: Score): string {
  if (score.type === 'BOOLEAN') {
    return score.value === 1 ? 'Yes' : 'No'
  }

  if (score.type === 'CATEGORICAL') {
    return score.string_value || '-'
  }

  // NUMERIC
  return score.value !== undefined ? score.value.toFixed(2) : '-'
}

/**
 * Get background class for boolean scores
 */
function getBooleanBgClass(score: Score): string {
  if (score.type !== 'BOOLEAN') return ''
  return score.value === 1
    ? 'bg-green-50 dark:bg-green-950/30'
    : 'bg-red-50 dark:bg-red-950/30'
}

/**
 * Format source for display
 */
function formatSource(source: string): string {
  const sourceMap: Record<string, string> = {
    code: 'Code',
    llm: 'LLM',
    human: 'Human',
    API: 'API',
    ANNOTATION: 'Annotation',
    EVAL: 'Evaluation',
  }
  return sourceMap[source] || source
}

// ============================================================================
// ScoreItem Component - Single score display with hover card
// ============================================================================

interface ScoreItemProps {
  score: Score
}

function ScoreItem({ score }: ScoreItemProps) {
  const variant = getScoreBadgeVariant(score)
  const displayValue = getScoreDisplayValue(score)
  const booleanBg = getBooleanBgClass(score)
  const hasComment = !!score.comment
  const hasMetadata =
    score.evaluator_config && Object.keys(score.evaluator_config).length > 0

  return (
    <HoverCard openDelay={200}>
      <HoverCardTrigger asChild>
        <div
          className={cn(
            'flex items-center justify-between py-2 px-3 rounded-md cursor-pointer transition-colors',
            'hover:bg-muted/50',
            booleanBg
          )}
        >
          <div className="flex items-center gap-2 min-w-0">
            {/* Score badge */}
            <Badge variant={variant} className="shrink-0">
              {score.name}: {displayValue}
            </Badge>

            {/* Source indicator */}
            <span className="text-xs text-muted-foreground capitalize shrink-0">
              {formatSource(score.source)}
            </span>

            {/* Indicators for comment/metadata */}
            <div className="flex items-center gap-1">
              {hasComment && (
                <Tooltip>
                  <TooltipTrigger asChild>
                    <MessageSquare className="h-3 w-3 text-muted-foreground" />
                  </TooltipTrigger>
                  <TooltipContent>Has comment</TooltipContent>
                </Tooltip>
              )}
              {hasMetadata && (
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Braces className="h-3 w-3 text-muted-foreground" />
                  </TooltipTrigger>
                  <TooltipContent>Has metadata</TooltipContent>
                </Tooltip>
              )}
            </div>
          </div>

          {/* Timestamp */}
          <span className="text-xs text-muted-foreground shrink-0 ml-2">
            {formatDistanceToNow(new Date(score.timestamp), { addSuffix: true })}
          </span>
        </div>
      </HoverCardTrigger>

      <HoverCardContent className="w-80" align="start">
        <div className="space-y-3">
          {/* Header */}
          <div className="space-y-1">
            <h4 className="text-sm font-semibold">{score.name}</h4>
            <div className="flex items-center gap-2 text-xs text-muted-foreground">
              <span>Type: {score.type}</span>
              <span>•</span>
              <span>Source: {formatSource(score.source)}</span>
            </div>
          </div>

          {/* Value */}
          <div className="flex items-center gap-2">
            <span className="text-sm text-muted-foreground">Value:</span>
            <Badge variant={variant}>{displayValue}</Badge>
          </div>

          {/* Comment */}
          {hasComment && (
            <div className="space-y-1">
              <div className="flex items-center gap-1 text-xs text-muted-foreground">
                <MessageSquare className="h-3 w-3" />
                <span>Comment</span>
              </div>
              <p className="text-sm line-clamp-3">{score.comment}</p>
            </div>
          )}

          {/* Evaluator info */}
          {score.evaluator_name && (
            <div className="text-xs text-muted-foreground">
              Evaluator: {score.evaluator_name}
              {score.evaluator_version && ` v${score.evaluator_version}`}
            </div>
          )}

          {/* Metadata */}
          {hasMetadata && (
            <div className="space-y-1">
              <div className="flex items-center gap-1 text-xs text-muted-foreground">
                <Braces className="h-3 w-3" />
                <span>Metadata</span>
              </div>
              <pre className="text-xs bg-muted p-2 rounded overflow-x-auto max-h-24">
                {JSON.stringify(score.evaluator_config, null, 2)}
              </pre>
            </div>
          )}

          {/* Timestamp */}
          <div className="text-xs text-muted-foreground border-t pt-2">
            {format(new Date(score.timestamp), 'PPpp')}
          </div>
        </div>
      </HoverCardContent>
    </HoverCard>
  )
}

// ============================================================================
// Grouping Logic
// ============================================================================

/**
 * Group scores by span_id
 * - Trace-level: span_id is null, empty, or equals trace_id
 * - Per-span: grouped by span_id with span name lookup
 */
function groupScoresBySpan(
  scores: Score[],
  traceId: string,
  spans?: Span[]
): GroupedScores {
  const traceScores: Score[] = []
  const spanScoresMap: Map<string, Score[]> = new Map()

  for (const score of scores) {
    // Trace-level scores: span_id is null, empty, or matches trace_id
    if (
      !score.span_id ||
      score.span_id === traceId ||
      score.span_id === score.trace_id
    ) {
      traceScores.push(score)
    } else {
      const existing = spanScoresMap.get(score.span_id) || []
      existing.push(score)
      spanScoresMap.set(score.span_id, existing)
    }
  }

  // Convert map to array with span name lookup
  const spanScores = Array.from(spanScoresMap.entries()).map(
    ([spanId, scoreList]) => {
      const span = spans?.find((s) => s.span_id === spanId)
      return {
        spanId,
        spanName: span?.span_name || `Span ${spanId.substring(0, 8)}...`,
        scores: scoreList,
      }
    }
  )

  return { traceScores, spanScores }
}

// ============================================================================
// Main Component
// ============================================================================

export function ScoresTabContent({
  projectId,
  traceId,
  spans,
}: ScoresTabContentProps) {
  const { data: scores, isLoading, error } = useTraceScoresQuery(projectId, traceId)

  // Loading state
  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="flex flex-col items-center space-y-2">
          <div className="h-6 w-6 animate-spin rounded-full border-2 border-primary border-t-transparent" />
          <p className="text-sm text-muted-foreground">Loading scores...</p>
        </div>
      </div>
    )
  }

  // Error state
  if (error) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-center">
        <AlertCircle className="h-12 w-12 text-muted-foreground/50 mb-4" />
        <p className="text-sm text-destructive">Failed to load scores</p>
        <p className="text-xs text-muted-foreground mt-1">
          {error instanceof Error ? error.message : 'An error occurred'}
        </p>
      </div>
    )
  }

  // Empty state
  if (!scores || scores.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-center">
        <Star className="h-12 w-12 text-muted-foreground/50 mb-4" />
        <p className="text-sm text-muted-foreground">No scores for this trace</p>
        <p className="text-xs text-muted-foreground/70 mt-1">
          Add scores via SDK using{' '}
          <code className="text-xs bg-muted px-1 py-0.5 rounded">
            brokle.score()
          </code>
        </p>
      </div>
    )
  }

  // Group scores
  const { traceScores, spanScores } = groupScoresBySpan(scores, traceId, spans)

  return (
    <div className="space-y-1">
      {/* Trace-level scores */}
      {traceScores.length > 0 && (
        <CollapsibleSection
          title="Trace Scores"
          icon={<Star className="h-4 w-4 text-muted-foreground" />}
          count={traceScores.length}
          defaultExpanded={true}
        >
          <div className="space-y-1">
            {traceScores.map((score) => (
              <ScoreItem key={score.id} score={score} />
            ))}
          </div>
        </CollapsibleSection>
      )}

      {/* Per-span scores */}
      {spanScores.map(({ spanId, spanName, scores: spanScoreList }) => (
        <CollapsibleSection
          key={spanId}
          title={spanName}
          count={spanScoreList.length}
          defaultExpanded={false}
        >
          <div className="space-y-1">
            {spanScoreList.map((score) => (
              <ScoreItem key={score.id} score={score} />
            ))}
          </div>
        </CollapsibleSection>
      ))}

      {/* Edge case: All scores are span-level but no trace scores */}
      {traceScores.length === 0 && spanScores.length > 0 && (
        <p className="text-xs text-muted-foreground px-1 py-2">
          All {scores.length} scores are attached to specific spans
        </p>
      )}
    </div>
  )
}
