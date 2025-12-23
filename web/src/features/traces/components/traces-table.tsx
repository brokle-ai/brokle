'use client'

import { useCallback, useEffect, useMemo, useState, useTransition } from 'react'
import { useSearchParams } from 'next/navigation'
import type { ColumnFiltersState, SortingState, VisibilityState } from '@tanstack/react-table'
import {
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
} from '@tanstack/react-table'
import { useTableSearchParams } from '@/hooks/use-table-search-params'
import { useTableNavigation } from '../hooks/use-table-navigation'
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
  isFetching?: boolean
}

export function TracesTable({ data, totalCount, isFetching }: DataTableProps) {
  const searchParams = useSearchParams()
  const [isPending, startTransition] = useTransition()
  const { setCurrentPageTraceIds } = useTraces()
  const { openTrace } = useTraceDetailState()

  // Parse URL state (single source of truth)
  const {
    page,
    pageSize,
    filter,
    status,
    sortBy,
    sortOrder,
  } = useTableSearchParams(searchParams)

  // Local UI-only state (not in URL)
  const [rowSelection, setRowSelection] = useState({})
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({})

  // Convert parsed params to React Table format
  const pagination = { pageIndex: page - 1, pageSize }
  const columnFilters: ColumnFiltersState = [
    ...(status.length > 0 ? [{ id: 'status', value: status }] : []),
  ]
  const sorting: SortingState =
    sortBy && sortOrder ? [{ id: sortBy, desc: sortOrder === 'desc' }] : []
  const globalFilter = filter

  // Get navigation handlers
  const { handlePageChange, handleSearch, handleFilter, handleSort, handleReset } =
    useTableNavigation({
      searchParams,
      onSearchChange: () => setRowSelection({}),
    })

  // Wrap navigation handlers with startTransition
  const onPaginationChange = useCallback(
    (paginationUpdater: any) => {
      const newPagination =
        typeof paginationUpdater === 'function'
          ? paginationUpdater(pagination)
          : paginationUpdater
      startTransition(() => {
        handlePageChange(newPagination)
      })
    },
    [pagination, handlePageChange]
  )

  const onGlobalFilterChange = useCallback(
    (filterValue: string) => {
      startTransition(() => {
        handleSearch(filterValue)
      })
    },
    [handleSearch]
  )

  const onColumnFiltersChange = useCallback(
    (filterUpdater: any) => {
      const newFilters =
        typeof filterUpdater === 'function' ? filterUpdater(columnFilters) : filterUpdater
      startTransition(() => {
        handleFilter(newFilters)
      })
    },
    [columnFilters, handleFilter]
  )

  const onSortingChange = useCallback(
    (sortingUpdater: any) => {
      const newSorting =
        typeof sortingUpdater === 'function' ? sortingUpdater(sorting) : sortingUpdater
      startTransition(() => {
        handleSort(newSorting)
      })
    },
    [sorting, handleSort]
  )

  // Initialize React Table
  const table = useReactTable({
    data,
    columns,
    pageCount: totalCount ? Math.ceil(totalCount / pageSize) : -1,
    state: {
      sorting,
      columnVisibility,
      rowSelection,
      columnFilters,
      globalFilter,
      pagination,
    },
    manualPagination: true,
    manualFiltering: true,
    manualSorting: true,
    enableRowSelection: true,
    getRowId: (row) => row.trace_id,
    onRowSelectionChange: setRowSelection,
    onColumnVisibilityChange: setColumnVisibility,
    onPaginationChange,
    onGlobalFilterChange,
    onColumnFiltersChange,
    onSortingChange,
    globalFilterFn: (row, _columnId, filterValue) => {
      const traceId = String(row.getValue('trace_id')).toLowerCase()
      const name = String(row.getValue('name')).toLowerCase()
      const searchValue = String(filterValue).toLowerCase()
      return traceId.includes(searchValue) || name.includes(searchValue)
    },
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
  })

  // Optimized: Derive trace IDs from data (only changes when data changes)
  const rowIds = useMemo(
    () => data.map((trace) => trace.trace_id),
    [data]
  )

  // Sync current page trace IDs to context (for prev/next navigation)
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
      <DataTableToolbar table={table} isPending={isPending || isFetching} onReset={handleReset} />
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
      <DataTablePagination table={table} isPending={isPending || isFetching} />
      <DataTableBulkActions table={table} />
    </div>
    <TraceDetailContainer />
    </>
  )
}
