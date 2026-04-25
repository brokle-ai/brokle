import * as React from 'react'
import { Download, FileJson, FileSpreadsheet, Loader2 } from 'lucide-react'
import { toast } from 'sonner'
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
import type { Span, TraceDetail } from '../api/types'
import { downloadFile, traceToCSV, traceToJSON } from '../utils/export-utils'

type ExportFormat = 'csv' | 'json'

interface ExportTraceButtonProps {
  trace: TraceDetail
  spans: Span[]
  variant?: 'icon' | 'default'
  className?: string
}

/**
 * ExportTraceButton — drop-down for client-side CSV / JSON export of a
 * trace's span tree. The backend has no `GET /traces/{id}/export`
 * endpoint, so we serialise locally; this matches web/'s parity
 * implementation.
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
        downloadFile(
          traceToCSV(trace, spans),
          filename,
          'text/csv;charset=utf-8',
        )
      } else {
        downloadFile(traceToJSON(trace, spans), filename, 'application/json')
      }

      toast.success(`Trace exported as ${format.toUpperCase()}`)
    } catch (err) {
      console.error('Export failed:', err)
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
                  variant="ghost"
                  size="icon"
                  className="h-8 w-8"
                  disabled={isExporting}
                >
                  {isExporting ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    <Download className="h-4 w-4" />
                  )}
                </Button>
              </DropdownMenuTrigger>
            </TooltipTrigger>
            <TooltipContent>Export Trace</TooltipContent>
          </Tooltip>
        </TooltipProvider>
        <DropdownMenuContent align="end">
          <DropdownMenuItem onClick={() => handleExport('csv')}>
            <FileSpreadsheet className="mr-2 h-4 w-4" />
            Export as CSV
          </DropdownMenuItem>
          <DropdownMenuItem onClick={() => handleExport('json')}>
            <FileJson className="mr-2 h-4 w-4" />
            Export as JSON
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    )
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="outline"
          size="sm"
          className={className}
          disabled={isExporting}
        >
          {isExporting ? (
            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
          ) : (
            <Download className="mr-2 h-4 w-4" />
          )}
          Export
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem onClick={() => handleExport('csv')}>
          <FileSpreadsheet className="mr-2 h-4 w-4" />
          Export as CSV
        </DropdownMenuItem>
        <DropdownMenuItem onClick={() => handleExport('json')}>
          <FileJson className="mr-2 h-4 w-4" />
          Export as JSON
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
