import { useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { ListPlus, Plus, Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import { rawFetch } from '@/lib/api/client'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { queueListQueryOptions, queuesKeys } from '../api/queries'
import type { AnnotationQueue, QueueStats } from '../api/types'

// The backend item endpoints expose a nominal ObjectType of 'trace' or
// 'span'. Kept in this file (not the api/types surface) because this
// button is currently the only caller and we prefer local narrowing over
// a public enum until a second consumer lands.
type ObjectType = 'trace' | 'span'

interface AddToQueueButtonProps {
  /**
   * Org + project slugs are passed in so the button can build the
   * "Create New Queue" navigation target using TanStack Router's typed
   * params API — no URL-substring parsing. Traces hosts this button
   * inside a typed route, so it already has both.
   */
  orgId: string
  projectId: string
  objectId: string
  objectType: ObjectType
  variant?: 'default' | 'ghost' | 'outline'
  size?: 'default' | 'sm' | 'icon'
}

interface AddItemsToQueueVariables {
  queueId: string
  items: Array<{
    object_id: string
    object_type: ObjectType
    priority?: number
  }>
}

/**
 * Renders a dropdown of active annotation queues the current object
 * (trace/span) can be added to. Clicking an entry POSTs to the queue's
 * items endpoint. Errors classified as duplicates surface as a toast
 * info ("already in queue") rather than a hard failure.
 */
export function AddToQueueButton({
  orgId,
  projectId,
  objectId,
  objectType,
  variant = 'ghost',
  size = 'icon',
}: AddToQueueButtonProps) {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [open, setOpen] = useState(false)
  const [addingToQueueId, setAddingToQueueId] = useState<string | null>(null)

  // Pull the active-queue list up-front when the trigger mounts. The
  // list endpoint returns queue+stats pairs so we can render the
  // pending-item counts in the dropdown without a second roundtrip.
  const { data: queuesData, isLoading } = useQuery(
    queueListQueryOptions(projectId, { page: 1, limit: 100, status: 'active' }),
  )

  const addToQueueMutation = useMutation({
    mutationFn: async ({ queueId, items }: AddItemsToQueueVariables) => {
      const resp = await rawFetch(
        `/api/v1/projects/${projectId}/annotation-queues/${queueId}/items`,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ items }),
        },
      )
      return resp.json()
    },
    onSuccess: (_result, variables) => {
      // Refresh the targeted queue's items/stats and the queue list so
      // the pending-count badges reflect the new item immediately.
      queryClient.invalidateQueries({
        queryKey: queuesKeys.detail(variables.queueId),
      })
      queryClient.invalidateQueries({
        queryKey: queuesKeys.lists(),
      })
    },
  })

  const handleAddToQueue = async (queueId: string, queueName: string) => {
    setAddingToQueueId(queueId)
    try {
      await addToQueueMutation.mutateAsync({
        queueId,
        items: [{ object_id: objectId, object_type: objectType, priority: 0 }],
      })
      toast.success('Added to Queue', {
        description: `Added to "${queueName}"`,
      })
      setOpen(false)
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to add item to queue'
      if (message.toLowerCase().includes('already') || message.toLowerCase().includes('duplicate')) {
        toast.info('Already in Queue', {
          description: `This ${objectType} is already in "${queueName}"`,
        })
      } else {
        toast.error('Failed to Add', {
          description: message,
        })
      }
    } finally {
      setAddingToQueueId(null)
    }
  }

  // `queueListQueryOptions` already filters server-side when `status`
  // is passed, but we defensively re-filter client-side so an upstream
  // contract drift does not surface archived/paused queues.
  const activeQueues =
    queuesData?.data?.filter((q: { queue: AnnotationQueue; stats: QueueStats }) =>
      q.queue.status === 'active',
    ) ?? []

  return (
    <DropdownMenu open={open} onOpenChange={setOpen}>
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <DropdownMenuTrigger asChild>
              <Button variant={variant} size={size} className={size === 'icon' ? 'h-8 w-8' : ''}>
                <ListPlus className="h-4 w-4" />
                {size !== 'icon' && <span className="ml-2">Add to Queue</span>}
              </Button>
            </DropdownMenuTrigger>
          </TooltipTrigger>
          <TooltipContent>Add to Annotation Queue</TooltipContent>
        </Tooltip>
      </TooltipProvider>

      <DropdownMenuContent align="end" className="w-[250px]">
        <DropdownMenuLabel className="flex items-center gap-2">
          <ListPlus className="h-4 w-4" />
          Add to Queue
        </DropdownMenuLabel>
        <DropdownMenuSeparator />

        {isLoading ? (
          <div className="flex items-center justify-center py-4">
            <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
          </div>
        ) : activeQueues.length === 0 ? (
          <div className="py-4 px-2 text-center text-sm text-muted-foreground">
            No annotation queues available.
            <br />
            Create one to get started.
          </div>
        ) : (
          activeQueues.map((queueWithStats) => {
            const queue = queueWithStats.queue
            const stats = queueWithStats.stats
            const isAdding = addingToQueueId === queue.id

            return (
              <DropdownMenuItem
                key={queue.id}
                onClick={() => handleAddToQueue(queue.id, queue.name)}
                disabled={isAdding}
                className="flex items-center justify-between gap-2 cursor-pointer"
              >
                <div className="flex items-center gap-2 min-w-0">
                  {isAdding ? (
                    <Loader2 className="h-4 w-4 animate-spin shrink-0" />
                  ) : (
                    <Plus className="h-4 w-4 shrink-0 text-muted-foreground" />
                  )}
                  <span className="truncate">{queue.name}</span>
                </div>
                <Badge variant="secondary" className="text-xs shrink-0">
                  {stats.pending_items} pending
                </Badge>
              </DropdownMenuItem>
            )
          })
        )}

        <DropdownMenuSeparator />
        <DropdownMenuItem
          onClick={() => {
            setOpen(false)
            navigate({
              to: '/o/$orgId/p/$projectId/annotation-queues',
              params: { orgId, projectId },
              search: { page: 1, limit: 20 },
            })
          }}
          className="text-muted-foreground"
        >
          <Plus className="mr-2 h-4 w-4" />
          Create New Queue
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
