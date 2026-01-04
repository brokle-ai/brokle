'use client'

import Link from 'next/link'
import { formatDistanceToNow } from 'date-fns'
import { LayoutDashboard, MoreVertical, Trash2, Pencil } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { useDashboards } from '../context/dashboards-context'
import type { Dashboard } from '../types'

interface DashboardCardProps {
  dashboard: Dashboard
}

export function DashboardCard({ dashboard }: DashboardCardProps) {
  const { setOpen, setCurrentRow, projectSlug } = useDashboards()

  const handleEdit = () => {
    setCurrentRow(dashboard)
    setOpen('edit')
  }

  const handleDelete = () => {
    setCurrentRow(dashboard)
    setOpen('delete')
  }

  const widgetCount = dashboard.config?.widgets?.length ?? 0

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-lg font-medium">
          <Link
            href={`/projects/${projectSlug}/dashboards/${dashboard.id}`}
            className="hover:underline"
          >
            {dashboard.name}
          </Link>
        </CardTitle>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon">
              <MoreVertical className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
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
      <CardContent>
        <p className="text-sm text-muted-foreground mb-4 line-clamp-2">
          {dashboard.description || 'No description'}
        </p>
        <div className="flex items-center justify-between text-sm">
          <span className="flex items-center gap-1 text-muted-foreground">
            <LayoutDashboard className="h-4 w-4" />
            {widgetCount} widget{widgetCount !== 1 ? 's' : ''}
          </span>
          <span className="text-muted-foreground">
            {formatDistanceToNow(new Date(dashboard.created_at), { addSuffix: true })}
          </span>
        </div>
      </CardContent>
    </Card>
  )
}
