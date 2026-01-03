'use client'

import { useCallback, useEffect, useMemo, useState, useTransition } from 'react'
import { useRouter } from 'next/navigation'
import {
  flexRender,
  getCoreRowModel,
  useReactTable,
  type ColumnFiltersState,
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
import { DataTablePagination } from '@/components/data-table'
import { usePromptsTableState } from '../../hooks/use-prompts-table-state'
import { PromptsToolbar } from './prompts-toolbar'
import type { PromptListItem, PromptType } from '../../types'
import { createPromptsColumns } from './prompts-columns'
import { PromptsDeleteDialog } from './prompts-delete-dialog'

interface PromptsTableProps {
  data: PromptListItem[]
  totalCount: number
  isFetching?: boolean
  protectedLabels?: string[]
  projectSlug: string
}

export function PromptsTable({
  data,
  totalCount,
  isFetching,
  protectedLabels = [],
  projectSlug,
}: PromptsTableProps) {
  const router = useRouter()
  const [isPending, startTransition] = useTransition()

  // URL state via nuqs (single source of truth)
  const tableState = usePromptsTableState()
  const { page, pageSize, search, types, sortBy, sortOrder, setTypes, setPagination, setSorting, resetAll } = tableState

  // Local UI-only state (not in URL)
  const [rowSelection, setRowSelection] = useState({})
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({})
  const [deletePrompt, setDeletePrompt] = useState<PromptListItem | null>(null)

  // Clear row selection when search changes (prevents stale selections after filtering)
  useEffect(() => {
    setRowSelection({})
  }, [search])

  // Convert URL state to React Table format (memoized to stabilize references)
  const pagination = useMemo(
    (): PaginationState => ({ pageIndex: page - 1, pageSize }),
    [page, pageSize]
  )
  const columnFilters = useMemo(
    (): ColumnFiltersState => [
      ...(types.length > 0 ? [{ id: 'type', value: types }] : []),
    ],
    [types]
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


  const onColumnFiltersChange = useCallback(
    (filterUpdater: Updater<ColumnFiltersState>) => {
      const newFilters: ColumnFiltersState =
        typeof filterUpdater === 'function' ? filterUpdater(columnFilters) : filterUpdater
      startTransition(() => {
        // Extract type filter values (multi-select)
        const typeFilterValue = newFilters.find((f) => f.id === 'type')?.value as PromptType[] | undefined
        setTypes(typeFilterValue || [])
        setRowSelection({})
      })
    },
    [columnFilters, setTypes]
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
          setSorting(sort.id, sort.desc ? 'desc' : 'asc')
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
    onColumnFiltersChange,
    onSortingChange,
    getCoreRowModel: getCoreRowModel(),
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
      <PromptsToolbar table={table} tableState={tableState} onReset={handleReset} />
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
