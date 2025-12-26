'use client'

import {
  ChevronLeftIcon,
  ChevronRightIcon,
  DoubleArrowLeftIcon,
  DoubleArrowRightIcon,
} from '@radix-ui/react-icons'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

interface CardListPaginationProps {
  page: number
  pageSize: number
  totalCount: number
  onPageChange: (page: number) => void
  onPageSizeChange?: (pageSize: number) => void
  isPending?: boolean
  pageSizeOptions?: number[]
}

export function CardListPagination({
  page,
  pageSize,
  totalCount,
  onPageChange,
  onPageSizeChange,
  isPending = false,
  pageSizeOptions = [12, 24, 36, 48],
}: CardListPaginationProps) {
  const totalPages = Math.ceil(totalCount / pageSize)
  const canPreviousPage = page > 1
  const canNextPage = page < totalPages

  if (totalCount === 0) {
    return null
  }

  return (
    <div
      className='flex items-center justify-end overflow-clip px-2 mt-4'
      style={{ overflowClipMargin: 1 }}
    >
      <div className='flex items-center sm:space-x-6 lg:space-x-8'>
        {onPageSizeChange && (
          <div className='flex items-center space-x-2'>
            <p className='hidden text-sm font-medium sm:block'>Items per page</p>
            <Select
              value={`${pageSize}`}
              onValueChange={(value) => {
                onPageSizeChange(Number(value))
              }}
            >
              <SelectTrigger className='h-8 w-[70px]'>
                <SelectValue placeholder={pageSize} />
              </SelectTrigger>
              <SelectContent side='top'>
                {pageSizeOptions.map((size) => (
                  <SelectItem key={size} value={`${size}`}>
                    {size}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        )}
        <div className='flex w-[100px] items-center justify-center text-sm font-medium'>
          Page {page} of {totalPages}
        </div>
        <div className='flex items-center space-x-2'>
          <Button
            variant='outline'
            className='hidden h-8 w-8 p-0 lg:flex'
            onClick={() => onPageChange(1)}
            disabled={!canPreviousPage || isPending}
          >
            <span className='sr-only'>Go to first page</span>
            <DoubleArrowLeftIcon className='h-4 w-4' />
          </Button>
          <Button
            variant='outline'
            className='h-8 w-8 p-0'
            onClick={() => onPageChange(page - 1)}
            disabled={!canPreviousPage || isPending}
          >
            <span className='sr-only'>Go to previous page</span>
            <ChevronLeftIcon className='h-4 w-4' />
          </Button>
          <Button
            variant='outline'
            className='h-8 w-8 p-0'
            onClick={() => onPageChange(page + 1)}
            disabled={!canNextPage || isPending}
          >
            <span className='sr-only'>Go to next page</span>
            <ChevronRightIcon className='h-4 w-4' />
          </Button>
          <Button
            variant='outline'
            className='hidden h-8 w-8 p-0 lg:flex'
            onClick={() => onPageChange(totalPages)}
            disabled={!canNextPage || isPending}
          >
            <span className='sr-only'>Go to last page</span>
            <DoubleArrowRightIcon className='h-4 w-4' />
          </Button>
        </div>
      </div>
    </div>
  )
}
