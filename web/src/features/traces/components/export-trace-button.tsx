'use client'

import * as React from 'react'
import { Download, FileJson, FileSpreadsheet, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { toast } from 'sonner'
import type { Trace, Span } from '../data/schema'
import { traceToCSV, traceToJSON, downloadFile } from '../utils/export-utils'

interface ExportTraceButtonProps {
  trace: Trace
  spans: Span[]
  variant?: 'default' | 'icon'
  className?: string
}

type ExportFormat = 'csv' | 'json'

/**
 * ExportTraceButton - Export trace data as CSV or JSON
 *
 * Provides a dropdown menu to export trace and span data.
 * Two variants:
 * - icon: Compact icon-only button (default for header)
 * - default: Button with text label
 */
export function ExportTraceButton({
  trace,
  spans,
  variant = 'icon',
  className,
}: ExportTraceButtonProps) {
  const [isExporting, setIsExporting] = React.useState(false)

  const handleExport = async (format: ExportFormat) => {
    setIsExporting(true)
    try {
      const timestamp = new Date().toISOString().split('T')[0]
      const filename = `trace_${trace.trace_id.slice(0, 8)}_${timestamp}.${format}`

      if (format === 'csv') {
        const content = traceToCSV(trace, spans)
        downloadFile(content, filename, 'text/csv;charset=utf-8')
      } else {
        const content = traceToJSON(trace, spans)
        downloadFile(content, filename, 'application/json')
      }

      toast.success(`Trace exported as ${format.toUpperCase()}`)
    } catch (error) {
      console.error('Export failed:', error)
      toast.error('Failed to export trace')
    } finally {
      setIsExporting(false)
    }
  }

  if (variant === 'icon') {
    return (
      <DropdownMenu>
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger asChild>
              <DropdownMenuTrigger asChild>
                <Button
                  variant='ghost'
                  size='icon'
                  className='h-8 w-8'
                  disabled={isExporting}
                >
                  {isExporting ? (
                    <Loader2 className='h-4 w-4 animate-spin' />
                  ) : (
                    <Download className='h-4 w-4' />
                  )}
                </Button>
              </DropdownMenuTrigger>
            </TooltipTrigger>
            <TooltipContent>Export Trace</TooltipContent>
          </Tooltip>
        </TooltipProvider>
        <DropdownMenuContent align='end'>
          <DropdownMenuItem onClick={() => handleExport('csv')}>
            <FileSpreadsheet className='mr-2 h-4 w-4' />
            Export as CSV
          </DropdownMenuItem>
          <DropdownMenuItem onClick={() => handleExport('json')}>
            <FileJson className='mr-2 h-4 w-4' />
            Export as JSON
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    )
  }

  // Default button variant (with text)
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant='outline'
          size='sm'
          className={className}
          disabled={isExporting}
        >
          {isExporting ? (
            <Loader2 className='mr-2 h-4 w-4 animate-spin' />
          ) : (
            <Download className='mr-2 h-4 w-4' />
          )}
          Export
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align='end'>
        <DropdownMenuItem onClick={() => handleExport('csv')}>
          <FileSpreadsheet className='mr-2 h-4 w-4' />
          Export as CSV
        </DropdownMenuItem>
        <DropdownMenuItem onClick={() => handleExport('json')}>
          <FileJson className='mr-2 h-4 w-4' />
          Export as JSON
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
