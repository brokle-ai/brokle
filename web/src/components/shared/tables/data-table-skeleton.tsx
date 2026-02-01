'use client'

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Skeleton } from '@/components/ui/skeleton'

interface DataTableSkeletonProps {
  /** Number of columns to render in skeleton */
  columns?: number
  /** Number of rows to render in skeleton */
  rows?: number
  /** Show toolbar skeleton above table */
  showToolbar?: boolean
  /** Show pagination skeleton below table */
  showPagination?: boolean
  /** Number of toolbar skeleton items (for filters) */
  toolbarSlots?: number
  /** Custom column widths for more realistic skeleton */
  columnWidths?: number[]
}

/**
 * Generic skeleton component for data tables.
 *
 * Provides a consistent loading state across all table implementations.
 * Configurable columns, rows, and optional toolbar/pagination sections.
 *
 * @example
 * ```tsx
 * // Basic usage
 * <DataTableSkeleton />
 *
 * // Customized
 * <DataTableSkeleton
 *   columns={5}
 *   rows={10}
 *   showToolbar
 *   showPagination
 *   toolbarSlots={3}
 * />
 * ```
 */
export function DataTableSkeleton({
  columns = 6,
  rows = 5,
  showToolbar = true,
  showPagination = true,
  toolbarSlots = 2,
  columnWidths,
}: DataTableSkeletonProps) {
  // Default column widths if not provided
  const widths = columnWidths || Array(columns).fill(80)

  return (
    <div className="space-y-4">
      {/* Toolbar skeleton */}
      {showToolbar && (
        <div className="flex items-center gap-3">
          <Skeleton className="h-9 w-[200px]" />
          {Array(toolbarSlots - 1)
            .fill(0)
            .map((_, index) => (
              <Skeleton key={index} className="h-9 w-[130px]" />
            ))}
        </div>
      )}

      {/* Table skeleton */}
      <div className="overflow-hidden rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              {Array(columns)
                .fill(0)
                .map((_, index) => (
                  <TableHead key={index}>
                    <Skeleton className="h-6 w-20" />
                  </TableHead>
                ))}
            </TableRow>
          </TableHeader>
          <TableBody>
            {Array(rows)
              .fill(0)
              .map((_, rowIndex) => (
                <TableRow key={rowIndex}>
                  {Array(columns)
                    .fill(0)
                    .map((_, colIndex) => (
                      <TableCell key={colIndex}>
                        <Skeleton
                          className="h-6"
                          style={{ width: widths[colIndex] || 64 }}
                        />
                      </TableCell>
                    ))}
                </TableRow>
              ))}
          </TableBody>
        </Table>
      </div>

      {/* Pagination skeleton */}
      {showPagination && (
        <div className="flex items-center justify-between px-2">
          <Skeleton className="h-5 w-24" />
          <div className="flex items-center gap-4">
            <Skeleton className="h-8 w-[130px]" />
            <Skeleton className="h-5 w-20" />
            <Skeleton className="h-8 w-20" />
          </div>
        </div>
      )}
    </div>
  )
}
