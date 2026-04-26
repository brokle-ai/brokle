import { useEffect, useMemo, useRef, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { AlertCircle } from 'lucide-react'
import { toast } from 'sonner'
import { Sheet, SheetContent, SheetTitle } from '@/components/ui/sheet'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'
import {
  traceDetailQueryOptions,
  traceSpansQueryOptions,
  traceAnnotationsQueryOptions,
  traceCommentsQueryOptions,
  createTraceComment,
  createTraceCommentReply,
  deleteTraceAnnotation,
  toggleTraceCommentReaction,
  tracesKeys,
} from '../api/queries'
import type { Span, TraceAnnotation, TraceDetail } from '../api/types'
import { SpanTree } from './span-tree'
import { SpanDetailPanel } from './span-detail-panel'
import { IoPreview } from './io-preview'
import { AnnotationsDrawer } from './annotations-drawer'
import { CommentList } from './comment-list'
import { PeekSheetHeader } from './peek-sheet/peek-sheet-header'
import { usePeekNavigation } from '../hooks/use-peek-navigation'
import { useDetailNavigation } from '../hooks/use-detail-navigation'
import { usePeekSheetState } from '../hooks/use-peek-sheet-state'
import type { TraceTab } from '../types/peek'

interface TracesPeekViewProps {
  orgId: string
  projectId: string
}

// TracesPeekView — mounts as a right-side sheet when `?peek=<traceId>`
// is present in the URL. Composes the existing SpanTree +
// SpanDetailPanel + IoPreview + annotations/comments drawers into a
// tabbed side-sheet. UX contract:
//   - slide-in sheet, modal={false} so the list behind stays visible
//     (pointer interactions with the list still propagate closure)
//   - keyboard: ← / → prev-next, Esc close
//   - tabs: IO / Span Tree / Metadata / Scores / Comments, state in URL
//   - header: trace name + status + duration/cost + prev-next +
//     add-to-queue + maximize (full page) + close
export function TracesPeekView({ orgId, projectId }: TracesPeekViewProps) {
  const { peekId, selectedTab } = usePeekSheetState()
  const { closePeek, expandPeek, setTab } = usePeekNavigation()
  const { handlePrev, handleNext, canGoPrev, canGoNext } =
    useDetailNavigation()

  // ------------------------------------------------------------------
  // Trace + spans fetch
  //
  // Both queries are guarded on `peekId` — when the sheet is closed we
  // don't prefetch. Stale windows match the list query's 30s.
  // ------------------------------------------------------------------
  const traceQuery = useQuery({
    ...traceDetailQueryOptions(projectId, peekId ?? ''),
    enabled: !!peekId,
  })
  const spansQuery = useQuery({
    ...traceSpansQueryOptions(projectId, peekId ?? ''),
    enabled: !!peekId,
  })

  // ------------------------------------------------------------------
  // Keyboard shortcuts — ref pattern to avoid re-binding the listener
  // on every handler identity change.
  // ------------------------------------------------------------------
  const handlersRef = useRef({ handlePrev, handleNext, closePeek })
  useEffect(() => {
    handlersRef.current = { handlePrev, handleNext, closePeek }
  }, [handlePrev, handleNext, closePeek])

  useEffect(() => {
    if (!peekId) return
    const onKey = (e: KeyboardEvent) => {
      const target = e.target as HTMLElement | null
      // Swallow nav shortcuts while the user is typing in the comments
      // textarea or any other form field inside the sheet.
      if (
        target &&
        (target.tagName === 'INPUT' ||
          target.tagName === 'TEXTAREA' ||
          target.isContentEditable)
      ) {
        if (e.key === 'Escape') {
          handlersRef.current.closePeek()
          e.preventDefault()
        }
        return
      }
      if (e.key === 'ArrowLeft' && !e.metaKey && !e.ctrlKey) {
        e.preventDefault()
        handlersRef.current.handlePrev()
      } else if (e.key === 'ArrowRight' && !e.metaKey && !e.ctrlKey) {
        e.preventDefault()
        handlersRef.current.handleNext()
      } else if (e.key === 'Escape') {
        e.preventDefault()
        handlersRef.current.closePeek()
      }
    }
    document.addEventListener('keydown', onKey)
    return () => document.removeEventListener('keydown', onKey)
  }, [peekId])

  if (!peekId) return null

  const trace = traceQuery.data
  const spans = spansQuery.data ?? []
  const isLoading = traceQuery.isLoading || spansQuery.isLoading
  const error = traceQuery.error || spansQuery.error

  return (
    <Sheet
      open={!!peekId}
      onOpenChange={(open) => {
        if (!open) closePeek()
      }}
      modal={false}
    >
      <SheetContent
        side="right"
        className="flex max-h-full min-h-0 w-full min-w-[70vw] flex-col gap-0 overflow-hidden rounded-l-xl p-0 sm:max-w-none"
        // Without `onPointerDownOutside` the sheet auto-closes on the
        // first click in the list behind. modal={false} lets clicks
        // propagate; `preventDefault` keeps the sheet open through
        // them — matching web/.
        onPointerDownOutside={(e) => e.preventDefault()}
        onInteractOutside={(e) => e.preventDefault()}
        tabIndex={-1}
        hideCloseButton
      >
        <SheetTitle className="sr-only">Trace details</SheetTitle>

        {isLoading ? (
          <PeekSheetSkeleton />
        ) : error || !trace ? (
          <PeekSheetError
            message={
              error instanceof Error
                ? error.message
                : 'Trace not found'
            }
          />
        ) : (
          <PeekSheetBody
            trace={trace}
            spans={spans}
            orgId={orgId}
            projectId={projectId}
            selectedTab={selectedTab}
            onTabChange={setTab}
            onPrev={handlePrev}
            onNext={handleNext}
            onClose={closePeek}
            onExpand={() => expandPeek(false)}
            onExpandNewTab={() => expandPeek(true)}
            canGoPrev={canGoPrev}
            canGoNext={canGoNext}
          />
        )}
      </SheetContent>
    </Sheet>
  )
}

// ----------------------------------------------------------------------
// Body — header + tabs
// ----------------------------------------------------------------------

interface PeekSheetBodyProps {
  trace: TraceDetail
  spans: Span[]
  orgId: string
  projectId: string
  selectedTab: TraceTab
  onTabChange: (tab: TraceTab) => void
  onPrev: () => void
  onNext: () => void
  onClose: () => void
  onExpand: () => void
  onExpandNewTab: () => void
  canGoPrev: boolean
  canGoNext: boolean
}

function PeekSheetBody({
  trace,
  spans,
  orgId,
  projectId,
  selectedTab,
  onTabChange,
  onPrev,
  onNext,
  onClose,
  onExpand,
  onExpandNewTab,
  canGoPrev,
  canGoNext,
}: PeekSheetBodyProps) {
  // Track the span selected in the Span Tree tab. Default to the root
  // span (no parent) or the first span if the tree is flat.
  const [selectedSpanId, setSelectedSpanId] = useState<string | undefined>()
  useEffect(() => {
    setSelectedSpanId(
      spans.find((s) => !s.parent_span_id)?.span_id ?? spans[0]?.span_id,
    )
  }, [spans])

  const selectedSpan = useMemo(
    () => spans.find((s) => s.span_id === selectedSpanId),
    [spans, selectedSpanId],
  )

  // IO preview sources — root span's input/output if available, else
  // fall through to the first non-empty span. The peek view is a
  // preview, not the full log view; surfacing a meaningful sample
  // matters more than showing everything.
  const ioSpan = useMemo(() => {
    const root = spans.find((s) => !s.parent_span_id)
    if (root && (root.input || root.output)) return root
    return spans.find((s) => s.input || s.output) ?? root ?? spans[0]
  }, [spans])

  return (
    <div className="flex h-full flex-col">
      <PeekSheetHeader
        trace={trace}
        orgId={orgId}
        projectId={projectId}
        onPrevious={onPrev}
        onNext={onNext}
        onExpand={onExpand}
        onExpandNewTab={onExpandNewTab}
        onClose={onClose}
        hasPrevious={canGoPrev}
        hasNext={canGoNext}
      />

      <Tabs
        value={selectedTab}
        onValueChange={(v) => onTabChange(v as TraceTab)}
        className="flex min-h-0 flex-1 flex-col"
      >
        <div className="border-b px-4">
          <TabsList className="h-9">
            <TabsTrigger value="io" className="text-xs">
              IO Preview
            </TabsTrigger>
            <TabsTrigger value="tree" className="text-xs">
              Span Tree
              <Badge variant="outline" className="ml-1.5 h-4 px-1 text-[10px]">
                {spans.length}
              </Badge>
            </TabsTrigger>
            <TabsTrigger value="metadata" className="text-xs">
              Metadata
            </TabsTrigger>
            <TabsTrigger value="scores" className="text-xs">
              Scores
            </TabsTrigger>
            <TabsTrigger value="comments" className="text-xs">
              Comments
            </TabsTrigger>
          </TabsList>
        </div>

        <TabsContent
          value="io"
          className="mt-0 min-h-0 flex-1 overflow-y-auto p-4"
        >
          <IoTabContent span={ioSpan} />
        </TabsContent>

        <TabsContent
          value="tree"
          className="mt-0 min-h-0 flex-1 overflow-y-auto p-4"
        >
          {spans.length === 0 ? (
            <EmptyState message="No spans in this trace" />
          ) : (
            <div className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(320px,420px)]">
              <SpanTree
                spans={spans}
                selectedSpanId={selectedSpanId}
                onSpanSelect={(s) => setSelectedSpanId(s.span_id)}
              />
              {selectedSpan ? (
                <div className="lg:sticky lg:top-0">
                  <SpanDetailPanel span={selectedSpan} />
                </div>
              ) : null}
            </div>
          )}
        </TabsContent>

        <TabsContent
          value="metadata"
          className="mt-0 min-h-0 flex-1 overflow-y-auto p-4"
        >
          <MetadataTabContent trace={trace} />
        </TabsContent>

        <TabsContent
          value="scores"
          className="mt-0 min-h-0 flex-1 overflow-y-auto p-4"
        >
          <ScoresTabContent projectId={projectId} traceId={trace.trace_id} />
        </TabsContent>

        <TabsContent
          value="comments"
          className="mt-0 min-h-0 flex-1 overflow-y-auto p-4"
        >
          <CommentsTabContent
            projectId={projectId}
            traceId={trace.trace_id}
          />
        </TabsContent>
      </Tabs>
    </div>
  )
}

