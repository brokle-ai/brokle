'use client'

import { useCallback, useState } from 'react'
import {
  type ColumnDef,
  type ExpandedState,
  flexRender,
  getCoreRowModel,
  getExpandedRowModel,
  useReactTable,
  type Row,
} from '@tanstack/react-table'
import { formatDistanceToNow, differenceInMinutes } from 'date-fns'
import { Copy, ChevronRight, ChevronDown, MessageSquare, User } from 'lucide-react'
import { useProjectSessions, type Session } from '../hooks/use-project-sessions'
import { useTabState } from '../hooks/use-tab-state'
import { useTracesTableState } from '../hooks/use-traces-table-state'
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
import { formatDuration, formatCost } from '../utils/format-helpers'
import { toast } from 'sonner'

/**
 * Format duration between two dates as human-readable
 */
function formatSessionDuration(firstTrace: Date, lastTrace: Date): string {
  const minutes = differenceInMinutes(lastTrace, firstTrace)
  if (minutes < 1) return '< 1 min'
  if (minutes < 60) return `${minutes} min`
  const hours = Math.floor(minutes / 60)
  const remainingMinutes = minutes % 60
  if (hours < 24) return `${hours}h ${remainingMinutes}m`
  const days = Math.floor(hours / 24)
  return `${days}d ${hours % 24}h`
}

/**
 * Sessions table columns
 */
const sessionsColumns: ColumnDef<Session>[] = [
  {
    id: 'expander',
    header: () => null,
    cell: ({ row }) => {
      return (
        <Button
          variant='ghost'
          size='icon'
          className='h-6 w-6'
          onClick={(e) => {
            e.stopPropagation()
            row.toggleExpanded()
          }}
        >
          {row.getIsExpanded() ? (
            <ChevronDown className='h-4 w-4' />
          ) : (
            <ChevronRight className='h-4 w-4' />
          )}
        </Button>
      )
    },
    enableSorting: false,
  },
  {
    accessorKey: 'session_id',
    header: ({ column }) => <DataTableColumnHeader column={column} title='Session ID' />,
    cell: ({ row }) => {
      const sessionId = row.getValue('session_id') as string
      const isNoSession = sessionId === 'no-session'
      const displayId = isNoSession ? 'No Session' : sessionId.substring(0, 16)

      const handleCopy = (e: React.MouseEvent) => {
        e.stopPropagation()
        if (!isNoSession) {
          navigator.clipboard.writeText(sessionId)
          toast.success('Session ID copied to clipboard')
        }
      }

      return (
        <div className='flex items-center gap-2'>
          <span className={`font-mono text-sm ${isNoSession ? 'text-muted-foreground italic' : ''}`}>
            {displayId}
          </span>
          {!isNoSession && (
            <Button variant='ghost' size='icon' className='h-5 w-5' onClick={handleCopy}>
              <Copy className='h-3 w-3' />
            </Button>
          )}
        </div>
      )
    },
    enableSorting: false,
  },
  {
    accessorKey: 'trace_count',
    header: ({ column }) => <DataTableColumnHeader column={column} title='Traces' />,
    cell: ({ row }) => {
      const count = row.getValue('trace_count') as number
      return (
        <Badge variant='outline' className='font-mono'>
          {count}
        </Badge>
      )
    },
    enableSorting: true,
  },
  {
    id: 'duration',
    header: ({ column }) => <DataTableColumnHeader column={column} title='Duration' />,
    cell: ({ row }) => {
      const session = row.original
      return (
        <div className='font-mono text-sm'>
          {formatSessionDuration(session.first_trace, session.last_trace)}
        </div>
      )
    },
    enableSorting: false,
  },
  {
    accessorKey: 'total_tokens',
    header: ({ column }) => <DataTableColumnHeader column={column} title='Tokens' />,
    cell: ({ row }) => {
      const tokens = row.getValue('total_tokens') as number
      if (!tokens) return <span className='text-muted-foreground'>-</span>
      return <div className='font-mono text-sm'>{tokens.toLocaleString()}</div>
    },
    enableSorting: true,
  },
  {
    accessorKey: 'total_cost',
    header: ({ column }) => <DataTableColumnHeader column={column} title='Cost' />,
    cell: ({ row }) => {
      const cost = row.getValue('total_cost') as number
      return <div className='font-mono text-sm'>{formatCost(cost)}</div>
    },
    enableSorting: true,
  },
  {
    accessorKey: 'user_ids',
    header: ({ column }) => <DataTableColumnHeader column={column} title='Users' />,
    cell: ({ row }) => {
      const userIds = row.getValue('user_ids') as string[]
      if (!userIds.length) return <span className='text-muted-foreground'>-</span>

      return (
        <div className='flex items-center gap-1'>
          <User className='h-3.5 w-3.5 text-muted-foreground' />
          <span className='text-sm'>{userIds.length}</span>
        </div>
      )
    },
    enableSorting: false,
  },
  {
    accessorKey: 'last_trace',
    header: ({ column }) => <DataTableColumnHeader column={column} title='Last Activity' />,
    cell: ({ row }) => {
      const lastTrace = row.getValue('last_trace') as Date

      if (!lastTrace || !(lastTrace instanceof Date) || isNaN(lastTrace.getTime())) {
        return <div className='text-sm text-muted-foreground'>-</div>
      }

      return (
        <div className='text-sm text-muted-foreground'>
          {formatDistanceToNow(lastTrace, { addSuffix: true })}
        </div>
      )
    },
    enableSorting: true,
  },
]

