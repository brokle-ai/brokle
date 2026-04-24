import { useCallback, useEffect, useMemo, useState } from 'react'
import { Link, useNavigate } from '@tanstack/react-router'
import {
  useMutation,
  useQuery,
  useQueryClient,
  useSuspenseQuery,
} from '@tanstack/react-query'
import { toast } from 'sonner'
import { ArrowLeft, CheckCircle, Loader2, SkipForward } from 'lucide-react'
import { rawFetch } from '@/lib/api/client'
import { BrokleError } from '@/lib/api/errors'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { IoPreview } from '@/features/traces/components'
import {
  traceDetailQueryOptions,
  traceSpansQueryOptions,
} from '@/features/traces/api/queries'
import { scoreConfigsQueryOptions } from '@/features/scores/api/queries'
import type { ScoreConfig } from '@/features/scores/api/types'
import type { Span } from '@/features/traces/api/types'
import {
  claimNextItem,
  completeItem,
  queueDetailQueryOptions,
  queuesKeys,
  skipItem,
} from '../api/queries'
import type {
  QueueItem,
  ScoreSubmission,
} from '../api/types'
import { ScoreInput, type ScoreInputValue } from './score-input'

interface ReviewItemProps {
  orgId: string
  projectId: string
  queueId: string
}

// Scores map keyed by score-config id. One entry per configured score
// even before the reviewer enters a value — keeps component state
// decoupled from "which inputs exist" so the submit path can straight
// map over it without nil-checks exploding the control flow.
type ScoresState = Record<string, ScoreInputValue>

// When a span-typed item arrives we must resolve its parent trace ID
// before the trace viewer can render. The endpoint mirrors the
// backend's span-get response shape (same `Span` DTO as the list).
async function getSpanByID(spanId: string): Promise<Span> {
  const resp = await rawFetch(
    `/api/v1/spans/${encodeURIComponent(spanId)}`,
    { method: 'GET' },
  )
  return (await resp.json()) as Span
}

