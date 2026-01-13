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
import { QueueForm } from './queue-form'
import { useCreateQueueMutation } from '../hooks/use-annotation-queues'
import type { CreateQueueRequest } from '../types'

interface CreateQueueDialogProps {
  projectId: string
}

export function CreateQueueDialog({ projectId }: CreateQueueDialogProps) {
  const [open, setOpen] = useState(false)
  const createMutation = useCreateQueueMutation(projectId)

  const handleSubmit = async (data: CreateQueueRequest) => {
    await createMutation.mutateAsync(data)
    setOpen(false)
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button>
          <Plus className="mr-2 h-4 w-4" />
          Create Queue
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[550px]">
        <DialogHeader>
          <DialogTitle>Create Annotation Queue</DialogTitle>
          <DialogDescription>
            Create a new queue for human-in-the-loop evaluation. Add traces or spans for annotators to review and score.
          </DialogDescription>
        </DialogHeader>
        <QueueForm
          onSubmit={handleSubmit}
          onCancel={() => setOpen(false)}
          isLoading={createMutation.isPending}
        />
      </DialogContent>
    </Dialog>
  )
}
