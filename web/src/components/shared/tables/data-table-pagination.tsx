'use client'

import {
  ChevronLeftIcon,
  ChevronRightIcon,
  DoubleArrowLeftIcon,
  DoubleArrowRightIcon,
} from '@radix-ui/react-icons'
import type { Table } from '@tanstack/react-table'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

/**
 * Server-side pagination metadata
 * Use when the table uses manualPagination: true
 */
interface PaginationMeta {
  page: number
  pageSize: number
  total: number
  totalPages: number
  hasNextPage: boolean
  hasPreviousPage: boolean
}

interface DataTablePaginationProps<TData> {
  /** TanStack Table instance */
  table: Table<TData>
  /** Available page sizes for the dropdown */
  pageSizes?: number[]
  /** Show row selection count (defaults to false) */
  showSelectedRows?: boolean
  /** Server-side pagination metadata (for manualPagination tables) */
  serverPagination?: PaginationMeta | null
  /** Show loading state on pagination controls */
  isPending?: boolean
  /** Total count label (e.g., "10 total items") - only shown when not showing selected rows */
  totalLabel?: string
}

/**
 * Unified pagination component for data tables.
 *
 * Supports both client-side and server-side (manual) pagination:
 * - Client-side: Uses TanStack Table's built-in pagination
 * - Server-side: Uses serverPagination prop for page/total info
 *
 * For server-side pagination with manual page control (using nuqs/URL state),
 * set up the table with onPaginationChange callback that syncs to URL state.
 *
 * @example
 * ```tsx
 * // Client-side pagination
 * <DataTablePagination table={table} />
 *
 * // Server-side with URL state
 * <DataTablePagination
 *   table={table}
 *   isPending={isPending}
 * />
 *
 * // With server pagination meta and total label
 * <DataTablePagination
 *   table={table}
 *   serverPagination={paginationMeta}
 *   totalLabel={`${total} evaluators`}
 * />
 * ```
 */
export function DataTablePagination<TData>({
  table,
  pageSizes = [10, 20, 30, 40, 50],
  showSelectedRows = false,
  serverPagination,
  isPending = false,
  totalLabel,
}: DataTablePaginationProps<TData>) {
  // Determine pagination values from server or client
  const currentPage = serverPagination
    ? serverPagination.page
    : table.getState().pagination.pageIndex + 1
  const totalPages = serverPagination
    ? serverPagination.totalPages
    : table.getPageCount()
  const canPrevious = serverPagination
    ? serverPagination.hasPreviousPage
    : table.getCanPreviousPage()
  const canNext = serverPagination
    ? serverPagination.hasNextPage
    : table.getCanNextPage()

  return (
    <div
      className="flex items-center justify-between overflow-clip px-2"
      style={{ overflowClipMargin: 1 }}
    >
      {/* Left side: row selection count OR total label */}
      <div className="text-muted-foreground hidden flex-1 text-sm sm:block">
        {showSelectedRows ? (
          <>
            {table.getFilteredSelectedRowModel().rows.length} of{' '}
            {serverPagination
              ? serverPagination.total
              : table.getFilteredRowModel().rows.length}{' '}
            row(s) selected.
          </>
        ) : totalLabel ? (
          totalLabel
        ) : null}
      </div>

      {/* Right side: pagination controls */}
      <div className="flex items-center sm:space-x-6 lg:space-x-8">
        {/* Rows per page selector */}
        <div className="flex items-center space-x-2">
          <p className="hidden text-sm font-medium sm:block">Rows per page</p>
          <Select
            value={`${table.getState().pagination.pageSize}`}
            onValueChange={(value) => {
              const newPageSize = Number(value)
              // Use setPagination to trigger onPaginationChange callback
              // This works for both client-side and manual pagination
              table.setPagination({
                pageIndex: 0, // Reset to first page on page size change
                pageSize: newPageSize,
              })
            }}
          >
            <SelectTrigger className="h-8 w-[70px]">
              <SelectValue placeholder={table.getState().pagination.pageSize} />
            </SelectTrigger>
            <SelectContent side="top">
              {pageSizes.map((pageSize) => (
                <SelectItem key={pageSize} value={`${pageSize}`}>
                  {pageSize}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {/* Page indicator */}
        <div className="flex w-[100px] items-center justify-center text-sm font-medium">
          Page {currentPage} of {totalPages}
        </div>

        {/* Navigation buttons */}
        <div className="flex items-center space-x-2">
          <Button
            variant="outline"
            className="hidden h-8 w-8 p-0 lg:flex"
            onClick={() => table.setPageIndex(0)}
            disabled={!canPrevious || isPending}
          >
            <span className="sr-only">Go to first page</span>
            <DoubleArrowLeftIcon className="h-4 w-4" />
          </Button>
          <Button
            variant="outline"
            className="h-8 w-8 p-0"
            onClick={() => table.previousPage()}
            disabled={!canPrevious || isPending}
          >
            <span className="sr-only">Go to previous page</span>
            <ChevronLeftIcon className="h-4 w-4" />
          </Button>
          <Button
            variant="outline"
            className="h-8 w-8 p-0"
            onClick={() => table.nextPage()}
            disabled={!canNext || isPending}
          >
            <span className="sr-only">Go to next page</span>
            <ChevronRightIcon className="h-4 w-4" />
          </Button>
          <Button
            variant="outline"
            className="hidden h-8 w-8 p-0 lg:flex"
            onClick={() => table.setPageIndex(totalPages - 1)}
            disabled={!canNext || isPending}
          >
            <span className="sr-only">Go to last page</span>
            <DoubleArrowRightIcon className="h-4 w-4" />
          </Button>
        </div>
      </div>
    </div>
  )
}
