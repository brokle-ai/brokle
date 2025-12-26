'use client'

import { Cross2Icon } from '@radix-ui/react-icons'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

interface CardListToolbarProps {
  searchPlaceholder?: string
  searchValue: string
  onSearchChange: (value: string) => void
  isPending?: boolean
  onReset?: () => void
  isFiltered?: boolean
  children?: React.ReactNode
}

export function CardListToolbar({
  searchPlaceholder = 'Search...',
  searchValue,
  onSearchChange,
  isPending = false,
  onReset,
  isFiltered = false,
  children,
}: CardListToolbarProps) {
  return (
    <div className='flex items-center justify-between mb-4'>
      <div className='flex flex-1 flex-col-reverse items-start gap-y-2 sm:flex-row sm:items-center sm:space-x-2'>
        <Input
          placeholder={searchPlaceholder}
          value={searchValue}
          onChange={(event) => onSearchChange(event.target.value)}
          className='h-8 w-[150px] lg:w-[250px]'
          disabled={isPending}
        />
        {children && <div className='flex gap-x-2'>{children}</div>}
        {isFiltered && onReset && (
          <Button
            variant='ghost'
            onClick={onReset}
            className='h-8 px-2 lg:px-3'
            disabled={isPending}
          >
            Reset
            <Cross2Icon className='ms-2 h-4 w-4' />
          </Button>
        )}
      </div>
    </div>
  )
}
