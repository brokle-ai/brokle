import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { ListPlus, Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  addQueueItems,
  queueListQueryOptions,
  queuesKeys,
} from '../api/queries'
import type { ObjectType } from '../api/types'

interface BulkAddToQueueButtonProps {
  projectId: string
  objectIds: string[]
  objectType?: ObjectType
  /** Disable the trigger entirely (e.g. when no rows selected). */
  disabled?: boolean
}

/**
 * Bulk-add UI for the traces table's selection toolbar. Renders a
 * trigger that opens a dialog with a queue picker (active queues
 * only) and submits the currently selected IDs to the chosen queue.
 *
 * Decoupled from `AddToQueueButton` (single-trace dropdown on the
 * trace-detail header) because the bulk surface needs an explicit
 * dialog + count + summary, not a one-click dropdown menu.
 */
export function BulkAddToQueueButton({
  projectId,
  objectIds,
  objectType = 'trace',
  disabled,
}: BulkAddToQueueButtonProps) {
  const [open, setOpen] = useState(false)
  const [queueId, setQueueId] = useState<string>('')
  const queryClient = useQueryClient()

  const queuesQuery = useQuery({
    ...queueListQueryOptions(projectId, {
      page: 1,
      limit: 100,
      status: 'active',
    }),
    enabled: open,
  })
  const queues = queuesQuery.data?.data ?? []

  const mutation = useMutation({
    mutationFn: async () => {
      if (!queueId) throw new Error('No queue selected')
      return addQueueItems(projectId, queueId, {
        items: objectIds.map((id) => ({
          object_id: id,
          object_type: objectType,
          priority: 0,
        })),
      })
    },
    onSuccess: (resp) => {
      const queueName =
        queues.find((q) => q.queue.id === queueId)?.queue.name ?? 'queue'
      queryClient.invalidateQueries({ queryKey: queuesKeys.detail(queueId) })
      queryClient.invalidateQueries({ queryKey: queuesKeys.lists() })
      toast.success(`Added ${resp.created} item${resp.created === 1 ? '' : 's'}`, {
        description: `Added to "${queueName}".`,
      })
      setOpen(false)
      setQueueId('')
    },
    onError: (err) => {
      toast.error('Failed to add to queue', {
        description: err instanceof Error ? err.message : 'Please try again.',
      })
    },
  })

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button
          variant="outline"
          size="icon"
          className="size-8"
          disabled={disabled || objectIds.length === 0}
          aria-label="Add selected to annotation queue"
          title="Add selected to annotation queue"
        >
          <ListPlus />
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Add to Annotation Queue</DialogTitle>
          <DialogDescription>
            Add {objectIds.length} selected{' '}
            {objectType === 'trace' ? 'trace' : 'span'}
            {objectIds.length === 1 ? '' : 's'} to a queue for review.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-2">
          <div className="space-y-2">
            <Label htmlFor="queue">Queue</Label>
            {queuesQuery.isLoading ? (
              <div className="flex h-10 items-center gap-2 rounded-md border px-3 py-2 text-sm text-muted-foreground">
                <Loader2 className="h-4 w-4 animate-spin" />
                Loading queues...
              </div>
            ) : queues.length === 0 ? (
              <p className="text-sm text-muted-foreground">
                No active queues. Create one first.
              </p>
            ) : (
              <Select value={queueId} onValueChange={setQueueId}>
                <SelectTrigger id="queue">
                  <SelectValue placeholder="Select a queue..." />
                </SelectTrigger>
                <SelectContent>
                  {queues.map((q) => (
                    <SelectItem key={q.queue.id} value={q.queue.id}>
                      <span className="flex items-center gap-2">
                        <span>{q.queue.name}</span>
                        <Badge variant="secondary" className="text-xs">
                          {q.stats.pending_items} pending
                        </Badge>
                      </span>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => setOpen(false)}>
            Cancel
          </Button>
          <Button
            onClick={() => mutation.mutate()}
            disabled={!queueId || mutation.isPending}
          >
            {mutation.isPending ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Adding...
              </>
            ) : (
              `Add ${objectIds.length}`
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
