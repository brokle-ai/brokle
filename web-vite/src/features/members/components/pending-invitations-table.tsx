import { useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { BrokleError } from '@/lib/api/errors'
import { memberKeys, revokeInvitation } from '../api/queries'
import type { Invitation } from '../api/types'

interface PendingInvitationsTableProps {
  orgId: string
  rows: Invitation[]
}

function formatTimestamp(iso: string | undefined): string {
  if (!iso) return '—'
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString()
}

function StatusBadge({ status }: { status: Invitation['status'] }) {
  if (status === 'pending') return <Badge variant="outline">Pending</Badge>
  if (status === 'accepted') return <Badge variant="secondary">Accepted</Badge>
  if (status === 'expired') return <Badge variant="destructive">Expired</Badge>
  return <Badge variant="destructive">Revoked</Badge>
}

function inviterLabel(inv: Invitation): string {
  const { inviter } = inv
  if (!inviter) return '—'
  const name = [inviter.first_name, inviter.last_name].filter(Boolean).join(' ')
  return name || inviter.email
}

function errorMessage(err: unknown, fallback: string): string {
  if (err instanceof BrokleError) return err.message
  if (err instanceof Error) return err.message
  return fallback
}

// Pending invitations table. Rendered below the members list on the
// /settings/members route. Revoke hits DELETE on the invitation and
// optimistically busts the list cache via invalidation.
export function PendingInvitationsTable({
  orgId,
  rows,
}: PendingInvitationsTableProps) {
  const queryClient = useQueryClient()
  const revokeMutation = useMutation({
    mutationFn: async (invitationId: string) =>
      revokeInvitation(orgId, invitationId),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: memberKeys.invitationList(orgId),
      })
      toast.success('Invitation revoked')
    },
    onError: (err) => {
      toast.error('Failed to revoke invitation', {
        description: errorMessage(err, 'Could not revoke invitation.'),
      })
    },
  })

  if (rows.length === 0) {
    return (
      <div className="rounded-lg border p-8 text-center">
        <p className="text-sm text-muted-foreground">No pending invitations</p>
      </div>
    )
  }

  return (
    <div className="rounded-lg border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Email</TableHead>
            <TableHead>Role</TableHead>
            <TableHead>Invited by</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Expires</TableHead>
            <TableHead className="w-[1%] text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.map((inv) => (
            <TableRow key={inv.id}>
              <TableCell className="font-medium">{inv.email}</TableCell>
              <TableCell className="text-muted-foreground">
                {inv.role?.display_name ?? inv.role?.name ?? '—'}
              </TableCell>
              <TableCell className="text-muted-foreground">
                {inviterLabel(inv)}
              </TableCell>
              <TableCell>
                <StatusBadge status={inv.status} />
              </TableCell>
              <TableCell className="text-muted-foreground">
                {formatTimestamp(inv.expires_at)}
              </TableCell>
              <TableCell className="text-right">
                {inv.status === 'pending' && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => revokeMutation.mutate(inv.id)}
                    disabled={
                      revokeMutation.isPending &&
                      revokeMutation.variables === inv.id
                    }
                  >
                    Revoke
                  </Button>
                )}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
