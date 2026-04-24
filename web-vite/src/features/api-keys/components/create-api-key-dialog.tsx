import { useEffect, useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { apiKeyKeys, createApiKey } from '../api/queries'
import type { ApiKey, ApiKeyExpiryOption } from '../api/types'

interface CreateApiKeyDialogProps {
  projectId: string
  open: boolean
  onOpenChange: (open: boolean) => void
}

// Two-step flow: (1) name + expiry form; (2) once the request
// succeeds, swap to a read-only confirmation that shows the full
// plaintext key ONE TIME with a copy button. The list cache is
// invalidated only after the user dismisses the confirmation — there
// is no value in refreshing the table while the dialog still has the
// one-shot secret on screen.
export function CreateApiKeyDialog({
  projectId,
  open,
  onOpenChange,
}: CreateApiKeyDialogProps) {
  const queryClient = useQueryClient()

  const [name, setName] = useState('')
  const [expiry, setExpiry] = useState<ApiKeyExpiryOption>('90days')
  const [error, setError] = useState<string | null>(null)
  const [createdKey, setCreatedKey] = useState<ApiKey | null>(null)

  // Reset local state on close so the next "Create key" run starts
  // fresh. Intentional that this also runs when we programmatically
  // close after dismissing the confirmation step.
  useEffect(() => {
    if (!open) {
      setName('')
      setExpiry('90days')
      setError(null)
      setCreatedKey(null)
    }
  }, [open])

  const mutation = useMutation({
    mutationFn: async () =>
      createApiKey(projectId, { name: name.trim(), expiry_option: expiry }),
    onSuccess: (key) => {
      setCreatedKey(key)
      toast.success(`API key "${key.name}" created`)
    },
    onError: (err) => {
      const msg =
        err instanceof Error ? err.message : 'Failed to create API key'
      setError(msg)
      toast.error(msg)
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    if (name.trim().length < 2) {
      setError('Name must be at least 2 characters')
      return
    }
    mutation.mutate()
  }

  const handleClose = async () => {
    // The backend returned the plaintext key. Once the user closes
    // the confirmation we invalidate the list so the new preview
    // appears, then release the dialog.
    if (createdKey) {
      await queryClient.invalidateQueries({ queryKey: apiKeyKeys.lists() })
    }
    onOpenChange(false)
  }

  const handleCopy = async () => {
    if (!createdKey) return
    try {
      await navigator.clipboard.writeText(createdKey.key)
      toast.success('API key copied to clipboard')
    } catch {
      toast.error('Copy failed — select the text and copy manually')
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(next) => {
        // Prevent clicking outside from dropping the one-shot secret —
        // the user must explicitly hit "Done" after copying.
        if (!next && createdKey) return
        onOpenChange(next)
      }}
    >
      <DialogContent className="sm:max-w-md">
        {createdKey ? (
          <>
            <DialogHeader>
              <DialogTitle>Copy your API key now</DialogTitle>
              <DialogDescription>
                This is the only time you&apos;ll see the full key. Store it
                somewhere safe — once you close this dialog it&apos;s gone.
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-2">
              <Label htmlFor="api-key-value">API key</Label>
              <Input
                id="api-key-value"
                readOnly
                value={createdKey.key}
                onFocus={(e) => e.currentTarget.select()}
                className="font-mono text-xs"
              />
            </div>

            <DialogFooter>
              <Button variant="outline" onClick={handleCopy}>
                Copy
              </Button>
              <Button onClick={handleClose}>Done</Button>
            </DialogFooter>
          </>
        ) : (
          <form onSubmit={handleSubmit}>
            <DialogHeader>
              <DialogTitle>Create API key</DialogTitle>
              <DialogDescription>
                Issue a new API key for SDK clients to ingest telemetry on this
                project.
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-4 py-4">
              <div className="space-y-2">
                <Label htmlFor="key-name">Name</Label>
                <Input
                  id="key-name"
                  required
                  autoFocus
                  minLength={2}
                  maxLength={100}
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="e.g. production-ingest"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="key-expiry">Expires</Label>
                <Select
                  value={expiry}
                  onValueChange={(v) => setExpiry(v as ApiKeyExpiryOption)}
                >
                  <SelectTrigger id="key-expiry">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="30days">In 30 days</SelectItem>
                    <SelectItem value="90days">In 90 days</SelectItem>
                    <SelectItem value="never">Never</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              {error && (
                <p className="text-sm text-destructive" role="alert">
                  {error}
                </p>
              )}
            </div>

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
                {mutation.isPending ? 'Creating…' : 'Create key'}
              </Button>
            </DialogFooter>
          </form>
        )}
      </DialogContent>
    </Dialog>
  )
}
