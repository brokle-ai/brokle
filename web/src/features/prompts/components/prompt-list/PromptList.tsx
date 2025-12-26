'use client'

import { useCallback, useMemo, useState, useTransition } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import {
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
  type ColumnFiltersState,
  type SortingState,
  type VisibilityState,
} from '@tanstack/react-table'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { DataTablePagination } from '@/components/data-table'
import { useTableSearchParams } from '@/hooks/use-table-search-params'
import { usePromptsTableNavigation } from '../../hooks/use-prompts-table-navigation'
import { PromptsToolbar } from './prompts-toolbar'
import type { PromptListItem } from '../../types'
import { createPromptsColumns } from './prompts-columns'
import { PromptsDeleteDialog } from './prompts-delete-dialog'

interface PromptsTableProps {
  data: PromptListItem[]
  totalCount: number
  isFetching?: boolean
  protectedLabels?: string[]
  projectSlug: string
  orgSlug: string
}

export function PromptsTable({
  data,
  totalCount,
  isFetching,
  protectedLabels = [],
  projectSlug,
  orgSlug,
}: PromptsTableProps) {
  const router = useRouter()
  const searchParams = useSearchParams()
  const [isPending, startTransition] = useTransition()

  // Parse URL state (single source of truth)
  const {
    page,
    pageSize,
    filter,
    type: typeFilter,
    sortBy,
    sortOrder,
  } = useTableSearchParams(searchParams)

  // Local UI-only state (not in URL)
  const [rowSelection, setRowSelection] = useState({})
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({})
  const [deletePrompt, setDeletePrompt] = useState<PromptListItem | null>(null)

  // Convert parsed params to React Table format
  const pagination = { pageIndex: page - 1, pageSize }
  const columnFilters: ColumnFiltersState = [
    ...(typeFilter.length > 0 ? [{ id: 'type', value: typeFilter }] : []),
  ]
  const sorting: SortingState =
    sortBy && sortOrder ? [{ id: sortBy, desc: sortOrder === 'desc' }] : []
  const globalFilter = filter

  // Get navigation handlers
  const { handlePageChange, handleSearch, handleFilter, handleSort, handleReset } =
    usePromptsTableNavigation({
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

  // Create columns with actions
  const columns = useMemo(
    () =>
      createPromptsColumns({
        protectedLabels,
        onEdit: (prompt) => {
          router.push(`/projects/${projectSlug}/prompts/${prompt.id}`)
        },
        onDelete: (prompt) => {
          setDeletePrompt(prompt)
        },
        onPlayground: (prompt) => {
          router.push(`/projects/${projectSlug}/prompts/${prompt.id}/playground`)
        },
        onViewHistory: (prompt) => {
          router.push(`/projects/${projectSlug}/prompts/${prompt.id}/versions`)
        },
      }),
    [projectSlug, protectedLabels, router]
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
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
  })

  // Row click handler
  const handleRowClick = useCallback(
    (prompt: PromptListItem, e: React.MouseEvent) => {
      // Ignore if click target is interactive element
      if ((e.target as HTMLElement).closest('[role="checkbox"], button, a')) {
        return
      }
      router.push(`/projects/${projectSlug}/prompts/${prompt.id}`)
    },
    [router, projectSlug]
  )

  return (
    <div className="space-y-4">
      <PromptsToolbar table={table} isPending={isPending || isFetching} onReset={handleReset} />
      <div className="overflow-hidden rounded-md border">
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
                  data-state={row.getIsSelected() && 'selected'}
                  className="cursor-pointer hover:bg-muted/50"
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
                <TableCell colSpan={columns.length} className="h-24 text-center">
                  No prompts found.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
      <DataTablePagination table={table} isPending={isPending || isFetching} />

      {/* Delete Dialog */}
      <PromptsDeleteDialog
        prompt={deletePrompt}
        open={!!deletePrompt}
        onOpenChange={(open) => !open && setDeletePrompt(null)}
      />
    </div>
  )
}
