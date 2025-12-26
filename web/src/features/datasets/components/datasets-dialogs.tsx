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
import { useDatasets } from '../context/datasets-context'
import { useDeleteDatasetMutation } from '../hooks/use-datasets'

export function DatasetsDialogs() {
  const { open, setOpen, currentRow, setCurrentRow, projectId } = useDatasets()
  const deleteMutation = useDeleteDatasetMutation(projectId ?? '')

  const handleDeleteConfirm = async () => {
    if (!currentRow) return
    await deleteMutation.mutateAsync({
      datasetId: currentRow.id,
      datasetName: currentRow.name,
    })
    setOpen(null)
    setCurrentRow(null)
  }

  const handleClose = () => {
    setOpen(null)
    setCurrentRow(null)
  }

  return (
    <AlertDialog open={open === 'delete'} onOpenChange={(isOpen) => !isOpen && handleClose()}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Dataset</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete &quot;{currentRow?.name}&quot;? This will also
              delete all items in this dataset. This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={deleteMutation.isPending}>
              Cancel
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDeleteConfirm}
              disabled={deleteMutation.isPending}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {deleteMutation.isPending ? (
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
  )
}