// ----------------------------------------------------------------------
// Tab contents
// ----------------------------------------------------------------------

function IoTabContent({ span }: { span: Span | undefined }) {
  if (!span) {
    return <EmptyState message="No IO data for this trace" />
  }
  return (
    <div className="space-y-4">
      <section>
        <h3 className="mb-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
          Input
        </h3>
        <IoPreview value={span.input} label="Input" />
      </section>
      <section>
        <h3 className="mb-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
          Output
        </h3>
        <IoPreview value={span.output} label="Output" />
      </section>
    </div>
  )
}

function MetadataTabContent({ trace }: { trace: TraceDetail }) {
  const rows: Array<[string, React.ReactNode]> = [
    ['Trace ID', <span key="id" className="font-mono text-xs">{trace.trace_id}</span>],
    ['Root span ID', <span key="rid" className="font-mono text-xs">{trace.root_span_id}</span>],
    ['Name', trace.name],
    ['Service', trace.service_name ?? '—'],
    ['Model', trace.model_name ?? '—'],
    ['Provider', trace.provider_name ?? '—'],
    ['User ID', trace.user_id ?? '—'],
    ['Session ID', trace.session_id ?? '—'],
    ['Start time', new Date(trace.start_time).toLocaleString()],
    [
      'End time',
      trace.end_time ? new Date(trace.end_time).toLocaleString() : '—',
    ],
    ['Spans', trace.span_count.toLocaleString()],
    ['Error spans', trace.error_span_count.toLocaleString()],
    ['Input tokens', trace.input_tokens.toLocaleString()],
    ['Output tokens', trace.output_tokens.toLocaleString()],
    ['Total tokens', trace.total_tokens.toLocaleString()],
    [
      'Tags',
      trace.tags && trace.tags.length > 0 ? (
        <div key="tags" className="flex flex-wrap gap-1">
          {trace.tags.map((t) => (
            <Badge key={t} variant="outline">
              {t}
            </Badge>
          ))}
        </div>
      ) : (
        '—'
      ),
    ],
  ]
  return (
    <dl className="grid grid-cols-[minmax(140px,200px)_minmax(0,1fr)] gap-x-4 gap-y-2 text-sm">
      {rows.map(([k, v]) => (
        <div key={k} className="contents">
          <dt className="text-muted-foreground">{k}</dt>
          <dd className="min-w-0 break-words">{v}</dd>
        </div>
      ))}
    </dl>
  )
}

