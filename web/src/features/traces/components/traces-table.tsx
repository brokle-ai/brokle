'use client'

import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  flexRender,
  getCoreRowModel,
  useReactTable,
} from '@tanstack/react-table'
import { useTracesTableState } from '../hooks/use-traces-table-state'
import { useColumnVisibility } from '../hooks/use-column-visibility'
import { useTraceDetailState } from '../hooks/use-trace-detail-state'
import { useTraces } from '../context/traces-context'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { type Trace } from '../data/schema'
import { DataTablePagination } from '@/components/data-table'
import { DataTableBulkActions } from './data-table-bulk-actions'
import { DataTableToolbar } from './data-table-toolbar'
import { tracesColumns as columns } from './traces-columns'
import { TraceDetailContainer } from './trace-detail-container'

type DataTableProps = {
  data: Trace[]
  totalCount?: number
  isFetching?: boolean  // For subtle indicators (pagination spinner)
}

export function TracesTable({ data, totalCount, isFetching }: DataTableProps) {
  const { setCurrentPageTraceIds } = useTraces()
  const { openTrace } = useTraceDetailState()

  // Centralized URL state management
  const tableState = useTracesTableState()

  // localStorage-based column visibility
  const { visibility, setVisibility } = useColumnVisibility()

  // Local UI-only state
  const [rowSelection, setRowSelection] = useState({})

  // Initialize React Table with manual control
  const table = useReactTable({
    data,
    columns,
    pageCount: totalCount ? Math.ceil(totalCount / tableState.pageSize) : -1,
    state: {
      columnVisibility: visibility,
      rowSelection,
      pagination: {
        pageIndex: tableState.page - 1,
        pageSize: tableState.pageSize,
      },
      sorting: tableState.sortBy
        ? [{ id: tableState.sortBy, desc: tableState.sortOrder === 'desc' }]
        : [],
      globalFilter: tableState.search || '',
    },
    manualPagination: true,
    manualFiltering: true,
    manualSorting: true,
    enableRowSelection: true,
    getRowId: (row) => row.trace_id,
    onRowSelectionChange: setRowSelection,
    onColumnVisibilityChange: setVisibility,
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

  // Derive trace IDs from data for prev/next navigation
  const rowIds = useMemo(
    () => data.map((trace) => trace.trace_id),
    [data]
  )

  // Sync current page trace IDs to context
  useEffect(() => {
    setCurrentPageTraceIds(rowIds)
  }, [rowIds, setCurrentPageTraceIds])

  // Row click handler for peek
  const handleRowClick = useCallback(
    (trace: Trace, e: React.MouseEvent) => {
      // Ignore if click target is interactive element
      if ((e.target as HTMLElement).closest('[role="checkbox"], button, a')) {
        return
      }
      openTrace(trace.trace_id)
    },
    [openTrace]
  )

  return (
    <>
      <div className='space-y-4 max-sm:has-[div[role="toolbar"]]:mb-16'>
        <DataTableToolbar
          table={table}
          tableState={tableState}
        />
        <div className='overflow-hidden rounded-md border'>
          <Table>
            <TableHeader>
              {table.getHeaderGroups().map((headerGroup) => (
                <TableRow key={headerGroup.id}>
                  {headerGroup.headers.map((header) => {
                    return (
                      <TableHead key={header.id} colSpan={header.colSpan}>
                        {header.isPlaceholder
                          ? null
                          : flexRender(header.column.columnDef.header, header.getContext())}
                      </TableHead>
                    )
                  })}
                </TableRow>
              ))}
            </TableHeader>
            <TableBody>
              {table.getRowModel().rows?.length ? (
                table.getRowModel().rows.map((row) => (
                  <TableRow
                    key={row.id}
                    data-state={row.getIsSelected() && 'selected'}
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
                  <TableCell colSpan={columns.length} className='h-24 text-center'>
                    No traces found.
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>
        <DataTablePagination table={table} isPending={isFetching} />
        <DataTableBulkActions table={table} />
      </div>
      <TraceDetailContainer />
    </>
  )
}
