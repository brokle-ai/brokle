'use client'

import { formatDistanceToNow } from 'date-fns'
import Link from 'next/link'
import {
  Clock,
  CheckCircle,
  XCircle,
  Lock,
  MoreVertical,
  ExternalLink,
  Play,
} from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import type { QueueItem, ItemStatus, ObjectType } from '../types'

interface ItemCardProps {
  item: QueueItem
  projectSlug: string
  onAnnotate?: (item: QueueItem) => void
  onViewObject?: (item: QueueItem) => void
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

function getObjectTypeBadge(type: ObjectType) {
  return (
    <Badge variant="outline" className="uppercase text-xs">
      {type}
    </Badge>
  )
}

function getObjectLink(objectId: string, objectType: ObjectType, projectSlug: string) {
  if (objectType === 'trace') {
    return `/projects/${projectSlug}/traces/${objectId}`
  }
  // For SPAN, we would need the trace ID - for now link to traces with filter
  return `/projects/${projectSlug}/traces?spanId=${objectId}`
}

export function ItemCard({ item, projectSlug, onAnnotate, onViewObject }: ItemCardProps) {
  const isLocked = !!item.locked_at && !!item.locked_by_user_id
  const isPending = item.status === 'pending'

  return (
    <Card>
      <CardHeader className="flex flex-row items-start justify-between pb-2">
        <div className="space-y-1">
          <CardTitle className="text-sm font-mono">
            {item.object_id.substring(0, 16)}...
          </CardTitle>
          <div className="flex items-center gap-2">
            {getObjectTypeBadge(item.object_type)}
            {getStatusBadge(item.status)}
          </div>
        </div>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon" className="h-8 w-8">
              <MoreVertical className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            {isPending && onAnnotate && (
              <DropdownMenuItem onClick={() => onAnnotate(item)}>
                <Play className="mr-2 h-4 w-4" />
                Annotate
              </DropdownMenuItem>
            )}
            <DropdownMenuItem asChild>
              <Link href={getObjectLink(item.object_id, item.object_type, projectSlug)}>
                <ExternalLink className="mr-2 h-4 w-4" />
                View {item.object_type.toLowerCase()}
              </Link>
            </DropdownMenuItem>
            {onViewObject && (
              <DropdownMenuItem onClick={() => onViewObject(item)}>
                <ExternalLink className="mr-2 h-4 w-4" />
                View Details
              </DropdownMenuItem>
            )}
          </DropdownMenuContent>
        </DropdownMenu>
      </CardHeader>
      <CardContent className="space-y-3">
        {/* Priority */}
        <div className="flex items-center justify-between text-sm">
          <span className="text-muted-foreground">Priority</span>
          <span className="font-medium">{item.priority}</span>
        </div>

        {/* Lock Status */}
        {isLocked && (
          <div className="flex items-center gap-1 text-sm text-muted-foreground">
            <Lock className="h-3 w-3" />
            <span>
              Locked {formatDistanceToNow(new Date(item.locked_at!), { addSuffix: true })}
            </span>
          </div>
        )}

        {/* Completion Info */}
        {item.completed_at && (
          <div className="flex items-center gap-1 text-sm text-muted-foreground">
            <CheckCircle className="h-3 w-3 text-green-600" />
            <span>
              Completed {formatDistanceToNow(new Date(item.completed_at), { addSuffix: true })}
            </span>
          </div>
        )}

        {/* Footer */}
        <div className="flex items-center justify-between text-xs text-muted-foreground pt-2 border-t">
          <span>
            Created {formatDistanceToNow(new Date(item.created_at), { addSuffix: true })}
          </span>
        </div>
      </CardContent>
    </Card>
  )
}
