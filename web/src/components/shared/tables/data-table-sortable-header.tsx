'use client'

import { ArrowUpDown, ArrowUp, ArrowDown } from 'lucide-react'
import { Button } from '@/components/ui/button'

interface SortableColumnHeaderProps<TField extends string> {
  label: string
  field: TField
  currentSort: TField | null
  currentOrder: 'asc' | 'desc' | null
  onSort: (field: TField | null, order: 'asc' | 'desc' | null) => void
}

/**
 * Generic sortable column header component for URL-state aware tables.
 *
 * Features:
 * - 3-state cycling: unsorted → desc → asc → unsorted
 * - Visual feedback with sort direction icons
 * - Generic type support for any sort field enum
 *
 * @example
 * ```tsx
 * <SortableColumnHeader
 *   label="Name"
 *   field="name"
 *   currentSort={sortBy}
 *   currentOrder={sortOrder}
 *   onSort={onSortChange}
 * />
 * ```
 */
export function SortableColumnHeader<TField extends string>({
  label,
  field,
  currentSort,
  currentOrder,
  onSort,
}: SortableColumnHeaderProps<TField>) {
  const isActive = currentSort === field

  const handleClick = () => {
    if (!isActive) {
      // First click: sort desc
      onSort(field, 'desc')
    } else if (currentOrder === 'desc') {
      // Second click: sort asc
      onSort(field, 'asc')
    } else {
      // Third click: clear sort
      onSort(null, null)
    }
  }

  return (
    <Button
      variant="ghost"
      size="sm"
      className="-ml-3 h-8 data-[state=open]:bg-accent"
      onClick={handleClick}
    >
      <span>{label}</span>
      {isActive ? (
        currentOrder === 'desc' ? (
          <ArrowDown className="ml-1.5 h-3.5 w-3.5" />
        ) : (
          <ArrowUp className="ml-1.5 h-3.5 w-3.5" />
        )
      ) : (
        <ArrowUpDown className="ml-1.5 h-3.5 w-3.5 opacity-50" />
      )}
    </Button>
  )
}
