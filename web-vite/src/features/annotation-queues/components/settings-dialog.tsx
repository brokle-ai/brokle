import { useState, type ReactNode } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
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
import { queuesKeys, updateQueue } from '../api/queries'
import type { AnnotationQueue, CreateQueueRequest } from '../api/types'

interface SettingsDialogProps {
  projectId: string
  queue: AnnotationQueue
  trigger?: ReactNode
}

/**
 * Wraps the shared `QueueForm` in a dialog for the queue-detail "gear"
 * action. Submission targets the PUT /queues/{queueId} endpoint and
 * invalidates the queue-detail + list caches so name/status/score-config
 * changes propagate without a manual refetch.
 */
export function SettingsDialog({
  projectId,
  queue,
  trigger,
}: SettingsDialogProps) {
  const [open, setOpen] = useState(false)
  const queryClient = useQueryClient()

  const mutation = useMutation({
    mutationFn: async (values: CreateQueueRequest) =>
      updateQueue(projectId, queue.id, {
        name: values.name,
        description: values.description,
        instructions: values.instructions,
        score_config_ids: values.score_config_ids,
        settings: values.settings,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queuesKeys.detail(queue.id) })
      queryClient.invalidateQueries({ queryKey: queuesKeys.lists() })
      toast.success('Queue settings saved')
      setOpen(false)
    },
    onError: (err) => {
      toast.error('Failed to save settings', {
        description: err instanceof Error ? err.message : 'Please try again.',
      })
    },
  })

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        {trigger ?? (
          <Button variant="outline" size="sm">
            <Settings className="mr-2 h-4 w-4" />
            Settings
          </Button>
        )}
      </DialogTrigger>
      <DialogContent className="sm:max-w-[550px] max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Queue Settings</DialogTitle>
          <DialogDescription>
            Update settings for &ldquo;{queue.name}&rdquo;.
          </DialogDescription>
        </DialogHeader>
        <QueueForm
          projectId={projectId}
          queue={queue}
          onSubmit={(values) => mutation.mutate(values)}
          onCancel={() => setOpen(false)}
          isLoading={mutation.isPending}
        />
      </DialogContent>
    </Dialog>
  )
}
