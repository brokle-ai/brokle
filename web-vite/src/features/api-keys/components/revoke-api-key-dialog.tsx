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
import { apiKeyKeys, revokeApiKey } from '../api/queries'

interface RevokeApiKeyDialogProps {
  projectId: string
  keyId: string
  keyName: string
  open: boolean
  onOpenChange: (open: boolean) => void
}

// Destructive confirm on API-key revocation. The underlying DELETE is
// permanent — SDK clients carrying the key start getting 401 on their
// next request.
export function RevokeApiKeyDialog({
  projectId,
  keyId,
  keyName,
  open,
  onOpenChange,
}: RevokeApiKeyDialogProps) {
  const queryClient = useQueryClient()

  const mutation = useMutation({
    mutationFn: async () => revokeApiKey(projectId, keyId),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: apiKeyKeys.lists() })
      toast.success(`API key "${keyName}" revoked`)
      onOpenChange(false)
    },
    onError: (err) => {
      toast.error(
        err instanceof Error ? err.message : 'Failed to revoke API key',
      )
    },
  })

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Revoke API key?</AlertDialogTitle>
          <AlertDialogDescription>
            {`Revoke "${keyName}"? Any client using this key will start receiving 401 on its next request. This can't be undone.`}
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
            {mutation.isPending ? 'Revoking…' : 'Revoke'}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
