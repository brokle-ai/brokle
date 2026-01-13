'use client'

import { useState } from 'react'
import { Settings } from 'lucide-react'
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
import { useUpdateQueueMutation } from '../hooks/use-annotation-queues'
import type { AnnotationQueue, CreateQueueRequest } from '../types'

interface SettingsDialogProps {
  projectId: string
  queue: AnnotationQueue
  trigger?: React.ReactNode
}

export function SettingsDialog({ projectId, queue, trigger }: SettingsDialogProps) {
  const [open, setOpen] = useState(false)
  const updateMutation = useUpdateQueueMutation(projectId, queue.id)

  const handleSubmit = async (data: CreateQueueRequest) => {
    await updateMutation.mutateAsync({
      name: data.name,
      description: data.description,
      instructions: data.instructions,
      settings: data.settings,
    })
    setOpen(false)
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        {trigger || (
          <Button variant="outline" size="sm">
            <Settings className="mr-2 h-4 w-4" />
            Settings
          </Button>
        )}
      </DialogTrigger>
      <DialogContent className="sm:max-w-[550px]">
        <DialogHeader>
          <DialogTitle>Queue Settings</DialogTitle>
          <DialogDescription>
            Update settings for &quot;{queue.name}&quot;.
          </DialogDescription>
        </DialogHeader>
        <QueueForm
          queue={queue}
          onSubmit={handleSubmit}
          onCancel={() => setOpen(false)}
          isLoading={updateMutation.isPending}
        />
      </DialogContent>
    </Dialog>
  )
}
