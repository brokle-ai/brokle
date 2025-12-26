'use client'

import Link from 'next/link'
import { formatDistanceToNow } from 'date-fns'
import { Database, MoreVertical, Trash2, Pencil } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { useDatasets } from '../context/datasets-context'
import type { Dataset } from '../types'

interface DatasetCardProps {
  dataset: Dataset
}

export function DatasetCard({ dataset }: DatasetCardProps) {
  const { setOpen, setCurrentRow, projectSlug } = useDatasets()

  const handleEdit = () => {
    setCurrentRow(dataset)
    setOpen('edit')
  }

  const handleDelete = () => {
    setCurrentRow(dataset)
    setOpen('delete')
  }

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-lg font-medium">
          <Link
            href={`/projects/${projectSlug}/datasets/${dataset.id}`}
            className="hover:underline"
          >
            {dataset.name}
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
          {dataset.description || 'No description'}
        </p>
        <div className="flex items-center justify-between text-sm">
          <span className="flex items-center gap-1 text-muted-foreground">
            <Database className="h-4 w-4" />
            Dataset
          </span>
          <span className="text-muted-foreground">
            {formatDistanceToNow(new Date(dataset.created_at), { addSuffix: true })}
          </span>
        </div>
      </CardContent>
    </Card>
  )
}
