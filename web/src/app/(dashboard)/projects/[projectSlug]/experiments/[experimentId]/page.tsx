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
  useExperimentQuery,
  useUpdateExperimentMutation,
  useDeleteExperimentMutation,
  ExperimentForm,
  ExperimentItemTable,
  ExperimentStatusBadge,
} from '@/features/experiments'
import type { UpdateExperimentRequest } from '@/features/experiments'

export default function ExperimentDetailPage() {
  const params = useParams<{ projectSlug: string; experimentId: string }>()
  const router = useRouter()
  const { currentProject, hasProject, isLoading: projectLoading } = useProjectOnly()

  const [isEditDialogOpen, setIsEditDialogOpen] = useState(false)
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false)

  const projectId = currentProject?.id ?? ''
  const experimentId = params.experimentId

  const { data: experiment, isLoading: experimentLoading } = useExperimentQuery(
    projectId,
    experimentId
  )
  const updateMutation = useUpdateExperimentMutation(projectId, experimentId)
  const deleteMutation = useDeleteExperimentMutation(projectId)

  const isLoading = projectLoading || experimentLoading

  const handleUpdate = async (data: UpdateExperimentRequest) => {
    await updateMutation.mutateAsync(data)
    setIsEditDialogOpen(false)
  }

  const handleDelete = async () => {
    if (experiment) {
      await deleteMutation.mutateAsync({
        experimentId: experiment.id,
        experimentName: experiment.name,
      })
      router.push(`/projects/${params.projectSlug}/experiments`)
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

  if (!experiment) {
    return (
      <>
        <DashboardHeader />
        <Main>
          <div className="flex flex-col items-center justify-center py-12">
            <p className="text-lg font-medium">Experiment not found</p>
            <Link
              href={`/projects/${params.projectSlug}/experiments`}
              className="text-sm text-muted-foreground hover:underline mt-2"
            >
              Back to experiments
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
                  href={`/projects/${params.projectSlug}/experiments`}
                  className="text-muted-foreground hover:text-foreground"
                >
                  <ArrowLeft className="h-5 w-5" />
                </Link>
                <h1 className="text-2xl font-bold tracking-tight">
                  {experiment.name}
                </h1>
                <ExperimentStatusBadge status={experiment.status} />
              </div>
              {experiment.description && (
                <p className="text-muted-foreground ml-8">
                  {experiment.description}
                </p>
              )}
              <p className="text-sm text-muted-foreground ml-8">
                Created{' '}
                {formatDistanceToNow(new Date(experiment.created_at), {
                  addSuffix: true,
                })}
                {experiment.started_at && (
                  <>
                    {' • Started '}
                    {formatDistanceToNow(new Date(experiment.started_at), {
                      addSuffix: true,
                    })}
                  </>
                )}
                {experiment.completed_at && (
                  <>
                    {' • Completed '}
                    {formatDistanceToNow(new Date(experiment.completed_at), {
                      addSuffix: true,
                    })}
                  </>
                )}
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
              <h2 className="text-lg font-medium">Experiment Items</h2>
            </div>
            <ExperimentItemTable
              projectId={projectId}
              projectSlug={params.projectSlug}
              experimentId={experimentId}
            />
          </div>
        </div>
      </Main>

      {/* Edit Dialog */}
      <Dialog open={isEditDialogOpen} onOpenChange={setIsEditDialogOpen}>
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
            onCancel={() => setIsEditDialogOpen(false)}
            isLoading={updateMutation.isPending}
          />
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation */}
      <AlertDialog open={isDeleteDialogOpen} onOpenChange={setIsDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Experiment</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete &quot;{experiment.name}&quot;? This
              will also delete all items in this experiment. This action cannot be
              undone.
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
