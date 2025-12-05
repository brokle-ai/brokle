'use client'

import * as React from 'react'
import { Button } from '@/components/ui/button'
import { ChevronDown, ChevronRight, Copy, Check } from 'lucide-react'
import { cn } from '@/lib/utils'

/**
 * PathValueTree - Expandable tree view for OTEL Map/Object data
 *
 * Features:
 * - Path | Value column layout
 * - Expandable/collapsible nested objects and arrays
 * - Type-based value coloring (string=green, number=blue, boolean=purple)
 * - Copy button on hover for any value
 * - Handles OTEL dotted keys (e.g., gen_ai.request.model)
 */

interface PathValueTreeProps {
  data: Record<string, unknown> | unknown[] | null | undefined
  maxInitialDepth?: number // Default: 1
  showCopyButtons?: boolean // Default: true
  className?: string
}

interface TreeNodeProps {
  path: string
  value: unknown
  depth: number
  maxInitialDepth: number
  showCopyButtons: boolean
  isLast: boolean
}

/**
 * Get display text and color for a value based on its type
 */
function getValueDisplay(value: unknown): { text: string; color: string; isExpandable: boolean } {
  if (value === null) {
    return { text: 'null', color: 'text-muted-foreground italic', isExpandable: false }
  }
  if (value === undefined) {
    return { text: 'undefined', color: 'text-muted-foreground italic', isExpandable: false }
  }
  if (typeof value === 'string') {
    // Truncate long strings
    const display = value.length > 100 ? `${value.slice(0, 100)}...` : value
    return { text: `"${display}"`, color: 'text-green-600 dark:text-green-400', isExpandable: false }
  }
  if (typeof value === 'number') {
    return { text: String(value), color: 'text-blue-600 dark:text-blue-400', isExpandable: false }
  }
  if (typeof value === 'boolean') {
    return { text: String(value), color: 'text-purple-600 dark:text-purple-400', isExpandable: false }
  }
  if (Array.isArray(value)) {
    if (value.length === 0) {
      return { text: '[]', color: 'text-muted-foreground', isExpandable: false }
    }
    return { text: `${value.length} items`, color: 'text-muted-foreground', isExpandable: true }
  }
  if (typeof value === 'object') {
    const keys = Object.keys(value as object)
    if (keys.length === 0) {
      return { text: '{}', color: 'text-muted-foreground', isExpandable: false }
    }
    return { text: `${keys.length} keys`, color: 'text-muted-foreground', isExpandable: true }
  }
  return { text: String(value), color: 'text-foreground', isExpandable: false }
}

/**
 * Copy button component with feedback
 */
