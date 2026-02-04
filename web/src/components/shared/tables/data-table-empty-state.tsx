'use client'

import type { ReactNode } from 'react'
import { Button } from '@/components/ui/button'

interface DataTableEmptyStateProps {
  /** Main title text */
  title?: string
  /** Descriptive text below title */
  description?: string
  /** Icon to display (centered, with muted styling) */
  icon?: ReactNode
  /** Whether filters are currently active */
  hasFilters?: boolean
  /** Callback to clear all filters */
  onClearFilters?: () => void
  /** Optional custom action button/element */
  action?: ReactNode
}

/**
 * Empty state component for data tables.
 *
 * Displays when no data matches current filters or when table is empty.
 * Supports both filtered-to-zero states (with clear filters) and
 * truly empty states (with optional CTA).
 *
 * @example
 * ```tsx
 * // Filtered to zero results
 * <DataTableEmptyState
 *   title="No results found"
 *   description="Try adjusting your filters"
 *   hasFilters={true}
 *   onClearFilters={handleReset}
 * />
 *
 * // Truly empty state with icon
 * <DataTableEmptyState
 *   icon={<Database className="h-full w-full" />}
 *   title="No datasets yet"
 *   description="Create your first dataset to get started"
 *   action={<Button>Create Dataset</Button>}
 * />
 * ```
 */
export function DataTableEmptyState({
  title = 'No results found',
  description,
  icon,
  hasFilters = false,
  onClearFilters,
  action,
}: DataTableEmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center py-12">
      {icon && (
        <div className="flex h-16 w-16 items-center justify-center rounded-full bg-muted mb-4">
          <div className="h-8 w-8 text-muted-foreground">{icon}</div>
        </div>
      )}
      <div className="flex flex-col items-center text-center max-w-md space-y-2">
        <p className="text-muted-foreground font-medium">{title}</p>
        {description && (
          <p className="text-sm text-muted-foreground">{description}</p>
        )}
        {hasFilters && onClearFilters && (
          <Button variant="ghost" size="sm" onClick={onClearFilters} className="mt-2">
            Clear filters
          </Button>
        )}
        {action && <div className="mt-4">{action}</div>}
      </div>
    </div>
  )
}
