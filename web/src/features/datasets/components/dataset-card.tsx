'use client'

import { useState } from 'react'
import Link from 'next/link'
import { formatDistanceToNow } from 'date-fns'
import { Database, MoreVertical, Trash2, Pencil, Loader2 } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
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
import { useDeleteDatasetMutation } from '../hooks/use-datasets'
import type { Dataset } from '../types'

interface DatasetCardProps {
  dataset: Dataset
  projectId: string
  projectSlug: string
  onEdit?: (dataset: Dataset) => void
}

export function DatasetCard({ dataset, projectId, projectSlug, onEdit }: DatasetCardProps) {
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)
  const deleteMutation = useDeleteDatasetMutation(projectId)

  const handleDelete = async () => {
    await deleteMutation.mutateAsync({
      datasetId: dataset.id,
      datasetName: dataset.name,
    })
    setShowDeleteDialog(false)
  }

  return (
    <>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between pb-2">
          <CardTitle className="text-lg font-medium">
            <Link
              href={`/projects/${projectSlug}/datasets/${dataset.id}`}
              className="hover:underline"
            >
              {dataset.name}
            </Link>
          </CardTitle>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon">
                <MoreVertical className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              {onEdit && (
                <>
                  <DropdownMenuItem onClick={() => onEdit(dataset)}>
                    <Pencil className="mr-2 h-4 w-4" />
                    Edit
                  </DropdownMenuItem>
                  <DropdownMenuSeparator />
                </>
              )}
              <DropdownMenuItem
                className="text-destructive focus:text-destructive"
                onClick={() => setShowDeleteDialog(true)}
              >
                <Trash2 className="mr-2 h-4 w-4" />
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground mb-4 line-clamp-2">
            {dataset.description || 'No description'}
          </p>
          <div className="flex items-center justify-between text-sm">
            <span className="flex items-center gap-1 text-muted-foreground">
              <Database className="h-4 w-4" />
              Dataset
            </span>
            <span className="text-muted-foreground">
              {formatDistanceToNow(new Date(dataset.created_at), { addSuffix: true })}
            </span>
          </div>
        </CardContent>
      </Card>

      <AlertDialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
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
