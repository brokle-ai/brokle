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
import { deleteEvaluator, evaluatorsKeys } from '../api/queries'

interface EvaluatorDeleteButtonProps {
  projectId: string
  evaluatorId: string
  onDeleted: () => void
}

export function EvaluatorDeleteButton({
  projectId,
  evaluatorId,
  onDeleted,
}: EvaluatorDeleteButtonProps) {
  const [open, setOpen] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const queryClient = useQueryClient()

  const mutation = useMutation({
    mutationFn: () => deleteEvaluator(projectId, evaluatorId),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: evaluatorsKeys.lists() })
      queryClient.removeQueries({
        queryKey: evaluatorsKeys.detail(evaluatorId),
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
            : 'Failed to delete evaluator',
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
          <AlertDialogTitle>Delete this evaluator?</AlertDialogTitle>
          <AlertDialogDescription>
            This permanently removes the evaluator and stops any
            automatic scoring driven by it. Scores it previously
            produced are retained.
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
            {mutation.isPending ? 'Deleting…' : 'Delete evaluator'}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
