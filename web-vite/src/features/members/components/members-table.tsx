import { useState } from 'react'
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
import { RemoveMemberDialog } from './remove-member-dialog'
import type { MemberListItem } from '../api/types'

interface MembersTableProps {
  orgId: string
  rows: MemberListItem[]
}

function formatTimestamp(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString()
}

function StatusBadge({ status }: { status: string }) {
  if (status === 'active') {
    return <Badge variant="secondary">Active</Badge>
  }
  if (status === 'invited') {
    return <Badge variant="outline">Invited</Badge>
  }
  if (status === 'suspended') {
    return <Badge variant="destructive">Suspended</Badge>
  }
  return <Badge variant="outline">{status}</Badge>
}

export function MembersTable({ orgId, rows }: MembersTableProps) {
  // One dialog instance per table — the row click sets which user it
  // targets. Keeps the component tree flat vs. a dialog per row.
  const [removeTarget, setRemoveTarget] = useState<MemberListItem | null>(null)

  if (rows.length === 0) {
    return (
      <div className="rounded-lg border p-12 text-center">
        <p className="text-sm text-muted-foreground">No members yet</p>
        <p className="mt-1 text-xs text-muted-foreground">
          Invite teammates to collaborate in this organization.
        </p>
      </div>
    )
  }

  return (
    <>
      <div className="rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>User</TableHead>
              <TableHead>Role</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Joined</TableHead>
              <TableHead className="w-[1%] text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {rows.map((m) => (
              <TableRow key={m.user_id}>
                <TableCell className="font-mono text-xs text-muted-foreground">
                  {m.user_id.slice(0, 8)}…
                </TableCell>
                <TableCell className="font-mono text-xs text-muted-foreground">
                  {m.role_id.slice(0, 8)}…
                </TableCell>
                <TableCell>
                  <StatusBadge status={m.status} />
                </TableCell>
                <TableCell className="text-muted-foreground">
                  {formatTimestamp(m.joined_at)}
                </TableCell>
                <TableCell className="text-right">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setRemoveTarget(m)}
                  >
                    Remove
                  </Button>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>

      {removeTarget && (
        <RemoveMemberDialog
          orgId={orgId}
          userId={removeTarget.user_id}
          displayLabel={`${removeTarget.user_id.slice(0, 8)}…`}
          open={!!removeTarget}
          onOpenChange={(open) => {
            if (!open) setRemoveTarget(null)
          }}
        />
      )}
    </>
  )
}
