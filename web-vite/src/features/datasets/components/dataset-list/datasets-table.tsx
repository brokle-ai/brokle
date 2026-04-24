import { useCallback, useEffect, useMemo, useState, useTransition } from 'react'
import { useNavigate, useParams } from '@tanstack/react-router'
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
import { useDatasetsTableState, type SortField } from '../../hooks/use-datasets-table-state'
import { DatasetsToolbar } from './datasets-toolbar'
import { createDatasetsColumns } from './datasets-columns'
import type { DatasetWithItemCount } from '../../types'

interface DatasetsTableProps {
  data: DatasetWithItemCount[]
  totalCount: number
  isLoading?: boolean
  isFetching?: boolean
  projectSlug: string
  onDelete?: (dataset: DatasetWithItemCount) => void
}

export function DatasetsTable({
  data,
  totalCount,
  isLoading = false,
  isFetching,
  projectSlug: _projectSlug,
  onDelete,
}: DatasetsTableProps) {
  const navigate = useNavigate()
  const routeParams = useParams({ strict: false }) as {
    orgId?: string
    projectId?: string
  }
  const orgId = routeParams.orgId ?? ''
  const projectId = routeParams.projectId ?? ''
  const [isPending, startTransition] = useTransition()

  // URL state via nuqs (single source of truth)
  const tableState = useDatasetsTableState()
  const { page, pageSize, search, sortBy, sortOrder, setPagination, setSorting, resetAll } = tableState

  // Local UI-only state (not in URL)
  const [rowSelection, setRowSelection] = useState({})
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({})

  // Clear row selection when search changes (prevents stale selections after filtering)
  useEffect(() => {
    setRowSelection({})
  }, [search])

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
        const [sort] = newSorting
        if (!sort) {
          setSorting(null, null)
        } else {
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

  // Create columns with actions
  const columns = useMemo(
    () =>
      createDatasetsColumns({
        onEdit: (dataset) => {
          navigate({
            to: '/o/$orgId/p/$projectId/datasets/$datasetId',
            params: { orgId, projectId, datasetId: dataset.id },
          })
        },
        onDelete: onDelete,
        onRunExperiment: (dataset) => {
          // experiments/new search schema doesn't model datasetId — use
          // a raw href to avoid TanStack's typed-search mismatch.
          window.location.href = `/o/${orgId}/p/${projectId}/experiments/new?datasetId=${dataset.id}`
        },
        onViewVersions: (dataset) => {
          navigate({
            to: '/o/$orgId/p/$projectId/datasets/$datasetId/versions',
            params: { orgId, projectId, datasetId: dataset.id },
          })
        },
      }),
    [orgId, projectId, navigate, onDelete],
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
    (dataset: DatasetWithItemCount, e: React.MouseEvent) => {
      // Ignore if click target is interactive element
      if ((e.target as HTMLElement).closest('[role="checkbox"], button, a')) {
        return
      }
      navigate({
        to: '/o/$orgId/p/$projectId/datasets/$datasetId',
        params: { orgId, projectId, datasetId: dataset.id },
      })
    },
    [navigate, orgId, projectId],
  )

  // Loading state
  if (isLoading) {
    return <DataTableSkeleton columns={6} rows={5} toolbarSlots={2} />
  }

  return (
    <div className="space-y-4">
      <DatasetsToolbar table={table} tableState={tableState} onReset={handleReset} />
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
                  No datasets found.
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
