'use client'

import { useState, useCallback } from 'react'
import { useRouter, usePathname } from 'next/navigation'
import { useUsers } from '../context/users-context'
import {
  ColumnDef,
  RowData,
  SortingState,
  VisibilityState,
  ColumnFiltersState,
  PaginationState,
  OnChangeFn,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  useReactTable,
} from '@tanstack/react-table'
import { cn } from '@/lib/utils'
import { useTableUrlState } from '@/hooks/use-table-url-state'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { DataTablePagination } from '@/components/shared/tables/data-table-pagination'
import { DataTableToolbar } from '@/components/shared/tables/data-table-toolbar'
import { FilterField } from '@/components/shared/tables/data-table'
import { TableSkeleton } from '@/components/shared/tables/table-skeleton'
import { User } from '../data/schema'
import { DataTableBulkActions } from './data-table-bulk-actions'
import { userTypes } from '../data/data'

declare module '@tanstack/react-table' {
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  interface ColumnMeta<TData extends RowData, TValue> {
    className: string
  }
}

interface PaginationMeta {
  page: number
  pageSize: number
  total: number
  totalPages: number
  hasNextPage: boolean
  hasPreviousPage: boolean
}

interface DataTableProps {
  columns: ColumnDef<User>[]
  data: User[]
  pagination?: PaginationMeta | null
  loading?: boolean
}

