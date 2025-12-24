'use client'

import { useMemo, useState } from 'react'
import { useRouter } from 'next/navigation'
import {
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
  type ColumnFiltersState,
  type SortingState,
  type VisibilityState,
} from '@tanstack/react-table'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Search, X } from 'lucide-react'
import type { PromptListItem, PromptType } from '../../types'
import { createPromptsColumns } from './prompts-columns'
import { PromptsDeleteDialog } from './prompts-delete-dialog'

interface PromptsTableProps {
  data: PromptListItem[]
  totalCount: number
  page: number
  pageSize: number
  totalPages: number
  isLoading?: boolean
  protectedLabels?: string[]
  projectSlug: string
  orgSlug: string
  onPageChange?: (page: number) => void
  onSearch?: (query: string) => void
  onTypeFilter?: (type: PromptType | undefined) => void
}

export function PromptsTable({
  data,
  totalCount,
  page,
  pageSize,
  totalPages,
  isLoading,
  protectedLabels = [],
  projectSlug,
  orgSlug,
  onPageChange,
  onSearch,
  onTypeFilter,
}: PromptsTableProps) {
  const router = useRouter()
  const [sorting, setSorting] = useState<SortingState>([])
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([])
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({})
  const [rowSelection, setRowSelection] = useState({})
  const [searchValue, setSearchValue] = useState('')
  const [typeFilter, setTypeFilter] = useState<string>('')
  const [deletePrompt, setDeletePrompt] = useState<PromptListItem | null>(null)

  const columns = useMemo(
    () =>
      createPromptsColumns({
        protectedLabels,
        onEdit: (prompt) => {
          router.push(`/projects/${projectSlug}/prompts/${prompt.id}`)
        },
        onDelete: (prompt) => {
          setDeletePrompt(prompt)
        },
        onPlayground: (prompt) => {
          router.push(`/projects/${projectSlug}/prompts/${prompt.id}/playground`)
        },
        onViewHistory: (prompt) => {
          router.push(`/projects/${projectSlug}/prompts/${prompt.id}/versions`)
        },
      }),
    [projectSlug, protectedLabels, router]
  )

  const table = useReactTable({
    data,
    columns,
    pageCount: totalPages,
    state: {
      sorting,
      columnVisibility,
      rowSelection,
      columnFilters,
      pagination: { pageIndex: page - 1, pageSize },
    },
    enableRowSelection: true,
    manualPagination: true,
    onRowSelectionChange: setRowSelection,
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    onColumnVisibilityChange: setColumnVisibility,
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
  })

  const handleSearch = (value: string) => {
    setSearchValue(value)
    onSearch?.(value)
  }

  const handleTypeFilter = (value: string) => {
    setTypeFilter(value)
    onTypeFilter?.(value === 'all' ? undefined : (value as PromptType))
  }

  const handleRowClick = (prompt: PromptListItem, e: React.MouseEvent) => {
    if ((e.target as HTMLElement).closest('[role="checkbox"], button')) {
      return
    }
    router.push(`/projects/${projectSlug}/prompts/${prompt.id}`)
  }

  return (
    <div className="space-y-4">
      {/* Toolbar */}
      <div className="flex items-center justify-between gap-4">
        <div className="flex flex-1 items-center gap-2">
          <div className="relative flex-1 max-w-sm">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder="Search prompts..."
              value={searchValue}
              onChange={(e) => handleSearch(e.target.value)}
              className="pl-9"
            />
            {searchValue && (
              <Button
                variant="ghost"
                size="sm"
                className="absolute right-1 top-1/2 h-6 w-6 -translate-y-1/2 p-0"
                onClick={() => handleSearch('')}
              >
                <X className="h-4 w-4" />
              </Button>
            )}
          </div>
          <Select value={typeFilter || 'all'} onValueChange={handleTypeFilter}>
            <SelectTrigger className="w-[120px]" size="sm">
              <SelectValue placeholder="Type" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All types</SelectItem>
              <SelectItem value="text">Text</SelectItem>
              <SelectItem value="chat">Chat</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      {/* Table */}
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <TableHead key={header.id} colSpan={header.colSpan}>
                    {header.isPlaceholder
                      ? null
                      : flexRender(header.column.columnDef.header, header.getContext())}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {isLoading ? (
              <TableRow>
                <TableCell colSpan={columns.length} className="h-24 text-center">
                  Loading prompts...
                </TableCell>
              </TableRow>
            ) : table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  data-state={row.getIsSelected() && 'selected'}
                  className="cursor-pointer hover:bg-muted/50"
                  onClick={(e) => handleRowClick(row.original, e)}
                >
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id}>
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell colSpan={columns.length} className="h-24 text-center">
                  No prompts found.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      {/* Pagination */}
      <div className="flex items-center justify-between px-2">
        <div className="flex-1 text-sm text-muted-foreground">
          {table.getFilteredSelectedRowModel().rows.length} of {totalCount} row(s) selected.
        </div>
        <div className="flex items-center space-x-6 lg:space-x-8">
          <div className="flex items-center space-x-2">
            <p className="text-sm font-medium">Page {page} of {totalPages}</p>
          </div>
          <div className="flex items-center space-x-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => onPageChange?.(page - 1)}
              disabled={page <= 1}
            >
              Previous
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => onPageChange?.(page + 1)}
              disabled={page >= totalPages}
            >
              Next
            </Button>
          </div>
        </div>
      </div>

      {/* Delete Dialog */}
      <PromptsDeleteDialog
        prompt={deletePrompt}
        open={!!deletePrompt}
        onOpenChange={(open) => !open && setDeletePrompt(null)}
      />
    </div>
  )
}
