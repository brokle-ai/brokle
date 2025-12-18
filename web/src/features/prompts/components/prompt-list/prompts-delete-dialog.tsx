'use client'

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
import { useDeletePromptMutation } from '../../hooks/use-prompts-queries'
import type { PromptListItem } from '../../types'

interface PromptsDeleteDialogProps {
  prompt: PromptListItem | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function PromptsDeleteDialog({ prompt, open, onOpenChange }: PromptsDeleteDialogProps) {
  const { currentProject } = useProjectOnly()
  const deleteMutation = useDeletePromptMutation(currentProject?.id || '')

  const handleDelete = async () => {
    if (!prompt || !currentProject?.id) return

    await deleteMutation.mutateAsync({
      promptId: prompt.id,
      promptName: prompt.name,
    })
    onOpenChange(false)
  }

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete prompt</AlertDialogTitle>
          <AlertDialogDescription>
            Are you sure you want to delete <strong>{prompt?.name}</strong>? This action cannot be
            undone. All versions and labels associated with this prompt will be permanently deleted.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction
            onClick={handleDelete}
            disabled={deleteMutation.isPending}
            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
          >
            {deleteMutation.isPending ? 'Deleting...' : 'Delete'}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
