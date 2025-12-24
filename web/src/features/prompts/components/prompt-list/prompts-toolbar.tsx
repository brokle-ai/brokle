'use client'

import { Cross2Icon } from '@radix-ui/react-icons'
import { type Table } from '@tanstack/react-table'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { DataTableViewOptions, DataTableFacetedFilter } from '@/components/data-table'

const promptTypes = [
  { label: 'Text', value: 'text' },
  { label: 'Chat', value: 'chat' },
]

type PromptsToolbarProps<TData> = {
  table: Table<TData>
  isPending?: boolean
  onReset?: () => void
}

export function PromptsToolbar<TData>({
  table,
  isPending = false,
  onReset,
}: PromptsToolbarProps<TData>) {
  const isFiltered =
    table.getState().columnFilters.length > 0 || table.getState().globalFilter

  const handleReset = () => {
    onReset?.()
  }

  return (
    <div className='flex items-center justify-between'>
      <div className='flex flex-1 flex-col-reverse items-start gap-y-2 sm:flex-row sm:items-center sm:space-x-2'>
        <Input
          placeholder='Filter by name...'
          value={table.getState().globalFilter ?? ''}
          onChange={(event) => table.setGlobalFilter(event.target.value)}
          className='h-8 w-[150px] lg:w-[250px]'
          disabled={isPending}
        />
        <div className='flex gap-x-2'>
          {table.getColumn('type') && (
            <DataTableFacetedFilter
              column={table.getColumn('type')}
              title='Type'
              options={promptTypes}
            />
          )}
        </div>
        {isFiltered && (
          <Button
            variant='ghost'
            onClick={handleReset}
            className='h-8 px-2 lg:px-3'
            disabled={isPending}
          >
            Reset
            <Cross2Icon className='ms-2 h-4 w-4' />
          </Button>
        )}
      </div>
      <DataTableViewOptions table={table} />
    </div>
  )
}
