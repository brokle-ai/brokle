'use client'

import { Loader2 } from 'lucide-react'
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
import { EditEvaluatorDialog } from './edit-evaluator-dialog'
import { useEvaluatorDetail } from '../context/evaluator-detail-context'

export function EvaluatorDetailDialogs() {
  const {
    evaluator,
    open,
    setOpen,
    handleDelete,
    isDeleting,
    projectId,
  } = useEvaluatorDetail()

  if (!evaluator) return null

  return (
    <>
      <EditEvaluatorDialog
        projectId={projectId}
        evaluator={evaluator}
        open={open === 'edit'}
        onOpenChange={(isOpen) => setOpen(isOpen ? 'edit' : null)}
      />

      <AlertDialog
        open={open === 'delete'}
        onOpenChange={(isOpen) => setOpen(isOpen ? 'delete' : null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Evaluator</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete &quot;{evaluator.name}&quot;? This action cannot be undone.
              The evaluator will stop evaluating spans immediately.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isDeleting}>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              disabled={isDeleting}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {isDeleting ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Deleting...
                </>
              ) : (
                'Delete'
              )}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
