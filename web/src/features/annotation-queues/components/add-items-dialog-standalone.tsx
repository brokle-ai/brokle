'use client'

import { useState } from 'react'
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
import { AddItemsForm } from './add-items-form'
import { useAddItemsMutation } from '../hooks/use-annotation-queues'
import type { AddItemsBatchRequest } from '../types'

interface AddItemsDialogStandaloneProps {
  projectId: string
  queueId: string
  queueName: string
  trigger?: React.ReactNode
}

export function AddItemsDialogStandalone({
  projectId,
  queueId,
  queueName,
  trigger,
}: AddItemsDialogStandaloneProps) {
  const [open, setOpen] = useState(false)
  const addItemsMutation = useAddItemsMutation(projectId, queueId)

  const handleSubmit = async (data: AddItemsBatchRequest) => {
    await addItemsMutation.mutateAsync(data)
    setOpen(false)
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        {trigger || (
          <Button variant="outline" size="sm">
            <Plus className="mr-2 h-4 w-4" />
            Add Items
          </Button>
        )}
      </DialogTrigger>
      <DialogContent className="sm:max-w-[550px]">
        <DialogHeader>
          <DialogTitle>Add Items to Queue</DialogTitle>
          <DialogDescription>
            Add traces or spans to &quot;{queueName}&quot; for annotation.
          </DialogDescription>
        </DialogHeader>
        <AddItemsForm
          onSubmit={handleSubmit}
          onCancel={() => setOpen(false)}
          isLoading={addItemsMutation.isPending}
        />
      </DialogContent>
    </Dialog>
  )
}