function ScoresTabContent({
  projectId,
  traceId,
}: {
  projectId: string
  traceId: string
}) {
  const queryClient = useQueryClient()
  const { data: annotations = [], isLoading } = useQuery(
    traceAnnotationsQueryOptions(projectId, traceId),
  )

  const deleteMutation = useMutation({
    mutationFn: (scoreId: string) =>
      deleteTraceAnnotation(projectId, traceId, scoreId),
    onSuccess: () =>
      queryClient.invalidateQueries({
        queryKey: tracesKeys.annotations(traceId, projectId),
      }),
    onError: (err) =>
      toast.error('Failed to delete annotation', {
        description: err instanceof Error ? err.message : 'Please try again.',
      }),
  })

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <p className="text-xs text-muted-foreground">
          {annotations.length.toLocaleString()} annotation
          {annotations.length === 1 ? '' : 's'}
        </p>
        <AnnotationsDrawer projectId={projectId} traceId={traceId} />
      </div>

      {isLoading ? (
        <Skeleton className="h-24 w-full" />
      ) : annotations.length === 0 ? (
        <EmptyState message="No annotations yet" />
      ) : (
        <ul className="space-y-2">
          {annotations.map((a) => (
            <AnnotationRow
              key={a.id}
              annotation={a}
              onDelete={() => deleteMutation.mutate(a.id)}
              isDeleting={
                deleteMutation.isPending && deleteMutation.variables === a.id
              }
            />
          ))}
        </ul>
      )}
    </div>
  )
}

