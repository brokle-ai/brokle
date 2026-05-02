import { useEffect, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
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
import {
  inviteMember,
  memberKeys,
  organizationRolesQueryOptions,
} from '../api/queries'
import type { InviteMemberRequest } from '../api/types'

interface InviteMemberDialogProps {
  orgId: string
  open: boolean
  onOpenChange: (open: boolean) => void
}

// "Invite member" flow. Email + role select + optional message. The
// role list comes from GET /api/v1/rbac/roles?scope_type=organization
// and we hide `owner` (only one owner per org is allowed).
export function InviteMemberDialog({
  orgId,
  open,
  onOpenChange,
}: InviteMemberDialogProps) {
  const queryClient = useQueryClient()
  const rolesQuery = useQuery(organizationRolesQueryOptions())

  const [email, setEmail] = useState('')
  const [roleId, setRoleId] = useState<string>('')
  const [message, setMessage] = useState('')
  const [error, setError] = useState<string | null>(null)

  // Default role selection: prefer `developer`, fall back to the first
  // non-owner role returned by the backend. Runs each time the list
  // resolves; never overrides an explicit user choice.
  useEffect(() => {
    if (!rolesQuery.data || roleId) return
    const assignable = rolesQuery.data.data.filter((r) => r.name !== 'owner')
    const dev = assignable.find((r) => r.name === 'developer')
    setRoleId((dev ?? assignable[0])?.id ?? '')
  }, [rolesQuery.data, roleId])

  // Reset form state whenever the dialog closes so the next open
  // starts clean.
  useEffect(() => {
    if (!open) {
      setEmail('')
      setMessage('')
      setError(null)
    }
  }, [open])

  const mutation = useMutation({
    mutationFn: async (data: InviteMemberRequest) =>
      inviteMember(orgId, data),
    onSuccess: async (_inv, vars) => {
      await queryClient.invalidateQueries({
        queryKey: memberKeys.invitationList(orgId),
      })
      await queryClient.invalidateQueries({ queryKey: memberKeys.lists() })
      toast.success(`Invitation sent to ${vars.email}`)
      onOpenChange(false)
    },
    onError: (err) => {
      const msg =
        err instanceof Error ? err.message : 'Failed to send invitation'
      setError(msg)
      toast.error(msg)
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    if (!email.trim()) {
      setError('Email is required')
      return
    }
    if (!roleId) {
      setError('Select a role')
      return
    }
    mutation.mutate({
      email: email.trim(),
      role_id: roleId,
      message: message.trim() ? message.trim() : undefined,
    })
  }

  const assignableRoles =
    rolesQuery.data?.data.filter((r) => r.name !== 'owner') ?? []

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Invite member</DialogTitle>
          <DialogDescription>
            Send an invitation email to join this organization.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="invite-email">Email</Label>
            <Input
              id="invite-email"
              type="email"
              required
              autoFocus
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="teammate@example.com"
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="invite-role">Role</Label>
            <Select value={roleId} onValueChange={setRoleId}>
              <SelectTrigger id="invite-role">
                <SelectValue placeholder="Select a role" />
              </SelectTrigger>
              <SelectContent>
                {assignableRoles.map((r) => (
                  <SelectItem key={r.id} value={r.id} className="capitalize">
                    {r.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {rolesQuery.isError && (
              <p className="text-xs text-destructive">
                Failed to load roles.{' '}
                {rolesQuery.error instanceof Error
                  ? rolesQuery.error.message
                  : ''}
              </p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="invite-message">Message (optional)</Label>
            <Input
              id="invite-message"
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              maxLength={500}
              placeholder="Welcome aboard!"
            />
          </div>

          {error && (
            <p className="text-sm text-destructive" role="alert">
              {error}
            </p>
          )}

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
              {mutation.isPending ? 'Sending…' : 'Send invitation'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
