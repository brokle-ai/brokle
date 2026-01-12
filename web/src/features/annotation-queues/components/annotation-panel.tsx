'use client'

import { useState, useCallback } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Skeleton } from '@/components/ui/skeleton'
import {
  ChevronRight,
  SkipForward,
  CheckCircle,
  AlertCircle,
  Info,
  Loader2,
} from 'lucide-react'
import { ScoreInputForm } from './score-input-form'
import {
  useClaimNextItemMutation,
  useCompleteItemMutation,
  useSkipItemMutation,
  useReleaseItemMutation,
} from '../hooks/use-annotation-queues'
import type { QueueItem, AnnotationQueue, ScoreSubmission } from '../types'

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
          {/* Placeholder for trace/span data viewer */}
          <div className="rounded-lg border bg-muted/50 p-4 mb-4">
            <p className="text-sm text-muted-foreground text-center">
              Trace/span viewer will be integrated here.
              <br />
              Object ID: <code className="font-mono">{currentItem.object_id}</code>
            </p>
          </div>

          {/* Score Input Form */}
          <ScoreInputForm
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
