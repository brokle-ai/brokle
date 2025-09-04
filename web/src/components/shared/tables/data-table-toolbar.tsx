'use client'

import * as React from 'react'
import { Cross2Icon } from '@radix-ui/react-icons'
import { Table } from '@tanstack/react-table'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { FilterField } from './data-table'
import { DataTableFacetedFilter } from './data-table-faceted-filter'
import { DataTableViewOptions } from './data-table-view-options'

interface DataTableToolbarProps<TData> {
  table: Table<TData>
  searchField?: string
  filterFields?: FilterField[]
  children?: React.ReactNode
}

export function DataTableToolbar<TData>({
  table,
  searchField,
  filterFields = [],
  children,
}: DataTableToolbarProps<TData>) {
  const isFiltered = table.getState().columnFilters.length > 0

  return (
    <div className='flex items-center justify-between'>
      <div className='flex flex-1 flex-col-reverse items-start gap-y-2 sm:flex-row sm:items-center sm:space-x-2'>
        {searchField && (
          <Input
            placeholder={`Filter by ${searchField}...`}
            value={(table.getColumn(searchField)?.getFilterValue() as string) ?? ''}
            onChange={(event) =>
              table.getColumn(searchField)?.setFilterValue(event.target.value)
            }
            className='h-8 w-[150px] lg:w-[250px]'
          />
        )}
        <div className='flex gap-x-2'>
          {filterFields.map((field) => {
            const column = table.getColumn(field.key)
            return (
              column && (
                <DataTableFacetedFilter
                  key={field.key}
                  column={column}
                  title={field.title}
                  options={field.options}
                />
              )
            )
          })}
        </div>
        {isFiltered && (
          <Button
            variant='ghost'
            onClick={() => table.resetColumnFilters()}
            className='h-8 px-2 lg:px-3'
          >
            Reset
            <Cross2Icon className='ml-2 h-4 w-4' />
          </Button>
        )}
        {children}
      </div>
      <DataTableViewOptions table={table} />
    </div>
  )
}