'use client'

import { formatDistanceToNow } from 'date-fns'
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
import { ExternalLink, Lock, CheckCircle, XCircle, Clock } from 'lucide-react'
import type { QueueItem, ItemStatus } from '../types'

interface QueueItemTableProps {
  items: QueueItem[]
  onViewItem?: (item: QueueItem) => void
}

function getStatusBadge(status: ItemStatus) {
  switch (status) {
    case 'pending':
      return (
        <Badge variant="outline">
          <Clock className="mr-1 h-3 w-3" />
          Pending
        </Badge>
      )
    case 'completed':
      return (
        <Badge variant="default" className="bg-green-600">
          <CheckCircle className="mr-1 h-3 w-3" />
          Completed
        </Badge>
      )
    case 'skipped':
      return (
        <Badge variant="secondary">
          <XCircle className="mr-1 h-3 w-3" />
          Skipped
        </Badge>
      )
  }
}

export function QueueItemTable({ items, onViewItem }: QueueItemTableProps) {
  if (items.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-center">
        <p className="text-muted-foreground">No items in this queue yet.</p>
      </div>
    )
  }

  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Object ID</TableHead>
            <TableHead>Type</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Priority</TableHead>
            <TableHead>Locked</TableHead>
            <TableHead>Created</TableHead>
            <TableHead className="w-[100px]">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {items.map((item) => (
            <TableRow key={item.id}>
              <TableCell className="font-mono text-sm">
                {item.object_id.substring(0, 16)}...
              </TableCell>
              <TableCell>
                <Badge variant="outline">{item.object_type}</Badge>
              </TableCell>
              <TableCell>{getStatusBadge(item.status)}</TableCell>
              <TableCell>{item.priority}</TableCell>
              <TableCell>
                {item.locked_at ? (
                  <span className="flex items-center text-sm text-muted-foreground">
                    <Lock className="mr-1 h-3 w-3" />
                    {formatDistanceToNow(new Date(item.locked_at), { addSuffix: true })}
                  </span>
                ) : (
                  <span className="text-muted-foreground">-</span>
                )}
              </TableCell>
              <TableCell className="text-sm text-muted-foreground">
                {formatDistanceToNow(new Date(item.created_at), { addSuffix: true })}
              </TableCell>
              <TableCell>
                {onViewItem && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => onViewItem(item)}
                  >
                    <ExternalLink className="h-4 w-4" />
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
