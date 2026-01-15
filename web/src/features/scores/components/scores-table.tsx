'use client'

import { useCallback, useState, useTransition } from 'react'
import Link from 'next/link'
import { useSearchParams, useRouter, usePathname } from 'next/navigation'
import type { ColumnDef } from '@tanstack/react-table'
import {
  flexRender,
  getCoreRowModel,
  useReactTable,
} from '@tanstack/react-table'
import { ExternalLink } from 'lucide-react'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Input } from '@/components/ui/input'
import {
  ChevronLeftIcon,
  ChevronRightIcon,
} from '@radix-ui/react-icons'
import type { Score, ScoreDataType, ScoreSource } from '../types'
import type { Pagination } from '@/lib/api/core/types'
import { ScoreValueCell } from './score-value-cell'
import { formatDistanceToNow } from 'date-fns'

interface ScoresTableProps {
  data: Score[]
  pagination: Pagination
  projectSlug: string
  loading?: boolean
  error?: string
}

const columns: ColumnDef<Score>[] = [
  {
    accessorKey: 'name',
    header: 'Name',
    cell: ({ row }) => (
      <span className="font-medium">{row.original.name}</span>
    ),
  },
  {
    id: 'value',
    header: 'Value & Source',
    cell: ({ row }) => <ScoreValueCell score={row.original} />,
  },
  {
    accessorKey: 'data_type',
    header: 'Type',
    cell: ({ row }) => {
      const typeLabels: Record<ScoreDataType, string> = {
        NUMERIC: 'Numeric',
        BOOLEAN: 'Boolean',
        CATEGORICAL: 'Categorical',
      }
      return (
        <span className="text-muted-foreground text-sm">
          {typeLabels[row.original.data_type] || row.original.data_type}
        </span>
      )
    },
  },
  {
    accessorKey: 'trace_id',
    header: 'Trace',
    cell: ({ row, table }) => {
      const traceId = row.original.trace_id
      if (!traceId) {
        return <span className="text-muted-foreground">-</span>
      }
      const { projectSlug } = table.options.meta as { projectSlug: string }
      return (
        <Link
          href={`/projects/${projectSlug}/traces/${traceId}`}
          className="inline-flex items-center gap-1 text-sm text-blue-600 hover:text-blue-700 hover:underline"
          onClick={(e) => e.stopPropagation()}
        >
          <span className="font-mono truncate max-w-[120px]">
            {traceId.slice(0, 8)}...
          </span>
          <ExternalLink className="h-3 w-3" />
        </Link>
      )
    },
  },
  {
    accessorKey: 'span_id',
    header: 'Span',
    cell: ({ row }) =>
      row.original.span_id ? (
        <span className="font-mono text-sm text-muted-foreground truncate max-w-[100px]">
          {row.original.span_id.slice(0, 8)}...
        </span>
      ) : (
        <span className="text-muted-foreground">-</span>
      ),
  },
  {
    accessorKey: 'timestamp',
    header: 'Created',
    cell: ({ row }) => (
      <span className="text-sm text-muted-foreground">
        {formatDistanceToNow(new Date(row.original.timestamp), { addSuffix: true })}
      </span>
    ),
  },
]

