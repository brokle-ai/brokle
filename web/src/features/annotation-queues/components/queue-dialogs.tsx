'use client'

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { QueueForm } from './queue-form'
import { AddItemsForm } from './add-items-form'
import { useAnnotationQueues } from '../context/annotation-queues-context'
import {
  useUpdateQueueMutation,
  useDeleteQueueMutation,
  useAddItemsMutation,
} from '../hooks/use-annotation-queues'
import type { CreateQueueRequest, AddItemsBatchRequest } from '../types'

export function QueueDialogs() {
  const { open, setOpen, currentRow, projectId } = useAnnotationQueues()

  if (!projectId) return null

  return (
    <>
      {/* Edit Queue Dialog */}
      <EditQueueDialog
        open={open === 'edit'}
        onOpenChange={(isOpen) => setOpen(isOpen ? 'edit' : null)}
        projectId={projectId}
      />

      {/* Delete Queue Dialog */}
      <DeleteQueueDialog
        open={open === 'delete'}
        onOpenChange={(isOpen) => setOpen(isOpen ? 'delete' : null)}
        projectId={projectId}
      />

      {/* Add Items Dialog */}
      <AddItemsDialog
        open={open === 'add-items'}
        onOpenChange={(isOpen) => setOpen(isOpen ? 'add-items' : null)}
        projectId={projectId}
      />
    </>
  )
}

// Edit Queue Dialog
interface EditQueueDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  projectId: string
}

function EditQueueDialog({ open, onOpenChange, projectId }: EditQueueDialogProps) {
  const { currentRow } = useAnnotationQueues()
  const updateMutation = useUpdateQueueMutation(projectId, currentRow?.id ?? '')

  const handleSubmit = async (data: CreateQueueRequest) => {
    if (!currentRow) return
    await updateMutation.mutateAsync({
      name: data.name,
      description: data.description,
      instructions: data.instructions,
      settings: data.settings,
    })
    onOpenChange(false)
  }

  if (!currentRow) return null

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[550px]">
        <DialogHeader>
          <DialogTitle>Edit Annotation Queue</DialogTitle>
          <DialogDescription>
            Update the queue settings and instructions.
          </DialogDescription>
        </DialogHeader>
        <QueueForm
          queue={currentRow}
          onSubmit={handleSubmit}
          onCancel={() => onOpenChange(false)}
          isLoading={updateMutation.isPending}
        />
      </DialogContent>
    </Dialog>
  )
}

// Delete Queue Dialog
interface DeleteQueueDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  projectId: string
}

function DeleteQueueDialog({ open, onOpenChange, projectId }: DeleteQueueDialogProps) {
  const { currentRow } = useAnnotationQueues()
  const deleteMutation = useDeleteQueueMutation(projectId)

  const handleDelete = async () => {
    if (!currentRow) return
    await deleteMutation.mutateAsync({
      queueId: currentRow.id,
      queueName: currentRow.name,
    })
    onOpenChange(false)
  }

  if (!currentRow) return null

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete Annotation Queue</AlertDialogTitle>
          <AlertDialogDescription>
            Are you sure you want to delete &quot;{currentRow.name}&quot;? This will
            permanently delete the queue and all its items. This action cannot be
            undone.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction
            onClick={handleDelete}
            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
          >
            {deleteMutation.isPending ? 'Deleting...' : 'Delete'}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}

// Add Items Dialog
interface AddItemsDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  projectId: string
}

function AddItemsDialog({ open, onOpenChange, projectId }: AddItemsDialogProps) {
  const { currentRow } = useAnnotationQueues()
  const addItemsMutation = useAddItemsMutation(projectId, currentRow?.id ?? '')

  const handleSubmit = async (data: AddItemsBatchRequest) => {
    if (!currentRow) return
    await addItemsMutation.mutateAsync(data)
    onOpenChange(false)
  }

  if (!currentRow) return null

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[550px]">
        <DialogHeader>
          <DialogTitle>Add Items to Queue</DialogTitle>
          <DialogDescription>
            Add traces or spans to &quot;{currentRow.name}&quot; for annotation.
          </DialogDescription>
        </DialogHeader>
        <AddItemsForm
          onSubmit={handleSubmit}
          onCancel={() => onOpenChange(false)}
          isLoading={addItemsMutation.isPending}
        />
      </DialogContent>
    </Dialog>
  )
}
