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
import { useDashboards } from '../context/dashboards-context'
import {
  useDeleteDashboardMutation,
  useUpdateDashboardMutation,
} from '../hooks/use-dashboards-queries'
import { DashboardForm } from './dashboard-form'
import type { CreateDashboardRequest } from '../types'

export function DashboardsDialogs() {
  const { open, setOpen, currentRow, setCurrentRow, projectId } = useDashboards()
  const deleteMutation = useDeleteDashboardMutation(projectId ?? '')
  const updateMutation = useUpdateDashboardMutation(projectId ?? '', currentRow?.id ?? '')

  const handleEditSubmit = async (data: CreateDashboardRequest) => {
    if (!currentRow) return
    await updateMutation.mutateAsync({
      name: data.name,
      description: data.description,
    })
    handleClose()
  }

  const handleDeleteConfirm = async () => {
    if (!currentRow) return
    await deleteMutation.mutateAsync({
      dashboardId: currentRow.id,
      dashboardName: currentRow.name,
    })
    handleClose()
  }

  const handleClose = () => {
    setOpen(null)
    setCurrentRow(null)
  }

  return (
    <>
      {/* Edit Dialog */}
      <Dialog open={open === 'edit'} onOpenChange={(isOpen) => !isOpen && handleClose()}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle>Edit Dashboard</DialogTitle>
            <DialogDescription>
              Update dashboard settings and configuration.
            </DialogDescription>
          </DialogHeader>
          {currentRow && (
            <DashboardForm
              dashboard={currentRow}
              onSubmit={handleEditSubmit}
              onCancel={handleClose}
              isLoading={updateMutation.isPending}
            />
          )}
        </DialogContent>
      </Dialog>

      {/* Delete Dialog */}
      <AlertDialog
        open={open === 'delete'}
        onOpenChange={(isOpen) => !isOpen && handleClose()}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Dashboard</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete &quot;{currentRow?.name}&quot;?
              <span className="block mt-2">
                This action cannot be undone.
              </span>
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
    </>
  )
}
