'use client'

import { type ColumnDef } from '@tanstack/react-table'
import { Checkbox } from '@/components/ui/checkbox'
import { Badge } from '@/components/ui/badge'
import { DataTableColumnHeader } from './data-table-column-header'
import { DataTableRowActions } from './data-table-row-actions'
import { statuses, statusCodeToString } from '../data/constants'
import type { Trace } from '../data/schema'
import { formatDuration, formatCost } from '../utils/format-helpers'
import { formatDistanceToNow } from 'date-fns'
import { Copy, Star } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { toast } from 'sonner'
import {
  HoverCard,
  HoverCardContent,
  HoverCardTrigger,
} from '@/components/ui/hover-card'
import { ScoreTagList, type Score as ScoresFeatureScore } from '@/features/scores'
import type { Score } from '../data/schema'

// Convert traces Score to scores feature Score for ScoreTagList
function toScoresFeatureScore(score: Score): ScoresFeatureScore {
  return {
    id: score.id,
    project_id: score.project_id,
    trace_id: score.trace_id,
    span_id: score.span_id,
    name: score.name,
    value: score.value,
    string_value: score.string_value,
    type: score.type as ScoresFeatureScore['type'],
    source: mapSource(score.source),
    reason: score.comment,
    metadata: score.evaluator_config as Record<string, unknown>,
    timestamp: score.timestamp instanceof Date
      ? score.timestamp.toISOString()
      : String(score.timestamp),
  }
}

function mapSource(source: string): ScoresFeatureScore['source'] {
  const sourceMap: Record<string, ScoresFeatureScore['source']> = {
    API: 'code',
    api: 'code',
    code: 'code',
    EVAL: 'llm',
    eval: 'llm',
    llm: 'llm',
    ANNOTATION: 'human',
    annotation: 'human',
    human: 'human',
  }
  return sourceMap[source] || 'code'
}

