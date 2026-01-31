'use client'

import * as React from 'react'
import { Star, AlertCircle, Pencil } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { cn } from '@/lib/utils'
import { CollapsibleSection } from './collapsible-section'
import { useTraceScoresQuery } from '../../hooks/use-trace-scores'
import { useUpdateAnnotation, useDeleteAnnotation } from '../../hooks/use-annotations'
import { useScoreConfigsQuery } from '@/features/scores'
import type { Score, Span } from '../../data/schema'
import type { Score as ScoresFeatureScore, ScoreConfig } from '@/features/scores/types'
import { ScoreTag } from '@/features/scores/components/score-tag'
import { AnnotationFormDialog } from '@/features/scores/components/annotation/annotation-form-dialog'

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
// Type Conversion Helpers
// ============================================================================

/**
 * Convert traces/data/schema Score to scores/types Score
 * The traces schema uses different field names
 */
function toScoresFeatureScore(score: Score): ScoresFeatureScore {
  return {
    id: score.id,
    project_id: score.project_id,
    trace_id: score.trace_id,
    span_id: score.span_id,
    name: score.name,
    value: score.value,
    string_value: score.string_value,
    data_type: score.data_type as ScoresFeatureScore['data_type'],
    source: mapSource(score.source),
    reason: score.comment, // traces schema uses 'comment', scores uses 'reason'
    metadata: score.evaluator_config as Record<string, unknown>,
    timestamp: score.timestamp instanceof Date
      ? score.timestamp.toISOString()
      : String(score.timestamp),
  }
}

/**
 * Map source values between schemas
 */
function mapSource(source: string): ScoresFeatureScore['source'] {
  const sourceMap: Record<string, ScoresFeatureScore['source']> = {
    API: 'code',
    api: 'code',
    code: 'code',
    EVAL: 'llm',
    eval: 'llm',
    llm: 'llm',
    ANNOTATION: 'human',
    annotation: 'human',
    human: 'human',
  }
  return sourceMap[source] || 'code'
}

/**
 * Check if a score is editable (human annotation)
 */
function isEditableScore(score: Score): boolean {
  const source = score.source.toLowerCase()
  return source === 'annotation' || source === 'human'
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
// ScoreItem Component - Single score with edit capability
// ============================================================================

interface ScoreItemProps {
  score: Score
  config?: ScoreConfig
  isEditable: boolean
  onEdit: (score: Score) => void
  onDelete: (scoreId: string) => void
}

function ScoreItem({ score, config, isEditable, onEdit, onDelete }: ScoreItemProps) {
  const convertedScore = toScoresFeatureScore(score)

  return (
    <div className="group flex items-center gap-2 py-1">
      <ScoreTag
        score={convertedScore}
        onDelete={isEditable ? () => onDelete(score.id) : undefined}
      />

      {isEditable && (
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="ghost"
              size="icon"
              className="h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity"
              onClick={() => onEdit(score)}
            >
              <Pencil className="h-3 w-3" />
              <span className="sr-only">Edit</span>
            </Button>
          </TooltipTrigger>
          <TooltipContent>Edit annotation</TooltipContent>
        </Tooltip>
      )}
    </div>
  )
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
  const { data: scoreConfigs } = useScoreConfigsQuery(projectId)

  // Edit dialog state
  const [editingScore, setEditingScore] = React.useState<Score | null>(null)

  // Mutations
  const updateMutation = useUpdateAnnotation(projectId, traceId)
  const deleteMutation = useDeleteAnnotation(projectId, traceId)

  // Find config for a score by name
  const getConfigForScore = React.useCallback((scoreName: string): ScoreConfig | undefined => {
    return scoreConfigs?.configs?.find(config => config.name === scoreName)
  }, [scoreConfigs])

  // Handle save from edit dialog
  const handleSave = React.useCallback((data: { value?: number | null; string_value?: string | null; reason?: string | null }) => {
    if (!editingScore) return

    updateMutation.mutate(
      { scoreId: editingScore.id, data },
      { onSuccess: () => setEditingScore(null) }
    )
  }, [editingScore, updateMutation])

  // Handle delete
  const handleDelete = React.useCallback(() => {
    if (!editingScore) return

    deleteMutation.mutate(editingScore.id, {
      onSuccess: () => setEditingScore(null)
    })
  }, [editingScore, deleteMutation])

  // Handle delete from tag (without opening dialog)
  const handleDeleteDirect = React.useCallback((scoreId: string) => {
    deleteMutation.mutate(scoreId)
  }, [deleteMutation])

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
    <>
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
                <ScoreItem
                  key={score.id}
                  score={score}
                  config={getConfigForScore(score.name)}
                  isEditable={isEditableScore(score)}
                  onEdit={setEditingScore}
                  onDelete={handleDeleteDirect}
                />
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
                <ScoreItem
                  key={score.id}
                  score={score}
                  config={getConfigForScore(score.name)}
                  isEditable={isEditableScore(score)}
                  onEdit={setEditingScore}
                  onDelete={handleDeleteDirect}
                />
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

      {/* Edit dialog */}
      {editingScore && (
        <AnnotationFormDialog
          score={toScoresFeatureScore(editingScore)}
          config={getConfigForScore(editingScore.name)}
          open={!!editingScore}
          onOpenChange={(open) => !open && setEditingScore(null)}
          onSave={handleSave}
          onDelete={handleDelete}
          isSaving={updateMutation.isPending}
          isDeleting={deleteMutation.isPending}
        />
      )}
    </>
  )
}
