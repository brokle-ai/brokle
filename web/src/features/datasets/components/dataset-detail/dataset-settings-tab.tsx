'use client'

import { useState, useEffect } from 'react'
import { formatDistanceToNow } from 'date-fns'
import { Loader2, AlertTriangle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
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
import { Skeleton } from '@/components/ui/skeleton'
import { useDatasetDetail } from '../../context/dataset-detail-context'

export function DatasetSettingsTab() {
  const {
    dataset,
    isLoading,
    handleUpdate,
    handleDelete,
    isUpdating,
    isDeleting,
  } = useDatasetDetail()

  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [confirmName, setConfirmName] = useState('')
  const [hasChanges, setHasChanges] = useState(false)

  // Initialize form when dataset loads
  useEffect(() => {
    if (dataset) {
      setName(dataset.name)
      setDescription(dataset.description || '')
      setHasChanges(false)
    }
  }, [dataset])

  // Reset form to current dataset values
  const resetForm = () => {
    if (dataset) {
      setName(dataset.name)
      setDescription(dataset.description || '')
      setHasChanges(false)
    }
  }

  const handleNameChange = (value: string) => {
    setName(value)
    setHasChanges(value !== dataset?.name || description !== (dataset?.description || ''))
  }

  const handleDescriptionChange = (value: string) => {
    setDescription(value)
    setHasChanges(name !== dataset?.name || value !== (dataset?.description || ''))
  }

  const handleSave = async () => {
    await handleUpdate({
      name: name !== dataset?.name ? name : undefined,
      description: description !== (dataset?.description || '') ? description : undefined,
    })
    setHasChanges(false)
  }

  const handleDeleteConfirm = async () => {
    await handleDelete()
    setDeleteDialogOpen(false)
  }

  if (isLoading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-[200px]" />
        <Skeleton className="h-[150px]" />
      </div>
    )
  }

  if (!dataset) {
    return null
  }

  const canDelete = confirmName === dataset.name

  return (
    <div className="space-y-6 max-w-2xl">
      {/* General Settings */}
      <Card>
        <CardHeader>
          <CardTitle>General</CardTitle>
          <CardDescription>
            Update your dataset name and description
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="name">Name</Label>
            <Input
              id="name"
              value={name}
              onChange={(e) => handleNameChange(e.target.value)}
              placeholder="Dataset name"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="description">Description</Label>
            <Textarea
              id="description"
              value={description}
              onChange={(e) => handleDescriptionChange(e.target.value)}
              placeholder="Optional description"
              rows={3}
            />
          </div>
        </CardContent>
        <CardFooter className="flex justify-between border-t pt-6">
          <div className="text-sm text-muted-foreground">
            Last updated {formatDistanceToNow(new Date(dataset.updated_at), { addSuffix: true })}
          </div>
          <div className="flex gap-2">
            <Button
              variant="outline"
              onClick={resetForm}
              disabled={!hasChanges || isUpdating}
            >
              Cancel
            </Button>
            <Button
              onClick={handleSave}
              disabled={!hasChanges || isUpdating}
            >
              {isUpdating && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Save Changes
            </Button>
          </div>
        </CardFooter>
      </Card>

      {/* Danger Zone */}
      <Card className="border-destructive/50">
        <CardHeader>
          <CardTitle className="text-destructive flex items-center gap-2">
            <AlertTriangle className="h-5 w-5" />
            Danger Zone
          </CardTitle>
          <CardDescription>
            Irreversible and destructive actions
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-between rounded-lg border border-destructive/20 p-4">
            <div>
              <h4 className="font-medium">Delete this dataset</h4>
              <p className="text-sm text-muted-foreground">
                Permanently delete this dataset and all its items. This action cannot be undone.
              </p>
            </div>
            <Button
              variant="destructive"
              onClick={() => setDeleteDialogOpen(true)}
            >
              Delete Dataset
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Delete Confirmation Dialog */}
      <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Dataset</AlertDialogTitle>
            <AlertDialogDescription>
              This action cannot be undone. This will permanently delete the dataset
              <strong className="text-foreground"> &quot;{dataset.name}&quot; </strong>
              and all of its items.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <div className="py-4">
            <Label htmlFor="confirm-name" className="text-sm">
              Type <strong>{dataset.name}</strong> to confirm:
            </Label>
            <Input
              id="confirm-name"
              value={confirmName}
              onChange={(e) => setConfirmName(e.target.value)}
              placeholder={dataset.name}
              className="mt-2"
            />
          </div>
          <AlertDialogFooter>
            <AlertDialogCancel
              onClick={() => setConfirmName('')}
              disabled={isDeleting}
            >
              Cancel
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDeleteConfirm}
              disabled={!canDelete || isDeleting}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {isDeleting ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Deleting...
                </>
              ) : (
                'Delete Dataset'
              )}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
