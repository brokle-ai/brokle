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
import { RevokeApiKeyDialog } from './revoke-api-key-dialog'
import type { ApiKeyListItem } from '../api/types'

interface ApiKeysTableProps {
  projectId: string
  rows: ApiKeyListItem[]
}

function formatTimestamp(iso: string | undefined): string {
  if (!iso) return '—'
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString()
}

function StatusBadge({ status }: { status: ApiKeyListItem['status'] }) {
  if (status === 'active') {
    return <Badge variant="secondary">Active</Badge>
  }
  return <Badge variant="destructive">Expired</Badge>
}

export function ApiKeysTable({ projectId, rows }: ApiKeysTableProps) {
  // One revoke dialog per table — the row button sets which key it
  // targets. Keeps the tree shallow vs. a dialog per row.
  const [revokeTarget, setRevokeTarget] = useState<ApiKeyListItem | null>(null)

  if (rows.length === 0) {
    return (
      <div className="rounded-lg border p-12 text-center">
        <p className="text-sm text-muted-foreground">No API keys yet</p>
        <p className="mt-1 text-xs text-muted-foreground">
          Generate an API key to let SDK clients ingest telemetry for this
          project.
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
              <TableHead>Name</TableHead>
              <TableHead>Preview</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Last used</TableHead>
              <TableHead>Created</TableHead>
              <TableHead>Expires</TableHead>
              <TableHead className="w-[1%] text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {rows.map((k) => (
              <TableRow key={k.id}>
                <TableCell className="font-medium">{k.name}</TableCell>
                <TableCell className="font-mono text-xs text-muted-foreground">
                  {k.key_preview}
                </TableCell>
                <TableCell>
                  <StatusBadge status={k.status} />
                </TableCell>
                <TableCell className="text-muted-foreground">
                  {formatTimestamp(k.last_used)}
                </TableCell>
                <TableCell className="text-muted-foreground">
                  {formatTimestamp(k.created_at)}
                </TableCell>
                <TableCell className="text-muted-foreground">
                  {formatTimestamp(k.expires_at)}
                </TableCell>
                <TableCell className="text-right">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setRevokeTarget(k)}
                  >
                    Revoke
                  </Button>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>

      {revokeTarget && (
        <RevokeApiKeyDialog
          projectId={projectId}
          keyId={revokeTarget.id}
          keyName={revokeTarget.name}
          open={!!revokeTarget}
          onOpenChange={(open) => {
            if (!open) setRevokeTarget(null)
          }}
        />
      )}
    </>
  )
}