function AnnotationRow({
  annotation,
  onDelete,
  isDeleting,
}: {
  annotation: TraceAnnotation
  onDelete: () => void
  isDeleting: boolean
}) {
  const displayValue =
    annotation.type === 'CATEGORICAL'
      ? (annotation.string_value ?? '—')
      : annotation.type === 'BOOLEAN'
        ? annotation.value === 1
          ? 'true'
          : 'false'
        : (annotation.value ?? '—').toString()
  return (
    <li className="flex items-start justify-between gap-3 rounded-md border p-3">
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium">{annotation.name}</span>
          <Badge variant="outline" className="text-[10px]">
            {annotation.type}
          </Badge>
          <Badge variant="secondary" className="text-[10px]">
            {annotation.source}
          </Badge>
        </div>
        <p className="mt-1 text-sm font-mono">{displayValue}</p>
        {annotation.reason ? (
          <p className="mt-1 text-xs text-muted-foreground">
            {annotation.reason}
          </p>
        ) : null}
      </div>
      <button
        type="button"
        onClick={onDelete}
        disabled={isDeleting}
        className="text-xs text-muted-foreground hover:text-destructive disabled:opacity-50"
      >
        {isDeleting ? 'Deleting…' : 'Delete'}
      </button>
    </li>
  )
}

