import { useMemo, useState, type ReactNode } from 'react'
import {
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  useReactTable,
  type ColumnFiltersState,
  type VisibilityState,
  type RowSelectionState,
  type SortingState,
} from '@tanstack/react-table'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { DataTablePagination } from '@/components/shared/tables/data-table-pagination'
import type { TraceListItem, Pagination } from '../api/types'
import type { TracesFilterValue } from './traces-filter-bar'
import { buildTracesColumns } from './traces-columns'
import { TracesToolbar } from './data-table-toolbar'
import { TracesBulkActions } from './data-table-bulk-actions'

export type TracesSortKey =
  | 'duration'
  | 'model_name'
  | 'total_cost'
  | 'total_tokens'
  | 'span_count'
  | 'start_time'

export interface TracesTableServerControl {
  /** Server pagination metadata (from the list response). */
  pagination: Pagination
  /** Page-size choices to surface in the pagination footer. */
  pageSizes?: number[]
  /** URL-backed filter bar state. */
  filterValue: TracesFilterValue
  onFilterChange: (next: TracesFilterValue) => void
  /** Pagination change → parent updates URL. */
  onPaginationChange: (page: number, pageSize: number) => void
  /** Sort change → parent updates URL. */
  sortBy?: TracesSortKey | null
  sortDir?: 'asc' | 'desc' | null
  onSortChange?: (
    sortBy: TracesSortKey | null,
    sortDir: 'asc' | 'desc' | null,
  ) => void
  /** Subtle spinner during loader refetch. */
  isFetching?: boolean
  /** Row-level action callbacks. Parent owns navigation + mutations. */
  onViewDetail?: (trace: TraceListItem) => void
  onAddToDataset?: (trace: TraceListItem) => void
  onDelete?: (trace: TraceListItem) => void
  /** Whole-row click (opens peek / detail). */
  onRowClick?: (trace: TraceListItem) => void
}

export interface TracesTableProps {
  rows: TraceListItem[]
  /**
   * Link renderer for the name column — parent owns routing so the
   * table stays framework-agnostic. Shared by both embedded (sessions
   * detail) and full (traces index) modes.
   */
  renderNameLink?: (trace: TraceListItem, children: ReactNode) => ReactNode
  /**
   * When provided, renders the full TanStack Table experience: toolbar
   * with filters + faceted filters + column-visibility menu, pagination
   * footer, and bulk-actions strip. When omitted (embedded mode, e.g.
   * the sessions detail page), the table renders rows-only.
   */
  server?: TracesTableServerControl
  /**
   * Project scope for bulk-mutation actions (e.g. add-to-queue). Only
   * read inside the bulk-actions branch — the embedded sessions-detail
   * caller has no bulk surface and can omit it.
   */
  projectId?: string
}

/**
 * The sort-key → server column name mapping is intentional. The Go
 * backend exposes a fixed enum of sortable columns (see
 * `internal/transport/http/handlers/observability/dashboard_types.go`:
 * `ListTracesInput.SortBy`). The faceted filters remain purely
 * client-side — they're narrowing the current page of results — while
 * pagination + sort round-trip to the server.
 */