export function UsersTable({ columns, data, pagination: serverPagination, loading = false }: DataTableProps) {
  // Next.js navigation hooks
  const router = useRouter()
  const pathname = usePathname()
  
  // Users context for dialog management
  const { setOpen, setCurrentRow } = useUsers()

  // Local UI-only states
  const [rowSelection, setRowSelection] = useState({})
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({})

  // URL state reading (pagination handled by server, sorting from URL)
  const { columnFilters, sorting } = useTableUrlState({
    globalFilter: { enabled: false },
    sorting: { enabled: true },
    columnFilters: [
      { columnId: 'username', searchKey: 'username', type: 'string' },
      { columnId: 'status', searchKey: 'status', type: 'array' },
      { columnId: 'role', searchKey: 'role', type: 'array' },
    ],
  })

  // Create pagination state from server data
  const pagination = serverPagination
    ? {
        pageIndex: serverPagination.page - 1, // Convert to 0-based
        pageSize: serverPagination.pageSize,
      }
    : { pageIndex: 0, pageSize: 10 }

  // Pagination change handler using Next.js standards
  const onPaginationChange: OnChangeFn<PaginationState> = useCallback((updater) => {
    const newPagination = typeof updater === 'function' ? updater(pagination) : updater
    const params = new URLSearchParams(window.location.search)
    
    // Update page (convert from 0-based to 1-based)
    const newPage = newPagination.pageIndex + 1
    if (newPage > 1) {
      params.set('page', newPage.toString())
    } else {
      params.delete('page')
    }
    
    // Update page size
    if (newPagination.pageSize !== 10) {
      params.set('pageSize', newPagination.pageSize.toString())
    } else {
      params.delete('pageSize')
    }
    
    const newUrl = `${pathname}?${params.toString()}`
    router.replace(newUrl)
  }, [pagination, pathname, router])

  // Column filters change handler
  const onColumnFiltersChange: OnChangeFn<ColumnFiltersState> = useCallback((updater) => {
    const newFilters = typeof updater === 'function' ? updater(columnFilters) : updater
    const params = new URLSearchParams(window.location.search)
    
    // Reset to first page when filtering
    params.delete('page')
    
    // Update filters
    params.delete('username')
    params.delete('status')
    params.delete('role')
    
    newFilters.forEach((filter) => {
      if (filter.id === 'username' && typeof filter.value === 'string' && filter.value.trim()) {
        params.set('username', filter.value)
      } else if (filter.id === 'status' && Array.isArray(filter.value) && filter.value.length > 0) {
        params.set('status', JSON.stringify(filter.value))
      } else if (filter.id === 'role' && Array.isArray(filter.value) && filter.value.length > 0) {
        params.set('role', JSON.stringify(filter.value))
      }
    })
    
    router.replace(`${pathname}?${params.toString()}`)
  }, [columnFilters, pathname, router])

  // Sorting change handler
  const onSortingChange: OnChangeFn<SortingState> = useCallback((updater) => {
    const newSorting = typeof updater === 'function' ? updater(sorting) : updater
    const params = new URLSearchParams(window.location.search)
    
    // Reset to first page when sorting
    params.delete('page')
    
    // Update sorting parameters
    if (newSorting.length > 0) {
      const sort = newSorting[0] // Only support single column sorting for now
      params.set('sortBy', sort.id)
      params.set('sortOrder', sort.desc ? 'desc' : 'asc')
    } else {
      params.delete('sortBy')
      params.delete('sortOrder')
    }
    
    router.replace(`${pathname}?${params.toString()}`)
  }, [sorting, pathname, router])

  // Filter fields configuration for the shared toolbar
  const filterFields: FilterField[] = [
    {
      key: 'status',
      title: 'Status',
      options: [
        { label: 'Active', value: 'active' },
        { label: 'Inactive', value: 'inactive' },
        { label: 'Invited', value: 'invited' },
        { label: 'Suspended', value: 'suspended' },
      ],
    },
    {
      key: 'role',
      title: 'Role',
      options: userTypes.map((type) => ({
        label: type.label,
        value: type.value,
        icon: type.icon,
      })),
    },
  ]

  const table = useReactTable({
    data,
    columns,
    state: {
      sorting,
      pagination,
      rowSelection,
      columnFilters,
      columnVisibility,
    },
    enableRowSelection: true,
    onPaginationChange,
    onColumnFiltersChange,
    onSortingChange,
    onRowSelectionChange: setRowSelection,
    onColumnVisibilityChange: setColumnVisibility,
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    // Server handles pagination and sorting
    manualPagination: true,
    manualSorting: true,
    pageCount: serverPagination?.totalPages ?? 1,
  })

  return (
    <div className='space-y-4 max-sm:has-[div[role="toolbar"]]:mb-16'>
      {/* Screen reader live region for status updates */}
      <div 
        aria-live="polite" 
        aria-atomic="true" 
        className="sr-only"
        role="status"
      >
        {serverPagination && (
          `Showing page ${serverPagination.page} of ${serverPagination.totalPages}, ${serverPagination.total} users total`
        )}
      </div>
      
      <DataTableToolbar 
        table={table} 
        searchField="username"
        filterFields={filterFields}
      />
      <div className='overflow-hidden rounded-md border'>
        <Table role="table" aria-label="Users table with pagination and filtering" className="table-fixed w-full">
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id} className='group/row'>
                {headerGroup.headers.map((header) => {
                  return (
                    <TableHead
                      key={header.id}
                      colSpan={header.colSpan}
                      className={cn(
                        'bg-background group-hover/row:bg-muted group-data-[state=selected]/row:bg-muted',
                        header.column.columnDef.meta?.className ?? ''
                      )}
                    >
                      {header.isPlaceholder
                        ? null
                        : flexRender(
                            header.column.columnDef.header,
                            header.getContext()
                          )}
                    </TableHead>
                  )
                })}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {loading ? (
              <TableSkeleton columns={columns.length} rows={serverPagination?.pageSize ?? 10} />
            ) : table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  data-state={row.getIsSelected() && 'selected'}
                  className='group/row cursor-pointer'
                  onClick={() => {
                    // Open edit dialog with selected user
                    setCurrentRow(row.original)
                    setOpen('edit')
                  }}
                >
                  {row.getVisibleCells().map((cell) => (
                    <TableCell
                      key={cell.id}
                      className={cn(
                        'bg-background group-hover/row:bg-muted group-data-[state=selected]/row:bg-muted',
                        cell.column.columnDef.meta?.className ?? ''
                      )}
                    >
                      {flexRender(
                        cell.column.columnDef.cell,
                        cell.getContext()
                      )}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell
                  colSpan={columns.length}
                  className='h-24 text-center'
                >
                  No results.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
      <DataTablePagination table={table} serverPagination={serverPagination} />
      <DataTableBulkActions table={table} />
    </div>
  )
}