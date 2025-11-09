'use client'

import { useCallback, useState, useTransition } from 'react'
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
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { type Task } from '../data/schema'
import { DataTableBulkActions } from './data-table-bulk-actions'
import { DataTablePagination } from './data-table-pagination'
import { DataTableToolbar } from './data-table-toolbar'
import { tasksColumns as columns } from './tasks-columns'

type DataTableProps = {
  data: Task[]
  totalCount?: number
}

export function TasksTable({ data, totalCount }: DataTableProps) {
  const searchParams = useSearchParams()
  const [isPending, startTransition] = useTransition()

  // Parse URL state (single source of truth)
  const {
    page,
    pageSize,
    filter,
    status,
    priority,
    sortBy,
    sortOrder,
  } = useTableSearchParams(searchParams)

  // Local UI-only state (not in URL)
  const [rowSelection, setRowSelection] = useState({})
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({})

  // Convert parsed params to React Table format
  const pagination = { pageIndex: page - 1, pageSize } // Convert to 0-indexed
  const columnFilters: ColumnFiltersState = [
    ...(status.length > 0 ? [{ id: 'status', value: status }] : []),
    ...(priority.length > 0 ? [{ id: 'priority', value: priority }] : []),
  ]
  const sorting: SortingState =
    sortBy && sortOrder ? [{ id: sortBy, desc: sortOrder === 'desc' }] : []
  const globalFilter = filter

  // Get navigation handlers
  const { handlePageChange, handleSearch, handleFilter, handleSort, handleReset } =
    useTableNavigation({
      searchParams,
      onSearchChange: () => setRowSelection({}), // Clear selection on filter/search
    })

  // Wrap navigation handlers with startTransition for smooth UX
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
    getRowId: (row) => row.id,
    onRowSelectionChange: setRowSelection,
    onColumnVisibilityChange: setColumnVisibility,
    onPaginationChange,
    onGlobalFilterChange,
    onColumnFiltersChange,
    onSortingChange,
    globalFilterFn: (row, _columnId, filterValue) => {
      const id = String(row.getValue('id')).toLowerCase()
      const title = String(row.getValue('title')).toLowerCase()
      const searchValue = String(filterValue).toLowerCase()
      return id.includes(searchValue) || title.includes(searchValue)
    },
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
  })

  return (
    <div className='space-y-4 max-sm:has-[div[role="toolbar"]]:mb-16'>
      <DataTableToolbar table={table} isPending={isPending} onReset={handleReset} />
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
                <TableRow key={row.id} data-state={row.getIsSelected() && 'selected'}>
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
                  No results.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
      <DataTablePagination table={table} isPending={isPending} />
      <DataTableBulkActions table={table} />
    </div>
  )
}