export function ScoresTable({
  data,
  pagination,
  projectSlug,
  loading = false,
  error,
}: ScoresTableProps) {
  const router = useRouter()
  const pathname = usePathname()
  const searchParams = useSearchParams()
  const [isPending, startTransition] = useTransition()
  const [nameFilter, setNameFilter] = useState(searchParams.get('name') || '')
  const [sourceFilter, setSourceFilter] = useState(searchParams.get('source') || '')
  const [dataTypeFilter, setDataTypeFilter] = useState(searchParams.get('data_type') || '')

  const createQueryString = useCallback(
    (params: Record<string, string | number | undefined>) => {
      const current = new URLSearchParams(Array.from(searchParams.entries()))
      Object.entries(params).forEach(([key, value]) => {
        if (value === undefined || value === '') {
          current.delete(key)
        } else {
          current.set(key, String(value))
        }
      })
      return current.toString()
    },
    [searchParams]
  )

  const navigate = useCallback(
    (params: Record<string, string | number | undefined>) => {
      startTransition(() => {
        const queryString = createQueryString(params)
        router.push(`${pathname}?${queryString}`)
      })
    },
    [createQueryString, pathname, router]
  )

  const handleNameSearch = useCallback(() => {
    navigate({ name: nameFilter || undefined, page: 1 })
  }, [nameFilter, navigate])

  const handleSourceChange = useCallback(
    (value: string) => {
      setSourceFilter(value)
      navigate({ source: value === 'all' ? undefined : value, page: 1 })
    },
    [navigate]
  )

  const handleDataTypeChange = useCallback(
    (value: string) => {
      setDataTypeFilter(value)
      navigate({ data_type: value === 'all' ? undefined : value, page: 1 })
    },
    [navigate]
  )

  const handlePageChange = useCallback(
    (newPage: number) => {
      navigate({ page: newPage })
    },
    [navigate]
  )

  const handlePageSizeChange = useCallback(
    (newSize: number) => {
      navigate({ limit: newSize, page: 1 })
    },
    [navigate]
  )

  const table = useReactTable({
    data,
    columns,
    pageCount: pagination.totalPages,
    state: {
      pagination: {
        pageIndex: pagination.page - 1,
        pageSize: pagination.limit,
      },
    },
    manualPagination: true,
    getCoreRowModel: getCoreRowModel(),
    meta: { projectSlug },
  })

  if (loading) {
    return (
      <div className="space-y-4">
        <div className="flex items-center gap-4">
          <Skeleton className="h-9 w-[200px]" />
          <Skeleton className="h-9 w-[120px]" />
          <Skeleton className="h-9 w-[120px]" />
        </div>
        <div className="overflow-hidden rounded-md border">
          <Table>
            <TableHeader>
              <TableRow>
                {columns.map((_, index) => (
                  <TableHead key={index}>
                    <Skeleton className="h-6 w-20" />
                  </TableHead>
                ))}
              </TableRow>
            </TableHeader>
            <TableBody>
              {Array(5)
                .fill(0)
                .map((_, index) => (
                  <TableRow key={index}>
                    {columns.map((_, colIndex) => (
                      <TableCell key={colIndex}>
                        <Skeleton className="h-6 w-16" />
                      </TableCell>
                    ))}
                  </TableRow>
                ))}
            </TableBody>
          </Table>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="text-center">
          <p className="text-muted-foreground">Failed to load scores</p>
          <p className="text-destructive text-sm mt-1">{error}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center gap-4">
        <div className="flex items-center gap-2">
          <Input
            placeholder="Filter by name..."
            aria-label="Filter scores by name"
            value={nameFilter}
            onChange={(e) => setNameFilter(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleNameSearch()}
            className="h-9 w-[200px]"
          />
          <Button
            variant="outline"
            size="sm"
            onClick={handleNameSearch}
            disabled={isPending}
          >
            Search
          </Button>
        </div>

        <Select value={sourceFilter || 'all'} onValueChange={handleSourceChange}>
          <SelectTrigger className="h-9 w-[120px]" aria-label="Filter by source">
            <SelectValue placeholder="Source" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Sources</SelectItem>
            <SelectItem value="human">Human</SelectItem>
            <SelectItem value="code">SDK</SelectItem>
            <SelectItem value="llm">LLM</SelectItem>
          </SelectContent>
        </Select>

        <Select value={dataTypeFilter || 'all'} onValueChange={handleDataTypeChange}>
          <SelectTrigger className="h-9 w-[130px]" aria-label="Filter by data type">
            <SelectValue placeholder="Data Type" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Types</SelectItem>
            <SelectItem value="NUMERIC">Numeric</SelectItem>
            <SelectItem value="BOOLEAN">Boolean</SelectItem>
            <SelectItem value="CATEGORICAL">Categorical</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <div className="overflow-hidden rounded-md border">
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
            {table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow key={row.id}>
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
                  No scores found.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      <div className="flex items-center justify-between px-2">
        <div className="text-sm text-muted-foreground">
          {pagination.total} total score{pagination.total !== 1 ? 's' : ''}
        </div>
        <div className="flex items-center space-x-6">
          <div className="flex items-center space-x-2">
            <p className="text-sm font-medium">Rows per page</p>
            <Select
              value={String(pagination.limit)}
              onValueChange={(value) => handlePageSizeChange(Number(value))}
            >
              <SelectTrigger className="h-8 w-[70px]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {[10, 20, 50, 100].map((size) => (
                  <SelectItem key={size} value={String(size)}>
                    {size}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="flex w-[100px] items-center justify-center text-sm font-medium">
            Page {pagination.page} of {pagination.totalPages}
          </div>
          <div className="flex items-center space-x-2">
            <Button
              variant="outline"
              size="icon"
              className="h-8 w-8"
              onClick={() => handlePageChange(pagination.page - 1)}
              disabled={!pagination.hasPrev || isPending}
              aria-label="Go to previous page"
            >
              <ChevronLeftIcon className="h-4 w-4" aria-hidden="true" />
            </Button>
            <Button
              variant="outline"
              size="icon"
              className="h-8 w-8"
              onClick={() => handlePageChange(pagination.page + 1)}
              disabled={!pagination.hasNext || isPending}
              aria-label="Go to next page"
            >
              <ChevronRightIcon className="h-4 w-4" aria-hidden="true" />
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}
