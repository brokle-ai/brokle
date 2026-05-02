import { useState, type ReactNode } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { addQueueItems, queuesKeys } from '../api/queries'
import { AddItemsForm } from './add-items-form'
import type { AddItemsBatchRequest } from '../api/types'

interface AddItemsDialogStandaloneProps {
  projectId: string
  queueId: string
  queueName: string
  trigger?: ReactNode
}

/**
 * Stand-alone "Add items" dialog hosted on the queue-detail header.
 * Wraps `AddItemsForm` and POSTs the batch to the bulk-add endpoint.
 * Invalidates the queue-detail (stats) + items-list caches so the new
 * items appear in the table without a manual refresh.
 */
export function AddItemsDialogStandalone({
  projectId,
  queueId,
  queueName,
  trigger,
}: AddItemsDialogStandaloneProps) {
  const [open, setOpen] = useState(false)
  const queryClient = useQueryClient()

  const mutation = useMutation({
    mutationFn: (data: AddItemsBatchRequest) =>
      addQueueItems(projectId, queueId, data),
    onSuccess: (resp) => {
      queryClient.invalidateQueries({ queryKey: queuesKeys.detail(queueId) })
      queryClient.invalidateQueries({ queryKey: queuesKeys.lists() })
      toast.success(`Added ${resp.created} item${resp.created === 1 ? '' : 's'}`, {
        description: `Items added to "${queueName}".`,
      })
      setOpen(false)
    },
    onError: (err) => {
      toast.error('Failed to add items', {
        description: err instanceof Error ? err.message : 'Please try again.',
      })
    },
  })

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        {trigger ?? (
          <Button variant="outline" size="sm">
            <Plus className="mr-2 h-4 w-4" />
            Add items
          </Button>
        )}
      </DialogTrigger>
      <DialogContent className="sm:max-w-[500px] max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Add Items to Queue</DialogTitle>
          <DialogDescription>
            Add traces or spans to &ldquo;{queueName}&rdquo; for review.
          </DialogDescription>
        </DialogHeader>
        <AddItemsForm
          onSubmit={(data) => mutation.mutate(data)}
          onCancel={() => setOpen(false)}
          isLoading={mutation.isPending}
        />
      </DialogContent>
    </Dialog>
  )
}
