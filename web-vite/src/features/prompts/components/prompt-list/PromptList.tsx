import { useCallback, useEffect, useMemo, useState, useTransition } from 'react'
import { useNavigate, useParams } from '@tanstack/react-router'
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
  projectId: string
}

export function PromptsTable({
  data,
  totalCount,
  isFetching,
  protectedLabels = [],
  projectId,
}: PromptsTableProps) {
  const navigate = useNavigate()
  const routeParams = useParams({ strict: false }) as {
    orgId?: string
    projectId?: string
  }
  const orgId = routeParams.orgId ?? ''
  const pProjectId = routeParams.projectId ?? ''
  const [isPending, startTransition] = useTransition()

  const tableState = usePromptsTableState()
  const {
    page,
    pageSize,
    search,
    types,
    sortBy,
    sortOrder,
    setTypes,
    setPagination,
    setSorting,
    resetAll,
  } = tableState

  const [rowSelection, setRowSelection] = useState({})
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({})
  const [deletePrompt, setDeletePrompt] = useState<PromptListItem | null>(null)

  useEffect(() => {
    setRowSelection({})
  }, [search])

  const pagination = useMemo(
    (): PaginationState => ({ pageIndex: page - 1, pageSize }),
    [page, pageSize],
  )
  const columnFilters = useMemo(
    (): ColumnFiltersState => [
      ...(types.length > 0 ? [{ id: 'type', value: types }] : []),
    ],
    [types],
  )
  const sorting = useMemo(
    (): SortingState =>
      sortBy && sortOrder ? [{ id: sortBy, desc: sortOrder === 'desc' }] : [],
    [sortBy, sortOrder],
  )
  const globalFilter = search || ''

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
    [pagination, setPagination],
  )

  const onColumnFiltersChange = useCallback(
    (filterUpdater: Updater<ColumnFiltersState>) => {
      const newFilters: ColumnFiltersState =
        typeof filterUpdater === 'function'
          ? filterUpdater(columnFilters)
          : filterUpdater
      startTransition(() => {
        const typeFilterValue = newFilters.find((f) => f.id === 'type')?.value as
          | PromptType[]
          | undefined
        setTypes(typeFilterValue || [])
        setRowSelection({})
      })
    },
    [columnFilters, setTypes],
  )

  const onSortingChange = useCallback(
    (sortingUpdater: Updater<SortingState>) => {
      const newSorting: SortingState =
        typeof sortingUpdater === 'function'
          ? sortingUpdater(sorting)
          : sortingUpdater
      startTransition(() => {
        if (newSorting.length === 0) {
          setSorting(null, null)
        } else {
          const [sort] = newSorting
          setSorting(sort.id, sort.desc ? 'desc' : 'asc')
        }
      })
    },
    [sorting, setSorting],
  )

  const handleReset = useCallback(() => {
    startTransition(() => {
      resetAll()
      setRowSelection({})
    })
  }, [resetAll])

  const goToDetail = useCallback(
    (promptId: string) => {
      void navigate({
        to: '/o/$orgId/p/$projectId/prompts/$promptId',
        params: { orgId, projectId: pProjectId, promptId },
      })
    },
    [navigate, orgId, pProjectId],
  )

  const goToEdit = useCallback(
    (promptId: string) => {
      void navigate({
        to: '/o/$orgId/p/$projectId/prompts/$promptId/edit',
        params: { orgId, projectId: pProjectId, promptId },
      })
    },
    [navigate, orgId, pProjectId],
  )

  const columns = useMemo(
    () =>
      createPromptsColumns({
        protectedLabels,
        onEdit: (prompt) => goToEdit(prompt.id),
        onDelete: (prompt) => setDeletePrompt(prompt),
        onViewHistory: (prompt) => goToDetail(prompt.id),
      }),
    [protectedLabels, goToDetail, goToEdit],
  )

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

  const handleRowClick = useCallback(
    (prompt: PromptListItem, e: React.MouseEvent) => {
      if ((e.target as HTMLElement).closest('[role="checkbox"], button, a')) {
        return
      }
      goToDetail(prompt.id)
    },
    [goToDetail],
  )

  return (
    <div className="space-y-4">
      <PromptsToolbar
        table={table}
        tableState={tableState}
        onReset={handleReset}
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
                      : flexRender(
                          header.column.columnDef.header,
                          header.getContext(),
                        )}
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
                      {flexRender(
                        cell.column.columnDef.cell,
                        cell.getContext(),
                      )}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell
                  colSpan={columns.length}
                  className="h-24 text-center"
                >
                  No prompts found.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
      <DataTablePagination
        table={table}
        isPending={isPending || isFetching}
      />

      <PromptsDeleteDialog
        projectId={projectId}
        prompt={deletePrompt}
        open={!!deletePrompt}
        onOpenChange={(open) => !open && setDeletePrompt(null)}
      />
    </div>
  )
}
