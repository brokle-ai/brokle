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
import { toast } from 'sonner'
import { cn } from '@/lib/utils'
import type { TimeRange } from '@/components/shared/time-range-picker'

interface ExportUsageButtonProps {
  organizationId: string
  timeRange: TimeRange
  className?: string
}

type ExportFormat = 'csv' | 'json'

function getTimeRangeDates(timeRange: TimeRange): { from: string; to: string } {
  const now = new Date()
  let from: Date
  let to: Date = now

  if (timeRange.relative && timeRange.relative !== 'custom') {
    const days: Record<string, number> = {
      '1h': 1 / 24,
      '24h': 1,
      '7d': 7,
      '30d': 30,
      '90d': 90,
    }
    const daysAgo = days[timeRange.relative] || 30
    from = new Date(now.getTime() - daysAgo * 24 * 60 * 60 * 1000)
  } else if (timeRange.from && timeRange.to) {
    from = new Date(timeRange.from)
    to = new Date(timeRange.to)
  } else {
    from = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000)
  }

  return {
    from: from.toISOString(),
    to: to.toISOString(),
  }
}

export function ExportUsageButton({
  organizationId,
  timeRange,
  className,
}: ExportUsageButtonProps) {
  const [isExporting, setIsExporting] = React.useState(false)

  const handleExport = async (format: ExportFormat) => {
    setIsExporting(true)
    try {
      const { from, to } = getTimeRangeDates(timeRange)
      const params = new URLSearchParams({
        from,
        to,
        format,
      })

      const response = await fetch(
        `/api/v1/organizations/${organizationId}/usage/export?${params}`,
        {
          method: 'GET',
          credentials: 'include',
        }
      )

      if (!response.ok) {
        throw new Error('Export failed')
      }

      // Get filename from Content-Disposition header or generate one
      const contentDisposition = response.headers.get('Content-Disposition')
      let filename = `usage_export.${format}`
      if (contentDisposition) {
        const match = contentDisposition.match(/filename=([^;]+)/)
        if (match) {
          filename = match[1].trim()
        }
      }

      // Download the file
      const blob = await response.blob()
      const url = window.URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = filename
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      window.URL.revokeObjectURL(url)

      toast.success(`Usage data exported as ${format.toUpperCase()}`)
    } catch (error) {
      toast.error('Failed to export usage data')
    } finally {
      setIsExporting(false)
    }
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
