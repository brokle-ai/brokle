'use client'

import { TableCell, TableRow } from '@/components/ui/table'
import { Skeleton } from '@/components/ui/skeleton'

interface TableSkeletonProps {
  columns: number
  rows?: number
}

export function TableSkeleton({ columns, rows = 10 }: TableSkeletonProps) {
  return (
    <>
      {Array.from({ length: rows }).map((_, rowIndex) => (
        <TableRow key={`skeleton-row-${rowIndex}`} className="group/row">
          {Array.from({ length: columns }).map((_, colIndex) => (
            <TableCell 
              key={`skeleton-cell-${rowIndex}-${colIndex}`}
              className="bg-background group-hover/row:bg-muted"
            >
              {colIndex === 0 ? (
                // Column 1: Select checkbox - w-10 min-w-10
                <div className="flex items-center">
                  <Skeleton className="h-4 w-4 rounded" />
                </div>
              ) : colIndex === 1 ? (
                // Column 2: Username - w-36 min-w-36 max-w-36
                <Skeleton className="h-4 w-32 max-w-36" />
              ) : colIndex === 2 ? (
                // Column 3: Full Name - w-36 min-w-36 max-w-36
                <Skeleton className="h-4 w-32 max-w-36" />
              ) : colIndex === 3 ? (
                // Column 4: Email - w-48 min-w-48
                <Skeleton className="h-4 w-44" />
              ) : colIndex === 4 ? (
                // Column 5: Phone Number - w-32 min-w-32
                <Skeleton className="h-4 w-28" />
              ) : colIndex === 5 ? (
                // Column 6: Status Badge - w-20 min-w-20
                <Skeleton className="h-6 w-16 rounded-full" />
              ) : colIndex === 6 ? (
                // Column 7: Role with icon - w-28 min-w-28
                <div className="flex items-center space-x-2">
                  <Skeleton className="h-4 w-4 rounded" />
                  <Skeleton className="h-4 w-20" />
                </div>
              ) : (
                // Column 8: Actions - w-12 min-w-12
                <Skeleton className="h-4 w-6" />
              )}
            </TableCell>
          ))}
        </TableRow>
      ))}
    </>
  )
}