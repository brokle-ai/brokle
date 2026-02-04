'use client'

import { useCallback, useEffect, useMemo, useState, useTransition } from 'react'
import { useRouter } from 'next/navigation'
import {
  flexRender,
  getCoreRowModel,
  useReactTable,
  type PaginationState,
  type SortingState,
  type Updater,
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
import { DataTablePagination, DataTableSkeleton } from '@/components/data-table'
import { useExperimentsTableState, type SortField } from '../../hooks/use-experiments-table-state'
import { ExperimentsToolbar } from './experiments-toolbar'
import { createExperimentsColumns } from './experiments-columns'
import type { Experiment } from '../../types'

interface ExperimentsTableProps {
  data: Experiment[]
  totalCount: number
  isLoading?: boolean
  isFetching?: boolean
  projectSlug: string
  onDelete?: (experiment: Experiment) => void
  onRerun?: (experiment: Experiment) => void
}

export function ExperimentsTable({
  data,
  totalCount,
  isLoading = false,
  isFetching,
  projectSlug,
  onDelete,
  onRerun,
}: ExperimentsTableProps) {
  const router = useRouter()
  const [isPending, startTransition] = useTransition()

  // URL state via nuqs (single source of truth)
  const tableState = useExperimentsTableState()
  const { page, pageSize, search, sortBy, sortOrder, setPagination, setSorting, resetAll } = tableState

  // Local UI-only state (not in URL)
  const [rowSelection, setRowSelection] = useState<Record<string, boolean>>({})
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({})

  // Clear row selection when search or filters change
  useEffect(() => {
    setRowSelection({})
  }, [search, tableState.status])

  // Convert URL state to React Table format (memoized to stabilize references)
  const pagination = useMemo(
    (): PaginationState => ({ pageIndex: page - 1, pageSize }),
    [page, pageSize]
  )
  const sorting = useMemo(
    (): SortingState =>
      sortBy && sortOrder ? [{ id: sortBy, desc: sortOrder === 'desc' }] : [],
    [sortBy, sortOrder]
  )
  const globalFilter = search || ''

  // Handlers using nuqs setters (wrapped in startTransition for smooth UX)
  const onPaginationChange = useCallback(
    (paginationUpdater: Updater<PaginationState>) => {
      const newPagination =
        typeof paginationUpdater === 'function'
          ? paginationUpdater(pagination)
          : paginationUpdater
      startTransition(() => {
        setPagination(newPagination.pageIndex + 1, newPagination.pageSize)
      })
    },
    [pagination, setPagination]
  )

  const onSortingChange = useCallback(
    (sortingUpdater: Updater<SortingState>) => {
      const newSorting: SortingState =
        typeof sortingUpdater === 'function' ? sortingUpdater(sorting) : sortingUpdater
      startTransition(() => {
        if (newSorting.length === 0) {
          setSorting(null, null)
        } else {
          const [sort] = newSorting
          setSorting(sort.id as SortField, sort.desc ? 'desc' : 'asc')
        }
      })
    },
    [sorting, setSorting]
  )

  const handleReset = useCallback(() => {
    startTransition(() => {
      resetAll()
      setRowSelection({})
    })
  }, [resetAll])

  // Navigate to compare view with selected experiments
  const handleCompare = useCallback(
    (selectedIds: string[]) => {
      const params = new URLSearchParams()
      selectedIds.forEach((id) => params.append('ids', id))
      router.push(`/projects/${projectSlug}/experiments/compare?${params.toString()}`)
    },
    [router, projectSlug]
  )

  // Create columns with actions
  const columns = useMemo(
    () =>
      createExperimentsColumns({
        onView: (experiment) => {
          router.push(`/projects/${projectSlug}/experiments/${experiment.id}`)
        },
        onEdit: (experiment) => {
          router.push(`/projects/${projectSlug}/experiments/${experiment.id}`)
        },
        onDelete,
        onRerun,
      }),
    [projectSlug, router, onDelete, onRerun]
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
    onSortingChange,
    getCoreRowModel: getCoreRowModel(),
  })

  // Row click handler
  const handleRowClick = useCallback(
    (experiment: Experiment, e: React.MouseEvent) => {
      // Ignore if click target is interactive element
      if ((e.target as HTMLElement).closest('[role="checkbox"], button, a')) {
        return
      }
      router.push(`/projects/${projectSlug}/experiments/${experiment.id}`)
    },
    [router, projectSlug]
  )

  // Count selected rows
  const selectedCount = Object.keys(rowSelection).filter((key) => rowSelection[key]).length

  // Loading state
  if (isLoading) {
    return <DataTableSkeleton columns={7} rows={5} toolbarSlots={2} />
  }

  return (
    <div className="space-y-4">
      <ExperimentsToolbar
        table={table}
        tableState={tableState}
        onReset={handleReset}
        onCompare={handleCompare}
        selectedCount={selectedCount}
      />
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
                  No experiments found.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
      <DataTablePagination table={table} isPending={isPending || isFetching} />
    </div>
  )
}
