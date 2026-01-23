'use client'

import { useState } from 'react'
import { ChevronRight, ChevronDown, Copy, Check } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import type { RowHeight } from './types'

interface JsonCellProps {
  value: Record<string, unknown> | unknown[] | undefined | null
  rowHeight?: RowHeight
  className?: string
}

const ROW_HEIGHT_CONFIG: Record<RowHeight, { maxChars: number; showPreview: boolean }> = {
  small: { maxChars: 50, showPreview: false },
  medium: { maxChars: 100, showPreview: true },
  large: { maxChars: 500, showPreview: true },
}

export function JsonCell({ value, rowHeight = 'medium', className }: JsonCellProps) {
  const [copied, setCopied] = useState(false)
  const [isOpen, setIsOpen] = useState(false)

  if (!value) {
    return <span className="text-muted-foreground">-</span>
  }

  const jsonString = JSON.stringify(value, null, 2)
  const compactString = JSON.stringify(value)
  const config = ROW_HEIGHT_CONFIG[rowHeight]
  const isTruncated = compactString.length > config.maxChars
  const displayString = isTruncated
    ? compactString.slice(0, config.maxChars) + '...'
    : compactString

  const handleCopy = async (e: React.MouseEvent) => {
    e.stopPropagation()
    await navigator.clipboard.writeText(jsonString)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <Popover open={isOpen} onOpenChange={setIsOpen}>
      <PopoverTrigger asChild>
        <button
          className={cn(
            'flex items-center gap-1 text-left max-w-full',
            'hover:bg-muted/50 rounded px-1 -mx-1 transition-colors',
            className
          )}
        >
          {isTruncated && (
            isOpen ? (
              <ChevronDown className="h-3 w-3 shrink-0 text-muted-foreground" />
            ) : (
              <ChevronRight className="h-3 w-3 shrink-0 text-muted-foreground" />
            )
          )}
          <code className="text-xs bg-muted px-2 py-1 rounded font-mono truncate">
            {displayString}
          </code>
        </button>
      </PopoverTrigger>
      <PopoverContent
        className="w-[400px] max-h-[400px] p-0"
        align="start"
        side="bottom"
      >
        <div className="flex items-center justify-between border-b px-3 py-2 bg-muted/50">
          <span className="text-xs font-medium text-muted-foreground">
            JSON Preview
          </span>
          <Button
            variant="ghost"
            size="sm"
            className="h-6 px-2"
            onClick={handleCopy}
          >
            {copied ? (
              <Check className="h-3 w-3 text-green-500" />
            ) : (
              <Copy className="h-3 w-3" />
            )}
          </Button>
        </div>
        <div className="overflow-auto max-h-[350px] p-3">
          <JsonTree data={value} />
        </div>
      </PopoverContent>
    </Popover>
  )
}

interface JsonTreeProps {
  data: unknown
  level?: number
}

function JsonTree({ data, level = 0 }: JsonTreeProps) {
  const [expanded, setExpanded] = useState(level < 2)

  if (data === null) {
    return <span className="text-orange-500">null</span>
  }

  if (typeof data === 'boolean') {
    return <span className="text-purple-500">{String(data)}</span>
  }

  if (typeof data === 'number') {
    return <span className="text-blue-500">{data}</span>
  }

  if (typeof data === 'string') {
    return <span className="text-green-600">&quot;{data}&quot;</span>
  }

  if (Array.isArray(data)) {
    if (data.length === 0) {
      return <span className="text-muted-foreground">[]</span>
    }

    return (
      <div className="font-mono text-xs">
        <button
          onClick={() => setExpanded(!expanded)}
          className="flex items-center gap-1 hover:bg-muted/50 rounded"
          aria-expanded={expanded}
          aria-label={`${expanded ? 'Collapse' : 'Expand'} array with ${data.length} items`}
        >
          {expanded ? (
            <ChevronDown className="h-3 w-3" />
          ) : (
            <ChevronRight className="h-3 w-3" />
          )}
          <span className="text-muted-foreground">
            Array[{data.length}]
          </span>
        </button>
        {expanded && (
          <div className="ml-4 border-l pl-2 mt-1 space-y-1">
            {data.map((item, index) => (
              <div key={index} className="flex gap-2">
                <span className="text-muted-foreground shrink-0">{index}:</span>
                <JsonTree data={item} level={level + 1} />
              </div>
            ))}
          </div>
        )}
      </div>
    )
  }

  if (typeof data === 'object') {
    const entries = Object.entries(data)
    if (entries.length === 0) {
      return <span className="text-muted-foreground">{'{}'}</span>
    }

    return (
      <div className="font-mono text-xs">
        <button
          onClick={() => setExpanded(!expanded)}
          className="flex items-center gap-1 hover:bg-muted/50 rounded"
          aria-expanded={expanded}
          aria-label={`${expanded ? 'Collapse' : 'Expand'} object with ${entries.length} properties`}
        >
          {expanded ? (
            <ChevronDown className="h-3 w-3" />
          ) : (
            <ChevronRight className="h-3 w-3" />
          )}
          <span className="text-muted-foreground">
            Object{'{'}
            {entries.length}
            {'}'}
          </span>
        </button>
        {expanded && (
          <div className="ml-4 border-l pl-2 mt-1 space-y-1">
            {entries.map(([key, value]) => (
              <div key={key} className="flex gap-2">
                <span className="text-cyan-600 shrink-0">{key}:</span>
                <JsonTree data={value} level={level + 1} />
              </div>
            ))}
          </div>
        )}
      </div>
    )
  }

  return <span className="text-muted-foreground">{String(data)}</span>
}
