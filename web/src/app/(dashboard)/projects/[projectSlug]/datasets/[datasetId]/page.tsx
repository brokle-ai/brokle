'use client'

import { useState } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { formatDistanceToNow } from 'date-fns'
import { ArrowLeft, Pencil, Trash2, Loader2 } from 'lucide-react'
import Link from 'next/link'
import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
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
import { useProjectOnly } from '@/features/projects'
import {
  useDatasetQuery,
  useUpdateDatasetMutation,
  useDeleteDatasetMutation,
  DatasetForm,
  DatasetItemTable,
  AddDatasetItemDialog,
} from '@/features/datasets'
import type { UpdateDatasetRequest } from '@/features/datasets'

export default function DatasetDetailPage() {
  const params = useParams<{ projectSlug: string; datasetId: string }>()
  const router = useRouter()
  const { currentProject, hasProject, isLoading: projectLoading } = useProjectOnly()

  const [isEditDialogOpen, setIsEditDialogOpen] = useState(false)
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false)

  const projectId = currentProject?.id ?? ''
  const datasetId = params.datasetId

  const { data: dataset, isLoading: datasetLoading } = useDatasetQuery(projectId, datasetId)
  const updateMutation = useUpdateDatasetMutation(projectId, datasetId)
  const deleteMutation = useDeleteDatasetMutation(projectId)

  const isLoading = projectLoading || datasetLoading

  const handleUpdate = async (data: UpdateDatasetRequest) => {
    await updateMutation.mutateAsync(data)
    setIsEditDialogOpen(false)
  }

  const handleDelete = async () => {
    if (dataset) {
      await deleteMutation.mutateAsync({
        datasetId: dataset.id,
        datasetName: dataset.name,
      })
      router.push(`/projects/${params.projectSlug}/datasets`)
    }
  }

  if (isLoading) {
    return (
      <>
        <DashboardHeader />
        <Main>
          <div className="space-y-6">
            <div className="flex items-center gap-4">
              <Skeleton className="h-10 w-10" />
              <Skeleton className="h-8 w-64" />
            </div>
            <Skeleton className="h-4 w-48" />
            <div className="space-y-2">
              {Array.from({ length: 5 }).map((_, i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          </div>
        </Main>
      </>
    )
  }

  if (!hasProject || !currentProject) {
    return (
      <>
        <DashboardHeader />
        <Main>
          <div className="flex items-center justify-center py-12">
            <p className="text-muted-foreground">No project selected</p>
          </div>
        </Main>
      </>
    )
  }

  if (!dataset) {
    return (
      <>
        <DashboardHeader />
        <Main>
          <div className="flex flex-col items-center justify-center py-12">
            <p className="text-lg font-medium">Dataset not found</p>
            <Link
              href={`/projects/${params.projectSlug}/datasets`}
              className="text-sm text-muted-foreground hover:underline mt-2"
            >
              Back to datasets
            </Link>
          </div>
        </Main>
      </>
    )
  }

  return (
    <>
      <DashboardHeader />
      <Main>
        <div className="space-y-6">
          {/* Header */}
          <div className="flex items-start justify-between">
            <div className="space-y-1">
              <div className="flex items-center gap-3">
                <Link
                  href={`/projects/${params.projectSlug}/datasets`}
                  className="text-muted-foreground hover:text-foreground"
                >
                  <ArrowLeft className="h-5 w-5" />
                </Link>
                <h1 className="text-2xl font-bold tracking-tight">{dataset.name}</h1>
              </div>
              {dataset.description && (
                <p className="text-muted-foreground ml-8">{dataset.description}</p>
              )}
              <p className="text-sm text-muted-foreground ml-8">
                Created {formatDistanceToNow(new Date(dataset.created_at), { addSuffix: true })}
              </p>
            </div>
            <div className="flex items-center gap-2">
              <Button variant="outline" onClick={() => setIsEditDialogOpen(true)}>
                <Pencil className="mr-2 h-4 w-4" />
                Edit
              </Button>
              <Button
                variant="outline"
                className="text-destructive hover:text-destructive hover:bg-destructive/10"
                onClick={() => setIsDeleteDialogOpen(true)}
              >
                <Trash2 className="mr-2 h-4 w-4" />
                Delete
              </Button>
            </div>
          </div>

          {/* Items Section */}
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-medium">Items</h2>
              <AddDatasetItemDialog projectId={projectId} datasetId={datasetId} />
            </div>
            <DatasetItemTable projectId={projectId} datasetId={datasetId} />
          </div>
        </div>
      </Main>

      {/* Edit Dialog */}
      <Dialog open={isEditDialogOpen} onOpenChange={setIsEditDialogOpen}>
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
            onCancel={() => setIsEditDialogOpen(false)}
            isLoading={updateMutation.isPending}
          />
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation */}
      <AlertDialog open={isDeleteDialogOpen} onOpenChange={setIsDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Dataset</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete &quot;{dataset.name}&quot;? This will also
              delete all items in this dataset. This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={deleteMutation.isPending}>
              Cancel
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
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
