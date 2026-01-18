'use client'

import { memo } from 'react'
import { Handle, Position, type NodeProps } from 'reactflow'
import {
  Brain,
  Bot,
  Layers,
  MessageSquare,
  Workflow,
  Cpu,
  Globe,
  Circle,
  Clock,
  AlertCircle,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import type { Span } from '../../data/schema'
import type { SpanCategory } from '../../utils/span-type-detector'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'

/**
 * Data structure for span nodes in the graph
 */
export interface SpanNodeData {
  span: Span
  category: SpanCategory
  label: string
  duration?: string
  tokens?: number
  cost?: number
  hasError: boolean
  isSelected: boolean
  model?: string
  statusCode?: string
}

/**
 * Icon mapping for span categories
 */
const CategoryIcon: Record<SpanCategory, React.ComponentType<{ className?: string }>> = {
  llm: Brain,
  agent: Bot,
  batch: Layers,
  conversation: MessageSquare,
  pipeline: Workflow,
  worker: Cpu,
  api: Globe,
  generic: Circle,
}

/**
 * Node styling based on span category
 * Matches the existing SPAN_CATEGORY_COLORS from span-type-detector.ts
 */
const categoryStyles: Record<SpanCategory, string> = {
  llm: 'border-purple-400 dark:border-purple-500 bg-purple-50 dark:bg-purple-950/40',
  agent: 'border-orange-400 dark:border-orange-500 bg-orange-50 dark:bg-orange-950/40',
  batch: 'border-cyan-400 dark:border-cyan-500 bg-cyan-50 dark:bg-cyan-950/40',
  conversation: 'border-blue-400 dark:border-blue-500 bg-blue-50 dark:bg-blue-950/40',
  pipeline: 'border-green-400 dark:border-green-500 bg-green-50 dark:bg-green-950/40',
  worker: 'border-teal-400 dark:border-teal-500 bg-teal-50 dark:bg-teal-950/40',
  api: 'border-amber-400 dark:border-amber-500 bg-amber-50 dark:bg-amber-950/40',
  generic: 'border-gray-300 dark:border-gray-600 bg-muted',
}

/**
 * Icon color classes for each category
 */
const iconColorClasses: Record<SpanCategory, string> = {
  llm: 'text-purple-600 dark:text-purple-400',
  agent: 'text-orange-600 dark:text-orange-400',
  batch: 'text-cyan-600 dark:text-cyan-400',
  conversation: 'text-blue-600 dark:text-blue-400',
  pipeline: 'text-green-600 dark:text-green-400',
  worker: 'text-teal-600 dark:text-teal-400',
  api: 'text-amber-600 dark:text-amber-400',
  generic: 'text-muted-foreground',
}

/**
 * Truncate span name for display
 */
function truncateName(name: string, maxLength = 25): string {
  if (name.length <= maxLength) return name
  return name.slice(0, maxLength - 3) + '...'
}

/**
 * Format duration from nanoseconds to human readable
 */
function formatDuration(durationNs: number | undefined): string | undefined {
  if (!durationNs) return undefined
  const ms = durationNs / 1_000_000
  if (ms < 1000) return `${Math.round(ms)}ms`
  return `${(ms / 1000).toFixed(1)}s`
}

/**
 * SpanNode - Custom node component for displaying spans in the graph
 *
 * Features:
 * - Category-based icon and colors
 * - Error indicator (red border)
 * - Selected state (ring highlight)
 * - Compact metrics display
 * - Truncated name with tooltip
 */
function SpanNodeComponent({ data, selected }: NodeProps<SpanNodeData>) {
  const Icon = CategoryIcon[data.category]
  const isSelected = data.isSelected || selected
  const duration = formatDuration(data.span.duration)
  const calculatedTokens = (data.span.gen_ai_usage_input_tokens || 0) +
    (data.span.gen_ai_usage_output_tokens || 0)
  const totalTokens = data.tokens ?? (calculatedTokens > 0 ? calculatedTokens : undefined)

  return (
    <>
      {/* Target handle (top) */}
      <Handle
        type="target"
        position={Position.Top}
        className="!w-2 !h-2 !bg-muted-foreground/50"
      />

      <Tooltip>
        <TooltipTrigger asChild>
          <div
            className={cn(
              'px-3 py-2 rounded-lg border-2 min-w-[160px] max-w-[220px]',
              'cursor-pointer transition-all duration-150',
              'hover:shadow-md',
              categoryStyles[data.category],
              // Error state
              data.hasError && 'border-red-500 dark:border-red-400 bg-red-50 dark:bg-red-950/40',
              // Selected state
              isSelected && 'ring-2 ring-primary ring-offset-2 ring-offset-background'
            )}
          >
            {/* Header with icon and name */}
            <div className="flex items-center gap-2">
              <Icon
                className={cn(
                  'h-4 w-4 flex-shrink-0',
                  data.hasError
                    ? 'text-red-600 dark:text-red-400'
                    : iconColorClasses[data.category]
                )}
              />
              <span className="text-xs font-medium truncate flex-1 text-foreground">
                {truncateName(data.label)}
              </span>
              {data.hasError && (
                <AlertCircle className="h-3.5 w-3.5 text-red-500 dark:text-red-400 flex-shrink-0" />
              )}
            </div>

            {/* Metrics row */}
            <div className="flex items-center gap-3 mt-1.5 text-[10px] text-muted-foreground">
              {duration && (
                <div className="flex items-center gap-1">
                  <Clock className="h-3 w-3" />
                  <span>{duration}</span>
                </div>
              )}
              {totalTokens && totalTokens > 0 && (
                <span>{totalTokens.toLocaleString()} tok</span>
              )}
              {data.model && (
                <span className="truncate max-w-[80px]">{data.model}</span>
              )}
            </div>
          </div>
        </TooltipTrigger>
        <TooltipContent side="top" className="max-w-[300px]">
          <div className="space-y-1">
            <p className="font-medium">{data.label}</p>
            {data.span.span_id && (
              <p className="text-xs text-muted-foreground font-mono">
                {data.span.span_id.slice(0, 16)}...
              </p>
            )}
            {data.statusCode && (
              <p className="text-xs">
                Status: <span className={data.hasError ? 'text-red-500' : 'text-green-500'}>
                  {data.statusCode}
                </span>
              </p>
            )}
          </div>
        </TooltipContent>
      </Tooltip>

      {/* Source handle (bottom) */}
      <Handle
        type="source"
        position={Position.Bottom}
        className="!w-2 !h-2 !bg-muted-foreground/50"
      />
    </>
  )
}

// Memoize for performance
export const SpanNode = memo(SpanNodeComponent)
