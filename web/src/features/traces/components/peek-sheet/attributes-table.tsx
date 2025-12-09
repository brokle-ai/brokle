'use client'

import * as React from 'react'
import { Copy, Check, ChevronDown, ChevronRight } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

// ============================================================================
// Types
// ============================================================================

interface AttributesTableProps {
  data: Record<string, any> | undefined | null
  className?: string
  emptyMessage?: string
}

// ============================================================================
// Copy Button Component
// ============================================================================

function CopyButton({ value, className }: { value: string; className?: string }) {
  const [copied, setCopied] = React.useState(false)

  const handleCopy = async (e: React.MouseEvent) => {
    e.stopPropagation()
    await navigator.clipboard.writeText(value)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <Button
      variant='ghost'
      size='icon'
      className={cn('h-5 w-5 opacity-0 group-hover:opacity-100 transition-opacity', className)}
      onClick={handleCopy}
      title='Copy value'
    >
      {copied ? (
        <Check className='h-3 w-3 text-green-500' />
      ) : (
        <Copy className='h-3 w-3 text-muted-foreground' />
      )}
    </Button>
  )
}

// ============================================================================
// Value Renderer - Handles different value types
// ============================================================================

function ValueRenderer({ value }: { value: any }) {
  if (value === null) {
    return <span className='text-muted-foreground italic'>null</span>
  }

  if (value === undefined) {
    return <span className='text-muted-foreground italic'>undefined</span>
  }

  if (typeof value === 'boolean') {
    return (
      <span className={cn('font-mono text-sm', value ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400')}>
        {String(value)}
      </span>
    )
  }

  if (typeof value === 'number') {
    return <span className='font-mono text-sm text-blue-600 dark:text-blue-400'>{value}</span>
  }

  if (typeof value === 'string') {
    // Check if it's a URL
    if (value.startsWith('http://') || value.startsWith('https://')) {
      return (
        <a
          href={value}
          target='_blank'
          rel='noopener noreferrer'
          className='font-mono text-sm text-primary underline underline-offset-2 hover:text-primary/80 break-all'
        >
          {value}
        </a>
      )
    }
    return <span className='font-mono text-sm text-foreground break-all'>{value}</span>
  }

  if (Array.isArray(value)) {
    if (value.length === 0) {
      return <span className='text-muted-foreground italic'>[]</span>
    }
    return (
      <span className='font-mono text-sm text-foreground'>
        [{value.length} items]
      </span>
    )
  }

  if (typeof value === 'object') {
    const keys = Object.keys(value)
    if (keys.length === 0) {
      return <span className='text-muted-foreground italic'>{'{}'}</span>
    }
    return (
      <span className='font-mono text-sm text-foreground'>
        {'{' + keys.length + ' keys}'}
      </span>
    )
  }

  return <span className='font-mono text-sm text-foreground'>{String(value)}</span>
}

// ============================================================================
// Expandable Row - For nested objects/arrays
// ============================================================================

interface ExpandableRowProps {
  attrKey: string
  value: any
  depth?: number
}

function ExpandableRow({ attrKey, value, depth = 0 }: ExpandableRowProps) {
  const [isExpanded, setIsExpanded] = React.useState(false)
  const isExpandable = typeof value === 'object' && value !== null && (Array.isArray(value) ? value.length > 0 : Object.keys(value).length > 0)

  const stringValue = React.useMemo(() => {
    try {
      return JSON.stringify(value)
    } catch {
      return String(value)
    }
  }, [value])

  if (!isExpandable) {
    return (
      <div
        className={cn(
          'group flex items-center gap-3 py-2 px-3 hover:bg-muted/50 rounded-md transition-colors',
          depth > 0 && 'ml-4 border-l border-border/50'
        )}
      >
        {/* Spacer for alignment with expandable rows */}
        <div className='w-4 flex-shrink-0' />
        <span className='text-sm text-muted-foreground font-mono min-w-[180px] flex-shrink-0 truncate' title={attrKey}>
          {attrKey}
        </span>
        <div className='flex-1 min-w-0'>
          <ValueRenderer value={value} />
        </div>
        <CopyButton value={stringValue} />
      </div>
    )
  }

  return (
    <div className={cn(depth > 0 && 'ml-4 border-l border-border/50')}>
      <div
        className='group flex items-center gap-3 py-2 px-3 hover:bg-muted/50 rounded-md cursor-pointer transition-colors'
        onClick={() => setIsExpanded(!isExpanded)}
      >
        <div className='w-4 flex-shrink-0'>
          {isExpanded ? (
            <ChevronDown className='h-4 w-4 text-muted-foreground' />
          ) : (
            <ChevronRight className='h-4 w-4 text-muted-foreground' />
          )}
        </div>
        <span className='text-sm text-muted-foreground font-mono min-w-[180px] flex-shrink-0 truncate' title={attrKey}>
          {attrKey}
        </span>
        <div className='flex-1 min-w-0'>
          <ValueRenderer value={value} />
        </div>
        <CopyButton value={stringValue} />
      </div>

      {isExpanded && (
        <div className='pb-1'>
          {Array.isArray(value) ? (
            value.map((item, index) => (
              <ExpandableRow key={index} attrKey={`[${index}]`} value={item} depth={depth + 1} />
            ))
          ) : (
            Object.entries(value).map(([key, val]) => (
              <ExpandableRow key={key} attrKey={key} value={val} depth={depth + 1} />
            ))
          )}
        </div>
      )}
    </div>
  )
}

// ============================================================================
// AttributesTable - Main Component
// ============================================================================

/**
 * AttributesTable - Clean key-value table for OTEL attributes
 *
 * Features:
 * - Key-value table layout (like OpenLIT screenshot)
 * - Expandable nested objects/arrays
 * - Copy button on hover
 * - Syntax highlighting by type (strings, numbers, booleans)
 * - Clickable URLs
 */
export function AttributesTable({
  data,
  className,
  emptyMessage = 'No attributes',
}: AttributesTableProps) {
  if (!data || Object.keys(data).length === 0) {
    return (
      <div className={cn('py-6 text-center', className)}>
        <p className='text-sm text-muted-foreground italic'>{emptyMessage}</p>
      </div>
    )
  }

  // Sort keys alphabetically for consistent display
  const sortedKeys = Object.keys(data).sort()

  return (
    <div className={cn('space-y-0.5', className)}>
      {/* Header row */}
      <div className='flex items-center gap-3 py-1.5 px-3 text-xs text-muted-foreground font-medium border-b border-border/50'>
        <div className='w-4 flex-shrink-0' />
        <span className='min-w-[180px] flex-shrink-0'>Key</span>
        <span className='flex-1'>Value</span>
        <div className='w-5' />
      </div>

      {/* Attribute rows */}
      {sortedKeys.map((key) => (
        <ExpandableRow key={key} attrKey={key} value={data[key]} />
      ))}
    </div>
  )
}