export function ReviewItem({ orgId, projectId, queueId }: ReviewItemProps) {
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const { data: detail } = useSuspenseQuery(
    queueDetailQueryOptions(projectId, queueId),
  )
  const { data: configsPage } = useSuspenseQuery(
    scoreConfigsQueryOptions(projectId),
  )

  const { queue } = detail

  // Filter project-level configs to only those attached to this queue.
  // Preserves the order encoded in queue.score_config_ids so reviewers
  // see inputs in the order the queue owner configured them.
  const queueConfigs = useMemo<ScoreConfig[]>(() => {
    const byId = new Map(configsPage.data.map((c) => [c.id, c]))
    return queue.score_config_ids
      .map((id) => byId.get(id))
      .filter((c): c is ScoreConfig => Boolean(c))
  }, [configsPage.data, queue.score_config_ids])

  const [currentItem, setCurrentItem] = useState<QueueItem | null>(null)
  const [seenItemIds, setSeenItemIds] = useState<string[]>([])
  const [scores, setScores] = useState<ScoresState>({})
  const [noMoreItems, setNoMoreItems] = useState(false)

  // Reset per-item form state whenever a new item is claimed.
  const resetScoresFor = useCallback((configs: ScoreConfig[]) => {
    const next: ScoresState = {}
    for (const c of configs) {
      next[c.id] = { value: null, comment: '' }
    }
    setScores(next)
  }, [])

  const claimMutation = useMutation({
    mutationFn: (seen: string[]) =>
      claimNextItem(projectId, queueId, { seen_item_ids: seen }),
    onSuccess: (item) => {
      setCurrentItem(item)
      setNoMoreItems(false)
      resetScoresFor(queueConfigs)
    },
    onError: (err) => {
      // 404 from the backend means "no claimable items left" — surface
      // as an end-of-queue empty state rather than a toast.
      if (err instanceof BrokleError && err.status === 404) {
        setCurrentItem(null)
        setNoMoreItems(true)
        return
      }
      toast.error('Failed to claim next item', {
        description:
          err instanceof Error ? err.message : 'Please try again.',
      })
    },
  })

  const completeMutation = useMutation({
    mutationFn: async (submissions: ScoreSubmission[]) => {
      if (!currentItem) throw new Error('No active item to complete')
      return completeItem(projectId, queueId, currentItem.id, {
        scores: submissions,
      })
    },
    onSuccess: () => {
      // Invalidate the queue's stats cache so the detail view reflects
      // new counts when the reviewer exits.
      queryClient.invalidateQueries({
        queryKey: queuesKeys.detail(queueId),
      })
    },
  })

  const skipMutation = useMutation({
    mutationFn: async () => {
      if (!currentItem) throw new Error('No active item to skip')
      return skipItem(projectId, queueId, currentItem.id, {})
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: queuesKeys.detail(queueId),
      })
    },
  })

  // Automatic first claim on mount — the spec's loader contract asks
  // for "the next claimable item". Claim is a side-effecting POST so
  // it can't live in a route loader; effect-on-mount is the compromise
  // and mirrors the reference `annotation-panel.tsx`.
  useEffect(() => {
    if (currentItem === null && !noMoreItems && !claimMutation.isPending) {
      claimMutation.mutate(seenItemIds)
    }
    // Intentionally scoped — we want this to fire ONLY when idle state
    // becomes eligible, not on every render.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentItem, noMoreItems])

  const buildSubmissions = useCallback((): ScoreSubmission[] => {
    const out: ScoreSubmission[] = []
    for (const c of queueConfigs) {
      const entry = scores[c.id]
      if (!entry || entry.value === null) continue
      out.push({
        score_config_id: c.id,
        value: entry.value,
        comment: entry.comment.trim() || undefined,
      })
    }
    return out
  }, [queueConfigs, scores])

  const handleSubmitAndNext = useCallback(async () => {
    if (!currentItem) return
    const submissions = buildSubmissions()
    try {
      await completeMutation.mutateAsync(submissions)
      const completedId = currentItem.id
      const nextSeen = [...seenItemIds, completedId]
      setSeenItemIds(nextSeen)
      setCurrentItem(null)
      claimMutation.mutate(nextSeen)
    } catch (err) {
      toast.error('Failed to submit scores', {
        description:
          err instanceof Error ? err.message : 'Please try again.',
      })
    }
  }, [
    buildSubmissions,
    claimMutation,
    completeMutation,
    currentItem,
    seenItemIds,
  ])

  const handleSkip = useCallback(async () => {
    if (!currentItem) return
    try {
      await skipMutation.mutateAsync()
      const skippedId = currentItem.id
      const nextSeen = [...seenItemIds, skippedId]
      setSeenItemIds(nextSeen)
      setCurrentItem(null)
      claimMutation.mutate(nextSeen)
    } catch (err) {
      toast.error('Failed to skip item', {
        description:
          err instanceof Error ? err.message : 'Please try again.',
      })
    }
  }, [claimMutation, currentItem, seenItemIds, skipMutation])

  const handleExit = useCallback(() => {
    navigate({
      to: '/o/$orgId/p/$projectId/annotation-queues/$queueId',
      params: { orgId, projectId, queueId },
      search: { page: 1, limit: 20, itemStatus: undefined },
    })
  }, [navigate, orgId, projectId, queueId])

  const busy =
    claimMutation.isPending ||
    completeMutation.isPending ||
    skipMutation.isPending

  const canSubmit = useMemo(() => {
    // Require at least one score filled so a completion carries real
    // signal; reviewers who want to pass without scoring should skip.
    return buildSubmissions().length > 0 && !busy
  }, [buildSubmissions, busy])

  return (
    <main className="mx-auto max-w-6xl space-y-4 p-6">
      <nav className="flex items-center justify-between">
        <Button
          variant="ghost"
          size="sm"
          onClick={handleExit}
          className="gap-1 px-2"
        >
          <ArrowLeft className="h-4 w-4" />
          Exit review
        </Button>
        <div className="text-sm text-muted-foreground">
          Reviewing{' '}
          <Link
            to="/o/$orgId/p/$projectId/annotation-queues/$queueId"
            params={{ orgId, projectId, queueId }}
            search={{ page: 1, limit: 20, itemStatus: undefined }}
            className="font-medium text-foreground hover:underline"
          >
            {queue.name}
          </Link>
        </div>
      </nav>

      {noMoreItems ? (
        <Card>
          <CardHeader>
            <CardTitle>All caught up</CardTitle>
            <CardDescription>
              There are no more pending items in this queue.
            </CardDescription>
          </CardHeader>
          <CardContent className="flex gap-2">
            <Button onClick={handleExit}>Back to queue</Button>
            <Button
              variant="outline"
              onClick={() => {
                setNoMoreItems(false)
                claimMutation.mutate(seenItemIds)
              }}
            >
              Try again
            </Button>
          </CardContent>
        </Card>
      ) : !currentItem ? (
        <Card>
          <CardHeader>
            <CardTitle>Loading next item…</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              <Skeleton className="h-6 w-48" />
              <Skeleton className="h-24 w-full" />
              <Skeleton className="h-24 w-full" />
            </div>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_380px]">
          <Card>
            <CardHeader>
              <div className="flex items-start justify-between gap-2">
                <div>
                  <CardTitle>
                    {currentItem.object_type === 'trace' ? 'Trace' : 'Span'}
                  </CardTitle>
                  <CardDescription className="font-mono text-xs">
                    {currentItem.object_id}
                  </CardDescription>
                </div>
                <Badge variant="outline" className="text-xs">
                  {currentItem.object_type}
                </Badge>
              </div>
            </CardHeader>
            <CardContent>
              <ReviewObjectView
                objectType={currentItem.object_type}
                objectId={currentItem.object_id}
              />
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="text-base">Scores</CardTitle>
              {queue.instructions ? (
                <Alert className="mt-2">
                  <AlertTitle className="text-xs font-semibold">
                    Instructions
                  </AlertTitle>
                  <AlertDescription className="whitespace-pre-wrap text-xs">
                    {queue.instructions}
                  </AlertDescription>
                </Alert>
              ) : null}
            </CardHeader>
            <CardContent className="space-y-3">
              {queueConfigs.length === 0 ? (
                <Alert variant="destructive">
                  <AlertTitle>No score configs</AlertTitle>
                  <AlertDescription>
                    This queue has no attached score configs. Configure at
                    least one before reviewing.
                  </AlertDescription>
                </Alert>
              ) : (
                queueConfigs.map((c) => (
                  <ScoreInput
                    key={c.id}
                    config={c}
                    value={
                      scores[c.id] ?? { value: null, comment: '' }
                    }
                    onChange={(next) =>
                      setScores((prev) => ({ ...prev, [c.id]: next }))
                    }
                    disabled={busy}
                  />
                ))
              )}

              <div className="flex flex-col gap-2 pt-2">
                <Button
                  onClick={handleSubmitAndNext}
                  disabled={!canSubmit}
                >
                  {completeMutation.isPending ? (
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  ) : (
                    <CheckCircle className="mr-2 h-4 w-4" />
                  )}
                  Submit & claim next
                </Button>
                <Button
                  variant="outline"
                  onClick={handleSkip}
                  disabled={busy}
                >
                  {skipMutation.isPending ? (
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  ) : (
                    <SkipForward className="mr-2 h-4 w-4" />
                  )}
                  Skip
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      )}
    </main>
  )
}

