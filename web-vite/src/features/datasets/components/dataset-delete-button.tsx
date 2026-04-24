import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import { Button } from '@/components/ui/button'
import { BrokleError } from '@/lib/api/errors'
import { datasetsKeys, deleteDataset } from '../api/queries'

interface DatasetDeleteButtonProps {
  projectId: string
  datasetId: string
  onDeleted: () => void
}

// AlertDialog-wrapped destructive action. Parent owns the post-delete
// navigation via `onDeleted` — we stay transport-pure and only
// invalidate caches on success.
export function DatasetDeleteButton({
  projectId,
  datasetId,
  onDeleted,
}: DatasetDeleteButtonProps) {
  const [open, setOpen] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const queryClient = useQueryClient()

  const mutation = useMutation({
    mutationFn: () => deleteDataset(projectId, datasetId),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: datasetsKeys.lists() })
      queryClient.removeQueries({ queryKey: datasetsKeys.detail(datasetId) })
      setOpen(false)
      onDeleted()
    },
    onError: (err) => {
      setError(
        err instanceof BrokleError
          ? err.message
          : err instanceof Error
            ? err.message
            : 'Failed to delete dataset',
      )
    },
  })

  return (
    <AlertDialog
      open={open}
      onOpenChange={(next) => {
        setOpen(next)
        if (!next) setError(null)
      }}
    >
      <AlertDialogTrigger asChild>
        <Button variant="outline" size="sm">
          Delete
        </Button>
      </AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete this dataset?</AlertDialogTitle>
          <AlertDialogDescription>
            This permanently removes the dataset and all of its items.
            Experiments that ran against it will retain their own item
            snapshots. This action cannot be undone.
          </AlertDialogDescription>
        </AlertDialogHeader>
        {error ? (
          <div className="rounded-md border border-destructive/40 bg-destructive/5 p-3 text-sm text-destructive">
            {error}
          </div>
        ) : null}
        <AlertDialogFooter>
          <AlertDialogCancel disabled={mutation.isPending}>
            Cancel
          </AlertDialogCancel>
          <AlertDialogAction
            disabled={mutation.isPending}
            onClick={(e) => {
              // AlertDialogAction auto-closes on click — prevent that
              // so we can keep the dialog open if the request fails.
              e.preventDefault()
              setError(null)
              mutation.mutate()
            }}
          >
            {mutation.isPending ? 'Deleting…' : 'Delete dataset'}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