export const tracesColumns: ColumnDef<Trace>[] = [
  {
    id: 'select',
    header: ({ table }) => (
      <Checkbox
        checked={
          table.getIsAllPageRowsSelected() ||
          (table.getIsSomePageRowsSelected() && 'indeterminate')
        }
        onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
        aria-label='Select all'
        className='translate-y-0.5'
      />
    ),
    cell: ({ row }) => (
      <Checkbox
        checked={row.getIsSelected()}
        onCheckedChange={(value) => row.toggleSelected(!!value)}
        aria-label='Select row'
        className='translate-y-0.5'
      />
    ),
    enableSorting: false,
    enableHiding: false,
  },
  {
    accessorKey: 'trace_id',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Trace ID' />
    ),
    cell: ({ row }) => {
      const traceId = row.getValue('trace_id') as string
      const shortId = traceId.substring(0, 8)

      const handleCopy = (e: React.MouseEvent) => {
        e.stopPropagation()
        navigator.clipboard.writeText(traceId)
        toast.success('Trace ID copied to clipboard')
      }

      return (
        <div className='flex items-center gap-1'>
          <span className='font-mono text-xs'>{shortId}</span>
          <Button
            variant='ghost'
            size='icon'
            className='h-5 w-5'
            onClick={handleCopy}
          >
            <Copy className='h-3 w-3' />
          </Button>
        </div>
      )
    },
    enableSorting: false, // Backend doesn't support trace_id sorting
    enableHiding: false,
  },
  {
    accessorKey: 'name',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Name' />
    ),
    cell: ({ row }) => {
      return <div className='font-medium'>{row.getValue('name')}</div>
    },
    enableSorting: false, // Backend doesn't support name sorting
  },
  {
    accessorKey: 'duration',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Duration' />
    ),
    cell: ({ row }) => {
      const duration = row.getValue('duration') as number | undefined
      return <div className='font-mono text-sm'>{formatDuration(duration)}</div>
    },
    enableSorting: true,
  },
  {
    accessorKey: 'status_code',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Status' />
    ),
    cell: ({ row }) => {
      const statusCode = row.getValue('status_code') as number
      const statusStr = statusCodeToString(statusCode) // Convert UInt8 to string
      const status = statuses.find((s) => s.value === statusStr)

      if (!status) return null

      const StatusIcon = status.icon

      return (
        <div className='flex items-center gap-2'>
          <StatusIcon className='h-4 w-4 text-muted-foreground' />
          <span>{status.label}</span>
        </div>
      )
    },
    filterFn: (row, id, value) => {
      const statusCode = row.getValue(id) as number
      const statusStr = statusCodeToString(statusCode)
      return value.includes(statusStr)
    },
    enableSorting: false,
  },
  {
    accessorKey: 'model_name',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Model' />
    ),
    cell: ({ row }) => {
      const model = row.getValue('model_name') as string | undefined
      if (!model) return <span className='text-muted-foreground'>-</span>
      return (
        <Badge variant='secondary' className='font-mono text-xs'>
          {model}
        </Badge>
      )
    },
    filterFn: (row, id, value) => {
      const model = row.getValue(id) as string | undefined
      if (!model) return false
      return value.includes(model)
    },
    enableSorting: true,
  },
  {
    accessorKey: 'provider_name',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Provider' />
    ),
    cell: ({ row }) => {
      const provider = row.getValue('provider_name') as string | undefined
      if (!provider) return <span className='text-muted-foreground'>-</span>
      return (
        <span className='text-sm capitalize'>{provider}</span>
      )
    },
    filterFn: (row, id, value) => {
      const provider = row.getValue(id) as string | undefined
      if (!provider) return false
      return value.includes(provider)
    },
    enableSorting: false, // Backend doesn't support provider_name sorting
  },
  {
    accessorKey: 'cost',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Cost' />
    ),
    cell: ({ row }) => {
      const cost = row.getValue('cost') as number | string | undefined
      return <div className='font-mono text-sm'>{formatCost(cost)}</div>
    },
    enableSorting: true,
  },
  {
    accessorKey: 'tokens',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Tokens' />
    ),
    cell: ({ row }) => {
      const tokens = row.getValue('tokens') as number | undefined
      if (!tokens) return <span className='text-muted-foreground'>-</span>
      return (
        <div className='font-mono text-sm'>
          {tokens.toLocaleString()}
        </div>
      )
    },
    enableSorting: true,
  },
  {
    accessorKey: 'spanCount',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Spans' />
    ),
    cell: ({ row }) => {
      const count = row.getValue('spanCount') as number
      return (
        <Badge variant='outline' className='font-mono'>
          {count}
        </Badge>
      )
    },
    enableSorting: true,
  },
  {
    accessorKey: 'scores',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Scores' />
    ),
    cell: ({ row }) => {
      const scores = row.original.scores
      const count = scores?.length ?? 0

      if (count === 0) {
        return <span className='text-muted-foreground text-sm'>—</span>
      }

      const convertedScores = scores!.map(toScoresFeatureScore)

      return (
        <HoverCard openDelay={300}>
          <HoverCardTrigger asChild>
            <Badge
              variant='secondary'
              className='font-mono cursor-pointer hover:bg-secondary/80 gap-1'
            >
              <Star className='h-3 w-3' />
              {count}
            </Badge>
          </HoverCardTrigger>
          <HoverCardContent className='w-auto max-w-xs' align='start'>
            <div className='space-y-2'>
              <div className='text-xs font-medium text-muted-foreground'>
                Scores ({count})
              </div>
              <ScoreTagList scores={convertedScores} maxVisible={5} />
            </div>
          </HoverCardContent>
        </HoverCard>
      )
    },
    enableSorting: false,
  },
  {
    accessorKey: 'start_time',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Timestamp' />
    ),
    cell: ({ row }) => {
      const startTime = row.getValue('start_time') as Date

      // Safety check for invalid dates
      if (!startTime || !(startTime instanceof Date) || isNaN(startTime.getTime())) {
        return <div className='text-sm text-muted-foreground'>-</div>
      }

      return (
        <div className='text-sm text-muted-foreground'>
          {formatDistanceToNow(startTime, { addSuffix: true })}
        </div>
      )
    },
    enableSorting: true,
  },
  {
    id: 'actions',
    cell: ({ row }) => <DataTableRowActions row={row} />,
    enableSorting: false,
    enableHiding: false,
  },
]