function CommentsTabContent({
  projectId,
  traceId,
}: {
  projectId: string
  traceId: string
}) {
  const queryClient = useQueryClient()
  const { data, isLoading } = useQuery(
    traceCommentsQueryOptions(projectId, traceId),
  )
  const comments = data?.comments ?? []

  const invalidate = () =>
    queryClient.invalidateQueries({
      queryKey: tracesKeys.comments(traceId, projectId),
    })

  const replyMutation = useMutation({
    mutationFn: ({ parentId, content }: { parentId: string; content: string }) =>
      createTraceCommentReply(projectId, traceId, parentId, { content }),
    onSuccess: invalidate,
    onError: (err) =>
      toast.error('Failed to post reply', {
        description: err instanceof Error ? err.message : 'Please try again.',
      }),
  })

  const reactionMutation = useMutation({
    mutationFn: ({ commentId, emoji }: { commentId: string; emoji: string }) =>
      toggleTraceCommentReaction(projectId, traceId, commentId, { emoji }),
    onSuccess: invalidate,
    onError: (err) =>
      toast.error('Failed to toggle reaction', {
        description: err instanceof Error ? err.message : 'Please try again.',
      }),
  })

  const [draft, setDraft] = useState('')
  const createMutation = useMutation({
    mutationFn: (content: string) =>
      createTraceComment(projectId, traceId, { content }),
    onSuccess: () => {
      setDraft('')
      invalidate()
    },
    onError: (err) =>
      toast.error('Failed to post comment', {
        description: err instanceof Error ? err.message : 'Please try again.',
      }),
  })

  const trimmed = draft.trim()
  const canSubmit =
    trimmed.length > 0 && trimmed.length <= 10_000 && !createMutation.isPending

  return (
    <div className="space-y-4">
      <div>
        <textarea
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          placeholder="Add a comment…"
          rows={3}
          maxLength={10_000}
          className="w-full resize-none rounded-md border bg-background p-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
        />
        <div className="mt-2 flex items-center justify-between">
          <span className="text-xs text-muted-foreground">
            {draft.length.toLocaleString()} / 10,000
          </span>
          <button
            type="button"
            disabled={!canSubmit}
            onClick={() => createMutation.mutate(trimmed)}
            className="rounded-md border bg-primary px-3 py-1 text-xs font-medium text-primary-foreground disabled:opacity-50"
          >
            {createMutation.isPending ? 'Posting…' : 'Post'}
          </button>
        </div>
      </div>

      {isLoading ? (
        <Skeleton className="h-24 w-full" />
      ) : comments.length === 0 ? (
        <EmptyState message="No comments yet" />
      ) : (
        <CommentList
          comments={comments}
          onReply={(parentId, content) =>
            replyMutation.mutate({ parentId, content })
          }
          onToggleReaction={(commentId, emoji) =>
            reactionMutation.mutate({ commentId, emoji })
          }
          replyPendingId={
            replyMutation.isPending
              ? replyMutation.variables?.parentId
              : undefined
          }
          reactionPendingId={
            reactionMutation.isPending
              ? reactionMutation.variables?.commentId
              : undefined
          }
        />
      )}
    </div>
  )
}

// ----------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------

function PeekSheetSkeleton() {
  return (
    <div className="flex h-full flex-col">
      <div className="space-y-3 border-b p-4">
        <Skeleton className="h-6 w-48" />
        <div className="flex gap-2">
          <Skeleton className="h-5 w-16" />
          <Skeleton className="h-5 w-20" />
          <Skeleton className="h-5 w-24" />
        </div>
      </div>
      <div className="flex-1 space-y-3 p-4">
        <Skeleton className="h-8 w-full" />
        <Skeleton className="h-32 w-full" />
        <Skeleton className="h-32 w-full" />
      </div>
    </div>
  )
}

function PeekSheetError({ message }: { message: string }) {
  return (
    <div className="flex h-full flex-col items-center justify-center py-12 text-destructive">
      <AlertCircle className="mb-4 h-12 w-12" />
      <h3 className="mb-2 text-lg font-semibold">Failed to load trace</h3>
      <p className="text-sm text-muted-foreground">{message}</p>
    </div>
  )
}

function EmptyState({ message }: { message: string }) {
  return (
    <div className="flex items-center justify-center py-12">
      <p className="text-sm text-muted-foreground">{message}</p>
    </div>
  )
}
