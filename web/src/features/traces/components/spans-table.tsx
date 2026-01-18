'use client'

import { useCallback, useMemo } from 'react'
import { type ColumnDef, flexRender, getCoreRowModel, useReactTable } from '@tanstack/react-table'
import { formatDistanceToNow } from 'date-fns'
import { Copy, ExternalLink, Layers } from 'lucide-react'
import { useProjectSpans } from '../hooks/use-project-spans'
import { useTraceDetailState } from '../hooks/use-trace-detail-state'
import { ItemBadge } from './item-badge'
import { DataTableColumnHeader } from './data-table-column-header'
import { DataTablePagination } from '@/components/data-table'
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
import { DataTableEmptyState } from '@/components/data-table'
import { LoadingSpinner } from '@/components/guards/loading-spinner'
import { statuses, statusCodeToString } from '../data/constants'
import { formatDuration, formatCost } from '../utils/format-helpers'
import { toast } from 'sonner'
import type { Span } from '../data/schema'

/**
 * Columns for flat spans table
 *
 * Based on competitive analysis:
 * - Opik: span_name, trace_id, status, duration, model, tokens, cost, timestamp
 * - Phoenix: span_kind (color-coded), name, duration, model, tokens, cost
 */
const spansColumns: ColumnDef<Span>[] = [
  {
    accessorKey: 'span_name',
    header: ({ column }) => <DataTableColumnHeader column={column} title='Name' />,
    cell: ({ row }) => {
      const name = row.getValue('span_name') as string
      return <div className='font-medium max-w-[200px] truncate'>{name}</div>
    },
    enableSorting: true,
  },
  {
    accessorKey: 'span_type',
    header: ({ column }) => <DataTableColumnHeader column={column} title='Type' />,
    cell: ({ row }) => {
      const spanType = row.original.span_type
      return <ItemBadge spanType={spanType} showLabel />
    },
    enableSorting: false,
  },
  {
    accessorKey: 'trace_id',
    header: ({ column }) => <DataTableColumnHeader column={column} title='Trace ID' />,
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
          <span className='font-mono text-xs text-muted-foreground'>{shortId}</span>
          <Button variant='ghost' size='icon' className='h-5 w-5' onClick={handleCopy}>
            <Copy className='h-3 w-3' />
          </Button>
        </div>
      )
    },
    enableSorting: false,
  },
  {
    accessorKey: 'status_code',
    header: ({ column }) => <DataTableColumnHeader column={column} title='Status' />,
    cell: ({ row }) => {
      const statusCode = row.getValue('status_code') as number
      const statusStr = statusCodeToString(statusCode)
      const status = statuses.find((s) => s.value === statusStr)

      if (!status) return null

      const StatusIcon = status.icon

      return (
        <div className='flex items-center gap-2'>
          <StatusIcon className='h-4 w-4 text-muted-foreground' />
          <span className='text-sm'>{status.label}</span>
        </div>
      )
    },
    enableSorting: false,
  },
  {
    accessorKey: 'duration',
    header: ({ column }) => <DataTableColumnHeader column={column} title='Duration' />,
    cell: ({ row }) => {
      const duration = row.getValue('duration') as number | undefined
      // Duration color thresholds (Phoenix pattern)
      const getDurationColor = (ns?: number) => {
        if (!ns) return ''
        const ms = ns / 1_000_000
        if (ms < 500) return 'text-green-600 dark:text-green-400'
        if (ms < 2000) return 'text-yellow-600 dark:text-yellow-400'
        return 'text-red-600 dark:text-red-400'
      }
      return (
        <div className={`font-mono text-sm ${getDurationColor(duration)}`}>
          {formatDuration(duration)}
        </div>
      )
    },
    enableSorting: true,
  },
  {
    accessorKey: 'model_name',
    header: ({ column }) => <DataTableColumnHeader column={column} title='Model' />,
    cell: ({ row }) => {
      const model = row.original.model_name || row.original.gen_ai_request_model
      if (!model) return <span className='text-muted-foreground'>-</span>
      return (
        <Badge variant='secondary' className='font-mono text-xs'>
          {model}
        </Badge>
      )
    },
    enableSorting: true,
  },
  {
    id: 'tokens',
    header: ({ column }) => <DataTableColumnHeader column={column} title='Tokens' />,
    cell: ({ row }) => {
      const inputTokens = row.original.gen_ai_usage_input_tokens
      const outputTokens = row.original.gen_ai_usage_output_tokens
      const total = (inputTokens || 0) + (outputTokens || 0)

      if (!total) return <span className='text-muted-foreground'>-</span>

      return (
        <div className='font-mono text-sm'>
          {total.toLocaleString()}
        </div>
      )
    },
    enableSorting: false,
  },
  {
    accessorKey: 'total_cost',
    header: ({ column }) => <DataTableColumnHeader column={column} title='Cost' />,
    cell: ({ row }) => {
      const cost = row.getValue('total_cost') as number | undefined
      return <div className='font-mono text-sm'>{formatCost(cost)}</div>
    },
    enableSorting: true,
  },
  {
    accessorKey: 'start_time',
    header: ({ column }) => <DataTableColumnHeader column={column} title='Timestamp' />,
    cell: ({ row }) => {
      const startTime = row.getValue('start_time') as Date

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
]

export function SpansTable() {
  const { data, totalCount, isLoading, isFetching, error, hasProject, refetch, tableState } =
    useProjectSpans()
  const { openTrace } = useTraceDetailState()

  const isInitialLoad = isLoading && data.length === 0 && !tableState.hasActiveFilters
  const isEmptyProject = !isLoading && totalCount === 0 && !tableState.hasActiveFilters

  const table = useReactTable({
    data,
    columns: spansColumns,
    pageCount: totalCount ? Math.ceil(totalCount / tableState.pageSize) : -1,
    state: {
      pagination: {
        pageIndex: tableState.page - 1,
        pageSize: tableState.pageSize,
      },
      sorting: tableState.sortBy
        ? [{ id: tableState.sortBy, desc: tableState.sortOrder === 'desc' }]
        : [],
    },
    manualPagination: true,
    manualSorting: true,
    getRowId: (row) => row.span_id,
    onPaginationChange: (updater) => {
      const current = { pageIndex: tableState.page - 1, pageSize: tableState.pageSize }
      const next = typeof updater === 'function' ? updater(current) : updater
      tableState.setPagination(next.pageIndex + 1, next.pageSize)
    },
    onSortingChange: (updater) => {
      const current = tableState.sortBy
        ? [{ id: tableState.sortBy, desc: tableState.sortOrder === 'desc' }]
        : []
      const next = typeof updater === 'function' ? updater(current) : updater
      if (next.length > 0) {
        tableState.setSorting(next[0].id, next[0].desc ? 'desc' : 'asc')
      } else {
        tableState.setSorting(null, null)
      }
    },
    getCoreRowModel: getCoreRowModel(),
  })

  // Row click handler - opens the trace containing this span
  const handleRowClick = useCallback(
    (span: Span, e: React.MouseEvent) => {
      if ((e.target as HTMLElement).closest('button, a')) {
        return
      }
      // Open trace detail - will show the span in the trace view
      openTrace(span.trace_id)
    },
    [openTrace]
  )

  // Initial loading state
  if (isInitialLoad) {
    return (
      <div className='flex flex-1 items-center justify-center py-16'>
        <LoadingSpinner message='Loading spans...' />
      </div>
    )
  }

  // Error state
  if (error) {
    return (
      <div className='flex flex-col items-center justify-center py-12 space-y-4'>
        <div className='rounded-lg bg-destructive/10 p-6 text-center max-w-md'>
          <h3 className='font-semibold text-destructive mb-2'>Failed to load spans</h3>
          <p className='text-sm text-muted-foreground mb-4'>{error}</p>
          <button
            onClick={() => refetch()}
            className='inline-flex items-center justify-center rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors'
          >
            Try Again
          </button>
        </div>
      </div>
    )
  }

  // Empty state
  if (isEmptyProject) {
    return (
      <DataTableEmptyState
        icon={<Layers className='h-full w-full' />}
        title='No spans yet'
        description='Spans will appear here when traces contain nested operations.'
      />
    )
  }

  return (
    <div className='space-y-4'>
      <div className='overflow-hidden rounded-md border'>
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <TableHead key={header.id} colSpan={header.colSpan}>
                    {header.isPlaceholder
                      ? null
                      : flexRender(header.column.columnDef.header, header.getContext())}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  className='cursor-pointer hover:bg-muted/50'
                  onClick={(e) => handleRowClick(row.original, e)}
                >
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id}>
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell colSpan={spansColumns.length} className='h-24 text-center'>
                  No spans found.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
      <DataTablePagination table={table} isPending={isFetching} />
    </div>
  )
}
