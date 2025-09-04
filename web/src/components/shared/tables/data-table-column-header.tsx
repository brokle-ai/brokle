'use client'

import {
  ArrowDownIcon,
  ArrowUpIcon,
  CaretSortIcon,
} from '@radix-ui/react-icons'
import { Column } from '@tanstack/react-table'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'

interface DataTableColumnHeaderProps<TData, TValue>
  extends React.HTMLAttributes<HTMLDivElement> {
  column: Column<TData, TValue>
  title: string
}

export function DataTableColumnHeader<TData, TValue>({
  column,
  title,
  className,
}: DataTableColumnHeaderProps<TData, TValue>) {
  if (!column.getCanSort()) {
    return <div className={cn(className)}>{title}</div>
  }

  const handleClick = () => {
    const currentSort = column.getIsSorted()
    
    if (!currentSort) {
      // No sorting → Sort ascending
      column.toggleSorting(false)
    } else if (currentSort === 'asc') {
      // Ascending → Sort descending
      column.toggleSorting(true)
    } else {
      // Descending → Clear sorting
      column.clearSorting()
    }
  }

  return (
    <div className={cn('flex items-center space-x-2', className)}>
      <Button
        variant='ghost'
        size='sm'
        className='hover:bg-accent -ml-3 h-8 cursor-pointer'
        onClick={(e) => {
          e.stopPropagation()
          handleClick()
        }}
        aria-label={`Sort by ${title}`}
      >
        <span>{title}</span>
        {column.getIsSorted() === 'desc' ? (
          <ArrowDownIcon className='ml-2 h-4 w-4' />
        ) : column.getIsSorted() === 'asc' ? (
          <ArrowUpIcon className='ml-2 h-4 w-4' />
        ) : (
          <CaretSortIcon className='ml-2 h-4 w-4' />
        )}
      </Button>
    </div>
  )
}