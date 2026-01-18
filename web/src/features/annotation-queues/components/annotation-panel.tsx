'use client'

import { useState, useCallback } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Skeleton } from '@/components/ui/skeleton'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible'
import {
  ChevronRight,
  ChevronDown,
  SkipForward,
  CheckCircle,
  AlertCircle,
  Info,
  Loader2,
  Clock,
  Coins,
  Cpu,
} from 'lucide-react'
import { ScoreInputForm } from './score-input-form'
import {
  useClaimNextItemMutation,
  useCompleteItemMutation,
  useSkipItemMutation,
  useReleaseItemMutation,
} from '../hooks/use-annotation-queues'
import { getTraceById, getSpansForTrace, getSpanById } from '@/features/traces/api/traces-api'
import { traceQueryKeys } from '@/features/traces/hooks/trace-query-keys'
import type { Trace, Span } from '@/features/traces/data/schema'
import type { QueueItem, AnnotationQueue, ScoreSubmission } from '../types'

// ============================================================================
// Trace Viewer Components
// ============================================================================

function formatDuration(nanoseconds?: number): string {
  if (nanoseconds == null) return '-'
  const ms = nanoseconds / 1_000_000
  if (ms < 1000) return `${ms.toFixed(0)}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

function formatCost(cost?: number): string {
  if (cost == null) return '-'
  if (cost < 0.01) return `$${cost.toFixed(4)}`
  return `$${cost.toFixed(3)}`
}

function formatTokens(count?: number): string {
  if (count == null) return '-'
  return count.toLocaleString()
}

interface ContentSectionProps {
  title: string
  content: string | null | undefined
  defaultExpanded?: boolean
}

function ContentSection({ title, content, defaultExpanded = true }: ContentSectionProps) {
  const [isOpen, setIsOpen] = useState(defaultExpanded)

  if (!content) return null

  // Try to parse and format JSON content
  let displayContent = content
  let isJson = false
  try {
    const parsed = JSON.parse(content)
    displayContent = JSON.stringify(parsed, null, 2)
    isJson = true
  } catch {
    // Not JSON, use as-is
  }

  return (
    <Collapsible open={isOpen} onOpenChange={setIsOpen}>
      <CollapsibleTrigger className="flex items-center gap-2 w-full py-2 hover:bg-muted/50 rounded-md px-2 -mx-2">
        {isOpen ? (
          <ChevronDown className="h-4 w-4 text-muted-foreground" />
        ) : (
          <ChevronRight className="h-4 w-4 text-muted-foreground" />
        )}
        <span className="text-sm font-medium">{title}</span>
        {isJson && (
          <Badge variant="secondary" className="text-xs ml-auto">JSON</Badge>
        )}
      </CollapsibleTrigger>
      <CollapsibleContent>
        <pre className="mt-2 p-3 bg-muted/50 rounded-md text-xs font-mono overflow-x-auto max-h-[300px] whitespace-pre-wrap break-words">
          {displayContent}
        </pre>
      </CollapsibleContent>
    </Collapsible>
  )
}

interface TraceViewerProps {
  trace: Trace | null | undefined
  spans: Span[]
  objectType: 'trace' | 'span'
  objectId: string
  isLoading: boolean
}

function TraceViewer({ trace, spans, objectType, objectId, isLoading }: TraceViewerProps) {
  if (isLoading) {
    return (
      <div className="space-y-3">
        <Skeleton className="h-6 w-48" />
        <Skeleton className="h-24 w-full" />
        <Skeleton className="h-24 w-full" />
      </div>
    )
  }

  if (!trace) {
    return (
      <div className="text-center py-8 text-muted-foreground">
        <AlertCircle className="h-8 w-8 mx-auto mb-2 opacity-50" />
        <p className="text-sm">Failed to load trace data</p>
      </div>
    )
  }

  // For span type, find the specific span
  const targetSpan = objectType === 'span'
    ? spans.find(s => s.span_id === objectId)
    : spans.find(s => !s.parent_span_id) || spans[0]

  return (
    <div className="space-y-4">
      {/* Metadata Badges */}
      <div className="flex flex-wrap gap-2">
        {trace.model_name && (
          <Badge variant="outline" className="text-xs">
            <Cpu className="h-3 w-3 mr-1" />
            {trace.model_name}
          </Badge>
        )}
        {trace.duration != null && (
          <Badge variant="outline" className="text-xs">
            <Clock className="h-3 w-3 mr-1" />
            {formatDuration(trace.duration)}
          </Badge>
        )}
        {trace.cost != null && (
          <Badge variant="outline" className="text-xs">
            <Coins className="h-3 w-3 mr-1" />
            {formatCost(trace.cost)}
          </Badge>
        )}
        {trace.tokens != null && (
          <Badge variant="outline" className="text-xs">
            {formatTokens(trace.tokens)} tokens
          </Badge>
        )}
        {trace.has_error && (
          <Badge variant="destructive" className="text-xs">
            <AlertCircle className="h-3 w-3 mr-1" />
            Error
          </Badge>
        )}
      </div>

      {/* Tabs for different views */}
      <Tabs defaultValue="preview" className="w-full">
        <TabsList className="h-8">
          <TabsTrigger value="preview" className="text-xs h-7 px-3">Preview</TabsTrigger>
          {spans.length > 1 && (
            <TabsTrigger value="spans" className="text-xs h-7 px-3">
              Spans ({spans.length})
            </TabsTrigger>
          )}
        </TabsList>

        <TabsContent value="preview" className="mt-3 space-y-2">
          {/* Error message if present */}
          {targetSpan?.status_message && (
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription className="font-mono text-xs">
                {targetSpan.status_message}
              </AlertDescription>
            </Alert>
          )}

          {/* Input/Output sections */}
          <ContentSection title="Input" content={targetSpan?.input} defaultExpanded={true} />
          <ContentSection title="Output" content={targetSpan?.output} defaultExpanded={true} />

          {/* Attributes if any */}
          {targetSpan?.attributes && Object.keys(targetSpan.attributes).length > 0 && (
            <ContentSection
              title="Attributes"
              content={JSON.stringify(targetSpan.attributes)}
              defaultExpanded={false}
            />
          )}
        </TabsContent>

        {spans.length > 1 && (
          <TabsContent value="spans" className="mt-3">
            <div className="space-y-2 max-h-[400px] overflow-y-auto">
              {spans.map((span) => (
                <div
                  key={span.span_id}
                  className={`p-3 rounded-md border ${span.span_id === objectId ? 'border-primary bg-primary/5' : 'hover:bg-muted/50'}`}
                >
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-medium truncate">{span.span_name}</span>
                    <div className="flex items-center gap-2">
                      {span.span_type && (
                        <Badge variant="secondary" className="text-xs">{span.span_type}</Badge>
                      )}
                      <span className="text-xs text-muted-foreground">
                        {formatDuration(span.duration)}
                      </span>
                    </div>
                  </div>
                  {span.gen_ai_request_model && (
                    <p className="text-xs text-muted-foreground mt-1">
                      {span.gen_ai_request_model}
                    </p>
                  )}
                </div>
              ))}
            </div>
          </TabsContent>
        )}
      </Tabs>
    </div>
  )
}

// ============================================================================
// Main Component
// ============================================================================

interface AnnotationPanelProps {
  projectId: string
  queue: AnnotationQueue
  currentItem: QueueItem | null
  onItemClaimed: (item: QueueItem) => void
  onItemCompleted: () => void
  onItemSkipped: () => void
}

export function AnnotationPanel({
  projectId,
  queue,
  currentItem,
  onItemClaimed,
  onItemCompleted,
  onItemSkipped,
}: AnnotationPanelProps) {
  const [seenItemIds, setSeenItemIds] = useState<string[]>([])
  const [scores, setScores] = useState<ScoreSubmission[]>([])

  const claimMutation = useClaimNextItemMutation(projectId, queue.id)
  const completeMutation = useCompleteItemMutation(projectId, queue.id)
  const skipMutation = useSkipItemMutation(projectId, queue.id)
  const releaseMutation = useReleaseItemMutation(projectId, queue.id)

  // Fetch trace data when item is claimed
  const { data: trace, isLoading: isTraceLoading } = useQuery<Trace | null, Error>({
    queryKey: traceQueryKeys.detail(projectId, currentItem?.object_id || ''),
    queryFn: async (): Promise<Trace | null> => {
      if (!currentItem) return null
      if (currentItem.object_type === 'trace') {
        return getTraceById(projectId, currentItem.object_id)
      }
      // For span type, we need to get the span first, then its trace
      const span = await getSpanById(projectId, currentItem.object_id)
      return getTraceById(projectId, span.trace_id)
    },
    enabled: !!currentItem,
    staleTime: 30_000,
  })

  // Fetch spans for the trace
  const { data: spans = [], isLoading: isSpansLoading } = useQuery<Span[], Error>({
    queryKey: traceQueryKeys.spans(projectId, trace?.trace_id || ''),
    queryFn: (): Promise<Span[]> => getSpansForTrace(projectId, trace!.trace_id),
    enabled: !!trace?.trace_id,
    staleTime: 30_000,
  })

  const isTraceDataLoading = isTraceLoading || isSpansLoading

  const handleClaimNext = useCallback(async () => {
    try {
      const item = await claimMutation.mutateAsync({ seen_item_ids: seenItemIds })
      onItemClaimed(item)
      setScores([]) // Reset scores for new item
    } catch {
      // Error is handled by the mutation
    }
  }, [claimMutation, seenItemIds, onItemClaimed])

  const handleComplete = useCallback(async () => {
    if (!currentItem) return
    try {
      await completeMutation.mutateAsync({
        itemId: currentItem.id,
        data: { scores },
      })
      setSeenItemIds((prev) => [...prev, currentItem.id])
      onItemCompleted()
      // Automatically claim next
      handleClaimNext()
    } catch {
      // Error is handled by the mutation
    }
  }, [currentItem, completeMutation, scores, onItemCompleted, handleClaimNext])

  const handleSkip = useCallback(async () => {
    if (!currentItem) return
    try {
      await skipMutation.mutateAsync({
        itemId: currentItem.id,
        data: { reason: 'Skipped by annotator' },
      })
      setSeenItemIds((prev) => [...prev, currentItem.id])
      onItemSkipped()
      // Automatically claim next
      handleClaimNext()
    } catch {
      // Error is handled by the mutation
    }
  }, [currentItem, skipMutation, onItemSkipped, handleClaimNext])

  const handleRelease = useCallback(async () => {
    if (!currentItem) return
    try {
      await releaseMutation.mutateAsync(currentItem.id)
      onItemSkipped()
    } catch {
      // Error is handled by the mutation
    }
  }, [currentItem, releaseMutation, onItemSkipped])

  const isLoading =
    claimMutation.isPending ||
    completeMutation.isPending ||
    skipMutation.isPending ||
    releaseMutation.isPending

  // No current item - show claim button
  if (!currentItem) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Start Annotating</CardTitle>
          <CardDescription>
            Claim an item from the queue to begin reviewing and scoring.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Instructions */}
          {queue.instructions && (
            <Alert>
              <Info className="h-4 w-4" />
              <AlertTitle>Instructions</AlertTitle>
              <AlertDescription className="whitespace-pre-wrap">
                {queue.instructions}
              </AlertDescription>
            </Alert>
          )}

          <Button
            onClick={handleClaimNext}
            disabled={isLoading}
            size="lg"
            className="w-full"
          >
            {isLoading ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Loading...
              </>
            ) : (
              <>
                <ChevronRight className="mr-2 h-4 w-4" />
                Claim Next Item
              </>
            )}
          </Button>

          {claimMutation.error && (
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertTitle>No Items Available</AlertTitle>
              <AlertDescription>
                There are no pending items in this queue. Check back later or add more items.
              </AlertDescription>
            </Alert>
          )}
        </CardContent>
      </Card>
    )
  }

  // Show current item with scoring form
  return (
    <div className="space-y-4">
      {/* Instructions (collapsed) */}
      {queue.instructions && (
        <Alert>
          <Info className="h-4 w-4" />
          <AlertTitle>Instructions</AlertTitle>
          <AlertDescription className="whitespace-pre-wrap line-clamp-2">
            {queue.instructions}
          </AlertDescription>
        </Alert>
      )}

      {/* Current Item Info */}
      <Card>
        <CardHeader className="pb-3">
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="text-lg">Current Item</CardTitle>
              <CardDescription>
                {currentItem.object_type}: {currentItem.object_id}
              </CardDescription>
            </div>
            <Button
              variant="ghost"
              size="sm"
              onClick={handleRelease}
              disabled={isLoading}
            >
              Release Lock
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {/* Trace/Span Viewer */}
          <div className="mb-4">
            {trace ? (
              <TraceViewer
                trace={trace}
                spans={spans}
                objectType={currentItem.object_type as 'trace' | 'span'}
                objectId={currentItem.object_id}
                isLoading={isTraceDataLoading}
              />
            ) : isTraceDataLoading ? (
              <div className="space-y-3">
                <Skeleton className="h-6 w-48" />
                <Skeleton className="h-24 w-full" />
                <Skeleton className="h-24 w-full" />
              </div>
            ) : (
              <div className="text-center py-8 text-muted-foreground">
                <AlertCircle className="h-8 w-8 mx-auto mb-2 opacity-50" />
                <p className="text-sm">Unable to load trace data</p>
                <p className="text-xs mt-1">Object ID: <code className="font-mono">{currentItem.object_id}</code></p>
              </div>
            )}
          </div>

          {/* Score Input Form */}
          <ScoreInputForm
            projectId={projectId}
            queueId={queue.id}
            scoreConfigIds={queue.score_config_ids}
            scores={scores}
            onScoresChange={setScores}
          />
        </CardContent>
      </Card>

      {/* Action Buttons */}
      <div className="flex gap-3">
        <Button
          variant="outline"
          onClick={handleSkip}
          disabled={isLoading}
          className="flex-1"
        >
          {skipMutation.isPending ? (
            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
          ) : (
            <SkipForward className="mr-2 h-4 w-4" />
          )}
          Skip
        </Button>
        <Button
          onClick={handleComplete}
          disabled={isLoading}
          className="flex-1"
        >
          {completeMutation.isPending ? (
            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
          ) : (
            <CheckCircle className="mr-2 h-4 w-4" />
          )}
          Submit & Next
        </Button>
      </div>
    </div>
  )
}

// Loading skeleton for annotation panel
export function AnnotationPanelSkeleton() {
  return (
    <Card>
      <CardHeader>
        <Skeleton className="h-6 w-32" />
        <Skeleton className="h-4 w-64 mt-1" />
      </CardHeader>
      <CardContent className="space-y-4">
        <Skeleton className="h-32 w-full" />
        <Skeleton className="h-10 w-full" />
      </CardContent>
    </Card>
  )
}
