'use client'

import * as React from 'react'
import {
  Clock,
  Timer,
  AlertCircle,
  CheckCircle2,
  MinusCircle,
  Hash,
  DollarSign,
  Brain,
  Bot,
  Layers,
  Globe,
  Cpu,
  Workflow,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import type { Span } from '../../data/schema'
import {
  detectSpanCategory,
  SPAN_CATEGORY_LABELS,
  SPAN_CATEGORY_COLORS,
  type SpanCategory,
} from '../../utils/span-type-detector'
import { formatDuration, formatCost } from '../../utils/format-helpers'

/**
 * AdaptiveMetricsGrid - Displays metrics based on span category
 *
 * Features:
 * - Category badge (LLM, AGENT, API, etc.)
 * - Base metrics: Duration, Start Time, Status
 * - Category-specific metrics based on available data
 */

interface AdaptiveMetricsGridProps {
  span: Span
  className?: string
}

interface MetricItem {
  label: string
  value: string | number | null | undefined
  icon?: React.ReactNode
}

/**
 * Format time for display (HH:mm:ss.SSS)
 */
function formatTime(date: Date | undefined | null): string {
  if (!date || !(date instanceof Date) || isNaN(date.getTime())) {
    return '-'
  }
  return date.toLocaleTimeString('en-US', {
    hour12: false,
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    fractionalSecondDigits: 3,
  })
}

/**
 * Get status display with icon and color
 */
function getStatusDisplay(statusCode: number): { label: string; icon: React.ReactNode; color: string } {
  switch (statusCode) {
    case 1: // OK
      return {
        label: 'OK',
        icon: <CheckCircle2 className='h-3 w-3' />,
        color: 'text-green-600 dark:text-green-400',
      }
    case 2: // ERROR
      return {
        label: 'ERROR',
        icon: <AlertCircle className='h-3 w-3' />,
        color: 'text-red-600 dark:text-red-400',
      }
    default: // UNSET (0)
      return {
        label: 'UNSET',
        icon: <MinusCircle className='h-3 w-3' />,
        color: 'text-muted-foreground',
      }
  }
}

/**
 * Get the icon for a span category
 */
function getCategoryIcon(category: SpanCategory): React.ReactNode {
  switch (category) {
    case 'llm':
      return <Brain className='h-3 w-3' />
    case 'agent':
      return <Bot className='h-3 w-3' />
    case 'batch':
      return <Layers className='h-3 w-3' />
    case 'api':
      return <Globe className='h-3 w-3' />
    case 'worker':
      return <Cpu className='h-3 w-3' />
    case 'pipeline':
      return <Workflow className='h-3 w-3' />
    default:
      return null
  }
}

/**
 * Get metrics for a span based on its category
 */
function getMetricsForSpan(span: Span, category: SpanCategory): MetricItem[][] {
  const status = getStatusDisplay(span.status_code)

  // Base metrics - always shown
  const baseMetrics: MetricItem[] = [
    {
      label: 'Duration',
      value: formatDuration(span.duration),
      icon: <Timer className='h-3 w-3 text-muted-foreground' />,
    },
    {
      label: 'Start Time',
      value: formatTime(span.start_time),
      icon: <Clock className='h-3 w-3 text-muted-foreground' />,
    },
    {
      label: 'Status',
      value: status.label,
      icon: <span className={status.color}>{status.icon}</span>,
    },
  ]

  const attrs = span.attributes ?? {}

  switch (category) {
    case 'llm': {
      const inputTokens = span.usage_details?.input ?? span.gen_ai_usage_input_tokens
      const outputTokens = span.usage_details?.output ?? span.gen_ai_usage_output_tokens
      const totalTokens = span.usage_details?.total ?? (inputTokens && outputTokens ? inputTokens + outputTokens : undefined)

      return [
        baseMetrics,
        [
          {
            label: 'Input Tokens',
            value: inputTokens ?? '-',
            icon: <Hash className='h-3 w-3 text-muted-foreground' />,
          },
          {
            label: 'Output Tokens',
            value: outputTokens ?? '-',
            icon: <Hash className='h-3 w-3 text-muted-foreground' />,
          },
          {
            label: 'Total Cost',
            value: formatCost(span.total_cost),
            icon: <DollarSign className='h-3 w-3 text-muted-foreground' />,
          },
        ],
        [
          {
            label: 'Model',
            value: (attrs['gen_ai.request.model'] as string) ?? span.model_name ?? span.gen_ai_request_model ?? '-',
          },
          {
            label: 'Provider',
            value: (attrs['gen_ai.system'] as string) ?? span.provider_name ?? span.gen_ai_provider_name ?? '-',
          },
        ],
      ]
    }

    case 'agent':
      return [
        baseMetrics,
        [
          { label: 'Agent', value: (attrs['agent.name'] as string) ?? '-' },
          { label: 'Iteration', value: (attrs['iteration'] as number) ?? '-' },
          { label: 'Tool', value: (attrs['tool.name'] as string) ?? '-' },
        ],
      ]

    case 'batch':
      return [
        baseMetrics,
        [
          { label: 'Batch ID', value: (attrs['batch.id'] as string) ?? '-' },
          { label: 'Batch Size', value: (attrs['batch.size'] as number) ?? '-' },
          { label: 'Item Index', value: (attrs['item.index'] as number) ?? '-' },
        ],
      ]

    case 'api':
      return [
        baseMetrics,
        [
          { label: 'Method', value: (attrs['http.method'] as string) ?? '-' },
          { label: 'URL', value: (attrs['http.url'] as string) ?? '-' },
        ],
      ]

    case 'worker':
      return [
        baseMetrics,
        [
          { label: 'Worker ID', value: (attrs['worker_id'] as string) ?? '-' },
          { label: 'Task', value: (attrs['task'] as string) ?? '-' },
          { label: 'Children', value: (attrs['child_count'] as number) ?? '-' },
        ],
      ]

    case 'pipeline':
      return [
        baseMetrics,
        [
          { label: 'Pipeline', value: (attrs['pipeline.name'] as string) ?? '-' },
          { label: 'Version', value: (attrs['pipeline.version'] as string) ?? '-' },
        ],
      ]

    case 'conversation':
      return [
        baseMetrics,
        [
          { label: 'Conversation ID', value: (attrs['conversation.id'] as string) ?? '-' },
          { label: 'Turn', value: (attrs['conversation.turn'] as number) ?? '-' },
          { label: 'Total Turns', value: (attrs['conversation.turns'] as number) ?? '-' },
        ],
      ]

    default: // generic
      return [baseMetrics]
  }
}

/**
 * Single metric cell
 */
function MetricCell({ item }: { item: MetricItem }) {
  return (
    <div className='flex flex-col gap-0.5'>
      <span className='text-xs text-muted-foreground flex items-center gap-1'>
        {item.icon}
        {item.label}
      </span>
      <span className='text-sm font-medium truncate' title={String(item.value ?? '-')}>
        {item.value ?? '-'}
      </span>
    </div>
  )
}

/**
 * Category badge component
 */
function CategoryBadge({ category }: { category: SpanCategory }) {
  const colors = SPAN_CATEGORY_COLORS[category]
  const label = SPAN_CATEGORY_LABELS[category]
  const icon = getCategoryIcon(category)

  return (
    <Badge
      variant='outline'
      className={cn(
        'text-xs font-medium px-2 py-0.5 gap-1',
        colors.bg,
        colors.text,
        colors.border
      )}
    >
      {icon}
      {label}
    </Badge>
  )
}

/**
 * AdaptiveMetricsGrid - Main component
 */
export function AdaptiveMetricsGrid({ span, className }: AdaptiveMetricsGridProps) {
  const category = detectSpanCategory(span.span_name, span.attributes)
  const metricsRows = getMetricsForSpan(span, category)

  return (
    <div className={cn('space-y-3', className)}>
      {/* Category Badge */}
      <div className='flex items-center gap-2'>
        <CategoryBadge category={category} />
      </div>

      {/* Metrics Grid */}
      <div className='bg-muted/30 rounded-lg p-3 space-y-3'>
        {metricsRows.map((row, rowIndex) => (
          <div
            key={rowIndex}
            className={cn(
              'grid gap-4',
              row.length === 2 ? 'grid-cols-2' : 'grid-cols-3'
            )}
          >
            {row.map((item, itemIndex) => (
              <MetricCell key={itemIndex} item={item} />
            ))}
          </div>
        ))}
      </div>
    </div>
  )
}

/**
 * Export category detection for use elsewhere
 */
export { detectSpanCategory, CategoryBadge }