/**
 * Expanded row showing traces in the session
 */
function SessionTracesSubRow({ row }: { row: Row<Session> }) {
  const traces = row.original.traces
  const { setTab } = useTabState()
  const tracesTableState = useTracesTableState()

  const handleViewInTraces = (sessionId: string) => {
    // Navigate to traces tab filtered by session
    tracesTableState.setFilters([
      {
        id: 'session-filter',
        column: 'session_id',
        operator: '=',
        value: sessionId,
      },
    ])
    setTab('traces')
  }

  return (
    <div className='bg-muted/30 p-4 space-y-3'>
      <div className='flex items-center justify-between'>
        <span className='text-sm font-medium'>
          {traces.length} trace{traces.length !== 1 ? 's' : ''} in this session
        </span>
        {row.original.session_id !== 'no-session' && (
          <Button
            variant='outline'
            size='sm'
            onClick={() => handleViewInTraces(row.original.session_id)}
          >
            View in Traces
          </Button>
        )}
      </div>
      <div className='grid gap-2'>
        {traces.slice(0, 5).map((trace) => (
          <div
            key={trace.trace_id}
            className='flex items-center justify-between rounded-md border bg-background p-2 text-sm'
          >
            <div className='flex items-center gap-3'>
              <span className='font-mono text-xs text-muted-foreground'>
                {trace.trace_id.substring(0, 8)}
              </span>
              <span className='font-medium'>{trace.name}</span>
            </div>
            <div className='flex items-center gap-4 text-muted-foreground'>
              <span className='font-mono text-xs'>{formatDuration(trace.duration)}</span>
              <span className='text-xs'>
                {formatDistanceToNow(trace.start_time, { addSuffix: true })}
              </span>
            </div>
          </div>
        ))}
        {traces.length > 5 && (
          <div className='text-sm text-muted-foreground text-center py-1'>
            ... and {traces.length - 5} more traces
          </div>
        )}
      </div>
    </div>
  )
}

export function SessionsTable() {
  const { data, totalCount, isLoading, isFetching, error, hasProject, refetch, tableState } =
    useProjectSessions()
  const [expanded, setExpanded] = useState<ExpandedState>({})

  const isInitialLoad = isLoading && data.length === 0 && !tableState.hasActiveFilters
  const isEmpty = !isLoading && totalCount === 0 && !tableState.hasActiveFilters

  const table = useReactTable({
    data,
    columns: sessionsColumns,
    pageCount: totalCount ? Math.ceil(totalCount / tableState.pageSize) : -1,
    state: {
      pagination: {
        pageIndex: tableState.page - 1,
        pageSize: tableState.pageSize,
      },
      sorting: tableState.sortBy
        ? [{ id: tableState.sortBy, desc: tableState.sortOrder === 'desc' }]
        : [],
      expanded,
    },
    manualPagination: true,
    manualSorting: true,
    getRowId: (row) => row.session_id,
    onExpandedChange: setExpanded,
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
    getExpandedRowModel: getExpandedRowModel(),
  })

  if (isInitialLoad) {
    return (
      <div className='flex flex-1 items-center justify-center py-16'>
        <LoadingSpinner message='Loading sessions...' />
      </div>
    )
  }

  if (error) {
    return (
      <div className='flex flex-col items-center justify-center py-12 space-y-4'>
        <div className='rounded-lg bg-destructive/10 p-6 text-center max-w-md'>
          <h3 className='font-semibold text-destructive mb-2'>Failed to load sessions</h3>
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

  if (isEmpty) {
    return (
      <DataTableEmptyState
        icon={<MessageSquare className='h-full w-full' />}
        title='No sessions yet'
        description='Sessions group multi-turn conversations. Add session_id to your traces to see them here.'
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
                <>
                  <TableRow
                    key={row.id}
                    className='cursor-pointer hover:bg-muted/50'
                    onClick={() => row.toggleExpanded()}
                  >
                    {row.getVisibleCells().map((cell) => (
                      <TableCell key={cell.id}>
                        {flexRender(cell.column.columnDef.cell, cell.getContext())}
                      </TableCell>
                    ))}
                  </TableRow>
                  {row.getIsExpanded() && (
                    <TableRow key={`${row.id}-expanded`}>
                      <TableCell colSpan={sessionsColumns.length} className='p-0'>
                        <SessionTracesSubRow row={row} />
                      </TableCell>
                    </TableRow>
                  )}
                </>
              ))
            ) : (
              <TableRow>
                <TableCell colSpan={sessionsColumns.length} className='h-24 text-center'>
                  No sessions found.
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
