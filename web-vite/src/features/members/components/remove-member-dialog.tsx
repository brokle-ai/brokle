import { useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
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
import { memberKeys, removeMember } from '../api/queries'

interface RemoveMemberDialogProps {
  orgId: string
  userId: string
  open: boolean
  onOpenChange: (open: boolean) => void
  // Used for the confirmation copy — truncated UUID if the list
  // endpoint hasn't been extended with identity fields yet.
  displayLabel: string
}

// Destructive confirm on member removal. Revokes the user's membership
// at the org level; downstream project memberships cascade server-side.
export function RemoveMemberDialog({
  orgId,
  userId,
  open,
  onOpenChange,
  displayLabel,
}: RemoveMemberDialogProps) {
  const queryClient = useQueryClient()

  const mutation = useMutation({
    mutationFn: async () => removeMember(orgId, userId),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: memberKeys.lists() })
      toast.success('Member removed')
      onOpenChange(false)
    },
    onError: (err) => {
      toast.error(
        err instanceof Error ? err.message : 'Failed to remove member',
      )
    },
  })

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Remove member?</AlertDialogTitle>
          <AlertDialogDescription>
            {`Remove ${displayLabel} from this organization? They'll lose access to every project immediately.`}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={mutation.isPending}>
            Cancel
          </AlertDialogCancel>
          <AlertDialogAction
            onClick={(e) => {
              e.preventDefault()
              mutation.mutate()
            }}
            disabled={mutation.isPending}
            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
          >
            {mutation.isPending ? 'Removing…' : 'Remove'}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
