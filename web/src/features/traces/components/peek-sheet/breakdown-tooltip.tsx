'use client'

import * as React from 'react'
import { Info } from 'lucide-react'
import {
  HoverCard,
  HoverCardContent,
  HoverCardTrigger,
} from '@/components/ui/hover-card'
import { cn } from '@/lib/utils'

interface BreakdownEntry {
  label: string
  value: number
}

interface BreakdownTooltipProps {
  /** The breakdown details - can be a Map, Record, or null */
  details: Map<string, number> | Record<string, number> | null | undefined
  /** Whether this is a cost breakdown (uses $ formatting) or usage (uses number formatting) */
  isCost?: boolean
  /** Optional total override (if not provided, calculates from details) */
  total?: number
  /** The trigger element */
  children: React.ReactNode
  /** Additional class name for the trigger */
  className?: string
}

/**
 * Format a number for display
 * - Cost: $0.001234 (6 decimals for small values)
 * - Usage: 1,234 (locale formatted)
 */
function formatValue(value: number, isCost: boolean): string {
  if (isCost) {
    // For very small costs, show more decimals
    if (value > 0 && value < 0.01) {
      return `$${value.toFixed(6)}`
    }
    return `$${value.toFixed(4)}`
  }
  return value.toLocaleString()
}

/**
 * Convert Map or Record to array of entries, sorted by value descending
 */
function getEntries(
  details: Map<string, number> | Record<string, number> | null | undefined
): BreakdownEntry[] {
  if (!details) return []

  const entries: BreakdownEntry[] = []

  if (details instanceof Map) {
    details.forEach((value, key) => {
      entries.push({ label: key, value })
    })
  } else {
    Object.entries(details).forEach(([key, value]) => {
      if (typeof value === 'number') {
        entries.push({ label: key, value })
      }
    })
  }

  // Sort by value descending
  return entries.sort((a, b) => b.value - a.value)
}

/**
 * Categorize entries into input, output, and other
 */
function categorizeEntries(entries: BreakdownEntry[]): {
  input: BreakdownEntry[]
  output: BreakdownEntry[]
  other: BreakdownEntry[]
} {
  const input: BreakdownEntry[] = []
  const output: BreakdownEntry[] = []
  const other: BreakdownEntry[] = []

  for (const entry of entries) {
    const label = entry.label.toLowerCase()
    if (label.includes('input') || label.includes('prompt')) {
      input.push(entry)
    } else if (label.includes('output') || label.includes('completion')) {
      output.push(entry)
    } else {
      other.push(entry)
    }
  }

  return { input, output, other }
}

/**
 * Render a section of the breakdown
 */
function BreakdownSection({
  title,
  entries,
  isCost,
  className,
}: {
  title: string
  entries: BreakdownEntry[]
  isCost: boolean
  className?: string
}) {
  if (entries.length === 0) return null

  const total = entries.reduce((sum, e) => sum + e.value, 0)

  return (
    <div className={cn('space-y-1', className)}>
      <div className='text-xs font-medium text-muted-foreground'>{title}</div>
      {entries.map((entry) => (
        <div key={entry.label} className='flex justify-between text-xs'>
          <span className='text-muted-foreground truncate max-w-[120px]' title={entry.label}>
            {entry.label}
          </span>
          <span className='font-mono ml-2'>{formatValue(entry.value, isCost)}</span>
        </div>
      ))}
      {entries.length > 1 && (
        <div className='flex justify-between text-xs font-medium border-t pt-1 mt-1'>
          <span>Subtotal</span>
          <span className='font-mono'>{formatValue(total, isCost)}</span>
        </div>
      )}
    </div>
  )
}

/**
 * BreakdownTooltip - Detailed cost/usage breakdown tooltip
 *
 * Shows a hover card with detailed breakdown of costs or token usage,
 * categorized into input, output, and other sections.
 */
export function BreakdownTooltip({
  details,
  isCost = false,
  total: totalOverride,
  children,
  className,
}: BreakdownTooltipProps) {
  const entries = React.useMemo(() => getEntries(details), [details])
  const { input, output, other } = React.useMemo(
    () => categorizeEntries(entries),
    [entries]
  )

  // Calculate total
  const calculatedTotal = entries.reduce((sum, e) => sum + e.value, 0)
  const total = totalOverride ?? calculatedTotal

  // If no details, just render children without tooltip
  if (entries.length === 0) {
    return <>{children}</>
  }

  return (
    <HoverCard openDelay={200} closeDelay={100}>
      <HoverCardTrigger asChild>
        <button
          type='button'
          className={cn(
            'inline-flex items-center gap-1 hover:opacity-80 transition-opacity cursor-help',
            className
          )}
        >
          {children}
          <Info className='h-3 w-3 text-muted-foreground' />
        </button>
      </HoverCardTrigger>
      <HoverCardContent
        align='start'
        className='w-64 p-3'
        side='bottom'
      >
        <div className='space-y-3'>
          <div className='text-sm font-medium'>
            {isCost ? 'Cost Breakdown' : 'Usage Breakdown'}
          </div>

          {input.length > 0 && (
            <BreakdownSection title='Input' entries={input} isCost={isCost} />
          )}

          {output.length > 0 && (
            <BreakdownSection title='Output' entries={output} isCost={isCost} />
          )}

          {other.length > 0 && (
            <BreakdownSection title='Other' entries={other} isCost={isCost} />
          )}

          {/* Total */}
          <div className='flex justify-between text-sm font-semibold border-t pt-2'>
            <span>Total</span>
            <span className='font-mono'>{formatValue(total, isCost)}</span>
          </div>
        </div>
      </HoverCardContent>
    </HoverCard>
  )
}

/**
 * Simplified breakdown badge with inline info icon
 * Use this to wrap existing badges
 */
export function WithBreakdown({
  details,
  isCost,
  total,
  children,
}: Omit<BreakdownTooltipProps, 'className'>) {
  return (
    <BreakdownTooltip details={details} isCost={isCost} total={total}>
      {children}
    </BreakdownTooltip>
  )
}
