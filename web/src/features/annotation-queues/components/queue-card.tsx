'use client'

import Link from 'next/link'
import { formatDistanceToNow } from 'date-fns'
import {
  ClipboardList,
  MoreVertical,
  Trash2,
  Pencil,
  Plus,
  Pause,
  Play,
  Archive,
} from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { useAnnotationQueues } from '../context/annotation-queues-context'
import type { QueueWithStats, QueueStatus } from '../types'

interface QueueCardProps {
  data: QueueWithStats
}

function getStatusVariant(status: QueueStatus): 'default' | 'secondary' | 'outline' {
  switch (status) {
    case 'active':
      return 'default'
    case 'paused':
      return 'secondary'
    case 'archived':
      return 'outline'
  }
}

function getStatusIcon(status: QueueStatus) {
  switch (status) {
    case 'active':
      return Play
    case 'paused':
      return Pause
    case 'archived':
      return Archive
  }
}

export function QueueCard({ data }: QueueCardProps) {
  const { queue, stats } = data
  const { setOpen, setCurrentRow, projectSlug } = useAnnotationQueues()

  const handleEdit = () => {
    setCurrentRow(queue)
    setOpen('edit')
  }

  const handleDelete = () => {
    setCurrentRow(queue)
    setOpen('delete')
  }

  const handleAddItems = () => {
    setCurrentRow(queue)
    setOpen('add-items')
  }

  const completionPercentage =
    stats.total_items > 0
      ? Math.round(((stats.completed_items + stats.skipped_items) / stats.total_items) * 100)
      : 0

  const StatusIcon = getStatusIcon(queue.status)

  return (
    <Card>
      <CardHeader className="flex flex-row items-start justify-between pb-2">
        <div className="space-y-1">
          <CardTitle className="text-lg font-medium">
            <Link
              href={`/projects/${projectSlug}/annotation-queues/${queue.id}`}
              className="hover:underline"
            >
              {queue.name}
            </Link>
          </CardTitle>
          <Badge variant={getStatusVariant(queue.status)} className="capitalize">
            <StatusIcon className="mr-1 h-3 w-3" />
            {queue.status}
          </Badge>
        </div>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon">
              <MoreVertical className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={handleAddItems}>
              <Plus className="mr-2 h-4 w-4" />
              Add Items
            </DropdownMenuItem>
            <DropdownMenuItem onClick={handleEdit}>
              <Pencil className="mr-2 h-4 w-4" />
              Edit
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              className="text-destructive focus:text-destructive"
              onClick={handleDelete}
            >
              <Trash2 className="mr-2 h-4 w-4" />
              Delete
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </CardHeader>
      <CardContent className="space-y-4">
        <p className="text-sm text-muted-foreground line-clamp-2">
          {queue.description || 'No description'}
        </p>

        {/* Progress Bar */}
        <div className="space-y-1">
          <div className="flex items-center justify-between text-sm">
            <span className="text-muted-foreground">Progress</span>
            <span className="font-medium">{completionPercentage}%</span>
          </div>
          <Progress value={completionPercentage} className="h-2" />
        </div>

        {/* Stats */}
        <div className="grid grid-cols-4 gap-2 text-center text-xs">
          <div>
            <div className="font-medium">{stats.pending_items}</div>
            <div className="text-muted-foreground">Pending</div>
          </div>
          <div>
            <div className="font-medium">{stats.in_progress_items}</div>
            <div className="text-muted-foreground">In Progress</div>
          </div>
          <div>
            <div className="font-medium">{stats.completed_items}</div>
            <div className="text-muted-foreground">Completed</div>
          </div>
          <div>
            <div className="font-medium">{stats.skipped_items}</div>
            <div className="text-muted-foreground">Skipped</div>
          </div>
        </div>

        {/* Footer */}
        <div className="flex items-center justify-between text-sm pt-2 border-t">
          <span className="flex items-center gap-1 text-muted-foreground">
            <ClipboardList className="h-4 w-4" />
            {stats.total_items} items
          </span>
          <span className="text-muted-foreground">
            {formatDistanceToNow(new Date(queue.created_at), { addSuffix: true })}
          </span>
        </div>
      </CardContent>
    </Card>
  )
}