function CopyButton({ value }: { value: string }) {
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
      className='h-5 w-5 opacity-0 group-hover:opacity-100 transition-opacity'
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

/**
 * Single tree node - recursive component
 */
function TreeNode({
  path,
  value,
  depth,
  maxInitialDepth,
  showCopyButtons,
  isLast,
}: TreeNodeProps) {
  const { text, color, isExpandable } = getValueDisplay(value)
  const [isOpen, setIsOpen] = React.useState(depth < maxInitialDepth)

  // Get children if expandable
  const children = React.useMemo(() => {
    if (!isExpandable) return []
    if (Array.isArray(value)) {
      return value.map((item, index) => ({
        key: `[${index}]`,
        value: item,
      }))
    }
    if (typeof value === 'object' && value !== null) {
      return Object.entries(value as Record<string, unknown>).map(([key, val]) => ({
        key,
        value: val,
      }))
    }
    return []
  }, [value, isExpandable])

  // Stringify value for copy
  const copyValue = React.useMemo(() => {
    if (typeof value === 'string') return value
    if (value === null || value === undefined) return String(value)
    return JSON.stringify(value, null, 2)
  }, [value])

  return (
    <div className={cn('text-sm', depth > 0 && 'ml-4')}>
      <div
        className={cn(
          'group flex items-center gap-1 py-0.5 rounded hover:bg-muted/50 cursor-default',
          isExpandable && 'cursor-pointer'
        )}
        onClick={isExpandable ? () => setIsOpen(!isOpen) : undefined}
      >
        {/* Expand/collapse icon */}
        <div className='w-4 flex-shrink-0'>
          {isExpandable ? (
            isOpen ? (
              <ChevronDown className='h-3 w-3 text-muted-foreground' />
            ) : (
              <ChevronRight className='h-3 w-3 text-muted-foreground' />
            )
          ) : null}
        </div>

        {/* Path/Key */}
        <span className='font-mono text-xs text-foreground min-w-0 flex-shrink-0'>
          {path}
        </span>

        {/* Separator */}
        <span className='text-muted-foreground mx-1'>:</span>

        {/* Value */}
        <span className={cn('font-mono text-xs truncate', color)}>{text}</span>

        {/* Copy button */}
        {showCopyButtons && <CopyButton value={copyValue} />}
      </div>

      {/* Children */}
      {isOpen && isExpandable && children.length > 0 && (
        <div className='border-l border-border/50 ml-2'>
          {children.map((child, index) => (
            <TreeNode
              key={child.key}
              path={child.key}
              value={child.value}
              depth={depth + 1}
              maxInitialDepth={maxInitialDepth}
              showCopyButtons={showCopyButtons}
              isLast={index === children.length - 1}
            />
          ))}
        </div>
      )}
    </div>
  )
}

/**
 * PathValueTree - Main component
 */
export function PathValueTree({
  data,
  maxInitialDepth = 1,
  showCopyButtons = true,
  className,
}: PathValueTreeProps) {
  if (data === null || data === undefined) {
    return (
      <div className={cn('text-sm text-muted-foreground italic p-2', className)}>
        No data
      </div>
    )
  }

  // Get root level entries
  const entries = React.useMemo(() => {
    if (Array.isArray(data)) {
      return data.map((item, index) => ({
        key: `[${index}]`,
        value: item,
      }))
    }
    return Object.entries(data).map(([key, value]) => ({
      key,
      value,
    }))
  }, [data])

  if (entries.length === 0) {
    return (
      <div className={cn('text-sm text-muted-foreground italic p-2', className)}>
        Empty {Array.isArray(data) ? 'array' : 'object'}
      </div>
    )
  }

  return (
    <div className={cn('font-mono', className)}>
      {entries.map((entry, index) => (
        <TreeNode
          key={entry.key}
          path={entry.key}
          value={entry.value}
          depth={0}
          maxInitialDepth={maxInitialDepth}
          showCopyButtons={showCopyButtons}
          isLast={index === entries.length - 1}
        />
      ))}
    </div>
  )
}

/**
 * FlatKeyValueList - Simple flat key-value display (no nesting)
 * Good for OTEL dotted keys like gen_ai.request.model
 */
export function FlatKeyValueList({
  data,
  showCopyButtons = true,
  className,
}: {
  data: Record<string, unknown> | null | undefined
  showCopyButtons?: boolean
  className?: string
}) {
  if (!data || Object.keys(data).length === 0) {
    return (
      <div className={cn('text-sm text-muted-foreground italic p-2', className)}>
        No attributes
      </div>
    )
  }

  // Sort keys alphabetically for consistent display
  const sortedEntries = Object.entries(data).sort(([a], [b]) => a.localeCompare(b))

  return (
    <div className={cn('space-y-0.5', className)}>
      {sortedEntries.map(([key, value]) => {
        const { text, color } = getValueDisplay(value)
        const copyValue = typeof value === 'string' ? value : JSON.stringify(value)

        return (
          <div
            key={key}
            className='group flex items-center gap-2 py-0.5 px-1 rounded hover:bg-muted/50'
          >
            <span className='font-mono text-xs text-muted-foreground min-w-0 flex-shrink-0'>
              {key}
            </span>
            <span className='text-muted-foreground'>:</span>
            <span className={cn('font-mono text-xs truncate flex-1', color)}>{text}</span>
            {showCopyButtons && <CopyButton value={copyValue} />}
          </div>
        )
      })}
    </div>
  )
}
