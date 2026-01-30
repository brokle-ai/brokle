'use client'

import { useState } from 'react'
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
import { Copy, ChevronRight, ChevronDown, User, Clock, Coins, Hash, AlertTriangle, ExternalLink } from 'lucide-react'
import { useRouter } from 'next/navigation'
import { type Session, useSessionsTableState } from '../hooks/use-project-sessions'
import { DataTableColumnHeader } from '@/features/traces/components/data-table-column-header'
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
import { formatDuration, formatCost } from '@/features/traces/utils/format-helpers'
import { toast } from 'sonner'
import { useProjectOnly } from '@/features/projects'

interface SessionsTableProps {
  data: Session[]
  totalCount: number
  isFetching: boolean
}

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
    accessorKey: 'sessionId',
    header: ({ column }) => <DataTableColumnHeader column={column} title='Session ID' />,
    cell: ({ row }) => {
      const sessionId = row.getValue('sessionId') as string
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
    accessorKey: 'traceCount',
    header: ({ column }) => <DataTableColumnHeader column={column} title='Traces' />,
    cell: ({ row }) => {
      const count = row.getValue('traceCount') as number
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
          {formatSessionDuration(session.firstTrace, session.lastTrace)}
        </div>
      )
    },
    enableSorting: false,
  },
  {
    accessorKey: 'totalTokens',
    header: ({ column }) => <DataTableColumnHeader column={column} title='Tokens' />,
    cell: ({ row }) => {
      const tokens = row.getValue('totalTokens') as number
      if (!tokens) return <span className='text-muted-foreground'>-</span>
      return <div className='font-mono text-sm'>{tokens.toLocaleString()}</div>
    },
    enableSorting: true,
  },
  {
    accessorKey: 'totalCost',
    header: ({ column }) => <DataTableColumnHeader column={column} title='Cost' />,
    cell: ({ row }) => {
      const cost = row.getValue('totalCost') as number
      return <div className='font-mono text-sm'>{formatCost(cost)}</div>
    },
    enableSorting: true,
  },
  {
    accessorKey: 'userIds',
    header: ({ column }) => <DataTableColumnHeader column={column} title='Users' />,
    cell: ({ row }) => {
      const userIds = row.getValue('userIds') as string[]
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
    accessorKey: 'lastTrace',
    header: ({ column }) => <DataTableColumnHeader column={column} title='Last Activity' />,
    cell: ({ row }) => {
      const lastTrace = row.getValue('lastTrace') as Date

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
 * Expanded row showing aggregate session stats with link to traces view
 */
function SessionTracesSubRow({ row }: { row: Row<Session> }) {
  const router = useRouter()
  const { currentProject } = useProjectOnly()
  const session = row.original

  const handleViewInTraces = () => {
    if (currentProject?.slug && session.sessionId !== 'no-session') {
      // Navigate to traces page with session filter
      router.push(`/projects/${currentProject.slug}/traces?sessionId=${session.sessionId}`)
    }
  }

  return (
    <div className='bg-muted/30 p-4 space-y-4'>
      {/* Session Summary Header */}
      <div className='flex items-center justify-between'>
        <span className='text-sm font-medium'>
          Session Summary
        </span>
        {session.sessionId !== 'no-session' && (
          <Button
            variant='outline'
            size='sm'
            onClick={handleViewInTraces}
            className='gap-2'
          >
            <ExternalLink className='h-3.5 w-3.5' />
            View Traces
          </Button>
        )}
      </div>

      {/* Aggregate Stats Grid */}
      <div className='grid grid-cols-2 md:grid-cols-4 gap-4'>
        {/* Trace Count */}
        <div className='rounded-md border bg-background p-3'>
          <div className='flex items-center gap-2 text-muted-foreground text-xs mb-1'>
            <Hash className='h-3.5 w-3.5' />
            Traces
          </div>
          <div className='font-mono text-lg font-semibold'>
            {session.traceCount.toLocaleString()}
          </div>
        </div>

        {/* Duration */}
        <div className='rounded-md border bg-background p-3'>
          <div className='flex items-center gap-2 text-muted-foreground text-xs mb-1'>
            <Clock className='h-3.5 w-3.5' />
            Duration
          </div>
          <div className='font-mono text-lg font-semibold'>
            {formatSessionDuration(session.firstTrace, session.lastTrace)}
          </div>
        </div>

        {/* Tokens */}
        <div className='rounded-md border bg-background p-3'>
          <div className='flex items-center gap-2 text-muted-foreground text-xs mb-1'>
            <Hash className='h-3.5 w-3.5' />
            Total Tokens
          </div>
          <div className='font-mono text-lg font-semibold'>
            {session.totalTokens ? session.totalTokens.toLocaleString() : '-'}
          </div>
        </div>

        {/* Cost */}
        <div className='rounded-md border bg-background p-3'>
          <div className='flex items-center gap-2 text-muted-foreground text-xs mb-1'>
            <Coins className='h-3.5 w-3.5' />
            Total Cost
          </div>
          <div className='font-mono text-lg font-semibold'>
            {formatCost(session.totalCost)}
          </div>
        </div>
      </div>

      {/* Additional Info Row */}
      <div className='flex flex-wrap gap-4 text-sm text-muted-foreground'>
        {/* Error Count */}
        {session.errorCount > 0 && (
          <div className='flex items-center gap-1.5'>
            <AlertTriangle className='h-4 w-4 text-destructive' />
            <span className='text-destructive font-medium'>{session.errorCount} error{session.errorCount !== 1 ? 's' : ''}</span>
          </div>
        )}

        {/* Users */}
        {session.userIds.length > 0 && (
          <div className='flex items-center gap-1.5'>
            <User className='h-4 w-4' />
            <span>
              {session.userIds.length} user{session.userIds.length !== 1 ? 's' : ''}
              {session.userIds.length <= 3 && (
                <span className='ml-1 text-xs'>
                  ({session.userIds.join(', ')})
                </span>
              )}
            </span>
          </div>
        )}

        {/* Time Range */}
        <div className='flex items-center gap-1.5'>
          <Clock className='h-4 w-4' />
          <span>
            {formatDistanceToNow(session.firstTrace, { addSuffix: true })} — {formatDistanceToNow(session.lastTrace, { addSuffix: true })}
          </span>
        </div>
      </div>
    </div>
  )
}

export function SessionsTable({ data, totalCount, isFetching }: SessionsTableProps) {
  const tableState = useSessionsTableState()
  const [expanded, setExpanded] = useState<ExpandedState>({})

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
    getRowId: (row) => row.sessionId,
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
