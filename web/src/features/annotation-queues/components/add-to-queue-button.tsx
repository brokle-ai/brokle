'use client'

import { useState } from 'react'
import { ListPlus, Plus, Loader2 } from 'lucide-react'
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
import { toast } from 'sonner'
import { useAnnotationQueuesQuery, useAddItemsToQueueMutation } from '../hooks/use-annotation-queues'
import type { ObjectType } from '../types'

interface AddToQueueButtonProps {
  projectId: string
  objectId: string
  objectType: ObjectType
  variant?: 'default' | 'ghost' | 'outline'
  size?: 'default' | 'sm' | 'icon'
}

export function AddToQueueButton({
  projectId,
  objectId,
  objectType,
  variant = 'ghost',
  size = 'icon',
}: AddToQueueButtonProps) {
  const [open, setOpen] = useState(false)
  const [addingToQueueId, setAddingToQueueId] = useState<string | null>(null)

  const { data: queuesData, isLoading } = useAnnotationQueuesQuery(projectId)
  const addToQueueMutation = useAddItemsToQueueMutation(projectId)

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
      // Check if it's a duplicate error
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

  // Filter to only active queues
  const activeQueues = queuesData?.queues?.filter((q) => q.queue.status === 'active') ?? []

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
            // Navigate to queues page - we can't use router here since we don't have org/project slug
            // This is a simple workaround
            window.location.href = window.location.pathname.replace(/\/traces.*/, '/annotation-queues')
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
