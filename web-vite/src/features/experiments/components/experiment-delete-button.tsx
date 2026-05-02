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
import { deleteExperiment, experimentsKeys } from '../api/queries'

interface ExperimentDeleteButtonProps {
  projectId: string
  experimentId: string
  onDeleted: () => void
}

export function ExperimentDeleteButton({
  projectId,
  experimentId,
  onDeleted,
}: ExperimentDeleteButtonProps) {
  const [open, setOpen] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const queryClient = useQueryClient()

  const mutation = useMutation({
    mutationFn: () => deleteExperiment(projectId, experimentId),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: experimentsKeys.lists() })
      queryClient.removeQueries({
        queryKey: experimentsKeys.detail(experimentId),
      })
      setOpen(false)
      onDeleted()
    },
    onError: (err) => {
      setError(
        err instanceof BrokleError
          ? err.message
          : err instanceof Error
            ? err.message
            : 'Failed to delete experiment',
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
          <AlertDialogTitle>Delete this experiment?</AlertDialogTitle>
          <AlertDialogDescription>
            This permanently removes the experiment and its run
            artefacts. Traces and scores it produced are retained.
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
              e.preventDefault()
              setError(null)
              mutation.mutate()
            }}
          >
            {mutation.isPending ? 'Deleting…' : 'Delete experiment'}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
