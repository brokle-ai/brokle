import { useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { formatDistanceToNow } from 'date-fns'
import {
  ClipboardCheck,
  Hash,
  List as ListIcon,
  Loader2,
  Plus,
  ToggleLeft,
  Trash2,
} from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet'
import { Textarea } from '@/components/ui/textarea'
import { cn } from '@/lib/utils'
import {
  scoreConfigsQueryOptions,
} from '@/features/scores/api/queries'
import type { ScoreConfig } from '@/features/scores/api/types'
import { ScoreInput, type ScoreInputValue } from '@/features/annotation-queues/components/score-input'
import {
  createTraceAnnotation,
  deleteTraceAnnotation,
  traceAnnotationsQueryOptions,
  tracesKeys,
} from '../api/queries'
import type {
  CreateTraceAnnotationRequest,
  TraceAnnotation,
} from '../api/types'

interface AnnotationsDrawerProps {
  projectId: string
  traceId: string
  className?: string
}

export function AnnotationsDrawer({
  projectId,
  traceId,
  className,
}: AnnotationsDrawerProps) {
  const queryClient = useQueryClient()
  const [open, setOpen] = useState(false)
  const [configId, setConfigId] = useState<string>('')
  const [entry, setEntry] = useState<ScoreInputValue>({
    value: null,
    comment: '',
  })

  const annotationsQuery = useQuery(
    traceAnnotationsQueryOptions(projectId, traceId),
  )
  const configsQuery = useQuery(scoreConfigsQueryOptions(projectId))

  const configs = configsQuery.data?.data ?? []
  const selectedConfig: ScoreConfig | undefined = useMemo(
    () => configs.find((c) => c.id === configId),
    [configId, configs],
  )

  const invalidate = () =>
    queryClient.invalidateQueries({
      queryKey: tracesKeys.annotations(traceId, projectId),
    })

  const createMutation = useMutation({
    mutationFn: (data: CreateTraceAnnotationRequest) =>
      createTraceAnnotation(projectId, traceId, data),
    onSuccess: () => {
      setConfigId('')
      setEntry({ value: null, comment: '' })
      invalidate()
    },
    onError: (err) => {
      toast.error('Failed to add annotation', {
        description:
          err instanceof Error ? err.message : 'Please try again.',
      })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (scoreId: string) =>
      deleteTraceAnnotation(projectId, traceId, scoreId),
    onSuccess: invalidate,
    onError: (err) => {
      toast.error('Failed to delete annotation', {
        description:
          err instanceof Error ? err.message : 'Please try again.',
      })
    },
  })

  const handleSubmit = () => {
    if (!selectedConfig || entry.value === null) return
    const base: CreateTraceAnnotationRequest = {
      name: selectedConfig.name,
      type: selectedConfig.type,
      reason: entry.comment.trim() || undefined,
    }
    if (selectedConfig.type === 'NUMERIC') {
      if (typeof entry.value !== 'number') return
      base.value = entry.value
    } else if (selectedConfig.type === 'BOOLEAN') {
      if (typeof entry.value !== 'boolean') return
      // Backend encodes BOOLEAN as 1.0/0.0 in the numeric `value`
      // field — see `observability.Score` domain entity. The wire
      // contract doesn't accept a native boolean here.
      base.value = entry.value ? 1 : 0
    } else if (selectedConfig.type === 'CATEGORICAL') {
      if (typeof entry.value !== 'string') return
      base.string_value = entry.value
    }
    createMutation.mutate(base)
  }

  const annotations = annotationsQuery.data ?? []
  const human = annotations.filter((a) => a.source === 'annotation')
  const automated = annotations.filter((a) => a.source !== 'annotation')
  const total = annotations.length
  const badgeText = total > 99 ? '99+' : String(total)

  const canSubmit = Boolean(selectedConfig) && entry.value !== null

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetTrigger asChild>
        <Button
          variant="ghost"
          size="sm"
          className={cn('relative h-8 gap-1 px-2', className)}
        >
          <ClipboardCheck className="h-4 w-4" />
          <span className="text-xs">Annotations</span>
          {total > 0 ? (
            <Badge
              variant="secondary"
              className="ml-1 h-4 min-w-4 px-1 text-[10px]"
            >
              {badgeText}
            </Badge>
          ) : null}
        </Button>
      </SheetTrigger>
      <SheetContent
        side="right"
        className="flex w-full flex-col gap-0 p-0 sm:max-w-md lg:max-w-lg"
      >
        <SheetHeader className="border-b px-4 py-3">
          <SheetTitle className="flex items-center gap-2 text-base">
            <ClipboardCheck className="h-4 w-4" />
            Annotations & Scores
            {total > 0 ? (
              <Badge variant="secondary">{total}</Badge>
            ) : null}
          </SheetTitle>
        </SheetHeader>

        <div className="border-b px-4 py-4">
          <div className="space-y-3">
            <div className="space-y-2">
              <Label htmlFor="annotation-score-config">Score</Label>
              <Select
                value={configId}
                onValueChange={(v) => {
                  setConfigId(v)
                  setEntry({ value: null, comment: '' })
                }}
                disabled={createMutation.isPending}
              >
                <SelectTrigger id="annotation-score-config">
                  <SelectValue
                    placeholder={
                      configsQuery.isLoading
                        ? 'Loading score configs...'
                        : 'Select a score type...'
                    }
                  />
                </SelectTrigger>
                <SelectContent>
                  {configs.map((c) => (
                    <SelectItem key={c.id} value={c.id}>
                      <span className="flex items-center gap-2">
                        <span>{c.name}</span>
                        <span className="text-xs text-muted-foreground">
                          ({c.type.toLowerCase()})
                        </span>
                      </span>
                    </SelectItem>
                  ))}
                  {configs.length === 0 && !configsQuery.isLoading ? (
                    <div className="px-2 py-1.5 text-xs text-muted-foreground">
                      No score configs yet — create one in Scores
                      settings.
                    </div>
                  ) : null}
                </SelectContent>
              </Select>
            </div>

            {selectedConfig ? (
              <>
                <ScoreInput
                  config={selectedConfig}
                  value={entry}
                  onChange={setEntry}
                  disabled={createMutation.isPending}
                />
                <Button
                  type="button"
                  className="w-full"
                  onClick={handleSubmit}
                  disabled={!canSubmit || createMutation.isPending}
                >
                  {createMutation.isPending ? (
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  ) : (
                    <Plus className="mr-2 h-4 w-4" />
                  )}
                  Add annotation
                </Button>
              </>
            ) : null}
          </div>
        </div>

        <div className="flex-1 overflow-y-auto px-4 py-4">
          {annotationsQuery.isLoading ? (
            <div className="flex items-center justify-center py-8 text-muted-foreground">
              <Loader2 className="h-5 w-5 animate-spin" />
            </div>
          ) : total === 0 ? (
            <div className="py-10 text-center">
              <p className="text-sm text-muted-foreground">
                No annotations yet.
              </p>
              <p className="mt-1 text-xs text-muted-foreground">
                Add one above to get started.
              </p>
            </div>
          ) : (
            <div className="space-y-5">
              {human.length > 0 ? (
                <AnnotationSection
                  title="Human annotations"
                  variant="primary"
                  annotations={human}
                  onDelete={(id) => deleteMutation.mutate(id)}
                  deletingId={
                    deleteMutation.isPending
                      ? deleteMutation.variables
                      : undefined
                  }
                />
              ) : null}
              {automated.length > 0 ? (
                <AnnotationSection
                  title="Automated scores"
                  variant="muted"
                  annotations={automated}
                />
              ) : null}
            </div>
          )}
        </div>
      </SheetContent>
    </Sheet>
  )
}

interface AnnotationSectionProps {
  title: string
  variant: 'primary' | 'muted'
  annotations: TraceAnnotation[]
  onDelete?: (id: string) => void
  deletingId?: string
}

function AnnotationSection({
  title,
  variant,
  annotations,
  onDelete,
  deletingId,
}: AnnotationSectionProps) {
  return (
    <section className="space-y-2">
      <div className="flex items-center gap-2">
        <h4
          className={cn(
            'text-sm font-medium',
            variant === 'muted' && 'text-muted-foreground',
          )}
        >
          {title}
        </h4>
        <Badge
          variant={variant === 'primary' ? 'outline' : 'secondary'}
          className="text-xs"
        >
          {annotations.length}
        </Badge>
      </div>
      <div className="space-y-2">
        {annotations.map((a) => (
          <AnnotationRow
            key={a.id}
            annotation={a}
            onDelete={onDelete}
            isDeleting={deletingId === a.id}
          />
        ))}
      </div>
    </section>
  )
}

interface AnnotationRowProps {
  annotation: TraceAnnotation
  onDelete?: (id: string) => void
  isDeleting?: boolean
}

function AnnotationRow({
  annotation,
  onDelete,
  isDeleting,
}: AnnotationRowProps) {
  const Icon =
    annotation.type === 'BOOLEAN'
      ? ToggleLeft
      : annotation.type === 'CATEGORICAL'
        ? ListIcon
        : Hash

  const formattedValue = (() => {
    if (annotation.type === 'BOOLEAN') {
      if (annotation.value === 1) return 'Yes'
      if (annotation.value === 0) return 'No'
      return '—'
    }
    if (annotation.type === 'CATEGORICAL') {
      return annotation.string_value ?? '—'
    }
    return annotation.value !== undefined
      ? String(annotation.value)
      : '—'
  })()

  const timeAgo = (() => {
    const d = new Date(annotation.timestamp)
    if (Number.isNaN(d.getTime())) return annotation.timestamp
    return formatDistanceToNow(d, { addSuffix: true })
  })()

  return (
    <div className="group rounded-md border p-3">
      <div className="flex items-start justify-between gap-2">
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <Icon className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
            <span className="truncate text-sm font-medium">
              {annotation.name}
            </span>
            {annotation.source !== 'annotation' ? (
              <Badge variant="outline" className="px-1.5 py-0 text-[10px]">
                {annotation.source}
              </Badge>
            ) : null}
          </div>
          <p className="mt-0.5 text-xs text-muted-foreground">{timeAgo}</p>
        </div>
        <div className="flex items-center gap-1">
          <Badge variant="secondary" className="font-mono text-xs">
            {formattedValue}
          </Badge>
          {onDelete && annotation.source === 'annotation' ? (
            <Button
              variant="ghost"
              size="icon"
              className="h-7 w-7 opacity-0 transition-opacity group-hover:opacity-100"
              onClick={() => onDelete(annotation.id)}
              disabled={isDeleting}
              aria-label="Delete annotation"
            >
              {isDeleting ? (
                <Loader2 className="h-3.5 w-3.5 animate-spin" />
              ) : (
                <Trash2 className="h-3.5 w-3.5 text-muted-foreground hover:text-destructive" />
              )}
            </Button>
          ) : null}
        </div>
      </div>
      {annotation.reason ? (
        <Textarea
          value={annotation.reason}
          readOnly
          rows={2}
          className="mt-2 resize-none border-0 bg-muted/40 p-2 text-xs text-muted-foreground"
        />
      ) : null}
    </div>
  )
}
