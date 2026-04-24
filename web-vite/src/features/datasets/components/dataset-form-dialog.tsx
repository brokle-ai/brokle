import { useEffect, useState, type FormEvent } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { BrokleError } from '@/lib/api/errors'
import {
  createDataset,
  datasetsKeys,
  updateDataset,
} from '../api/queries'
import type { DatasetDetail } from '../api/types'

type Mode =
  | { kind: 'create' }
  | { kind: 'edit'; dataset: DatasetDetail }

interface DatasetFormDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  projectId: string
  mode: Mode
  // Optional callback when the create/update succeeds — the parent
  // uses this to navigate into the new detail page on create.
  onSuccess?: (dataset: DatasetDetail) => void
}

function extractError(err: unknown): string {
  if (err instanceof BrokleError) return err.message
  if (err instanceof Error) return err.message
  return 'Something went wrong'
}

// Shared create + edit dialog. In edit mode, `mode.dataset` seeds the
// initial state; any change to `open` re-syncs so reopening for a
// different row doesn't leak stale field values.
export function DatasetFormDialog({
  open,
  onOpenChange,
  projectId,
  mode,
  onSuccess,
}: DatasetFormDialogProps) {
  const initialName = mode.kind === 'edit' ? mode.dataset.name : ''
  const initialDesc =
    mode.kind === 'edit' ? mode.dataset.description ?? '' : ''

  const [name, setName] = useState(initialName)
  const [description, setDescription] = useState(initialDesc)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (open) {
      setName(initialName)
      setDescription(initialDesc)
      setError(null)
    }
  }, [open, initialName, initialDesc])

  const queryClient = useQueryClient()

  const mutation = useMutation({
    mutationFn: async () => {
      const trimmedDesc = description.trim()
      if (mode.kind === 'create') {
        return createDataset(projectId, {
          name,
          description: trimmedDesc.length > 0 ? trimmedDesc : undefined,
        })
      }
      return updateDataset(projectId, mode.dataset.id, {
        name,
        // Empty string clears the description — send an empty string
        // rather than omitting so the backend writes NULL semantics.
        description: trimmedDesc,
      })
    },
    onSuccess: async (result) => {
      await queryClient.invalidateQueries({ queryKey: datasetsKeys.lists() })
      if (mode.kind === 'edit') {
        await queryClient.invalidateQueries({
          queryKey: datasetsKeys.detail(mode.dataset.id),
        })
      }
      onOpenChange(false)
      onSuccess?.(result)
    },
    onError: (err) => setError(extractError(err)),
  })

  function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    setError(null)
    mutation.mutate()
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          <DialogHeader>
            <DialogTitle>
              {mode.kind === 'create' ? 'New dataset' : 'Edit dataset'}
            </DialogTitle>
            <DialogDescription>
              {mode.kind === 'create'
                ? 'Create a dataset to curate inputs and expected outputs for your evaluations.'
                : 'Rename or update the description of this dataset.'}
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-2">
            <Label htmlFor="dataset-name">Name</Label>
            <Input
              id="dataset-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              disabled={mutation.isPending}
              required
              maxLength={255}
              placeholder="support-tickets-v1"
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="dataset-description">Description</Label>
            <Textarea
              id="dataset-description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              disabled={mutation.isPending}
              rows={3}
              placeholder="Optional context on what this dataset contains."
            />
          </div>

          {error ? (
            <div className="rounded-md border border-destructive/40 bg-destructive/5 p-3 text-sm text-destructive">
              {error}
            </div>
          ) : null}

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={mutation.isPending}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={mutation.isPending}>
              {mutation.isPending
                ? 'Saving…'
                : mode.kind === 'create'
                  ? 'Create dataset'
                  : 'Save changes'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
