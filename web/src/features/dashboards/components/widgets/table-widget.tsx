'use client'

import { useMemo } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { cn } from '@/lib/utils'
import type { Widget } from '../../types'

interface TableWidgetProps {
  widget: Widget
  data: TableData | null
  isLoading: boolean
  error?: string
}

interface TableData {
  columns: ColumnDefinition[]
  rows: Record<string, unknown>[]
}

interface ColumnDefinition {
  key: string
  label: string
  type?: 'string' | 'number' | 'date' | 'boolean'
  align?: 'left' | 'center' | 'right'
  width?: string
  format?: (value: unknown) => string
}

function formatCellValue(value: unknown, column: ColumnDefinition): string {
  if (value === null || value === undefined) return '-'

  if (column.format) {
    return column.format(value)
  }

  switch (column.type) {
    case 'number':
      if (typeof value === 'number') {
        if (value >= 1000000) return `${(value / 1000000).toFixed(1)}M`
        if (value >= 1000) return `${(value / 1000).toFixed(1)}K`
        if (Number.isInteger(value)) return String(value)
        return value.toFixed(2)
      }
      return String(value)
    case 'date':
      if (value instanceof Date) {
        return value.toLocaleDateString()
      }
      if (typeof value === 'string') {
        try {
          return new Date(value).toLocaleDateString()
        } catch {
          return value
        }
      }
      return String(value)
    case 'boolean':
      return value ? 'Yes' : 'No'
    default:
      return String(value)
  }
}

function getCellAlignment(column: ColumnDefinition): string {
  if (column.align) {
    return column.align === 'center' ? 'text-center' :
           column.align === 'right' ? 'text-right' : 'text-left'
  }
  // Default alignment based on type
  return column.type === 'number' ? 'text-right' : 'text-left'
}

export function TableWidget({ widget, data, isLoading, error }: TableWidgetProps) {
  const maxRows = widget.config?.maxRows as number || 10
  const showHeader = widget.config?.showHeader !== false
  const compact = widget.config?.compact === true

  // Auto-derive columns from data if not provided
  const columns = useMemo((): ColumnDefinition[] => {
    if (data?.columns && data.columns.length > 0) {
      return data.columns
    }
    if (data?.rows && data.rows.length > 0) {
      return Object.keys(data.rows[0]).map(key => ({
        key,
        label: key.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase()),
        type: typeof data.rows[0][key] === 'number' ? 'number' as const : 'string' as const
      }))
    }
    return []
  }, [data])

  const displayedRows = useMemo(() => {
    if (!data?.rows) return []
    return data.rows.slice(0, maxRows)
  }, [data?.rows, maxRows])

  if (isLoading) {
    return (
      <Card className="h-full">
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium">
            <Skeleton className="h-4 w-32" />
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            <Skeleton className="h-8 w-full" />
            <Skeleton className="h-6 w-full" />
            <Skeleton className="h-6 w-full" />
            <Skeleton className="h-6 w-full" />
          </div>
        </CardContent>
      </Card>
    )
  }

  if (error) {
    return (
      <Card className="h-full">
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium">{widget.title}</CardTitle>
        </CardHeader>
        <CardContent className="flex items-center justify-center h-[200px]">
          <p className="text-sm text-destructive">{error}</p>
        </CardContent>
      </Card>
    )
  }

  if (!data || displayedRows.length === 0 || columns.length === 0) {
    return (
      <Card className="h-full">
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium">{widget.title}</CardTitle>
        </CardHeader>
        <CardContent className="flex items-center justify-center h-[200px]">
          <p className="text-sm text-muted-foreground">No data available</p>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card className="h-full">
      <CardHeader className="pb-2">
        <CardTitle className="text-sm font-medium">{widget.title}</CardTitle>
        {widget.description && (
          <CardDescription className="text-xs">{widget.description}</CardDescription>
        )}
      </CardHeader>
      <CardContent className="overflow-auto">
        <Table>
          {showHeader && (
            <TableHeader>
              <TableRow>
                {columns.map((column) => (
                  <TableHead
                    key={column.key}
                    className={cn(
                      getCellAlignment(column),
                      compact ? 'py-1 px-2 text-xs' : 'text-xs',
                      column.width && `w-[${column.width}]`
                    )}
                  >
                    {column.label}
                  </TableHead>
                ))}
              </TableRow>
            </TableHeader>
          )}
          <TableBody>
            {displayedRows.map((row, rowIndex) => (
              <TableRow key={rowIndex}>
                {columns.map((column) => (
                  <TableCell
                    key={column.key}
                    className={cn(
                      getCellAlignment(column),
                      compact ? 'py-1 px-2 text-xs' : 'text-xs'
                    )}
                  >
                    {formatCellValue(row[column.key], column)}
                  </TableCell>
                ))}
              </TableRow>
            ))}
          </TableBody>
        </Table>
        {data.rows.length > maxRows && (
          <p className="text-xs text-muted-foreground text-center mt-2">
            Showing {maxRows} of {data.rows.length} rows
          </p>
        )}
      </CardContent>
    </Card>
  )
}

export type { TableData, ColumnDefinition }
