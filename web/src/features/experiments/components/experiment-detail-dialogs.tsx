'use client'

import { Loader2 } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
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
import { ExperimentForm } from './experiment-form'
import { useExperimentDetail } from '../context/experiment-detail-context'

export function ExperimentDetailDialogs() {
  const {
    experiment,
    open,
    setOpen,
    handleUpdate,
    handleDelete,
    handleRerun,
    isUpdating,
    isDeleting,
    isRerunning,
  } = useExperimentDetail()

  if (!experiment) return null

  return (
    <>
      <Dialog open={open === 'edit'} onOpenChange={(isOpen) => setOpen(isOpen ? 'edit' : null)}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle>Edit Experiment</DialogTitle>
            <DialogDescription>
              Update the experiment name and description.
            </DialogDescription>
          </DialogHeader>
          <ExperimentForm
            experiment={experiment}
            onSubmit={handleUpdate}
            onCancel={() => setOpen(null)}
            isLoading={isUpdating}
          />
        </DialogContent>
      </Dialog>

      <AlertDialog
        open={open === 'delete'}
        onOpenChange={(isOpen) => setOpen(isOpen ? 'delete' : null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Experiment</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete &quot;{experiment.name}&quot;? This will also
              delete all items in this experiment. This action cannot be undone.
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

      <AlertDialog
        open={open === 'rerun'}
        onOpenChange={(isOpen) => setOpen(isOpen ? 'rerun' : null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Re-run Experiment</AlertDialogTitle>
            <AlertDialogDescription>
              This will create a new experiment &quot;{experiment.name} (Re-run)&quot; using the
              same dataset. You can then run your evaluation task to populate results with
              potentially different configurations or model versions.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isRerunning}>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => handleRerun()}
              disabled={isRerunning}
            >
              {isRerunning ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Creating...
                </>
              ) : (
                'Create Re-run'
              )}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
