'use client'

import { type ColumnDef } from '@tanstack/react-table'
import { Checkbox } from '@/components/ui/checkbox'
import { Badge } from '@/components/ui/badge'
import { DataTableColumnHeader } from './data-table-column-header'
import { DataTableRowActions } from './data-table-row-actions'
import { statuses, statusCodeToString } from '../data/constants'
import type { Trace } from '../data/schema'
import { formatDistanceToNow } from 'date-fns'
import { Copy } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { toast } from 'sonner'

// Helper to format duration
function formatDuration(ms: number | undefined): string {
  if (!ms) return '-'
  if (ms < 1000) return `${ms}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

// Helper to format cost
function formatCost(cost: number | undefined): string {
  if (!cost) return '-'
  return `$${cost.toFixed(4)}`
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
    accessorKey: 'trace_id', // Updated from 'id'
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
    enableSorting: true,
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
    enableSorting: true,
  },
  {
    accessorKey: 'duration_ms',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Duration' />
    ),
    cell: ({ row }) => {
      const duration = row.getValue('duration_ms') as number | undefined
      return <div className='font-mono text-sm'>{formatDuration(duration)}</div>
    },
    enableSorting: true,
  },
  {
    accessorKey: 'status_code', // Updated from 'status'
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
    accessorKey: 'cost',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Cost' />
    ),
    cell: ({ row }) => {
      const cost = row.getValue('cost') as number | undefined
      return <div className='font-mono text-sm'>{formatCost(cost)}</div>
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