interface ReviewObjectViewProps {
  objectType: 'trace' | 'span'
  objectId: string
}

// Minimal inline viewer — renders the root span's input/output via the
// shared IoPreview. For span-typed items we first look up the span to
// discover its trace_id, then fetch the trace header + spans; for
// trace-typed items we skip the span lookup. Keeping this local to
// the review route rather than reusing TraceDetail keeps the surface
// focused on "read the prompt, score it" without importing the full
// trace timeline UI.
function ReviewObjectView({ objectType, objectId }: ReviewObjectViewProps) {
  const spanLookup = useQuery({
    queryKey: ['review-span-lookup', objectId],
    queryFn: () => getSpanByID(objectId),
    enabled: objectType === 'span',
    staleTime: 30 * 1000,
  })

  const traceId =
    objectType === 'trace' ? objectId : spanLookup.data?.trace_id
  const trace = useQuery({
    ...traceDetailQueryOptions(traceId ?? ''),
    enabled: Boolean(traceId),
  })
  const spans = useQuery({
    ...traceSpansQueryOptions(traceId ?? ''),
    enabled: Boolean(traceId),
  })

  const targetSpan = useMemo(() => {
    if (!spans.data) return undefined
    if (objectType === 'span') {
      return spans.data.find((s) => s.span_id === objectId)
    }
    return (
      spans.data.find((s) => !s.parent_span_id) ?? spans.data[0]
    )
  }, [objectId, objectType, spans.data])

  if (
    spanLookup.isLoading ||
    trace.isLoading ||
    spans.isLoading ||
    !traceId
  ) {
    return (
      <div className="space-y-3">
        <Skeleton className="h-6 w-48" />
        <Skeleton className="h-24 w-full" />
        <Skeleton className="h-24 w-full" />
      </div>
    )
  }

  if (trace.error || spans.error || spanLookup.error) {
    return (
      <Alert variant="destructive">
        <AlertTitle>Unable to load trace data</AlertTitle>
        <AlertDescription>
          {(trace.error ?? spans.error ?? spanLookup.error)?.message ??
            'Unknown error'}
        </AlertDescription>
      </Alert>
    )
  }

  if (!trace.data || !targetSpan) {
    return (
      <p className="text-sm text-muted-foreground">
        No renderable content for this object.
      </p>
    )
  }

  return (
    <div className="space-y-3">
      <div className="flex flex-wrap gap-2">
        {trace.data.name ? (
          <Badge variant="secondary">{trace.data.name}</Badge>
        ) : null}
        {targetSpan.model_name ? (
          <Badge variant="outline">{targetSpan.model_name}</Badge>
        ) : null}
        {targetSpan.provider_name ? (
          <Badge variant="outline">{targetSpan.provider_name}</Badge>
        ) : null}
      </div>
      <IoPreview value={targetSpan.input} label="Input" />
      <IoPreview value={targetSpan.output} label="Output" />
    </div>
  )
}
