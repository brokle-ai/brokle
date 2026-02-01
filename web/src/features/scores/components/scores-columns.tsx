'use client'

import Link from 'next/link'
import type { ColumnDef } from '@tanstack/react-table'
import { ExternalLink } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'
import { SortableColumnHeader } from '@/components/shared/tables'
import type { Score } from '../types'
import type { ScoreSortField } from '../hooks/use-scores-table-state'
import { ScoreTag } from './score-tag'
import { getDataTypeIndicator, getSourceIndicator } from '../lib/score-colors'

interface CreateScoresColumnsOptions {
  projectSlug: string
  sortBy: ScoreSortField
  sortOrder: 'asc' | 'desc'
  onSortChange: (
    field: ScoreSortField | null,
    order: 'asc' | 'desc' | null
  ) => void
  onDeleteScore?: (scoreId: string) => void
}

export function createScoresColumns({
  projectSlug,
  sortBy,
  sortOrder,
  onSortChange,
  onDeleteScore,
}: CreateScoresColumnsOptions): ColumnDef<Score>[] {
  return [
    {
      accessorKey: 'name',
      header: () => (
        <SortableColumnHeader
          label="Score"
          field="name"
          currentSort={sortBy}
          currentOrder={sortOrder}
          onSort={onSortChange}
        />
      ),
      cell: ({ row }) => (
        <ScoreTag score={row.original} onDelete={onDeleteScore} compact={false} />
      ),
    },
    {
      accessorKey: 'type',
      header: () => (
        <SortableColumnHeader
          label="Type"
          field="type"
          currentSort={sortBy}
          currentOrder={sortOrder}
          onSort={onSortChange}
        />
      ),
      cell: ({ row }) => {
        const { symbol, label } = getDataTypeIndicator(row.original.type)
        return (
          <div className="flex items-center gap-1.5">
            <span className="text-muted-foreground font-mono">{symbol}</span>
            <span className="text-sm text-muted-foreground">{label}</span>
          </div>
        )
      },
    },
    {
      accessorKey: 'source',
      header: () => (
        <SortableColumnHeader
          label="Source"
          field="source"
          currentSort={sortBy}
          currentOrder={sortOrder}
          onSort={onSortChange}
        />
      ),
      cell: ({ row }) => {
        const { label, className } = getSourceIndicator(row.original.source)
        return (
          <Badge variant="outline" className={cn('text-xs', className)}>
            {label}
          </Badge>
        )
      },
    },
    {
      accessorKey: 'trace_id',
      header: 'Trace',
      cell: ({ row }) => {
        const traceId = row.original.trace_id
        if (!traceId) {
          return <span className="text-muted-foreground">-</span>
        }
        return (
          <Link
            href={`/projects/${projectSlug}/traces/${traceId}`}
            className="inline-flex items-center gap-1 text-sm text-blue-600 hover:text-blue-700 hover:underline dark:text-blue-400"
            onClick={(e) => e.stopPropagation()}
          >
            <span className="font-mono truncate max-w-[100px]">
              {traceId.slice(0, 8)}...
            </span>
            <ExternalLink className="h-3 w-3 flex-shrink-0" />
          </Link>
        )
      },
    },
    {
      accessorKey: 'span_id',
      header: 'Span',
      cell: ({ row }) =>
        row.original.span_id ? (
          <span className="font-mono text-sm text-muted-foreground truncate max-w-[80px]">
            {row.original.span_id.slice(0, 8)}...
          </span>
        ) : (
          <span className="text-muted-foreground">-</span>
        ),
    },
    {
      accessorKey: 'timestamp',
      header: () => (
        <SortableColumnHeader
          label="Created"
          field="timestamp"
          currentSort={sortBy}
          currentOrder={sortOrder}
          onSort={onSortChange}
        />
      ),
      cell: ({ row }) => (
        <span className="text-sm text-muted-foreground whitespace-nowrap">
          {formatDistanceToNow(new Date(row.original.timestamp), {
            addSuffix: true,
          })}
        </span>
      ),
    },
  ]
}