export function TracesTable({
  rows,
  renderNameLink,
  server,
  projectId,
}: TracesTableProps) {
  const [rowSelection, setRowSelection] = useState<RowSelectionState>({})
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([])
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({})

  const columns = useMemo(
    () =>
      buildTracesColumns({
        renderNameLink,
        onViewDetail: server?.onViewDetail,
        onAddToDataset: server?.onAddToDataset,
        onDelete: server?.onDelete,
      }),
    [
      renderNameLink,
      server?.onViewDetail,
      server?.onAddToDataset,
      server?.onDelete,
    ],
  )

  const sorting: SortingState = useMemo(() => {
    if (!server?.sortBy) return []
    return [{ id: server.sortBy, desc: server.sortDir === 'desc' }]
  }, [server?.sortBy, server?.sortDir])

  const table = useReactTable({
    data: rows,
    columns,
    pageCount: server
      ? Math.max(1, server.pagination.total_pages)
      : undefined,
    state: {
      rowSelection,
      columnFilters,
      columnVisibility,
      sorting,
      pagination: server
        ? {
            pageIndex: server.pagination.page - 1,
            pageSize: server.pagination.limit,
          }
        : undefined,
    },
    manualPagination: !!server,
    manualSorting: !!server,
    enableRowSelection: true,
    getRowId: (row) => row.trace_id,
    onRowSelectionChange: setRowSelection,
    onColumnFiltersChange: setColumnFilters,
    onColumnVisibilityChange: setColumnVisibility,
    onPaginationChange: server
      ? (updater) => {
          const current = {
            pageIndex: server.pagination.page - 1,
            pageSize: server.pagination.limit,
          }
          const next = typeof updater === 'function' ? updater(current) : updater
          server.onPaginationChange(next.pageIndex + 1, next.pageSize)
        }
      : undefined,
    onSortingChange: server?.onSortChange
      ? (updater) => {
          const current: SortingState = server.sortBy
            ? [{ id: server.sortBy, desc: server.sortDir === 'desc' }]
            : []
          const next = typeof updater === 'function' ? updater(current) : updater
          if (next.length > 0) {
            server.onSortChange?.(
              next[0].id as TracesSortKey,
              next[0].desc ? 'desc' : 'asc',
            )
          } else {
            server.onSortChange?.(null, null)
          }
        }
      : undefined,
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
  })

  const { modelFacets, providerFacets } = useMemo(() => {
    const models = new Set<string>()
    const providers = new Set<string>()
    for (const row of rows) {
      if (row.model_name) models.add(row.model_name)
      if (row.provider_name) providers.add(row.provider_name)
    }
    return {
      modelFacets: Array.from(models).sort(),
      providerFacets: Array.from(providers).sort(),
    }
  }, [rows])

  // Embedded mode (sessions detail): rows-only render with the name
  // link renderer, no toolbar / pagination / bulk actions.
  if (!server) {
    return (
      <div className="overflow-hidden rounded-md border">
        <TableShell table={table} columnCount={columns.length} />
      </div>
    )
  }

  return (
    <div className='space-y-4 max-sm:has-[div[role="toolbar"]]:mb-16'>
      <TracesToolbar
        table={table}
        filterValue={server.filterValue}
        onFilterChange={server.onFilterChange}
        modelFacets={modelFacets}
        providerFacets={providerFacets}
      />
      <div className="overflow-hidden rounded-md border">
        <TableShell
          table={table}
          columnCount={columns.length}
          onRowClick={server.onRowClick}
        />
      </div>
      <DataTablePagination
        table={table}
        isPending={!!server.isFetching}
        showSelectedRows
        serverPagination={{
          page: server.pagination.page,
          pageSize: server.pagination.limit,
          total: server.pagination.total,
          totalPages: Math.max(1, server.pagination.total_pages),
          hasNextPage: server.pagination.has_next,
          hasPreviousPage: server.pagination.has_prev,
        }}
        pageSizes={server.pageSizes}
      />
      {projectId ? <TracesBulkActions table={table} projectId={projectId} /> : null}
    </div>
  )
}

function TableShell({
  table,
  columnCount,
  onRowClick,
}: {
  table: ReturnType<typeof useReactTable<TraceListItem>>
  columnCount: number
  onRowClick?: (trace: TraceListItem) => void
}) {
  const rowModel = table.getRowModel()
  return (
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
        {rowModel.rows.length > 0 ? (
          rowModel.rows.map((row) => (
            <TableRow
              key={row.id}
              data-state={row.getIsSelected() && 'selected'}
              className={onRowClick ? 'cursor-pointer hover:bg-muted/50' : undefined}
              onClick={
                onRowClick
                  ? (e) => {
                      const target = e.target as HTMLElement
                      if (target.closest('[role="checkbox"], button, a')) return
                      onRowClick(row.original)
                    }
                  : undefined
              }
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
            <TableCell colSpan={columnCount} className="h-24 text-center">
              No traces found.
            </TableCell>
          </TableRow>
        )}
      </TableBody>
    </Table>
  )
}
