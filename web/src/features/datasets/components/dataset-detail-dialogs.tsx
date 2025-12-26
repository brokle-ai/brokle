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
import { DatasetForm } from './dataset-form'
import { useDatasetDetail } from '../context/dataset-detail-context'

export function DatasetDetailDialogs() {
  const {
    dataset,
    open,
    setOpen,
    handleUpdate,
    handleDelete,
    isUpdating,
    isDeleting,
  } = useDatasetDetail()

  if (!dataset) return null

  return (
    <>
      <Dialog open={open === 'edit'} onOpenChange={(isOpen) => setOpen(isOpen ? 'edit' : null)}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle>Edit Dataset</DialogTitle>
            <DialogDescription>
              Update the dataset name and description.
            </DialogDescription>
          </DialogHeader>
          <DatasetForm
            dataset={dataset}
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
            <AlertDialogTitle>Delete Dataset</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete &quot;{dataset.name}&quot;? This will also
              delete all items in this dataset. This action cannot be undone.
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
