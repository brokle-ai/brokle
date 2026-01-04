'use client'

import { useMemo, useState, useCallback } from 'react'
import { Code, Copy, Check } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { cn } from '@/lib/utils'
import type { WidgetQuery, WidgetViewType } from '../../types'

interface QueryPreviewProps {
  query?: WidgetQuery
  className?: string
}

function useCopyToClipboard() {
  const [copied, setCopied] = useState(false)

  const copy = useCallback(async (text: string) => {
    try {
      await navigator.clipboard.writeText(text)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      console.error('Failed to copy to clipboard')
    }
  }, [])

  return { copied, copy }
}

/**
 * Generates a human-readable representation of the widget query.
 * This is for preview purposes only - actual execution happens on the backend.
 */
function generateQueryPreview(query: WidgetQuery): string {
  const lines: string[] = []

  // SELECT clause
  const selectParts: string[] = []

  if (query.measures && query.measures.length > 0) {
    selectParts.push(...query.measures.map((m) => `${m}(...)`))
  }

  if (query.dimensions && query.dimensions.length > 0) {
    selectParts.push(...query.dimensions)
  }

  if (selectParts.length > 0) {
    lines.push(`SELECT ${selectParts.join(', ')}`)
  } else {
    lines.push('SELECT *')
  }

  // FROM clause
  const viewLabels: Record<WidgetViewType, string> = {
    traces: 'otel_traces (root spans)',
    spans: 'otel_traces (all spans)',
    scores: 'quality_scores',
  }
  lines.push(`FROM ${viewLabels[query.view] ?? query.view}`)

  // WHERE clause
  if (query.filters && query.filters.length > 0) {
    const filterParts = query.filters.map((f) => {
      const value = typeof f.value === 'string' ? `'${f.value}'` : String(f.value)
      return `${f.field} ${f.operator} ${value}`
    })
    lines.push(`WHERE ${filterParts.join(' AND ')}`)
  }

  // GROUP BY clause
  if (query.dimensions && query.dimensions.length > 0) {
    lines.push(`GROUP BY ${query.dimensions.join(', ')}`)
  }

  // ORDER BY clause
  if (query.order_by) {
    const direction = query.order_dir ?? 'desc'
    lines.push(`ORDER BY ${query.order_by} ${direction.toUpperCase()}`)
  }

  // LIMIT clause
  if (query.limit) {
    lines.push(`LIMIT ${query.limit}`)
  }

  return lines.join('\n')
}

export function QueryPreview({ query, className }: QueryPreviewProps) {
  const { copied, copy } = useCopyToClipboard()

  const preview = useMemo(() => {
    if (!query) {
      return '-- No query configured\n-- Select a view and measures to begin'
    }
    return generateQueryPreview(query)
  }, [query])

  const handleCopy = () => {
    copy(preview)
  }

  return (
    <Card className={cn('', className)}>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-sm font-medium flex items-center gap-2">
            <Code className="h-4 w-4" />
            Query Preview
          </CardTitle>
          <Button
            variant="ghost"
            size="sm"
            onClick={handleCopy}
            className="h-7 px-2"
          >
            {copied ? (
              <Check className="h-3.5 w-3.5 text-green-500" />
            ) : (
              <Copy className="h-3.5 w-3.5" />
            )}
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        <pre className="text-xs font-mono bg-muted p-3 rounded-md overflow-x-auto whitespace-pre-wrap">
          {preview}
        </pre>
        <p className="text-xs text-muted-foreground mt-2">
          This is a simplified preview. The actual query is optimized for ClickHouse execution.
        </p>
      </CardContent>
    </Card>
  )
}
