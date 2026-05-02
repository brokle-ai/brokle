import { useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
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
import { deleteTraceScore, scoresKeys } from '../api/queries'

interface DeleteScoreDialogProps {
  scoreId: string
  projectId: string
  traceId: string | undefined
  scoreName: string
  open: boolean
  onOpenChange: (open: boolean) => void
}

// Destructive confirm on score deletion. Backend only allows deleting
// trace-scoped scores — dataset/experiment scores can't be deleted
// through this surface yet. Guard at the UI by disabling the action
// when `trace_id` is undefined.
export function DeleteScoreDialog({
  scoreId,
  projectId,
  traceId,
  scoreName,
  open,
  onOpenChange,
}: DeleteScoreDialogProps) {
  const queryClient = useQueryClient()

  const mutation = useMutation({
    mutationFn: async () => {
      if (!traceId) throw new Error('Only trace-scoped scores can be deleted')
      return deleteTraceScore(projectId, traceId, scoreId)
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: scoresKeys.lists() })
      toast.success(`Score "${scoreName}" deleted`)
      onOpenChange(false)
    },
    onError: (err) => {
      toast.error(
        err instanceof Error ? err.message : 'Failed to delete score',
      )
    },
  })

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete score?</AlertDialogTitle>
          <AlertDialogDescription>
            {`Delete "${scoreName}"? This removes the score annotation from the associated trace permanently.`}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={mutation.isPending}>
            Cancel
          </AlertDialogCancel>
          <AlertDialogAction
            disabled={mutation.isPending || !traceId}
            onClick={(e) => {
              e.preventDefault()
              mutation.mutate()
            }}
          >
            {mutation.isPending ? 'Deleting…' : 'Delete'}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
